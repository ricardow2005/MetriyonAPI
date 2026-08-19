package storage

import (
	"testing"
	"testing/fstest"
	"time"

	"forge-api-client/internal/models"
)

const testSchema = `CREATE TABLE IF NOT EXISTS workspaces(id TEXT PRIMARY KEY,name TEXT NOT NULL,created_at TEXT NOT NULL,updated_at TEXT NOT NULL);CREATE TABLE IF NOT EXISTS collections(id TEXT PRIMARY KEY,workspace_id TEXT NOT NULL,name TEXT NOT NULL,description TEXT NOT NULL DEFAULT '',variables_json TEXT NOT NULL DEFAULT '{}',sort_order INTEGER NOT NULL DEFAULT 0);CREATE TABLE IF NOT EXISTS folders(id TEXT PRIMARY KEY,collection_id TEXT NOT NULL,parent_id TEXT NOT NULL DEFAULT '',name TEXT NOT NULL,sort_order INTEGER NOT NULL DEFAULT 0);CREATE TABLE IF NOT EXISTS requests(id TEXT PRIMARY KEY,workspace_id TEXT NOT NULL,collection_id TEXT NOT NULL DEFAULT '',folder_id TEXT NOT NULL DEFAULT '',name TEXT NOT NULL,definition_json TEXT NOT NULL,created_at TEXT NOT NULL,updated_at TEXT NOT NULL);CREATE TABLE IF NOT EXISTS environments(id TEXT PRIMARY KEY,workspace_id TEXT NOT NULL,name TEXT NOT NULL,variables_json TEXT NOT NULL DEFAULT '[]');CREATE TABLE IF NOT EXISTS history(id TEXT PRIMARY KEY,workspace_id TEXT NOT NULL,executed_at TEXT NOT NULL,method TEXT NOT NULL,url TEXT NOT NULL,name TEXT NOT NULL,status_code INTEGER NOT NULL,duration_ms INTEGER NOT NULL,size INTEGER NOT NULL,request_json TEXT NOT NULL,response_json TEXT NOT NULL);CREATE TABLE IF NOT EXISTS load_test_runs(id TEXT PRIMARY KEY,workspace_id TEXT NOT NULL,request_id TEXT NOT NULL,request_name TEXT NOT NULL,started_at TEXT NOT NULL,concurrency INTEGER NOT NULL,total_requests INTEGER NOT NULL,delay_ms INTEGER NOT NULL,result_json TEXT NOT NULL);CREATE TABLE IF NOT EXISTS flows(id TEXT PRIMARY KEY,workspace_id TEXT NOT NULL,name TEXT NOT NULL,definition_json TEXT NOT NULL,created_at TEXT NOT NULL,updated_at TEXT NOT NULL);CREATE TABLE IF NOT EXISTS settings(key TEXT PRIMARY KEY,value_json TEXT NOT NULL);CREATE TABLE IF NOT EXISTS open_tabs(request_id TEXT PRIMARY KEY,sort_order INTEGER NOT NULL);`

func TestStorePersistsState(t *testing.T) {
	migrations := fstest.MapFS{"migrations/001.sql": {Data: []byte(testSchema)}}
	store, err := Open(t.TempDir()+"/test.db", migrations)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	now := time.Now().UTC()
	workspace := models.Workspace{ID: "workspace", Name: "Test", CreatedAt: now, UpdatedAt: now}
	if err := store.SaveWorkspace(workspace); err != nil {
		t.Fatal(err)
	}
	request := models.RequestDefinition{ID: "request", WorkspaceID: workspace.ID, Name: "Get", Method: "GET", Protocol: "REST", URL: "https://example.test", CreatedAt: now, UpdatedAt: now}
	if err := store.SaveRequest(request); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveTabs([]string{request.ID}); err != nil {
		t.Fatal(err)
	}
	flow := models.Flow{ID: "flow", WorkspaceID: workspace.ID, Name: "Client flow", Nodes: []models.FlowNode{{ID: "node", RequestID: request.ID, Name: request.Name}}, Edges: []models.FlowEdge{}, CreatedAt: now, UpdatedAt: now}
	if err := store.SaveFlow(flow); err != nil {
		t.Fatal(err)
	}
	state, err := store.LoadState()
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Workspaces) != 1 || len(state.Requests) != 1 || len(state.OpenTabs) != 1 || len(state.Flows) != 1 {
		t.Fatalf("state not restored: %#v", state)
	}
	run := models.LoadTestRun{ID: "run", WorkspaceID: workspace.ID, RequestID: request.ID, RequestName: request.Name, StartedAt: now, Concurrency: 2, TotalRequests: 4, Result: models.LoadTestResult{Total: 4, Successful: 4}}
	if err := store.SaveLoadTest(run); err != nil {
		t.Fatal(err)
	}
	state, err = store.LoadState()
	if err != nil || len(state.LoadTests) != 1 {
		t.Fatalf("load test not persisted: %#v %v", state.LoadTests, err)
	}
	if err := store.ClearLoadTests(request.ID); err != nil {
		t.Fatal(err)
	}
	state, err = store.LoadState()
	if err != nil || len(state.LoadTests) != 0 {
		t.Fatalf("load test scenario not cleared: %#v %v", state.LoadTests, err)
	}
	collection := models.Collection{ID: "collection", WorkspaceID: workspace.ID, Name: "Collection", Variables: map[string]string{}}
	if err := store.SaveCollection(collection); err != nil {
		t.Fatal(err)
	}
	folder := models.Folder{ID: "folder", CollectionID: collection.ID, Name: "Folder"}
	if err := store.SaveFolder(folder); err != nil {
		t.Fatal(err)
	}
	request.CollectionID = collection.ID
	request.FolderID = folder.ID
	if err := store.SaveRequest(request); err != nil {
		t.Fatal(err)
	}
	targetCollection := models.Collection{ID: "target", WorkspaceID: workspace.ID, Name: "Target", Variables: map[string]string{}}
	if err := store.SaveCollection(targetCollection); err != nil {
		t.Fatal(err)
	}
	if err := store.MoveFolder(models.MoveFolderInput{FolderID: folder.ID, CollectionID: targetCollection.ID, SortOrder: 3}); err != nil {
		t.Fatal(err)
	}
	state, err = store.LoadState()
	if err != nil {
		t.Fatal(err)
	}
	if state.Folders[0].CollectionID != targetCollection.ID || state.Requests[0].CollectionID != targetCollection.ID {
		t.Fatalf("folder contents were not moved together: %#v %#v", state.Folders, state.Requests)
	}
	if err := store.Delete("folder", folder.ID); err != nil {
		t.Fatal(err)
	}
	state, err = store.LoadState()
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Requests) != 1 || state.Requests[0].FolderID != "" {
		t.Fatalf("folder deletion left an orphan request: %#v", state.Requests)
	}
	if err := store.Delete("collection", targetCollection.ID); err != nil {
		t.Fatal(err)
	}
	state, err = store.LoadState()
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Requests) != 0 || len(state.OpenTabs) != 0 {
		t.Fatalf("collection deletion did not cascade: %#v", state)
	}
}
