package stonfi

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/tonkeeper/tonapi-go"
	"github.com/xssnick/tonutils-go/address"
	"slices"
	"testing"
)

func TestSwapTransferNotificationWithReferral(t *testing.T) {
	client, _ := tonapi.New()

	// Regular swap without ref
	params := tonapi.GetTraceParams{TraceID: "e4bf0ba74bc636ebda771731f3308849b2744b088298047af0d6f2603bdb23d8"}
	trace, _ := client.GetTrace(context.Background(), params)
	notification := findRouterTransferNotificationNodes(trace)[0]

	message, _ := V1NotificationFromTrace(notification)

	assert.Equal(t, message.Hash, "e4bf0ba74bc636ebda771731f3308849b2744b088298047af0d6f2603bdb23d8")
	assert.Equal(t, message.Lt, uint64(49884086000002))
	//assert.Equal(t, message.TransactionTime, tm)
	assert.Equal(t, message.QueryId, uint64(1728730147170))
	assert.Equal(t, message.Amount, uint64(11539116566))
	assert.Equal(t, message.Sender, address.MustParseRawAddr("0:a3fd8c4d3a5bf76f43f8bab26df4a64cc98ea8aedb44c275d0ed3cea09486947"))
	assert.Equal(t, message.TokenWallet, address.MustParseRawAddr("0:1150b518b2626ad51899f98887f8824b70065456455f7fe2813f012699a4061f"))
	assert.Equal(t, message.MinOut, uint64(2830035335))
	assert.Equal(t, message.ToAddress, address.MustParseRawAddr("0:a3fd8c4d3a5bf76f43f8bab26df4a64cc98ea8aedb44c275d0ed3cea09486947"))
	assert.Equal(t, message.ReferralAddress, address.MustParseRawAddr("0:bdf6cf18679ba1a0b5ff09cd6670c99da146ddc4785a27b35b5dc04593e34734"))
}

func TestParsePaymentRequestMessage(t *testing.T) {
	client, _ := tonapi.New()

	params := tonapi.GetTraceParams{TraceID: "d680dca3d9c2448ec69282a35e08a739b04f22d8c2e23f573b63beb0cd62f3a6"}
	trace, _ := client.GetTrace(context.Background(), params)
	notification := findRouterTransferNotificationNodes(trace)[0]

	payments := findPaymentsForNotification(notification)

	index := slices.IndexFunc(payments, func(trace *tonapi.Trace) bool {
		return trace.Transaction.Hash == "d680dca3d9c2448ec69282a35e08a739b04f22d8c2e23f573b63beb0cd62f3a6"
	})

	payment := payments[index]
	message, _ := PaymentRequestFromTrace(payment)

	assert.Equal(t, message.Hash, "d680dca3d9c2448ec69282a35e08a739b04f22d8c2e23f573b63beb0cd62f3a6")
	assert.Equal(t, message.Lt, uint64(49884885000003))
	//assert.Equal(t, message.TransactionTime, tm)
	assert.Equal(t, message.QueryId, uint64(165066398389))
	assert.Equal(t, message.Owner, address.MustParseRawAddr("0:c5f5ca55b18af2a46f9a479ae81504b5fc0ba2b43062a6f5311d4783a5e447ed"))
	assert.Equal(t, message.ExitCode, uint64(3326308581))
	assert.Equal(t, message.Amount0Out, uint64(0))
	assert.Equal(t, message.Token0Address, address.MustParseRawAddr("0:f38723ef1e85f751e34de3ab108ff6e2e3837b9f2ad156560ad709ce7392d5c8"))
	assert.Equal(t, message.Amount1Out, uint64(30999999999))
	assert.Equal(t, message.Token1Address, address.MustParseRawAddr("0:1150b518b2626ad51899f98887f8824b70065456455f7fe2813f012699a4061f"))
}
