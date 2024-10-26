package jettons

import "TonArb/models"

func JettonInfoByMaster(master string) (*ChainTokenInfo, error) {
	chainTokenInfo, e := TokenDefinitionByMasterRetries(master, 4)
	if e != nil {
		return nil, e
	}

	if chainTokenInfo.Name != "" && chainTokenInfo.Symbol != "" {
		return chainTokenInfo, nil
	}

	tonviewerTokenInfo, er := JettonInfoFromMasterPageRetries(master, 4)
	if er != nil {
		tonviewerTokenInfo = &models.TonviewerTokenInfo{}
	}

	result := &ChainTokenInfo{
		JettonAddress: chainTokenInfo.JettonAddress,
		Decimals:      chainTokenInfo.Decimals,
	}

	if chainTokenInfo.Name == "" {
		result.Name = tonviewerTokenInfo.TokenName
	} else {
		result.Name = chainTokenInfo.Name
	}

	if chainTokenInfo.Symbol == "" {
		result.Symbol = tonviewerTokenInfo.TokenSymbol
	} else {
		result.Symbol = chainTokenInfo.Symbol
	}

	return result, nil
}
