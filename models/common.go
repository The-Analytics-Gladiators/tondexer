package models

import (
	"github.com/tonkeeper/tonapi-go"
	"time"
)

type RawTransactionWithHash struct {
	RawTransaction *tonapi.GetRawTransactionsOK
	Hash           string
	Lt             uint64
	Time           time.Time
}
