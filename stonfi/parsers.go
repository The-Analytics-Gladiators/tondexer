package stonfi

import (
	"encoding/hex"
	"errors"
	"github.com/tonkeeper/tonapi-go"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
	"time"
	"tondexer/models"
)

const SwapOpCode = 630424929
const SwapOkPaymentCode = 3326308581
const SwapRefPaymentCode = 1158120768

const TransferNotificationCode = 1935855772
const PaymentRequestCode = 4181439551

const StonfiRouter = "EQB3ncyBUTjZUA5EnFKR5_EnOMI9V1tTEAAPaiU71gc4TiUt"

func PaymentRequestFromTrace(trace *tonapi.Trace) (*models.PayoutRequest, error) {
	transaction, e := ParseRawTransaction(trace.Transaction.Raw)
	if e != nil {
		return nil, e
	}
	message := transaction.IO.In.AsInternal()
	cll := message.Body.BeginParse()

	msgCode := cll.MustLoadUInt(32) // Message code
	if msgCode != PaymentRequestCode {
		return nil, errors.New("invalid payment request code")
	}

	queryId := cll.MustLoadUInt(64)
	owner := cll.MustLoadAddr()
	exitCode := cll.MustLoadUInt(32)
	if exitCode != SwapOkPaymentCode && exitCode != SwapRefPaymentCode {
		return nil, errors.New("invalid payment refOrOk request code")
	}

	ref := cll.MustLoadRef()
	amount0Out := ref.MustLoadCoins()
	token0Address := ref.MustLoadAddr()
	amount1Out := ref.MustLoadCoins()
	token1Address := ref.MustLoadAddr()

	return &models.PayoutRequest{
		Hash:                trace.Transaction.Hash,
		Lt:                  message.CreatedLT,
		TransactionTime:     time.UnixMilli(trace.Transaction.Utime * 1000),
		EventCatchTime:      time.Now(),
		QueryId:             queryId,
		Owner:               owner,
		ExitCode:            exitCode,
		Amount0Out:          amount0Out,
		Token0WalletAddress: token0Address,
		Amount1Out:          amount1Out,
		Token1WalletAddress: token1Address,
	}, nil
}

func ParsePaymentRequestMessage(message *tlb.InternalMessage, rawTransactionWithHash *models.RawTransactionWithHash) *models.PayoutRequest {
	cll := message.Body.BeginParse()

	msgCode := cll.MustLoadUInt(32) // Message code
	if msgCode != PaymentRequestCode {
		return nil
	}

	queryId := cll.MustLoadUInt(64)
	owner := cll.MustLoadAddr()
	exitCode := cll.MustLoadUInt(32)
	//cll.MustLoadUInt(32)
	if exitCode != SwapOkPaymentCode && exitCode != SwapRefPaymentCode {
		return nil
	}

	ref := cll.MustLoadRef()
	amount0Out := ref.MustLoadCoins()
	token0Address := ref.MustLoadAddr()
	amount1Out := ref.MustLoadCoins()
	token1Address := ref.MustLoadAddr()

	return &models.PayoutRequest{
		Hash:                rawTransactionWithHash.Hash,
		Lt:                  rawTransactionWithHash.Lt,
		TransactionTime:     rawTransactionWithHash.TransactionTime,
		EventCatchTime:      rawTransactionWithHash.CatchEventTime,
		QueryId:             queryId,
		Owner:               owner,
		ExitCode:            exitCode,
		Amount0Out:          amount0Out,
		Token0WalletAddress: token0Address,
		Amount1Out:          amount1Out,
		Token1WalletAddress: token1Address,
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

func V1NotificationFromTrace(trace *tonapi.Trace) (*models.SwapTransferNotification, error) {
	transaction, e := ParseRawTransaction(trace.Transaction.Raw)
	if e != nil {
		return nil, e
	}
	message := transaction.IO.In.AsInternal()

	cll := message.Body.BeginParse()
	msgCode := cll.MustLoadUInt(32) // Message code
	if msgCode != TransferNotificationCode {
		return nil, errors.New("unknown transfer notification code")
	}

	queryId := cll.MustLoadUInt(64)
	jettonAmount := cll.MustLoadCoins()
	fromUser := cll.MustLoadAddr()

	ref := cll.MustLoadRef()
	transferredOp := ref.MustLoadUInt(32)
	tokenWallet1 := ref.MustLoadAddr()

	if transferredOp != SwapOpCode {
		return nil, errors.New("unknown swap code for notification")
	}

	minOut := ref.MustLoadCoins()
	toAddress := ref.MustLoadAddr()
	hasRef := ref.MustLoadBoolBit()

	var refAddress *address.Address
	if hasRef {
		refAddress = ref.MustLoadAddr()
	}
	return &models.SwapTransferNotification{
		Hash: trace.Transaction.Hash,
		//Lt:              uint64(trace.Transaction.Lt),
		Lt:              message.CreatedLT,
		TransactionTime: time.UnixMilli(trace.Transaction.Utime * 1000),
		EventCatchTime:  time.Now(),
		QueryId:         queryId,
		Amount:          jettonAmount,
		Sender:          fromUser,
		TokenWallet:     tokenWallet1,
		MinOut:          minOut,
		ToAddress:       toAddress,
		ReferralAddress: refAddress,
	}, nil
}

func ParseSwapTransferNotificationMessage(message *tlb.InternalMessage, rawTransactionWithHash *models.RawTransactionWithHash) *models.SwapTransferNotification {
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

	return &models.SwapTransferNotification{
		Hash:            rawTransactionWithHash.Hash,
		Lt:              rawTransactionWithHash.Lt,
		TransactionTime: rawTransactionWithHash.TransactionTime,
		EventCatchTime:  rawTransactionWithHash.CatchEventTime,
		QueryId:         queryId,
		Amount:          jettonAmount,
		Sender:          fromUser,
		TokenWallet:     tokenWallet1,
		MinOut:          minOut,
		ToAddress:       toAddress,
		ReferralAddress: refAddress,
	}
}
