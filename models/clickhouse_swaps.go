package models

import "time"

type SwapCH struct {
	Dex               string    `ch:"dex"`
	Hashes            []string  `ch:"hashes"`
	Lt                uint64    `ch:"lt"`
	Time              time.Time `ch:"time"`
	JettonIn          string    `ch:"jetton_in"`
	AmountIn          uint64    `ch:"amount_in"`
	JettonInSymbol    string    `ch:"jetton_in_symbol"`
	JettonInName      string    `ch:"jetton_in_name"`
	JettonInUsdRate   float64   `ch:"jetton_in_usd_rate"`
	JettonInDecimals  uint64    `ch:"jetton_in_decimals"`
	JettonOut         string    `ch:"jetton_out"`
	AmountOut         uint64    `ch:"amount_out"`
	JettonOutSymbol   string    `ch:"jetton_out_symbol"`
	JettonOutName     string    `ch:"jetton_out_name"`
	JettonOutUsdRate  float64   `ch:"jetton_out_usd_rate"`
	JettonOutDecimals uint64    `ch:"jetton_out_decimals"`
	MinAmountOut      uint64    `ch:"min_amount_out"`
	PoolAddress       string    `ch:"pool_address"`
	Sender            string    `ch:"sender"`
	ReferralAddress   string    `ch:"referral_address"`
	ReferralAmount    uint64    `ch:"referral_amount"`
	CatchTime         time.Time `ch:"catch_time"`
}
