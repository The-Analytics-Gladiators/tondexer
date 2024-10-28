package models

import "time"

type ClickhouseJetton struct {
	Name     string `ch:"name"`
	Symbol   string `ch:"symbol"`
	Master   string `ch:"master"`
	Decimals uint64 `ch:"decimals"`
}

type WalletJetton struct {
	Wallet string `ch:"wallet"`
	Master string `ch:"master"`
}

type JettonRate struct {
	Time     time.Time `ch:"time"`
	Name     string    `ch:"name"`
	Symbol   string    `ch:"symbol"`
	Master   string    `ch:"master"`
	Decimals uint64    `ch:"decimals"`
	Rate     float64   `ch:"rate"`
}
