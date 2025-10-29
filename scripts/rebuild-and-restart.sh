#!/bin/bash

################################################################################
# Rebuild and Restart All Containers
#
# This script rebuilds Docker images and restarts all containers with new changes
################################################################################

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}  Aureo VPN - Rebuild and Restart${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo -e "${YELLOW}Step 1/5: Stopping all containers...${NC}"
docker-compose down || true
echo -e "${GREEN}✓ Containers stopped${NC}"
echo ""

echo -e "${YELLOW}Step 2/5: Cleaning up old images (optional)...${NC}"
read -p "$(echo -e ${CYAN}Do you want to remove old images? [y/N]: ${NC})" REMOVE_IMAGES
if [[ "$REMOVE_IMAGES" =~ ^[Yy]$ ]]; then
    docker-compose rm -f || true
    echo -e "${GREEN}✓ Old images removed${NC}"
else
    echo -e "${BLUE}Skipping image removal${NC}"
fi
echo ""

echo -e "${YELLOW}Step 3/5: Building new images with latest changes...${NC}"
docker-compose build --no-cache
echo -e "${GREEN}✓ Images built successfully${NC}"
echo ""

echo -e "${YELLOW}Step 4/5: Starting all containers...${NC}"
docker-compose up -d
echo -e "${GREEN}✓ Containers started${NC}"
echo ""

echo -e "${YELLOW}Step 5/5: Checking container status...${NC}"
echo ""
docker-compose ps
echo ""

echo -e "${YELLOW}Waiting for services to be healthy...${NC}"
sleep 5

echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✓ All services restarted successfully!${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

echo -e "${CYAN}Service logs:${NC}"
echo -e "  View all logs:        ${YELLOW}docker-compose logs -f${NC}"
echo -e "  View API logs:        ${YELLOW}docker-compose logs -f api-gateway${NC}"
echo -e "  View VPN node logs:   ${YELLOW}docker-compose logs -f vpn-node${NC}"
echo -e "  View DB logs:         ${YELLOW}docker-compose logs -f postgres${NC}"
echo ""

echo -e "${CYAN}Quick commands:${NC}"
echo -e "  Stop all:             ${YELLOW}docker-compose stop${NC}"
echo -e "  Restart all:          ${YELLOW}docker-compose restart${NC}"
echo -e "  View status:          ${YELLOW}docker-compose ps${NC}"
echo ""
