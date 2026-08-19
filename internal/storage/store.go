package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	"forge-api-client/internal/models"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(path string, migrations fs.FS) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.migrate(migrations); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate(migrations fs.FS) error {
	if _, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		return fmt.Errorf("prepare migrations: %w", err)
	}
	entries, err := fs.ReadDir(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		var exists int
		if err := s.db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version=?`, entry.Name()).Scan(&exists); err != nil {
			return err
		}
		if exists > 0 {
			continue
		}
		body, err := fs.ReadFile(migrations, "migrations/"+entry.Name())
		if err != nil {
			return err
		}
		tx, err := s.db.Begin()
		if err != nil {
			return err
		}
		if _, err = tx.Exec(string(body)); err == nil {
			_, err = tx.Exec(`INSERT INTO schema_migrations(version,applied_at) VALUES(?,?)`, entry.Name(), time.Now().UTC().Format(time.RFC3339Nano))
		}
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", entry.Name(), err)
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func defaultSettings() models.Settings {
	return models.Settings{Theme: "dark", StoreHistoryBodies: true, HistoryRetentionDays: 30, MaxResponsePreviewSize: 10 * 1024 * 1024, SidebarWidth: 288, EditorSplitPercent: 52, LogLevel: "INFO", WindowWidth: 1440, WindowHeight: 900}
}

func (s *Store) LoadState() (models.AppState, error) {
	state := models.AppState{Settings: defaultSettings()}
	var err error
	if state.Workspaces, err = s.listWorkspaces(); err != nil {
		return state, err
	}
	if state.Collections, err = s.listCollections(); err != nil {
		return state, err
	}
	if state.Folders, err = s.listFolders(); err != nil {
		return state, err
	}
	if state.Requests, err = s.listRequests(); err != nil {
		return state, err
	}
	if state.Environments, err = s.listEnvironments(); err != nil {
		return state, err
	}
	if state.History, err = s.listHistory(250); err != nil {
		return state, err
	}
	if state.LoadTests, err = s.listLoadTests(100); err != nil {
		return state, err
	}
	if state.Flows, err = s.listFlows(); err != nil {
		return state, err
	}
	if state.OpenTabs, err = s.listTabs(); err != nil {
		return state, err
	}
	var settingsJSON string
	if err := s.db.QueryRow(`SELECT value_json FROM settings WHERE key='app'`).Scan(&settingsJSON); err == nil {
		_ = json.Unmarshal([]byte(settingsJSON), &state.Settings)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return state, err
	}
	return state, nil
}

func (s *Store) SaveWorkspace(v models.Workspace) error {
	_, err := s.db.Exec(`INSERT INTO workspaces(id,name,created_at,updated_at) VALUES(?,?,?,?) ON CONFLICT(id) DO UPDATE SET name=excluded.name,updated_at=excluded.updated_at`, v.ID, v.Name, v.CreatedAt.UTC().Format(time.RFC3339Nano), v.UpdatedAt.UTC().Format(time.RFC3339Nano))
	return err
}

func (s *Store) SaveCollection(v models.Collection) error {
	b, _ := json.Marshal(v.Variables)
	_, err := s.db.Exec(`INSERT INTO collections(id,workspace_id,name,description,variables_json,sort_order) VALUES(?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET workspace_id=excluded.workspace_id,name=excluded.name,description=excluded.description,variables_json=excluded.variables_json,sort_order=excluded.sort_order`, v.ID, v.WorkspaceID, v.Name, v.Description, string(b), v.SortOrder)
	return err
}

func (s *Store) SaveFolder(v models.Folder) error {
	_, err := s.db.Exec(`INSERT INTO folders(id,collection_id,parent_id,name,sort_order) VALUES(?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET collection_id=excluded.collection_id,parent_id=excluded.parent_id,name=excluded.name,sort_order=excluded.sort_order`, v.ID, v.CollectionID, v.ParentID, v.Name, v.SortOrder)
	return err
}

func (s *Store) SaveRequest(v models.RequestDefinition) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`INSERT INTO requests(id,workspace_id,collection_id,folder_id,name,definition_json,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET workspace_id=excluded.workspace_id,collection_id=excluded.collection_id,folder_id=excluded.folder_id,name=excluded.name,definition_json=excluded.definition_json,updated_at=excluded.updated_at`, v.ID, v.WorkspaceID, v.CollectionID, v.FolderID, v.Name, string(b), v.CreatedAt.UTC().Format(time.RFC3339Nano), v.UpdatedAt.UTC().Format(time.RFC3339Nano))
	return err
}

func (s *Store) SaveEnvironment(v models.Environment) error {
	b, _ := json.Marshal(v.Variables)
	_, err := s.db.Exec(`INSERT INTO environments(id,workspace_id,name,variables_json) VALUES(?,?,?,?) ON CONFLICT(id) DO UPDATE SET workspace_id=excluded.workspace_id,name=excluded.name,variables_json=excluded.variables_json`, v.ID, v.WorkspaceID, v.Name, string(b))
	return err
}
func (s *Store) SaveHistory(v models.HistoryEntry) error {
	rq, _ := json.Marshal(v.Request)
	rs, _ := json.Marshal(v.Response)
	_, err := s.db.Exec(`INSERT INTO history(id,workspace_id,executed_at,method,url,name,status_code,duration_ms,size,request_json,response_json) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, v.ID, v.WorkspaceID, v.ExecutedAt.UTC().Format(time.RFC3339Nano), v.Method, v.URL, v.Name, v.StatusCode, v.DurationMs, v.Size, string(rq), string(rs))
	return err
}
func (s *Store) SaveLoadTest(v models.LoadTestRun) error {
	result, err := json.Marshal(v.Result)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`INSERT INTO load_test_runs(id,workspace_id,request_id,request_name,started_at,concurrency,total_requests,delay_ms,result_json) VALUES(?,?,?,?,?,?,?,?,?)`, v.ID, v.WorkspaceID, v.RequestID, v.RequestName, v.StartedAt.UTC().Format(time.RFC3339Nano), v.Concurrency, v.TotalRequests, v.DelayMs, string(result))
	return err
}

func (s *Store) SaveFlow(v models.Flow) error {
	definition, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`INSERT INTO flows(id,workspace_id,name,definition_json,created_at,updated_at) VALUES(?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET workspace_id=excluded.workspace_id,name=excluded.name,definition_json=excluded.definition_json,updated_at=excluded.updated_at`, v.ID, v.WorkspaceID, v.Name, string(definition), v.CreatedAt.UTC().Format(time.RFC3339Nano), v.UpdatedAt.UTC().Format(time.RFC3339Nano))
	return err
}

