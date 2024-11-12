package stonfiv2

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/tonkeeper/tonapi-go"
	"github.com/xssnick/tonutils-go/address"
	"math/big"
	"testing"
)

func Test_findRouterTransferNotificationNodes(t *testing.T) {
	client, _ := tonapi.New()

	params := tonapi.GetTraceParams{TraceID: "49c2837771b6b07b0984ac3d29d57aa92e8ddb85854765fe50b42239e28bfa34"}
	trace, _ := client.GetTrace(context.Background(), params)

	notificationTraces := findRouterTransferNotificationNodes(trace)

	assert.Equal(t, 1, len(notificationTraces))
	assert.Equal(t, "2e51309e3c0cefca28582f169038c2a0f2b49c7461b9a15beab5077b87394c66", notificationTraces[0].Transaction.Hash)
}

func Test_findPoolAddressForNotification(t *testing.T) {
	client, _ := tonapi.New()

	params := tonapi.GetTraceParams{TraceID: "49c2837771b6b07b0984ac3d29d57aa92e8ddb85854765fe50b42239e28bfa34"}
	trace, _ := client.GetTrace(context.Background(), params)

	notificationTrace := findRouterTransferNotificationNodes(trace)[0]

	poolAddress := findPoolAddressForNotification(notificationTrace)

	assert.Equal(t, address.MustParseAddr("EQBOdxSlGhjxjVYVwi6blHxfS1tWsOXesMrA1npnuiWeOKgI"), poolAddress)
}

func Test_findPayoutForNotification(t *testing.T) {
	client, _ := tonapi.New()

	params := tonapi.GetTraceParams{TraceID: "49c2837771b6b07b0984ac3d29d57aa92e8ddb85854765fe50b42239e28bfa34"}
	trace, _ := client.GetTrace(context.Background(), params)

	notificationTrace := findRouterTransferNotificationNodes(trace)[0]
	payout := findPayoutForNotification(notificationTrace)

	assert.Equal(t, "0cb1e23eb1c59f2420f16e5fd250cac37f9fb493bb130ddaa571be8fbe256e7e", payout.Transaction.Hash)
}

func Test_findVaultPayoutForNotification(t *testing.T) {
	client, _ := tonapi.New()

	params := tonapi.GetTraceParams{TraceID: "49c2837771b6b07b0984ac3d29d57aa92e8ddb85854765fe50b42239e28bfa34"}
	trace, _ := client.GetTrace(context.Background(), params)

	notificationTrace := findRouterTransferNotificationNodes(trace)[0]
	payout := findVaultPayoutForNotification(notificationTrace)

	assert.Equal(t, "8bcefb3d042c10b4d86b817cc7ad85723c419855ddcac8d43e2e7a2f24cd4bf9", payout.Transaction.Hash)
}

