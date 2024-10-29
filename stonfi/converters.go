package stonfi

import (
	"TonArb/jettons"
	"TonArb/models"
	"log"
	"slices"
)

func ToChSwap(swap *StonfiV1Swap,
	cache func(string) *jettons.ChainTokenInfo,
	rateCache func(string) *float64) *models.SwapCH {

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
	var amountIn uint64
	var walletOut string
	var amountOut uint64
	if swapPayment.Amount0Out == 0 {
		walletIn = swapPayment.Token0Address.String()
		amountIn = swap.Notification.Amount
		walletOut = swapPayment.Token1Address.String()
		amountOut = swapPayment.Amount1Out
	} else {
		walletIn = swapPayment.Token1Address.String()
		amountIn = swap.Notification.Amount
		walletOut = swapPayment.Token0Address.String()
		amountOut = swapPayment.Amount0Out
	}

	var referralAddress string
	var referralAmount uint64

	if swap.Referral != nil {
		referralSwap := swap.Referral

		referralAddress = referralSwap.Owner.String()
		if referralSwap.Amount0Out == 0 {
			referralAmount = referralSwap.Amount1Out
		} else {
			referralAmount = referralSwap.Amount0Out
		}

		hashes = append(hashes, referralSwap.Hash)
	}

	tokenInInfo := cache(walletIn)

	var tokenInUsdRate float64 = 0
	var tokenInSymbol string = ""
	var tokenInName string = ""
	var jettonMasterIn string = ""
	var tokenInDecimals uint = 9

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
	var tokenOutSymbol string = ""
	var tokenOutName string = ""
	var jettonMasterOut string
	var tokenOutDecimals uint = 9

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

	return &models.SwapCH{
		Dex:               "StonfiV1",
		Hashes:            hashes,
		Lt:                swap.Notification.Lt,
		Time:              swap.Notification.TransactionTime,
		JettonIn:          jettonMasterIn,
		AmountIn:          amountIn,
		JettonInSymbol:    tokenInSymbol,
		JettonInName:      tokenInName,
		JettonInUsdRate:   tokenInUsdRate,
		JettonInDecimals:  uint64(tokenInDecimals),
		JettonOut:         jettonMasterOut,
		AmountOut:         amountOut,
		JettonOutSymbol:   tokenOutSymbol,
		JettonOutName:     tokenOutName,
		JettonOutUsdRate:  tokenOutUsdRate,
		JettonOutDecimals: uint64(tokenOutDecimals),
		MinAmountOut:      swap.Notification.MinOut,
		Sender:            swap.Notification.ToAddress.String(),
		ReferralAddress:   referralAddress,
		ReferralAmount:    referralAmount,
		CatchTime:         swap.Notification.EventCatchTime,
	}
}

func ToChModel(re *StonfiV1RelatedEvents, cache func(string) *jettons.ChainTokenInfo, rateCache func(string) *float64) *models.SwapCH {
	if re.Notification == nil {
		return nil
	}

	if len(re.Payments) == 1 &&
		re.Notification.ReferralAddress != nil &&
		re.Payments[0].Owner.Equals(re.Notification.ReferralAddress) {
		log.Printf("Tx with missing Payment, but with Ref! %v \n", re.Notification.Hash)
		return nil
	}

	swapPaymentIndex := slices.IndexFunc(re.Payments, func(p *models.PaymentRequest) bool {
		return p.Owner.Equals(re.Notification.ToAddress)
	})
	if swapPaymentIndex == -1 {
		return nil
	}
	swapPayment := re.Payments[swapPaymentIndex]

	hashes := []string{re.Notification.Hash, swapPayment.Hash}

	var walletIn string
	var amountIn uint64
	var walletOut string
	var amountOut uint64
	if swapPayment.Amount0Out == 0 {
		walletIn = swapPayment.Token0Address.String()
		amountIn = re.Notification.Amount
		walletOut = swapPayment.Token1Address.String()
		amountOut = swapPayment.Amount1Out
	} else {
		walletIn = swapPayment.Token1Address.String()
		amountIn = re.Notification.Amount
		walletOut = swapPayment.Token0Address.String()
		amountOut = swapPayment.Amount0Out
	}

	var referralAddress string
	var referralAmount uint64

	if len(re.Payments) == 2 {
		var referralSwapIndex int
		if swapPaymentIndex == 0 {
			referralSwapIndex = 1
		} else {
			referralSwapIndex = 0
		}
		referralSwap := re.Payments[referralSwapIndex]

		referralAddress = referralSwap.Owner.String()
		if referralSwap.Amount0Out == 0 {
			referralAmount = referralSwap.Amount1Out
		} else {
			referralAmount = referralSwap.Amount0Out
		}

		hashes = append(hashes, referralSwap.Hash)
	}

	tokenInInfo := cache(walletIn)

	var tokenInUsdRate float64 = 0
	var tokenInSymbol string = ""
	var tokenInName string = ""
	var jettonMasterIn string = ""
	var tokenInDecimals uint = 9

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
	var tokenOutSymbol string = ""
	var tokenOutName string = ""
	var jettonMasterOut string
	var tokenOutDecimals uint = 9

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

	return &models.SwapCH{
		Dex:               "StonfiV1",
		Hashes:            hashes,
		Lt:                re.Notification.Lt,
		Time:              re.Notification.TransactionTime,
		JettonIn:          jettonMasterIn,
		AmountIn:          amountIn,
		JettonInSymbol:    tokenInSymbol,
		JettonInName:      tokenInName,
		JettonInUsdRate:   tokenInUsdRate,
		JettonInDecimals:  uint64(tokenInDecimals),
		JettonOut:         jettonMasterOut,
		AmountOut:         amountOut,
		JettonOutSymbol:   tokenOutSymbol,
		JettonOutName:     tokenOutName,
		JettonOutUsdRate:  tokenOutUsdRate,
		JettonOutDecimals: uint64(tokenOutDecimals),
		MinAmountOut:      re.Notification.MinOut,
		Sender:            re.Notification.ToAddress.String(),
		ReferralAddress:   referralAddress,
		ReferralAmount:    referralAmount,
		CatchTime:         re.Notification.EventCatchTime,
	}
}
