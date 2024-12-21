package models

import (
	"github.com/xssnick/tonutils-go/address"
	"math/big"
	"time"
)

type SwapPoolInfo struct {
	Hash     string
	Lt       uint64
	Address  *address.Address
	Sender   *address.Address
	JettonIn *address.Address
	AmountIn *big.Int
	Limit    *big.Int
}

type DedustSwapInfo struct {
	TraceID          string
	Lt               uint64    `ch:"lt"`
	Time             time.Time `ch:"time"`
	Sender           *address.Address
	InWalletAddress  *address.Address
	PoolsInfo        []*SwapPoolInfo
	OutWalletAddress *address.Address
	OutAmount        *big.Int
	CatchTime        time.Time
}
