# Emby-In-One-Next

<div align="center">

### The Next-Generation Emby Aggregation & Playback Gateway
**Unify Multiple Emby Servers into One Seamless, High-Performance Media Gateway**

[![Release](https://img.shields.io/github/v/release/Takumilove/Emby-In-One-Next?color=brightgreen&label=Release&logo=github)](https://github.com/Takumilove/Emby-In-One-Next/releases)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker&logoColor=white)](https://www.docker.com/)
[![License: GPL v3](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20amd64%20%7C%20arm64%20%7C%20armv7-lightgrey)]()

[📖 Quick Start](#-quick-start) • [🌟 Why Next](#-why-emby-in-one-next) • [🤝 Lineage & Acknowledgements](#-lineage--acknowledgements) • [⚡ Features](#-features-overview) • [🛠️ Maintenance](#-cli-management-menu) • [中文文档](README.md)

</div>

---

## 🌟 Why Emby-In-One-Next?

**Emby-In-One-Next** is an advanced, high-performance aggregation and playback gateway designed specifically for media enthusiasts. It seamlessly merges your multiple Emby servers (self-hosted, shared, netdisk-mounted, etc.) into a unified portal accessible by any standard Emby or third-party client.

#### Overcoming the Critical Pain Points of Legacy Solutions:

| Challenge / Pain Point | Legacy / Upstream Behavior | 🚀 Emby-In-One-Next Solution |
| :--- | :--- | :--- |
| **Home "Resume" Cards** | Batch querying `/Items` triggers `403 Forbidden` on non-admin upstream accounts, leaving Resume cards blank | **Adaptive User-Scope Passthrough**: Queries `/Users/{userId}/Items` with fallback to `/Items`; concurrency fallback fetches items one-by-one, ensuring 100% card recovery |
| **Missing Progress Bars** | Upstream missing `RunTimeTicks` causes clients to drop Resume cards entirely | **Dynamic Metric Enrichment**: Automatically injects duration, poster tags, and calculates `PlayedPercentage` based on local records for perfect progress rendering |
| **Client Incompatibilities** | AfuseKt, Yamby, Infuse, and SenPlayer crash or drop playback due to lowercase routes or missing endpoints | **Full Matrix Routing**: Comprehensive bidirectional routing for `/Users/{userId}/...` and general endpoints; case-insensitive route normalization; supports JSON, Form, and Query auth |
| **TV Shows "NextUp"** | Inability to aggregate unplayed next episodes across servers, breaking binge-watching continuity | **Cross-Server NextUp Aggregation**: Intelligently aggregates unplayed next episodes across upstream servers, filtering finished and invalid items |
| **Media Sources & Concurrency** | Third-party players calling PlaybackInfo bypass multi-version merging and concurrency limits | **PlaybackInfo User Endpoints**: Fully wired to trigger multi-server version merging (MediaSources) and enforce per-server `maxConcurrent` throttling |
| **Cloud/Netdisk Streaming Costs** | Proxy mode consumes massive VPS bandwidth and breaks on 302 image/stream redirects | **Netdisk 302 Direct Stream + Image Follow**: Redirects players to upstream CDN signed URLs directly; automatically resolves image redirects to eliminate broken posters |
| **Multi-User Data Desync** | Sharing accounts with family/friends overwrites shared watch history and progress | **True Multi-User Isolation**: Local SQLite engine records watch history, favorites, and played state independently for each sub-user without touching upstream |
| **Upgrades & Migration** | Updating can cause configuration overwrite or data directory desync | **Seamless In-Place Migration**: Automatically detects and migrates legacy `/opt/emby-in-one` data, preserving command aliases and configs |

---

## 🤝 Lineage & Acknowledgements

**Emby-In-One-Next** honors the open-source spirit with profound gratitude:

- This project is a next-generation evolutionary continuation based on [ArizeSky/Emby-In-One](https://github.com/ArizeSky/Emby-In-One) by [@chen3861229](https://github.com/chen3861229). We thank the original author for the pioneering architecture and Go virtualization foundation.
- Special thanks to community contributor [@ssabv](https://github.com/ssabv) for pioneering netdisk 302 stream pass-through, paged library scanning, and image redirect handling.
- **The Next Mission**: As the original repository matured and became less active in addressing complex third-party client bugs, `Emby-In-One-Next` steps forward to carry the torch. We completely refactored cross-server playback, eliminated playback dropouts, expanded user-scoped routing, and established the standalone **V1.0.0** release timeline.
- Open-sourced under the **GNU General Public License v3.0 (GPL-3.0)**.

---

## ⚡ Features Overview

### 🎬 1. Cross-Server Aggregation & Seamless Playback
- **Unified "Resume" (Continue Watching)**: Consolidates progress from all servers right onto the home screen with accurate progress bars.
- **Cross-Server "NextUp"**: Automatically surfaces the next episode across all servers without manual switching.
- **Multi-Version Stream Merging**: Merges identical movies/shows into a single detail card (4K HDR, 1080P, Remux) for player source selection.
- **4-Tier Metadata Prioritization**: Flagged priority > Chinese language > Summary length > Server order to deliver the best scraped posters and info.

### 📱 2. Comprehensive Client Compatibility
- **Universal Ecosystem Support**: Official Emby clients (Web, Android TV, iOS, macOS, Windows), AfuseKt, Yamby, Infuse, VidHub, SenPlayer, etc.
- **Case-Insensitive Route Normalization**: Automatically normalizes Emby endpoints to withstand lowercase client requests (e.g. `/Users/authenticatebyname`).
- **Full Route Surface**: Complete coverage for `Similar`, `ThemeMedia`, `Search/Hints`, `Latest`, `Seasons`, and `Episodes`.

### 👥 3. True Multi-User Isolation
- **Independent Sub-Users**: Create distinct sub-users on the gateway without provisioning additional upstream Emby accounts.
- **Per-User Library Visibility**: Hide specific libraries per user (even per server) to declutter family members' home screens.
- **Isolated Watch History**: Play progress, played checks, and favorites persist in local SQLite, fully decoupled from the shared upstream account.
- **Role-Based Server Allocation**: Restrict sub-users to only designated upstream nodes.

### 🚀 4. Performance, Bandwidth Savings & Flow Control
- **Concurrent Dispatch & Grace Period**: Fan-out requests across servers with a configurable grace period so fast servers respond immediately while slower ones catch up.
- **Netdisk 302 Direct Stream (`directRedirect`)**: 302-redirects media streams to CDN signed links for netdisk/STRM libraries, saving VPS bandwidth.
- **Anti-Burst Paged Scanning (`pagedScan`)**: Splits large scan requests into 200-item chunks for customized upstream servers that reject 5000-item queries.
- **Strict Concurrency Limits (`maxConcurrent`)**: Restricts concurrent playback sessions per upstream to prevent account bans, returning 429 when saturated.
- **Multiple Streaming Lines**: Define primary and failover streaming URLs (internal NAT, foreign acceleration) with automatic failover.

### 🛡️ 5. Security & High Availability
- **Non-Root Execution**: Runs by default as an unprivileged user (UID 1000).
- **Salted Password Hashing**: Admin credentials stored via salted scrypt hashing.
- **SSRF Defense & Rate Limiting**: Intercepts loopback/private IP scanning in admin APIs; locks out brute-force login attempts for 15 minutes.
- **Node Failover & Auto-Recovery**: Automatically bypasses offline upstream servers and falls back to alternate instances.

### 💻 6. Simple Operations & Developer Experience
- **Embedded Web Admin Panel**: Vue 3 + Tailwind CSS dashboard bundled into the binary with 0 external CDN dependencies.
- **Interactive SSH CLI**: Type `emby-in-one-next` (or `emby-in-one`) to launch a color-coded management menu for updates, credentials, and logs.
- **Pure Static Binary**: Go compilation with zero glibc dependencies for rock-solid portability.

---

## 📖 Quick Start

We strongly recommend deploying via the **Release One-Click Script** (downloads precompiled static binaries for your CPU architecture with zero build overhead). Docker Compose is also available.

### Method 1: Release One-Click Install (Recommended)

Supported on all modern Linux distributions (Debian, Ubuntu, CentOS, AlmaLinux, Arch, etc.):

```bash
curl -fsSL -o release-install.sh https://raw.githubusercontent.com/Takumilove/Emby-In-One-Next/main/release-install.sh
sudo bash release-install.sh
```

This script will:
- Detect your system CPU architecture (`amd64`, `arm64`, `armv7`, `mips`, `mipsle`, `riscv64`) and download the matching static binary;
- Initialize `/opt/emby-in-one-next`, **automatically migrating legacy data if `/opt/emby-in-one` exists**;
- Generate and encrypt a secure random admin password;
- Register and start the `emby-in-one-next.service` systemd service;
- Install the `emby-in-one-next` CLI tool and configure the `emby-in-one` command alias.

### Method 2: Docker Compose

1. Create directories and grant ownership (container runs as UID 1000):
```bash
mkdir -p /opt/emby-in-one-next/{config,data}
chown -R 1000:1000 /opt/emby-in-one-next/config /opt/emby-in-one-next/data
cd /opt/emby-in-one-next
```

2. Create `docker-compose.yml`:
```yaml
services:
  emby-in-one-next:
    image: ghcr.io/takumilove/emby-in-one-next:latest
    container_name: emby-in-one-next
    user: "1000:1000"
    ports:
      - "8096:8096"
    volumes:
      - ./config:/app/config
      - ./data:/app/data
    restart: unless-stopped
    security_opt:
      - no-new-privileges:true
    logging:
      driver: json-file
      options:
        max-size: "10m"
        max-file: "3"
```

3. Launch:
```bash
docker compose up -d
```

### Method 3: Build from Source

Requires Go 1.23+ and a C toolchain (for vendored SQLite CGO):

```bash
git clone https://github.com/Takumilove/Emby-In-One-Next.git
cd Emby-In-One-Next
mkdir -p config data
go run ./cmd/emby-in-one-next
```

### Endpoints & Credentials

- **Emby Client Server URL**: `http://<SERVER_IP>:8096`
- **Web Admin Dashboard**: `http://<SERVER_IP>:8096/admin`
- **Admin Username**: `admin`
- **Admin Password**: Printed at the end of installation. Reset anytime via the SSH CLI menu.

---

## 🛠️ CLI Management Menu

Launch the interactive terminal menu anytime:
```bash
emby-in-one-next
# Or via legacy alias:
emby-in-one
```

Available actions:
1. **Service Controls**: Start, stop, or restart the service (auto-routes to systemd or Docker Compose)
2. **Online Update**: Check and apply updates to the latest Release seamlessly
3. **Security Maintenance**: Inspect admin account, reset admin password, change admin username
4. **User Management**: View sub-user list, provision or delete sub-users
5. **Log Inspection**: Tail the latest 50 lines of logs in real-time
6. **Clean Uninstallation**: Option to keep or purge data/config folders

---

## ⚙️ Configuration Reference

Configuration lives in `config/config.yaml` (modifications in the Web Admin take effect **instantly** without service restart):

```yaml
# dataDir: "/opt/emby-in-one-next/data"    # Runtime data directory (top-level key)

server:
  port: 8096
  name: "Emby-In-One-Next"
  # trustProxy: true        # MUST be true if behind a reverse proxy (Nginx, Caddy, Cloudflare)

admin:
  username: "admin"
  password: "your-strong-password"    # Encrypted into scrypt hash after first startup

playback:
  mode: "proxy"          # "proxy" or "redirect"

timeouts:
  api: 30000             # Single API request timeout (ms)
  global: 15000          # Aggregation total deadline (ms)
  login: 30000           # Upstream login timeout (ms)
  healthCheck: 30000     # Offline node probe timeout (ms)
  healthInterval: 60000  # Health check interval (ms)
  searchGracePeriod: 3000     # Search aggregation grace period (ms)
  metadataGracePeriod: 3000   # Metadata fan-out grace period (ms)
  latestGracePeriod: 0        # "Latest items" grace period (ms)

proxies: []
  # - id: "proxy-1"
  #   name: "HK Proxy"
  #   url: "http://user:pass@ip:port"

upstream:
  - name: "Primary Server"
    url: "https://emby.example.com"
    username: "user"
    password: "pass"
    # Or use API Key:
    # apiKey: "your-api-key"
    playbackMode: "proxy"                      # proxy | redirect
    spoofClient: "none"                        # none | passthrough | infuse | custom
    streamingUrls:                             # Multiple streaming lines (failover)
      - "https://cdn.example.com"
      - "https://backup.example.com"
    followRedirects: true                      # Follow 302 redirects
    directRedirect: true                       # Pass-through netdisk 302 direct links
    pagedScan: false                           # Scan in 200-item chunks
    maxConcurrent: 3                           # Concurrency limit (0 = unconstrained)
    priorityMetadata: false                    # Prefer metadata from this server on conflict
```

---

## 👥 Multi-User Management & Independent History

### User Roles

| Role | Access Scope | Play History & Progress |
| :--- | :--- | :--- |
| **Admin (`admin`)** | Full access to libraries, admin panel, and admin API | Synchronized directly with upstream account |
| **Sub-User (`user`)** | Restricted to assigned upstream servers; blocked from admin APIs | **Locally recorded in SQLite**, isolated from upstream |

### How Independent History Works

1. **Resume & NextUp**: When sub-users stream media on any client, progress reports write to local SQLite (`mappings.db`). Home Resume cards calculate strictly against their own records.
2. **Favorites & Played Status**: Marking items as played or favorited stays local to each sub-user.
3. **Library Filtering**: Customize which upstream libraries are visible per user directly in the Admin Panel.

---

## 🛡️ Security & Troubleshooting

### Reverse Proxy Trust (`trustProxy`)
When operating behind a reverse proxy (Nginx, Caddy, Cloudflare), **ensure `trustProxy: true` is configured** in `config.yaml` and configure your proxy to **overwrite** rather than append client IP headers:
```nginx
proxy_set_header X-Real-IP $remote_addr;
proxy_set_header X-Forwarded-For $remote_addr;
```
This prevents malicious callers from spoofing headers to bypass login brute-force limits.

### Log Files
- **File Location**: `data/emby-in-one-next.log` (10MB per file with 3 automatic rotated backups).
- **Viewing Logs**: Access via SSH menu option `10` or download/clear directly inside the Web Admin Dashboard.

---

## 🏗️ Architecture

```text
Emby-In-One-Next/
├── cmd/emby-in-one-next/           # Application entrypoint & CLI commands
├── internal/backend/               # Core aggregation gateway engine
│   ├── routes.go                   # Route registration and normalization
│   ├── media.go                    # Media fan-out, deduplication & priority
│   ├── media_resume.go             # Cross-server Resume with 403 fallback
│   ├── media_nextup.go             # Cross-server NextUp aggregation
│   ├── media_playback.go           # PlaybackInfo proxy, multi-source merging & concurrency
│   ├── streamproxy.go              # Media streaming proxy, netdisk 302 pass-through & HLS
│   ├── idstore.go                  # SQLite bidirectional virtual ID mapping
│   ├── watch_store.go              # Per-user isolated watch history database
│   ├── identity.go                 # Client identification & 5-tier passthrough spoofing
│   ├── ssrf.go                     # SSRF defense & DNS rebinding protection
│   └── logger.go                   # Structured dual-output rotating logging engine
├── public/                         # Self-hosted embedded Admin Panel (Vue 3 + Tailwind CSS)
├── Dockerfile                      # Multi-stage container build
├── docker-compose.yml              # Docker Compose definition
├── install.sh                      # Source installation script
├── release-install.sh              # Precompiled binary installation script
└── emby-in-one-next-cli.sh         # Interactive SSH CLI management script
```

---

## 📄 Disclaimer

This project is created strictly for personal learning, research, and technical media library aggregation. Please respect upstream terms of service and applicable local laws. The author assumes no liability for upstream account suspension or data loss resulting from client emulation.

---

## 📜 License

This project is open-sourced under the [GNU General Public License v3.0 (GPL-3.0)](LICENSE).
