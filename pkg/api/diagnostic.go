package api

import (
	"regexp"
	"strings"
)

var (
	diagnosticURL    = regexp.MustCompile(`(?i)(?:https?|wss?)://[^\s"'<>]+`)
	diagnosticSecret = regexp.MustCompile(`(?i)\b(?:authorization|access_token|client_secret|clientSecret|appSecret|file_info|auth_token|upload_id|[a-z_]*openid)\b[\s"':=]+[^\s,;}\]]+`)
	diagnosticOpaque = regexp.MustCompile(`[A-Za-z0-9_+/=\-]{16,}`)
)

func redactAPIDiagnostic(message string, secrets ...string) string {
	for _, secret := range secrets {
		if secret != "" {
			message = strings.ReplaceAll(message, secret, "[redacted]")
		}
	}
	message = diagnosticURL.ReplaceAllString(message, "[redacted URL]")
	message = diagnosticSecret.ReplaceAllString(message, "[redacted field]")
	message = diagnosticOpaque.ReplaceAllString(message, "[redacted value]")
	runes := []rune(strings.ToValidUTF8(message, "�"))
	if len(runes) > 512 {
		return string(runes[:512]) + "…"
	}
	return string(runes)
}
