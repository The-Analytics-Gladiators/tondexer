package models

import (
	"fmt"
	"github.com/xssnick/tonutils-go/address"
	"math/big"
	"time"
)

type SwapInfo struct {
	Notification *SwapTransferNotification
	Payment      *PayoutRequest
	Referral     *PayoutRequest
	PoolAddress  *address.Address
}

type SwapTransferNotification struct {
	Hash            string
	Lt              uint64
	TransactionTime time.Time
	EventCatchTime  time.Time
	QueryId         uint64
	Amount          *big.Int
	Sender          *address.Address // from_user in contract
	TokenWallet     *address.Address // Router Outer Jetton wallet. In fact isn't used
	MinOut          *big.Int
	ToAddress       *address.Address // also in fact unused
	ReferralAddress *address.Address
}

func (tn *SwapTransferNotification) String() string {
	return fmt.Sprintf("SwapTransferNotification(Lt: %v, TransactionTime: %v, QueryId: %v, Amount: %v, Sender: %v, TokenWallet: %v, "+
		"MinOut: %v, ToAddress: %v, ReferralAddress: %v, Hash: %v)",
		tn.Lt, tn.TransactionTime, tn.QueryId, tn.Amount, tn.Sender, tn.TokenWallet,
		tn.MinOut, tn.ToAddress, tn.ReferralAddress, tn.Hash)
}

type PayoutRequest struct {
	Hash                string
	Lt                  uint64
	TransactionTime     time.Time
	EventCatchTime      time.Time
	QueryId             uint64
	Owner               *address.Address
	ExitCode            uint64
	Amount0Out          *big.Int
	Token0WalletAddress *address.Address
	Amount1Out          *big.Int
	Token1WalletAddress *address.Address
}

func (pr *PayoutRequest) String() string {
	return fmt.Sprintf("PayoutRequest(Lt: %v, TransactionTime: %v, QueryId: %v, Owner: %v, ExitCode: %v, Amount0Out: %v, "+
		"Token0WalletAddress: %v, Amount1Out: %v, Token1WalletAddress: %v, Hash: %v)",
		pr.Lt, pr.TransactionTime, pr.QueryId, pr.Owner, pr.ExitCode, pr.Amount0Out,
		pr.Token0WalletAddress, pr.Amount1Out, pr.Token1WalletAddress, pr.Hash)
}
