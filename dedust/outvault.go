package dedust

type OutVaultJsonBody struct {
	QueryID       uint64 `json:"query_id"`
	Proof         string `json:"proof"`
	Amount        string `json:"amount"`
	RecipientAddr string `json:"recipient_addr"`
}
