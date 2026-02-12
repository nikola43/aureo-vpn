# 🌐 P2P Network

Aureo uses libp2p for decentralized node discovery, eliminating single points of failure.

---

## Discovery Mechanisms

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
