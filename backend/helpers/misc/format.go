package misc

import "strings"

func Capitalize(s string) string {
	return strings.ToUpper(string(s[0])) + s[1:]
}
