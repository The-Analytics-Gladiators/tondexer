package stonfi

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/tonkeeper/tonapi-go"
	"github.com/xssnick/tonutils-go/address"
	"math/big"
	"slices"
	"testing"
	"tondexer/common"
	"tondexer/models"
)

func TestFindRouterNotificationNodes(t *testing.T) {
	client, _ := tonapi.New()

	// Regular swap without ref
	params := tonapi.GetTraceParams{TraceID: "a69e0f2a244c7807ba6e8ecfe3da1ebac48b080782d828166fd0db5ee27d1238"}
	trace, _ := client.GetTrace(context.Background(), params)
	notifications := findRouterTransferNotificationNodes(trace)

	assert.Equal(t, 1, len(notifications))
	assert.Equal(t, "0d42e08d83e30eaec5455425f032b79aa54f477af5310081f732be3112d5d70d", notifications[0].Transaction.Hash)

	// 2 parallel swaps without refs
	params = tonapi.GetTraceParams{TraceID: "1eb52591ace42a6c364b436cfc08a018c6c45684daf888936340af6145c712cc"}
	trace, _ = client.GetTrace(context.Background(), params)
	notifications = findRouterTransferNotificationNodes(trace)

	assert.Equal(t, 2, len(notifications))

	hashes := []string{"7abb04587405a0bc3047c8012653a2f9a2683c9a2de098835263fa7b1b2b9624", "3648ab7b96037d9983ef957ba019d6fdd4d5ba64fa95f790d550d49b6b4e65c3"}
	assert.True(t, common.Contains(hashes, notifications[0].Transaction.Hash))
	assert.True(t, common.Contains(hashes, notifications[1].Transaction.Hash))
}

func TestFindPaymentForNotification(t *testing.T) {
	client, _ := tonapi.New()

	// Regular swap without ref
	params := tonapi.GetTraceParams{TraceID: "a69e0f2a244c7807ba6e8ecfe3da1ebac48b080782d828166fd0db5ee27d1238"}
	trace, _ := client.GetTrace(context.Background(), params)
	notification := findRouterTransferNotificationNodes(trace)[0]

	payments := findPaymentsForNotification(notification)

	assert.Equal(t, 1, len(payments))
	assert.Equal(t, "a70f3c5d8a09f2414f55af769c37cbfdd8304e68a22d63ab04df9af5d14bda4f", payments[0].Transaction.Hash)

	// regular swap with ref
	params = tonapi.GetTraceParams{TraceID: "c666610443281d4395e101fdc70f157ea9fd907da38c1350c3f57d28a492ac97"}
	trace, _ = client.GetTrace(context.Background(), params)
	notification = findRouterTransferNotificationNodes(trace)[0]

	payments = findPaymentsForNotification(notification)

	assert.Equal(t, 2, len(payments))

	hashes := []string{"7bed7e6f09a8dc73b5a05f0ab2004ce18c76d49f2c1c63d07a767a7ab1b3871c", "c666610443281d4395e101fdc70f157ea9fd907da38c1350c3f57d28a492ac97"}
	assert.True(t, common.Contains(hashes, payments[0].Transaction.Hash))
	assert.True(t, common.Contains(hashes, payments[1].Transaction.Hash))
}

func TestFindPoolAddressForNotification(t *testing.T) {
	client, _ := tonapi.New()

	params := tonapi.GetTraceParams{TraceID: "c666610443281d4395e101fdc70f157ea9fd907da38c1350c3f57d28a492ac97"}
	trace, _ := client.GetTrace(context.Background(), params)
	notification := findRouterTransferNotificationNodes(trace)[0]

	addr := findPoolAddressForNotification(notification)
	assert.Equal(t, address.MustParseAddr("EQD7AY9ov7urTou069yON36pH0Okd-bpd1hxYxA5-Sp54Cza"), addr)
}

