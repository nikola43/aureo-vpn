#!/bin/bash

################################################################################
# Aureo VPN Node Setup Script
#
# This script helps operators set up a VPN node quickly and easily.
# It will:
#   1. Install required dependencies (WireGuard, OpenVPN)
#   2. Configure the node
#   3. Register the node with the Aureo VPN network
#   4. Start VPN services
#
# Usage:
#   sudo ./setup-node.sh
################################################################################

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Banner
echo -e "${PURPLE}"
cat << "EOF"
   _                         __     ______  _   _
  / \  _   _ _ __ ___  ___   \ \   / /  _ \| \ | |
 / _ \| | | | '__/ _ \/ _ \   \ \ / /| |_) |  \| |
/ ___ \ |_| | | |  __/ (_) |   \ V / |  __/| |\  |
/_/   \_\__,_|_|  \___|\___/     \_/  |_|   |_| \_|

        Node Setup Script - Start Earning Crypto!
EOF
echo -e "${NC}"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Error: Please run as root (use sudo)${NC}"
    exit 1
fi

# Detect OS
echo -e "${CYAN}🔍 Detecting operating system...${NC}"
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
    OS_VERSION=$VERSION_ID
    echo -e "${GREEN}✓ Detected: $PRETTY_NAME${NC}"
else
    echo -e "${RED}Error: Cannot detect OS${NC}"
    exit 1
fi

# Function to install packages on Ubuntu/Debian
install_ubuntu() {
    echo -e "${CYAN}📦 Installing dependencies on Ubuntu/Debian...${NC}"
    apt-get update -qq
    apt-get install -y -qq \
        wireguard \
        wireguard-tools \
        openvpn \
        easy-rsa \
        iptables \
        curl \
        jq \
        net-tools \
        resolvconf
    echo -e "${GREEN}✓ Dependencies installed${NC}"
}

# Function to install packages on CentOS/RHEL
install_centos() {
    echo -e "${CYAN}📦 Installing dependencies on CentOS/RHEL...${NC}"
    yum install -y epel-release
    yum install -y \
        wireguard-tools \
        openvpn \
        easy-rsa \
        iptables \
        curl \
        jq \
        net-tools
    echo -e "${GREEN}✓ Dependencies installed${NC}"
}

# Install based on OS
case "$OS" in
    ubuntu|debian)
        install_ubuntu
        ;;
    centos|rhel|fedora)
        install_centos
        ;;
    *)
        echo -e "${RED}Error: Unsupported OS: $OS${NC}"
        exit 1
        ;;
esac

