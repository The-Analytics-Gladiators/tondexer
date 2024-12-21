package main

import (
	"context"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/tonkeeper/tonapi-go"
	"log"
	"os"
	"time"
	"tondexer/arbitrage"
	"tondexer/common"
	"tondexer/core"
	"tondexer/dedust"
	"tondexer/jettons"
	"tondexer/models"
	"tondexer/persistence"
	"tondexer/stonfi"
	"tondexer/stonfiv2"
)

type Config struct {
	ConsoleToken      string   `yaml:"console_token" env:"CONSOLE_TOKEN" env-default:""`
	DbHost            string   `yaml:"db_host" env:"DB_HOST" env-default:"localhost"`
	DbPort            uint     `yaml:"db_port" env:"DB_PORT" env-default:"9000"`
	DbUser            string   `yaml:"db_user" env:"DB_USER" env-default:"default"`
	DbPassword        string   `yaml:"db_password" env:"DB_PASSWORD" env-default:""`
	DbName            string   `yaml:"db_name" env:"DB_NAME" env-default:"default"`
	StonfiV1Addresses []string `yaml:"stonfiv1_addresses" env:"STONFIV1_ADDRESSES" env-default:""`
	StonfiV2Addresses []string `yaml:"stonfiv2_addresses" env:"STONFIV2_ADDRESSES" env-default:""`
	DedustAddresses   []string `yaml:"dedust_addresses" env:"DEDUST_ADDRESSES" env-default:""`
}

func subscribeToAccounts(streamingApi *tonapi.StreamingAPI, accounts []string, incomingTransactionsChannel chan string) {
	for {
		e := streamingApi.WebsocketHandleRequests(context.Background(), func(ws tonapi.Websocket) error {
			ws.SetTransactionHandler(func(data tonapi.TransactionEventData) {
				//log.Printf("New tx with hash: %v lt: %v \n", data.TxHash, data.Lt)
				go func() {
					incomingTransactionsChannel <- data.TxHash
				}()
			})
			if err := ws.SubscribeToTransactions(accounts, nil); err != nil {
				return err
			}

			return nil
		})
		if e != nil {
			log.Printf("Streaming failed! for accounts %v: %v \n", accounts, e)
		}
		time.Sleep(1 * time.Second)
	}
}

func swapInfoWithDex(infos []*models.SwapInfo, dex string) []core.Pair[*models.SwapInfo, string] {
	return common.Map(infos, func(swapInfo *models.SwapInfo) core.Pair[*models.SwapInfo, string] {
		return core.Pair[*models.SwapInfo, string]{
			First:  swapInfo,
			Second: dex,
		}
	})
}

