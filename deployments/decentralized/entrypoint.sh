#!/bin/bash
set -e

echo "================================================="
echo "   AUREO VPN - Decentralized Node"
echo "================================================="

# Auto-detect public IP if not set
if [ -z "$PUBLIC_IP" ]; then
    echo "Detecting public IP..."
    PUBLIC_IP=$(curl -s --max-time 5 https://api.ipify.org 2>/dev/null || \
                curl -s --max-time 5 https://ifconfig.me/ip 2>/dev/null || \
                curl -s --max-time 5 https://icanhazip.com 2>/dev/null || \
                echo "")
    if [ -n "$PUBLIC_IP" ]; then
        echo "Detected public IP: $PUBLIC_IP"
    else
        echo "Warning: Could not detect public IP"
    fi
fi

# Auto-detect location if not set
if [ -z "$COUNTRY_CODE" ] && [ -n "$PUBLIC_IP" ]; then
    echo "Detecting location..."
    LOCATION=$(curl -s --max-time 5 "http://ip-api.com/json/$PUBLIC_IP" 2>/dev/null || echo "{}")
    if [ -n "$LOCATION" ] && [ "$LOCATION" != "{}" ]; then
        COUNTRY=$(echo "$LOCATION" | grep -o '"country":"[^"]*"' | cut -d'"' -f4)
        COUNTRY_CODE=$(echo "$LOCATION" | grep -o '"countryCode":"[^"]*"' | cut -d'"' -f4)
        CITY=$(echo "$LOCATION" | grep -o '"city":"[^"]*"' | cut -d'"' -f4)
        echo "Detected location: $CITY, $COUNTRY ($COUNTRY_CODE)"
    fi
fi

# Enable IP forwarding
echo "Enabling IP forwarding..."
echo 1 > /proc/sys/net/ipv4/ip_forward 2>/dev/null || true

# Build command arguments
CMD_ARGS=""
[ -n "$NODE_NAME" ] && CMD_ARGS="$CMD_ARGS --name=$NODE_NAME"
[ -n "$DATA_DIR" ] && CMD_ARGS="$CMD_ARGS --data-dir=$DATA_DIR"
[ -n "$PUBLIC_IP" ] && CMD_ARGS="$CMD_ARGS --public-ip=$PUBLIC_IP"
[ -n "$API_PORT" ] && CMD_ARGS="$CMD_ARGS --api-port=$API_PORT"
[ -n "$P2P_PORT" ] && CMD_ARGS="$CMD_ARGS --p2p-port=$P2P_PORT"
[ -n "$WG_PORT" ] && CMD_ARGS="$CMD_ARGS --wg-port=$WG_PORT"
[ -n "$COUNTRY" ] && CMD_ARGS="$CMD_ARGS --country=$COUNTRY"
[ -n "$COUNTRY_CODE" ] && CMD_ARGS="$CMD_ARGS --country-code=$COUNTRY_CODE"
[ -n "$CITY" ] && CMD_ARGS="$CMD_ARGS --city=$CITY"
[ -n "$BOOTSTRAP_PEERS" ] && CMD_ARGS="$CMD_ARGS --bootstrap=$BOOTSTRAP_PEERS"
[ "$ENABLE_API" = "false" ] && CMD_ARGS="$CMD_ARGS --enable-api=false"
[ "$ENABLE_VPN" = "false" ] && CMD_ARGS="$CMD_ARGS --enable-vpn=false"

echo "Starting aureo-node with args: $CMD_ARGS"
echo "================================================="

# Execute the node
exec /usr/local/bin/aureo-node $CMD_ARGS
