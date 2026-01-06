# Pvdify

Heroku-style container deployments made simple.

## Overview

Pvdify is a lightweight PaaS (Platform as a Service) that makes deploying containerized applications as simple as Heroku. Deploy any Docker container with zero configuration.

## Features

- **Simple Deployments** - Deploy containers with a single command
- **Automatic HTTPS** - Free SSL certificates via Let's Encrypt
- **Custom Domains** - Add your own domain with DNS configuration
- **Config Vars** - Secure environment variable management
- **Release History** - Track and rollback deployments
- **Process Scaling** - Scale your dynos up or down

## Quick Start

### Create an App

```bash
pvdify apps:create my-app
```

### Deploy

```bash
pvdify deploy my-app --image nginx:latest
```

Your app will be available at `https://my-app.pvdify.win`

## CLI Commands

| Command | Description |
|---------|-------------|
| `pvdify apps` | List all apps |
| `pvdify apps:create NAME` | Create a new app |
| `pvdify deploy NAME --image IMAGE` | Deploy a container |
| `pvdify config:set KEY=value` | Set config var |
| `pvdify domains:add NAME DOMAIN` | Add custom domain |
| `pvdify ps:scale web=2` | Scale processes |
| `pvdify releases` | List releases |

## Admin Dashboard

Access the web dashboard at [admin.pvdify.win](https://admin.pvdify.win)

## Status

Check system status at [admin.pvdify.win/status](https://admin.pvdify.win/status)

## License

MIT
