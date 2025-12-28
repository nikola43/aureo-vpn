# Aureo VPN

A decentralized VPN platform with cryptocurrency rewards for node operators. Built with Go, featuring WireGuard and OpenVPN protocol support, P2P node discovery, and blockchain-based payment system.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Features](#features)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Configuration](#configuration)
- [How the System Works](#how-the-system-works)
- [Running Your Own Node](#running-your-own-node)
- [API Reference](#api-reference)
- [Docker Deployment](#docker-deployment)
- [Security](#security)
- [Development](#development)
- [Testing](#testing)
- [Troubleshooting](#troubleshooting)

---

## Overview

Aureo VPN is a decentralized VPN service that allows anyone to run VPN nodes and earn cryptocurrency rewards. The platform consists of:

- **API Gateway**: REST API server handling user authentication, node discovery, and session management
- **Control Server**: Orchestrates nodes, manages the network, and processes payments
- **VPN Nodes**: Client-facing servers running WireGuard/OpenVPN
- **CLI Tool**: Command-line interface for administration
- **P2P Network**: Decentralized node discovery using libp2p

### Key Highlights

- **Multi-Protocol**: Support for both WireGuard (modern, fast) and OpenVPN (compatible)
- **Decentralized**: P2P network for node discovery without single point of failure
- **Crypto Rewards**: Node operators earn ETH, BTC, or LTC based on bandwidth served
- **Reputation System**: Quality-based rewards with tier progression
- **Privacy-Focused**: Kill switch, DNS leak protection, multi-hop routing

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              AUREO VPN PLATFORM                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐                    │
│  │   Desktop   │     │   Mobile    │     │    Web      │                    │
│  │   Client    │     │   Client    │     │  Dashboard  │                    │
│  └──────┬──────┘     └──────┬──────┘     └──────┬──────┘                    │
│         │                   │                   │                           │
│         └───────────────────┼───────────────────┘                           │
│                             │                                               │
│                             ▼                                               │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                         API GATEWAY                                   │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐              │   │
│  │  │   Auth   │  │  Nodes   │  │ Sessions │  │ Operator │              │   │
│  │  │ Handler  │  │ Handler  │  │ Handler  │  │ Handler  │              │   │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘              │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                             │                                               │
│              ┌──────────────┼──────────────┐                                │
│              ▼              ▼              ▼                                │
│  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐                   │
│  │    SQLite     │  │    Redis      │  │ Control Server│                   │
│  │   Database    │  │    Cache      │  │               │                   │
│  └───────────────┘  └───────────────┘  └───────┬───────┘                   │
│                                                 │                           │
│                                                 ▼                           │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                       P2P NETWORK (libp2p)                           │   │
│  │  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐   │   │
│  │  │   Kademlia DHT  │    │    Gossipsub    │    │      mDNS       │   │   │
│  │  │  (Peer Routing) │    │   (Pub/Sub)     │    │  (Local Disc.)  │   │   │
│  │  └─────────────────┘    └─────────────────┘    └─────────────────┘   │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                         VPN NODES                                     │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │   │
│  │  │  Node US-1  │  │  Node EU-1  │  │  Node AP-1  │  │  Node ...   │  │   │
│  │  │  WireGuard  │  │  OpenVPN    │  │  WireGuard  │  │             │  │   │
│  │  │  51820/UDP  │  │  1194/UDP   │  │  51820/UDP  │  │             │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘  │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                      BLOCKCHAIN LAYER                                 │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                   │   │
│  │  │  Ethereum   │  │   Bitcoin   │  │  Litecoin   │                   │   │
│  │  │   Payouts   │  │   Payouts   │  │   Payouts   │                   │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘                   │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Component Overview

| Component | Description | Port |
|-----------|-------------|------|
| API Gateway | REST API server (Fiber) | 8080 |
| Control Server | Network orchestrator | 8081 |
| VPN Node | WireGuard/OpenVPN server | 51820/UDP, 1194/UDP |
| SQLite | Embedded database | - (file-based) |
| Redis | Caching & rate limiting | 6379 |
| Prometheus | Metrics collection | 9090 |
| Grafana | Metrics visualization | 3000 |

---

## Features

### VPN Features

- **WireGuard Protocol**: Modern, fast, secure VPN protocol with Curve25519 encryption
- **OpenVPN Protocol**: Legacy protocol support for maximum compatibility
- **Kill Switch**: Prevents traffic leaks when VPN disconnects
- **DNS Leak Protection**: Routes DNS queries through encrypted tunnel
- **Split Tunneling**: Route only specific traffic through VPN
- **Multi-Hop**: Chain through multiple nodes for enhanced privacy
- **Obfuscation**: Disguise VPN traffic as regular HTTPS

### Platform Features

- **Decentralized Network**: P2P node discovery without central servers
- **Crypto Payments**: Earn ETH, BTC, or LTC for running nodes
- **Reputation System**: Quality-based rewards (Bronze → Platinum tiers)
- **Auto Load Balancing**: Smart node selection based on load score
- **Real-time Metrics**: Prometheus + Grafana monitoring
- **Operator Dashboard**: Web interface for node operators

### Security Features

- **JWT Authentication**: Secure token-based auth with refresh tokens
- **Rate Limiting**: Protection against abuse
- **CORS Protection**: Configurable cross-origin security
- **TLS Support**: HTTPS encryption for API
- **Password Hashing**: bcrypt with secure defaults
- **Encrypted Storage**: Private keys encrypted at rest

---

## Quick Start

### Prerequisites

- Go 1.24+
- Docker & Docker Compose (optional)
- Redis 7+ (optional, for caching)
- WireGuard tools (for VPN nodes)

### Using Docker (Recommended)

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

### Manual Setup

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

> **Note**: SQLite database file is created automatically at startup. No external database setup required.

---

## Installation

### Building from Source

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
- `bin/api-gateway` - API server
- `bin/control-server` - Control plane
- `bin/vpn-node` - VPN server
- `bin/cli` - Management CLI

### System Requirements

**Minimum (Development)**:
- 2 CPU cores
- 2GB RAM
- 10GB disk

**Recommended (Production)**:
- 4+ CPU cores
- 8GB+ RAM
- 50GB+ SSD
- 1Gbps network

> **Note**: SQLite is used as the embedded database, eliminating the need for external database servers.

---

## Configuration

### Environment Variables

Create a `.env` file or set environment variables:

```bash
# =============================================================================
# DATABASE CONFIGURATION (SQLite)
# =============================================================================
DB_PATH=./data/aureo.db                # SQLite database file path
DB_MAX_CONNECTIONS=10                  # Max concurrent connections

# =============================================================================
# JWT AUTHENTICATION
# =============================================================================
JWT_SECRET=your-32-byte-secret-key-here-change-in-prod
JWT_ACCESS_DURATION=15m                # Access token lifetime
JWT_REFRESH_DURATION=168h              # Refresh token lifetime (7 days)

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
RATE_LIMIT_MAX_REQUESTS=100            # Max requests per window
RATE_LIMIT_WINDOW_MINUTES=1            # Rate limit window

CORS_ENABLED=true
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://your-domain.com

TRUSTED_PROXIES=127.0.0.1              # For reverse proxy setups

# =============================================================================
# TLS/HTTPS (Production)
# =============================================================================
TLS_ENABLED=false
TLS_CERT_PATH=/etc/ssl/certs/server.crt
TLS_KEY_PATH=/etc/ssl/private/server.key

# =============================================================================
# LOGGING
# =============================================================================
LOG_LEVEL=info                         # debug, info, warn, error
LOG_FORMAT=json                        # json or text

# =============================================================================
# VPN NODE CONFIGURATION
# =============================================================================
NODE_ID=                               # Auto-generated if empty
NODE_NAME=us-east-1                    # Human-readable name
NODE_COUNTRY=US
NODE_CITY=New York
NODE_LATITUDE=40.7128
NODE_LONGITUDE=-74.0060

DEFAULT_PROTOCOL=wireguard             # wireguard or openvpn
WIREGUARD_PORT=51820
OPENVPN_PORT=1194

# =============================================================================
# VPN FEATURES
# =============================================================================
SESSION_TIMEOUT=24h
MAX_SESSIONS_PER_USER=5
DATA_TRANSFER_LIMIT_GB=0               # 0 = unlimited

ENABLE_KILL_SWITCH=true
ENABLE_DNS_PROTECTION=true
ENABLE_MULTIHOP=true
ENABLE_OBFUSCATION=false
ENABLE_SPLIT_TUNNELING=true

# =============================================================================
# BLOCKCHAIN CONFIGURATION
# =============================================================================
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

# =============================================================================
# P2P NETWORK
# =============================================================================
P2P_ENABLED=true
P2P_PORT=4001
P2P_BOOTSTRAP_PEERS=                   # Comma-separated multiaddrs
P2P_MAX_PEERS=100
P2P_MAX_NODES=1000

# =============================================================================
# METRICS
# =============================================================================
METRICS_ENABLED=true
METRICS_PORT=9090
```

### Reward Tiers Configuration

Node operators earn based on their tier:

| Tier | Rate/GB | Min Uptime | Min Reputation | Bonus |
|------|---------|------------|----------------|-------|
| Bronze | $0.010 | 50% | 0 | 1.0x |
| Silver | $0.015 | 80% | 60 | 1.2x |
| Gold | $0.020 | 90% | 75 | 1.5x |
| Platinum | $0.030 | 95% | 90 | 2.0x |

**Earning Formula**:
```
earnings = bandwidth_gb × rate_per_gb × quality_multiplier × duration_bonus

quality_multiplier = 0.5 + (quality_score / 100)
duration_bonus = 1.1 (>60min) or 1.2 (>180min)
```

---

## How the System Works

### User Flow: Registration to VPN Connection

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           USER JOURNEY                                   │
└─────────────────────────────────────────────────────────────────────────┘

1. REGISTRATION
   ┌──────────┐      POST /api/v1/auth/register       ┌──────────────┐
   │  Client  │ ──────────────────────────────────────▶│ API Gateway  │
   │          │                                        │              │
   │          │◀────────────────────────────────────── │  • Validate  │
   │          │       { access_token, refresh_token }  │  • Hash pwd  │
   └──────────┘                                        │  • Create    │
                                                       └──────────────┘

2. GET AVAILABLE NODES
   ┌──────────┐      GET /api/v1/nodes?country=US     ┌──────────────┐
   │  Client  │ ──────────────────────────────────────▶│ API Gateway  │
   │          │                                        │              │
   │          │◀────────────────────────────────────── │  • Query DB  │
   │          │       [{ node_id, ip, port, load }]    │  • Query P2P │
   └──────────┘                                        │  • Sort by   │
                                                       │    load      │
                                                       └──────────────┘

3. CREATE VPN SESSION
   ┌──────────┐      POST /api/v1/sessions            ┌──────────────┐
   │  Client  │ ──────────────────────────────────────▶│ API Gateway  │
   │          │       { node_id, protocol }            │              │
   │          │                                        │  • Allocate  │
   │          │◀────────────────────────────────────── │    tunnel IP │
   │          │       { session_id, tunnel_ip,         │  • Generate  │
   │          │         server_public_key, config }    │    keypair   │
   └──────────┘                                        │  • Add peer  │
                                                       └──────┬───────┘
                                                              │
                                                              ▼
                                                       ┌──────────────┐
                                                       │   VPN Node   │
                                                       │              │
                                                       │  • Add peer  │
                                                       │    to WG     │
                                                       │  • Start     │
                                                       │    monitor   │
                                                       └──────────────┘

4. ESTABLISH VPN TUNNEL
   ┌──────────┐         WireGuard Handshake           ┌──────────────┐
   │  Client  │ ══════════════════════════════════════▶│   VPN Node   │
   │          │         UDP:51820                      │              │
   │          │◀══════════════════════════════════════ │  Encrypted   │
   │          │         Encrypted Tunnel               │  Tunnel      │
   └──────────┘                                        └──────────────┘

5. SESSION MONITORING (Background)
                                                       ┌──────────────┐
                                                       │   VPN Node   │
                                                       │              │
   Every 30s:                                          │ • Track      │
   ┌─────────────────────────────────────────────────▶│   bytes      │
   │                                                   │ • Calculate  │
   │                                                   │   bandwidth  │
   │                                                   │ • Record     │
   │   Every 15min: Flush earnings                     │   earnings   │
   │   ┌─────────────────────────────────────────────▶│              │
   │   │                                               └──────┬───────┘
   │   │                                                      │
   │   │                                                      ▼
   │   │                                               ┌──────────────┐
   │   │                                               │   Database   │
   │   │                                               │              │
   └───┴──────────────────────────────────────────────▶│ • Sessions   │
                                                       │ • Earnings   │
                                                       │ • Operators  │
                                                       └──────────────┘

6. DISCONNECT
   ┌──────────┐      DELETE /api/v1/sessions/:id      ┌──────────────┐
   │  Client  │ ──────────────────────────────────────▶│ API Gateway  │
   │          │                                        │              │
   │          │◀────────────────────────────────────── │  • Remove    │
   │          │       { bytes_sent, bytes_recv,        │    peer      │
   └──────────┘         duration, earnings }           │  • Finalize  │
                                                       │    earnings  │
                                                       └──────────────┘
```

### VPN Protocol Details

#### WireGuard Implementation

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        WIREGUARD TUNNEL                                  │
└─────────────────────────────────────────────────────────────────────────┘

Key Generation (Curve25519):
  Private Key: 32 random bytes with bit clamping
  Public Key:  X25519(private_key, basepoint)
  Pre-shared:  32 random bytes (optional extra security)

Connection Flow:
  1. Client generates keypair
  2. Server generates keypair
  3. Exchange public keys via API
  4. Server adds client as peer with allocated IP
  5. Client configures local interface
  6. Noise Protocol handshake (IKpsk2)
  7. ChaCha20-Poly1305 encrypted tunnel

Config Example (Client):
┌────────────────────────────────────────┐
│ [Interface]                            │
│ PrivateKey = <client_private_key>      │
│ Address = 10.8.0.2/32                  │
│ DNS = 1.1.1.1, 8.8.8.8                │
│                                        │
│ [Peer]                                 │
│ PublicKey = <server_public_key>        │
│ Endpoint = vpn.example.com:51820       │
│ AllowedIPs = 0.0.0.0/0                │
│ PersistentKeepalive = 25               │
└────────────────────────────────────────┘
```

#### OpenVPN Implementation

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        OPENVPN TUNNEL                                    │
└─────────────────────────────────────────────────────────────────────────┘

Certificate Generation:
  • RSA 2048-bit keys
  • Self-signed CA certificate
  • Server/Client certificates signed by CA
  • 1-year validity

Connection Flow:
  1. TLS handshake with certificate validation
  2. Key exchange and session key generation
  3. UDP tunnel with AES-256-GCM encryption
  4. HMAC-SHA256 authentication

Config Example (Client):
┌────────────────────────────────────────┐
│ client                                 │
│ dev tun                                │
│ proto udp                              │
│ remote vpn.example.com 1194            │
│ resolv-retry infinite                  │
│ nobind                                 │
│ persist-key                            │
│ persist-tun                            │
│ cipher AES-256-GCM                     │
│ auth SHA256                            │
│ <ca>                                   │
│ -----BEGIN CERTIFICATE-----            │
│ ...                                    │
│ </ca>                                  │
└────────────────────────────────────────┘
```

### P2P Network Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                      P2P NETWORK (libp2p)                                │
└─────────────────────────────────────────────────────────────────────────┘

Discovery Mechanisms:
  • Kademlia DHT: Distributed hash table for peer routing
  • Gossipsub: Pub/sub messaging for announcements
  • mDNS: Local network discovery

Topics:
  /aureo/nodes/announce/1.0.0   - New node announcements
  /aureo/nodes/heartbeat/1.0.0  - Status updates (every 30s)

NodeInfo Structure:
┌────────────────────────────────────────┐
│ {                                      │
│   "node_id": "uuid",                   │
│   "peer_id": "libp2p-peer-id",        │
│   "addresses": ["/ip4/.../tcp/4001"],  │
│   "country": "US",                     │
│   "city": "New York",                  │
│   "latitude": 40.7128,                 │
│   "longitude": -74.0060,               │
│   "max_connections": 1000,             │
│   "load_score": 25,                    │
│   "bandwidth_mbps": 1000,              │
│   "status": "online",                  │
│   "signature": "..."                   │
│ }                                      │
└────────────────────────────────────────┘

Message Flow:
  ┌─────────┐     announce      ┌─────────┐     announce      ┌─────────┐
  │ Node A  │ ────────────────▶ │   DHT   │ ────────────────▶ │ Node B  │
  └─────────┘                   └─────────┘                   └─────────┘
       │                                                           │
       │                     heartbeat (30s)                       │
       └───────────────────────────────────────────────────────────┘
```

### Blockchain Payment Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                      PAYMENT FLOW                                        │
└─────────────────────────────────────────────────────────────────────────┘

Earning Calculation:
  ┌──────────────────────────────────────────────────────────────────────┐
  │                                                                      │
  │  Session Ends → Calculate Bandwidth → Apply Tier Rate → Store       │
  │                                                                      │
  │  Example:                                                            │
  │    Bandwidth: 50 GB                                                  │
  │    Tier: Gold ($0.02/GB)                                            │
  │    Quality Score: 85%                                               │
  │    Duration: 4 hours (1.2x bonus)                                   │
  │                                                                      │
  │    Earnings = 50 × $0.02 × (0.5 + 0.85) × 1.2 = $1.62               │
  │                                                                      │
  └──────────────────────────────────────────────────────────────────────┘

Payout Process:
  ┌─────────────┐                          ┌─────────────┐
  │  Operator   │ ──── Request Payout ───▶ │    API      │
  │             │                          │             │
  │  Wallet:    │                          │  • Verify   │
  │  0x1234...  │                          │    balance  │
  └─────────────┘                          │  • Check    │
                                           │    minimum  │
                                           └──────┬──────┘
                                                  │
                                                  ▼
                                           ┌─────────────┐
                                           │ Blockchain  │
                                           │   Service   │
                                           │             │
                                           │ • Create TX │
                                           │ • Sign      │
                                           │ • Broadcast │
                                           └──────┬──────┘
                                                  │
                     ┌────────────────────────────┼────────────────────────────┐
                     ▼                            ▼                            ▼
              ┌─────────────┐              ┌─────────────┐              ┌─────────────┐
              │  Ethereum   │              │   Bitcoin   │              │  Litecoin   │
              │   Network   │              │   Network   │              │   Network   │
              └─────────────┘              └─────────────┘              └─────────────┘
```

---

## Running Your Own Node

### Overview

Running a VPN node allows you to:
- Earn cryptocurrency rewards for bandwidth served
- Contribute to the decentralized network
- Help users access private, secure internet

### Requirements

**Hardware**:
- 2+ CPU cores
- 4GB+ RAM
- 100GB+ SSD
- 100Mbps+ network (1Gbps recommended)
- Static IP address

**Software**:
- Linux (Ubuntu 22.04 recommended)
- Docker & Docker Compose
- WireGuard kernel module

**Network**:
- Open port 51820/UDP (WireGuard)
- Open port 1194/UDP (OpenVPN, optional)
- Open port 4001/TCP (P2P, optional)

### Step 1: Register as Operator

First, create an account and register as a node operator:

```bash
# Register user account
curl -X POST https://api.aureo-vpn.com/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "operator@example.com",
    "username": "my_node_operator",
    "password": "secure_password_123"
  }'

# Save the access token from response
export TOKEN="<access_token_from_response>"

# Register as operator with your crypto wallet
curl -X POST https://api.aureo-vpn.com/api/v1/operator/register \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "wallet_address": "0xYourEthereumWalletAddress",
    "wallet_currency": "ethereum",
    "company_name": "My VPN Nodes LLC",
    "contact_email": "operator@example.com"
  }'
```

Wait for admin verification (you'll receive an email).

### Step 2: Create Your Node

After verification, create a node entry:

```bash
curl -X POST https://api.aureo-vpn.com/api/v1/operator/nodes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "us-east-node-1",
    "hostname": "vpn-us.example.com",
    "ip_address": "YOUR_SERVER_IP",
    "country": "US",
    "city": "New York",
    "latitude": 40.7128,
    "longitude": -74.0060,
    "wireguard_port": 51820,
    "openvpn_port": 1194,
    "max_connections": 500
  }'

# Response includes:
# - node_id
# - private_key (save this securely!)
# - public_key
# - subnet (e.g., 10.8.0.0/24)
```

**Save the `node_id` and `private_key` from the response!**

### Step 3: Server Setup

SSH into your server and install dependencies:

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install WireGuard
sudo apt install -y wireguard wireguard-tools

# Install Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Enable IP forwarding
echo "net.ipv4.ip_forward = 1" | sudo tee /etc/sysctl.d/99-wireguard.conf
echo "net.ipv6.conf.all.forwarding = 1" | sudo tee -a /etc/sysctl.d/99-wireguard.conf
sudo sysctl -p /etc/sysctl.d/99-wireguard.conf
```

### Step 4: Configure the Node

Create configuration directory:

```bash
sudo mkdir -p /etc/aureo-vpn
sudo chmod 700 /etc/aureo-vpn
```

Create environment file `/etc/aureo-vpn/.env`:

```bash
# Node Identity (from Step 2)
NODE_ID=your-node-uuid-here
NODE_PRIVATE_KEY=your-wireguard-private-key

# API Connection
API_URL=https://api.aureo-vpn.com
API_TOKEN=your-operator-access-token

# VPN Configuration
WIREGUARD_PORT=51820
WIREGUARD_INTERFACE=wg0
VPN_SUBNET=10.8.0.0/24
VPN_GATEWAY=10.8.0.1

# Server Info
NODE_NAME=us-east-node-1
NODE_COUNTRY=US
NODE_CITY=New York

# P2P Network (optional)
P2P_ENABLED=true
P2P_PORT=4001

# Logging
LOG_LEVEL=info
```

### Step 5: Run with Docker

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  vpn-node:
    image: ghcr.io/nikola43/aureo-vpn-node:latest
    container_name: aureo-vpn-node
    restart: unless-stopped
    env_file:
      - /etc/aureo-vpn/.env
    cap_add:
      - NET_ADMIN
      - SYS_MODULE
    sysctls:
      - net.ipv4.ip_forward=1
      - net.ipv4.conf.all.src_valid_mark=1
    volumes:
      - /etc/aureo-vpn:/etc/aureo-vpn:ro
      - /lib/modules:/lib/modules:ro
    devices:
      - /dev/net/tun:/dev/net/tun
    ports:
      - "51820:51820/udp"
      - "1194:1194/udp"
      - "4001:4001/tcp"
    networks:
      - vpn-network
    healthcheck:
      test: ["CMD", "wg", "show", "wg0"]
      interval: 30s
      timeout: 10s
      retries: 3

networks:
  vpn-network:
    driver: bridge
```

Start the node:

```bash
docker-compose up -d

# Check logs
docker-compose logs -f

# Verify WireGuard interface
docker exec aureo-vpn-node wg show
```

### Step 6: Manual Setup (Alternative)

If you prefer running without Docker:

```bash
# Clone repository
git clone https://github.com/nikola43/aureo-vpn.git
cd aureo-vpn

# Build VPN node binary
make build-node

# Configure WireGuard interface
sudo ip link add dev wg0 type wireguard
sudo wg set wg0 private-key /etc/aureo-vpn/privatekey listen-port 51820
sudo ip addr add 10.8.0.1/24 dev wg0
sudo ip link set wg0 up

# Configure firewall (iptables)
sudo iptables -A FORWARD -i wg0 -j ACCEPT
sudo iptables -A FORWARD -o wg0 -j ACCEPT
sudo iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE

# Run the node
sudo ./bin/vpn-node
```

### Step 7: Configure Firewall

```bash
# UFW (Ubuntu)
sudo ufw allow 51820/udp comment "WireGuard VPN"
sudo ufw allow 1194/udp comment "OpenVPN"
sudo ufw allow 4001/tcp comment "P2P Network"

# Or iptables
sudo iptables -A INPUT -p udp --dport 51820 -j ACCEPT
sudo iptables -A INPUT -p udp --dport 1194 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 4001 -j ACCEPT
```

### Step 8: Verify Node Status

```bash
# Check node is visible in network
curl -X GET https://api.aureo-vpn.com/api/v1/nodes \
  -H "Authorization: Bearer $TOKEN" | jq '.[] | select(.id == "YOUR_NODE_ID")'

# Response should show:
# {
#   "id": "your-node-id",
#   "status": "online",
#   "load_score": 0,
#   "connections": 0,
#   ...
# }
```

### Step 9: Monitor Your Node

```bash
# View operator dashboard
curl -X GET https://api.aureo-vpn.com/api/v1/operator/dashboard \
  -H "Authorization: Bearer $TOKEN"

# Response:
# {
#   "total_earned_usd": 45.67,
#   "pending_payout_usd": 12.34,
#   "active_nodes": 1,
#   "total_bandwidth_gb": 2345.67,
#   "average_uptime_percent": 99.5,
#   "reputation_score": 85,
#   "current_tier": "gold"
# }

# View earnings history
curl -X GET https://api.aureo-vpn.com/api/v1/operator/earnings \
  -H "Authorization: Bearer $TOKEN"
```

### Step 10: Request Payout

When your pending balance exceeds $10:

```bash
curl -X POST https://api.aureo-vpn.com/api/v1/operator/payout/request \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json"

# Response:
# {
#   "payout_id": "uuid",
#   "amount_usd": 45.67,
#   "amount_crypto": "0.0228",
#   "currency": "ethereum",
#   "status": "pending",
#   "estimated_completion": "2024-01-15T12:00:00Z"
# }
```

### Node Maintenance

```bash
# Update node software
docker-compose pull
docker-compose up -d

# View real-time stats
docker exec aureo-vpn-node wg show wg0

# Check peer connections
docker exec aureo-vpn-node wg show wg0 dump | wc -l

# Restart node
docker-compose restart

# Stop node gracefully
docker-compose down
```

### Troubleshooting Node Issues

**Node shows offline**:
```bash
# Check if process is running
docker ps | grep aureo-vpn-node

# Check logs for errors
docker-compose logs --tail=100

# Verify API connectivity
curl -I https://api.aureo-vpn.com/health
```

**No connections received**:
```bash
# Verify port is open
sudo netstat -ulnp | grep 51820

# Test from external host
nc -zvu your-server-ip 51820

# Check firewall
sudo iptables -L -n | grep 51820
```

**WireGuard interface issues**:
```bash
# Recreate interface
docker-compose down
sudo ip link delete wg0 2>/dev/null || true
docker-compose up -d
```

---

## API Reference

### Authentication

#### Register User
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "username": "johndoe",
  "password": "SecurePassword123"
}

Response 201:
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "johndoe",
    "subscription_tier": "free"
  },
  "access_token": "eyJhbG...",
  "refresh_token": "eyJhbG...",
  "expires_in": 900
}
```

#### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePassword123"
}

Response 200:
{
  "user": { ... },
  "access_token": "eyJhbG...",
  "refresh_token": "eyJhbG...",
  "expires_in": 900
}
```

#### Refresh Token
```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbG..."
}

Response 200:
{
  "access_token": "eyJhbG...",
  "refresh_token": "eyJhbG...",
  "expires_in": 900
}
```

### Nodes

#### List Nodes
```http
GET /api/v1/nodes?country=US&protocol=wireguard&limit=10
Authorization: Bearer <token>

Response 200:
{
  "nodes": [
    {
      "id": "uuid",
      "name": "us-east-1",
      "hostname": "us1.vpn.aureo.io",
      "ip_address": "203.0.113.1",
      "country": "US",
      "city": "New York",
      "load_score": 25,
      "wireguard_port": 51820,
      "wireguard_public_key": "base64...",
      "status": "online",
      "latency_ms": 15,
      "features": ["multihop", "dns_protection"]
    }
  ],
  "total": 1
}
```

#### Get Best Node
```http
GET /api/v1/nodes/best?country=US&protocol=wireguard
Authorization: Bearer <token>

Response 200:
{
  "node": { ... }
}
```

### Sessions

#### Create Session
```http
POST /api/v1/sessions
Authorization: Bearer <token>
Content-Type: application/json

{
  "node_id": "uuid",
  "protocol": "wireguard",
  "enable_kill_switch": true,
  "enable_dns_protection": true
}

Response 201:
{
  "session": {
    "id": "uuid",
    "node_id": "uuid",
    "tunnel_ip": "10.8.0.25",
    "protocol": "wireguard",
    "status": "active",
    "created_at": "2024-01-15T10:30:00Z"
  },
  "config": "[Interface]\nPrivateKey = ...\n..."
}
```

#### Disconnect Session
```http
DELETE /api/v1/sessions/:id
Authorization: Bearer <token>

Response 200:
{
  "session": {
    "id": "uuid",
    "status": "disconnected",
    "bytes_sent": 1234567890,
    "bytes_received": 9876543210,
    "duration_seconds": 3600
  }
}
```

### Operator Endpoints

#### Register as Operator
```http
POST /api/v1/operator/register
Authorization: Bearer <token>
Content-Type: application/json

{
  "wallet_address": "0x1234...",
  "wallet_currency": "ethereum",
  "company_name": "My VPN LLC",
  "contact_email": "contact@myvpn.com"
}

Response 201:
{
  "operator": {
    "id": "uuid",
    "status": "pending",
    "wallet_address": "0x1234...",
    "wallet_currency": "ethereum"
  },
  "message": "Registration submitted. Awaiting verification."
}
```

#### Create Operator Node
```http
POST /api/v1/operator/nodes
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "us-east-1",
  "hostname": "vpn.example.com",
  "ip_address": "203.0.113.1",
  "country": "US",
  "city": "New York",
  "latitude": 40.7128,
  "longitude": -74.0060,
  "wireguard_port": 51820,
  "max_connections": 500
}

Response 201:
{
  "node": {
    "id": "uuid",
    "name": "us-east-1",
    "wireguard_public_key": "base64...",
    "subnet": "10.8.0.0/24"
  },
  "private_key": "base64...",  // Save this securely!
  "setup_instructions": "..."
}
```

#### Get Operator Dashboard
```http
GET /api/v1/operator/dashboard
Authorization: Bearer <token>

Response 200:
{
  "total_earned_usd": 156.78,
  "pending_payout_usd": 23.45,
  "total_paid_usd": 133.33,
  "active_nodes": 2,
  "total_bandwidth_gb": 5678.90,
  "total_sessions": 12345,
  "average_uptime_percent": 99.2,
  "reputation_score": 87,
  "current_tier": "gold",
  "nodes": [
    {
      "id": "uuid",
      "name": "us-east-1",
      "status": "online",
      "load_score": 45,
      "active_connections": 123,
      "bandwidth_today_gb": 456.78,
      "earnings_today_usd": 9.12
    }
  ]
}
```

#### Request Payout
```http
POST /api/v1/operator/payout/request
Authorization: Bearer <token>

Response 201:
{
  "payout": {
    "id": "uuid",
    "amount_usd": 23.45,
    "amount_crypto": "0.0117",
    "currency": "ethereum",
    "wallet_address": "0x1234...",
    "status": "pending",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

### Health Endpoints

```http
GET /health
Response 200: {"status": "ok"}

GET /ready
Response 200: {"status": "ready", "database": "connected", "redis": "connected"}

GET /metrics
Response 200: # Prometheus metrics format
```

---

## Docker Deployment

### Production Setup

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  redis:
    image: redis:7-alpine
    restart: always
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data

  api-gateway:
    image: ghcr.io/nikola43/aureo-api-gateway:latest
    restart: always
    env_file: .env
    ports:
      - "8080:8080"
    volumes:
      - sqlite_data:/app/data    # SQLite database persistence
    depends_on:
      redis:
        condition: service_started
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  control-server:
    image: ghcr.io/nikola43/aureo-control-server:latest
    restart: always
    env_file: .env
    volumes:
      - sqlite_data:/app/data    # Shared SQLite database
    depends_on:
      - api-gateway

  prometheus:
    image: prom/prometheus:latest
    restart: always
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus_data:/prometheus
    ports:
      - "9090:9090"

  grafana:
    image: grafana/grafana:latest
    restart: always
    environment:
      GF_SECURITY_ADMIN_PASSWORD: ${GRAFANA_PASSWORD}
    volumes:
      - grafana_data:/var/lib/grafana
    ports:
      - "3000:3000"

volumes:
  sqlite_data:
  redis_data:
  prometheus_data:
  grafana_data:
```

> **Note**: SQLite database is stored in a Docker volume for persistence. The database file is created automatically on first run.

### Build Custom Images

```bash
# Build all images
docker-compose -f docker-compose.yml build

# Build specific service
docker build -t aureo-api-gateway -f deployments/docker/Dockerfile.api .
docker build -t aureo-vpn-node -f deployments/docker/Dockerfile.node .
```

---

## Security

### Best Practices

1. **Secrets Management**
   - Use strong, unique `JWT_SECRET` (32+ bytes)
   - Store private keys in secure vault (HashiCorp Vault, AWS Secrets Manager)
   - Rotate secrets regularly

2. **Network Security**
   - Enable TLS for all API traffic
   - Configure firewall rules
   - Use VPN or private network for internal services
   - Enable rate limiting

3. **Database Security**
   - Restrict file permissions on SQLite database (chmod 600)
   - Store database file in secure location
   - Regular backups of the database file
   - Consider encryption at rest for sensitive deployments

4. **Monitoring**
   - Enable audit logging
   - Set up alerts for suspicious activity
   - Monitor failed login attempts
   - Track API abuse patterns

### Security Headers

The API Gateway automatically sets:
- `X-Frame-Options: DENY`
- `X-Content-Type-Options: nosniff`
- `X-XSS-Protection: 1; mode=block`
- `Content-Security-Policy: default-src 'self'`
- `Strict-Transport-Security: max-age=31536000`

---

## Development

### Project Structure

```
aureo-vpn/
├── cmd/                    # Application entry points
│   ├── api-gateway/       # REST API server
│   ├── control-server/    # Control plane
│   ├── vpn-node/          # VPN node service
│   └── cli/               # CLI tool
├── pkg/                    # Shared packages
│   ├── auth/              # JWT authentication
│   ├── crypto/            # Cryptographic utilities
│   ├── database/          # Database connection
│   ├── models/            # Data models (GORM)
│   ├── protocols/         # VPN protocols
│   │   ├── wireguard/    # WireGuard implementation
│   │   └── openvpn/      # OpenVPN implementation
│   ├── blockchain/        # Crypto payments
│   ├── rewards/           # Reward calculation
│   ├── p2p/               # P2P networking
│   └── metrics/           # Prometheus metrics
├── internal/               # Private packages
│   ├── api/               # API handlers
│   ├── control/           # Control server logic
│   ├── node/              # Node service logic
│   └── security/          # Security features
├── deployments/            # Deployment configs
│   ├── docker/            # Docker Compose
│   └── kubernetes/        # K8s manifests
├── tests/                  # Test files
├── web/                    # Web dashboards
└── scripts/               # Utility scripts
```

### Development Mode

```bash
# Run API with live reload
make dev-api

# Run control server with live reload
make dev-control

# Run VPN node (requires root)
sudo make dev-node

# Run all tests
make test

# Run linter
make lint

# Format code
make fmt
```

### Adding New Features

1. **Database Changes**
   ```bash
   # Create new migration
   make migrate-create name=add_new_feature

   # Run migrations
   make migrate-up
   ```

2. **API Endpoints**
   - Add handler in `internal/api/handlers/`
   - Register route in `internal/api/routes.go`
   - Add tests in `tests/`

3. **Models**
   - Define in `pkg/models/`
   - Add migrations
   - Update related handlers

---

## Testing

### Running Tests

```bash
# All tests
make test

# Unit tests only
make test-unit

# Integration tests
make test-integration

# Specific package
go test -v ./pkg/auth/...

# With coverage
make coverage

# Race detection
go test -race ./...
```

### Test Database

```bash
# SQLite test database is created automatically in memory or temp file

# Run integration tests (uses separate test database)
make test-integration

# Run with specific test database path
DB_PATH=./test_aureo.db make test-integration

# Clean up test database
rm -f ./test_aureo.db
```

---

## Troubleshooting

### Common Issues

**Database issues**:
```bash
# Check SQLite database file exists
ls -la ./data/aureo.db

# Check file permissions
stat ./data/aureo.db

# Verify database integrity
sqlite3 ./data/aureo.db "PRAGMA integrity_check;"

# Check environment variable
echo $DB_PATH
```

**JWT token invalid**:
```bash
# Verify JWT_SECRET is set
echo $JWT_SECRET | wc -c  # Should be 32+

# Check token format
echo $TOKEN | cut -d. -f2 | base64 -d 2>/dev/null | jq .
```

**WireGuard not working**:
```bash
# Check kernel module
lsmod | grep wireguard

# Install if missing
sudo apt install wireguard-dkms wireguard-tools

# Check interface
ip link show wg0
wg show wg0
```

**Port already in use**:
```bash
# Find process using port
sudo lsof -i :8080

# Kill process
sudo kill -9 <PID>
```

### Logs

```bash
# API Gateway logs
docker-compose logs -f api-gateway

# All service logs
docker-compose logs -f

# System logs (VPN node)
journalctl -u aureo-vpn-node -f
```

---

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

### Code Style

- Follow Go conventions
- Run `make lint` before committing
- Add tests for new features
- Update documentation

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Support

- **Issues**: https://github.com/nikola43/aureo-vpn/issues
- **Documentation**: https://docs.aureo-vpn.com
- **Discord**: https://discord.gg/aureo-vpn

---

## Acknowledgments

- [WireGuard](https://www.wireguard.com/) - Modern VPN protocol
- [libp2p](https://libp2p.io/) - P2P networking stack
- [Fiber](https://gofiber.io/) - Express-inspired web framework
- [GORM](https://gorm.io/) - The fantastic ORM library for Go
