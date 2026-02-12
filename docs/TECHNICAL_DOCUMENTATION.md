<div align="center">

# 📖 Aureo VPN — Technical Documentation

### Complete Developer Reference

*For the Aureo decentralized VPN platform.*

</div>

---

## Table of Contents

1. [🔭 System Overview](#1--system-overview)
2. [🏗️ Architecture](#2-️-architecture)
3. [🗂️ Data Models](#3-️-data-models)
4. [📖 API Reference](#4--api-reference)
5. [🔑 Authentication Flow](#5--authentication-flow)
6. [🔗 VPN Connection Flow](#6--vpn-connection-flow)
7. [🌐 P2P Network Architecture](#7--p2p-network-architecture)
8. [💰 Blockchain & Rewards System](#8--blockchain--rewards-system)
9. [🎛️ Control Server](#9-️-control-server)
10. [📡 VPN Node Service](#10--vpn-node-service)
11. [🛡️ Security Model](#11-️-security-model)
12. [📊 Monitoring & Metrics](#12--monitoring--metrics)
13. [📱 Mobile App Architecture](#13--mobile-app-architecture)
14. [🖥️ Desktop App Architecture](#14-️-desktop-app-architecture)
15. [🐳 Deployment](#15--deployment)
16. [🧪 Testing](#16--testing)

---

## 🔭 1. System Overview

Aureo VPN is a decentralized VPN platform consisting of three independent packages:

```
aureo/
├── aureo-vpn/       # Backend (Go 1.24, Fiber v2, GORM, libp2p)
├── aureo-app/       # Mobile client (Expo 54, React Native 0.83, TypeScript)
└── aureo-desktop/   # Desktop client (Go + Wails 2.x)
```

### 🏗️ High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLIENTS                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │  Mobile App   │  │ Desktop App  │  │    CLI Tool          │  │
│  │ (Expo/RN)     │  │ (Wails/Go)   │  │  (Cobra/Go)         │  │
│  └──────┬───────┘  └──────┬───────┘  └──────────┬───────────┘  │
└─────────┼──────────────────┼────────────────────┼───────────────┘
          │                  │                    │
          │ HTTPS/REST       │ HTTPS/REST         │ HTTPS/REST
          ▼                  ▼                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                    API GATEWAY (Fiber v2)                        │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────────┐  │
│  │   Auth    │ │  Nodes   │ │ Sessions │ │    Operators     │  │
│  │  Handler  │ │  Handler │ │  Handler │ │    Handler       │  │
│  └──────────┘ └──────────┘ └──────────┘ └──────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ Middleware: JWT Auth | Rate Limit | CORS | Metrics       │  │
│  └──────────────────────────────────────────────────────────┘  │
└────────────────────────┬────────────────────────────────────────┘
                         │
          ┌──────────────┼──────────────┐
          ▼              ▼              ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│  PostgreSQL  │ │   Control    │ │  Blockchain  │
│  / SQLite    │ │   Server     │ │   Service    │
│  (GORM)      │ │              │ │ (ETH/BTC/LTC)│
└──────────────┘ └──────┬───────┘ └──────────────┘
                        │
                        ▼
              ┌──────────────────┐
              │    VPN Nodes     │
              │  ┌────────────┐  │
              │  │ WireGuard  │  │
              │  │  Manager   │  │
              │  ├────────────┤  │
              │  │  P2P Host  │◄─┼──── libp2p (Kademlia DHT + GossipSub)
              │  │ (libp2p)   │  │
              │  ├────────────┤  │
              │  │  Traffic   │  │
              │  │  Monitor   │  │
              │  └────────────┘  │
              └──────────────────┘
```

---

## 🏗️ 2. Architecture

### ⚙️ Backend Binaries

Four binaries are built from `aureo-vpn/cmd/`:

| Binary | Source | Purpose | Default Port |
|--------|--------|---------|-------------|
| `api-gateway` | `cmd/api-gateway/` | REST API server (user-facing) | 8080 |
| `control-server` | `cmd/control-server/` | Network orchestrator | - |
| `vpn-node` | `cmd/vpn-node/` | Individual VPN node (requires sudo) | 51820 (WG) |
| `cli` | `cmd/cli/` | CLI management tool (Cobra) | - |

Additionally, `cmd/aureo-node/` provides an all-in-one decentralized node with embedded SQLite and P2P.

### 📁 Package Structure

```
aureo-vpn/
├── cmd/
│   ├── api-gateway/       # REST API entry point
│   ├── control-server/    # Network orchestrator
│   ├── vpn-node/          # VPN node service
│   ├── aureo-node/        # All-in-one decentralized node
│   └── cli/               # CLI tool (Cobra)
├── internal/
│   ├── api/               # HTTP handlers (handlers.go)
│   ├── control/           # Control server logic
│   └── node/              # VPN node service logic
├── pkg/
│   ├── auth/              # JWT token service + auth service
│   ├── blockchain/        # Multi-chain crypto integration
│   ├── config/            # Configuration loader (env vars)
│   ├── crypto/            # Password hashing (bcrypt)
│   ├── database/          # GORM (PostgreSQL + SQLite)
│   ├── errors/            # AppError type
│   ├── logger/            # Structured logging (slog)
│   ├── metrics/           # Prometheus metrics
│   ├── middleware/        # Auth + admin + rate limit middleware
│   ├── models/            # GORM models (6 files)
│   ├── operator/          # Operator service
│   ├── p2p/               # libp2p host, registry, client
│   ├── protocols/
│   │   ├── wireguard/     # WireGuard manager + key generation
│   │   ├── openvpn/       # OpenVPN config (planned)
│   │   └── ipsec/         # IPSec config (planned)
│   ├── rewards/           # Crypto reward tiers + payout logic
│   └── security/          # Input validation, privacy filter
└── tests/
    ├── unit/
    └── integration/
```

### 🔗 Dependency Graph

```
api-gateway
  ├── auth (JWT + bcrypt)
  ├── blockchain (ETH/BTC/LTC)
  ├── rewards (tiers + payouts)
  ├── operator (service layer)
  ├── middleware (auth, admin, rate limit)
  ├── metrics (Prometheus)
  ├── database (GORM)
  └── security (validation, privacy)

vpn-node
  ├── protocols/wireguard (interface mgmt)
  ├── p2p (libp2p host, registry)
  ├── rewards (earnings recording)
  ├── metrics (Prometheus)
  └── database (GORM)

control-server
  ├── database (GORM)
  └── models (VPNNode, Session)
```

---

## 🗂️ 3. Data Models

All models use UUID primary keys with GORM. Database: PostgreSQL (production) or SQLite (development/nodes).

### 👤 User

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Primary key (`gen_random_uuid()`) |
| `email` | string | Unique, not null |
| `password_hash` | string | bcrypt hash (JSON: `-`) |
| `username` | string | Unique, not null |
| `full_name` | string | Optional display name |
| `is_active` | bool | Default: `true` |
| `is_admin` | bool | Default: `false` |
| `subscription_tier` | string | `free`, `basic`, `premium` |
| `subscription_expiry` | timestamp | Subscription end date |
| `data_transferred_gb` | float64 | Cumulative data usage |
| `connection_count` | int64 | Total connections made |
| `two_factor_enabled` | bool | 2FA status |
| `two_factor_secret` | string | TOTP secret (JSON: `-`) |
| `created_at` | timestamp | |
| `updated_at` | timestamp | |
| `deleted_at` | timestamp | Soft delete (GORM) |

**Relations:** `Sessions[]`, `Configs[]`

### 📡 VPNNode

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Primary key |
| `name` | string | Unique node name |
| `hostname` | string | DNS hostname |
| `country` | string | Full country name |
| `country_code` | string(2) | ISO 3166-1 alpha-2 |
| `city` | string | City name |
| `latitude` / `longitude` | float64 | Geolocation |
| `public_ip` | string | Public IPv4 address |
| `internal_ip` | string | WireGuard subnet IP (e.g. `10.0.0.1`) |
| `ipv6_address` | string | IPv6 address |
| `max_connections` | int | Default: 1000 |
| `current_connections` | int | Live count |
| `cpu_usage` | float64 | CPU % (0-100) |
| `memory_usage` | float64 | Memory % (0-100) |
| `bandwidth_usage_gbps` | float64 | Current throughput |
| `total_bandwidth_kb` | int64 | Cumulative traffic (KB) |
| `load_score` | float64 | 0-100, lower is better |
| `status` | string | `online`, `offline`, `maintenance` |
| `is_active` | bool | Enabled/disabled |
| `last_heartbeat` | timestamp | Last heartbeat received |
| `latency` | int | Milliseconds |
| `supports_wireguard` | bool | Default: `true` |
| `supports_openvpn` | bool | Default: `true` |
| `wireguard_port` | int | Default: 51820 |
| `openvpn_port` | int | Default: 1194 |
| `public_key` | string | WireGuard public key |
| `private_key_encrypted` | string | WireGuard private key (JSON: `-`) |
| `supports_multihop` | bool | Multi-hop routing |
| `supports_obfuscation` | bool | Traffic obfuscation |
| `supports_socks5` | bool | SOCKS5 proxy |
| `operator_id` | UUID? | FK to NodeOperator (null = company-owned) |
| `is_operator_owned` | bool | Community-operated flag |
| `uptime_percentage` | float64 | Historical uptime % |
| `total_earned_usd` | float64 | Cumulative operator earnings |

**Load Score Formula:**
```
LoadScore = (ConnectionLoad * 0.4) + (CPULoad * 0.3) + (MemoryLoad * 0.3)
where ConnectionLoad = (CurrentConnections / MaxConnections) * 100
```

> ⚠️ **Health Check:** Node is healthy if `IsActive && Status == "online" && LastHeartbeat < 2min && LoadScore <= 90`

### 🔗 Session

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Primary key |
| `user_id` | UUID | FK to User |
| `node_id` | UUID | FK to VPNNode |
| `protocol` | string | `wireguard`, `openvpn` |
| `client_ip` | string | Anonymized hash of real IP |
| `tunnel_ip` | string | VPN tunnel IP (e.g. `10.0.0.2`) |
| `public_key` | string | Client's WireGuard public key |
| `private_key` | string | Encrypted (JSON: `-`) |
| `status` | string | `pending`, `active`, `disconnected`, `terminated`, `pending_disconnect` |
| `connected_at` | timestamp | Connection start time |
| `disconnected_at` | timestamp? | Connection end time |
| `bytes_sent` | int64 | Upload bytes |
| `bytes_received` | int64 | Download bytes |
| `data_used_gb` | float64 | `(sent + received) / 1024^3` |
| `latency` | int | Milliseconds |
| `packet_loss` | float64 | Percentage |
| `last_keepalive` | timestamp | Last WireGuard handshake |
| `split_tunnel_enabled` | bool | |
| `kill_switch_enabled` | bool | Default: `true` |
| `dns_leak_protection` | bool | Default: `true` |
| `is_multihop` | bool | Multi-hop active |
| `next_hop_node_id` | UUID? | Second hop node |
| `client_version` | string | App version |
| `device_type` | string | `desktop`, `mobile`, `router` |
| `os_type` | string | `linux`, `macos`, `windows`, `ios`, `android` |

### 💼 NodeOperator

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Primary key |
| `user_id` | UUID | FK to User (unique) |
| `wallet_address` | string | Crypto wallet (unique) |
| `wallet_type` | string | `ethereum`, `bitcoin`, `litecoin` |
| `status` | string | `pending`, `active`, `suspended`, `banned` |
| `is_verified` | bool | Admin-verified |
| `total_earned` | float64 | Total USD earned (paid) |
| `pending_payout` | float64 | Confirmed, awaiting payout |
| `total_nodes_created` | int | Lifetime nodes |
| `active_nodes_count` | int | Currently online |
| `total_bandwidth_kb` | int64 | Cumulative traffic |
| `total_connections_served` | int64 | Lifetime connections |
| `average_uptime` | float64 | Across all nodes (%) |
| `reputation_score` | float64 | 0-100, starts at 50 |
| `stake_amount` | float64 | Security deposit (USD) |
| `stake_status` | string | `none`, `staked`, `locked`, `slashed` |
| `kyc_status` | string | `not_required`, `pending`, `approved`, `rejected` |

**Reputation Score Calculation:**
```
Base:       50 points
+ Uptime:   (AverageUptime / 100) * 30    (max 30 pts)
+ Ratings:  (AvgUserRating / 5) * 20      (max 20 pts)
+ Traffic:  100GB+ = 5pts, 1TB+ = 10pts   (max 10 pts)
+ Stake:    $100+ = 5pts, $1000+ = 10pts   (max 10 pts)
= Cap: 100
```

**Relations:** `Nodes[]`, `Earnings[]`, `Payouts[]`

### 💰 OperatorEarning

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Primary key |
| `operator_id` | UUID | FK to NodeOperator |
| `node_id` | UUID | FK to VPNNode |
| `session_id` | UUID | FK to Session |
| `bandwidth_kb` | int64 | Traffic served |
| `duration_minutes` | int | Session duration |
| `rate_per_gb` | float64 | USD/GB (from tier) |
| `amount_usd` | float64 | Earned amount |
| `status` | string | `pending`, `confirmed`, `paid` |
| `connection_quality` | float64 | 0-100 quality score |
| `user_rating` | int | 1-5 user rating |

### 💳 OperatorPayout

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Primary key |
| `operator_id` | UUID | FK to NodeOperator |
| `amount_usd` | float64 | Payout amount in USD |
| `crypto_amount` | float64 | Amount in cryptocurrency |
| `crypto_currency` | string | `ETH`, `BTC`, `LTC` |
| `exchange_rate` | float64 | USD per crypto unit |
| `wallet_address` | string | Destination wallet |
| `transaction_hash` | string | Blockchain tx hash |
| `transaction_fee` | float64 | Network fee |
| `status` | string | `pending`, `processing`, `completed`, `failed` |
| `payout_method` | string | `blockchain`, `exchange`, `manual` |
| `failure_reason` | string | Error details if failed |

### 🏆 NodeReward

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Primary key |
| `tier_name` | string | `bronze`, `silver`, `gold`, `platinum` |
| `min_reputation_score` | float64 | Required reputation |
| `min_uptime_percent` | float64 | Required uptime |
| `base_rate_per_gb` | float64 | USD per GB |
| `bonus_multiplier` | float64 | Rate multiplier |
| `min_bandwidth` | int | Required Mbps |
| `max_latency` | int | Maximum ms |
| `is_active` | bool | Tier enabled |

### 📊 NodePerformanceMetric

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Primary key |
| `node_id` | UUID | FK to VPNNode |
| `metric_date` | date | Metric date |
| `hour` | int | 0-23 |
| `uptime_minutes` | int | Minutes online |
| `downtime_minutes` | int | Minutes offline |
| `connections_served` | int | Connections in period |
| `bandwidth_gb` | float64 | Traffic in period |
| `average_latency_ms` | int | Avg latency |
| `availability_score` | float64 | 0-100 |
| `performance_score` | float64 | 0-100 |
| `user_satisfaction` | float64 | 0-100 |
| `earnings_usd` | float64 | Earnings in period |

### ⚙️ Config

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Primary key |
| `user_id` | UUID | FK to User |
| `node_id` | UUID | FK to VPNNode |
| `protocol` | string | `wireguard`, `openvpn` |
| `config_name` | string | Display name |
| `config_content` | text | Encrypted config file (JSON: `-`) |
| `config_hash` | string | Content hash |
| `public_key` | string | WireGuard public key |
| `private_key` | string | Encrypted (JSON: `-`) |
| `dns_servers` | string | Comma-separated DNS |
| `allowed_ips` | string | Split tunnel IPs |
| `mtu` | int | Default: 1420 |
| `persistent_keepalive` | int | Default: 25 seconds |
| `is_active` | bool | Config enabled |
| `times_used` | int64 | Usage count |
| `expires_at` | timestamp? | Optional expiration |

### 🗄️ Local Models (SQLite - per-node)

Used by `aureo-node` for standalone operation:

- **LocalUser** - Node-local user registration
- **LocalSession** - Node-local session tracking
- **LocalNodeConfig** - Key-value config store
- **NodeIdentity** - Node ID, P2P keys, WireGuard keys

---

## 📖 4. API Reference

Base URL: `/api/v1`

### ❤️ Health & Metrics

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | No | Liveness check |
| GET | `/ready` | No | Readiness check (DB status) |
| GET | `/metrics` | No | Prometheus metrics (if enabled) |

### 🔑 Authentication

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/v1/auth/register` | No | Register new user |
| POST | `/api/v1/auth/login` | No | Login |
| POST | `/api/v1/auth/refresh` | No | Refresh access token |

**POST /api/v1/auth/register**
```json
// Request
{ "email": "user@example.com", "password": "securepass", "username": "johndoe" }

// Response (201)
{
  "access_token": "eyJhbG...",
  "refresh_token": "eyJhbG...",
  "user": { "id": "uuid", "email": "...", "username": "..." }
}
```

**POST /api/v1/auth/login**
```json
// Request
{ "email": "user@example.com", "password": "securepass" }

// Response (200) — same shape as register
```

**POST /api/v1/auth/refresh**
```json
// Request
{ "refresh_token": "eyJhbG..." }

// Response (200)
{ "access_token": "eyJhbG..." }
```

### 👤 User

All routes require `Authorization: Bearer <access_token>`.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/user/profile` | Get authenticated user profile |
| PUT | `/api/v1/user/profile` | Update username/email |
| GET | `/api/v1/user/sessions` | Get active sessions |
| GET | `/api/v1/user/stats` | Get usage statistics |
| PUT | `/api/v1/user/password` | Change password |

**PUT /api/v1/user/profile**
```json
// Request
{ "username": "newname", "email": "new@email.com" }
```

**PUT /api/v1/user/password**
```json
// Request
{ "old_password": "current", "new_password": "newpass123" }
```

**GET /api/v1/user/stats**
```json
// Response
{
  "total_sessions": 42,
  "active_sessions": 1,
  "data_transferred_gb": 15.7
}
```

### 📡 Nodes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/nodes` | List available nodes |
| GET | `/api/v1/nodes/best` | Get best node by load |
| GET | `/api/v1/nodes/:id` | Get specific node |

**GET /api/v1/nodes?country=US&protocol=wireguard&source=auto**

Query params:
- `country` — ISO 3166-1 alpha-2 code
- `protocol` — `wireguard` or `openvpn`
- `source` — `p2p`, `db`, or `auto` (default)

```json
// Response
{
  "nodes": [...],
  "count": 25,
  "source": "p2p"  // or "database"
}
```

**GET /api/v1/nodes/best?protocol=wireguard&country=US**
```json
// Response — single node object
{
  "id": "uuid",
  "name": "US-East-1",
  "public_ip": "1.2.3.4",
  "load_score": 12.5,
  ...
}
```

### 🔗 Sessions

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/sessions` | Create VPN session |
| GET | `/api/v1/sessions/:id` | Get session details |
| DELETE | `/api/v1/sessions/:id` | Disconnect session |

**POST /api/v1/sessions**
```json
// Request
{ "node_id": "uuid", "protocol": "wireguard" }

// Response (201)
{
  "id": "session-uuid",
  "user_id": "...",
  "node_id": "...",
  "protocol": "wireguard",
  "status": "pending",
  "tunnel_ip": "",
  "kill_switch_enabled": true,
  "dns_leak_protection": true
}
```

### ⚙️ Config

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/config/generate` | Generate WireGuard config |
| GET | `/api/v1/config/:id` | Get saved config |
| GET | `/api/v1/config` | List user configs |

**POST /api/v1/config/generate** — Primary WireGuard connection endpoint
```json
// Request
{ "node_id": "uuid", "public_key": "base64-wg-pubkey" }

// Response (201)
{
  "session_id": "uuid",
  "server_public_key": "base64...",
  "server_endpoint": "1.2.3.4:51820",
  "client_ip": "10.0.0.2",
  "dns": "1.1.1.1,8.8.8.8",
  "allowed_ips": "0.0.0.0/0",
  "keepalive": 25
}
```

### 💼 Operator

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/v1/operator/register` | JWT | Register as operator |
| POST | `/api/v1/operator/nodes` | JWT | Create operator node |
| GET | `/api/v1/operator/nodes` | JWT | List operator nodes |
| GET | `/api/v1/operator/stats` | JWT | Operator statistics |
| GET | `/api/v1/operator/earnings` | JWT | Earnings history (paginated) |
| GET | `/api/v1/operator/payouts` | JWT | Payout history (paginated) |
| POST | `/api/v1/operator/payout/request` | JWT | Request manual payout |
| GET | `/api/v1/operator/dashboard` | JWT | Full dashboard data |
| GET | `/api/v1/operator/rewards/tiers` | No | Get reward tiers (public) |

**POST /api/v1/operator/register**
```json
// Request
{
  "wallet_address": "0x...",
  "wallet_type": "ethereum",
  "email": "operator@example.com",
  "country": "US"
}
```

**GET /api/v1/operator/stats**
```json
// Response
{
  "operator_id": "uuid",
  "total_earned": 150.50,
  "pending_payout": 25.00,
  "earnings_today": 2.30,
  "earnings_week": 18.50,
  "earnings_month": 65.00,
  "active_nodes": 3,
  "total_bandwidth_gb": 1024.5,
  "reputation_score": 82.5,
  "average_uptime": 99.2,
  "current_tier": "gold",
  "rate_per_gb": 0.03,
  "connected_users": 47,
  "current_traffic": 250.0
}
```

### 🔐 Admin

All admin routes require JWT + `is_admin: true`.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/admin/nodes` | List all nodes |
| POST | `/api/v1/admin/nodes` | Create node |
| PUT | `/api/v1/admin/nodes/:id` | Update node |
| DELETE | `/api/v1/admin/nodes/:id` | Delete node |
| GET | `/api/v1/admin/users` | List all users |
| GET | `/api/v1/admin/users/:id` | Get user |
| PUT | `/api/v1/admin/users/:id` | Update user |
| DELETE | `/api/v1/admin/users/:id` | Delete user |
| GET | `/api/v1/admin/stats` | System statistics |
| GET | `/api/v1/admin/sessions` | All sessions |
| PUT | `/api/v1/admin/operators/:id/verify` | Verify operator |

**GET /api/v1/admin/stats**
```json
{
  "total_users": 1500,
  "active_users": 1200,
  "total_nodes": 50,
  "online_nodes": 45,
  "total_sessions": 50000,
  "active_sessions": 320
}
```

---

## 🔑 5. Authentication Flow

### 🔐 JWT Configuration

- **Algorithm:** HS256 (HMAC-SHA256)
- **Issuer:** `aureo-vpn`
- **Access Token Duration:** 15 minutes (configurable)
- **Refresh Token Duration:** 7 days (configurable)

### 📋 Token Claims

```go
type Claims struct {
    UserID    uuid.UUID  // User's UUID
    Email     string
    Username  string
    IsAdmin   bool
    TokenType string     // "access" or "refresh"
    jwt.RegisteredClaims // ExpiresAt, IssuedAt, NotBefore, Issuer, Subject
}
```

### 🔄 Flow Diagram

```
                    ┌──────────┐
                    │  Client  │
                    └────┬─────┘
                         │
    1. POST /auth/register or /auth/login
    ┌────────────────────┤
    │                    ▼
    │            ┌───────────────┐
    │            │  Auth Service │
    │            ├───────────────┤
    │            │ - Hash password│ (bcrypt)
    │            │ - Find/Create │ user
    │            │ - Generate    │ token pair
    │            └───────┬───────┘
    │                    │
    │    { access_token, refresh_token, user }
    │◄───────────────────┤
    │                    │
    │  2. Authenticated requests
    │  Authorization: Bearer <access_token>
    ├───────────────────►│
    │                    │
    │            ┌───────────────┐
    │            │Auth Middleware│
    │            ├───────────────┤
    │            │ - Parse Bearer│
    │            │ - Verify JWT  │
    │            │ - Extract     │ claims
    │            │ - Set ctx     │ locals
    │            └───────┬───────┘
    │                    │
    │  3. When access token expires (401)
    │  POST /auth/refresh { refresh_token }
    ├───────────────────►│
    │                    │
    │            ┌───────────────┐
    │            │Token Service  │
    │            ├───────────────┤
    │            │ - Verify      │ refresh token
    │            │ - Check type  │ == "refresh"
    │            │ - Generate    │ new access token
    │            └───────┬───────┘
    │                    │
    │    { access_token }│
    │◄───────────────────┘
```

### ⚙️ Middleware Stack

```
Request → Recover → RequestID → Compress → Logger → CORS → Metrics → RateLimit → [Auth] → Handler
```

---

## 🔗 6. VPN Connection Flow

### 📱 Client-Side (Desktop + Mobile)

```
┌──────────────────────────────────────────────────────────────┐
│                    CLIENT CONNECTION FLOW                     │
│                                                              │
│  1. Generate WireGuard key pair (X25519)                     │
│     ├── Desktop: `wg genkey` + `wg pubkey`                   │
│     └── Mobile: tweetnacl.box.keyPair() + expo-crypto PRNG   │
│                                                              │
│  2. POST /api/v1/config/generate                             │
│     { node_id: "uuid", public_key: "base64..." }            │
│                                                              │
│  3. Receive server config:                                   │
│     { session_id, server_public_key, server_endpoint,        │
│       client_ip, dns, allowed_ips, keepalive }               │
│                                                              │
│  4. Write WireGuard config file                              │
│     [Interface]                                              │
│     PrivateKey = <client_private_key>                        │
│     Address = <client_ip>/32                                 │
│     DNS = 1.1.1.1,8.8.8.8                                   │
│                                                              │
│     [Peer]                                                   │
│     PublicKey = <server_public_key>                           │
│     Endpoint = <server_endpoint>                             │
│     AllowedIPs = 0.0.0.0/0                                   │
│     PersistentKeepalive = 25                                 │
│                                                              │
│  5. Activate tunnel                                          │
│     ├── Desktop: `wg-quick up` (admin elevation)             │
│     └── Mobile: NativeModule.startTunnel(config)             │
│                                                              │
│  6. Poll stats every 2-3 seconds                             │
│     ├── Desktop: `wg show` → parse bytes/handshake           │
│     └── Mobile: VPNModule.getStatistics() + GET /sessions/:id│
│                                                              │
│  7. Disconnect                                               │
│     ├── Stop native tunnel                                   │
│     ├── DELETE /api/v1/sessions/:id                          │
│     └── Clean up config files                                │
└──────────────────────────────────────────────────────────────┘
```

### 🖥️ Server-Side (VPN Node)

```
┌──────────────────────────────────────────────────────────────┐
│                    NODE SESSION LIFECYCLE                     │
│                                                              │
│  1. watchPendingSessions() polls DB every 5 seconds          │
│     └── Finds sessions with status = "pending"               │
│                                                              │
│  2. provisionSession()                                       │
│     ├── Check node capacity                                  │
│     ├── Use client's public_key (or generate server-side)    │
│     ├── Allocate tunnel IP if not allocated                   │
│     ├── `wg set wg0 peer <pubkey> allowed-ips <ip>/32`       │
│     ├── Update session → status = "active"                   │
│     └── Increment node current_connections                   │
│                                                              │
│  3. trafficMonitor() runs every 1 second                     │
│     ├── `wg show wg0 dump` → parse peer stats               │
│     ├── Update bytes_sent, bytes_received per session        │
│     ├── Flush stats to DB every 3 seconds                    │
│     └── Flush earnings every 10 minutes                      │
│                                                              │
│  4. sessionMonitor() runs every 1 minute                     │
│     └── Disconnect sessions with no keepalive > 10 minutes   │
│                                                              │
│  5. processDisconnects() polls DB every 5 seconds            │
│     ├── Finds sessions with status = "pending_disconnect"    │
│     ├── `wg set wg0 peer <pubkey> remove`                    │
│     ├── Update session → status = "disconnected"             │
│     └── Decrement node current_connections                   │
│                                                              │
│  6. heartbeatLoop() runs every 30 seconds                    │
│     ├── Update last_heartbeat + status in DB                 │
│     └── Broadcast to P2P network                             │
└──────────────────────────────────────────────────────────────┘
```

### 🌐 IP Allocation

Clients receive IPs from the node's `/24` subnet. The node uses `.1`, clients get `.2-.254`:

```go
func allocateClientIP(nodeIP string, usedIPs []string) string {
    // nodeIP = "10.0.0.1"
    // Iterates .2 through .254, skipping used IPs
    // Returns first available, e.g. "10.0.0.2"
}
```

---

## 🌐 7. P2P Network Architecture

### 📋 Overview

The P2P layer uses [libp2p](https://libp2p.io/) for decentralized node discovery. It is **optional** — the system works with database-only node lists.

> 💡 **Note:** The P2P layer is optional — the system works with database-only node lists.

### ⚙️ Components

```
┌─────────────────────────────────────────────┐
│               P2P Host (libp2p)             │
│                                             │
│  ┌───────────┐  ┌───────────┐  ┌─────────┐│
│  │ Kademlia  │  │ GossipSub │  │  mDNS   ││
│  │   DHT     │  │  PubSub   │  │  Local  ││
│  └─────┬─────┘  └─────┬─────┘  └────┬────┘│
│        │              │             │      │
│  ┌─────┴──────────────┴─────────────┴────┐ │
│  │              Registry                  │ │
│  │  (in-memory node store, thread-safe)   │ │
│  │  Max: 1000 nodes, timeout: 2 min       │ │
│  └────────────────────────────────────────┘ │
│                                             │
│  Identity: Ed25519 key pair (base64)        │
│  Transport: TCP + QUIC-v1                   │
│  NAT: UPnP, hole punching, relay           │
└─────────────────────────────────────────────┘
```

### 🔗 Protocol IDs

| Protocol | ID | Purpose |
|----------|----|---------|
| Node Info | `/aureo/nodeinfo/1.0.0` | Request node metadata |
| Node List | `/aureo/nodelist/1.0.0` | Request known nodes |
| Health Check | `/aureo/health/1.0.0` | Peer health check |

### 📡 PubSub Topics

| Topic | Purpose | Interval |
|-------|---------|----------|
| `/aureo/nodes/announce/1.0.0` | Node joins + full info updates | 5 minutes |
| `/aureo/nodes/heartbeat/1.0.0` | Status, load, connections | 30 seconds |

### 📨 Message Types

**AnnounceMessage** — Full node info broadcast:
```json
{
  "node": {
    "id": "uuid",
    "peer_id": "QmPeer...",
    "name": "US-East-1",
    "public_ip": "1.2.3.4",
    "country": "United States",
    "country_code": "US",
    "wireguard_port": 51820,
    "public_key": "base64...",
    "max_connections": 1000,
    "current_connections": 45,
    "load_score": 15.2,
    "status": "online",
    "supports_wireguard": true,
    "supports_openvpn": true,
    "is_operator_owned": false,
    "reputation": 85.0,
    "multiaddrs": ["/ip4/1.2.3.4/tcp/4001/p2p/QmPeer..."]
  },
  "timestamp": 1700000000000
}
```

**HeartbeatMessage** — Lightweight status update:
```json
{
  "node_id": "uuid",
  "peer_id": "QmPeer...",
  "status": "online",
  "current_connections": 47,
  "load_score": 16.1,
  "cpu_usage": 25.0,
  "memory_usage": 40.0,
  "bandwidth_gbps": 0.5,
  "timestamp": 1700000030000
}
```

### ⚙️ Default Configuration

```go
Config{
    ListenAddrs:       ["/ip4/0.0.0.0/tcp/4001", "/ip4/0.0.0.0/udp/4001/quic-v1"],
    EnableDHT:         true,
    DHTServerMode:     true,
    EnableMDNS:        true,
    EnablePubSub:      true,
    HeartbeatInterval: 30s,
    AnnounceInterval:  5m,
    NodeTimeout:       2m,
    MaxPeers:          100,
    MaxNodes:          1000,
    DataDir:           "./data/p2p",
}
```

### 🔄 Background Loops

| Loop | Interval | Action |
|------|----------|--------|
| `heartbeatLoop` | 30s | Broadcast `HeartbeatMessage` to PubSub |
| `announceLoop` | 5m | Broadcast `AnnounceMessage` to PubSub |
| `discoveryLoop` | 30s | Find peers via DHT, request node info |
| `cleanupLoop` | 5m | Mark offline nodes, remove stale entries, persist registry |

---

## 💰 8. Blockchain & Rewards System

### ⛓️ Multi-Chain Support

| Chain | Library | RPC Config |
|-------|---------|------------|
| Ethereum | go-ethereum | URL + Private Key + Chain ID |
| Bitcoin | RPC client | URL + User + Password |
| Litecoin | RPC client | URL + User + Password |

### ⚙️ Blockchain Service API

```go
// Send a crypto payout
SendTransaction(ctx, walletType, toAddress, amountUSD) → (*Transaction, error)

// Check transaction status
GetTransactionStatus(ctx, walletType, txHash) → (*Transaction, error)

// Validate crypto address format
ValidateAddress(walletType, address) → (bool, error)

// Get wallet balance
GetBalance(ctx, walletType) → (*big.Float, error)

// Estimate network fee
EstimateFee(ctx, walletType, amountUSD) → (*big.Float, error)
```

### 🏆 Reward Tiers

| Tier | Reputation | Uptime | Rate/GB | Bonus | Min Bandwidth | Max Latency |
|------|-----------|--------|---------|-------|---------------|-------------|
| 🥉 Bronze | 0+ | 50%+ | $0.010 | 1.0x | 50 Mbps | 150ms |
| 🥈 Silver | 60+ | 80%+ | $0.015 | 1.2x | 100 Mbps | 100ms |
| 🥇 Gold | 75+ | 90%+ | $0.020 | 1.5x | 200 Mbps | 75ms |
| 💎 Platinum | 90+ | 95%+ | $0.030 | 2.0x | 500 Mbps | 50ms |

### 💸 Earnings Calculation

```go
func CalculateEarnings(bandwidthGB, durationMinutes, ratePerGB, qualityScore float64) float64 {
    baseEarnings := bandwidthGB * ratePerGB

    // Quality multiplier: 0.5x to 1.5x
    qualityMultiplier := 0.5 + (qualityScore / 100.0)

    // Duration bonus: encourage stable connections
    durationBonus := 1.0
    if durationMinutes > 60  { durationBonus = 1.1 }
    if durationMinutes > 180 { durationBonus = 1.2 }

    return baseEarnings * qualityMultiplier * durationBonus
}
```

### 💳 Payout Pipeline

```
Session active → trafficMonitor (every 1s)
    │
    ├── Accumulate PendingBandwidthKB per session
    │
    └── Every 10 minutes → flushEarnings()
        │
        ├── RecordEarning() → OperatorEarning (status: "pending")
        │   ├── Calculate quality score
        │   ├── Determine reward tier
        │   └── Calculate USD amount
        │
        ├── ConfirmEarnings() → status: "confirmed"
        │
        └── ProcessPayouts() (when pending_payout >= minimum)
            │
            ├── getCryptoConversion() → exchange rate
            ├── Create OperatorPayout record
            └── executeBlockchainTransaction()
                ├── SendTransaction() → blockchain
                ├── Poll for confirmation (30 attempts, 10s intervals)
                ├── Update payout status → "completed"
                └── Update operator: pending_payout -= amount
```

---

## 🎛️ 9. Control Server

The control server (`internal/control/server.go`) manages the VPN network infrastructure.

### 🔄 Background Tasks

| Task | Interval | Action |
|------|----------|--------|
| `healthCheckLoop` | 1 minute | Check node heartbeats; mark offline if >2min since last heartbeat; mark online if heartbeat resumed |
| `loadBalancerLoop` | 30 seconds | Recalculate `LoadScore` for all online nodes; warn if score > 80 |
| `cleanupLoop` | 1 hour | Delete disconnected sessions >30 days old; delete expired configs; fix orphaned sessions (active session on offline node → disconnected) |

### 📡 Node Selection

Best node is selected by: `ORDER BY load_score ASC, latency ASC` with filters for country and protocol.

---

## 📡 10. VPN Node Service

The VPN node service (`internal/node/service.go`) manages the WireGuard interface and client sessions.

### ⚙️ Service Components

```
Service
├── wgManager          # WireGuard interface operations
├── rewardService      # Earnings recording
├── p2pHost            # libp2p network (optional)
├── activeSessions     # map[uuid.UUID]*SessionInfo
├── trafficMonitor     # Real-time bandwidth tracking
└── Background tasks:
    ├── heartbeatLoop       (30s)  — Update DB + P2P status
    ├── sessionMonitor      (1m)   — Disconnect inactive sessions
    ├── metricsCollector    (15s)  — Update Prometheus metrics
    ├── trafficMonitor      (1s)   — Parse WG stats, update per-session bytes
    └── watchPendingSessions (5s)  — Provision pending, process disconnects
```

### 🔧 WireGuard Interface Setup

```go
ServerConfig{
    PrivateKey: "...",
    Address:    "10.0.0.1/24",
    ListenPort: 51820,
    PostUp: [
        "iptables -A FORWARD -i wg0 -j ACCEPT",
        "iptables -A FORWARD -o wg0 -j ACCEPT",
        "iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE",
    ],
    PostDown: [
        "iptables -D FORWARD -i wg0 -j ACCEPT",
        "iptables -D FORWARD -o wg0 -j ACCEPT",
        "iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE",
    ],
}
```

### 🔄 Session Lifecycle on Node

1. **Pending** — Created by API gateway; node polls for pending sessions every 5s
2. **Active** — Node adds WireGuard peer; allocates tunnel IP; starts traffic tracking
3. **Active (monitoring)** — Traffic stats flushed to DB every 3s; earnings flushed every 10min
4. **Pending Disconnect** — Client requests disconnect via API; node polls every 5s
5. **Disconnected** — Node removes WireGuard peer; final stats flush; update DB

---

## 🛡️ 11. Security Model

### ✅ Input Validation (`pkg/security/validation.go`)

| Validator | Rules |
|-----------|-------|
| Email | Max 254 chars, RFC 5321, valid TLD |
| Username | 3-64 chars, `^[a-zA-Z][a-zA-Z0-9_-]{2,63}$` |
| Password | 6-128 chars, common password blocklist |
| UUID | RFC 4122 format |
| IP | `net.ParseIP()` validation |
| Port | 1-65535 |
| Hostname | Max 253 chars, valid DNS |
| Country Code | 2-char ISO 3166-1 alpha-2, whitelist of 35 countries |
| Protocol | Whitelist: `wireguard`, `openvpn`, `ikev2`, `ipsec` |
| Crypto Currency | Whitelist: `ethereum`, `bitcoin`, `litecoin`, `monero` |
| Wallet Address | Per-currency regex (ETH: `0x` + 40 hex, BTC: legacy/segwit/bech32, LTC: L/M/ltc1) |
| Max Input | 10,000 chars default |

### 🚨 Attack Detection

- **SQL Injection:** Regex patterns for `SELECT`, `INSERT`, `DROP`, `UNION`, `--`, `/*`, `xp_`, `sp_`
- **XSS:** Patterns for `<script>`, `on*=` event handlers, `javascript:`, `<iframe>`, `<object>`

### 🛡️ Privacy Filter (`pkg/security/privacy.go`)

- **IP Anonymization:** `AnonymizeIP(ip)` → SHA256 hash (first 4 bytes), no real IPs stored
- **Sensitive Data Redaction:** Regex-based redaction of IPv4/v6, emails, JWTs, API keys, PEM keys, credit cards, Ethereum addresses
- **Log Sanitization:** All log messages pass through `SanitizeLogMessage()`

### 📋 Data Retention Policy (GDPR)

| Data Type | Retention |
|-----------|-----------|
| Active sessions | While active |
| Completed sessions | 24 hours |
| Security logs | 90 days |
| Access logs | 7 days |
| Error logs | 30 days |
| Inactive users | 1 year |
| Deleted users | 30 days post-deletion |
| Payment records | 7 years (legal) |

### 🔐 Password Security

- **Hashing:** bcrypt (`pkg/crypto`)
- **Common Passwords:** Blocklist of top-15 common passwords
- **High-Security Mode:** 16+ chars, upper + lower + digit + special, max 2 repeating chars

### 🔑 Authentication Security

- **Bearer Token:** `Authorization: Bearer <JWT>`
- **Token Verification:** HMAC-SHA256 signature check, expiration check, signing method validation
- **Refresh Flow:** Mutex-protected to prevent concurrent refresh race conditions
- **Admin Middleware:** Checks `is_admin` claim from JWT

---

## 📊 12. Monitoring & Metrics

### Prometheus Metrics Catalog

All metrics are prefixed with `aureo_vpn_` or `aureo_p2p_`.

#### 🌐 HTTP Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `http_requests_total` | Counter | method, path, status | Total HTTP requests |
| `http_request_duration_seconds` | Histogram | method, path | Request latency |

#### 🔗 VPN Connection Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `active_connections` | Gauge | protocol, node | Current active connections |
| `connections_total` | Counter | protocol, node, status | Total connections |
| `data_transferred_bytes` | Counter | direction, protocol, node | Data transferred |
| `connection_duration_seconds` | Histogram | protocol, node | Connection duration |

#### 📡 Node Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `node_status` | Gauge | node, country, city | 1=online, 0=offline |
| `node_load_score` | Gauge | node | Load score (0-100) |
| `node_cpu_usage_percent` | Gauge | node | CPU usage % |
| `node_memory_usage_percent` | Gauge | node | Memory usage % |
| `node_bandwidth_gbps` | Gauge | node | Bandwidth in Gbps |

#### 🔑 Authentication Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `login_attempts_total` | Counter | status | Login attempts |
| `user_registrations_total` | Counter | - | Total registrations |
| `token_generations_total` | Counter | type | Token generations |

#### 🌐 P2P Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `aureo_p2p_connected_peers` | Gauge | Connected P2P peers |
| `aureo_p2p_known_nodes` | Gauge | Known nodes in registry |

#### 📊 VPN Traffic Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `bytes_received_total` | Counter | Total bytes received |
| `bytes_sent_total` | Counter | Total bytes sent |

#### 🗄️ Database Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `database_queries_total` | Counter | operation, table | Query count |
| `database_query_duration_seconds` | Histogram | operation, table | Query latency |

---

## 📱 13. Mobile App Architecture

### ⚙️ Tech Stack

- **Framework:** Expo 54 + React Native 0.83
- **Language:** TypeScript
- **Routing:** Expo Router (file-based, `app/` directory)
- **State:** Zustand stores (`src/stores/`)
- **API Client:** Axios with JWT interceptors
- **Path Alias:** `@/*` maps to project root
- **Experiments:** React Compiler, typed routes

### 🗂️ Store Architecture

```
src/stores/
├── auth.store.ts       # User auth, tokens, session restore
├── vpn.store.ts        # VPN connection, stats, sessions
├── settings.store.ts   # App settings, API URL
└── profiles.store.ts   # VPN profiles
```

#### 🔑 Auth Store (`useAuthStore`)

- **Secure Storage:** Tokens in `expo-secure-store` (keys: `aureo_access_token`, `aureo_refresh_token`, `aureo_user_data`)
- **Actions:** `login()`, `register()`, `logout()`, `refreshTokens()`, `restoreSession()`
- **Session Restore:** On app launch, reads SecureStore → optimistic auth with cached user → background profile refresh → fallback to token refresh → logout if both fail
- **Circular Dependency Prevention:** Uses `require()` for lazy-loading related stores

#### 🔗 VPN Store (`useVPNStore`)

- **State:** `connectionStatus`, `activeSession`, `connectedNode`, `clientIP`, speeds, bytes, recent connections
- **Key Generation:** `tweetnacl.box.keyPair()` for WireGuard X25519 keys, seeded with `expo-crypto` PRNG
- **Connection Flow:**
  1. Generate WireGuard key pair
  2. `POST /config/generate` with public key
  3. Start native tunnel via `VPNModule.startTunnel(config)`
  4. Poll stats every 3 seconds (native + API fallback)
- **Quick Connect:** `GET /nodes/best` → auto-select protocol → connect

### 📱 Native VPN Module (`src/native/VPNModule.ts`)

Bridge to platform-specific tunnel implementations:

```typescript
VPNModule = {
  startTunnel(config: TunnelConfig): Promise<void>,
  stopTunnel(): Promise<void>,
  getStatus(): Promise<VPNStatus>,
  getStatistics(): Promise<VPNStatistics>,
  onStatusChanged(callback): Subscription,
  onStatsUpdated(callback): Subscription,
}
```

- **iOS:** NetworkExtension packet-tunnel-provider
- **Android:** Android VpnService
- **Native Module Name:** `AureoVPN` (registered on both platforms)
- **Events:** `onVPNStatusChanged`, `onVPNStatsUpdated`
- **Dev Mode:** Graceful fallback when native module unavailable

### 🌐 API Client (`src/api/client.ts`)

- **Base URL:** `https://api.aureovpn.com` (configurable via settings store)
- **Timeout:** 30 seconds
- **Request Interceptor:** Attaches Bearer token, re-resolves base URL per request
- **Response Interceptor:** On 401 → mutex-protected token refresh → retry original request
- **Error Normalization:** All errors wrapped in `ApiError` class with `status`, `code`, `errorKey`
- **Circular Dependency:** Stores accessed via `require()` at call-time

---

## 🖥️ 14. Desktop App Architecture

### ⚙️ Tech Stack

- **Framework:** Wails 2.x (Go backend + embedded web frontend)
- **Backend:** Go
- **Frontend:** HTML/JS/CSS (static, in `frontend/dist/`)
- **Window:** 1024x768, dark theme
- **Build:** Cross-platform (macOS universal, Windows amd64, Linux amd64)

### 🔗 Go<->JS Bridge

The `App` struct methods are automatically exposed to the frontend via Wails bindings:

```go
type App struct {
    ctx        context.Context
    apiClient  *api.Client       // HTTP client for backend API
    vpnManager *vpn.WireGuardManager  // WireGuard tunnel management
    user       *models.User      // Currently logged-in user
    session    *models.Session   // Current VPN session
    nodeID     string            // Connected node ID
    nodeName   string            // Connected node name
    configDir  string            // ~/.aureo-vpn
}
```

**Exposed Methods:**

| Category | Methods |
|----------|---------|
| Auth | `Login`, `Register`, `Logout`, `CheckSavedSession` |
| Session | `SaveSession`, `LoadSession`, `DeleteSession` |
| VPN | `ConnectToVPN`, `DisconnectVPN`, `IsConnected`, `GetVPNStats` |
| Nodes | `GetNodes`, `GetBestNode`, `GetNode`, `GenerateConfig` |
| User | `GetCurrentUser`, `GetUserProfile`, `GetUserStats`, `GetAllSessions` |
| Config | `SetAPIURL`, `GetAPIURL` |

### 🔒 WireGuard Management (`internal/vpn/wireguard.go`)

- **Key Generation:** `wg genkey` + `wg pubkey` commands
- **Config Path:** `~/.aureo-vpn/wg0.conf`
- **Connection (macOS):** `wg-quick up` via `osascript` for admin elevation
- **Disconnection:** `wg-quick down` with admin privileges
- **Stats:** Parses `sudo wg show` output for bytes sent/received and last handshake

### 💾 Session Persistence

Sessions stored in `~/.aureo-vpn/session.json` with 0600 permissions:
```json
{
  "access_token": "...",
  "refresh_token": "...",
  "user": { ... },
  "api_url": "https://api.aureovpn.com"
}
```

### 🎨 Frontend UI

- **Design:** Premium dark theme (Gold #F59E0B, Cyber Blue #3B82F6, Dark #030712)
- **Map:** Leaflet.js with node markers, color-coded load indicators
- **Tabs:** Servers (search + list), Stats, Settings
- **Quick Actions:** Quick Connect, Secure Core, P2P Friendly, Random
- **Real-time:** Speed monitoring every 2s, connection timer every 1s

---

## 🐳 15. Deployment

### 🔑 Environment Variables

```bash
# Server
PORT=8080
HOST=0.0.0.0
TLS_ENABLED=false
TLS_CERT_FILE=/path/to/cert.pem
TLS_KEY_FILE=/path/to/key.pem

# Database (PostgreSQL)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=aureo_vpn
DB_SSLMODE=disable
DB_MAX_IDLE_CONNS=10
DB_MAX_OPEN_CONNS=100

# JWT
JWT_SECRET=minimum-32-character-secret-key-here
JWT_ACCESS_DURATION=15m
JWT_REFRESH_DURATION=168h    # 7 days

# Blockchain
ETHEREUM_RPC_URL=https://eth-mainnet.alchemyapi.io/v2/KEY
ETHEREUM_PRIVATE_KEY=0x...
ETHEREUM_CHAIN_ID=1
BITCOIN_RPC_URL=http://localhost:8332
BITCOIN_RPC_USER=bitcoin
BITCOIN_RPC_PASSWORD=password
LITECOIN_RPC_URL=http://localhost:9332
LITECOIN_RPC_USER=litecoin
LITECOIN_RPC_PASSWORD=password

# Security
CORS_ENABLED=true
CORS_ALLOWED_ORIGINS=https://app.aureovpn.com
RATE_LIMIT_ENABLED=true
RATE_LIMIT_MAX_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# Metrics
METRICS_ENABLED=true
METRICS_PATH=/metrics

# Logging
LOG_LEVEL=info          # debug, info, warn, error
LOG_FORMAT=json         # json, text
LOG_ADD_SOURCE=false
ENVIRONMENT=production  # development, staging, production

# VPN Node
P2P_PORT=4001
P2P_ENABLE_DHT=true
P2P_ENABLE_MDNS=true
DEFAULT_PROTOCOL=wireguard
ENABLE_KILL_SWITCH=true
```

### 🔨 Build Commands

```bash
# Backend
cd aureo-vpn
make build              # Build all 4 binaries
make build-api          # API Gateway only
make docker-build       # Docker images
make docker-up          # docker-compose up

# Mobile
cd aureo-app
npm start               # Expo dev server
npm run ios             # iOS simulator
npm run android         # Android emulator

# Desktop
cd aureo-desktop
make dev                # Dev mode with hot reload
make build              # Current platform
make build-all          # macOS + Windows + Linux
```

### 🗄️ Database Migration Order

Migrations respect foreign key dependencies:

1. `User`, `NodeReward` (independent)
2. `NodeOperator` (depends on User)
3. `VPNNode` (depends on NodeOperator)
4. `Session`, `Config` (depend on User, VPNNode)
5. `OperatorEarning`, `OperatorPayout`, `NodePerformanceMetric` (depend on multiple)

### 🚀 Server Startup Sequence

1. Load config from environment variables
2. Initialize structured logger
3. Connect to database (retry up to 5x with exponential backoff)
4. Run GORM auto-migrations
5. Initialize JWT token service
6. Initialize blockchain service (optional, warns if unavailable)
7. Initialize reward service + seed reward tiers
8. Initialize operator service
9. Create Fiber app with middleware stack
10. Register routes
11. Start server (TLS optional)
12. Wait for SIGINT/SIGTERM → graceful shutdown

---

## 🧪 16. Testing

### 📁 Test Structure

```
aureo-vpn/tests/
├── unit/               # Unit tests
│   ├── auth_test.go    # JWT validation, password hashing
│   └── ...
└── integration/        # Integration tests
    ├── vpn_flow_test.go      # End-to-end connection flow
    ├── earnings_flow_test.go # Operator payout flow
    └── ...
```

### ⚙️ Commands

```bash
# All tests with race detection and coverage
make test

# Unit tests only
make test-unit
# Equivalent: go test -v -race ./tests/unit/...

# Integration tests only
make test-integration
# Equivalent: go test -v -race ./tests/integration/...

# Single test
cd aureo-vpn && go test -v -race ./tests/unit/ -run TestName

# Linting
make lint       # golangci-lint run ./...
make fmt        # go fmt ./...
make vet        # go vet ./...

# Security scanning
make security-scan  # gosec ./...

# Mobile
cd aureo-app && npm run lint  # ESLint (expo lint)
```

---
<div align="center">

Built with ❤️ by the Aureo team

</div>
