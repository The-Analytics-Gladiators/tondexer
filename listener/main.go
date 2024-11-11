package main

import (
	"context"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/tonkeeper/tonapi-go"
	"log"
	"os"
	"time"
	"tondexer/core"
	"tondexer/dedust"
	"tondexer/jettons"
	"tondexer/models"
	"tondexer/persistence"
	"tondexer/stonfi"
	"tondexer/stonfiv2"
)

//type IncomingHash struct {
//	Hash string
//}

func subscribeToAccounts(streamingApi *tonapi.StreamingAPI, accounts []string, incomingTransactionsChannel chan string) {
	for {
		e := streamingApi.WebsocketHandleRequests(context.Background(), func(ws tonapi.Websocket) error {
			ws.SetTransactionHandler(func(data tonapi.TransactionEventData) {
				log.Printf("New tx with hash: %v lt: %v \n", data.TxHash, data.Lt)
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
	}
}

func swapInfoWithDex(infos []*models.SwapInfo, dex string) []core.Pair[*models.SwapInfo, string] {
	return core.Map(infos, func(swapInfo *models.SwapInfo) core.Pair[*models.SwapInfo, string] {
		return core.Pair[*models.SwapInfo, string]{
			First:  swapInfo,
			Second: dex,
		}
	})
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

	client, _ := tonapi.New(tonapi.WithToken(cfg.ConsoleToken))
	consoleApi := &core.TonConsoleApi{Client: client}
	incomingTransactionsChannel := make(chan string)

	allSubscribers := append(stonfiV1Accounts, append(stonfiv2.Routers, dedust.VaultAddresses...)...)
	log.Printf("Subscribing to %v addresses... \n", len(allSubscribers))
	for _, chunk := range core.ChunkArray(allSubscribers, 10) {
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

	processedTransactions := core.NewEvictableSet[string](10 * time.Minute)
	go func() {
		for transactionHashes := range readyTransactionsChannel {
			var modelsCh []*models.SwapCH
			for _, transactionHash := range transactionHashes {
				if processedTransactions.Exists(transactionHash) {
					continue
				}
				trace, e := consoleApi.GetTraceByHash(transactionHash)
				if e != nil {
					log.Printf("Unable to get trace %v \n", e)
					continue
				}
				for _, transaction := range stonfi.GetAllTransactionsFromTrace(trace) {
					processedTransactions.Add(transaction.Hash)
				}

				stonfiV1Swaps := swapInfoWithDex(stonfi.ExtractStonfiSwapsFromRootTrace(trace), "StonfiV1")
				stonfiV2Swaps := swapInfoWithDex(stonfiv2.ExtractStonfiV2SwapsFromRootTrace(trace), "StonfiV2")
				dedustSwaps := swapInfoWithDex(dedust.ExtractDedustSwapsFromRootTrace(trace), "DeDust")

				swaps := append(stonfiV1Swaps, append(stonfiV2Swaps, dedustSwaps...)...)

				modelsCh = append(modelsCh, core.Map(swaps, func(pair core.Pair[*models.SwapInfo, string]) *models.SwapCH {
					return stonfi.ToChSwap(pair.First, pair.Second, jettonInfoCacheFunction, usdRateCacheFunction)
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
