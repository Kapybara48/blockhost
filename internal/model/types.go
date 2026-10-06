package model

// Metrics holds system performance telemetry.
type Metrics struct {
	CPU     int     `json:"cpu"`
	RAMUsed float64 `json:"ramUsed"`
	TPS     float64 `json:"tps"`
}

// Profile represents a hardware/resource allocation tier.
type Profile struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	RAM         int    `json:"ram"`
	VCPU        int    `json:"vcpu"`
	Price       int    `json:"price"`
	Description string `json:"description"`
}

// Schedule represents an automated profile switching rule.
type Schedule struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Days            []string `json:"days"`
	Time            string   `json:"time"` // HH:MM format
	TargetProfileID string   `json:"targetProfileId"`
	Enabled         bool     `json:"enabled"`
}

// Player represents a player state in the server.
type Player struct {
	Name   string `json:"name"`
	IP     string `json:"ip"`
	Ping   int    `json:"ping"`
	OP     bool   `json:"op"`
	Joined string `json:"joined"`
	Online bool   `json:"online"`
}

// WhitelistEntry represents an allowed player.
type WhitelistEntry struct {
	Name      string `json:"name"`
	AddedDate string `json:"addedDate"`
	AddedBy   string `json:"addedBy"`
}

// BanEntry represents a banned player.
type BanEntry struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
	Date   string `json:"date"`
	Banner string `json:"banner"`
}

// ServerConfig represents minecraft server settings.
type ServerConfig struct {
	MOTD       string `json:"motd"`
	ServerName string `json:"serverName"`
	Core       string `json:"core"`
	Gamemode   string `json:"gamemode"`
	Difficulty string `json:"difficulty"`
	MaxPlayers int    `json:"maxPlayers"`
	Port       int    `json:"port"`
	PVP        bool   `json:"pvp"`
	Whitelist  bool   `json:"whitelist"`
}

// Backup represents a saved server snapshot archive.
type Backup struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Date     string `json:"date"`
	Size     string `json:"size"`
	Type     string `json:"type"`
	FilePath string `json:"filePath,omitempty"`
}

// LogEntry represents a console log message.
type LogEntry struct {
	Time  string `json:"time"`
	Level string `json:"level"` // INFO, WARN, ERROR, CHAT, SYSTEM
	Text  string `json:"text"`
}

// ActiveEvent represents a temporary high-performance event boost.
type ActiveEvent struct {
	Name              string `json:"name"`
	ProfileID         string `json:"profileId"`
	EndsAt            int64  `json:"endsAt"` // Unix timestamp in milliseconds
	DurationMins      int    `json:"durationMins"`
	OriginalProfileID string `json:"originalProfileId"`
}

// AppState represents the complete application state.
type AppState struct {
	ServerStatus    string           `json:"serverStatus"` // 'online' | 'offline' | 'restarting'
	UptimeSeconds   int64            `json:"uptimeSeconds"`
	ActiveProfileID string           `json:"activeProfileId"`
	Metrics         Metrics          `json:"metrics"`
	ActiveEvent     *ActiveEvent     `json:"activeEvent"`
	Profiles        []Profile        `json:"profiles"`
	Schedules       []Schedule       `json:"schedules"`
	Players         []Player         `json:"players"`
	Whitelist       []WhitelistEntry `json:"whitelist"`
	Banlist         []BanEntry       `json:"banlist"`
	Config          ServerConfig     `json:"config"`
	Backups         []Backup         `json:"backups"`
	Logs            []LogEntry       `json:"logs"`
}
