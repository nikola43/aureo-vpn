# Aureo VPN - Decentralized Node Operator Guide

## 🚀 Become a VPN Node Operator and Earn Crypto!

Aureo VPN is a decentralized VPN network where anyone can run a VPN node and earn cryptocurrency rewards by providing VPN services to users. This guide will help you get started as a node operator.

---

## 💰 Earning Potential

### Reward Tiers

| Tier | Min Uptime | Min Reputation | Rate per GB | Monthly Potential* |
|------|-----------|----------------|-------------|-------------------|
| **Bronze** | 50% | 0-59 | $0.01 | $30 - $300 |
| **Silver** | 80% | 60-74 | $0.015 | $225 - $450 |
| **Gold** | 90% | 75-89 | $0.02 | $600 - $1,200 |
| **Platinum** | 95% | 90-100 | $0.03 | $900 - $1,800 |

*Based on serving 100-1000 GB/day

### Calculation Examples

**Example 1: Bronze Tier (Starting Out)**
- Daily bandwidth: 100 GB
- Rate: $0.01 per GB
- Monthly earning: 100 GB/day × 30 days × $0.01 = **$30/month**

**Example 2: Silver Tier (Growing)**
- Daily bandwidth: 500 GB
- Rate: $0.015 per GB
- Uptime bonus: 1.2x
- Monthly earning: 500 × 30 × $0.015 × 1.2 = **$270/month**

**Example 3: Platinum Tier (Professional)**
- Daily bandwidth: 1000 GB
- Rate: $0.03 per GB
- Quality bonus: 1.5x
- Monthly earning: 1000 × 30 × $0.03 × 1.5 = **$1,350/month**

### Bonus Multipliers

- **Quality Score**: 0.5x - 1.5x based on performance
- **Session Duration**: +10% for sessions > 1 hour, +20% for > 3 hours
- **Reputation**: Higher tier = higher base rate

---

## 📋 Requirements

### Minimum Hardware

- **CPU**: 2+ cores
- **RAM**: 2 GB minimum (4 GB recommended)
- **Storage**: 10 GB free space
- **Bandwidth**: 50 Mbps minimum (100+ Mbps recommended)
- **OS**: Linux (Ubuntu 20.04+, Debian 10+, CentOS 8+) or macOS

### Network Requirements

- **Public IP**: Static or dynamic (dynamic with DDNS)
- **Open Ports**: Ability to forward ports (51820 UDP, 1194 UDP)
- **Uptime**: 95%+ for optimal earnings
- **Latency**: < 100ms for best tier

### Crypto Wallet

You need a wallet address for receiving rewards. Supported cryptocurrencies:
- **Ethereum (ETH)** - Recommended
- **Bitcoin (BTC)**
- **Litecoin (LTC)**

---

## 🎯 Quick Start (5 Minutes)

### Option 1: Automated Setup Script

The easiest way to get started is using our automated setup script:

```bash
# Download and run the operator setup script
wget https://raw.githubusercontent.com/nikola43/aureo-vpn/main/scripts/become-node-operator.sh
chmod +x become-node-operator.sh
./become-node-operator.sh
```

The script will:
1. ✅ Check system requirements
2. ✅ Install dependencies (Docker, WireGuard, etc.)
3. ✅ Create your operator account
4. ✅ Setup your crypto wallet
5. ✅ Configure and start your VPN node
6. ✅ Show your earnings dashboard

### Option 2: Manual Setup

#### Step 1: Install Dependencies

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install -y wireguard wireguard-tools iptables docker.io
```

**CentOS/RHEL:**
```bash
sudo yum install -y wireguard-tools iptables docker
```

**macOS:**
```bash
brew install wireguard-tools docker
```

#### Step 2: Register as Operator

```bash
# Create account
curl -X POST https://api.aureo-vpn.com/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "your@email.com",
    "password": "SecurePassword123!",
    "username": "yourname"
  }'

