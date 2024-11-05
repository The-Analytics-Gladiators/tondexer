package main

import (
	"context"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/tonkeeper/tonapi-go"
	"log"
	"os"
	"time"
	"tondexer/core"
	"tondexer/jettons"
	"tondexer/persistence"
	"tondexer/stonfi"
	"tondexer/stonfiv2"

	"tondexer/models"
)

const StonfiRouterV2 = "EQBCl1JANkTpMpJ9N3lZktPMpp2btRe2vVwHon0la8ibRied"

type IncomingHash struct {
	Hash string
	Dex  string
}

func main() {
	var cfg core.Config

	if err := cleanenv.ReadConfig(os.Args[1], &cfg); err != nil {
		panic(err)
	}
	jettonInfoCache, e := jettons.InitJettonInfoCache(&cfg)
	if e != nil {
		panic(e)
	}

	walletMasterCache, e := jettons.InitWalletJettonCache(&cfg)
	if e != nil {
		panic(e)
	}

	usdRateCache, err := jettons.InitUsdRateCache(&cfg)
	if err != nil {
		panic(e)
	}

	streamingApi := tonapi.NewStreamingAPI(tonapi.WithStreamingToken(cfg.ConsoleToken))

	stonfiV1Accounts := []string{stonfi.StonfiRouter}
	stonfiV2Accounts := []string{StonfiRouterV2}

	client, _ := tonapi.New(tonapi.WithToken(cfg.ConsoleToken))
	incomingTransactionsChannel := make(chan *IncomingHash)

	go func() {
		for {
			e := streamingApi.WebsocketHandleRequests(context.Background(), func(ws tonapi.Websocket) error {
				ws.SetTransactionHandler(func(data tonapi.TransactionEventData) {
					log.Printf("New tx with hash: %v lt: %v \n", data.TxHash, data.Lt)
					go func() {
						incomingHash := &IncomingHash{
							Hash: data.TxHash,
							Dex:  "StonfiV2",
						}
						incomingTransactionsChannel <- incomingHash
					}()
				})
				if err := ws.SubscribeToTransactions(stonfiV2Accounts, nil); err != nil {
					return err
				}
				return nil
			})
			if e != nil {
				log.Printf("Streaming failed! %v \n", e)
			}
		}
	}()

	go func() {
		for {
			e := streamingApi.WebsocketHandleRequests(context.Background(), func(ws tonapi.Websocket) error {
				ws.SetTransactionHandler(func(data tonapi.TransactionEventData) {
					log.Printf("New tx with hash: %v lt: %v \n", data.TxHash, data.Lt)
					go func() {
						incomingHash := &IncomingHash{
							Hash: data.TxHash,
							Dex:  "StonfiV1",
						}
						incomingTransactionsChannel <- incomingHash
					}()
				})
				if err := ws.SubscribeToTransactions(stonfiV1Accounts, nil); err != nil {
					return err
				}
				return nil
			})
			if e != nil {
				log.Printf("Streaming failed! %v \n", e)
			}
		}
	}()

	readyTransactionsChannel := make(chan []*IncomingHash)

	transactionsWaitingList := &core.WaitingList[*IncomingHash]{
		ExpirationSeconds: 40 * time.Second,
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
			}
		}
	}()

	jettonInfoCacheFunction := func(wallet string) *jettons.ChainTokenInfo {
		master, e := walletMasterCache.Get(context.Background(), wallet)
		if e != nil {
			return nil
		}

		info, e := jettonInfoCache.Get(context.Background(), master.(*models.WalletJetton).Master)
		if e != nil {
			log.Printf("Unable to get jetton info for %v \n", master.(*models.WalletJetton).Master)
			return nil
		}
		return info.(*jettons.ChainTokenInfo)
	}

	usdRateCacheFunction := func(wallet string) *float64 {
		rate, e := usdRateCache.Get(context.Background(), wallet)
		if e != nil {
			return nil
		}
		return rate.(*float64)
	}
	swapChChannel := make(chan []*models.SwapCH)

	processedTransactions := core.NewEvictableSet[string](3 * time.Minute)
	go func() {
		for transactionHashes := range readyTransactionsChannel {
			var modelsCh []*models.SwapCH
			for _, transactionHash := range transactionHashes {
				if processedTransactions.Exists(transactionHash.Hash) {
					continue
				}
				params := tonapi.GetTraceParams{TraceID: transactionHash.Hash}
				trace, e := client.GetTrace(context.Background(), params)
				if e != nil {
					log.Printf("Unable to get trace %v \n", e)
					continue
				}
				for _, transaction := range stonfi.GetAllTransactionsFromTrace(trace) {
					processedTransactions.Add(transaction.Hash)
				}

				var swaps []*models.SwapInfo
				if transactionHash.Dex == "StonfiV1" {
					swaps = stonfi.ExtractStonfiSwapsFromRootTrace(trace)
				} else if transactionHash.Dex == "StonfiV2" {
					swaps = stonfiv2.ExtractStonfiV2SwapsFromRootTrace(trace)
				}

				modelsCh = append(modelsCh, core.Map(swaps, func(swap *models.SwapInfo) *models.SwapCH {
					return stonfi.ToChSwap(swap, transactionHash.Dex, jettonInfoCacheFunction, usdRateCacheFunction)
				})...)

			}
			go func() {
				swapChChannel <- core.Filter(modelsCh, func(ch *models.SwapCH) bool {
					return ch != nil
				})
			}()
			processedTransactions.Evict()
		}
	}()

	for {
		select {
		case chModels := <-swapChChannel:
			if chModels != nil {
				e := persistence.SaveSwapsToClickhouse(&cfg, chModels)
				if e != nil {
					log.Printf("Warning: Unable to save models %v\n", e)
				}
			}
		}
	}
}
