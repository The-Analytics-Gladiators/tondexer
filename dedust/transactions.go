package dedust

import (
	"encoding/json"
	"github.com/tonkeeper/tonapi-go"
	"github.com/xssnick/tonutils-go/address"
	"log"
	"strconv"
	"time"
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

var stonfiPtonWallet = address.MustParseAddr("EQARULUYsmJq1RiZ-YiH-IJLcAZUVkVff-KBPwEmmaQGH6aC")

type DedustSwapTraces struct {
	OptInVaultWalletTrace  OptTrace
	InVaultTrace           *tonapi.Trace
	PoolTrace              *tonapi.Trace
	OutVaultTrace          *tonapi.Trace
	OptOutVaultWalletTrace OptTrace
}

func findSwapTraces(root *tonapi.Trace) []*DedustSwapTraces {

	var traverse func(trace *tonapi.Trace, previousTrace *tonapi.Trace)

	var result []*DedustSwapTraces

	traverse = func(trace *tonapi.Trace, previousTrace *tonapi.Trace) {
		inMsg := trace.Transaction.InMsg
		if core.Contains(trace.Interfaces, "dedust_vault") &&
			inMsg.IsSet() &&
			inMsg.Value.OpCode.IsSet() &&
			(inMsg.Value.OpCode.Value == dedustSwapOpCode || inMsg.Value.OpCode.Value == core.JettonNotifyOpCode) &&
			len(trace.Children) == 1 &&
			core.Contains(trace.Children[0].Interfaces, "dedust_pool") &&
			trace.Children[0].Transaction.InMsg.Set &&
			trace.Children[0].Transaction.InMsg.Value.OpCode.IsSet() &&
			trace.Children[0].Transaction.InMsg.Value.OpCode.Value == dedustSwapInternalOpCode &&
			len(trace.Children[0].Children) == 1 &&
			core.Contains(trace.Children[0].Children[0].Interfaces, "dedust_vault") &&
			trace.Children[0].Children[0].Transaction.InMsg.Set &&
			trace.Children[0].Children[0].Transaction.InMsg.Value.OpCode.IsSet() &&
			trace.Children[0].Children[0].Transaction.InMsg.Value.OpCode.Value == dedustPayoutFromPoolOpCode &&
			// Failed payout are through the vault
			trace.Children[0].Children[0].Transaction.Account.Address != trace.Transaction.Account.Address {

			swapTraces := &DedustSwapTraces{
				InVaultTrace:  trace,
				PoolTrace:     &trace.Children[0],
				OutVaultTrace: &trace.Children[0].Children[0],
			}
			if inMsg.Value.OpCode.Value == dedustSwapOpCode {
				// means that input token = TON
				// => output Token != TON
				if len(trace.Children[0].Children[0].Children) == 1 {
					swapTraces.OptOutVaultWalletTrace = OptTrace{Set: true, Trace: &trace.Children[0].Children[0].Children[0]}
					swapTraces.OptInVaultWalletTrace = OptTrace{Set: false}
				}
			}
			if inMsg.Value.OpCode.Value == core.JettonNotifyOpCode {
				// means the input token != TON
				if previousTrace != nil {
					swapTraces.OptInVaultWalletTrace = OptTrace{Set: true, Trace: previousTrace}
					//swapTraces.OptOutVaultWalletTrace = OptTrace{Set: false}
				}
			}
			if len(swapTraces.OutVaultTrace.Children) == 1 &&
				swapTraces.OutVaultTrace.Children[0].Transaction.InMsg.IsSet() &&
				swapTraces.OutVaultTrace.Children[0].Transaction.InMsg.Value.OpCode.IsSet() &&
				swapTraces.OutVaultTrace.Children[0].Transaction.InMsg.Value.OpCode.Value == jettonTransferOpCode {
				//means output token != TON
				swapTraces.OptOutVaultWalletTrace = OptTrace{Set: true, Trace: &swapTraces.OutVaultTrace.Children[0]}
			}

			result = append(result, swapTraces)
			if swapTraces.OptOutVaultWalletTrace.Set {
				for _, child := range swapTraces.OptOutVaultWalletTrace.Trace.Children {
					traverse(&child, swapTraces.OptOutVaultWalletTrace.Trace)
				}
			} else {
				for _, child := range swapTraces.OutVaultTrace.Children {
					traverse(&child, swapTraces.OutVaultTrace)
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

func swapInfoFromDedustTraces(swapTraces *DedustSwapTraces) (*models.SwapInfo, error) {
	var poolJson PoolJsonBody
	if err := json.Unmarshal(swapTraces.PoolTrace.Transaction.InMsg.Value.DecodedBody, &poolJson); err != nil {
		return nil, err
	}

	swapTransferNotification, err := notificationFromSwapTraces(swapTraces, poolJson)
	if err != nil {
		return nil, err
	}

	payment, err := paymentFromSwapTraces(swapTraces)
	if err != nil {
		return nil, err
	}

	// I don't know where the referral amount is..
	var referral *models.PayoutRequest
	if swapTransferNotification.ReferralAddress != nil {
		copyPayment := *payment
		copyPayment.Amount1Out = 0
		referral = &copyPayment
	}

	return &models.SwapInfo{
		Notification: swapTransferNotification,
		Payment:      payment,
		Referral:     referral,
		PoolAddress:  swapTraces.PoolTrace.Transaction.Account.Address,
	}, nil
}

func notificationFromSwapTraces(swapTraces *DedustSwapTraces, poolJson PoolJsonBody) (*models.SwapTransferNotification, error) {

	var queryId uint64
	var amount uint64
	var minOut uint64
	var referralAddress *address.Address

	if !swapTraces.OptInVaultWalletTrace.Set {
		// source token is TON
		var inVaultJson InVaultJsonBodyForTon
		if err := json.Unmarshal(swapTraces.InVaultTrace.Transaction.InMsg.Value.DecodedBody, &inVaultJson); err != nil {
			return nil, err
		}
		queryId = inVaultJson.QueryID

		var err error
		amount, err = strconv.ParseUint(inVaultJson.Amount, 10, 64)
		if err != nil {
			log.Printf("error parsing amount for inVault %v: %v\n", swapTraces.InVaultTrace.Transaction.Hash, err)
		}
		minOut, err = strconv.ParseUint(inVaultJson.Step.Params.Limit, 10, 64)
		if err != nil {
			log.Printf("error parsing limit for inVault %v: %v\n", swapTraces.InVaultTrace.Transaction.Hash, err)
		}
		if inVaultJson.SwapParams.ReferralAddr != "" {
			referralAddress = address.MustParseRawAddr(inVaultJson.SwapParams.ReferralAddr)
		}
	} else {
		// source token is not TON
		var inVaultJson InVaultBodyForToken
		if err := json.Unmarshal(swapTraces.InVaultTrace.Transaction.InMsg.Value.DecodedBody, &inVaultJson); err != nil {
			return nil, err
		}
		queryId = inVaultJson.QueryID

		var err error
		amount, err = strconv.ParseUint(inVaultJson.Amount, 10, 64)
		if err != nil {
			log.Printf("error parsing amount for inVault %v: %v\n", swapTraces.InVaultTrace.Transaction.Hash, err)
		}

		minOut, err = strconv.ParseUint(inVaultJson.ForwardPayload.Value.Value.Step.Params.Limit, 10, 64)
		if err != nil {
			log.Printf("error parsing limit for inVault %v: %v\n", swapTraces.InVaultTrace.Transaction.Hash, err)
		}

		if inVaultJson.ForwardPayload.Value.Value.SwapParams.ReferralAddr != "" {
			referralAddress = address.MustParseRawAddr(inVaultJson.ForwardPayload.Value.Value.SwapParams.ReferralAddr)
		}
	}

	var sender *address.Address
	if poolJson.SenderAddr != "" {
		sender = address.MustParseRawAddr(poolJson.SenderAddr)
	}

	return &models.SwapTransferNotification{
		Hash:            swapTraces.InVaultTrace.Transaction.Hash,
		Lt:              uint64(swapTraces.InVaultTrace.Transaction.Lt),
		TransactionTime: time.UnixMilli(swapTraces.InVaultTrace.Transaction.Utime * 1000),
		EventCatchTime:  time.Now(),
		QueryId:         queryId,
		Amount:          amount,
		Sender:          sender,
		TokenWallet:     nil, // unused in fact
		MinOut:          minOut,
		ToAddress:       nil,
		ReferralAddress: referralAddress,
	}, nil
}

func paymentFromSwapTraces(swapTraces *DedustSwapTraces) (*models.PayoutRequest, error) {

	var outVaultJson OutVaultJsonBody
	if err := json.Unmarshal(swapTraces.OutVaultTrace.Transaction.InMsg.Value.DecodedBody, &outVaultJson); err != nil {
		return nil, err
	}
	var tokenInWalletAddress *address.Address
	if swapTraces.OptInVaultWalletTrace.Set {
		tokenInWalletAddress = address.MustParseRawAddr(swapTraces.OptInVaultWalletTrace.Trace.Transaction.Account.Address)
	} else {
		// otherwise it is TON. Taking stonfi pTON wallet just for reference
		tokenInWalletAddress = stonfiPtonWallet
	}
	var tokenOutWalletAddress *address.Address
	if swapTraces.OptOutVaultWalletTrace.Set {
		tokenOutWalletAddress = address.MustParseRawAddr(swapTraces.OptOutVaultWalletTrace.Trace.Transaction.Account.Address)
	} else {
		tokenOutWalletAddress = stonfiPtonWallet
	}

	amountOut, err := strconv.ParseUint(outVaultJson.Amount, 10, 64)
	if err != nil {
		log.Printf("Unable to parse amount '%v' from outVault for Hash %v: %v \n",
			outVaultJson.Amount, swapTraces.OutVaultTrace.Transaction.Hash, err)
	}

	return &models.PayoutRequest{
		Hash:                swapTraces.OutVaultTrace.Transaction.Hash,
		Lt:                  uint64(swapTraces.OutVaultTrace.Transaction.Lt),
		TransactionTime:     time.UnixMilli(swapTraces.OutVaultTrace.Transaction.Utime * 1000),
		EventCatchTime:      time.Now(),
		QueryId:             outVaultJson.QueryID,
		Owner:               nil, // Used only for referrals
		ExitCode:            0,   // No exit code
		Amount0Out:          0,   // Here we treat token0 as alsways IN token
		Token0WalletAddress: tokenInWalletAddress,
		Amount1Out:          amountOut,
		Token1WalletAddress: tokenOutWalletAddress,
	}, nil
}