package rcon

import (
	"os"
	"strings"
	"testing"

	"blockhost/internal/store"
)

func TestConsoleCommands(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "blockhost-rcon-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	st, err := store.New(tempDir)
	if err != nil {
		t.Fatalf("failed to init store: %v", err)
	}
	defer st.Close()

	executor := NewExecutor(st)

	// Test help
	executor.Execute("help")
	logs := st.GetState().Logs
	if len(logs) == 0 {
		t.Errorf("expected logs after help command")
	}

	// Test say
	executor.Execute("say Hello World")
	latest := st.GetState().Logs[len(st.GetState().Logs)-1]
	if !strings.Contains(latest.Text, "Hello World") {
		t.Errorf("expected latest log to contain 'Hello World', got '%s'", latest.Text)
	}

	// Test profile switch
	executor.Execute("profile eco")
	if st.GetState().ActiveProfileID != "eco" {
		t.Errorf("expected active profile to be 'eco', got '%s'", st.GetState().ActiveProfileID)
	}

	// Test kick
	executor.Execute("kick DanKrejci")
	for _, p := range st.GetState().Players {
		if p.Name == "DanKrejci" && p.Online {
			t.Errorf("expected DanKrejci to be kicked offline")
		}
	}

	// Test clear
	executor.Execute("clear")
	if len(st.GetState().Logs) != 0 {
		t.Errorf("expected logs to be empty after clear")
	}
}
