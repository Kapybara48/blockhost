package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"blockhost/internal/model"
	"blockhost/internal/rcon"
	"blockhost/internal/store"
)

// Server encapsulates the API handler and services.
type Server struct {
	store    *store.Store
	executor *rcon.Executor
	webDir   string
}

// NewServer creates a new API server.
func NewServer(s *store.Store, webDir string) *Server {
	return &Server{
		store:    s,
		executor: rcon.NewExecutor(s),
		webDir:   webDir,
	}
}

// WriteJSON sends a JSON response with status code.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// WriteError sends a JSON error response.
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}

// GetState returns current server state.
func (s *Server) GetState(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, s.store.GetState())
}

// Server Lifecycle
func (s *Server) StartServer(w http.ResponseWriter, r *http.Request) {
	s.store.StartServer()
	WriteJSON(w, http.StatusOK, map[string]string{"status": "starting"})
}

func (s *Server) StopServer(w http.ResponseWriter, r *http.Request) {
	s.store.StopServer()
	WriteJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

func (s *Server) RestartServer(w http.ResponseWriter, r *http.Request) {
	s.store.RestartServer()
	WriteJSON(w, http.StatusOK, map[string]string{"status": "restarting"})
}

// Profile Handlers
func (s *Server) ActivateProfile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProfileID string `json:"profileId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ProfileID == "" {
		WriteError(w, http.StatusBadRequest, "Invalid request body: profileId required")
		return
	}

	p, err := s.store.ActivateProfile(req.ProfileID, false)
	if err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, p)
}

func (s *Server) CreateProfile(w http.ResponseWriter, r *http.Request) {
	var p model.Profile
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || p.Name == "" {
		WriteError(w, http.StatusBadRequest, "Invalid profile payload")
		return
	}

	if err := s.store.AddProfile(p); err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, p)
}

// Event Boost Handlers
func (s *Server) LaunchEvent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name         string `json:"name"`
		ProfileID    string `json:"profileId"`
		DurationMins int    `json:"durationMins"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ProfileID == "" {
		WriteError(w, http.StatusBadRequest, "Invalid event payload")
		return
	}
	if req.DurationMins <= 0 {
		req.DurationMins = 60
	}
	if req.Name == "" {
		req.Name = "Komunitní Turnaj"
	}

	if err := s.store.LaunchEvent(req.Name, req.ProfileID, req.DurationMins); err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "launched"})
}

func (s *Server) EndEvent(w http.ResponseWriter, r *http.Request) {
	if err := s.store.EndEvent(); err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ended"})
}

// Scheduler Handlers
func (s *Server) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	var sch model.Schedule
	if err := json.NewDecoder(r.Body).Decode(&sch); err != nil || sch.Name == "" || len(sch.Days) == 0 {
		WriteError(w, http.StatusBadRequest, "Invalid schedule payload")
		return
	}
	sch.Enabled = true

	if err := s.store.AddSchedule(sch); err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, sch)
}

func (s *Server) ToggleSchedule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid payload")
		return
	}

	if err := s.store.ToggleSchedule(id, req.Enabled); err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (s *Server) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.store.DeleteSchedule(id); err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// Player Handlers
func (s *Server) KickPlayer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		WriteError(w, http.StatusBadRequest, "Player name required")
		return
	}

	if !s.store.KickPlayer(req.Name) {
		WriteError(w, http.StatusNotFound, "Player not found or offline")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "kicked"})
}

func (s *Server) BanPlayer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name   string `json:"name"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		WriteError(w, http.StatusBadRequest, "Player name required")
		return
	}

	s.store.BanPlayer(req.Name, req.Reason)
	WriteJSON(w, http.StatusOK, map[string]string{"status": "banned"})
}

func (s *Server) UnbanPlayer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		WriteError(w, http.StatusBadRequest, "Player name required")
		return
	}

	s.store.UnbanPlayer(req.Name)
	WriteJSON(w, http.StatusOK, map[string]string{"status": "unbanned"})
}

func (s *Server) TogglePlayerOP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		WriteError(w, http.StatusBadRequest, "Player name required")
		return
	}

	ok, isOP := s.store.TogglePlayerOP(req.Name)
	if !ok {
		WriteError(w, http.StatusNotFound, "Player not found")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{"status": "updated", "op": isOP})
}

func (s *Server) SimulateJoin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		WriteError(w, http.StatusBadRequest, "Player name required")
		return
	}

	if err := s.store.AddPlayerJoin(req.Name); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "joined"})
}

func (s *Server) AddWhitelist(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		WriteError(w, http.StatusBadRequest, "Player name required")
		return
	}

	s.store.AddWhitelist(req.Name)
	WriteJSON(w, http.StatusOK, map[string]string{"status": "added"})
}

func (s *Server) RemoveWhitelist(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		WriteError(w, http.StatusBadRequest, "Player name required")
		return
	}

	s.store.RemoveWhitelist(name)
	WriteJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

// Config Handlers
func (s *Server) GetConfig(w http.ResponseWriter, r *http.Request) {
	state := s.store.GetState()
	WriteJSON(w, http.StatusOK, state.Config)
}

func (s *Server) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var cfg model.ServerConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid configuration body")
		return
	}

	if err := s.store.UpdateConfig(cfg); err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, cfg)
}

func (s *Server) GetRawConfig(w http.ResponseWriter, r *http.Request) {
	raw, err := s.store.GetRawConfig()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(raw))
}

// Backup Handlers
func (s *Server) CreateBackup(w http.ResponseWriter, r *http.Request) {
	bk, err := s.store.CreateBackup()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, bk)
}

func (s *Server) RestoreBackup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.store.RestoreBackup(id); err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "restoring"})
}

func (s *Server) DownloadBackup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	filePath, downloadName, err := s.store.GetBackupPath(id)
	if err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}

	f, err := os.Open(filePath)
	if err != nil {
		WriteError(w, http.StatusNotFound, "Backup archive file not found on disk")
		return
	}
	defer f.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, downloadName))
	w.Header().Set("Content-Type", "application/gzip")
	_, _ = io.Copy(w, f)
}

func (s *Server) DeleteBackup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.store.DeleteBackup(id); err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// Console & Logs Handlers
func (s *Server) ExecuteConsole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Command string `json:"command"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Command == "" {
		WriteError(w, http.StatusBadRequest, "Command required")
		return
	}

	s.executor.Execute(req.Command)
	WriteJSON(w, http.StatusOK, map[string]string{"status": "executed"})
}

