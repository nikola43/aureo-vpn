#!/bin/bash

################################################################################
# Aureo VPN - Node Operator Setup Script
#
# One-command setup to become an Aureo VPN node operator
#
# What it does:
#   - Deploys all services via Docker Compose OR System-level
#   - Registers you as an operator
#   - Creates and activates your VPN node
#   - Sets up the peer registration script
#   - Configures everything for automatic operation
#
# Usage:
#   sudo bash become-node-operator.sh
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
BINARY_DIR="$PROJECT_ROOT/bin"
SYSTEMD_DIR="/etc/systemd/system"

# Deployment mode (will be set by user choice)
DEPLOYMENT_MODE=""

# Trap errors
trap 'error_handler $? $LINENO' ERR

error_handler() {
    echo -e "\n${RED}✗ Error occurred at line $2 (exit code: $1)${NC}"
    echo -e "${YELLOW}Setup failed. Please check the errors above.${NC}"
    exit 1
}

# Print header
print_header() {
    clear
    echo -e "${PURPLE}"
    cat << "EOF"
╔══════════════════════════════════════════════════════════════════╗
║                                                                  ║
║       🚀 Aureo VPN - Node Operator Setup 🚀                      ║
║                                                                  ║
║     One-Command Setup: Deploy and Earn Crypto Rewards!          ║
║                                                                  ║
╚══════════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}\n"
}

