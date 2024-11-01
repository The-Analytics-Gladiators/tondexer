package persistence

import (
	"fmt"
	"time"
	"tondexer/core"
)

const UsdInField = "(amount_in / pow(10, jetton_in_decimals)) * jetton_in_usd_rate"
const UsdOutField = "(amount_out / pow(10, jetton_out_decimals)) * jetton_out_usd_rate"

type EnrichedSwapCH struct {
	Time              time.Time `ch:"time"`
	Dex               string    `ch:"dex"`
	Hashes            []string  `ch:"hashes"`
	Sender            string    `ch:"sender"`
	JettonInMaster    string    `ch:"jetton_in"`
	JettonInSymbol    string    `ch:"jetton_in_symbol"`
	JettonInName      string    `ch:"jetton_in_name"`
	JettonInUsdRate   float64   `ch:"jetton_in_usd_rate"`
	JettonInDecimals  uint64    `ch:"jetton_in_decimals"`
	AmountIn          uint64    `ch:"amount_in"`
	InUsd             float64   `ch:"in_usd"`
	JettonOut         string    `ch:"jetton_out"`
	JettonOutSymbol   string    `ch:"jetton_out_symbol"`
	JettonOutName     string    `ch:"jetton_out_name"`
	JettonOutUsdRate  float64   `ch:"jetton_out_usd_rate"`
	JettonOutDecimals uint64    `ch:"jetton_out_decimals"`
	AmountOut         uint64    `ch:"amount_out"`
	OutUsd            float64   `ch:"out_usd"`
	MinAmountOut      uint64    `ch:"min_amount_out"`
	ReferralAddress   string    `ch:"referral_address"`
	ReferralAmount    uint64    `ch:"referral_amount"`
	ReferralUsd       float64   `ch:"referral_usd"`
}

func LastSwapsSqlQuery(config *core.Config, limit uint64) string {
	return fmt.Sprint(`
SELECT
	time,
	dex,
	hashes,
	sender,
	jetton_in,
	jetton_in_symbol,
	jetton_in_name,
	jetton_in_usd_rate,
	jetton_in_decimals,
	amount_in,
	floor(`, UsdInField, `, 3) AS in_usd,
	jetton_out,
	jetton_out_symbol,
	jetton_out_name,
	jetton_out_usd_rate,
	jetton_out_decimals,
	amount_out,
	floor(`, UsdOutField, `, 3) AS out_usd,
	min_amount_out,
	referral_address,
	referral_amount,
	floor((referral_amount / pow(10, jetton_out_decimals)) * jetton_out_usd_rate, 4) AS referral_usd
FROM `, config.DbName, `.swaps
ORDER BY time DESC
LIMIT `, limit)
}
