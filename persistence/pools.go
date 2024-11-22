package persistence

import (
	"fmt"
	"math/big"
	"tondexer/core"
	"tondexer/models"
)

type PoolVolume struct {
	PoolAddress       string   `json:"pool_address" ch:"pool_address"`
	JettonIn          string   `json:"jetton_in" ch:"jetton_in"`
	AmountIn          *big.Int `json:"amount_in" ch:"in_amount"`
	AmountInUsd       float64  `json:"amount_in_usd" ch:"amount_in_usd"`
	JettonInName      string   `json:"jetton_in_name" ch:"jetton_in_name"`
	JettonInSymbol    string   `json:"jetton_in_symbol" ch:"jetton_in_symbol"`
	JettonInDecimals  uint64   `json:"jetton_in_decimals" ch:"in_jetton_decimals"`
	JettonOut         string   `json:"jetton_out" ch:"jetton_out"`
	AmountOut         *big.Int `json:"amount_out" ch:"out_amount"`
	AmountOutUsd      float64  `json:"amount_out_usd" ch:"amount_out_usd"`
	JettonOutName     string   `json:"jetton_out_name" ch:"jetton_out_name"`
	JettonOutSymbol   string   `json:"jetton_out_symbol" ch:"jetton_out_symbol"`
	JettonOutDecimals uint64   `json:"jetton_out_decimals" ch:"out_jetton_decimals"`
	AmountUsd         float64  `json:"amount_usd" ch:"amount_usd"`
	Dex               string   `json:"dex" ch:"pool_dex"`
}

func TopPoolsRequest(config *core.Config, period models.Period, dex models.Dex) string {
	periodParams := models.PeriodParamsMap[period]
	return fmt.Sprint(`
SELECT
    pool_address,
    anyHeavy(jetton_in) AS jetton_in,
    sum(amount_in) AS in_amount,
    sum(`, UsdInField, `) AS amount_in_usd,
    anyHeavy(jetton_in_name) AS jetton_in_name,
    anyHeavy(`, Symbol("jetton_in_symbol"), `) AS jetton_in_symbol,
    anyHeavy(jetton_in_decimals) AS in_jetton_decimals,
    anyHeavy(jetton_out) AS jetton_out,
    sum(amount_out) AS out_amount,
    sum(`, UsdOutField, `) AS amount_out_usd,
    anyHeavy(jetton_out_name) AS jetton_out_name,
    anyHeavy(`, Symbol("jetton_out_symbol"), `) AS jetton_out_symbol,
    anyHeavy(jetton_out_decimals) AS out_jetton_decimals,
    (amount_in_usd + amount_out_usd) / 2 AS amount_usd,
	anyHeavy(dex) as pool_dex
FROM `, config.DbName, `.swaps
WHERE time >= `, periodParams.ToStartOf, `(subtractDays(now(), `, periodParams.WindowInDays, `))
AND `, dex.WhereStatement("dex"), `
AND `, UsdInField, ` < 1000000 AND `, UsdOutField, ` < 1000000
GROUP BY pool_address
ORDER BY amount_usd DESC
LIMIT 15
`)
}
