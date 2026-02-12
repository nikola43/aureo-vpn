# ⭐ Key Features

A comprehensive set of privacy and security features built into the platform.

---

## 🔐 Military-Grade Encryption

Every byte of traffic is encrypted with WireGuard's ChaCha20-Poly1305 cipher suite, providing authenticated encryption with associated data (AEAD). Key exchange uses Curve25519 elliptic curve Diffie-Hellman, with BLAKE2s for hashing and SipHash for hashtable keys.

## 🛡️ Zero-Log Architecture

Aureo implements privacy at the protocol level, not just the policy level:

- **IP Anonymization** — Client IPs are hashed using SHA-256 before storage. The original IP address never touches the database.
- **Privacy Filter** — All log messages pass through a privacy filter that redacts IP addresses, email addresses, JWT tokens, API keys, and other sensitive data using pattern-based detection.
- **Minimal Data Retention** — Active session data is retained only while the session is active. Completed session metadata is purged within 24 hours.
- **No Traffic Inspection** — VPN nodes forward encrypted packets without inspecting, logging, or storing traffic content.

## 🌐 Decentralized Node Network

The network has no single point of failure:

- Nodes discover each other via distributed hash table (no central registry)
- Any node can join or leave without affecting the network
- Load balancing is automatic — users are directed to the fastest, least-loaded node
- Geographic diversity is incentivized through the reward system

## 🔄 Multi-Protocol Support

While WireGuard is the primary protocol, Aureo supports multiple VPN protocols:

- **WireGuard** — Primary, fastest, most secure (port 51820)
- **OpenVPN** — Legacy compatibility (port 1194)
- **IPSec/IKEv2** — Enterprise compatibility

## 🔗 Multi-Hop VPN

For users requiring enhanced privacy, Aureo supports multi-hop connections where traffic is routed through two or more nodes in different jurisdictions:

```
User → Node A (Germany) → Node B (Switzerland) → Internet
```

Each hop adds a layer of encryption, making traffic correlation attacks exponentially more difficult.

## 🚫 Kill Switch & DNS Leak Protection

- **Kill Switch** — If the VPN connection drops, all internet traffic is blocked to prevent data leaks. Enabled by default.
- **DNS Leak Protection** — DNS queries are routed through the VPN tunnel to prevent ISP DNS snooping. Uses privacy-respecting DNS servers (1.1.1.1, 8.8.8.8).

## 💰 Community-Operated Nodes with Crypto Rewards

Node operators earn cryptocurrency for providing bandwidth to the network:

| Tier | Requirements | Earnings Rate | Bonus |
|------|-------------|---------------|-------|
| 🥉 Bronze | 50% uptime | $0.010/GB | 1.0x |
| 🥈 Silver | 80% uptime, 60+ reputation | $0.015/GB | 1.2x |
| 🥇 Gold | 90% uptime, 75+ reputation | $0.020/GB | 1.5x |
| 💎 Platinum | 95% uptime, 90+ reputation | $0.030/GB | 2.0x |

Quality bonuses reward stable, long-running connections. Operators who maintain higher uptime, faster speeds, and better user ratings earn progressively more.

## 📱 Cross-Platform Support

| Platform | Technology | Features |
|----------|-----------|----------|
| iOS | Expo + React Native | Native WireGuard via NetworkExtension |
| Android | Expo + React Native | Native WireGuard via VpnService |
| macOS | Go + Wails 2.x | Native WireGuard via wg-quick |
| Windows | Go + Wails 2.x | Native WireGuard integration |
| Linux | Go + Wails 2.x | Native WireGuard via wg-quick |
| CLI | Go + Cobra | Full management from terminal |
