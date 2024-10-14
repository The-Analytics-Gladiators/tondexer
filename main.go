package main

import (
	"context"
	"fmt"
	"github.com/tonkeeper/tonapi-go"
	"github.com/xssnick/tonutils-go/address"
	"log"
	"os"
	"time"
)

const TransferNotificationCode = 1935855772
const PaymentRequestCode = 4181439551

const StonfiRouter = "EQB3ncyBUTjZUA5EnFKR5_EnOMI9V1tTEAAPaiU71gc4TiUt"
const StonfiRouterV2 = "EQBCl1JANkTpMpJ9N3lZktPMpp2btRe2vVwHon0la8ibRied"

//TIP To run your code, right-click the code and select <b>Run</b>. Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.

func main() {
	consoleToken := os.Getenv("CONSOLE_TOKEN")

	streamingApi := tonapi.NewStreamingAPI(tonapi.WithStreamingToken(consoleToken))

	accounts := []string{StonfiRouter}

	client, _ := tonapi.New()
	tonClient := &TonClient{client}

	rawTransactionChannel := make(chan *tonapi.GetRawTransactionsOK)
	go streamingApi.WebsocketHandleRequests(context.Background(), func(ws tonapi.Websocket) error {
		ws.SetTransactionHandler(func(data tonapi.TransactionEventData) {
			log.Printf("New tx with hash: %v lt: %v \n", data.TxHash, data.Lt)
			go tonClient.FetchRawTransactionFromHashToChannel(&data, rawTransactionChannel)
		})
		if err := ws.SubscribeToTransactions(accounts, nil); err != nil {
			return err
		}
		return nil
	})

	stonfiTransferNotificationsChannel := make(chan *SwapTransferNotification)
	stonfiPaymentRequestChannel := make(chan *PaymentRequest)

	go func() {
		for rawTransaction := range rawTransactionChannel {
			transaction, _ := ParseRawTransaction(rawTransaction.Transactions)

			slice := transaction.IO.In.AsInternal().Body.BeginParse()
			msgCode := slice.MustLoadUInt(32)
			if msgCode == TransferNotificationCode {
				transferNotification := ParseSwapTransferNotificationMessage(transaction.IO.In.AsInternal())
				//log.Printf("Message: %v \n ", transferNotification.String())
				stonfiTransferNotificationsChannel <- transferNotification
			} else if msgCode == PaymentRequestCode {
				paymentRequest := ParsePaymentRequestMessage(transaction.IO.In.AsInternal())
				//log.Printf("Message: %v \n ", paymentRequest.String())
				stonfiPaymentRequestChannel <- paymentRequest
			}
		}
	}()

	var transferNotificationsList []*SwapTransferNotification
	var paymentRequestList []*PaymentRequest

	lastJoinTime := time.Now().Second()
	for {
		select {
		case transferNotification := <-stonfiTransferNotificationsChannel:
			transferNotificationsList = append(transferNotificationsList, transferNotification)
			transferNotificationsList, paymentRequestList = matchStonfi(transferNotificationsList, paymentRequestList, lastJoinTime)
			lastJoinTime = time.Now().Second()
		case paymentRequest := <-stonfiPaymentRequestChannel:
			paymentRequestList = append(paymentRequestList, paymentRequest)
			transferNotificationsList, paymentRequestList = matchStonfi(transferNotificationsList, paymentRequestList, lastJoinTime)
			lastJoinTime = time.Now().Second()
		}
	}

}

func matchStonfi(transferNotificationsList []*SwapTransferNotification,
	paymentRequestList []*PaymentRequest, lastJoinTime int) (
	[]*SwapTransferNotification, []*PaymentRequest) {

	var toDeleteQueryIds []uint64
	if time.Now().Second()-lastJoinTime > 5 {
		fmt.Printf("================= CALCULATING ======================\n")
		for transferNotificationIndex := range transferNotificationsList {
			transferNotification := transferNotificationsList[transferNotificationIndex]
			if transferNotification.QueryId == 0 {
				continue
			}

			var matchedTransferNotificationsList []*SwapTransferNotification
			var matchedPaymentRequestList []*PaymentRequest

			for paymentRequestIndex := range paymentRequestList {
				paymentRequest := paymentRequestList[paymentRequestIndex]

				if transferNotification.QueryId == paymentRequest.QueryId {
					if len(toDeleteQueryIds) == 0 || toDeleteQueryIds[len(toDeleteQueryIds)-1] != paymentRequest.QueryId {
						toDeleteQueryIds = append(toDeleteQueryIds, paymentRequest.QueryId)
						matchedTransferNotificationsList = append(matchedTransferNotificationsList, transferNotification)
					}
					matchedPaymentRequestList = append(matchedPaymentRequestList, paymentRequest)
				}
			}
			if len(matchedTransferNotificationsList) != 0 {
				for _, transferNotification := range matchedTransferNotificationsList {
					log.Printf("Message: %v  ", transferNotification.String())
				}
				for _, paymentRequest := range matchedPaymentRequestList {
					log.Printf("Message: %v  ", paymentRequest.String())
				}
				log.Printf("======================================================\n")
			}
		}

		var filteredTransferNotifications []*SwapTransferNotification
		var filteredPaymentRequests []*PaymentRequest

		//for _, queryId := range toDeleteQueryIds {
		for _, transferNotification := range transferNotificationsList {
			if !contains(toDeleteQueryIds, transferNotification.QueryId) {
				filteredTransferNotifications = append(filteredTransferNotifications, transferNotification)
			}
		}

		for _, paymentRequest := range paymentRequestList {
			if !contains(toDeleteQueryIds, paymentRequest.QueryId) {
				filteredPaymentRequests = append(filteredPaymentRequests, paymentRequest)
			}
		}
		return filteredTransferNotifications, filteredPaymentRequests
		//}
	} else {
		return transferNotificationsList, paymentRequestList
	}
}

func contains(slice []uint64, value uint64) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

type TonClient struct {
	*tonapi.Client
}

func (client *TonClient) FetchRawTransactionFromHashToChannel(data *tonapi.TransactionEventData, chnl chan *tonapi.GetRawTransactionsOK) {
	//log.Printf("Transaction data : %v \n", data)
	if rawTransaction, e := client.FetchRawTransactionFromHash(data); e != nil {
		log.Printf("Smth wrong with getting raw transaction, %v \n", e)
	} else {
		chnl <- rawTransaction
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
