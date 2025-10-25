package token

import "strings"

func IsStableOrVolatile(ticker string) string {
	if IsStablecoin(ticker) == true {
		return "Stable"
	}

	return "Volatile"
}

func IsStablecoin(ticker string) bool {
	t := strings.ToUpper(ticker)
	switch t {
	case "DAI", "EUR", "ZEUR", "ZUSD":
		return true
	default:
		return strings.Contains(t, "USD")
	}
}
