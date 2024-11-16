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
	route.GET("/api/volumeHistory", periodDexArrayRequest(&cfg, func(config *core.Config, period models.Period, dex models.Dex) ([]persistence.VolumeHistoryEntry, error) {
		return persistence.ReadArrayFromClickhouse[persistence.VolumeHistoryEntry](config, persistence.VolumeHistorySqlQuery(config, period, dex))
	}))
	route.GET("/api/swaps/top", periodDexArrayRequest(&cfg, func(config *core.Config, period models.Period, dex models.Dex) ([]persistence.EnrichedSwapCH, error) {
		return persistence.ReadArrayFromClickhouse[persistence.EnrichedSwapCH](config, persistence.TopSwapsSqlQuery(config, period, dex))
	}))
	route.GET("/api/pools/top", periodDexArrayRequest[persistence.PoolVolume](&cfg, func(config *core.Config, period models.Period, dex models.Dex) ([]persistence.PoolVolume, error) {
		return persistence.ReadArrayFromClickhouse[persistence.PoolVolume](&cfg, persistence.TopPoolsRequest(config, period, dex))
	}))
	route.GET("api/jettons/top", periodDexArrayRequest[persistence.JettonVolume](&cfg, func(config *core.Config, period models.Period, dex models.Dex) ([]persistence.JettonVolume, error) {
		return persistence.ReadArrayFromClickhouse[persistence.JettonVolume](&cfg, persistence.TopJettonRequest(config, period, dex))
	}))
	route.GET("/api/users/top", periodDexArrayRequest[persistence.UserVolume](&cfg, func(config *core.Config, period models.Period, dex models.Dex) ([]persistence.UserVolume, error) {
		return persistence.ReadArrayFromClickhouse[persistence.UserVolume](&cfg, persistence.TopUsersRequest(config, period, dex))
	}))
	route.GET("/api/referrers/top", periodDexArrayRequest[persistence.UserVolume](&cfg, func(config *core.Config, period models.Period, dex models.Dex) ([]persistence.UserVolume, error) {
		return persistence.ReadArrayFromClickhouse[persistence.UserVolume](&cfg, persistence.TopReferrersRequest(config, period, dex))
	}))
	route.GET("/api/profiters/top", periodDexArrayRequest[persistence.UserVolume](&cfg, func(config *core.Config, period models.Period, dex models.Dex) ([]persistence.UserVolume, error) {
		//Deprecated
		return persistence.ReadArrayFromClickhouse[persistence.UserVolume](&cfg, persistence.TopUsersProfiters(config, period))
	}))
	route.GET("/api/arbitrages/volumeHistory", periodArrayRequest(&cfg, func(config *core.Config, period models.Period) ([]persistence.ArbitrageHistoryEntry, error) {
		return persistence.ReadArrayFromClickhouse[persistence.ArbitrageHistoryEntry](config, persistence.ArbitrageHistorySqlQuery(config, period))
	}))
	route.GET("/api/arbitrages/latest", latestArbitrages(&cfg))

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

func periodDexArrayRequest[T any](cfg *core.Config, fetchEntitiesFunc func(config *core.Config, period models.Period, dex models.Dex) ([]T, error)) func(c *gin.Context) {
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

		entities, e := fetchEntitiesFunc(cfg, period, dex)
		if e != nil {
			log.Printf("Error queryin entities: %v\n", e)
			c.JSON(500, gin.H{"msg": e.Error()})
		}

		c.JSON(200, entities)
	}
}

func periodArrayRequest[T any](cfg *core.Config, fetchEntitiesFunc func(config *core.Config, period models.Period) ([]T, error)) func(c *gin.Context) {
	return func(c *gin.Context) {
		var request PeriodRequest
		if err := c.ShouldBindQuery(&request); err != nil {
			log.Printf("Error binding request %v\n", err)
			c.JSON(400, gin.H{"msg": err.Error()})
			return
		}
		period, e := models.ParsePeriod(request.Period)
		if e != nil {
			log.Printf("Invalid period: %v\n", request.Period)
			c.JSON(400, gin.H{"msg": e.Error()})
		}

		entities, e := fetchEntitiesFunc(cfg, period)
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

		summary, e := persistence.ReadSummaryStats(cfg, period, dex)
		if e != nil {
			log.Printf("Error querying summary: %v\n", e)
			c.JSON(500, gin.H{"msg": e.Error()})
		}

		c.JSON(200, summary)
	}
}
