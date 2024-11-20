package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"tondexer/core"
	"tondexer/models"
	"tondexer/persistence"
)

type DexPeriodRequest struct {
	Period string `form:"period" binding:"required,oneof=day week month"`
	Dex    string `form:"dex" binding:"omitempty,oneof=all stonfi dedust"`
}

func main() {
	var cfg core.Config

	if err := cleanenv.ReadConfig(os.Args[1], &cfg); err != nil {
		panic(err)
	}

	route := gin.Default()

	route.GET("/api/summary", oneRowPeriodDexRequest[persistence.SummaryStats](&cfg, func(cfg *core.Config, period models.Period, dex models.Dex) string {
		return persistence.SwapsSummarySql(cfg, period, dex)
	}))
	route.GET("/api/swaps/latest", latestSwaps(&cfg))
	route.GET("/api/volumeHistory", periodDexArrayRequest[persistence.VolumeHistoryEntry](&cfg, func(config *core.Config, period models.Period, dex models.Dex) string {
		return persistence.VolumeHistorySqlQuery(config, period, dex)
	}))
	route.GET("/api/swaps/top", periodDexArrayRequest[persistence.EnrichedSwapCH](&cfg, func(config *core.Config, period models.Period, dex models.Dex) string {
		return persistence.TopSwapsSqlQuery(config, period, dex)
	}))
	route.GET("/api/pools/top", periodDexArrayRequest[persistence.PoolVolume](&cfg, func(config *core.Config, period models.Period, dex models.Dex) string {
		return persistence.TopPoolsRequest(config, period, dex)
	}))
	route.GET("api/jettons/top", periodDexArrayRequest[persistence.JettonVolume](&cfg, func(config *core.Config, period models.Period, dex models.Dex) string {
		return persistence.TopJettonRequest(config, period, dex)
	}))
	route.GET("/api/users/top", periodDexArrayRequest[persistence.UserVolume](&cfg, func(config *core.Config, period models.Period, dex models.Dex) string {
		return persistence.TopUsersRequest(config, period, dex)
	}))
	route.GET("/api/referrers/top", periodDexArrayRequest[persistence.UserVolume](&cfg, func(config *core.Config, period models.Period, dex models.Dex) string {
		return persistence.TopReferrersRequest(config, period, dex)
	}))
	route.GET("/api/profiters/top", periodDexArrayRequest[persistence.UserVolume](&cfg, func(config *core.Config, period models.Period, dex models.Dex) string {
		//Deprecated
		return persistence.TopUsersProfiters(config, period)
	}))
	route.GET("/api/swaps/distribution", oneRowPeriodDexRequest[persistence.SwapDistribution](&cfg, func(cfg *core.Config, period models.Period, dex models.Dex) string {
		return persistence.SwapsDistributionSqlQuery(cfg, period, dex)
	}))

	route.GET("/api/arbitrages/latest", latestArbitrages(&cfg))
	route.GET("/api/arbitrages/top", periodDexArrayRequest[persistence.EnrichedArbitrageCH](&cfg, func(config *core.Config, period models.Period, dex models.Dex) string {
		return persistence.TopArbitragesSqlQuery(config, period)
	}))
	route.GET("/api/arbitrages/volumeHistory", periodDexArrayRequest[persistence.ArbitrageHistoryEntry](&cfg, func(config *core.Config, period models.Period, _ models.Dex) string {
		return persistence.ArbitrageHistorySqlQuery(config, period)
	}))
	route.GET("/api/arbitrages/distribution", oneRowPeriodDexRequest[persistence.ArbitrageDistribution](&cfg, func(cfg *core.Config, period models.Period, _ models.Dex) string {
		return persistence.ArbitrageDistributionSqlQuery(cfg, period)
	}))
	route.GET("/api/arbitrages/users/top", periodDexArrayRequest[persistence.TopArbitrageUser](&cfg, func(cfg *core.Config, period models.Period, dex models.Dex) string {
		return persistence.TopArbitrageUsersSql(cfg, period)
	}))
	route.GET("/api/arbitrages/jettons/top", periodDexArrayRequest[persistence.TopArbitrageJetton](&cfg, func(cfg *core.Config, period models.Period, dex models.Dex) string {
		return persistence.TopArbitrageJettonsSql(cfg, period)
	}))

	route.Run(":8088")
}

