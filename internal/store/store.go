package store

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"blockhost/internal/model"
)

// Store manages thread-safe state persistence, backups, and background tickers.
type Store struct {
	mu          sync.RWMutex
	state       model.AppState
	dataDir     string
	backupsDir  string
	stateFile   string
	propsFile   string
	subscribers map[chan model.AppState]struct{}
	subMu       sync.Mutex
	stopChan    chan struct{}
}

// Czech day of week abbreviations
var dayMap = map[time.Weekday]string{
	time.Monday:    "Po",
	time.Tuesday:   "Út",
	time.Wednesday: "St",
	time.Thursday:  "Čt",
	time.Friday:    "Pá",
	time.Saturday:  "So",
	time.Sunday:    "Ne",
}

// New creates a new Store instance.
func New(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	backupsDir := filepath.Join(dataDir, "backups")
	if err := os.MkdirAll(backupsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backups directory: %w", err)
	}

	s := &Store{
		dataDir:     dataDir,
		backupsDir:  backupsDir,
		stateFile:   filepath.Join(dataDir, "state.json"),
		propsFile:   filepath.Join(dataDir, "server.properties"),
		subscribers: make(map[chan model.AppState]struct{}),
		stopChan:    make(chan struct{}),
	}

	// Initialize default state or load existing
	if err := s.loadOrInit(); err != nil {
		return nil, err
	}

	// Start background simulation ticker
	go s.runTicker()

	return s, nil
}

// Close gracefully terminates background routines.
func (s *Store) Close() {
	close(s.stopChan)
}

// Subscribe returns a channel that receives state updates.
func (s *Store) Subscribe() chan model.AppState {
	s.subMu.Lock()
	defer s.subMu.Unlock()
	ch := make(chan model.AppState, 10)
	s.subscribers[ch] = struct{}{}
	return ch
}

// Unsubscribe removes a state listener channel.
func (s *Store) Unsubscribe(ch chan model.AppState) {
	s.subMu.Lock()
	defer s.subMu.Unlock()
	delete(s.subscribers, ch)
	close(ch)
}

func (s *Store) broadcastLocked() {
	copyState := s.cloneState(s.state)
	s.subMu.Lock()
	defer s.subMu.Unlock()

	for ch := range s.subscribers {
		select {
		case ch <- copyState:
		default:
			// Non-blocking drop if channel is full
		}
	}
}

func (s *Store) broadcast() {
	s.mu.RLock()
	copyState := s.cloneState(s.state)
	s.mu.RUnlock()

	s.subMu.Lock()
	defer s.subMu.Unlock()
	for ch := range s.subscribers {
		select {
		case ch <- copyState:
		default:
		}
	}
}

// GetState returns a clone of the current state.
func (s *Store) GetState() model.AppState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cloneState(s.state)
}

func (s *Store) cloneState(src model.AppState) model.AppState {
	// Deep copy via json roundtrip or struct copy
	bytes, _ := json.Marshal(src)
	var dst model.AppState
	_ = json.Unmarshal(bytes, &dst)
	return dst
}

