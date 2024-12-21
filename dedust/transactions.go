package dedust

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tonkeeper/tonapi-go"
	"github.com/xssnick/tonutils-go/address"
	"log"
	"math/big"
	"time"
	"tondexer/common"
	"tondexer/core"
	"tondexer/models"
)

type OptTrace struct {
	Trace *tonapi.Trace
	Set   bool
}

const dedustSwapInternalOpCode = "0x61ee542d"
const dedustSwapOpCode = "0xea06185d"
const dedustPayoutFromPoolOpCode = "0xad4eb6f5"
const dedustPayoutOpCode = "0x474f86cf"
const jettonTransferOpCode = "0x0f8a7ea5"
const dedustSwapPeerOpCode = "0x72aca8aa"

type DedustSwapTraces struct {
	Root                   *tonapi.Trace
	OptInVaultWalletTrace  OptTrace
	InVaultTrace           *tonapi.Trace
	PoolTraces             []*tonapi.Trace
	OutVaultTrace          *tonapi.Trace
	OptOutVaultWalletTrace OptTrace
}

func ExtractDedustSwapsFromRootTrace(root *tonapi.Trace) []*models.DedustSwapInfo {
	infos := common.Map(findSwapTraces(root), func(t *DedustSwapTraces) *models.DedustSwapInfo {
		info, e := dedustSwapInfoFromDedustTraces(t)
		if e != nil {
			log.Printf("Error extracting dedust Swap Info from %v: %v", t.InVaultTrace.Transaction.Hash, e)
			return nil
		}
		return info
	})
	return common.Filter(infos, func(info *models.DedustSwapInfo) bool { return info != nil })
}

func vaultInMessageProperOpCode(inMsg tonapi.OptMessage) bool {
	return inMsg.IsSet() &&
		inMsg.Value.OpCode.IsSet() &&
		(inMsg.Value.OpCode.Value == dedustSwapOpCode || inMsg.Value.OpCode.Value == core.JettonNotifyOpCode)
}

func poolChildIsVaultAndHasProperCode(child tonapi.Trace) bool {
	return common.Contains(child.Interfaces, "dedust_vault") &&
		child.Transaction.InMsg.Set &&
		child.Transaction.InMsg.Value.OpCode.IsSet() &&
		child.Transaction.InMsg.Value.OpCode.Value == dedustPayoutFromPoolOpCode
}

func GetAfterInVaultPoolChain(inVaultTrace *tonapi.Trace) []*tonapi.Trace {
	currentTrace := inVaultTrace
	var result []*tonapi.Trace

	i := 1
	for len(currentTrace.Children) == 1 &&
		common.Contains(currentTrace.Children[0].Interfaces, "dedust_pool") &&
		currentTrace.Children[0].Transaction.InMsg.Set &&
		currentTrace.Children[0].Transaction.InMsg.Value.OpCode.IsSet() &&
		(i == 1 && currentTrace.Children[0].Transaction.InMsg.Value.OpCode.Value == dedustSwapInternalOpCode ||
			i > 1 && currentTrace.Children[0].Transaction.InMsg.Value.OpCode.Value == dedustSwapPeerOpCode) {

		result = append(result, &currentTrace.Children[0])
		currentTrace = &currentTrace.Children[0]
		i++
	}

	return result
}

func findSwapTraces(root *tonapi.Trace) []*DedustSwapTraces {
	var traverse func(trace *tonapi.Trace, previousTrace *tonapi.Trace)

	var result []*DedustSwapTraces

	traverse = func(trace *tonapi.Trace, previousTrace *tonapi.Trace) {
		inMsg := trace.Transaction.InMsg
		if common.Contains(trace.Interfaces, "dedust_vault") &&
			vaultInMessageProperOpCode(inMsg) {

			poolChain := GetAfterInVaultPoolChain(trace)
			if len(poolChain) == 0 {
				for _, child := range trace.Children {
					traverse(&child, trace)
				}
			} else {
				lastPoolTrace := poolChain[len(poolChain)-1]
				if len(lastPoolTrace.Children) == 1 &&
					poolChildIsVaultAndHasProperCode(lastPoolTrace.Children[0]) {

					swapTraces := &DedustSwapTraces{
						Root:         root,
						InVaultTrace: trace,
						PoolTraces:   poolChain,
						//OutVaultTrace: &trace.Children[0].Children[0],
					}

					if inMsg.Value.OpCode.Value == dedustSwapOpCode {
						// means that input token = TON
						swapTraces.OptInVaultWalletTrace = OptTrace{Set: false}
					}
					if inMsg.Value.OpCode.Value == core.JettonNotifyOpCode {
						// means the input token != TON
						if previousTrace != nil {
							swapTraces.OptInVaultWalletTrace = OptTrace{Set: true, Trace: previousTrace}
						}
					}
					outVaultTrace := lastPoolTrace.Children[0]
					swapTraces.OutVaultTrace = &outVaultTrace
					if len(swapTraces.OutVaultTrace.Children) == 1 &&
						swapTraces.OutVaultTrace.Children[0].Transaction.InMsg.IsSet() &&
						swapTraces.OutVaultTrace.Children[0].Transaction.InMsg.Value.OpCode.IsSet() &&
						swapTraces.OutVaultTrace.Children[0].Transaction.InMsg.Value.OpCode.Value == jettonTransferOpCode {
						//means output token != TON
						swapTraces.OptOutVaultWalletTrace = OptTrace{Set: true, Trace: &swapTraces.OutVaultTrace.Children[0]}
					} else {
						swapTraces.OptOutVaultWalletTrace = OptTrace{Set: false}
					}
					result = append(result, swapTraces)
					for _, child := range swapTraces.OutVaultTrace.Children {
						traverse(&child, lastPoolTrace)
					}
				} else {
					for _, child := range lastPoolTrace.Children {
						traverse(&child, lastPoolTrace)
					}
				}
			}
		} else {
			for _, child := range trace.Children {
				traverse(&child, trace)
			}
		}
	}

	traverse(root, nil)
	return result
}

