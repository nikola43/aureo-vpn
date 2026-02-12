<div align="center">

# 👑 Aureo VPN: The Decentralized Privacy Network

### Whitepaper v1.0

*A VPN owned by everyone, controlled by no one.*

</div>

---

## 🎯 Executive Summary

The global VPN market is projected to exceed $100 billion by 2030, driven by growing privacy concerns, increasing cyber threats, and the expanding remote workforce. Yet the industry remains dominated by centralized providers that require users to trust a single entity with all their internet traffic — creating the exact vulnerability they claim to solve.

**Aureo VPN** reimagines the VPN from the ground up. By combining military-grade WireGuard encryption with a decentralized node network powered by blockchain incentives, Aureo creates a VPN ecosystem where no single entity controls the infrastructure. Community operators around the world run nodes and earn cryptocurrency rewards, while users enjoy faster, more private, and more resilient connectivity.

> 💡 **Key Insight:** Traditional VPNs simply move trust from your ISP to the VPN provider. Aureo eliminates the need for trust entirely through decentralization, open-source code, and on-chain transparency.

**Key highlights:**

- 🌐 Decentralized node network with no single point of failure
- 🔒 WireGuard protocol (ChaCha20-Poly1305 encryption)
- 🛡️ Zero-log architecture with IP anonymization at the protocol level
- 💰 Multi-chain crypto rewards for node operators (ETH, BTC, LTC)
- 📱 Cross-platform: iOS, Android, macOS, Windows, Linux
- 📡 Peer-to-peer node discovery via Kademlia DHT

---

## ⚠️ The Problem

### Centralized VPNs are fundamentally broken

Traditional VPN providers suffer from inherent architectural flaws:

| Issue | Centralized VPNs | Aureo VPN |
|-------|-----------------|-----------|
| **Trust Model** | ❌ "Trust us" — unverifiable no-log claims | ✅ Open source, auditable, verify yourself |
| **Single Point of Failure** | ❌ One server seizure compromises all | ✅ Distributed P2P network |
| **Censorship Resistance** | ❌ Known IP ranges easily blocked | ✅ Thousands of independent community nodes |
| **Transparency** | ❌ Closed source, opaque policies | ✅ Public code, on-chain payments |
| **Economic Model** | ❌ Only the provider profits | ✅ Community operators earn crypto rewards |

**🔐 Single Point of Trust**
When you connect to a centralized VPN, you are simply moving your trust from your ISP to the VPN provider. The provider can see, log, and potentially monetize all your traffic. Despite "no-log" marketing claims, users have no way to verify these promises. Multiple major VPN providers have been caught logging user data despite explicit no-log policies.

**💥 Single Point of Failure**
Centralized infrastructure means a single server seizure, legal order, or data breach can compromise the entire network. Government authorities can compel providers to install surveillance capabilities or shut down operations entirely.

**🚫 Censorship Vulnerability**
Authoritarian regimes can block centralized VPN providers by targeting their known server IP ranges. Once identified, the entire service becomes inaccessible in that region.

**🐌 Performance Bottlenecks**
Traffic must route through provider-owned data centers, creating bottlenecks and adding unnecessary latency. Users in underserved regions face limited server options and degraded performance.

**🕶️ Opaque Business Models**
Many "free" VPN providers monetize user data through advertising, analytics, or direct data sales. Even paid services operate as black boxes — users cannot audit the infrastructure handling their sensitive traffic.

---

## 💡 Our Solution

### A VPN owned by everyone, controlled by no one

Aureo VPN replaces the traditional centralized model with a decentralized architecture where the community IS the infrastructure.

```
Traditional VPN                    Aureo VPN

 ┌─────────┐                       ┌─────────┐
 │  User   │                       │  User   │
 └────┬────┘                       └────┬────┘
      │                                 │
      ▼                                 ▼
 ┌─────────┐                    ┌───────────────┐
 │ Company │   Single point     │  Decentralized│   No single
 │ Server  │   of failure       │  Node Network │   point of
 │ Farm    │   & trust          │  (Community)  │   failure
 └────┬────┘                    └───────┬───────┘
      │                                 │
      ▼                                 ▼
 ┌─────────┐                    ┌───────────────┐
 │Internet │                    │   Internet    │
 └─────────┘                    └───────────────┘
```

**Three pillars of the Aureo network:**

1. 🏗️ **Community-Operated Nodes** — Anyone can run an Aureo node and contribute bandwidth to the network. Node operators earn cryptocurrency rewards proportional to their contribution.

2. 📡 **Peer-to-Peer Discovery** — Nodes discover each other through a distributed hash table (Kademlia DHT) and gossip protocol, eliminating the need for central registries that can be censored or seized.

3. ⛓️ **Blockchain-Powered Incentives** — Smart contracts and multi-chain integration (Ethereum, Bitcoin, Litecoin) ensure transparent, automated compensation for node operators.

