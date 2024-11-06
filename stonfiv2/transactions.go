package stonfiv2

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

const payoutOpCode = "stonfi_pay_to_v2"
const payoutOpRefCode = "stonfi_pay_vault_v2"

const stonfiRouterV2 = "stonfi_router_v2"

type SwapTraces struct {
	Notification *tonapi.Trace
	Payout       *tonapi.Trace
	VaultPayout  *tonapi.Trace
	Pool         *address.Address
}

func ExtractStonfiV2SwapsFromRootTrace(trace *tonapi.Trace) []*models.SwapInfo {
	infos := core.Map(findSwapTraces(trace), func(t *SwapTraces) *models.SwapInfo {
		return swapTracesToSwapInfo(t)
	})
	return core.Filter(infos, func(info *models.SwapInfo) bool {
		return info != nil
	})
}

type NotificationJsonBody struct {
	QueryID        uint64         `json:"query_id"`
	Amount         string         `json:"amount"`
	Sender         string         `json:"sender"`
	ForwardPayload ForwardPayload `json:"forward_payload"`
}

type ForwardPayload struct {
	IsRight bool         `json:"is_right"`
	Value   PayloadValue `json:"value"`
}

type PayloadValue struct {
	SumType string    `json:"sum_type"`
	OpCode  int       `json:"op_code"`
	Value   SwapValue `json:"value"`
}

type SwapValue struct {
	TokenWallet1    string        `json:"token_wallet1"`
	RefundAddress   string        `json:"refund_address"`
	ExcessesAddress string        `json:"excesses_address"`
	TxDeadline      int64         `json:"tx_deadline"`
	CrossSwapBody   CrossSwapBody `json:"cross_swap_body"`
}

type CrossSwapBody struct {
	MinOut   string `json:"min_out"`
	Receiver string `json:"receiver"`
	FwdGas   string `json:"fwd_gas"`
	//CustomPayload interface{} `json:"custom_payload"`
	RefundFwdGas string `json:"refund_fwd_gas"`
	//RefundPayload interface{} `json:"refund_payload"`
	RefFee     int    `json:"ref_fee"`
	RefAddress string `json:"ref_address"`
}

func parseTraceNotification(notification *tonapi.Trace) (*models.SwapTransferNotification, error) {
	var notificationInfo NotificationJsonBody
	err := json.Unmarshal(notification.Transaction.InMsg.Value.DecodedBody, &notificationInfo)
	if err != nil {
		return nil, err
	}

	amount, e := strconv.ParseUint(notificationInfo.Amount, 10, 64)
	if e != nil {
		log.Printf("error parsing amount for notification %v: %v\n", notification.Transaction.Hash, e)
	}
	minAmount, e := strconv.ParseUint(notificationInfo.ForwardPayload.Value.Value.CrossSwapBody.MinOut, 10, 64)
	if e != nil {
		log.Printf("error parsing minAmount for notification %v: %v\n", notification.Transaction.Hash, e)
	}
	var referralAddress *address.Address
	if notificationInfo.ForwardPayload.Value.Value.CrossSwapBody.RefAddress != "" {
		referralAddress = address.MustParseRawAddr(notificationInfo.ForwardPayload.Value.Value.CrossSwapBody.RefAddress)
	}
	return &models.SwapTransferNotification{
		Hash:            notification.Transaction.Hash,
		Lt:              uint64(notification.Transaction.Lt),
		TransactionTime: time.UnixMilli(notification.Transaction.Utime * 1000),
		EventCatchTime:  time.Now(),
		QueryId:         uint64(notificationInfo.QueryID),
		Amount:          amount,
		Sender:          address.MustParseRawAddr(notificationInfo.Sender),
		TokenWallet:     address.MustParseRawAddr(notificationInfo.ForwardPayload.Value.Value.TokenWallet1),
		MinOut:          minAmount,
		ToAddress:       address.MustParseRawAddr(notificationInfo.ForwardPayload.Value.Value.CrossSwapBody.Receiver),
		ReferralAddress: referralAddress,
	}, nil
}

