#!/bin/bash

################################################################################
# Aureo VPN - Node Restart Script
#
# Restarts VPN node and related services safely
#
# Usage:
#   sudo bash restart-node.sh [options]
#
# Options:
#   --all           Restart all services (database, redis, api, vpn node)
#   --node-only     Restart only VPN node (default)
#   --api           Restart API gateway
#   --control       Restart control server
#   --force         Force restart (skip confirmations)
#
################################################################################

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Project paths
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONFIG_DIR="$HOME/.aureo-vpn"
DOCKER_COMPOSE_FILE="$PROJECT_ROOT/deployments/docker/docker-compose.yml"

# Default options
RESTART_ALL=false
RESTART_NODE=true
RESTART_API=false
RESTART_CONTROL=false
FORCE=false
DEPLOYMENT_MODE=""

# Trap errors
trap 'error_handler $? $LINENO' ERR

error_handler() {
    echo -e "\n${RED}✗ Error occurred at line $2 (exit code: $1)${NC}"
    echo -e "${YELLOW}Restart failed. Check the errors above.${NC}"
    exit 1
}

# Detect Docker Compose command
detect_docker_compose() {
    if docker compose version >/dev/null 2>&1; then
        DOCKER_COMPOSE="docker compose"
    elif command -v docker-compose >/dev/null 2>&1; then
        DOCKER_COMPOSE="docker-compose"
    else
        return 1
    fi
    return 0
}

# Command exists helper
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Banner
print_header() {
    clear
    echo -e "${PURPLE}"
    cat << "EOF"
╔══════════════════════════════════════════════════════════════════╗
║                                                                  ║
║       🔄 Aureo VPN - Node Restart Script 🔄                      ║
║                                                                  ║
║     Safely restart your VPN node and related services           ║
║                                                                  ║
╚══════════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}\n"
}

# Section header
section() {
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  $1${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

# Check if running as root
check_root() {
    if [ "$EUID" -ne 0 ]; then
        echo -e "${RED}Please run as root (use sudo)${NC}"
        exit 1
    fi
}

# Detect deployment mode
detect_deployment_mode() {
    section "🔍 Detecting Deployment Mode"

    # Check if Docker is being used
    if [ -f "$DOCKER_COMPOSE_FILE" ] && docker ps 2>/dev/null | grep -q "aureo"; then
        DEPLOYMENT_MODE="docker"
        detect_docker_compose
        echo -e "${GREEN}✓ Docker deployment detected${NC}"
        echo -e "${BLUE}  Docker Compose: $DOCKER_COMPOSE${NC}"
    # Check if systemd services exist
    elif systemctl list-unit-files 2>/dev/null | grep -q "aureo-"; then
        DEPLOYMENT_MODE="system"
        echo -e "${GREEN}✓ System deployment detected${NC}"
    else
        echo -e "${RED}✗ Could not detect deployment mode${NC}"
        echo -e "${YELLOW}No Docker containers or systemd services found${NC}"
        echo -e "${YELLOW}Please run become-node-operator.sh first${NC}"
        exit 1
    fi
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --all)
                RESTART_ALL=true
                RESTART_NODE=true
                RESTART_API=true
                RESTART_CONTROL=true
                shift
                ;;
            --node-only)
                RESTART_NODE=true
                RESTART_API=false
                RESTART_CONTROL=false
                shift
                ;;
            --api)
                RESTART_API=true
                shift
                ;;
            --control)
                RESTART_CONTROL=true
                shift
                ;;
            --force)
                FORCE=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                echo -e "${RED}Unknown option: $1${NC}"
                show_help
                exit 1
                ;;
        esac
    done
}

# Show help
show_help() {
    echo "Usage: sudo bash restart-node.sh [options]"
    echo ""
    echo "Options:"
    echo "  --all           Restart all services"
    echo "  --node-only     Restart only VPN node (default)"
    echo "  --api           Restart API gateway"
    echo "  --control       Restart control server"
    echo "  --force         Force restart (skip confirmations)"
    echo "  -h, --help      Show this help message"
    echo ""
}

# Confirm restart
confirm_restart() {
    if [ "$FORCE" = true ]; then
        return 0
    fi

    echo -e "${YELLOW}Services to be restarted:${NC}"
    [ "$RESTART_NODE" = true ] && echo "  • VPN Node"
    [ "$RESTART_API" = true ] && echo "  • API Gateway"
    [ "$RESTART_CONTROL" = true ] && echo "  • Control Server"
    echo ""

    read -p "Continue with restart? (y/n): " CONFIRM
    if [[ ! $CONFIRM =~ ^[Yy]$ ]]; then
        echo "Restart cancelled."
        exit 0
    fi
}

