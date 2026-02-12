# 🐳 Deployment

This page covers environment variables, build commands, database migrations, and server startup sequence for all Aureo VPN components.

---

## Environment Variables

```bash
# ──────────────────────────────────────────────
# Server
# ──────────────────────────────────────────────
PORT=8080                           # HTTP port
HOST=0.0.0.0                       # Bind address
READ_TIMEOUT=10s                   # HTTP read timeout
WRITE_TIMEOUT=10s                  # HTTP write timeout
IDLE_TIMEOUT=120s                  # HTTP idle timeout
SHUTDOWN_TIMEOUT=30s               # Graceful shutdown timeout
BODY_LIMIT=10485760                # Max request body (10MB)
TLS_ENABLED=false                  # Enable HTTPS
TLS_CERT_FILE=                     # TLS certificate path
TLS_KEY_FILE=                      # TLS private key path

# ──────────────────────────────────────────────
# Database (PostgreSQL)
# ──────────────────────────────────────────────
DB_HOST=localhost                   # Database host
DB_PORT=5432                       # Database port
DB_USER=postgres                   # Database user
DB_PASSWORD=                       # Database password (required in production)
DB_NAME=aureo_vpn                  # Database name
DB_SSL_MODE=disable                # SSL mode (disable, require, verify-full)
DB_TIMEZONE=UTC                    # Timezone
DB_MAX_IDLE_CONNS=10               # Max idle connections
DB_MAX_OPEN_CONNS=100              # Max open connections
DB_CONN_MAX_LIFETIME=1h            # Connection max lifetime
DB_LOG_LEVEL=warn                  # GORM log level (silent, error, warn, info)

# ──────────────────────────────────────────────
# JWT
# ──────────────────────────────────────────────
JWT_SECRET=                        # HMAC secret (min 32 chars in production)
JWT_ACCESS_DURATION=15m            # Access token lifetime
JWT_REFRESH_DURATION=168h          # Refresh token lifetime (7 days)
JWT_ISSUER=aureo-vpn               # Token issuer

# ──────────────────────────────────────────────
# Blockchain (Ethereum)
# ──────────────────────────────────────────────
ETHEREUM_RPC_URL=                  # Ethereum JSON-RPC endpoint
ETHEREUM_PRIVATE_KEY=              # Payout wallet private key
ETHEREUM_CHAIN_ID=1                # Chain ID (1=mainnet, 5=goerli, 11155111=sepolia)

# ──────────────────────────────────────────────
# Blockchain (Bitcoin)
# ──────────────────────────────────────────────
BITCOIN_RPC_URL=                   # Bitcoin Core RPC URL
BITCOIN_RPC_USER=                  # RPC username
BITCOIN_RPC_PASSWORD=              # RPC password

# ──────────────────────────────────────────────
# Blockchain (Litecoin)
# ──────────────────────────────────────────────
LITECOIN_RPC_URL=                  # Litecoin Core RPC URL
LITECOIN_RPC_USER=                 # RPC username
LITECOIN_RPC_PASSWORD=             # RPC password

# ──────────────────────────────────────────────
# Security
# ──────────────────────────────────────────────
CORS_ENABLED=true                  # Enable CORS middleware
CORS_ALLOWED_ORIGINS=              # Comma-separated origins (required in production)
CORS_ALLOW_CREDENTIALS=false       # Allow credentials
CORS_MAX_AGE=3600                  # Preflight cache seconds
RATE_LIMIT_ENABLED=true            # Enable rate limiting
RATE_LIMIT_MAX_REQUESTS=100        # Max requests per window
RATE_LIMIT_WINDOW=1m               # Rate limit window
TRUSTED_PROXIES=                   # Comma-separated trusted proxy IPs
PASSWORD_MIN_LENGTH=8              # Minimum password length
MAX_LOGIN_ATTEMPTS=5               # Before lockout
LOCKOUT_DURATION=15m               # Lockout duration

# ──────────────────────────────────────────────
# Metrics
# ──────────────────────────────────────────────
METRICS_ENABLED=true               # Enable Prometheus metrics
METRICS_PORT=9090                  # Metrics port
METRICS_PATH=/metrics              # Scrape endpoint

# ──────────────────────────────────────────────
# Logging
# ──────────────────────────────────────────────
LOG_LEVEL=info                     # Log level (debug, info, warn, error)
LOG_FORMAT=json                    # Log format (json, text)
LOG_ADD_SOURCE=true                # Include source file in logs
ENVIRONMENT=development            # Environment (development, production)
SERVICE_NAME=aureo-vpn             # Service name in logs

# ──────────────────────────────────────────────
# VPN Node
# ──────────────────────────────────────────────
NODE_ID=                           # UUID for this node
DEFAULT_PROTOCOL=wireguard         # Default VPN protocol
SESSION_TIMEOUT=24h                # Max session duration
MAX_SESSIONS_PER_USER=5            # Concurrent sessions per user
DATA_TRANSFER_LIMIT_GB=0           # 0 = unlimited
ENABLE_KILL_SWITCH=true            # Kill switch default
ENABLE_DNS_PROTECTION=true         # DNS leak protection default
ENABLE_MULTIHOP=true               # Multi-hop support
ENABLE_OBFUSCATION=true            # Traffic obfuscation support
```

