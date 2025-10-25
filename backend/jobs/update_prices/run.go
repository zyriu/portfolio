package update_prices

import (
	"context"
	"fmt"
	"strings"

	"github.com/zyriu/portfolio/backend/helpers/coingecko"
	"github.com/zyriu/portfolio/backend/helpers/grist"
	"github.com/zyriu/portfolio/backend/helpers/jobstatus"
)

func Run(ctx context.Context, _ ...any) error {
	updateStatus := jobstatus.GetStatusUpdater(ctx)

	updateStatus("Initializing Grist client...")
	g, err := grist.InitiateClient()
	if err != nil {
		return err
	}

	updateStatus("Fetching price records from Grist Prices table...")
	// If your table has Name/Chain/Contract, you can add them here, e.g.:
	// records, err := g.GetRecords(ctx, "Prices", "fields=Ticker,Coingecko_ID,Name,Chain,Contract")
	records, err := g.GetRecords(ctx, "Prices", "fields=Ticker,Coingecko_ID")
	if err != nil {
		return err
	}

	updateStatus(fmt.Sprintf("Found %d price records in Grist", len(records.Records)))

	// Identify rows missing CoinGecko IDs (skip USD as it doesn't need one)
	var missingIDs []grist.Record
	for _, r := range records.Records {
		// Check ticker first
		tickerRaw, ok := r.Fields["Ticker"]
		if ok && tickerRaw != nil {
			ticker := strings.TrimSpace(strings.ToUpper(fmt.Sprint(tickerRaw)))
			if ticker == "USD" {
				continue // USD doesn't need a CoinGecko ID
			}
		}

		raw, ok := r.Fields["Coingecko_ID"]
		if !ok || raw == nil || strings.TrimSpace(fmt.Sprint(raw)) == "" {
			missingIDs = append(missingIDs, r)
		}
	}

	// Try to resolve missing IDs
	if len(missingIDs) > 0 {
		updateStatus(fmt.Sprintf("Found %d records without CoinGecko ID, attempting to match...", len(missingIDs)))

		updateStatus("Fetching CoinGecko coins list (with platforms)...")
		coinsList, err := coingecko.FetchCoinsListWithPlatforms(ctx)
		if err != nil {
			updateStatus(fmt.Sprintf("⚠️  Failed to fetch coins list: %v", err))
		} else {
			updateStatus(fmt.Sprintf("✓ Received %d coins from CoinGecko", len(coinsList)))

			// Build robust index (symbol -> []coins, name -> []coins)
			idx := coingecko.BuildIndex(coinsList)

			var idUpserts []grist.Upsert
			matched := 0

			for _, r := range missingIDs {
				tickerRaw, ok := r.Fields["Ticker"]
				if !ok || tickerRaw == nil {
					continue
				}
				ticker := strings.TrimSpace(fmt.Sprint(tickerRaw))
				if ticker == "" {
					continue
				}

				// Optionally pull Name/Chain/Contract if your table has them
				var name, chain, contract string
				if v, ok := r.Fields["Name"]; ok && v != nil {
					name = strings.TrimSpace(fmt.Sprint(v))
				}
				if v, ok := r.Fields["Chain"]; ok && v != nil {
					chain = strings.TrimSpace(fmt.Sprint(v))
				}
				if v, ok := r.Fields["Contract"]; ok && v != nil {
					contract = strings.TrimSpace(fmt.Sprint(v))
				}

				rec := coingecko.Record{
					Ticker:   ticker,
					Name:     name,
					Chain:    strings.ToLower(chain),
					Contract: contract,
				}

				id, err := coingecko.ResolveID(ctx, idx, rec)
				if err != nil {
					updateStatus(fmt.Sprintf("⚠️  Resolve failed for %s: %v", ticker, err))
					continue
				}
				if id == "" {
					continue
				}

				idUpserts = append(idUpserts, grist.Upsert{
					Require: map[string]any{
						"Ticker": ticker,
					},
					Fields: map[string]any{
						"Coingecko_ID": id,
					},
				})
				matched++
			}

			if len(idUpserts) > 0 {
				updateStatus(fmt.Sprintf("Matched %d CoinGecko IDs, updating Grist...", matched))
				if err := g.UpsertRecords(ctx, "Prices", idUpserts, grist.UpsertOpts{}); err != nil {
					updateStatus(fmt.Sprintf("⚠️  Failed to update CoinGecko IDs: %v", err))
				} else {
					updateStatus(fmt.Sprintf("✓ Successfully updated %d CoinGecko IDs", len(idUpserts)))

					// Re-fetch records to get the updated IDs (pulling same fields as before)
					updateStatus("Re-fetching price records with updated IDs...")
					records, err = g.GetRecords(ctx, "Prices", "fields=Ticker,Coingecko_ID")
					if err != nil {
						return err
					}
				}
			} else {
				updateStatus("⚠️  No CoinGecko IDs could be matched (likely due to symbol collisions). Consider adding Name or Chain/Contract columns for better disambiguation.")
			}
		}
	}

	updateStatus("Fetching current prices from CoinGecko API...")
	prices, err := coingecko.FetchSimplePrices(ctx, records)
	if err != nil {
		return err
	}

	updateStatus(fmt.Sprintf("✓ Received %d prices from CoinGecko", len(prices)))
	updateStatus("Processing and matching prices with tokens...")
	var upserts []grist.Upsert
	skipped := 0
	for _, r := range records.Records {
		raw, ok := r.Fields["Coingecko_ID"]
		if !ok || raw == nil {
			skipped++
			continue
		}

		coinID := strings.TrimSpace(fmt.Sprint(raw))
		if coinID == "" {
			skipped++
			continue
		}

		price, ok := prices[coinID]
		if !ok {
			skipped++
			continue
		}

		upserts = append(upserts, grist.Upsert{
			Require: map[string]any{
				"Coingecko_ID": coinID,
			},
			Fields: map[string]any{
				"Price": price,
			},
		})
	}

	if skipped > 0 {
		updateStatus(fmt.Sprintf("Skipped %d tokens (missing CoinGecko ID or price not available)", skipped))
	}

	if len(upserts) > 0 {
		updateStatus(fmt.Sprintf("Upserting %d prices to Grist...", len(upserts)))
		if err := g.UpsertRecords(ctx, "Prices", upserts, grist.UpsertOpts{}); err != nil {
			return err
		}
		updateStatus(fmt.Sprintf("✓ Successfully updated %d prices", len(upserts)))
	} else {
		updateStatus("⚠️  No prices to update")
	}

	return nil
}
