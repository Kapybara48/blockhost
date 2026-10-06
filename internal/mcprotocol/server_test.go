package mcprotocol

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"net"
	"os"
	"testing"
	"time"

	"blockhost/internal/store"
)

func TestMinecraftProtocolHandshakeAndPing(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "blockhost-mc-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	st, err := store.New(tempDir)
	if err != nil {
		t.Fatalf("failed to init store: %v", err)
	}
	defer st.Close()

	// Configure port 25599 for test
	cfg := st.GetState().Config
	cfg.Port = 25599
	cfg.MOTD = "Unit Test MOTD"
	_ = st.UpdateConfig(cfg)

	mcSrv := NewServer(st)
	if err := mcSrv.Start(); err != nil {
		t.Fatalf("failed to start mc server on port 25599: %v", err)
	}
	defer mcSrv.Stop()

	// Connect as client
	conn, err := net.DialTimeout("tcp", "127.0.0.1:25599", 2*time.Second)
	if err != nil {
		t.Fatalf("failed to connect to mc server: %v", err)
	}
	defer conn.Close()

	// 1. Send Handshake packet (Packet ID 0x00, protocol 765, address "localhost", port 25599, nextState 1)
	var hsBody bytes.Buffer
	_ = writeVarInt(&hsBody, 0x00) // Packet ID
	_ = writeVarInt(&hsBody, 765)  // Protocol version 1.20.4
	_ = writeString(&hsBody, "localhost")
	_ = binary.Write(&hsBody, binary.BigEndian, uint16(25599))
	_ = writeVarInt(&hsBody, 1) // Next state 1 = status

	// Write frame
	var hsFrame bytes.Buffer
	_ = writeVarInt(&hsFrame, hsBody.Len())
	hsFrame.Write(hsBody.Bytes())
	if _, err := conn.Write(hsFrame.Bytes()); err != nil {
		t.Fatalf("failed to write handshake: %v", err)
	}

	// 2. Send Status Request packet (Packet ID 0x00, 0 payload)
	var reqFrame bytes.Buffer
	_ = writeVarInt(&reqFrame, 1)    // len 1
	_ = writeVarInt(&reqFrame, 0x00) // ID 0
	if _, err := conn.Write(reqFrame.Bytes()); err != nil {
		t.Fatalf("failed to write status request: %v", err)
	}

	// 3. Read Status Response
	pktID, payload, err := readPacket(conn)
	if err != nil {
		t.Fatalf("failed to read status response: %v", err)
	}
	if pktID != 0x00 {
		t.Fatalf("expected packet ID 0x00, got 0x%02x", pktID)
	}

	pr := bytes.NewReader(payload)
	jsonStr, err := readString(pr)
	if err != nil {
		t.Fatalf("failed to read json string: %v", err)
	}

	var statusMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &statusMap); err != nil {
		t.Fatalf("failed to unmarshal status json: %v", err)
	}

	desc := statusMap["description"].(map[string]interface{})
	if desc["text"] != "Unit Test MOTD" {
		t.Errorf("expected MOTD 'Unit Test MOTD', got '%v'", desc["text"])
	}

	// 4. Send Ping packet (Packet ID 0x01, 8 bytes timestamp)
	var pingBody bytes.Buffer
	_ = writeVarInt(&pingBody, 0x01)
	testTimestamp := int64(1234567890)
	_ = binary.Write(&pingBody, binary.BigEndian, testTimestamp)

	var pingFrame bytes.Buffer
	_ = writeVarInt(&pingFrame, pingBody.Len())
	pingFrame.Write(pingBody.Bytes())
	if _, err := conn.Write(pingFrame.Bytes()); err != nil {
		t.Fatalf("failed to write ping: %v", err)
	}

	// 5. Read Pong
	pongID, pongPayload, err := readPacket(conn)
	if err != nil {
		t.Fatalf("failed to read pong: %v", err)
	}
	if pongID != 0x01 {
		t.Errorf("expected pong ID 0x01, got 0x%02x", pongID)
	}
	var receivedTimestamp int64
	_ = binary.Read(bytes.NewReader(pongPayload), binary.BigEndian, &receivedTimestamp)
	if receivedTimestamp != testTimestamp {
		t.Errorf("expected pong timestamp %d, got %d", testTimestamp, receivedTimestamp)
	}
}

func TestMinecraftProtocolLogin(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "blockhost-mc-login-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	st, err := store.New(tempDir)
	if err != nil {
		t.Fatalf("failed to init store: %v", err)
	}
	defer st.Close()

	cfg := st.GetState().Config
	cfg.Port = 25598
	cfg.Whitelist = false
	_ = st.UpdateConfig(cfg)

	mcSrv := NewServer(st)
	if err := mcSrv.Start(); err != nil {
		t.Fatalf("failed to start mc server: %v", err)
	}
	defer mcSrv.Stop()

	conn, err := net.DialTimeout("tcp", "127.0.0.1:25598", 2*time.Second)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	// 1. Handshake with nextState 2 (Login)
	var hsBody bytes.Buffer
	_ = writeVarInt(&hsBody, 0x00)
	_ = writeVarInt(&hsBody, 765)
	_ = writeString(&hsBody, "localhost")
	_ = binary.Write(&hsBody, binary.BigEndian, uint16(25598))
	_ = writeVarInt(&hsBody, 2) // Next state 2 = Login

	var hsFrame bytes.Buffer
	_ = writeVarInt(&hsFrame, hsBody.Len())
	hsFrame.Write(hsBody.Bytes())
	_, _ = conn.Write(hsFrame.Bytes())

	// 2. Login Start (0x00, username "RealPlayer")
	var loginBody bytes.Buffer
	_ = writeVarInt(&loginBody, 0x00)
	_ = writeString(&loginBody, "RealPlayer")

	var loginFrame bytes.Buffer
	_ = writeVarInt(&loginFrame, loginBody.Len())
	loginFrame.Write(loginBody.Bytes())
	_, _ = conn.Write(loginFrame.Bytes())

	// 3. Read Disconnect / Welcome packet
	pktID, payload, err := readPacket(conn)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	if pktID != 0x00 {
		t.Errorf("expected packet 0x00, got 0x%02x", pktID)
	}

	var chatMap map[string]string
	r := bytes.NewReader(payload)
	str, _ := readString(r)
	_ = json.Unmarshal([]byte(str), &chatMap)

	if chatMap["text"] == "" {
		t.Errorf("expected response text component, got empty")
	}

	// Verify player was recorded online in store
	found := false
	for _, p := range st.GetState().Players {
		if p.Name == "RealPlayer" && p.Online {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected RealPlayer to be registered online in state")
	}
}
