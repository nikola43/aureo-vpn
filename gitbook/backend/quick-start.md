# 🚀 Quick Start

Get the Aureo VPN backend running in minutes.

---

## 📋 Prerequisites

- Go 1.24+
- Docker & Docker Compose (optional)
- Redis 7+ (optional, for caching)
- WireGuard tools (for VPN nodes)

---

## 🐳 Using Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/nikola43/aureo-vpn.git
cd aureo-vpn

# Start all services
cd deployments/docker
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f api-gateway
```

The services will be available at:
- API Gateway: http://localhost:8080
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)
- Dashboard: http://localhost:5000

---

## 🔧 Manual Setup

```bash
# Clone and enter directory
git clone https://github.com/nikola43/aureo-vpn.git
cd aureo-vpn

# Install dependencies
make setup

# Configure environment
cp .env.example .env
# Edit .env with your settings

# Start Redis (optional, for caching/rate limiting)
docker run -d --name aureo-redis \
  -p 6379:6379 \
  redis:7-alpine

# Build all services
make build

# Run API Gateway (SQLite database is created automatically)
./bin/api-gateway

# Run Control Server (in another terminal)
./bin/control-server

# Run VPN Node (requires root)
sudo ./bin/vpn-node
```

{% hint style="info" %}
SQLite database file is created automatically at startup. No external database setup required.
{% endhint %}

---

## 📦 Building from Source

```bash
# Build all components
make build

# Or build individually
make build-api       # API Gateway
make build-control   # Control Server
make build-node      # VPN Node
make build-cli       # CLI Tool
```

### Binary Outputs

After building:
- `bin/api-gateway` — API server
- `bin/control-server` — Control plane
- `bin/vpn-node` — VPN server
- `bin/cli` — Management CLI

---

## 💻 System Requirements

**Minimum (Development)**:
- 2 CPU cores
- 2GB RAM
- 10GB disk

**Recommended (Production)**:
- 4+ CPU cores
- 8GB+ RAM
- 50GB+ SSD
- 1Gbps network
