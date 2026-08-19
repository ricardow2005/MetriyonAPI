package curl

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"unicode"

	"forge-api-client/internal/models"
	"github.com/google/uuid"
)

func Import(command string) (models.RequestDefinition, error) {
	tokens, err := tokenize(normalizeLineContinuations(command))
	if err != nil {
		return models.RequestDefinition{}, err
	}
	if len(tokens) == 0 || strings.ToLower(tokens[0]) != "curl" {
		return models.RequestDefinition{}, fmt.Errorf("o comando deve começar com curl")
	}

	r := models.RequestDefinition{
		ID:             uuid.NewString(),
		Name:           "Imported request",
		Protocol:       "REST",
		Method:         "GET",
		BodyType:       "none",
		VerifySSL:      true,
		FollowRedirect: true,
		TimeoutSeconds: 30,
		Auth:           models.AuthDefinition{Type: "none"},
		Variables:      map[string]string{},
		Params:         []models.KeyValue{},
		Headers:        []models.KeyValue{},
		Multipart:      []models.MultipartPart{},
	}

	forceQuery := false
	dataQuery := url.Values{}

	for i := 1; i < len(tokens); i++ {
		token := tokens[i]
		option, inlineValue, hasInlineValue := splitLongOption(token)
		if hasInlineValue {
			token = option
		}

		next := func() (string, error) {
			if hasInlineValue {
				hasInlineValue = false
				return inlineValue, nil
			}
			if i+1 >= len(tokens) {
				return "", fmt.Errorf("valor ausente após %s", token)
			}
			i++
			return tokens[i], nil
		}

		switch token {
		case "-X", "--request":
			v, e := next()
			if e != nil {
				return r, e
			}
			r.Method = strings.ToUpper(v)

		case "-H", "--header":
			v, e := next()
			if e != nil {
				return r, e
			}
			if e := importHeader(&r, v); e != nil {
				return r, e
			}

		case "-d", "--data", "--data-raw", "--data-ascii", "--data-binary":
			v, e := next()
			if e != nil {
				return r, e
			}
			if forceQuery {
				appendEncodedData(dataQuery, v)
				continue
			}
			r.Body = joinBodyData(r.Body, v)
			r.BodyType = detectBodyType(r.Body)
			if r.Method == "GET" {
				r.Method = "POST"
			}

		case "--data-urlencode":
			v, e := next()
			if e != nil {
				return r, e
			}
			if forceQuery {
				appendEncodedData(dataQuery, v)
				continue
			}
			r.Body = joinBodyData(r.Body, encodeDataValue(v))
			r.BodyType = "form"
			if r.Method == "GET" {
				r.Method = "POST"
			}

		case "--json":
			v, e := next()
			if e != nil {
				return r, e
			}
			r.Body = v
			r.BodyType = "json"
			if !hasHeader(r.Headers, "Content-Type") {
				addHeader(&r, "Content-Type", "application/json")
			}
			if !hasHeader(r.Headers, "Accept") {
				addHeader(&r, "Accept", "application/json")
			}
			if r.Method == "GET" {
				r.Method = "POST"
			}

		case "-F", "--form":
			v, e := next()
			if e != nil {
				return r, e
			}
			parts := strings.SplitN(v, "=", 2)
			if len(parts) != 2 {
				return r, fmt.Errorf("campo multipart inválido: %s", v)
			}
			kind := "TEXT"
			value := parts[1]
			if strings.HasPrefix(value, "@") {
				kind = "FILE"
				value = strings.TrimPrefix(value, "@")
			}
			r.Multipart = append(r.Multipart, models.MultipartPart{ID: uuid.NewString(), Enabled: true, Type: kind, Key: parts[0], Value: value})
			r.BodyType = "multipart"
			if r.Method == "GET" {
				r.Method = "POST"
			}

		case "-u", "--user":
			v, e := next()
			if e != nil {
				return r, e
			}
			setBasicAuth(&r, v)

		case "-A", "--user-agent":
			v, e := next()
			if e != nil {
				return r, e
			}
			addHeader(&r, "User-Agent", v)

		case "-b", "--cookie":
			v, e := next()
			if e != nil {
				return r, e
			}
			addHeader(&r, "Cookie", v)

		case "-G", "--get":
			forceQuery = true
			if r.Method == "POST" {
				r.Method = "GET"
			}

		case "-k", "--insecure":
			r.VerifySSL = false

		case "-L", "--location":
			r.FollowRedirect = true

		case "--url":
			v, e := next()
			if e != nil {
				return r, e
			}
			r.URL = v

		default:
			if !strings.HasPrefix(token, "-") && r.URL == "" {
				r.URL = token
			}
		}
	}

	if r.URL == "" {
		return r, fmt.Errorf("URL não encontrada no comando cURL")
	}

	parsed, parseErr := url.Parse(r.URL)
	if parseErr != nil {
		return r, fmt.Errorf("URL inválida no comando cURL: %w", parseErr)
	}
	for key, values := range parsed.Query() {
		for _, value := range values {
			r.Params = append(r.Params, models.KeyValue{ID: uuid.NewString(), Enabled: true, Key: key, Value: value})
		}
	}
	for key, values := range dataQuery {
		for _, value := range values {
			r.Params = append(r.Params, models.KeyValue{ID: uuid.NewString(), Enabled: true, Key: key, Value: value})
		}
	}
	parsed.RawQuery = ""
	r.URL = parsed.String()

	return r, nil
}

func normalizeLineContinuations(command string) string {
	replacer := strings.NewReplacer(
		"\\\r\n", " ",
		"\\\n", " ",
		"\\\r", " ",
		"`\r\n", " ",
		"`\n", " ",
	)
	return replacer.Replace(strings.TrimSpace(command))
}

