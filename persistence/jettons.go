package persistence

import (
	"fmt"
	"tondexer/core"
	"tondexer/models"
)

type JettonVolume struct {
	JettonAddress  string  `json:"jetton_address" ch:"jetton_address"`
	JettonSymbol   string  `json:"jetton_symbol" ch:"jetton_symbol"`
	JettonName     string  `json:"jetton_name" ch:"jetton_name"`
	JettonDecimals uint64  `json:"jetton_decimals" ch:"jetton_decimals"`
	JettonAmount   uint64  `json:"jetton_amount" ch:"jetton_amount"`
	JettonUsd      float64 `json:"jetton_usd" ch:"jetton_usd"`
}

func TopJettonRequest(config *core.Config, period models.Period) string {
	periodParams := models.PeriodParamsMap[period]

	return fmt.Sprint(`
SELECT
    jetton_address,
    any(jetton_symbol) AS jetton_symbol,
    any(jetton_name) AS jetton_name,
    any(jetton_decimals) AS jetton_decimals,
    sum(amount) AS jetton_amount,
    sum(jetton_usd) AS jetton_usd
FROM
(
    SELECT
		time,
        jetton_in AS jetton_address,
        jetton_in_symbol AS jetton_symbol,
        jetton_in_name AS jetton_name,
    	jetton_in_decimals AS jetton_decimals,
        amount_in AS amount,
        `, UsdInField, ` AS jetton_usd
    FROM swaps
    UNION ALL
    SELECT
		time,
        jetton_out AS jetton_address,
        jetton_out_symbol AS jetton_symbol,
        jetton_out_name AS jetton_name,
		jetton_out_decimals AS jetton_decimals,
        amount_out AS amount,
        `, UsdOutField, ` AS jetton_usd
    FROM `, config.DbName, `.swaps
)
WHERE time >= `, periodParams.ToStartOf, `(subtractDays(now(), `, periodParams.WindowInDays, `))
GROUP BY jetton_address
ORDER BY jetton_usd DESC
LIMIT 15
`)
}
