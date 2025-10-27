#!/bin/bash

################################################################################
# Aureo VPN Client
#
# Simple VPN client for connecting to Aureo VPN nodes
#
# Commands:
#   login      - Login with email and password
#   logout     - Logout and clear session
#   list       - List available VPN nodes
#   connect    - Connect to a VPN node
#   disconnect - Disconnect from VPN
#   status     - Show connection status
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
API_URL="${API_URL:-http://155.138.238.145:8080/api/v1}"
CONFIG_DIR="/tmp/.aureo-vpn-$USER"
SESSION_FILE="$CONFIG_DIR/.session"
CONNECTION_FILE="$CONFIG_DIR/.connection"
WG_CONFIG="$CONFIG_DIR/wg0.conf"
WG_INTERFACE="wg-aureo"

# Create config directory
mkdir -p "$CONFIG_DIR"

# Banner
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

# Login
cmd_login() {
    banner
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  Login${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    read -p "$(echo -e ${CYAN}Email: ${NC})" EMAIL
    read -sp "$(echo -e ${CYAN}Password: ${NC})" PASSWORD
    echo ""
    echo ""

    if [ -z "$EMAIL" ] || [ -z "$PASSWORD" ]; then
        echo -e "${RED}✗ Email and password required${NC}"
        exit 1
    fi

    echo -e "${CYAN}Logging in...${NC}"

    RESPONSE=$(curl -s -X POST "$API_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

    if echo "$RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
        echo -e "${RED}✗ Login failed: $(echo "$RESPONSE" | jq -r '.error')${NC}"
        exit 1
    fi

    TOKEN=$(echo "$RESPONSE" | jq -r '.access_token // empty')
    if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
        echo -e "${RED}✗ Login failed${NC}"
        exit 1
    fi

    echo "TOKEN=$TOKEN" > "$SESSION_FILE"
    echo "EMAIL=$EMAIL" >> "$SESSION_FILE"
    chmod 600 "$SESSION_FILE"

    echo -e "${GREEN}✓ Logged in as $EMAIL${NC}"
    echo ""
}

# Logout
cmd_logout() {
    banner
    if [ -f "$SESSION_FILE" ]; then
        rm -f "$SESSION_FILE"
        echo -e "${GREEN}✓ Logged out${NC}"
    else
        echo -e "${YELLOW}No active session${NC}"
    fi
}

# Load session
load_session() {
    if [ ! -f "$SESSION_FILE" ]; then
        echo -e "${YELLOW}Not logged in. Please run: $0 login${NC}"
        exit 1
    fi
    source "$SESSION_FILE"
}

# List nodes
cmd_list() {
    banner
    load_session

    echo -e "${CYAN}Available VPN Nodes:${NC}"
    echo ""

    RESPONSE=$(curl -s -X GET "$API_URL/operator/nodes" \
        -H "Authorization: Bearer $TOKEN")

    if echo "$RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
        ERROR_MSG=$(echo "$RESPONSE" | jq -r '.error')
        if [[ "$ERROR_MSG" == *"expired"* ]] || [[ "$ERROR_MSG" == *"invalid"* ]]; then
            echo -e "${RED}✗ Session expired. Please login again: $0 login${NC}"
        else
            echo -e "${RED}✗ Failed to fetch nodes: $ERROR_MSG${NC}"
        fi
        exit 1
    fi

    echo "$RESPONSE" | jq -r '.nodes // [] | .[] |
        "\u001b[32m[\(.id | .[0:8])]\u001b[0m \u001b[1;37m\(.name)\u001b[0m
  📍 \(.city), \(.country)
  🌐 \(.public_ip)
  ⚡ Status: \(.status)
  📊 Uptime: \(.uptime_percentage)%
"' 2>/dev/null || echo -e "${YELLOW}No nodes available${NC}"
}

# Connect
cmd_connect() {
    banner
    load_session

    # Check if already connected
    if [ -f "$CONNECTION_FILE" ]; then
        echo -e "${YELLOW}Already connected. Disconnect first with: $0 disconnect${NC}"
        exit 1
    fi

    echo -e "${CYAN}Fetching available nodes...${NC}"
    echo ""

    RESPONSE=$(curl -s -X GET "$API_URL/operator/nodes" \
        -H "Authorization: Bearer $TOKEN")

    # Check for authentication errors
    if echo "$RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
        ERROR_MSG=$(echo "$RESPONSE" | jq -r '.error')
        if [[ "$ERROR_MSG" == *"expired"* ]] || [[ "$ERROR_MSG" == *"invalid"* ]]; then
            echo -e "${RED}✗ Session expired. Please login again: $0 login${NC}"
        else
            echo -e "${RED}✗ Failed to fetch nodes: $ERROR_MSG${NC}"
        fi
        exit 1
    fi

    # Parse nodes into arrays (include public_key and wireguard_port)
    NODES=$(echo "$RESPONSE" | jq -r '.nodes // [] | .[] | "\(.id)|\(.name)|\(.public_ip)|\(.city),\(.country)|\(.status)|\(.public_key)|\(.wireguard_port)"')

    if [ -z "$NODES" ]; then
        echo -e "${RED}✗ No nodes available${NC}"
        exit 1
    fi

    # Display nodes with numbers
    echo -e "${CYAN}Available Nodes:${NC}"
    echo ""

    i=1
    while IFS='|' read -r id name ip location status pubkey wgport; do
        STATUS_COLOR="${GREEN}"
        [ "$status" != "online" ] && STATUS_COLOR="${YELLOW}"
        echo -e "  ${CYAN}[$i]${NC} ${BLUE}$name${NC}"
        echo -e "      📍 $location"
        echo -e "      🌐 $ip"
        echo -e "      ⚡ ${STATUS_COLOR}$status${NC}"
        echo ""
        i=$((i+1))
    done <<< "$NODES"

    # Ask user to select
    read -p "$(echo -e ${CYAN}Select node [1-$((i-1))]: ${NC})" SELECTION

    if ! [[ "$SELECTION" =~ ^[0-9]+$ ]] || [ "$SELECTION" -lt 1 ] || [ "$SELECTION" -ge "$i" ]; then
        echo -e "${RED}✗ Invalid selection${NC}"
        exit 1
    fi

    # Get selected node
    NODE=$(echo "$NODES" | sed -n "${SELECTION}p")
    NODE_ID=$(echo "$NODE" | cut -d'|' -f1)
    NODE_NAME=$(echo "$NODE" | cut -d'|' -f2)
    NODE_IP=$(echo "$NODE" | cut -d'|' -f3)
    WG_PUBKEY=$(echo "$NODE" | cut -d'|' -f6)
    WG_PORT=$(echo "$NODE" | cut -d'|' -f7)

    echo ""
    echo -e "${CYAN}Connecting to ${BLUE}$NODE_NAME${NC}..."

    # Generate WireGuard keys
    echo -e "${CYAN}Generating WireGuard keys...${NC}"
    WG_PRIVATE=$(wg genkey)
    WG_PUBLIC=$(echo "$WG_PRIVATE" | wg pubkey)

    # Register peer with API
    echo -e "${CYAN}Registering with VPN server...${NC}"
    REGISTER_RESPONSE=$(curl -s -X POST "$API_URL/config/generate" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{\"public_key\":\"$WG_PUBLIC\",\"node_id\":\"$NODE_ID\"}")

    # Check for errors
    if echo "$REGISTER_RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
        ERROR_MSG=$(echo "$REGISTER_RESPONSE" | jq -r '.error')
        echo -e "${RED}✗ Registration failed: $ERROR_MSG${NC}"
        exit 1
    fi

    # Extract configuration from response
    SERVER_PUBKEY=$(echo "$REGISTER_RESPONSE" | jq -r '.server_public_key')
    SERVER_ENDPOINT=$(echo "$REGISTER_RESPONSE" | jq -r '.server_endpoint')
    CLIENT_IP=$(echo "$REGISTER_RESPONSE" | jq -r '.client_ip')
    DNS_SERVERS=$(echo "$REGISTER_RESPONSE" | jq -r '.dns')

    if [ -z "$SERVER_PUBKEY" ] || [ "$SERVER_PUBKEY" = "null" ]; then
        echo -e "${RED}✗ Failed to get configuration from server${NC}"
        echo "$REGISTER_RESPONSE"
        exit 1
    fi

    echo -e "${GREEN}✓ Registered successfully${NC}"
    echo -e "${CYAN}  Your VPN IP: ${GREEN}$CLIENT_IP${NC}"

    # Create WireGuard config
    cat > "$WG_CONFIG" << EOF
[Interface]
PrivateKey = $WG_PRIVATE
Address = $CLIENT_IP/32
DNS = $DNS_SERVERS

[Peer]
PublicKey = $SERVER_PUBKEY
Endpoint = $SERVER_ENDPOINT
AllowedIPs = 0.0.0.0/0
PersistentKeepalive = 25
EOF

    chmod 600 "$WG_CONFIG"

    # Connect via WireGuard
    echo -e "${CYAN}Starting VPN tunnel...${NC}"

    if ! command -v wg-quick >/dev/null 2>&1; then
        echo -e "${RED}✗ WireGuard not installed${NC}"
        echo -e "${YELLOW}Install with: brew install wireguard-tools${NC}"
        exit 1
    fi

    sudo wg-quick up "$WG_CONFIG" 2>&1 | grep -v "Warning"

    if [ $? -eq 0 ]; then
        # Save connection info
        echo "NODE_ID=$NODE_ID" > "$CONNECTION_FILE"
        echo "NODE_NAME=$NODE_NAME" >> "$CONNECTION_FILE"
        echo "NODE_IP=$NODE_IP" >> "$CONNECTION_FILE"
        echo "CONNECTED_AT=\"$(date '+%Y-%m-%d %H:%M:%S')\"" >> "$CONNECTION_FILE"

        echo ""
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${GREEN}✓ Connected to $NODE_NAME${NC}"
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo ""
        echo -e "${CYAN}Your IP: ${GREEN}$NODE_IP${NC}"
        echo -e "${CYAN}Location: ${GREEN}$(echo "$NODE" | cut -d'|' -f4)${NC}"
        echo ""
        echo -e "${YELLOW}VPN is active. To disconnect, run: ${CYAN}$0 disconnect${NC}"
        echo ""
    else
        echo -e "${RED}✗ Failed to connect${NC}"
        exit 1
    fi
}

# Disconnect
cmd_disconnect() {
    banner

    if [ ! -f "$CONNECTION_FILE" ]; then
        echo -e "${YELLOW}Not connected${NC}"
        exit 0
    fi

    source "$CONNECTION_FILE"

    echo -e "${CYAN}Disconnecting from $NODE_NAME...${NC}"

    sudo wg-quick down "$WG_CONFIG" 2>&1 | grep -v "Warning"

    rm -f "$CONNECTION_FILE"
    rm -f "$WG_CONFIG"

    echo -e "${GREEN}✓ Disconnected${NC}"
}

# Status
cmd_status() {
    banner

    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  Status${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    # Check session
    if [ -f "$SESSION_FILE" ]; then
        source "$SESSION_FILE"
        echo -e "${GREEN}✓ Logged in${NC} as ${CYAN}$EMAIL${NC}"
    else
        echo -e "${YELLOW}✗ Not logged in${NC}"
    fi

    echo ""

    # Check connection
    if [ -f "$CONNECTION_FILE" ]; then
        source "$CONNECTION_FILE"
        echo -e "${GREEN}✓ Connected${NC} to ${CYAN}$NODE_NAME${NC}"
        echo -e "  📍 IP: ${CYAN}$NODE_IP${NC}"
        echo -e "  ⏱  Since: ${CYAN}$CONNECTED_AT${NC}"

        # Show WireGuard stats if available
        if sudo wg show "$WG_INTERFACE" 2>/dev/null | grep -q "interface"; then
            echo ""
            echo -e "${CYAN}Traffic:${NC}"
            sudo wg show "$WG_INTERFACE" 2>/dev/null | grep -E "transfer:" | sed 's/^/  /'
        fi
    else
        echo -e "${YELLOW}✗ Not connected${NC}"
    fi

    echo ""
}

# Help
cmd_help() {
    banner
    echo "Usage: $0 <command>"
    echo ""
    echo "Commands:"
    echo "  login       - Login with email and password"
    echo "  logout      - Logout and clear session"
    echo "  list        - List available VPN nodes"
    echo "  connect     - Connect to a VPN node (interactive)"
    echo "  disconnect  - Disconnect from VPN"
    echo "  status      - Show connection status"
    echo "  help        - Show this help"
    echo ""
    echo "Examples:"
    echo "  $0 login              # Login first"
    echo "  $0 list               # See available nodes"
    echo "  $0 connect            # Connect to a node"
    echo "  $0 status             # Check connection"
    echo "  $0 disconnect         # Disconnect"
    echo ""
}

# Main
COMMAND=${1:-help}

case "$COMMAND" in
    login)
        cmd_login
        ;;
    logout)
        cmd_logout
        ;;
    list|ls)
        cmd_list
        ;;
    connect|conn)
        cmd_connect
        ;;
    disconnect|disc)
        cmd_disconnect
        ;;
    status|st)
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
