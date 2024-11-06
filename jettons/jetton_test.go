package jettons

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJmntToken(t *testing.T) {
	tonApi, _ := GetTonApi()

	master, _ := tonApi.MasterByWallet("EQC1cHQGpDX9tji9VZbRcPue_0J-q4HmMlyldYDCIQNU8_JS")
	jetton, _ := JettonInfoByMaster(master.String())

	assert.Equal(t, uint(18), jetton.Decimals)
	assert.Equal(t, "jMNT", jetton.Symbol)
	assert.Equal(t, "Mantle", jetton.Name)
}

func TestLKYToken(t *testing.T) {
	jetton, _ := JettonInfoByMaster("EQCIXQn940RNcOk6GzSorRSiA9WZC9xUz-6lyhl6Ap6na2sh")

	assert.Equal(t, uint(9), jetton.Decimals)
	assert.Equal(t, "wNOT", jetton.Symbol)
	assert.Equal(t, "Shards of Notcoin NFT bond", jetton.Name)
}
