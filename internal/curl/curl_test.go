package curl

import (
	"strings"
	"testing"
)

func TestImportExport(t *testing.T) {
	command := `curl --location 'https://api.example.com/users?page=1' --header 'Content-Type: application/json' --header 'Authorization: Bearer token' --data-raw '{"name":"Forge"}'`
	request, err := Import(command)
	if err != nil {
		t.Fatal(err)
	}
	if request.Method != "POST" || request.BodyType != "json" || len(request.Params) != 1 || len(request.Headers) != 2 {
		t.Fatalf("unexpected request: %#v", request)
	}
	exported := Export(request)
	if !strings.Contains(exported, "--request 'POST'") || !strings.Contains(exported, "Content-Type: application/json") {
		t.Fatalf("unexpected export: %s", exported)
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
