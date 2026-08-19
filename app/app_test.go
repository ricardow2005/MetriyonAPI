package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"forge-api-client/internal/models"
	"forge-api-client/internal/security"
	"forge-api-client/internal/storage"
)

func TestRequestSecretsEncryptedAtRest(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.Open(filepath.Join(dir, "test.db"), os.DirFS(".."))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	protector, err := security.NewProtector(dir)
	if err != nil {
		t.Fatal(err)
	}
	application := New(store, protector)
	application.ctx = context.Background()
	now := time.Now().UTC()
	workspace := models.Workspace{ID: "workspace", Name: "Test", CreatedAt: now, UpdatedAt: now}
	if err := store.SaveWorkspace(workspace); err != nil {
		t.Fatal(err)
	}
	request := models.RequestDefinition{ID: "request", WorkspaceID: workspace.ID, Name: "Secret", Protocol: "REST", Method: "GET", URL: "https://example.test", Auth: models.AuthDefinition{Type: "bearer", Token: "plain-token"}, Headers: []models.KeyValue{{Enabled: true, Key: "Authorization", Value: "Bearer plain-token"}}, CreatedAt: now, UpdatedAt: now}
	if _, err := application.SaveRequest(request); err != nil {
		t.Fatal(err)
	}
	raw, err := store.LoadState()
	if err != nil {
		t.Fatal(err)
	}
	if raw.Requests[0].Auth.Token == "plain-token" || raw.Requests[0].Headers[0].Value == "Bearer plain-token" {
		t.Fatal("request secret was stored in plaintext")
	}
	state, err := application.Bootstrap()
	if err != nil {
		t.Fatal(err)
	}
	if state.Requests[0].Auth.Token != "plain-token" || state.Requests[0].Headers[0].Value != "Bearer plain-token" {
		t.Fatal("request secrets were not decrypted for use")
	}
	duplicated, err := application.DuplicateRequest(state.Requests[0])
	if err != nil {
		t.Fatal(err)
	}
	if duplicated.ID == state.Requests[0].ID || duplicated.Name != "Secret Copy" || duplicated.Auth.Token != "plain-token" || duplicated.URL != state.Requests[0].URL {
		t.Fatalf("request was not duplicated correctly: %#v", duplicated)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	defer server.Close()
	loadRequest := request
	loadRequest.ID = "load-request"
	loadRequest.URL = server.URL
	loadRequest.Auth = models.AuthDefinition{Type: "none"}
	loadRequest.Headers = nil
	run, err := application.RunLoadTest(models.LoadTestInput{ExecutionID: "load", Request: loadRequest, Concurrency: 3, TotalRequests: 12})
	if err != nil {
		t.Fatal(err)
	}
	if run.Result.Total != 12 || run.Result.Successful != 12 || run.Result.Failed != 0 || len(run.Result.Samples) != 12 {
		t.Fatalf("unexpected load test result: %#v", run.Result)
	}
}

func TestRunFlowExtractsJSONIntoNextRequest(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.Open(filepath.Join(dir, "test.db"), os.DirFS(".."))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	protector, err := security.NewProtector(dir)
	if err != nil {
		t.Fatal(err)
	}
	application := New(store, protector)
	application.ctx = context.Background()
	var receivedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if request.URL.Path == "/clients" {
			_, _ = w.Write([]byte(`{"client":{"nrCliente":42}}`))
			return
		}
		receivedPath = request.URL.Path
		_, _ = w.Write([]byte(`{"created":true}`))
	}))
	defer server.Close()
	now := time.Now().UTC()
	workspace := models.Workspace{ID: "workspace", Name: "Test", CreatedAt: now, UpdatedAt: now}
	if err := store.SaveWorkspace(workspace); err != nil {
		t.Fatal(err)
	}
	first := models.RequestDefinition{ID: "first", WorkspaceID: workspace.ID, Name: "requestClients", Protocol: "REST", Method: "GET", URL: server.URL + "/clients", Auth: models.AuthDefinition{Type: "none"}, Variables: map[string]string{}, CreatedAt: now, UpdatedAt: now}
	second := models.RequestDefinition{ID: "second", WorkspaceID: workspace.ID, Name: "postNewCliente", Protocol: "REST", Method: "POST", URL: server.URL + "/clients/{{nrCliente}}", Auth: models.AuthDefinition{Type: "none"}, Variables: map[string]string{}, CreatedAt: now, UpdatedAt: now}
	if _, err := application.SaveRequest(first); err != nil {
		t.Fatal(err)
	}
	if _, err := application.SaveRequest(second); err != nil {
		t.Fatal(err)
	}
	flow := models.Flow{ID: "flow", WorkspaceID: workspace.ID, Name: "Clients", Nodes: []models.FlowNode{{ID: "node-1", RequestID: first.ID, Name: first.Name, X: 100, Y: 100, Mappings: []models.FlowMapping{{Path: "client.nrCliente", Variable: "nrCliente"}}}, {ID: "node-2", RequestID: second.ID, Name: second.Name, X: 400, Y: 100}}, Edges: []models.FlowEdge{{ID: "edge", SourceID: "node-1", TargetID: "node-2"}}}
	result, err := application.RunFlow(models.FlowRunInput{ExecutionID: "execution", Flow: flow})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Successful || len(result.Nodes) != 2 || result.Variables["nrCliente"] != "42" || receivedPath != "/clients/42" {
		t.Fatalf("flow did not propagate response value: %#v, path=%q", result, receivedPath)
	}
}

