package jettons

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestNovaToken(t *testing.T) {
	tokenDefinition, _ := JettonDefinitionByWallet("EQCudjvsLIf37Tzw7xdZjr_kEYziTJE5hKVDQIeLkbCLamn3")

	assert.Equal(t, tokenDefinition.Decimals, uint(9))
	assert.Equal(t, "NOVA", tokenDefinition.Symbol)
	assert.Equal(t, "Dragunova", tokenDefinition.Name)
}

func TestUsdToken(t *testing.T) {
	tokenDefinition, e := JettonDefinitionByWallet("EQCI2sZ8zq25yub6rHEY8FwPqV3zbCqS5oasOdljENCjh0bs")
	if e != nil {
		panic(e)
	}

	assert.Equal(t, tokenDefinition.Decimals, uint(6))
	assert.Equal(t, tokenDefinition.Symbol, "USD₮")
	assert.Equal(t, tokenDefinition.Name, "Tether USD")
}

func TestJmnt(t *testing.T) {

	tokenDefinition, e := JettonDefinitionByWallet("EQC1cHQGpDX9tji9VZbRcPue_0J-q4HmMlyldYDCIQNU8_JS")
	if e != nil {
		panic(e)
	}
	assert.Equal(t, uint(18), tokenDefinition.Decimals)
	//assert.Equal(t, "jMNT", tokenDefinition.Symbol) // json url is not resolving
	//assert.Equal(t, "Mantle", tokenDefinition.Name)
}

func TestOffChainToken(t *testing.T) {
	tokenDefinition, e := JettonDefinitionByWallet("EQARULUYsmJq1RiZ-YiH-IJLcAZUVkVff-KBPwEmmaQGH6aC")
	if e != nil {
		panic(e)
	}

	assert.Equal(t, tokenDefinition.Decimals, uint(9))
	assert.Equal(t, tokenDefinition.Symbol, "pTON")
	assert.Equal(t, tokenDefinition.Name, "Proxy TON")
}

func TestOffChainTokenWithOtherTypesInJson(t *testing.T) {
	tokenDefinition, e := JettonDefinitionByWallet("EQBeQi4AixFN8r8-m3Sm3hHOZYCPQP0vl34RLoayeP0GkVm_")
	if e != nil {
		panic(e)
	}

	assert.Equal(t, tokenDefinition.Decimals, uint(9))
	assert.Equal(t, tokenDefinition.Symbol, "APE")
	assert.Equal(t, tokenDefinition.Name, "Ton Apes")
}

func TestGrimReaper(t *testing.T) {
	tokenDefinition, e := JettonDefinitionByWallet("EQC2BKK8-VtvrWVw2EvVZ4ncd-hz7b_HMWEuBy_Nz9aZ7Qpq")
	if e != nil {
		panic(e)
	}

	assert.Equal(t, tokenDefinition.Decimals, uint(9))
	assert.Equal(t, tokenDefinition.Symbol, "GRIM ")
	assert.Equal(t, tokenDefinition.Name, "Grim Reaper")
}

func TestTroopToken(t *testing.T) {

	tokenDefinition, e := JettonDefinitionByWallet("EQApoGZGcC9lfB4yofGcylIINSwcmR1_IF7Zs3f4wSRtysOm")
	if e != nil {
		panic(e)
	}

	assert.Equal(t, tokenDefinition.Decimals, uint(9))
	assert.Equal(t, tokenDefinition.Symbol, "GRIM ")
	assert.Equal(t, tokenDefinition.Name, "Grim Reaper")
}

