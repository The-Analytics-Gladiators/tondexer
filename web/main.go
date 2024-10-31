package main

import (
	"TonArb/core"
	"TonArb/models"
	"TonArb/persistence"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
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

	route.GET("/summary", summary(&cfg))
	route.GET("/swaps/latest", latestSwaps(&cfg))
	route.GET("/volumeHistory", periodArrayRequest(&cfg, func(config *core.Config, period models.Period) ([]persistence.VolumeHistoryEntry, error) {
		return persistence.ReadArrayFromClickhouse[persistence.VolumeHistoryEntry](config, persistence.VolumeHistorySqlQuery(config, period))
	}))

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
		entitiesJson, e := json.Marshal(entities)
		if e != nil {
			c.JSON(500, gin.H{"msg": e.Error()})
			return
		}

		c.JSON(200, string(entitiesJson))
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

		swaps, e := persistence.ReadArrayFromClickhouse[persistence.EnrichedSwapCH](cfg, persistence.LastSwapsSqlQuery(cfg, request.Limit))

		if e != nil {
			c.JSON(200, gin.H{"msg": e.Error()})
			return
		}

		swapsJson, err := json.Marshal(swaps)
		if err != nil {
			c.JSON(500, gin.H{"msg": e.Error()})
			return
		}
		c.JSON(200, string(swapsJson))
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

		summaryJson, e := json.Marshal(summary)
		if e != nil {
			c.JSON(500, gin.H{"msg": e.Error()})
			return
		}

		c.JSON(200, string(summaryJson))
	}
}
