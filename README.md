# Hub Dokploy 🚀

An extremely lightweight application developed in Go (Golang) that serves as a **central shortcut HUB** for all your applications hosted on **Dokploy**.

It dynamically scans Traefik configuration files in `/etc/dokploy/traefik/dynamic/` and automatically generates a modern, responsive dashboard with direct links to each application with a configured domain.

---

## ✨ Features

- ⚡ **Ultra Lightweight**: Consumes less than 10 MB of RAM with a static binary compiled in Go.
- 🎨 **Modern Interface (Dark Mode)**: Sleek dashboard with Glassmorphism effect, real-time search, and protocol badges (HTTPS/HTTP).
- 🔍 **Real-Time Filter**: Instant search input to filter applications by name or domain.
- 📋 **1-Click URL Copy**: Convenient buttons to copy application links directly to the clipboard.
- 🐋 **Containerized**: Multi-stage Dockerfile ready for direct deployment on Dokploy or Docker Swarm.
- 🔄 **Dynamic Updates**: Reads Traefik configurations directly from the mounted directory without needing a database.

---

## 🛠️ How It Works

Dokploy stores Traefik dynamic routes in the host directory `/etc/dokploy/traefik/dynamic/`. **Hub Dokploy** reads these files (`.yml`, `.yaml`, `.json`), extracts domain rules from Traefik (`Host(...)`), maps the protocol (HTTPS/HTTP), and automatically generates the application listing.

---

## 🚀 How to Deploy on Dokploy

### Option 1: Via Dockerfile (Recommended on Dokploy)

1. Create a new application in Dokploy pointing to your Git repository.
2. Select **Provider: Dockerfile**.
3. Under **Volumes**, add the following host volume mapping:
   - **Host Path**: `/etc/dokploy/traefik/dynamic`
   - **Container Path**: `/etc/dokploy/traefik/dynamic`
   - **Mode**: `ro` (Read Only)
4. Click **Deploy**.

---

### Option 2: Via Docker Compose

You can also use the included `docker-compose.yml` file:

```yaml
version: '3.8'

services:
  hub-dokploy:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: hub-dokploy
    restart: always
    ports:
      - "8007:8007"
    volumes:
      - /etc/dokploy/traefik/dynamic:/etc/dokploy/traefik/dynamic:ro
    environment:
      - PORT=8007
      - DYNAMIC_CONFIG_DIR=/etc/dokploy/traefik/dynamic
```

---

## ⚙️ Environment Variables

| Variable | Default | Description |
| :--- | :--- | :--- |
| `PORT` | `8007` | Port where the HTTP server will listen |
| `DYNAMIC_CONFIG_DIR` | `/etc/dokploy/traefik/dynamic` | Path to the Traefik dynamic configuration directory |

---

## 💻 Local Development

```bash
# Install dependencies
go mod download

# Run unit tests
go test -v ./...

# Build and run the application locally (using test directory or default)
DYNAMIC_CONFIG_DIR=./test_configs go run .
```
