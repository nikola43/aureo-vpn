# ✨ Introduction

Aureo VPN is a decentralized VPN platform that reimagines internet privacy. Instead of trusting a single company with all your traffic, Aureo distributes the infrastructure across a global network of community-operated nodes.

---

## Features

| | Feature | Description |
|---|---|---|
| 🛡️ | **Military-Grade Encryption** | WireGuard with ChaCha20-Poly1305 — the most modern, audited VPN protocol available |
| 🌐 | **Fully Decentralized** | P2P node discovery via Kademlia DHT — no central servers, no single point of failure |
| 💰 | **Crypto Rewards** | Node operators earn ETH, BTC, or LTC for every GB of bandwidth they serve |
| 📱 | **Cross-Platform** | Native apps for iOS, Android, macOS, Windows, and Linux — plus a CLI for power users |
| 🔒 | **Zero-Log Architecture** | IP anonymization at the protocol level — your real IP never touches a database |
| ⚡ | **Lightning Fast** | WireGuard's minimal 4,000-line codebase delivers the highest throughput of any VPN protocol |

---

## How It Works

### 👤 For Users

1. **Open** the Aureo app on any platform
2. **Tap** "Quick Connect" — automatically selects the fastest, lowest-load node
3. **Done** — all traffic is encrypted with WireGuard (ChaCha20-Poly1305), your real IP is hidden

### 📡 For Node Operators

Run a VPN node on your server and earn crypto for every GB of bandwidth you serve:

| Tier | Rate/GB | Min Uptime | Bonus |
|------|---------|------------|-------|
| 🥉 Bronze | $0.010 | 50% | 1.0x |
| 🥈 Silver | $0.015 | 80% | 1.2x |
| 🥇 Gold | $0.020 | 90% | 1.5x |
| 💎 Platinum | $0.030 | 95% | 2.0x |

---

## Quick Start

### Backend

```bash
cd aureo-vpn

# Install dependencies & build
make setup && make build

# Start API Gateway (SQLite created automatically)
./bin/api-gateway

# Start Control Server (in another terminal)
./bin/control-server
```

Or use Docker:

```bash
cd aureo-vpn/deployments/docker
docker-compose up -d
```

### Mobile App

```bash
cd aureo-app

npm install
npm start          # Expo dev server
npm run ios        # iOS Simulator
npm run android    # Android Emulator
```

### Desktop App

```bash
cd aureo-desktop

make dev           # Dev mode with hot reload
make build         # Build for current platform
```