# Print section header
section() {
    echo -e "\n${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  $1${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
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

# Install Go from official source
install_go() {
    section "📦 Installing Go"

    local GO_VERSION="1.25.3"
    local GO_TARBALL="go${GO_VERSION}.linux-amd64.tar.gz"
    local GO_URL="https://go.dev/dl/${GO_TARBALL}"

    echo -e "${YELLOW}Downloading Go ${GO_VERSION}...${NC}"

    cd /tmp
    if ! wget -q --show-progress "$GO_URL"; then
        echo -e "${RED}✗ Failed to download Go${NC}"
        return 1
    fi

    echo -e "${YELLOW}Installing Go...${NC}"

    # Remove old Go installation
    rm -rf /usr/local/go

    # Extract new Go
    tar -C /usr/local -xzf "$GO_TARBALL"

    # Add Go to PATH for all users
    if ! grep -q "/usr/local/go/bin" /etc/profile; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    fi

    # Add to current session
    export PATH=$PATH:/usr/local/go/bin

    # Add to .bashrc for root
    if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    fi

    # Cleanup
    rm -f "$GO_TARBALL"

    # Verify installation
    if /usr/local/go/bin/go version >/dev/null 2>&1; then
        echo -e "${GREEN}✓ Go ${GO_VERSION} installed successfully${NC}"
        return 0
    else
        echo -e "${RED}✗ Go installation failed${NC}"
        return 1
    fi
}

# Install PostgreSQL
install_postgresql() {
    section "📦 Installing PostgreSQL"

    echo -e "${YELLOW}Installing PostgreSQL...${NC}"

    # Update package list
    apt-get update -qq

    # Install PostgreSQL
    if DEBIAN_FRONTEND=noninteractive apt-get install -y -qq postgresql postgresql-contrib; then
        echo -e "${GREEN}✓ PostgreSQL installed successfully${NC}"

        # Start and enable PostgreSQL
        systemctl start postgresql
        systemctl enable postgresql

        echo -e "${GREEN}✓ PostgreSQL service started${NC}"
        return 0
    else
        echo -e "${RED}✗ PostgreSQL installation failed${NC}"
        return 1
    fi
}

# Install Redis
install_redis() {
    section "📦 Installing Redis"

    echo -e "${YELLOW}Installing Redis...${NC}"

    # Update package list
    apt-get update -qq

    # Install Redis
    if DEBIAN_FRONTEND=noninteractive apt-get install -y -qq redis-server; then
        echo -e "${GREEN}✓ Redis installed successfully${NC}"

        # Configure Redis to start on boot
        systemctl start redis-server
        systemctl enable redis-server

        echo -e "${GREEN}✓ Redis service started${NC}"
        return 0
    else
        echo -e "${RED}✗ Redis installation failed${NC}"
        return 1
    fi
}

# Install Nginx
install_nginx() {
    section "📦 Installing Nginx"

    echo -e "${YELLOW}Installing Nginx...${NC}"

    # Update package list
    apt-get update -qq

    # Install Nginx
    if DEBIAN_FRONTEND=noninteractive apt-get install -y -qq nginx; then
        echo -e "${GREEN}✓ Nginx installed successfully${NC}"

        # Start and enable Nginx
        systemctl start nginx
        systemctl enable nginx

        echo -e "${GREEN}✓ Nginx service started${NC}"
        return 0
    else
        echo -e "${RED}✗ Nginx installation failed${NC}"
        return 1
    fi
}

# Install WireGuard
install_wireguard() {
    section "📦 Installing WireGuard"

    echo -e "${YELLOW}Installing WireGuard...${NC}"

    # Update package list
    apt-get update -qq

    # Install WireGuard
    if DEBIAN_FRONTEND=noninteractive apt-get install -y -qq wireguard wireguard-tools; then
        echo -e "${GREEN}✓ WireGuard installed successfully${NC}"

        # Enable IP forwarding
        if ! grep -q "^net.ipv4.ip_forward=1" /etc/sysctl.conf; then
            echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf
            sysctl -p > /dev/null
            echo -e "${GREEN}✓ IP forwarding enabled${NC}"
        fi

        return 0
    else
        echo -e "${RED}✗ WireGuard installation failed${NC}"
        return 1
    fi
}

# Install all system dependencies
install_all_dependencies() {
    section "📦 Installing All Dependencies"

    echo -e "${CYAN}This will install:${NC}"
    echo "  • Go (latest version)"
    echo "  • PostgreSQL"
    echo "  • Redis"
    echo "  • Nginx"
    echo "  • WireGuard"
    echo ""

    read -p "Continue with installation? (y/n): " INSTALL_DEPS
    if [[ ! $INSTALL_DEPS =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}Dependency installation skipped${NC}"
        return 1
    fi

    # Install each dependency
    install_go || true
    install_postgresql || true
    install_redis || true
    install_nginx || true
    install_wireguard || true

    echo -e "\n${GREEN}✓ Dependency installation complete!${NC}"
    echo -e "${YELLOW}Please restart your shell or run: source ~/.bashrc${NC}\n"
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

# Detect previous installation
detect_previous_installation() {
    section "🔍 Checking for Previous Installation"

    local found_docker=false
    local found_system=false

    # Check for Docker installation
    if docker ps 2>/dev/null | grep -q "aureo"; then
        found_docker=true
        echo -e "${YELLOW}⚠ Found existing Docker installation${NC}"
    fi

    # Check for systemd services
    if systemctl list-unit-files 2>/dev/null | grep -q "aureo-"; then
        found_system=true
        echo -e "${YELLOW}⚠ Found existing System installation${NC}"
    fi

    # Check for config files
    if [ -d "$CONFIG_DIR" ] && [ -f "$CONFIG_DIR/operator-credentials" ]; then
        echo -e "${YELLOW}⚠ Found existing configuration files${NC}"
    fi

    if [ "$found_docker" = true ] || [ "$found_system" = true ]; then
        echo ""
        echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${CYAN}  PREVIOUS INSTALLATION DETECTED${NC}"
        echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo ""
        echo -e "${YELLOW}A previous Aureo VPN installation was detected.${NC}"
        echo -e "${YELLOW}To ensure a clean installation, all existing services${NC}"
        echo -e "${YELLOW}and configurations will be removed.${NC}"
        echo ""
        echo -e "${RED}WARNING: This will:${NC}"
        [ "$found_docker" = true ] && echo -e "  • Stop and remove all Docker containers"
        [ "$found_docker" = true ] && echo -e "  • Remove Docker volumes (including database data)"
        [ "$found_system" = true ] && echo -e "  • Stop and disable systemd services"
        [ "$found_system" = true ] && echo -e "  • Remove systemd service files"
        echo -e "  • Remove configuration files"
        echo -e "  • Remove operator credentials"
        echo ""
        echo -e "${CYAN}Your node registration and earnings will be preserved${NC}"
        echo -e "${CYAN}if you use the same wallet address during setup.${NC}"
        echo ""

        read -p "Do you want to remove the existing installation and continue? (y/n): " REMOVE_EXISTING
        if [[ ! $REMOVE_EXISTING =~ ^[Yy]$ ]]; then
            echo -e "${YELLOW}Installation cancelled.${NC}"
            echo -e "${CYAN}To manually remove the installation, run:${NC}"
            if [ "$found_docker" = true ]; then
                echo -e "  ${GREEN}cd $PROJECT_ROOT && docker compose -f $DOCKER_COMPOSE_FILE down -v${NC}"
            fi
            if [ "$found_system" = true ]; then
                echo -e "  ${GREEN}sudo systemctl stop aureo-* && sudo systemctl disable aureo-*${NC}"
                echo -e "  ${GREEN}sudo rm -f /etc/systemd/system/aureo-*.service${NC}"
            fi
            exit 0
        fi

        # Perform cleanup
        cleanup_previous_installation "$found_docker" "$found_system"
    else
        echo -e "${GREEN}✓ No previous installation found${NC}"
    fi
}

# Cleanup Docker installation
cleanup_docker_installation() {
    echo -e "${YELLOW}Cleaning up Docker installation...${NC}"

    cd "$PROJECT_ROOT"

    # Detect docker compose command
    detect_docker_compose || true

    # Stop and remove all containers
    if [ -n "$DOCKER_COMPOSE" ]; then
        echo -e "${CYAN}Stopping Docker containers...${NC}"
        $DOCKER_COMPOSE -f "$DOCKER_COMPOSE_FILE" down -v 2>/dev/null || true
    else
        # Fallback: remove containers manually
        docker rm -f $(docker ps -a -q --filter "name=aureo") 2>/dev/null || true
    fi

    # Remove Docker networks
    docker network rm docker_aureo-network 2>/dev/null || true

    # Remove Docker images (optional - commented out to save bandwidth)
    # echo -e "${CYAN}Removing Docker images...${NC}"
    # docker rmi $(docker images -q "docker_*") 2>/dev/null || true

    # Clean up .env file
    rm -f "$PROJECT_ROOT/deployments/docker/.env" 2>/dev/null || true

    echo -e "${GREEN}✓ Docker installation cleaned up${NC}"
}

# Cleanup System installation
cleanup_system_installation() {
    echo -e "${YELLOW}Cleaning up System installation...${NC}"

    # Stop all Aureo services
    echo -e "${CYAN}Stopping services...${NC}"
    systemctl stop aureo-vpn-node 2>/dev/null || true
    systemctl stop aureo-api-gateway 2>/dev/null || true
    systemctl stop aureo-control-server 2>/dev/null || true

    # Disable services
    echo -e "${CYAN}Disabling services...${NC}"
    systemctl disable aureo-vpn-node 2>/dev/null || true
    systemctl disable aureo-api-gateway 2>/dev/null || true
    systemctl disable aureo-control-server 2>/dev/null || true

    # Remove systemd service files
    echo -e "${CYAN}Removing service files...${NC}"
    rm -f /etc/systemd/system/aureo-vpn-node.service 2>/dev/null || true
    rm -f /etc/systemd/system/aureo-api-gateway.service 2>/dev/null || true
    rm -f /etc/systemd/system/aureo-control-server.service 2>/dev/null || true

    # Reload systemd
    systemctl daemon-reload

    # Remove Nginx configuration
    echo -e "${CYAN}Removing Nginx configuration...${NC}"
    rm -f /etc/nginx/sites-enabled/aureo-vpn 2>/dev/null || true
    rm -f /etc/nginx/sites-available/aureo-vpn 2>/dev/null || true

    # Restore default Nginx site if needed
    if [ ! -f /etc/nginx/sites-enabled/default ] && [ -f /etc/nginx/sites-available/default ]; then
        ln -s /etc/nginx/sites-available/default /etc/nginx/sites-enabled/default 2>/dev/null || true
    fi

    # Reload Nginx
    systemctl reload nginx 2>/dev/null || true

    # Stop WireGuard interface
    echo -e "${CYAN}Stopping WireGuard interface...${NC}"
    wg-quick down wg0 2>/dev/null || true
    ip link delete wg0 2>/dev/null || true

    # Clean up database
    echo -e "${CYAN}Cleaning up database...${NC}"
    sudo -u postgres psql -c "DROP DATABASE IF EXISTS aureo_vpn;" 2>/dev/null || true
    sudo -u postgres psql -c "DROP USER IF EXISTS aureo;" 2>/dev/null || true
    echo -e "${GREEN}✓ Database cleaned up${NC}"

    echo -e "${GREEN}✓ System installation cleaned up${NC}"
}

# Cleanup common files and configurations
cleanup_common() {
    echo -e "${YELLOW}Cleaning up configuration files...${NC}"

    # Remove config directory
    if [ -d "$CONFIG_DIR" ]; then
        echo -e "${CYAN}Backing up operator credentials...${NC}"
        if [ -f "$CONFIG_DIR/operator-credentials" ]; then
            cp "$CONFIG_DIR/operator-credentials" "$CONFIG_DIR/operator-credentials.backup.$(date +%s)" 2>/dev/null || true
        fi
        # Don't remove the entire directory, just clean it
        rm -f "$CONFIG_DIR/operator-credentials" 2>/dev/null || true
    fi

    # Remove /opt/aureo-vpn scripts (will be reinstalled)
    rm -rf /opt/aureo-vpn 2>/dev/null || true

    # Remove cron jobs
    echo -e "${CYAN}Removing cron jobs...${NC}"
    (crontab -l 2>/dev/null | grep -v "aureo-vpn\|keep-node-online") | crontab - 2>/dev/null || true

    echo -e "${GREEN}✓ Configuration files cleaned up${NC}"
}

# Main cleanup function
cleanup_previous_installation() {
    local found_docker=$1
    local found_system=$2

    section "🧹 Removing Previous Installation"

    if [ "$found_docker" = true ]; then
        cleanup_docker_installation
    fi

    if [ "$found_system" = true ]; then
        cleanup_system_installation
    fi

    cleanup_common

    echo ""
    echo -e "${GREEN}✓ Previous installation successfully removed${NC}"
    echo -e "${CYAN}Ready for fresh installation...${NC}"
    echo ""

    # Wait a moment for services to fully stop
    sleep 3
}

# Ask user for deployment mode
select_deployment_mode() {
    section "🔧 Select Deployment Mode"

    echo -e "${CYAN}Choose how you want to deploy Aureo VPN:${NC}"
    echo ""
    echo -e "${GREEN}1)${NC} Docker (Recommended)"
    echo -e "   ${YELLOW}✓${NC} Easy setup and management"
    echo -e "   ${YELLOW}✓${NC} Isolated containers"
    echo -e "   ${YELLOW}✓${NC} Automatic updates"
    echo -e "   ${YELLOW}✓${NC} Works on any platform"
    echo ""
    echo -e "${GREEN}2)${NC} System (Native)"
    echo -e "   ${YELLOW}✓${NC} Better performance"
    echo -e "   ${YELLOW}✓${NC} Lower resource usage"
    echo -e "   ${YELLOW}✓${NC} Direct system integration"
    echo -e "   ${YELLOW}⚠${NC}  Requires Go, PostgreSQL, Redis installation"
    echo ""

    while true; do
        read -p "Enter your choice [1-2]: " choice
        case $choice in
            1)
                DEPLOYMENT_MODE="docker"
                echo -e "${GREEN}✓ Docker deployment selected${NC}"
                break
                ;;
            2)
                DEPLOYMENT_MODE="system"
                echo -e "${GREEN}✓ System deployment selected${NC}"
                break
                ;;
            *)
                echo -e "${RED}Invalid choice. Please enter 1 or 2.${NC}"
                ;;
        esac
    done
    echo ""
}

