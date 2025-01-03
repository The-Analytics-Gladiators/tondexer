package stonfiv2

import (
	"encoding/json"
	"github.com/tonkeeper/tonapi-go"
	"github.com/xssnick/tonutils-go/address"
	"log"
	"math/big"
	"time"
	"tondexer/models"
)

type PayoutJsonBody struct {
	QueryID         uint64         `json:"query_id"`
	ToAddress       string         `json:"to_address"`
	ExcessesAddress string         `json:"excesses_address"`
	OriginalCaller  string         `json:"original_caller"`
	ExitCode        int64          `json:"exit_code"`
	CustomPayload   interface{}    `json:"custom_payload"`
	AdditionalInfo  AdditionalInfo `json:"additional_info"`
}

type AdditionalInfo struct {
	FwdTonAmount  string `json:"fwd_ton_amount"`
	Amount0Out    string `json:"amount0_out"`
	Token0Address string `json:"token0_address"`
	Amount1Out    string `json:"amount1_out"`
	Token1Address string `json:"token1_address"`
}

func parseTracePayout(payout *tonapi.Trace) (*models.PayoutRequest, error) {
	var payoutJson PayoutJsonBody
	if err := json.Unmarshal(payout.Transaction.InMsg.Value.DecodedBody, &payoutJson); err != nil {
		return nil, err
	}

	amount0Out, success := new(big.Int).SetString(payoutJson.AdditionalInfo.Amount0Out, 10)
	if !success {
		log.Printf("error parsing amount0Out '%v' for payout %v\n", payoutJson.AdditionalInfo.Amount0Out, payout.Transaction.Hash)
	}
	amount1Out, success := new(big.Int).SetString(payoutJson.AdditionalInfo.Amount1Out, 10)
	if !success {
		log.Printf("error parsing amount1Out '%v' for payout %v\n", payoutJson.AdditionalInfo.Amount1Out, payout.Transaction.Hash)
	}
	return &models.PayoutRequest{
		Hash:                payout.Transaction.Hash,
		Lt:                  uint64(payout.Transaction.Lt),
		TransactionTime:     time.UnixMilli(payout.Transaction.Utime * 1000),
		EventCatchTime:      time.Now(),
		QueryId:             payoutJson.QueryID,
		Owner:               address.MustParseRawAddr(payoutJson.ToAddress),
		ExitCode:            uint64(payoutJson.ExitCode),
		Amount0Out:          amount0Out,
		Token0WalletAddress: address.MustParseRawAddr(payoutJson.AdditionalInfo.Token0Address),
		Amount1Out:          amount1Out,
		Token1WalletAddress: address.MustParseRawAddr(payoutJson.AdditionalInfo.Token1Address),
	}, nil
}
