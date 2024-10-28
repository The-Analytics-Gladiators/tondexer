package persistence

import (
	"TonArb/models"
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"log"
)

func connection() (driver.Conn, error) {
	return clickhouse.Open(&clickhouse.Options{
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
}

func WriteToClickhouse[T any](entities []*T, table string, batchFunc func(driver.Batch, *T) error) error {
	conn, err := connection()

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

	batch, err := conn.PrepareBatch(context.Background(), "INSERT INTO default."+table)
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

func ReadClickhouseJettons() ([]models.ClickhouseJetton, error) {
	conn, err := connection()
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	var result []models.ClickhouseJetton

	if err = conn.Select(context.Background(), &result, `
		SELECT name, symbol, master, decimals FROM clickhouse_jetton`); err != nil {
		return nil, err
	}

	return result, nil
}

func ReadWalletMasters() ([]models.WalletJetton, error) {
	conn, err := connection()
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	var result []models.WalletJetton

	if err = conn.Select(context.Background(), &result, `
		SELECT wallet, master FROM wallet_to_master`); err != nil {
		return nil, err
	}

	return result, nil
}

func ReadStonfiRouterWallets() ([]string, error) {
	conn, err := connection()
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	var result []struct {
		Token string `ch:"token"`
	}

	if err = conn.Select(context.Background(), &result, `
		SELECT DISTINCT tokens AS token
		FROM
		(
			SELECT [token_in, token_out] AS tokens
			FROM swaps
		)
		ARRAY JOIN tokens`); err != nil {
		return nil, err
	}

	var strings []string

	for _, row := range result {
		strings = append(strings, row.Token)
	}

	return strings, nil
}

func SaveSwapsToClickhouse(modelsBatch []*models.SwapCH) error {
	conn, err := connection()

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

	batch, err := conn.PrepareBatch(context.Background(), "INSERT INTO default.swaps")
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
