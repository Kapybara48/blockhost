package mcprotocol

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"blockhost/internal/model"
	"blockhost/internal/store"
)

// Server is a native Go Minecraft protocol server that listens on port 25565.
type Server struct {
	store    *store.Store
	mu       sync.Mutex
	listener net.Listener
	stopChan chan struct{}
	running  bool
}

// NewServer creates a new Minecraft protocol server.
func NewServer(s *store.Store) *Server {
	return &Server{
		store: s,
	}
}

// Start listens on the configured Minecraft port (default 25565).
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	state := s.store.GetState()
	port := state.Config.Port
	if port <= 0 {
		port = 25565
	}

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to bind Minecraft TCP port %d: %w", port, err)
	}

	s.listener = l
	s.stopChan = make(chan struct{})
	s.running = true

	log.Printf("🎮 Minecraft Protocol Server listening on %s (TCP)", addr)

	go s.acceptLoop(l, s.stopChan)
	return nil
}

// Stop closes the Minecraft TCP listener.
func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.running = false
	if s.stopChan != nil {
		close(s.stopChan)
	}
	if s.listener != nil {
		_ = s.listener.Close()
	}
	log.Printf("Minecraft Protocol Server stopped on TCP port.")
}

func (s *Server) acceptLoop(l net.Listener, stopChan chan struct{}) {
	for {
		conn, err := l.Accept()
		if err != nil {
			select {
			case <-stopChan:
				return
			default:
				return
			}
		}

		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))

	// 1. Read Handshake Packet
	pktID, payload, err := readPacket(conn)
	if err != nil || pktID != 0x00 {
		return
	}

	r := bytes.NewReader(payload)
	protoVersion, err := readVarInt(r)
	if err != nil {
		return
	}
	_, err = readString(r) // serverAddress
	if err != nil {
		return
	}
	var serverPort uint16
	if err := binary.Read(r, binary.BigEndian, &serverPort); err != nil {
		return
	}
	nextState, err := readVarInt(r)
	if err != nil {
		return
	}

	state := s.store.GetState()

	// State 1: Server List Ping (Status)
	if nextState == 1 {
		// Read Status Request (0x00)
		pktID, _, err := readPacket(conn)
		if err != nil || pktID != 0x00 {
			return
		}

		// Count online players
		onlineCount := 0
		var samplePlayers []map[string]string
		for _, p := range state.Players {
			if p.Online {
				onlineCount++
				samplePlayers = append(samplePlayers, map[string]string{
					"name": p.Name,
					"id":   "00000000-0000-0000-0000-000000000000",
				})
			}
		}

		// Respond with Server List Ping JSON
		statusJSON := map[string]interface{}{
			"version": map[string]interface{}{
				"name":     state.Config.Core,
				"protocol": protoVersion, // Echo back client's protocol version so ping shows green bars
			},
			"players": map[string]interface{}{
				"max":    state.Config.MaxPlayers,
				"online": onlineCount,
				"sample": samplePlayers,
			},
			"description": map[string]interface{}{
				"text": state.Config.MOTD,
			},
			"enforcesSecureChat": false,
		}

		jsonBytes, err := json.Marshal(statusJSON)
		if err != nil {
			return
		}

		var respBuf bytes.Buffer
		_ = writeString(&respBuf, string(jsonBytes))
		_ = sendPacket(conn, 0x00, respBuf.Bytes())

		// Read Ping Packet (0x01)
		pktID, pingPayload, err := readPacket(conn)
		if err != nil || pktID != 0x01 {
			return
		}

		// Send Pong Packet (0x01)
		_ = sendPacket(conn, 0x01, pingPayload)
		return
	}

	// State 2: Login Start
	if nextState == 2 {
		pktID, loginPayload, err := readPacket(conn)
		if err != nil || pktID != 0x00 {
			return
		}

		lr := bytes.NewReader(loginPayload)
		username, err := readString(lr)
		if err != nil || username == "" {
			username = "Player"
		}

		// Check Whitelist
		if state.Config.Whitelist {
			whitelisted := false
			for _, w := range state.Whitelist {
				if strings.EqualFold(w.Name, username) {
					whitelisted = true
					break
				}
			}
			if !whitelisted {
				s.store.AddLog("WARN", fmt.Sprintf("[Server] Připojení odmítnuto: Hráč %s není na whitelistu", username))
				sendDisconnect(conn, "§c[BlockHost] Nejste na whitelistu tohoto serveru!\n§7Požádejte administrátora o přístup.")
				return
			}
		}

		// Check Banlist
		for _, b := range state.Banlist {
			if strings.EqualFold(b.Name, username) {
				s.store.AddLog("WARN", fmt.Sprintf("[Server] Připojení odmítnuto: Hráč %s je zabanován (%s)", username, b.Reason))
				sendDisconnect(conn, fmt.Sprintf("§c[BlockHost] Byl jste zabanován ze serveru!\n§eDůvod: §f%s\n§7Admin: %s (%s)", b.Reason, b.Banner, b.Date))
				return
			}
		}

		// Register online player in Store
		_ = s.store.AddPlayerJoin(username)
		s.store.AddLog("INFO", fmt.Sprintf("[Server] %s[%s] se úspěšně připojil přes Minecraft klienta (Protokol %d)", username, conn.RemoteAddr().String(), protoVersion))

		var currentProfile model.Profile
		for _, p := range state.Profiles {
			if p.ID == state.ActiveProfileID {
				currentProfile = p
				break
			}
		}
		if currentProfile.RAM == 0 {
			currentProfile.Name = "Standard Community"
			currentProfile.RAM = 6
			currentProfile.VCPU = 2
		}

		welcomeMsg := fmt.Sprintf(
			"§6§l==============================================§r\n"+
				"§a§l         BlockHost Minecraft Server           §r\n"+
				"§6§l==============================================§r\n\n"+
				"§fVítej, §e§l%s§r§f!\n"+
				"§7Spojení s herním serverem bylo §a§lÚSPĚŠNĚ NAVÁZÁNO§r§7.\n\n"+
				"§fServer: §b%s (%s)§r\n"+
				"§fAktivní profil: §e%s (%d GB RAM, %d vCPU)§r\n"+
				"§fStav serveru: §a● Online§r  |  §fTPS: §a20.0§r\n"+
				"§fMOTD: §f%s§r\n\n"+
				"§7[Server běží pod správou BlockHost Go Backend]§r\n"+
				"§aVaše herní spojení bylo zaznamenáno na webovém dashboardu!§r\n"+
				"§6§l==============================================§r",
			username,
			state.Config.ServerName, state.Config.Core,
			currentProfile.Name, currentProfile.RAM, currentProfile.VCPU,
			state.Config.MOTD,
		)

		sendDisconnect(conn, welcomeMsg)
	}
}

