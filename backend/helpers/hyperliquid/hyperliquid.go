package hyperliquid

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/zyriu/portfolio/backend/helpers/misc"
)

const apiBaseURL = "https://api.hyperliquid.xyz"

type Hyperliquid struct{}

type UserFill struct {
	Coin      string `json:"coin"`
	ClosedPnL string `json:"closedPnl"`
	Tid       int64  `json:"tid"`
	Time      int64  `json:"time"`
	Fee       string `json:"fee"`
	FeeToken  string `json:"feeToken"`
	Sz        string `json:"sz"`
	Px        string `json:"px"`
	Side      string `json:"side"`
}

type SpotClearinghouseState struct {
	Balances []struct {
		Coin     string `json:"coin"`
		Token    int    `json:"token"`
		Hold     string `json:"hold"`
		Total    string `json:"total"`
		EntryNtl string `json:"entryNtl"`
	} `json:"balances"`
}

func InitiateClient() (Hyperliquid, error) {
	var hyperliquid Hyperliquid

	return hyperliquid, nil
}

func (h *Hyperliquid) GetUserFills(ctx context.Context, user string) ([]UserFill, error) {
	path := "/info"
	endpoint := apiBaseURL + path

	body := struct {
		Type            string `json:"type"`
		User            string `json:"user"`
		AggregateByTime bool   `json:"aggregateByTime"`
	}{
		Type:            "userFills",
		User:            user,
		AggregateByTime: true,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	resp, err := h.queryAPI(ctx, b, endpoint)
	if err != nil {
		return nil, err
	}

	var userFills []UserFill
	if err := json.Unmarshal(resp, &userFills); err != nil {
		return nil, err
	}

	return userFills, nil
}

func (h *Hyperliquid) GetUserFillsByTime(ctx context.Context, user string, startTime int64) ([]UserFill, error) {
	path := "/info"
	endpoint := apiBaseURL + path

	body := struct {
		Type            string `json:"type"`
		User            string `json:"user"`
		AggregateByTime bool   `json:"aggregateByTime"`
		StartTime       int64  `json:"startTime"`
	}{
		Type:            "userFillsByTime",
		User:            user,
		AggregateByTime: true,
		StartTime:       startTime * 1000,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	resp, err := h.queryAPI(ctx, b, endpoint)
	if err != nil {
		return nil, err
	}

	var userFills []UserFill
	if err := json.Unmarshal(resp, &userFills); err != nil {
		return nil, err
	}

	return userFills, nil
}

func (h *Hyperliquid) GetSpotBalances(ctx context.Context, user string) (SpotClearinghouseState, error) {
	path := "/info"
	endpoint := apiBaseURL + path

	var balances SpotClearinghouseState

	body := struct {
		Type string `json:"type"`
		User string `json:"user"`
	}{
		Type: "spotClearinghouseState",
		User: user,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return balances, fmt.Errorf("marshal payload: %w", err)
	}

	resp, err := h.queryAPI(ctx, b, endpoint)
	if err != nil {
		return balances, err
	}

	if err := json.Unmarshal(resp, &balances); err != nil {
		return balances, err
	}

	return balances, nil
}

func (h *Hyperliquid) IsPerpetuals(ticker string) bool {
	if ticker[0] == '@' {
		return false
	}

	switch ticker {
	case "PURR/USDC", "UFART", "UPUMP", "UBTC", "UXPL", "UETH", "USOL":
		return false
	default:
		return true
	}
}

func (h *Hyperliquid) NormalizeTicker(ticker string) string {
	ticker, _, _ = strings.Cut(ticker, "/")

	switch ticker {
	case "@74":
		return "OMNIX"
	case "@85":
		return "PIP"
	case "@107", "@207":
		return "HYPE"
	case "@116":
		return "MUNCH"
	case "@125":
		return "VAULT"
	case "@142", "UBTC":
		return "BTC"
	case "@151", "UETH":
		return "ETH"
	case "@156", "USOL":
		return "SOL"
	case "@162", "UFART":
		return "FARTCOIN"
	case "@166":
		return "USDT0"
	case "@188", "UPUMP":
		return "PUMP"
	default:
		return ticker
	}
}

func (h *Hyperliquid) queryAPI(ctx context.Context, body []byte, endpoint string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp := misc.DoWithRetry(ctx, req)
	if resp.Err != nil {
		return nil, resp.Err
	}

	return resp.Body, nil
}
