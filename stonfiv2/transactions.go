package stonfiv2

import (
	"github.com/tonkeeper/tonapi-go"
	"github.com/xssnick/tonutils-go/address"
	"log"
	"tondexer/core"
	"tondexer/models"
)

const payoutOpCode = "stonfi_pay_to_v2"
const payoutOpRefCode = "stonfi_pay_vault_v2"

const stonfiRouterV2 = "stonfi_router_v2"

type StonfiV2SwapTraces struct {
	Root         *tonapi.Trace
	Notification *tonapi.Trace
	Payout       *tonapi.Trace
	VaultPayout  *tonapi.Trace
	Pool         *address.Address
}

func ExtractStonfiV2SwapsFromRootTrace(trace *tonapi.Trace) []*models.SwapInfo {
	infos := core.Map(findSwapTraces(trace), func(t *StonfiV2SwapTraces) *models.SwapInfo {
		return swapTracesToSwapInfo(t)
	})
	return core.Filter(infos, func(info *models.SwapInfo) bool {
		return info != nil
	})
}

func swapTracesToSwapInfo(swapTraces *StonfiV2SwapTraces) *models.SwapInfo {
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
		TraceID:      swapTraces.Root.Transaction.Hash,
		Notification: swapTransferNotification,
		PoolAddress:  swapTraces.Pool,
		Payment:      payout,
		Referral:     referral,
	}
}

func findSwapTraces(root *tonapi.Trace) []*StonfiV2SwapTraces {
	var swapTraces []*StonfiV2SwapTraces

	var findNextSwap func(trace *tonapi.Trace)

	findNextSwap = func(trace *tonapi.Trace) {
		notifications := findRouterTransferNotificationNodes(trace)
		for _, notification := range notifications {
			poolAddress := findPoolAddressForNotification(notification)
			payout := findPayoutForNotification(notification)
			vaultPayout := findVaultPayoutForNotification(notification)
			if payout != nil {
				swapTraces = append(swapTraces, &StonfiV2SwapTraces{
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
	for _, swapTrace := range swapTraces {
		swapTrace.Root = root
	}
	return swapTraces
}

func findRouterTransferNotificationNodes(root *tonapi.Trace) []*tonapi.Trace {
	var traces []*tonapi.Trace

	var traverse func(trace *tonapi.Trace)

	traverse = func(trace *tonapi.Trace) {

		if trace.Transaction.InMsg.IsSet() &&
			trace.Transaction.InMsg.Value.OpCode.Value == core.JettonNotifyOpCode && // Jetton notification
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
