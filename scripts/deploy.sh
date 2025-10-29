#!/bin/bash

################################################################################
# Aureo VPN Deployment Script
#
# Usage:
#   ./deploy.sh rebuild    - Rebuild images and restart all containers
#   ./deploy.sh restart    - Quick restart without rebuild
#   ./deploy.sh logs       - Show logs for all services
#   ./deploy.sh status     - Show container status
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

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

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

# Rebuild and restart
cmd_rebuild() {
    banner
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  Rebuild and Restart${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    echo -e "${YELLOW}[1/6] Stopping containers...${NC}"
    docker-compose down
    echo -e "${GREEN}✓ Containers stopped${NC}"
    echo ""

    echo -e "${YELLOW}[2/6] Removing old containers...${NC}"
    docker-compose rm -f || true
    echo -e "${GREEN}✓ Old containers removed${NC}"
    echo ""

    echo -e "${YELLOW}[3/6] Building Go binaries...${NC}"
    go build -o bin/api-gateway ./cmd/api-gateway
    go build -o bin/vpn-node ./cmd/vpn-node
    echo -e "${GREEN}✓ Go binaries built${NC}"
    echo ""

    echo -e "${YELLOW}[4/6] Building Docker images...${NC}"
    docker-compose build
    echo -e "${GREEN}✓ Docker images built${NC}"
    echo ""

    echo -e "${YELLOW}[5/6] Starting containers...${NC}"
    docker-compose up -d
    echo -e "${GREEN}✓ Containers started${NC}"
    echo ""

    echo -e "${YELLOW}[6/6] Waiting for services...${NC}"
    sleep 5
    echo ""

    docker-compose ps
    echo ""

    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}✓ Deployment complete!${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    show_info
}

# Quick restart
cmd_restart() {
    banner
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  Quick Restart${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    echo -e "${YELLOW}Restarting containers...${NC}"
    docker-compose restart
    echo ""

    docker-compose ps
    echo ""

    echo -e "${GREEN}✓ Restart complete!${NC}"
    echo ""
}

# Show logs
cmd_logs() {
    banner
    echo -e "${CYAN}Showing logs for all services (Ctrl+C to exit)...${NC}"
    echo ""
    docker-compose logs -f --tail=100
}

# Show status
cmd_status() {
    banner
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  Container Status${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    docker-compose ps
    echo ""

    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  Resource Usage${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    docker stats --no-stream $(docker-compose ps -q)
    echo ""
}

# Show helpful information
show_info() {
    echo -e "${CYAN}Useful commands:${NC}"
    echo -e "  View logs:            ${YELLOW}./scripts/deploy.sh logs${NC}"
    echo -e "  Check status:         ${YELLOW}./scripts/deploy.sh status${NC}"
    echo -e "  Quick restart:        ${YELLOW}./scripts/deploy.sh restart${NC}"
    echo ""
    echo -e "${CYAN}Individual service logs:${NC}"
    echo -e "  API Gateway:          ${YELLOW}docker-compose logs -f api-gateway${NC}"
    echo -e "  VPN Node:             ${YELLOW}docker-compose logs -f vpn-node${NC}"
    echo -e "  PostgreSQL:           ${YELLOW}docker-compose logs -f postgres${NC}"
    echo -e "  Redis:                ${YELLOW}docker-compose logs -f redis${NC}"
    echo ""
    echo -e "${CYAN}API Endpoints:${NC}"
    echo -e "  Health:               ${YELLOW}http://localhost:8080/health${NC}"
    echo -e "  API Docs:             ${YELLOW}http://localhost:8080/api/v1${NC}"
    echo ""
}

# Help
cmd_help() {
    banner
    echo "Usage: $0 <command>"
    echo ""
    echo "Commands:"
    echo "  rebuild    - Rebuild images and restart all containers (use after code changes)"
    echo "  restart    - Quick restart without rebuild"
    echo "  logs       - Show logs for all services (follow mode)"
    echo "  status     - Show container status and resource usage"
    echo "  help       - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 rebuild     # After making code changes"
    echo "  $0 restart     # Quick restart without rebuild"
    echo "  $0 logs        # View live logs"
    echo "  $0 status      # Check if containers are running"
    echo ""
}

# Main
COMMAND=${1:-help}

case "$COMMAND" in
    rebuild|build)
        cmd_rebuild
        ;;
    restart|start)
        cmd_restart
        ;;
    logs|log)
        cmd_logs
        ;;
    status|ps)
        cmd_status
        ;;
    help|-h|--help)
        cmd_help
        ;;
    *)
        echo -e "${RED}Unknown command: $COMMAND${NC}"
        echo ""
        cmd_help
        exit 1
        ;;
esac
