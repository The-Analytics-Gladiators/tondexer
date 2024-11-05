package jettons

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/tonkeeper/tonapi-go"
	"log"
	"testing"
	"tondexer/stonfiv2"
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

func TestTrace(t *testing.T) {
	client, _ := tonapi.New()

	params := tonapi.GetTraceParams{TraceID: "8bcefb3d042c10b4d86b817cc7ad85723c419855ddcac8d43e2e7a2f24cd4bf9"}
	trace, _ := client.GetTrace(context.Background(), params)

	var response stonfiv2.NotificationJsonBody
	err := json.Unmarshal(trace.Children[0].Children[0].Children[0].Transaction.InMsg.Value.DecodedBody, &response)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%v  \n", response)
}
