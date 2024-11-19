package arbitrage

import (
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"tondexer/core"
	"tondexer/jettons"
	"tondexer/models"
)

func FindArbitragesAndDeleteThemFromSetGeneric(swapSet *core.EvictableSet[*models.SwapCH]) []*models.ArbitrageCH {
	all := swapSet.Elements()

	inMap := map[string]*models.SwapCH{}
	for _, model := range all {
		inMap[tokenAmountHash(jettons.Contract(model.JettonIn), model.AmountIn, model.JettonInDecimals)] = model
	}

	var arbitrages []*models.ArbitrageCH
	for _, swap := range all {
		if arbitrage, participatedSwaps := findArbitrageChain(swap, inMap); arbitrage != nil {
			arbitrages = append(arbitrages, arbitrage)
			for _, participatedSwap := range participatedSwaps {
				swapSet.Remove(participatedSwap)
			}
		}
	}

	return arbitrages
}

func findArbitrageChain(firstSwap *models.SwapCH, inMap map[string]*models.SwapCH) (*models.ArbitrageCH, []*models.SwapCH) {
	jettonOut := jettons.Contract(firstSwap.JettonOut)
	amountOut := firstSwap.AmountOut
	decimalsOut := firstSwap.JettonOutDecimals

	swaps := []*models.SwapCH{firstSwap}

	i := 0
	for jettons.Contract(jettonOut) != jettons.Contract(firstSwap.JettonIn) && i <= 10 { // breaking the loop
		if nextSwap, exists := inMap[tokenAmountHash(jettons.Contract(jettonOut), amountOut, decimalsOut)]; exists {
			for _, seenSwap := range swaps {
				if nextSwap.Hashes[0] == seenSwap.Hashes[0] {
					// break the endless loop!
					return nil, []*models.SwapCH{}
				}
			}
			swaps = append(swaps, nextSwap)
			jettonOut = nextSwap.JettonOut
			amountOut = nextSwap.AmountOut
			decimalsOut = nextSwap.JettonOutDecimals
			i++
		} else {
			return nil, []*models.SwapCH{}
		}
	}
	if i >= 10 {
		return nil, []*models.SwapCH{}
	}
	return SwapsToArbitrage(swaps), swaps
}

func tokenAmountHash(token string, amount *big.Int, decimals uint64) string {
	h := sha256.New()
	h.Write([]byte(token))
	exp := big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(decimals-1)), nil)
	modAmount := big.NewInt(0).Div(amount, exp)
	h.Write([]byte(modAmount.String()))
	return hex.EncodeToString(h.Sum(nil))
}

func mapWithFirstArbitrage[T any](swaps []*models.SwapCH, firstMap func(*models.SwapCH) T, secondMap func(*models.SwapCH) T) []T {
	result := []T{firstMap(swaps[0])}
	rest := core.Map(swaps, func(swap *models.SwapCH) T {
		return secondMap(swap)
	})
	return append(result, rest...)
}

func SwapsToArbitrage(swaps []*models.SwapCH) *models.ArbitrageCH {
	return &models.ArbitrageCH{
		Sender:          swaps[0].Sender,
		Time:            swaps[0].Time,
		AmountIn:        swaps[0].AmountIn,
		Jetton:          swaps[0].JettonIn,
		JettonName:      swaps[0].JettonInName,
		JettonSymbol:    swaps[0].JettonInSymbol,
		JettonUsdRate:   swaps[0].JettonInUsdRate,
		JettonDecimals:  swaps[0].JettonInDecimals,
		AmountOut:       swaps[len(swaps)-1].AmountOut,
		AmountsPath:     mapWithFirstArbitrage(swaps, func(swap *models.SwapCH) *big.Int { return swap.AmountIn }, func(swap *models.SwapCH) *big.Int { return swap.AmountOut }),
		JettonsPath:     mapWithFirstArbitrage(swaps, func(swap *models.SwapCH) string { return swap.JettonIn }, func(swap *models.SwapCH) string { return swap.JettonOut }),
		JettonNames:     mapWithFirstArbitrage(swaps, func(swap *models.SwapCH) string { return swap.JettonInName }, func(swap *models.SwapCH) string { return swap.JettonOutName }),
		JettonSymbols:   mapWithFirstArbitrage(swaps, func(swap *models.SwapCH) string { return swap.JettonInSymbol }, func(swap *models.SwapCH) string { return swap.JettonOutSymbol }),
		JettonUsdRates:  mapWithFirstArbitrage(swaps, func(swap *models.SwapCH) float64 { return swap.JettonInUsdRate }, func(swap *models.SwapCH) float64 { return swap.JettonOutUsdRate }),
		JettonsDecimals: mapWithFirstArbitrage(swaps, func(swap *models.SwapCH) uint64 { return swap.JettonInDecimals }, func(swap *models.SwapCH) uint64 { return swap.JettonOutDecimals }),
		PoolsPath:       core.Map(swaps, func(swap *models.SwapCH) string { return swap.PoolAddress }),
		TraceIDs:        core.Map(swaps, func(swap *models.SwapCH) string { return swap.TraceID }),
		Dexes:           core.Map(swaps, func(swap *models.SwapCH) string { return swap.Dex }),
		Senders:         core.Map(swaps, func(swap *models.SwapCH) string { return swap.Sender }),
	}
}
