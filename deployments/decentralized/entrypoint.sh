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
if [ -z "$COUNTRY_CODE" ]; then
    echo "Detecting location..."
    LOCATION=$(curl -s --max-time 5 "https://geo.kamero.ai/api/geo" 2>/dev/null || echo "{}")
    if [ -n "$LOCATION" ] && [ "$LOCATION" != "{}" ]; then
        [ -z "$CITY" ] && CITY=$(echo "$LOCATION" | grep -o '"city":"[^"]*"' | cut -d'"' -f4)
        COUNTRY_CODE=$(echo "$LOCATION" | grep -o '"country":"[^"]*"' | cut -d'"' -f4)
        [ -z "$COUNTRY" ] && COUNTRY="$COUNTRY_CODE"
        [ -z "$LATITUDE" ] && LATITUDE=$(echo "$LOCATION" | grep -o '"latitude":"[^"]*"' | cut -d'"' -f4)
        [ -z "$LONGITUDE" ] && LONGITUDE=$(echo "$LOCATION" | grep -o '"longitude":"[^"]*"' | cut -d'"' -f4)
        [ -z "$PUBLIC_IP" ] && PUBLIC_IP=$(echo "$LOCATION" | grep -o '"ip":"[^"]*"' | cut -d'"' -f4)
        echo "Detected location: $CITY, $COUNTRY_CODE [$LATITUDE, $LONGITUDE]"
    fi
fi

# Enable IP forwarding
echo "Enabling IP forwarding..."
echo 1 > /proc/sys/net/ipv4/ip_forward 2>/dev/null || true

# Build command arguments
CMD_ARRAY=()
[ -n "$NODE_NAME" ] && CMD_ARRAY+=("--name=$NODE_NAME")
[ -n "$DATA_DIR" ] && CMD_ARRAY+=("--data-dir=$DATA_DIR")
[ -n "$PUBLIC_IP" ] && CMD_ARRAY+=("--public-ip=$PUBLIC_IP")
[ -n "$API_PORT" ] && CMD_ARRAY+=("--api-port=$API_PORT")
[ -n "$P2P_PORT" ] && CMD_ARRAY+=("--p2p-port=$P2P_PORT")
[ -n "$WG_PORT" ] && CMD_ARRAY+=("--wg-port=$WG_PORT")
[ -n "$COUNTRY" ] && CMD_ARRAY+=("--country=$COUNTRY")
[ -n "$COUNTRY_CODE" ] && CMD_ARRAY+=("--country-code=$COUNTRY_CODE")
[ -n "$CITY" ] && CMD_ARRAY+=("--city=$CITY")
[ -n "$LATITUDE" ] && CMD_ARRAY+=("--latitude=$LATITUDE")
[ -n "$LONGITUDE" ] && CMD_ARRAY+=("--longitude=$LONGITUDE")
[ -n "$BOOTSTRAP_PEERS" ] && CMD_ARRAY+=("--bootstrap=$BOOTSTRAP_PEERS")
[ "$ENABLE_API" = "false" ] && CMD_ARRAY+=("--enable-api=false")
[ "$ENABLE_VPN" = "false" ] && CMD_ARRAY+=("--enable-vpn=false")

echo "Starting aureo-node with args: ${CMD_ARRAY[*]}"
echo "================================================="

# Execute the node
exec /usr/local/bin/aureo-node "${CMD_ARRAY[@]}"
