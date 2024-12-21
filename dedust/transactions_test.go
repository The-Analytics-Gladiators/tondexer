package dedust

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/tonkeeper/tonapi-go"
	"github.com/xssnick/tonutils-go/address"
	"math/big"
	"testing"
)

func TestFindDedustSwapTokenForTonTrace(t *testing.T) {
	client, _ := tonapi.New()
	params := tonapi.GetTraceParams{TraceID: "5b4a7d422346a9fd28b574623f1ec54502ca346f7cc3d31175377e6029bc0bc5"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapTraces := findSwapTraces(trace)

	assert.Equal(t, 1, len(swapTraces))
	swapTrace := swapTraces[0]
	assert.Equal(t, "5b4a7d422346a9fd28b574623f1ec54502ca346f7cc3d31175377e6029bc0bc5", swapTrace.Root.Transaction.Hash)

	assert.True(t, swapTrace.OptInVaultWalletTrace.Set)
	assert.Equal(t, "689678158a29f2148d5779a9e462df34696da846822bc955081998c0e7face1e", swapTrace.OptInVaultWalletTrace.Trace.Transaction.Hash)

	assert.Equal(t, "c0ca596d0e907dd7464bf6ef539e5d56d583341e9c7ccba27e6d37a6d90445b9", swapTrace.InVaultTrace.Transaction.Hash)
	assert.Equal(t, "dfdd2c20590c28d58965320bce181f4530d2b93c3d01a2ddabda67c67bc57559", swapTrace.PoolTraces[0].Transaction.Hash)
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

	assert.Equal(t, "982eeb30aa2751375badc07ba4d644f0a273fb5dfadd2adc5016e4f0311885fa", swapTrace.Root.Transaction.Hash)
	assert.False(t, swapTrace.OptInVaultWalletTrace.Set)

	assert.Equal(t, "3b0411ae3fe1fae4ec0cef2f4ce7ced1864d5f93481259d9a12fac235625b030", swapTrace.InVaultTrace.Transaction.Hash)
	assert.Equal(t, "e8300a4fec12e395f61698817eb25c729902669d81f7f51391668fbc26a95f42", swapTrace.PoolTraces[0].Transaction.Hash)
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

	assert.Equal(t, "87bcc2cfbfc23228d7c27d179bea253c9acfa5899183f05e857a5ec4f63e204c", swapTrace.Root.Transaction.Hash)

	assert.True(t, swapTrace.OptInVaultWalletTrace.Set)
	assert.Equal(t, "b938693a3708fd35d22b42f044e1265cebed37d942ce6f290976945fbf9464e5", swapTrace.OptInVaultWalletTrace.Trace.Transaction.Hash)

	assert.Equal(t, "b59046a551499a63ba541dc17c22f628f8e36bc1b997f70aff6da8f5aab02058", swapTrace.InVaultTrace.Transaction.Hash)
	assert.Equal(t, "9b0278734bf46115e434a3a48798c9de4e1b728ab1dba9112edbb78b2d90b123", swapTrace.PoolTraces[0].Transaction.Hash)
	assert.Equal(t, "c63df8e487c3f516847f56c596913b4ed4267a15ca7a5b12557827c05fbbb8f6", swapTrace.OutVaultTrace.Transaction.Hash)

	assert.True(t, swapTrace.OptOutVaultWalletTrace.Set)
	assert.Equal(t, "d7b1548b7cdea086f44b0a96976501e71e567d0cc490f436db0bc9c95efc8b43", swapTrace.OptOutVaultWalletTrace.Trace.Transaction.Hash)
}

func TestFindSwapTracesForThreeCycle(t *testing.T) {
	client, _ := tonapi.New()
	params := tonapi.GetTraceParams{TraceID: "528f4366cc050178566b635a39ab7810c31d95f82aaeeb60b0c0595d2300202f"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapTraces := findSwapTraces(trace)

	assert.Equal(t, 1, len(swapTraces))
	swapTrace := swapTraces[0]

	assert.Equal(t, "528f4366cc050178566b635a39ab7810c31d95f82aaeeb60b0c0595d2300202f", swapTrace.Root.Transaction.Hash)

	assert.False(t, swapTrace.OptInVaultWalletTrace.Set)
	assert.False(t, swapTrace.OptOutVaultWalletTrace.Set)

	assert.Equal(t, "97288e4f9833781ed777d14683f739b06149d6e3c4620af1ca490359bef032fb", swapTrace.InVaultTrace.Transaction.Hash)
	assert.Equal(t, "5703c142bd3f031c1150bf8d2a5cd3c14a2a21637e0b2913c73925dabf657215", swapTrace.OutVaultTrace.Transaction.Hash)

	assert.Equal(t, 3, len(swapTrace.PoolTraces))

	assert.Equal(t, "93edc8b21280725ccb584f3e433deafdb1bd1e61ed2416f3e47ebe79ba9a8667", swapTrace.PoolTraces[0].Transaction.Hash)
	assert.Equal(t, "1a13847107f7a046bc5a1058d0046f7af93cf8d4e8f60ecf52587de557142ee1", swapTrace.PoolTraces[1].Transaction.Hash)
	assert.Equal(t, "31cf09018cf50e6808c978e8bd51e04cc2fda10892f23381ebee832233ea8ee9", swapTrace.PoolTraces[2].Transaction.Hash)
}

//---------------------------- NEW

func TestDedustSwapInfoForOneSwap(t *testing.T) {
	client, _ := tonapi.New()
	params := tonapi.GetTraceParams{TraceID: "87bcc2cfbfc23228d7c27d179bea253c9acfa5899183f05e857a5ec4f63e204c"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapTraces := findSwapTraces(trace)

	assert.Equal(t, 1, len(swapTraces))
	swapTrace := swapTraces[0]

	dedustSwapInfo, _ := dedustSwapInfoFromDedustTraces(swapTrace)

	assert.Equal(t, "87bcc2cfbfc23228d7c27d179bea253c9acfa5899183f05e857a5ec4f63e204c", dedustSwapInfo.TraceID)
	assert.Equal(t, address.MustParseAddr("UQBXWw3PUqcSoImaoVHRQ3mfn7425UpyidE3HbI6WgiX56ao"), dedustSwapInfo.Sender)
	assert.Equal(t, address.MustParseAddr("EQCSdAZtek06kK-FGDkPl1HMnGTaIQhN7be_FPtToOrkPV0o"), dedustSwapInfo.InWalletAddress)
	assert.Equal(t, address.MustParseAddr("EQCI2sZ8zq25yub6rHEY8FwPqV3zbCqS5oasOdljENCjh0bs"), dedustSwapInfo.OutWalletAddress)
	assert.Equal(t, big.NewInt(24150666), dedustSwapInfo.OutAmount)

	assert.Equal(t, 1, len(dedustSwapInfo.PoolsInfo))
	poolInfo := dedustSwapInfo.PoolsInfo[0]
	assert.Equal(t, "9b0278734bf46115e434a3a48798c9de4e1b728ab1dba9112edbb78b2d90b123", poolInfo.Hash)
	assert.Equal(t, address.MustParseAddr("EQCm92zFBkLe_qcFDp7WBvI6JFSDsm4WbDPvZ7xNd7nPL_6M"), poolInfo.Address)
	assert.Equal(t, address.MustParseAddr("UQBXWw3PUqcSoImaoVHRQ3mfn7425UpyidE3HbI6WgiX56ao"), poolInfo.Sender)
	assert.Nil(t, poolInfo.JettonIn)
	assert.Equal(t, big.NewInt(4717904120), poolInfo.AmountIn)
	assert.Equal(t, big.NewInt(23426146), poolInfo.Limit)
}

func TestDedustSwapInfoForThreeSwap(t *testing.T) {
	client, _ := tonapi.New()
	params := tonapi.GetTraceParams{TraceID: "fad824432d05c95ccd0e0926d5a16d39356a24366c48bd5e81e7b9ce12a9ad37"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapTraces := findSwapTraces(trace)

	assert.Equal(t, 1, len(swapTraces))
	swapTrace := swapTraces[0]

	dedustSwapInfo, _ := dedustSwapInfoFromDedustTraces(swapTrace)
	assert.Equal(t, 3, len(dedustSwapInfo.PoolsInfo))
	poolsInfo := dedustSwapInfo.PoolsInfo

	assert.Equal(t, "fad824432d05c95ccd0e0926d5a16d39356a24366c48bd5e81e7b9ce12a9ad37", dedustSwapInfo.TraceID)
	assert.Equal(t, address.MustParseAddr("UQDQ7jqqGUsLNDYwTTHo-E14ehHBPv1oVIw3Jam7_7SZBZoX"), dedustSwapInfo.Sender)
	assert.Nil(t, dedustSwapInfo.InWalletAddress)
	assert.Nil(t, dedustSwapInfo.OutWalletAddress)
	assert.Equal(t, big.NewInt(32438750071), dedustSwapInfo.OutAmount)

	assert.Equal(t, address.MustParseAddr("EQCk6tGPlFoQ_1TgZJjuiulfSJz5aoJgnyy29eLsXtOmeYDw"), poolsInfo[0].Address)
	assert.Equal(t, "8afe3eb5942ea610c7cbbef69b83be46dcb135c5524f2cc1a8236591e25b77f3", poolsInfo[0].Hash)
	assert.Equal(t, address.MustParseAddr("UQDQ7jqqGUsLNDYwTTHo-E14ehHBPv1oVIw3Jam7_7SZBZoX"), poolsInfo[0].Sender)
	assert.Nil(t, poolsInfo[0].JettonIn)
	assert.Equal(t, big.NewInt(32267507305), poolsInfo[0].AmountIn)
	assert.Equal(t, big.NewInt(170803689), poolsInfo[0].Limit)

	assert.Equal(t, address.MustParseAddr("EQC6Ckm9EFlZGgQqip3IlvZFkCh8fY_-P39V9puQwe4T_fp7"), poolsInfo[1].Address)
	assert.Equal(t, "3af798b963d11fe5b3cfacfb8a8d57ffc1e4a9e122e202750b83b1af6b3b8567", poolsInfo[1].Hash)
	assert.Equal(t, address.MustParseAddr("UQDQ7jqqGUsLNDYwTTHo-E14ehHBPv1oVIw3Jam7_7SZBZoX"), poolsInfo[1].Sender)
	assert.Equal(t, address.MustParseRawAddr("0:729c13b6df2c07cbf0a06ab63d34af454f3d320ec1bcd8fb5c6d24d0806a17c2"), poolsInfo[1].JettonIn)
	assert.Equal(t, big.NewInt(171334827), poolsInfo[1].AmountIn)
	assert.Equal(t, big.NewInt(167185834), poolsInfo[1].Limit)

	assert.Equal(t, address.MustParseAddr("EQA-X_yo3fzzbDbJ_0bzFWKqtRuZFIRa1sJsveZJ1YpViO3r"), poolsInfo[2].Address)
	assert.Equal(t, "abf04efdd07426d57c047ec17a7f756557dd94a0ea4e14dc33d5b4137dc20c1c", poolsInfo[2].Hash)
	assert.Equal(t, address.MustParseAddr("UQDQ7jqqGUsLNDYwTTHo-E14ehHBPv1oVIw3Jam7_7SZBZoX"), poolsInfo[2].Sender)
	assert.Equal(t, address.MustParseRawAddr("0:b113a994b5024a16719f69139328eb759596c38a25f59028b146fecdc3621dfe"), poolsInfo[2].JettonIn)
	assert.Equal(t, big.NewInt(168211927), poolsInfo[2].AmountIn)
	assert.Equal(t, big.NewInt(32103129526), poolsInfo[2].Limit)
}

func TestSeveralSwapsWithFailed(t *testing.T) {
	client, _ := tonapi.New()
	params := tonapi.GetTraceParams{TraceID: "6932e9372c48b340f0bccc4cb2a06c12a0093e7241c9bbfebcee753205bbd6a7"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapTraces := findSwapTraces(trace)

	assert.Equal(t, 1, len(swapTraces))
	swapTrace := swapTraces[0]

	dedustSwapInfo, _ := dedustSwapInfoFromDedustTraces(swapTrace)
	assert.Equal(t, 1, len(dedustSwapInfo.PoolsInfo))
	poolsInfo := dedustSwapInfo.PoolsInfo

	assert.Equal(t, "6932e9372c48b340f0bccc4cb2a06c12a0093e7241c9bbfebcee753205bbd6a7", dedustSwapInfo.TraceID)
	assert.Equal(t, address.MustParseAddr("UQDQ7jqqGUsLNDYwTTHo-E14ehHBPv1oVIw3Jam7_7SZBZoX"), dedustSwapInfo.Sender)
	assert.Nil(t, dedustSwapInfo.InWalletAddress)
	assert.Equal(t, address.MustParseAddr("EQCI2sZ8zq25yub6rHEY8FwPqV3zbCqS5oasOdljENCjh0bs"), dedustSwapInfo.OutWalletAddress)
	assert.Equal(t, big.NewInt(33256486), dedustSwapInfo.OutAmount)

	assert.Equal(t, address.MustParseAddr("EQA-X_yo3fzzbDbJ_0bzFWKqtRuZFIRa1sJsveZJ1YpViO3r"), poolsInfo[0].Address)
	assert.Equal(t, "b5ef6d64f4017e436083c2c63d7db35540c007e50595e89dc4bf813d4f1c143d", poolsInfo[0].Hash)
	assert.Equal(t, address.MustParseAddr("UQDQ7jqqGUsLNDYwTTHo-E14ehHBPv1oVIw3Jam7_7SZBZoX"), poolsInfo[0].Sender)
	assert.Nil(t, poolsInfo[0].JettonIn)
	assert.Equal(t, big.NewInt(6006011959), poolsInfo[0].AmountIn)
	assert.Equal(t, big.NewInt(32834405), poolsInfo[0].Limit)
}

//---------------------------- OLD

func TestNotificationParsingFromTonToTokenSwap(t *testing.T) {
	client, _ := tonapi.New()
	params := tonapi.GetTraceParams{TraceID: "3b0411ae3fe1fae4ec0cef2f4ce7ced1864d5f93481259d9a12fac235625b030"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapTraces := findSwapTraces(trace)

	dedustSwapInfo, _ := dedustSwapInfoFromDedustTraces(swapTraces[0])

	assert.Equal(t, "982eeb30aa2751375badc07ba4d644f0a273fb5dfadd2adc5016e4f0311885fa", dedustSwapInfo.TraceID)
	assert.Equal(t, address.MustParseAddr("UQD3rKv4wASePQAwESk02fOD8V5lPVZ9KjVnbCEo51afkguH"), dedustSwapInfo.Sender)
	assert.Nil(t, dedustSwapInfo.InWalletAddress)
	assert.Equal(t, address.MustParseAddr("EQC4Nm5aWcLYHWqH87fm0Wk5vqcJ5Am_TKcr-Q9jChQu1m7j"), dedustSwapInfo.OutWalletAddress)
	assert.Equal(t, big.NewInt(1139001183), dedustSwapInfo.OutAmount)

	assert.Equal(t, 1, len(dedustSwapInfo.PoolsInfo))
	poolInfo := dedustSwapInfo.PoolsInfo[0]
	assert.Equal(t, "e8300a4fec12e395f61698817eb25c729902669d81f7f51391668fbc26a95f42", poolInfo.Hash)
	assert.Equal(t, address.MustParseAddr("EQBTbu-Q5sOShpEXfEUAq9lL378Fog2DdVcem8mebNND_lsf"), poolInfo.Address)
	assert.Equal(t, address.MustParseAddr("UQD3rKv4wASePQAwESk02fOD8V5lPVZ9KjVnbCEo51afkguH"), poolInfo.Sender)
	assert.Nil(t, poolInfo.JettonIn)
	assert.Equal(t, big.NewInt(20000000000), poolInfo.AmountIn)
	assert.Equal(t, big.NewInt(1082051124), poolInfo.Limit)
}

func TestNotificationParsingToTokenFromTonSwap(t *testing.T) {
	client, _ := tonapi.New()
	params := tonapi.GetTraceParams{TraceID: "b59046a551499a63ba541dc17c22f628f8e36bc1b997f70aff6da8f5aab02058"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapTraces := findSwapTraces(trace)

	dedustSwapInfo, _ := dedustSwapInfoFromDedustTraces(swapTraces[0])

	assert.Equal(t, "87bcc2cfbfc23228d7c27d179bea253c9acfa5899183f05e857a5ec4f63e204c", dedustSwapInfo.TraceID)
	assert.Equal(t, address.MustParseAddr("UQBXWw3PUqcSoImaoVHRQ3mfn7425UpyidE3HbI6WgiX56ao"), dedustSwapInfo.Sender)
	assert.Equal(t, address.MustParseAddr("EQCSdAZtek06kK-FGDkPl1HMnGTaIQhN7be_FPtToOrkPV0o"), dedustSwapInfo.InWalletAddress)
	assert.Equal(t, address.MustParseAddr("EQCI2sZ8zq25yub6rHEY8FwPqV3zbCqS5oasOdljENCjh0bs"), dedustSwapInfo.OutWalletAddress)
	assert.Equal(t, big.NewInt(24150666), dedustSwapInfo.OutAmount)

	assert.Equal(t, 1, len(dedustSwapInfo.PoolsInfo))
	poolInfo := dedustSwapInfo.PoolsInfo[0]
	assert.Equal(t, "9b0278734bf46115e434a3a48798c9de4e1b728ab1dba9112edbb78b2d90b123", poolInfo.Hash)
	assert.Equal(t, address.MustParseAddr("EQCm92zFBkLe_qcFDp7WBvI6JFSDsm4WbDPvZ7xNd7nPL_6M"), poolInfo.Address)
	assert.Equal(t, address.MustParseAddr("UQBXWw3PUqcSoImaoVHRQ3mfn7425UpyidE3HbI6WgiX56ao"), poolInfo.Sender)
	assert.Nil(t, poolInfo.JettonIn)
	assert.Equal(t, big.NewInt(4717904120), poolInfo.AmountIn)
	assert.Equal(t, big.NewInt(23426146), poolInfo.Limit)
}

func TestTraceIdIsSetForSwapInfo(t *testing.T) {
	client, _ := tonapi.New()
	params := tonapi.GetTraceParams{TraceID: "87bcc2cfbfc23228d7c27d179bea253c9acfa5899183f05e857a5ec4f63e204c"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapTraces := findSwapTraces(trace)

	swapInfo, _ := dedustSwapInfoFromDedustTraces(swapTraces[0])

	assert.Equal(t, "87bcc2cfbfc23228d7c27d179bea253c9acfa5899183f05e857a5ec4f63e204c", swapInfo.TraceID)
}

func TestPaymentParsingForTokenForTokenSwap(t *testing.T) {
	client, _ := tonapi.New()
	params := tonapi.GetTraceParams{TraceID: "c63df8e487c3f516847f56c596913b4ed4267a15ca7a5b12557827c05fbbb8f6"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapTraces := findSwapTraces(trace)

	dedustSwapInfo, _ := dedustSwapInfoFromDedustTraces(swapTraces[0])

	assert.Equal(t, "87bcc2cfbfc23228d7c27d179bea253c9acfa5899183f05e857a5ec4f63e204c", dedustSwapInfo.TraceID)
	assert.Equal(t, address.MustParseAddr("UQBXWw3PUqcSoImaoVHRQ3mfn7425UpyidE3HbI6WgiX56ao"), dedustSwapInfo.Sender)
	assert.Equal(t, address.MustParseAddr("EQCSdAZtek06kK-FGDkPl1HMnGTaIQhN7be_FPtToOrkPV0o"), dedustSwapInfo.InWalletAddress)
	assert.Equal(t, address.MustParseAddr("EQCI2sZ8zq25yub6rHEY8FwPqV3zbCqS5oasOdljENCjh0bs"), dedustSwapInfo.OutWalletAddress)
	assert.Equal(t, big.NewInt(24150666), dedustSwapInfo.OutAmount)

	assert.Equal(t, 1, len(dedustSwapInfo.PoolsInfo))
	poolInfo := dedustSwapInfo.PoolsInfo[0]
	assert.Equal(t, "9b0278734bf46115e434a3a48798c9de4e1b728ab1dba9112edbb78b2d90b123", poolInfo.Hash)
	assert.Equal(t, address.MustParseAddr("EQCm92zFBkLe_qcFDp7WBvI6JFSDsm4WbDPvZ7xNd7nPL_6M"), poolInfo.Address)
	assert.Equal(t, address.MustParseAddr("UQBXWw3PUqcSoImaoVHRQ3mfn7425UpyidE3HbI6WgiX56ao"), poolInfo.Sender)
	assert.Nil(t, poolInfo.JettonIn)
	assert.Equal(t, big.NewInt(4717904120), poolInfo.AmountIn)
	assert.Equal(t, big.NewInt(23426146), poolInfo.Limit)
}

func TestPaymentParsingForTokenForTonSwap(t *testing.T) {
	client, _ := tonapi.New()
	params := tonapi.GetTraceParams{TraceID: "5b4a7d422346a9fd28b574623f1ec54502ca346f7cc3d31175377e6029bc0bc5"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapTraces := findSwapTraces(trace)

	dedustSwapInfo, _ := dedustSwapInfoFromDedustTraces(swapTraces[0])

	assert.Equal(t, "5b4a7d422346a9fd28b574623f1ec54502ca346f7cc3d31175377e6029bc0bc5", dedustSwapInfo.TraceID)
	assert.Equal(t, address.MustParseAddr("UQAGuoF4qb-KLJlAO0PAjyigenjHNRHv-YiqspwLzB0D6ZmE"), dedustSwapInfo.Sender)
	assert.Equal(t, address.MustParseAddr("EQC4Nm5aWcLYHWqH87fm0Wk5vqcJ5Am_TKcr-Q9jChQu1m7j"), dedustSwapInfo.InWalletAddress)
	assert.Nil(t, dedustSwapInfo.OutWalletAddress)
	assert.Equal(t, big.NewInt(2616434113), dedustSwapInfo.OutAmount)

	assert.Equal(t, 1, len(dedustSwapInfo.PoolsInfo))
	poolInfo := dedustSwapInfo.PoolsInfo[0]
	assert.Equal(t, "dfdd2c20590c28d58965320bce181f4530d2b93c3d01a2ddabda67c67bc57559", poolInfo.Hash)
	assert.Equal(t, address.MustParseAddr("EQBTbu-Q5sOShpEXfEUAq9lL378Fog2DdVcem8mebNND_lsf"), poolInfo.Address)
	assert.Equal(t, address.MustParseAddr("UQAGuoF4qb-KLJlAO0PAjyigenjHNRHv-YiqspwLzB0D6ZmE"), poolInfo.Sender)
	assert.Nil(t, poolInfo.JettonIn)
	assert.Equal(t, big.NewInt(150000000), poolInfo.AmountIn)
	assert.Equal(t, big.NewInt(2485612407), poolInfo.Limit)
}

func TestPaymentParsingForTonForTokenSwap(t *testing.T) {
	client, _ := tonapi.New()
	params := tonapi.GetTraceParams{TraceID: "3b0411ae3fe1fae4ec0cef2f4ce7ced1864d5f93481259d9a12fac235625b030"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapTraces := findSwapTraces(trace)

	dedustSwapInfo, _ := dedustSwapInfoFromDedustTraces(swapTraces[0])

	assert.Equal(t, "982eeb30aa2751375badc07ba4d644f0a273fb5dfadd2adc5016e4f0311885fa", dedustSwapInfo.TraceID)
	assert.Equal(t, address.MustParseAddr("UQD3rKv4wASePQAwESk02fOD8V5lPVZ9KjVnbCEo51afkguH"), dedustSwapInfo.Sender)
	assert.Nil(t, dedustSwapInfo.InWalletAddress)
	assert.Equal(t, address.MustParseAddr("EQC4Nm5aWcLYHWqH87fm0Wk5vqcJ5Am_TKcr-Q9jChQu1m7j"), dedustSwapInfo.OutWalletAddress)
	assert.Equal(t, big.NewInt(1139001183), dedustSwapInfo.OutAmount)

	assert.Equal(t, 1, len(dedustSwapInfo.PoolsInfo))
	poolInfo := dedustSwapInfo.PoolsInfo[0]
	assert.Equal(t, "e8300a4fec12e395f61698817eb25c729902669d81f7f51391668fbc26a95f42", poolInfo.Hash)
	assert.Equal(t, address.MustParseAddr("EQBTbu-Q5sOShpEXfEUAq9lL378Fog2DdVcem8mebNND_lsf"), poolInfo.Address)
	assert.Equal(t, address.MustParseAddr("UQD3rKv4wASePQAwESk02fOD8V5lPVZ9KjVnbCEo51afkguH"), poolInfo.Sender)
	assert.Nil(t, poolInfo.JettonIn)
	assert.Equal(t, big.NewInt(20000000000), poolInfo.AmountIn)
	assert.Equal(t, big.NewInt(1082051124), poolInfo.Limit)
}

func TestParseBigLimitAndAmount(t *testing.T) {
	client, _ := tonapi.New()
	params := tonapi.GetTraceParams{TraceID: "0b4bf597f1a07d0a97ab8cc7c1e961c2013ba974c4708487eab2afc7ba3e0b76"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapTraces := findSwapTraces(trace)

	dedustSwapInfo, _ := dedustSwapInfoFromDedustTraces(swapTraces[0])

	assert.Equal(t, "7b292e40f678bd0cb85a338b6a9431a599923db67672a3deba10117218b5e7dc", dedustSwapInfo.TraceID)
	assert.Equal(t, address.MustParseAddr("UQCGLlk6QbchZAxgkIb7kHPT1G75B5GfD3KK4Mf4XJQrV7m1"), dedustSwapInfo.Sender)
	assert.Nil(t, dedustSwapInfo.InWalletAddress)
	assert.Equal(t, address.MustParseAddr("EQBYkoAByPc5wgVoLY3UqRdnOAuDgmrv5piz3nqOzSaaCQwO"), dedustSwapInfo.OutWalletAddress)
	amount, _ := new(big.Int).SetString("1865030432608179035859", 10)
	assert.Equal(t, amount, dedustSwapInfo.OutAmount)

	assert.Equal(t, 1, len(dedustSwapInfo.PoolsInfo))
	poolInfo := dedustSwapInfo.PoolsInfo[0]
	assert.Equal(t, "c6537f854b0dc23e77c3e3f1e97c2c469a53fa174a6197c5bdbd7c695aba25d7", poolInfo.Hash)
	assert.Equal(t, address.MustParseAddr("EQC9KPFTUcoHql3xHFp4Dh1nHCNP8jA7jr6Xz0pBOO0X8bur"), poolInfo.Address)
	assert.Equal(t, address.MustParseAddr("UQCGLlk6QbchZAxgkIb7kHPT1G75B5GfD3KK4Mf4XJQrV7m1"), poolInfo.Sender)
	assert.Nil(t, poolInfo.JettonIn)
	assert.Equal(t, big.NewInt(92880000000), poolInfo.AmountIn)
	limit, _ := new(big.Int).SetString("1863165402175570856823", 10)
	assert.Equal(t, limit, poolInfo.Limit)
}

func TestWhenInitialAccountHasTwoOutgoingMessages(t *testing.T) {
	client, _ := tonapi.New()
	params := tonapi.GetTraceParams{TraceID: "5ac93cd223409580fcdd4c0899e9a88f3acf831cb1fffbb742ca0ec6103a0cbb"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapTraces := findSwapTraces(trace)
	assert.Equal(t, 1, len(swapTraces))
	dedustSwapInfo, _ := dedustSwapInfoFromDedustTraces(swapTraces[0])

	assert.Equal(t, "5ac93cd223409580fcdd4c0899e9a88f3acf831cb1fffbb742ca0ec6103a0cbb", dedustSwapInfo.TraceID)
	assert.Equal(t, address.MustParseAddr("UQAfjQ8EnhchNZdHLlE3D8Y9aIkhIZMMcjUUlf7pGvkMGG7z"), dedustSwapInfo.Sender)
	assert.Nil(t, dedustSwapInfo.InWalletAddress)
	assert.Equal(t, address.MustParseAddr("EQAH_hFK0FBx5Yd2qfsAV0uRZ-m92X0EzqXvvnMfGPzVmENN"), dedustSwapInfo.OutWalletAddress)
	assert.Equal(t, big.NewInt(149088615976), dedustSwapInfo.OutAmount)

	assert.Equal(t, 1, len(dedustSwapInfo.PoolsInfo))
	poolInfo := dedustSwapInfo.PoolsInfo[0]
	assert.Equal(t, "ff25f72299d2f25c25c5ee511dbe0ff33faa5b2966cde26f1301a6daa67c199c", poolInfo.Hash)
	assert.Equal(t, address.MustParseAddr("EQBGXZ9ddZeWypx8EkJieHJX75ct0bpkmu0Y4YoYr3NM0Z9e"), poolInfo.Address)
	assert.Equal(t, address.MustParseRawAddr("0:1f8d0f049e17213597472e51370fc63d68892121930c72351495fee91af90c18"), poolInfo.Sender)
	assert.Nil(t, poolInfo.JettonIn)
	assert.Equal(t, big.NewInt(10000000000), poolInfo.AmountIn)
	assert.Equal(t, big.NewInt(141634185178), poolInfo.Limit)
}

func TestRootTrace(t *testing.T) {
	client, _ := tonapi.New()
	params := tonapi.GetTraceParams{TraceID: "dd4500f05ae7c10d96bd446e1eada3aa5eae08e74dbf6b165d32ba498e63b46d"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapTraces := findSwapTraces(trace)

	assert.Equal(t, "5ac93cd223409580fcdd4c0899e9a88f3acf831cb1fffbb742ca0ec6103a0cbb", swapTraces[0].Root.Transaction.Hash)
}

func TestTraceIFForSwapInfo(t *testing.T) {
	client, _ := tonapi.New()
	params := tonapi.GetTraceParams{TraceID: "0b4bf597f1a07d0a97ab8cc7c1e961c2013ba974c4708487eab2afc7ba3e0b76"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapInfos := ExtractDedustSwapsFromRootTrace(trace)

	assert.Equal(t, 1, len(swapInfos))
	assert.Equal(t, "7b292e40f678bd0cb85a338b6a9431a599923db67672a3deba10117218b5e7dc", swapInfos[0].TraceID)
}

// https://tonviewer.com/transaction/3bc93c9d4696ec75b1e44106f613215e4eaab00894f80bb6cd58ee5aba67b39a both dedust and stonfiv2 and arbitrage
// https://tonviewer.com/transaction/bbea35f5402a7c99530eaa83ce13daa444ad8ebe47d60257e362848f9336ca91 - 3 swaps stonfi and dedust

// 3-cycle, but slightly unmached by out-in: https://tonviewer.com/transaction/04fada65227c3fac0fd34ff731035fe2a2a498e1dcb06c53d6a98de8d1193e9c
