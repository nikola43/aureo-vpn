# 📡 VPN Node Service

The VPN Node Service is the individual node daemon that manages WireGuard tunnels, tracks sessions, monitors traffic, earns rewards, and participates in the P2P network.

Source: `internal/node/service.go`

---

## Service Components

```
┌──────────────────────────────────────────────────────────────────┐
│                        VPN Node Service                          │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐    │
│  │                    Core Components                        │    │
│  │                                                          │    │
│  │  wgManager         wireguard.Manager   WireGuard iface   │    │
│  │  rewardService     rewards.RewardService  Earnings calc  │    │
│  │  p2pHost           p2p.Host            P2P networking     │    │
│  │  activeSessions    map[uuid.UUID]*SessionInfo             │    │
│  │  db                *gorm.DB            Database access    │    │
│  └──────────────────────────────────────────────────────────┘    │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐    │
│  │                  Background Tasks                         │    │
│  │                                                          │    │
│  │  heartbeatLoop       30s   Update DB + P2P status        │    │
│  │  sessionMonitor      1min  Disconnect inactive sessions  │    │
│  │  metricsCollector    15s   Load score + Prometheus        │    │
│  │  trafficMonitor      1s    WG stats + earnings accrual   │    │
│  │  watchPendingSessions 5s   Provision + disconnect queue  │    │
│  └──────────────────────────────────────────────────────────┘    │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐    │
│  │                   Traffic Monitor                         │    │
│  │                                                          │    │
│  │  lastBytesSent       int64   Previous total sent          │    │
│  │  lastBytesReceived   int64   Previous total received      │    │
│  │  lastTrafficCheck    time.Time  Last measurement time    │    │
│  │                                                          │    │
│  │  Per-session tracking:                                   │    │
│  │    BytesSent, BytesReceived (cumulative)                 │    │
│  │    LastBytesSent, LastBytesReceived (delta baseline)     │    │
│  │    PendingBandwidthKB (earnings accumulator)             │    │
│  │    LastEarningsFlush (10min interval)                    │    │
│  │    LastStatsFlush (3s interval for client polling)       │    │
│  └──────────────────────────────────────────────────────────┘    │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

---

## WireGuard Interface Setup

When the node starts, it configures the WireGuard interface:

```go
config := wireguard.ServerConfig{
    PrivateKey: privateKey,
    Address:    node.InternalIP + "/24",  // e.g., "10.0.0.1/24"
    ListenPort: node.WireGuardPort,       // default: 51820
    PostUp: []string{
        "iptables -A FORWARD -i wg0 -j ACCEPT",
        "iptables -A FORWARD -o wg0 -j ACCEPT",
        "iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE",
    },
    PostDown: []string{
        "iptables -D FORWARD -i wg0 -j ACCEPT",
        "iptables -D FORWARD -o wg0 -j ACCEPT",
        "iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE",
    },
}
```

The `PostUp` rules enable IP forwarding and NAT masquerading so VPN clients can access the internet through the node. `PostDown` cleans up on shutdown.

---

## Startup Sequence

1. Load node config from database (or auto-register if not found)
2. Generate WireGuard keypair if missing, store in database
3. Setup WireGuard interface (`wg0`) with iptables rules
4. Initialize P2P network (optional, continues without if it fails)
5. Restore active sessions from database (re-add WireGuard peers)
6. Start 5 background goroutines (heartbeat, session monitor, metrics, traffic, pending sessions)
7. Log "VPN Node Service started successfully"

---

## Session Lifecycle on Node

```
┌──────────────────────────────────────────────────────────────────┐
│                     Session States on Node                       │
│                                                                  │
│  ┌─────────┐                                                     │
│  │ Pending  │  API creates session with status="pending"         │
│  │          │  Node polls every 5s for pending sessions          │
│  └────┬────┘                                                     │
│       │                                                          │
│       │ provisionSession():                                      │
│       │   - Check node capacity                                  │
│       │   - Use client public key (or generate server-side)      │
│       │   - Allocate tunnel IP if needed                         │
│       │   - wgManager.AddPeer(publicKey, allowedIPs)             │
│       │   - UPDATE session SET status='active'                   │
│       │   - Add to activeSessions map                            │
│       │                                                          │
│  ┌────▼────┐                                                     │
│  │ Active  │  trafficMonitor reads WG stats every 1s             │
│  │         │  Updates BytesSent, BytesReceived per session       │
│  │         │  Flushes stats to DB every 3s                       │
│  │         │  Flushes earnings every 10min                       │
│  └────┬────┘                                                     │
│       │                                                          │
│  ┌────▼─────────────────┐                                        │
│  │ Active (monitoring)  │  sessionMonitor checks every 1min      │
│  │                      │  If LastKeepalive > 10min:             │
│  │                      │    → auto-disconnect (inactive client) │
│  └────┬─────────────────┘                                        │
│       │                                                          │
│  ┌────▼──────────────┐                                           │
│  │ Pending Disconnect│  Client sends DELETE /sessions/:id        │
│  │                   │  API sets status="pending_disconnect"     │
│  │                   │  Node polls every 5s                      │
│  └────┬──────────────┘                                           │
│       │                                                          │
│       │ disconnectSession():                                     │
│       │   - Flush remaining earnings                             │
│       │   - Flush final traffic stats                            │
│       │   - wgManager.RemovePeer(publicKey)                      │
│       │   - UPDATE session SET status='disconnected'             │
│       │   - Decrement node current_connections                   │
│       │   - Remove from activeSessions map                       │
│       │   - Update Prometheus metrics                            │
│       │                                                          │
│  ┌────▼────────┐                                                 │
│  │Disconnected │  Final state. Cleaned up by control server      │
│  │             │  after 30 days.                                  │
│  └─────────────┘                                                 │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

---

## Graceful Shutdown

1. Cancel context (signals all goroutines to stop)
2. Wait up to 30 seconds for all 5 background goroutines to finish
3. Stop P2P network
4. Disconnect all active sessions (flush earnings, remove WG peers, update DB)
5. Log "VPN Node Service stopped"

---

## Session Restoration

On startup, the node restores active sessions from the database:

1. Query sessions with `node_id = self AND status = 'active'`
2. Get current WireGuard peers via `wg show`
3. For each session:
   - Skip if missing public key or tunnel IP
   - If peer not in WireGuard, re-add it
   - Add to `activeSessions` map with current stats

---

## Auto-Registration

If the node ID does not exist in the database, the node auto-registers itself:

```go
node = models.VPNNode{
    ID:            nodeID,
    Name:          "node-" + nodeID[:8],
    Hostname:      os.Hostname(),
    Country:       "Unknown",
    CountryCode:   "XX",
    City:          "Unknown",
    PublicIP:      "0.0.0.0",
    InternalIP:    "10.0.0.1",
    WireGuardPort: 51820,
    OpenVPNPort:   1194,
    Status:        "online",
    IsActive:      true,
}
```

---

## P2P Integration

If P2P initialization succeeds, the node:

1. Creates a `p2p.Host` with the configured identity
2. Sets local node info (converts database model to P2P `NodeInfo`)
3. Registers callbacks for node join/leave events
4. Every heartbeat (30s), updates the P2P host with current load stats
5. The P2P host handles announce/heartbeat broadcasting automatically

If P2P fails, the node continues operating in database-only mode.
