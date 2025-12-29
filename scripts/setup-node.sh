#!/bin/bash

################################################################################
# Aureo VPN - Decentralized Node Setup
#
# One-command setup to run an Aureo VPN node with full monitoring stack
#
# Features:
#   - Fully decentralized (no central database)
#   - P2P node discovery (libp2p DHT + GossipSub)
#   - SQLite for local storage
#   - WireGuard VPN
#   - Operator Dashboard
#   - Prometheus + Grafana monitoring
#   - Nginx reverse proxy
#
# Supported OS:
#   - Ubuntu 20.04, 22.04, 24.04
#   - Debian 10, 11, 12
#   - CentOS 7, 8, 9
#   - RHEL 7, 8, 9
#   - Fedora 38+
#   - Alpine Linux
#
# Usage:
#   sudo bash setup-node.sh
#
################################################################################

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEPLOY_DIR="$PROJECT_ROOT/deployments/decentralized"
CONFIG_DIR="$HOME/.aureo-vpn"
LOG_FILE="/tmp/aureo-setup-$(date +%Y%m%d-%H%M%S).log"

# OS Detection variables
OS_TYPE=""
OS_VERSION=""
PKG_MANAGER=""

# Error handler
error_exit() {
    echo -e "\n${RED}Error: $1${NC}"
    echo -e "${YELLOW}Setup failed. Check log file: $LOG_FILE${NC}"
    exit 1
}

# Log function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >> "$LOG_FILE"
}

# Print header
print_header() {
    clear
    echo -e "${PURPLE}"
    cat << "EOF"
╔══════════════════════════════════════════════════════════════════╗
║                                                                  ║
║       🚀 Aureo VPN - Decentralized Node Setup 🚀                 ║
║                                                                  ║
║           Fully Decentralized • P2P Discovery                    ║
║           No Central Database • Easy Setup                       ║
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
        error_exit "Please run as root (use sudo)"
    fi
}

# Command exists helper
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

################################################################################
# OS DETECTION
################################################################################

detect_os() {
    section "🔍 Detecting Operating System"

    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS_TYPE=$ID
        OS_VERSION=$VERSION_ID
    elif [ -f /etc/redhat-release ]; then
        OS_TYPE="rhel"
        OS_VERSION=$(cat /etc/redhat-release | grep -oE '[0-9]+\.[0-9]+' | head -1)
    elif [ -f /etc/alpine-release ]; then
        OS_TYPE="alpine"
        OS_VERSION=$(cat /etc/alpine-release)
    else
        error_exit "Unsupported operating system"
    fi

    # Normalize OS type
    case "$OS_TYPE" in
        ubuntu|debian)
            PKG_MANAGER="apt"
            ;;
        centos|rhel|rocky|almalinux|fedora)
            if command_exists dnf; then
                PKG_MANAGER="dnf"
            else
                PKG_MANAGER="yum"
            fi
            ;;
        alpine)
            PKG_MANAGER="apk"
            ;;
        *)
            error_exit "Unsupported OS: $OS_TYPE"
            ;;
    esac

    echo -e "${GREEN}✓ Detected: $OS_TYPE $OS_VERSION (package manager: $PKG_MANAGER)${NC}"
    log "OS detected: $OS_TYPE $OS_VERSION, package manager: $PKG_MANAGER"
}

################################################################################
# PACKAGE INSTALLATION
################################################################################

# Update package manager
update_packages() {
    echo -e "${CYAN}Updating package lists...${NC}"

    case "$PKG_MANAGER" in
        apt)
            DEBIAN_FRONTEND=noninteractive apt-get update -qq >> "$LOG_FILE" 2>&1
            ;;
        dnf)
            dnf check-update -q >> "$LOG_FILE" 2>&1 || true
            ;;
        yum)
            yum check-update -q >> "$LOG_FILE" 2>&1 || true
            ;;
        apk)
            apk update >> "$LOG_FILE" 2>&1
            ;;
    esac

    echo -e "${GREEN}✓ Package lists updated${NC}"
}

# Install a package
install_package() {
    local package=$1
    local package_alt=${2:-$1}  # Alternative package name for different distros

    case "$PKG_MANAGER" in
        apt)
            DEBIAN_FRONTEND=noninteractive apt-get install -y -qq "$package" >> "$LOG_FILE" 2>&1
            ;;
        dnf)
            dnf install -y -q "$package_alt" >> "$LOG_FILE" 2>&1
            ;;
        yum)
            yum install -y -q "$package_alt" >> "$LOG_FILE" 2>&1
            ;;
        apk)
            apk add --no-cache "$package" >> "$LOG_FILE" 2>&1
            ;;
    esac
}

