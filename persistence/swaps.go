package persistence

import (
	"fmt"
	"math/big"
	"time"
	"tondexer/core"
	"tondexer/models"
)

const UsdInField = "(amount_in / pow(10, jetton_in_decimals)) * jetton_in_usd_rate"
const UsdOutField = "(amount_out / pow(10, jetton_out_decimals)) * jetton_out_usd_rate"
const UsdReferralField = "(referral_amount / pow(10, jetton_out_decimals)) * jetton_out_usd_rate"

func UsdField(amountType string) string {
	return fmt.Sprint("(amount_", amountType, " / pow(10, jetton_decimals) * jetton_usd_rate)")
}

var enrichedSwapSelect = fmt.Sprint(`
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
	amount_in / pow(10, jetton_in_decimals) AS amount_jetton_in,
	floor(`, UsdInField, `, 2) AS in_usd,
	jetton_out,
	jetton_out_symbol,
	jetton_out_name,
	jetton_out_usd_rate,
	jetton_out_decimals,
	amount_out,
	amount_out / pow(10, jetton_out_decimals) AS amount_jetton_out,
	floor(`, UsdOutField, `, 2) AS out_usd,
	min_amount_out,
	referral_address,
	referral_amount,
	floor(`, UsdReferralField, `, 2) AS referral_usd,
	trace_id
`)

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
	AmountIn          *big.Int  `ch:"amount_in"`
	AmountJettonIn    float64   `ch:"amount_jetton_in"`
	InUsd             float64   `ch:"in_usd"`
	JettonOut         string    `ch:"jetton_out"`
	JettonOutSymbol   string    `ch:"jetton_out_symbol"`
	JettonOutName     string    `ch:"jetton_out_name"`
	JettonOutUsdRate  float64   `ch:"jetton_out_usd_rate"`
	JettonOutDecimals uint64    `ch:"jetton_out_decimals"`
	AmountOut         *big.Int  `ch:"amount_out"`
	AmountJettonOut   float64   `ch:"amount_jetton_out"`
	OutUsd            float64   `ch:"out_usd"`
	MinAmountOut      *big.Int  `ch:"min_amount_out"`
	ReferralAddress   string    `ch:"referral_address"`
	ReferralAmount    *big.Int  `ch:"referral_amount"`
	ReferralUsd       float64   `ch:"referral_usd"`
	TraceID           string    `ch:"trace_id"`
}

func LatestSwapsSqlQuery(config *core.Config, limit uint64) string {
	return fmt.Sprint(
		enrichedSwapSelect, `
FROM `, config.DbName, `.swaps
ORDER BY time DESC
LIMIT `, limit)
}

func TopSwapsSqlQuery(config *core.Config, period models.Period) string {
	periodParams := models.PeriodParamsMap[period]
	return fmt.Sprint(enrichedSwapSelect,
		`
FROM `, config.DbName, `.swaps
WHERE time >= `, periodParams.ToStartOf, `(subtractDays(now(), `, periodParams.WindowInDays, `))
ORDER BY (in_usd + out_usd) DESC
LIMIT 15
`)
}