# Get node information
echo ""
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}           Node Configuration${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Auto-detect public IP
PUBLIC_IP=$(curl -s https://api.ipify.org)
echo -e "${CYAN}Auto-detected public IP: ${GREEN}$PUBLIC_IP${NC}"

# Auto-detect location
echo -e "${CYAN}🌍 Detecting location...${NC}"
LOCATION=$(curl -s "https://ipapi.co/$PUBLIC_IP/json/")
AUTO_COUNTRY=$(echo "$LOCATION" | jq -r '.country_name')
AUTO_COUNTRY_CODE=$(echo "$LOCATION" | jq -r '.country_code')
AUTO_CITY=$(echo "$LOCATION" | jq -r '.city')
AUTO_LAT=$(echo "$LOCATION" | jq -r '.latitude')
AUTO_LON=$(echo "$LOCATION" | jq -r '.longitude')

echo -e "${GREEN}✓ Location: $AUTO_CITY, $AUTO_COUNTRY${NC}"

# Get hostname
HOSTNAME=$(hostname)
echo -e "${CYAN}Auto-detected hostname: ${GREEN}$HOSTNAME${NC}"

# Ask for node name
read -p "$(echo -e ${CYAN}Enter node name [default: aureo-node-$HOSTNAME]: ${NC})" NODE_NAME
NODE_NAME=${NODE_NAME:-"aureo-node-$HOSTNAME"}

# Ask for API URL
read -p "$(echo -e ${CYAN}Enter API URL [default: http://localhost:8080/api/v1]: ${NC})" API_URL
API_URL=${API_URL:-"http://localhost:8080/api/v1"}

# Ask for JWT token
echo ""
echo -e "${YELLOW}You need your JWT access token from the dashboard.${NC}"
echo -e "${YELLOW}You can find it in your browser's localStorage (key: access_token)${NC}"
echo ""
read -sp "$(echo -e ${CYAN}Enter your JWT token: ${NC})" JWT_TOKEN
echo ""

if [ -z "$JWT_TOKEN" ]; then
    echo -e "${RED}Error: JWT token is required${NC}"
    exit 1
fi

# WireGuard port
read -p "$(echo -e ${CYAN}WireGuard port [default: 51820]: ${NC})" WG_PORT
WG_PORT=${WG_PORT:-51820}

# OpenVPN port
read -p "$(echo -e ${CYAN}OpenVPN port [default: 1194]: ${NC})" OVPN_PORT
OVPN_PORT=${OVPN_PORT:-1194}

# Configure WireGuard
echo ""
echo -e "${CYAN}🔧 Configuring WireGuard...${NC}"

# Generate WireGuard keys
WG_PRIVATE_KEY=$(wg genkey)
WG_PUBLIC_KEY=$(echo "$WG_PRIVATE_KEY" | wg pubkey)

# Create WireGuard config directory
mkdir -p /etc/wireguard

# Create WireGuard config
cat > /etc/wireguard/wg0.conf << EOF
[Interface]
PrivateKey = $WG_PRIVATE_KEY
Address = 10.8.0.1/24
ListenPort = $WG_PORT
PostUp = iptables -A FORWARD -i wg0 -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
PostDown = iptables -D FORWARD -i wg0 -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE
EOF

chmod 600 /etc/wireguard/wg0.conf

echo -e "${GREEN}✓ WireGuard configured${NC}"

# Enable IP forwarding
echo -e "${CYAN}🔧 Enabling IP forwarding...${NC}"
echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf
echo "net.ipv6.conf.all.forwarding=1" >> /etc/sysctl.conf
sysctl -p > /dev/null 2>&1
echo -e "${GREEN}✓ IP forwarding enabled${NC}"

# Configure firewall
echo -e "${CYAN}🔧 Configuring firewall...${NC}"
iptables -A INPUT -p udp --dport $WG_PORT -j ACCEPT
iptables -A INPUT -p udp --dport $OVPN_PORT -j ACCEPT
iptables -A INPUT -p tcp --dport $OVPN_PORT -j ACCEPT
echo -e "${GREEN}✓ Firewall configured${NC}"

# Register node with API
echo ""
echo -e "${CYAN}📡 Registering node with Aureo VPN network...${NC}"

REGISTER_RESPONSE=$(curl -s -X POST "$API_URL/operator/nodes" \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"name\": \"$NODE_NAME\",
        \"hostname\": \"$HOSTNAME\",
        \"public_ip\": \"$PUBLIC_IP\",
        \"country\": \"$AUTO_COUNTRY\",
        \"country_code\": \"$AUTO_COUNTRY_CODE\",
        \"city\": \"$AUTO_CITY\",
        \"wireguard_port\": $WG_PORT,
        \"openvpn_port\": $OVPN_PORT,
        \"latitude\": $AUTO_LAT,
        \"longitude\": $AUTO_LON
    }")

# Check if registration was successful
if echo "$REGISTER_RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
    ERROR_MSG=$(echo "$REGISTER_RESPONSE" | jq -r '.error')
    echo -e "${RED}✗ Registration failed: $ERROR_MSG${NC}"
    exit 1
fi

NODE_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.node.id')

if [ -z "$NODE_ID" ] || [ "$NODE_ID" = "null" ]; then
    echo -e "${RED}✗ Registration failed${NC}"
    echo "$REGISTER_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✓ Node registered successfully!${NC}"
echo -e "${GREEN}  Node ID: $NODE_ID${NC}"

# Start WireGuard
echo ""
echo -e "${CYAN}🚀 Starting VPN services...${NC}"
systemctl enable wg-quick@wg0
systemctl start wg-quick@wg0

if systemctl is-active --quiet wg-quick@wg0; then
    echo -e "${GREEN}✓ WireGuard started${NC}"
else
    echo -e "${RED}✗ Failed to start WireGuard${NC}"
fi

# Display summary
echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}           🎉 Node Setup Complete!${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${CYAN}Node Details:${NC}"
echo -e "  Name:       ${GREEN}$NODE_NAME${NC}"
echo -e "  Node ID:    ${GREEN}$NODE_ID${NC}"
echo -e "  Public IP:  ${GREEN}$PUBLIC_IP${NC}"
echo -e "  Location:   ${GREEN}$AUTO_CITY, $AUTO_COUNTRY${NC}"
echo -e "  WireGuard:  ${GREEN}$PUBLIC_IP:$WG_PORT${NC}"
echo -e "  OpenVPN:    ${GREEN}$PUBLIC_IP:$OVPN_PORT${NC}"
echo ""
echo -e "${YELLOW}💰 Your node is now online and earning crypto rewards!${NC}"
echo ""
echo -e "${CYAN}View your earnings at: ${BLUE}$API_URL/../dashboard${NC}"
echo ""
echo -e "${CYAN}Useful Commands:${NC}"
echo -e "  Check status:    ${GREEN}sudo systemctl status wg-quick@wg0${NC}"
echo -e "  View logs:       ${GREEN}sudo journalctl -u wg-quick@wg0 -f${NC}"
echo -e "  Restart node:    ${GREEN}sudo systemctl restart wg-quick@wg0${NC}"
echo -e "  Stop node:       ${GREEN}sudo systemctl stop wg-quick@wg0${NC}"
echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# Save config
mkdir -p ~/.aureo-vpn
cat > ~/.aureo-vpn/node-config.json << EOF
{
    "node_id": "$NODE_ID",
    "node_name": "$NODE_NAME",
    "public_ip": "$PUBLIC_IP",
    "api_url": "$API_URL",
    "wireguard_port": $WG_PORT,
    "openvpn_port": $OVPN_PORT
}
EOF

echo -e "${CYAN}Configuration saved to: ${GREEN}~/.aureo-vpn/node-config.json${NC}"
echo ""
