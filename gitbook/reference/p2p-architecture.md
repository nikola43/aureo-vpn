# 🌐 P2P Architecture

Aureo VPN includes an optional peer-to-peer layer built on **libp2p** for decentralized node discovery. The system can run with or without P2P -- when disabled, node discovery falls back to the central database.

Source: `pkg/p2p/`

---

## Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         P2P Host                                │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │ Kademlia DHT │  │  GossipSub   │  │       mDNS           │  │
│  │              │  │  (PubSub)    │  │  (Local Discovery)   │  │
│  │  - Peer      │  │              │  │                      │  │
│  │    routing   │  │  - Announce  │  │  Service name:       │  │
│  │  - Content   │  │    topic     │  │  "aureo-vpn"         │  │
│  │    discovery │  │  - Heartbeat │  │                      │  │
│  │  - Server    │  │    topic     │  │  Auto-connects       │  │
│  │    mode      │  │              │  │  LAN peers           │  │
│  └──────────────┘  └──────────────┘  └──────────────────────┘  │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                     Registry                             │   │
│  │  map[uuid.UUID]*NodeInfo   (max 1000 nodes)             │   │
│  │  map[peer.ID]uuid.UUID     (peer ID index)              │   │
│  │  Node timeout: 2 minutes                                │   │
│  │  Stale cleanup: 3x timeout (6 minutes)                  │   │
│  │  Eviction: oldest node when at capacity                 │   │
│  │  Persistence: ./data/p2p/nodes.json                     │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌──────────────┐  ┌──────────────────────────────────────┐    │
│  │  Ed25519     │  │         Transport                    │    │
│  │  Identity    │  │  /ip4/0.0.0.0/tcp/4001              │    │
│  │  Key         │  │  /ip4/0.0.0.0/udp/4001/quic-v1      │    │
│  │              │  │  + NAT port mapping                  │    │
│  │  base64      │  │  + Relay + Hole punching             │    │
│  │  encoded     │  │                                      │    │
│  └──────────────┘  └──────────────────────────────────────┘    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## Protocol IDs

Direct peer-to-peer request/response protocols:

| Protocol ID | Description | Handler |
|---|---|---|
| `/aureo/nodeinfo/1.0.0` | Request a peer's node info | Returns local `NodeInfo` as JSON |
| `/aureo/nodelist/1.0.0` | Request known active nodes | Returns `[]*NodeInfo` as JSON |
| `/aureo/health/1.0.0` | Health check | Returns `{"status":"ok","timestamp":...}` |

---

## PubSub Topics

GossipSub topics for network-wide broadcast:

| Topic | Interval | Message Type | Purpose |
|---|---|---|---|
| `/aureo/nodes/announce/1.0.0` | 5 minutes | `AnnounceMessage` | Full node info broadcast (join/update) |
| `/aureo/nodes/heartbeat/1.0.0` | 30 seconds | `HeartbeatMessage` | Lightweight status update |

---

## Message Types

### AnnounceMessage

Broadcast when a node joins the network or periodically to share full node info:

```json
{
  "node": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "peer_id": "12D3KooWExample...",
    "name": "node-us-east-1",
    "version": "1.0.0",
    "public_ip": "203.0.113.42",
    "multiaddrs": [
      "/ip4/203.0.113.42/tcp/4001/p2p/12D3KooWExample..."
    ],
    "wireguard_port": 51820,
    "openvpn_port": 1194,
    "public_key": "WG_PUBLIC_KEY_BASE64...",
    "country": "United States",
    "country_code": "US",
    "city": "New York",
    "latitude": 40.7128,
    "longitude": -74.0060,
    "max_connections": 1000,
    "current_connections": 42,
    "load_score": 15.3,
    "cpu_usage": 23.5,
    "memory_usage": 45.2,
    "bandwidth_gbps": 0.85,
    "status": "online",
    "last_heartbeat": "2025-01-15T10:30:00Z",
    "uptime_percentage": 99.95,
    "supports_wireguard": true,
    "supports_openvpn": true,
    "supports_multihop": false,
    "supports_obfuscation": false,
    "is_operator_owned": true,
    "reputation": 87.5,
    "stake_amount": 500.0
  },
  "timestamp": 1705312200000000000
}
```