func splitLongOption(token string) (string, string, bool) {
	if !strings.HasPrefix(token, "--") {
		return token, "", false
	}
	parts := strings.SplitN(token, "=", 2)
	if len(parts) != 2 {
		return token, "", false
	}
	return parts[0], parts[1], true
}

func importHeader(r *models.RequestDefinition, header string) error {
	parts := strings.SplitN(header, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("header inválido: %s", header)
	}
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	if strings.EqualFold(key, "Authorization") {
		if importAuthorization(r, value) {
			return nil
		}
	}
	if isCommonAPIKeyHeader(key) && strings.EqualFold(r.Auth.Type, "none") {
		r.Auth.Type = "apikey"
		r.Auth.Key = key
		r.Auth.Value = value
		r.Auth.AddTo = "header"
		return nil
	}
	addHeader(r, key, value)
	return nil
}

func isCommonAPIKeyHeader(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	return normalized == "x-api-key" || normalized == "api-key" || normalized == "x-apikey" || normalized == "apikey"
}

func importAuthorization(r *models.RequestDefinition, value string) bool {
	if strings.HasPrefix(strings.ToLower(value), "bearer ") {
		r.Auth.Type = "bearer"
		r.Auth.Token = strings.TrimSpace(value[len("Bearer "):])
		return true
	}
	if strings.HasPrefix(strings.ToLower(value), "basic ") {
		encoded := strings.TrimSpace(value[len("Basic "):])
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err == nil {
			setBasicAuth(r, string(decoded))
			return true
		}
	}
	return false
}

func setBasicAuth(r *models.RequestDefinition, credentials string) {
	parts := strings.SplitN(credentials, ":", 2)
	r.Auth.Type = "basic"
	r.Auth.Username = parts[0]
	if len(parts) > 1 {
		r.Auth.Password = parts[1]
	}
}

func addHeader(r *models.RequestDefinition, key, value string) {
	r.Headers = append(r.Headers, models.KeyValue{ID: uuid.NewString(), Enabled: true, Key: strings.TrimSpace(key), Value: strings.TrimSpace(value)})
}

func hasHeader(headers []models.KeyValue, key string) bool {
	for _, header := range headers {
		if header.Enabled && strings.EqualFold(header.Key, key) {
			return true
		}
	}
	return false
}

func appendEncodedData(values url.Values, raw string) {
	parts := strings.SplitN(raw, "=", 2)
	if len(parts) == 1 {
		values.Add(parts[0], "")
		return
	}
	values.Add(parts[0], parts[1])
}

func encodeDataValue(raw string) string {
	parts := strings.SplitN(raw, "=", 2)
	if len(parts) == 1 {
		return url.QueryEscape(parts[0])
	}
	return url.QueryEscape(parts[0]) + "=" + url.QueryEscape(parts[1])
}

func joinBodyData(current, value string) string {
	if current == "" {
		return value
	}
	return current + "&" + value
}

func Export(r models.RequestDefinition) string {
	lines := []string{"curl --request " + shellQuote(strings.ToUpper(r.Method))}
	u := r.URL
	if parsed, err := url.Parse(u); err == nil {
		query := parsed.Query()
		for _, p := range r.Params {
			if p.Enabled && p.Key != "" {
				query.Add(p.Key, p.Value)
			}
		}
		parsed.RawQuery = query.Encode()
		u = parsed.String()
	}
	lines = append(lines, "  --url "+shellQuote(u))
	for _, h := range r.Headers {
		if h.Enabled && h.Key != "" {
			lines = append(lines, "  --header "+shellQuote(h.Key+": "+h.Value))
		}
	}
	switch strings.ToLower(r.Auth.Type) {
	case "basic":
		lines = append(lines, "  --user "+shellQuote(r.Auth.Username+":"+r.Auth.Password))
	case "bearer":
		lines = append(lines, "  --header "+shellQuote("Authorization: Bearer "+r.Auth.Token))
	case "apikey":
		if !strings.EqualFold(r.Auth.AddTo, "query") {
			lines = append(lines, "  --header "+shellQuote(r.Auth.Key+": "+r.Auth.Value))
		}
	}
	if r.Body != "" {
		lines = append(lines, "  --data-raw "+shellQuote(r.Body))
	}
	for _, part := range r.Multipart {
		if part.Enabled && part.Key != "" {
			value := part.Value
			if strings.EqualFold(part.Type, "FILE") {
				value = "@" + value
			}
			lines = append(lines, "  --form "+shellQuote(part.Key+"="+value))
		}
	}
	return strings.Join(lines, " \\\n")
}

func shellQuote(v string) string { return "'" + strings.ReplaceAll(v, "'", `'\''`) + "'" }

func detectBodyType(v string) string {
	trimmed := strings.TrimSpace(v)
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		return "json"
	}
	if strings.HasPrefix(trimmed, "<") {
		return "xml"
	}
	return "text"
}

func tokenize(input string) ([]string, error) {
	var out []string
	var current strings.Builder
	var quote rune
	escaped := false
	flush := func() {
		if current.Len() > 0 {
			out = append(out, current.String())
			current.Reset()
		}
	}
	for _, r := range input {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' && quote != '\'' {
			escaped = true
			continue
		}
		if quote != 0 {
			if r == quote {
				quote = 0
			} else {
				current.WriteRune(r)
			}
			continue
		}
		if r == '\'' || r == '"' {
			quote = r
			continue
		}
		if unicode.IsSpace(r) {
			flush()
			continue
		}
		current.WriteRune(r)
	}
	if quote != 0 {
		return nil, fmt.Errorf("aspas não fechadas no comando cURL")
	}
	if escaped {
		current.WriteRune('\\')
	}
	flush()
	return out, nil
}
