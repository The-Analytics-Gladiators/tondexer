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

type VaultPayoutJsonBody struct {
	QueryID         uint64                    `json:"query_id"`
	Owner           string                    `json:"owner"`
	ExcessesAddress string                    `json:"excesses_address"`
	AdditionalInfo  VaultPayoutAdditionalInfo `json:"additional_info"`
}

type VaultPayoutAdditionalInfo struct {
	Amount0Out    string `json:"amount0_out"`
	Token0Address string `json:"token0_address"`
	Amount1Out    string `json:"amount1_out"`
	Token1Address string `json:"token1_address"`
}

func parseTraceVaultPayout(vaultPayout *tonapi.Trace) (*models.PayoutRequest, error) {
	var payoutJson VaultPayoutJsonBody
	if err := json.Unmarshal(vaultPayout.Transaction.InMsg.Value.DecodedBody, &payoutJson); err != nil {
		return nil, err
	}
	amount0Out, success := new(big.Int).SetString(payoutJson.AdditionalInfo.Amount0Out, 10)
	if !success {
		log.Printf("error parsing amount0Out '%v' for payout %v\n", payoutJson.AdditionalInfo.Amount0Out, vaultPayout.Transaction.Hash)
	}
	amount1Out, success := new(big.Int).SetString(payoutJson.AdditionalInfo.Amount1Out, 10)
	if !success {
		log.Printf("error parsing amount1Out '%v' for payout %v\n", payoutJson.AdditionalInfo.Amount1Out, vaultPayout.Transaction.Hash)
	}

	return &models.PayoutRequest{
		Hash:                vaultPayout.Transaction.Hash,
		Lt:                  uint64(vaultPayout.Transaction.Lt),
		TransactionTime:     time.UnixMilli(vaultPayout.Transaction.Utime * 1000),
		EventCatchTime:      time.Now(),
		QueryId:             payoutJson.QueryID,
		Owner:               address.MustParseRawAddr(payoutJson.Owner),
		ExitCode:            0,
		Amount0Out:          amount0Out,
		Token0WalletAddress: address.MustParseRawAddr(payoutJson.AdditionalInfo.Token0Address),
		Amount1Out:          amount1Out,
		Token1WalletAddress: address.MustParseRawAddr(payoutJson.AdditionalInfo.Token1Address),
	}, nil
}
