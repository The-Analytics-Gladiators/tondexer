package models

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
