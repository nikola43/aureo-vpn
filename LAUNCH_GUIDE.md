# Aureo VPN - Production Launch Guide

This guide describes how to deploy the Aureo VPN platform (API Gateway, Control Server, VPN Nodes, Dashboard) for production.

## 1. Prerequisites

- **Docker & Docker Compose**: Ensure Docker Engine 20.10+ and Docker Compose 2.0+ are installed.
- **Domain Name**: Point your domain (e.g., `vpn.example.com`) to your server's IP.
- **Ports**: Open the following ports in your firewall:
  - `80`/`443` (HTTP/HTTPS for Dashboard & API)
  - `51820` UDP (WireGuard)
  - `1194` UDP (OpenVPN)

## 2. Environment Setup

Run the helper script to generate a secure `.env` file:

```bash
./scripts/generate-env.sh
```

**Critical**: Edit the generated `.env` file and set your blockchain credentials:

```bash
nano .env
# Set ETHEREUM_RPC_URL and ETHEREUM_PRIVATE_KEY
```

## 3. Deployment

Start the entire stack using Docker Compose:

```bash
cd deployments/docker
docker-compose up -d --build
```

This will launch:
1.  **PostgreSQL**: Database
2.  **Redis**: Caching & Rate Limiting
3.  **API Gateway**: Main backend (Port 8080)
4.  **Control Server**: Node management
5.  **VPN Node**: WireGuard/OpenVPN server (Port 51820/1194)
6.  **Dashboard**: Operator UI (Port 5000)
7.  **Prometheus & Grafana**: Monitoring (Port 9090 & 3000)

## 4. Verification

Check the health of the services:

```bash
# Check API Health
curl http://localhost:8080/health
# Response: {"status":"healthy", "database":"connected"}

# Check Dashboard
# Open http://your-server-ip:5000 in your browser
```

## 5. Post-Deployment Steps

### Register as an Operator

1.  Open the Dashboard (`http://your-server-ip:5000`).
2.  Sign up for a new account.
3.  Go to "Register Operator" and connect your wallet.

### Connect a Client

1.  Download the Aureo VPN Client.
2.  Login with a client account (register via API or Client UI).
3.  Select a node and connect.

## 6. Troubleshooting

**View Logs**:
```bash
docker-compose logs -f api-gateway
docker-compose logs -f vpn-node-1
```

**Restart Services**:
```bash
docker-compose restart api-gateway
```

**Database Access**:
```bash
docker-compose exec postgres psql -U postgres -d aureo_vpn
```
