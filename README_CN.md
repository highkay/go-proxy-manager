# Go Proxy Manager (GPM)

[English](README.md) | [中文文档](README_CN.md)

GPM 是一个用 Go 语言编写的轻量级、高性能代理池管理工具。它负责聚合、验证并通过简单的 HTTP API 分发代理列表。

## 功能特性

- **多源抓取**: 支持文本 (Text) 和 JSON 格式的代理源。
- **多协议支持**: 支持 HTTP, HTTPS, SOCKS4, SOCKS5 协议。
- **并发验证**: 使用高性能协程池 (Worker Pool) 进行代理验证。
- **自动排序**: 返回的代理列表按延迟 (Latency) 自动升序排列。
- **健康监控**: 定期重新验证已存储的代理，自动剔除失效代理。
- **持久化**: 停机时自动保存有效代理至磁盘，重启后立即加载。
- **结构化日志**: 使用 `log/slog` 输出 JSON 格式日志，易于监控集成。
- **容器化**: 支持 Docker 和 Docker Compose 一键部署。

## 快速开始

### 前置要求

- Go 1.22+
- Docker & Docker Compose (可选)

### 配置文件

配置文件位于 `configs/config.yaml`。你可以在此处配置应用参数及代理源。

```yaml
app:
  port: 8080           # 服务端口
  log_level: "info"    # 日志级别
  thread_count: 50     # 验证并发数
  cache_file: "data/proxies.json" # 代理缓存文件路径

validation:
  target_urls:         # 用于验证代理有效性的目标 URL
    - "https://stock.finance.sina.com.cn/..."
  timeout: 10s         # 验证超时时间
  interval: 180s       # 已存代理重验间隔

sources:               # 代理源列表
  - url: "http://example.com/proxies.txt"
    type: "text"
    interval: 3600s
```

### 安装与运行

#### 使用 Makefile

```bash
# 编译二进制文件
make build

# 运行应用
make run
```

#### 使用 Docker

`deploy/docker-compose.yml` 已经配置了 `data/` 目录的挂载，确保代理数据在重启后保留。

```bash
# 构建并启动容器
docker-compose -f deploy/docker-compose.yml up -d
```

## API 接口文档

### 获取代理列表

返回按延迟排序的可用代理列表。

**接口地址**: `GET /api/v1/proxies`

**请求参数**:
- `limit` (可选): 返回的代理数量限制。
- `format` (可选): 返回格式。默认为 `json`，可选 `text`。
  - `text`: 返回纯文本格式 `protocol://ip:port`，每行一个。

**响应示例 (JSON)**:
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

**响应示例 (Text)**:
```text
http://1.2.3.4:8080
socks5://5.6.7.8:1080
```

### 健康检查

**接口地址**: `GET /health`

## 开发指南

### 项目结构

- `cmd/gpm`: 程序入口。
- `internal/`: 核心逻辑 (fetcher, checker, manager, store, server)。
- `pkg/`: 公共包 (logger)。
- `configs/`: 配置文件。
- `build/`: Dockerfile 构建文件。
- `deploy/`: Docker Compose 部署文件。

### 许可证

MIT
