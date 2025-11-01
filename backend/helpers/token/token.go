package token

import "strings"

func GetAssetType(ticker string) string {
	if IsStablecoin(ticker) {
		return "Stable"
	}

	if strings.Contains(ticker, "XAUT") {
		return "Commodity"
	}

	return "Volatile"
}

func IsStablecoin(ticker string) bool {
	t := strings.ToUpper(ticker)
	switch t {
	case "DAI", "EUR", "USX", "ZEUR", "ZUSD":
		return true
	default:
		return strings.Contains(t, "USD")
	}
}
