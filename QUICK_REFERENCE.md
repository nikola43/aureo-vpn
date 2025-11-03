# Aureo VPN - Quick Reference Guide

## 🚀 Getting Started in 5 Minutes

### Prerequisites
```bash
- Go 1.22+
- Docker & Docker Compose
- PostgreSQL 15+ (or use Docker)
- Git
```

### Installation
```bash
# 1. Clone repository
git clone https://github.com/nikola43/aureo-vpn.git
cd aureo-vpn

# 2. Run setup (installs dependencies, creates DB, builds binaries)
make setup

# 3. Start services with Docker
make docker-up

# 4. Check if running
curl http://localhost:8080/health
```

## 📦 Project Structure

```
aureo-vpn/
├── cmd/                       # Applications
│   ├── api-gateway/          # REST API server
│   ├── control-server/       # Control plane
│   ├── vpn-node/            # VPN node service
│   └── cli/                 # Management CLI
├── pkg/                      # Shared libraries
│   ├── auth/                # JWT authentication
│   ├── crypto/              # Encryption (AES, ChaCha20)
│   ├── protocols/           # WireGuard, OpenVPN, IKEv2
│   ├── payment/             # Cryptocurrency
│   ├── proxy/               # SOCKS5
│   └── transparency/        # Warrant canary
├── internal/                # Private code
│   ├── api/                # API handlers
│   ├── node/               # Node service
│   ├── control/            # Control logic
│   └── security/           # Security features
└── deployments/            # Docker & K8s configs
```

## 🔧 Common Commands

### Make Targets
```bash
make help              # Show all available commands
make build            # Build all binaries
make test             # Run tests
make docker-up        # Start with Docker Compose
make docker-down      # Stop Docker services
make lint             # Run linter
make clean            # Clean build artifacts
```

### CLI Commands
```bash
# Node management
./bin/aureo-vpn node create --name "US-1" --hostname "vpn.example.com" \
  --ip "1.2.3.4" --country "United States" --country-code "US" --city "New York"

./bin/aureo-vpn node list
./bin/aureo-vpn node delete <node-id>

# Configuration
./bin/aureo-vpn config generate --user <user-id> --node <node-id> \
  --protocol wireguard --output client.conf

# Statistics
./bin/aureo-vpn stats
./bin/aureo-vpn user list
```

### API Endpoints
```bash
# Authentication
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"pass123","username":"user"}'

curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"pass123"}'

# Get nodes
curl -X GET "http://localhost:8080/api/v1/nodes?country=US" \
  -H "Authorization: Bearer <token>"

# Get best node
curl -X GET "http://localhost:8080/api/v1/nodes/best?protocol=wireguard" \
  -H "Authorization: Bearer <token>"

# Health check
curl http://localhost:8080/health

# Metrics
curl http://localhost:8080/metrics
```

## 🐳 Docker Quick Reference

```bash
# Start all services
docker-compose -f deployments/docker/docker-compose.yml up -d

# View logs
docker-compose logs -f api-gateway
docker-compose logs -f vpn-node-1

# Stop services
docker-compose down

# Rebuild images
docker-compose build --no-cache

# Check status
docker-compose ps
```

## 🎯 Environment Variables

### Required
```bash
DB_HOST=localhost          # PostgreSQL host
DB_PORT=5432              # PostgreSQL port
DB_USER=postgres          # Database user
DB_PASSWORD=postgres      # Database password
DB_NAME=aureo_vpn        # Database name
JWT_SECRET=your-secret    # JWT signing key
```

### Optional
```bash
PORT=8080                 # API Gateway port
REDIS_HOST=localhost      # Redis host (for rate limiting)
REDIS_PORT=6379          # Redis port
NODE_ID=<uuid>           # VPN node ID (for vpn-node service)
LOG_LEVEL=info           # Logging level
ENVIRONMENT=development   # Environment
```

## 🔐 Security Features Quick Access

### Kill Switch
```go
import "github.com/nikola43/aureo-vpn/internal/security"

killSwitch := security.NewKillSwitch("wg0")
killSwitch.Enable()                    // Activate
killSwitch.AllowVPNServer(ip, port, proto)  // Whitelist
killSwitch.Disable()                   // Deactivate
```

### DNS Protection
```go
dnsManager := security.NewDNSManager([]string{"1.1.1.1", "1.0.0.1"})
dnsManager.EnableDNSProtection()
dnsManager.DisableDNSProtection()
```

### WebRTC Protection
```go
webrtc := security.NewWebRTCProtection()
webrtc.Enable()
webrtc.Disable()
```

