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
# Usage:
#   sudo bash setup-node.sh
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
NC='\033[0m'

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEPLOY_DIR="$PROJECT_ROOT/deployments/decentralized"
CONFIG_DIR="$HOME/.aureo-vpn"

# Trap errors
trap 'error_handler $? $LINENO' ERR

error_handler() {
    echo -e "\n${RED}Error occurred at line $2 (exit code: $1)${NC}"
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
        echo -e "${RED}Please run as root (use sudo)${NC}"
        exit 1
    fi
}

# Command exists helper
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Install Docker
install_docker() {
    section "📦 Installing Docker"

    echo -e "${YELLOW}Installing Docker...${NC}"

    # Install prerequisites
    apt-get update -qq
    DEBIAN_FRONTEND=noninteractive apt-get install -y -qq \
        ca-certificates curl gnupg lsb-release

    # Add Docker GPG key
    install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg 2>/dev/null || true
    chmod a+r /etc/apt/keyrings/docker.gpg

    # Add Docker repository
    echo \
        "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
        $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
        tee /etc/apt/sources.list.d/docker.list > /dev/null

    # Install Docker
    apt-get update -qq
    DEBIAN_FRONTEND=noninteractive apt-get install -y -qq \
        docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

    # Start Docker
    systemctl start docker
    systemctl enable docker

    echo -e "${GREEN}✓ Docker installed successfully${NC}"
}

# Check prerequisites
check_prerequisites() {
    section "🔍 Checking Prerequisites"

    # Check Docker
    if ! command_exists docker; then
        echo -e "${YELLOW}Docker not found. Installing...${NC}"
        install_docker
    else
        echo -e "${GREEN}✓ Docker installed${NC}"
    fi

    # Check Docker Compose
    if docker compose version >/dev/null 2>&1; then
        DOCKER_COMPOSE="docker compose"
        echo -e "${GREEN}✓ Docker Compose installed${NC}"
    elif command_exists docker-compose; then
        DOCKER_COMPOSE="docker-compose"
        echo -e "${GREEN}✓ Docker Compose installed${NC}"
    else
        echo -e "${RED}Docker Compose not found${NC}"
        exit 1
    fi

    # Check Docker is running
    if ! docker ps >/dev/null 2>&1; then
        echo -e "${YELLOW}Starting Docker...${NC}"
        systemctl start docker
        sleep 3
    fi
    echo -e "${GREEN}✓ Docker daemon running${NC}"

    # Check project files
    if [ ! -f "$DEPLOY_DIR/docker-compose.yml" ]; then
        echo -e "${RED}✗ Deployment files not found at $DEPLOY_DIR${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Project files found${NC}"
}

# Detect public IP and location
detect_info() {
    section "🌍 Detecting Server Information"

    echo -e "${CYAN}Detecting public IP...${NC}"
    PUBLIC_IP=$(curl -s --max-time 5 https://api.ipify.org 2>/dev/null || \
                curl -s --max-time 5 https://ifconfig.me/ip 2>/dev/null || \
                curl -s --max-time 5 https://icanhazip.com 2>/dev/null || \
                echo "")

    if [ -n "$PUBLIC_IP" ]; then
        echo -e "${GREEN}✓ Public IP: $PUBLIC_IP${NC}"
    else
        echo -e "${YELLOW}⚠ Could not detect public IP${NC}"
    fi

    echo -e "${CYAN}Detecting location...${NC}"
    if [ -n "$PUBLIC_IP" ]; then
        LOCATION=$(curl -s --max-time 5 "http://ip-api.com/json/$PUBLIC_IP" 2>/dev/null || echo "{}")
        if [ -n "$LOCATION" ] && [ "$LOCATION" != "{}" ]; then
            AUTO_COUNTRY=$(echo "$LOCATION" | grep -o '"country":"[^"]*"' | cut -d'"' -f4)
            AUTO_COUNTRY_CODE=$(echo "$LOCATION" | grep -o '"countryCode":"[^"]*"' | cut -d'"' -f4)
            AUTO_CITY=$(echo "$LOCATION" | grep -o '"city":"[^"]*"' | cut -d'"' -f4)
            echo -e "${GREEN}✓ Location: $AUTO_CITY, $AUTO_COUNTRY ($AUTO_COUNTRY_CODE)${NC}"
        fi
    fi
}

# Get user configuration
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

# Get admin user configuration
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

