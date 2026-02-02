# Go Proxy Manager (GPM) - Development Plan

GPM is a lightweight, high-performance proxy pool manager written in Go. It aggregates, validates, and serves proxy lists via a simple HTTP API.

## Core Features
- Multi-source proxy fetching (HTTP/HTTPS, JSON/Text).
- Concurrent proxy validation (Worker Pool).
- Sorted proxy list (by latency) via HTTP API.
- Configurable fetch and check intervals.
- Structured logging with `slog`.
- Docker and Docker Compose support.

## Technical Stack
- **Language:** Go 1.22+
- **Configuration:** YAML (gopkg.in/yaml.v3)
- **Logging:** `log/slog`
- **Concurrency:** Goroutines, Channels, `sync.RWMutex`, Worker Pool
- **API:** `net/http` (Standard Library)
- **Deployment:** Docker Multi-stage build

## Project Layout
```text
go-proxy-manager/
├── cmd/gpm/main.go           # Entry point
├── configs/config.yaml       # Default configuration
├── internal/
│   ├── config/               # Config loading
│   ├── model/                # Data structures (Proxy, Source)
│   ├── fetcher/              # Proxy fetching logic
│   ├── checker/              # Proxy validation (concurrent)
│   ├── manager/              # Coordinator
│   ├── store/                # Thread-safe in-memory storage
│   └── server/               # HTTP API server
├── pkg/logger/               # Logging wrapper
├── build/Dockerfile          # Containerization
├── deploy/docker-compose.yml # Orchestration
└── AGENTS.md                 # Development plan
```

## Implementation Phases

### Phase 1: Models & Configuration
- Define `Proxy` and `Source` structures.
- Implement YAML config loading.

### Phase 2: Fetcher & Checker
- Implement Fetcher for Text and JSON sources.
- Implement concurrent Checker using a Worker Pool.

### Phase 3: Store & Manager
- Implement thread-safe storage with latency-based sorting.
- Implement Manager to orchestrate fetching and checking cycles.

### Phase 4: Server & CLI
- Implement `net/http` server with `/api/v1/proxies` endpoint.
- Implement `main.go` for initialization.

### Phase 5: Docker & Optimization
- Write Dockerfile and docker-compose.yml.
- Add health checks and graceful shutdown.
