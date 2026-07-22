package core

import "strings"

// SanitizeString removes NULL bytes and ASCII control characters (0x00-0x1F except newline and tab) from input strings.
func SanitizeString(s string) string {
	s = strings.ReplaceAll(s, "\x00", "")
	var sb strings.Builder
	sb.Grow(len(s))
	for _, r := range s {
		if r >= 32 || r == '\n' || r == '\r' || r == '\t' {
			sb.WriteRune(r)
		}
	}
	return strings.TrimSpace(sb.String())
}
