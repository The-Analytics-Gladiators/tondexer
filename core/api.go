package core

import (
	"context"
	"errors"
	"github.com/sethvargo/go-retry"
	"github.com/tonkeeper/tonapi-go"
	"log"
	"strconv"
	"time"
	"tondexer/models"
)

type TonConsoleApi struct {
	*tonapi.Client
}

func (api *TonConsoleApi) GetTraceByHash(hash string) (*tonapi.Trace, error) {
	backoff := retry.WithMaxRetries(4, retry.NewExponential(1*time.Second))
	params := tonapi.GetTraceParams{TraceID: hash}
	return retry.DoValue(context.Background(), backoff, func(ctx context.Context) (*tonapi.Trace, error) {
		internal, err := api.GetTrace(context.Background(), params)
		return internal, retry.RetryableError(err)
	})
}

func (api *TonConsoleApi) JettonInfoByMaster(master string) (*models.ChainTokenInfo, error) {
	backoff := retry.WithMaxRetries(4, retry.NewExponential(1*time.Second))
	return retry.DoValue(context.Background(), backoff, func(ctx context.Context) (*models.ChainTokenInfo, error) {
		internal, err := api.jettonInfoByMasterInternal(master)
		return internal, retry.RetryableError(err)
	})
}

func (api *TonConsoleApi) jettonInfoByMasterInternal(master string) (*models.ChainTokenInfo, error) {
	params := tonapi.GetJettonInfoParams{
		AccountID: master,
	}
	jettonInfo, e := api.GetJettonInfo(context.Background(), params)
	if e != nil {
		return nil, e
	}

	decimals, e := strconv.ParseUint(jettonInfo.Metadata.Decimals, 10, 64)
	if e != nil {
		decimals = 9
		log.Printf("Error parsing decimals for %v: %v \n", master, e)
	}

	return &models.ChainTokenInfo{
		Name:          jettonInfo.Metadata.Name,
		Symbol:        jettonInfo.Metadata.Symbol,
		Decimals:      decimals,
		JettonAddress: master,
	}, nil
}

func (api *TonConsoleApi) JettonRateToUsdByMaster(master string) (float64, error) {
	backoff := retry.WithMaxRetries(4, retry.NewExponential(1*time.Second))
	return retry.DoValue(context.Background(), backoff, func(ctx context.Context) (float64, error) {
		internal, err := api.jettonRateToUsdByMasterInternal(master)
		return internal, retry.RetryableError(err)
	})
}

func (api *TonConsoleApi) jettonRateToUsdByMasterInternal(master string) (float64, error) {
	params := tonapi.GetRatesParams{
		Tokens:     []string{master},
		Currencies: []string{"usd"},
	}

	rates, e := api.GetRates(context.Background(), params)
	if e != nil {
		return 0, e
	}
	if tokenRates, exists := rates.Rates[master]; exists {
		if rate, exists2 := tokenRates.Prices.Value["USD"]; exists2 {
			return rate, nil
		}
	}

	return 0, errors.New("no usd rate for master " + master)
}
