package variables

import (
	"fmt"
	"regexp"
	"strings"
)

var placeholder = regexp.MustCompile(`\{\{\s*([A-Za-z0-9_.-]+)\s*\}\}`)

func Merge(scopes ...map[string]string) map[string]string {
	out := map[string]string{}
	for _, scope := range scopes {
		for key, value := range scope {
			out[key] = value
		}
	}
	return out
}

func Resolve(input string, values map[string]string) (string, []string) {
	missingSet := map[string]bool{}
	result := placeholder.ReplaceAllStringFunc(input, func(token string) string {
		match := placeholder.FindStringSubmatch(token)
		if value, ok := values[match[1]]; ok {
			return value
		}
		missingSet[match[1]] = true
		return token
	})
	missing := make([]string, 0, len(missingSet))
	for key := range missingSet {
		missing = append(missing, key)
	}
	return result, missing
}

func ResolveStrict(input string, values map[string]string) (string, error) {
	result, missing := Resolve(input, values)
	if len(missing) > 0 {
		return "", fmt.Errorf("variáveis não definidas: %s", strings.Join(missing, ", "))
	}
	return result, nil
}