func sendDisconnect(conn net.Conn, reason string) {
	chatJSON, _ := json.Marshal(map[string]string{"text": reason})
	var body bytes.Buffer
	_ = writeString(&body, string(chatJSON))
	_ = sendPacket(conn, 0x00, body.Bytes())
}

func readVarInt(r io.Reader) (int, error) {
	var val int
	var pos int
	buf := make([]byte, 1)
	for {
		_, err := r.Read(buf)
		if err != nil {
			return 0, err
		}
		b := buf[0]
		val |= int(b&0x7F) << pos
		if (b & 0x80) == 0 {
			break
		}
		pos += 7
		if pos >= 35 {
			return 0, fmt.Errorf("varint exceeds 35 bits")
		}
	}
	return val, nil
}

func writeVarInt(w io.Writer, val int) error {
	u := uint32(val)
	for {
		b := byte(u & 0x7F)
		u >>= 7
		if u != 0 {
			b |= 0x80
		}
		if _, err := w.Write([]byte{b}); err != nil {
			return err
		}
		if u == 0 {
			break
		}
	}
	return nil
}

func readString(r io.Reader) (string, error) {
	length, err := readVarInt(r)
	if err != nil {
		return "", err
	}
	if length < 0 || length > 32767 {
		return "", fmt.Errorf("invalid string length %d", length)
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

func writeString(w io.Writer, s string) error {
	if err := writeVarInt(w, len(s)); err != nil {
		return err
	}
	_, err := w.Write([]byte(s))
	return err
}

func readPacket(r io.Reader) (int, []byte, error) {
	pktLen, err := readVarInt(r)
	if err != nil {
		return 0, nil, err
	}
	if pktLen <= 0 || pktLen > 2097152 {
		return 0, nil, fmt.Errorf("invalid packet length %d", pktLen)
	}
	raw := make([]byte, pktLen)
	if _, err := io.ReadFull(r, raw); err != nil {
		return 0, nil, err
	}

	buf := bytes.NewReader(raw)
	pktID, err := readVarInt(buf)
	if err != nil {
		return 0, nil, err
	}
	payload := raw[len(raw)-buf.Len():]
	return pktID, payload, nil
}

func sendPacket(conn net.Conn, packetID int, payload []byte) error {
	var body bytes.Buffer
	if err := writeVarInt(&body, packetID); err != nil {
		return err
	}
	body.Write(payload)

	var frame bytes.Buffer
	if err := writeVarInt(&frame, body.Len()); err != nil {
		return err
	}
	frame.Write(body.Bytes())

	_, err := conn.Write(frame.Bytes())
	return err
}
