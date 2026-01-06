# Pvdify

Heroku-style container deployments made simple.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Overview

Pvdify is a lightweight, self-hosted PaaS (Platform as a Service) that brings Heroku-style deployments to your own infrastructure. Deploy any Docker/OCI container with zero configuration, automatic HTTPS, and a clean web dashboard.

## Features

- **Zero-Config Deployments** - Deploy any container image with a single command
- **Automatic HTTPS** - Free SSL certificates via Let's Encrypt or Cloudflare
- **Custom Domains** - One-click Cloudflare DNS integration
- **Config Vars** - Secure environment variable management (12-factor style)
- **Release Management** - Version tracking with instant rollbacks
- **Process Scaling** - Scale dynos horizontally as needed
- **Web Dashboard** - Modern, mobile-friendly admin interface
- **RESTful API** - Full-featured API for automation and integrations
- **CLI Tool** - Heroku-compatible command-line interface
- **GitHub Integration** - Deploy directly from GitHub Actions

---

## System Requirements

### Minimum Hardware

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| CPU | 2 cores | 4+ cores |
| RAM | 2 GB | 4+ GB |
| Storage | 20 GB SSD | 50+ GB SSD |
| Network | 100 Mbps | 1 Gbps |

### Software Requirements

| Component | Version | Purpose |
|-----------|---------|---------|
| **Operating System** | AlmaLinux 8+, RHEL 8+, Ubuntu 22.04+ | Host OS |
| **Go** | 1.21+ | Build pvdifyd and CLI |
| **Node.js** | 18+ LTS | Build admin UI |
| **Podman** | 4.0+ | Container runtime (rootless supported) |
| **SQLite** | 3.35+ | Application database |
| **Systemd** | 239+ | Service management |

### Network Requirements

| Port | Protocol | Purpose |
|------|----------|---------|
| 80 | TCP | HTTP redirect to HTTPS |
| 443 | TCP | HTTPS (Admin UI, API, App traffic) |
| 9443 | TCP | Internal API (localhost only) |

### External Services (Optional)

| Service | Purpose |
|---------|---------|
| **Cloudflare** | DNS management, SSL, CDN, DDoS protection |
| **Let's Encrypt** | Free SSL certificates (if not using Cloudflare) |
| **Container Registry** | Docker Hub, GitHub Container Registry, or private |

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Internet                                     │
└───────────────────────────────┬─────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    Cloudflare (Optional)                             │
│            DNS, SSL Termination, CDN, WAF                           │
└───────────────────────────────┬─────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     Reverse Proxy Layer                              │
│              (Caddy / LiteSpeed / Nginx)                            │
│                                                                      │
│   ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────────┐ │
│   │ admin.*.win │  │  app1.*.win │  │        app2.*.win           │ │
│   └──────┬──────┘  └──────┬──────┘  └──────────────┬──────────────┘ │
└──────────┼────────────────┼────────────────────────┼────────────────┘
           │                │                        │
           ▼                ▼                        ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        Pvdify Stack                                  │
├─────────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────────┐  │
│  │  Admin UI   │  │  Pvdify CLI │  │      GitHub Extension       │  │
│  │ (SvelteKit) │  │    (Go)     │  │       (gh-pvdify)           │  │
│  └──────┬──────┘  └──────┬──────┘  └──────────────┬──────────────┘  │
│         │                │                        │                  │
│         └────────────────┼────────────────────────┘                  │
│                          ▼                                           │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │                      pvdifyd (Go)                              │  │
│  │                 Control Plane Daemon                           │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────────┐  │  │
│  │  │   REST   │ │  SQLite  │ │  Podman  │ │     Systemd      │  │  │
│  │  │   API    │ │    DB    │ │  Client  │ │    Generator     │  │  │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────────────┘  │  │
│  └───────────────────────────────────────────────────────────────┘  │
│                          │                                           │
│                          ▼                                           │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │                 Container Runtime (Podman)                     │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────────┐  │  │
│  │  │  App 1   │ │  App 2   │ │  App 3   │ │      App N       │  │  │
│  │  │ :3000    │ │ :3001    │ │ :3002    │ │     :300N        │  │  │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────────────┘  │  │
│  └───────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
```

### Component Overview

| Component | Technology | Description |
|-----------|------------|-------------|
| **pvdifyd** | Go 1.21+ | Control plane daemon - REST API, SQLite database, container orchestration |
| **admin-ui** | SvelteKit 2 + Tailwind CSS | Responsive web dashboard for app management |
| **cli** | Go + Cobra | Heroku-compatible command-line interface |
| **gh-pvdify** | Bash | GitHub CLI extension for CI/CD pipelines |

### Data Flow

1. **User Request** → Cloudflare (DNS/SSL) → Reverse Proxy → pvdifyd API
2. **App Traffic** → Cloudflare → Reverse Proxy → Container (via port mapping)
3. **Deployment** → CLI/API → pvdifyd → Podman → Systemd unit → Running container

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

### 3. Add a Custom Domain (Optional)

```bash
pvdify domains:add my-app myapp.example.com
```

### 4. View Your App

Your app is live at `https://my-app.yourdomain.com`

