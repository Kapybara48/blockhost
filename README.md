# BlockHost 🎮

> [!CAUTION]
> ### ⚠️ School Project Disclaimer
> **This repository is strictly a school project created for academic and educational purposes.**
>
> - **DO NOT USE** this project in production or real-world environments.
> - **DO NOT DEPLOY** or rely on this code for managing active servers or sensitive data.
> - **1000% vibe coded** ✨ — enter at your own risk.
> - It is provided as-is without any warranties, maintenance, guarantees, or security auditing.

---

## 📌 About the Project

**BlockHost** is a web-based dashboard prototype designed for centralized management, monitoring, and automation of Minecraft servers. It was developed by Daniel Krejčí as a student assignment.

### Planned / Demonstrated Features

- **Server Instance Control:** Start, stop, restart, and kill server instances.
- **Interactive Console:** Live console logs and command input interface.
- **Resource Monitoring:** CPU, RAM, disk usage, online players, and TPS metrics.
- **Configuration Management:** Web-based editing of server settings and `server.properties`.
- **Player Management:** Active player inspection, whitelist/banlist administration, and operator rights.
- **Backup & Restore System:** Manual and scheduled backup management.
- **Dynamic Performance Profiles:** Cost and resource tiers (e.g., Eco, Standard, Turbo) with scheduled scaling.

---

## 🛠️ Tech Stack & Methodology

- **Frontend:** HTML5, CSS3, Vanilla JavaScript (Single-Page App)
- **Backend:** Golang (Standard Library: `net/http`, `archive/tar`, `compress/gzip`, `encoding/json`, `sync`, `time`)
- **Real-Time Communication:** Server-Sent Events (SSE)
- **Methodology:** 1000% vibe coded ✨
- **Purpose:** Educational prototype / functional web application

---

## 🚀 Running the Go Backend

Make sure [Go](https://go.dev/) is installed (Go 1.22+):

```bash
# Run backend directly
go run .

# Or build the binary
go build -o blockhost-server .
./blockhost-server -port 8080
```

Once running, navigate to:
- 🌐 **Web Dashboard:** [http://localhost:8080/](http://localhost:8080/)
- 📡 **REST API:** [http://localhost:8080/api/state](http://localhost:8080/api/state)
- ⚡ **Live SSE Stream:** [http://localhost:8080/api/events/stream](http://localhost:8080/api/events/stream)

### Available Command-Line Flags

| Flag | Default | Description |
|---|---|---|
| `-port` | `8080` | Port for the HTTP server |
| `-data` | `./data` | Directory for JSON state, `server.properties`, and `.tar.gz` backups |
| `-web` | `.` | Directory containing the frontend files (`index.html`, `style.css`, `app.js`) |

---

## 👤 Author

- **Daniel Krejčí** — School Project

---

## 📄 License & Terms

This software is provided strictly for educational review and grading. No support, updates, or suitability for any particular purpose is guaranteed.
