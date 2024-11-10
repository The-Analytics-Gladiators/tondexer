package dedust

type InVaultJsonBodyForTon struct {
	QueryID    uint64     `json:"query_id"`
	Amount     string     `json:"amount"`
	Step       Step       `json:"step"`
	SwapParams SwapParams `json:"swap_params"`
}

type InVaultBodyForToken struct {
	QueryID        uint64         `json:"query_id"`
	Amount         string         `json:"amount"`
	Sender         string         `json:"sender"`
	ForwardPayload ForwardPayload `json:"forward_payload"`
}

type Step struct {
	PoolAddr string `json:"pool_addr"`
	Params   Params `json:"params"` // Use a pointer so it can be null
}

type ForwardPayload struct {
	IsRight bool  `json:"is_right"`
	Value   Value `json:"value"`
}

type Value struct {
	SumType string       `json:"sum_type"`
	OpCode  int64        `json:"op_code"`
	Value   ValueDetails `json:"value"`
}

type ValueDetails struct {
	Step       Step       `json:"step"`
	SwapParams SwapParams `json:"swap_params"`
}