func (s *Store) loadOrInit() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.stateFile)
	if err == nil {
		if err := json.Unmarshal(data, &s.state); err == nil {
			_ = s.writePropertiesLocked()
			return nil
		}
	}

	// Set default initial state
	s.state = model.AppState{
		ServerStatus:    "online",
		UptimeSeconds:   14320,
		ActiveProfileID: "standard",
		Metrics: model.Metrics{
			CPU:     28,
			RAMUsed: 3.2,
			TPS:     20.0,
		},
		ActiveEvent: nil,
		Profiles: []model.Profile{
			{
				ID:          "eco",
				Name:        "Eco Starter",
				RAM:         2,
				VCPU:        1,
				Price:       19,
				Description: "Úsporný režim pro noční provoz a minimální zátěž",
			},
			{
				ID:          "standard",
				Name:        "Standard Community",
				RAM:         6,
				VCPU:        2,
				Price:       49,
				Description: "Optimální profil pro běžné denní hraní a základní pluginy",
			},
			{
				ID:          "pro",
				Name:        "Pro Guild",
				RAM:         10,
				VCPU:        4,
				Price:       89,
				Description: "Vysoký výkon pro větší komunitu a náročnější modifikace",
			},
			{
				ID:          "ultra",
				Name:        "Ultra Event",
				RAM:         16,
				VCPU:        6,
				Price:       149,
				Description: "Maximální výkon pro velké herní eventy, minihry a turnaje",
			},
		},
		Schedules: []model.Schedule{
			{
				ID:              "sch-1",
				Name:            "Noční útlum",
				Days:            []string{"Po", "Út", "St", "Čt", "Pá"},
				Time:            "01:00",
				TargetProfileID: "eco",
				Enabled:         true,
			},
			{
				ID:              "sch-2",
				Name:            "Odpolední špička",
				Days:            []string{"Po", "Út", "St", "Čt", "Pá"},
				Time:            "14:30",
				TargetProfileID: "standard",
				Enabled:         true,
			},
			{
				ID:              "sch-3",
				Name:            "Víkendový boost",
				Days:            []string{"So", "Ne"},
				Time:            "10:00",
				TargetProfileID: "pro",
				Enabled:         true,
			},
		},
		Players: []model.Player{
			{Name: "DanKrejci", IP: "192.168.1.15", Ping: 18, OP: true, Joined: "10:14", Online: true},
			{Name: "MinerBob", IP: "178.40.12.91", Ping: 32, OP: false, Joined: "11:05", Online: true},
			{Name: "PixelQueen", IP: "89.102.4.21", Ping: 24, OP: false, Joined: "11:22", Online: true},
			{Name: "SteveMaster", IP: "37.188.19.4", Ping: 45, OP: false, Joined: "11:45", Online: true},
		},
		Whitelist: []model.WhitelistEntry{
			{Name: "DanKrejci", AddedDate: "2026-10-01", AddedBy: "CONSOLE"},
			{Name: "MinerBob", AddedDate: "2026-10-01", AddedBy: "DanKrejci"},
			{Name: "PixelQueen", AddedDate: "2026-10-02", AddedBy: "DanKrejci"},
			{Name: "SteveMaster", AddedDate: "2026-10-02", AddedBy: "DanKrejci"},
			{Name: "AlexCraft", AddedDate: "2026-10-03", AddedBy: "DanKrejci"},
			{Name: "RedstoneWiz", AddedDate: "2026-10-04", AddedBy: "DanKrejci"},
		},
		Banlist: []model.BanEntry{
			{Name: "Griefer99", Reason: "Ničení cizích staveb (Griefing)", Date: "2026-10-03", Banner: "DanKrejci"},
			{Name: "XrayUser", Reason: "Použití nepovolených módů (X-Ray)", Date: "2026-10-04", Banner: "CONSOLE"},
		},
		Config: model.ServerConfig{
			MOTD:       "Vítejte na BlockHost Minecraft Serveru!",
			ServerName: "BlockHost SMP",
			Core:       "Paper 1.20.4",
			Gamemode:   "survival",
			Difficulty: "normal",
			MaxPlayers: 40,
			Port:       25565,
			PVP:        true,
			Whitelist:  true,
		},
		Backups: []model.Backup{
			{ID: "bk-1", Name: "Auto-Zaloha-Svet-Daily", Date: "2026-10-06 04:00", Size: "428 MB", Type: "Automatická"},
			{ID: "bk-2", Name: "Pred-Aktualizaci-Paper", Date: "2026-10-05 18:30", Size: "415 MB", Type: "Manuální"},
			{ID: "bk-3", Name: "Tydenni-Archiv-Full", Date: "2026-10-01 02:00", Size: "610 MB", Type: "Automatická"},
		},
		Logs: []model.LogEntry{
			{Time: "11:58:12", Level: "INFO", Text: "[Server] Starting Minecraft server version 1.20.4"},
			{Time: "11:58:15", Level: "INFO", Text: "[Server] Loading properties and world chunks..."},
			{Time: "11:58:20", Level: "INFO", Text: "[BlockHost] Dynamic profile active: Standard (6 GB RAM, 2 vCPU)"},
			{Time: "11:58:22", Level: "INFO", Text: "[Server] Done (7.21s)! For help, type \"help\""},
			{Time: "12:01:05", Level: "CHAT", Text: "<DanKrejci> Vítejte na serveru BlockHost!"},
			{Time: "12:05:40", Level: "INFO", Text: "[BlockHost-RCON] RCON daemon running on port 25575"},
		},
	}

	_ = s.saveStateLocked()
	_ = s.writePropertiesLocked()
	return nil
}

