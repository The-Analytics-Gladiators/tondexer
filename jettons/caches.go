package jettons

import (
	"TonArb/models"
	"TonArb/persistence"
	"context"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/eko/gocache/lib/v4/cache"
	gocache_store "github.com/eko/gocache/store/go_cache/v4"
	gocache "github.com/patrickmn/go-cache"
	"log"
	"time"
)

func initCache[T any](
	table string,
	cacheFunction func(key any) (*T, error),
	batchAppendFunc func(batch driver.Batch, model *T) error,
	initCacheFunction func(cacheManager *cache.LoadableCache[any]) error) (*cache.LoadableCache[any], error) {

	gocacheClient := gocache.New(gocache.NoExpiration, gocache.NoExpiration)
	gocacheStore := gocache_store.NewGoCache(gocacheClient)

	chChannel := make(chan *T)
	ticker := time.NewTicker(10 * time.Second)

	loadFunction := func(ctx context.Context, key any) (any, error) {
		result, err := cacheFunction(key)
		if result != nil {
			go func() { chChannel <- result }()
		}
		return result, err
	}

	var toPersist []*T

	if table != "" {
		go func() {
			for {
				select {
				case model := <-chChannel:
					toPersist = append(toPersist, model)
				case <-ticker.C:
					if len(toPersist) > 0 {
						if e := persistence.WriteToClickhouse(toPersist, table, batchAppendFunc); e == nil {
							toPersist = []*T{}
						}
					}
				}
			}
		}()
	}

	// any because go-cache is supporting only any
	cacheManager := cache.NewLoadable[any](loadFunction, gocacheStore)

	if e := initCacheFunction(cacheManager); e != nil {
		return nil, e
	}

	return cacheManager, nil
}

func InitJettonInfoCache() (*cache.LoadableCache[any], error) {
	return initCache[ChainTokenInfo](
		"clickhouse_jetton",
		func(key any) (*ChainTokenInfo, error) {
			return JettonInfoByMaster(key.(string))
		},
		func(batch driver.Batch, model *ChainTokenInfo) error {
			return batch.Append(
				model.Name,
				model.Symbol,
				model.JettonAddress,
				uint64(model.Decimals),
			)
		},
		func(cacheManager *cache.LoadableCache[any]) error {
			chJettons, e := persistence.ReadClickhouseJettons()
			if e != nil {
				return e
			}

			for _, jetton := range chJettons {
				if e := cacheManager.Set(context.Background(), jetton.Master, &ChainTokenInfo{
					Name:          jetton.Name,
					Symbol:        jetton.Symbol,
					Decimals:      uint(jetton.Decimals),
					JettonAddress: jetton.Master,
				}); e != nil {
					log.Printf("Unable to set jetton cache entry %v \n", jetton)
				}
			}
			return nil
		})
}

func InitUsdRateCache() (*cache.LoadableCache[any], error) {
	return initCache[float64](
		"",
		func(key any) (*float64, error) {
			jettonInfo, e := JettonInfoFromMasterPageRetries(key.(string), 4)
			if e != nil {
				return nil, e
			}
			return &jettonInfo.TokenToUsd, nil
		},
		func(batch driver.Batch, model *float64) error { return nil },
		func(cacheManager *cache.LoadableCache[any]) error {
			go func() {
				walletToMasters, e := persistence.ReadWalletMasters()
				if e != nil {
					return
				}
				log.Printf("Warming up rate cache with %v entities \n", len(walletToMasters))
				for _, walletToMaster := range walletToMasters {

					res, e := cacheManager.Get(context.Background(), walletToMaster.Master)
					if e == nil {
						log.Printf("For master %v rate is %v \n", walletToMaster.Master, *res.(*float64))
					} else {
						log.Printf("Nil rate for %v master \n", walletToMaster.Master)
					}
				}
			}()

			return nil
		},
	)
}

func InitWalletJettonCache() (*cache.LoadableCache[any], error) {
	return initCache[models.WalletJetton](
		"wallet_to_master",
		func(key any) (*models.WalletJetton, error) {
			tonApi, e := GetTonApi()
			if e != nil {
				return nil, e
			}
			master, e := tonApi.MasterByWallet(key.(string))
			if e != nil {
				return nil, e
			}
			return &models.WalletJetton{
				Wallet: key.(string),
				Master: master.String()}, nil
		},
		func(batch driver.Batch, model *models.WalletJetton) error {
			return batch.Append(
				model.Wallet,
				model.Master,
			)
		},
		func(cacheManager *cache.LoadableCache[any]) error {
			walletToMasters, e := persistence.ReadWalletMasters()
			if e != nil {
				return e
			}

			for _, walletToMaster := range walletToMasters {
				cacheManager.Set(context.Background(), walletToMaster.Wallet, &walletToMaster)
			}
			return nil
		},
	)
}
