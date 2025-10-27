#!/bin/bash

################################################################################
# Aureo VPN - Start Script
#
# Starts all VPN Docker services
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
echo -e "${CYAN}  Starting Aureo VPN Services${NC}"
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
    echo -e "${YELLOW}Please start Docker and try again${NC}"
    exit 1
fi

# Navigate to docker directory
cd "$DOCKER_DIR"

echo -e "${CYAN}Building images...${NC}"
docker-compose build

echo ""
echo -e "${CYAN}Starting services...${NC}"
docker-compose up -d

echo ""
echo -e "${CYAN}Waiting for services to be healthy...${NC}"
sleep 5

# Check service status
echo ""
echo -e "${CYAN}Service Status:${NC}"
docker-compose ps

echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✓ Aureo VPN Services Started${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${CYAN}Services available at:${NC}"
echo -e "  ${BLUE}API Gateway:${NC}    http://localhost:8080"
echo -e "  ${BLUE}Grafana:${NC}        http://localhost:3000 (admin/admin)"
echo -e "  ${BLUE}Prometheus:${NC}     http://localhost:9090"
echo -e "  ${BLUE}PostgreSQL:${NC}     localhost:5432"
echo -e "  ${BLUE}Redis:${NC}          localhost:6379"
echo ""
echo -e "${YELLOW}To view logs:${NC}    docker-compose -f $DOCKER_DIR/docker-compose.yml logs -f"
echo -e "${YELLOW}To stop:${NC}        ./scripts/stop-vpn.sh"
echo ""