func (s *Store) saveStateLocked() error {
	bytes, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.stateFile, bytes, 0644)
}

func (s *Store) writePropertiesLocked() error {
	cfg := s.state.Config
	content := fmt.Sprintf(`# Minecraft server properties - BlockHost Generated
# %s
motd=%s
server-name=%s
server-port=%d
gamemode=%s
difficulty=%s
max-players=%d
pvp=%t
white-list=%t
view-distance=10
simulation-distance=8
enable-rcon=true
rcon.port=25575
rcon.password=********
online-mode=true
allow-nether=true
enable-command-block=true
`, time.Now().Format(time.RFC3339), cfg.MOTD, cfg.ServerName, cfg.Port, cfg.Gamemode, cfg.Difficulty, cfg.MaxPlayers, cfg.PVP, cfg.Whitelist)

	return os.WriteFile(s.propsFile, []byte(content), 0644)
}

// AddLogLocked appends a log entry (caller must hold lock or use AddLog).
func (s *Store) AddLogLocked(level, text string) {
	timeStr := time.Now().Format("15:04:05")
	s.state.Logs = append(s.state.Logs, model.LogEntry{
		Time:  timeStr,
		Level: level,
		Text:  text,
	})
	if len(s.state.Logs) > 200 {
		s.state.Logs = s.state.Logs[len(s.state.Logs)-200:]
	}
}

// AddLog appends a log entry safely.
func (s *Store) AddLog(level, text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AddLogLocked(level, text)
	_ = s.saveStateLocked()
	s.broadcastLocked()
}

// ClearLogs clears log history.
func (s *Store) ClearLogs() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Logs = []model.LogEntry{}
	_ = s.saveStateLocked()
	s.broadcastLocked()
}

// SetServerStatus changes server status (start, stop, restart).
func (s *Store) SetServerStatus(status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.ServerStatus = status
	if status == "offline" {
		s.state.Metrics.CPU = 0
		s.state.Metrics.RAMUsed = 0
		s.state.Metrics.TPS = 0
		s.AddLogLocked("WARN", "[Server] Stopping server: saving players, saving worlds")
		s.AddLogLocked("INFO", "[Server] Thread shutdown complete.")
	} else if status == "online" {
		s.state.UptimeSeconds = 0
		s.AddLogLocked("INFO", "[Server] Minecraft server loaded chunks and opened network socket (25565)")
		s.AddLogLocked("INFO", fmt.Sprintf("[Server] BlockHost profile: %s", strings.ToUpper(s.state.ActiveProfileID)))
	}
	_ = s.saveStateLocked()
	s.broadcastLocked()
}

// StartServer sets up restarting -> online flow.
func (s *Store) StartServer() {
	s.mu.Lock()
	if s.state.ServerStatus == "online" {
		s.mu.Unlock()
		return
	}
	s.state.ServerStatus = "restarting"
	s.AddLogLocked("INFO", "[BlockHost] Starting server process...")
	_ = s.saveStateLocked()
	s.mu.Unlock()
	s.broadcastLocked()

	time.AfterFunc(2*time.Second, func() {
		s.SetServerStatus("online")
	})
}

// StopServer stops server.
func (s *Store) StopServer() {
	s.SetServerStatus("offline")
}

// RestartServer restarts server.
func (s *Store) RestartServer() {
	s.mu.Lock()
	if s.state.ServerStatus == "restarting" {
		s.mu.Unlock()
		return
	}
	s.state.ServerStatus = "restarting"
	s.AddLogLocked("WARN", "[Server] Restarting Minecraft server process...")
	_ = s.saveStateLocked()
	s.mu.Unlock()
	s.broadcastLocked()

	time.AfterFunc(2500*time.Millisecond, func() {
		s.SetServerStatus("online")
	})
}

