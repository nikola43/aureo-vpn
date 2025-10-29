#!/bin/bash

################################################################################
# Aureo VPN - Node Operator Setup Script
#
# One-command setup to become an Aureo VPN node operator
#
# What it does:
#   - Deploys all services via Docker Compose
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

# Check prerequisites
check_prerequisites() {
    section "🔍 Checking Prerequisites"

    # Check Docker
    if ! command_exists docker; then
        echo -e "${RED}✗ Docker is not installed${NC}"
        echo -e "${YELLOW}Please install Docker first: https://docs.docker.com/get-docker/${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Docker installed${NC}"

    # Check Docker Compose
    if ! docker compose version >/dev/null 2>&1 && ! command_exists docker-compose; then
        echo -e "${RED}✗ Docker Compose is not installed${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Docker Compose installed${NC}"

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

# Deploy services
deploy_services() {
    section "🐳 Deploying Services"

    cd "$PROJECT_ROOT"

    echo -e "${YELLOW}Building and starting containers...${NC}"

    # Stop any existing containers
    docker compose -f "$DOCKER_COMPOSE_FILE" down 2>/dev/null || true

    # Build and start services
    docker compose -f "$DOCKER_COMPOSE_FILE" up -d --build

    echo -e "${YELLOW}Waiting for services to be ready...${NC}"
    sleep 10

    # Check services are running
    if ! docker compose -f "$DOCKER_COMPOSE_FILE" ps | grep -q "Up"; then
        echo -e "${RED}✗ Services failed to start${NC}"
        echo -e "${YELLOW}Check logs with: docker compose -f $DOCKER_COMPOSE_FILE logs${NC}"
        exit 1
    fi

    echo -e "${GREEN}✓ All services deployed successfully${NC}"
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
        echo -e "${RED}✗ Failed to register as operator${NC}"
        echo "Response: $OPERATOR_RESPONSE"
        exit 1
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
    docker exec aureo-vpn-db psql -U postgres -d aureo_vpn -c \
        "UPDATE node_operators SET status='active', is_verified=true, verified_at=NOW() WHERE wallet_address='$WALLET_ADDRESS';" \
        >/dev/null 2>&1

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

# Register VPN node
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

    # Get NODE_ID from environment or generate
    if [ -z "$NODE_ID_1" ]; then
        # Generate a UUID for the node
        NODE_ID_1=$(cat /proc/sys/kernel/random/uuid)

        # Update docker-compose environment
        echo "NODE_ID_1=$NODE_ID_1" >> "$PROJECT_ROOT/deployments/docker/.env"

        # Restart vpn-node container with new NODE_ID
        docker compose -f "$DOCKER_COMPOSE_FILE" restart vpn-node-1
        sleep 5
    fi

    # Register node with API
    echo -e "\n${BLUE}Registering VPN node with network...${NC}"

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
        echo -e "${GREEN}✓ VPN node registered successfully${NC}"
        echo -e "${BLUE}  Node ID: $NODE_ID${NC}"

        # Set node status to online
        echo -e "${CYAN}Activating node...${NC}"
        docker exec aureo-vpn-db psql -U postgres -d aureo_vpn -c \
            "UPDATE vpn_nodes SET status='online', last_heartbeat=NOW() WHERE id='$NODE_ID';" \
            >/dev/null 2>&1

        # Get WireGuard server public key and update node
        echo -e "${CYAN}Configuring WireGuard...${NC}"
        WG_PUBLIC_KEY=$(docker exec aureo-vpn-node-1 wg show wg0 public-key 2>/dev/null || echo "")
        if [ -n "$WG_PUBLIC_KEY" ]; then
            docker exec aureo-vpn-db psql -U postgres -d aureo_vpn -c \
                "UPDATE vpn_nodes SET public_key='$WG_PUBLIC_KEY' WHERE id='$NODE_ID';" \
                >/dev/null 2>&1
            echo -e "${GREEN}✓ WireGuard public key configured${NC}"
        fi

        # Recalculate operator stats
        docker exec aureo-vpn-db psql -U postgres -d aureo_vpn -c \
            "UPDATE node_operators SET active_nodes_count = (SELECT COUNT(*) FROM vpn_nodes WHERE operator_id = node_operators.id AND status = 'online' AND is_active = true);" \
            >/dev/null 2>&1

        echo -e "${GREEN}✓ Node activated and online${NC}"

        # Save node info
        cat >> "$CONFIG_DIR/operator-credentials" << EOF
NODE_ID=$NODE_ID
NODE_NAME=$NODE_NAME
PUBLIC_IP=$PUBLIC_IP
INTERNAL_IP=$INTERNAL_IP
EOF
    else
        echo -e "${RED}✗ Failed to register node${NC}"
        echo "Response: $NODE_RESPONSE"
        exit 1
    fi
}

# Setup monitoring script
setup_monitoring() {
    section "📊 Setting Up Monitoring"

    # Create keep-node-online script
    cat > /opt/aureo-vpn/keep-node-online.sh << 'EOF'
#!/bin/bash
# Keep node status as online
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
        { cat; echo "* * * * * /opt/aureo-vpn/keep-node-online.sh"; } | crontab -

    echo -e "${GREEN}✓ Monitoring configured${NC}"
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
    echo -e "  ✓ API Gateway:          Running on http://localhost:8080"
    echo -e "  ✓ VPN Node:             Running on $PUBLIC_IP:51820"
    echo -e "  ✓ Web Dashboard:        Running on http://localhost:3001"
    echo -e "  ✓ Node Status:          Registered and Active"
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
    echo -e "  View all containers:    ${GREEN}docker compose -f $DOCKER_COMPOSE_FILE ps${NC}"
    echo -e "  View logs (all):        ${GREEN}docker compose -f $DOCKER_COMPOSE_FILE logs -f${NC}"
    echo -e "  View VPN node logs:     ${GREEN}docker compose -f $DOCKER_COMPOSE_FILE logs -f vpn-node-1${NC}"
    echo -e "  Check WireGuard:        ${GREEN}docker exec aureo-vpn-node-1 wg show wg0${NC}"
    echo -e "  Restart services:       ${GREEN}./scripts/deploy.sh restart${NC}"
    echo -e "  Rebuild after changes:  ${GREEN}./scripts/deploy.sh rebuild${NC}"
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
    check_prerequisites

    echo ""
    echo -e "${YELLOW}This script will:${NC}"
    echo "  ✓ Deploy all services using Docker Compose"
    echo "  ✓ Create your operator account"
    echo "  ✓ Register and activate your VPN node"
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

    deploy_services
    setup_peer_script
    create_operator_account
    register_node
    setup_monitoring
    print_summary

    echo -e "\n${GREEN}🚀 Node Operator Setup Completed Successfully!${NC}\n"
}

# Run main
main "$@"
