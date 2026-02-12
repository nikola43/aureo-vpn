# ⚙️ Configuration

All configuration is done via environment variables. Create a `.env` file in the project root or set variables directly.

---

## 🔑 Environment Variables

### DATABASE CONFIGURATION (SQLite)

```bash
DB_PATH=./data/aureo.db                # SQLite database file path
DB_MAX_CONNECTIONS=10                  # Max concurrent connections
```

{% hint style="info" %}
SQLite is the default embedded database. The file is created automatically on first startup. No external database server is required.
{% endhint %}

---

### JWT AUTHENTICATION

```bash
JWT_SECRET=your-32-byte-secret-key-here-change-in-prod
JWT_ACCESS_DURATION=15m                # Access token lifetime
JWT_REFRESH_DURATION=168h              # Refresh token lifetime (7 days)
```

Generate a strong secret:

```bash
openssl rand -base64 32
```

---

### SERVER CONFIGURATION

```bash
HOST=0.0.0.0
PORT=8080
READ_TIMEOUT=10s
WRITE_TIMEOUT=10s
IDLE_TIMEOUT=120s
SHUTDOWN_TIMEOUT=30s
```

---

### REDIS CONFIGURATION

```bash
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

{% hint style="warning" %}
Redis is optional. It is used for caching and rate limiting. The system works without it but rate limiting will be disabled.
{% endhint %}

---

### SECURITY

```bash
# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_MAX_REQUESTS=100            # Max requests per window
RATE_LIMIT_WINDOW_MINUTES=1            # Rate limit window in minutes

# CORS
CORS_ENABLED=true
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://your-domain.com

# Trusted Proxies
TRUSTED_PROXIES=127.0.0.1              # For reverse proxy setups
```

Rate limit tiers:

| Tier | Requests/min |
|------|-------------|
| Anonymous | 100 |
| Authenticated | 1,000 |
| Premium | 5,000 |

Rate limit headers included in all responses:

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 998
X-RateLimit-Reset: 1640995200
```

---

### TLS/HTTPS

```bash
TLS_ENABLED=false
TLS_CERT_PATH=/etc/ssl/certs/server.crt
TLS_KEY_PATH=/etc/ssl/private/server.key
```

{% hint style="danger" %}
Always enable TLS in production. Never send JWT tokens over unencrypted HTTP.
{% endhint %}

---

### LOGGING

```bash
LOG_LEVEL=info                         # debug, info, warn, error
LOG_FORMAT=json                        # json or text
```

---

### VPN NODE CONFIGURATION

```bash
NODE_ID=                               # Auto-generated if empty
NODE_NAME=us-east-1                    # Human-readable name
NODE_COUNTRY=US
NODE_CITY=New York
NODE_LATITUDE=40.7128
NODE_LONGITUDE=-74.0060

DEFAULT_PROTOCOL=wireguard             # wireguard or openvpn
WIREGUARD_PORT=51820
OPENVPN_PORT=1194
```

---

### VPN FEATURES

```bash
SESSION_TIMEOUT=24h
MAX_SESSIONS_PER_USER=5
DATA_TRANSFER_LIMIT_GB=0               # 0 = unlimited

ENABLE_KILL_SWITCH=true
ENABLE_DNS_PROTECTION=true
ENABLE_MULTIHOP=true
ENABLE_OBFUSCATION=false
ENABLE_SPLIT_TUNNELING=true
```

---

### BLOCKCHAIN CONFIGURATION

```bash
# Ethereum (Primary)
ETHEREUM_RPC_URL=https://mainnet.infura.io/v3/YOUR_PROJECT_ID
ETHEREUM_PRIVATE_KEY=0x...             # Hot wallet private key
ETHEREUM_CHAIN_ID=1                    # 1=mainnet, 11155111=sepolia

# Bitcoin (Optional)
BITCOIN_RPC_URL=http://localhost:8332
BITCOIN_RPC_USER=bitcoinrpc
BITCOIN_RPC_PASSWORD=your_password

# Litecoin (Optional)
LITECOIN_RPC_URL=http://localhost:9332
LITECOIN_RPC_USER=litecoinrpc
LITECOIN_RPC_PASSWORD=your_password
```

{% hint style="danger" %}
Never commit private keys to version control. Use a secrets manager (HashiCorp Vault, AWS Secrets Manager) in production.
{% endhint %}

---

### P2P NETWORK

```bash
P2P_ENABLED=true
P2P_PORT=4001
P2P_BOOTSTRAP_PEERS=                   # Comma-separated multiaddrs
P2P_MAX_PEERS=100
P2P_MAX_NODES=1000
```

---

### METRICS

```bash
METRICS_ENABLED=true
METRICS_PORT=9090
```

