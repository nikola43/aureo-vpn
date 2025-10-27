#!/bin/bash

################################################################################
# Aureo VPN - Stop Script
#
# Stops all VPN Docker services
################################################################################

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DOCKER_DIR="$PROJECT_ROOT/deployments/docker"

# Banner
banner() {
    echo -e "${PURPLE}"
    cat << "EOF"
   _                         __     ______  _   _
  / \  _   _ _ __ ___  ___   \ \   / /  _ \| \ | |
 / _ \| | | | '__/ _ \/ _ \   \ \ / /| |_) |  \| |
/ ___ \ |_| | | |  __/ (_) |   \ V / |  __/| |\  |
/_/   \_\__,_|_|  \___|\___/     \_/  |_|   |_| \_|

EOF
    echo -e "${NC}"
}

banner

echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}  Stopping Aureo VPN Services${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Check if docker-compose.yml exists
if [ ! -f "$DOCKER_DIR/docker-compose.yml" ]; then
    echo -e "${RED}✗ docker-compose.yml not found at $DOCKER_DIR${NC}"
    exit 1
fi

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
    echo -e "${RED}✗ Docker is not running${NC}"
    exit 1
fi

# Navigate to docker directory
cd "$DOCKER_DIR"

# Check if any containers are running
if [ -z "$(docker-compose ps -q)" ]; then
    echo -e "${YELLOW}No running services found${NC}"
    exit 0
fi

echo -e "${CYAN}Stopping services...${NC}"
docker-compose down

echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✓ Aureo VPN Services Stopped${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${YELLOW}To start services:${NC}     ./scripts/start-vpn.sh"
echo -e "${YELLOW}To remove volumes:${NC}     docker-compose -f $DOCKER_DIR/docker-compose.yml down -v"
echo ""