# Install essential system packages
install_system_dependencies() {
    section "📦 Installing System Dependencies"

    echo -e "${CYAN}Installing essential packages...${NC}"

    local packages_apt="curl wget git ca-certificates gnupg lsb-release software-properties-common apt-transport-https iptables iproute2 jq"
    local packages_dnf="curl wget git ca-certificates gnupg2 iptables iproute jq"
    local packages_apk="curl wget git ca-certificates gnupg iptables iproute2 jq bash"

    case "$PKG_MANAGER" in
        apt)
            DEBIAN_FRONTEND=noninteractive apt-get install -y -qq $packages_apt >> "$LOG_FILE" 2>&1
            ;;
        dnf|yum)
            $PKG_MANAGER install -y -q $packages_dnf >> "$LOG_FILE" 2>&1
            ;;
        apk)
            apk add --no-cache $packages_apk >> "$LOG_FILE" 2>&1
            ;;
    esac

    echo -e "${GREEN}✓ System dependencies installed${NC}"
}

################################################################################
# DOCKER INSTALLATION
################################################################################

install_docker_ubuntu_debian() {
    echo -e "${CYAN}Installing Docker for $OS_TYPE...${NC}"

    # Remove old versions
    apt-get remove -y docker docker-engine docker.io containerd runc >> "$LOG_FILE" 2>&1 || true

    # Install prerequisites
    DEBIAN_FRONTEND=noninteractive apt-get install -y -qq \
        ca-certificates curl gnupg lsb-release >> "$LOG_FILE" 2>&1

    # Add Docker GPG key
    install -m 0755 -d /etc/apt/keyrings
    if [ ! -f /etc/apt/keyrings/docker.gpg ]; then
        curl -fsSL "https://download.docker.com/linux/$OS_TYPE/gpg" | gpg --dearmor -o /etc/apt/keyrings/docker.gpg 2>/dev/null
        chmod a+r /etc/apt/keyrings/docker.gpg
    fi

    # Add Docker repository
    local codename
    if [ -n "$VERSION_CODENAME" ]; then
        codename=$VERSION_CODENAME
    else
        codename=$(lsb_release -cs 2>/dev/null || echo "focal")
    fi

    echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/$OS_TYPE $codename stable" | \
        tee /etc/apt/sources.list.d/docker.list > /dev/null

    # Install Docker
    apt-get update -qq >> "$LOG_FILE" 2>&1
    DEBIAN_FRONTEND=noninteractive apt-get install -y -qq \
        docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin >> "$LOG_FILE" 2>&1
}

install_docker_centos_rhel() {
    echo -e "${CYAN}Installing Docker for $OS_TYPE...${NC}"

    # Remove old versions
    $PKG_MANAGER remove -y docker docker-client docker-client-latest docker-common \
        docker-latest docker-latest-logrotate docker-logrotate docker-engine >> "$LOG_FILE" 2>&1 || true

    # Install prerequisites
    $PKG_MANAGER install -y yum-utils >> "$LOG_FILE" 2>&1

    # Add Docker repository
    yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo >> "$LOG_FILE" 2>&1

    # Install Docker
    $PKG_MANAGER install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin >> "$LOG_FILE" 2>&1
}

install_docker_fedora() {
    echo -e "${CYAN}Installing Docker for Fedora...${NC}"

    # Remove old versions
    dnf remove -y docker docker-client docker-client-latest docker-common \
        docker-latest docker-latest-logrotate docker-logrotate docker-selinux \
        docker-engine-selinux docker-engine >> "$LOG_FILE" 2>&1 || true

    # Install prerequisites
    dnf install -y dnf-plugins-core >> "$LOG_FILE" 2>&1

    # Add Docker repository
    dnf config-manager --add-repo https://download.docker.com/linux/fedora/docker-ce.repo >> "$LOG_FILE" 2>&1

    # Install Docker
    dnf install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin >> "$LOG_FILE" 2>&1
}

install_docker_alpine() {
    echo -e "${CYAN}Installing Docker for Alpine...${NC}"

    apk add --no-cache docker docker-cli docker-cli-compose >> "$LOG_FILE" 2>&1
}

