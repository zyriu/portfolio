package lighter

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/zyriu/portfolio/backend/helpers/misc"
)

const apiBaseURL = "https://mainnet.zklighter.elliot.ai"

func InitiateClient() (Lighter, error) {
	var lighter Lighter

	return lighter, nil
}

func (l *Lighter) GetAccounts(ctx context.Context, l1Address string) ([]Account, error) {
	path := "/api/v1/account"
	endpoint := apiBaseURL + path

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("by", "l1_address")
	q.Add("value", l1Address)
	req.URL.RawQuery = q.Encode()

	req.Header.Add("accept", "application/json")

	resp := misc.DoWithRetry(ctx, req)
	if resp.Err != nil {
		return nil, resp.Err
	}

	var result AccountsResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, err
	}

	accounts := make([]Account, 0, len(result.Accounts))
	accounts = append(accounts, result.Accounts...)

	return accounts, nil
}

func (l *Lighter) GetSubAccounts(ctx context.Context, l1Address string) ([]SubAccount, error) {
	path := "/api/v1/accountsByL1Address"
	endpoint := apiBaseURL + path

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("l1_address", l1Address)
	req.URL.RawQuery = q.Encode()

	req.Header.Add("accept", "application/json")

	resp := misc.DoWithRetry(ctx, req)
	if resp.Err != nil {
		return nil, resp.Err
	}

	var result SubAccountsResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, err
	}

	subAccounts := make([]SubAccount, 0, len(result.SubAccounts))
	subAccounts = append(subAccounts, result.SubAccounts...)

	return subAccounts, nil
}
