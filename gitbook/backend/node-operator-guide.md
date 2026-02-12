# 📡 Node Operator Guide

A comprehensive guide for running your own Aureo VPN node and earning cryptocurrency rewards.

---

## Overview

Running a VPN node allows you to:

- Earn cryptocurrency rewards for bandwidth served
- Contribute to the decentralized network
- Help users access private, secure internet

You can earn up to **$0.030/GB** at the Platinum tier, with additional bonuses for quality and session duration.

---

## Requirements

### Hardware

**Minimum**:
- 2 CPU cores
- 2 GB RAM
- 50 GB SSD
- 100 Mbps network
- Static IP address

**Recommended**:
- 4+ CPU cores
- 8 GB RAM
- 100 GB SSD
- 1 Gbps network
- Static IP address
- DDoS protection

### Software

- Linux (Ubuntu 22.04 recommended)
- Docker & Docker Compose
- WireGuard kernel module

### Network

| Port | Protocol | Service | Required |
|------|----------|---------|----------|
| 51820 | UDP | WireGuard | Yes |
| 1194 | UDP | OpenVPN | Optional |
| 4001 | TCP | P2P Network | Optional |

---

## Step 1: Register as Operator

First, create a user account and register as a node operator:

```bash
# Register user account
curl -X POST https://api.aureo-vpn.com/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "operator@example.com",
    "username": "my_node_operator",
    "password": "secure_password_123"
  }'

# Save the access token from response
export TOKEN="<access_token_from_response>"

# Register as operator with your crypto wallet
curl -X POST https://api.aureo-vpn.com/api/v1/operator/register \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "wallet_address": "0xYourEthereumWalletAddress",
    "wallet_currency": "ethereum",
    "company_name": "My VPN Nodes LLC",
    "contact_email": "operator@example.com"
  }'
```

Wait for admin verification (you will receive an email).

---

## Step 2: Create Your Node

After verification, create a node entry:

```bash
curl -X POST https://api.aureo-vpn.com/api/v1/operator/nodes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "us-east-node-1",
    "hostname": "vpn-us.example.com",
    "ip_address": "YOUR_SERVER_IP",
    "country": "US",
    "city": "New York",
    "latitude": 40.7128,
    "longitude": -74.0060,
    "wireguard_port": 51820,
    "openvpn_port": 1194,
    "max_connections": 500
  }'

# Response includes:
# - node_id
# - private_key (save this securely!)
# - public_key
# - subnet (e.g., 10.8.0.0/24)
```

{% hint style="danger" %}
Save the `node_id` and `private_key` from the response immediately. The private key is only returned once and cannot be retrieved again.
{% endhint %}

---

## Step 3: Server Setup

SSH into your server and install dependencies:

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install WireGuard
sudo apt install -y wireguard wireguard-tools

# Install Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Enable IP forwarding
echo "net.ipv4.ip_forward = 1" | sudo tee /etc/sysctl.d/99-wireguard.conf
echo "net.ipv6.conf.all.forwarding = 1" | sudo tee -a /etc/sysctl.d/99-wireguard.conf
sudo sysctl -p /etc/sysctl.d/99-wireguard.conf
```

---

## Step 4: Configure the Node

Create configuration directory:

```bash
sudo mkdir -p /etc/aureo-vpn
sudo chmod 700 /etc/aureo-vpn
```

Create environment file `/etc/aureo-vpn/.env`:

```bash
# Node Identity (from Step 2)
NODE_ID=your-node-uuid-here
NODE_PRIVATE_KEY=your-wireguard-private-key

# API Connection
API_URL=https://api.aureo-vpn.com
API_TOKEN=your-operator-access-token

# VPN Configuration
WIREGUARD_PORT=51820
WIREGUARD_INTERFACE=wg0
VPN_SUBNET=10.8.0.0/24
VPN_GATEWAY=10.8.0.1

# Server Info
NODE_NAME=us-east-node-1
NODE_COUNTRY=US
NODE_CITY=New York

# P2P Network (optional)
P2P_ENABLED=true
P2P_PORT=4001

# Logging
LOG_LEVEL=info
```

---

## Step 5: Run with Docker

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  vpn-node:
    image: ghcr.io/nikola43/aureo-vpn-node:latest
    container_name: aureo-vpn-node
    restart: unless-stopped
    env_file:
      - /etc/aureo-vpn/.env
    cap_add:
      - NET_ADMIN
      - SYS_MODULE
    sysctls:
      - net.ipv4.ip_forward=1
      - net.ipv4.conf.all.src_valid_mark=1
    volumes:
      - /etc/aureo-vpn:/etc/aureo-vpn:ro
      - /lib/modules:/lib/modules:ro
    devices:
      - /dev/net/tun:/dev/net/tun
    ports:
      - "51820:51820/udp"
      - "1194:1194/udp"
      - "4001:4001/tcp"
    networks:
      - vpn-network
    healthcheck:
      test: ["CMD", "wg", "show", "wg0"]
      interval: 30s
      timeout: 10s
      retries: 3

networks:
  vpn-network:
    driver: bridge
```

