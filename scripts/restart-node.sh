#!/bin/bash

#
# Aureo VPN - Restart VPN Node Script
#
# This script restarts the VPN node service
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Project paths
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONFIG_DIR="/tmp/.aureo-vpn-root"
DOCKER_COMPOSE_FILE="$PROJECT_ROOT/deployments/docker/docker-compose.yml"

# Detect Docker Compose command
detect_docker_compose() {
    if docker compose version >/dev/null 2>&1; then
        DOCKER_COMPOSE="docker compose"
    elif command -v docker-compose >/dev/null 2>&1; then
        DOCKER_COMPOSE="docker-compose"
    else
        echo -e "${RED}Error: Neither 'docker compose' nor 'docker-compose' found${NC}"
        exit 1
    fi
}

# Banner
banner() {
    echo -e "${CYAN}"
    echo "   _                         __     ______  _   _"
    echo "  / \  _   _ _ __ ___  ___   \ \   / /  _ \| \ | |"
    echo " / _ \| | | | '__/ _ \/ _ \   \ \ / /| |_) |  \| |"
    echo "/ ___ \ |_| | | |  __/ (_) |   \ V / |  __/| |\  |"
    echo "/_/   \_\__,_|_|  \___|\___/     \_/  |_|   |_| \_|"
    echo -e "${NC}"
    echo ""
}

section() {
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  $1${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Please run as root (use sudo)${NC}"
    exit 1
fi

banner

detect_docker_compose

section "🔄 Restarting VPN Node"

# Check if .env file exists
if [ ! -f "$PROJECT_ROOT/deployments/docker/.env" ]; then
    echo -e "${RED}✗ .env file not found${NC}"
    echo -e "${YELLOW}Please run become-node-operator.sh first${NC}"
    exit 1
fi

# Get NODE_ID from .env
NODE_ID=$(grep NODE_ID_1 "$PROJECT_ROOT/deployments/docker/.env" | cut -d'=' -f2)

if [ -z "$NODE_ID" ]; then
    echo -e "${RED}✗ NODE_ID not found in .env file${NC}"
    exit 1
fi

echo -e "${BLUE}Node ID: $NODE_ID${NC}"
echo ""

# Stop VPN node
echo -e "${YELLOW}Stopping VPN node...${NC}"
docker rm -f aureo-vpn-node-1 2>/dev/null || true
echo -e "${GREEN}✓ VPN node stopped${NC}"

# Start VPN node
echo -e "${YELLOW}Starting VPN node...${NC}"
docker run -d \
    --name aureo-vpn-node-1 \
    --network docker_aureo-network \
    --cap-add NET_ADMIN \
    --cap-add SYS_MODULE \
    --sysctl net.ipv4.ip_forward=1 \
    --sysctl net.ipv6.conf.all.forwarding=1 \
    -e NODE_ID=$NODE_ID \
    -e DB_HOST=postgres \
    -e REDIS_HOST=redis \
    docker_vpn-node-1:latest >/dev/null

echo -e "${GREEN}✓ VPN node started${NC}"

# Wait for node to start
echo ""
echo -e "${YELLOW}Waiting for node to initialize...${NC}"
sleep 5

# Check if node is running
if docker ps | grep -q aureo-vpn-node-1; then
    echo -e "${GREEN}✓ Node is running${NC}"

    # Get WireGuard public key
    echo ""
    echo -e "${YELLOW}Getting WireGuard public key...${NC}"
    WG_PUBLIC_KEY=$(docker exec aureo-vpn-node-1 wg show wg0 public-key 2>/dev/null || echo "")

    if [ -n "$WG_PUBLIC_KEY" ]; then
        echo -e "${GREEN}✓ Public Key: $WG_PUBLIC_KEY${NC}"

        # Update database
        echo ""
        echo -e "${YELLOW}Updating database...${NC}"
        docker exec aureo-vpn-db psql -U postgres -d aureo_vpn -c \
            "UPDATE vpn_nodes SET public_key='$WG_PUBLIC_KEY', status='online', last_heartbeat=NOW() WHERE id='$NODE_ID';" \
            >/dev/null 2>&1

        echo -e "${GREEN}✓ Database updated${NC}"
    else
        echo -e "${YELLOW}⚠ Could not retrieve WireGuard public key${NC}"
    fi

    # Show node status
    echo ""
    section "📊 Node Status"
    docker exec aureo-vpn-node-1 wg show wg0 2>/dev/null || echo "WireGuard interface not ready yet"

else
    echo -e "${RED}✗ Node failed to start${NC}"
    echo ""
    echo -e "${YELLOW}Logs:${NC}"
    docker logs aureo-vpn-node-1 --tail 20
    exit 1
fi

echo ""
section "✅ VPN Node Restarted Successfully"
echo ""
echo -e "${GREEN}Your VPN node is now running!${NC}"
echo ""