type PayoutJsonBody struct {
	QueryID         uint64         `json:"query_id"`
	ToAddress       string         `json:"to_address"`
	ExcessesAddress string         `json:"excesses_address"`
	OriginalCaller  string         `json:"original_caller"`
	ExitCode        int64          `json:"exit_code"`
	CustomPayload   interface{}    `json:"custom_payload"`
	AdditionalInfo  AdditionalInfo `json:"additional_info"`
}

type AdditionalInfo struct {
	FwdTonAmount  string `json:"fwd_ton_amount"`
	Amount0Out    string `json:"amount0_out"`
	Token0Address string `json:"token0_address"`
	Amount1Out    string `json:"amount1_out"`
	Token1Address string `json:"token1_address"`
}

func parseTracePayout(payout *tonapi.Trace) (*models.PayoutRequest, error) {
	var payoutJson PayoutJsonBody
	if err := json.Unmarshal(payout.Transaction.InMsg.Value.DecodedBody, &payoutJson); err != nil {
		return nil, err
	}

	amount0Out, e := strconv.ParseUint(payoutJson.AdditionalInfo.Amount0Out, 10, 64)
	if e != nil {
		log.Printf("error parsing amount0Out for payout %v: %v\n", payout.Transaction.Hash, e)
	}
	amount1Out, e := strconv.ParseUint(payoutJson.AdditionalInfo.Amount1Out, 10, 64)
	if e != nil {
		log.Printf("error parsing amount1Out for payout %v: %v\n", payout.Transaction.Hash, e)
	}
	return &models.PayoutRequest{
		Hash:            payout.Transaction.Hash,
		Lt:              uint64(payout.Transaction.Lt),
		TransactionTime: time.UnixMilli(payout.Transaction.Utime * 1000),
		EventCatchTime:  time.Now(),
		QueryId:         uint64(payoutJson.QueryID),
		Owner:           address.MustParseRawAddr(payoutJson.ToAddress),
		ExitCode:        uint64(payoutJson.ExitCode),
		Amount0Out:      amount0Out,
		Token0Address:   address.MustParseRawAddr(payoutJson.AdditionalInfo.Token0Address),
		Amount1Out:      amount1Out,
		Token1Address:   address.MustParseRawAddr(payoutJson.AdditionalInfo.Token1Address),
	}, nil
}

type VaultPayoutJsonBody struct {
	QueryID         uint64                    `json:"query_id"`
	Owner           string                    `json:"owner"`
	ExcessesAddress string                    `json:"excesses_address"`
	AdditionalInfo  VaultPayoutAdditionalInfo `json:"additional_info"`
}

type VaultPayoutAdditionalInfo struct {
	Amount0Out    string `json:"amount0_out"`
	Token0Address string `json:"token0_address"`
	Amount1Out    string `json:"amount1_out"`
	Token1Address string `json:"token1_address"`
}

func parseTraceVaultPayout(vaultPayout *tonapi.Trace) (*models.PayoutRequest, error) {
	var payoutJson VaultPayoutJsonBody
	if err := json.Unmarshal(vaultPayout.Transaction.InMsg.Value.DecodedBody, &payoutJson); err != nil {
		return nil, err
	}
	amount0Out, e := strconv.ParseUint(payoutJson.AdditionalInfo.Amount0Out, 10, 64)
	if e != nil {
		log.Printf("error parsing amount0Out for payout %v: %v\n", vaultPayout.Transaction.Hash, e)
	}
	amount1Out, e := strconv.ParseUint(payoutJson.AdditionalInfo.Amount1Out, 10, 64)
	if e != nil {
		log.Printf("error parsing amount1Out for payout %v: %v\n", vaultPayout.Transaction.Hash, e)
	}

	return &models.PayoutRequest{
		Hash:            vaultPayout.Transaction.Hash,
		Lt:              uint64(vaultPayout.Transaction.Lt),
		TransactionTime: time.UnixMilli(vaultPayout.Transaction.Utime * 1000),
		EventCatchTime:  time.Now(),
		QueryId:         uint64(payoutJson.QueryID),
		Owner:           address.MustParseRawAddr(payoutJson.Owner),
		ExitCode:        0,
		Amount0Out:      amount0Out,
		Token0Address:   address.MustParseRawAddr(payoutJson.AdditionalInfo.Token0Address),
		Amount1Out:      amount1Out,
		Token1Address:   address.MustParseRawAddr(payoutJson.AdditionalInfo.Token1Address),
	}, nil
}

