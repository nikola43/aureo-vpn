# Aureo VPN Node Operator Guide

A comprehensive guide for running your own Aureo VPN node and earning cryptocurrency rewards.

## Quick Start

```bash
# 1. Register as user (API)
curl -X POST https://api.aureo-vpn.com/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"you@example.com","username":"operator1","password":"secure123"}'

# 2. Register as operator (with your crypto wallet)
curl -X POST https://api.aureo-vpn.com/api/v1/operator/register \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"wallet_address":"0xYourEthWallet","wallet_currency":"ethereum"}'

# 3. Create node (after admin verification)
curl -X POST https://api.aureo-vpn.com/api/v1/operator/nodes \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"my-node","ip_address":"YOUR_SERVER_IP","country":"US"}'

# 4. Deploy node
docker run -d --name aureo-node \
  --cap-add=NET_ADMIN \
  -e NODE_ID=YOUR_NODE_ID \
  -e NODE_PRIVATE_KEY=YOUR_PRIVATE_KEY \
  -p 51820:51820/udp \
  ghcr.io/nikola43/aureo-vpn-node:latest
```

---

## Reward Tiers

| Tier | Rate/GB | Min Uptime | Min Reputation | Bonus |
|------|---------|------------|----------------|-------|
| **Bronze** | $0.010 | 50% | 0 | 1.0x |
| **Silver** | $0.015 | 80% | 60 | 1.2x |
| **Gold** | $0.020 | 90% | 75 | 1.5x |
| **Platinum** | $0.030 | 95% | 90 | 2.0x |

### Earning Example

```
Bandwidth: 100 GB
Tier: Gold ($0.02/GB)
Quality Score: 90%
Duration: 5 hours (1.2x bonus)

Base: 100 × $0.02 = $2.00
Quality: × (0.5 + 0.9) = $2.80
Duration Bonus: × 1.2 = $3.36
Tier Bonus: × 1.5 = $5.04

Total Earnings: $5.04
```

---

## Server Requirements

### Minimum
- 2 CPU cores
- 2 GB RAM
- 50 GB SSD
- 100 Mbps network
- Static IP

### Recommended
- 4+ CPU cores
- 8 GB RAM
- 100 GB SSD
- 1 Gbps network
- Static IP
- DDoS protection

### Required Ports

| Port | Protocol | Service |
|------|----------|---------|
| 51820 | UDP | WireGuard |
| 1194 | UDP | OpenVPN (optional) |
| 4001 | TCP | P2P Network (optional) |

---

## Detailed Setup

### Step 1: Prepare Server (Ubuntu 22.04)

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install dependencies
sudo apt install -y wireguard wireguard-tools curl jq

# Install Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Enable IP forwarding
cat <<EOF | sudo tee /etc/sysctl.d/99-aureo.conf
net.ipv4.ip_forward = 1
net.ipv6.conf.all.forwarding = 1
net.ipv4.conf.all.src_valid_mark = 1
EOF
sudo sysctl -p /etc/sysctl.d/99-aureo.conf

# Reboot or re-login to apply Docker group
```

### Step 2: Configure Firewall

```bash
# UFW (Recommended)
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 51820/udp comment "WireGuard"
sudo ufw allow 1194/udp comment "OpenVPN"
sudo ufw allow 4001/tcp comment "P2P"
sudo ufw enable

# Or iptables
sudo iptables -A INPUT -p udp --dport 51820 -j ACCEPT
sudo iptables -A INPUT -p udp --dport 1194 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 4001 -j ACCEPT
sudo iptables -A FORWARD -i wg0 -j ACCEPT
sudo iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
```

### Step 3: Register Operator

```bash
# Set API URL
API_URL="https://api.aureo-vpn.com"

# Register account
REGISTER_RESP=$(curl -s -X POST $API_URL/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "operator@example.com",
    "username": "my_operator",
    "password": "SecurePassword123!"
  }')

# Extract token
TOKEN=$(echo $REGISTER_RESP | jq -r '.access_token')
echo "Access Token: $TOKEN"

# Register as operator
curl -X POST $API_URL/api/v1/operator/register \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "wallet_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f5bC3F",
    "wallet_currency": "ethereum",
    "company_name": "My VPN Nodes",
    "contact_email": "contact@mynodes.com"
  }'

echo "Operator registration submitted. Wait for admin verification."
```

### Step 4: Create Node Entry

After verification:

```bash
# Create node
NODE_RESP=$(curl -s -X POST $API_URL/api/v1/operator/nodes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "us-east-node-1",
    "hostname": "us1.myvpn.com",
    "ip_address": "203.0.113.10",
    "country": "US",
    "city": "New York",
    "latitude": 40.7128,
    "longitude": -74.0060,
    "wireguard_port": 51820,
    "openvpn_port": 1194,
    "max_connections": 500
  }')

