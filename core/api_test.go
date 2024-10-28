package core

import (
	"github.com/stretchr/testify/assert"
	"github.com/tonkeeper/tonapi-go"
	"testing"
)

func TestAllTransactionsFromTrace(t *testing.T) {
	client, _ := tonapi.New()
	tonClient := &TonClient{Client: client}

	transactions, _ := tonClient.FetchTransactionsFromTraceByTransactionHash("16dffefb014aca7bffd37a3ca20a723b8e3dc730b1903ddbdc8e63c0442c1c8c")

	assert.Equal(t, 12, len(transactions))
}
