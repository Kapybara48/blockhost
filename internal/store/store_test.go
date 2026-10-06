package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStoreLifecycleAndPersistence(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "blockhost-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	s, err := New(tempDir)
	if err != nil {
		t.Fatalf("failed to init store: %v", err)
	}
	defer s.Close()

	initialState := s.GetState()
	if initialState.ServerStatus != "online" {
		t.Errorf("expected initial status 'online', got '%s'", initialState.ServerStatus)
	}

	// Test StopServer
	s.StopServer()
	stateAfterStop := s.GetState()
	if stateAfterStop.ServerStatus != "offline" {
		t.Errorf("expected status 'offline', got '%s'", stateAfterStop.ServerStatus)
	}
	if stateAfterStop.Metrics.CPU != 0 || stateAfterStop.Metrics.RAMUsed != 0 {
		t.Errorf("expected metrics to be 0 when offline, got CPU: %d, RAM: %.1f", stateAfterStop.Metrics.CPU, stateAfterStop.Metrics.RAMUsed)
	}

	// Test StartServer
	s.StartServer()
	// Initial transition state is restarting
	if s.GetState().ServerStatus != "restarting" {
		t.Errorf("expected status 'restarting', got '%s'", s.GetState().ServerStatus)
	}

	// Wait for start timer
	time.Sleep(2200 * time.Millisecond)
	if s.GetState().ServerStatus != "online" {
		t.Errorf("expected status 'online', got '%s'", s.GetState().ServerStatus)
	}

	// Test Profile Switching
	p, err := s.ActivateProfile("pro", false)
	if err != nil {
		t.Fatalf("failed to activate profile: %v", err)
	}
	if p.ID != "pro" || s.GetState().ActiveProfileID != "pro" {
		t.Errorf("expected active profile 'pro', got '%s'", s.GetState().ActiveProfileID)
	}

	// Test Event Booster
	err = s.LaunchEvent("Test Cup", "ultra", 1)
	if err != nil {
		t.Fatalf("failed to launch event: %v", err)
	}
	if s.GetState().ActiveEvent == nil || s.GetState().ActiveEvent.Name != "Test Cup" {
		t.Errorf("active event not set properly")
	}
	if s.GetState().ActiveProfileID != "ultra" {
		t.Errorf("expected event profile 'ultra', got '%s'", s.GetState().ActiveProfileID)
	}

	err = s.EndEvent()
	if err != nil {
		t.Fatalf("failed to end event: %v", err)
	}
	if s.GetState().ActiveEvent != nil {
		t.Errorf("expected active event to be nil")
	}
	if s.GetState().ActiveProfileID != "pro" {
		t.Errorf("expected profile to revert to 'pro', got '%s'", s.GetState().ActiveProfileID)
	}

	// Test Players
	s.KickPlayer("MinerBob")
	for _, pl := range s.GetState().Players {
		if pl.Name == "MinerBob" && pl.Online {
			t.Errorf("expected MinerBob to be offline")
		}
	}

	s.BanPlayer("MinerBob", "Test Rule Break")
	banFound := false
	for _, b := range s.GetState().Banlist {
		if b.Name == "MinerBob" {
			banFound = true
			break
		}
	}
	if !banFound {
		t.Errorf("expected MinerBob in banlist")
	}

	s.UnbanPlayer("MinerBob")
	for _, b := range s.GetState().Banlist {
		if b.Name == "MinerBob" {
			t.Errorf("expected MinerBob to be removed from banlist")
		}
	}

	// Test Backup Creation and Restore
	bk, err := s.CreateBackup()
	if err != nil {
		t.Fatalf("failed to create backup: %v", err)
	}
	if _, err := os.Stat(bk.FilePath); err != nil {
		t.Errorf("backup file does not exist on disk: %v", err)
	}

	// Update config then restore
	cfg := s.GetState().Config
	cfg.ServerName = "Modified Name"
	_ = s.UpdateConfig(cfg)
	if s.GetState().Config.ServerName != "Modified Name" {
		t.Errorf("failed to update config")
	}

	err = s.RestoreBackup(bk.ID)
	if err != nil {
		t.Fatalf("failed to restore backup: %v", err)
	}
	if s.GetState().Config.ServerName != "BlockHost SMP" {
		t.Errorf("expected restored server name 'BlockHost SMP', got '%s'", s.GetState().Config.ServerName)
	}

	// Delete backup
	err = s.DeleteBackup(bk.ID)
	if err != nil {
		t.Fatalf("failed to delete backup: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tempDir, "backups", bk.ID+".tar.gz")); !os.IsNotExist(err) {
		t.Errorf("backup archive was not removed from disk")
	}
}
