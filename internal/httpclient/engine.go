package httpclient

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"strings"
	"sync"
	"time"

	"forge-api-client/internal/models"
)

type Engine struct {
	mu      sync.Mutex
	cancels map[string]context.CancelFunc
}

func New() *Engine { return &Engine{cancels: map[string]context.CancelFunc{}} }

func (e *Engine) Cancel(executionID string) bool {
	e.mu.Lock()
	cancel, ok := e.cancels[executionID]
	e.mu.Unlock()
	if ok {
		cancel()
	}
	return ok
}

func (e *Engine) Execute(ctx context.Context, executionID string, request *http.Request, options models.RequestDefinition, maxPreview int64) (models.ExecuteRequestResult, error) {
	result := models.ExecuteRequestResult{ResolvedURL: request.URL.String(), Redirects: []models.Redirect{}}
	ctx, cancel := context.WithCancel(ctx)
	e.mu.Lock()
	e.cancels[executionID] = cancel
	e.mu.Unlock()
	defer func() { cancel(); e.mu.Lock(); delete(e.cancels, executionID); e.mu.Unlock() }()

	trace := newTrace()
	request = request.WithContext(httptrace.WithClientTrace(ctx, trace.clientTrace()))
	redirects := []models.Redirect{}
	isSOAP := strings.EqualFold(options.Protocol, "SOAP")
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2:     !isSOAP,
		DisableKeepAlives:     isSOAP,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: !options.VerifySSL}, // #nosec G402 -- explicit user setting
		TLSHandshakeTimeout:   15 * time.Second,
		ResponseHeaderTimeout: time.Duration(max(options.TimeoutSeconds, 1)) * time.Second,
		IdleConnTimeout:       90 * time.Second,
	}
	client := &http.Client{Transport: transport, Timeout: time.Duration(max(options.TimeoutSeconds, 1)) * time.Second}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		redirects = append(redirects, models.Redirect{StatusCode: req.Response.StatusCode, URL: req.URL.String()})
		if !options.FollowRedirect {
			return http.ErrUseLastResponse
		}
		if len(via) >= 10 {
			return fmt.Errorf("limite de 10 redirecionamentos excedido")
		}
		return nil
	}

	started := time.Now()
	response, err := client.Do(request)
	if err != nil {
		result.DurationMs = time.Since(started).Milliseconds()
		result.TechnicalError = err.Error()
		return result, classifyError(err, request.URL.Hostname())
	}
	defer response.Body.Close()
	trace.responseReceived = time.Now()
	limit := maxPreview
	if limit <= 0 {
		limit = 10 * 1024 * 1024
	}
	body, readErr := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if readErr != nil {
		return result, fmt.Errorf("não foi possível ler a resposta: %w", readErr)
	}
	trace.downloadDone = time.Now()
	if int64(len(body)) > limit {
		body = body[:limit]
		result.Truncated = true
	}
	result.StatusCode = response.StatusCode
	result.Status = response.Status
	result.Headers = response.Header.Clone()
	for _, cookie := range response.Cookies() {
		result.Cookies = append(result.Cookies, cookie.String())
	}
	result.ContentType = response.Header.Get("Content-Type")
	result.DurationMs = time.Since(started).Milliseconds()
	result.Size = response.ContentLength
	if result.Size < 0 {
		result.Size = int64(len(body))
	}
	result.Redirects = redirects
	result.Binary = !isTextContent(result.ContentType)
	if result.Binary {
		result.BodyBase64 = base64.StdEncoding.EncodeToString(body)
	} else {
		result.Body = string(body)
	}
	result.Timings = trace.timings(started)
	return result, nil
}

func isTextContent(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return strings.HasPrefix(contentType, "text/") || strings.Contains(contentType, "json") || strings.Contains(contentType, "xml") || strings.Contains(contentType, "javascript") || strings.Contains(contentType, "graphql") || contentType == ""
}

func classifyError(err error, host string) error {
	if err == context.Canceled || strings.Contains(err.Error(), "context canceled") {
		return fmt.Errorf("requisição cancelada")
	}
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return fmt.Errorf("tempo limite excedido ao conectar com %s", host)
	}
	return fmt.Errorf("não foi possível conectar ao servidor %s: %w", host, err)
}

type traceState struct{ dnsStart, dnsDone, connectStart, connectDone, tlsStart, tlsDone, wroteRequest, firstByte, responseReceived, downloadDone time.Time }

func newTrace() *traceState { return &traceState{} }
func (t *traceState) clientTrace() *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		DNSStart: func(httptrace.DNSStartInfo) { t.dnsStart = time.Now() }, DNSDone: func(httptrace.DNSDoneInfo) { t.dnsDone = time.Now() },
		ConnectStart: func(_, _ string) { t.connectStart = time.Now() }, ConnectDone: func(_, _ string, _ error) { t.connectDone = time.Now() },
		TLSHandshakeStart: func() { t.tlsStart = time.Now() }, TLSHandshakeDone: func(tls.ConnectionState, error) { t.tlsDone = time.Now() },
		WroteRequest: func(httptrace.WroteRequestInfo) { t.wroteRequest = time.Now() }, GotFirstResponseByte: func() { t.firstByte = time.Now() },
	}
}
func millis(start, end time.Time) int64 {
	if start.IsZero() || end.IsZero() {
		return 0
	}
	return end.Sub(start).Milliseconds()
}
func (t *traceState) timings(started time.Time) models.Timing {
	return models.Timing{DNS: millis(t.dnsStart, t.dnsDone), Connect: millis(t.connectStart, t.connectDone), TLS: millis(t.tlsStart, t.tlsDone), TTFB: millis(t.wroteRequest, t.firstByte), Download: millis(t.responseReceived, t.downloadDone), Total: millis(started, t.downloadDone)}
}