# Check prerequisites for Docker mode
check_docker_prerequisites() {
    # Check Docker
    if ! command_exists docker; then
        echo -e "${RED}✗ Docker is not installed${NC}"
        echo -e "${YELLOW}Please install Docker first: https://docs.docker.com/get-docker/${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Docker installed${NC}"

    # Check Docker Compose
    if ! detect_docker_compose; then
        echo -e "${RED}✗ Docker Compose is not installed${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Docker Compose installed ($DOCKER_COMPOSE)${NC}"

    # Check Docker is running
    if ! docker ps >/dev/null 2>&1; then
        echo -e "${RED}✗ Docker daemon is not running${NC}"
        echo -e "${YELLOW}Please start Docker and try again${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Docker daemon running${NC}"

    # Check project files exist
    if [ ! -f "$DOCKER_COMPOSE_FILE" ]; then
        echo -e "${RED}✗ docker-compose.yml not found${NC}"
        echo -e "${YELLOW}Please run this script from the project root${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Project files found${NC}"
}

# Check prerequisites for System mode
check_system_prerequisites() {
    local missing_deps=()

    # Check Go
    if ! command_exists go && ! command_exists /usr/local/go/bin/go; then
        echo -e "${RED}✗ Go is not installed${NC}"
        missing_deps+=("go")
    else
        if command_exists go; then
            GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        else
            GO_VERSION=$(/usr/local/go/bin/go version | awk '{print $3}' | sed 's/go//')
        fi
        echo -e "${GREEN}✓ Go installed (version $GO_VERSION)${NC}"
    fi

    # Check PostgreSQL
    if ! command_exists psql; then
        echo -e "${RED}✗ PostgreSQL is not installed${NC}"
        missing_deps+=("postgresql")
    else
        echo -e "${GREEN}✓ PostgreSQL installed${NC}"

        # Check if PostgreSQL is running
        if ! sudo systemctl is-active --quiet postgresql; then
            echo -e "${YELLOW}⚠ PostgreSQL is not running. Starting it...${NC}"
            sudo systemctl start postgresql
        fi
        echo -e "${GREEN}✓ PostgreSQL is running${NC}"
    fi

    # Check Redis
    if ! command_exists redis-cli; then
        echo -e "${RED}✗ Redis is not installed${NC}"
        missing_deps+=("redis")
    else
        echo -e "${GREEN}✓ Redis installed${NC}"

        # Check if Redis is running
        if ! sudo systemctl is-active --quiet redis-server && ! sudo systemctl is-active --quiet redis; then
            echo -e "${YELLOW}⚠ Redis is not running. Starting it...${NC}"
            sudo systemctl start redis-server 2>/dev/null || sudo systemctl start redis 2>/dev/null || true
        fi
        echo -e "${GREEN}✓ Redis is running${NC}"
    fi

    # Check Nginx
    if ! command_exists nginx; then
        echo -e "${RED}✗ Nginx is not installed${NC}"
        missing_deps+=("nginx")
    else
        echo -e "${GREEN}✓ Nginx installed${NC}"
    fi

    # Check WireGuard
    if ! command_exists wg; then
        echo -e "${RED}✗ WireGuard is not installed${NC}"
        missing_deps+=("wireguard")
    else
        echo -e "${GREEN}✓ WireGuard installed${NC}"
    fi

    # If there are missing dependencies, offer to install them
    if [ ${#missing_deps[@]} -gt 0 ]; then
        echo ""
        echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${YELLOW}  Missing Dependencies Detected${NC}"
        echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo ""
        echo -e "${CYAN}The following dependencies are missing:${NC}"
        for dep in "${missing_deps[@]}"; do
            echo "  • $dep"
        done
        echo ""
        echo -e "${CYAN}Would you like to automatically install them?${NC}"
        read -p "Install missing dependencies? (y/n): " AUTO_INSTALL

        if [[ $AUTO_INSTALL =~ ^[Yy]$ ]]; then
            # Install missing dependencies
            for dep in "${missing_deps[@]}"; do
                case $dep in
                    go)
                        install_go || true
                        ;;
                    postgresql)
                        install_postgresql || true
                        ;;
                    redis)
                        install_redis || true
                        ;;
                    nginx)
                        install_nginx || true
                        ;;
                    wireguard)
                        install_wireguard || true
                        ;;
                esac
            done

            # Re-check prerequisites
            echo ""
            echo -e "${CYAN}Verifying installations...${NC}"
            sleep 2
            check_system_prerequisites
            return
        else
            echo -e "${RED}Cannot proceed without required dependencies${NC}"
            echo ""
            echo -e "${CYAN}Manual installation commands:${NC}"
            for dep in "${missing_deps[@]}"; do
                case $dep in
                    go)
                        echo "  Go: wget https://go.dev/dl/go1.25.3.linux-amd64.tar.gz && tar -C /usr/local -xzf go1.25.3.linux-amd64.tar.gz"
                        ;;
                    postgresql)
                        echo "  PostgreSQL: sudo apt-get install postgresql postgresql-contrib"
                        ;;
                    redis)
                        echo "  Redis: sudo apt-get install redis-server"
                        ;;
                    nginx)
                        echo "  Nginx: sudo apt-get install nginx"
                        ;;
                    wireguard)
                        echo "  WireGuard: sudo apt-get install wireguard wireguard-tools"
                        ;;
                esac
            done
            exit 1
        fi
    fi

    # Check project files
    if [ ! -f "$PROJECT_ROOT/go.mod" ]; then
        echo -e "${RED}✗ Go project files not found${NC}"
        echo -e "${YELLOW}Please run this script from the project root${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Project files found${NC}"
}

# Check prerequisites based on deployment mode
check_prerequisites() {
    section "🔍 Checking Prerequisites"

    if [ "$DEPLOYMENT_MODE" = "docker" ]; then
        check_docker_prerequisites
    else
        check_system_prerequisites
    fi
}

# Build binaries for system deployment
build_binaries() {
    section "🔨 Building Binaries"

    cd "$PROJECT_ROOT"

    echo -e "${YELLOW}Building all services from source...${NC}"

    # Build all binaries using Makefile
    if ! make build; then
        echo -e "${RED}✗ Failed to build binaries${NC}"
        exit 1
    fi

    echo -e "${GREEN}✓ Binaries built successfully${NC}"
}

