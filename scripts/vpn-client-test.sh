#!/bin/bash

################################################################################
# Aureo VPN Client Test Tool
#
# This tool helps test VPN node connectivity and performance
# Features:
#   - List available nodes
#   - Generate WireGuard client config
#   - Test node connectivity
#   - Monitor traffic
#   - Check node health
#
# Usage:
#   ./vpn-client-test.sh [command]
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
API_URL="${API_URL:-http://localhost:8080/api/v1}"
CONFIG_DIR="$HOME/.aureo-vpn"

# Banner
show_banner() {
    echo -e "${PURPLE}"
    cat << "EOF"
   _                         __     ______  _   _
  / \  _   _ _ __ ___  ___   \ \   / /  _ \| \ | |
 / _ \| | | | '__/ _ \/ _ \   \ \ / /| |_) |  \| |
/ ___ \ |_| | | |  __/ (_) |   \ V / |  __/| |\  |
/_/   \_\__,_|_|  \___|\___/     \_/  |_|   |_| \_|

        VPN Client Test Tool - Test Your Nodes!
EOF
    echo -e "${NC}"
}

# Login and get access token
login() {
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  Login to Aureo VPN${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    # Ask for email
    read -p "$(echo -e ${CYAN}Email: ${NC})" EMAIL

    # Ask for password (hidden)
    read -sp "$(echo -e ${CYAN}Password: ${NC})" PASSWORD
    echo ""
    echo ""

    if [ -z "$EMAIL" ] || [ -z "$PASSWORD" ]; then
        echo -e "${RED}✗ Email and password are required${NC}"
        exit 1
    fi

    echo -e "${CYAN}🔐 Logging in...${NC}"

    # Make login request
    LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$EMAIL\",
            \"password\": \"$PASSWORD\"
        }")

    # Check for errors
    if echo "$LOGIN_RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
        ERROR_MSG=$(echo "$LOGIN_RESPONSE" | jq -r '.error')
        echo -e "${RED}✗ Login failed: $ERROR_MSG${NC}"
        exit 1
    fi

    # Extract access token
    ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token // empty')

    if [ -z "$ACCESS_TOKEN" ] || [ "$ACCESS_TOKEN" = "null" ]; then
        echo -e "${RED}✗ Login failed: No access token received${NC}"
        echo "Response: $LOGIN_RESPONSE"
        exit 1
    fi

    echo -e "${GREEN}✓ Login successful!${NC}"
    echo ""

    # Save token to temporary session file
    # Ensure config directory exists and is writable
    if [ ! -d "$CONFIG_DIR" ]; then
        mkdir -p "$CONFIG_DIR" 2>/dev/null || {
            echo -e "${YELLOW}⚠ Cannot create $CONFIG_DIR, using /tmp instead${NC}"
            CONFIG_DIR="/tmp/.aureo-vpn-$USER"
            mkdir -p "$CONFIG_DIR"
        }
    fi

    # Test if directory is writable
    if [ ! -w "$CONFIG_DIR" ]; then
        echo -e "${YELLOW}⚠ $CONFIG_DIR not writable, using /tmp instead${NC}"
        CONFIG_DIR="/tmp/.aureo-vpn-$USER"
        mkdir -p "$CONFIG_DIR"
    fi

    SESSION_FILE="$CONFIG_DIR/.session"
    echo "ACCESS_TOKEN=$ACCESS_TOKEN" > "$SESSION_FILE"
    echo "EMAIL=$EMAIL" >> "$SESSION_FILE"
    chmod 600 "$SESSION_FILE"

    echo -e "${GREEN}✓ Session saved to $SESSION_FILE${NC}"
    echo ""
}

