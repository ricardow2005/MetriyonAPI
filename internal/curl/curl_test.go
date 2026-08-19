package curl

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestImportExport(t *testing.T) {
	command := `curl --location 'https://api.example.com/users?page=1' --header 'Content-Type: application/json' --header 'Authorization: Bearer token' --data-raw '{"name":"Forge"}'`
	request, err := Import(command)
	if err != nil {
		t.Fatal(err)
	}
	if request.Method != "POST" || request.BodyType != "json" || len(request.Params) != 1 || len(request.Headers) != 1 {
		t.Fatalf("unexpected request: %#v", request)
	}
	if request.Auth.Type != "bearer" || request.Auth.Token != "token" {
		t.Fatalf("bearer auth was not imported: %#v", request.Auth)
	}
	exported := Export(request)
	if !strings.Contains(exported, "--request 'POST'") || !strings.Contains(exported, "Content-Type: application/json") || !strings.Contains(exported, "Authorization: Bearer token") {
		t.Fatalf("unexpected export: %s", exported)
	}
}

func TestImportMultilineCurlWithLF(t *testing.T) {
	command := "curl -X 'GET' \\\n  'http://hostinger-vps:8196/api/v1/clients?active=true&page=2' \\\n  -H 'accept: application/json'"

	request, err := Import(command)
	if err != nil {
		t.Fatal(err)
	}
	if request.URL != "http://hostinger-vps:8196/api/v1/clients" {
		t.Fatalf("endpoint was not imported: %q", request.URL)
	}
	if request.Method != "GET" {
		t.Fatalf("method was not imported: %q", request.Method)
	}
	if len(request.Params) != 2 {
		t.Fatalf("query params were not imported: %#v", request.Params)
	}
	if len(request.Headers) != 1 || !strings.EqualFold(request.Headers[0].Key, "accept") || request.Headers[0].Value != "application/json" {
		t.Fatalf("header was not imported: %#v", request.Headers)
	}
}

func TestImportMultilineCurlWithCRLF(t *testing.T) {
	command := "curl -X POST \\\r\n  --url 'https://api.example.com/items?id=42' \\\r\n  --header 'Content-Type: application/json' \\\r\n  --data-raw '{\"name\":\"Metriyon\"}'"
	request, err := Import(command)
	if err != nil {
		t.Fatal(err)
	}
	if request.URL != "https://api.example.com/items" || len(request.Params) != 1 || request.BodyType != "json" || request.Method != "POST" {
		t.Fatalf("unexpected multiline import: %#v", request)
	}
}

func TestImportBasicAuthFromAuthorizationHeader(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("user:secret"))
	request, err := Import(`curl 'https://api.example.com/private' -H 'Authorization: Basic ` + encoded + `'`)
	if err != nil {
		t.Fatal(err)
	}
	if request.Auth.Type != "basic" || request.Auth.Username != "user" || request.Auth.Password != "secret" {
		t.Fatalf("basic auth was not imported: %#v", request.Auth)
	}
	if len(request.Headers) != 0 {
		t.Fatalf("recognized authorization header should not be duplicated: %#v", request.Headers)
	}
}

func TestImportGetWithDataUrlencodeAsParams(t *testing.T) {
	request, err := Import(`curl --get 'https://api.example.com/search' --data-urlencode 'q=hello world' --data-urlencode 'page=2'`)
	if err != nil {
		t.Fatal(err)
	}
	if request.Method != "GET" || len(request.Params) != 2 || request.Body != "" {
		t.Fatalf("GET query data was not imported correctly: %#v", request)
	}
}

func TestImportLongOptionsWithEquals(t *testing.T) {
	request, err := Import(`curl --request=DELETE --url='https://api.example.com/items/10' --header='Accept: application/json'`)
	if err != nil {
		t.Fatal(err)
	}
	if request.Method != "DELETE" || request.URL != "https://api.example.com/items/10" || len(request.Headers) != 1 {
		t.Fatalf("inline long options were not imported: %#v", request)
	}
}

func TestImportAPIKeyHeader(t *testing.T) {
	request, err := Import(`curl 'https://api.example.com/private' -H 'X-API-Key: secret-key'`)
	if err != nil {
		t.Fatal(err)
	}
	if request.Auth.Type != "apikey" || request.Auth.Key != "X-API-Key" || request.Auth.Value != "secret-key" || request.Auth.AddTo != "header" {
		t.Fatalf("API key auth was not imported: %#v", request.Auth)
	}
}

func TestImportMultipart(t *testing.T) {
	request, err := Import(`curl -X POST https://example.test/upload -F 'description=test' -F 'image=@C:\\image.png'`)
	if err != nil {
		t.Fatal(err)
	}
	if request.BodyType != "multipart" || len(request.Multipart) != 2 || request.Multipart[1].Type != "FILE" {
		t.Fatalf("unexpected multipart: %#v", request.Multipart)
	}
}