func TestAll(t *testing.T) {
	all := []string{
		"EQBfuIvgYZlLaejrSNqPQ2m2YzacBdhOQ8tkHa6G9kV6e5st",
		"EQARULUYsmJq1RiZ-YiH-IJLcAZUVkVff-KBPwEmmaQGH6aC",
		"EQBO7JIbnU1WoNlGdgFtScJrObHXkBp-FT5mAz8UagiG9KQR",
		"EQB-KQ0uUCQfDY711JezW8J3XqZVEQCohyY9pEvqhmw5JXMq",
		"EQCJBWME6WFzV7VgY_dle1ntXdosFsJtelm52l-gcQ2ozL0v",
		"EQAVgzVMA56gOSX1XVBV2VnyO31NlBy2HLx8mmD7hrxpwjy_",
		"EQCdcT_DkhRxmwuvXDZWX1PboSLoJGP5cRzx3GXcaZwO4eao",
		"EQBeQi4AixFN8r8-m3Sm3hHOZYCPQP0vl34RLoayeP0GkVm_",
		"EQCHuSJBqmpX3zEnFGDBCcVN_ZiaGuoDL2EH0sZdboh5zkwy",
		"EQC2BKK8-VtvrWVw2EvVZ4ncd-hz7b_HMWEuBy_Nz9aZ7Qpq",
		"EQBtdvw3n5U3p1AMtsmGfM45fYcJgjmjkEb8lFl6oRP1pdYy",
		"EQC1cHQGpDX9tji9VZbRcPue_0J-q4HmMlyldYDCIQNU8_JS",
		"EQBTrdcnXvdyDyKk37sVEaDFjnveqT131bPlWQgEzeLIQtVS",
		"EQApoGZGcC9lfB4yofGcylIINSwcmR1_IF7Zs3f4wSRtysOm",
		"EQD_gLpaqIeUD6QfKhvYK73i2v6GyOTl2etk39dYwDz4ZMGo",
		"EQAn0fI9QmJPcj9Cq68dI-_C5jpex7DhLlS7Amu-ENQmMHg4",
		"EQA3ceeRUrUzbV8yejeuchGrliSdLqBUJj4jujh2kEHgGzRc",
		"EQDoYvNW_oMJjZfVo28i-oc9c-UN0uPT8yr7VLb6GyJT_0Wx",
		"EQAtP6lWEX3UU70qcoIOh4eY2oK5jPYzazCIkPpBkQ7oZu2b",
		"EQA0E2pZ-3LcsNPyif5JQByHKPvKEjWrdZTREmXofFLSofLd",
		"EQAV9YK2MGX390ZvHukr4QUasxNOBdPTpVSs8SNiQohH_sQQ",
		"EQBNnswn6thUP8uRt6zRk0KGxh_qUW6A0HDeUrZBo6Nvajop",
		"EQCV4z92XwW3ZVF_m5cDgPHFrSyIbaXhEvlCJhPrcHxYTiks",
		"EQDWPbvjVbj2Nhfv8UiUbwZ50U29Aw-hMnmsgeALU1rpC0b7",
		"EQBLyVLoZ_V2GW0QdO-9WB0w_sNI1vzESIyFQ7dYOvvvxVe4",
		"EQCsRdQ0Pf-yf9YE_rDhss1Tza5B5QWS4KlkwCfW-PnxLLY5",
		"EQA_4uah8neMWm1HDhg-fSCJTQH3OzooYOfhuAosAarJP_nk",
		"EQBkTakCu-m6D1uh5a-EeEgKdlLQwP2_MkscH3kNBRsSMJSv",
		"EQAUrAcsVikSMtfNk93sEgI1xeXPXiAn9Ju8WqJ25dIk2O19",
		"EQCtnKtOwpmNKKjOd2a2wPQO42Spc5AKBoBB0EHFjLvB1_mD",
		"EQA62J8t4bUHCwkbPvelbvdTgH-oe-uEhD1bat_NRXDQT-6j",
		"EQDoy43vqt59atjOBiUpzsczpoQT7CPqe7LQGywB3Jb0nOVr",
		"EQBAdjC3miSH38aAfqvhT-pf_bVmav7ys8XRZcNvArslq2z4",
		"EQBS-pGT9ifGPquy2Jbo0s2C2vCymuPoSKTigIxOlEXQ1Abj",
		"EQAWA6ttYHPI-5NH84pQYKC0ZZ_06bGpQ_aVuBFjefAHOdVn",
		"EQB9OLalTSd789znUFcvUBVovcLJHeARXiUVjsvQKulWa5PT",
		"EQCcHSj2ac3bBtxjYhHykf8--h5ZOCGiQkcWfEULUmS31iTl",
		"EQA99waI-F2_nB0ac0LGfux5CRB9nCV-wm3I5WKH2_UbVakX",
		"EQBtZ7zTFBphyQ3M5-5Zh4GLlmOmUdMvCvetfCoNDyjxLkIe",
		"EQAg7P41kE7nYucrXLxqGGKJr3Su4nbQULZASt7YojGFz4A-",
		"EQBXXKJSXD1xtF_s4ctcxxUF3DMKQBYDfJ0f73743NPB2rzT",
		"EQB9DN6FxGtRvVKCH4U9a_AAWOtM6LHgO0db6G-g1NYV_Zus",
		"EQBgno13J_5TNHV2qvvY2WDgDN1CggKTz5QuR8XlXHsXMIVI",
		"EQCbXOfowe2ZUVhVF-yhoHNalpu_c3I3mR1NY_4HZi3WtsTX",
		"EQCfmPsO-fy0OZ7FBAudMoioWJkJree0Anb1k2m3q8Piap4f",
		"EQBbBfps5ARlUI16-QiWHYoK4STubuOUAH1mFqfz4QAODwxM",
		"EQBDS96KfuZGg9DeU1QKS-8Bi0qH4HZOdJjZ8NIoVhLZ60-V",
		"EQAnlZHdGmI_fS9UddaU_xCl33MBv8WqRWN1uIJ7pCgbas37",
		"EQDxLHFL-AZIlsjnExSnfvO1KWLt7oYdrTlby0L6dlmJu4ys",
		"EQD3-WgJdOBbTui2nBAIy1Jq1yFTqVP9esGHHcTNZ1TEUVm4",
		"EQDJStlU6l5rpB05eQLNlM2mkzlAIwUoKH40lm0s4NxrQyeD",
		"EQAjYIEexuuxM4kc81Xx0ZjYyP8wRqi1VWNSCY4Awb03tt-0",
		"EQC1fMRcvxjS09EP58YKpFr80hgCQw_9uCYAhDaBBYs7krdH",
		"EQCRxIBkZMHTaXfT_Ax4W240tWwftc9mnggWdinSDBx9Zlay",
		"EQAgCFIap_IbMHSSfvWVaqlHYtPyVNMHtXBtbodfRQw7HGAh",
		"EQCnK9g_M8xsjiu01aVZRc7ZhHt7QSCh7T7EPyhwTu_0ozHl",
		"EQDQsp1oKCWWuH3j0O3JkaTGW_YvybzMsqAF_jfPNDpTepfe",
		"EQDzhyPvHoX3UeNN46sQj_bi44N7nyrRVlYK1wnOc5LVyKtk",
		"EQCJPHa2LIQbC1S7VVyI9GeHgQETlit3Rh2uZ1ibgVo3gbEn",
		"EQDfXgXrizT5aRHx2Y87ToBT0EsMvIlGQa5LsktKRBhDk7b_",
		"EQBxrGD_0O2JpABRF0h8rErndueAmUw8o8tbvDl3QKzac-7W",
		"EQAL2KjJo96afRImx_O2i3Wpr5aDRdv7AiIQolNcmJYTRNRW",
		"EQCz8N5auD98KEDqNQrF_dJnRcz5EGPRyZuzyuZG3g3YfpWS",
		"EQClduQdi2qBwWogiqFgC9ZkP7SKCb8S430a4qtYXMx7AeFU",
	}

	for _, a := range all {
		if tokenDefinition, e := JettonDefinitionByWalletRetry(a, 4); e == nil {
			log.Printf("%v, %v %v %v", a, tokenDefinition.Name, tokenDefinition.Symbol, tokenDefinition.Decimals)
		} else {
			log.Printf("Error %v %v", a, e)
		}
	}
}
