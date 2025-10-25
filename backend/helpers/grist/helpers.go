package grist

import (
	"fmt"
	"net/url"
	"strings"
)

func (g *Grist) generateRecordsUrl(table string, query string) string {
	path := fmt.Sprintf("/api/docs/%s/tables/%s/records", g.DocId, table)
	endpoint := apiBaseURL + path

	v := url.Values{}
	if query != "" {
		for _, kv := range strings.Split(query, "&") {
			if kv == "" {
				continue
			}
			parts := strings.SplitN(kv, "=", 2)
			if len(parts) == 2 {
				v.Set(parts[0], parts[1])
			}
		}
	}

	if enc := v.Encode(); enc != "" {
		return endpoint + "?" + enc
	}

	return endpoint
}

func (g *Grist) generateSqlUrl(query string) string {
	path := fmt.Sprintf("/api/docs/%s/sql", g.DocId)
	endpoint := apiBaseURL + path

	v := url.Values{}
	if query != "" {
		for _, kv := range strings.Split(query, "&") {
			if kv == "" {
				continue
			}
			parts := strings.SplitN(kv, "=", 2)
			if len(parts) == 2 {
				v.Set(parts[0], parts[1])
			}
		}
	}

	if enc := v.Encode(); enc != "" {
		return endpoint + "?" + enc
	}

	return endpoint
}
