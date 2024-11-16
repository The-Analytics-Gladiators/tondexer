package jettons

import (
	"context"
	"github.com/sethvargo/go-retry"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
	"time"
)

type TonApi struct {
	Api *ton.APIClientWrapped
}

func (tonApi *TonApi) RunGetMethodRetries(ctx context.Context,
	block *ton.BlockIDExt,
	address *address.Address,
	method string,
	retries uint64) (*ton.ExecutionResult, error) {

	backoff := retry.WithMaxRetries(retries, retry.NewExponential(1*time.Second))
	return retry.DoValue(ctx, backoff, func(ctx context.Context) (*ton.ExecutionResult, error) {
		return (*tonApi.Api).RunGetMethod(ctx, block, address, method)
	})
}

func GetTonApi() (*TonApi, error) {
	client := liteclient.NewConnectionPool()

	configUrl := "https://ton.org/global.config.json"
	err := client.AddConnectionsFromConfigUrl(context.Background(), configUrl)
	if err != nil {
		return nil, err
	}
	a := ton.NewAPIClient(client)
	api := a.WithRetry()

	wApi := TonApi{Api: &api}
	return &wApi, nil
}

func (tonApi *TonApi) MasterByWallet(wallet string) (*address.Address, error) {
	backoff := retry.WithMaxRetries(5, retry.NewFibonacci(1*time.Second))
	return retry.DoValue(context.Background(), backoff, func(ctx context.Context) (*address.Address, error) {
		result, err := tonApi.masterByWalletInternal(wallet)
		return result, retry.RetryableError(err)
	})
}

func (tonApi *TonApi) masterByWalletInternal(wallet string) (*address.Address, error) {
	wApi, err := GetTonApi()
	if err != nil {
		return nil, err
	}

	block, err := (*wApi.Api).CurrentMasterchainInfo(context.Background())
	if err != nil {
		return nil, err
	}
	res, err := wApi.RunGetMethodRetries(context.Background(), block, address.MustParseAddr(wallet), "get_wallet_data", 4)

	if err != nil {
		return nil, err
	}

	jettonMasterAddress := res.MustSlice(2).MustLoadAddr()

	return jettonMasterAddress, nil
}

func Contract(address string) string {
	// converting stonfi proxy ton v2 to v1
	if address == "EQBnGWMCf3-FZZq1W4IWcWiGAc3PHuZ0_H-7sad2oY00o83S" {
		return "EQCM3B12QK1e4yZSf8GtBRT0aLMNyEsBc_DhVfRRtOEffLez"
	} else {
		return address
	}
}
