package main

import (
	"encoding/hex"
	"fmt"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

const SwapOpCode = 630424929

type SwapTransferNotification struct {
	QueryId         uint64
	Amount          uint64
	Sender          *address.Address // from_user in contract
	TokenWallet     *address.Address
	MinOut          uint64
	ToAddress       *address.Address
	ReferralAddress *address.Address
}

func (tn *SwapTransferNotification) String() string {
	return fmt.Sprintf("QueryId: %v, Amount: %v, Sender: %v, TokenWallet: %v, "+
		"MinOut: %v, ToAddress: %v, ReferralAddress: %v",
		tn.QueryId, tn.Amount, tn.Sender, tn.TokenWallet,
		tn.MinOut, tn.ToAddress, tn.ReferralAddress)
}

func ParseSwapTransferNotificationMessage(message *tlb.InternalMessage) *SwapTransferNotification {
	cell := message.Body.BeginParse()

	msgCode := cell.MustLoadUInt(32) // Message code
	if msgCode != TransferNotificationCode {
		return nil
	}

	queryId := cell.MustLoadUInt(64)
	jettonAmount := cell.MustLoadCoins()
	fromUser := cell.MustLoadAddr()

	ref := cell.MustLoadRef()
	transferredOp := ref.MustLoadUInt(32)
	tokenWallet1 := ref.MustLoadAddr()

	if transferredOp != SwapOpCode {
		return nil
	}

	minOut := ref.MustLoadCoins()
	toAddress := ref.MustLoadAddr()
	hasRef := ref.MustLoadBoolBit()

	var refAddress *address.Address
	if hasRef {
		refAddress = ref.MustLoadAddr()
	}

	return &SwapTransferNotification{
		QueryId:         queryId,
		Amount:          jettonAmount,
		Sender:          fromUser,
		TokenWallet:     tokenWallet1,
		MinOut:          minOut,
		ToAddress:       toAddress,
		ReferralAddress: refAddress,
	}
}

func ParseRawTransaction(transactions string) (*tlb.Transaction, error) {
	hx, _ := hex.DecodeString(transactions)
	cl, _ := cell.FromBOC(hx)

	var tx tlb.Transaction
	if err := tlb.LoadFromCell(&tx, cl.BeginParse()); err != nil {
		return nil, err
	}
	return &tx, nil
}