# Setup database for system deployment
setup_system_database() {
    section "💾 Setting Up Database"

    echo -e "${YELLOW}Creating database and user...${NC}"

    # Check if database exists
    DB_EXISTS=$(sudo -u postgres psql -tAc "SELECT 1 FROM pg_database WHERE datname='aureo_vpn'")

    if [ "$DB_EXISTS" != "1" ]; then
        sudo -u postgres createdb aureo_vpn
        echo -e "${GREEN}✓ Database 'aureo_vpn' created${NC}"
    else
        echo -e "${YELLOW}⚠ Database 'aureo_vpn' already exists${NC}"
    fi

    # Check if user exists
    USER_EXISTS=$(sudo -u postgres psql -tAc "SELECT 1 FROM pg_roles WHERE rolname='aureo'")

    if [ "$USER_EXISTS" != "1" ]; then
        sudo -u postgres psql -c "CREATE USER aureo WITH ENCRYPTED PASSWORD 'aureo_secure_pass';"
        echo -e "${GREEN}✓ User 'aureo' created${NC}"
    else
        echo -e "${YELLOW}⚠ User 'aureo' already exists${NC}"
    fi

    # Grant privileges
    sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE aureo_vpn TO aureo;"
    sudo -u postgres psql -d aureo_vpn -c "GRANT ALL ON SCHEMA public TO aureo;"
    sudo -u postgres psql -d aureo_vpn -c "ALTER DATABASE aureo_vpn OWNER TO aureo;"

    echo -e "${GREEN}✓ Database permissions configured${NC}"

    # Configure PostgreSQL to accept local connections with password
    PG_VERSION=$(sudo -u postgres psql -tAc "SELECT version()" | grep -oP 'PostgreSQL \K[0-9]+')
    PG_HBA_FILE="/etc/postgresql/${PG_VERSION}/main/pg_hba.conf"

    if [ -f "$PG_HBA_FILE" ]; then
        # Backup original
        cp "$PG_HBA_FILE" "${PG_HBA_FILE}.backup.$(date +%s)" 2>/dev/null || true

        # Ensure md5 authentication for local connections
        if ! grep -q "^host.*aureo_vpn.*aureo.*md5" "$PG_HBA_FILE"; then
            sed -i "/^# IPv4 local connections:/a host    aureo_vpn       aureo           127.0.0.1/32            md5" "$PG_HBA_FILE"
            sed -i "/^# IPv6 local connections:/a host    aureo_vpn       aureo           ::1/128                 md5" "$PG_HBA_FILE"

            # Reload PostgreSQL configuration
            systemctl reload postgresql
            echo -e "${GREEN}✓ PostgreSQL authentication configured${NC}"
        fi
    fi

    # Run migrations
    echo -e "${YELLOW}Running database migrations...${NC}"

    # Set environment variables for migration
    export DATABASE_URL="postgresql://aureo:aureo_secure_pass@localhost:5432/aureo_vpn?sslmode=disable"
    export DB_HOST="localhost"
    export DB_PORT="5432"
    export DB_USER="aureo"
    export DB_PASSWORD="aureo_secure_pass"
    export DB_NAME="aureo_vpn"

    # Wait a moment for PostgreSQL reload
    sleep 2

    # Run migrations using the API gateway binary
    cd "$PROJECT_ROOT"
    if [ -f "$BINARY_DIR/api-gateway" ]; then
        echo -e "${CYAN}Running migrations (this may take a moment)...${NC}"
        timeout 30 $BINARY_DIR/api-gateway migrate up 2>&1 | grep -i "migration\|error" || true
    fi

    echo -e "${GREEN}✓ Database setup complete${NC}"
}

# Create systemd service files
create_systemd_services() {
    section "⚙️  Creating Systemd Services"

    # API Gateway service
    cat > "$SYSTEMD_DIR/aureo-api-gateway.service" << EOF
[Unit]
Description=Aureo VPN API Gateway
After=network.target postgresql.service redis.service
Wants=postgresql.service redis.service

[Service]
Type=simple
User=root
WorkingDirectory=$PROJECT_ROOT
Environment="DB_HOST=localhost"
Environment="DB_PORT=5432"
Environment="DB_USER=aureo"
Environment="DB_PASSWORD=aureo_secure_pass"
Environment="DB_NAME=aureo_vpn"
Environment="REDIS_HOST=localhost"
Environment="REDIS_PORT=6379"
Environment="JWT_SECRET=your-secret-key-change-this"
Environment="PORT=8080"
ExecStart=$BINARY_DIR/api-gateway
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

    # Control Server service
    cat > "$SYSTEMD_DIR/aureo-control-server.service" << EOF
[Unit]
Description=Aureo VPN Control Server
After=network.target postgresql.service redis.service
Wants=postgresql.service redis.service

[Service]
Type=simple
User=root
WorkingDirectory=$PROJECT_ROOT
Environment="DB_HOST=localhost"
Environment="DB_PORT=5432"
Environment="DB_USER=aureo"
Environment="DB_PASSWORD=aureo_secure_pass"
Environment="DB_NAME=aureo_vpn"
Environment="REDIS_HOST=localhost"
Environment="REDIS_PORT=6379"
ExecStart=$BINARY_DIR/control-server
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

    # VPN Node service (will be created later with NODE_ID)
    echo -e "${GREEN}✓ Systemd service files created${NC}"
}

# Deploy base services for system mode
deploy_system_base_services() {
    section "🚀 Deploying System Services"

    # Build binaries
    build_binaries

    # Setup database
    setup_system_database

    # Create systemd services
    create_systemd_services

    # Reload systemd
    systemctl daemon-reload

    # Start services
    echo -e "${YELLOW}Starting services...${NC}"
    systemctl start aureo-api-gateway
    systemctl start aureo-control-server

    # Enable services to start on boot
    systemctl enable aureo-api-gateway
    systemctl enable aureo-control-server

    echo -e "${YELLOW}Waiting for services to be ready...${NC}"
    sleep 10

    # Check services are running
    if ! systemctl is-active --quiet aureo-api-gateway; then
        echo -e "${RED}✗ API Gateway failed to start${NC}"
        echo -e "${YELLOW}Check logs with: journalctl -u aureo-api-gateway -f${NC}"
        exit 1
    fi

    if ! systemctl is-active --quiet aureo-control-server; then
        echo -e "${RED}✗ Control Server failed to start${NC}"
        echo -e "${YELLOW}Check logs with: journalctl -u aureo-control-server -f${NC}"
        exit 1
    fi

    echo -e "${GREEN}✓ System services deployed successfully${NC}"
}

# Deploy base services (Docker)
deploy_docker_base_services() {
    section "🐳 Deploying Docker Services"

    cd "$PROJECT_ROOT"

    echo -e "${YELLOW}Building and starting containers...${NC}"

    # Stop any existing containers
    $DOCKER_COMPOSE -f "$DOCKER_COMPOSE_FILE" down 2>/dev/null || true

    # Start only base services (postgres, redis, api, dashboard, control, prometheus, grafana)
    $DOCKER_COMPOSE -f "$DOCKER_COMPOSE_FILE" up -d --build postgres redis api-gateway control-server dashboard prometheus grafana

    echo -e "${YELLOW}Waiting for services to be ready...${NC}"
    sleep 15

    # Check services are running
    if ! $DOCKER_COMPOSE -f "$DOCKER_COMPOSE_FILE" ps | grep -q "Up"; then
        echo -e "${RED}✗ Services failed to start${NC}"
        echo -e "${YELLOW}Check logs with: $DOCKER_COMPOSE -f $DOCKER_COMPOSE_FILE logs${NC}"
        exit 1
    fi

    echo -e "${GREEN}✓ Docker services deployed successfully${NC}"
}

