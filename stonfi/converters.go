package stonfi

import (
	"TonArb/models"
	"log"
	"slices"
)

func ToChModel(re *models.StonfiV1RelatedEvents) *models.SwapCH {
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

	return &models.SwapCH{
		Hashes:          hashes,
		Lt:              re.Notification.Lt,
		Time:            re.Notification.Time,
		TokenIn:         tokenIn,
		AmountIn:        amountIn,
		TokenOut:        tokenOut,
		AmountOut:       amountOut,
		MinAmountOut:    re.Notification.MinOut,
		Sender:          re.Notification.ToAddress.String(),
		ReferralAddress: referralAddress,
		ReferralAmount:  referralAmount,
	}
}
