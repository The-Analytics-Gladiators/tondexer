package core

import (
	"context"
	"github.com/tonkeeper/tonapi-go"
	"github.com/xssnick/tonutils-go/address"
	"log"
	"time"
	"tondexer/models"
)

type TonClient struct {
	*tonapi.Client
}

func (client *TonClient) FetchRawTransactionFromHashToChannel(data *tonapi.TransactionEventData, chnl chan *models.RawTransactionWithHash) {
	if rawTransaction, e := client.FetchRawTransactionFromHash(data); e != nil {
		log.Printf("Smth wrong with getting raw transaction, %v \n", e)
	} else {
		chnl <- &models.RawTransactionWithHash{
			RawTransaction:  rawTransaction,
			Hash:            data.TxHash,
			Lt:              data.Lt,
			TransactionTime: time.Now(),
			CatchEventTime:  time.Now(),
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
