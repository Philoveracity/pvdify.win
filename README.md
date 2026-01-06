# Pvdify

Heroku-style container deployments made simple.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Overview

Pvdify is a lightweight, self-hosted PaaS (Platform as a Service) that brings Heroku-style deployments to your own infrastructure. Deploy any Docker/OCI container with zero configuration, automatic HTTPS, and a clean web dashboard.

**Live Demo:** [admin.pvdify.win](https://admin.pvdify.win)

## Features

- **Zero-Config Deployments** - Deploy any container image with a single command
- **Automatic HTTPS** - Free SSL certificates via Let's Encrypt
- **Custom Domains** - Point your own domains with simple DNS configuration
- **Config Vars** - Secure environment variable management (12-factor style)
- **Release Management** - Version tracking with instant rollbacks
- **Process Scaling** - Scale dynos up or down as needed
- **Web Dashboard** - Modern, mobile-friendly admin interface
- **RESTful API** - Full-featured API for automation and integrations
- **CLI Tool** - Heroku-compatible command-line interface

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Pvdify Stack                         │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │  Admin UI   │  │  Pvdify CLI │  │   GitHub Extension  │ │
│  │ (SvelteKit) │  │    (Go)     │  │    (gh-pvdify)      │ │
│  └──────┬──────┘  └──────┬──────┘  └──────────┬──────────┘ │
│         │                │                     │            │
│         └────────────────┼─────────────────────┘            │
│                          ▼                                  │
│  ┌───────────────────────────────────────────────────────┐ │
│  │                    pvdifyd (Go)                        │ │
│  │              Control Plane Daemon                      │ │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────────┐  │ │
│  │  │   API   │ │   DB    │ │ Podman  │ │   Systemd   │  │ │
│  │  │ Server  │ │ (SQLite)│ │ Client  │ │  Generator  │  │ │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────────┘  │ │
│  └───────────────────────────────────────────────────────┘ │
│                          │                                  │
│                          ▼                                  │
│  ┌───────────────────────────────────────────────────────┐ │
│  │              Container Runtime (Podman)                │ │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────────┐  │ │
│  │  │  App 1  │ │  App 2  │ │  App 3  │ │    App N    │  │ │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────────┘  │ │
│  └───────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Components

| Component | Technology | Description |
|-----------|------------|-------------|
| **pvdifyd** | Go | Control plane daemon - REST API, database, container orchestration |
| **admin-ui** | SvelteKit + Tailwind | Web dashboard for app management |
| **cli** | Go + Cobra | Command-line interface |
| **gh-pvdify** | Bash | GitHub CLI extension for CI/CD |

---

## Quick Start

### 1. Create an App

```bash
pvdify apps:create my-app
```

### 2. Deploy a Container

```bash
pvdify deploy my-app --image nginx:latest
```

### 3. View Your App

Your app is now live at `https://my-app.pvdify.win`

---

## CLI Reference

### Global Flags

| Flag | Environment Variable | Description |
|------|---------------------|-------------|
| `--api-url` | `PVDIFY_API_URL` | API endpoint (default: https://pvdify.win) |
| `--token` | `PVDIFY_TOKEN` | Authentication token |

### Apps

```bash
# List all apps
pvdify apps

# Create a new app
pvdify apps:create NAME [-e environment]
  -e, --environment   Environment: production (default), staging

# Show app details
pvdify apps:info NAME

# Delete an app
pvdify apps:delete NAME
```

### Deployments

```bash
# Deploy a container image
pvdify deploy NAME --image IMAGE
  -i, --image   Container image to deploy (required)

# List releases
pvdify releases NAME

# Rollback to previous release
pvdify rollback NAME
```

### Config Vars

```bash
# Show all config vars
pvdify config NAME

# Set config vars
pvdify config:set NAME KEY=VALUE [KEY=VALUE...]

# Unset config vars
pvdify config:unset NAME KEY [KEY...]
```

### Domains

```bash
# List domains
pvdify domains NAME

# Add a custom domain
pvdify domains:add NAME DOMAIN

# Remove a domain
pvdify domains:remove NAME DOMAIN
```

### Processes

```bash
# List processes
pvdify ps NAME

# Scale processes
pvdify ps:scale NAME TYPE=COUNT [TYPE=COUNT...]
  Example: pvdify ps:scale my-app web=3

# Restart all processes
pvdify ps:restart NAME
```

### Logs

```bash
# View logs
pvdify logs NAME [-n lines] [-f]
  -n, --lines    Number of lines (default: 100)
  -f, --follow   Stream logs in real-time
```

---

## REST API Reference

Base URL: `https://admin.pvdify.win/api/v1`

### Health Check

```
GET /health
```

Response:
```json
{
  "status": "ok",
  "version": "0.1.0"
}
```

### Apps

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/apps` | List all apps |
| `POST` | `/apps` | Create a new app |
| `GET` | `/apps/{name}` | Get app details |
| `PATCH` | `/apps/{name}` | Update app settings |
| `DELETE` | `/apps/{name}` | Delete an app |

#### Create App

```bash
curl -X POST https://admin.pvdify.win/api/v1/apps \
  -H "Content-Type: application/json" \
  -d '{"name": "my-app", "environment": "production"}'
```

Response:
```json
{
  "name": "my-app",
  "environment": "production",
  "status": "created",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### Releases

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/apps/{name}/releases` | List all releases |
| `POST` | `/apps/{name}/releases` | Create a new release (deploy) |
| `GET` | `/apps/{name}/releases/{version}` | Get specific release |
| `POST` | `/apps/{name}/rollback` | Rollback to previous release |

#### Deploy (Create Release)

```bash
curl -X POST https://admin.pvdify.win/api/v1/apps/my-app/releases \
  -H "Content-Type: application/json" \
  -d '{"image": "nginx:latest"}'
```

### Config Vars

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/apps/{name}/config` | Get all config vars |
| `PUT` | `/apps/{name}/config` | Set config vars |
| `DELETE` | `/apps/{name}/config/{key}` | Unset a config var |

#### Set Config

```bash
curl -X PUT https://admin.pvdify.win/api/v1/apps/my-app/config \
  -H "Content-Type: application/json" \
  -d '{"DATABASE_URL": "postgres://...", "API_KEY": "secret"}'
```

### Domains

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/apps/{name}/domains` | List domains |
| `POST` | `/apps/{name}/domains` | Add a domain |
| `DELETE` | `/apps/{name}/domains/{domain}` | Remove a domain |

### Processes

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/apps/{name}/ps` | List processes |
| `POST` | `/apps/{name}/ps/scale` | Scale processes |
| `POST` | `/apps/{name}/ps/restart` | Restart processes |

### Logs

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/apps/{name}/logs` | Get app logs |

---

## Admin Dashboard

The web dashboard provides a visual interface for all operations:

### Dashboard (`/`)
- View all apps at a glance
- Quick status indicators (running, stopped, deploying, failed)
- One-click navigation to app details

### Create App (`/apps/new`)
- Form-based app creation
- Environment selection (production/staging)
- Real-time name validation

### App Details (`/apps/{name}`)

**Overview Tab**
- Process/dyno status and scaling
- Latest release information
- Quick stats (releases, config vars, domains, dynos)

**Deploy Tab**
- CLI deployment instructions
- Full release history with rollback options

**Config Tab**
- View/hide config var values
- Secure environment variable management

**Settings Tab**
- App metadata and timestamps
- Domain management
- Danger zone (delete app)

### Status Page (`/status`)
- Real-time system health monitoring
- API, Database, and Container Runtime status
- Auto-refresh every 30 seconds

---

## Data Models

### App

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Unique app identifier |
| `environment` | string | `production` or `staging` |
| `status` | string | `created`, `running`, `stopped`, `failed`, `deleting` |
| `image` | string | Current container image |
| `bind_port` | int | Internal container port |
| `resources` | object | CPU/memory limits |
| `healthcheck` | object | Health check configuration |
| `created_at` | datetime | Creation timestamp |
| `updated_at` | datetime | Last update timestamp |

### Release

| Field | Type | Description |
|-------|------|-------------|
| `version` | int | Sequential release number |
| `image` | string | Container image for this release |
| `status` | string | `pending`, `deploying`, `current`, `rolled_back` |
| `created_at` | datetime | Deployment timestamp |

### Process

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | Process type (e.g., `web`, `worker`) |
| `count` | int | Number of instances |
| `command` | string | Override command (optional) |

---

## Configuration

### pvdifyd Configuration

The daemon can be configured via environment variables or config file:

| Variable | Default | Description |
|----------|---------|-------------|
| `PVDIFY_LISTEN` | `0.0.0.0:9443` | Listen address |
| `PVDIFY_DB_PATH` | `/var/lib/pvdify/pvdifyd.db` | SQLite database path |
| `PVDIFY_STATIC_DIR` | `/opt/pvdify/admin-ui/dist` | Admin UI static files |
| `PVDIFY_DEV` | `false` | Development mode |
| `PVDIFY_TLS_ENABLED` | `false` | Enable TLS |
| `PVDIFY_TLS_CERT` | - | TLS certificate path |
| `PVDIFY_TLS_KEY` | - | TLS private key path |

---

## Project Structure

```
pvdify.win/
├── pvdifyd/                 # Control plane daemon
│   ├── cmd/pvdifyd/         # Entry point
│   └── internal/
│       ├── api/             # HTTP handlers
│       ├── config/          # Configuration
│       ├── db/              # Database layer (SQLite)
│       ├── models/          # Data models
│       ├── podman/          # Container runtime client
│       ├── systemd/         # Unit file generator
│       └── tunnel/          # Cloudflare tunnel config
├── cli/                     # CLI tool
│   ├── cmd/pvdify/          # Commands
│   └── internal/client/     # API client
├── admin-ui/                # Web dashboard
│   └── src/
│       ├── routes/          # SvelteKit pages
│       └── app.css          # Tailwind styles
├── gh-pvdify/               # GitHub CLI extension
├── .gitignore
├── LICENSE
└── README.md
```

---

## Development

### Prerequisites

- Go 1.21+
- Node.js 18+
- Podman

### Build pvdifyd

```bash
cd pvdifyd
go build -o pvdifyd ./cmd/pvdifyd
```

### Build CLI

```bash
cd cli
go build -o pvdify ./cmd/pvdify
```

### Build Admin UI

```bash
cd admin-ui
npm install
npm run build
```

### Run Locally

```bash
# Start the daemon in dev mode
./pvdifyd --dev

# Admin UI available at http://localhost:9443
# API available at http://localhost:9443/api/v1
```

---

## Deployment

### Systemd Service

```ini
[Unit]
Description=Pvdify Control Plane Daemon
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/pvdifyd
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### Reverse Proxy (Caddy)

```
admin.pvdify.win {
    reverse_proxy localhost:9443
}

*.pvdify.win {
    reverse_proxy localhost:9443
}
```

---

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## License

MIT License - see [LICENSE](LICENSE) for details.

---

## Links

- **Dashboard:** [admin.pvdify.win](https://admin.pvdify.win)
- **Status:** [admin.pvdify.win/status](https://admin.pvdify.win/status)
- **GitHub:** [github.com/Philoveracity/pvdify.win](https://github.com/Philoveracity/pvdify.win)
