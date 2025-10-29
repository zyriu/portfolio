package pendle_markets

import (
	"strings"

	"github.com/zyriu/portfolio/backend/helpers/chain"
	"github.com/zyriu/portfolio/backend/helpers/grist"
	"github.com/zyriu/portfolio/backend/helpers/misc"
	"github.com/zyriu/portfolio/backend/helpers/pendle"
)

func generatePendleRecords(markets pendle.Markets) []grist.Upsert {
	records := make([]grist.Upsert, 0, len(markets.Markets))

	for _, market := range markets.Markets {
		var cats []string
		for _, id := range market.CategoryIDs {
			cats = append(cats, misc.Capitalize(id))
		}

		records = append(records, grist.Upsert{
			Require: map[string]any{
				"Chain":  chain.IDToName(market.ChainID),
				"Name":   market.Name,
				"Expiry": market.Expiry,
			},
			Fields: map[string]any{
				"Categories":  strings.Join(cats, ", "),
				"Is_New":      market.IsNew,
				"Is_Prime":    market.IsPrime,
				"Implied_APY": market.Details.ImpliedAPY,
				"Pendle_APY":  market.Details.PendleAPY,
				"Liquidity":   market.Details.Liquidity,
			},
		})
	}

	return records
}
