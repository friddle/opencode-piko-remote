# opencode-piko-remote

Expose [opencode](https://github.com/anomalyco/opencode) web interface remotely through [piko](https://github.com/andydunstall/piko) tunnel.

## Architecture

```
[Browser] → [Piko Server (nginx + piko)] → [Piko Agent] → [opencode web]
```

The client binary embeds opencode, starts it in web mode, and connects through piko to a remote server — making your AI coding assistant accessible from anywhere.

## Quick Start

### Client (on your dev machine)

```bash
opencode-piko /path/to/project \
  --name=my-dev \
  --remote=piko-server.example.com:8088 \
  --pass=your-password
```

Then access `http://piko-server.example.com:8088/my-dev/` in your browser.

### Server (public-facing, Docker)

```bash
cd server
docker compose up -d
```

Exposes:
- `:8088` — HTTP access (nginx → piko proxy, routed by endpoint name)
- `:8022` — Piko upstream port (for client connections)

## Client Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `[project]` | No | `.` | Working directory for opencode |
| `--name` | Yes | — | Piko endpoint name (also the URL path) |
| `--remote` | Yes | — | Piko server address `host:port` |
| `--user` | No | `opencode` | Auth username |
| `--pass` | No | — | Auth password |
| `--server-port` | No | `8022` | Piko upstream port |
| `--auto-exit` | No | `true` | Auto shutdown after 24h |

## Commands

```bash
opencode-piko [project] [flags]   # Start service
opencode-piko upgrade             # Update embedded opencode to latest
```

## Build

```bash
cd client

# Current platform
make build

# All 4 platforms (darwin/linux × amd64/arm64)
make build-all

# Docker image
make docker
```

Requires Go 1.23+ and internet access (downloads opencode from GitHub Releases during build).

## Docker (Client)

```bash
cd client
make download-opencode TARGET_OS=linux TARGET_ARCH=amd64
make docker
docker run -d opencode-piko:latest --name=dev --remote=your-server:8088 --pass=secret
```

## How It Works

1. On first run, extracts embedded opencode binary to `~/.opencode/bin/`
2. Starts `opencode web` on a random local port with `OPENCODE_SERVER_PASSWORD`
3. Starts piko agent connecting to the remote piko server
4. Traffic flows: browser → piko server → piko agent → opencode web

## License

MIT