# Restart Docker services
restart_docker_services() {
    section "🐳 Restarting Docker Services"

    cd "$PROJECT_ROOT"

    # Get NODE_ID from .env
    if [ -f "$PROJECT_ROOT/deployments/docker/.env" ]; then
        NODE_ID=$(grep NODE_ID_1 "$PROJECT_ROOT/deployments/docker/.env" 2>/dev/null | cut -d'=' -f2)
        if [ -n "$NODE_ID" ]; then
            echo -e "${BLUE}Node ID: $NODE_ID${NC}"
        fi
    fi

    if [ "$RESTART_NODE" = true ]; then
        echo -e "${YELLOW}Restarting VPN node...${NC}"
        $DOCKER_COMPOSE -f "$DOCKER_COMPOSE_FILE" restart vpn-node-1
        echo -e "${GREEN}✓ VPN node restarted${NC}"
    fi

    if [ "$RESTART_API" = true ]; then
        echo -e "${YELLOW}Restarting API gateway...${NC}"
        $DOCKER_COMPOSE -f "$DOCKER_COMPOSE_FILE" restart api-gateway
        echo -e "${GREEN}✓ API gateway restarted${NC}"
    fi

    if [ "$RESTART_CONTROL" = true ]; then
        echo -e "${YELLOW}Restarting control server...${NC}"
        $DOCKER_COMPOSE -f "$DOCKER_COMPOSE_FILE" restart control-server
        echo -e "${GREEN}✓ Control server restarted${NC}"
    fi

    echo -e "${YELLOW}Waiting for services to be ready...${NC}"
    sleep 5
}

# Restart System services
restart_system_services() {
    section "⚙️  Restarting System Services"

    # Get NODE_ID from config
    if [ -f "$CONFIG_DIR/operator-credentials" ]; then
        source "$CONFIG_DIR/operator-credentials"
        if [ -n "$NODE_ID" ]; then
            echo -e "${BLUE}Node ID: $NODE_ID${NC}"
        fi
    fi

    if [ "$RESTART_NODE" = true ]; then
        echo -e "${YELLOW}Restarting VPN node...${NC}"
        systemctl restart aureo-vpn-node
        echo -e "${GREEN}✓ VPN node restarted${NC}"
    fi

    if [ "$RESTART_API" = true ]; then
        echo -e "${YELLOW}Restarting API gateway...${NC}"
        systemctl restart aureo-api-gateway
        echo -e "${GREEN}✓ API gateway restarted${NC}"
    fi

    if [ "$RESTART_CONTROL" = true ]; then
        echo -e "${YELLOW}Restarting control server...${NC}"
        systemctl restart aureo-control-server
        echo -e "${GREEN}✓ Control server restarted${NC}"
    fi

    # Reload Nginx if needed
    if [ "$RESTART_API" = true ]; then
        echo -e "${YELLOW}Reloading Nginx...${NC}"
        systemctl reload nginx 2>/dev/null || true
        echo -e "${GREEN}✓ Nginx reloaded${NC}"
    fi

    echo -e "${YELLOW}Waiting for services to be ready...${NC}"
    sleep 5
}

# Check Docker service health
check_docker_health() {
    section "🏥 Checking Service Health"

    local all_healthy=true

    if [ "$RESTART_NODE" = true ]; then
        echo -n "VPN Node: "
        if docker ps | grep -q "aureo-vpn-node-1"; then
            echo -e "${GREEN}✓ Running${NC}"

            # Get WireGuard status
            if [ -n "$NODE_ID" ]; then
                WG_PUBLIC_KEY=$(docker exec aureo-vpn-node-1 wg show wg0 public-key 2>/dev/null || echo "")
                if [ -n "$WG_PUBLIC_KEY" ]; then
                    echo -e "  ${CYAN}WireGuard Public Key: $WG_PUBLIC_KEY${NC}"

                    # Update database
                    docker exec aureo-vpn-db psql -U postgres -d aureo_vpn -c \
                        "UPDATE vpn_nodes SET public_key='$WG_PUBLIC_KEY', status='online', last_heartbeat=NOW() WHERE id='$NODE_ID';" \
                        >/dev/null 2>&1 || true
                fi
            fi
        else
            echo -e "${RED}✗ Not running${NC}"
            all_healthy=false
        fi
    fi

    if [ "$RESTART_API" = true ]; then
        echo -n "API Gateway: "
        if docker ps | grep -q "aureo.*api-gateway" && curl -sf http://localhost:8080/health >/dev/null 2>&1; then
            echo -e "${GREEN}✓ Running and healthy${NC}"
        else
            echo -e "${RED}✗ Not healthy${NC}"
            all_healthy=false
        fi
    fi

    if [ "$RESTART_CONTROL" = true ]; then
        echo -n "Control Server: "
        if docker ps | grep -q "aureo.*control-server"; then
            echo -e "${GREEN}✓ Running${NC}"
        else
            echo -e "${RED}✗ Not running${NC}"
            all_healthy=false
        fi
    fi

    if [ "$all_healthy" = false ]; then
        echo -e "\n${YELLOW}⚠ Some services are not healthy${NC}"
        echo -e "${CYAN}Check logs: $DOCKER_COMPOSE -f $DOCKER_COMPOSE_FILE logs -f${NC}"
    fi
}

