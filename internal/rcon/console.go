package rcon

import (
	"fmt"
	"strings"

	"blockhost/internal/store"
)

// Executor executes console / RCON commands against the server store.
type Executor struct {
	store *store.Store
}

// NewExecutor creates a new command executor.
func NewExecutor(s *store.Store) *Executor {
	return &Executor{store: s}
}

// Execute processes a console command line string.
func (e *Executor) Execute(raw string) {
	cmdStr := strings.TrimSpace(raw)
	if cmdStr == "" {
		return
	}

	e.store.AddLog("INFO", fmt.Sprintf("> %s", cmdStr))

	parts := strings.Fields(cmdStr)
	cmd := strings.ToLower(strings.TrimPrefix(parts[0], "/"))
	args := parts[1:]

	state := e.store.GetState()
	if state.ServerStatus != "online" && cmd != "start" && cmd != "help" {
		e.store.AddLog("ERROR", "[RCON] Příkaz nelze vykonat: Server je offline.")
		return
	}

	switch cmd {
	case "help":
		e.store.AddLog("SYSTEM", "Dostupné příkazy:")
		e.store.AddLog("SYSTEM", "  help                 - Zobrazí nápovědu")
		e.store.AddLog("SYSTEM", "  list                 - Seznam připojených hráčů")
		e.store.AddLog("SYSTEM", "  tps                  - Zobrazí aktuální TPS")
		e.store.AddLog("SYSTEM", "  say <zpráva>         - Odešle hlášení do chatu všem hráčům")
		e.store.AddLog("SYSTEM", "  op <hráč>            - Přidělí hráči administrátorská práva")
		e.store.AddLog("SYSTEM", "  deop <hráč>          - Odebere hráči OP práva")
		e.store.AddLog("SYSTEM", "  kick <hráč>          - Vyhodí hráče ze serveru")
		e.store.AddLog("SYSTEM", "  ban <hráč>           - Zabanuje hráče")
		e.store.AddLog("SYSTEM", "  unban <hráč>         - Odbanuje hráče")
		e.store.AddLog("SYSTEM", "  profile <eco|standard|pro|ultra> - Okamžitě přepne výkonový profil")
		e.store.AddLog("SYSTEM", "  backup               - Spustí okamžité zálohování")
		e.store.AddLog("SYSTEM", "  start / stop / restart - Řízení stavu serveru")
		e.store.AddLog("SYSTEM", "  clear                - Vymaže okno konzole")

	case "list":
		var active []string
		for _, p := range state.Players {
			if p.Online {
				active = append(active, p.Name)
			}
		}
		playerList := "Žádní"
		if len(active) > 0 {
			playerList = strings.Join(active, ", ")
		}
		e.store.AddLog("INFO", fmt.Sprintf("Online hráči (%d/%d): %s", len(active), state.Config.MaxPlayers, playerList))

	case "tps":
		e.store.AddLog("INFO", fmt.Sprintf("TPS z posledních 1m, 5m, 15m: %.2f, 20.00, 20.00", state.Metrics.TPS))

	case "say":
		if len(args) == 0 {
			e.store.AddLog("ERROR", "Použití: say <zpráva>")
		} else {
			msg := strings.Join(args, " ")
			e.store.AddLog("CHAT", fmt.Sprintf("[Server] %s", msg))
		}

	case "op":
		if len(args) == 0 {
			e.store.AddLog("ERROR", "Použití: op <hráč>")
		} else {
			nick := args[0]
			ok, isOP := e.store.TogglePlayerOP(nick)
			if !ok {
				e.store.AddLog("WARN", fmt.Sprintf("Hráč %s nebyl nalezen v databázi.", nick))
			} else if !isOP {
				// We wanted to op him, but toggle turned it off -> turn on again
				e.store.TogglePlayerOP(nick)
			}
		}

	case "deop":
		if len(args) == 0 {
			e.store.AddLog("ERROR", "Použití: deop <hráč>")
		} else {
			nick := args[0]
			ok, isOP := e.store.TogglePlayerOP(nick)
			if !ok {
				e.store.AddLog("WARN", fmt.Sprintf("Hráč %s nebyl nalezen.", nick))
			} else if isOP {
				// We wanted to deop -> toggle again
				e.store.TogglePlayerOP(nick)
			}
		}

	case "kick":
		if len(args) == 0 {
			e.store.AddLog("ERROR", "Použití: kick <hráč>")
		} else {
			nick := args[0]
			if !e.store.KickPlayer(nick) {
				e.store.AddLog("WARN", fmt.Sprintf("Hráč %s není online.", nick))
			}
		}

	case "ban":
		if len(args) == 0 {
			e.store.AddLog("ERROR", "Použití: ban <hráč>")
		} else {
			e.store.BanPlayer(args[0], "Zabanováno přes konzoli")
		}

	case "unban":
		if len(args) == 0 {
			e.store.AddLog("ERROR", "Použití: unban <hráč>")
		} else {
			e.store.UnbanPlayer(args[0])
		}

	case "profile":
		if len(args) == 0 {
			e.store.AddLog("ERROR", "Použití: profile <eco|standard|pro|ultra>")
		} else {
			pID := strings.ToLower(args[0])
			if _, err := e.store.ActivateProfile(pID, false); err != nil {
				e.store.AddLog("ERROR", fmt.Sprintf("Chyba při změně profilu: %v", err))
			}
		}

	case "backup":
		_, _ = e.store.CreateBackup()

	case "clear":
		e.store.ClearLogs()

	case "start":
		e.store.StartServer()

	case "stop":
		e.store.StopServer()

	case "restart":
		e.store.RestartServer()

	default:
		e.store.AddLog("WARN", fmt.Sprintf(`Neznámý příkaz "%s". Pro seznam příkazů napište "help".`, cmd))
	}
}