func periodAndDexFromRequest(request DexPeriodRequest) (models.Period, models.Dex, error) {
	period, e := models.ParsePeriod(request.Period)
	if e != nil {
		return "", "", e
	}
	dex, e := models.ParseDex(request.Dex)
	if e != nil {
		return "", "", e
	}

	return period, dex, nil
}

func periodDexArrayRequest[T any](cfg *core.Config, sqlFunc func(cfg *core.Config, period models.Period, dex models.Dex) string) func(c *gin.Context) {
	return func(c *gin.Context) {
		var request DexPeriodRequest
		if err := c.ShouldBindQuery(&request); err != nil {
			log.Printf("Error binding request %v\n", err)
			c.JSON(400, gin.H{"msg": err.Error()})
			return
		}
		period, dex, e := periodAndDexFromRequest(request)
		if e != nil {
			log.Printf("Invalid request: %v - %v\n", request.Period, request.Dex)
			c.JSON(400, gin.H{"msg": e.Error()})
		}

		entities, e := persistence.ReadArrayFromClickhouse[T](cfg, sqlFunc(cfg, period, dex))
		if e != nil {
			log.Printf("Error queryin entities: %v\n", e)
			c.JSON(500, gin.H{"msg": e.Error()})
		}

		c.JSON(200, entities)
	}
}

func latestSwaps(cfg *core.Config) func(c *gin.Context) {
	return func(c *gin.Context) {
		var request struct {
			Limit uint64 `form:"limit"`
			Dex   string `form:"dex" binding:"omitempty,oneof=all stonfi dedust"`
		}
		if err := c.ShouldBindQuery(&request); err != nil {
			c.JSON(400, gin.H{"msg": err.Error()})
			return
		}
		dex, e := models.ParseDex(request.Dex)
		if e != nil {
			log.Printf("Invalid dex: %v\n", request.Dex)
			c.JSON(400, gin.H{"msg": e.Error()})
		}

		swaps, e := persistence.ReadArrayFromClickhouse[persistence.EnrichedSwapCH](cfg, persistence.LatestSwapsSqlQuery(cfg, request.Limit, dex))
		if e != nil {
			c.JSON(200, gin.H{"msg": e.Error()})
			return
		}

		c.JSON(200, swaps)
	}
}

func latestArbitrages(cfg *core.Config) func(c *gin.Context) {
	return func(c *gin.Context) {
		var request struct {
			Limit uint64 `form:"limit"`
		}
		if err := c.ShouldBindQuery(&request); err != nil {
			c.JSON(400, gin.H{"msg": err.Error()})
			return
		}

		swaps, e := persistence.ReadArrayFromClickhouse[persistence.EnrichedArbitrageCH](cfg, persistence.LatestArbitragesSqlQuery(cfg, request.Limit))
		if e != nil {
			c.JSON(200, gin.H{"msg": e.Error()})
			return
		}

		c.JSON(200, swaps)
	}
}

func oneRowPeriodDexRequest[T any](cfg *core.Config, sqlFunc func(cfg *core.Config, period models.Period, dex models.Dex) string) func(c *gin.Context) {
	return func(c *gin.Context) {
		var request DexPeriodRequest

		if err := c.ShouldBindQuery(&request); err != nil {
			c.JSON(400, gin.H{"msg": err.Error()})
			return
		}
		period, dex, e := periodAndDexFromRequest(request)
		if e != nil {
			log.Printf("Invalid request: %v - %v\n", request.Period, request.Dex)
			c.JSON(400, gin.H{"msg": e.Error()})
		}

		result, e := persistence.ReadSingleRow[T](cfg, sqlFunc(cfg, period, dex))
		if e != nil {
			log.Printf("Error querying one row: %v\n", e)
			c.JSON(500, gin.H{"msg": e.Error()})
		}

		c.JSON(200, result)
	}
}
