package core

import (
	"TonArb/models"
	"context"
	"github.com/tonkeeper/tonapi-go"
	"github.com/xssnick/tonutils-go/address"
	"log"
	"time"
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

func (client *TonClient) FetchTransactionsFromTraceByTransactionHash(hash string) ([]tonapi.Transaction, error) {
	params := tonapi.GetTraceParams{TraceID: hash}
	trace, e := client.GetTrace(context.Background(), params)
	if e != nil {
		return nil, e
	}

	transactions := getAllTransactions(*trace)

	return transactions, nil
}

func getAllTransactions(trace tonapi.Trace) []tonapi.Transaction {
	var transactions []tonapi.Transaction
	var traverse func(t tonapi.Trace)

	traverse = func(t tonapi.Trace) {
		transactions = append(transactions, t.Transaction)

		for _, child := range t.Children {
			traverse(child)
		}
	}

	traverse(trace)
	return transactions
}
