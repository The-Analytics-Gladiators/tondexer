package persistence

import (
	"fmt"
	"math/big"
	"time"
	"tondexer/core"
	"tondexer/models"
)

type ArbitrageHistoryEntry struct {
	Period    time.Time `json:"period" ch:"period"`
	UsdProfit float64   `json:"usd_profit" ch:"usd_profit"`
	UsdVolume float64   `json:"usd_volume" ch:"usd_volume"`
	Number    uint64    `json:"number" ch:"number"`
}

func ArbitrageHistorySqlQuery(config *core.Config, period models.Period) string {
	periodParams := models.PeriodParamsMap[period]

	return fmt.Sprint(`
SELECT `,
		periodParams.ToStartOf, `(time) AS period,
	sum(`, UsdField("out"), ` - `, UsdField("in"), `) AS usd_profit,
	sum(`, UsdField("in"), `) AS usd_volume,
	count() AS number
FROM `, config.DbName, `.arbitrages
WHERE time >= `, periodParams.ToStartOf, `(subtractDays(now(), `, periodParams.WindowInDays, `))
GROUP BY period
ORDER BY period ASC WITH FILL STEP `, periodParams.ToInterval, `(1)`,
	)
}

type EnrichedArbitrageCH struct {
	Time             time.Time  `json:"time" ch:"time"`
	Sender           string     `json:"sender" ch:"sender"`
	Traces           []string   `json:"traces" ch:"traces"`
	AmountIn         *big.Int   `json:"amount_in" ch:"amount_in"`
	AmountInJettons  float64    `json:"amount_in_jettons" ch:"amount_in_jettons"`
	AmountOut        *big.Int   `json:"amount_out" ch:"amount_out"`
	AmountOutJettons float64    `json:"amount_out_jettons" ch:"amount_out_jettons"`
	AmountInUSD      float64    `json:"amount_in_usd" ch:"amount_in_usd"`
	AmountOutUSD     float64    `json:"amount_out_usd" ch:"amount_out_usd"`
	Jetton           string     `json:"jetton" ch:"jetton"`
	JettonSymbol     string     `json:"jetton_symbol" ch:"jetton_symbol"`
	JettonName       string     `json:"jetton_name" ch:"jetton_name"`
	JettonUsdRate    float64    `json:"jetton_usd_rate" ch:"jetton_usd_rate"`
	JettonDecimals   uint64     `json:"jetton_decimals" ch:"jetton_decimals"`
	AmountsPath      []*big.Int `json:"amounts_path" ch:"amounts_path"`
	JettonsPath      []string   `json:"jettons_path" ch:"jettons_path"`
	JettonNames      []string   `json:"jetton_names" ch:"jetton_names"`
	JettonSymbols    []string   `json:"jetton_symbols" ch:"jetton_symbols"`
	JettonUsdRates   []float64  `json:"jetton_usd_rates" ch:"jetton_usd_rates"`
	JettonsDecimals  []uint64   `json:"jettons_decimals" ch:"jettons_decimals"`
	AmountsJettons   []float64  `json:"amounts_jettons" ch:"amounts_jettons"`
	AmountsUsdPath   []float64  `json:"amounts_usd_path" ch:"amounts_usd_path"`
	PoolsPath        []string   `json:"pools_path" ch:"pools_path"`
	Dexes            []string   `json:"dexes" ch:"dexes"`
}

func LatestArbitragesSqlQuery(config *core.Config, limit uint64) string {
	return fmt.Sprint(`
SELECT
    time,
    sender,
    arrayDistinct(trace_ids) AS traces,
    amount_in,
    toFloat64(amount_in) / pow(10, jetton_decimals) AS amount_in_jettons,
    amount_out,
    toFloat64(amount_out) / pow(10, jetton_decimals) AS amount_out_jettons,
    amount_in_jettons * jetton_usd_rate AS amount_in_usd,
    amount_out_jettons * jetton_usd_rate AS amount_out_usd,
    jetton,
    jetton_symbol,
    jetton_name,
    jetton_usd_rate,
    jetton_decimals,
    amounts_path,
    jettons_path,
    jetton_names,
    jetton_symbols,
    jetton_usd_rates,
    jettons_decimals,
    arrayMap(i -> (toFloat64(amounts_path[i]) / pow(10, jettons_decimals[i])), range(1, length(amounts_path) + 1)) AS amounts_jettons,
    arrayMap(i -> ((amounts_jettons[i]) * (jetton_usd_rates[i])), range(1, length(amounts_path) + 1)) AS amounts_usd_path,
    pools_path,
    dexes
FROM `, config.DbName, `.arbitrages
ORDER BY time DESC
LIMIT `, limit)
}