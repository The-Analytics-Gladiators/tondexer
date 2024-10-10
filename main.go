package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/tonkeeper/tonapi-go"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
	"log"
	"os"
	"strconv"
)

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
	streamingApi.WebsocketHandleRequests(context.Background(), func(ws tonapi.Websocket) error {
		ws.SetTransactionHandler(func(data tonapi.TransactionEventData) {
			fmt.Printf("New tx with hash: %v, lt: %v \n", data.TxHash, data.Lt)
			client, _ := tonapi.New()
			addr := address.MustParseRawAddr(data.AccountID.String())
			fmt.Printf("Address: %v \n", addr.String())

			params := tonapi.GetRawTransactionsParams{
				Hash:      data.TxHash,
				AccountID: addr.String(),
				Lt:        int64(data.Lt),
				Count:     100,
			}
			rawTransaction, e := client.GetRawTransactions(context.Background(), params)

			if e != nil {
				fmt.Println(e)
			} else {

				hx, _ := hex.DecodeString(rawTransaction.Transactions)
				cl, _ := cell.FromBOC(hx)

				var tx tlb.Transaction
				err := tlb.LoadFromCell(&tx, cl.BeginParse())
				if err != nil {
					fmt.Println(err)
				}

				msgCode := tx.IO.In.AsInternal().Body.BeginParse().MustLoadUInt(32)
				msgCodeHex := strconv.FormatInt(int64(msgCode), 16)

				log.Printf("msg code %v \n", msgCodeHex)
			}
		})
		if err := ws.SubscribeToTransactions(accounts, nil); err != nil {
			return err
		}
		return nil
	})

}

//TIP See GoLand help at <a href="https://www.jetbrains.com/help/go/">jetbrains.com/help/go/</a>.
// Also, you can try interactive lessons for GoLand by selecting 'Help | Learn IDE Features' from the main menu.