# Save your access token from the response
export ACCESS_TOKEN="your_token_here"

# Register as operator
curl -X POST https://api.aureo-vpn.com/api/v1/operator/register \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "wallet_address": "0xYourEthereumAddress",
    "wallet_type": "ethereum",
    "country": "United States",
    "email": "your@email.com"
  }'
```

#### Step 3: Create Node

```bash
curl -X POST https://api.aureo-vpn.com/api/v1/operator/nodes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "name": "My-VPN-Node-US",
    "hostname": "my-node.example.com",
    "public_ip": "YOUR_PUBLIC_IP",
    "country": "United States",
    "country_code": "US",
    "city": "New York",
    "wireguard_port": 51820,
    "openvpn_port": 1194,
    "is_operator_owned": true
  }'
```

#### Step 4: Start Node

```bash
# Download node software
wget https://github.com/nikola43/aureo-vpn/releases/latest/download/vpn-node-linux-amd64
chmod +x vpn-node-linux-amd64

# Create configuration
cat > config.env << EOF
NODE_ID=your_node_id_from_step3
ACCESS_TOKEN=$ACCESS_TOKEN
API_ENDPOINT=https://api.aureo-vpn.com
EOF

# Run node
sudo ./vpn-node-linux-amd64
```

---

## 📊 Monitoring Your Earnings

### Web Dashboard

Access your operator dashboard at:
```
https://app.aureo-vpn.com/operator/dashboard
```

Features:
- Real-time earnings
- Node status and performance
- Bandwidth usage
- Payout history
- Reputation score

### CLI Commands

```bash
# View operator stats
curl -H "Authorization: Bearer $ACCESS_TOKEN" \
  https://api.aureo-vpn.com/api/v1/operator/stats

# View earnings history
curl -H "Authorization: Bearer $ACCESS_TOKEN" \
  https://api.aureo-vpn.com/api/v1/operator/earnings

# View payout history
curl -H "Authorization: Bearer $ACCESS_TOKEN" \
  https://api.aureo-vpn.com/api/v1/operator/payouts

# Request payout (min $10)
curl -X POST -H "Authorization: Bearer $ACCESS_TOKEN" \
  https://api.aureo-vpn.com/api/v1/operator/payout/request
```

### Mobile App

Download the **Aureo VPN Operator** app:
- iOS: App Store
- Android: Google Play

---

## 💳 Payouts

### Payment Schedule

- **Automatic**: Weekly (every Monday) for balances ≥ $10
- **Manual**: Request anytime when balance ≥ $10
- **Processing Time**: 24-48 hours
- **Fee**: Network transaction fees only (deducted from payout)

### Payout Process

1. **Earnings Accumulate**: As users connect to your node
2. **Pending Status**: Earnings held for 24 hours (quality verification)
3. **Confirmed**: After quality check passes
4. **Payout**: Processed weekly or on request
5. **Blockchain Transaction**: Sent to your wallet
6. **Confirmation**: Email notification with transaction hash

### Minimum Thresholds

| Cryptocurrency | Minimum Payout |
|---------------|----------------|
| Ethereum (ETH) | $10 |
| Bitcoin (BTC) | $10 |
| Litecoin (LTC) | $10 |

---

## 🎯 Maximizing Earnings

### 1. Maintain High Uptime

- **Target**: 95%+ uptime
- **Impact**: Higher tier = higher pay rate
- **Tips**:
  - Use reliable hosting (AWS, DigitalOcean, etc.)
  - Set up monitoring and alerts
  - Have backup power supply

### 2. Provide Fast Connection

- **Target**: < 50ms latency
- **Impact**: +20% quality bonus
- **Tips**:
  - Use SSD storage
  - Optimize network configuration
  - Choose datacenter with good peering

### 3. Serve More Bandwidth

- **Target**: 500+ GB/day
- **Impact**: More usage = more earnings
- **Tips**:
  - Upgrade bandwidth capacity
  - Ensure no bandwidth caps
  - Monitor and scale resources

### 4. Build Reputation

- **Target**: 90+ reputation score
- **Impact**: Access to Platinum tier (3x base rate)
- **How**:
  - Maintain uptime
  - Provide quality service
  - Collect positive user ratings
  - Participate in operator community

### 5. Run Multiple Nodes

- **Maximum**: 10 nodes per operator
- **Strategy**: Geographic diversity
- **Benefit**: 10x earning potential
- **Locations**: Different countries/cities

---

## 🔧 Technical Configuration

### Firewall Rules

```bash
# Allow WireGuard
sudo ufw allow 51820/udp

