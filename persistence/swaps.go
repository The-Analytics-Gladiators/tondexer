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
	trace_id,
	pool_address
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
	PoolAddress       string    `ch:"pool_address"`
}

func LatestSwapsSqlQuery(config *core.Config, limit uint64, dex models.Dex) string {
	return fmt.Sprint(
		enrichedSwapSelect, `
FROM `, config.DbName, `.swaps
WHERE `, dex.WhereStatement("dex"), `
ORDER BY time DESC
LIMIT `, limit)
}

func TopSwapsSqlQuery(config *core.Config, period models.Period, dex models.Dex) string {
	periodParams := models.PeriodParamsMap[period]
	return fmt.Sprint(enrichedSwapSelect, `
FROM `, config.DbName, `.swaps
WHERE time >= `, periodParams.ToStartOf, `(subtractDays(now(), `, periodParams.WindowInDays, `))
AND `, dex.WhereStatement("dex"), `
ORDER BY (in_usd + out_usd) DESC
LIMIT 15
`)
}

type SwapDistribution struct {
	Usd_1        uint64 `ch:"usd_1" json:"usd_1"`
	Usd_1_5      uint64 `ch:"usd_1_5" json:"usd_1_5"`
	Usd_5_15     uint64 `ch:"usd_5_15" json:"usd_5_15"`
	Usd_15_50    uint64 `ch:"usd_15_50" json:"usd_15_50"`
	Usd_50_150   uint64 `ch:"usd_50_150" json:"usd_50_150"`
	Usd_150_500  uint64 `ch:"usd_150_500" json:"usd_150_500"`
	Usd_500_2000 uint64 `ch:"usd_500_2000" json:"usd_500_2000"`
	Usd_2000     uint64 `ch:"usd_2000" json:"usd_2000"`
}

func SwapsDistributionSqlQuery(config *core.Config, period models.Period, dex models.Dex) string {
	periodParams := models.PeriodParamsMap[period]
	return fmt.Sprint(`
SELECT
    countIf(usd <= 1) AS usd_1,
    countIf((usd > 1) AND (usd <= 5)) AS usd_1_5,
    countIf((usd > 5) AND (usd < 15)) AS usd_5_15,
    countIf((usd >= 15) AND (usd < 50)) AS usd_15_50,
    countIf((usd >= 50) AND (usd < 150)) AS usd_50_150,
    countIf((usd >= 150) AND (usd < 500)) AS usd_150_500,
    countIf((usd >= 500) AND (usd < 2000)) AS usd_500_2000,
    countIf(usd >= 2000) AS usd_2000
FROM
(
    SELECT
        (`, UsdInField, ` + `, UsdOutField, `) / 2 AS usd
	FROM `, config.DbName, `.swaps
	WHERE time >= `, periodParams.ToStartOf, `(subtractDays(now(), `, periodParams.WindowInDays, `))
	AND `, dex.WhereStatement("dex"), `
)
`)
}