---

## CLI Reference

### Installation

```bash
# Download latest release
curl -fsSL https://github.com/Philoveracity/pvdify.win/releases/latest/download/pvdify-linux-amd64 -o pvdify
chmod +x pvdify
sudo mv pvdify /usr/local/bin/
```

### Configuration

```bash
# Set API endpoint
export PVDIFY_API_URL="https://your-pvdify-server.com"

# Set authentication token (if enabled)
export PVDIFY_TOKEN="your-api-token"
```

### Global Flags

| Flag | Environment Variable | Description |
|------|---------------------|-------------|
| `--api-url` | `PVDIFY_API_URL` | Pvdify API endpoint |
| `--token` | `PVDIFY_TOKEN` | Authentication token |

### App Management

```bash
# List all apps
pvdify apps

# Create a new app
pvdify apps:create NAME [-e environment]
  -e, --environment   Environment: production (default), staging

# Show app details
pvdify apps:info NAME

# Delete an app (requires confirmation)
pvdify apps:delete NAME
```

### Deployments

```bash
# Deploy a container image
pvdify deploy NAME --image IMAGE
  -i, --image   Container image to deploy (required)

# Examples:
pvdify deploy my-app --image nginx:latest
pvdify deploy my-app --image ghcr.io/myorg/myapp:v1.2.3
pvdify deploy my-app --image my-registry.com/app:latest

# List releases
pvdify releases NAME

# Rollback to previous release
pvdify rollback NAME
```

### Config Vars (Environment Variables)

```bash
# Show all config vars (values hidden)
pvdify config NAME

# Set one or more config vars
pvdify config:set NAME KEY=VALUE [KEY=VALUE...]

# Examples:
pvdify config:set my-app DATABASE_URL=postgres://...
pvdify config:set my-app NODE_ENV=production PORT=3000

# Unset config vars
pvdify config:unset NAME KEY [KEY...]
```

### Custom Domains

```bash
# List domains for an app
pvdify domains NAME

# Add a custom domain
pvdify domains:add NAME DOMAIN

# Remove a domain
pvdify domains:remove NAME DOMAIN

# Example workflow:
pvdify domains:add my-app app.example.com
# Then add CNAME record: app.example.com → my-app.pvdify.win
```

### Process Management

```bash
# List processes (dynos)
pvdify ps NAME

# Scale processes
pvdify ps:scale NAME TYPE=COUNT [TYPE=COUNT...]

# Examples:
pvdify ps:scale my-app web=3        # Scale web to 3 instances
pvdify ps:scale my-app worker=2     # Scale workers

# Restart all processes
pvdify ps:restart NAME
```

### Logs

```bash
# View recent logs
pvdify logs NAME

# View more lines
pvdify logs NAME -n 500

# Stream logs in real-time
pvdify logs NAME -f
```

---

## REST API Reference

### Base URL

```
https://your-pvdify-server.com/api/v1
```

### Authentication

If authentication is enabled, include the token in the Authorization header:

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" https://api.example.com/api/v1/apps
```

### Health Check

```http
GET /health
```

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
curl -X POST https://api.example.com/api/v1/apps \
  -H "Content-Type: application/json" \
  -d '{"name": "my-app", "environment": "production"}'
```

#### Response

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
| `POST` | `/apps/{name}/releases` | Create release (deploy) |
| `GET` | `/apps/{name}/releases/{version}` | Get specific release |
| `POST` | `/apps/{name}/rollback` | Rollback to previous |

#### Deploy

```bash
curl -X POST https://api.example.com/api/v1/apps/my-app/releases \
  -H "Content-Type: application/json" \
  -d '{"image": "nginx:latest"}'
```

