# 🔭 Backend Overview

The Aureo VPN backend is built with Go 1.24 and provides all server-side infrastructure for the decentralized VPN platform.

---

## Components

The backend consists of four independent binaries:

- **API Gateway** — REST API server handling user authentication, node discovery, and session management
- **Control Server** — Orchestrates nodes, manages the network, and processes payments
- **VPN Nodes** — Client-facing servers running WireGuard/OpenVPN
- **CLI Tool** — Command-line interface for administration
- **P2P Network** — Decentralized node discovery using libp2p

{% hint style="info" %}
**Highlights:** Multi-protocol VPN (WireGuard + OpenVPN) · Decentralized P2P discovery · Crypto rewards for node operators · Zero-log privacy architecture
{% endhint %}

## Key Features

- **Multi-Protocol**: Support for both WireGuard (modern, fast) and OpenVPN (compatible)
- **Decentralized**: P2P network for node discovery without single point of failure
- **Crypto Rewards**: Node operators earn ETH, BTC, or LTC based on bandwidth served
- **Reputation System**: Quality-based rewards with tier progression
- **Privacy-Focused**: Kill switch, DNS leak protection, multi-hop routing

## Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                       AUREO VPN PLATFORM                              │
├──────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐            │
│  │   Desktop   │     │   Mobile    │     │    Web      │            │
│  │   Client    │     │   Client    │     │  Dashboard  │            │
│  └──────┬──────┘     └──────┬──────┘     └──────┬──────┘            │
│         └───────────────────┼───────────────────┘                    │
│                             ▼                                        │
│  ┌──────────────────────────────────────────────────────────────┐    │
│  │                     API GATEWAY (Fiber v2)                    │    │
│  │  Auth Handler | Nodes Handler | Sessions Handler | Operator   │    │
│  └──────────────────────────┬───────────────────────────────────┘    │
│              ┌───────────────┼──────────────┐                        │
│              ▼               ▼              ▼                        │
│  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐            │
│  │    SQLite     │  │    Control    │  │  Blockchain   │            │
│  │   Database    │  │    Server     │  │   Service     │            │
│  └───────────────┘  └───────┬───────┘  └───────────────┘            │
│                              ▼                                       │
│  ┌──────────────────────────────────────────────────────────────┐    │
│  │                   P2P NETWORK (libp2p)                        │    │
│  │  Kademlia DHT | Gossipsub | mDNS                              │    │
│  └──────────────────────────────────────────────────────────────┘    │
│                                                                       │
│  ┌──────────────────────────────────────────────────────────────┐    │
│  │  VPN NODES: Node US-1 (WG) | Node EU-1 (OVPN) | Node AP-1   │    │
│  └──────────────────────────────────────────────────────────────┘    │
│                                                                       │
│  ┌──────────────────────────────────────────────────────────────┐    │
│  │  BLOCKCHAIN: Ethereum | Bitcoin | Litecoin                    │    │
│  └──────────────────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────────────┘
```

| Component | Description | Port |
|-----------|-------------|------|
| API Gateway | REST API server (Fiber) | 8080 |
| Control Server | Network orchestrator | 8081 |
| VPN Node | WireGuard/OpenVPN server | 51820/UDP, 1194/UDP |
| SQLite | Embedded database | — (file-based) |
| Prometheus | Metrics collection | 9090 |
| Grafana | Metrics visualization | 3000 |