# Extract node credentials - SAVE THESE!
NODE_ID=$(echo $NODE_RESP | jq -r '.node.id')
PRIVATE_KEY=$(echo $NODE_RESP | jq -r '.private_key')
PUBLIC_KEY=$(echo $NODE_RESP | jq -r '.node.wireguard_public_key')

echo "Node ID: $NODE_ID"
echo "Private Key: $PRIVATE_KEY"
echo "Public Key: $PUBLIC_KEY"

# Save to file
cat > node-credentials.txt <<EOF
NODE_ID=$NODE_ID
PRIVATE_KEY=$PRIVATE_KEY
PUBLIC_KEY=$PUBLIC_KEY
EOF

chmod 600 node-credentials.txt
```

### Step 5: Deploy Node

**Option A: Docker Compose (Recommended)**

```yaml
# docker-compose.yml
version: '3.8'

services:
  vpn-node:
    image: ghcr.io/nikola43/aureo-vpn-node:latest
    container_name: aureo-node
    restart: unless-stopped
    environment:
      - NODE_ID=${NODE_ID}
      - NODE_PRIVATE_KEY=${PRIVATE_KEY}
      - API_URL=https://api.aureo-vpn.com
      - API_TOKEN=${TOKEN}
      - WIREGUARD_PORT=51820
      - LOG_LEVEL=info
    cap_add:
      - NET_ADMIN
      - SYS_MODULE
    sysctls:
      - net.ipv4.ip_forward=1
      - net.ipv4.conf.all.src_valid_mark=1
    devices:
      - /dev/net/tun:/dev/net/tun
    ports:
      - "51820:51820/udp"
      - "1194:1194/udp"
    volumes:
      - /lib/modules:/lib/modules:ro
    healthcheck:
      test: ["CMD", "wg", "show", "wg0"]
      interval: 30s
      timeout: 10s
      retries: 3
```

```bash
# Create .env file
cat > .env <<EOF
NODE_ID=your-node-id
PRIVATE_KEY=your-private-key
TOKEN=your-api-token
EOF

# Start node
docker-compose up -d

# View logs
docker-compose logs -f
```

**Option B: Docker Run**

```bash
docker run -d \
  --name aureo-node \
  --restart unless-stopped \
  --cap-add NET_ADMIN \
  --cap-add SYS_MODULE \
  --sysctl net.ipv4.ip_forward=1 \
  --device /dev/net/tun:/dev/net/tun \
  -p 51820:51820/udp \
  -p 1194:1194/udp \
  -e NODE_ID="$NODE_ID" \
  -e NODE_PRIVATE_KEY="$PRIVATE_KEY" \
  -e API_URL="https://api.aureo-vpn.com" \
  -e API_TOKEN="$TOKEN" \
  ghcr.io/nikola43/aureo-vpn-node:latest
```

**Option C: Binary (Manual)**

```bash
# Clone and build
git clone https://github.com/nikola43/aureo-vpn.git
cd aureo-vpn
make build-node

# Create config
sudo mkdir -p /etc/aureo-vpn
cat <<EOF | sudo tee /etc/aureo-vpn/.env
NODE_ID=$NODE_ID
NODE_PRIVATE_KEY=$PRIVATE_KEY
API_URL=https://api.aureo-vpn.com
API_TOKEN=$TOKEN
WIREGUARD_PORT=51820
LOG_LEVEL=info
EOF

# Setup WireGuard interface
sudo ip link add dev wg0 type wireguard

# Write private key
echo "$PRIVATE_KEY" | sudo tee /etc/wireguard/privatekey
sudo chmod 600 /etc/wireguard/privatekey

# Configure interface
sudo wg set wg0 \
  private-key /etc/wireguard/privatekey \
  listen-port 51820

sudo ip addr add 10.8.0.1/24 dev wg0
sudo ip link set wg0 up

# Run node
sudo ./bin/vpn-node
```

### Step 6: Verify Node Status

```bash
# Check WireGuard interface
wg show wg0

# Check node appears online
curl -s -X GET "$API_URL/api/v1/nodes" \
  -H "Authorization: Bearer $TOKEN" | \
  jq '.[] | select(.id == "'$NODE_ID'")'

# Expected output:
# {
#   "id": "your-node-id",
#   "name": "us-east-node-1",
#   "status": "online",
#   "load_score": 0,
#   ...
# }
```

---

## Monitoring

### Dashboard API

```bash
# Get dashboard stats
curl -s -X GET "$API_URL/api/v1/operator/dashboard" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Response:
# {
#   "total_earned_usd": 156.78,
#   "pending_payout_usd": 23.45,
#   "active_nodes": 1,
#   "total_bandwidth_gb": 5678.90,
#   "average_uptime_percent": 99.2,
#   "reputation_score": 87,
#   "current_tier": "gold",
#   "nodes": [...]
# }
```

### View Earnings

```bash
# List recent earnings
curl -s -X GET "$API_URL/api/v1/operator/earnings?limit=10" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# View earnings for specific node
curl -s -X GET "$API_URL/api/v1/operator/earnings?node_id=$NODE_ID" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

