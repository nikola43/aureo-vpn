#!/bin/bash

# Add WireGuard peer script for Docker deployment
# This script generates a WireGuard configuration for a new peer

PUBLIC_KEY="$1"

if [ -z "$PUBLIC_KEY" ]; then
    echo '{"error": "public_key is required"}' >&2
    exit 1
fi

# Generate a unique IP address for this peer
# For simplicity, use a counter-based approach
PEER_COUNT=$(docker exec aureo-vpn-db psql -U postgres -d aureo_vpn -t -c "SELECT COUNT(*) FROM wireguard_peers;" 2>/dev/null | tr -d ' ')
if [ -z "$PEER_COUNT" ] || [ "$PEER_COUNT" = "" ]; then
    PEER_COUNT=0
fi

# Calculate next IP (starting from 10.8.0.2)
PEER_NUM=$((PEER_COUNT + 2))
CLIENT_IP="10.8.0.$PEER_NUM"

# Get server public key from the running WireGuard interface
SERVER_PUBLIC_KEY=$(docker exec aureo-vpn-node-1 wg show wg0 public-key 2>/dev/null || echo "")

# If we can't get it from wg0, try getting from database
if [ -z "$SERVER_PUBLIC_KEY" ]; then
    SERVER_PUBLIC_KEY=$(docker exec aureo-vpn-db psql -U postgres -d aureo_vpn -t -c "SELECT public_key FROM vpn_nodes WHERE status='online' LIMIT 1;" 2>/dev/null | tr -d ' ')
fi

# If still no key, fail
if [ -z "$SERVER_PUBLIC_KEY" ]; then
    echo '{"error": "server public key not found"}' >&2
    exit 1
fi

# Get server endpoint (node's public IP)
SERVER_IP=$(docker exec aureo-vpn-db psql -U postgres -d aureo_vpn -t -c "SELECT public_ip FROM vpn_nodes WHERE status='online' LIMIT 1;" 2>/dev/null | tr -d ' ')
if [ -z "$SERVER_IP" ]; then
    SERVER_IP="155.138.238.145"  # Fallback to known IP
fi

SERVER_PORT="51820"
SERVER_ENDPOINT="$SERVER_IP:$SERVER_PORT"

# Store peer in database
docker exec aureo-vpn-db psql -U postgres -d aureo_vpn -c "
    CREATE TABLE IF NOT EXISTS wireguard_peers (
        id SERIAL PRIMARY KEY,
        public_key TEXT UNIQUE NOT NULL,
        client_ip TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT NOW()
    );
    INSERT INTO wireguard_peers (public_key, client_ip)
    VALUES ('$PUBLIC_KEY', '$CLIENT_IP')
    ON CONFLICT (public_key) DO UPDATE SET client_ip='$CLIENT_IP';
" >/dev/null 2>&1

# Add peer to WireGuard interface
docker exec aureo-vpn-node-1 wg set wg0 peer "$PUBLIC_KEY" allowed-ips "$CLIENT_IP/32" persistent-keepalive 25 >/dev/null 2>&1

# Verify peer was added
if ! docker exec aureo-vpn-node-1 wg show wg0 peers | grep -q "$PUBLIC_KEY"; then
    echo '{"error": "failed to add peer to wireguard interface"}' >&2
    exit 1
fi

# Return JSON configuration (matching the client script expectations)
# Use split routing (0.0.0.0/1 and 128.0.0.0/1) instead of 0.0.0.0/0
# This avoids conflicts with default routes on some systems
cat <<EOF
{
    "client_ip": "$CLIENT_IP",
    "dns": "1.1.1.1,8.8.8.8",
    "server_public_key": "$SERVER_PUBLIC_KEY",
    "server_endpoint": "$SERVER_ENDPOINT",
    "allowed_ips": "0.0.0.0/1, 128.0.0.0/1"
}
EOF

exit 0