func swapPoolsInfoFromSwapTraces(swapTraces *DedustSwapTraces) ([]*models.SwapPoolInfo, error) {
	var e error
	poolInfos := common.Map(swapTraces.PoolTraces, func(poolTrace *tonapi.Trace) *models.SwapPoolInfo {
		if poolTrace.Transaction.ComputePhase.Set &&
			poolTrace.Transaction.ComputePhase.Value.ExitCode.IsSet() &&
			poolTrace.Transaction.ComputePhase.Value.ExitCode.Value != 0 {
			return nil
		}

		var poolJson PoolJsonBody
		if err := json.Unmarshal(poolTrace.Transaction.InMsg.Value.DecodedBody, &poolJson); err != nil {
			e = err
		}
		poolAddress, err := address.ParseRawAddr(poolTrace.Transaction.Account.Address)
		if err != nil {
			e = err
		}
		var jettonIn *address.Address
		if poolJson.Asset != nil {
			jettonIn, err = address.ParseRawAddr(fmt.Sprintf("%v:%v", poolJson.Asset.Jetton.WorkchainId, poolJson.Asset.Jetton.Address))
			e = err
		}
		amount, p := new(big.Int).SetString(poolJson.Amount, 10)
		if !p {
			e = errors.New("invalid amount")
		}

		limit, _ := new(big.Int).SetString(poolJson.Current.Limit, 10)

		sender, err := address.ParseRawAddr(poolJson.SenderAddr)
		sender.SetBounce(false)
		if err != nil {
			e = err
		}

		return &models.SwapPoolInfo{
			Hash:     poolTrace.Transaction.Hash,
			Lt:       uint64(poolTrace.Transaction.Lt),
			Address:  poolAddress,
			Sender:   sender,
			JettonIn: jettonIn,
			AmountIn: amount,
			Limit:    limit,
		}
	})
	if e != nil {
		return nil, e
	}
	return common.FilterNonNill(poolInfos), nil
}

func dedustSwapInfoFromDedustTraces(swapTraces *DedustSwapTraces) (*models.DedustSwapInfo, error) {
	poolInfos, e := swapPoolsInfoFromSwapTraces(swapTraces)
	if e != nil {
		return nil, e
	}

	if len(poolInfos) == 0 {
		return nil, errors.New("no dedust swap pools")
	}

	sender := poolInfos[0].Sender

	var inWalletAddress *address.Address
	if swapTraces.OptInVaultWalletTrace.Set {
		inWalletAddress, e = address.ParseRawAddr(swapTraces.OptInVaultWalletTrace.Trace.Transaction.Account.Address)
		if e != nil {
			return nil, e
		}
	}

	var outWalletAddress *address.Address
	if swapTraces.OptOutVaultWalletTrace.Set {
		outWalletAddress, e = address.ParseRawAddr(swapTraces.OptOutVaultWalletTrace.Trace.Transaction.Account.Address)
		if e != nil {
			return nil, e
		}
	}
	var lt uint64
	var t time.Time
	if swapTraces.OptInVaultWalletTrace.Set {
		lt = uint64(swapTraces.OptInVaultWalletTrace.Trace.Transaction.Lt)
		t = time.UnixMilli(swapTraces.OptInVaultWalletTrace.Trace.Transaction.Utime * 1000)
	} else {
		lt = uint64(swapTraces.InVaultTrace.Transaction.Lt)
		t = time.UnixMilli(swapTraces.InVaultTrace.Transaction.Utime * 1000)
	}

	var outVaultJson OutVaultJsonBody
	if err := json.Unmarshal(swapTraces.OutVaultTrace.Transaction.InMsg.Value.DecodedBody, &outVaultJson); err != nil {
		return nil, err
	}
	amountOut, p := new(big.Int).SetString(outVaultJson.Amount, 10)
	if !p {
		return nil, errors.New("invalid amount out")
	}

	return &models.DedustSwapInfo{
		TraceID:          swapTraces.Root.Transaction.Hash,
		Lt:               lt,
		Time:             t,
		Sender:           sender,
		InWalletAddress:  inWalletAddress,
		PoolsInfo:        poolInfos,
		OutWalletAddress: outWalletAddress,
		OutAmount:        amountOut,
		CatchTime:        time.Now(),
	}, nil
}