### Real-time Monitoring Script

```bash
#!/bin/bash
# monitor.sh - Run: watch -n10 ./monitor.sh

API_URL="https://api.aureo-vpn.com"
TOKEN="YOUR_TOKEN"

# Get stats
STATS=$(curl -s -X GET "$API_URL/api/v1/operator/dashboard" \
  -H "Authorization: Bearer $TOKEN")

# Display
echo "========== AUREO NODE STATS =========="
echo "Earned Total:    \$$(echo $STATS | jq -r '.total_earned_usd')"
echo "Pending Payout:  \$$(echo $STATS | jq -r '.pending_payout_usd')"
echo "Bandwidth Today: $(echo $STATS | jq -r '.total_bandwidth_gb') GB"
echo "Active Sessions: $(docker exec aureo-node wg show wg0 | grep -c peer)"
echo "Uptime:          $(echo $STATS | jq -r '.average_uptime_percent')%"
echo "Reputation:      $(echo $STATS | jq -r '.reputation_score')/100"
echo "Tier:            $(echo $STATS | jq -r '.current_tier')"
echo "======================================"
```

---

## Payouts

### Request Payout

```bash
# Check pending balance
BALANCE=$(curl -s -X GET "$API_URL/api/v1/operator/dashboard" \
  -H "Authorization: Bearer $TOKEN" | jq -r '.pending_payout_usd')

echo "Pending balance: \$$BALANCE"

# Request payout (minimum $10)
if (( $(echo "$BALANCE >= 10" | bc -l) )); then
  curl -X POST "$API_URL/api/v1/operator/payout/request" \
    -H "Authorization: Bearer $TOKEN"
else
  echo "Minimum payout is $10. Current balance: \$$BALANCE"
fi
```

### View Payout History

```bash
curl -s -X GET "$API_URL/api/v1/operator/payouts" \
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

## Maintenance

### Update Node

```bash
# Docker Compose
docker-compose pull
docker-compose up -d

# Docker Run
docker pull ghcr.io/nikola43/aureo-vpn-node:latest
docker stop aureo-node
docker rm aureo-node
# Re-run docker run command
```

### Restart Node

```bash
docker-compose restart
# or
docker restart aureo-node
```

### View Logs

```bash
# All logs
docker-compose logs -f

# Last 100 lines
docker-compose logs --tail=100

# Filter by level
docker-compose logs | grep ERROR
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
# Check if container is running
docker ps | grep aureo-node

# Check logs for errors
docker logs aureo-node --tail=50

# Verify WireGuard interface
docker exec aureo-node wg show wg0

# Check API connectivity
curl -I https://api.aureo-vpn.com/health

# Restart container
docker restart aureo-node
```

### No Connections

```bash
# Verify port is accessible
nc -zvu YOUR_SERVER_IP 51820

# Check firewall
sudo iptables -L -n | grep 51820
sudo ufw status

# Test from external server
ssh external-server 'nc -zvu YOUR_SERVER_IP 51820'
```

### Low Earnings

1. **Check uptime** - Maintain >95% for Platinum tier
2. **Check quality score** - Reduce latency, minimize packet loss
3. **Check location** - Popular regions have more users
4. **Check bandwidth** - Higher capacity = more connections

### WireGuard Issues

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

### High Load

```bash
# Check current connections
docker exec aureo-node wg show wg0 | grep -c peer

# Check system resources
docker stats aureo-node

# Reduce max connections in API
curl -X PUT "$API_URL/api/v1/operator/nodes/$NODE_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"max_connections": 300}'
```

---

## Security Best Practices

1. **Keep credentials secure**
   - Never commit credentials to git
   - Use environment variables
   - Restrict file permissions (chmod 600)

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
   - Have recovery plan

---

## Support

- **Documentation**: https://docs.aureo-vpn.com
- **Discord**: https://discord.gg/aureo-vpn
- **Issues**: https://github.com/nikola43/aureo-vpn/issues
- **Email**: operators@aureo-vpn.com

---

## FAQ

**Q: How long until I start earning?**
A: Earnings begin as soon as users connect to your node. This depends on your location and network capacity.

**Q: What's the minimum payout?**
A: $10 USD equivalent in your chosen cryptocurrency.

**Q: How often are payouts processed?**
A: You can request a payout anytime once you reach the minimum. Transactions typically complete within 10-30 minutes.

**Q: Can I run multiple nodes?**
A: Yes, up to 10 nodes per operator account.

**Q: What happens if my node goes offline?**
A: Your uptime percentage decreases, which may affect your tier. Active sessions are gracefully disconnected.

**Q: How is reputation calculated?**
A: Based on uptime, connection quality, user feedback, and time as an operator.
