# 🔗 VPN Connection Flow

This page documents the complete lifecycle of a VPN connection, from the client generating keys to the node tearing down the tunnel.

---

## Client-Side Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         CLIENT APPLICATION                              │
│                                                                         │
│  1. Generate WireGuard key pair locally                                 │
│     ┌──────────────────────────────────┐                                │
│     │ wg genkey | tee privatekey       │                                │
│     │ cat privatekey | wg pubkey       │                                │
│     └──────────────────────────────────┘                                │
│                           │                                             │
│  2. POST /api/v1/config/generate                                        │
│     { node_id, public_key }                                             │
│                           │                                             │
│  3. Receive configuration                                               │
│     ┌──────────────────────────────────┐                                │
│     │ {                                │                                │
│     │   session_id,                    │                                │
│     │   server_public_key,             │                                │
│     │   server_endpoint: "1.2.3.4:51820",                              │
│     │   client_ip: "10.0.0.5",         │                                │
│     │   dns: "1.1.1.1,8.8.8.8",       │                                │
│     │   allowed_ips: "0.0.0.0/0",      │                                │
│     │   keepalive: 25                  │                                │
│     │ }                                │                                │
│     └──────────────────────────────────┘                                │
│                           │                                             │
│  4. Write WireGuard config file                                         │
│     [Interface]                                                         │
│     PrivateKey = <client_private_key>                                   │
│     Address = 10.0.0.5/24                                               │
│     DNS = 1.1.1.1,8.8.8.8                                              │
│     [Peer]                                                              │
│     PublicKey = <server_public_key>                                     │
│     Endpoint = 1.2.3.4:51820                                           │
│     AllowedIPs = 0.0.0.0/0                                             │
│     PersistentKeepalive = 25                                            │
│                           │                                             │
│  5. Activate tunnel (wg-quick up / NetworkExtension)                    │
│                           │                                             │
│  6. Poll GET /api/v1/sessions/:id for traffic stats                     │
│     { bytes_sent, bytes_received, status }                              │
│                           │                                             │
│  7. Disconnect: DELETE /api/v1/sessions/:id                             │
│     Sets status to "pending_disconnect"                                 │
│     Node picks it up and tears down peer                                │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

**Step-by-step:**

| Step | Action | Endpoint |
|---|---|---|
| 1 | Generate WireGuard keypair on client | Local |
| 2 | Request VPN config from API | `POST /api/v1/config/generate` |
| 3 | Receive server details and allocated IP | Response |
| 4 | Build WireGuard config file locally | Local |
| 5 | Activate WireGuard tunnel | OS-level |
| 6 | Poll session for traffic stats | `GET /api/v1/sessions/:id` |
| 7 | Request disconnection | `DELETE /api/v1/sessions/:id` |

---

## Server-Side Flow

The VPN node service (`internal/node/service.go`) runs several background loops that manage the server side of connections:

