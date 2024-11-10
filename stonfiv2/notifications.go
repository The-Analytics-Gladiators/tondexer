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

func parseTraceNotification(notification *tonapi.Trace) (*models.SwapTransferNotification, error) {
	var notificationInfo NotificationJsonBody
	err := json.Unmarshal(notification.Transaction.InMsg.Value.DecodedBody, &notificationInfo)
	if err != nil {
		return nil, err
	}

	amount, e := strconv.ParseUint(notificationInfo.Amount, 10, 64)
	if e != nil {
		log.Printf("error parsing amount for notification %v: %v\n", notification.Transaction.Hash, e)
	}
	minAmount, e := strconv.ParseUint(notificationInfo.ForwardPayload.Value.Value.CrossSwapBody.MinOut, 10, 64)
	if e != nil {
		log.Printf("error parsing minAmount for notification %v: %v\n", notification.Transaction.Hash, e)
	}
	var referralAddress *address.Address
	if notificationInfo.ForwardPayload.Value.Value.CrossSwapBody.RefAddress != "" {
		referralAddress = address.MustParseRawAddr(notificationInfo.ForwardPayload.Value.Value.CrossSwapBody.RefAddress)
	}
	return &models.SwapTransferNotification{
		Hash:            notification.Transaction.Hash,
		Lt:              uint64(notification.Transaction.Lt),
		TransactionTime: time.UnixMilli(notification.Transaction.Utime * 1000),
		EventCatchTime:  time.Now(),
		QueryId:         uint64(notificationInfo.QueryID),
		Amount:          amount,
		Sender:          address.MustParseRawAddr(notificationInfo.Sender),
		TokenWallet:     address.MustParseRawAddr(notificationInfo.ForwardPayload.Value.Value.TokenWallet1),
		MinOut:          minAmount,
		ToAddress:       address.MustParseRawAddr(notificationInfo.ForwardPayload.Value.Value.CrossSwapBody.Receiver),
		ReferralAddress: referralAddress,
	}, nil
}

type NotificationJsonBody struct {
	QueryID        uint64         `json:"query_id"`
	Amount         string         `json:"amount"`
	Sender         string         `json:"sender"`
	ForwardPayload ForwardPayload `json:"forward_payload"`
}

type ForwardPayload struct {
	IsRight bool         `json:"is_right"`
	Value   PayloadValue `json:"value"`
}

type PayloadValue struct {
	SumType string    `json:"sum_type"`
	OpCode  int       `json:"op_code"`
	Value   SwapValue `json:"value"`
}

type SwapValue struct {
	TokenWallet1    string        `json:"token_wallet1"`
	RefundAddress   string        `json:"refund_address"`
	ExcessesAddress string        `json:"excesses_address"`
	TxDeadline      int64         `json:"tx_deadline"`
	CrossSwapBody   CrossSwapBody `json:"cross_swap_body"`
}

type CrossSwapBody struct {
	MinOut   string `json:"min_out"`
	Receiver string `json:"receiver"`
	FwdGas   string `json:"fwd_gas"`
	//CustomPayload interface{} `json:"custom_payload"`
	RefundFwdGas string `json:"refund_fwd_gas"`
	//RefundPayload interface{} `json:"refund_payload"`
	RefFee     int    `json:"ref_fee"`
	RefAddress string `json:"ref_address"`
}
