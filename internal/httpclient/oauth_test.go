package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"forge-api-client/internal/models"
)

func TestAcquireOAuthClientCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.Form.Get("grant_type") != "client_credentials" || r.Form.Get("client_id") != "client" || r.Form.Get("client_secret") != "secret" {
			t.Fatalf("unexpected form: %#v", r.Form)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"access_token": "access", "refresh_token": "refresh", "token_type": "Bearer", "expires_in": 3600})
	}))
	defer server.Close()
	auth := models.AuthDefinition{OAuthFlow: "client_credentials", TokenURL: server.URL, ClientID: "client", ClientSecret: "secret", ClientAuth: "body"}
	result, err := AcquireOAuthToken(context.Background(), auth, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.AccessToken != "access" || result.RefreshToken != "refresh" || result.ExpiresAt == "" {
		t.Fatalf("unexpected token: %#v", result)
	}
}

func TestOAuthTokenNeedsRenewal(t *testing.T) {
	now := time.Now().UTC()
	cases := []struct {
		name string
		auth models.AuthDefinition
		want bool
	}{
		{name: "not oauth", auth: models.AuthDefinition{Type: "bearer"}, want: false},
		{name: "missing", auth: models.AuthDefinition{Type: "oauth2"}, want: true},
		{name: "no expiry", auth: models.AuthDefinition{Type: "oauth2", AccessToken: "token"}, want: false},
		{name: "valid", auth: models.AuthDefinition{Type: "oauth2", AccessToken: "token", ExpiresAt: now.Add(time.Hour).Format(time.RFC3339)}, want: false},
		{name: "near expiry", auth: models.AuthDefinition{Type: "oauth2", AccessToken: "token", ExpiresAt: now.Add(20 * time.Second).Format(time.RFC3339)}, want: true},
		{name: "expired", auth: models.AuthDefinition{Type: "oauth2", AccessToken: "token", ExpiresAt: now.Add(-time.Minute).Format(time.RFC3339)}, want: true},
		{name: "invalid expiry", auth: models.AuthDefinition{Type: "oauth2", AccessToken: "token", ExpiresAt: "invalid"}, want: true},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := OAuthTokenNeedsRenewal(test.auth, now); got != test.want {
				t.Fatalf("OAuthTokenNeedsRenewal() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestAcquireOAuthRefreshWithBasicClientAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "client" || password != "secret" {
			t.Fatal("client basic authentication missing")
		}
		_ = r.ParseForm()
		if r.Form.Get("grant_type") != "refresh_token" || r.Form.Get("refresh_token") != "old-refresh" {
			t.Fatalf("unexpected form: %#v", r.Form)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"new-access"}`))
	}))
	defer server.Close()
	auth := models.AuthDefinition{OAuthFlow: "refresh_token", TokenURL: server.URL, ClientID: "client", ClientSecret: "secret", ClientAuth: "basic", RefreshToken: "old-refresh"}
	result, err := AcquireOAuthToken(context.Background(), auth, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.AccessToken != "new-access" {
		t.Fatalf("unexpected token: %#v", result)
	}
}
