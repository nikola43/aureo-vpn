#!/bin/bash

################################################################################
# Quick Restart All Containers (without rebuild)
#
# This script quickly restarts all containers without rebuilding images
################################################################################

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}  Aureo VPN - Quick Restart${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo -e "${YELLOW}Restarting all containers...${NC}"
docker-compose restart
echo -e "${GREEN}✓ All containers restarted${NC}"
echo ""

echo -e "${CYAN}Container status:${NC}"
docker-compose ps
echo ""

echo -e "${GREEN}✓ Done!${NC}"
echo ""
