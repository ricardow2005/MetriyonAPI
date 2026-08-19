CREATE TABLE IF NOT EXISTS load_test_runs (
    id TEXT PRIMARY KEY,
    workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    request_id TEXT NOT NULL,
    request_name TEXT NOT NULL,
    started_at TEXT NOT NULL,
    concurrency INTEGER NOT NULL,
    total_requests INTEGER NOT NULL,
    delay_ms INTEGER NOT NULL,
    result_json TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_load_test_request_date ON load_test_runs(request_id, started_at DESC);
