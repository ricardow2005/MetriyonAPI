package wsdl

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"forge-api-client/internal/models"
)

type document struct {
	Name            string    `xml:"name,attr"`
	TargetNamespace string    `xml:"targetNamespace,attr"`
	Services        []service `xml:"service"`
	Bindings        []binding `xml:"binding"`
}
type service struct {
	Name  string `xml:"name,attr"`
	Ports []port `xml:"port"`
}
type port struct {
	Name      string    `xml:"name,attr"`
	Addresses []address `xml:",any"`
}
type address struct {
	XMLName  xml.Name
	Location string `xml:"location,attr"`
}
type binding struct {
	Name       string      `xml:"name,attr"`
	Operations []operation `xml:"operation"`
}
type operation struct {
	Name     string           `xml:"name,attr"`
	Children []operationChild `xml:",any"`
}
type operationChild struct {
	XMLName    xml.Name
	SOAPAction string `xml:"soapAction,attr"`
}

func Import(source string, fromURL bool) (models.WSDLImportResult, error) {
	var body []byte
	var err error
	if fromURL {
		client := http.Client{Timeout: 20 * time.Second}
		resp, e := client.Get(source)
		if e != nil {
			return models.WSDLImportResult{}, fmt.Errorf("baixar WSDL: %w", e)
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return models.WSDLImportResult{}, fmt.Errorf("baixar WSDL: HTTP %s", resp.Status)
		}
		body, err = io.ReadAll(io.LimitReader(resp.Body, 8*1024*1024))
	} else {
		body, err = os.ReadFile(source)
	}
	if err != nil {
		return models.WSDLImportResult{}, fmt.Errorf("ler WSDL: %w", err)
	}
	return Parse(body)
}
func Parse(body []byte) (models.WSDLImportResult, error) {
	var doc document
	if err := xml.Unmarshal(body, &doc); err != nil {
		return models.WSDLImportResult{}, fmt.Errorf("WSDL inválido: %w", err)
	}
	result := models.WSDLImportResult{Name: doc.Name, Operations: []models.WSDLOperation{}}
	if result.Name == "" {
		result.Name = "Imported SOAP Service"
	}
	for _, svc := range doc.Services {
		if result.Name == "Imported SOAP Service" && svc.Name != "" {
			result.Name = svc.Name
		}
		for _, p := range svc.Ports {
			for _, a := range p.Addresses {
				if a.XMLName.Local == "address" && a.Location != "" {
					result.Endpoint = a.Location
					break
				}
			}
		}
	}
	seen := map[string]bool{}
	for _, binding := range doc.Bindings {
		for _, op := range binding.Operations {
			if op.Name == "" || seen[op.Name] {
				continue
			}
			action := ""
			for _, child := range op.Children {
				if child.XMLName.Local == "operation" && child.SOAPAction != "" {
					action = child.SOAPAction
				}
			}
			result.Operations = append(result.Operations, models.WSDLOperation{Name: op.Name, SOAPAction: action, Envelope: envelope(doc.TargetNamespace, op.Name)})
			seen[op.Name] = true
		}
	}
	if len(result.Operations) == 0 {
		return result, fmt.Errorf("nenhuma operação SOAP foi encontrada no WSDL")
	}
	return result, nil
}
func envelope(namespace, operation string) string {
	if namespace == "" {
		namespace = "urn:service"
	}
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:tns="%s">
  <soapenv:Header/>
  <soapenv:Body>
    <tns:%s>
      <!-- Add operation parameters here -->
    </tns:%s>
  </soapenv:Body>
</soapenv:Envelope>`, escape(namespace), operation, operation)
}
func escape(v string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(v))
	return b.String()
}
