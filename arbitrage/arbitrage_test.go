package arbitrage

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/tonkeeper/tonapi-go"
	"math/big"
	"testing"
	"time"
	"tondexer/core"
	"tondexer/dedust"
	"tondexer/jettons"
	"tondexer/models"
	"tondexer/stonfi"
	"tondexer/stonfiv2"
)

//a531ac5b5b44df7ee5acc11f514338f3ae12b3a60db5ebed51f008b891c7f918','8535c00aab48da6e2b926c2c26821aab259b69837426b1306cc2946a603ca659

var client, _ = tonapi.New()

var tonApiClient, _ = tonapi.New()
var api = core.TonConsoleApi{Client: tonApiClient}
var chainTonApi, _ = jettons.GetTonApi()

func infoCache(wallet string) *models.ChainTokenInfo {

	master, _ := chainTonApi.MasterByWallet(wallet)
	info, _ := api.JettonInfoByMaster(master.String())
	return info
}
func rateCache(s string) *float64 {
	rate, _ := api.JettonRateToUsdByMaster(s)
	return &rate
}

func TestTwoCycle(t *testing.T) {
	params := tonapi.GetTraceParams{TraceID: "a531ac5b5b44df7ee5acc11f514338f3ae12b3a60db5ebed51f008b891c7f918"}
	trace, _ := client.GetTrace(context.Background(), params)
	stonfiSwapInfo := stonfi.ExtractStonfiSwapsFromRootTrace(trace)[0]
	stonfiChModel1 := models.ToChSwap(stonfiSwapInfo, "StonfiV1",
		infoCache, rateCache)

	params = tonapi.GetTraceParams{TraceID: "8535c00aab48da6e2b926c2c26821aab259b69837426b1306cc2946a603ca659"}
	trace, _ = client.GetTrace(context.Background(), params)

	trace, _ = client.GetTrace(context.Background(), params)
	stonfiSwapInfo = stonfi.ExtractStonfiSwapsFromRootTrace(trace)[0]
	stonfiChModel2 := models.ToChSwap(stonfiSwapInfo, "StonfiV1",
		infoCache, rateCache)

	set := core.NewEvictableSet[*models.SwapCH](time.Second * 0)
	set.Add(stonfiChModel1)
	set.Add(stonfiChModel2)

	arbitrages := FindArbitragesAndDeleteThemFromSetGeneric(set)
	allElements := set.Evict()
	assert.Equal(t, 0, len(allElements))
	assert.Equal(t, 1, len(arbitrages))

	//arbitrage := models.TwoSwapsToArbitrage(stonfiChModel1, stonfiChModel2)
	//
	//assert.Equal(t, "EQCM3B12QK1e4yZSf8GtBRT0aLMNyEsBc_DhVfRRtOEffLez", arbitrage.Jetton)
	// TODO other asserts
}

func TestThreeCycle(t *testing.T) {
	params := tonapi.GetTraceParams{TraceID: "d3951ae5db30660e5a23d22d485ad238041a227da025f2c0328f466634acacbd"}
	trace, _ := client.GetTrace(context.Background(), params)

	stonfiSwapInfo := stonfi.ExtractStonfiSwapsFromRootTrace(trace)[0]
	stonfiChModel := models.ToChSwap(stonfiSwapInfo, "StonfiV1",
		infoCache, rateCache)

	stonfi2SwapInfo := stonfiv2.ExtractStonfiV2SwapsFromRootTrace(trace)[0]
	stonfi2ChModel := models.ToChSwap(stonfi2SwapInfo, "StonfiV2",
		infoCache, rateCache)

	dedustSwapInfo := dedust.ExtractDedustSwapsFromRootTrace(trace)[0]
	dedustChModel := models.ToChSwap(dedustSwapInfo, "DeDust",
		infoCache, rateCache)

	set := core.NewEvictableSet[*models.SwapCH](time.Second * 0)
	set.Add(stonfiChModel)
	set.Add(stonfi2ChModel)
	set.Add(dedustChModel)

	arbitrages := FindArbitragesAndDeleteThemFromSetGeneric(set)

	allElements := set.Evict()
	assert.Equal(t, 0, len(allElements))
	assert.Equal(t, 1, len(arbitrages))

	arbitrage := arbitrages[0]

	assert.Equal(t, "EQA6UVoybsI7mFQQaqMLMVmQovCGBGx0rOUuyf2Q2GfGmvCN", arbitrage.Sender)
	assert.Equal(t, big.NewInt(3698350804), arbitrage.AmountIn)
	assert.Equal(t, big.NewInt(4836589584), arbitrage.AmountOut)
	assert.Equal(t, "EQCM3B12QK1e4yZSf8GtBRT0aLMNyEsBc_DhVfRRtOEffLez", arbitrage.Jetton)
	assert.Equal(t, "Proxy TON", arbitrage.JettonName)
	assert.Equal(t, "pTON", arbitrage.JettonSymbol)
	//assert.Equal(t, , arbitrage.JettonUsdRate)
	assert.Equal(t, uint64(9), arbitrage.JettonDecimals)
	//TODO !
	//assert.Equal(t, []*big.Int{}, arbitrage.AmountsPath)
	//assert.Equal(t, , arbitrage.JettonsPath)
	//assert.Equal(t, , arbitrage.JettonNames)
	//assert.Equal(t, , arbitrage.JettonSymbols)
	//assert.Equal(t, , arbitrage.JettonUsdRates)
	//assert.Equal(t, , arbitrage.JettonsDecimals)
	//
	//assert.Equal(t, , arbitrage.PoolsPath)
	assert.Equal(t, []string{"600bcbbca9c84ab81306c85e4ae3c67ff2b593435ed974da5b0e3dcbe22b9f52",
		"600bcbbca9c84ab81306c85e4ae3c67ff2b593435ed974da5b0e3dcbe22b9f52",
		"600bcbbca9c84ab81306c85e4ae3c67ff2b593435ed974da5b0e3dcbe22b9f52"}, arbitrage.TraceIDs)
	assert.Equal(t, []string{"StonfiV1", "DeDust", "StonfiV2"}, arbitrage.Dexes)

}
