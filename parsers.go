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

type PaymentRequest struct {
	QueryId       uint64
	Owner         *address.Address
	ExitCode      uint64
	Amount0Out    uint64
	Token0Address *address.Address
	Amount1Out    uint64
	Token1Address *address.Address
}

func (pr *PaymentRequest) String() string {
	return fmt.Sprintf("PaymentRequest(QueryId: %v, Owner: %v, ExitCode: %v, Amount0Out: %v, Token0Address: %v, Amount1Out: %v, Token1Address: %v)",
		pr.QueryId, pr.Owner, pr.ExitCode, pr.Amount0Out, pr.Token0Address, pr.Amount1Out, pr.Token1Address)
}

func ParsePaymentRequestMessage(message *tlb.InternalMessage) *PaymentRequest {
	cll := message.Body.BeginParse()

	msgCode := cll.MustLoadUInt(32) // Message code
	if msgCode != PaymentRequestCode {
		return nil
	}

	queryId := cll.MustLoadUInt(64)
	owner := cll.MustLoadAddr()
	exitCode := cll.MustLoadUInt(32)
	//cll.MustLoadUInt(32)

	ref := cll.MustLoadRef()
	amount0Out := ref.MustLoadCoins()
	token0Address := ref.MustLoadAddr()
	amount1Out := ref.MustLoadCoins()
	token1Address := ref.MustLoadAddr()

	return &PaymentRequest{
		QueryId:       queryId,
		Owner:         owner,
		ExitCode:      exitCode,
		Amount0Out:    amount0Out,
		Token0Address: token0Address,
		Amount1Out:    amount1Out,
		Token1Address: token1Address,
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

func (tn *SwapTransferNotification) String() string {
	return fmt.Sprintf("SwapTransferNotification(QueryId: %v, Amount: %v, Sender: %v, TokenWallet: %v, "+
		"MinOut: %v, ToAddress: %v, ReferralAddress: %v)",
		tn.QueryId, tn.Amount, tn.Sender, tn.TokenWallet,
		tn.MinOut, tn.ToAddress, tn.ReferralAddress)
}

func ParseSwapTransferNotificationMessage(message *tlb.InternalMessage) *SwapTransferNotification {
	cll := message.Body.BeginParse()

	msgCode := cll.MustLoadUInt(32) // Message code
	if msgCode != TransferNotificationCode {
		return nil
	}

	queryId := cll.MustLoadUInt(64)
	jettonAmount := cll.MustLoadCoins()
	fromUser := cll.MustLoadAddr()

	ref := cll.MustLoadRef()
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