func TestFlowRejectsCycles(t *testing.T) {
	flow := models.Flow{Nodes: []models.FlowNode{{ID: "a"}, {ID: "b"}}, Edges: []models.FlowEdge{{SourceID: "a", TargetID: "b"}, {SourceID: "b", TargetID: "a"}}}
	if _, err := flowExecutionOrder(flow); err == nil {
		t.Fatal("expected cyclic flow to be rejected")
	}
}

func TestExecuteRequestAutomaticallyRenewsOAuthToken(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.Open(filepath.Join(dir, "test.db"), os.DirFS(".."))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	protector, err := security.NewProtector(dir)
	if err != nil {
		t.Fatal(err)
	}
	application := New(store, protector)
	application.ctx = context.Background()
	tokenCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/oauth/token" {
			tokenCalls++
			_ = request.ParseForm()
			if request.Form.Get("grant_type") != "client_credentials" {
				t.Fatalf("unexpected OAuth grant: %q", request.Form.Get("grant_type"))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"auto-token","expires_in":3600}`))
			return
		}
		if request.Header.Get("Authorization") != "Bearer auto-token" {
			t.Fatalf("automatic access token was not applied: %q", request.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	now := time.Now().UTC()
	workspace := models.Workspace{ID: "oauth-workspace", Name: "OAuth", CreatedAt: now, UpdatedAt: now}
	if err := store.SaveWorkspace(workspace); err != nil {
		t.Fatal(err)
	}
	request := models.RequestDefinition{ID: "oauth-request", WorkspaceID: workspace.ID, Name: "OAuth request", Protocol: "REST", Method: "GET", URL: server.URL + "/resource", BodyType: "none", Variables: map[string]string{}, Auth: models.AuthDefinition{Type: "oauth2", OAuthFlow: "client_credentials", TokenURL: server.URL + "/oauth/token", ClientID: "client", ClientSecret: "secret", ClientAuth: "body"}, CreatedAt: now, UpdatedAt: now}
	if _, err := application.ExecuteRequest(models.ExecuteRequestInput{ExecutionID: "missing-token", Request: request}); err != nil {
		t.Fatal(err)
	}
	if tokenCalls != 1 {
		t.Fatalf("missing token should be acquired once, got %d calls", tokenCalls)
	}
	request.Auth.AccessToken = "auto-token"
	request.Auth.ExpiresAt = now.Add(time.Hour).Format(time.RFC3339)
	if _, err := application.ExecuteRequest(models.ExecuteRequestInput{ExecutionID: "valid-token", Request: request}); err != nil {
		t.Fatal(err)
	}
	if tokenCalls != 1 {
		t.Fatalf("valid token should be reused, got %d calls", tokenCalls)
	}
	request.Auth.ExpiresAt = now.Add(10 * time.Second).Format(time.RFC3339)
	if _, err := application.ExecuteRequest(models.ExecuteRequestInput{ExecutionID: "expiring-token", Request: request}); err != nil {
		t.Fatal(err)
	}
	if tokenCalls != 2 {
		t.Fatalf("expiring token should be renewed, got %d calls", tokenCalls)
	}
}

func TestRunFlowIfExecutesOnlySelectedBranch(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.Open(filepath.Join(dir, "test.db"), os.DirFS(".."))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	protector, err := security.NewProtector(dir)
	if err != nil {
		t.Fatal(err)
	}
	application := New(store, protector)
	application.ctx = context.Background()

	trueHits := 0
	falseHits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/source":
			_, _ = w.Write([]byte(`{"value":42}`))
		case "/true":
			trueHits++
			_, _ = w.Write([]byte(`{"branch":"true"}`))
		case "/false":
			falseHits++
			_, _ = w.Write([]byte(`{"branch":"false"}`))
		default:
			http.NotFound(w, request)
		}
	}))
	defer server.Close()

	now := time.Now().UTC()
	workspace := models.Workspace{ID: "if-workspace", Name: "IF", CreatedAt: now, UpdatedAt: now}
	if err := store.SaveWorkspace(workspace); err != nil {
		t.Fatal(err)
	}
	source := models.RequestDefinition{ID: "if-source", WorkspaceID: workspace.ID, Name: "Source", Protocol: "REST", Method: "GET", URL: server.URL + "/source", Auth: models.AuthDefinition{Type: "none"}, Variables: map[string]string{}, CreatedAt: now, UpdatedAt: now}
	trueRequest := models.RequestDefinition{ID: "if-true", WorkspaceID: workspace.ID, Name: "True", Protocol: "REST", Method: "GET", URL: server.URL + "/true", Auth: models.AuthDefinition{Type: "none"}, Variables: map[string]string{}, CreatedAt: now, UpdatedAt: now}
	falseRequest := models.RequestDefinition{ID: "if-false", WorkspaceID: workspace.ID, Name: "False", Protocol: "REST", Method: "GET", URL: server.URL + "/false", Auth: models.AuthDefinition{Type: "none"}, Variables: map[string]string{}, CreatedAt: now, UpdatedAt: now}
	for _, request := range []models.RequestDefinition{source, trueRequest, falseRequest} {
		if _, err := application.SaveRequest(request); err != nil {
			t.Fatal(err)
		}
	}

	flow := models.Flow{
		ID:          "if-flow",
		WorkspaceID: workspace.ID,
		Name:        "Conditional",
		Nodes: []models.FlowNode{
			{ID: "source", Type: "request", RequestID: source.ID, Name: source.Name, X: 100, Y: 100},
			{ID: "if", Type: "if", Name: "If", X: 400, Y: 100, ConvertTypes: true, Conditions: []models.FlowCondition{{Left: "body.value", Operator: "greater_than", Right: "40"}}},
			{ID: "true", Type: "request", RequestID: trueRequest.ID, Name: trueRequest.Name, X: 700, Y: 40},
			{ID: "false", Type: "request", RequestID: falseRequest.ID, Name: falseRequest.Name, X: 700, Y: 180},
		},
		Edges: []models.FlowEdge{
			{ID: "source-if", SourceID: "source", TargetID: "if", SourceHandle: "default"},
			{ID: "if-true", SourceID: "if", TargetID: "true", SourceHandle: "true"},
			{ID: "if-false", SourceID: "if", TargetID: "false", SourceHandle: "false"},
		},
	}

	result, err := application.RunFlow(models.FlowRunInput{ExecutionID: "if-execution", Flow: flow})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Successful {
		t.Fatalf("expected successful flow: %#v", result)
	}
	if trueHits != 1 || falseHits != 0 {
		t.Fatalf("wrong branch executed: true=%d false=%d", trueHits, falseHits)
	}
	var ifResult, falseResult *models.FlowNodeResult
	for index := range result.Nodes {
		switch result.Nodes[index].NodeID {
		case "if":
			ifResult = &result.Nodes[index]
		case "false":
			falseResult = &result.Nodes[index]
		}
	}
	if ifResult == nil || ifResult.Outcome != "true" {
		t.Fatalf("expected IF outcome true: %#v", ifResult)
	}
	if falseResult == nil || !falseResult.Skipped {
		t.Fatalf("expected false branch to be skipped: %#v", falseResult)
	}
}

func TestRunFlowExtractNodeCreatesVariableForNextRequest(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.Open(filepath.Join(dir, "test.db"), os.DirFS(".."))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	protector, err := security.NewProtector(dir)
	if err != nil {
		t.Fatal(err)
	}
	application := New(store, protector)
	application.ctx = context.Background()

	receivedPath := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if request.URL.Path == "/source" {
			_, _ = w.Write([]byte(`{"client":{"id":99}}`))
			return
		}
		receivedPath = request.URL.Path
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	now := time.Now().UTC()
	workspace := models.Workspace{ID: "extract-workspace", Name: "Extract", CreatedAt: now, UpdatedAt: now}
	if err := store.SaveWorkspace(workspace); err != nil {
		t.Fatal(err)
	}
	source := models.RequestDefinition{ID: "extract-source", WorkspaceID: workspace.ID, Name: "Source", Protocol: "REST", Method: "GET", URL: server.URL + "/source", Auth: models.AuthDefinition{Type: "none"}, Variables: map[string]string{}, CreatedAt: now, UpdatedAt: now}
	next := models.RequestDefinition{ID: "extract-next", WorkspaceID: workspace.ID, Name: "Next", Protocol: "REST", Method: "GET", URL: server.URL + "/clients/{{clientId}}", Auth: models.AuthDefinition{Type: "none"}, Variables: map[string]string{}, CreatedAt: now, UpdatedAt: now}
	for _, request := range []models.RequestDefinition{source, next} {
		if _, err := application.SaveRequest(request); err != nil {
			t.Fatal(err)
		}
	}

	flow := models.Flow{
		ID:          "extract-flow",
		WorkspaceID: workspace.ID,
		Name:        "Extract data",
		Nodes: []models.FlowNode{
			{ID: "source", Type: "request", RequestID: source.ID, Name: source.Name, X: 100, Y: 100},
			{ID: "extract", Type: "extract", Name: "Extract Data", X: 400, Y: 100, Mappings: []models.FlowMapping{{Path: "body.client.id", Variable: "clientId"}}},
			{ID: "next", Type: "request", RequestID: next.ID, Name: next.Name, X: 700, Y: 100},
		},
		Edges: []models.FlowEdge{
			{ID: "source-extract", SourceID: "source", TargetID: "extract"},
			{ID: "extract-next", SourceID: "extract", TargetID: "next"},
		},
	}

	result, err := application.RunFlow(models.FlowRunInput{ExecutionID: "extract-execution", Flow: flow})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Successful || result.Variables["clientId"] != "99" || receivedPath != "/clients/99" {
		t.Fatalf("extract node did not propagate variable: result=%#v path=%q", result, receivedPath)
	}
}

func TestRunFlowFilterSkipsDownstreamWhenConditionFails(t *testing.T) {
	response := models.ExecuteRequestResult{StatusCode: 500, Body: `{"ok":false}`}
	node := models.FlowNode{
		ID:           "filter",
		Type:         "filter",
		Name:         "Filter",
		ConvertTypes: true,
		Conditions:   []models.FlowCondition{{Left: "status", Operator: "less_than", Right: "400"}},
	}
	passed, err := evaluateFlowConditions(node, response, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	if passed {
		t.Fatal("500 should not pass a status < 400 filter")
	}
}
