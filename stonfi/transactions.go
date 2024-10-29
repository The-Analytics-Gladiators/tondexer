package stonfi

import (
	"TonArb/core"
	"TonArb/models"
	"github.com/tonkeeper/tonapi-go"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"log"
	"slices"
)

type StonfiV1Swap struct {
	Notification *models.SwapTransferNotification
	Payment      *models.PaymentRequest
	Referral     *models.PaymentRequest
}

func ExtractStonfiSwapsFromRootTrace(trace *tonapi.Trace) []*StonfiV1Swap {
	var swaps []core.RelatedEvents[tonapi.Trace, tonapi.Trace]

	var findNextSwap func(trace *tonapi.Trace)

	findNextSwap = func(trace *tonapi.Trace) {
		notifications := findRouterTransferNotificationNodes(trace)
		for _, notification := range notifications {
			payments := findPaymentsForNotification(notification)
			if len(payments) > 0 {
				swaps = append(swaps, core.RelatedEvents[tonapi.Trace, tonapi.Trace]{
					Notification: notification,
					Payments:     payments,
				})

				for _, payment := range payments {
					findNextSwap(payment)
				}
			}
		}
	}
	findNextSwap(trace)

	stonfiSwaps := core.Map(swaps, func(re core.RelatedEvents[tonapi.Trace, tonapi.Trace]) *StonfiV1Swap {
		return relatedEventsToStonfiSwap(re)
	})

	return core.Filter(stonfiSwaps, func(swap *StonfiV1Swap) bool { return swap != nil })
}

func relatedEventsToStonfiSwap(relatedEvents core.RelatedEvents[tonapi.Trace, tonapi.Trace]) *StonfiV1Swap {
	if relatedEvents.Notification == nil {
		log.Printf("Warning: no notification at relatedEvents \n")
		return nil
	}

	notification, e := V1NotificationFromTrace(relatedEvents.Notification)
	if e != nil {
		log.Printf("Warning: could not parse stonfi notification: %v . %v \n", relatedEvents.Notification.Transaction.Hash, e)
		return nil
	}

	if len(relatedEvents.Payments) == 0 || len(relatedEvents.Payments) > 2 {
		log.Printf("Warning: weird number of related events in swap: %v. %v \n", len(relatedEvents.Payments),
			relatedEvents.Notification.Transaction.Hash)
		return nil
	}

	allPayments := core.Map(relatedEvents.Payments, func(trace *tonapi.Trace) *models.PaymentRequest {
		if request, e := PaymentRequestFromTrace(trace); e == nil {
			return request
		}
		return nil
	})
	allPayments = core.Filter(allPayments, func(payment *models.PaymentRequest) bool { return payment != nil })

	paymentIndex := slices.IndexFunc(allPayments, func(request *models.PaymentRequest) bool {
		return request.ExitCode == SwapOkPaymentCode
	})
	if paymentIndex == -1 {
		log.Printf("Warning: no payout in swap: %v \n", relatedEvents.Notification.Transaction.Hash)
		return nil
	}
	payment := allPayments[paymentIndex]
	refPaymentIndex := slices.IndexFunc(allPayments, func(request *models.PaymentRequest) bool {
		return request.ExitCode == SwapRefPaymentCode
	})
	result := &StonfiV1Swap{
		Notification: notification,
		Payment:      payment,
	}

	if refPaymentIndex != -1 {
		refPayment := allPayments[refPaymentIndex]
		result.Referral = refPayment
	}
	return result
}

// Starts from node children
func findPaymentsForNotification(notification *tonapi.Trace) []*tonapi.Trace {
	var payments []*tonapi.Trace

	var traverse func(node *tonapi.Trace)

	traverse = func(trace *tonapi.Trace) {
		account := address.MustParseRawAddr(trace.Transaction.Account.Address)
		parsedTransaction, e := ParseRawTransaction(trace.Transaction.Raw)
		if e == nil &&
			account.String() == StonfiRouter &&
			parsedTransaction.IO.In.MsgType == tlb.MsgTypeInternal {

			slice := parsedTransaction.IO.In.AsInternal().Body.BeginParse()
			if msgCode, e := slice.LoadUInt(32); e == nil && msgCode == PaymentRequestCode {
				payments = append(payments, trace)
				return
			} else {
				//Some other message to router, may be a failed swap
				return
			}
		}
		for _, child := range trace.Children {
			traverse(&child)
		}
	}

	for _, child := range notification.Children {
		traverse(&child)
	}
	return payments
}

func findRouterTransferNotificationNodes(root *tonapi.Trace) []*tonapi.Trace {
	var traces []*tonapi.Trace

	var traverse func(trace *tonapi.Trace)

	traverse = func(trace *tonapi.Trace) {
		account := address.MustParseRawAddr(trace.Transaction.Account.Address)
		parsedTransaction, e := ParseRawTransaction(trace.Transaction.Raw)
		if e == nil &&
			account.String() == StonfiRouter &&
			parsedTransaction.IO.In.MsgType == tlb.MsgTypeInternal {
			slice := parsedTransaction.IO.In.AsInternal().Body.BeginParse()
			if msgCode, e := slice.LoadUInt(32); e == nil && msgCode == TransferNotificationCode {
				traces = append(traces, trace)
				return
			}
		}

		for _, child := range trace.Children {
			traverse(&child)
		}
	}

	traverse(root)

	return traces
}

func GetAllTransactionsFromTrace(trace *tonapi.Trace) []tonapi.Transaction {
	var transactions []tonapi.Transaction
	var traverse func(t *tonapi.Trace)

	traverse = func(t *tonapi.Trace) {
		transactions = append(transactions, t.Transaction)

		for _, child := range t.Children {
			traverse(&child)
		}
	}

	traverse(trace)
	return transactions
}