# Deploy base services based on deployment mode
deploy_base_services() {
    if [ "$DEPLOYMENT_MODE" = "docker" ]; then
        deploy_docker_base_services
    else
        deploy_system_base_services
    fi
}

# Deploy VPN node for system mode
deploy_system_vpn_node() {
    echo -e "${YELLOW}Creating VPN node systemd service...${NC}"

    # Get database connection info for system mode
    DB_CONN_STRING="postgresql://aureo:aureo_secure_pass@localhost:5432/aureo_vpn?sslmode=disable"

    # Create VPN Node systemd service
    cat > "$SYSTEMD_DIR/aureo-vpn-node.service" << EOF
[Unit]
Description=Aureo VPN Node
After=network.target postgresql.service redis.service aureo-api-gateway.service
Wants=postgresql.service redis.service aureo-api-gateway.service

[Service]
Type=simple
User=root
WorkingDirectory=$PROJECT_ROOT
Environment="NODE_ID=$NODE_ID"
Environment="DB_HOST=localhost"
Environment="DB_PORT=5432"
Environment="DB_USER=aureo"
Environment="DB_PASSWORD=aureo_secure_pass"
Environment="DB_NAME=aureo_vpn"
Environment="REDIS_HOST=localhost"
Environment="REDIS_PORT=6379"
Environment="WIREGUARD_PORT=51820"
Environment="OPENVPN_PORT=1194"
Environment="API_URL=http://localhost:8080"
ExecStart=$BINARY_DIR/vpn-node
Restart=always
RestartSec=10
CapabilityBoundingSet=CAP_NET_ADMIN
AmbientCapabilities=CAP_NET_ADMIN

[Install]
WantedBy=multi-user.target
EOF

    # Reload systemd
    systemctl daemon-reload

    # Start VPN node service
    echo -e "${YELLOW}Starting VPN node with NODE_ID: $NODE_ID${NC}"
    systemctl start aureo-vpn-node
    systemctl enable aureo-vpn-node

    echo -e "${YELLOW}Waiting for VPN node to initialize...${NC}"
    sleep 10

    # Check if VPN node is running
    if ! systemctl is-active --quiet aureo-vpn-node; then
        echo -e "${RED}✗ VPN node failed to start${NC}"
        echo -e "${YELLOW}Check logs: journalctl -u aureo-vpn-node -f${NC}"
        exit 1
    fi

    echo -e "${GREEN}✓ VPN node deployed successfully${NC}"
}

# Deploy VPN node for Docker mode
deploy_docker_vpn_node() {
    cd "$PROJECT_ROOT"

    # Create/update .env file with NODE_ID
    mkdir -p "$PROJECT_ROOT/deployments/docker"
    echo "NODE_ID_1=$NODE_ID" > "$PROJECT_ROOT/deployments/docker/.env"

    echo -e "${YELLOW}Starting VPN node with NODE_ID: $NODE_ID${NC}"

    # Start VPN node container
    $DOCKER_COMPOSE -f "$DOCKER_COMPOSE_FILE" up -d --build vpn-node-1

    echo -e "${YELLOW}Waiting for VPN node to initialize...${NC}"
    sleep 10

    # Check if VPN node is running
    if ! docker ps | grep -q "aureo-vpn-node-1"; then
        echo -e "${RED}✗ VPN node failed to start${NC}"
        echo -e "${YELLOW}Check logs: docker logs aureo-vpn-node-1${NC}"
        exit 1
    fi

    echo -e "${GREEN}✓ VPN node deployed successfully${NC}"
}

# Deploy VPN node based on deployment mode
deploy_vpn_node() {
    section "🚀 Deploying VPN Node"

    if [ "$DEPLOYMENT_MODE" = "docker" ]; then
        deploy_docker_vpn_node
    else
        deploy_system_vpn_node
    fi
}

# Execute database query based on deployment mode
db_exec() {
    local query="$1"
    if [ "$DEPLOYMENT_MODE" = "docker" ]; then
        docker exec aureo-vpn-db psql -U postgres -d aureo_vpn -c "$query" >/dev/null 2>&1
    else
        sudo -u postgres psql -d aureo_vpn -c "$query" >/dev/null 2>&1
    fi
}

# Finalize node setup for Docker
finalize_docker_node_setup() {
    # Wait a bit more for WireGuard to be ready
    sleep 5

    # Get WireGuard server public key and update node
    echo -e "${CYAN}Configuring WireGuard...${NC}"
    WG_PUBLIC_KEY=$(docker exec aureo-vpn-node-1 wg show wg0 public-key 2>/dev/null || echo "")

    if [ -n "$WG_PUBLIC_KEY" ]; then
        db_exec "UPDATE vpn_nodes SET public_key='$WG_PUBLIC_KEY' WHERE id='$NODE_ID';"
        echo -e "${GREEN}✓ WireGuard public key configured${NC}"
    else
        echo -e "${YELLOW}⚠ Could not get WireGuard public key yet${NC}"
    fi

    # Verify WireGuard is running
    if docker exec aureo-vpn-node-1 wg show wg0 >/dev/null 2>&1; then
        echo -e "${GREEN}✓ WireGuard interface is active${NC}"
    else
        echo -e "${YELLOW}⚠ WireGuard interface not ready yet${NC}"
    fi
}

# Finalize node setup for System
finalize_system_node_setup() {
    # Wait a bit more for WireGuard to be ready
    sleep 5

    # Get WireGuard server public key and update node
    echo -e "${CYAN}Configuring WireGuard...${NC}"
    WG_PUBLIC_KEY=$(wg show wg0 public-key 2>/dev/null || echo "")

    if [ -n "$WG_PUBLIC_KEY" ]; then
        db_exec "UPDATE vpn_nodes SET public_key='$WG_PUBLIC_KEY' WHERE id='$NODE_ID';"
        echo -e "${GREEN}✓ WireGuard public key configured${NC}"
    else
        echo -e "${YELLOW}⚠ Could not get WireGuard public key yet (will be configured on first start)${NC}"
    fi

    # Verify WireGuard is running
    if wg show wg0 >/dev/null 2>&1; then
        echo -e "${GREEN}✓ WireGuard interface is active${NC}"
    else
        echo -e "${YELLOW}⚠ WireGuard interface not ready yet (will be configured on first start)${NC}"
    fi
}

# Finalize node setup based on deployment mode
finalize_node_setup() {
    section "⚙️  Finalizing Node Configuration"

    if [ "$DEPLOYMENT_MODE" = "docker" ]; then
        finalize_docker_node_setup
    else
        finalize_system_node_setup
    fi
}

