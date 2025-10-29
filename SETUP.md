# Aureo VPN - Quick Setup Guide

## 🚀 Become a Node Operator in 5 Minutes

### Prerequisites

- Linux server (Ubuntu 20.04+ or Debian 11+ recommended)
- Docker and Docker Compose installed
- Root access
- Public IP address
- Crypto wallet address (ETH, BTC, or LTC)

### One-Command Setup

```bash
cd aureo-vpn
sudo bash scripts/become-node-operator.sh
```

That's it! The script will:
- ✅ Deploy all services via Docker Compose
- ✅ Create your operator account
- ✅ Register and activate your VPN node
- ✅ Configure WireGuard VPN server
- ✅ Setup peer registration
- ✅ Configure monitoring

### What You'll Need to Provide

During setup, you'll be prompted for:

1. **Email**: Your account email
2. **Username**: Your operator username
3. **Password**: Secure password (min 8 characters)
4. **Crypto Type**: Choose ETH, BTC, or LTC
5. **Wallet Address**: Your crypto wallet for rewards
6. **Node Name**: A name for your VPN node (optional)

### After Setup

Your VPN node will be running and earning rewards! Here's what's deployed:

- **API Gateway**: `http://localhost:8080`
- **Web Dashboard**: `http://localhost:3001`
- **Grafana Metrics**: `http://localhost:3000`
- **Prometheus**: `http://localhost:9090`
- **VPN Node**: Running on port `51820` (WireGuard)

### Useful Commands

```bash
# View all containers
docker compose -f deployments/docker/docker-compose.yml ps

# View logs
docker compose -f deployments/docker/docker-compose.yml logs -f

# Check WireGuard status
docker exec aureo-vpn-node-1 wg show wg0

# Restart services
./scripts/deploy.sh restart

# Rebuild after code changes
./scripts/deploy.sh rebuild
```

### Earnings Tiers

| Tier | Rate | Requirement |
|------|------|-------------|
| 🥉 Bronze | $0.01/GB | 50%+ uptime |
| 🥈 Silver | $0.015/GB | 80%+ uptime |
| 🥇 Gold | $0.02/GB | 90%+ uptime |
| 💎 Platinum | $0.03/GB | 95%+ uptime |

**Minimum payout**: $10
**Payment schedule**: Weekly (Fridays)

### Example Earnings

- 100 GB/day × 30 days × $0.01/GB = **$30/month**
- 500 GB/day × 30 days × $0.02/GB = **$300/month**
- 1000 GB/day × 30 days × $0.03/GB = **$900/month**

---

## 📱 Connect as a Client

### For Mac Users

```bash
# Install WireGuard
brew install wireguard-tools

# Connect to VPN
./scripts/aureo-vpn-mac.sh connect

# Disconnect
./scripts/aureo-vpn-mac.sh disconnect

# Check status
./scripts/aureo-vpn-mac.sh status
```

### For Linux Users

```bash
# Install WireGuard
sudo apt install wireguard-tools

# Connect to VPN
./scripts/aureo-vpn-linux.sh connect

# Disconnect
./scripts/aureo-vpn-linux.sh disconnect
```

---

## 🔧 Advanced Configuration

### Environment Variables

Edit `deployments/docker/.env` to customize:

```env
# Node ID (auto-generated)
NODE_ID_1=<your-node-uuid>

# Ports (optional, defaults shown)
# API_PORT=8080
# POSTGRES_PORT=5432
# WIREGUARD_PORT=51820
```

### Custom DNS Servers

Edit `/opt/aureo-vpn/add-wireguard-peer.sh` line 73:

```bash
"dns": "1.1.1.1,8.8.8.8",  # Change to your preferred DNS
```

### Firewall Configuration

Ensure these ports are open:

```bash
# WireGuard VPN
sudo ufw allow 51820/udp

# API Gateway (if accessing remotely)
sudo ufw allow 8080/tcp

# Web Dashboard (if accessing remotely)
sudo ufw allow 3001/tcp
```

---

## 🐛 Troubleshooting

### VPN Node Not Starting

```bash
# Check logs
docker logs aureo-vpn-node-1

# Common issue: No internal_ip set
docker exec aureo-vpn-db psql -U postgres -d aureo_vpn -c \
  "UPDATE vpn_nodes SET internal_ip = '10.8.0.1' WHERE public_ip = 'YOUR_IP';"

# Restart node
docker restart aureo-vpn-node-1
```

### Clients Can't Connect

```bash
# Check WireGuard is running
docker exec aureo-vpn-node-1 wg show wg0

# Verify firewall allows UDP 51820
sudo ufw status

# Check NAT rules
docker exec aureo-vpn-node-1 iptables -t nat -L -n -v
```

### No Internet Through VPN

```bash
# Check IP forwarding
docker exec aureo-vpn-node-1 sysctl net.ipv4.ip_forward

# Should output: net.ipv4.ip_forward = 1

# Check MASQUERADE rule exists
docker exec aureo-vpn-node-1 iptables -t nat -L POSTROUTING -n -v
# Should show MASQUERADE rule for eth0
```

### Database Issues

```bash
# Access database
docker exec -it aureo-vpn-db psql -U postgres -d aureo_vpn

# Check nodes
SELECT id, name, public_ip, internal_ip, status FROM vpn_nodes;

# Check operators
SELECT id, wallet_address, status, active_nodes_count FROM node_operators;
```

---

## 📚 Project Structure

```
aureo-vpn/
├── cmd/
│   ├── api-gateway/       # API server
│   ├── control-server/    # Control plane
│   └── vpn-node/          # VPN node service
├── internal/
│   ├── api/               # API handlers
│   ├── node/              # Node management
│   └── control/           # Control server logic
├── pkg/
│   ├── models/            # Database models
│   ├── protocols/         # WireGuard implementation
│   └── database/          # Database utilities
├── deployments/docker/    # Docker configs
├── scripts/
│   ├── become-node-operator.sh  # One-command setup
│   ├── add-wireguard-peer.sh    # Peer registration
│   ├── aureo-vpn-mac.sh         # Mac client
│   ├── aureo-vpn-linux.sh       # Linux client
│   └── deploy.sh                # Deployment helper
└── web/operator-dashboard # Web dashboard
```

---

## 🔒 Security Notes

- **Private Keys**: Currently stored in database. In production, use KMS/Vault
- **Passwords**: Hashed with Argon2
- **JWT Tokens**: 24h expiry for access tokens
- **API Auth**: All endpoints require authentication
- **Database**: Postgres with restricted access

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

---

## 📝 License

[Your License Here]

---

## 🆘 Support

- **Documentation**: Check `/docs` folder
- **Issues**: Open an issue on GitHub
- **Email**: support@aureovpn.com

---

## ✨ Credits

Built with:
- Go (Backend)
- WireGuard (VPN Protocol)
- PostgreSQL (Database)
- Docker (Containerization)
- React (Dashboard)
- Fiber (Web Framework)

---

**Happy Earning! 🚀💰**