# Load credentials
load_credentials() {
    # Try environment variable first
    if [ -n "$AUREO_ACCESS_TOKEN" ]; then
        ACCESS_TOKEN="$AUREO_ACCESS_TOKEN"
        return 0
    fi

    # Check for session file in multiple locations
    for DIR in "$CONFIG_DIR" "/tmp/.aureo-vpn-$USER"; do
        SESSION_FILE="$DIR/.session"
        if [ -f "$SESSION_FILE" ]; then
            CONFIG_DIR="$DIR"
            break
        fi
    done

    # Try session file (from recent login)
    SESSION_FILE="$CONFIG_DIR/.session"
    if [ -f "$SESSION_FILE" ]; then
        # Check if session is less than 24 hours old
        if [ "$(find "$SESSION_FILE" -mtime -1 2>/dev/null)" ]; then
            source "$SESSION_FILE" 2>/dev/null
            if [ -n "$ACCESS_TOKEN" ]; then
                echo -e "${GREEN}✓ Using saved session${NC}"
                return 0
            fi
        fi
    fi

    # Try reading credentials file
    if [ -f "$CONFIG_DIR/operator-credentials" ]; then
        # Try to source the file (may fail if owned by root)
        if source "$CONFIG_DIR/operator-credentials" 2>/dev/null; then
            return 0
        else
            # If permission denied, try with sudo
            echo -e "${YELLOW}⚠ Credentials file requires elevated permissions${NC}"
            if sudo test -f "$CONFIG_DIR/operator-credentials" 2>/dev/null; then
                eval $(sudo cat "$CONFIG_DIR/operator-credentials" | grep -E "^(ACCESS_TOKEN|API_URL|NODE_ID)=" | sed 's/^/export /')
                if [ -n "$ACCESS_TOKEN" ]; then
                    return 0
                fi
            fi
        fi
    fi

    # If no credentials found, prompt for login
    echo -e "${YELLOW}⚠ No active session found${NC}"
    echo ""
    login
}

# List available nodes
list_nodes() {
    echo -e "${CYAN}📡 Fetching available nodes...${NC}\n"

    NODES_RESPONSE=$(curl -s -X GET "$API_URL/operator/nodes" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -H "Content-Type: application/json")

    if echo "$NODES_RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
        echo -e "${RED}✗ Failed to fetch nodes${NC}"
        exit 1
    fi

    # Parse and display nodes
    echo "$NODES_RESPONSE" | jq -r '.nodes[] |
        "[\(.id | .[0:8])...] \(.name)
  Location: \(.city), \(.country)
  IP: \(.public_ip)
  Status: \(.status)
  WireGuard: \(.public_ip):\(.wireguard_port)
  OpenVPN: \(.public_ip):\(.openvpn_port)
  Uptime: \(.uptime_percentage)%
"'
}

# Generate WireGuard config
generate_wireguard_config() {
    local NODE_ID=$1

    if [ -z "$NODE_ID" ]; then
        echo -e "${RED}Error: Node ID required${NC}"
        echo "Usage: $0 wireguard <node_id>"
        exit 1
    fi

    echo -e "${CYAN}🔐 Generating WireGuard configuration...${NC}\n"

    # Get node details
    NODE_RESPONSE=$(curl -s -X GET "$API_URL/operator/nodes/$NODE_ID" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -H "Content-Type: application/json")

    if echo "$NODE_RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
        echo -e "${RED}✗ Node not found${NC}"
        exit 1
    fi

    # Extract node info
    NODE_IP=$(echo "$NODE_RESPONSE" | jq -r '.node.public_ip')
    NODE_PORT=$(echo "$NODE_RESPONSE" | jq -r '.node.wireguard_port')
    NODE_PUBKEY=$(echo "$NODE_RESPONSE" | jq -r '.node.public_key')

    # Generate client keys
    echo -e "${CYAN}Generating client keys...${NC}"
    CLIENT_PRIVATE_KEY=$(wg genkey)
    CLIENT_PUBLIC_KEY=$(echo "$CLIENT_PRIVATE_KEY" | wg pubkey)

    # Create config
    CONFIG_FILE="$CONFIG_DIR/wireguard-$NODE_ID.conf"
    cat > "$CONFIG_FILE" << EOF
[Interface]
PrivateKey = $CLIENT_PRIVATE_KEY
Address = 10.8.0.2/32
DNS = 1.1.1.1, 8.8.8.8

[Peer]
PublicKey = $NODE_PUBKEY
Endpoint = $NODE_IP:$NODE_PORT
AllowedIPs = 0.0.0.0/0
PersistentKeepalive = 25
EOF

    chmod 600 "$CONFIG_FILE"

    echo -e "${GREEN}✓ Configuration generated: $CONFIG_FILE${NC}\n"
    echo -e "${YELLOW}To connect, run:${NC}"
    echo -e "${GREEN}sudo wg-quick up $CONFIG_FILE${NC}\n"
    echo -e "${YELLOW}To disconnect, run:${NC}"
    echo -e "${GREEN}sudo wg-quick down $CONFIG_FILE${NC}\n"
}

