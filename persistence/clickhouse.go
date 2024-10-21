package persistence

import (
	"TonArb/models"
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"log"
)

func SaveToClickhouse(modelsBatch []*models.SwapCH) {

	conn, _ := clickhouse.Open(&clickhouse.Options{
		Addr:         []string{fmt.Sprintf("%s:%d", "localhost", 9000)},
		Protocol:     clickhouse.Native,
		MaxOpenConns: 5,
		MaxIdleConns: 5,
		Auth:         clickhouse.Auth{
			//Database: env.Database,
			//Username: env.Username,
			//Password: env.Password,
		},
	})

	defer func(conn driver.Conn) {
		err := conn.Close()
		if err != nil {
			log.Printf("Close connection issue: %v \n", err)
		}
	}(conn)

	batch, err := conn.PrepareBatch(context.Background(), "INSERT INTO default.swaps")
	if err != nil {
		fmt.Printf("Unable to create batch  %v \n", err)
	} else {
		for _, model := range modelsBatch {
			log.Printf("CH Model %v \n", model)
			e := batch.Append(
				model.Hashes,
				model.Lt,
				model.Time,
				model.TokenIn,
				model.AmountIn,
				model.TokenOut,
				model.AmountOut,
				model.MinAmountOut,
				model.Sender,
				model.ReferralAddress,
				model.ReferralAmount,
			)

			if e != nil {
				log.Printf("Unable to add batch %v \n", e)
			}

		}

		if len(modelsBatch) != 0 {
			e := batch.Send()
			if e != nil {
				log.Printf("Clickhouse insert issue: %v \n", e)
			} else {
				modelsBatch = []*models.SwapCH{}
			}
		}
	}
}