# Check System service health
check_system_health() {
    section "🏥 Checking Service Health"

    local all_healthy=true

    if [ "$RESTART_NODE" = true ]; then
        echo -n "VPN Node: "
        if systemctl is-active --quiet aureo-vpn-node; then
            echo -e "${GREEN}✓ Running${NC}"

            # Get WireGuard status
            WG_PUBLIC_KEY=$(wg show wg0 public-key 2>/dev/null || echo "")
            if [ -n "$WG_PUBLIC_KEY" ]; then
                echo -e "  ${CYAN}WireGuard Public Key: $WG_PUBLIC_KEY${NC}"
            fi
        else
            echo -e "${RED}✗ Not running${NC}"
            all_healthy=false
        fi
    fi

    if [ "$RESTART_API" = true ]; then
        echo -n "API Gateway: "
        if systemctl is-active --quiet aureo-api-gateway && curl -sf http://localhost:8080/health >/dev/null 2>&1; then
            echo -e "${GREEN}✓ Running and healthy${NC}"
        else
            echo -e "${RED}✗ Not healthy${NC}"
            all_healthy=false
        fi
    fi

    if [ "$RESTART_CONTROL" = true ]; then
        echo -n "Control Server: "
        if systemctl is-active --quiet aureo-control-server; then
            echo -e "${GREEN}✓ Running${NC}"
        else
            echo -e "${RED}✗ Not running${NC}"
            all_healthy=false
        fi
    fi

    echo -n "Nginx: "
    if systemctl is-active --quiet nginx; then
        echo -e "${GREEN}✓ Running${NC}"
    else
        echo -e "${RED}✗ Not running${NC}"
        all_healthy=false
    fi

    if [ "$all_healthy" = false ]; then
        echo -e "\n${YELLOW}⚠ Some services are not healthy${NC}"
        echo -e "${CYAN}Check logs: journalctl -u aureo-* -f${NC}"
    fi
}

# Print summary
print_summary() {
    section "✅ Restart Complete"

    if [ -f "$CONFIG_DIR/operator-credentials" ]; then
        source "$CONFIG_DIR/operator-credentials"

        echo -e "${GREEN}Your VPN node has been restarted successfully!${NC}"
        echo ""
        echo -e "${CYAN}Node Information:${NC}"
        echo -e "  Node ID:    ${NODE_ID:-N/A}"
        echo -e "  Public IP:  ${PUBLIC_IP:-N/A}"
        echo ""
        echo -e "${CYAN}Useful Commands:${NC}"

        if [ "$DEPLOYMENT_MODE" = "docker" ]; then
            echo -e "  View logs:          ${GREEN}$DOCKER_COMPOSE -f $DOCKER_COMPOSE_FILE logs -f vpn-node-1${NC}"
            echo -e "  Check status:       ${GREEN}$DOCKER_COMPOSE -f $DOCKER_COMPOSE_FILE ps${NC}"
            echo -e "  View all logs:      ${GREEN}$DOCKER_COMPOSE -f $DOCKER_COMPOSE_FILE logs -f${NC}"
        else
            echo -e "  View node logs:     ${GREEN}journalctl -u aureo-vpn-node -f${NC}"
            echo -e "  View API logs:      ${GREEN}journalctl -u aureo-api-gateway -f${NC}"
            echo -e "  Check status:       ${GREEN}systemctl status aureo-*${NC}"
        fi
    else
        echo -e "${GREEN}Services restarted successfully!${NC}"
    fi
}

##############################################################################
# MAIN EXECUTION
##############################################################################

main() {
    print_header
    check_root

    # Parse arguments
    parse_args "$@"

    # Detect deployment mode
    detect_deployment_mode

    # Confirm restart
    confirm_restart

    # Restart services based on deployment mode
    if [ "$DEPLOYMENT_MODE" = "docker" ]; then
        restart_docker_services
        check_docker_health
    else
        restart_system_services
        check_system_health
    fi

    # Print summary
    print_summary

    echo -e "\n${GREEN}🔄 Node Restart Completed Successfully!${NC}\n"
}

# Run main
main "$@"
