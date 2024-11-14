package models

import (
	"math/big"
	"time"
)

type ArbitrageCH struct {
	Sender string    `json:"sender"`
	Time   time.Time `json:"time"`

	AmountIn       *big.Int `json:"amount_in"`
	AmountOut      *big.Int `json:"amount_out"`
	Jetton         string   `json:"jetton"`
	JettonName     string   `json:"jetton_name"`
	JettonSymbol   string   `json:"jetton_symbol"`
	JettonUsdRate  float64  `json:"jetton_usd_rate"`
	JettonDecimals uint64   `json:"jetton_decimals"`

	AmountsPath     []*big.Int `json:"amounts_path"`
	JettonsPath     []string   `json:"jettons_path"`
	JettonNames     []string   `json:"jetton_names"`
	JettonSymbols   []string   `json:"jetton_symbols"`
	JettonUsdRates  []float64  `json:"jetton_usd_rates"`
	JettonsDecimals []uint64   `json:"jettons_decimals"`

	PoolsPath []string `json:"pools_path"`
	TraceIDs  []string `json:"trace_ids"`
	Dexes     []string `json:"dexes"`
}

func SwapsToArbitrage(swap1, swap2 *SwapCH) *ArbitrageCH {
	return &ArbitrageCH{
		Sender:          swap1.Sender,
		Time:            swap1.Time,
		AmountIn:        swap1.AmountIn,
		Jetton:          swap1.JettonIn,
		JettonName:      swap1.JettonInName,
		JettonSymbol:    swap1.JettonInSymbol,
		JettonUsdRate:   swap1.JettonInUsdRate,
		JettonDecimals:  swap1.JettonInDecimals,
		AmountOut:       swap2.AmountOut,
		AmountsPath:     []*big.Int{swap1.AmountIn, swap1.AmountOut, swap1.AmountOut},
		JettonsPath:     []string{swap1.JettonIn, swap1.JettonOut, swap2.JettonOut},
		JettonNames:     []string{swap1.JettonInName, swap1.JettonOutName, swap2.JettonOutName},
		JettonSymbols:   []string{swap1.JettonInSymbol, swap1.JettonOutSymbol, swap2.JettonOutSymbol},
		JettonUsdRates:  []float64{swap1.JettonInUsdRate, swap1.JettonOutUsdRate, swap2.JettonOutUsdRate},
		JettonsDecimals: []uint64{swap1.JettonInDecimals, swap1.JettonOutDecimals, swap2.JettonOutDecimals},
		PoolsPath:       []string{swap1.PoolAddress, swap2.PoolAddress},
		TraceIDs:        []string{swap1.TraceID, swap2.TraceID},
		Dexes:           []string{swap1.Dex, swap2.Dex},
	}
}
