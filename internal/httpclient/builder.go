package httpclient

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"forge-api-client/internal/models"
	"forge-api-client/internal/variables"
)

func BuildRequest(ctx context.Context, definition models.RequestDefinition, values map[string]string) (*http.Request, error) {
	resolvedURL, err := variables.ResolveStrict(definition.URL, values)
	if err != nil {
		return nil, err
	}
	parsed, err := url.Parse(resolvedURL)
	if err != nil {
		return nil, fmt.Errorf("URL inválida: %w", err)
	}
	query := parsed.Query()
	for _, param := range definition.Params {
		if param.Enabled && param.Key != "" {
			key, _ := variables.ResolveStrict(param.Key, values)
			value, e := variables.ResolveStrict(param.Value, values)
			if e != nil {
				return nil, e
			}
			query.Add(key, value)
		}
	}
	parsed.RawQuery = query.Encode()
	body, contentType, err := buildBody(definition, values)
	if err != nil {
		return nil, err
	}
	method := strings.ToUpper(strings.TrimSpace(definition.Method))
	if method == "" {
		method = http.MethodGet
	}
	if strings.EqualFold(definition.Protocol, "SOAP") {
		method = http.MethodPost
	}
	request, err := http.NewRequestWithContext(ctx, method, parsed.String(), body)
	if err != nil {
		return nil, fmt.Errorf("criar requisição: %w", err)
	}
	for _, header := range definition.Headers {
		if header.Enabled && header.Key != "" {
			key, e := variables.ResolveStrict(header.Key, values)
			if e != nil {
				return nil, e
			}
			value, e := variables.ResolveStrict(header.Value, values)
			if e != nil {
				return nil, e
			}
			request.Header.Add(key, value)
		}
	}
	if contentType != "" && request.Header.Get("Content-Type") == "" {
		request.Header.Set("Content-Type", contentType)
	}
	if strings.EqualFold(definition.Protocol, "SOAP") {
		applySOAPHeaders(request, definition)
	}
	if err := applyAuth(request, definition.Auth, values); err != nil {
		return nil, err
	}
	return request, nil
}

func buildBody(def models.RequestDefinition, values map[string]string) (io.Reader, string, error) {
	if strings.EqualFold(def.Protocol, "SOAP") {
		body, err := variables.ResolveStrict(def.Body, values)
		return strings.NewReader(body), soapContentType(def.SOAPVersion), err
	}
	switch strings.ToLower(def.BodyType) {
	case "", "none":
		return nil, "", nil
	case "json":
		body, err := variables.ResolveStrict(def.Body, values)
		return strings.NewReader(body), "application/json", err
	case "xml":
		body, err := variables.ResolveStrict(def.Body, values)
		return strings.NewReader(body), "application/xml", err
	case "html":
		body, err := variables.ResolveStrict(def.Body, values)
		return strings.NewReader(body), "text/html", err
	case "text":
		body, err := variables.ResolveStrict(def.Body, values)
		return strings.NewReader(body), "text/plain", err
	case "form":
		form := url.Values{}
		for _, item := range def.Params {
			if item.Enabled && item.Key != "" {
				value, err := variables.ResolveStrict(item.Value, values)
				if err != nil {
					return nil, "", err
				}
				form.Add(item.Key, value)
			}
		}
		return strings.NewReader(form.Encode()), "application/x-www-form-urlencoded", nil
	case "multipart":
		var buffer bytes.Buffer
		writer := multipart.NewWriter(&buffer)
		for _, part := range def.Multipart {
			if !part.Enabled || part.Key == "" {
				continue
			}
			value, err := variables.ResolveStrict(part.Value, values)
			if err != nil {
				return nil, "", err
			}
			if strings.EqualFold(part.Type, "FILE") {
				file, err := os.Open(value)
				if err != nil {
					return nil, "", fmt.Errorf("abrir arquivo %s: %w", value, err)
				}
				field, err := writer.CreateFormFile(part.Key, filepath.Base(value))
				if err == nil {
					_, err = io.Copy(field, file)
				}
				file.Close()
				if err != nil {
					return nil, "", err
				}
			} else {
				if err := writer.WriteField(part.Key, value); err != nil {
					return nil, "", err
				}
			}
		}
		if err := writer.Close(); err != nil {
			return nil, "", err
		}
		return &buffer, writer.FormDataContentType(), nil
	default:
		body, err := variables.ResolveStrict(def.Body, values)
		return strings.NewReader(body), "application/octet-stream", err
	}
}

func applyAuth(request *http.Request, auth models.AuthDefinition, values map[string]string) error {
	switch strings.ToLower(auth.Type) {
	case "", "none":
		return nil
	case "basic":
		username, e := variables.ResolveStrict(auth.Username, values)
		if e != nil {
			return e
		}
		password, e := variables.ResolveStrict(auth.Password, values)
		if e != nil {
			return e
		}
		request.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(username+":"+password)))
	case "bearer":
		token, e := variables.ResolveStrict(auth.Token, values)
		if e != nil {
			return e
		}
		request.Header.Set("Authorization", "Bearer "+token)
	case "apikey":
		value, e := variables.ResolveStrict(auth.Value, values)
		if e != nil {
			return e
		}
		key, e := variables.ResolveStrict(auth.Key, values)
		if e != nil {
			return e
		}
		if strings.EqualFold(auth.AddTo, "query") {
			query := request.URL.Query()
			query.Set(key, value)
			request.URL.RawQuery = query.Encode()
		} else {
			request.Header.Set(key, value)
		}
	case "oauth2":
		token, e := variables.ResolveStrict(auth.AccessToken, values)
		if e != nil {
			return e
		}
		if strings.TrimSpace(token) == "" {
			return fmt.Errorf("obtenha um access token OAuth 2.0 antes de enviar a requisição")
		}
		request.Header.Set("Authorization", "Bearer "+token)
	default:
		return fmt.Errorf("autenticação %q ainda não é suportada nesta versão", auth.Type)
	}
	return nil
}
func soapContentType(version string) string {
	if version == "1.2" {
		return "application/soap+xml; charset=utf-8"
	}
	return "text/xml; charset=utf-8"
}
func applySOAPHeaders(request *http.Request, definition models.RequestDefinition) {
	if request.Header.Get("Content-Type") == "" {
		request.Header.Set("Content-Type", soapContentType(definition.SOAPVersion))
	}
	if definition.SOAPVersion != "1.2" && definition.SOAPAction != "" && request.Header.Get("SOAPAction") == "" {
		request.Header.Set("SOAPAction", `"`+strings.Trim(definition.SOAPAction, `"`)+`"`)
	}
}