### HeartbeatMessage

Lightweight status broadcast every 30 seconds:

```json
{
  "node_id": "550e8400-e29b-41d4-a716-446655440000",
  "peer_id": "12D3KooWExample...",
  "status": "online",
  "current_connections": 42,
  "load_score": 15.3,
  "cpu_usage": 23.5,
  "memory_usage": 45.2,
  "bandwidth_gbps": 0.85,
  "timestamp": 1705312200000000000
}
```

---

## Default Configuration

```go
func DefaultConfig() Config {
    return Config{
        ListenAddrs: []string{
            "/ip4/0.0.0.0/tcp/4001",
            "/ip4/0.0.0.0/udp/4001/quic-v1",
        },
        EnableDHT:         true,
        DHTServerMode:     true,
        EnableMDNS:        true,
        EnablePubSub:      true,
        HeartbeatInterval: 30 * time.Second,
        AnnounceInterval:  5 * time.Minute,
        NodeTimeout:       2 * time.Minute,
        MaxPeers:          100,
        MaxNodes:          1000,
        AnnounceTopic:     "/aureo/nodes/announce/1.0.0",
        HeartbeatTopic:    "/aureo/nodes/heartbeat/1.0.0",
        DataDir:           "./data/p2p",
    }
}
```

---

## Background Loops

| Loop | Interval | Description |
|---|---|---|
| `heartbeatLoop` | 30 seconds | Broadcasts `HeartbeatMessage` with current node stats via GossipSub |
| `announceLoop` | 5 minutes | Broadcasts full `AnnounceMessage` with complete node info. Also fires once on startup after 2s delay |
| `discoveryLoop` | 30 seconds | Uses DHT `RoutingDiscovery` to find new peers. Advertises self under `"aureo-vpn"` namespace. On discovery, requests node info via `/aureo/nodeinfo/1.0.0` protocol |
| `cleanupLoop` | 5 minutes | Marks nodes as offline if heartbeat is stale (> 2 min). Removes nodes with no heartbeat for 3x timeout (> 6 min). Persists registry to disk |

---

## Node Discovery Flow

```
1. Node starts → creates libp2p host (Ed25519 identity)
2. Bootstraps DHT with known peers (if configured)
3. Enables mDNS for LAN discovery (service: "aureo-vpn")
4. Joins GossipSub announce + heartbeat topics
5. Broadcasts initial AnnounceMessage (after 2s delay)
6. discoveryLoop finds peers via DHT RoutingDiscovery
7. On new peer: opens /aureo/nodeinfo/1.0.0 stream → gets NodeInfo
8. NodeInfo added to Registry (thread-safe, sorted by load)
9. Incoming heartbeats update Registry in real-time
10. API queries Registry for node selection (sorted by LoadScore ASC)
```

---

## Registry

The registry is an in-memory store with thread-safe access:

- **Capacity:** Max 1000 nodes (configurable). Evicts oldest node when full
- **Indexes:** By `uuid.UUID` and by `peer.ID`
- **Timeout:** Nodes marked offline after 2 minutes without heartbeat
- **Cleanup:** Nodes removed entirely after 6 minutes (3x timeout)
- **Persistence:** Saved to `./data/p2p/nodes.json` on cleanup and shutdown
- **Query:** Filter by protocol and country code, sorted by `LoadScore ASC`

---

## API Integration

The API gateway can optionally connect to the P2P network as a **client** (not a VPN node):

```
GET /api/v1/nodes?source=auto     → Try P2P first, fall back to database
GET /api/v1/nodes?source=p2p      → P2P only
GET /api/v1/nodes?source=db       → Database only
GET /api/v1/nodes/best?protocol=wireguard&country=US

GET /api/v1/p2p/status            → P2P network stats
GET /api/v1/p2p/countries         → Countries with active P2P nodes
```
