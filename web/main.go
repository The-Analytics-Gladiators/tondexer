package main

import (
	"TonArb/persistence"
	"encoding/json"
	"github.com/gin-gonic/gin"
)

func main() {
	route := gin.Default()

	route.GET("/swaps/latest", func(c *gin.Context) {
		var request struct {
			Limit uint64 `form:"limit"`
		}
		if err := c.ShouldBindQuery(&request); err != nil {
			c.JSON(400, gin.H{"msg": err.Error()})
			return
		}

		models, e := persistence.ReadLastSwaps(request.Limit)
		if e != nil {
			c.JSON(200, gin.H{"msg": e.Error()})
			return
		}

		jsonData, err := json.Marshal(models)
		if err != nil {
			c.JSON(200, gin.H{"msg": e.Error()})
			return
		}
		c.JSON(200, string(jsonData))
	})
	route.Run(":8088")
}
