# 🔧 Tech Stack

A complete overview of the technologies powering the Aureo VPN platform.

---

## Full Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| **Backend** | Go 1.24, Fiber v2, GORM | REST API, business logic, ORM |
| **VPN Protocol** | WireGuard | ChaCha20-Poly1305 encrypted tunnels |
| **P2P Network** | libp2p (Kademlia DHT, GossipSub) | Decentralized node discovery |
| **Blockchain** | go-ethereum, BTC/LTC RPC | Multi-chain crypto payments |
| **Database** | SQLite (dev) / PostgreSQL (prod) | Persistent storage |
| **Mobile** | Expo 54, React Native 0.81, TypeScript | iOS & Android apps |
| **State** | Zustand 5, React Query 5 | Client-side state management |
| **Desktop** | Go + Wails 2.x, Leaflet.js | Cross-platform desktop app |
| **Metrics** | Prometheus + Grafana | Real-time monitoring |
| **CI/CD** | GitHub Actions, Docker | Automated pipelines |

---

## Backend Details

| Component | Technology | Version |
|-----------|-----------|---------|
| Language | Go | 1.24 |
| HTTP Framework | Fiber | v2 |
| ORM | GORM | v2 |
| Authentication | JWT (golang-jwt) | v5 |
| Password Hashing | bcrypt | — |
| Crypto (ETH) | go-ethereum | — |
| P2P Networking | libp2p | — |
| Metrics | Prometheus client | — |
| Logging | slog (structured) | — |

## Mobile App Details

| Component | Technology | Version |
|-----------|-----------|---------|
| Runtime | Expo | 54.0 |
| Framework | React Native | 0.81 |
| Language | TypeScript | 5.9 |
| State | Zustand | 5.0 |
| Server State | React Query | 5.90 |
| HTTP | Axios | 1.13 |
| Animation | Reanimated | 4.1 |
| Routing | Expo Router | 6.0 |
| Crypto | tweetnacl | 1.0 |
| Storage | Expo SecureStore + MMKV | — |

## Desktop App Details

| Component | Technology | Version |
|-----------|-----------|---------|
| Framework | Wails | 2.x |
| Backend | Go | 1.21+ |
| Frontend | HTML/JS/CSS (static) | — |
| Maps | Leaflet.js | — |
| Window | 1024x768, dark theme | — |
| Build targets | macOS Universal, Windows amd64, Linux amd64 | — |