# Allow OpenVPN
sudo ufw allow 1194/udp

# Enable firewall
sudo ufw enable
```

### Systemd Service

```ini
[Unit]
Description=Aureo VPN Node
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/aureo-vpn
EnvironmentFile=/opt/aureo-vpn/config.env
ExecStart=/opt/aureo-vpn/vpn-node
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### Monitoring Setup

```bash
# Install Prometheus node exporter
wget https://github.com/prometheus/node_exporter/releases/download/v1.6.0/node_exporter-1.6.0.linux-amd64.tar.gz
tar xvfz node_exporter-1.6.0.linux-amd64.tar.gz
sudo mv node_exporter-1.6.0.linux-amd64/node_exporter /usr/local/bin/

# Run as service
sudo systemctl start node_exporter
```

---

## 🛡️ Security Best Practices

### 1. Secure Your Server

```bash
# Update system
sudo apt-get update && sudo apt-get upgrade -y

# Enable automatic security updates
sudo apt-get install -y unattended-upgrades

# Configure firewall
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 51820/udp
sudo ufw allow 1194/udp
sudo ufw enable
```

### 2. Protect Private Keys

- Never share your node private keys
- Store configuration files with restricted permissions:
  ```bash
  chmod 600 /opt/aureo-vpn/config.env
  ```
- Use encrypted backups

### 3. Monitor for Intrusions

```bash
# Install fail2ban
sudo apt-get install -y fail2ban

# Monitor logs
sudo journalctl -u aureo-vpn-node -f
```

### 4. Regular Backups

```bash
# Backup configuration
tar -czf aureo-backup-$(date +%Y%m%d).tar.gz /opt/aureo-vpn/config.env

# Store securely
scp aureo-backup-*.tar.gz user@backup-server:/backups/
```

---

## 📈 Scaling Your Operation

### Single Node → Professional Operator

**Month 1-2: Learning Phase**
- Run 1 node
- Learn the system
- Optimize performance
- Goal: Bronze → Silver tier

**Month 3-4: Growth Phase**
- Add 2-3 more nodes
- Different geographic locations
- Monitor and optimize
- Goal: Silver → Gold tier

**Month 5-6: Scale Phase**
- Scale to 5-10 nodes
- Automate monitoring
- Professional hosting
- Goal: Platinum tier

**Month 7+: Professional**
- Maximum 10 nodes
- Geographic diversity
- High availability setup
- Estimated: $5,000-$15,000/month

### Recommended Hosting Providers

| Provider | Pros | Cons | Cost/Month |
|----------|------|------|------------|
| **DigitalOcean** | Easy setup, good network | Limited locations | $40-200 |
| **AWS EC2** | Global coverage, reliable | Complex pricing | $50-300 |
| **Vultr** | Affordable, many locations | Variable performance | $30-150 |
| **Hetzner** | Cheap bandwidth | Europe-focused | $20-100 |
| **OVH** | Unlimited bandwidth | Mixed reviews | $25-120 |

---

## 🐛 Troubleshooting

### Node Won't Start

```bash
# Check logs
sudo journalctl -u aureo-vpn-node -n 100

# Verify configuration
cat /opt/aureo-vpn/config.env

# Test connectivity
curl -v https://api.aureo-vpn.com/health
```

### Low Earnings