// UpdateConfig updates server properties.
func (s *Store) UpdateConfig(cfg model.ServerConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.state.Config = cfg
	_ = s.writePropertiesLocked()
	s.AddLogLocked("INFO", fmt.Sprintf(`[Server] Configuration saved: MOTD="%s", Core=%s`, cfg.MOTD, cfg.Core))
	if err := s.saveStateLocked(); err != nil {
		return err
	}
	s.broadcastLocked()
	return nil
}

// GetRawConfig returns raw server.properties string.
func (s *Store) GetRawConfig() (string, error) {
	data, err := os.ReadFile(s.propsFile)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ActivateProfile changes current hardware profile.
func (s *Store) ActivateProfile(profileID string, automatic bool) (*model.Profile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var target *model.Profile
	for i := range s.state.Profiles {
		if s.state.Profiles[i].ID == profileID {
			target = &s.state.Profiles[i]
			break
		}
	}
	if target == nil {
		return nil, fmt.Errorf("profile '%s' not found", profileID)
	}

	s.state.ActiveProfileID = profileID
	s.AddLogLocked("INFO", fmt.Sprintf("[BlockHost] Profil změněn na: %s (%d GB RAM, %d vCPU)", target.Name, target.RAM, target.VCPU))
	_ = s.saveStateLocked()
	s.broadcastLocked()
	return target, nil
}

// AddProfile creates a new hardware profile.
func (s *Store) AddProfile(p model.Profile) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if p.ID == "" {
		p.ID = fmt.Sprintf("custom-%d", time.Now().UnixMilli())
	}
	s.state.Profiles = append(s.state.Profiles, p)
	_ = s.saveStateLocked()
	s.broadcastLocked()
	return nil
}

// LaunchEvent starts a temporary event boost.
func (s *Store) LaunchEvent(name, profileID string, durationMins int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	endsAt := time.Now().Add(time.Duration(durationMins) * time.Minute).UnixMilli()
	orig := s.state.ActiveProfileID

	s.state.ActiveEvent = &model.ActiveEvent{
		Name:              name,
		ProfileID:         profileID,
		EndsAt:            endsAt,
		DurationMins:      durationMins,
		OriginalProfileID: orig,
	}

	s.state.ActiveProfileID = profileID
	s.AddLogLocked("SYSTEM", fmt.Sprintf(`[BlockHost Event] Spuštěn event "%s" na dobu %d minut. Dočasný profil: %s`, name, durationMins, profileID))
	_ = s.saveStateLocked()
	s.broadcastLocked()
	return nil
}

// EndEvent terminates an active event early.
func (s *Store) EndEvent() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state.ActiveEvent == nil {
		return nil
	}

	orig := s.state.ActiveEvent.OriginalProfileID
	name := s.state.ActiveEvent.Name
	s.state.ActiveEvent = nil
	s.state.ActiveProfileID = orig

	s.AddLogLocked("SYSTEM", fmt.Sprintf(`[BlockHost Event] Event "%s" byl předčasně ukončen. Profil vrácen zpět.`, name))
	_ = s.saveStateLocked()
	s.broadcastLocked()
	return nil
}

// AddSchedule adds an automated scheduler rule.
func (s *Store) AddSchedule(sch model.Schedule) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sch.ID == "" {
		sch.ID = fmt.Sprintf("sch-%d", time.Now().UnixMilli())
	}
	s.state.Schedules = append(s.state.Schedules, sch)
	_ = s.saveStateLocked()
	s.broadcastLocked()
	return nil
}

// ToggleSchedule activates or deactivates a rule.
func (s *Store) ToggleSchedule(id string, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.state.Schedules {
		if s.state.Schedules[i].ID == id {
			s.state.Schedules[i].Enabled = enabled
			_ = s.saveStateLocked()
			s.broadcastLocked()
			return nil
		}
	}
	return fmt.Errorf("schedule rule '%s' not found", id)
}

// DeleteSchedule removes a schedule.
func (s *Store) DeleteSchedule(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filtered := s.state.Schedules[:0]
	for _, sch := range s.state.Schedules {
		if sch.ID != id {
			filtered = append(filtered, sch)
		}
	}
	s.state.Schedules = filtered
	_ = s.saveStateLocked()
	s.broadcastLocked()
	return nil
}