# Test node connectivity
test_node() {
    local NODE_ID=$1

    if [ -z "$NODE_ID" ]; then
        echo -e "${RED}Error: Node ID required${NC}"
        echo "Usage: $0 test <node_id>"
        exit 1
    fi

    echo -e "${CYAN}🧪 Testing node connectivity...${NC}\n"

    # Get node details
    NODE_RESPONSE=$(curl -s -X GET "$API_URL/operator/nodes/$NODE_ID" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -H "Content-Type: application/json")

    if echo "$NODE_RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
        echo -e "${RED}✗ Node not found${NC}"
        exit 1
    fi

    NODE_NAME=$(echo "$NODE_RESPONSE" | jq -r '.node.name')
    NODE_IP=$(echo "$NODE_RESPONSE" | jq -r '.node.public_ip')
    NODE_WG_PORT=$(echo "$NODE_RESPONSE" | jq -r '.node.wireguard_port')
    NODE_OVPN_PORT=$(echo "$NODE_RESPONSE" | jq -r '.node.openvpn_port')

    echo -e "${BLUE}Testing: $NODE_NAME${NC}"
    echo -e "${BLUE}IP: $NODE_IP${NC}\n"

    # Test 1: Ping
    echo -e "${CYAN}1. Testing ICMP (ping)...${NC}"
    if ping -c 3 -W 2 "$NODE_IP" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Ping successful${NC}\n"
    else
        echo -e "${YELLOW}⚠ Ping failed (ICMP may be blocked)${NC}\n"
    fi

    # Test 2: WireGuard port
    echo -e "${CYAN}2. Testing WireGuard port ($NODE_WG_PORT/udp)...${NC}"
    if command -v nc > /dev/null 2>&1; then
        if timeout 3 nc -uzv "$NODE_IP" "$NODE_WG_PORT" 2>&1 | grep -q "succeeded\|open"; then
            echo -e "${GREEN}✓ WireGuard port accessible${NC}\n"
        else
            echo -e "${YELLOW}⚠ WireGuard port may be closed or filtered${NC}\n"
        fi
    else
        echo -e "${YELLOW}⚠ netcat not installed, skipping port test${NC}\n"
    fi

    # Test 3: OpenVPN port
    echo -e "${CYAN}3. Testing OpenVPN port ($NODE_OVPN_PORT/tcp)...${NC}"
    if command -v nc > /dev/null 2>&1; then
        if timeout 3 nc -zv "$NODE_IP" "$NODE_OVPN_PORT" 2>&1 | grep -q "succeeded\|open"; then
            echo -e "${GREEN}✓ OpenVPN port accessible${NC}\n"
        else
            echo -e "${YELLOW}⚠ OpenVPN port may be closed or filtered${NC}\n"
        fi
    else
        echo -e "${YELLOW}⚠ netcat not installed, skipping port test${NC}\n"
    fi

    # Test 4: API health check
    echo -e "${CYAN}4. Testing API endpoint...${NC}"
    API_HEALTH=$(curl -s -X GET "$API_URL/../../health" -w "%{http_code}" -o /dev/null)
    if [ "$API_HEALTH" = "200" ]; then
        echo -e "${GREEN}✓ API is healthy${NC}\n"
    else
        echo -e "${RED}✗ API returned status $API_HEALTH${NC}\n"
    fi

    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}  Test Complete!${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

# Monitor node stats
monitor_node() {
    local NODE_ID=$1

    if [ -z "$NODE_ID" ]; then
        echo -e "${RED}Error: Node ID required${NC}"
        echo "Usage: $0 monitor <node_id>"
        exit 1
    fi

    echo -e "${CYAN}📊 Monitoring node (press Ctrl+C to stop)...${NC}\n"

    while true; do
        # Clear screen
        clear
        show_banner

        # Get node stats
        NODE_RESPONSE=$(curl -s -X GET "$API_URL/operator/nodes/$NODE_ID" \
            -H "Authorization: Bearer $ACCESS_TOKEN" \
            -H "Content-Type: application/json")

        if echo "$NODE_RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
            echo -e "${RED}✗ Failed to fetch node stats${NC}"
            sleep 5
            continue
        fi

        # Display stats
        echo "$NODE_RESPONSE" | jq -r '"
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  NODE: \(.node.name)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Status:              \(.node.status)
  Location:            \(.node.city), \(.node.country)
  IP:                  \(.node.public_ip)

  Current Connections: \(.node.current_connections) / \(.node.max_connections)
  CPU Usage:           \(.node.cpu_usage)%
  Memory Usage:        \(.node.memory_usage)%
  Bandwidth:           \(.node.bandwidth_usage_gbps) Gbps

  Uptime:              \(.node.uptime_percentage)%
  Total Earned:        $\(.node.total_earned_usd)

  Last Heartbeat:      \(.node.last_heartbeat)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
"'

        # Update every 3 seconds
        sleep 3
    done
}

# Show session status
status() {
    # Check for session file in multiple locations
    SESSION_FILE=""
    for DIR in "$CONFIG_DIR" "/tmp/.aureo-vpn-$USER"; do
        if [ -f "$DIR/.session" ]; then
            SESSION_FILE="$DIR/.session"
            CONFIG_DIR="$DIR"
            break
        fi
    done

    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  Session Status${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    if [ -n "$SESSION_FILE" ] && [ -f "$SESSION_FILE" ]; then
        source "$SESSION_FILE" 2>/dev/null

        # Check if session is expired (older than 24 hours)
        if [ "$(find "$SESSION_FILE" -mtime -1 2>/dev/null)" ]; then
            echo -e "${GREEN}✓ Active session found${NC}"
            echo -e "  Email:      ${CYAN}${EMAIL}${NC}"
            echo -e "  Token:      ${CYAN}${ACCESS_TOKEN:0:20}...${NC}"
            echo -e "  Expires:    ${CYAN}$(date -r "$SESSION_FILE" -v+24H '+%Y-%m-%d %H:%M:%S' 2>/dev/null || date -d @$(($(stat -f %m "$SESSION_FILE" 2>/dev/null || stat -c %Y "$SESSION_FILE") + 86400)) '+%Y-%m-%d %H:%M:%S' 2>/dev/null || echo "in 24h")${NC}"
        else
            echo -e "${YELLOW}⚠ Session expired${NC}"
            echo -e "  Please login again: ${CYAN}$0 login${NC}"
        fi
    else
        echo -e "${YELLOW}⚠ No active session${NC}"
        echo -e "  Login with: ${CYAN}$0 login${NC}"
    fi

    echo ""
}

# Logout
logout() {
    # Check for session file in multiple locations and remove all
    FOUND=false
    for DIR in "$CONFIG_DIR" "/tmp/.aureo-vpn-$USER"; do
        SESSION_FILE="$DIR/.session"
        if [ -f "$SESSION_FILE" ]; then
            rm -f "$SESSION_FILE"
            FOUND=true
        fi
    done

    if [ "$FOUND" = true ]; then
        echo -e "${GREEN}✓ Logged out successfully${NC}"
    else
        echo -e "${YELLOW}⚠ No active session found${NC}"
    fi
}

# Show help
show_help() {
    echo "Aureo VPN Client Test Tool"
    echo ""
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  login                 - Login with email and password"
    echo "  status                - Show current session status"
    echo "  logout                - Clear saved session"
    echo "  list                  - List all available nodes"
    echo "  wireguard <node_id>   - Generate WireGuard client config"
    echo "  test <node_id>        - Test node connectivity"
    echo "  monitor <node_id>     - Monitor node stats in real-time"
    echo "  help                  - Show this help message"
    echo ""
    echo "Authentication:"
    echo "  The tool will automatically prompt for login if no session exists."
    echo "  Sessions are valid for 24 hours and stored in ~/.aureo-vpn/.session"
    echo ""
    echo "  You can also set AUREO_ACCESS_TOKEN environment variable:"
    echo "    export AUREO_ACCESS_TOKEN='your-token-here'"
    echo ""
    echo "Examples:"
    echo "  $0 login"
    echo "  $0 list"
    echo "  $0 wireguard 88853565-2394-411b-85c0-f8baff02dd1e"
    echo "  $0 test 88853565-2394-411b-85c0-f8baff02dd1e"
    echo "  $0 monitor 88853565-2394-411b-85c0-f8baff02dd1e"
    echo "  $0 logout"
    echo ""
}

# Main
main() {
    show_banner

    # Create config directory if it doesn't exist
    mkdir -p "$CONFIG_DIR"

    # Parse command
    COMMAND=${1:-help}

    case "$COMMAND" in
        login)
            login
            ;;
        status)
            status
            ;;
        logout)
            logout
            ;;
        list)
            load_credentials
            list_nodes
            ;;
        wireguard)
            load_credentials
            generate_wireguard_config "$2"
            ;;
        test)
            load_credentials
            test_node "$2"
            ;;
        monitor)
            load_credentials
            monitor_node "$2"
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            echo -e "${RED}Unknown command: $COMMAND${NC}\n"
            show_help
            exit 1
            ;;
    esac
}

main "$@"
