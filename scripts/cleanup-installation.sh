#!/bin/bash

################################################################################
# Aureo VPN - Cleanup Installation Script
#
# Removes all Aureo VPN services and configurations
#
# Usage:
#   sudo bash cleanup-installation.sh [--force]
#
################################################################################

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONFIG_DIR="$HOME/.aureo-vpn"
DOCKER_COMPOSE_FILE="$PROJECT_ROOT/deployments/docker/docker-compose.yml"
FORCE=false

# Print header
print_header() {
    clear
    echo -e "${PURPLE}"
    cat << "EOF"
╔══════════════════════════════════════════════════════════════════╗
║                                                                  ║
║       🧹 Aureo VPN - Cleanup Script 🧹                          ║
║                                                                  ║
║     Remove all Aureo VPN services and configurations            ║
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

# Command exists helper
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Detect Docker Compose command
detect_docker_compose() {
    if docker compose version >/dev/null 2>&1; then
        DOCKER_COMPOSE="docker compose"
    elif command_exists docker-compose; then
        DOCKER_COMPOSE="docker-compose"
    else
        return 1
    fi
    return 0
}

# Detect what's installed
detect_installation() {
    section "🔍 Detecting Installation"

    local found_docker=false
    local found_system=false

    # Check for Docker installation
    if docker ps 2>/dev/null | grep -q "aureo"; then
        found_docker=true
        echo -e "${YELLOW}✓ Found Docker installation${NC}"
    fi

    # Check for systemd services
    if systemctl list-unit-files 2>/dev/null | grep -q "aureo-"; then
        found_system=true
        echo -e "${YELLOW}✓ Found System installation${NC}"
    fi

    # Check for config files
    if [ -d "$CONFIG_DIR" ] && [ -f "$CONFIG_DIR/operator-credentials" ]; then
        echo -e "${YELLOW}✓ Found configuration files${NC}"
    fi

    # Check for /opt/aureo-vpn
    if [ -d "/opt/aureo-vpn" ]; then
        echo -e "${YELLOW}✓ Found /opt/aureo-vpn directory${NC}"
    fi

    if [ "$found_docker" = false ] && [ "$found_system" = false ]; then
        echo -e "${GREEN}No Aureo VPN installation found${NC}"
        exit 0
    fi

    echo "$found_docker $found_system"
}

# Confirm cleanup
confirm_cleanup() {
    if [ "$FORCE" = true ]; then
        return 0
    fi

    echo ""
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${RED}  WARNING: THIS WILL REMOVE EVERYTHING!${NC}"
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "${YELLOW}This script will:${NC}"
    echo -e "  • Stop and remove all Docker containers"
    echo -e "  • Remove Docker volumes (including database data)"
    echo -e "  • Stop and disable all systemd services"
    echo -e "  • Remove systemd service files"
    echo -e "  • Remove Nginx configuration"
    echo -e "  • Remove configuration files and credentials"
    echo -e "  • Remove /opt/aureo-vpn directory"
    echo -e "  • Remove cron jobs"
    echo -e "  • Stop WireGuard interface"
    echo ""
    echo -e "${CYAN}Note: Your node registration in the database will remain.${NC}"
    echo -e "${CYAN}You can re-register with the same wallet address later.${NC}"
    echo ""

    read -p "Are you sure you want to continue? (yes/no): " CONFIRM
    if [[ ! $CONFIRM =~ ^[Yy][Ee][Ss]$ ]]; then
        echo -e "${YELLOW}Cleanup cancelled.${NC}"
        exit 0
    fi

    echo ""
    read -p "Type 'DELETE' to confirm: " CONFIRM2
    if [[ ! $CONFIRM2 == "DELETE" ]]; then
        echo -e "${YELLOW}Cleanup cancelled.${NC}"
        exit 0
    fi
}

# Cleanup Docker installation
cleanup_docker() {
    section "🐳 Cleaning Up Docker Installation"

    cd "$PROJECT_ROOT"

    # Detect docker compose command
    if detect_docker_compose; then
        echo -e "${YELLOW}Stopping and removing Docker containers...${NC}"
        $DOCKER_COMPOSE -f "$DOCKER_COMPOSE_FILE" down -v 2>/dev/null || true
        echo -e "${GREEN}✓ Docker containers removed${NC}"
    else
        echo -e "${YELLOW}Removing Docker containers manually...${NC}"
        docker rm -f $(docker ps -a -q --filter "name=aureo") 2>/dev/null || true
        echo -e "${GREEN}✓ Docker containers removed${NC}"
    fi

    # Remove Docker networks
    echo -e "${YELLOW}Removing Docker networks...${NC}"
    docker network rm docker_aureo-network 2>/dev/null || true
    echo -e "${GREEN}✓ Docker networks removed${NC}"

    # Clean up .env file
    echo -e "${YELLOW}Removing .env file...${NC}"
    rm -f "$PROJECT_ROOT/deployments/docker/.env" 2>/dev/null || true
    rm -f "$PROJECT_ROOT/deployments/docker/.env.backup"* 2>/dev/null || true
    echo -e "${GREEN}✓ .env files removed${NC}"
}

# Cleanup System installation
cleanup_system() {
    section "⚙️  Cleaning Up System Installation"

    # Stop all Aureo services
    echo -e "${YELLOW}Stopping services...${NC}"
    systemctl stop aureo-vpn-node 2>/dev/null || true
    systemctl stop aureo-api-gateway 2>/dev/null || true
    systemctl stop aureo-control-server 2>/dev/null || true
    echo -e "${GREEN}✓ Services stopped${NC}"

    # Disable services
    echo -e "${YELLOW}Disabling services...${NC}"
    systemctl disable aureo-vpn-node 2>/dev/null || true
    systemctl disable aureo-api-gateway 2>/dev/null || true
    systemctl disable aureo-control-server 2>/dev/null || true
    echo -e "${GREEN}✓ Services disabled${NC}"

    # Remove systemd service files
    echo -e "${YELLOW}Removing service files...${NC}"
    rm -f /etc/systemd/system/aureo-vpn-node.service 2>/dev/null || true
    rm -f /etc/systemd/system/aureo-api-gateway.service 2>/dev/null || true
    rm -f /etc/systemd/system/aureo-control-server.service 2>/dev/null || true
    echo -e "${GREEN}✓ Service files removed${NC}"

    # Reload systemd
    echo -e "${YELLOW}Reloading systemd...${NC}"
    systemctl daemon-reload
    echo -e "${GREEN}✓ Systemd reloaded${NC}"

    # Remove Nginx configuration
    echo -e "${YELLOW}Removing Nginx configuration...${NC}"
    rm -f /etc/nginx/sites-enabled/aureo-vpn 2>/dev/null || true
    rm -f /etc/nginx/sites-available/aureo-vpn 2>/dev/null || true

    # Restore default Nginx site if needed
    if [ ! -f /etc/nginx/sites-enabled/default ] && [ -f /etc/nginx/sites-available/default ]; then
        ln -s /etc/nginx/sites-available/default /etc/nginx/sites-enabled/default 2>/dev/null || true
    fi

    # Reload Nginx
    systemctl reload nginx 2>/dev/null || true
    echo -e "${GREEN}✓ Nginx configuration removed${NC}"

    # Stop WireGuard interface
    echo -e "${YELLOW}Stopping WireGuard interface...${NC}"
    wg-quick down wg0 2>/dev/null || true
    ip link delete wg0 2>/dev/null || true
    echo -e "${GREEN}✓ WireGuard stopped${NC}"

    # Clean up database
    echo -e "${YELLOW}Cleaning up database...${NC}"
    sudo -u postgres psql -c "DROP DATABASE IF EXISTS aureo_vpn;" 2>/dev/null || true
    sudo -u postgres psql -c "DROP USER IF EXISTS aureo;" 2>/dev/null || true
    echo -e "${GREEN}✓ Database cleaned up${NC}"
}

# Cleanup common files
cleanup_common() {
    section "📁 Cleaning Up Configuration Files"

    # Backup operator credentials if they exist
    if [ -f "$CONFIG_DIR/operator-credentials" ]; then
        echo -e "${YELLOW}Backing up operator credentials...${NC}"
        BACKUP_FILE="$CONFIG_DIR/operator-credentials.backup.$(date +%s)"
        cp "$CONFIG_DIR/operator-credentials" "$BACKUP_FILE" 2>/dev/null || true
        echo -e "${GREEN}✓ Credentials backed up to: $BACKUP_FILE${NC}"
    fi

    # Remove config directory
    echo -e "${YELLOW}Removing config directory...${NC}"
    rm -rf "$CONFIG_DIR" 2>/dev/null || true
    echo -e "${GREEN}✓ Config directory removed${NC}"

    # Remove /opt/aureo-vpn
    echo -e "${YELLOW}Removing /opt/aureo-vpn...${NC}"
    rm -rf /opt/aureo-vpn 2>/dev/null || true
    echo -e "${GREEN}✓ /opt/aureo-vpn removed${NC}"

    # Remove cron jobs
    echo -e "${YELLOW}Removing cron jobs...${NC}"
    (crontab -l 2>/dev/null | grep -v "aureo-vpn\|keep-node-online") | crontab - 2>/dev/null || true
    echo -e "${GREEN}✓ Cron jobs removed${NC}"

    # Remove binaries (optional)
    if [ -d "$PROJECT_ROOT/bin" ]; then
        echo -e "${YELLOW}Removing compiled binaries...${NC}"
        rm -rf "$PROJECT_ROOT/bin" 2>/dev/null || true
        echo -e "${GREEN}✓ Binaries removed${NC}"
    fi
}

# Print summary
print_summary() {
    section "✅ Cleanup Complete"

    echo -e "${GREEN}All Aureo VPN services and configurations have been removed.${NC}"
    echo ""
    echo -e "${CYAN}What was removed:${NC}"
    echo -e "  ✓ Docker containers and volumes"
    echo -e "  ✓ Systemd services"
    echo -e "  ✓ Nginx configuration"
    echo -e "  ✓ Configuration files"
    echo -e "  ✓ Cron jobs"
    echo -e "  ✓ WireGuard interface"
    echo ""
    echo -e "${CYAN}What was preserved:${NC}"
    echo -e "  ✓ Source code in $PROJECT_ROOT"
    echo -e "  ✓ Database data (if using system PostgreSQL)"
    echo -e "  ✓ Operator credentials backup (check $CONFIG_DIR/*.backup.*)"
    echo ""
    echo -e "${BLUE}To reinstall, run:${NC}"
    echo -e "  ${GREEN}sudo bash $PROJECT_ROOT/scripts/become-node-operator.sh${NC}"
    echo ""
}

# Parse arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --force)
                FORCE=true
                shift
                ;;
            -h|--help)
                echo "Usage: sudo bash cleanup-installation.sh [--force]"
                echo ""
                echo "Options:"
                echo "  --force    Skip confirmation prompts"
                echo "  -h, --help Show this help message"
                exit 0
                ;;
            *)
                echo -e "${RED}Unknown option: $1${NC}"
                exit 1
                ;;
        esac
    done
}

##############################################################################
# MAIN EXECUTION
##############################################################################

main() {
    print_header
    check_root

    # Parse arguments
    parse_args "$@"

    # Detect installation
    INSTALL_INFO=$(detect_installation)
    FOUND_DOCKER=$(echo $INSTALL_INFO | cut -d' ' -f1)
    FOUND_SYSTEM=$(echo $INSTALL_INFO | cut -d' ' -f2)

    # Confirm cleanup
    confirm_cleanup

    # Perform cleanup
    if [ "$FOUND_DOCKER" = "true" ]; then
        cleanup_docker
    fi

    if [ "$FOUND_SYSTEM" = "true" ]; then
        cleanup_system
    fi

    cleanup_common

    # Print summary
    print_summary

    echo -e "\n${GREEN}🧹 Cleanup Completed Successfully!${NC}\n"
}

# Run main
main "$@"
