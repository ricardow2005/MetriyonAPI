package httpclient

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"forge-api-client/internal/models"
)

func baseRequest() models.RequestDefinition {
	return models.RequestDefinition{Protocol: "REST", Method: "POST", URL: "https://{{host}}/users", Params: []models.KeyValue{}, Headers: []models.KeyValue{}, Auth: models.AuthDefinition{Type: "none"}, BodyType: "none", VerifySSL: true, FollowRedirect: true, TimeoutSeconds: 5}
}

func TestBuildRESTRequest(t *testing.T) {
	r := baseRequest()
	r.Params = []models.KeyValue{{Enabled: true, Key: "page", Value: "{{page}}"}, {Enabled: false, Key: "debug", Value: "true"}}
	r.Headers = []models.KeyValue{{Enabled: true, Key: "X-Test", Value: "{{value}}"}}
	r.Auth = models.AuthDefinition{Type: "basic", Username: "user", Password: "pass"}
	request, err := BuildRequest(context.Background(), r, map[string]string{"host": "example.test", "page": "2", "value": "ok"})
	if err != nil {
		t.Fatal(err)
	}
	if request.URL.String() != "https://example.test/users?page=2" {
		t.Fatalf("URL: %s", request.URL)
	}
	if request.Header.Get("X-Test") != "ok" {
		t.Fatal("header not resolved")
	}
	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:pass"))
	if request.Header.Get("Authorization") != want {
		t.Fatal("basic auth not applied")
	}
}

func TestBuildSOAP12Request(t *testing.T) {
	r := baseRequest()
	r.Protocol = "SOAP"
	r.SOAPVersion = "1.2"
	r.SOAPAction = "urn:Run"
	r.Body = `<Envelope>{{id}}</Envelope>`
	request, err := BuildRequest(context.Background(), r, map[string]string{"host": "soap.test", "id": "42"})
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(request.Body)
	if string(body) != `<Envelope>42</Envelope>` {
		t.Fatalf("body: %s", body)
	}
	if !strings.HasPrefix(request.Header.Get("Content-Type"), "application/soap+xml") {
		t.Fatalf("content type: %s", request.Header.Get("Content-Type"))
	}
	if request.Header.Get("SOAPAction") != "" {
		t.Fatal("SOAP 1.2 must not add SOAPAction header")
	}
}

func TestEngineExecuteWithHTTPTestServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	definition := baseRequest()
	definition.URL = server.URL
	definition.VerifySSL = true
	request, err := BuildRequest(context.Background(), definition, nil)
	if err != nil {
		t.Fatal(err)
	}
	result, err := New().Execute(context.Background(), "test", request, definition, 1024)
	if err != nil {
		t.Fatal(err)
	}
	if result.StatusCode != 201 || result.Body != `{"ok":true}` || result.Size == 0 {
		t.Fatalf("unexpected response: %#v", result)
	}
}

func TestBuildOAuthBearerRequest(t *testing.T) {
	r := baseRequest()
	r.Auth = models.AuthDefinition{Type: "oauth2", AccessToken: "{{accessToken}}"}
	request, err := BuildRequest(context.Background(), r, map[string]string{"host": "example.test", "accessToken": "oauth-token"})
	if err != nil {
		t.Fatal(err)
	}
	if request.Header.Get("Authorization") != "Bearer oauth-token" {
		t.Fatalf("authorization: %s", request.Header.Get("Authorization"))
	}
}
