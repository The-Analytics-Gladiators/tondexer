package models

import "time"

type SwapCH struct {
	Dex               string
	Hashes            []string
	Lt                uint64
	Time              time.Time
	JettonIn          string
	AmountIn          uint64
	JettonInSymbol    string
	JettonInName      string
	JettonInUsdRate   float64
	JettonInDecimals  uint64
	JettonOut         string
	AmountOut         uint64
	JettonOutSymbol   string
	JettonOutName     string
	JettonOutUsdRate  float64
	JettonOutDecimals uint64
	MinAmountOut      uint64
	Sender            string
	ReferralAddress   string
	ReferralAmount    uint64
}
