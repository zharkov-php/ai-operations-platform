package redaction

import (
	"regexp"
	"strings"
)

const replacement = "[REDACTED]"

var patterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9._~+/-]+=*`),
	regexp.MustCompile(`(?i)(api[_-]?key|password|authorization|cookie)\s*[:=]\s*[^\s,;]+`),
	regexp.MustCompile(`-----BEGIN(?: RSA| EC| OPENSSH)? PRIVATE KEY-----[\s\S]*?-----END(?: RSA| EC| OPENSSH)? PRIVATE KEY-----`),
}
var sensitiveNames = map[string]bool{"authorization": true, "cookie": true, "password": true, "api_key": true, "apikey": true, "access_token": true, "refresh_token": true, "private_key": true}

func Text(value string, redactContacts bool) string {
	result := value
	for _, pattern := range patterns {
		result = pattern.ReplaceAllString(result, replacement)
	}
	if redactContacts {
		result = regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`).ReplaceAllString(result, replacement)
		result = regexp.MustCompile(`\+?[0-9][0-9 ()-]{7,}[0-9]`).ReplaceAllString(result, replacement)
	}
	return result
}
func Metadata(value map[string]any, redactContacts bool) map[string]any {
	clean := make(map[string]any, len(value))
	for key, item := range value {
		if sensitiveNames[strings.ToLower(key)] {
			clean[key] = replacement
			continue
		}
		switch typed := item.(type) {
		case string:
			clean[key] = Text(typed, redactContacts)
		case map[string]any:
			clean[key] = Metadata(typed, redactContacts)
		case []any:
			items := make([]any, len(typed))
			for i, nested := range typed {
				if nestedMap, ok := nested.(map[string]any); ok {
					items[i] = Metadata(nestedMap, redactContacts)
				} else if text, ok := nested.(string); ok {
					items[i] = Text(text, redactContacts)
				} else {
					items[i] = nested
				}
			}
			clean[key] = items
		default:
			clean[key] = item
		}
	}
	return clean
}
func Preview(value string, max int) string {
	clean := Text(value, true)
	runes := []rune(clean)
	if len(runes) <= max {
		return clean
	}
	return string(runes[:max]) + "…"
}
