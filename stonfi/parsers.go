package stonfi

import (
	"TonArb/models"
	"encoding/hex"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
	"log"
)

const SwapOpCode = 630424929
const SwapOkPaymentCode = 3326308581
const SwapRefPaymentCode = 1158120768

const TransferNotificationCode = 1935855772
const PaymentRequestCode = 4181439551

const StonfiRouter = "EQB3ncyBUTjZUA5EnFKR5_EnOMI9V1tTEAAPaiU71gc4TiUt"
const StonfiRouterV2 = "EQBCl1JANkTpMpJ9N3lZktPMpp2btRe2vVwHon0la8ibRied"

func ParsePaymentRequestMessage(message *tlb.InternalMessage, rawTransactionWithHash *models.RawTransactionWithHash) *models.PaymentRequest {
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

	return &models.PaymentRequest{
		Hash:          rawTransactionWithHash.Hash,
		Lt:            rawTransactionWithHash.Lt,
		Time:          rawTransactionWithHash.Time,
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

	if !fromUser.Equals(toAddress) {
		log.Printf("!!!!! Different From and Sender %v \n", rawTransactionWithHash.Hash)
	}

	return &models.SwapTransferNotification{
		Hash:            rawTransactionWithHash.Hash,
		Lt:              rawTransactionWithHash.Lt,
		Time:            rawTransactionWithHash.Time,
		QueryId:         queryId,
		Amount:          jettonAmount,
		Sender:          fromUser,
		TokenWallet:     tokenWallet1,
		MinOut:          minOut,
		ToAddress:       toAddress,
		ReferralAddress: refAddress,
	}
}
