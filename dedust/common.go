package dedust

type Params struct {
	KindOut bool   `json:"kind_out"`
	Limit   string `json:"limit"`
}

type SwapParams struct {
	RecipientAddr string `json:"recipient_addr"`
	ReferralAddr  string `json:"referral_addr"`
}
