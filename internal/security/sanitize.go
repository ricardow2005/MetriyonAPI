package security

import (
	"regexp"
	"strings"
)

var sensitiveKeys = map[string]bool{
	"authorization": true, "proxy-authorization": true, "client-secret": true,
	"client_secret": true, "password": true, "token": true, "private-key": true,
}

func IsSensitive(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	if sensitiveKeys[normalized] {
		return true
	}
	return strings.Contains(normalized, "secret") || strings.Contains(normalized, "password") || strings.Contains(normalized, "token")
}

func Mask(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 4 {
		return "****"
	}
	return value[:4] + "****"
}

func SanitizeHeaders(headers map[string][]string) map[string][]string {
	out := make(map[string][]string, len(headers))
	for key, values := range headers {
		if IsSensitive(key) {
			out[key] = []string{"****"}
			continue
		}
		out[key] = append([]string(nil), values...)
	}
	return out
}

func SanitizeText(text string) string {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(authorization\s*[:=]\s*(?:bearer|basic)?\s*)[^\s,;]+`),
		regexp.MustCompile(`(?i)((?:client[_-]?secret|password|access[_-]?token|refresh[_-]?token)\s*[:=]\s*)[^\s,;]+`),
	}
	for _, pattern := range patterns {
		text = pattern.ReplaceAllString(text, `${1}****`)
	}
	return text
}