func swapTracesToSwapInfo(swapTraces *SwapTraces) *models.SwapInfo {
	swapTransferNotification, err := parseTraceNotification(swapTraces.Notification)
	if err != nil {
		log.Printf("error parsing notification info for trace %v: %v \n", swapTraces.Notification.Transaction.Hash, err)
		return nil
	}
	payout, err := parseTracePayout(swapTraces.Payout)
	if err != nil {
		log.Printf("error parsing payout for trace %v: %v \n", swapTraces.Payout.Transaction.Hash, err)
		return nil
	}

	if payout.ExitCode != 3326308581 {
		log.Printf("It is not the success payment for StonfiV2 transaction. Code %v, hash %v", payout.ExitCode, payout.Hash)
		return nil
	}

	var referral *models.PayoutRequest
	if swapTraces.VaultPayout != nil {
		referral, err = parseTraceVaultPayout(swapTraces.VaultPayout)
		if err != nil {
			log.Printf("error parsing vault payout for trace %v: %v \n", swapTraces.VaultPayout.Transaction.Hash, err)
		}
	}

	return &models.SwapInfo{
		Notification: swapTransferNotification,
		PoolAddress:  swapTraces.Pool.String(),
		Payment:      payout,
		Referral:     referral,
	}
}

func findSwapTraces(root *tonapi.Trace) []*SwapTraces {
	var swaps []*SwapTraces

	var findNextSwap func(trace *tonapi.Trace)

	findNextSwap = func(trace *tonapi.Trace) {
		notifications := findRouterTransferNotificationNodes(trace)
		for _, notification := range notifications {
			poolAddress := findPoolAddressForNotification(notification)
			payout := findPayoutForNotification(notification)
			vaultPayout := findVaultPayoutForNotification(notification)
			if payout != nil {
				swaps = append(swaps, &SwapTraces{
					Notification: notification,
					Payout:       payout,
					VaultPayout:  vaultPayout,
					Pool:         poolAddress,
				})
			}
			if payout != nil {
				findNextSwap(payout)
			}
			if vaultPayout != nil {
				findNextSwap(vaultPayout)
			}
		}
	}
	findNextSwap(root)
	return swaps
}

func findRouterTransferNotificationNodes(root *tonapi.Trace) []*tonapi.Trace {
	var traces []*tonapi.Trace

	var traverse func(trace *tonapi.Trace)

	traverse = func(trace *tonapi.Trace) {
		if trace.Transaction.InMsg.IsSet() &&
			trace.Transaction.InMsg.Value.OpCode.Value == "0x7362d09c" && // Jetton notification
			core.Contains(trace.Interfaces, stonfiRouterV2) {
			traces = append(traces, trace)
			return
		}

		for _, child := range trace.Children {
			traverse(&child)
		}
	}

	traverse(root)
	return traces
}

func findPoolAddressForNotification(notification *tonapi.Trace) *address.Address {
	children := notification.Children
	if len(children) == 0 {
		return nil
	}
	if core.Contains(notification.Children[0].Interfaces, "stonfi_pool_v2") {
		return address.MustParseRawAddr(children[0].Transaction.Account.Address)
	} else {
		return nil
	}
}

func findPayoutForNotification(notification *tonapi.Trace) *tonapi.Trace {
	return findForNotification(notification, payoutOpCode)
}

func findVaultPayoutForNotification(notification *tonapi.Trace) *tonapi.Trace {
	return findForNotification(notification, payoutOpRefCode)
}

func findForNotification(notification *tonapi.Trace, opCode string) *tonapi.Trace {
	var result *tonapi.Trace

	var traverse func(trace *tonapi.Trace)

	traverse = func(trace *tonapi.Trace) {
		if core.Contains(trace.Interfaces, stonfiRouterV2) {
			if trace.Transaction.InMsg.IsSet() &&
				trace.Transaction.InMsg.Value.DecodedOpName.Value == opCode {
				result = trace
			}
			return
		}
		for _, child := range trace.Children {
			traverse(&child)
		}
	}
	for _, child := range notification.Children {
		traverse(&child)
	}
	return result
}