// Player operations
func (s *Store) KickPlayer(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	found := false
	for i := range s.state.Players {
		if strings.EqualFold(s.state.Players[i].Name, name) {
			s.state.Players[i].Online = false
			found = true
			break
		}
	}
	if found {
		s.AddLogLocked("INFO", fmt.Sprintf("[Server] Kicked player %s from the server (by admin)", name))
		_ = s.saveStateLocked()
		s.broadcastLocked()
	}
	return found
}

func (s *Store) BanPlayer(name, reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Kick first
	for i := range s.state.Players {
		if strings.EqualFold(s.state.Players[i].Name, name) {
			s.state.Players[i].Online = false
			break
		}
	}

	// Add to banlist if not present
	already := false
	for _, b := range s.state.Banlist {
		if strings.EqualFold(b.Name, name) {
			already = true
			break
		}
	}
	if !already {
		if reason == "" {
			reason = "Zabanováno administrátorem"
		}
		s.state.Banlist = append(s.state.Banlist, model.BanEntry{
			Name:   name,
			Reason: reason,
			Date:   time.Now().Format("2006-01-02"),
			Banner: "Admin",
		})
		s.AddLogLocked("WARN", fmt.Sprintf("[Server] Banned player %s: %s", name, reason))
		_ = s.saveStateLocked()
		s.broadcastLocked()
	}
}

func (s *Store) UnbanPlayer(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	filtered := s.state.Banlist[:0]
	for _, b := range s.state.Banlist {
		if !strings.EqualFold(b.Name, name) {
			filtered = append(filtered, b)
		}
	}
	s.state.Banlist = filtered
	s.AddLogLocked("INFO", fmt.Sprintf("[Server] Unbanned player %s", name))
	_ = s.saveStateLocked()
	s.broadcastLocked()
}

func (s *Store) TogglePlayerOP(name string) (bool, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.state.Players {
		if strings.EqualFold(s.state.Players[i].Name, name) {
			s.state.Players[i].OP = !s.state.Players[i].OP
			action := "Opped"
			if !s.state.Players[i].OP {
				action = "De-opped"
			}
			s.AddLogLocked("INFO", fmt.Sprintf("[Server] %s player %s", action, s.state.Players[i].Name))
			_ = s.saveStateLocked()
			s.broadcastLocked()
			return true, s.state.Players[i].OP
		}
	}
	return false, false
}

func (s *Store) AddPlayerJoin(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state.ServerStatus != "online" {
		return fmt.Errorf("server is not online")
	}

	found := false
	for i := range s.state.Players {
		if strings.EqualFold(s.state.Players[i].Name, name) {
			s.state.Players[i].Online = true
			found = true
			break
		}
	}
	if !found {
		s.state.Players = append(s.state.Players, model.Player{
			Name:   name,
			IP:     fmt.Sprintf("192.168.1.%d", 20+rand.Intn(70)),
			Ping:   15 + rand.Intn(35),
			OP:     false,
			Joined: time.Now().Format("15:04"),
			Online: true,
		})
	}
	s.AddLogLocked("INFO", fmt.Sprintf("[Server] %s[/192.168.1.55:54321] logged in with entity id %d", name, 100+rand.Intn(900)))
	_ = s.saveStateLocked()
	s.broadcastLocked()
	return nil
}

func (s *Store) AddWhitelist(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, w := range s.state.Whitelist {
		if strings.EqualFold(w.Name, name) {
			return
		}
	}
	s.state.Whitelist = append(s.state.Whitelist, model.WhitelistEntry{
		Name:      name,
		AddedDate: time.Now().Format("2006-01-02"),
		AddedBy:   "Admin",
	})
	s.AddLogLocked("INFO", fmt.Sprintf("[Server] Added %s to whitelist", name))
	_ = s.saveStateLocked()
	s.broadcastLocked()
}

func (s *Store) RemoveWhitelist(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	filtered := s.state.Whitelist[:0]
	for _, w := range s.state.Whitelist {
		if !strings.EqualFold(w.Name, name) {
			filtered = append(filtered, w)
		}
	}
	s.state.Whitelist = filtered
	s.AddLogLocked("INFO", fmt.Sprintf("[Server] Removed %s from whitelist", name))
	_ = s.saveStateLocked()
	s.broadcastLocked()
}

