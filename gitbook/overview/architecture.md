# 🏗️ Architecture

Aureo VPN is built as a three-layer system: clients at the top, backend services in the middle, and a blockchain payment layer at the bottom.

---

## System Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                           CLIENTS                                    │
│                                                                      │
│   📱 Mobile App          🖥️ Desktop App          ⌨️ CLI Tool         │
│   (Expo / React Native)  (Go / Wails 2.x)       (Go / Cobra)       │
│                                                                      │
└───────────────┬──────────────────┬──────────────────┬────────────────┘
                │                  │                  │
                │   HTTPS / REST   │                  │
                ▼                  ▼                  ▼
┌──────────────────────────────────────────────────────────────────────┐
│                       BACKEND SERVICES                                │
│                                                                       │
│   🔑 API Gateway     →     🎛️ Control Server     →     📡 VPN Nodes  │
│   (Fiber v2)               (Orchestrator)              (WireGuard)   │
│                                                                       │
│   🌐 P2P Network (libp2p)  ←→  Kademlia DHT + GossipSub            │
│                                                                       │
└───────────────────────────────┬───────────────────────────────────────┘
                                │
                                ▼
┌──────────────────────────────────────────────────────────────────────┐
│                      ⛓️ BLOCKCHAIN LAYER                              │
│                                                                       │
│         Ethereum (ETH)    ·    Bitcoin (BTC)    ·    Litecoin (LTC)   │
│                     Transparent crypto payouts                        │
└──────────────────────────────────────────────────────────────────────┘
```

---

## Monorepo Structure

```
aureo/
├── 🔧 aureo-vpn/          Backend infrastructure (Go 1.24, Fiber, GORM, libp2p)
│   ├── cmd/                4 binaries: api-gateway, control-server, vpn-node, cli
│   ├── pkg/                Shared packages (auth, blockchain, p2p, protocols)
│   └── internal/           Private packages (handlers, control, node)
│
├── 📱 aureo-app/           Mobile client (Expo 54, React Native, TypeScript)
│   ├── app/                File-based routing (Expo Router)
│   ├── src/                Stores, API client, hooks, native modules
│   └── plugins/            Native VPN plugins (iOS NetworkExtension, Android VpnService)
│
├── 🖥️ aureo-desktop/       Desktop client (Go + Wails 2.x)
│   ├── app.go              Wails App — Go methods exposed to JS
│   ├── internal/           API client + WireGuard manager
│   └── frontend/           Embedded web UI (Leaflet.js map)
│
└── 📖 docs/                Whitepaper, technical documentation
```

---

## Component Overview

| Component | Description | Default Port |
|-----------|-------------|-------------|
| API Gateway | REST API server (Fiber v2) — user-facing | 8080 |
| Control Server | Network orchestrator | 8081 |
| VPN Node | WireGuard/OpenVPN server (requires sudo) | 51820 (WG), 1194 (OVPN) |
| CLI | Management tool (Cobra) | — |
| SQLite | Embedded database | — (file-based) |
| Prometheus | Metrics collection | 9090 |
| Grafana | Metrics visualization | 3000 |