Prometheus metrics exposed:

```
aureo_vpn_active_connections
aureo_vpn_node_load_score
aureo_vpn_data_transferred_bytes
aureo_vpn_connection_duration_seconds
aureo_vpn_http_requests_total
```

---

## 💰 Reward Tiers Configuration

Node operators earn based on their tier:

| Tier | Rate/GB | Min Uptime | Min Reputation | Bonus |
|------|---------|------------|----------------|-------|
| 🥉 Bronze | $0.010 | 50% | 0 | 1.0x |
| 🥈 Silver | $0.015 | 80% | 60 | 1.2x |
| 🥇 Gold | $0.020 | 90% | 75 | 1.5x |
| 💎 Platinum | $0.030 | 95% | 90 | 2.0x |

### Earning Formula

```
earnings = bandwidth_gb × rate_per_gb × quality_multiplier × duration_bonus

quality_multiplier = 0.5 + (quality_score / 100)
duration_bonus = 1.1 (>60min) or 1.2 (>180min)
```

### Earning Example

```
Bandwidth: 50 GB
Tier: Gold ($0.02/GB)
Quality Score: 85%
Duration: 4 hours (1.2x bonus)

Earnings = 50 × $0.02 × (0.5 + 0.85) × 1.2 = $1.62
```

---

## 📄 Full `.env` Example

```bash
# =============================================================================
# DATABASE CONFIGURATION (SQLite)
# =============================================================================
DB_PATH=./data/aureo.db
DB_MAX_CONNECTIONS=10

# =============================================================================
# JWT AUTHENTICATION
# =============================================================================
JWT_SECRET=your-32-byte-secret-key-here-change-in-prod
JWT_ACCESS_DURATION=15m
JWT_REFRESH_DURATION=168h

# =============================================================================
# SERVER CONFIGURATION
# =============================================================================
HOST=0.0.0.0
PORT=8080
READ_TIMEOUT=10s
WRITE_TIMEOUT=10s
IDLE_TIMEOUT=120s
SHUTDOWN_TIMEOUT=30s

# =============================================================================
# REDIS CONFIGURATION
# =============================================================================
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# =============================================================================
# SECURITY
# =============================================================================
RATE_LIMIT_ENABLED=true
RATE_LIMIT_MAX_REQUESTS=100
RATE_LIMIT_WINDOW_MINUTES=1
CORS_ENABLED=true
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://your-domain.com
TRUSTED_PROXIES=127.0.0.1

# =============================================================================
# TLS/HTTPS (Production)
# =============================================================================
TLS_ENABLED=false
TLS_CERT_PATH=/etc/ssl/certs/server.crt
TLS_KEY_PATH=/etc/ssl/private/server.key

# =============================================================================
# LOGGING
# =============================================================================
LOG_LEVEL=info
LOG_FORMAT=json

# =============================================================================
# VPN NODE CONFIGURATION
# =============================================================================
NODE_ID=
NODE_NAME=us-east-1
NODE_COUNTRY=US
NODE_CITY=New York
NODE_LATITUDE=40.7128
NODE_LONGITUDE=-74.0060
DEFAULT_PROTOCOL=wireguard
WIREGUARD_PORT=51820
OPENVPN_PORT=1194

# =============================================================================
# VPN FEATURES
# =============================================================================
SESSION_TIMEOUT=24h
MAX_SESSIONS_PER_USER=5
DATA_TRANSFER_LIMIT_GB=0
ENABLE_KILL_SWITCH=true
ENABLE_DNS_PROTECTION=true
ENABLE_MULTIHOP=true
ENABLE_OBFUSCATION=false
ENABLE_SPLIT_TUNNELING=true

# =============================================================================
# BLOCKCHAIN CONFIGURATION
# =============================================================================
ETHEREUM_RPC_URL=https://mainnet.infura.io/v3/YOUR_PROJECT_ID
ETHEREUM_PRIVATE_KEY=0x...
ETHEREUM_CHAIN_ID=1
BITCOIN_RPC_URL=http://localhost:8332
BITCOIN_RPC_USER=bitcoinrpc
BITCOIN_RPC_PASSWORD=your_password
LITECOIN_RPC_URL=http://localhost:9332
LITECOIN_RPC_USER=litecoinrpc
LITECOIN_RPC_PASSWORD=your_password

# =============================================================================
# P2P NETWORK
# =============================================================================
P2P_ENABLED=true
P2P_PORT=4001
P2P_BOOTSTRAP_PEERS=
P2P_MAX_PEERS=100
P2P_MAX_NODES=1000

# =============================================================================
# METRICS
# =============================================================================
METRICS_ENABLED=true
METRICS_PORT=9090
```
