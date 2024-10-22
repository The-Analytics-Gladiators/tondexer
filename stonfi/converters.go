package stonfi

import (
	"TonArb/models"
	"log"
	"slices"
)

func ToChModel(re *models.StonfiV1RelatedEvents, cache func(string) *models.TokenInfo) *models.SwapCH {
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
	swapPayment := re.Payments[swapPaymentIndex]

	hashes := []string{re.Notification.Hash, swapPayment.Hash}

	var tokenIn string
	var amountIn uint64
	var tokenOut string
	var amountOut uint64
	if swapPayment.Amount0Out == 0 {
		tokenIn = swapPayment.Token0Address.String()
		amountIn = re.Notification.Amount
		tokenOut = swapPayment.Token1Address.String()
		amountOut = swapPayment.Amount1Out
	} else {
		tokenIn = swapPayment.Token1Address.String()
		amountIn = re.Notification.Amount
		tokenOut = swapPayment.Token0Address.String()
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

	tokenInInfo := cache(tokenIn)

	var tokenInUsdRate float64 = 0
	var tokenInSymbol string = ""
	var tokenInName string = ""

	if tokenInInfo != nil {
		tokenInUsdRate = tokenInInfo.TokenToUsd
		tokenInSymbol = tokenInInfo.TokenSymbol
		tokenInName = tokenInInfo.TokenName
	}

	tokenOutInfo := cache(tokenOut)

	var tokenOutUsdRate float64 = 0
	var tokenOutSymbol string = ""
	var tokenOutName string = ""

	if tokenInInfo != nil {
		tokenOutUsdRate = tokenOutInfo.TokenToUsd
		tokenOutSymbol = tokenOutInfo.TokenSymbol
		tokenOutName = tokenOutInfo.TokenName
	}

	return &models.SwapCH{
		Dex:             "StonfiV1",
		Hashes:          hashes,
		Lt:              re.Notification.Lt,
		Time:            re.Notification.Time,
		TokenIn:         tokenIn,
		AmountIn:        amountIn,
		TokenInSymbol:   tokenInSymbol,
		TokenInName:     tokenInName,
		TokenInUsdRate:  tokenInUsdRate,
		TokenOut:        tokenOut,
		AmountOut:       amountOut,
		TokenOutSymbol:  tokenOutSymbol,
		TokenOutName:    tokenOutName,
		TokenOutUsdRate: tokenOutUsdRate,
		MinAmountOut:    re.Notification.MinOut,
		Sender:          re.Notification.ToAddress.String(),
		ReferralAddress: referralAddress,
		ReferralAmount:  referralAmount,
	}
}
