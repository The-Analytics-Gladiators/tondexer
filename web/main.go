package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"tondexer/core"
	"tondexer/models"
	"tondexer/persistence"
)

type PeriodRequest struct {
	Period string `form:"period" binding:"required,oneof=day week month"`
}

func main() {
	var cfg core.Config

	if err := cleanenv.ReadConfig(os.Args[1], &cfg); err != nil {
		panic(err)
	}

	route := gin.Default()

	route.GET("/api/summary", summary(&cfg))
	route.GET("/api/swaps/latest", latestSwaps(&cfg))
	route.GET("/api/volumeHistory", periodArrayRequest(&cfg, func(config *core.Config, period models.Period) ([]persistence.VolumeHistoryEntry, error) {
		return persistence.ReadArrayFromClickhouse[persistence.VolumeHistoryEntry](config, persistence.VolumeHistorySqlQuery(config, period))
	}))
	route.GET("/api/swaps/top", periodArrayRequest(&cfg, func(config *core.Config, period models.Period) ([]persistence.EnrichedSwapCH, error) {
		return persistence.ReadArrayFromClickhouse[persistence.EnrichedSwapCH](config, persistence.TopSwapsSqlQuery(config, period))
	}))
	route.GET("/api/pools/top", periodArrayRequest[persistence.PoolVolume](&cfg, func(config *core.Config, period models.Period) ([]persistence.PoolVolume, error) {
		return persistence.ReadArrayFromClickhouse[persistence.PoolVolume](&cfg, persistence.TopPoolsRequest(config, period))
	}))
	route.GET("api/jettons/top", periodArrayRequest[persistence.JettonVolume](&cfg, func(config *core.Config, period models.Period) ([]persistence.JettonVolume, error) {
		return persistence.ReadArrayFromClickhouse[persistence.JettonVolume](&cfg, persistence.TopJettonRequest(config, period))
	}))
	route.GET("/api/users/top", periodArrayRequest[persistence.UserVolume](&cfg, func(config *core.Config, period models.Period) ([]persistence.UserVolume, error) {
		return persistence.ReadArrayFromClickhouse[persistence.UserVolume](&cfg, persistence.TopUsersRequest(config, period))
	}))
	route.GET("/api/referrers/top", periodArrayRequest[persistence.UserVolume](&cfg, func(config *core.Config, period models.Period) ([]persistence.UserVolume, error) {
		return persistence.ReadArrayFromClickhouse[persistence.UserVolume](&cfg, persistence.TopReferrersRequest(config, period))
	}))
	route.GET("/api/profiters/top", periodArrayRequest[persistence.UserVolume](&cfg, func(config *core.Config, period models.Period) ([]persistence.UserVolume, error) {
		return persistence.ReadArrayFromClickhouse[persistence.UserVolume](&cfg, persistence.TopUsersProfiters(config, period))
	}))
	route.GET("/api/arbitrages/volumeHistory", periodArrayRequest(&cfg, func(config *core.Config, period models.Period) ([]persistence.ArbitrageHistoryEntry, error) {
		return persistence.ReadArrayFromClickhouse[persistence.ArbitrageHistoryEntry](config, persistence.ArbitrageHistorySqlQuery(config, period))
	}))
	route.GET("/api/arbitrages/latest", latestArbitrages(&cfg))

	route.Run(":8088")
}

func periodArrayRequest[T any](cfg *core.Config, fetchEntitiesFunc func(config *core.Config, period models.Period) ([]T, error)) func(c *gin.Context) {
	return func(c *gin.Context) {
		var request PeriodRequest
		if err := c.ShouldBindQuery(&request); err != nil {
			c.JSON(400, gin.H{"msg": err.Error()})
			return
		}
		period, e := models.ParsePeriod(request.Period)
		if e != nil {
			c.JSON(400, gin.H{"msg": e.Error()})
		}

		entities, e := fetchEntitiesFunc(cfg, period)
		if e != nil {
			c.JSON(500, gin.H{"msg": e.Error()})
		}

		c.JSON(200, entities)
	}
}

func latestSwaps(cfg *core.Config) func(c *gin.Context) {
	return func(c *gin.Context) {
		var request struct {
			Limit uint64 `form:"limit"`
		}
		if err := c.ShouldBindQuery(&request); err != nil {
			c.JSON(400, gin.H{"msg": err.Error()})
			return
		}

		swaps, e := persistence.ReadArrayFromClickhouse[persistence.EnrichedSwapCH](cfg, persistence.LatestSwapsSqlQuery(cfg, request.Limit))
		if e != nil {
			c.JSON(200, gin.H{"msg": e.Error()})
			return
		}

		c.JSON(200, swaps)
	}
}

// TODO refactor latest*
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

func summary(cfg *core.Config) func(c *gin.Context) {
	return func(c *gin.Context) {
		var request PeriodRequest

		if err := c.ShouldBindQuery(&request); err != nil {
			c.JSON(400, gin.H{"msg": err.Error()})
			return
		}
		period, e := models.ParsePeriod(request.Period)
		if e != nil {
			c.JSON(400, gin.H{"msg": e.Error()})
		}

		summary, e := persistence.ReadSummaryStats(cfg, period)
		if e != nil {
			c.JSON(500, gin.H{"msg": e.Error()})
		}

		c.JSON(200, summary)
	}
}