// Backup operations
func (s *Store) CreateBackup() (*model.Backup, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	bkID := fmt.Sprintf("bk-%d", now.UnixMilli())
	bkName := fmt.Sprintf("Backup-%s", now.Format("2006-01-02-15-04"))
	tarGzFile := filepath.Join(s.backupsDir, bkID+".tar.gz")

	// Create real tar.gz file
	f, err := os.Create(tarGzFile)
	if err != nil {
		return nil, err
	}
	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	// Add state snapshot inside archive
	stateData, _ := json.MarshalIndent(s.state, "", "  ")
	hdrState := &tar.Header{
		Name:    "state_snapshot.json",
		Mode:    0600,
		Size:    int64(len(stateData)),
		ModTime: now,
	}
	_ = tw.WriteHeader(hdrState)
	_, _ = tw.Write(stateData)

	// Add server.properties inside archive
	propsData, _ := os.ReadFile(s.propsFile)
	if len(propsData) > 0 {
		hdrProps := &tar.Header{
			Name:    "server.properties",
			Mode:    0600,
			Size:    int64(len(propsData)),
			ModTime: now,
		}
		_ = tw.WriteHeader(hdrProps)
		_, _ = tw.Write(propsData)
	}

	// Add dummy world seed chunk mock
	worldMock := []byte(fmt.Sprintf("BlockHost Level Data Chunk Seed: %d", now.UnixNano()))
	hdrWorld := &tar.Header{
		Name:    "world/level.dat",
		Mode:    0600,
		Size:    int64(len(worldMock)),
		ModTime: now,
	}
	_ = tw.WriteHeader(hdrWorld)
	_, _ = tw.Write(worldMock)

	_ = tw.Close()
	_ = gw.Close()
	_ = f.Close()

	fi, err := os.Stat(tarGzFile)
	sizeStr := "412 MB"
	if err == nil {
		sizeStr = fmt.Sprintf("%.1f KB", float64(fi.Size())/1024.0)
	}

	bk := model.Backup{
		ID:       bkID,
		Name:     bkName,
		Date:     now.Format("2006-01-02 15:04"),
		Size:     sizeStr,
		Type:     "Manuální",
		FilePath: tarGzFile,
	}

	s.state.Backups = append([]model.Backup{bk}, s.state.Backups...)
	s.AddLogLocked("INFO", fmt.Sprintf(`[BlockHost Backup] Záloha "%s" úspěšně vytvořena.`, bk.Name))
	_ = s.saveStateLocked()
	s.broadcastLocked()
	return &bk, nil
}

func (s *Store) RestoreBackup(id string) error {
	s.mu.Lock()

	var target *model.Backup
	for _, b := range s.state.Backups {
		if b.ID == id {
			target = &b
			break
		}
	}
	if target == nil {
		s.mu.Unlock()
		return fmt.Errorf("backup '%s' not found", id)
	}

	tarPath := target.FilePath
	if tarPath == "" {
		tarPath = filepath.Join(s.backupsDir, id+".tar.gz")
	}

	f, err := os.Open(tarPath)
	if err == nil {
		defer f.Close()
		gr, err := gzip.NewReader(f)
		if err == nil {
			defer gr.Close()
			tr := tar.NewReader(gr)
			for {
				hdr, err := tr.Next()
				if err != nil {
					break
				}
				if hdr.Name == "state_snapshot.json" {
					var restoredState model.AppState
					if err := json.NewDecoder(tr).Decode(&restoredState); err == nil {
						// Keep current backups list intact
						oldBackups := s.state.Backups
						s.state = restoredState
						s.state.Backups = oldBackups
					}
				}
			}
		}
	}

	s.AddLogLocked("WARN", fmt.Sprintf(`[BlockHost Backup] Obnova dat ze zálohy: %s`, target.Name))
	_ = s.saveStateLocked()
	_ = s.writePropertiesLocked()
	s.mu.Unlock()

	s.RestartServer()
	return nil
}

