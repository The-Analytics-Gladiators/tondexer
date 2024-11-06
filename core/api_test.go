package core

import (
	"github.com/stretchr/testify/assert"
	"github.com/tonkeeper/tonapi-go"
	"testing"
)

func TestJmntToken(t *testing.T) {
	client, _ := tonapi.New()
	consoleApi := TonConsoleApi{Client: client}

	jetton, _ := consoleApi.JettonInfoByMaster("EQAEuikLQVh2lDMrV99nTHqFL_TXEyCEJ1xKMuPT60tfvdps")

	assert.Equal(t, uint64(18), jetton.Decimals)
	assert.Equal(t, "jMNT", jetton.Symbol)
	assert.Equal(t, "Mantle", jetton.Name)
}

func TestLKYToken(t *testing.T) {
	client, _ := tonapi.New()
	consoleApi := TonConsoleApi{Client: client}
	jetton, _ := consoleApi.JettonInfoByMaster("EQCIXQn940RNcOk6GzSorRSiA9WZC9xUz-6lyhl6Ap6na2sh")

	assert.Equal(t, uint64(9), jetton.Decimals)
	assert.Equal(t, "wNOT", jetton.Symbol)
	assert.Equal(t, "Shards of Notcoin NFT bond", jetton.Name)
}

func TestRateOfJetton(t *testing.T) {
	client, _ := tonapi.New()
	consoleApi := TonConsoleApi{Client: client}

	rate, _ := consoleApi.JettonRateToUsdByMaster("EQBynBO23ywHy_CgarY9NK9FTz0yDsG82PtcbSTQgGoXwiuA")

	assert.GreaterOrEqual(t, rate, 0.9)
	assert.LessOrEqual(t, rate, 1.1)
}

func TestRateOfNilJetton(t *testing.T) {

	client, _ := tonapi.New()
	consoleApi := TonConsoleApi{Client: client}

	rate, _ := consoleApi.JettonRateToUsdByMaster("EQA4ewTwGfu7pVQC-ZjI3soFwF6XrXJONjY6MmGzL8pWCDj3")

	println(rate)
}