### Config Vars

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/apps/{name}/config` | Get all config vars |
| `PUT` | `/apps/{name}/config` | Set config vars |
| `DELETE` | `/apps/{name}/config/{key}` | Unset a config var |

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
| `GET` | `/apps/{name}/logs` | Get application logs |

---

## Admin Dashboard

The web dashboard provides a visual interface for all operations.

### Pages

| Route | Description |
|-------|-------------|
| `/` | Dashboard - View all apps with status |
| `/apps/new` | Create a new application |
| `/apps/{name}` | App details with tabs |
| `/status` | System health monitoring |

### App Details Tabs

**Overview**
- Process/dyno status with scaling controls
- Latest release information
- Quick stats (releases, config vars, domains, dynos)

**Deploy**
- CLI deployment instructions
- Full release history
- One-click rollback

**Config**
- View/hide sensitive config values
- Add/remove environment variables

**Settings**
- App metadata and timestamps
- Domain management with Cloudflare integration
- Danger zone (delete app)

---

## Cloudflare Integration

Pvdify integrates with Cloudflare for DNS management and SSL.

### Setup

1. **API Token**: Create a Cloudflare API token with Zone:DNS:Edit permissions
2. **Configure**: Set `CLOUDFLARE_API_TOKEN` environment variable on your Pvdify server
3. **Use**: The admin UI will show "Connect to Cloudflare" buttons for domain management

### Features

- **One-Click DNS**: Add CNAME records directly from the dashboard
- **Automatic SSL**: Cloudflare provides free SSL certificates
- **Proxy Mode**: Enable Cloudflare proxy for CDN and DDoS protection

---

## Data Models

### App

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Unique identifier (lowercase, alphanumeric, hyphens) |
| `environment` | string | `production` or `staging` |
| `status` | string | `created`, `running`, `stopped`, `failed`, `deleting` |
| `image` | string | Current container image |
| `bind_port` | int | Container port to expose |
| `resources` | object | CPU/memory limits |
| `healthcheck` | object | Health check configuration |
| `created_at` | datetime | Creation timestamp |
| `updated_at` | datetime | Last modification |

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

## Security Considerations

### Network Security

- Run pvdifyd on localhost only (use reverse proxy for external access)
- Enable HTTPS via Cloudflare or Let's Encrypt
- Use firewall rules to restrict port access

### Authentication

- Enable API token authentication in production
- Use environment variables for sensitive configuration
- Rotate tokens periodically

### Container Security

- Run containers as non-root users
- Use read-only root filesystems where possible
- Scan images for vulnerabilities before deployment
- Use private registries for proprietary images

### Data Security

- Config vars are stored encrypted at rest
- Database file permissions should be restricted
- Regular backups recommended

---

## Project Structure

```
pvdify.win/
├── pvdifyd/                 # Control plane daemon (Go)
│   ├── cmd/pvdifyd/         # Main entry point
│   └── internal/
│       ├── api/             # REST API handlers
│       ├── config/          # Configuration management
│       ├── db/              # SQLite database layer
│       ├── models/          # Data structures
│       ├── podman/          # Container runtime client
│       └── systemd/         # Unit file generator
├── cli/                     # Command-line tool (Go)
│   ├── cmd/pvdify/          # CLI commands
│   └── internal/client/     # API client library
├── admin-ui/                # Web dashboard (SvelteKit)
│   ├── src/routes/          # Page components
│   └── src/app.css          # Tailwind styles
├── gh-pvdify/               # GitHub CLI extension
├── .gitignore
├── LICENSE
└── README.md
```

---

## Development

### Prerequisites

- Go 1.21+
- Node.js 18+ LTS
- Podman 4.0+

### Build All Components

```bash
# Build control plane daemon
cd pvdifyd && go build -o pvdifyd ./cmd/pvdifyd

# Build CLI
cd cli && go build -o pvdify ./cmd/pvdify

# Build admin UI
cd admin-ui && npm install && npm run build
```

### Run Development Server

```bash
# Start pvdifyd in dev mode
./pvdifyd --dev

# Access:
# - Admin UI: http://localhost:9443
# - API: http://localhost:9443/api/v1
```

---

## Deployment Guide

### 1. Install Dependencies

```bash
# AlmaLinux/RHEL
sudo dnf install -y podman sqlite

# Ubuntu/Debian
sudo apt install -y podman sqlite3
```

### 2. Install Pvdify

```bash
# Download and install binaries
sudo cp pvdifyd /usr/local/bin/
sudo cp pvdify /usr/local/bin/

# Create data directory
sudo mkdir -p /var/lib/pvdify
```

### 3. Create Systemd Service

```ini
# /etc/systemd/system/pvdifyd.service
[Unit]
Description=Pvdify Control Plane Daemon
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/pvdifyd
Restart=always
RestartSec=5
Environment=PVDIFY_STATIC_DIR=/opt/pvdify/admin-ui/dist

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now pvdifyd
```

### 4. Configure Reverse Proxy

**Caddy (Recommended)**
```
your-domain.com {
    reverse_proxy localhost:9443
}
```

**Nginx**
```nginx
server {
    listen 443 ssl;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:9443;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

---

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`go test ./...`)
5. Commit (`git commit -m 'Add amazing feature'`)
6. Push (`git push origin feature/amazing-feature`)
7. Open a Pull Request

---

## License

MIT License - see [LICENSE](LICENSE) for details.

---

## Links

- **GitHub:** [github.com/Philoveracity/pvdify.win](https://github.com/Philoveracity/pvdify.win)
- **Issues:** [github.com/Philoveracity/pvdify.win/issues](https://github.com/Philoveracity/pvdify.win/issues)
