#!/bin/bash

# Aureo VPN - Decentralized Node Operator Setup Script
# This script helps users easily set up a VPN node and start earning crypto rewards

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
API_ENDPOINT="${AUREO_API_ENDPOINT:-https://api.aureo-vpn.com}"
MIN_BANDWIDTH_MBPS=50
MIN_RAM_GB=2
MIN_DISK_GB=10

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
    ╔═══════════════════════════════════════════════════════════╗
    ║                                                           ║
    ║           🚀 Aureo VPN Node Operator Setup 🚀            ║
    ║                                                           ║
    ║        Earn Crypto Rewards by Running a VPN Node!        ║
    ║                                                           ║
    ╚═══════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}\n"
}

# Print section header
section() {
    echo -e "\n${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  $1${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Get user input
get_input() {
    local prompt="$1"
    local var_name="$2"
    local default="$3"

    if [ -n "$default" ]; then
        read -p "$(echo -e ${BLUE}${prompt}${NC} [${default}]: )" input
        eval $var_name="${input:-$default}"
    else
        read -p "$(echo -e ${BLUE}${prompt}${NC}: )" input
        eval $var_name="$input"
    fi
}

# Check system requirements
check_requirements() {
    section "📋 Step 1: Checking System Requirements"

    local all_good=true

    # Check OS
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo -e "${GREEN}✓${NC} Operating System: Linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        echo -e "${YELLOW}⚠${NC} Operating System: macOS (Limited support)"
    else
        echo -e "${RED}✗${NC} Unsupported operating system"
        all_good=false
    fi

    # Check RAM
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        total_ram=$(free -g | awk '/^Mem:/{print $2}')
    else
        total_ram=$(sysctl -n hw.memsize | awk '{print $1/1024/1024/1024}')
    fi

    if [ "$total_ram" -ge "$MIN_RAM_GB" ]; then
        echo -e "${GREEN}✓${NC} RAM: ${total_ram}GB (required: ${MIN_RAM_GB}GB)"
    else
        echo -e "${RED}✗${NC} RAM: ${total_ram}GB (required: ${MIN_RAM_GB}GB)"
        all_good=false
    fi

    # Check disk space
    disk_free=$(df -BG / | awk 'NR==2{print $4}' | sed 's/G//')
    if [ "$disk_free" -ge "$MIN_DISK_GB" ]; then
        echo -e "${GREEN}✓${NC} Disk Space: ${disk_free}GB free (required: ${MIN_DISK_GB}GB)"
    else
        echo -e "${RED}✗${NC} Disk Space: ${disk_free}GB free (required: ${MIN_DISK_GB}GB)"
        all_good=false
    fi

    # Check network bandwidth
    echo -e "${BLUE}ℹ${NC} Checking network speed..."
    if command_exists speedtest-cli; then
        download_speed=$(speedtest-cli --simple 2>/dev/null | grep Download | awk '{print $2}')
        if (( $(echo "$download_speed >= $MIN_BANDWIDTH_MBPS" | bc -l) )); then
            echo -e "${GREEN}✓${NC} Download Speed: ${download_speed} Mbps (required: ${MIN_BANDWIDTH_MBPS} Mbps)"
        else
            echo -e "${YELLOW}⚠${NC} Download Speed: ${download_speed} Mbps (recommended: ${MIN_BANDWIDTH_MBPS}+ Mbps)"
        fi
    else
        echo -e "${YELLOW}⚠${NC} speedtest-cli not installed, skipping speed test"
        echo -e "${BLUE}ℹ${NC} Please ensure you have at least ${MIN_BANDWIDTH_MBPS} Mbps bandwidth"
    fi

    if [ "$all_good" = false ]; then
        echo -e "\n${RED}✗ System does not meet minimum requirements${NC}"
        exit 1
    fi

    echo -e "\n${GREEN}✓ All system requirements met!${NC}"
}

# Install dependencies
install_dependencies() {
    section "📦 Step 2: Installing Dependencies"

    # Detect package manager
    if command_exists apt-get; then
        PM="apt-get"
        PM_UPDATE="sudo apt-get update"
        PM_INSTALL="sudo apt-get install -y"
    elif command_exists yum; then
        PM="yum"
        PM_UPDATE="sudo yum check-update"
        PM_INSTALL="sudo yum install -y"
    elif command_exists brew; then
        PM="brew"
        PM_UPDATE="brew update"
        PM_INSTALL="brew install"
    else
        echo -e "${RED}✗ No supported package manager found${NC}"
        exit 1
    fi

    echo "Updating package lists..."
    $PM_UPDATE || true

    # Install required packages
    PACKAGES="curl wget git wireguard wireguard-tools iptables jq"

    echo "Installing packages: $PACKAGES"
    for package in $PACKAGES; do
        if ! command_exists $package; then
            echo "Installing $package..."
            $PM_INSTALL $package || echo -e "${YELLOW}⚠ Could not install $package${NC}"
        else
            echo -e "${GREEN}✓${NC} $package already installed"
        fi
    done

    # Check Docker
    if ! command_exists docker; then
        echo -e "\n${YELLOW}Docker not installed. Do you want to install Docker? (recommended)${NC}"
        read -p "Install Docker? (y/n): " install_docker
        if [[ $install_docker =~ ^[Yy]$ ]]; then
            echo "Installing Docker..."
            curl -fsSL https://get.docker.com | sh
            sudo usermod -aG docker $USER
            echo -e "${GREEN}✓ Docker installed${NC}"
            echo -e "${YELLOW}⚠ Please log out and back in for Docker permissions to take effect${NC}"
        fi
    else
        echo -e "${GREEN}✓ Docker already installed${NC}"
    fi

    echo -e "\n${GREEN}✓ Dependencies installed successfully!${NC}"
}

# Create operator account
create_operator_account() {
    section "👤 Step 3: Creating Operator Account"

    echo "Let's create your node operator account!"
    echo ""

    # Get user details
    get_input "Email address" EMAIL
    get_input "Username" USERNAME
    get_input "Password (min 8 characters)" PASSWORD
    get_input "Full Name" FULL_NAME ""

    # Validate email
    if [[ ! "$EMAIL" =~ ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
        echo -e "${RED}✗ Invalid email address${NC}"
        exit 1
    fi

    # Register user
    echo -e "\n${BLUE}Creating account...${NC}"

    REGISTER_RESPONSE=$(curl -s -X POST "${API_ENDPOINT}/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$EMAIL\",
            \"password\": \"$PASSWORD\",
            \"username\": \"$USERNAME\",
            \"full_name\": \"$FULL_NAME\"
        }")

    ACCESS_TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.access_token')

    if [ "$ACCESS_TOKEN" != "null" ] && [ -n "$ACCESS_TOKEN" ]; then
        echo -e "${GREEN}✓ Account created successfully!${NC}"
        echo "Access Token: ${ACCESS_TOKEN:0:20}..."

        # Save credentials
        mkdir -p ~/.aureo-vpn
        cat > ~/.aureo-vpn/credentials << EOF
EMAIL=$EMAIL
USERNAME=$USERNAME
ACCESS_TOKEN=$ACCESS_TOKEN
EOF
        chmod 600 ~/.aureo-vpn/credentials
    else
        echo -e "${RED}✗ Failed to create account${NC}"
        echo "Response: $REGISTER_RESPONSE"
        exit 1
    fi
}

# Setup crypto wallet
setup_wallet() {
    section "💰 Step 4: Crypto Wallet Setup"

    echo "Choose your preferred cryptocurrency for receiving rewards:"
    echo ""
    echo "  1) Ethereum (ETH)"
    echo "  2) Bitcoin (BTC)"
    echo "  3) Litecoin (LTC)"
    echo ""

    read -p "Enter choice [1-3]: " crypto_choice

    case $crypto_choice in
        1)
            CRYPTO_TYPE="ethereum"
            ;;
        2)
            CRYPTO_TYPE="bitcoin"
            ;;
        3)
            CRYPTO_TYPE="litecoin"
            ;;
        *)
            echo -e "${RED}Invalid choice${NC}"
            exit 1
            ;;
    esac

    echo ""
    get_input "Enter your ${CRYPTO_TYPE} wallet address" WALLET_ADDRESS

    # Validate wallet address format (basic validation)
    if [ ${#WALLET_ADDRESS} -lt 26 ]; then
        echo -e "${RED}✗ Invalid wallet address (too short)${NC}"
        exit 1
    fi

    # Register as operator
    echo -e "\n${BLUE}Registering as node operator...${NC}"

    OPERATOR_RESPONSE=$(curl -s -X POST "${API_ENDPOINT}/api/v1/operator/register" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -d "{
            \"wallet_address\": \"$WALLET_ADDRESS\",
            \"wallet_type\": \"$CRYPTO_TYPE\",
            \"country\": \"$(curl -s https://ipapi.co/country_name/)\",
            \"email\": \"$EMAIL\"
        }")

    OPERATOR_ID=$(echo $OPERATOR_RESPONSE | jq -r '.operator_id')

    if [ "$OPERATOR_ID" != "null" ] && [ -n "$OPERATOR_ID" ]; then
        echo -e "${GREEN}✓ Operator account created!${NC}"
        echo "Operator ID: $OPERATOR_ID"

        # Save operator info
        cat >> ~/.aureo-vpn/credentials << EOF
OPERATOR_ID=$OPERATOR_ID
WALLET_ADDRESS=$WALLET_ADDRESS
CRYPTO_TYPE=$CRYPTO_TYPE
EOF
    else
        echo -e "${RED}✗ Failed to register as operator${NC}"
        echo "Response: $OPERATOR_RESPONSE"
        exit 1
    fi
}

# Setup VPN node
setup_node() {
    section "🖥️  Step 5: VPN Node Setup"

    echo "Let's configure your VPN node!"
    echo ""

    # Get node details
    get_input "Node name (e.g., My-VPN-Node-US)" NODE_NAME
    get_input "Public IP address" PUBLIC_IP "$(curl -s https://ipapi.co/ip/)"
    get_input "WireGuard port" WIREGUARD_PORT "51820"
    get_input "OpenVPN port" OPENVPN_PORT "1194"

    # Detect location
    COUNTRY=$(curl -s https://ipapi.co/country_name/)
    COUNTRY_CODE=$(curl -s https://ipapi.co/country/)
    CITY=$(curl -s https://ipapi.co/city/)

    echo -e "\nDetected location: ${CITY}, ${COUNTRY}"

    read -p "Is this correct? (y/n): " location_correct
    if [[ ! $location_correct =~ ^[Yy]$ ]]; then
        get_input "Country" COUNTRY
        get_input "Country Code (2 letters)" COUNTRY_CODE
        get_input "City" CITY
    fi

    # Create node
    echo -e "\n${BLUE}Registering VPN node...${NC}"

    NODE_RESPONSE=$(curl -s -X POST "${API_ENDPOINT}/api/v1/operator/nodes" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -d "{
            \"name\": \"$NODE_NAME\",
            \"hostname\": \"$NODE_NAME.aureo-vpn.network\",
            \"public_ip\": \"$PUBLIC_IP\",
            \"country\": \"$COUNTRY\",
            \"country_code\": \"$COUNTRY_CODE\",
            \"city\": \"$CITY\",
            \"wireguard_port\": $WIREGUARD_PORT,
            \"openvpn_port\": $OPENVPN_PORT,
            \"is_operator_owned\": true
        }")

    NODE_ID=$(echo $NODE_RESPONSE | jq -r '.node_id')
    PUBLIC_KEY=$(echo $NODE_RESPONSE | jq -r '.public_key')

    if [ "$NODE_ID" != "null" ] && [ -n "$NODE_ID" ]; then
        echo -e "${GREEN}✓ VPN node registered!${NC}"
        echo "Node ID: $NODE_ID"
        echo "Public Key: $PUBLIC_KEY"

        # Save node info
        cat >> ~/.aureo-vpn/credentials << EOF
NODE_ID=$NODE_ID
NODE_NAME=$NODE_NAME
PUBLIC_IP=$PUBLIC_IP
PUBLIC_KEY=$PUBLIC_KEY
WIREGUARD_PORT=$WIREGUARD_PORT
OPENVPN_PORT=$OPENVPN_PORT
EOF
    else
        echo -e "${RED}✗ Failed to register node${NC}"
        echo "Response: $NODE_RESPONSE"
        exit 1
    fi
}

# Setup firewall
setup_firewall() {
    section "🔥 Step 6: Configuring Firewall"

    echo "Configuring firewall rules..."

    # Allow VPN ports
    if command_exists ufw; then
        sudo ufw allow $WIREGUARD_PORT/udp comment "Aureo VPN WireGuard"
        sudo ufw allow $OPENVPN_PORT/udp comment "Aureo VPN OpenVPN"
        echo -e "${GREEN}✓ UFW firewall configured${NC}"
    elif command_exists firewall-cmd; then
        sudo firewall-cmd --permanent --add-port=$WIREGUARD_PORT/udp
        sudo firewall-cmd --permanent --add-port=$OPENVPN_PORT/udp
        sudo firewall-cmd --reload
        echo -e "${GREEN}✓ firewalld configured${NC}"
    else
        echo -e "${YELLOW}⚠ No firewall detected. Please manually allow ports:${NC}"
        echo "  - WireGuard: $WIREGUARD_PORT/udp"
        echo "  - OpenVPN: $OPENVPN_PORT/udp"
    fi
}

# Install node software
install_node_software() {
    section "⚙️  Step 7: Installing Node Software"

    # Create installation directory
    sudo mkdir -p /opt/aureo-vpn
    cd /opt/aureo-vpn

    # Download node software
    echo "Downloading Aureo VPN node software..."
    LATEST_VERSION=$(curl -s https://api.github.com/repos/nikola43/aureo-vpn/releases/latest | jq -r '.tag_name')

    sudo wget -O vpn-node "https://github.com/nikola43/aureo-vpn/releases/download/${LATEST_VERSION}/vpn-node-linux-amd64"
    sudo chmod +x vpn-node

    # Create configuration
    sudo tee /opt/aureo-vpn/config.env > /dev/null << EOF
# Aureo VPN Node Configuration
NODE_ID=$NODE_ID
API_ENDPOINT=$API_ENDPOINT
ACCESS_TOKEN=$ACCESS_TOKEN

# Database connection (connects to central server)
DB_HOST=db.aureo-vpn.com
DB_PORT=5432
DB_USER=node_operator
DB_PASSWORD=$(openssl rand -base64 32)
DB_NAME=aureo_vpn

# Node settings
WIREGUARD_PORT=$WIREGUARD_PORT
OPENVPN_PORT=$OPENVPN_PORT
LOG_LEVEL=info
EOF

    # Create systemd service
    sudo tee /etc/systemd/system/aureo-vpn-node.service > /dev/null << EOF
[Unit]
Description=Aureo VPN Node
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/aureo-vpn
EnvironmentFile=/opt/aureo-vpn/config.env
ExecStart=/opt/aureo-vpn/vpn-node
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

    # Enable and start service
    sudo systemctl daemon-reload
    sudo systemctl enable aureo-vpn-node
    sudo systemctl start aureo-vpn-node

    echo -e "${GREEN}✓ Node software installed and started!${NC}"

    # Check status
    sleep 2
    if sudo systemctl is-active --quiet aureo-vpn-node; then
        echo -e "${GREEN}✓ Node is running${NC}"
    else
        echo -e "${YELLOW}⚠ Node may not be running. Check logs with: sudo journalctl -u aureo-vpn-node -f${NC}"
    fi
}

# Display earnings info
show_earnings_info() {
    section "💎 Step 8: Earnings Information"

    # Get current reward tier
    TIER_INFO=$(curl -s -X GET "${API_ENDPOINT}/api/v1/operator/rewards/tiers" \
        -H "Authorization: Bearer $ACCESS_TOKEN")

    echo "📊 Current Reward Tiers:"
    echo ""
    echo "$TIER_INFO" | jq -r '.tiers[] | "  \(.tier_name | ascii_upcase): $\(.base_rate_per_gb) per GB (Min uptime: \(.min_uptime_percent)%)"'

    echo ""
    echo "💰 Earnings Calculator:"
    echo "  - 100 GB/day × 30 days × \$0.01/GB = \$30/month (Bronze tier)"
    echo "  - 500 GB/day × 30 days × \$0.015/GB = \$225/month (Silver tier)"
    echo "  - 1000 GB/day × 30 days × \$0.02/GB = \$600/month (Gold tier)"
    echo ""
    echo "🎯 Tips to Maximize Earnings:"
    echo "  - Maintain 95%+ uptime"
    echo "  - Ensure fast, stable connection"
    echo "  - Keep latency < 50ms"
    echo "  - Provide excellent service quality"
    echo ""
    echo "📈 Track your earnings:"
    echo "  Dashboard: ${API_ENDPOINT}/operator/dashboard"
    echo "  Or use: aureo-vpn operator stats"
}

# Final summary
print_summary() {
    section "🎉 Setup Complete!"

    cat << EOF
${GREEN}✓ Your Aureo VPN node is now running!${NC}

${CYAN}Node Information:${NC}
  Node ID:      $NODE_ID
  Node Name:    $NODE_NAME
  Public IP:    $PUBLIC_IP
  Status:       Running

${CYAN}Earnings:${NC}
  Wallet:       $WALLET_ADDRESS ($CRYPTO_TYPE)
  Operator ID:  $OPERATOR_ID

${CYAN}Useful Commands:${NC}
  Check status:     sudo systemctl status aureo-vpn-node
  View logs:        sudo journalctl -u aureo-vpn-node -f
  Restart node:     sudo systemctl restart aureo-vpn-node
  View earnings:    curl -H "Authorization: Bearer $ACCESS_TOKEN" ${API_ENDPOINT}/api/v1/operator/stats

${CYAN}Dashboard:${NC}
  Web:    ${API_ENDPOINT}/operator/dashboard
  Mobile: Download Aureo VPN Operator app

${CYAN}Next Steps:${NC}
  1. Monitor your node's performance
  2. Maintain 95%+ uptime for maximum earnings
  3. Join our operator community: https://community.aureo-vpn.com
  4. Set up monitoring alerts

${YELLOW}Important:${NC}
  - Keep your node online 24/7 for best earnings
  - Minimum payout threshold: \$10
  - Payouts are processed weekly
  - Contact support: operator-support@aureo-vpn.com

${GREEN}Thank you for joining the Aureo VPN network!${NC}
${GREEN}Start earning crypto rewards by providing VPN services!${NC}

EOF
}

# Main execution
main() {
    print_header

    echo -e "${CYAN}This script will help you:${NC}"
    echo "  ✓ Check system requirements"
    echo "  ✓ Install necessary dependencies"
    echo "  ✓ Create your operator account"
    echo "  ✓ Setup your crypto wallet"
    echo "  ✓ Configure and start your VPN node"
    echo "  ✓ Start earning crypto rewards!"
    echo ""

    read -p "Ready to get started? (y/n): " start_setup
    if [[ ! $start_setup =~ ^[Yy]$ ]]; then
        echo "Setup cancelled."
        exit 0
    fi

    check_requirements
    install_dependencies
    create_operator_account
    setup_wallet
    setup_node
    setup_firewall
    install_node_software
    show_earnings_info
    print_summary

    echo -e "\n${GREEN}🚀 Setup completed successfully!${NC}\n"
}

# Run main function
main "$@"
