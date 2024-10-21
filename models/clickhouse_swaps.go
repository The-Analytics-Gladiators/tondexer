package models

import "time"

type SwapCH struct {
	Hashes          []string
	Lt              uint64
	Time            time.Time
	TokenIn         string
	AmountIn        uint64
	TokenOut        string
	AmountOut       uint64
	MinAmountOut    uint64
	Sender          string
	ReferralAddress string
	ReferralAmount  uint64
}
