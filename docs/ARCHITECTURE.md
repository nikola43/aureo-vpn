# Aureo VPN Architecture

## Overview

Aureo VPN is a production-ready, enterprise-grade VPN service designed for scalability, security, and high availability. The architecture follows microservices principles with clear separation of concerns.

## System Components

### 1. API Gateway
**Purpose**: Single entry point for all client requests

**Responsibilities**:
- User authentication and authorization (JWT)
- Rate limiting and request throttling
- Request routing to appropriate services
- API versioning
- Metrics collection

**Technology Stack**:
- Go Fiber (HTTP framework)
- PostgreSQL (data persistence)
- Redis (caching & rate limiting)
- Prometheus (metrics)

**Endpoints**:
```
/api/v1/auth/*        - Authentication endpoints
/api/v1/user/*        - User management
/api/v1/nodes/*       - VPN node discovery
/api/v1/admin/*       - Admin operations
/health               - Health check
/metrics              - Prometheus metrics
```

### 2. Control Server
**Purpose**: Orchestrates the VPN infrastructure

**Responsibilities**:
- Node health monitoring
- Load balancing across nodes
- Session management
- Automatic failover
- Resource cleanup
- Statistics aggregation

**Background Tasks**:
- Health check loop (every 1 minute)
- Load balancer loop (every 30 seconds)
- Cleanup loop (every 1 hour)

### 3. VPN Node Service
**Purpose**: Handles actual VPN connections

**Responsibilities**:
- Tunnel creation and management
- Protocol implementation (WireGuard/OpenVPN)
- Traffic encryption/decryption
- Connection monitoring
- Resource tracking

**Supported Protocols**:
- **WireGuard**: Modern, fast, secure
- **OpenVPN**: Traditional, widely compatible

**Background Tasks**:
- Heartbeat to control server (every 30 seconds)
- Session monitoring (every 1 minute)
- Metrics collection (every 15 seconds)

### 4. CLI Tool
**Purpose**: Administrative command-line interface

**Commands**:
```bash
aureo-vpn node create    - Create VPN node
aureo-vpn node list      - List all nodes
aureo-vpn node delete    - Delete node
aureo-vpn config generate - Generate client config
aureo-vpn user list      - List users
aureo-vpn stats          - View statistics
```

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         Clients                             │
│  (Desktop, Mobile, Router - WireGuard/OpenVPN)             │
└────────────────────┬────────────────────────────────────────┘
                     │
                     │ HTTPS/TLS
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                     Load Balancer                           │
│                   (NGINX / AWS ALB)                         │
└────────────────────┬────────────────────────────────────────┘
                     │
        ┌────────────┴────────────┬──────────────┐
        ▼                         ▼              ▼
┌───────────────┐        ┌───────────────┐  ┌──────────────┐
│ API Gateway 1 │        │ API Gateway 2 │  │API Gateway N │
└───────┬───────┘        └───────┬───────┘  └──────┬───────┘
        │                        │                  │
        └────────────┬───────────┴──────────────────┘
                     │
        ┌────────────┴────────────┬─────────────┐
        ▼                         ▼             ▼
┌───────────────┐        ┌───────────────┐  ┌────────────┐
│ PostgreSQL    │        │  Redis Cache  │  │ Prometheus │
│   (Primary)   │        │               │  │  Metrics   │
└───────┬───────┘        └───────────────┘  └────────────┘
        │
        │ Replication
        ▼
┌───────────────┐
│ PostgreSQL    │
│  (Replicas)   │
└───────────────┘


┌─────────────────────────────────────────────────────────────┐
│                   Control Server Cluster                     │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                  │
│  │Control-1 │  │Control-2 │  │Control-N │                  │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘                  │
│       │             │             │                          │
│       └─────────────┴─────────────┘                          │
└────────────────────┬────────────────────────────────────────┘
                     │
        ┌────────────┴────────────┬──────────────┐
        ▼                         ▼              ▼