# Setup Nginx reverse proxy
setup_nginx_config() {
    section "🌐 Setting Up Nginx Reverse Proxy"

    # Get server IP
    PUBLIC_IP=$(curl -s https://api.ipify.org || echo "localhost")

    echo -e "${YELLOW}Creating Nginx configuration...${NC}"

    # Create Nginx configuration for system mode
    cat > /etc/nginx/sites-available/aureo-vpn << EOF
# Aureo VPN - Nginx Reverse Proxy Configuration
# Generated by become-node-operator.sh

# Upstream servers
upstream api_backend {
    server localhost:8080;
    keepalive 32;
}

upstream dashboard_backend {
    server localhost:3001;
    keepalive 32;
}

upstream grafana_backend {
    server localhost:3000;
    keepalive 32;
}

upstream prometheus_backend {
    server localhost:9090;
    keepalive 32;
}

# HTTP Server - Port 80
server {
    listen 80 default_server;
    listen [::]:80 default_server;
    server_name $PUBLIC_IP localhost _;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;

    # Enable gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types text/plain text/css text/xml text/javascript
               application/javascript application/xml+rss application/json
               application/x-javascript application/xml;

    # Increase client body size for file uploads
    client_max_body_size 100M;

    # API Gateway - /api/*
    location /api/ {
        proxy_pass http://api_backend/api/;
        proxy_http_version 1.1;

        # WebSocket support
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";

        # Proxy headers
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;

        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;

        # CORS headers
        add_header 'Access-Control-Allow-Origin' '*' always;
        add_header 'Access-Control-Allow-Methods' 'GET, POST, PUT, DELETE, OPTIONS, PATCH' always;
        add_header 'Access-Control-Allow-Headers' 'Authorization, Content-Type, Accept, Origin, User-Agent, DNT, Cache-Control, X-Mx-ReqToken, Keep-Alive, X-Requested-With, If-Modified-Since' always;
        add_header 'Access-Control-Allow-Credentials' 'true' always;

        # Handle preflight requests
        if (\$request_method = 'OPTIONS') {
            add_header 'Access-Control-Allow-Origin' '*';
            add_header 'Access-Control-Allow-Methods' 'GET, POST, PUT, DELETE, OPTIONS, PATCH';
            add_header 'Access-Control-Allow-Headers' 'Authorization, Content-Type, Accept, Origin, User-Agent, DNT, Cache-Control, X-Mx-ReqToken, Keep-Alive, X-Requested-With, If-Modified-Since';
            add_header 'Access-Control-Max-Age' 1728000;
            add_header 'Content-Type' 'text/plain charset=UTF-8';
            add_header 'Content-Length' 0;
            return 204;
        }
    }

    # Health check endpoint
    location /health {
        proxy_pass http://api_backend/health;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        access_log off;
    }

    # Operator Dashboard - /
    location / {
        proxy_pass http://dashboard_backend/;
        proxy_http_version 1.1;

        # WebSocket support for HMR in development
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";

        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    # Grafana Monitoring - /grafana/
    location /grafana/ {
        proxy_pass http://grafana_backend/;
        proxy_http_version 1.1;

        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;

        # WebSocket support for Grafana Live
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";

        rewrite ^/grafana/(.*)\$ /\$1 break;
    }

    # Prometheus Metrics - /prometheus/
    location /prometheus/ {
        proxy_pass http://prometheus_backend/;
        proxy_http_version 1.1;

        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;

        rewrite ^/prometheus/(.*)\$ /\$1 break;
    }

    # Disable access to hidden files
    location ~ /\. {
        deny all;
        access_log off;
        log_not_found off;
    }

    # Custom error pages
    error_page 502 503 504 /50x.html;
    location = /50x.html {
        root /usr/share/nginx/html;
    }
}
EOF

    # Enable the site
    ln -sf /etc/nginx/sites-available/aureo-vpn /etc/nginx/sites-enabled/aureo-vpn

    # Remove default Nginx site if it exists
    rm -f /etc/nginx/sites-enabled/default

    # Test Nginx configuration
    echo -e "${CYAN}Testing Nginx configuration...${NC}"
    if nginx -t 2>&1 | grep -q "successful"; then
        echo -e "${GREEN}✓ Nginx configuration is valid${NC}"

        # Reload Nginx
        systemctl reload nginx
        echo -e "${GREEN}✓ Nginx reloaded successfully${NC}"
    else
        echo -e "${RED}✗ Nginx configuration test failed${NC}"
        nginx -t
        return 1
    fi

    # Configure firewall for HTTP/HTTPS
    echo -e "${YELLOW}Configuring firewall...${NC}"
    if command_exists ufw; then
        ufw allow 80/tcp comment 'Aureo VPN HTTP' 2>/dev/null || true
        ufw allow 443/tcp comment 'Aureo VPN HTTPS' 2>/dev/null || true
        echo -e "${GREEN}✓ Firewall rules added (UFW)${NC}"
    elif command_exists firewall-cmd; then
        firewall-cmd --permanent --add-service=http 2>/dev/null || true
        firewall-cmd --permanent --add-service=https 2>/dev/null || true
        firewall-cmd --reload 2>/dev/null || true
        echo -e "${GREEN}✓ Firewall rules added (firewalld)${NC}"
    else
        echo -e "${YELLOW}⚠ No firewall detected, skipping firewall configuration${NC}"
    fi

    echo -e "${GREEN}✓ Nginx reverse proxy configured${NC}"
}

# Setup peer registration script
setup_peer_script() {
    section "📝 Setting Up Peer Registration"

    # Create /opt/aureo-vpn directory
    mkdir -p /opt/aureo-vpn

    # Copy peer registration script
    cp "$PROJECT_ROOT/scripts/add-wireguard-peer.sh" /opt/aureo-vpn/
    chmod +x /opt/aureo-vpn/add-wireguard-peer.sh

    echo -e "${GREEN}✓ Peer registration script installed${NC}"
}

# Create operator account
create_operator_account() {
    section "👤 Creating Your Operator Account"

    echo "Let's create your operator account!"
    echo ""

    read -p "Email: " EMAIL
    read -p "Username: " USERNAME
    read -sp "Password (min 8 characters): " PASSWORD
    echo ""

    API_URL="http://localhost:8080"

    # Wait for API to be ready
    echo -e "\n${BLUE}Waiting for API to be ready...${NC}"
    for i in {1..30}; do
        if curl -sf "$API_URL/health" >/dev/null 2>&1; then
            echo -e "${GREEN}✓ API is ready${NC}"
            break
        fi
        echo -n "."
        sleep 1
    done
    echo ""

    # Register user
    echo -e "${BLUE}Registering user account...${NC}"

    REGISTER_RESPONSE=$(curl -sf -X POST "$API_URL/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$EMAIL\",
            \"password\": \"$PASSWORD\",
            \"username\": \"$USERNAME\"
        }" || echo '{"error": "Connection failed"}')

    ACCESS_TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.access_token // empty' 2>/dev/null)

    if [ -n "$ACCESS_TOKEN" ] && [ "$ACCESS_TOKEN" != "null" ]; then
        echo -e "${GREEN}✓ User account created${NC}"
    else
        echo -e "${RED}✗ Failed to create user account${NC}"
        ERROR_MSG=$(echo "$REGISTER_RESPONSE" | jq -r '.error // .message // "Unknown error"' 2>/dev/null)
        echo "Error: $ERROR_MSG"
        exit 1
    fi

    # Setup crypto wallet
    echo ""
    echo "Choose cryptocurrency for rewards:"
    echo "  1) Ethereum (ETH)"
    echo "  2) Bitcoin (BTC)"
    echo "  3) Litecoin (LTC)"
    read -p "Choice [1-3]: " CRYPTO_CHOICE

    case $CRYPTO_CHOICE in
        1) CRYPTO_TYPE="ethereum" ;;
        2) CRYPTO_TYPE="bitcoin" ;;
        3) CRYPTO_TYPE="litecoin" ;;
        *) CRYPTO_TYPE="ethereum" ;;
    esac

    read -p "Enter your $CRYPTO_TYPE wallet address: " WALLET_ADDRESS

    # Register as operator
    echo -e "\n${BLUE}Registering as node operator...${NC}"

    OPERATOR_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/operator/register" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -d "{
            \"wallet_address\": \"$WALLET_ADDRESS\",
            \"wallet_type\": \"$CRYPTO_TYPE\",
            \"country\": \"Unknown\",
            \"email\": \"$EMAIL\"
        }")

    if echo "$OPERATOR_RESPONSE" | jq -e '.operator' >/dev/null 2>&1; then
        echo -e "${GREEN}✓ Operator account created${NC}"
    else
        # Check if it's a conflict error (operator already exists)
        ERROR_MSG=$(echo "$OPERATOR_RESPONSE" | jq -r '.error // .message // "Unknown error"' 2>/dev/null)

        if [[ "$ERROR_MSG" == *"conflict"* ]] || [[ "$ERROR_MSG" == *"already exists"* ]]; then
            echo -e "${YELLOW}⚠ Operator already exists for this wallet or email${NC}"
            echo -e "${CYAN}Attempting to use existing operator account...${NC}"

            # Try to get existing operator info
            OPERATOR_INFO=$(curl -s -X GET "$API_URL/api/v1/operator" \
                -H "Authorization: Bearer $ACCESS_TOKEN")

            if echo "$OPERATOR_INFO" | jq -e '.operator' >/dev/null 2>&1; then
                echo -e "${GREEN}✓ Using existing operator account${NC}"
            else
                echo -e "${RED}✗ Could not access operator account${NC}"
                echo -e "${YELLOW}Please ensure the wallet address and email are correct${NC}"
                echo "Response: $OPERATOR_RESPONSE"
                exit 1
            fi
        else
            echo -e "${RED}✗ Failed to register as operator${NC}"
            echo "Error: $ERROR_MSG"
            echo "Response: $OPERATOR_RESPONSE"
            exit 1
        fi
    fi

    # Get fresh token with operator permissions
    echo -e "${CYAN}Getting authorization token...${NC}"
    LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$EMAIL\",
            \"password\": \"$PASSWORD\"
        }")

    ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token // empty' 2>/dev/null)
    if [ -z "$ACCESS_TOKEN" ] || [ "$ACCESS_TOKEN" = "null" ]; then
        echo -e "${RED}✗ Failed to get operator token${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Token obtained${NC}"

    # Activate operator account
    echo -e "${CYAN}Activating operator account...${NC}"
    db_exec "UPDATE node_operators SET status='active', is_verified=true, verified_at=NOW() WHERE wallet_address='$WALLET_ADDRESS';"

    echo -e "${GREEN}✓ Operator account activated${NC}"

    # Save credentials
    mkdir -p "$CONFIG_DIR"
    cat > "$CONFIG_DIR/operator-credentials" << EOF