func TestExtractSwapsFromRegularSwapWithoutRef(t *testing.T) {
	client, _ := tonapi.New()

	params := tonapi.GetTraceParams{TraceID: "a69e0f2a244c7807ba6e8ecfe3da1ebac48b080782d828166fd0db5ee27d1238"}
	trace, _ := client.GetTrace(context.Background(), params)

	swaps := ExtractStonfiSwapsFromRootTrace(trace)

	assert.Equal(t, 1, len(swaps))
	swap := swaps[0]

	assert.Nil(t, swap.Referral)

	assert.Equal(t, "0d42e08d83e30eaec5455425f032b79aa54f477af5310081f732be3112d5d70d", swap.Notification.Hash)
	assert.Equal(t, "a70f3c5d8a09f2414f55af769c37cbfdd8304e68a22d63ab04df9af5d14bda4f", swap.Payment.Hash)
	assert.Equal(t, address.MustParseAddr("EQAZFS5dJ8STrKcn5VnptYsoKILXbAYaDhdJJjbzUrNkDdH_"), swap.PoolAddress)
}

func TestExtractSwapsFromRegularSwapWithRef(t *testing.T) {
	client, _ := tonapi.New()

	params := tonapi.GetTraceParams{TraceID: "179156c9a79fe218d5cccae289f84762f16d080198221d2103049566e176a17e"}
	trace, _ := client.GetTrace(context.Background(), params)

	swaps := ExtractStonfiSwapsFromRootTrace(trace)

	assert.Equal(t, 1, len(swaps))
	swap := swaps[0]

	assert.Equal(t, "179156c9a79fe218d5cccae289f84762f16d080198221d2103049566e176a17e", swap.Notification.Hash)
	assert.Equal(t, "0cefafbf33b0562f1ddc55f9c659e7579c6a423b082959d665aca718923694f6", swap.Payment.Hash)
	assert.Equal(t, "b09e51b3a7593cf0764305752782d81c13b68971317f6319cfde5609ad342bcb", swap.Referral.Hash)
	assert.Equal(t, address.MustParseAddr("EQBJ_X3ysvgOGUo6XB3eUTCvagarGeA3X-QD3lxSqZzQbQ4w"), swap.PoolAddress)
}

func TestExtractSwapsFromParallelSwaps(t *testing.T) {
	client, _ := tonapi.New()

	params := tonapi.GetTraceParams{TraceID: "3648ab7b96037d9983ef957ba019d6fdd4d5ba64fa95f790d550d49b6b4e65c3"}
	trace, _ := client.GetTrace(context.Background(), params)

	swaps := ExtractStonfiSwapsFromRootTrace(trace)

	assert.Equal(t, 2, len(swaps))

	firstSwapIndex := slices.IndexFunc(swaps, func(swap *models.SwapInfo) bool {
		return swap.Notification.Hash == "7abb04587405a0bc3047c8012653a2f9a2683c9a2de098835263fa7b1b2b9624"
	})
	firstSwap := swaps[firstSwapIndex]

	secondSwapIndex := slices.IndexFunc(swaps, func(swap *models.SwapInfo) bool {
		return swap.Notification.Hash == "3648ab7b96037d9983ef957ba019d6fdd4d5ba64fa95f790d550d49b6b4e65c3"
	})
	secondSwap := swaps[secondSwapIndex]

	assert.Equal(t, "7abb04587405a0bc3047c8012653a2f9a2683c9a2de098835263fa7b1b2b9624", firstSwap.Notification.Hash)
	assert.Equal(t, "dc25076fcbef1bfa39cb95e0106bac59aff657a3edf226d706ccf5ebe6fddad4", firstSwap.Payment.Hash)
	assert.Equal(t, address.MustParseAddr("EQCaY8Ifl2S6lRBMBJeY35LIuMXPc8JfItWG4tl7lBGrSoR2"), firstSwap.PoolAddress)
	assert.Nil(t, firstSwap.Referral)

	assert.Equal(t, "3648ab7b96037d9983ef957ba019d6fdd4d5ba64fa95f790d550d49b6b4e65c3", secondSwap.Notification.Hash)
	assert.Equal(t, "1eb52591ace42a6c364b436cfc08a018c6c45684daf888936340af6145c712cc", secondSwap.Payment.Hash)
	assert.Equal(t, address.MustParseAddr("EQCaY8Ifl2S6lRBMBJeY35LIuMXPc8JfItWG4tl7lBGrSoR2"), secondSwap.PoolAddress)
	assert.Nil(t, secondSwap.Referral)
}

