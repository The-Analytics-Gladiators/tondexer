package persistence

import (
	"TonArb/core"
	"TonArb/models"
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"log"
)

func connection(config *core.Config) (driver.Conn, error) {
	return clickhouse.Open(&clickhouse.Options{
		Addr:         []string{fmt.Sprintf("%s:%d", config.DbHost, config.DbPort)},
		Protocol:     clickhouse.Native,
		MaxOpenConns: 5,
		MaxIdleConns: 5,
		Auth: clickhouse.Auth{
			Database: config.DbName,
			Username: config.DbUser,
			Password: config.DbPassword,
		},
	})
}

func WriteToClickhouse[T any](config *core.Config, entities []*T, table string, batchFunc func(driver.Batch, *T) error) error {
	conn, err := connection(config)

	if err != nil {
		log.Printf("Open connection issue: %v \n", err)
		return err
	}

	defer func(conn driver.Conn) {
		err := conn.Close()
		if err != nil {
			log.Printf("Close connection issue: %v \n", err)
		}
	}(conn)

	batch, err := conn.PrepareBatch(context.Background(), "INSERT INTO "+config.DbName+"."+table)
	if err != nil {
		fmt.Printf("Unable to create batch  %v \n", err)
		return err
	} else {
		for _, model := range entities {
			log.Printf("CH Model %v \n", model)
			e := batchFunc(batch, model)

			if e != nil {
				log.Printf("Unable to add batch to %v: %v \n", table, e)
				return err
			}

		}

		if len(entities) != 0 {
			e := batch.Send()
			if e != nil {
				log.Printf("Clickhouse insert issue: %v \n", e)
				return err
			}
		}
		return nil
	}
}

func ReadClickhouseJettons(config *core.Config) ([]models.ClickhouseJetton, error) {
	conn, err := connection(config)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	var result []models.ClickhouseJetton

	if err = conn.Select(context.Background(), &result, fmt.Sprintf(`
		SELECT name, symbol, master, decimals FROM %v.clickhouse_jetton`, config.DbName)); err != nil {
		return nil, err
	}

	return result, nil
}

func ReadWalletMasters(config *core.Config) ([]models.WalletJetton, error) {
	conn, err := connection(config)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	var result []models.WalletJetton

	if err = conn.Select(context.Background(), &result, fmt.Sprintf(`
		SELECT wallet, master FROM %v.wallet_to_master`, config.DbName)); err != nil {
		return nil, err
	}

	return result, nil
}

type SummaryStats struct {
	Volume       uint64 `ch:"volume" json:"volume"`
	Number       uint64 `ch:"number" json:"number"`
	UniqueTokens uint64 `ch:"unique_tokens" json:"unique_tokens"`
	UniqueUsers  uint64 `ch:"unique_users" json:"unique_users"`
}

func ReadSummaryStats(config *core.Config, period models.Period) (*SummaryStats, error) {
	conn, err := connection(config)
	if err != nil {
		return nil, err
	}

	defer conn.Close()
	periodParams := models.PeriodParamsMap[period]

	row := conn.QueryRow(context.Background(), fmt.Sprintf(`
SELECT
    toUInt64(sum(%v) + sum(%v)) AS volume,
    count() AS number,
    length(groupUniqArrayArray([jetton_in, jetton_out])) AS unique_tokens,
    uniq(sender) AS unique_users
FROM %v.swaps
WHERE time >= %v(subtractDays(now(), %v))`, UsdInField, UsdOutField,
		config.DbName, periodParams.ToStartOf, periodParams.WindowInDays))

	var stats SummaryStats
	err = row.ScanStruct(&stats)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func SaveSwapsToClickhouse(config *core.Config, modelsBatch []*models.SwapCH) error {
	conn, err := connection(config)

	if err != nil {
		log.Printf("Open connection issue: %v \n", err)
		return err
	}

	defer func(conn driver.Conn) {
		err := conn.Close()
		if err != nil {
			log.Printf("Close connection issue: %v \n", err)
		}
	}(conn)

	batch, err := conn.PrepareBatch(context.Background(), fmt.Sprintf("INSERT INTO %v.swaps", config.DbName))
	if err != nil {
		fmt.Printf("Unable to create batch  %v \n", err)
		return err
	} else {
		for _, model := range modelsBatch {
			log.Printf("CH Model %v \n", model)
			e := batch.Append(
				model.Dex,
				model.Hashes,
				model.Lt,
				model.Time,
				model.JettonIn,
				model.AmountIn,
				model.JettonInSymbol,
				model.JettonInName,
				model.JettonInUsdRate,
				model.JettonInDecimals,
				model.JettonOut,
				model.AmountOut,
				model.JettonOutSymbol,
				model.JettonOutName,
				model.JettonOutUsdRate,
				model.JettonOutDecimals,
				model.MinAmountOut,
				model.Sender,
				model.ReferralAddress,
				model.ReferralAmount,
				model.CatchTime,
			)

			if e != nil {
				log.Printf("Unable to add batch %v \n", e)
				return err
			}

		}

		if len(modelsBatch) != 0 {
			e := batch.Send()
			if e != nil {
				log.Printf("Clickhouse insert issue: %v \n", e)
				return err
			}
		}
		return nil
	}
}

func ReadArrayFromClickhouse[T any](config *core.Config, query string) ([]T, error) {
	conn, err := connection(config)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	rows, err := conn.Query(context.Background(), query)

	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()
	if err != nil {
		return nil, err
	}

	var ts []T
	for rows.Next() {
		var t T
		if e := rows.ScanStruct(&t); e != nil {
			return nil, e
		}
		ts = append(ts, t)
	}

	return ts, nil
}