func Test_findNextSwapTraces(t *testing.T) {
	client, _ := tonapi.New()

	params := tonapi.GetTraceParams{TraceID: "05761a40ef710ce6ec6d6b7a22e717701e7cedfc21cd2f8f372a0f627a9981a8"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapTraces := findSwapTraces(trace)

	assert.Equal(t, 1, len(swapTraces))

	assert.Equal(t, "da078ef96bd798a9c635cad5bcbe728f1f163c0ad9d5695a91664a09c09a2493", swapTraces[0].Notification.Transaction.Hash)
	assert.Equal(t, "9f16a7012fdc003475d84286367d8f61656d7acb2c217cbc1a191bbe637d3853", swapTraces[0].Payout.Transaction.Hash)
	assert.Equal(t, "722bc3e2a5a5edc0579090beb9dd4eb572f35103c6eb371742b4a78aa2f3a18d", swapTraces[0].VaultPayout.Transaction.Hash)
	assert.Equal(t, address.MustParseAddr("EQBOdxSlGhjxjVYVwi6blHxfS1tWsOXesMrA1npnuiWeOKgI"), swapTraces[0].Pool)
}

func Test_ExtractStonfiV2SwapsFromRootTrace(t *testing.T) {
	client, _ := tonapi.New()

	params := tonapi.GetTraceParams{TraceID: "093e92969e33af9d162c23724d4581a7c137ceecab5641bb9a1c4b191c34a95a"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapInfos := ExtractStonfiV2SwapsFromRootTrace(trace)

	assert.Equal(t, 1, len(swapInfos))

	swapInfo := swapInfos[0]

	assert.Equal(t, address.MustParseAddr("EQBOdxSlGhjxjVYVwi6blHxfS1tWsOXesMrA1npnuiWeOKgI"), swapInfo.PoolAddress)

	assert.Equal(t, "2e51309e3c0cefca28582f169038c2a0f2b49c7461b9a15beab5077b87394c66", swapInfo.Notification.Hash)
	assert.Equal(t, uint64(50536543000003), swapInfo.Notification.Lt)
	//assert.Equal(t, , swapInfo.Notification.TransactionTime)
	assert.Equal(t, uint64(43265945712075), swapInfo.Notification.QueryId)
	assert.Equal(t, big.NewInt(1500000), swapInfo.Notification.Amount)
	assert.Equal(t, address.MustParseRawAddr("0:2fe5505095f1bd92f308baaf996287ef248bee540269e13cc816164152fdf7e8"), swapInfo.Notification.Sender)
	assert.Equal(t, address.MustParseRawAddr("0:433b0d3c9cb130afd4d35f25c388fb81a201d933fd51938d8ab3c87e15090fdf"), swapInfo.Notification.TokenWallet)
	assert.Equal(t, big.NewInt(309640709), swapInfo.Notification.MinOut)
	assert.Equal(t, address.MustParseRawAddr("0:2fe5505095f1bd92f308baaf996287ef248bee540269e13cc816164152fdf7e8"), swapInfo.Notification.ToAddress)
	assert.Equal(t, address.MustParseRawAddr("0:b089166b7d44a530cd1dcfe1e6e0a5b522a685cf53f0c60dfeb748d00eddfaa8"), swapInfo.Notification.ReferralAddress)

	assert.Equal(t, "0cb1e23eb1c59f2420f16e5fd250cac37f9fb493bb130ddaa571be8fbe256e7e", swapInfo.Payment.Hash)
	assert.Equal(t, uint64(50536543000009), swapInfo.Payment.Lt)
	assert.Equal(t, uint64(43265945712075), swapInfo.Payment.QueryId)
	assert.Equal(t, address.MustParseRawAddr("0:2fe5505095f1bd92f308baaf996287ef248bee540269e13cc816164152fdf7e8"), swapInfo.Payment.Owner)
	assert.Equal(t, uint64(3326308581), swapInfo.Payment.ExitCode)
	assert.Equal(t, big.NewInt(0), swapInfo.Payment.Amount0Out)
	assert.Equal(t, address.MustParseRawAddr("0:40a0fe4e243dc71295bb6ea73491a3a020594c814ce2937219fd1a6fb308a4b5"), swapInfo.Payment.Token0WalletAddress)
	assert.Equal(t, big.NewInt(310885183), swapInfo.Payment.Amount1Out)
	assert.Equal(t, address.MustParseRawAddr("0:433b0d3c9cb130afd4d35f25c388fb81a201d933fd51938d8ab3c87e15090fdf"), swapInfo.Payment.Token1WalletAddress)

	assert.Equal(t, "8bcefb3d042c10b4d86b817cc7ad85723c419855ddcac8d43e2e7a2f24cd4bf9", swapInfo.Referral.Hash)
	assert.Equal(t, uint64(50536543000007), swapInfo.Referral.Lt)
	assert.Equal(t, uint64(43265945712075), swapInfo.Referral.QueryId)
	assert.Equal(t, address.MustParseRawAddr("0:b089166b7d44a530cd1dcfe1e6e0a5b522a685cf53f0c60dfeb748d00eddfaa8"), swapInfo.Referral.Owner)
	assert.Equal(t, uint64(0), swapInfo.Referral.ExitCode)
	assert.Equal(t, big.NewInt(0), swapInfo.Referral.Amount0Out)
	assert.Equal(t, address.MustParseRawAddr("0:40a0fe4e243dc71295bb6ea73491a3a020594c814ce2937219fd1a6fb308a4b5"), swapInfo.Referral.Token0WalletAddress)
	assert.Equal(t, big.NewInt(311509), swapInfo.Referral.Amount1Out)
	assert.Equal(t, address.MustParseRawAddr("0:433b0d3c9cb130afd4d35f25c388fb81a201d933fd51938d8ab3c87e15090fdf"), swapInfo.Referral.Token1WalletAddress)
}

func Test_DoNotParseFailedTransaction(t *testing.T) {
	client, _ := tonapi.New()

	params := tonapi.GetTraceParams{TraceID: "b960db5ada0013fa0d70639e258863d48861e06648f5201b89d7cd34bc6c7442"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapInfos := ExtractStonfiV2SwapsFromRootTrace(trace)

	assert.Equal(t, 0, len(swapInfos))
}

func Test_TransactionWithSmthFailed(t *testing.T) {
	// in fact query id does not fit into int64 - only uint64
	client, _ := tonapi.New()

	params := tonapi.GetTraceParams{TraceID: "fd88effd16246914a578fddf8484ca3f919d94e6534e43ec5abb0674a0ce0c54"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapInfos := ExtractStonfiV2SwapsFromRootTrace(trace)

	assert.Equal(t, 1, len(swapInfos))
}
