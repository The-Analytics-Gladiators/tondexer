package models

import (
	"log"
	"math/big"
)

func DedustSwapInfoToChSwap(info *DedustSwapInfo,
	walletToMasterCache func(string) *ChainTokenInfo,
	masterJettonCacheFunc func(string) *ChainTokenInfo,
	rateCache func(string) *float64) []*SwapCH {

	var swapChs []*SwapCH
	for i, poolInfo := range info.PoolsInfo {

		var jettonIn *ChainTokenInfo
		if i == 0 {
			if info.InWalletAddress == nil { //Then it's TON
				jettonIn = masterJettonCacheFunc("EQCM3B12QK1e4yZSf8GtBRT0aLMNyEsBc_DhVfRRtOEffLez")
			} else {
				jettonIn = walletToMasterCache(info.InWalletAddress.String())
			}
		} else {
			if poolInfo.JettonIn != nil {
				jettonIn = masterJettonCacheFunc(poolInfo.JettonIn.String())
			}
		}

		var tokenInUsdRate float64 = 0
		var tokenInSymbol = ""
		var tokenInName = ""
		var jettonMasterIn = ""
		var tokenInDecimals uint64 = 9

		if jettonIn != nil {
			rate := rateCache(jettonIn.JettonAddress)
			if rate != nil {
				tokenInUsdRate = *rate
			}
			tokenInSymbol = jettonIn.Symbol
			tokenInName = jettonIn.Name
			jettonMasterIn = jettonIn.JettonAddress
			tokenInDecimals = jettonIn.Decimals
		}

		amountIn := poolInfo.AmountIn
		var amountOut *big.Int
		if i < len(info.PoolsInfo)-1 {
			amountOut = info.PoolsInfo[i+1].AmountIn
		} else {
			amountOut = info.OutAmount
		}

		var jettonOut *ChainTokenInfo
		if i == len(info.PoolsInfo)-1 {
			if info.OutWalletAddress == nil { //then it's TON
				jettonOut = masterJettonCacheFunc("EQCM3B12QK1e4yZSf8GtBRT0aLMNyEsBc_DhVfRRtOEffLez")
			} else {
				jettonOut = walletToMasterCache(info.OutWalletAddress.String())
			}
		} else {
			jettonOut = masterJettonCacheFunc(info.PoolsInfo[i+1].JettonIn.String())
		}

		var tokenOutUsdRate float64 = 0
		var tokenOutSymbol = ""
		var tokenOutName = ""
		var jettonMasterOut string
		var tokenOutDecimals uint64 = 9

		if jettonOut != nil {
			rate := rateCache(jettonOut.JettonAddress)
			if rate != nil {
				tokenOutUsdRate = *rate
			}
			tokenOutSymbol = jettonOut.Symbol
			tokenOutName = jettonOut.Name
			jettonMasterOut = jettonOut.JettonAddress
			tokenOutDecimals = jettonOut.Decimals
		}
		limit := poolInfo.Limit

		swapChs = append(swapChs, &SwapCH{
			Dex:               DeDust,
			Hashes:            []string{poolInfo.Hash},
			Lt:                poolInfo.Lt,
			Time:              info.Time,
			JettonIn:          jettonMasterIn,
			AmountIn:          amountIn,
			JettonInSymbol:    tokenInSymbol,
			JettonInName:      tokenInName,
			JettonInUsdRate:   tokenInUsdRate,
			JettonInDecimals:  tokenInDecimals,
			JettonOut:         jettonMasterOut,
			AmountOut:         amountOut,
			JettonOutSymbol:   tokenOutSymbol,
			JettonOutName:     tokenOutName,
			JettonOutUsdRate:  tokenOutUsdRate,
			JettonOutDecimals: tokenOutDecimals,
			MinAmountOut:      limit,
			PoolAddress:       poolInfo.Address.String(),
			Sender:            poolInfo.Sender.String(),
			ReferralAddress:   "",
			ReferralAmount:    nil,
			CatchTime:         info.CatchTime,
			TraceID:           info.TraceID,
		})
	}

	return swapChs

}

