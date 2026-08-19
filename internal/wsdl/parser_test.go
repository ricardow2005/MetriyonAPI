package wsdl

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testWSDL = `<?xml version="1.0"?><definitions xmlns="http://schemas.xmlsoap.org/wsdl/" xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/" name="CustomerService" targetNamespace="urn:customer"><binding name="CustomerBinding"><operation name="GetCustomer"><soap:operation soapAction="urn:GetCustomer"/></operation></binding><service name="CustomerService"><port name="CustomerPort"><soap:address location="https://example.test/soap"/></port></service></definitions>`

func TestParseWSDL(t *testing.T) {
	result, err := Parse([]byte(testWSDL))
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "CustomerService" || result.Endpoint != "https://example.test/soap" || len(result.Operations) != 1 || result.Operations[0].SOAPAction != "urn:GetCustomer" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if result.Operations[0].Envelope == "" {
		t.Fatal("envelope was not generated")
	}
}

func TestImportWSDLFromURLUsesCompatibilityHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.UserAgent(), "Mozilla/5.0") || !strings.Contains(r.Header.Get("Accept"), "text/xml") {
			http.Error(w, "legacy server rejected client", http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "text/xml")
		_, _ = w.Write([]byte(testWSDL))
	}))
	defer server.Close()

	result, err := Import(server.URL+"/service.asmx?WSDL", true)
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "CustomerService" || len(result.Operations) != 1 {
		t.Fatalf("unexpected result: %#v", result)
	}
}
