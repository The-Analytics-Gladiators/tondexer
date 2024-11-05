package models

import (
	"fmt"
	"github.com/xssnick/tonutils-go/address"
	"time"
)

type SwapInfo struct {
	Notification *SwapTransferNotification
	Payment      *PayoutRequest
	Referral     *PayoutRequest
	PoolAddress  string
}

type SwapTransferNotification struct {
	Hash            string
	Lt              uint64
	TransactionTime time.Time
	EventCatchTime  time.Time
	QueryId         uint64
	Amount          uint64
	Sender          *address.Address // from_user in contract
	TokenWallet     *address.Address // Router Outer Jetton wallet
	MinOut          uint64
	ToAddress       *address.Address
	ReferralAddress *address.Address
}

func (tn *SwapTransferNotification) String() string {
	return fmt.Sprintf("SwapTransferNotification(Lt: %v, TransactionTime: %v, QueryId: %v, Amount: %v, Sender: %v, TokenWallet: %v, "+
		"MinOut: %v, ToAddress: %v, ReferralAddress: %v, Hash: %v)",
		tn.Lt, tn.TransactionTime, tn.QueryId, tn.Amount, tn.Sender, tn.TokenWallet,
		tn.MinOut, tn.ToAddress, tn.ReferralAddress, tn.Hash)
}

type PayoutRequest struct {
	Hash            string
	Lt              uint64
	TransactionTime time.Time
	EventCatchTime  time.Time
	QueryId         uint64
	Owner           *address.Address
	ExitCode        uint64
	Amount0Out      uint64
	Token0Address   *address.Address
	Amount1Out      uint64
	Token1Address   *address.Address
}

func (pr *PayoutRequest) String() string {
	return fmt.Sprintf("PayoutRequest(Lt: %v, TransactionTime: %v, QueryId: %v, Owner: %v, ExitCode: %v, Amount0Out: %v, "+
		"Token0Address: %v, Amount1Out: %v, Token1Address: %v, Hash: %v)",
		pr.Lt, pr.TransactionTime, pr.QueryId, pr.Owner, pr.ExitCode, pr.Amount0Out,
		pr.Token0Address, pr.Amount1Out, pr.Token1Address, pr.Hash)
}
