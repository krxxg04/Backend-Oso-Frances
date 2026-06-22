package shared

import "strings"

func NormalizeKey(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
