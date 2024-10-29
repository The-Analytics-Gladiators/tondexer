package main

import (
	"TonArb/core"
	"TonArb/jettons"
	"TonArb/persistence"
	"TonArb/stonfi"
	"context"
	"github.com/tonkeeper/tonapi-go"
	"log"
	"os"
	"time"

	"TonArb/models"
)

func main() {

	jettonInfoCache, e := jettons.InitJettonInfoCache()
	if e != nil {
		panic(e)
	}

	walletMasterCache, e := jettons.InitWalletJettonCache()
	if e != nil {
		panic(e)
	}

	usdRateCache, err := jettons.InitUsdRateCache()
	if err != nil {
		panic(e)
	}

	consoleToken := os.Getenv("CONSOLE_TOKEN")
	streamingApi := tonapi.NewStreamingAPI(tonapi.WithStreamingToken(consoleToken))

	accounts := []string{stonfi.StonfiRouter}

	client, _ := tonapi.New(tonapi.WithToken(consoleToken))
	incomingTransactionsChannel := make(chan string)

	go func() {
		for {
			e := streamingApi.WebsocketHandleRequests(context.Background(), func(ws tonapi.Websocket) error {
				ws.SetTransactionHandler(func(data tonapi.TransactionEventData) {
					log.Printf("New tx with hash: %v lt: %v \n", data.TxHash, data.Lt)
					//go tonClient.FetchRawTransactionFromHashToChannel(&data, rawTransactionWithHashChannel)
					go func() { incomingTransactionsChannel <- data.TxHash }()
				})
				if err := ws.SubscribeToTransactions(accounts, nil); err != nil {
					return err
				}
				return nil
			})
			if e != nil {
				log.Printf("Streaming failed! %v \n", e)
			}
		}
	}()

	readyTransactionsChannel := make(chan []string)

	transactionsWaitingList := &core.WaitingList[string]{
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
				if processedTransactions.Exists(transactionHash) {
					continue
				}
				params := tonapi.GetTraceParams{TraceID: transactionHash}
				trace, e := client.GetTrace(context.Background(), params)
				if e != nil {
					continue
				}
				for _, transaction := range stonfi.GetAllTransactionsFromTrace(trace) {
					processedTransactions.Add(transaction.Hash)
				}

				swaps := stonfi.ExtractStonfiSwapsFromRootTrace(trace)
				modelsCh = append(modelsCh, core.Map(swaps, func(swap *stonfi.StonfiV1Swap) *models.SwapCH {
					return stonfi.ToChSwap(swap, jettonInfoCacheFunction, usdRateCacheFunction)
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
				e := persistence.SaveSwapsToClickhouse(chModels)
				if e != nil {
					log.Printf("Warning: Unable to save models %v\n", e)
				}
			}
		}
	}
}