func main() {
	var cfg Config

	freeConsoleClient, _ := tonapi.New() // free one for the rates
	freeConsoleApi := core.TonConsoleApi{Client: freeConsoleClient}

	if err := cleanenv.ReadConfig(os.Args[1], &cfg); err != nil {
		panic(err)
	}
	dbConfig := core.DbConfig{
		DbHost:     cfg.DbHost,
		DbPort:     cfg.DbPort,
		DbUser:     cfg.DbUser,
		DbPassword: cfg.DbPassword,
		DbName:     cfg.DbName,
	}

	jettonInfoCache, e := jettons.InitJettonInfoCache(&dbConfig, &freeConsoleApi)
	if e != nil {
		panic(e)
	}

	walletMasterCache, e := jettons.InitWalletJettonCache(&dbConfig)
	if e != nil {
		panic(e)
	}

	usdRateCache, err := jettons.InitUsdRateCache(&dbConfig, &freeConsoleApi)
	if err != nil {
		panic(e)
	}

	streamingApi := tonapi.NewStreamingAPI(tonapi.WithStreamingToken(cfg.ConsoleToken))

	stonfiV1Accounts := []string{stonfi.StonfiRouter}

	client, _ := tonapi.New(tonapi.WithToken(cfg.ConsoleToken))
	consoleApi := &core.TonConsoleApi{Client: client}
	incomingTransactionsChannel := make(chan string)

	allSubscribers := append(stonfiV1Accounts, append(stonfiv2.Routers, dedust.VaultAddresses...)...)
	log.Printf("Subscribing to %v addresses... \n", len(allSubscribers))
	for _, chunk := range common.ChunkArray(allSubscribers, 10) {
		go subscribeToAccounts(streamingApi, chunk, incomingTransactionsChannel)
	}

	readyTransactionsChannel := make(chan []string)

	transactionsWaitingList := &core.WaitingList[string]{
		ExpirationSeconds: 70 * time.Second,
	}
	transactionsTicker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case transactionHash := <-incomingTransactionsChannel:
				transactionsWaitingList.Add(transactionHash)
			case <-transactionsTicker.C:
				evicted := transactionsWaitingList.Evict()
				go func() { readyTransactionsChannel <- evicted }()
				log.Printf("%v transaction hashes was sent for processing\n", len(evicted))
			}
		}
	}()

	walletToMasterJettonCacheFunc := func(wallet string) *models.ChainTokenInfo {
		master, e := walletMasterCache.Get(context.Background(), wallet)
		if e != nil {
			return nil
		}

		info, e := jettonInfoCache.Get(context.Background(), master.(*models.WalletJetton).Master)
		if e != nil {
			log.Printf("Unable to get jetton info for %v \n", master.(*models.WalletJetton).Master)
			return nil
		}
		return info.(*models.ChainTokenInfo)
	}

	masterJettonCacheFunc := func(master string) *models.ChainTokenInfo {
		info, e := jettonInfoCache.Get(context.Background(), master)
		if e != nil {
			log.Printf("Unable to get jetton info for %v \n", master)
			return nil
		}
		return info.(*models.ChainTokenInfo)
	}

	usdRateCacheFunction := func(wallet string) *float64 {
		rate, e := usdRateCache.Get(context.Background(), wallet)
		if e != nil {
			return nil
		}
		return rate.(*float64)
	}
	swapChChannel := make(chan []*models.SwapCH)
	swapChArbitrageDetectorChannel := make(chan []*models.SwapCH)

	alreadySeenHashes := core.NewEvictableSet[string](3 * time.Minute)

	savedToChTransactionsHashes := core.NewEvictableSet[string](15 * time.Minute)
	go func() {
		for transactionHashes := range readyTransactionsChannel {
			var modelsCh []*models.SwapCH
			for _, transactionHash := range transactionHashes {
				if alreadySeenHashes.Exists(transactionHash) {
					continue
				}
				trace, e := consoleApi.GetTraceByHash(transactionHash)
				if e != nil {
					log.Printf("Unable to get trace %v \n", e)
					continue
				}
				for _, transaction := range stonfi.GetAllTransactionsFromTrace(trace) {
					alreadySeenHashes.Add(transaction.Hash)
				}

				stonfiV1Swaps := swapInfoWithDex(stonfi.ExtractStonfiSwapsFromRootTrace(trace), models.StonfiV1)
				stonfiV2Swaps := swapInfoWithDex(stonfiv2.ExtractStonfiV2SwapsFromRootTrace(trace), models.StonfiV2)

				dedustSwaps := dedust.ExtractDedustSwapsFromRootTrace(trace)

				modelsCh = append(modelsCh, common.Map(stonfiV1Swaps, func(pair core.Pair[*models.SwapInfo, string]) *models.SwapCH {
					return models.ToChSwap(pair.First, pair.Second, walletToMasterJettonCacheFunc, usdRateCacheFunction)
				})...)

				modelsCh = append(modelsCh, common.Map(stonfiV2Swaps, func(pair core.Pair[*models.SwapInfo, string]) *models.SwapCH {
					return models.ToChSwap(pair.First, pair.Second, walletToMasterJettonCacheFunc, usdRateCacheFunction)
				})...)

				for _, dedustSwap := range dedustSwaps {
					modelsCh = append(modelsCh, models.DedustSwapInfoToChSwap(dedustSwap, walletToMasterJettonCacheFunc, masterJettonCacheFunc, usdRateCacheFunction)...)
				}

			}
			notNullModels := common.Filter(modelsCh, func(ch *models.SwapCH) bool {
				return ch != nil
			})
			newModels := common.Filter(notNullModels, func(ch *models.SwapCH) bool {
				contains := false
				for _, hash := range ch.Hashes {
					if savedToChTransactionsHashes.Exists(hash) {
						contains = true
					}
				}
				return !contains
			})
			for _, swap := range newModels {
				for _, hash := range swap.Hashes {
					savedToChTransactionsHashes.Add(hash)
				}
			}

			go func() {
				swapChChannel <- newModels
			}()
			go func() {
				swapChArbitrageDetectorChannel <- newModels
			}()

			alreadySeenHashes.Evict()
			savedToChTransactionsHashes.Evict()
		}
	}()

	go func() {
		for {
			select {
			case chModels := <-swapChChannel:
				if chModels != nil {
					e := persistence.SaveSwapsToClickhouse(&dbConfig, chModels)
					if e != nil {
						log.Printf("Warning: Unable to save models %v\n", e)
					}
				}
			}
		}
	}()

	processedChModels := core.NewEvictableSet[*models.SwapCH](15 * time.Minute)
	for {
		chModels := <-swapChArbitrageDetectorChannel
		for _, model := range chModels {
			processedChModels.Add(model)
		}

		arbitrages := arbitrage.FindArbitragesAndDeleteThemFromSetGeneric(processedChModels)
		if len(arbitrages) > 0 {
			if e := persistence.WriteArbitragesToClickhouse(&dbConfig, arbitrages); e != nil {
				log.Printf("Warning: Unable to save arbitrages %v\n", e)
			}
		}
		processedChModels.Evict()
	}
}
