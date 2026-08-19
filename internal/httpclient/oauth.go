package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"forge-api-client/internal/models"
	"forge-api-client/internal/security"
	"forge-api-client/internal/variables"
)

type oauthResponse struct {
	AccessToken      string      `json:"access_token"`
	RefreshToken     string      `json:"refresh_token"`
	TokenType        string      `json:"token_type"`
	Scope            string      `json:"scope"`
	ExpiresIn        interface{} `json:"expires_in"`
	Error            string      `json:"error"`
	ErrorDescription string      `json:"error_description"`
}

const oauthRenewalWindow = 30 * time.Second

func OAuthTokenNeedsRenewal(auth models.AuthDefinition, now time.Time) bool {
	if !strings.EqualFold(auth.Type, "oauth2") {
		return false
	}
	if strings.TrimSpace(auth.AccessToken) == "" {
		return true
	}
	if strings.TrimSpace(auth.ExpiresAt) == "" {
		return false
	}
	expiresAt, err := time.Parse(time.RFC3339, auth.ExpiresAt)
	if err != nil {
		return true
	}
	return !expiresAt.After(now.UTC().Add(oauthRenewalWindow))
}

func EnsureOAuthToken(ctx context.Context, auth *models.AuthDefinition, values map[string]string, force bool) (bool, error) {
	if auth == nil || !strings.EqualFold(auth.Type, "oauth2") || (!force && !OAuthTokenNeedsRenewal(*auth, time.Now())) {
		return false, nil
	}
	token, err := AcquireOAuthToken(ctx, *auth, values)
	if err != nil {
		return false, err
	}
	auth.AccessToken = token.AccessToken
	auth.ExpiresAt = token.ExpiresAt
	if token.RefreshToken != "" {
		auth.RefreshToken = token.RefreshToken
	}
	return true, nil
}

func AcquireOAuthToken(ctx context.Context, auth models.AuthDefinition, values map[string]string) (models.OAuthTokenResult, error) {
	resolve := func(value string) (string, error) { return variables.ResolveStrict(value, values) }
	tokenURL, err := resolve(auth.TokenURL)
	if err != nil {
		return models.OAuthTokenResult{}, err
	}
	if strings.TrimSpace(tokenURL) == "" {
		return models.OAuthTokenResult{}, fmt.Errorf("informe a Token URL")
	}
	clientID, err := resolve(auth.ClientID)
	if err != nil {
		return models.OAuthTokenResult{}, err
	}
	clientSecret, err := resolve(auth.ClientSecret)
	if err != nil {
		return models.OAuthTokenResult{}, err
	}
	form := url.Values{}
	flow := auth.OAuthFlow
	if flow == "" {
		flow = "client_credentials"
	}
	form.Set("grant_type", flow)
	if auth.Scope != "" {
		scope, e := resolve(auth.Scope)
		if e != nil {
			return models.OAuthTokenResult{}, e
		}
		form.Set("scope", scope)
	}
	switch flow {
	case "client_credentials":
	case "password":
		username, e := resolve(auth.OAuthUsername)
		if e != nil {
			return models.OAuthTokenResult{}, e
		}
		password, e := resolve(auth.OAuthPassword)
		if e != nil {
			return models.OAuthTokenResult{}, e
		}
		form.Set("username", username)
		form.Set("password", password)
	case "refresh_token":
		refresh, e := resolve(auth.RefreshToken)
		if e != nil {
			return models.OAuthTokenResult{}, e
		}
		if refresh == "" {
			return models.OAuthTokenResult{}, fmt.Errorf("informe o refresh token")
		}
		form.Set("refresh_token", refresh)
	default:
		return models.OAuthTokenResult{}, fmt.Errorf("fluxo OAuth 2.0 %q não suportado", flow)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return models.OAuthTokenResult{}, fmt.Errorf("criar requisição OAuth 2.0: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")
	if auth.ClientAuth == "basic" {
		request.SetBasicAuth(clientID, clientSecret)
	} else {
		form.Set("client_id", clientID)
		form.Set("client_secret", clientSecret)
		request.Body = io.NopCloser(strings.NewReader(form.Encode()))
		request.ContentLength = int64(len(form.Encode()))
	}
	client := &http.Client{Timeout: 30 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return models.OAuthTokenResult{}, fmt.Errorf("não foi possível obter o token OAuth 2.0: %w", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 1024*1024))
	if err != nil {
		return models.OAuthTokenResult{}, fmt.Errorf("ler resposta OAuth 2.0: %w", err)
	}
	var payload oauthResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return models.OAuthTokenResult{}, fmt.Errorf("resposta OAuth 2.0 inválida (HTTP %d)", response.StatusCode)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 || payload.Error != "" {
		detail := payload.ErrorDescription
		if detail == "" {
			detail = payload.Error
		}
		if detail == "" {
			detail = response.Status
		}
		return models.OAuthTokenResult{}, fmt.Errorf("servidor OAuth 2.0 recusou o token: %s", security.SanitizeText(detail))
	}
	if payload.AccessToken == "" {
		return models.OAuthTokenResult{}, fmt.Errorf("a resposta OAuth 2.0 não contém access_token")
	}
	expiresAt := ""
	if seconds := expiresSeconds(payload.ExpiresIn); seconds > 0 {
		expiresAt = time.Now().UTC().Add(time.Duration(seconds) * time.Second).Format(time.RFC3339)
	}
	return models.OAuthTokenResult{AccessToken: payload.AccessToken, RefreshToken: payload.RefreshToken, TokenType: payload.TokenType, ExpiresAt: expiresAt, Scope: payload.Scope}, nil
}

func expiresSeconds(value interface{}) int64 {
	switch v := value.(type) {
	case float64:
		return int64(v)
	case string:
		seconds, _ := strconv.ParseInt(v, 10, 64)
		return seconds
	default:
		return 0
	}
}
