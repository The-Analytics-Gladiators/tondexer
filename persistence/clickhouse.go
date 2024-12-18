package persistence

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"log"
	"tondexer/core"
	"tondexer/models"
)

func connection(config *core.DbConfig) (driver.Conn, error) {
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

func WriteToClickhouse[T any](config *core.DbConfig, entities []*T, table string, batchFunc func(driver.Batch, *T) error) error {
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
			} else {
				log.Printf("Batch of %v entities has been written to %v\n", len(entities), table)
			}
		}
		return nil
	}
}

func WriteArbitragesToClickhouse(config *core.DbConfig, arbitrages []*models.ArbitrageCH) error {
	return WriteToClickhouse(config, arbitrages, "arbitrages", func(batch driver.Batch, model *models.ArbitrageCH) error {
		return batch.Append(
			model.Sender,
			model.Time,

			model.AmountIn,
			model.AmountOut,
			model.Jetton,
			model.JettonName,
			model.JettonSymbol,
			model.JettonUsdRate,
			model.JettonDecimals,

			model.AmountsPath,
			model.JettonsPath,
			model.JettonNames,
			model.JettonSymbols,
			model.JettonUsdRates,
			model.JettonsDecimals,

			model.PoolsPath,
			model.TraceIDs,
			model.Dexes,
			model.Senders,
		)
	})
}

func ReadClickhouseJettons(config *core.DbConfig) ([]models.ClickhouseJetton, error) {
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

func ReadWalletMasters(config *core.DbConfig) ([]models.WalletJetton, error) {
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

func ReadSingleRow[T any](config *core.DbConfig, sql string) (*T, error) {
	conn, err := connection(config)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	row := conn.QueryRow(context.Background(), sql)

	var res T
	err = row.ScanStruct(&res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func SaveSwapsToClickhouse(config *core.DbConfig, modelsBatch []*models.SwapCH) error {
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
				model.PoolAddress,
				model.Sender,
				model.ReferralAddress,
				model.ReferralAmount,
				model.CatchTime,
				model.TraceID,
			)

			if e != nil {
				log.Printf("Unable to add batch %v \n", e)
				return err
			}

		}

		if len(modelsBatch) != 0 {
			e := batch.Send()
			log.Printf("Batch of %v swaps has been written \n", len(modelsBatch))
			if e != nil {
				log.Printf("Clickhouse insert issue: %v \n", e)
				return err
			}
		}
		return nil
	}
}

func ReadArrayFromClickhouse[T any](config *core.DbConfig, query string) ([]T, error) {
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
