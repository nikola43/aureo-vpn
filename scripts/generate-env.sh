#!/bin/bash
set -e

# Generate random secure strings
generate_secret() {
    openssl rand -hex 32
}

generate_node_id() {
    uuidgen
}

# Check if .env already exists
if [ -f .env ]; then
    echo ".env file already exists. Skipping generation."
    exit 0
fi

echo "Generating .env file for production..."

JWT_SECRET=$(generate_secret)
NODE_ID_1=$(generate_node_id)
DB_PASSWORD=$(generate_secret)

cat > .env <<EOF
# Environment
ENVIRONMENT=production

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=${DB_PASSWORD}
DB_NAME=aureo_vpn
DB_SSL_MODE=disable

# JWT
JWT_SECRET=${JWT_SECRET}

# Blockchain (Ethereum - Required for payouts)
# Replace with your actual RPC URL and Private Key
ETHEREUM_RPC_URL=https://mainnet.infura.io/v3/YOUR_INFURA_KEY
ETHEREUM_PRIVATE_KEY=YOUR_PRIVATE_KEY_HERE

# Node IDs
NODE_ID_1=${NODE_ID_1}

# Grafana
GF_SECURITY_ADMIN_PASSWORD=admin

echo ".env file generated successfully!"
echo "IMPORTANT: Please edit .env and set your ETHEREUM_RPC_URL and ETHEREUM_PRIVATE_KEY before launching."
EOF