# Create admin user via API
create_admin_user() {
    section "👤 Creating Admin User"

    echo -e "${CYAN}Registering admin user...${NC}"

    # Wait a bit more for API to be fully ready
    sleep 2

    REGISTER_RESPONSE=$(curl -s -X POST "http://localhost:8080/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$ADMIN_EMAIL\",\"username\":\"$ADMIN_USERNAME\",\"password\":\"$ADMIN_PASSWORD\"}" 2>/dev/null)

    # Check for errors
    if echo "$REGISTER_RESPONSE" | grep -q '"error"'; then
        ERROR_MSG=$(echo "$REGISTER_RESPONSE" | grep -o '"error":"[^"]*"' | cut -d'"' -f4)
        if [ "$ERROR_MSG" = "user already exists" ]; then
            echo -e "${YELLOW}⚠ User already exists. You can login with existing credentials.${NC}"
        else
            echo -e "${RED}✗ Failed to create admin user: $ERROR_MSG${NC}"
            echo -e "${YELLOW}You can create a user manually later:${NC}"
            echo -e "  curl -X POST http://$PUBLIC_IP:8080/api/v1/auth/register \\"
            echo -e "    -H 'Content-Type: application/json' \\"
            echo -e "    -d '{\"email\":\"your@email.com\",\"username\":\"admin\",\"password\":\"yourpass\"}'"
        fi
    else
        echo -e "${GREEN}✓ Admin user created successfully${NC}"
        ADMIN_CREATED=true
    fi
}

# Create .env file
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

# Configure firewall
configure_firewall() {
    section "🔥 Configuring Firewall"

    if command_exists ufw; then
        echo -e "${CYAN}Configuring UFW firewall...${NC}"
        ufw allow 80/tcp comment 'Aureo HTTP' 2>/dev/null || true
        ufw allow 8080/tcp comment 'Aureo API' 2>/dev/null || true
        ufw allow 4001/tcp comment 'Aureo P2P TCP' 2>/dev/null || true
        ufw allow 4001/udp comment 'Aureo P2P QUIC' 2>/dev/null || true
        ufw allow 51820/udp comment 'Aureo WireGuard' 2>/dev/null || true
        ufw allow 3000/tcp comment 'Grafana' 2>/dev/null || true
        ufw allow 9090/tcp comment 'Prometheus' 2>/dev/null || true
        echo -e "${GREEN}✓ Firewall rules added${NC}"
    elif command_exists firewall-cmd; then
        echo -e "${CYAN}Configuring firewalld...${NC}"
        firewall-cmd --permanent --add-port=80/tcp 2>/dev/null || true
        firewall-cmd --permanent --add-port=8080/tcp 2>/dev/null || true
        firewall-cmd --permanent --add-port=4001/tcp 2>/dev/null || true
        firewall-cmd --permanent --add-port=4001/udp 2>/dev/null || true
        firewall-cmd --permanent --add-port=51820/udp 2>/dev/null || true
        firewall-cmd --permanent --add-port=3000/tcp 2>/dev/null || true
        firewall-cmd --permanent --add-port=9090/tcp 2>/dev/null || true
        firewall-cmd --reload 2>/dev/null || true
        echo -e "${GREEN}✓ Firewall rules added${NC}"
    else
        echo -e "${YELLOW}⚠ No firewall detected. Make sure ports 80, 4001, 51820 are open.${NC}"
    fi
}

# Deploy services
deploy_services() {
    section "🚀 Deploying Services"

    cd "$DEPLOY_DIR"

    echo -e "${CYAN}Building and starting containers...${NC}"
    echo -e "${YELLOW}This may take a few minutes on first run...${NC}\n"

    # Stop any existing containers
    $DOCKER_COMPOSE down 2>/dev/null || true

    # Build and start
    $DOCKER_COMPOSE up -d --build

    echo -e "\n${GREEN}✓ Services deployed${NC}"
}

# Wait for services
wait_for_services() {
    section "⏳ Waiting for Services"

    echo -e "${CYAN}Waiting for services to be ready...${NC}"

    # Wait for API
    for i in {1..60}; do
        if curl -s --max-time 2 http://localhost:8080/health &>/dev/null; then
            echo -e "${GREEN}✓ API is ready${NC}"
            break
        fi
        echo -n "."
        sleep 2
    done

    # Wait for Dashboard
    for i in {1..30}; do
        if curl -s --max-time 2 http://localhost:5000 &>/dev/null; then
            echo -e "${GREEN}✓ Dashboard is ready${NC}"
            break
        fi
        echo -n "."
        sleep 2
    done

    echo ""
}

