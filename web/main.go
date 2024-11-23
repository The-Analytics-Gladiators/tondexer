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

type Config struct {
	DbHost     string `yaml:"db_host" env:"DB_HOST" env-default:"localhost"`
	DbPort     uint   `yaml:"db_port" env:"DB_PORT" env-default:"9000"`
	DbUser     string `yaml:"db_user" env:"DB_USER" env-default:"default"`
	DbPassword string `yaml:"db_password" env:"DB_PASSWORD" env-default:""`
	DbName     string `yaml:"db_name" env:"DB_NAME" env-default:"default"`
}

func main() {
	var cfg Config

	if err := cleanenv.ReadConfig(os.Args[1], &cfg); err != nil {
		panic(err)
	}

	dbConfig := core.DbConfig{
		DbHost:     cfg.DbHost,
		DbPort:     cfg.DbPort,
		DbUser:     cfg.DbUser,
		DbPassword: cfg.DbPassword,
		DbName:     cfg.DbName,
	}

	route := gin.Default()

	route.GET("/api/summary", oneRowPeriodDexRequest[persistence.SummaryStats](&dbConfig, func(cfg *core.DbConfig, period models.Period, dex models.Dex) string {
		return persistence.SwapsSummarySql(cfg, period, dex)
	}))
	route.GET("/api/swaps/latest", latestSwaps(&dbConfig))
	route.GET("/api/volumeHistory", periodDexArrayRequest[persistence.VolumeHistoryEntry](&dbConfig, func(config *core.DbConfig, period models.Period, dex models.Dex) string {
		return persistence.VolumeHistorySqlQuery(config, period, dex)
	}))
	route.GET("/api/swaps/top", periodDexArrayRequest[persistence.EnrichedSwapCH](&dbConfig, func(config *core.DbConfig, period models.Period, dex models.Dex) string {
		return persistence.TopSwapsSqlQuery(config, period, dex)
	}))
	route.GET("/api/pools/top", periodDexArrayRequest[persistence.PoolVolume](&dbConfig, func(config *core.DbConfig, period models.Period, dex models.Dex) string {
		return persistence.TopPoolsRequest(config, period, dex)
	}))
	route.GET("api/jettons/top", periodDexArrayRequest[persistence.JettonVolume](&dbConfig, func(config *core.DbConfig, period models.Period, dex models.Dex) string {
		return persistence.TopJettonRequest(config, period, dex)
	}))
	route.GET("/api/users/top", periodDexArrayRequest[persistence.UserVolume](&dbConfig, func(config *core.DbConfig, period models.Period, dex models.Dex) string {
		return persistence.TopUsersRequest(config, period, dex)
	}))
	route.GET("/api/referrers/top", periodDexArrayRequest[persistence.UserVolume](&dbConfig, func(config *core.DbConfig, period models.Period, dex models.Dex) string {
		return persistence.TopReferrersRequest(config, period, dex)
	}))
	route.GET("/api/profiters/top", periodDexArrayRequest[persistence.UserVolume](&dbConfig, func(config *core.DbConfig, period models.Period, dex models.Dex) string {
		//Deprecated
		return persistence.TopUsersProfiters(config, period)
	}))
	route.GET("/api/swaps/distribution", oneRowPeriodDexRequest[persistence.SwapDistribution](&dbConfig, func(cfg *core.DbConfig, period models.Period, dex models.Dex) string {
		return persistence.SwapsDistributionSqlQuery(cfg, period, dex)
	}))

	route.GET("/api/arbitrages/latest", latestArbitrages(&dbConfig))
	route.GET("/api/arbitrages/top", periodDexArrayRequest[persistence.EnrichedArbitrageCH](&dbConfig, func(config *core.DbConfig, period models.Period, dex models.Dex) string {
		return persistence.TopArbitragesSqlQuery(config, period)
	}))
	route.GET("/api/arbitrages/volumeHistory", periodDexArrayRequest[persistence.ArbitrageHistoryEntry](&dbConfig, func(config *core.DbConfig, period models.Period, _ models.Dex) string {
		return persistence.ArbitrageHistorySqlQuery(config, period)
	}))
	route.GET("/api/arbitrages/distribution", oneRowPeriodDexRequest[persistence.ArbitrageDistribution](&dbConfig, func(cfg *core.DbConfig, period models.Period, _ models.Dex) string {
		return persistence.ArbitrageDistributionSqlQuery(cfg, period)
	}))
	route.GET("/api/arbitrages/users/top", periodDexArrayRequest[persistence.TopArbitrageUser](&dbConfig, func(cfg *core.DbConfig, period models.Period, dex models.Dex) string {
		return persistence.TopArbitrageUsersSql(cfg, period)
	}))
	route.GET("/api/arbitrages/jettons/top", periodDexArrayRequest[persistence.TopArbitrageJetton](&dbConfig, func(cfg *core.DbConfig, period models.Period, dex models.Dex) string {
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

func periodDexArrayRequest[T any](cfg *core.DbConfig, sqlFunc func(cfg *core.DbConfig, period models.Period, dex models.Dex) string) func(c *gin.Context) {
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

func latestSwaps(cfg *core.DbConfig) func(c *gin.Context) {
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

func latestArbitrages(cfg *core.DbConfig) func(c *gin.Context) {
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

func oneRowPeriodDexRequest[T any](cfg *core.DbConfig, sqlFunc func(cfg *core.DbConfig, period models.Period, dex models.Dex) string) func(c *gin.Context) {
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
