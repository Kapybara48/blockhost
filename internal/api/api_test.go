package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"blockhost/internal/model"
	"blockhost/internal/store"
)

func setupTestServer(t *testing.T) (*Server, *store.Store, func()) {
	tempDir, err := os.MkdirTemp("", "blockhost-api-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	st, err := store.New(tempDir)
	if err != nil {
		t.Fatalf("failed to init store: %v", err)
	}

	srv := NewServer(st, ".")
	cleanup := func() {
		st.Close()
		os.RemoveAll(tempDir)
	}
	return srv, st, cleanup
}

func TestAPIEndpoints(t *testing.T) {
	srv, _, cleanup := setupTestServer(t)
	defer cleanup()

	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)

	// 1. GET /api/state
	req := httptest.NewRequest("GET", "/api/state", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/state returned code %d", rec.Code)
	}
	var state model.AppState
	if err := json.Unmarshal(rec.Body.Bytes(), &state); err != nil {
		t.Fatalf("failed to parse state json: %v", err)
	}
	if state.ServerStatus != "online" {
		t.Errorf("expected serverStatus 'online', got '%s'", state.ServerStatus)
	}

	// 2. POST /api/server/stop
	req = httptest.NewRequest("POST", "/api/server/stop", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("POST /api/server/stop returned %d", rec.Code)
	}

	// 3. POST /api/profiles/activate
	body, _ := json.Marshal(map[string]string{"profileId": "ultra"})
	req = httptest.NewRequest("POST", "/api/profiles/activate", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("POST /api/profiles/activate returned %d", rec.Code)
	}

	// 4. POST /api/console
	body, _ = json.Marshal(map[string]string{"command": "help"})
	req = httptest.NewRequest("POST", "/api/console", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("POST /api/console returned %d", rec.Code)
	}

	// 5. POST /api/backups
	req = httptest.NewRequest("POST", "/api/backups", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /api/backups returned %d", rec.Code)
	}
	var createdBackup model.Backup
	if err := json.Unmarshal(rec.Body.Bytes(), &createdBackup); err != nil {
		t.Fatalf("failed to parse backup json: %v", err)
	}

	// 6. GET /api/backups/{id}/download
	req = httptest.NewRequest("GET", "/api/backups/"+createdBackup.ID+"/download", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("download backup returned %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/gzip" {
		t.Errorf("expected Content-Type application/gzip, got %s", rec.Header().Get("Content-Type"))
	}
}