install_docker() {
    section "🐳 Installing Docker"

    if command_exists docker; then
        echo -e "${GREEN}✓ Docker already installed${NC}"
        # Check version
        local docker_version=$(docker --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+' | head -1)
        echo -e "${CYAN}  Version: $docker_version${NC}"
    else
        case "$OS_TYPE" in
            ubuntu|debian)
                install_docker_ubuntu_debian
                ;;
            centos|rhel|rocky|almalinux)
                install_docker_centos_rhel
                ;;
            fedora)
                install_docker_fedora
                ;;
            alpine)
                install_docker_alpine
                ;;
            *)
                error_exit "Docker installation not supported for $OS_TYPE"
                ;;
        esac
        echo -e "${GREEN}✓ Docker installed${NC}"
    fi

    # Start and enable Docker
    start_docker_service

    # Verify Docker Compose
    verify_docker_compose
}

start_docker_service() {
    echo -e "${CYAN}Starting Docker service...${NC}"

    if command_exists systemctl; then
        systemctl start docker >> "$LOG_FILE" 2>&1 || true
        systemctl enable docker >> "$LOG_FILE" 2>&1 || true
    elif command_exists rc-service; then
        rc-service docker start >> "$LOG_FILE" 2>&1 || true
        rc-update add docker boot >> "$LOG_FILE" 2>&1 || true
    elif command_exists service; then
        service docker start >> "$LOG_FILE" 2>&1 || true
    fi

    # Wait for Docker to be ready
    local retries=0
    while ! docker ps >/dev/null 2>&1 && [ $retries -lt 30 ]; do
        sleep 1
        retries=$((retries + 1))
    done

    if docker ps >/dev/null 2>&1; then
        echo -e "${GREEN}✓ Docker daemon running${NC}"
    else
        error_exit "Failed to start Docker daemon"
    fi
}

verify_docker_compose() {
    if docker compose version >/dev/null 2>&1; then
        DOCKER_COMPOSE="docker compose"
        echo -e "${GREEN}✓ Docker Compose (plugin) available${NC}"
    elif command_exists docker-compose; then
        DOCKER_COMPOSE="docker-compose"
        echo -e "${GREEN}✓ Docker Compose (standalone) available${NC}"
    else
        # Try to install docker-compose standalone
        echo -e "${YELLOW}Installing Docker Compose...${NC}"
        curl -fsSL "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" \
            -o /usr/local/bin/docker-compose >> "$LOG_FILE" 2>&1
        chmod +x /usr/local/bin/docker-compose
        DOCKER_COMPOSE="docker-compose"
        echo -e "${GREEN}✓ Docker Compose installed${NC}"
    fi
}

################################################################################
# SYSTEM CONFIGURATION
################################################################################

configure_system() {
    section "⚙️  Configuring System"

    # Enable IP forwarding
    echo -e "${CYAN}Enabling IP forwarding...${NC}"

    # Set immediately
    sysctl -w net.ipv4.ip_forward=1 >> "$LOG_FILE" 2>&1 || true
    sysctl -w net.ipv6.conf.all.forwarding=1 >> "$LOG_FILE" 2>&1 || true

    # Make persistent
    if [ ! -f /etc/sysctl.d/99-aureo-vpn.conf ]; then
        cat > /etc/sysctl.d/99-aureo-vpn.conf << EOF
# Aureo VPN - Network Configuration
net.ipv4.ip_forward = 1
net.ipv6.conf.all.forwarding = 1
net.core.rmem_max = 26214400
net.core.wmem_max = 26214400
EOF
        sysctl -p /etc/sysctl.d/99-aureo-vpn.conf >> "$LOG_FILE" 2>&1 || true
    fi
    echo -e "${GREEN}✓ IP forwarding enabled${NC}"

    # Load required kernel modules
    echo -e "${CYAN}Loading kernel modules...${NC}"
    modprobe ip_tables >> "$LOG_FILE" 2>&1 || true
    modprobe iptable_nat >> "$LOG_FILE" 2>&1 || true
    modprobe iptable_filter >> "$LOG_FILE" 2>&1 || true

    # Try to load WireGuard module (may not be available on all systems)
    modprobe wireguard >> "$LOG_FILE" 2>&1 || true

    echo -e "${GREEN}✓ Kernel modules loaded${NC}"

    # Create TUN device if not exists
    if [ ! -c /dev/net/tun ]; then
        echo -e "${CYAN}Creating TUN device...${NC}"
        mkdir -p /dev/net
        mknod /dev/net/tun c 10 200 2>/dev/null || true
        chmod 666 /dev/net/tun 2>/dev/null || true
    fi
    echo -e "${GREEN}✓ TUN device ready${NC}"
}

################################################################################
# FIREWALL CONFIGURATION
################################################################################

