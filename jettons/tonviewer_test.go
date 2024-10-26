package jettons

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMMMToken(t *testing.T) {
	tokenInfo, _ := TokenInfoFromJettonWalletPage("EQB8A-pt27DSbPcbBIvpElgI383iZi3ImznQaR3mgi0wn7dY")

	assert.Equal(t, "MMM", tokenInfo.TokenSymbol)
	assert.Equal(t, "MMM2049", tokenInfo.TokenName)
	assert.Greater(t, tokenInfo.TokenToUsd, float64(0))
}

func TestDogsToken(t *testing.T) {
	tokenInfo, _ := TokenInfoFromJettonWalletPage("EQAwbPGaCGCwLstAeRQvNMWaYC2r-83uEi6jGRcVIxS8sQz4")

	assert.Equal(t, "DOGS", tokenInfo.TokenSymbol)
	assert.Equal(t, "Dogs", tokenInfo.TokenName)
	assert.Greater(t, tokenInfo.TokenToUsd, float64(0))
}

func TestOrbitToken(t *testing.T) {
	tokenInfo, _ := TokenInfoFromJettonWalletPage("EQD3-WgJdOBbTui2nBAIy1Jq1yFTqVP9esGHHcTNZ1TEUVm4")

	assert.Equal(t, "oETH", tokenInfo.TokenSymbol)
	assert.Equal(t, "Orbit Bridge Ton Ethereum", tokenInfo.TokenName)
	assert.GreaterOrEqual(t, tokenInfo.TokenToUsd, float64(0))
}

func TestGrimToken(t *testing.T) {
	tokenInfo, _ := TokenInfoFromJettonWalletPage("EQBRNSoB5gAoC_FV8T3qkLQVGNAa-_qSM-ILxekw7IiJvGVi")

	assert.Equal(t, "GRIM", tokenInfo.TokenSymbol)
	assert.Equal(t, "Grim Reaper", tokenInfo.TokenName)
	assert.GreaterOrEqual(t, tokenInfo.TokenToUsd, float64(0))
}

func TestMasterJUSDToken(t *testing.T) {
	tokenInfo, _ := JettonInfoFromMasterPageRetries("EQBynBO23ywHy_CgarY9NK9FTz0yDsG82PtcbSTQgGoXwiuA", 4)

	assert.Equal(t, "jUSDT", tokenInfo.TokenSymbol)
	assert.Equal(t, "jUSDT", tokenInfo.TokenName)
	assert.GreaterOrEqual(t, tokenInfo.TokenToUsd, 0.9)

	assert.LessOrEqual(t, tokenInfo.TokenToUsd, 1.1)
}

//Empty page is returned by TonViewer
//func TestMasterLKYToken(t *testing.T) {
//	tokenInfo, _ := JettonInfoFromMasterPage("EQCIXQn940RNcOk6GzSorRSiA9WZC9xUz-6lyhl6Ap6na2sh")
//
//	assert.Equal(t, "jUSDT", tokenInfo.TokenSymbol)
//	assert.Equal(t, "jUSDT", tokenInfo.TokenName)
//	assert.GreaterOrEqual(t, tokenInfo.TokenToUsd, 0.9)
//
//	assert.LessOrEqual(t, tokenInfo.TokenToUsd, 1.1)
//}
