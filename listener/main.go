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

type IncomingHash struct {
	Hash string
	Dex  string
}

func subscribeToAccounts(streamingApi *tonapi.StreamingAPI, dex string, accounts []string, incomingTransactionsChannel chan *IncomingHash) {
	for {
		e := streamingApi.WebsocketHandleRequests(context.Background(), func(ws tonapi.Websocket) error {
			ws.SetTransactionHandler(func(data tonapi.TransactionEventData) {
				log.Printf("%v: new tx with hash: %v lt: %v \n", dex, data.TxHash, data.Lt)
				go func() {
					incomingHash := &IncomingHash{
						Hash: data.TxHash,
						Dex:  dex,
					}
					incomingTransactionsChannel <- incomingHash
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
	}
}

func main() {
	var cfg core.Config

	freeConsoleClient, _ := tonapi.New() // free one for the rates
	freeConsoleApi := core.TonConsoleApi{Client: freeConsoleClient}

	if err := cleanenv.ReadConfig(os.Args[1], &cfg); err != nil {
		panic(err)
	}
	jettonInfoCache, e := jettons.InitJettonInfoCache(&cfg, &freeConsoleApi)
	if e != nil {
		panic(e)
	}

	walletMasterCache, e := jettons.InitWalletJettonCache(&cfg)
	if e != nil {
		panic(e)
	}

	usdRateCache, err := jettons.InitUsdRateCache(&cfg, &freeConsoleApi)
	if err != nil {
		panic(e)
	}

	streamingApi := tonapi.NewStreamingAPI(tonapi.WithStreamingToken(cfg.ConsoleToken))

	stonfiV1Accounts := []string{stonfi.StonfiRouter}

	v2RoutersChunks := core.ChunkArray(stonfiv2.Routers, 10)

	client, _ := tonapi.New(tonapi.WithToken(cfg.ConsoleToken))
	consoleApi := &core.TonConsoleApi{Client: client}
	incomingTransactionsChannel := make(chan *IncomingHash)

	go subscribeToAccounts(streamingApi, "StonfiV1", stonfiV1Accounts, incomingTransactionsChannel)
	for _, chunk := range v2RoutersChunks {
		go subscribeToAccounts(streamingApi, "StonfiV2", chunk, incomingTransactionsChannel)
	}

	readyTransactionsChannel := make(chan []*IncomingHash)

	transactionsWaitingList := &core.WaitingList[*IncomingHash]{
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
			}
		}
	}()

	jettonInfoCacheFunction := func(wallet string) *models.ChainTokenInfo {
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
				trace, e := consoleApi.GetTraceByHash(transactionHash.Hash)
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
