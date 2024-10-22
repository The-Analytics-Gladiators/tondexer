package tonviewer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMMMToken(t *testing.T) {
	tokenInfo, _ := FetchTokenInfo("EQB8A-pt27DSbPcbBIvpElgI383iZi3ImznQaR3mgi0wn7dY")

	assert.Equal(t, "MMM", tokenInfo.TokenSymbol)
	assert.Equal(t, "MMM2049", tokenInfo.TokenName)
	assert.Greater(t, tokenInfo.TokenToUsd, float64(0))
}

func TestDogsToken(t *testing.T) {
	tokenInfo, _ := FetchTokenInfo("EQAwbPGaCGCwLstAeRQvNMWaYC2r-83uEi6jGRcVIxS8sQz4")

	assert.Equal(t, "DOGS", tokenInfo.TokenSymbol)
	assert.Equal(t, "Dogs", tokenInfo.TokenName)
	assert.Greater(t, tokenInfo.TokenToUsd, float64(0))
}

func TestOrbitToken(t *testing.T) {
	tokenInfo, _ := FetchTokenInfo("EQD3-WgJdOBbTui2nBAIy1Jq1yFTqVP9esGHHcTNZ1TEUVm4")

	assert.Equal(t, "oETH", tokenInfo.TokenSymbol)
	assert.Equal(t, "Orbit Bridge Ton Ethereum", tokenInfo.TokenName)
	assert.GreaterOrEqual(t, tokenInfo.TokenToUsd, float64(0))
}

func TestGrimToken(t *testing.T) {
	tokenInfo, _ := FetchTokenInfo("EQBRNSoB5gAoC_FV8T3qkLQVGNAa-_qSM-ILxekw7IiJvGVi")

	assert.Equal(t, "GRIM", tokenInfo.TokenSymbol)
	assert.Equal(t, "Grim Reaper", tokenInfo.TokenName)
	assert.GreaterOrEqual(t, tokenInfo.TokenToUsd, float64(0))
}