func (s *Store) DeleteBackup(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var pathToDelete string
	filtered := s.state.Backups[:0]
	for _, b := range s.state.Backups {
		if b.ID == id {
			pathToDelete = b.FilePath
			if pathToDelete == "" {
				pathToDelete = filepath.Join(s.backupsDir, id+".tar.gz")
			}
		} else {
			filtered = append(filtered, b)
		}
	}
	s.state.Backups = filtered
	if pathToDelete != "" {
		_ = os.Remove(pathToDelete)
	}
	_ = s.saveStateLocked()
	s.broadcastLocked()
	return nil
}

func (s *Store) GetBackupPath(id string) (string, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, b := range s.state.Backups {
		if b.ID == id {
			path := b.FilePath
			if path == "" {
				path = filepath.Join(s.backupsDir, id+".tar.gz")
			}
			return path, b.Name + ".tar.gz", nil
		}
	}
	return "", "", fmt.Errorf("backup '%s' not found", id)
}

// runTicker manages metric fluctuations, active events, and scheduling checks.
func (s *Store) runTicker() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case now := <-ticker.C:
			s.tick(now)
		}
	}
}

func (s *Store) tick(now time.Time) {
	s.mu.Lock()
	stateChanged := false

	// 1. Check Active Event Expiration
	if s.state.ActiveEvent != nil {
		if now.UnixMilli() >= s.state.ActiveEvent.EndsAt {
			finishedEvent := *s.state.ActiveEvent
			s.state.ActiveEvent = nil
			s.state.ActiveProfileID = finishedEvent.OriginalProfileID
			s.AddLogLocked("SYSTEM", fmt.Sprintf(`[BlockHost Event] Event "%s" skončil. Profil automaticky vrácen na výchozí.`, finishedEvent.Name))
			stateChanged = true
		}
	}

	// 2. Metrics fluctuation if server is online
	if s.state.ServerStatus == "online" {
		s.state.UptimeSeconds++

		var currentProfile model.Profile
		for _, p := range s.state.Profiles {
			if p.ID == s.state.ActiveProfileID {
				currentProfile = p
				break
			}
		}
		if currentProfile.RAM == 0 {
			currentProfile.RAM = 6
		}

		onlineCount := 0
		for _, p := range s.state.Players {
			if p.Online {
				onlineCount++
			}
		}

		// Realistic CPU calculation
		baseCPU := 15.0 + float64(onlineCount)*3.5
		jitter := (rand.Float64() * 8.0) - 4.0
		calcCPU := int(math.Round(baseCPU + jitter))
		if calcCPU < 5 {
			calcCPU = 5
		}
		if calcCPU > 99 {
			calcCPU = 99
		}
		s.state.Metrics.CPU = calcCPU

		// Realistic RAM calculation
		baseRAM := 1.8 + float64(onlineCount)*0.35
		ramJitter := (rand.Float64() * 0.2) - 0.1
		calcRAM := math.Max(1.0, math.Min(float64(currentProfile.RAM)*0.95, baseRAM+ramJitter))
		s.state.Metrics.RAMUsed = math.Round(calcRAM*10) / 10

		// TPS calculation
		if s.state.Metrics.CPU > 90 {
			s.state.Metrics.TPS = math.Max(16.5, 20.0-(rand.Float64()*2.0))
		} else {
			s.state.Metrics.TPS = math.Min(20.0, 19.85+(rand.Float64()*0.15))
		}

		stateChanged = true
	}

	// 3. Scheduler checking (every minute on 0th second)
	if now.Second() == 0 {
		dayAbbr := dayMap[now.Weekday()]
		currTime := now.Format("15:04")

		for _, rule := range s.state.Schedules {
			if !rule.Enabled {
				continue
			}
			if rule.Time == currTime {
				dayMatches := false
				for _, d := range rule.Days {
					if d == dayAbbr {
						dayMatches = true
						break
					}
				}
				if dayMatches && s.state.ActiveProfileID != rule.TargetProfileID {
					s.state.ActiveProfileID = rule.TargetProfileID
					s.AddLogLocked("SYSTEM", fmt.Sprintf(`[BlockHost Scheduler] Pravidlo "%s" aktivováno. Profil automaticky přepnut na %s.`, rule.Name, rule.TargetProfileID))
					stateChanged = true
				}
			}
		}
	}

	if stateChanged {
		s.broadcastLocked()
	}
	s.mu.Unlock()
}
