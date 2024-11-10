package dedust

type Params struct {
	KindOut bool   `json:"kind_out"`
	Limit   string `json:"limit"`
}

type SwapParams struct {
	RecipientAddr string `json:"recipient_addr"`
	ReferralAddr  string `json:"referral_addr"`
}

var VaultAddresses = []string{
	"EQDa4VOnTYlLvDJ0gZjNYm5PXfSmmtL6Vs6A_CZEtXCNICq_", // TON
	"EQAYqo4u7VF0fa4DPAebk4g9lBytj2VFny7pzXR0trjtXQaO", // USDT
	"EQDK7GXQoAL5so_kUhe-pjl7GqXM2S_-rJcnEOPXIvjRu5Ii", // DOGS
	"EQAJxbmwhwJw-FomFlt1RYukdeTltiCmWlsurQsQMRJhZfdB", // NOT
	"EQBN18kAXsBfvshv6gYu6rGGfK5kmUEohTSR9Tb6IiT8Zt_n", // HMSTR
	//"EQDT8yu8p0Lw5VLL8lmSUIsg4QKwsBE0ta79YW56RCpnA_-5", // CATI
	"EQBHL09fmrKCG0bSCyu5uP5Q88ubbB02oshzpSio9VpXCFNx", // WALL
	"EQBUoov-LV8Rr-9HzuQFYMGmGyhqHCj9EUFVKXtMHYwLxMWt", // REDO
	"EQBl0I22t0Ca8BroobN5-RmIbnfg3cwe8lg3tHEmR_kifi0S", // CATS
	"EQA7_udkPzrx4FwrrEwlZ9jk5OrQBMJZO1MOGxHJxJGsIaLZ", // FISH
	"EQAPiKEI8uhRcePptx2Kbzq7VvJocpYYUHsubry0mXpu7AXo", // XROCK
	"EQBeWd2_71HcPmAoTX2i9h0HWehA3_G76lxk90yyXmKXuje7", // JETTON
	"EQCXLa2eZNFnvlGx-cpCywps7AcUoowkqCqDS9aNErm_Y3di", // tsTON
	"EQCpvAlOOzqcI_NUxwB0dqw_zd0Sr2rtS0FoM35HaSxTI7-K", // GRAM
	"EQB_0ZmfV8bFhm_J_2tcNvdTuOCGT2i_t4FrTArZhXyxizoW", // durev
	"EQC4UoXs2gCXRuA4JoaDjmWPtN85yunHMM5wlbrmUVo9NMXD", // jUSDT
	"EQCnA5iADgj804vBKzyxVxxgIVUw0SPICA04_-o9bzdHOrI4", // PUNK

	"EQACpR7Dc3393EVHkZ-7pg7zZMB5j7DAh2NNteRzK2wPGqk1", // stTON
	"EQBUmROmkDcMarrpvfj0-iOY0AC769_ykBpvIarLnNisd4jw", // TCAT
	"EQAdOnd5xujz_QmtKLE_PEETNNx8ZREAY6wD5dkji_5jsYrD", // чебурашка
	"EQAf4BMoiqPf0U2ADoNiEatTemiw3UXkt5H90aQpeSKC2l7f", // DUST
}