---

## ⚙️ How Aureo Works

### 👤 For Users

Connecting to Aureo is as simple as pressing a button:

```
1. Open the Aureo app
   └── Available on iOS, Android, macOS, Windows, Linux

2. Tap "Quick Connect"
   └── Automatically selects the fastest, lowest-load node

3. You're protected
   └── All traffic encrypted with WireGuard (ChaCha20-Poly1305)
   └── Your real IP is hidden from every website and service
   └── Kill switch prevents any data leaks if connection drops
```

Behind the scenes:

```
┌────────────┐     1. Generate      ┌──────────────────────┐
│            │     WireGuard keys   │                      │
│   Aureo    │ ──────────────────►  │   Aureo API Gateway  │
│   App      │     2. Register      │                      │
│            │     public key       │  - Select best node  │
│            │ ◄──────────────────  │  - Allocate tunnel IP│
│            │     3. Receive       │  - Return server key │
│            │     server config    │                      │
└─────┬──────┘                      └──────────────────────┘
      │
      │  4. Establish encrypted
      │     WireGuard tunnel
      ▼
┌──────────────┐
│   VPN Node   │  5. All traffic routed
│  (Community  │     through encrypted tunnel
│   Operated)  │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│   Internet   │  Your real IP is hidden
└──────────────┘
```

### 📡 For Node Operators

Running an Aureo node turns your server into a revenue-generating privacy infrastructure:

```
1. Register as an operator
   └── Provide your crypto wallet address (ETH, BTC, or LTC)

2. Deploy a node
   └── Run the Aureo node software on your server
   └── Automatic P2P network integration

3. Earn rewards
   └── Get paid per GB of bandwidth served
   └── Higher earnings for better performance
   └── Automated payouts to your crypto wallet
```

---

## 🔧 Technology Stack

### 🔒 WireGuard Protocol

