package persistence

import (
	"fmt"
	"math/big"
	"time"
	"tondexer/core"
	"tondexer/models"
)

type VolumeHistoryEntry struct {
	Period          time.Time `json:"period" ch:"period"`
	StonfiVolumeUsd *big.Int  `json:"stonfi_volume" ch:"stonfi_volume_usd"`
	DedustVolumeUsd *big.Int  `json:"dedust_volume" ch:"dedust_volume_usd"`
	Number          uint64    `json:"number" ch:"number"`
}

func VolumeHistorySqlQuery(config *core.Config, period models.Period, dex models.Dex) string {
	periodParams := models.PeriodParamsMap[period]
	return fmt.Sprint(`
SELECT `,
		periodParams.ToStartOf, `(time) AS period,
    toUInt256((sumIf(`, UsdInField, `, dex = 'StonfiV1' OR dex = 'StonfiV2') + sumIf(`, UsdOutField, `, dex = 'StonfiV1' OR dex = 'StonfiV2')) / 2) AS stonfi_volume_usd,
    toUInt256((sumIf(`, UsdInField, `, dex = 'DeDust') + sumIf(`, UsdOutField, `, dex = 'DeDust')) / 2) AS dedust_volume_usd,
    count() AS number
FROM `, config.DbName, `.swaps
WHERE time >= `, periodParams.ToStartOf, `(subtractDays(now(), `, periodParams.WindowInDays, `))
AND `, dex.WhereStatement("dex"), `
GROUP BY period
ORDER BY period ASC WITH FILL STEP `, periodParams.ToInterval, `(1)`)
}
