package dedust

type PoolJsonBody struct {
	QueryID    uint64     `json:"query_id"`
	Proof      string     `json:"proof"`
	Asset      *Asset     `json:"asset"`
	Amount     string     `json:"amount"`
	SenderAddr string     `json:"sender_addr"`
	Current    Params     `json:"current"`
	SwapParams SwapParams `json:"swap_params"`
}

type Asset struct {
	SumType string  `json:"sum_type"`
	Jetton  *Jetton `json:"jetton"`
}

type Jetton struct {
	WorkchainId int    `json:"workchain_id"`
	Address     string `json:"address"`
}
