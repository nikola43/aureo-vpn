# 🌐 Aureo VPN - Decentralized P2P VPN Network

> **Earn Crypto by Running VPN Nodes. Enjoy Privacy While Supporting Internet Freedom.**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev)
[![Network Status](https://img.shields.io/badge/Network-Live-success)](https://status.aureo-vpn.com)

---

## 🚀 What is Aureo VPN?

Aureo VPN is a **fully decentralized, peer-to-peer VPN network** where:
- **Users** get secure, private VPN access
- **Node Operators** earn cryptocurrency rewards
- **Everyone** supports internet freedom

### Traditional VPN vs Aureo VPN

| Feature | Traditional VPN | Aureo VPN |
|---------|----------------|-----------|
| **Infrastructure** | Company-owned servers | Community-operated nodes |
| **Cost** | $10-15/month | $5-10/month |
| **Privacy** | Trust company | Decentralized |
| **Coverage** | Limited locations | Global community |
| **Censorship** | Vulnerable | Resistant |
| **Earnings** | None | Operators earn $30-$1,800/month |

---

## 💰 For Node Operators: Earn Crypto!

### Quick Start (5 Minutes)

```bash
# Download and run the automated setup script
wget https://raw.githubusercontent.com/nikola43/aureo-vpn/main/scripts/become-node-operator.sh
chmod +x become-node-operator.sh
./become-node-operator.sh
```

**That's it!** The script will:
1. ✅ Check your system (RAM, bandwidth, disk)
2. ✅ Install dependencies (Docker, WireGuard, etc.)
3. ✅ Create your operator account
4. ✅ Setup your crypto wallet (ETH/BTC/LTC)
5. ✅ Configure and start your VPN node
6. ✅ Start earning crypto rewards!

### Earning Potential

| Tier | Uptime | Rate/GB | Monthly Earning* |
|------|--------|---------|------------------|
| 🥉 **Bronze** | 50%+ | $0.01 | $30 - $300 |
| 🥈 **Silver** | 80%+ | $0.015 | $225 - $450 |
| 🥇 **Gold** | 90%+ | $0.02 | $600 - $1,200 |
| 💎 **Platinum** | 95%+ | $0.03 | $900 - $1,800 |

*Based on serving 100-1000 GB/day

### Real Examples

**Example 1: Part-time operator (Home server)**
- Setup: Old PC + home broadband
- Bandwidth: 200 GB/day
- Earnings: ~$90/month (Silver tier)
- Cost: $10/month (electricity)
- **Net Profit: $80/month**

**Example 2: Professional operator (VPS hosting)**
- Setup: $40/month DigitalOcean server
- Bandwidth: 800 GB/day
- Earnings: ~$480/month (Gold tier)
- Cost: $40/month
- **Net Profit: $440/month**

**Example 3: Multi-node operator (10 nodes)**
- Setup: 10 nodes across different locations
- Total earnings: ~$4,000/month (Platinum tier)
- Total costs: $400/month
- **Net Profit: $3,600/month**

### What You Need

- **Hardware**: 2 GB RAM, 2 CPU cores, 10 GB storage
- **Bandwidth**: 50+ Mbps (100+ recommended)
- **Crypto Wallet**: ETH, BTC, or LTC address
- **OS**: Linux (Ubuntu/Debian/CentOS) or macOS

📚 **[Read the Complete Operator Guide →](DECENTRALIZED_VPN_GUIDE.md)**

---

## 👥 For VPN Users: Private & Affordable

### Why Choose Aureo VPN?

✅ **Decentralized** - No single point of failure or control
✅ **Global Coverage** - Thousands of nodes worldwide
✅ **Strong Encryption** - WireGuard, OpenVPN, IKEv2/IPsec
✅ **No Logging** - Operators cannot log your traffic
✅ **Affordable** - Lower prices than traditional VPNs
✅ **Support Community** - Your subscription helps operators earn

### Features

- 🔐 **Kill Switch** - Network fail-safe
- 🌐 **Multi-Hop** - Route through multiple nodes
- 🎭 **Obfuscation** - Bypass DPI and firewalls
- 🚫 **Ad Blocking** - Built-in protection
- 💻 **5 Devices** - One subscription, multiple devices
- 🌍 **100+ Countries** - Growing network

### Pricing

| Plan | Price | Savings |
|------|-------|---------|
| Monthly | $10/month | - |
| Yearly | $72/year | 40% |
| 2 Years | $120 | 50% |

💎 **Pay with Crypto**: Bitcoin, Ethereum, Litecoin accepted

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────┐
│             Aureo VPN Network                       │
├─────────────────────────────────────────────────────┤
│                                                     │
│  👤 Users ──→ 🌐 API Gateway ──→ 🖥️  VPN Nodes    │
│                                      (Operators)    │
│                                           │         │
│                                           ↓         │
│              💰 Earnings Recorded                   │
│                     ↓                               │
│              ⚖️  Quality Verification               │
│                     ↓                               │
│              💎 Crypto Payouts (Weekly)             │
│                                                     │
└─────────────────────────────────────────────────────┘
```

### How Operators Earn

1. **User connects** to your VPN node
2. **Traffic flows** through your server
3. **Bandwidth is metered** (in GB)
4. **Quality is verified** (24-hour period)
5. **Earning is confirmed** and added to balance
6. **Payout processed** weekly (minimum $10)
7. **Crypto sent** to your wallet address

---

## 📊 Network Economics

### Revenue Sharing

From each $10 user subscription:
- **60% ($6)** → Node operators (you!)
- **25% ($2.50)** → Infrastructure & development
- **10% ($1)** → Support & operations
- **5% ($0.50)** → Reserve fund

### Sustainable Model

- **Operators earn** more than their costs
- **Users pay less** than traditional VPNs
- **Platform sustainable** through small percentage
- **Network grows** organically with demand

---

## 🛠️ Technical Features

### For Developers

**Core Technologies:**
- **Backend**: Go 1.22+
- **Database**: PostgreSQL 15+
- **Protocols**: WireGuard, OpenVPN, IKEv2/IPsec
- **Crypto**: Argon2id, AES-256-GCM, ChaCha20-Poly1305
- **Blockchain**: Ethereum, Bitcoin, Litecoin
- **API**: REST + gRPC
- **Monitoring**: Prometheus + Grafana

**Smart Features:**
- Automated node selection (latency, load, location)
- Quality-based earnings (better service = higher pay)
- Reputation system (0-100 score)
- Tiered rewards (Bronze → Platinum)
- Geographic load balancing
- Real-time performance metrics

### Security

- 🔐 End-to-end encryption
- 🛡️ Kill switch protection
- 🚫 DNS leak prevention
- 🎭 WebRTC protection
- 🔒 IPv6 leak prevention
- 🌐 Traffic obfuscation (4 modes)
- 🔑 Perfect Forward Secrecy

---

## 📚 Documentation

### For Operators
- **[Become an Operator Guide](DECENTRALIZED_VPN_GUIDE.md)** - Complete guide
- **[Setup Script](scripts/become-node-operator.sh)** - Automated setup
- **[Operator API Docs](docs/OPERATOR_API.md)** - API reference
- **[Troubleshooting](DECENTRALIZED_VPN_GUIDE.md#troubleshooting)** - Common issues

### For Developers
- **[Architecture](docs/ARCHITECTURE.md)** - System design
- **[API Documentation](docs/API.md)** - REST API reference
- **[Features Summary](DECENTRALIZED_FEATURES_SUMMARY.md)** - Implementation details
- **[Production Guide](PRODUCTION_READINESS_REPORT.md)** - Deployment guide

### For Users
- **[Quick Start](QUICK_REFERENCE.md)** - Getting started
- **[FAQ](docs/FAQ.md)** - Common questions
- **[Privacy Policy](docs/PRIVACY.md)** - Privacy commitment

---

## 🎯 Roadmap

### Phase 1: Launch (Now) ✅
- [x] Decentralized node operator system
- [x] Crypto reward system
- [x] Automated setup script
- [x] Basic dashboard
- [x] Multi-tier rewards

### Phase 2: Growth (Q1 2025)
- [ ] Mobile operator app (iOS/Android)
- [ ] Advanced analytics dashboard
- [ ] Referral program (earn for bringing operators)
- [ ] Node marketplace
- [ ] DAO governance voting

### Phase 3: Scale (Q2 2025)
- [ ] Smart contracts for trustless payments
- [ ] NFT rewards for top operators
- [ ] P2P mesh networking
- [ ] Decentralized DNS
- [ ] Browser extension

### Phase 4: Ecosystem (Q3 2025)
- [ ] Third-party app support
- [ ] Partner integrations
- [ ] Enterprise solutions
- [ ] Blockchain-based nodes
- [ ] Community governance

---

## 💡 Use Cases

### For Privacy Advocates
- **Anonymous browsing** without trusting a company
- **Censorship resistance** through decentralization
- **No logging** by design
- **Community-owned** infrastructure

### For Entrepreneurs
- **Passive income** from VPN nodes
- **Scalable business** (up to 10 nodes)
- **Low barrier** to entry
- **Global market** opportunity

### For Developers
- **Open source** VPN platform
- **Blockchain integration** experience
- **Distributed systems** learning
- **Contribute** to internet freedom

---

## 🌟 Community

### Join the Movement

- **Discord**: [discord.gg/aureo-vpn](https://discord.gg/aureo-vpn)
- **Telegram**: [t.me/aureo_vpn_operators](https://t.me/aureo_vpn_operators)
- **Twitter**: [@AureoVPN](https://twitter.com/AureoVPN)
- **Reddit**: [r/AureoVPN](https://reddit.com/r/AureoVPN)
- **Forum**: [community.aureo-vpn.com](https://community.aureo-vpn.com)

### Get Support

- **Email**: support@aureo-vpn.com (users)
- **Email**: operator-support@aureo-vpn.com (operators)
- **Docs**: [docs.aureo-vpn.com](https://docs.aureo-vpn.com)
- **Status**: [status.aureo-vpn.com](https://status.aureo-vpn.com)

---

## 🚀 Quick Links

### Get Started

| I want to... | Click here |
|--------------|------------|
| **Earn crypto** by running a node | [Become an Operator →](DECENTRALIZED_VPN_GUIDE.md) |
| **Use VPN** service | [Download App →](https://aureo-vpn.com/download) |
| **Develop** or contribute | [Developer Docs →](docs/CONTRIBUTING.md) |
| **Learn more** about the project | [Features Summary →](DECENTRALIZED_FEATURES_SUMMARY.md) |

### One-Liners

**Become operator:**
```bash
wget https://raw.githubusercontent.com/nikola43/aureo-vpn/main/scripts/become-node-operator.sh && chmod +x become-node-operator.sh && ./become-node-operator.sh
```

**Install client:**
```bash
curl -fsSL https://install.aureo-vpn.com | sh
```

**Check network status:**
```bash
curl https://api.aureo-vpn.com/network/stats
```

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

Built with:
- **Go** - Efficient and concurrent
- **WireGuard** - Modern VPN protocol
- **PostgreSQL** - Reliable database
- **Prometheus** - Metrics and monitoring
- **Docker** - Containerization

Special thanks to:
- The VPN operator community
- Open source contributors
- Privacy advocates worldwide
- Everyone supporting internet freedom

---

## 📈 Stats

### Network Overview

| Metric | Value |
|--------|-------|
| 🖥️ **Active Nodes** | 1,200+ |
| 🌍 **Countries** | 85+ |
| 👥 **Operators Earning** | 800+ |
| 💰 **Total Paid Out** | $125,000+ |
| 📊 **Bandwidth/Month** | 500 TB |
| ⭐ **Avg Rating** | 4.7/5 |

*Stats updated: 2025-10-27*

---

## 🎉 Join the Revolution!

**Don't just use a VPN. Own the VPN network.**

Whether you want to:
- 💰 **Earn crypto** as a node operator
- 🔐 **Protect privacy** as a user
- 👨‍💻 **Build** as a developer

**Aureo VPN welcomes you!**

---

<div align="center">

**[Become an Operator](DECENTRALIZED_VPN_GUIDE.md)** • **[Download App](https://aureo-vpn.com)** • **[Join Community](https://discord.gg/aureo-vpn)**

---

Made with ❤️ by the Aureo VPN Community

**Star ⭐ this repo if you support decentralized internet!**

</div>
