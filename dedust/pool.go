package dedust

type PoolJsonBody struct {
	QueryID    uint64     `json:"query_id"`
	Proof      string     `json:"proof"`
	Amount     string     `json:"amount"`
	SenderAddr string     `json:"sender_addr"`
	Current    Params     `json:"current"`
	SwapParams SwapParams `json:"swap_params"`
}
