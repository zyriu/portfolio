package grist

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/zyriu/portfolio/backend/helpers/chain"
	"github.com/zyriu/portfolio/backend/helpers/misc"
	"github.com/zyriu/portfolio/backend/helpers/settings"
)

const apiBaseURL = "https://docs.getgrist.com"

func InitiateClient() (Grist, error) {
	var grist Grist

	// Load settings
	settingsData, err := settings.GetCurrentSettings()
	if err != nil {
		return grist, fmt.Errorf("failed to load settings: %v", err)
	}

	grist.ApiKey = settingsData.Grist.APIKey
	grist.DocId = settingsData.Grist.DocumentID
	return grist, nil
}

func (g *Grist) BackupDocument(ctx context.Context, outputPath string) error {
	path := fmt.Sprintf("/api/docs/%s/download", g.DocId)
	endpoint := apiBaseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+g.ApiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func (g *Grist) DeleteRecords(ctx context.Context, table string, recordsIDs []int64) error {
	path := fmt.Sprintf("/api/docs/%s/tables/%s/data/delete", g.DocId, table)
	endpoint := apiBaseURL + path

	body, err := json.Marshal(recordsIDs)
	if err != nil {
		return fmt.Errorf("marshal ids: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+g.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	resp := misc.DoWithRetry(ctx, req)
	return resp.Err
}

func (g *Grist) GetExchangeLatestTrades(ctx context.Context, exchange string, limit uint64) ([]Trade, error) {
	var latestTrades []Trade

	query := fmt.Sprintf("filter={\"Exchange\":[\"%s\"]}&sort=Time&limit=%d", exchange, limit)
	resp, err := g.GetRecords(ctx, "Trades", query)
	if err != nil {
		return latestTrades, err
	}

	latestTrades = make([]Trade, 0, len(resp.Records))
	for _, rec := range resp.Records {
		b, err := json.Marshal(rec.Fields)
		if err != nil {
			return nil, err
		}

		var gt Trade
		if err := json.Unmarshal(b, &gt); err != nil {
			return nil, err
		}

		latestTrades = append(latestTrades, gt)
	}

	return latestTrades, nil
}

func (g *Grist) GetMatchingRecordsAmount(ctx context.Context, table string, where string) (int64, error) {
	query := fmt.Sprintf(`q=SELECT COUNT(*) AS n FROM "%s" WHERE %s`, table, where)
	u := g.generateSqlUrl(query)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+g.ApiKey)

	resp := misc.DoWithRetry(ctx, req)
	if resp.Err != nil {
		return 0, fmt.Errorf("grist.GetRecords: %s", resp.Err)
	}

	if resp.Status != http.StatusOK {
		return 0, fmt.Errorf("grist.GetRecords: code %d: %s", resp.Status, resp.Err)
	}

	var out struct {
		Records []struct {
			Fields struct {
				N json.Number `json:"n"`
			} `json:"fields"`
		} `json:"records"`
	}

	if err := json.Unmarshal(resp.Body, &out); err != nil {
		return 0, err
	}

	count, _ := out.Records[0].Fields.N.Int64()
	return count, nil
}

func (g *Grist) GetRecords(ctx context.Context, table string, query string) (Records, error) {
	var records Records

	endpoint := g.generateRecordsUrl(table, query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return records, err
	}

	req.Header.Set("Authorization", "Bearer "+g.ApiKey)

	resp := misc.DoWithRetry(ctx, req)
	if resp.Err != nil {
		return records, fmt.Errorf("grist.GetRecords: %s", resp.Err)
	}

	if resp.Status != http.StatusOK {
		return records, fmt.Errorf("grist.GetRecords: code %d: %s", resp.Status, resp.Err)
	}

	if err := json.Unmarshal(resp.Body, &records); err != nil {
		return records, err
	}

	return records, nil
}

func (g *Grist) FetchBook(ctx context.Context) (Book, error) {
	book := make(Book)

	endpoint := g.generateRecordsUrl("Book", "")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return book, err
	}

	req.Header.Set("Authorization", "Bearer "+g.ApiKey)

	resp := misc.DoWithRetry(ctx, req)
	if resp.Err != nil {
		return book, fmt.Errorf("grist.GetRecords: %s", resp.Err)
	}

	if resp.Status != http.StatusOK {
		return book, fmt.Errorf("grist.GetRecords: code %d: %s", resp.Status, resp.Err)
	}

	var records struct {
		Records []struct {
			Fields BookEntry `json:"fields"`
		} `json:"records"`
	}

	if err := json.Unmarshal(resp.Body, &records); err != nil {
		return book, err
	}

	for _, r := range records.Records {
		key := fmt.Sprintf("%s-%s-%s", r.Fields.Exchange, r.Fields.Market, r.Fields.Ticker)
		book[key] = r.Fields
	}

	return book, nil
}

func (g *Grist) FetchPrices(ctx context.Context) (Prices, error) {
	prices := make(Prices)

	endpoint := g.generateRecordsUrl("Prices", "")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return prices, err
	}

	req.Header.Set("Authorization", "Bearer "+g.ApiKey)

	resp := misc.DoWithRetry(ctx, req)
	if resp.Err != nil {
		return prices, fmt.Errorf("grist.GetRecords: %s", resp.Err)
	}

	if resp.Status != http.StatusOK {
		return prices, fmt.Errorf("grist.GetRecords: code %d: %s", resp.Status, resp.Err)
	}

	var records struct {
		Records []struct {
			Fields Price `json:"fields"`
		} `json:"records"`
	}

	if err := json.Unmarshal(resp.Body, &records); err != nil {
		return prices, err
	}

	for _, r := range records.Records {
		prices[r.Fields.Ticker] = r.Fields.Price
	}

	return prices, nil
}