configure_firewall() {
    section "🔥 Configuring Firewall"

    local ports="22/tcp 80/tcp 8080/tcp 4001/tcp 4001/udp 51820/udp 3000/tcp 9090/tcp"

    if command_exists ufw; then
        echo -e "${CYAN}Configuring UFW firewall...${NC}"

        # Enable UFW if not enabled
        ufw --force enable >> "$LOG_FILE" 2>&1 || true

        # Allow SSH first (safety)
        ufw allow 22/tcp comment 'SSH' >> "$LOG_FILE" 2>&1 || true

        # Allow Aureo ports
        ufw allow 80/tcp comment 'Aureo HTTP' >> "$LOG_FILE" 2>&1 || true
        ufw allow 8080/tcp comment 'Aureo API' >> "$LOG_FILE" 2>&1 || true
        ufw allow 4001/tcp comment 'Aureo P2P TCP' >> "$LOG_FILE" 2>&1 || true
        ufw allow 4001/udp comment 'Aureo P2P QUIC' >> "$LOG_FILE" 2>&1 || true
        ufw allow 51820/udp comment 'Aureo WireGuard' >> "$LOG_FILE" 2>&1 || true
        ufw allow 3000/tcp comment 'Grafana' >> "$LOG_FILE" 2>&1 || true
        ufw allow 9090/tcp comment 'Prometheus' >> "$LOG_FILE" 2>&1 || true

        echo -e "${GREEN}✓ UFW firewall configured${NC}"

    elif command_exists firewall-cmd; then
        echo -e "${CYAN}Configuring firewalld...${NC}"

        # Start firewalld if not running
        systemctl start firewalld >> "$LOG_FILE" 2>&1 || true

        firewall-cmd --permanent --add-port=22/tcp >> "$LOG_FILE" 2>&1 || true
        firewall-cmd --permanent --add-port=80/tcp >> "$LOG_FILE" 2>&1 || true
        firewall-cmd --permanent --add-port=8080/tcp >> "$LOG_FILE" 2>&1 || true
        firewall-cmd --permanent --add-port=4001/tcp >> "$LOG_FILE" 2>&1 || true
        firewall-cmd --permanent --add-port=4001/udp >> "$LOG_FILE" 2>&1 || true
        firewall-cmd --permanent --add-port=51820/udp >> "$LOG_FILE" 2>&1 || true
        firewall-cmd --permanent --add-port=3000/tcp >> "$LOG_FILE" 2>&1 || true
        firewall-cmd --permanent --add-port=9090/tcp >> "$LOG_FILE" 2>&1 || true
        firewall-cmd --permanent --add-masquerade >> "$LOG_FILE" 2>&1 || true
        firewall-cmd --reload >> "$LOG_FILE" 2>&1 || true

        echo -e "${GREEN}✓ firewalld configured${NC}"

    elif command_exists iptables; then
        echo -e "${CYAN}Configuring iptables...${NC}"

        # Allow established connections
        iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT >> "$LOG_FILE" 2>&1 || true

        # Allow loopback
        iptables -A INPUT -i lo -j ACCEPT >> "$LOG_FILE" 2>&1 || true

        # Allow required ports
        for port_proto in $ports; do
            local port=$(echo $port_proto | cut -d'/' -f1)
            local proto=$(echo $port_proto | cut -d'/' -f2)
            iptables -A INPUT -p $proto --dport $port -j ACCEPT >> "$LOG_FILE" 2>&1 || true
        done

        # Enable NAT for VPN
        iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE >> "$LOG_FILE" 2>&1 || true

        # Save iptables rules
        if command_exists iptables-save; then
            iptables-save > /etc/iptables.rules 2>/dev/null || true
        fi

        echo -e "${GREEN}✓ iptables configured${NC}"
    else
        echo -e "${YELLOW}⚠ No firewall detected. Ensure ports 80, 4001, 8080, 51820 are open.${NC}"
    fi
}

################################################################################
# NETWORK DETECTION
################################################################################

