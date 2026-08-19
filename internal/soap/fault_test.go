package soap

import "testing"

func TestDetectFault11(t *testing.T) {
	fault := DetectFault(`<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body><soap:Fault><faultcode>soap:Server</faultcode><faultstring>Customer not found</faultstring></soap:Fault></soap:Body></soap:Envelope>`)
	if fault == nil || fault.Code != "soap:Server" || fault.Message != "Customer not found" {
		t.Fatalf("unexpected fault: %#v", fault)
	}
}