┌───────────────┐        ┌───────────────┐  ┌──────────────┐
│  VPN Node 1   │        │  VPN Node 2   │  │ VPN Node N   │
│   (US-East)   │        │   (EU-West)   │  │ (AP-South)   │
│               │        │               │  │              │
│ ┌───────────┐ │        │ ┌───────────┐ │  │┌───────────┐│
│ │ WireGuard │ │        │ │ WireGuard │ │  ││ WireGuard ││
│ └───────────┘ │        │ └───────────┘ │  │└───────────┘│
│ ┌───────────┐ │        │ ┌───────────┐ │  │┌───────────┐│
│ │  OpenVPN  │ │        │ │  OpenVPN  │ │  ││  OpenVPN  ││
│ └───────────┘ │        │ └───────────┘ │  │└───────────┘│
└───────────────┘        └───────────────┘  └──────────────┘
```

## Data Flow

### 1. User Registration/Login
```
Client → API Gateway → Auth Service → Database
                                   ↓
                              JWT Tokens
                                   ↓
Client ← API Gateway ← Auth Service
```

### 2. VPN Connection Establishment
```
1. Client requests best node
   Client → API Gateway → Control Server → Database
                                        ↓
                                   Node Selection
                                        ↓
   Client ← API Gateway ← Control Server

2. Client connects to VPN node
   Client → VPN Node → Session Creation → Database
                    ↓
             Tunnel Setup (WireGuard/OpenVPN)
                    ↓
   Client ← VPN Node (Encrypted connection established)

3. Heartbeat & monitoring
   VPN Node → Control Server (every 30s)
           ↓
      Update Status
```

### 3. Traffic Flow
```
Client Traffic → VPN Tunnel → VPN Node → Internet
                    ↑
              (Encrypted)

Kill Switch ensures if tunnel drops:
Client Traffic → BLOCKED (no leak)
```

## Security Architecture

### Authentication & Authorization
```
┌──────────────┐
│    Client    │
└──────┬───────┘
       │ POST /auth/login
       │ {email, password}
       ▼
┌─────────────────┐
│  API Gateway    │
│                 │
│ ┌─────────────┐ │
│ │  Argon2id   │ │ ← Password hashing
│ │  Hasher     │ │
│ └─────────────┘ │
│                 │
│ ┌─────────────┐ │
│ │JWT Generator│ │ ← Token generation
│ │  (HS256)    │ │
│ └─────────────┘ │
└────────┬────────┘
         │
         │ Access Token (15 min)
         │ Refresh Token (7 days)
         ▼
   ┌──────────┐
   │  Client  │
   └──────────┘
```

### Encryption Layers

**Layer 1: Transport (TLS 1.3)**
```
Client ←─── HTTPS/TLS 1.3 ───→ API Gateway
         (Control Plane)
