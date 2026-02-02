# Go Proxy Manager (GPM)

[English](README.md) | [中文文档](README_CN.md)

GPM is a lightweight, high-performance proxy pool manager written in Go. It aggregates, validates, and serves proxy lists via a simple HTTP API.

## Features

- **Multi-source fetching**: Supports text and JSON sources.
- **Protocol support**: HTTP, HTTPS, SOCKS4, SOCKS5.
- **Concurrent validation**: High-performance worker pool for proxy checking.
- **Automatic sorting**: Proxies are returned sorted by latency.
- **Health monitoring**: Periodic re-validation of stored proxies.
- **Persistence**: Save valid proxies to disk on shutdown and reload on restart.
- **Structured logging**: JSON logs using `log/slog`.
- **Containerized**: Docker and Docker Compose ready.

## Getting Started

### Prerequisites

- Go 1.22+
- Docker & Docker Compose (optional)

### Configuration

The configuration is located in `configs/config.yaml`. You can add your proxy sources there.

```yaml
app:
  port: 8080
  log_level: "info"
  thread_count: 50
  cache_file: "data/proxies.json" # Path to store proxy cache

validation:
  target_urls:
    - "https://www.google.com"
  timeout: 5s
  interval: 60s

sources:
  - url: "http://example.com/proxies.txt"
    type: "text"
    interval: 300s
```

### Installation & Running

#### Using Makefile

```bash
# Build the binary
make build

# Run the application
make run
```

#### Using Docker

The `deploy/docker-compose.yml` includes a volume mount for `data/` to ensure proxies are persisted across restarts.

```bash
# Build and start the container
docker-compose -f deploy/docker-compose.yml up -d
```

## API Reference

### Get Proxies

Returns a list of valid proxies sorted by latency.

**Endpoint**: `GET /api/v1/proxies`

**Parameters**:
- `limit` (optional): Number of proxies to return.
- `format` (optional): Output format. Default is `json`. Set to `text` for a plain text list (protocol://ip:port).

**Response (JSON)**:
```json
[
  {
    "url": "http://1.2.3.4:8080",
    "protocol": "http",
    "ip": "1.2.3.4",
    "port": 8080,
    "latency": 150000000,
    "last_check": "2024-02-02T12:00:00Z",
    "fail_count": 0,
    "source": "http://example.com/proxies.txt"
  }
]
```

**Response (Text)**:
```text
http://1.2.3.4:8080
socks5://5.6.7.8:1080
```

### Health Check

**Endpoint**: `GET /health`

## Development

### Project Structure

- `cmd/gpm`: Application entry point.
- `internal/`: Core logic (fetcher, checker, manager, store, server).
- `pkg/`: Shared packages (logger).
- `configs/`: Configuration files.
- `build/`: Dockerfile.
- `deploy/`: Docker Compose.

### License

MIT