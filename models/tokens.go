package models

type TonviewerTokenInfo struct {
	TokenSymbol string
	TokenName   string
	TokenToUsd  float64
}

type WalletJettonKey struct {
	Dex    string
	Wallet string
}