func (s *Server) ClearLogs(w http.ResponseWriter, r *http.Request) {
	s.store.ClearLogs()
	WriteJSON(w, http.StatusOK, map[string]string{"status": "cleared"})
}

// Server-Sent Events (SSE) Stream
func (s *Server) StreamEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		WriteError(w, http.StatusInternalServerError, "Streaming unsupported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Send initial state immediately
	initialBytes, _ := json.Marshal(s.store.GetState())
	fmt.Fprintf(w, "data: %s\n\n", initialBytes)
	flusher.Flush()

	subChan := s.store.Subscribe()
	defer s.store.Unsubscribe(subChan)

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case state, ok := <-subChan:
			if !ok {
				return
			}
			bytes, err := json.Marshal(state)
			if err == nil {
				fmt.Fprintf(w, "data: %s\n\n", bytes)
				flusher.Flush()
			}
		}
	}
}

// RegisterRoutes registers all REST and SSE endpoints on mux.
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	// CORS and OPTIONS helper
	mux.HandleFunc("GET /api/state", s.GetState)
	mux.HandleFunc("GET /api/events/stream", s.StreamEvents)

	// Server lifecycle
	mux.HandleFunc("POST /api/server/start", s.StartServer)
	mux.HandleFunc("POST /api/server/stop", s.StopServer)
	mux.HandleFunc("POST /api/server/restart", s.RestartServer)

	// Profiles
	mux.HandleFunc("POST /api/profiles", s.CreateProfile)
	mux.HandleFunc("POST /api/profiles/activate", s.ActivateProfile)

	// Events
	mux.HandleFunc("POST /api/events/launch", s.LaunchEvent)
	mux.HandleFunc("POST /api/events/end", s.EndEvent)

	// Schedules
	mux.HandleFunc("POST /api/schedules", s.CreateSchedule)
	mux.HandleFunc("PATCH /api/schedules/{id}/toggle", s.ToggleSchedule)
	mux.HandleFunc("DELETE /api/schedules/{id}", s.DeleteSchedule)

	// Players
	mux.HandleFunc("POST /api/players/kick", s.KickPlayer)
	mux.HandleFunc("POST /api/players/ban", s.BanPlayer)
	mux.HandleFunc("POST /api/players/unban", s.UnbanPlayer)
	mux.HandleFunc("POST /api/players/op", s.TogglePlayerOP)
	mux.HandleFunc("POST /api/players/join", s.SimulateJoin)

	// Whitelist
	mux.HandleFunc("POST /api/whitelist", s.AddWhitelist)
	mux.HandleFunc("DELETE /api/whitelist/{name}", s.RemoveWhitelist)

	// Config
	mux.HandleFunc("GET /api/config", s.GetConfig)
	mux.HandleFunc("POST /api/config", s.UpdateConfig)
	mux.HandleFunc("GET /api/config/raw", s.GetRawConfig)

	// Backups
	mux.HandleFunc("POST /api/backups", s.CreateBackup)
	mux.HandleFunc("POST /api/backups/{id}/restore", s.RestoreBackup)
	mux.HandleFunc("GET /api/backups/{id}/download", s.DownloadBackup)
	mux.HandleFunc("DELETE /api/backups/{id}", s.DeleteBackup)

	// Console & Logs
	mux.HandleFunc("POST /api/console", s.ExecuteConsole)
	mux.HandleFunc("DELETE /api/logs", s.ClearLogs)

	// Static file serving from workspace
	fileServer := http.FileServer(http.Dir(s.webDir))
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		// If requesting /api that doesn't exist, return 404 json
		if strings.HasPrefix(r.URL.Path, "/api/") {
			WriteError(w, http.StatusNotFound, "API endpoint not found")
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}
