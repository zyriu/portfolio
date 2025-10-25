package kraken

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/zyriu/portfolio/backend/helpers/misc"
	"github.com/zyriu/portfolio/backend/helpers/settings"
)

const apiBaseURL = "https://api.kraken.com"

func InitiateClient() (Kraken, error) {
	var kraken Kraken

	// Load settings
	settingsData, err := settings.GetCurrentSettings()
	if err != nil {
		return kraken, fmt.Errorf("failed to load settings: %v", err)
	}

	kraken.ApiKey = settingsData.Exchanges.Kraken.APIKey
	kraken.ApiSecret = settingsData.Exchanges.Kraken.APISecret
	return kraken, nil
}

func (k *Kraken) GetBalances(ctx context.Context) (map[string]string, error) {
	path := "/0/private/Balance"

	body, _ := k.queryAPI(ctx, url.Values{}, path)
	var balances Balances
	if err := json.Unmarshal(body, &balances); err != nil {
		return nil, err
	}

	if len(balances.Error) > 0 {
		return nil, io.EOF
	}

	return balances.Result, nil
}

func (k *Kraken) GetBaseAndQuote(pair string) (string, string) {
	krakenQuotes := []string{"ZUSD", "USDT", "USDC", "USD", "ZSGD", "SGD", "ZEUR", "EUR", "XBTC", "XXBT", "XBT"}

	for _, quote := range krakenQuotes {
		if strings.HasSuffix(pair, quote) {
			base := pair[:len(pair)-len(quote)]
			return k.NormalizeTicker(base), k.NormalizeTicker(quote)
		}
	}

	return pair, ""
}

func (k *Kraken) GetTradesHistory(ctx context.Context, offset int64) (TradesHistory, error) {
	path := "/0/private/TradesHistory"

	form := url.Values{}
	form.Set("ofs", fmt.Sprintf("%d", offset))

	body, _ := k.queryAPI(ctx, form, path)

	var resp TradesHistory
	if err := json.Unmarshal(body, &resp); err != nil {
		return resp, err
	}

	if len(resp.Error) > 0 {
		return resp, fmt.Errorf("kraken error: %v", resp.Error)
	}

	return resp, nil
}

func (k *Kraken) NormalizeTicker(ticker string) string {
	switch ticker {
	case "XXRP", "XZEC", "XETH", "ZUSD", "ZEUR", "XXDG", "XLTC", "XXMR":
		return ticker[1:]
	case "XBT", "XXBT", "XBT.F":
		return "BTC"
	case "TAO.F":
		return "TAO"
	case "SOL.F":
		return "SOL"
	default:
		return ticker
	}
}

func (k *Kraken) queryAPI(ctx context.Context, data url.Values, path string) ([]byte, error) {
	endpoint := apiBaseURL + path

	nonce := strconv.FormatInt(time.Now().UnixNano(), 10)
	data.Set("nonce", nonce)

	sha := sha256.New()
	sha.Write([]byte(nonce + data.Encode()))
	shaSum := sha.Sum(nil)

	decodedSecret, err := base64.StdEncoding.DecodeString(k.ApiSecret)
	if err != nil {
		return nil, err
	}

	mac := hmac.New(sha512.New, decodedSecret)
	mac.Write(append([]byte(path), shaSum...))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("API-Key", k.ApiKey)
	req.Header.Set("API-Sign", signature)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp := misc.DoWithRetry(ctx, req)
	if resp.Err != nil {
		return nil, err
	}

	return resp.Body, nil
}