```

**Layer 2: VPN Tunnel**
```
WireGuard: ChaCha20-Poly1305 or AES-256-GCM
OpenVPN:   AES-256-GCM with SHA256 HMAC
```

**Layer 3: Data at Rest**
```
Database: Encrypted columns (passwords, keys)
Backups:  Full disk encryption
```

### Security Features

1. **Kill Switch**
   - Blocks all non-VPN traffic
   - Prevents IP leaks on disconnect
   - iptables-based implementation

2. **DNS Leak Protection**
   - Forces DNS through VPN
   - Blocks non-VPN DNS queries
   - Custom DNS servers (1.1.1.1, 1.0.0.1)

3. **Split Tunneling**
   - Route specific apps/IPs through VPN
   - Exclude certain traffic from VPN
   - Custom routing tables

4. **IPv6 Leak Prevention**
   - Routes IPv6 through tunnel
   - Blocks IPv6 if not supported

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    subscription_tier VARCHAR(20),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### VPN Nodes Table
```sql
CREATE TABLE vpn_nodes (
    id UUID PRIMARY KEY,
    name VARCHAR(100) UNIQUE,
    public_ip VARCHAR(45),
    country VARCHAR(100),
    city VARCHAR(100),
    max_connections INTEGER,
    current_connections INTEGER,
    load_score FLOAT,
    status VARCHAR(20),
    last_heartbeat TIMESTAMP,
    created_at TIMESTAMP
);
```

### Sessions Table
```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    node_id UUID REFERENCES vpn_nodes(id),
    protocol VARCHAR(20),
    tunnel_ip VARCHAR(45),
    status VARCHAR(20),
    connected_at TIMESTAMP,
    disconnected_at TIMESTAMP,
    bytes_sent BIGINT,
    bytes_received BIGINT
);
```

## Scalability

### Horizontal Scaling

**API Gateway**:
- Stateless design
- Scale to N instances
- Load balanced

**VPN Nodes**:
- Independent instances
- Geographic distribution
- Auto-scaling based on load

**Control Server**:
- Multiple instances for HA
- Leader election for tasks
- Shared state in database

### Vertical Scaling

**Database**:
- Connection pooling
- Read replicas
- Query optimization
- Partitioning by date

**VPN Nodes**:
- Multi-core processing
- Hardware acceleration
- Kernel tuning

## Monitoring & Observability

### Metrics Collected

**Node Metrics**:
- `aureo_vpn_active_connections` - Active connections per node
- `aureo_vpn_node_load_score` - Load score (0-100)
- `aureo_vpn_node_cpu_usage` - CPU utilization
- `aureo_vpn_data_transferred_bytes` - Data transfer

**API Metrics**:
- `aureo_vpn_http_requests_total` - Request count
- `aureo_vpn_http_request_duration` - Latency
- `aureo_vpn_login_attempts_total` - Auth attempts

**Database Metrics**:
- `aureo_vpn_database_queries_total` - Query count
- `aureo_vpn_database_query_duration` - Query time

### Logging

**Structured Logging**:
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "info",
  "service": "api-gateway",
  "message": "User login successful",
  "user_id": "uuid",
  "ip": "1.2.3.4"
}
```

## Disaster Recovery

### Backup Strategy
- **Database**: Daily full backups, hourly incrementals
- **Config**: Version controlled in Git
- **Logs**: 30-day retention

### Recovery Procedures
1. Database restore from backup
2. Redeploy services from Docker images
3. DNS failover to backup region
4. Session migration to healthy nodes

## Future Enhancements

1. **Multi-Hop VPN**: Route through 2+ nodes
2. **Obfuscation**: Disguise VPN traffic
3. **P2P Support**: Port forwarding
4. **Dedicated IPs**: Static IPs for users
5. **WebRTC Protection**: Prevent WebRTC leaks
6. **Tor Integration**: Tor over VPN
7. **SOCKS5 Proxy**: Additional proxy layer
8. **Mobile SDKs**: Native iOS/Android support

## Performance Targets

| Metric | Target | Measured |
|--------|--------|----------|
| API Latency (p95) | < 50ms | TBD |
| VPN Throughput | 1-10 Gbps | Hardware dependent |
| Connection Setup | < 2s | TBD |
| Node Capacity | 1000+ concurrent | TBD |
| Database Queries | < 10ms | TBD |
| Uptime | 99.9% | TBD |

## Technology Choices Rationale

**Go**: High performance, great concurrency, native compilation

**WireGuard**: Modern, fast, secure, built into Linux kernel

**PostgreSQL**: Robust, ACID compliant, JSON support

**Prometheus**: Industry standard, pull-based, flexible

**Docker/K8s**: Portable, scalable, industry standard

**Fiber**: Fast HTTP framework, Express-like API

**JWT**: Stateless auth, scalable, standard

## Compliance & Standards

- **GDPR**: User data protection, right to deletion
- **SOC 2**: Security controls and monitoring
- **ISO 27001**: Information security management
- **No-logs Policy**: No connection logs stored