detect_info() {
    section "🌍 Detecting Server Information"

    echo -e "${CYAN}Detecting public IP...${NC}"

    # Try multiple IP detection services
    PUBLIC_IP=""
    for service in "https://api.ipify.org" "https://ifconfig.me/ip" "https://icanhazip.com" "https://ipecho.net/plain"; do
        PUBLIC_IP=$(curl -s --max-time 5 "$service" 2>/dev/null | tr -d '\n')
        if [[ "$PUBLIC_IP" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
            break
        fi
        PUBLIC_IP=""
    done

    if [ -n "$PUBLIC_IP" ]; then
        echo -e "${GREEN}✓ Public IP: $PUBLIC_IP${NC}"
    else
        echo -e "${YELLOW}⚠ Could not detect public IP automatically${NC}"
        read -p "Enter your public IP address: " PUBLIC_IP
        if [ -z "$PUBLIC_IP" ]; then
            error_exit "Public IP is required"
        fi
    fi

    echo -e "${CYAN}Detecting location...${NC}"
    if [ -n "$PUBLIC_IP" ]; then
        LOCATION=$(curl -s --max-time 5 "http://ip-api.com/json/$PUBLIC_IP" 2>/dev/null || echo "{}")
        if [ -n "$LOCATION" ] && [ "$LOCATION" != "{}" ]; then
            AUTO_COUNTRY=$(echo "$LOCATION" | grep -o '"country":"[^"]*"' | cut -d'"' -f4)
            AUTO_COUNTRY_CODE=$(echo "$LOCATION" | grep -o '"countryCode":"[^"]*"' | cut -d'"' -f4)
            AUTO_CITY=$(echo "$LOCATION" | grep -o '"city":"[^"]*"' | cut -d'"' -f4)
            if [ -n "$AUTO_CITY" ] && [ -n "$AUTO_COUNTRY" ]; then
                echo -e "${GREEN}✓ Location: $AUTO_CITY, $AUTO_COUNTRY ($AUTO_COUNTRY_CODE)${NC}"
            fi
        fi
    fi
}

################################################################################
# USER CONFIGURATION
################################################################################

get_config() {
    section "⚙️  Node Configuration"

    echo -e "${CYAN}Configure your node (press Enter for defaults):${NC}\n"

    # Node name
    HOSTNAME=$(hostname)
    read -p "Node name [aureo-$HOSTNAME]: " NODE_NAME
    NODE_NAME=${NODE_NAME:-"aureo-$HOSTNAME"}

    # Country code
    read -p "Country code [${AUTO_COUNTRY_CODE:-US}]: " COUNTRY_CODE
    COUNTRY_CODE=${COUNTRY_CODE:-${AUTO_COUNTRY_CODE:-US}}

    # Country name
    read -p "Country name [${AUTO_COUNTRY:-Unknown}]: " COUNTRY
    COUNTRY=${COUNTRY:-${AUTO_COUNTRY:-Unknown}}

    # City
    read -p "City [${AUTO_CITY:-Unknown}]: " CITY
    CITY=${CITY:-${AUTO_CITY:-Unknown}}

    # Grafana password
    read -sp "Grafana admin password [admin]: " GRAFANA_PASSWORD
    echo ""
    GRAFANA_PASSWORD=${GRAFANA_PASSWORD:-admin}

    # Bootstrap peers
    echo ""
    echo -e "${YELLOW}Bootstrap peers (optional)${NC}"
    echo -e "Enter multiaddr of another node to connect to the network."
    echo -e "Example: /ip4/1.2.3.4/tcp/4001/p2p/12D3KooW..."
    read -p "Bootstrap peer [leave empty]: " BOOTSTRAP_PEERS

    echo ""
}

get_admin_config() {
    section "👤 Admin User Setup"

    echo -e "${CYAN}Create an admin account to connect to the VPN:${NC}\n"

    # Admin email
    while true; do
        read -p "Admin email: " ADMIN_EMAIL
        if [[ "$ADMIN_EMAIL" =~ ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
            break
        else
            echo -e "${RED}Please enter a valid email address${NC}"
        fi
    done

    # Admin username
    read -p "Admin username [admin]: " ADMIN_USERNAME
    ADMIN_USERNAME=${ADMIN_USERNAME:-admin}

    # Admin password
    while true; do
        read -sp "Admin password (min 6 chars): " ADMIN_PASSWORD
        echo ""
        if [ ${#ADMIN_PASSWORD} -ge 6 ]; then
            read -sp "Confirm password: " ADMIN_PASSWORD_CONFIRM
            echo ""
            if [ "$ADMIN_PASSWORD" = "$ADMIN_PASSWORD_CONFIRM" ]; then
                break
            else
                echo -e "${RED}Passwords don't match. Try again.${NC}"
            fi
        else
            echo -e "${RED}Password must be at least 6 characters${NC}"
        fi
    done

    echo ""
}

################################################################################
# CONFIGURATION FILES
################################################################################

create_env_file() {
    section "📝 Creating Configuration"

    cat > "$DEPLOY_DIR/.env" << EOF
# Aureo VPN - Decentralized Node Configuration
# Generated by setup-node.sh at $(date)

# Node Configuration
NODE_NAME=$NODE_NAME
PUBLIC_IP=$PUBLIC_IP
COUNTRY=$COUNTRY
COUNTRY_CODE=$COUNTRY_CODE
CITY=$CITY

# P2P Configuration
BOOTSTRAP_PEERS=$BOOTSTRAP_PEERS

# Grafana
GRAFANA_USER=admin
GRAFANA_PASSWORD=$GRAFANA_PASSWORD

# Generated at
SETUP_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
EOF

    chmod 600 "$DEPLOY_DIR/.env"
    echo -e "${GREEN}✓ Configuration saved${NC}"
}

################################################################################
# BUILD VERIFICATION
################################################################################

verify_build_requirements() {
    section "🔧 Verifying Build Requirements"

    local errors=0

    echo -e "${CYAN}Checking required files...${NC}"

    # Check Dockerfiles
    if [ ! -f "$DEPLOY_DIR/Dockerfile" ]; then
        echo -e "${RED}✗ Missing: $DEPLOY_DIR/Dockerfile${NC}"
        errors=$((errors + 1))
    else
        echo -e "${GREEN}✓ Dockerfile${NC}"
    fi

    if [ ! -f "$DEPLOY_DIR/Dockerfile.dashboard" ]; then
        echo -e "${RED}✗ Missing: $DEPLOY_DIR/Dockerfile.dashboard${NC}"
        errors=$((errors + 1))
    else
        echo -e "${GREEN}✓ Dockerfile.dashboard${NC}"
    fi

    # Check docker-compose
    if [ ! -f "$DEPLOY_DIR/docker-compose.yml" ]; then
        echo -e "${RED}✗ Missing: $DEPLOY_DIR/docker-compose.yml${NC}"
        errors=$((errors + 1))
    else
        echo -e "${GREEN}✓ docker-compose.yml${NC}"
    fi

    # Check control-panel package.json
    if [ ! -f "$PROJECT_ROOT/control-panel/package.json" ]; then
        echo -e "${RED}✗ Missing: control-panel/package.json${NC}"
        errors=$((errors + 1))
    else
        echo -e "${GREEN}✓ control-panel/package.json${NC}"
    fi

    # Check Go files
    if [ ! -f "$PROJECT_ROOT/go.mod" ]; then
        echo -e "${RED}✗ Missing: go.mod${NC}"
        errors=$((errors + 1))
    else
        echo -e "${GREEN}✓ go.mod${NC}"
    fi

    if [ $errors -gt 0 ]; then
        error_exit "Missing $errors required files. Please check your installation."
    fi

    echo -e "\n${GREEN}✓ All build requirements verified${NC}"
}

################################################################################
# DEPLOYMENT
################################################################################

deploy_services() {
    section "🚀 Deploying Services"

    cd "$DEPLOY_DIR"

    echo -e "${CYAN}Building and starting containers...${NC}"
    echo -e "${YELLOW}This may take 5-10 minutes on first run...${NC}\n"

    # Stop any existing containers
    $DOCKER_COMPOSE down >> "$LOG_FILE" 2>&1 || true

    # Clean up any stale build cache
    docker builder prune -f >> "$LOG_FILE" 2>&1 || true

    # Try to build and start (with retry logic)
    local max_retries=3
    local retry=0
    local build_success=false

    while [ $retry -lt $max_retries ] && [ "$build_success" = "false" ]; do
        if [ $retry -gt 0 ]; then
            echo -e "\n${YELLOW}Retrying build (attempt $((retry + 1))/$max_retries)...${NC}"
            echo -e "${CYAN}Cleaning build cache and retrying...${NC}\n"
            docker builder prune -af >> "$LOG_FILE" 2>&1 || true
            sleep 5
        fi

        # Run docker compose with timeout
        if timeout 600 $DOCKER_COMPOSE up -d --build >> "$LOG_FILE" 2>&1; then
            # Verify containers are running
            sleep 5
            if docker ps | grep -q "aureo-node"; then
                build_success=true
            else
                echo -e "${YELLOW}Containers not running, retrying...${NC}"
                retry=$((retry + 1))
            fi
        else
            echo -e "${YELLOW}Build command failed, retrying...${NC}"
            retry=$((retry + 1))
        fi
    done

    if [ "$build_success" = "false" ]; then
        echo -e "\n${RED}✗ Failed to build containers after $max_retries attempts${NC}"
        echo -e "${YELLOW}Check the log file for details: $LOG_FILE${NC}"
        echo -e "${YELLOW}Try running manually:${NC}"
        echo -e "  cd $DEPLOY_DIR && docker compose up -d --build"
        exit 1
    fi

    echo -e "\n${GREEN}✓ Services deployed${NC}"
}

wait_for_services() {
    section "⏳ Waiting for Services"

    echo -e "${CYAN}Waiting for services to be ready...${NC}"

    # Wait for API
    local api_ready=false
    for i in {1..60}; do
        if curl -s --max-time 2 http://localhost:8080/health &>/dev/null; then
            echo -e "${GREEN}✓ API is ready${NC}"
            api_ready=true
            break
        fi
        echo -n "."
        sleep 2
    done

    if [ "$api_ready" = "false" ]; then
        echo -e "\n${YELLOW}⚠ API not responding yet, continuing anyway...${NC}"
    fi

    # Wait for Dashboard
    local dashboard_ready=false
    for i in {1..30}; do
        if curl -s --max-time 2 http://localhost:5000 &>/dev/null; then
            echo -e "${GREEN}✓ Dashboard is ready${NC}"
            dashboard_ready=true
            break
        fi
        echo -n "."
        sleep 2
    done

    if [ "$dashboard_ready" = "false" ]; then
        echo -e "\n${YELLOW}⚠ Dashboard not responding yet, continuing anyway...${NC}"
    fi

    echo ""
}

create_admin_user() {
    section "👤 Creating Admin User"

    echo -e "${CYAN}Registering admin user...${NC}"

    # Wait a bit more for API to be fully ready
    sleep 3

    local retries=0
    local max_retries=5

    while [ $retries -lt $max_retries ]; do
        REGISTER_RESPONSE=$(curl -s -X POST "http://localhost:8080/api/v1/auth/register" \
            -H "Content-Type: application/json" \
            -d "{\"email\":\"$ADMIN_EMAIL\",\"username\":\"$ADMIN_USERNAME\",\"password\":\"$ADMIN_PASSWORD\"}" 2>/dev/null)

        if [ -n "$REGISTER_RESPONSE" ]; then
            break
        fi

        retries=$((retries + 1))
        sleep 2
    done

    # Check for errors
    if echo "$REGISTER_RESPONSE" | grep -q '"error"'; then
        ERROR_MSG=$(echo "$REGISTER_RESPONSE" | grep -o '"error":"[^"]*"' | cut -d'"' -f4)
        if [ "$ERROR_MSG" = "user already exists" ]; then
            echo -e "${YELLOW}⚠ User already exists. You can login with existing credentials.${NC}"
        else
            echo -e "${YELLOW}⚠ Note: $ERROR_MSG${NC}"
            echo -e "${CYAN}You can create a user manually via the dashboard.${NC}"
        fi
    else
        echo -e "${GREEN}✓ Admin user created successfully${NC}"
        ADMIN_CREATED=true
    fi
}

get_node_info() {
    NODE_INFO=$(curl -s http://localhost:8080/info 2>/dev/null || echo "{}")
    P2P_STATUS=$(curl -s http://localhost:8080/api/v1/p2p/status 2>/dev/null || echo "{}")

    NODE_ID=$(echo "$NODE_INFO" | grep -o '"id":"[^"]*"' | cut -d'"' -f4 || echo "")
    PEER_ID=$(echo "$P2P_STATUS" | grep -o '"peer_id":"[^"]*"' | cut -d'"' -f4 || echo "")
}

create_management_script() {
    cat > /usr/local/bin/aureo << 'SCRIPT'
#!/bin/bash
DEPLOY_DIR="${AUREO_DEPLOY_DIR:-/opt/aureo-vpn}"

cd "$DEPLOY_DIR" 2>/dev/null || cd "$(dirname "$(readlink -f "$0")")/../deployments/decentralized" 2>/dev/null

case "$1" in
    start)
        docker compose up -d
        ;;
    stop)
        docker compose down
        ;;
    restart)
        docker compose restart
        ;;
    logs)
        docker logs -f ${2:-aureo-node}
        ;;
    status)
        curl -s http://localhost:8080/info 2>/dev/null | python3 -m json.tool 2>/dev/null || echo "Node not running"
        ;;
    peers)
        curl -s http://localhost:8080/api/v1/p2p/peers 2>/dev/null | python3 -m json.tool 2>/dev/null || echo "Node not running"
        ;;
    nodes)
        curl -s http://localhost:8080/api/v1/nodes 2>/dev/null | python3 -m json.tool 2>/dev/null || echo "Node not running"
        ;;
    ps)
        docker compose ps
        ;;
    update)
        echo "Pulling latest changes..."
        git pull
        echo "Rebuilding containers..."
        docker compose up -d --build
        ;;
    *)
        echo "Aureo VPN Node Management"
        echo ""
        echo "Usage: aureo <command>"
        echo ""
        echo "Commands:"
        echo "  start          Start all services"
        echo "  stop           Stop all services"
        echo "  restart        Restart all services"
        echo "  logs [service] View logs (default: aureo-node)"
        echo "  status         Show node status"
        echo "  peers          List connected P2P peers"
        echo "  nodes          List known VPN nodes"
        echo "  ps             Show container status"
        echo "  update         Update and rebuild containers"
        ;;
