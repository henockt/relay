package auth

import "strings"

// only big consumer providers can sign up, to keep out throwaway addresses.
// keep in sync with web/src/lib/allowed-emails.ts
var allowedDomains = map[string]bool{
	"gmail.com":      true,
	"googlemail.com": true,
	"icloud.com":     true,
	"me.com":         true,
	"mac.com":        true,
	"outlook.com":    true,
	"hotmail.com":    true,
	"live.com":       true,
	"msn.com":        true,
	"yahoo.com":      true,
	"proton.me":      true,
	"protonmail.com": true,
}

// AllowedEmail reports whether the email's domain is on the allowlist.
func AllowedEmail(email string) bool {
	email = strings.ToLower(strings.TrimSpace(email))
	at := strings.LastIndex(email, "@")
	if at <= 0 {
		return false
	}
	return allowedDomains[email[at+1:]]
}
