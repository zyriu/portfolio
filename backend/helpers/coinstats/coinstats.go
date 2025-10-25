package coinstats

import (
	"context"
	"fmt"
	"net/http"

	"github.com/zyriu/portfolio/backend/helpers/misc"
	"github.com/zyriu/portfolio/backend/helpers/settings"
)

const apiBaseURL = "https://openapiv1.coinstats.app"

type Coinstats struct {
	ApiKey string
}

type EvmMultiChainBalances struct {
	Blockchain string `json:"blockchain"`
	Address    string `json:"address"`
	Balances   []struct {
		CoinID          string  `json:"coinId"`
		Amount          float64 `json:"amount"`
		Chain           string  `json:"chain"`
		Name            string  `json:"name"`
		Symbol          string  `json:"symbol"`
		Price           float64 `json:"price"`
		WalletAddress   string  `json:"walletAddress"`
		ContractAddress string  `json:"contractAddress"`
	} `json:"balances"`
}

type SingleChainBalance struct {
	Balances []struct {
		CoinID          string  `json:"coinId"`
		Amount          float64 `json:"amount"`
		ContractAddress string  `json:"contractAddress,omitempty"`
		Decimals        int64   `json:"decimals,omitempty"`
		Chain           string  `json:"chain"`
		Name            string  `json:"name"`
		Symbol          string  `json:"symbol"`
		Price           float64 `json:"price"`
	}
}

func InitiateClient() (Coinstats, error) {
	var coinstats Coinstats

	// Load settings
	settingsData, err := settings.GetCurrentSettings()
	if err != nil {
		return coinstats, fmt.Errorf("failed to load settings: %v", err)
	}

	coinstats.ApiKey = settingsData.OnChain.CoinStatsAPIKey
	return coinstats, nil
}

func (c *Coinstats) BuildEVMBalancesRequest(ctx context.Context, address string) (*http.Request, error) {
	path := "/wallet/balances"
	endpoint := fmt.Sprintf("%s%s?address=%s&networks=all", apiBaseURL, path, address)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-KEY", c.ApiKey)
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func (c *Coinstats) BuildSingleBalanceRequest(ctx context.Context, address string, connectionId string) (*http.Request, error) {
	path := "/wallet/balance"
	endpoint := fmt.Sprintf("%s%s?address=%s&connectionId=%s", apiBaseURL, path, address, connectionId)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-KEY", c.ApiKey)
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func (c *Coinstats) NormalizeBlockchainForGrist(blockchain string) string {
	switch blockchain {
	case "arbitrum-one":
		return "Arbitrum"
	case "bsc":
		return "BSC"
	case "hyperevm":
		return "HyperEVM"
	default:
		return misc.Capitalize(blockchain)
	}
}

func (c *Coinstats) NormalizeTickerForGrist(symbol string) string {
	switch symbol {
	case "USDE":
		return "USDe"
	default:
		return symbol
	}
}
