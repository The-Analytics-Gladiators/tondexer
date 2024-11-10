package stonfiv2

import (
	"encoding/json"
	"github.com/tonkeeper/tonapi-go"
	"github.com/xssnick/tonutils-go/address"
	"log"
	"strconv"
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
	amount0Out, e := strconv.ParseUint(payoutJson.AdditionalInfo.Amount0Out, 10, 64)
	if e != nil {
		log.Printf("error parsing amount0Out for payout %v: %v\n", vaultPayout.Transaction.Hash, e)
	}
	amount1Out, e := strconv.ParseUint(payoutJson.AdditionalInfo.Amount1Out, 10, 64)
	if e != nil {
		log.Printf("error parsing amount1Out for payout %v: %v\n", vaultPayout.Transaction.Hash, e)
	}

	return &models.PayoutRequest{
		Hash:                vaultPayout.Transaction.Hash,
		Lt:                  uint64(vaultPayout.Transaction.Lt),
		TransactionTime:     time.UnixMilli(vaultPayout.Transaction.Utime * 1000),
		EventCatchTime:      time.Now(),
		QueryId:             uint64(payoutJson.QueryID),
		Owner:               address.MustParseRawAddr(payoutJson.Owner),
		ExitCode:            0,
		Amount0Out:          amount0Out,
		Token0WalletAddress: address.MustParseRawAddr(payoutJson.AdditionalInfo.Token0Address),
		Amount1Out:          amount1Out,
		Token1WalletAddress: address.MustParseRawAddr(payoutJson.AdditionalInfo.Token1Address),
	}, nil
}
