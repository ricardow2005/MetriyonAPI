package wsdl

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"forge-api-client/internal/models"
)

const maxWSDLSize = 8 * 1024 * 1024

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
		body, err = downloadWSDL(source)
	} else {
		body, err = os.ReadFile(source)
	}
	if err != nil {
		return models.WSDLImportResult{}, fmt.Errorf("ler WSDL: %w", err)
	}
	return Parse(body)
}

func downloadWSDL(source string) ([]byte, error) {
	parsed, err := url.Parse(strings.TrimSpace(source))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("URL WSDL inválida: %s", source)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("protocolo WSDL não suportado: %s", parsed.Scheme)
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 0}).DialContext,
		DisableKeepAlives:     true,
		ForceAttemptHTTP2:     false,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
	}
	client := &http.Client{Timeout: 25 * time.Second, Transport: transport}

	req, err := http.NewRequest(http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("criar requisição WSDL: %w", err)
	}
	// Alguns serviços SOAP/IIS legados encerram conexões de clientes HTTP genéricos.
	// Esses cabeçalhos também tornam a requisição equivalente ao acesso manual pelo navegador.
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 MetriyonAPI/WSDL")
	req.Header.Set("Accept", "application/wsdl+xml, application/xml, text/xml, */*")
	req.Header.Set("Accept-Encoding", "identity")
	req.Header.Set("Connection", "close")
	req.Close = true

	resp, requestErr := client.Do(req)
	if requestErr == nil {
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("baixar WSDL: HTTP %s", resp.Status)
		}
		return readLimitedWSDL(resp.Body)
	}

	// Fallback para endpoints HTTP muito antigos que encerram HTTP/1.1 antes de responder.
	// O navegador costuma conseguir abrir esses WSDLs, enquanto o net/http pode receber
	// "connection was forcibly closed by the remote host".
	if parsed.Scheme == "http" {
		if body, fallbackErr := downloadWSDLHTTP10(parsed); fallbackErr == nil {
			return body, nil
		}
	}
	return nil, fmt.Errorf("baixar WSDL: %w", requestErr)
}

func downloadWSDLHTTP10(parsed *url.URL) ([]byte, error) {
	host := parsed.Host
	if !strings.Contains(host, ":") {
		host += ":80"
	}
	conn, err := net.DialTimeout("tcp", host, 10*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(20 * time.Second))

	path := parsed.RequestURI()
	if path == "" {
		path = "/"
	}
	hostHeader := parsed.Hostname()
	request := fmt.Sprintf(
		"GET %s HTTP/1.0\r\nHost: %s\r\nUser-Agent: Mozilla/5.0 MetriyonAPI/WSDL\r\nAccept: application/wsdl+xml, application/xml, text/xml, */*\r\nConnection: close\r\n\r\n",
		path,
		hostHeader,
	)
	if _, err := io.WriteString(conn, request); err != nil {
		return nil, err
	}

	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: http.MethodGet})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %s", resp.Status)
	}
	return readLimitedWSDL(resp.Body)
}

func readLimitedWSDL(reader io.Reader) ([]byte, error) {
	limited := io.LimitReader(reader, maxWSDLSize+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if len(body) > maxWSDLSize {
		return nil, fmt.Errorf("WSDL excede o limite de %d MB", maxWSDLSize/(1024*1024))
	}
	return body, nil
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