**Check:**
1. Node uptime: `systemctl status aureo-vpn-node`
2. Network speed: `speedtest-cli`
3. Latency: `ping 8.8.8.8`
4. Reputation score: Check dashboard

**Fix:**
- Ensure 95%+ uptime
- Upgrade bandwidth
- Reduce latency
- Improve service quality

### Payout Issues

**Common Causes:**
- Balance < $10
- Invalid wallet address
- Network congestion

**Solutions:**
- Wait for minimum threshold
- Verify wallet address
- Check payout history in dashboard

### Connection Rejected

```bash
# Check firewall
sudo ufw status

# Verify ports are open
sudo netstat -tulpn | grep :51820
sudo netstat -tulpn | grep :1194

# Check node status
curl -H "Authorization: Bearer $ACCESS_TOKEN" \
  https://api.aureo-vpn.com/api/v1/operator/nodes
```

---

## 📞 Support & Community

### Get Help

- **Documentation**: https://docs.aureo-vpn.com/operators
- **Discord**: https://discord.gg/aureo-vpn
- **Telegram**: https://t.me/aureo_vpn_operators
- **Email**: operator-support@aureo-vpn.com

### Operator Community

- Share tips and best practices
- Get help from experienced operators
- Stay updated on network changes
- Vote on governance proposals

### Report Issues

```bash
# Generate debug report
sudo aureo-vpn-node --debug-report

# Submit via GitHub
https://github.com/nikola43/aureo-vpn/issues
```

---

## 📜 Terms & Conditions

### Operator Agreement

By becoming a node operator, you agree to:

1. **Service Quality**: Maintain 80%+ uptime
2. **Legal Compliance**: Follow local laws
3. **No Logging**: Not log user traffic data
4. **Fair Use**: Not abuse the network
5. **Security**: Keep your node secure

### Penalties

- **Downtime >20%**: Reduced tier
- **Poor Quality**: Earnings reduction
- **Policy Violation**: Account suspension
- **Fraud**: Permanent ban + stake slashing

### Rewards Program Rules

- Earnings calculated based on actual usage
- Quality verification period: 24 hours
- Minimum payout: $10
- Payout processing: Weekly
- Transaction fees: Operator pays
- Rates subject to change with 30-day notice

---

## 🎓 FAQ

### Q: Do I need technical knowledge?
**A:** Basic Linux knowledge helps, but our automated script makes it easy for beginners.

### Q: Can I run a node from home?
**A:** Yes, if you have a good connection and can forward ports. However, datacenter hosting is recommended for best earnings.

### Q: How much can I really earn?
**A:** Earnings vary based on location, performance, and user demand. Realistic range: $30-$1,500/month per node.

### Q: What if my node goes offline?
**A:** Short outages (< 1 hour) have minimal impact. Extended downtime will lower your tier and reduce earnings.

### Q: Can I change my wallet address?
**A:** Yes, but pending payouts will be sent to the old address. Contact support for address changes.

### Q: Are there any hidden fees?
**A:** No hidden fees. You only pay blockchain transaction fees during payouts (typically $1-5).

### Q: How is traffic metered?
**A:** Bandwidth usage is measured in GB and reported every hour by your node software.

### Q: Can I run multiple nodes?
**A:** Yes, up to 10 nodes per operator for geographic diversity.

### Q: What happens to my earnings if I stop?
**A:** Your pending balance will be paid out when it reaches $10, even after stopping.

### Q: Is this legal?
**A:** Yes, running a VPN node is legal in most countries. Check your local laws.

---

## 🚀 Get Started Now!

Ready to start earning? Run this single command:

```bash
wget https://raw.githubusercontent.com/nikola43/aureo-vpn/main/scripts/become-node-operator.sh && chmod +x become-node-operator.sh && ./become-node-operator.sh
```

Or visit: **https://app.aureo-vpn.com/become-operator**

---

**Join thousands of operators earning crypto by providing VPN services!**

*Last Updated: 2025-10-27*
*Version: 1.0.0*