### Production Requirements

The following are enforced when `ENVIRONMENT=production`:

- `JWT_SECRET` must be set (minimum 32 characters)
- `DB_PASSWORD` must be set
- `CORS_ALLOWED_ORIGINS` must be set
- If `TLS_ENABLED=true`, both `TLS_CERT_FILE` and `TLS_KEY_FILE` must be set

---

## Build Commands

### Backend (aureo-vpn)

```bash
cd aureo-vpn

make build              # Build all 4 binaries (api-gateway, control-server, vpn-node, cli)
make build-api          # Build API Gateway only → bin/api-gateway
make build-control      # Build Control Server only → bin/control-server
make build-node         # Build VPN Node only → bin/vpn-node
make build-cli          # Build CLI only → bin/aureo (with git version tag)
make install            # Copy all binaries to /usr/local/bin (sudo)
make install-cli        # Install CLI only

# Docker
make docker-build       # Build Docker images for all services
make docker-up          # Start with Docker Compose
make docker-down        # Stop Docker Compose
make docker-logs        # Tail Docker Compose logs

# Run
make run-api            # Build and run API Gateway
make run-control        # Build and run Control Server
make run-node           # Build and run VPN Node (sudo)
make dev-api            # Run API Gateway with go run (dev mode)
make dev-control        # Run Control Server with go run
make dev-node           # Run VPN Node with go run (sudo)
```

### Mobile App (aureo-app)

```bash
cd aureo-app

npm start               # Start Expo dev server
npm run ios             # Run on iOS simulator
npm run android         # Run on Android emulator
npm run lint            # ESLint
```

### Desktop Client (aureo-desktop)

```bash
cd aureo-desktop

make dev                # Dev mode with hot reload
make build              # Build for current platform
make build-all          # Build for macOS, Windows, Linux
make generate           # Generate Wails bindings
```

---

## Database Migration Order

GORM `AutoMigrate` handles table creation. The order matters due to foreign key constraints:

| Order | Models | Dependencies |
|---|---|---|
| 1 | `User`, `NodeReward` | None |
| 2 | `NodeOperator` | `User` (via `UserID`) |
| 3 | `VPNNode` | `NodeOperator` (via `OperatorID`, nullable) |
| 4 | `Session`, `Config` | `User` (via `UserID`), `VPNNode` (via `NodeID`) |
| 5 | `OperatorEarning`, `OperatorPayout`, `NodePerformanceMetric` | `NodeOperator`, `VPNNode`, `Session` |

For VPN nodes with local SQLite, a separate migration runs `AllLocalModels()`:

```go
// LocalUser, LocalSession, LocalNodeConfig, NodeIdentity
database.AutoMigrateLocal()
```

---

## Server Startup Sequence

The API Gateway follows this startup sequence (`cmd/api-gateway/main.go`):

```
 1. Load configuration from environment variables
 2. Initialize structured logger (JSON format, with source)
 3. Connect to database with retry logic
    └── Up to 5 attempts with exponential backoff (2s, 4s, 8s, 16s, 32s)
 4. Run database AutoMigrate
 5. Initialize auth service (TokenService + AuthService)
 6. Initialize blockchain service (optional, warns if unconfigured)
 7. Initialize reward service + seed reward tiers
 8. Initialize operator service
 9. Initialize API handlers (with dependency injection)
10. Create Fiber app with production config
    ├── Set up global middleware stack
    │   (Recover → RequestID → Compress → Logger → CORS → Metrics → RateLimit)
    ├── Register health/readiness endpoints
    ├── Register API v1 routes
    │   ├── /api/v1/auth/*        (public)
    │   ├── /api/v1/user/*        (authenticated)
    │   ├── /api/v1/nodes/*       (authenticated)
    │   ├── /api/v1/sessions/*    (authenticated)
    │   ├── /api/v1/config/*      (authenticated)
    │   ├── /api/v1/operator/*    (authenticated)
    │   └── /api/v1/admin/*       (admin only)
    └── Register 404 handler
11. Start HTTP server in goroutine
    └── TLS if configured, plain HTTP otherwise
12. Wait for shutdown signal (SIGINT, SIGTERM)
    ├── Create shutdown context (30s timeout)
    ├── Attempt graceful shutdown
    ├── Close database connection
    └── Log "server stopped gracefully"
```