Start the node:

```bash
docker-compose up -d

# Check logs
docker-compose logs -f

# Verify WireGuard interface
docker exec aureo-vpn-node wg show
```

---

## Step 6: Manual Setup (Alternative)

If you prefer running without Docker:

```bash
# Clone repository
git clone https://github.com/nikola43/aureo-vpn.git
cd aureo-vpn

# Build VPN node binary
make build-node

# Configure WireGuard interface
sudo ip link add dev wg0 type wireguard
sudo wg set wg0 private-key /etc/aureo-vpn/privatekey listen-port 51820
sudo ip addr add 10.8.0.1/24 dev wg0
sudo ip link set wg0 up

# Configure firewall (iptables)
sudo iptables -A FORWARD -i wg0 -j ACCEPT
sudo iptables -A FORWARD -o wg0 -j ACCEPT
sudo iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE

# Run the node
sudo ./bin/vpn-node
```

---

## Step 7: Configure Firewall

```bash
# UFW (Ubuntu)
sudo ufw allow 51820/udp comment "WireGuard VPN"
sudo ufw allow 1194/udp comment "OpenVPN"
sudo ufw allow 4001/tcp comment "P2P Network"

# Or iptables
sudo iptables -A INPUT -p udp --dport 51820 -j ACCEPT
sudo iptables -A INPUT -p udp --dport 1194 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 4001 -j ACCEPT
```

---

## Step 8: Verify Node Status

```bash
# Check node is visible in network
curl -X GET https://api.aureo-vpn.com/api/v1/nodes \
  -H "Authorization: Bearer $TOKEN" | jq '.[] | select(.id == "YOUR_NODE_ID")'

# Response should show:
# {
#   "id": "your-node-id",
#   "status": "online",
#   "load_score": 0,
#   "connections": 0,
#   ...
# }
```

---

## Step 9: Monitor Your Node

```bash
# View operator dashboard
curl -X GET https://api.aureo-vpn.com/api/v1/operator/dashboard \
  -H "Authorization: Bearer $TOKEN"

# Response:
# {
#   "total_earned_usd": 45.67,
#   "pending_payout_usd": 12.34,
#   "active_nodes": 1,
#   "total_bandwidth_gb": 2345.67,
#   "average_uptime_percent": 99.5,
#   "reputation_score": 85,
#   "current_tier": "gold"
# }

# View earnings history
curl -X GET https://api.aureo-vpn.com/api/v1/operator/earnings \
  -H "Authorization: Bearer $TOKEN"
```

### Real-time Monitoring Script

```bash
#!/bin/bash
# monitor.sh - Run: watch -n10 ./monitor.sh

API_URL="https://api.aureo-vpn.com"
TOKEN="YOUR_TOKEN"

STATS=$(curl -s -X GET "$API_URL/api/v1/operator/dashboard" \
  -H "Authorization: Bearer $TOKEN")

echo "========== AUREO NODE STATS =========="
echo "Earned Total:    \$$(echo $STATS | jq -r '.total_earned_usd')"
echo "Pending Payout:  \$$(echo $STATS | jq -r '.pending_payout_usd')"
echo "Bandwidth Today: $(echo $STATS | jq -r '.total_bandwidth_gb') GB"
echo "Active Sessions: $(docker exec aureo-vpn-node wg show wg0 | grep -c peer)"
echo "Uptime:          $(echo $STATS | jq -r '.average_uptime_percent')%"
echo "Reputation:      $(echo $STATS | jq -r '.reputation_score')/100"
echo "Tier:            $(echo $STATS | jq -r '.current_tier')"
echo "======================================"
```

---

## Step 10: Request Payout

When your pending balance exceeds $10:

```bash
curl -X POST https://api.aureo-vpn.com/api/v1/operator/payout/request \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json"

# Response:
# {
#   "payout_id": "uuid",
#   "amount_usd": 45.67,
#   "amount_crypto": "0.0228",
#   "currency": "ethereum",
#   "status": "pending",
#   "estimated_completion": "2024-01-15T12:00:00Z"
# }
```

### View Payout History

