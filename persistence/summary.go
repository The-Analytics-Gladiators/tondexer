package persistence

import (
	"fmt"
	"tondexer/core"
	"tondexer/models"
)

func SwapsSummarySql(config *core.DbConfig, period models.Period, dex models.Dex) string {
	periodParams := models.PeriodParamsMap[period]
	return fmt.Sprintf(`
SELECT
    toUInt64((sum(%v) + sum(%v)) / 2) AS volume,
    count() AS number,
    length(groupUniqArrayArray([jetton_in, jetton_out])) AS unique_tokens,
    uniq(sender) AS unique_users
FROM %v.swaps
WHERE time >= %v(subtractDays(now(), %v))
AND %v
AND %v < 1000000 AND %v < 1000000`, UsdInField, UsdOutField,
		config.DbName, periodParams.ToStartOf, periodParams.WindowInDays,
		dex.WhereStatement("dex"),
		UsdInField, UsdOutField)
}
