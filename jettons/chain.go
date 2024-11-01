package jettons

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"github.com/sethvargo/go-retry"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
	cell2 "github.com/xssnick/tonutils-go/tvm/cell"
	"strconv"
	"time"
	"tondexer/core"
)

type ChainTokenInfo struct {
	Name          string
	Symbol        string
	Decimals      uint
	JettonAddress string
}

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

func dictKey(key string) *cell2.Cell {
	bytes := sha256.Sum256([]byte(key))
	return cell2.BeginCell().MustStoreSlice(bytes[:], uint(len(bytes)*8)).EndCell()
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

func JettonDefinitionByWalletRetry(walletAddr string, retries uint64) (*ChainTokenInfo, error) {
	backoff := retry.WithMaxRetries(retries, retry.NewFibonacci(1*time.Second))
	return retry.DoValue(context.Background(), backoff, func(ctx context.Context) (*ChainTokenInfo, error) {
		wallet, err := JettonDefinitionByWallet(walletAddr)
		return wallet, retry.RetryableError(err)
	})
}

func TokenDefinitionByMasterRetries(masterAddr string, retries uint64) (*ChainTokenInfo, error) {
	backoff := retry.WithMaxRetries(retries, retry.NewFibonacci(1*time.Second))
	return retry.DoValue(context.Background(), backoff, func(ctx context.Context) (*ChainTokenInfo, error) {
		jettonInfo, err := TokenDefinitionByMaster(masterAddr)
		return jettonInfo, retry.RetryableError(err)
	})
}

func TokenDefinitionByMaster(masterAddr string) (*ChainTokenInfo, error) {
	wApi, err := GetTonApi()

	block, err := (*wApi.Api).CurrentMasterchainInfo(context.Background())
	if err != nil {
		return nil, err
	}
	resp, err := wApi.RunGetMethodRetries(context.Background(), block, address.MustParseAddr(masterAddr), "get_jetton_data", 3)
	if err != nil {
		return nil, err
	}

	cell := resp.MustCell(3)
	cs := cell.BeginParse()

	result := &ChainTokenInfo{}
	result.Decimals = 9 // Default
	result.JettonAddress = masterAddr
	uri := ""
	if cs.MustLoadUInt(8) != 0 {
		//OFFCHAIN
		if binary, er := cs.LoadBinarySnake(); er == nil {
			uri = string(binary)
		}
	} else {
		dct := cs.MustLoadDict(256)

		decimalsKey := dictKey("decimals")
		nameKey := dictKey("name")
		symbolKey := dictKey("symbol")

		if decimalsValue, e := dct.LoadValue(decimalsKey); e == nil {
			if decimalBytes, er := decimalsValue.LoadBinarySnake(); er == nil {
				if decimals, er2 := strconv.Atoi(string(decimalBytes[1:])); er2 == nil {
					result.Decimals = uint(decimals)
				}
			}
		}

		if nameValue, e := dct.LoadValue(nameKey); e == nil {
			if nameBytes, er := nameValue.LoadBinarySnake(); er == nil {
				result.Name = string(nameBytes[1:])
			}
		}

		if symbolValue, e := dct.LoadValue(symbolKey); e == nil {
			if symbolBytes, er := symbolValue.LoadBinarySnake(); er == nil {
				result.Symbol = string(symbolBytes[1:])
			}
		}

		if uriCell, e := dct.LoadValue(dictKey("uri")); e == nil {
			if uriBytes, er := uriCell.LoadBinarySnake(); er == nil {
				uri = string(uriBytes[1:])
			}
		}
	}

	if uri != "" {
		body, e2 := core.GetRetry(context.Background(), uri, 3)
		if e2 == nil {
			var jsonMap map[string]any

			e4 := json.Unmarshal(body, &jsonMap)
			if e4 == nil {
				if name, exists := jsonMap["name"]; exists {
					result.Name = name.(string)
				}
				if symbol, exists := jsonMap["symbol"]; exists {
					result.Symbol = symbol.(string)
				}
				if decimalsString, exists := jsonMap["decimals"]; exists {
					switch decimalsString.(type) {
					case string:
						if decimals, e5 := strconv.Atoi(decimalsString.(string)); e5 == nil {
							result.Decimals = uint(decimals)
						}
					case float64:
						result.Decimals = uint(decimalsString.(float64))
					}
				}
			}
		}
	}

	return result, nil
}

func JettonDefinitionByWallet(walletAddr string) (*ChainTokenInfo, error) {
	addr := address.MustParseAddr(walletAddr)

	wApi, err := GetTonApi()
	if err != nil {
		return nil, err
	}

	jettonMasterAddress, err := wApi.MasterByWallet(addr.String())
	if err != nil {
		return nil, err
	}

	return TokenDefinitionByMaster(jettonMasterAddress.String())
}
