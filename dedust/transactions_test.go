package dedust

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/tonkeeper/tonapi-go"
	"github.com/xssnick/tonutils-go/address"
	"testing"
)

func TestFindDedustSwapTokenForTonTrace(t *testing.T) {
	client, _ := tonapi.New()
	params := tonapi.GetTraceParams{TraceID: "5b4a7d422346a9fd28b574623f1ec54502ca346f7cc3d31175377e6029bc0bc5"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapTraces := findSwapTraces(trace)

	assert.Equal(t, 1, len(swapTraces))
	swapTrace := swapTraces[0]

	assert.True(t, swapTrace.OptInVaultWalletTrace.Set)
	assert.Equal(t, "689678158a29f2148d5779a9e462df34696da846822bc955081998c0e7face1e", swapTrace.OptInVaultWalletTrace.Trace.Transaction.Hash)

	assert.Equal(t, "c0ca596d0e907dd7464bf6ef539e5d56d583341e9c7ccba27e6d37a6d90445b9", swapTrace.InVaultTrace.Transaction.Hash)
	assert.Equal(t, "dfdd2c20590c28d58965320bce181f4530d2b93c3d01a2ddabda67c67bc57559", swapTrace.PoolTrace.Transaction.Hash)
	assert.Equal(t, "8b4dfcba22e7eb087ce2aeb9f4c7ab453ce4ec568daea036b916b181042206fc", swapTrace.OutVaultTrace.Transaction.Hash)

	assert.False(t, swapTrace.OptOutVaultWalletTrace.Set)
}

func TestFindDedustSwapTonForTokenTrace(t *testing.T) {
	client, _ := tonapi.New()
	params := tonapi.GetTraceParams{TraceID: "e4a2ea99a1bf8b8812edf9487d70ce9e9c77a1b72345e3abff071624eebf801d"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapTraces := findSwapTraces(trace)

	assert.Equal(t, 1, len(swapTraces))
	swapTrace := swapTraces[0]

	assert.False(t, swapTrace.OptInVaultWalletTrace.Set)

	assert.Equal(t, "3b0411ae3fe1fae4ec0cef2f4ce7ced1864d5f93481259d9a12fac235625b030", swapTrace.InVaultTrace.Transaction.Hash)
	assert.Equal(t, "e8300a4fec12e395f61698817eb25c729902669d81f7f51391668fbc26a95f42", swapTrace.PoolTrace.Transaction.Hash)
	assert.Equal(t, "e4a2ea99a1bf8b8812edf9487d70ce9e9c77a1b72345e3abff071624eebf801d", swapTrace.OutVaultTrace.Transaction.Hash)

	assert.True(t, swapTrace.OptOutVaultWalletTrace.Set)
	assert.Equal(t, "1856ee5a758a58c1f3f0407c07abd5110200c1bd7c61dd3a90e26267f3de444f", swapTrace.OptOutVaultWalletTrace.Trace.Transaction.Hash)
}

func TestFindDedustSwapTokenForToken(t *testing.T) {
	client, _ := tonapi.New()
	params := tonapi.GetTraceParams{TraceID: "87bcc2cfbfc23228d7c27d179bea253c9acfa5899183f05e857a5ec4f63e204c"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapTraces := findSwapTraces(trace)

	assert.Equal(t, 1, len(swapTraces))
	swapTrace := swapTraces[0]

	assert.True(t, swapTrace.OptInVaultWalletTrace.Set)
	assert.Equal(t, "b938693a3708fd35d22b42f044e1265cebed37d942ce6f290976945fbf9464e5", swapTrace.OptInVaultWalletTrace.Trace.Transaction.Hash)

	assert.Equal(t, "b59046a551499a63ba541dc17c22f628f8e36bc1b997f70aff6da8f5aab02058", swapTrace.InVaultTrace.Transaction.Hash)
	assert.Equal(t, "9b0278734bf46115e434a3a48798c9de4e1b728ab1dba9112edbb78b2d90b123", swapTrace.PoolTrace.Transaction.Hash)
	assert.Equal(t, "c63df8e487c3f516847f56c596913b4ed4267a15ca7a5b12557827c05fbbb8f6", swapTrace.OutVaultTrace.Transaction.Hash)

	assert.True(t, swapTrace.OptOutVaultWalletTrace.Set)
	assert.Equal(t, "d7b1548b7cdea086f44b0a96976501e71e567d0cc490f436db0bc9c95efc8b43", swapTrace.OptOutVaultWalletTrace.Trace.Transaction.Hash)
}

func TestNotificationParsingFromTonToTokenSwap(t *testing.T) {
	client, _ := tonapi.New()
	params := tonapi.GetTraceParams{TraceID: "3b0411ae3fe1fae4ec0cef2f4ce7ced1864d5f93481259d9a12fac235625b030"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapTraces := findSwapTraces(trace)

	swapInfo, _ := swapInfoFromDedustTraces(swapTraces[0])

	notification := swapInfo.Notification

	assert.Equal(t, "3b0411ae3fe1fae4ec0cef2f4ce7ced1864d5f93481259d9a12fac235625b030", notification.Hash)
	assert.Equal(t, uint64(50703791000003), notification.Lt)
	//assert.Equal(t, , , notificationFromSwapTraces.TransactionTime)
	//assert.Equal(t, , notificationFromSwapTraces.)
	assert.Equal(t, uint64(275062886401), notification.QueryId)
	assert.Equal(t, uint64(20000000000), notification.Amount)
	assert.Equal(t, address.MustParseRawAddr("0:f7acabf8c0049e3d0030112934d9f383f15e653d567d2a35676c2128e7569f92"), notification.Sender)
	assert.Nil(t, notification.TokenWallet)
	assert.Equal(t, uint64(1082051124), notification.MinOut)
	assert.Nil(t, notification.ToAddress)
	assert.Equal(t, address.MustParseRawAddr("0:a8d9fb483fafc657d0f504f11cb499ab4a9c961348bb0eebfd5542ab52029bfb"), notification.ReferralAddress)
}

func TestNotificationParsingToTokenFromTonSwap(t *testing.T) {
	client, _ := tonapi.New()
	params := tonapi.GetTraceParams{TraceID: "b59046a551499a63ba541dc17c22f628f8e36bc1b997f70aff6da8f5aab02058"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapTraces := findSwapTraces(trace)

	swapInfo, _ := swapInfoFromDedustTraces(swapTraces[0])

	notification := swapInfo.Notification

	assert.Equal(t, "b59046a551499a63ba541dc17c22f628f8e36bc1b997f70aff6da8f5aab02058", notification.Hash)
	assert.Equal(t, uint64(50705113000001), notification.Lt)
	//assert.Equal(t, , , notificationFromSwapTraces.TransactionTime)
	//assert.Equal(t, , notificationFromSwapTraces.)
	assert.Equal(t, uint64(275063046657), notification.QueryId)
	assert.Equal(t, uint64(4717904120), notification.Amount)
	assert.Equal(t, address.MustParseRawAddr("0:575b0dcf52a712a0899aa151d143799f9fbe36e54a7289d1371db23a5a0897e7"), notification.Sender)
	assert.Nil(t, notification.TokenWallet)
	assert.Equal(t, uint64(23426146), notification.MinOut)
	assert.Nil(t, notification.ToAddress)
	assert.Equal(t, address.MustParseRawAddr("0:06fe05fea040552ce0090cfa9a93a53fecf7639b71f8eb4abedbe8398c9a98b7"), notification.ReferralAddress)
}

func TestPaymentParsingForTokenForTokenSwap(t *testing.T) {
	client, _ := tonapi.New()
	params := tonapi.GetTraceParams{TraceID: "c63df8e487c3f516847f56c596913b4ed4267a15ca7a5b12557827c05fbbb8f6"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapTraces := findSwapTraces(trace)

	swapInfo, _ := swapInfoFromDedustTraces(swapTraces[0])

	payment := swapInfo.Payment

	assert.Equal(t, "c63df8e487c3f516847f56c596913b4ed4267a15ca7a5b12557827c05fbbb8f6", payment.Hash)
	assert.Equal(t, uint64(50705120000001), payment.Lt)
	//assert.Equal(t, , payment.)
	//assert.Equal(t, , payment.)
	assert.Equal(t, uint64(275063046657), payment.QueryId)
	assert.Nil(t, payment.Owner)
	assert.Equal(t, uint64(0), payment.ExitCode)
	assert.Equal(t, uint64(0), payment.Amount0Out)
	assert.Equal(t, address.MustParseAddr("EQCSdAZtek06kK-FGDkPl1HMnGTaIQhN7be_FPtToOrkPV0o"), payment.Token0WalletAddress)
	assert.Equal(t, uint64(24150666), payment.Amount1Out)
	assert.Equal(t, address.MustParseAddr("EQCI2sZ8zq25yub6rHEY8FwPqV3zbCqS5oasOdljENCjh0bs"), payment.Token1WalletAddress)

	assert.Equal(t, uint64(0), swapInfo.Referral.Amount1Out)
}

func TestPaymentParsingForTokenForTonSwap(t *testing.T) {
	client, _ := tonapi.New()
	params := tonapi.GetTraceParams{TraceID: "5b4a7d422346a9fd28b574623f1ec54502ca346f7cc3d31175377e6029bc0bc5"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapTraces := findSwapTraces(trace)

	swapInfo, _ := swapInfoFromDedustTraces(swapTraces[0])

	payment := swapInfo.Payment

	assert.Equal(t, "8b4dfcba22e7eb087ce2aeb9f4c7ab453ce4ec568daea036b916b181042206fc", payment.Hash)
	assert.Equal(t, uint64(50704425000001), payment.Lt)
	//assert.Equal(t, , payment.)
	//assert.Equal(t, , payment.)
	//assert.Equal(t, uint64(12344582930068503000), payment.QueryId) // tonviewer shows another id O_o
	assert.Nil(t, payment.Owner)
	assert.Equal(t, uint64(0), payment.ExitCode)
	assert.Equal(t, uint64(0), payment.Amount0Out)
	assert.Equal(t, address.MustParseAddr("EQC4Nm5aWcLYHWqH87fm0Wk5vqcJ5Am_TKcr-Q9jChQu1m7j"), payment.Token0WalletAddress)
	assert.Equal(t, uint64(2616434113), payment.Amount1Out)
	assert.Equal(t, address.MustParseAddr("EQARULUYsmJq1RiZ-YiH-IJLcAZUVkVff-KBPwEmmaQGH6aC"), payment.Token1WalletAddress)

	assert.Nil(t, swapInfo.Referral)
}

func TestPaymentParsingForTonForTokenSwap(t *testing.T) {
	client, _ := tonapi.New()
	params := tonapi.GetTraceParams{TraceID: "3b0411ae3fe1fae4ec0cef2f4ce7ced1864d5f93481259d9a12fac235625b030"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapTraces := findSwapTraces(trace)

	swapInfo, _ := swapInfoFromDedustTraces(swapTraces[0])

	payment := swapInfo.Payment

	assert.Equal(t, "e4a2ea99a1bf8b8812edf9487d70ce9e9c77a1b72345e3abff071624eebf801d", payment.Hash)
	assert.Equal(t, uint64(50703797000001), payment.Lt)
	//assert.Equal(t, , payment.)
	//assert.Equal(t, , payment.)
	assert.Equal(t, uint64(275062886401), payment.QueryId) // tonviewer shows another id O_o
	assert.Nil(t, payment.Owner)
	assert.Equal(t, uint64(0), payment.ExitCode)
	assert.Equal(t, uint64(0), payment.Amount0Out)
	assert.Equal(t, address.MustParseAddr("EQARULUYsmJq1RiZ-YiH-IJLcAZUVkVff-KBPwEmmaQGH6aC"), payment.Token0WalletAddress)
	assert.Equal(t, uint64(1139001183), payment.Amount1Out)
	assert.Equal(t, address.MustParseAddr("EQC4Nm5aWcLYHWqH87fm0Wk5vqcJ5Am_TKcr-Q9jChQu1m7j"), payment.Token1WalletAddress)

	assert.Equal(t, uint64(0), swapInfo.Referral.Amount1Out)
}