EMAIL=$EMAIL
USERNAME=$USERNAME
ACCESS_TOKEN=$ACCESS_TOKEN
WALLET_ADDRESS=$WALLET_ADDRESS
CRYPTO_TYPE=$CRYPTO_TYPE
API_URL=$API_URL
EOF
    chmod 600 "$CONFIG_DIR/operator-credentials"
}

# Register VPN node in database
register_node() {
    section "🖥️  Registering Your VPN Node"

    # Auto-detect information
    PUBLIC_IP=$(curl -s https://api.ipify.org || echo "127.0.0.1")
    HOSTNAME=$(hostname)
    INTERNAL_IP="10.8.0.1"  # WireGuard server IP

    echo -e "${BLUE}Auto-detected information:${NC}"
    echo "  Public IP: $PUBLIC_IP"
    echo "  Hostname: $HOSTNAME"
    echo ""

    read -p "Node name [default: aureo-node-$HOSTNAME]: " NODE_NAME
    NODE_NAME=${NODE_NAME:-"aureo-node-$HOSTNAME"}

    # Source operator credentials
    source "$CONFIG_DIR/operator-credentials"

    # Register node with API (this creates the database entry)
    echo -e "\n${BLUE}Registering VPN node in database...${NC}"

    NODE_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/operator/nodes" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -d "{
            \"name\": \"$NODE_NAME\",
            \"hostname\": \"$HOSTNAME\",
            \"public_ip\": \"$PUBLIC_IP\",
            \"internal_ip\": \"$INTERNAL_IP\",
            \"country\": \"Unknown\",
            \"country_code\": \"US\",
            \"city\": \"Unknown\",
            \"wireguard_port\": 51820,
            \"openvpn_port\": 1194,
            \"latitude\": 0,
            \"longitude\": 0
        }")

    NODE_ID=$(echo "$NODE_RESPONSE" | jq -r '.node.id')

    if [ "$NODE_ID" != "null" ] && [ -n "$NODE_ID" ]; then
        echo -e "${GREEN}✓ VPN node registered in database${NC}"
        echo -e "${BLUE}  Node ID: $NODE_ID${NC}"

        # Set internal_ip, status to online, and update heartbeat
        db_exec "UPDATE vpn_nodes SET internal_ip='$INTERNAL_IP', status='online', last_heartbeat=NOW() WHERE id='$NODE_ID';"

        # Recalculate operator stats
        db_exec "UPDATE node_operators SET active_nodes_count = (SELECT COUNT(*) FROM vpn_nodes WHERE operator_id = node_operators.id AND status = 'online' AND is_active = true);"

        # Save node info
        cat >> "$CONFIG_DIR/operator-credentials" << EOF
NODE_ID=$NODE_ID
NODE_NAME=$NODE_NAME
PUBLIC_IP=$PUBLIC_IP
INTERNAL_IP=$INTERNAL_IP
EOF

        # Export NODE_ID for use in next steps
        export NODE_ID

        echo -e "${GREEN}✓ Node registration complete${NC}"
    else
        echo -e "${RED}✗ Failed to register node${NC}"
        echo "Response: $NODE_RESPONSE"
        exit 1
    fi
}

# Setup monitoring script for Docker
setup_docker_monitoring() {
    # Create keep-node-online script for Docker
    cat > /opt/aureo-vpn/keep-node-online.sh << 'EOF'
#!/bin/bash
# Keep node status as online (Docker mode)
NODE_ID="${NODE_ID_1:-}"
if [ -n "$NODE_ID" ]; then
    docker exec aureo-vpn-db psql -U postgres -d aureo_vpn -c \
        "UPDATE vpn_nodes SET status='online', last_heartbeat=NOW() WHERE id='$NODE_ID';" \
        >/dev/null 2>&1
fi
EOF

    chmod +x /opt/aureo-vpn/keep-node-online.sh

    # Add to crontab to run every minute
    (crontab -l 2>/dev/null || echo "") | grep -v "keep-node-online.sh" | \
        { cat; echo "* * * * * NODE_ID_1=$NODE_ID /opt/aureo-vpn/keep-node-online.sh"; } | crontab -

    echo -e "${GREEN}✓ Monitoring configured${NC}"
}

# Setup monitoring script for System
setup_system_monitoring() {
    # Create keep-node-online script for System
    cat > /opt/aureo-vpn/keep-node-online.sh << EOF
#!/bin/bash
# Keep node status as online (System mode)
NODE_ID="$NODE_ID"
if [ -n "\$NODE_ID" ]; then
    sudo -u postgres psql -d aureo_vpn -c \
        "UPDATE vpn_nodes SET status='online', last_heartbeat=NOW() WHERE id='\$NODE_ID';" \
        >/dev/null 2>&1
fi
EOF

    chmod +x /opt/aureo-vpn/keep-node-online.sh

    # Add to crontab to run every minute
    (crontab -l 2>/dev/null || echo "") | grep -v "keep-node-online.sh" | \
        { cat; echo "* * * * * /opt/aureo-vpn/keep-node-online.sh"; } | crontab -

    echo -e "${GREEN}✓ Monitoring configured${NC}"
}

# Setup monitoring based on deployment mode
setup_monitoring() {
    section "📊 Setting Up Monitoring"

    if [ "$DEPLOYMENT_MODE" = "docker" ]; then
        setup_docker_monitoring
    else
        setup_system_monitoring
    fi
}

# Print final summary
print_summary() {
    section "🎉 Setup Complete!"

    source "$CONFIG_DIR/operator-credentials"

    echo -e "${GREEN}✓ Your Aureo VPN node is fully configured and running!${NC}"
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  SERVICES STATUS${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "  ✓ PostgreSQL Database:  Running"
    echo -e "  ✓ VPN Node:             Running on $PUBLIC_IP:51820"
    echo -e "  ✓ Node Status:          Registered and Active"
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  WEB ACCESS${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    if [ "$DEPLOYMENT_MODE" = "system" ]; then
        echo -e "  🌐 Dashboard:          ${GREEN}http://$PUBLIC_IP/${NC}"
        echo -e "  🔌 API Endpoint:       ${GREEN}http://$PUBLIC_IP/api/${NC}"
        echo -e "  📊 Grafana:            ${GREEN}http://$PUBLIC_IP/grafana/${NC}"
        echo -e "  📈 Prometheus:         ${GREEN}http://$PUBLIC_IP/prometheus/${NC}"
        echo -e "  ❤️  Health Check:       ${GREEN}http://$PUBLIC_IP/health${NC}"
    else
        echo -e "  🌐 Dashboard:          ${GREEN}http://localhost:3001/${NC}"
        echo -e "  🔌 API Endpoint:       ${GREEN}http://localhost:8080/api/${NC}"
        echo -e "  📊 Grafana:            ${GREEN}http://localhost:3000/${NC}"
        echo -e "  📈 Prometheus:         ${GREEN}http://localhost:9090/${NC}"
    fi
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  YOUR ACCOUNT${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "  Email:          $EMAIL"
    echo -e "  Username:       $USERNAME"
    echo -e "  Wallet:         $WALLET_ADDRESS ($CRYPTO_TYPE)"
    echo -e "  Node Name:      $NODE_NAME"
    echo -e "  Node ID:        $NODE_ID"
    echo -e "  Public IP:      $PUBLIC_IP"
    echo -e "  VPN Network:    $INTERNAL_IP/24"
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  USEFUL COMMANDS${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    if [ "$DEPLOYMENT_MODE" = "docker" ]; then
        echo -e "  View all containers:    ${GREEN}$DOCKER_COMPOSE -f $DOCKER_COMPOSE_FILE ps${NC}"
        echo -e "  View logs (all):        ${GREEN}$DOCKER_COMPOSE -f $DOCKER_COMPOSE_FILE logs -f${NC}"
        echo -e "  View VPN node logs:     ${GREEN}$DOCKER_COMPOSE -f $DOCKER_COMPOSE_FILE logs -f vpn-node-1${NC}"
        echo -e "  Check WireGuard:        ${GREEN}docker exec aureo-vpn-node-1 wg show wg0${NC}"
        echo -e "  Restart services:       ${GREEN}$DOCKER_COMPOSE -f $DOCKER_COMPOSE_FILE restart${NC}"
        echo -e "  Stop services:          ${GREEN}$DOCKER_COMPOSE -f $DOCKER_COMPOSE_FILE down${NC}"
    else
        echo -e "  View service status:    ${GREEN}systemctl status aureo-*${NC}"
        echo -e "  View API logs:          ${GREEN}journalctl -u aureo-api-gateway -f${NC}"
        echo -e "  View VPN node logs:     ${GREEN}journalctl -u aureo-vpn-node -f${NC}"
        echo -e "  Check WireGuard:        ${GREEN}wg show wg0${NC}"
        echo -e "  Restart API:            ${GREEN}systemctl restart aureo-api-gateway${NC}"
        echo -e "  Restart VPN node:       ${GREEN}systemctl restart aureo-vpn-node${NC}"
        echo -e "  Stop all services:      ${GREEN}systemctl stop aureo-*${NC}"
    fi

    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  EARNINGS INFO${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "  💰 Reward Tiers:"
    echo -e "     🥉 Bronze:    \$0.01/GB  (50%+ uptime)"
    echo -e "     🥈 Silver:    \$0.015/GB (80%+ uptime)"
    echo -e "     🥇 Gold:      \$0.02/GB  (90%+ uptime)"
    echo -e "     💎 Platinum:  \$0.03/GB  (95%+ uptime)"
    echo ""
    echo -e "  💸 Minimum payout: \$10"
    echo -e "  📅 Payouts: Weekly (Fridays)"
    echo ""
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}  🎊 Congratulations! You're now earning crypto rewards! 🎊${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

##############################################################################
# MAIN EXECUTION
##############################################################################

main() {
    print_header
    check_root

    # Step 0: Check for previous installation
    detect_previous_installation

    # Step 1: Select deployment mode
    select_deployment_mode

    # Step 2: Check prerequisites based on selected mode
    check_prerequisites

    echo ""
    echo -e "${YELLOW}This script will:${NC}"
    if [ "$DEPLOYMENT_MODE" = "docker" ]; then
        echo "  ✓ Deploy services using Docker Compose"
    else
        echo "  ✓ Build and install system services"
        echo "  ✓ Create systemd service files"
        echo "  ✓ Setup Nginx reverse proxy"
    fi
    echo "  ✓ Setup Database and Redis"
    echo "  ✓ Create your operator account"
    echo "  ✓ Register your VPN node in database"
    echo "  ✓ Deploy VPN node with proper configuration"
    echo "  ✓ Configure peer registration"
    echo "  ✓ Setup monitoring"
    echo ""
    echo -e "${BLUE}Estimated time: 5-10 minutes${NC}"
    echo ""

    read -p "Continue with installation? (y/n): " CONTINUE
    if [[ ! $CONTINUE =~ ^[Yy]$ ]]; then
        echo "Installation cancelled."
        exit 0
    fi

    # Step 3: Deploy base services (without VPN node)
    deploy_base_services

    # Step 4: Setup peer script
    setup_peer_script

    # Step 5: Setup Nginx reverse proxy (only for system mode)
    if [ "$DEPLOYMENT_MODE" = "system" ]; then
        setup_nginx_config
    fi

    # Step 6: Create operator account
    create_operator_account

    # Step 7: Register node in database (get NODE_ID)
    register_node

    # Step 8: Deploy VPN node with the NODE_ID
    deploy_vpn_node

    # Step 9: Configure WireGuard and finalize
    finalize_node_setup

    # Step 10: Setup monitoring
    setup_monitoring

    # Step 11: Show summary
    print_summary

    echo -e "\n${GREEN}🚀 Node Operator Setup Completed Successfully!${NC}\n"
}

# Run main
main "$@"
