package persistence

import (
	"fmt"
	"math/big"
	"tondexer/core"
	"tondexer/models"
)

type JettonVolume struct {
	JettonAddress  string   `json:"jetton_address" ch:"jetton_address"`
	JettonSymbol   string   `json:"jetton_symbol" ch:"jetton_symbol"`
	JettonName     string   `json:"jetton_name" ch:"jetton_name"`
	JettonDecimals uint64   `json:"jetton_decimals" ch:"jetton_decimals"`
	JettonAmount   *big.Int `json:"jetton_amount" ch:"jetton_amount"`
	JettonUsd      float64  `json:"jetton_usd" ch:"jetton_usd"`
}

func TopJettonRequest(config *core.DbConfig, period models.Period, dex models.Dex) string {
	periodParams := models.PeriodParamsMap[period]

	return fmt.Sprint(`
SELECT
    any(jetton_address) AS jetton_address,
    jetton_symbol,
    any(jetton_name) AS jetton_name,
    any(jetton_decimals) AS jetton_decimals,
    sum(amount) AS jetton_amount,
    sum(jetton_usd_inner) AS jetton_usd
FROM
(
    SELECT
		time,
    	dex,
        jetton_in AS jetton_address,
        `, Symbol("jetton_in_symbol"), ` AS jetton_symbol,
        jetton_in_name AS jetton_name,
    	jetton_in_decimals AS jetton_decimals,
        amount_in AS amount,
        `, UsdInField, ` AS jetton_usd_inner
    FROM `, config.DbName, `.swaps
    UNION ALL
    SELECT
		time,
		dex,
        jetton_out AS jetton_address,
        `, Symbol("jetton_out_symbol"), ` AS jetton_symbol,
        jetton_out_name AS jetton_name,
		jetton_out_decimals AS jetton_decimals,
        amount_out AS amount,
        `, UsdOutField, ` AS jetton_usd_inner
    FROM `, config.DbName, `.swaps
)
WHERE time >= `, periodParams.ToStartOf, `(subtractDays(now(), `, periodParams.WindowInDays, `))
AND `, dex.WhereStatement("dex"), ` AND jetton_usd_inner < 1000000
GROUP BY jetton_symbol
ORDER BY jetton_usd DESC
LIMIT 10
`)
}