func (g *Grist) FetchTokens(ctx context.Context) (Tokens, error) {
	tokens := make(Tokens)

	endpoint := g.generateRecordsUrl("Tokens", "")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return tokens, err
	}

	req.Header.Set("Authorization", "Bearer "+g.ApiKey)

	resp := misc.DoWithRetry(ctx, req)
	if resp.Err != nil {
		return tokens, fmt.Errorf("grist.GetRecords: %s", resp.Err)
	}

	if resp.Status != http.StatusOK {
		return tokens, fmt.Errorf("grist.GetRecords: code %d: %s", resp.Status, resp.Err)
	}

	var records struct {
		Records []struct {
			Fields Token `json:"fields"`
		} `json:"records"`
	}

	if err := json.Unmarshal(resp.Body, &records); err != nil {
		return tokens, err
	}

	for _, r := range records.Records {
		key := fmt.Sprintf("%d-%s", chain.NameToID(r.Fields.Chain), r.Fields.Address)
		tokens[key] = r.Fields
	}

	return tokens, nil
}

func (g *Grist) UpsertRecords(ctx context.Context, table string, records []Upsert, opts UpsertOpts) error {
	endpoint := g.generateRecordsUrl(table, "")

	q := url.Values{}
	if opts.AllowEmptyRequire {
		q.Set("allow_empty_require", "true")
	}
	if opts.OnMany != "" {
		q.Set("on_many", opts.OnMany)
	}
	if enc := q.Encode(); enc != "" {
		sep := "&"
		if !strings.Contains(endpoint, "?") {
			sep = "?"
		}
		endpoint = endpoint + sep + enc
	}

	body, _ := json.Marshal(upsertPayload{Records: records})
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		return fmt.Errorf("err building request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+g.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	resp := misc.DoWithRetry(ctx, req)
	if resp.Err != nil {
		return fmt.Errorf("grist.UpsertRecords: %s", resp.Err)
	}

	if resp.Status != http.StatusOK {
		return fmt.Errorf("grist.UpsertRecords: %d: %s", resp.Status, resp.Err)
	}

	return nil
}