Aureo uses [WireGuard](https://www.wireguard.com/) as its primary VPN protocol — the most modern, audited, and performant VPN protocol available:

| Property | WireGuard | OpenVPN | IPSec |
|----------|-----------|---------|-------|
| Encryption | ChaCha20-Poly1305 | AES-256-GCM | AES-256-CBC |
| Key Exchange | Curve25519 | RSA/ECDH | IKEv2 |
| Code Lines | ~4,000 | ~100,000 | ~400,000 |
| Handshake | 1-RTT | Multi-RTT | Multi-RTT |
| Performance | Highest | Moderate | Moderate |
| Audit Surface | Minimal | Large | Large |

WireGuard's minimal codebase means a smaller attack surface and fewer potential vulnerabilities compared to legacy protocols.

### 🌐 Peer-to-Peer Network (libp2p)

Aureo's decentralized infrastructure is built on [libp2p](https://libp2p.io/), the same networking stack powering IPFS and Filecoin:

- **Kademlia DHT** — Distributed hash table for node discovery without central servers
- **GossipSub** — Efficient message propagation for node announcements and health updates
- **mDNS** — Local network discovery for mesh-network deployments
- **NAT Traversal** — UPnP, hole punching, and relay support for nodes behind firewalls
- **Transport Security** — Encrypted peer connections over TCP and QUIC

Nodes broadcast heartbeats every 30 seconds, announce their capabilities every 5 minutes, and are automatically removed from the network if unresponsive for 2 minutes.

### ⛓️ Multi-Chain Blockchain Integration

Operator rewards are paid in cryptocurrency with support for multiple blockchains:

- **Ethereum (ETH)** — Smart contract integration via go-ethereum
- **Bitcoin (BTC)** — Native BTC payments via RPC
- **Litecoin (LTC)** — Fast, low-fee alternative payments

All transactions are transparent and verifiable on their respective blockchains. Exchange rates are fetched in real-time to ensure fair USD-equivalent compensation.

---

## ⭐ Key Features

### 🔐 Military-Grade Encryption

Every byte of traffic is encrypted with WireGuard's ChaCha20-Poly1305 cipher suite, providing authenticated encryption with associated data (AEAD). Key exchange uses Curve25519 elliptic curve Diffie-Hellman, with BLAKE2s for hashing and SipHash for hashtable keys.

### 🛡️ Zero-Log Architecture

Aureo implements privacy at the protocol level, not just the policy level:

- **IP Anonymization** — Client IPs are hashed using SHA-256 before storage. The original IP address never touches the database.
- **Privacy Filter** — All log messages pass through a privacy filter that redacts IP addresses, email addresses, JWT tokens, API keys, and other sensitive data using pattern-based detection.
- **Minimal Data Retention** — Active session data is retained only while the session is active. Completed session metadata is purged within 24 hours.
- **No Traffic Inspection** — VPN nodes forward encrypted packets without inspecting, logging, or storing traffic content.

### 🌐 Decentralized Node Network

The network has no single point of failure:

- Nodes discover each other via distributed hash table (no central registry)
- Any node can join or leave without affecting the network
- Load balancing is automatic — users are directed to the fastest, least-loaded node
- Geographic diversity is incentivized through the reward system

### 🔄 Multi-Protocol Support

While WireGuard is the primary protocol, Aureo supports multiple VPN protocols:

- **WireGuard** — Primary, fastest, most secure (port 51820)
- **OpenVPN** — Legacy compatibility (port 1194)
- **IPSec/IKEv2** — Enterprise compatibility

### 🔗 Multi-Hop VPN

For users requiring enhanced privacy, Aureo supports multi-hop connections where traffic is routed through two or more nodes in different jurisdictions:

```
User → Node A (Germany) → Node B (Switzerland) → Internet
```

Each hop adds a layer of encryption, making traffic correlation attacks exponentially more difficult.

### 🚫 Kill Switch & DNS Leak Protection

- **Kill Switch** — If the VPN connection drops, all internet traffic is blocked to prevent data leaks. Enabled by default.
- **DNS Leak Protection** — DNS queries are routed through the VPN tunnel to prevent ISP DNS snooping. Uses privacy-respecting DNS servers (1.1.1.1, 8.8.8.8).

### 💰 Community-Operated Nodes with Crypto Rewards

Node operators earn cryptocurrency for providing bandwidth to the network:

| Tier | Requirements | Earnings Rate | Bonus |
|------|-------------|---------------|-------|
| 🥉 Bronze | 50% uptime | $0.010/GB | 1.0x |
| 🥈 Silver | 80% uptime, 60+ reputation | $0.015/GB | 1.2x |
| 🥇 Gold | 90% uptime, 75+ reputation | $0.020/GB | 1.5x |
| 💎 Platinum | 95% uptime, 90+ reputation | $0.030/GB | 2.0x |

Quality bonuses reward stable, long-running connections. Operators who maintain higher uptime, faster speeds, and better user ratings earn progressively more.

### 📱 Cross-Platform Support

| Platform | Technology | Features |
|----------|-----------|----------|
| iOS | Expo + React Native | Native WireGuard via NetworkExtension |
| Android | Expo + React Native | Native WireGuard via VpnService |
| macOS | Go + Wails 2.x | Native WireGuard via wg-quick |
| Windows | Go + Wails 2.x | Native WireGuard integration |
| Linux | Go + Wails 2.x | Native WireGuard via wg-quick |
| CLI | Go + Cobra | Full management from terminal |

---

## 💰 Node Operator Economy

### 💸 How Operators Earn

Node operators are compensated based on actual bandwidth served, adjusted for quality:

```
Earnings = BandwidthGB x RatePerGB x QualityMultiplier x DurationBonus
```

**Quality Multiplier** (0.5x - 1.5x): Based on connection quality score (latency, stability, uptime).

**Duration Bonus**: Longer stable connections earn more:
- Standard: 1.0x
- 1+ hour sessions: 1.1x
- 3+ hour sessions: 1.2x

### ⭐ Reputation System

Every operator starts with a base reputation score of 50 (out of 100). The score is composed of:

| Component | Max Points | Criteria |
|-----------|-----------|----------|
| Base | 50 | Starting score |
| Uptime | 30 | (Average Uptime % / 100) x 30 |
| User Ratings | 20 | (Average Rating / 5) x 20 |
| Bandwidth Served | 10 | 100GB+ = 5pts, 1TB+ = 10pts |
| Stake Amount | 10 | $100+ = 5pts, $1000+ = 10pts |

Higher reputation unlocks better reward tiers, creating a positive feedback loop that incentivizes quality infrastructure.

### 🔐 Staking

Operators can optionally stake cryptocurrency as a security deposit, demonstrating commitment to the network:

- Staked funds boost reputation score
- In cases of proven malicious behavior, stakes can be slashed
- Staking creates economic alignment between operators and users

### 💳 Payout Pipeline

Earnings flow through a transparent pipeline:

```
Traffic Served → Earnings Recorded (pending)
    → Quality Verified (confirmed)
    → Payout Threshold Reached
    → Crypto Conversion (real-time exchange rate)
    → Blockchain Transaction
    → Operator Wallet
```

Processing time: 24-48 hours from request to wallet receipt.

---

## 💳 Subscription Tiers

| Feature | Free | Basic | Premium |
|---------|------|-------|---------|
| Bandwidth | Limited | Unlimited | Unlimited |
| Server Access | Standard | All servers | All servers + Premium |
| Protocols | WireGuard | WireGuard + OpenVPN | All protocols |
| Multi-Hop | ❌ | ❌ | ✅ |
| Kill Switch | ✅ | ✅ | ✅ |
| DNS Protection | ✅ | ✅ | ✅ |
| Simultaneous Devices | 1 | 3 | 5 |
| Support | Community | Email | Priority |

---

## 🛡️ Security & Privacy

### 🏰 Defense-in-Depth Architecture

Aureo employs multiple layers of security:

1. 🔒 **Transport Security** — WireGuard with ChaCha20-Poly1305 AEAD encryption
2. 🔑 **Authentication** — JWT tokens with HMAC-SHA256, short-lived access tokens (15 min)
3. 🔐 **Password Security** — bcrypt hashing with common password blocklist
4. ✅ **Input Validation** — Comprehensive validation with SQL injection and XSS detection
5. 🚦 **Rate Limiting** — Per-IP request throttling to prevent abuse
6. 🛡️ **Privacy Filter** — Automatic redaction of PII from all system logs

### 🔍 Zero-Knowledge Architecture

Aureo is designed so that no single component has complete information:

- The **API Gateway** knows user accounts but not traffic content
- **VPN Nodes** forward encrypted traffic but do not have user account details
- **Client IPs** are hashed before storage — the real IP is never persisted
- **Payment processing** is on-chain — no centralized payment data

### 📜 Compliance

- **GDPR Ready** — Minimal data collection, defined retention periods, right to deletion
- **Privacy by Design** — Privacy protections are architectural, not just policy-based
- **Open Infrastructure** — Community-operated nodes can be independently audited
- **Warrant Canary** — Published and regularly updated statement confirming no government data requests

---

## 📱 Platform Support

### 📱 Mobile App

The Aureo mobile app provides a premium VPN experience:

- **Quick Connect** — One-tap connection to the fastest available server
- **Server Browser** — Browse nodes by country, filter by protocol
- **Real-Time Stats** — Live upload/download speed, data usage, connection duration
- **Connection Profiles** — Save preferred server configurations
- **Native VPN Integration** — Uses iOS NetworkExtension and Android VpnService for system-level VPN tunneling
- **Secure Storage** — Credentials stored in platform secure enclave (iOS Keychain / Android Keystore)
- **Background Operation** — VPN remains active when app is backgrounded

### 🖥️ Desktop App

The desktop application offers full-featured VPN management:

- **Interactive World Map** — Visualize all VPN nodes on a Leaflet.js-powered map
- **Server List** — Searchable, filterable server browser with real-time load indicators
- **Quick Actions** — Quick Connect, Secure Core, P2P Friendly, Random Server
- **System Integration** — Native WireGuard integration via wg-quick
- **Persistent Sessions** — Auto-reconnect on app launch
- **Cross-Platform** — Single codebase builds for macOS (Universal), Windows, and Linux

### ⌨️ CLI Tool

For power users and automation:

```bash
# Login
aureo login --email user@example.com --password secret

# Connect to best server
aureo connect --protocol wireguard

# Connect to specific country
aureo connect --country US --protocol wireguard

# Check status
aureo status

# Disconnect
aureo disconnect

# List available nodes
aureo nodes --country US --protocol wireguard

# View profile and stats
aureo profile
aureo stats
```

---

## 🗺️ Roadmap

### ✅ Phase 1 — Foundation (Completed)

- Core VPN infrastructure (API Gateway, Control Server, VPN Node)
- WireGuard protocol implementation
- User authentication and account management
- Mobile app (iOS + Android)
- Desktop app (macOS, Windows, Linux)
- P2P node discovery with libp2p

### 🚧 Phase 2 — Decentralization (In Progress)

- Node operator registration and management
- Multi-chain crypto rewards (ETH, BTC, LTC)
- Reputation and reward tier system
- Operator dashboard and earnings tracking
- Automated blockchain payouts

### 🔮 Phase 3 — Growth

- Multi-hop VPN routing
- Traffic obfuscation for censorship resistance
- Browser extension
- Smart DNS and ad blocking
- Expanded cryptocurrency support (Monero, Solana)
- Decentralized governance (operator voting)

### 🚀 Phase 4 — Scale

- 1000+ community-operated nodes worldwide
- Enterprise API for business integration
- White-label solutions
- Hardware VPN router support
- Post-quantum cryptography readiness
- Zero-knowledge proof integration for enhanced privacy

---

## 📬 Team & Contact

Aureo VPN is built by a team of privacy advocates, cryptography engineers, and distributed systems architects committed to building a more private internet.

**Website:** [aureovpn.com](https://aureovpn.com)
**Email:** contact@aureovpn.com
**GitHub:** [github.com/nikola43/aureo-vpn](https://github.com/nikola43/aureo-vpn)

---

<div align="center">

*This whitepaper is a living document and will be updated as the Aureo VPN platform evolves.*
*All technical claims are verifiable against the open-source codebase.*

**Aureo VPN — Privacy is a right, not a privilege.**

</div>