```bash
curl -s -X GET https://api.aureo-vpn.com/api/v1/operator/payouts \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Response:
# [
#   {
#     "id": "uuid",
#     "amount_usd": 45.67,
#     "amount_crypto": "0.0228",
#     "currency": "ethereum",
#     "transaction_hash": "0x...",
#     "status": "completed",
#     "processed_at": "2024-01-15T12:00:00Z"
#   }
# ]
```

---

## Reward Tiers

| Tier | Rate/GB | Min Uptime | Min Reputation | Bonus |
|------|---------|------------|----------------|-------|
| 🥉 Bronze | $0.010 | 50% | 0 | 1.0x |
| 🥈 Silver | $0.015 | 80% | 60 | 1.2x |
| 🥇 Gold | $0.020 | 90% | 75 | 1.5x |
| 💎 Platinum | $0.030 | 95% | 90 | 2.0x |

### Earning Example

```
Bandwidth: 100 GB
Tier: Gold ($0.02/GB)
Quality Score: 90%
Duration: 5 hours (1.2x bonus)

Base: 100 x $0.02 = $2.00
Quality: x (0.5 + 0.9) = $2.80
Duration Bonus: x 1.2 = $3.36
Tier Bonus: x 1.5 = $5.04

Total Earnings: $5.04
```

---

## Node Maintenance

```bash
# Update node software
docker-compose pull
docker-compose up -d

# View real-time stats
docker exec aureo-vpn-node wg show wg0

# Check peer connections
docker exec aureo-vpn-node wg show wg0 dump | wc -l

# Restart node
docker-compose restart

# Stop node gracefully
docker-compose down
```

### Backup Credentials

```bash
# Backup critical files
tar -czvf aureo-backup-$(date +%Y%m%d).tar.gz \
  node-credentials.txt \
  .env \
  docker-compose.yml

# Store securely (encrypted, offsite)
```

---

## Troubleshooting

### Node Shows Offline

```bash
# Check if process is running
docker ps | grep aureo-vpn-node

# Check logs for errors
docker-compose logs --tail=100

# Verify API connectivity
curl -I https://api.aureo-vpn.com/health

# Restart container
docker restart aureo-vpn-node
```

### No Connections Received

```bash
# Verify port is open
sudo netstat -ulnp | grep 51820

# Test from external host
nc -zvu your-server-ip 51820

# Check firewall
sudo iptables -L -n | grep 51820
sudo ufw status
```

### WireGuard Interface Issues

```bash
# Recreate interface
docker-compose down
sudo ip link delete wg0 2>/dev/null || true
docker-compose up -d

# Check kernel module
lsmod | grep wireguard

# Install module if missing
sudo apt install wireguard-dkms
```

### Low Earnings

1. **Check uptime** - Maintain >95% for Platinum tier
2. **Check quality score** - Reduce latency, minimize packet loss
3. **Check location** - Popular regions have more users
4. **Check bandwidth** - Higher capacity = more connections

### High Load

```bash
# Check current connections
docker exec aureo-vpn-node wg show wg0 | grep -c peer

# Check system resources
docker stats aureo-vpn-node

# Reduce max connections in API
curl -X PUT "https://api.aureo-vpn.com/api/v1/operator/nodes/$NODE_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"max_connections": 300}'
```

---

## Security Best Practices

1. **Keep credentials secure**
   - Never commit credentials to git
   - Use environment variables
   - Restrict file permissions (`chmod 600`)

2. **Regular updates**
   - Update OS regularly
   - Pull latest node image weekly
   - Monitor security advisories

3. **Network security**
   - Use firewall (UFW/iptables)
   - Disable password SSH (use keys)
   - Consider fail2ban

4. **Monitoring**
   - Set up alerts for downtime
   - Monitor disk space
   - Watch for unusual traffic

5. **Backups**
   - Backup credentials securely
   - Document your setup
   - Have a recovery plan

---

## FAQ

**Q: How long until I start earning?**
A: Earnings begin as soon as users connect to your node. This depends on your location and network capacity.

**Q: What is the minimum payout?**
A: $10 USD equivalent in your chosen cryptocurrency.

**Q: How often are payouts processed?**
A: You can request a payout anytime once you reach the minimum. Transactions typically complete within 10-30 minutes.

**Q: Can I run multiple nodes?**
A: Yes, up to 10 nodes per operator account.

**Q: What happens if my node goes offline?**
A: Your uptime percentage decreases, which may affect your tier. Active sessions are gracefully disconnected.

**Q: How is reputation calculated?**
A: Based on uptime, connection quality, user feedback, and time as an operator.
