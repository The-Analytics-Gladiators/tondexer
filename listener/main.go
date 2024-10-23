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
	"github.com/eko/gocache/lib/v4/cache"
	gocache_store "github.com/eko/gocache/store/go_cache/v4"
	gocache "github.com/patrickmn/go-cache"
	"github.com/xssnick/tonutils-go/address"
)

func main() {

	gocacheClient := gocache.New(time.Hour, gocache.NoExpiration)

	gocacheStore := gocache_store.NewGoCache(gocacheClient)

	loadFunction := func(ctx context.Context, key any) (any, error) {
		return jettons.TokenInfoFromJettonWalletPage(key.(string))
	}

	// any because go-cache is supporting only any
	cacheManager := cache.NewLoadable[any](loadFunction, gocacheStore)

	//cacheManager.

	wallets, e := persistence.ReadStonfiRouterWallets()
	if e != nil {
		panic(e)
	}

	for _, wallet := range wallets {
		//Warming up
		cacheManager.Get(context.Background(), wallet)
	}

	log.Printf("Read %v Stonfi RouterWallets", len(wallets))

	consoleToken := os.Getenv("CONSOLE_TOKEN")
	streamingApi := tonapi.NewStreamingAPI(tonapi.WithStreamingToken(consoleToken))

	accounts := []string{stonfi.StonfiRouter}

	client, _ := tonapi.New(tonapi.WithToken(consoleToken))
	tonClient := &TonClient{client}

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
			msgCode := slice.MustLoadUInt(32)
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

	events := &core.Events[models.SwapTransferNotification, models.PaymentRequest, int]{
		ExpireCondition: func(tm *int) bool {
			return int(time.Now().Unix())-*tm > 30
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

	tokenInfoCacheFunction := func(token string) *models.TokenInfo {
		info, e := cacheManager.Get(context.Background(), token)
		if e != nil {
			return nil
		}
		return info.(*models.TokenInfo)
	}

	go func() {
		for {
			select {
			case notification := <-stonfiTransferNotificationsChannel:
				pair := &core.Pair[*models.SwapTransferNotification, *int]{First: notification, Second: core.IntRef(int(time.Now().Unix()))}
				events.Notifications = append(events.Notifications, pair)
				newEvents, relatedEvents := match(events)
				events = newEvents

				for _, relatedEvent := range relatedEvents {
					chModel := stonfi.ToChModel(relatedEvent, tokenInfoCacheFunction)
					swapChChannel <- chModel
				}

			case a := <-stonfiPaymentRequestChannel:
				pair := &core.Pair[*models.PaymentRequest, *int]{First: a, Second: core.IntRef(int(time.Now().Unix()))}
				events.Payments = append(events.Payments, pair)
				newEvents, relatedEvents := match(events)
				events = newEvents

				for _, relatedEvent := range relatedEvents {

					chModel := stonfi.ToChModel(relatedEvent, tokenInfoCacheFunction)
					swapChChannel <- chModel
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
			e := persistence.SaveToClickhouse(modelsBatch)
			if e == nil {
				modelsBatch = []*models.SwapCH{}
			}
		}
	}
}

func match(events *core.Events[models.SwapTransferNotification, models.PaymentRequest, int]) (
	*models.StonfiV1Events, []*models.StonfiV1RelatedEvents) {

	newEvents, relatedEvents := events.Match()

	return newEvents, relatedEvents
}

type TonClient struct {
	*tonapi.Client
}

func (client *TonClient) FetchRawTransactionFromHashToChannel(data *tonapi.TransactionEventData, chnl chan *models.RawTransactionWithHash) {
	if rawTransaction, e := client.FetchRawTransactionFromHash(data); e != nil {
		log.Printf("Smth wrong with getting raw transaction, %v \n", e)
	} else {
		chnl <- &models.RawTransactionWithHash{
			RawTransaction: rawTransaction,
			Hash:           data.TxHash,
			Lt:             data.Lt,
			Time:           time.Now(),
		}
	}
}

func (client *TonClient) FetchRawTransactionFromHash(data *tonapi.TransactionEventData) (*tonapi.GetRawTransactionsOK, error) {
	addr := address.MustParseRawAddr(data.AccountID.String())

	params := tonapi.GetRawTransactionsParams{
		Hash:      data.TxHash,
		AccountID: addr.String(),
		Lt:        int64(data.Lt),
		Count:     100,
	}

	return client.GetRawTransactions(context.Background(), params)
}
