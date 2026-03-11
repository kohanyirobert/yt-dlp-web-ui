# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Environment

This project runs in a **devcontainer**. Never install dependencies directly (e.g., `apt install`, `go install`, `npm install -g`). Instead, add them to the devcontainer configuration (`.devcontainer/devcontainer.json` or its Dockerfile).

## Project Overview

yt-dlp-web-ui is a full-stack self-hosted web application that provides a browser UI for yt-dlp video downloads. It has a Go backend and a React/TypeScript frontend.

## Commands

### Starting the full app (embedded frontend)
The Go binary embeds the frontend, so the frontend must be built first:
```bash
cd frontend && pnpm install && pnpm build   # Build frontend into frontend/dist/
cd /workspaces/yt-dlp-web-ui && go run main.go -port 3033  # Serves UI + API at http://localhost:3033
```

### Starting frontend + backend separately (hot reload)
```bash
# Terminal 1: backend
cd /workspaces/yt-dlp-web-ui && go run main.go -port 3033

# Terminal 2: frontend dev server with hot reload
cd /workspaces/yt-dlp-web-ui/frontend && pnpm install && pnpm dev  # http://localhost:5173
```

### Backend (Go)
```bash
go run main.go                          # Run server (dev)
go build -o yt-dlp-webui main.go        # Build binary
CGO_ENABLED=0 go build -o yt-dlp-webui main.go  # Static binary
go test ./...                           # Run all tests
go test ./server/rpc/...                # Run tests in a specific package
```

### Frontend (pnpm)
```bash
cd frontend && pnpm install             # Install dependencies
cd frontend && pnpm dev                 # Dev server (http://localhost:5173)
cd frontend && pnpm build               # Production build
```

### Makefile shortcuts
```bash
make           # go run main.go
make fe        # Build frontend (pnpm install && pnpm build)
make dev       # Frontend dev server
make all       # Static Go binary
make multiarch # Build for linux/{amd64,arm64,armv6,armv7}
make clean     # Remove build directory
```

### CLI flags
```
-host    Bind host (default: 0.0.0.0)
-port    Bind port (default: 3033)
-conf    YAML config file path
-out     Downloads directory
-driver  yt-dlp binary path
-db      SQLite database path
-qs      Worker queue size
-auth    Enable authentication
-user    Username
-pass    Password
-fl      Enable file logging
```

## Architecture

### Backend (`/server`)

The Go server uses [chi](https://github.com/go-chi/chi) for routing and serves both the REST API and the embedded React SPA.

**Route structure:**
- `/` — Embedded frontend SPA (hash-based routing)
- `/rpc` — JSON-RPC 2.0 endpoint (WebSocket + HTTP)
- `/api/v1` — REST API for CRUD operations
- `/auth/*` — JWT and OpenID Connect auth
- `/filebrowser` — File system operations
- `/archive` — Download history (SQLite)
- `/subscriptions` — Scheduled download subscriptions
- `/twitch` — Twitch livestream monitor
- `/log` — Server-sent events log stream
- `/status` — Server status info
- `/openapi` — Swagger UI

**Core services** (all singletons, initialized in `server/server.go:RunBlocking`):
- `MemoryDB` — In-memory map of UUID → active download process; serialized to `session.dat` (gob) every 5 minutes for restart recovery
- `MessageQueue` — Internal pub-sub event bus for inter-component communication
- `Worker Pool` — Limits concurrent yt-dlp subprocesses (defaults to 2, forced to 1 on ≤2 CPU systems)
- `CronTaskRunner` — Runs subscription tasks on cron schedules
- `LivestreamMonitor` / `TwitchMonitor` — Poll for live streams at configured intervals

**Download lifecycle:**
1. User submits URL via RPC or REST
2. Handler creates a `Process` with UUID, stores in MemoryDB
3. yt-dlp subprocess spawned with `Setpgid=true` (separate process group for clean teardown)
4. Stdout parsed as JSON progress lines → MemoryDB updated
5. WebSocket broadcasts progress to connected frontend clients

**Configuration** (`server/config/config.go`):
- Singleton `Config` struct loaded from YAML file
- Environment variables override: `USERNAME`, `PASSWORD`, `JWT_SECRET`
- Auth choices: none, JWT (basic), or OpenID Connect

**Authentication middleware** (`server/middleware/`):
- Selected at startup based on config (`require_auth`, `use_openid`)
- JWT tokens passed via `X-Authentication` header or `token` query param

**Persistence:**
- `session.dat` — gob-encoded MemoryDB snapshot (active downloads); location from config `session_file_path`
- SQLite DB — archive history and subscription data; location from config `local_database_path`

**Package pattern:** Feature packages (`rest`, `archive`, `subscription`, `status`) follow a consistent structure: `container.go` (DI wiring), `provider.go` (singleton services via `sync.Once`), `service.go` (business logic), `handlers.go` (HTTP handlers). Each exposes `ApplyRouter()` returning a `func(chi.Router)` that `server/server.go` mounts with `r.Route()`.

**Embedding:** The frontend is compiled into the Go binary via `//go:embed frontend/dist/*` in `main.go`. For development, pass `-web <path>` to serve from the filesystem instead.

### Frontend (`/frontend`)

React 19 + TypeScript + Vite 6. UI components from MUI v6.

**State management:** [Jotai](https://jotai.org/) atoms (atomic, bottom-up). Atoms defined alongside components or in `atoms/` directory.

**Key patterns:**
- RxJS for streaming WebSocket data from the RPC endpoint
- `fp-ts` used for functional error handling in some utilities
- `react-virtuoso` for virtualized lists (large download queues)
- Hash-based routing (`#/downloads`, `#/settings`, etc.)

**API communication:**
- JSON-RPC 2.0 over WebSocket for download operations and real-time progress
- REST (`/api/v1`) for CRUD on subscriptions, archive, settings
- Auth token stored in state/localStorage and sent as header

### Proto / OpenAPI

- `/proto` — Protobuf definitions (currently informational, not compiled into the build)
- `/openapi` — OpenAPI 3.x spec served via Swagger UI at `/openapi`

## Git & GitHub Workflow

- Default branch is `main`. Always work on a **feature branch** and open a PR. Linear project history is required (squash or rebase merge only).
- **Always ask the user before merging a PR** so they can review it on GitHub first.
- CI workflows (`release.yml`, `docker-publish.yml`, `test-container.yml`) use `paths` whitelists — only app source changes (`main.go`, `go.*`, `server/**`, `frontend/**`, `Makefile`, `Dockerfile`) trigger builds/releases.

## Key File Locations

| Purpose | Path |
|---|---|
| Server entry point | `main.go` |
| Server bootstrap & routing | `server/server.go` |
| Config struct | `server/config/config.go` |
| MemoryDB (active downloads) | `server/internal/memory_db.go` |
| Process (yt-dlp subprocess) | `server/internal/process.go` |
| Worker pool & balancer | `server/internal/pool.go`, `server/internal/worker.go` |
| Message queue (pub-sub) | `server/internal/message_queue.go` |
| Auth middleware | `server/middleware/` |
| RPC service | `server/rpc/` |
| REST API | `server/rest/` |
| Frontend entry | `frontend/src/main.tsx` |
| Frontend atoms | `frontend/src/atoms/` |
| Vite config | `frontend/vite.config.mts` |
