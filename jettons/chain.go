package jettons

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
	cell2 "github.com/xssnick/tonutils-go/tvm/cell"
	"io"
	"net/http"
	"strconv"
)

type TokenDefinition struct {
	Name     string
	Symbol   string
	Decimals uint
}

func dictKey(key string) *cell2.Cell {
	bytes := sha256.Sum256([]byte(key))
	return cell2.BeginCell().MustStoreSlice(bytes[:], uint(len(bytes)*8)).EndCell()
}

func TokenDefinitionByWallet(walletAddr string) (*TokenDefinition, error) {
	client := liteclient.NewConnectionPool()

	configUrl := "https://ton.org/global.config.json"
	err := client.AddConnectionsFromConfigUrl(context.Background(), configUrl)
	if err != nil {
		return nil, err
	}
	a := ton.NewAPIClient(client)
	api := a.WithRetry()

	block, err := api.CurrentMasterchainInfo(context.Background())
	if err != nil {
		return nil, err
	}

	addr := address.MustParseAddr(walletAddr)

	res, err := api.RunGetMethod(context.Background(), block, addr, "get_wallet_data")
	if err != nil {
		return nil, err
	}

	jettonMasterAddress := res.MustSlice(2).MustLoadAddr()

	resp, err := api.RunGetMethod(context.Background(), block, jettonMasterAddress, "get_jetton_data")
	if err != nil {
		return nil, err
	}

	cell := resp.MustCell(3)
	cs := cell.BeginParse()

	result := &TokenDefinition{}
	result.Decimals = 9 // Default
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
		if resp, e2 := http.Get(uri); e2 == nil {
			defer resp.Body.Close()
			var jsonMap map[string]any

			if body, e3 := io.ReadAll(resp.Body); e3 == nil {
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
	}

	return result, nil
}
