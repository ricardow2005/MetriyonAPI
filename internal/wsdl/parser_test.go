package wsdl

import "testing"

func TestParseWSDL(t *testing.T) {
	source := []byte(`<?xml version="1.0"?><definitions xmlns="http://schemas.xmlsoap.org/wsdl/" xmlns:soap="http://schemas.xmlsoap.org/wsdl/soap/" name="CustomerService" targetNamespace="urn:customer"><binding name="CustomerBinding"><operation name="GetCustomer"><soap:operation soapAction="urn:GetCustomer"/></operation></binding><service name="CustomerService"><port name="CustomerPort"><soap:address location="https://example.test/soap"/></port></service></definitions>`)
	result, err := Parse(source)
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