```
┌──────────────────────────────────────────────────────────────────────┐
│                        VPN NODE SERVICE                              │
│                                                                      │
│  ┌─────────────────────┐                                             │
│  │ watchPendingSessions │ (every 5s)                                  │
│  │  ├─ processPendingSessions()                                      │
│  │  │   Query: status = "pending" AND node_id = self                 │
│  │  │   For each:                                                    │
│  │  │     provisionSession()                                         │
│  │  │       ├─ Check node capacity                                   │
│  │  │       ├─ Use client's public key (or generate if legacy)       │
│  │  │       ├─ Allocate tunnel IP (if not pre-allocated)             │
│  │  │       ├─ wgManager.AddPeer(publicKey, allowedIPs)              │
│  │  │       ├─ UPDATE session SET status='active'                    │
│  │  │       └─ Track in activeSessions map                           │
│  │  │                                                                │
│  │  └─ processDisconnects()                                          │
│  │      Query: status = "pending_disconnect" AND node_id = self      │
│  │      For each:                                                    │
│  │        DisconnectSession()                                        │
│  │          ├─ Flush remaining earnings                              │
│  │          ├─ Flush final traffic stats to DB                       │
│  │          ├─ wgManager.RemovePeer(publicKey)                       │
│  │          ├─ UPDATE session SET status='disconnected'              │
│  │          ├─ Decrement node current_connections                    │
│  │          └─ Remove from activeSessions map                        │
│  └─────────────────────┘                                             │
│                                                                      │
│  ┌─────────────────────┐                                             │
│  │   trafficMonitor    │ (every 1s)                                   │
│  │  ├─ wgManager.GetInterfaceStats()                                 │
│  │  ├─ For each active session:                                      │
│  │  │   ├─ Update BytesSent, BytesReceived from WG peer stats        │
│  │  │   ├─ Update LastKeepalive from WG handshake                    │
│  │  │   ├─ Accumulate PendingBandwidthKB                             │
│  │  │   ├─ Flush session stats to DB every 3s                        │
│  │  │   └─ Flush earnings every 10min if PendingBandwidthKB > 0     │
│  │  └─ Update node bandwidth_usage_gbps and total_bandwidth_kb      │
│  └─────────────────────┘                                             │
│                                                                      │
│  ┌─────────────────────┐                                             │
│  │   sessionMonitor    │ (every 1min)                                 │
│  │  └─ For each active session:                                      │
│  │      If LastKeepalive > 10 minutes ago:                           │
│  │        disconnectSession() (inactive client)                      │
│  └─────────────────────┘                                             │
│                                                                      │
│  ┌─────────────────────┐                                             │
│  │   heartbeatLoop     │ (every 30s)                                  │
│  │  ├─ Count active WG peers (wg show wg0 peers)                     │
│  │  ├─ UPDATE vpn_nodes SET last_heartbeat, status='online',         │
│  │  │        current_connections = peerCount                         │
│  │  ├─ Update P2P host with current stats                            │
│  │  └─ Update operator stats (pending payout, etc.)                  │
│  └─────────────────────┘                                             │
│                                                                      │
│  ┌─────────────────────┐                                             │
│  │  metricsCollector   │ (every 15s)                                  │
│  │  ├─ Recalculate LoadScore                                         │
│  │  └─ Update Prometheus gauges (status, load, CPU, memory, BW)      │
│  └─────────────────────┘                                             │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

---

## IP Allocation

The node uses a `/24` subnet for its WireGuard interface. The node itself takes `.1`, and clients are assigned `.2` through `.254`.

```go
func allocateClientIP(nodeIP string, usedIPs []string) string {
    // Parse node IP to get the base network
    // Assume /24 subnet, node uses .1, clients get .2-.254
    parts := strings.Split(nodeIP, ".")
    if len(parts) != 4 {
        return ""
    }

    base := strings.Join(parts[:3], ".")
    usedSet := make(map[string]bool)
    for _, ip := range usedIPs {
        usedSet[ip] = true
    }

    // Find first available IP (skip .1 which is the node)
    for i := 2; i <= 254; i++ {
        candidateIP := fmt.Sprintf("%s.%d", base, i)
        if !usedSet[candidateIP] {
            return candidateIP
        }
    }

    return ""
}
```

**Example:**

```
Node InternalIP: 10.0.0.1
Subnet:          10.0.0.0/24
Node address:    10.0.0.1 (gateway, NAT via iptables MASQUERADE)
Client 1:        10.0.0.2
Client 2:        10.0.0.3
...
Client 253:      10.0.0.254
```

Used IPs are queried from sessions with `status IN ('active', 'pending')` to avoid allocation conflicts.

---

## Privacy Protection

Client IP addresses are **never stored** in the database. Instead, a SHA-256 hash with a random salt is used for correlation purposes:

```go
session.ClientIP = privacyFilter.AnonymizeIP(c.IP())
// Produces: "ip:a1b2c3d4e5f6a7b8" (hash prefix, not the real IP)
```

This allows session correlation without revealing the user's real IP address.

---

## Session State Machine

```
                  ┌───────────┐
     POST /config │           │
     /generate    │  pending  │
    ─────────────>│           │
                  └─────┬─────┘
                        │
          provisionSession() on node
          (add WG peer, allocate IP)
                        │
                  ┌─────▼─────┐
                  │           │
                  │  active   │◄──── traffic flowing,
                  │           │      stats updated every 1-3s
                  └─────┬─────┘
                        │
           ┌────────────┼────────────┐
           │            │            │
    DELETE /sessions/:id  10min      admin/error
    (client request)    inactivity   force kill
           │            │            │
    ┌──────▼──────┐     │     ┌──────▼──────┐
    │  pending_   │     │     │             │
    │ disconnect  │     │     │ terminated  │
    └──────┬──────┘     │     │             │
           │            │     └─────────────┘
    processDisconnects()│
    (node picks up)     │
           │            │
    ┌──────▼────────────▼─┐
    │                     │
    │    disconnected     │
    │                     │
    └─────────────────────┘
```