func (s *Store) ClearLoadTests(requestID string) error {
	_, err := s.db.Exec(`DELETE FROM load_test_runs WHERE request_id=?`, requestID)
	return err
}
func (s *Store) SaveSettings(v models.Settings) error {
	b, _ := json.Marshal(v)
	_, err := s.db.Exec(`INSERT INTO settings(key,value_json) VALUES('app',?) ON CONFLICT(key) DO UPDATE SET value_json=excluded.value_json`, string(b))
	return err
}
func (s *Store) SaveTabs(ids []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM open_tabs`); err != nil {
		tx.Rollback()
		return err
	}
	for i, id := range ids {
		if _, err = tx.Exec(`INSERT INTO open_tabs(request_id,sort_order) VALUES(?,?)`, id, i); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) Delete(kind, id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	rollback := func(cause error) error { _ = tx.Rollback(); return cause }
	switch kind {
	case "request":
		if _, err = tx.Exec(`DELETE FROM open_tabs WHERE request_id=?`, id); err != nil {
			return rollback(err)
		}
		_, err = tx.Exec(`DELETE FROM requests WHERE id=?`, id)
	case "folder":
		rows, queryErr := tx.Query(`SELECT id,definition_json FROM requests WHERE folder_id=?`, id)
		if queryErr != nil {
			return rollback(queryErr)
		}
		type update struct{ id, definition string }
		updates := []update{}
		for rows.Next() {
			var requestID, raw string
			if scanErr := rows.Scan(&requestID, &raw); scanErr != nil {
				rows.Close()
				return rollback(scanErr)
			}
			var request models.RequestDefinition
			if json.Unmarshal([]byte(raw), &request) == nil {
				request.FolderID = ""
				body, _ := json.Marshal(request)
				updates = append(updates, update{requestID, string(body)})
			}
		}
		if rowsErr := rows.Err(); rowsErr != nil {
			rows.Close()
			return rollback(rowsErr)
		}
		rows.Close()
		for _, item := range updates {
			if _, err = tx.Exec(`UPDATE requests SET folder_id='',definition_json=? WHERE id=?`, item.definition, item.id); err != nil {
				return rollback(err)
			}
		}
		_, err = tx.Exec(`DELETE FROM folders WHERE id=?`, id)
	case "collection":
		if _, err = tx.Exec(`DELETE FROM open_tabs WHERE request_id IN (SELECT id FROM requests WHERE collection_id=?)`, id); err != nil {
			return rollback(err)
		}
		if _, err = tx.Exec(`DELETE FROM requests WHERE collection_id=?`, id); err != nil {
			return rollback(err)
		}
		_, err = tx.Exec(`DELETE FROM collections WHERE id=?`, id)
	case "workspace":
		if _, err = tx.Exec(`DELETE FROM open_tabs WHERE request_id IN (SELECT id FROM requests WHERE workspace_id=?)`, id); err != nil {
			return rollback(err)
		}
		_, err = tx.Exec(`DELETE FROM workspaces WHERE id=?`, id)
	case "environment":
		_, err = tx.Exec(`DELETE FROM environments WHERE id=?`, id)
	case "flow":
		_, err = tx.Exec(`DELETE FROM flows WHERE id=?`, id)
	default:
		return rollback(fmt.Errorf("unsupported entity type %q", kind))
	}
	if err != nil {
		return rollback(err)
	}
	return tx.Commit()
}

func (s *Store) MoveFolder(input models.MoveFolderInput) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	rollback := func(cause error) error { _ = tx.Rollback(); return cause }
	rows, err := tx.Query(`SELECT id,definition_json FROM requests WHERE folder_id=?`, input.FolderID)
	if err != nil {
		return rollback(err)
	}
	type update struct{ id, definition string }
	updates := []update{}
	for rows.Next() {
		var id, raw string
		if scanErr := rows.Scan(&id, &raw); scanErr != nil {
			rows.Close()
			return rollback(scanErr)
		}
		var request models.RequestDefinition
		if json.Unmarshal([]byte(raw), &request) == nil {
			request.CollectionID = input.CollectionID
			body, _ := json.Marshal(request)
			updates = append(updates, update{id, string(body)})
		}
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		rows.Close()
		return rollback(rowsErr)
	}
	rows.Close()
	if _, err = tx.Exec(`UPDATE folders SET collection_id=?,sort_order=? WHERE id=?`, input.CollectionID, input.SortOrder, input.FolderID); err != nil {
		return rollback(err)
	}
	for _, item := range updates {
		if _, err = tx.Exec(`UPDATE requests SET collection_id=?,definition_json=? WHERE id=?`, input.CollectionID, item.definition, item.id); err != nil {
			return rollback(err)
		}
	}
	return tx.Commit()
}
func (s *Store) ClearHistory(workspaceID string) error {
	_, err := s.db.Exec(`DELETE FROM history WHERE workspace_id=?`, workspaceID)
	return err
}

func (s *Store) PruneHistory(workspaceID string, retentionDays int) error {
	if retentionDays <= 0 {
		return nil
	}
	cutoff := time.Now().UTC().AddDate(0, 0, -retentionDays).Format(time.RFC3339Nano)
	_, err := s.db.Exec(`DELETE FROM history WHERE workspace_id=? AND executed_at<?`, workspaceID, cutoff)
	return err
}

func (s *Store) listWorkspaces() ([]models.Workspace, error) {
	rows, err := s.db.Query(`SELECT id,name,created_at,updated_at FROM workspaces ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Workspace{}
	for rows.Next() {
		var v models.Workspace
		var c, u string
		if err := rows.Scan(&v.ID, &v.Name, &c, &u); err != nil {
			return nil, err
		}
		v.CreatedAt, _ = time.Parse(time.RFC3339Nano, c)
		v.UpdatedAt, _ = time.Parse(time.RFC3339Nano, u)
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) listCollections() ([]models.Collection, error) {
	rows, err := s.db.Query(`SELECT id,workspace_id,name,description,variables_json,sort_order FROM collections ORDER BY sort_order,name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Collection{}
	for rows.Next() {
		var v models.Collection
		var j string
		if err := rows.Scan(&v.ID, &v.WorkspaceID, &v.Name, &v.Description, &j, &v.SortOrder); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(j), &v.Variables)
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) listFolders() ([]models.Folder, error) {
	rows, err := s.db.Query(`SELECT id,collection_id,parent_id,name,sort_order FROM folders ORDER BY sort_order,name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Folder{}
	for rows.Next() {
		var v models.Folder
		if err := rows.Scan(&v.ID, &v.CollectionID, &v.ParentID, &v.Name, &v.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) listRequests() ([]models.RequestDefinition, error) {
	rows, err := s.db.Query(`SELECT definition_json FROM requests ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.RequestDefinition{}
	for rows.Next() {
		var j string
		var v models.RequestDefinition
		if err := rows.Scan(&j); err != nil {
			return nil, err
		}
		if json.Unmarshal([]byte(j), &v) == nil {
			out = append(out, v)
		}
	}
	return out, rows.Err()
}
func (s *Store) listEnvironments() ([]models.Environment, error) {
	rows, err := s.db.Query(`SELECT id,workspace_id,name,variables_json FROM environments ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Environment{}
	for rows.Next() {
		var v models.Environment
		var j string
		if err := rows.Scan(&v.ID, &v.WorkspaceID, &v.Name, &j); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(j), &v.Variables)
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) listHistory(limit int) ([]models.HistoryEntry, error) {
	rows, err := s.db.Query(`SELECT id,workspace_id,executed_at,method,url,name,status_code,duration_ms,size,request_json,response_json FROM history ORDER BY executed_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.HistoryEntry{}
	for rows.Next() {
		var v models.HistoryEntry
		var at, rq, rs string
		if err := rows.Scan(&v.ID, &v.WorkspaceID, &at, &v.Method, &v.URL, &v.Name, &v.StatusCode, &v.DurationMs, &v.Size, &rq, &rs); err != nil {
			return nil, err
		}
		v.ExecutedAt, _ = time.Parse(time.RFC3339Nano, at)
		_ = json.Unmarshal([]byte(rq), &v.Request)
		_ = json.Unmarshal([]byte(rs), &v.Response)
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) listLoadTests(limit int) ([]models.LoadTestRun, error) {
	rows, err := s.db.Query(`SELECT id,workspace_id,request_id,request_name,started_at,concurrency,total_requests,delay_ms,result_json FROM load_test_runs ORDER BY started_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.LoadTestRun{}
	for rows.Next() {
		var v models.LoadTestRun
		var started, raw string
		if err := rows.Scan(&v.ID, &v.WorkspaceID, &v.RequestID, &v.RequestName, &started, &v.Concurrency, &v.TotalRequests, &v.DelayMs, &raw); err != nil {
			return nil, err
		}
		v.StartedAt, _ = time.Parse(time.RFC3339Nano, started)
		_ = json.Unmarshal([]byte(raw), &v.Result)
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) listFlows() ([]models.Flow, error) {
	rows, err := s.db.Query(`SELECT definition_json FROM flows ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Flow{}
	for rows.Next() {
		var raw string
		var flow models.Flow
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		if json.Unmarshal([]byte(raw), &flow) == nil {
			out = append(out, flow)
		}
	}
	return out, rows.Err()
}
func (s *Store) listTabs() ([]string, error) {
	rows, err := s.db.Query(`SELECT request_id FROM open_tabs ORDER BY sort_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
