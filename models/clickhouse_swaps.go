package models

import "time"

type SwapCH struct {
	Dex             string
	Hashes          []string
	Lt              uint64
	Time            time.Time
	TokenIn         string
	AmountIn        uint64
	TokenInSymbol   string
	TokenInName     string
	TokenInUsdRate  float64
	TokenOut        string
	AmountOut       uint64
	TokenOutSymbol  string
	TokenOutName    string
	TokenOutUsdRate float64
	MinAmountOut    uint64
	Sender          string
	ReferralAddress string
	ReferralAmount  uint64
}