# Get node info
get_node_info() {
    NODE_INFO=$(curl -s http://localhost:8080/info 2>/dev/null || echo "{}")
    P2P_STATUS=$(curl -s http://localhost:8080/api/v1/p2p/status 2>/dev/null || echo "{}")

    NODE_ID=$(echo "$NODE_INFO" | grep -o '"id":"[^"]*"' | cut -d'"' -f4 || echo "")
    PEER_ID=$(echo "$P2P_STATUS" | grep -o '"peer_id":"[^"]*"' | cut -d'"' -f4 || echo "")
}

# Create management script
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
        ;;
esac
SCRIPT

    chmod +x /usr/local/bin/aureo
    echo "export AUREO_DEPLOY_DIR=\"$DEPLOY_DIR\"" >> /etc/profile.d/aureo.sh
    chmod +x /etc/profile.d/aureo.sh
}

# Save node credentials
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

    # Save admin password separately (more secure)
    echo "$ADMIN_PASSWORD" > "$CONFIG_DIR/.admin-password"
    chmod 600 "$CONFIG_DIR/.admin-password"
}

# Print summary
print_summary() {
    section "🎉 Setup Complete!"

    echo -e "${GREEN}Your Aureo VPN node is running!${NC}\n"

    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  NODE INFORMATION${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "  Node Name:     ${GREEN}$NODE_NAME${NC}"
    echo -e "  Node ID:       ${GREEN}$NODE_ID${NC}"
    echo -e "  Peer ID:       ${GREEN}$PEER_ID${NC}"
    echo -e "  Location:      ${GREEN}$CITY, $COUNTRY ($COUNTRY_CODE)${NC}"
    echo -e "  Public IP:     ${GREEN}$PUBLIC_IP${NC}"
    echo ""

    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  VPN LOGIN CREDENTIALS${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "  📧 Email:         ${GREEN}$ADMIN_EMAIL${NC}"
    echo -e "  👤 Username:      ${GREEN}$ADMIN_USERNAME${NC}"
    echo -e "  🔑 Password:      ${GREEN}$ADMIN_PASSWORD${NC}"
    echo ""

    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  WEB ACCESS${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "  🌐 Dashboard:     ${GREEN}http://$PUBLIC_IP/${NC}"
    echo -e "  🔌 API:           ${GREEN}http://$PUBLIC_IP:8080/api/v1/${NC}"
    echo -e "  📊 Grafana:       ${GREEN}http://$PUBLIC_IP:3000/${NC}  (admin / $GRAFANA_PASSWORD)"
    echo -e "  📈 Prometheus:    ${GREEN}http://$PUBLIC_IP:9090/${NC}"
    echo -e "  ❤️  Health:        ${GREEN}http://$PUBLIC_IP:8080/health${NC}"
    echo ""

    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  P2P MULTIADDR (share with other nodes)${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "  ${YELLOW}/ip4/$PUBLIC_IP/tcp/4001/p2p/$PEER_ID${NC}"
    echo ""

    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  MANAGEMENT COMMANDS${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "  ${GREEN}aureo status${NC}    - Show node status"
    echo -e "  ${GREEN}aureo logs${NC}      - View node logs"
    echo -e "  ${GREEN}aureo peers${NC}     - List connected P2P peers"
    echo -e "  ${GREEN}aureo nodes${NC}     - List known VPN nodes"
    echo -e "  ${GREEN}aureo restart${NC}   - Restart all services"
    echo -e "  ${GREEN}aureo stop${NC}      - Stop all services"
    echo -e "  ${GREEN}aureo ps${NC}        - Show container status"
    echo ""

    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  PORTS${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "  80      - HTTP (Nginx reverse proxy)"
    echo -e "  8080    - API Server"
    echo -e "  4001    - P2P (TCP + QUIC)"
    echo -e "  51820   - WireGuard VPN"
    echo -e "  3000    - Grafana"
    echo -e "  9090    - Prometheus"
    echo ""

    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  CONNECT TO VPN${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "  From macOS:"
    echo -e "    ${GREEN}API_URL=http://$PUBLIC_IP:8080 ./scripts/aureo-vpn-mac.sh login${NC}"
    echo -e "    ${GREEN}API_URL=http://$PUBLIC_IP:8080 ./scripts/aureo-vpn-mac.sh connect${NC}"
    echo ""
    echo -e "  From Linux:"
    echo -e "    ${GREEN}API_URL=http://$PUBLIC_IP:8080 ./scripts/aureo-vpn-linux.sh login${NC}"
    echo -e "    ${GREEN}API_URL=http://$PUBLIC_IP:8080 ./scripts/aureo-vpn-linux.sh connect${NC}"
    echo ""

    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}  🎊 Node is ready to accept connections! 🎊${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

##############################################################################
# MAIN EXECUTION
##############################################################################

main() {
    print_header
    check_root
    check_prerequisites
    detect_info
    get_config
    get_admin_config
    create_env_file
    configure_firewall
    deploy_services
    wait_for_services
    create_admin_user
    get_node_info
    create_management_script
    save_credentials
    print_summary

    echo -e "\n${GREEN}🚀 Setup completed successfully!${NC}\n"
}

# Run main
main "$@"
