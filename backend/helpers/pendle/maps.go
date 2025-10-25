package pendle

import "fmt"

func GenerateAssetsMap(assets Assets) AssetsMap {
	assetsMap := make(AssetsMap, len(assets.Assets))
	for _, asset := range assets.Assets {
		key := fmt.Sprintf("%d-%s", asset.ChainID, asset.Address)
		assetsMap[key] = asset
	}

	return assetsMap
}

func GenerateMarketsMap(markets Markets) MarketsMap {
	marketsMap := make(MarketsMap, len(markets.Markets))
	for _, market := range markets.Markets {
		key := fmt.Sprintf("%d-%s", market.ChainID, market.Address)
		marketsMap[key] = market
	}

	return marketsMap
}
