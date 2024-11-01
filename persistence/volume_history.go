package persistence

import (
	"fmt"
	"time"
	"tondexer/core"
	"tondexer/models"
)

type VolumeHistoryEntry struct {
	Period    time.Time `json:"period" ch:"period"`
	VolumeUsd uint64    `json:"volume" ch:"volume_usd"`
	Number    uint64    `json:"number" ch:"number"`
}

func VolumeHistorySqlQuery(config *core.Config, period models.Period) string {
	periodParams := models.PeriodParamsMap[period]
	return fmt.Sprint(`
SELECT `,
		periodParams.ToStartOf, `(time) AS period,
    toUInt64(sum(`, UsdInField, `) + sum(`, UsdOutField, `)) AS volume_usd,
    count() AS number
FROM `, config.DbName, `.swaps
WHERE time >= `, periodParams.ToStartOf, `(subtractDays(now(), `, periodParams.WindowInDays, `))
GROUP BY period
ORDER BY period ASC WITH FILL STEP `, periodParams.ToInterval, `(1)`)
}
