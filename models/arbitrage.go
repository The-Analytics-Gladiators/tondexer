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