esac
SCRIPT

    chmod +x /usr/local/bin/aureo

    # Create profile script
    cat > /etc/profile.d/aureo.sh << EOF
export AUREO_DEPLOY_DIR="$DEPLOY_DIR"
EOF
    chmod +x /etc/profile.d/aureo.sh
}

save_credentials() {
    mkdir -p "$CONFIG_DIR"
    cat > "$CONFIG_DIR/node-info" << EOF
NODE_NAME=$NODE_NAME
NODE_ID=$NODE_ID
PEER_ID=$PEER_ID
PUBLIC_IP=$PUBLIC_IP
COUNTRY=$COUNTRY
COUNTRY_CODE=$COUNTRY_CODE
CITY=$CITY
DEPLOY_DIR=$DEPLOY_DIR
ADMIN_EMAIL=$ADMIN_EMAIL
ADMIN_USERNAME=$ADMIN_USERNAME
SETUP_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
EOF
    chmod 600 "$CONFIG_DIR/node-info"

    # Save admin password separately
    echo "$ADMIN_PASSWORD" > "$CONFIG_DIR/.admin-password"
    chmod 600 "$CONFIG_DIR/.admin-password"
}

print_summary() {
    section "🎉 Setup Complete!"

    echo -e "${GREEN}Your Aureo VPN node is running!${NC}\n"

    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  NODE INFORMATION${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "  Node Name:     ${GREEN}$NODE_NAME${NC}"
    echo -e "  Node ID:       ${GREEN}${NODE_ID:-detecting...}${NC}"
    echo -e "  Peer ID:       ${GREEN}${PEER_ID:-detecting...}${NC}"
    echo -e "  Location:      ${GREEN}$CITY, $COUNTRY ($COUNTRY_CODE)${NC}"
    echo -e "  Public IP:     ${GREEN}$PUBLIC_IP${NC}"
    echo ""

    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  VPN LOGIN CREDENTIALS${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "  Email:         ${GREEN}$ADMIN_EMAIL${NC}"
    echo -e "  Username:      ${GREEN}$ADMIN_USERNAME${NC}"
    echo -e "  Password:      ${GREEN}$ADMIN_PASSWORD${NC}"
    echo ""

    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  WEB ACCESS${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "  Dashboard:     ${GREEN}http://$PUBLIC_IP/${NC}"
    echo -e "  API:           ${GREEN}http://$PUBLIC_IP:8080/api/v1/${NC}"
    echo -e "  Grafana:       ${GREEN}http://$PUBLIC_IP:3000/${NC}  (admin / $GRAFANA_PASSWORD)"
    echo -e "  Prometheus:    ${GREEN}http://$PUBLIC_IP:9090/${NC}"
    echo -e "  Health:        ${GREEN}http://$PUBLIC_IP:8080/health${NC}"
    echo ""

    if [ -n "$PEER_ID" ]; then
        echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${CYAN}  P2P MULTIADDR (share with other nodes)${NC}"
        echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo ""
        echo -e "  ${YELLOW}/ip4/$PUBLIC_IP/tcp/4001/p2p/$PEER_ID${NC}"
        echo ""
    fi

    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  MANAGEMENT COMMANDS${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "  ${GREEN}aureo status${NC}    - Show node status"
    echo -e "  ${GREEN}aureo logs${NC}      - View node logs"
    echo -e "  ${GREEN}aureo peers${NC}     - List connected P2P peers"
    echo -e "  ${GREEN}aureo restart${NC}   - Restart all services"
    echo -e "  ${GREEN}aureo stop${NC}      - Stop all services"
    echo -e "  ${GREEN}aureo ps${NC}        - Show container status"
    echo ""

    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}  Node is ready to accept connections!${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "${CYAN}Log file: $LOG_FILE${NC}"
    echo ""
}

################################################################################
# MAIN EXECUTION
################################################################################

main() {
    # Initialize log file
    echo "Aureo VPN Setup - $(date)" > "$LOG_FILE"

    print_header
    check_root
    detect_os
    update_packages
    install_system_dependencies
    install_docker
    configure_system
    detect_info
    get_config
    get_admin_config
    create_env_file
    configure_firewall
    verify_build_requirements
    deploy_services
    wait_for_services
    create_admin_user
    get_node_info
    create_management_script
    save_credentials
    print_summary

    echo -e "\n${GREEN}Setup completed successfully!${NC}\n"
}

# Run main
main "$@"