func TestExtractSwapFromConsequentSwaps(t *testing.T) {
	client, _ := tonapi.New()

	params := tonapi.GetTraceParams{TraceID: "d2668f071f74f70493e18d3957f3d260dc7d6eeee68e1759a6f43af1aad11e85"}
	trace, _ := client.GetTrace(context.Background(), params)

	swaps := ExtractStonfiSwapsFromRootTrace(trace)

	assert.Equal(t, 2, len(swaps))

	firstSwapIndex := slices.IndexFunc(swaps, func(swap *models.SwapInfo) bool {
		return swap.Notification.Hash == "aa960e037a9b42b20dfb9b41344dae38365f7a90488d7e24f0914724b251011a"
	})
	firstSwap := swaps[firstSwapIndex]

	secondSwapIndex := slices.IndexFunc(swaps, func(swap *models.SwapInfo) bool {
		return swap.Notification.Hash == "d2668f071f74f70493e18d3957f3d260dc7d6eeee68e1759a6f43af1aad11e85"
	})
	secondSwap := swaps[secondSwapIndex]

	assert.Equal(t, "aa960e037a9b42b20dfb9b41344dae38365f7a90488d7e24f0914724b251011a", firstSwap.Notification.Hash)
	assert.Equal(t, "02f6fd022dbcedeca64b00d9ccb14e13e13774751de9be9112c9618c64aa6ec8", firstSwap.Payment.Hash)
	assert.Equal(t, address.MustParseAddr("EQBJ_X3ysvgOGUo6XB3eUTCvagarGeA3X-QD3lxSqZzQbQ4w"), firstSwap.PoolAddress)
	assert.Nil(t, firstSwap.Referral)

	assert.Equal(t, "d2668f071f74f70493e18d3957f3d260dc7d6eeee68e1759a6f43af1aad11e85", secondSwap.Notification.Hash)
	assert.Equal(t, "576f98c81f9f2811218584643d7df7db12a0fec401333c1dc0047058b042f955", secondSwap.Payment.Hash)
	assert.Equal(t, address.MustParseAddr("EQCnauv4pL2eF7xRqsJAzRMSeyAedsyKJ1Qi0JUcq9MXi-MN"), secondSwap.PoolAddress)
	assert.Nil(t, secondSwap.Referral)
}

func TestExtractSwapsFromRegularSwapWithRefSameAsSender(t *testing.T) {
	client, _ := tonapi.New()

	params := tonapi.GetTraceParams{TraceID: "f020893c2b8ac55b477211ad0be0eae87ef106092cfaa76b2e6448c140b612b5"}
	trace, _ := client.GetTrace(context.Background(), params)

	swaps := ExtractStonfiSwapsFromRootTrace(trace)

	assert.Equal(t, 1, len(swaps))
	swap := swaps[0]

	assert.Equal(t, "f020893c2b8ac55b477211ad0be0eae87ef106092cfaa76b2e6448c140b612b5", swap.Notification.Hash)
	assert.Equal(t, "3377b9fb2798bdaa039a0fe2a11cd9f7ce556820b72d7c40bf72f45f20bdf54c", swap.Payment.Hash)
	assert.Equal(t, big.NewInt(735041413167), swap.Payment.Amount0Out)
	assert.Equal(t, "de85e0c15b235677d5d1fc95c257c0b3921d0f170e04ff4d33831f089ce51f02", swap.Referral.Hash)
	assert.Equal(t, big.NewInt(736514443), swap.Referral.Amount0Out)
	assert.Equal(t, address.MustParseAddr("EQBCwe_IObXA4Mt3RbcHil2s4-v4YQS3wUDt1-DvZOceeMGO"), swap.PoolAddress)
}

func TestTraceIDOfSwapInfo(t *testing.T) {
	client, _ := tonapi.New()

	params := tonapi.GetTraceParams{TraceID: "f020893c2b8ac55b477211ad0be0eae87ef106092cfaa76b2e6448c140b612b5"}
	trace, _ := client.GetTrace(context.Background(), params)

	swaps := ExtractStonfiSwapsFromRootTrace(trace)

	assert.Equal(t, 1, len(swaps))
	swap := swaps[0]

	assert.Equal(t, "436a2721b414b982cdea4cde35d633550876cf3836f2e8cefec3e92d520467a7", swap.TraceID)
}
