package grist

import (
	"fmt"
	"strconv"
	"strings"
)

func ExtractColumnDataFromRecords[T comparable](records Records, field string) []T {
	data := make([]T, 0, len(records.Records))
	seen := make(map[T]struct{})

	for _, r := range records.Records {
		raw, ok := r.Fields[field]
		if !ok || raw == nil {
			continue
		}

		var v T
		if cast, ok := raw.(T); ok {
			v = cast
		} else {
			s := strings.TrimSpace(fmt.Sprint(raw))
			if s == "" {
				continue
			}

			var parsed T
			switch any(parsed).(type) {
			case string:
				parsed = any(s).(T)
			case int:
				if v, err := strconv.Atoi(s); err == nil {
					parsed = any(v).(T)
				}
			case int64:
				if v, err := strconv.ParseInt(s, 10, 64); err == nil {
					parsed = any(v).(T)
				}
			case float64:
				if v, err := strconv.ParseFloat(s, 64); err == nil {
					parsed = any(v).(T)
				}
			default:
				continue
			}

			v = parsed
		}

		if _, dup := seen[v]; dup {
			continue
		}

		seen[v] = struct{}{}
		data = append(data, v)
	}

	return data
}
