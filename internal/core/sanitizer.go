package core

import (
	"net/mail"
	"strings"
)

// SanitizeString removes NULL bytes and ASCII control characters from input strings.
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

// IsValidEmail uses Go's standard net/mail parser to validate email format per RFC 5322 standards.
func IsValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if len(email) < 5 || len(email) > 254 {
		return false
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return false
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	domain := parts[1]
	if !strings.Contains(domain, ".") || strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}

	dotParts := strings.Split(domain, ".")
	lastPart := dotParts[len(dotParts)-1]
	return len(lastPart) >= 2
}
