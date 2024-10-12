package main

import (
	"context"
	"github.com/tonkeeper/tonapi-go"
	"github.com/xssnick/tonutils-go/address"
	"log"
	"os"
)

const TransferNotificationCode = 1935855772

//TIP To run your code, right-click the code and select <b>Run</b>. Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.

func main() {
	consoleToken := os.Getenv("CONSOLE_TOKEN")

	streamingApi := tonapi.NewStreamingAPI(tonapi.WithStreamingToken(consoleToken))

	accounts := []string{"EQB3ncyBUTjZUA5EnFKR5_EnOMI9V1tTEAAPaiU71gc4TiUt"}

	//api, _ := tonapi.New(tonapi.WithToken(consoleToken))

	//api.GetBlockchainTransaction(context.Background(), tonapi.GetBlockchainTransactionParams { TransactionID: })

	//t, _ := api.GetBlockchainAccountTransactions(context.Background(), tonapi.GetBlockchainAccountTransactionsParams{AccountID: "EQB3ncyBUTjZUA5EnFKR5_EnOMI9V1tTEAAPaiU71gc4TiUt"})
	//
	//transactions := t.Transactions
	//for i := range transactions {
	//	transaction := transactions[i]
	//	inMsg := transaction.InMsg.Value
	//
	//	fmt.Printf("%v: %v \n", inMsg.DecodedOpName, transaction.Hash)
	//	//cl, _ := cell.FromBOC([]byte(inMsg.RawBody.Value))
	//	hx, _ := hex.DecodeString(inMsg.RawBody.Value)
	//	cl, _ := cell.FromBOC(hx)
	//
	//	msgCode := cl.BeginParse().MustLoadUInt(32)
	//
	//	//cl, _ := cell.FromBOC(inMsg.DecodedBody)
	//	fmt.Printf("%v \n", msgCode)
	//}
	//client, _ := tonapi.New(tonapi.WithToken(consoleToken))
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

	for rawTransaction := range rawTransactionChannel {
		transaction, _ := ParseRawTransaction(rawTransaction.Transactions)

		slice := transaction.IO.In.AsInternal().Body.BeginParse()
		msgCode := slice.MustLoadUInt(32)
		if msgCode == TransferNotificationCode {
			transferNotification := ParseSwapTransferNotificationMessage(transaction.IO.In.AsInternal())
			log.Printf("Message: %v \n ", transferNotification.String())
		}
	}

}

type TonClient struct {
	*tonapi.Client
}

func (client *TonClient) FetchRawTransactionFromHashToChannel(data *tonapi.TransactionEventData, chnl chan *tonapi.GetRawTransactionsOK) {
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
