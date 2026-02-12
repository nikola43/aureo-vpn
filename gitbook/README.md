# 👑 Aureo VPN

### The Decentralized Privacy Network with Crypto Rewards

*A VPN owned by everyone, controlled by no one.*

---

Aureo VPN reimagines the VPN from the ground up. By combining military-grade WireGuard encryption with a decentralized node network powered by blockchain incentives, Aureo creates a VPN ecosystem where no single entity controls the infrastructure. Community operators around the world run nodes and earn cryptocurrency rewards, while users enjoy faster, more private, and more resilient connectivity.

---

## Highlights

| | Feature | Description |
|---|---|---|
| 🛡️ | **Military-Grade Encryption** | WireGuard with ChaCha20-Poly1305 — the most modern, audited VPN protocol available |
| 🌐 | **Fully Decentralized** | P2P node discovery via Kademlia DHT — no central servers, no single point of failure |
| 💰 | **Crypto Rewards** | Node operators earn ETH, BTC, or LTC for every GB of bandwidth they serve |
| 📱 | **Cross-Platform** | Native apps for iOS, Android, macOS, Windows, and Linux — plus a CLI for power users |
| 🔒 | **Zero-Log Architecture** | IP anonymization at the protocol level — your real IP never touches a database |
| ⚡ | **Lightning Fast** | WireGuard's minimal 4,000-line codebase delivers the highest throughput of any VPN protocol |

---

## Quick Links

- [Introduction](overview/introduction.md) — Learn what Aureo VPN is and why it exists
- [Architecture](overview/architecture.md) — Understand the 3-layer system design
- [Whitepaper](whitepaper/executive-summary.md) — Read the full vision and roadmap
- [Backend Quick Start](backend/quick-start.md) — Get the backend running in minutes
- [API Reference](backend/api-reference/) — Full REST API documentation
- [Node Operator Guide](backend/node-operator-guide.md) — Run a node and earn crypto
- [Mobile App](mobile/overview.md) — iOS & Android client
- [Desktop App](desktop/overview.md) — macOS, Windows & Linux client
- [Technical Reference](reference/data-models.md) — Deep-dive into internals

---

## Repository Structure

```
aureo/
├── aureo-vpn/          Backend infrastructure (Go 1.24, Fiber, GORM, libp2p)
│   ├── cmd/            4 binaries: api-gateway, control-server, vpn-node, cli
│   ├── pkg/            Shared packages (auth, blockchain, p2p, protocols)
│   └── internal/       Private packages (handlers, control, node)
│
├── aureo-app/          Mobile client (Expo 54, React Native, TypeScript)
│   ├── app/            File-based routing (Expo Router)
│   ├── src/            Stores, API client, hooks, native modules
│   └── plugins/        Native VPN plugins (iOS NetworkExtension, Android VpnService)
│
├── aureo-desktop/      Desktop client (Go + Wails 2.x)
│   ├── app.go          Wails App — Go methods exposed to JS
│   ├── internal/       API client + WireGuard manager
│   └── frontend/       Embedded web UI (Leaflet.js map)
│
└── docs/               Whitepaper, technical documentation
```

---

*Privacy is a right, not a privilege.*
