package main

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/tonkeeper/tonapi-go"
	"math/big"
	"testing"
	"tondexer/core"
	"tondexer/dedust"
	"tondexer/jettons"
	"tondexer/models"
)

var client, _ = tonapi.New()

var tonApiClient, _ = tonapi.New()
var api = core.TonConsoleApi{Client: tonApiClient}
var chainTonApi, _ = jettons.GetTonApi()

func walletCache(wallet string) *models.ChainTokenInfo {

	master, _ := chainTonApi.MasterByWallet(wallet)
	info, _ := api.JettonInfoByMaster(master.String())
	return info
}

func masterCache(master string) *models.ChainTokenInfo {
	info, _ := api.JettonInfoByMaster(master)
	return info
}

func rateCache(s string) *float64 {
	rate, _ := api.JettonRateToUsdByMaster(s)
	return &rate
}

func TestSingleSwap(t *testing.T) {
	//03090d15f57f01a13b32b24cacc87ee82c34bafa5dc2302948449afbcda4cb8e
	client, _ := tonapi.New()
	params := tonapi.GetTraceParams{TraceID: "03090d15f57f01a13b32b24cacc87ee82c34bafa5dc2302948449afbcda4cb8e"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapInfos := dedust.ExtractDedustSwapsFromRootTrace(trace)
	assert.Equal(t, 1, len(swapInfos))

	swapsCh := models.DedustSwapInfoToChSwap(swapInfos[0], walletCache, masterCache, rateCache)
	assert.Equal(t, 1, len(swapsCh))

	swapCh := swapsCh[0]

	assert.Equal(t, "DeDust", swapCh.Dex)
	assert.Equal(t, []string{"5ae1d94b1a111b15ea7a9fcb8f63878e27b270ce43bee70be4553f3c8c344b23"}, swapCh.Hashes)
	assert.Equal(t, uint64(52081446000001), swapCh.Lt)
	assert.Equal(t, "EQC47093oX5Xhb0xuk2lCr2RhS8rj-vul61u4W2UH5ORmG_O", swapCh.JettonIn)
	assert.Equal(t, big.NewInt(13446214171778), swapCh.AmountIn)
	assert.Equal(t, "GRAM", swapCh.JettonInSymbol)
	assert.Equal(t, "Gram", swapCh.JettonInName)
	//assert.Equal(t, , swapCh.JettonInUsdRate)
	assert.Equal(t, uint64(9), swapCh.JettonInDecimals)
	assert.Equal(t, "EQCM3B12QK1e4yZSf8GtBRT0aLMNyEsBc_DhVfRRtOEffLez", swapCh.JettonOut)
	assert.Equal(t, big.NewInt(12144042998), swapCh.AmountOut)
	assert.Equal(t, "pTON", swapCh.JettonOutSymbol)
	assert.Equal(t, "Proxy TON", swapCh.JettonOutName)
	//assert.Equal(t, , swapCh.JettonOutUsdRate)
	assert.Equal(t, uint64(9), swapCh.JettonOutDecimals)
	assert.Equal(t, big.NewInt(11536840848), swapCh.MinAmountOut)
	assert.Equal(t, "EQAZZXXhnoNGCzIlSKYqY4vL-hHqdIAuNQXEgqMKg-CYCs1u", swapCh.PoolAddress)
	assert.Equal(t, "UQDQ7jqqGUsLNDYwTTHo-E14ehHBPv1oVIw3Jam7_7SZBZoX", swapCh.Sender)
	assert.Equal(t, "03090d15f57f01a13b32b24cacc87ee82c34bafa5dc2302948449afbcda4cb8e", swapCh.TraceID)
}

func TestThreeCycleSwap(t *testing.T) {
	client, _ := tonapi.New()
	params := tonapi.GetTraceParams{TraceID: "b3e9443a3d2c4a41863c71943ca3ba6b7e56beeb130918b195248299ff325fa2"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapInfos := dedust.ExtractDedustSwapsFromRootTrace(trace)
	assert.Equal(t, 1, len(swapInfos))

	swapsCh := models.DedustSwapInfoToChSwap(swapInfos[0], walletCache, masterCache, rateCache)
	assert.Equal(t, 3, len(swapsCh))

	assert.Equal(t, "DeDust", swapsCh[0].Dex)
	assert.Equal(t, []string{"ffff827fc3f522485992d132545c80cde7f1177520c5e59a2bf174ce7539c3ce"}, swapsCh[0].Hashes)
	assert.Equal(t, uint64(52083519000003), swapsCh[0].Lt)
	assert.Equal(t, "EQCM3B12QK1e4yZSf8GtBRT0aLMNyEsBc_DhVfRRtOEffLez", swapsCh[0].JettonIn)
	assert.Equal(t, big.NewInt(110575230709), swapsCh[0].AmountIn)
	assert.Equal(t, "pTON", swapsCh[0].JettonInSymbol)
	assert.Equal(t, "Proxy TON", swapsCh[0].JettonInName)
	assert.Equal(t, uint64(9), swapsCh[0].JettonInDecimals)
	assert.Equal(t, "EQDNhy-nxYFgUqzfUzImBEP67JqsyMIcyk2S5_RwNNEYku0k", swapsCh[0].JettonOut)
	assert.Equal(t, big.NewInt(105095392809), swapsCh[0].AmountOut)
	assert.Equal(t, "stTON", swapsCh[0].JettonOutSymbol)
	assert.Equal(t, "Staked TON", swapsCh[0].JettonOutName)
	assert.Equal(t, uint64(9), swapsCh[0].JettonOutDecimals)
	assert.Equal(t, big.NewInt(104885202023), swapsCh[0].MinAmountOut)
	assert.Equal(t, "EQCHFiQM_TTSIiKhUCmWSN4aPSTqxJ4VSBEyDFaZ4izyq95Y", swapsCh[0].PoolAddress)
	assert.Equal(t, "UQDQ7jqqGUsLNDYwTTHo-E14ehHBPv1oVIw3Jam7_7SZBZoX", swapsCh[0].Sender)
	assert.Equal(t, "b3e9443a3d2c4a41863c71943ca3ba6b7e56beeb130918b195248299ff325fa2", swapsCh[0].TraceID)

	assert.Equal(t, "DeDust", swapsCh[1].Dex)
	assert.Equal(t, []string{"68ec26561e71e4d06ea0b7c5d596d62ff79e14112f4a7762d5f006c9c9e0d064"}, swapsCh[1].Hashes)
	assert.Equal(t, uint64(52083519000003), swapsCh[1].Lt)
	assert.Equal(t, "EQDNhy-nxYFgUqzfUzImBEP67JqsyMIcyk2S5_RwNNEYku0k", swapsCh[1].JettonIn)
	assert.Equal(t, big.NewInt(105095392809), swapsCh[1].AmountIn)
	assert.Equal(t, "stTON", swapsCh[1].JettonInSymbol)
	assert.Equal(t, "Staked TON", swapsCh[1].JettonInName)
	assert.Equal(t, uint64(9), swapsCh[1].JettonInDecimals)
	assert.Equal(t, "EQCxE6mUtQJKFnGfaROTKOt1lZbDiiX1kCixRv7Nw2Id_sDs", swapsCh[1].JettonOut)
	assert.Equal(t, big.NewInt(610533973), swapsCh[1].AmountOut)
	assert.Equal(t, "USD₮", swapsCh[1].JettonOutSymbol)
	assert.Equal(t, "Tether USD", swapsCh[1].JettonOutName)
	assert.Equal(t, uint64(6), swapsCh[1].JettonOutDecimals)
	assert.Equal(t, big.NewInt(608213943), swapsCh[1].MinAmountOut)
	assert.Equal(t, "EQCm92zFBkLe_qcFDp7WBvI6JFSDsm4WbDPvZ7xNd7nPL_6M", swapsCh[1].PoolAddress)
	assert.Equal(t, "UQDQ7jqqGUsLNDYwTTHo-E14ehHBPv1oVIw3Jam7_7SZBZoX", swapsCh[1].Sender)
	assert.Equal(t, "b3e9443a3d2c4a41863c71943ca3ba6b7e56beeb130918b195248299ff325fa2", swapsCh[1].TraceID)

	assert.Equal(t, "DeDust", swapsCh[2].Dex)
	assert.Equal(t, []string{"7ce56a411c8ca7f2240aabced235222f36fa370177541f5d441f51ca28e3bf09"}, swapsCh[2].Hashes)
	assert.Equal(t, uint64(52083519000003), swapsCh[2].Lt)
	assert.Equal(t, "EQCxE6mUtQJKFnGfaROTKOt1lZbDiiX1kCixRv7Nw2Id_sDs", swapsCh[2].JettonIn)
	assert.Equal(t, big.NewInt(610533973), swapsCh[2].AmountIn)
	assert.Equal(t, "USD₮", swapsCh[2].JettonInSymbol)
	assert.Equal(t, "Tether USD", swapsCh[2].JettonInName)
	assert.Equal(t, uint64(6), swapsCh[2].JettonInDecimals)
	assert.Equal(t, "EQCM3B12QK1e4yZSf8GtBRT0aLMNyEsBc_DhVfRRtOEffLez", swapsCh[2].JettonOut)
	assert.Equal(t, big.NewInt(110652351574), swapsCh[2].AmountOut)
	assert.Equal(t, "pTON", swapsCh[2].JettonOutSymbol)
	assert.Equal(t, "Proxy TON", swapsCh[2].JettonOutName)
	assert.Equal(t, uint64(9), swapsCh[2].JettonOutDecimals)
	assert.Equal(t, big.NewInt(110021633170), swapsCh[2].MinAmountOut)
	assert.Equal(t, "EQA-X_yo3fzzbDbJ_0bzFWKqtRuZFIRa1sJsveZJ1YpViO3r", swapsCh[2].PoolAddress)
	assert.Equal(t, "UQDQ7jqqGUsLNDYwTTHo-E14ehHBPv1oVIw3Jam7_7SZBZoX", swapsCh[2].Sender)
	assert.Equal(t, "b3e9443a3d2c4a41863c71943ca3ba6b7e56beeb130918b195248299ff325fa2", swapsCh[2].TraceID)
}

func TestThreeCycleWithOneFailed(t *testing.T) {
	client, _ := tonapi.New()
	params := tonapi.GetTraceParams{TraceID: "d5c23c14919b0542928a1fe5e63d71110c82a5f7084e08442b2ba2589d89cc49"}
	trace, _ := client.GetTrace(context.Background(), params)

	swapInfos := dedust.ExtractDedustSwapsFromRootTrace(trace)
	assert.Equal(t, 1, len(swapInfos))

	swapsCh := models.DedustSwapInfoToChSwap(swapInfos[0], walletCache, masterCache, rateCache)
	assert.Equal(t, 2, len(swapsCh))

	assert.Equal(t, "DeDust", swapsCh[0].Dex)
	assert.Equal(t, []string{"9861959c84e36f3810b26197935455d3547a7a09607595b58af491d0edd3f586"}, swapsCh[0].Hashes)
	assert.Equal(t, uint64(52081415000009), swapsCh[0].Lt)
	assert.Equal(t, "EQCM3B12QK1e4yZSf8GtBRT0aLMNyEsBc_DhVfRRtOEffLez", swapsCh[0].JettonIn)
	assert.Equal(t, big.NewInt(12394374999), swapsCh[0].AmountIn)
	assert.Equal(t, "pTON", swapsCh[0].JettonInSymbol)
	assert.Equal(t, "Proxy TON", swapsCh[0].JettonInName)
	assert.Equal(t, uint64(9), swapsCh[0].JettonInDecimals)
	assert.Equal(t, "EQAWpz2_G0NKxlG2VvgFbgZGPt8Y1qe0cGj-4Yw5BfmYR5iF", swapsCh[0].JettonOut)
	assert.Equal(t, big.NewInt(157146238205606), swapsCh[0].AmountOut)
	assert.Equal(t, "MEM", swapsCh[0].JettonOutSymbol)
	assert.Equal(t, "Not Meme", swapsCh[0].JettonOutName)
	assert.Equal(t, uint64(9), swapsCh[0].JettonOutDecimals)
	assert.Equal(t, big.NewInt(156247559844978), swapsCh[0].MinAmountOut)
	assert.Equal(t, "EQApAQzWrHQFReeu92xG_vaWFgL30GEQvfTZca4ZyLeNftrK", swapsCh[0].PoolAddress)
	assert.Equal(t, "UQDQ7jqqGUsLNDYwTTHo-E14ehHBPv1oVIw3Jam7_7SZBZoX", swapsCh[0].Sender)
	assert.Equal(t, "d5c23c14919b0542928a1fe5e63d71110c82a5f7084e08442b2ba2589d89cc49", swapsCh[0].TraceID)

	assert.Equal(t, "DeDust", swapsCh[1].Dex)
	assert.Equal(t, []string{"945222e5cd11e6cc20a0fc98e88a2f2e042446df5a0cbceeb79d1b4ba2440c57"}, swapsCh[1].Hashes)
	assert.Equal(t, uint64(52081415000009), swapsCh[1].Lt)
	assert.Equal(t, "EQAWpz2_G0NKxlG2VvgFbgZGPt8Y1qe0cGj-4Yw5BfmYR5iF", swapsCh[1].JettonIn)
	assert.Equal(t, big.NewInt(157146238205606), swapsCh[1].AmountIn)
	assert.Equal(t, "MEM", swapsCh[1].JettonInSymbol)
	assert.Equal(t, "Not Meme", swapsCh[1].JettonInName)
	assert.Equal(t, uint64(9), swapsCh[1].JettonInDecimals)
	assert.Equal(t, "EQC47093oX5Xhb0xuk2lCr2RhS8rj-vul61u4W2UH5ORmG_O", swapsCh[1].JettonOut)
	assert.Equal(t, big.NewInt(13446214171778), swapsCh[1].AmountOut)
	assert.Equal(t, "GRAM", swapsCh[1].JettonOutSymbol)
	assert.Equal(t, "Gram", swapsCh[1].JettonOutName)
	assert.Equal(t, uint64(9), swapsCh[1].JettonOutDecimals)
	assert.Equal(t, big.NewInt(13438368888668), swapsCh[1].MinAmountOut)
	assert.Equal(t, "EQBmN1koBgN_0mjeKm6q2oh0FLPkHrt_Wo_vr0MAB1q1d_0K", swapsCh[1].PoolAddress)
	assert.Equal(t, "UQDQ7jqqGUsLNDYwTTHo-E14ehHBPv1oVIw3Jam7_7SZBZoX", swapsCh[1].Sender)
	assert.Equal(t, "d5c23c14919b0542928a1fe5e63d71110c82a5f7084e08442b2ba2589d89cc49", swapsCh[1].TraceID)
}
