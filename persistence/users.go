package persistence

import (
	"fmt"
	"tondexer/core"
	"tondexer/models"
)

type UserVolume struct {
	UserAddress string  `json:"user_address" ch:"sender"`
	AmountUsd   float64 `json:"amount_usd" ch:"amount_usd"`
	Tokens      uint64  `json:"tokens" ch:"tokens"`
	Count       uint64  `json:"count" ch:"count"`
}

func TopReferrersRequest(config *core.Config, period models.Period, dex models.Dex) string {
	periodParams := models.PeriodParamsMap[period]
	return fmt.Sprint(`
SELECT
    referral_address AS sender,
    sum(`, UsdReferralField, `) AS amount_usd,
    uniq(jetton_out) AS tokens,
    count() AS count
FROM `, config.DbName, `.swaps
WHERE time >= `, periodParams.ToStartOf, `(subtractDays(now(), `, periodParams.WindowInDays, `))
GROUP BY sender
ORDER BY amount_usd DESC
LIMIT 15
`)
}

func TopUsersRequest(config *core.Config, period models.Period, dex models.Dex) string {
	periodParams := models.PeriodParamsMap[period]
	return fmt.Sprint(`
SELECT
    sender,
    sum((`, UsdInField, ` + `, UsdOutField, `) / 2) AS amount_usd,
    uniqArray([jetton_in, jetton_out]) AS tokens,
    count() AS count
FROM `, config.DbName, `.swaps
WHERE time >= `, periodParams.ToStartOf, `(subtractDays(now(), `, periodParams.WindowInDays, `))
AND `, dex.WhereStatement("dex"), `
GROUP BY sender
ORDER BY amount_usd DESC
LIMIT 15
`)
}

func TopUsersProfiters(config *core.Config, period models.Period) string {
	periodParams := models.PeriodParamsMap[period]
	return fmt.Sprint(`
SELECT
    sender,
    sum(`, UsdOutField, ` - `, UsdInField, `) AS amount_usd,
    uniqArray([jetton_in, jetton_out]) AS tokens,
    count() AS count
FROM `, config.DbName, `.swaps
WHERE time >= `, periodParams.ToStartOf, `(subtractDays(now(), `, periodParams.WindowInDays, `))
AND jetton_in_usd_rate != 0 AND jetton_out_usd_rate != 0
GROUP BY sender
ORDER BY amount_usd DESC
LIMIT 15
`)
}
