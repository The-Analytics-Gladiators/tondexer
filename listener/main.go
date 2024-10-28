package main

import (
	"TonArb/core"
	"TonArb/jettons"
	"TonArb/persistence"
	"TonArb/stonfi"
	"context"
	"github.com/tonkeeper/tonapi-go"
	"github.com/xssnick/tonutils-go/address"
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
	tonClient := &core.TonClient{Client: client}

	rawTransactionWithHashChannel := make(chan *models.RawTransactionWithHash)

	stonfiTransferNotificationsChannel := make(chan *models.SwapTransferNotification)
	stonfiPaymentRequestChannel := make(chan *models.PaymentRequest)

	go func() {
		streamingApi.WebsocketHandleRequests(context.Background(), func(ws tonapi.Websocket) error {
			ws.SetTransactionHandler(func(data tonapi.TransactionEventData) {
				log.Printf("New tx with hash: %v lt: %v \n", data.TxHash, data.Lt)
				go tonClient.FetchRawTransactionFromHashToChannel(&data, rawTransactionWithHashChannel)
			})
			if err := ws.SubscribeToTransactions(accounts, nil); err != nil {
				return err
			}
			return nil
		})
	}()

	go func() {
		for rawTransactionWithHash := range rawTransactionWithHashChannel {
			transaction, _ := stonfi.ParseRawTransaction(rawTransactionWithHash.RawTransaction.Transactions)
			slice := transaction.IO.In.AsInternal().Body.BeginParse()
			msgCode, e := slice.LoadUInt(32)
			if e != nil {
				continue
			}
			if msgCode == stonfi.TransferNotificationCode {
				transferNotification := stonfi.ParseSwapTransferNotificationMessage(transaction.IO.In.AsInternal(), rawTransactionWithHash)
				if transferNotification != nil {
					stonfiTransferNotificationsChannel <- transferNotification
				}
			} else if msgCode == stonfi.PaymentRequestCode {
				paymentRequest := stonfi.ParsePaymentRequestMessage(transaction.IO.In.AsInternal(), rawTransactionWithHash)
				if paymentRequest != nil {
					stonfiPaymentRequestChannel <- paymentRequest
				}
			}
		}
	}()

	events := &core.Events[models.SwapTransferNotification, models.PaymentRequest, int64]{
		ExpireCondition: func(tm *int64) bool {
			return time.Now().Unix()-*tm > 30
		},
		NotificationWithPaymentMatchCondition: func(n *models.SwapTransferNotification, p *models.PaymentRequest) bool {
			if n.QueryId != p.QueryId {
				return false
			}
			return n.ToAddress.Equals(p.Owner) ||
				(n.ReferralAddress != nil && p.Owner.Equals(n.ReferralAddress))
		},
	}

	swapChChannel := make(chan *models.SwapCH)

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

	nonMatchedHashesChannel := make(chan []string)

	matchTicker := time.NewTicker(5 * time.Second)

	go func() {
		for {
			select {
			case <-matchTicker.C:
				newEvents, relatedEvents, orphanEvents := match(events)
				events = newEvents

				var nonMatchedHashes []string
				for _, relatedEvent := range relatedEvents {
					chModel := stonfi.ToChModel(relatedEvent, jettonInfoCacheFunction, usdRateCacheFunction)
					if chModel == nil {
						if relatedEvent.Notification != nil {
							nonMatchedHashes = append(nonMatchedHashes, relatedEvent.Notification.Hash)
						}
						hashes := core.Map(relatedEvent.Payments, func(t *models.PaymentRequest) string { return t.Hash })
						nonMatchedHashes = append(nonMatchedHashes, hashes...)
					}
					swapChChannel <- chModel
				}

				nonMatchedHashes = append(nonMatchedHashes, core.Map(orphanEvents.Notifications, func(t *models.SwapTransferNotification) string { return t.Hash })...)
				nonMatchedHashes = append(nonMatchedHashes, core.Map(orphanEvents.Payments, func(t *models.PaymentRequest) string { return t.Hash })...)

				if len(nonMatchedHashes) > 0 {
					go func() { nonMatchedHashesChannel <- nonMatchedHashes }()
				}
			case notification := <-stonfiTransferNotificationsChannel:
				pair := &core.Pair[*models.SwapTransferNotification, *int64]{First: notification, Second: core.Int64Ref(notification.EventCatchTime.Unix())}
				events.Notifications = append(events.Notifications, pair)
			case a := <-stonfiPaymentRequestChannel:
				pair := &core.Pair[*models.PaymentRequest, *int64]{First: a, Second: core.Int64Ref(a.EventCatchTime.Unix())}
				events.Payments = append(events.Payments, pair)
			}
		}
	}()

	go func() {
		for hashes := range nonMatchedHashesChannel {
			mp := make(map[string]tonapi.Transaction)

			log.Printf("hashes %v \n", hashes)
			for _, hash := range hashes {
				if _, exists := mp[hash]; !exists {
					if transactions, e := tonClient.FetchTransactionsFromTraceByTransactionHash(hash); e == nil {
						for _, transaction := range transactions {
							mp[transaction.Hash] = transaction
						}
					} else {
						log.Printf("Unable to fetch trace for %v \n", hash)
					}
				}
			}
			for _, transaction := range mp {
				addr := address.MustParseRawAddr(transaction.Account.Address)
				if addr.String() == stonfi.StonfiRouter {
					rt := &models.RawTransactionWithHash{
						RawTransaction: &tonapi.GetRawTransactionsOK{
							Transactions: transaction.Raw,
						},
						Hash:            transaction.Hash,
						Lt:              uint64(transaction.Lt),
						TransactionTime: time.UnixMilli(transaction.Utime * 1000),
						CatchEventTime:  time.Now(),
					}
					go func() { rawTransactionWithHashChannel <- rt }()
				}
			}
		}
	}()

	var modelsBatch []*models.SwapCH
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case model := <-swapChChannel:
			if model != nil {
				modelsBatch = append(modelsBatch, model)
			}
		case <-ticker.C:
			e := persistence.SaveSwapsToClickhouse(modelsBatch)
			if e == nil {
				modelsBatch = []*models.SwapCH{}
			}
		}
	}
}

func match(events *core.Events[models.SwapTransferNotification, models.PaymentRequest, int64]) (
	*stonfi.StonfiV1Events, []*stonfi.StonfiV1RelatedEvents, core.OrphanEvents[models.SwapTransferNotification, models.PaymentRequest]) {

	newEvents, relatedEvents, orphanEvents := events.Match()

	return newEvents, relatedEvents, orphanEvents
}