### Multi-Hop VPN
```go
multihop := security.NewMultiHopManager()
chain, err := multihop.CreateDoubleVPNChain(userID, "CH", "IS")
```

### Traffic Obfuscation
```go
obfuscation := security.NewObfuscationManager("stealth")
obfuscation.Enable()
```

## 📊 Monitoring

### Prometheus Metrics
```bash
# Access Prometheus
http://localhost:9090

# Access Grafana
http://localhost:3000
# Default: admin/admin

# View metrics
curl http://localhost:8080/metrics | grep aureo_vpn
```

### Key Metrics
```
aureo_vpn_active_connections       # Active VPN connections
aureo_vpn_node_load_score         # Node load (0-100)
aureo_vpn_http_requests_total     # API requests
aureo_vpn_data_transferred_bytes  # Data transfer
```

## 🧪 Testing

```bash
# Unit tests
make test-unit

# Integration tests
make test-integration

# All tests with coverage
make coverage

# Benchmarks
make bench

# Load testing
go test -bench=. -benchmem ./...
```

## 🐛 Troubleshooting

### Common Issues

**Port already in use**
```bash
# Find process using port 8080
lsof -i :8080
# Kill it
kill -9 <PID>
```

**Database connection failed**
```bash
# Test connection
psql -h localhost -U postgres -d aureo_vpn

# Reset database
make db-reset
```

**VPN node won't start**
```bash
# Check WireGuard
wg --version
sudo modprobe wireguard

# Check capabilities
sudo setcap cap_net_admin+ep ./bin/vpn-node
```

**Docker build fails**
```bash
# Clean Docker cache
docker system prune -a
docker-compose build --no-cache
```

## 📝 Configuration Examples

### WireGuard Client Config
```ini
[Interface]
PrivateKey = <client-private-key>
Address = 10.8.0.2/24
DNS = 1.1.1.1, 1.0.0.1

[Peer]
PublicKey = <server-public-key>
Endpoint = vpn.example.com:51820
AllowedIPs = 0.0.0.0/0, ::/0
PersistentKeepalive = 25
```

### Environment File (.env)
```bash
# Copy example and edit
cp .env.example .env
nano .env
```

### Docker Compose Override
```yaml
# docker-compose.override.yml
version: '3.8'
services:
  api-gateway:
    ports:
      - "8888:8080"
```

## 🔗 Useful Links

- **Documentation**: [docs/](docs/)
- **API Reference**: [docs/API.md](docs/API.md)
- **Architecture**: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
- **Deployment**: [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)
- **Features**: [docs/FEATURES.md](docs/FEATURES.md)

## 💡 Tips & Tricks

### Development Workflow
```bash
# Watch mode for API Gateway
make dev-api

# Run with race detector
go run -race cmd/api-gateway/main.go

# View SQL queries
DB_LOG_LEVEL=debug make run-api
```

### Performance Tuning
```bash
# Increase file descriptors
ulimit -n 65535

# Optimize PostgreSQL
# Add to postgresql.conf:
max_connections = 200
shared_buffers = 2GB
```

### Security Hardening
```bash
# Generate strong JWT secret
openssl rand -base64 32

# Create strong passwords
openssl rand -base64 16

# Check for vulnerabilities
make security-scan
```

## 🎓 Learning Path

1. **Start Here**: [README.md](README.md)
2. **Understand Architecture**: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
3. **Try API**: [docs/API.md](docs/API.md)
4. **Deploy**: [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)
5. **Explore Features**: [docs/FEATURES.md](docs/FEATURES.md)

## 📱 Quick API Test

```bash
# 1. Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!","username":"testuser"}'

# 2. Login & save token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!"}' | jq -r '.access_token')

# 3. Get nodes
curl -X GET http://localhost:8080/api/v1/nodes \
  -H "Authorization: Bearer $TOKEN"

# 4. Get profile
curl -X GET http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer $TOKEN"
```

## 🎯 Production Checklist

Before deploying to production:

- [ ] Change all default passwords
- [ ] Generate secure JWT secret
- [ ] Enable TLS/SSL
- [ ] Configure firewall rules
- [ ] Set up monitoring alerts
- [ ] Configure backups
- [ ] Enable rate limiting
- [ ] Test failover
- [ ] Document deployment
- [ ] Train operations team

## 📞 Support

- **Issues**: GitHub Issues
- **Email**: support@aureo-vpn.com
- **Documentation**: https://docs.aureo-vpn.com

---

**Pro Tip**: Star the repo and watch for updates! 🌟