func ToChSwap(swap *SwapInfo,
	dex string,
	cache func(string) *ChainTokenInfo,
	rateCache func(string) *float64) *SwapCH {

	if swap.Notification == nil {
		return nil
	}

	if swap.Payment == nil && swap.Referral != nil {
		log.Printf("Tx with missing Payment, but with Ref! %v \n", swap.Notification.Hash)
		return nil
	}

	swapPayment := swap.Payment

	hashes := []string{swap.Notification.Hash, swapPayment.Hash}

	var walletIn string
	var amountIn *big.Int
	var walletOut string
	var amountOut *big.Int
	if swapPayment.Amount0Out.Cmp(big.NewInt(0)) == 0 {
		walletIn = swapPayment.Token0WalletAddress.String()
		amountIn = swap.Notification.Amount
		walletOut = swapPayment.Token1WalletAddress.String()
		amountOut = swapPayment.Amount1Out
	} else {
		walletIn = swapPayment.Token1WalletAddress.String()
		amountIn = swap.Notification.Amount
		walletOut = swapPayment.Token0WalletAddress.String()
		amountOut = swapPayment.Amount0Out
	}

	var referralAddress string
	var referralAmount *big.Int

	if swap.Referral != nil {
		referralSwap := swap.Referral

		referralAddress = referralSwap.Owner.String()
		if referralSwap.Amount0Out.Cmp(big.NewInt(0)) == 0 {
			referralAmount = referralSwap.Amount1Out
		} else {
			referralAmount = referralSwap.Amount0Out
		}

		hashes = append(hashes, referralSwap.Hash)
	}

	tokenInInfo := cache(walletIn)

	var tokenInUsdRate float64 = 0
	var tokenInSymbol = ""
	var tokenInName = ""
	var jettonMasterIn = ""
	var tokenInDecimals uint64 = 9

	if tokenInInfo != nil {
		rate := rateCache(tokenInInfo.JettonAddress)
		if rate != nil {
			tokenInUsdRate = *rate
		}
		tokenInSymbol = tokenInInfo.Symbol
		tokenInName = tokenInInfo.Name
		jettonMasterIn = tokenInInfo.JettonAddress
		tokenInDecimals = tokenInInfo.Decimals
	}

	tokenOutInfo := cache(walletOut)

	var tokenOutUsdRate float64 = 0
	var tokenOutSymbol = ""
	var tokenOutName = ""
	var jettonMasterOut string
	var tokenOutDecimals uint64 = 9

	if tokenOutInfo != nil {
		rate := rateCache(tokenOutInfo.JettonAddress)
		if rate != nil {
			tokenOutUsdRate = *rate
		}
		tokenOutSymbol = tokenOutInfo.Symbol
		tokenOutName = tokenOutInfo.Name
		jettonMasterOut = tokenOutInfo.JettonAddress
		tokenOutDecimals = tokenOutInfo.Decimals
	}

	return &SwapCH{
		Dex:               dex,
		Hashes:            hashes,
		Lt:                swap.Notification.Lt,
		Time:              swap.Notification.TransactionTime,
		JettonIn:          jettonMasterIn,
		AmountIn:          amountIn,
		JettonInSymbol:    tokenInSymbol,
		JettonInName:      tokenInName,
		JettonInUsdRate:   tokenInUsdRate,
		JettonInDecimals:  tokenInDecimals,
		JettonOut:         jettonMasterOut,
		AmountOut:         amountOut,
		JettonOutSymbol:   tokenOutSymbol,
		JettonOutName:     tokenOutName,
		JettonOutUsdRate:  tokenOutUsdRate,
		JettonOutDecimals: tokenOutDecimals,
		MinAmountOut:      swap.Notification.MinOut,
		PoolAddress:       swap.PoolAddress.String(),
		Sender:            swap.Notification.Sender.String(),
		ReferralAddress:   referralAddress,
		ReferralAmount:    referralAmount,
		CatchTime:         swap.Notification.EventCatchTime,
		TraceID:           swap.TraceID,
	}
}
