# API Testing Guide - Aureo VPN Decentralized Operator System

## Overview

This guide provides complete API testing workflows for the Aureo VPN decentralized operator system, including:
- User registration and authentication
- Operator registration
- Node creation and management
- Earnings tracking
- Payout requests
- Dashboard access

All examples use `curl` for easy testing from the command line.

---

## Prerequisites

### 1. Start the API Gateway

```bash
# Set required environment variables
export JWT_SECRET="your-super-secret-jwt-key-min-32-chars-long"
export DB_HOST="localhost"
export DB_PORT="5432"
export DB_USER="postgres"
export DB_PASSWORD="your_password"
export DB_NAME="aureo_vpn"

# Run the API gateway
cd cmd/api-gateway
go run main.go
```

### 2. PostgreSQL Setup

```bash
# Create database
createdb aureo_vpn

# The application will auto-migrate tables on startup
```

### 3. Test Variables

```bash
# Base URL
API_URL="http://localhost:8080/api/v1"

# Will be populated during tests
ACCESS_TOKEN=""
OPERATOR_ID=""
NODE_ID=""
```

---

## Test Flow 1: Complete Operator Onboarding

### Step 1: User Registration

```bash
curl -X POST "${API_URL}/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "operator1@aureo-vpn.com",
    "password": "SecurePassword123!",
    "username": "operator1"
  }'
```

**Expected Response** (201 Created):
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "user": {
    "id": "uuid-here",
    "email": "operator1@aureo-vpn.com",
    "username": "operator1",
    "is_active": true
  }
}
```

**Save the token**:
```bash
ACCESS_TOKEN="eyJhbGc..."
```

### Step 2: Register as Operator

```bash
curl -X POST "${API_URL}/operator/register" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "wallet_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
    "wallet_type": "ethereum",
    "country": "United States",
    "email": "operator1@aureo-vpn.com",
    "phone_number": "+1234567890"
  }'
```

**Expected Response** (201 Created):
```json
{
  "operator": {
    "id": "operator-uuid",
    "user_id": "user-uuid",
    "wallet_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
    "wallet_type": "ethereum",
    "status": "pending",
    "reputation_score": 50.0,
    "country": "United States",
    "is_verified": false
  },
  "message": "Operator registered successfully. Please wait for verification."
}
```

### Step 3: Admin Verification (Admin Only)

```bash
# This would typically require admin auth
# For testing, you can verify directly in database:
psql -d aureo_vpn -c "UPDATE node_operators SET is_verified = true, status = 'active' WHERE email = 'operator1@aureo-vpn.com';"
```

### Step 4: Create VPN Node

Get your public IP:
```bash
PUBLIC_IP=$(curl -s ifconfig.me)
```

Create node:
```bash
curl -X POST "${API_URL}/operator/nodes" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "US-East-Node-1",
    "hostname": "node1.aureo-vpn.com",
    "public_ip": "'"${PUBLIC_IP}"'",
    "country": "United States",
    "country_code": "US",
    "city": "New York",
    "wireguard_port": 51820,
    "openvpn_port": 1194,
    "latitude": 40.7128,
    "longitude": -74.0060,
    "is_operator_owned": true
  }'
```

**Expected Response** (201 Created):
```json
{
  "node": {
    "id": "node-uuid",
    "name": "US-East-Node-1",
    "hostname": "node1.aureo-vpn.com",
    "public_ip": "1.2.3.4",
    "country": "United States",
    "status": "offline",
    "is_active": true,
    "operator_id": "operator-uuid"
  },
  "public_key": "OPERATOR_NODE_abc123",
  "message": "Node created successfully. Configure your node software with these credentials."
}
```

**Save node ID**:
```bash
NODE_ID="node-uuid"
```

---

## Test Flow 2: Earnings and Payouts

### Step 1: Simulate Session and Earnings

For testing, manually create earnings in the database:

```sql
-- Connect to database
psql -d aureo_vpn

-- Create a test session (find your node_id and user_id first)
INSERT INTO sessions (id, user_id, node_id, protocol, status, created_at)
VALUES (
  gen_random_uuid(),
  (SELECT user_id FROM node_operators WHERE email = 'operator1@aureo-vpn.com'),
  (SELECT id FROM vpn_nodes WHERE name = 'US-East-Node-1'),
  'wireguard',
  'completed',
  NOW()
);

-- Add earnings for the operator
INSERT INTO operator_earnings (
  id,
  operator_id,
  node_id,
  session_id,
  bandwidth_gb,
  duration_minutes,
  rate_per_gb,
  amount_usd,
  status,
  connection_quality,
  created_at
) VALUES (
  gen_random_uuid(),
  (SELECT id FROM node_operators WHERE email = 'operator1@aureo-vpn.com'),
  (SELECT id FROM vpn_nodes WHERE name = 'US-East-Node-1'),
  (SELECT id FROM sessions ORDER BY created_at DESC LIMIT 1),
  100.0,    -- 100 GB bandwidth
  180,      -- 3 hours
  0.02,     -- $0.02 per GB (Gold tier)
  2.40,     -- $2.40 earned
  'confirmed',
  95.0,
  NOW()
);

-- Update operator pending payout
UPDATE node_operators
SET pending_payout = pending_payout + 2.40
WHERE email = 'operator1@aureo-vpn.com';
```

### Step 2: Check Operator Stats

```bash
curl -X GET "${API_URL}/operator/stats" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

**Expected Response**:
```json
{
  "total_earned": 2.40,
  "pending_payout": 2.40,
  "total_paid": 0.00,
  "active_nodes": 1,
  "total_sessions": 1,
  "total_bandwidth_gb": 100.0,
  "reputation_score": 50.0,
  "current_tier": "bronze",
  "monthly_earnings_estimate": 72.00
}
```

### Step 3: View Earnings History

```bash
curl -X GET "${API_URL}/operator/earnings?limit=10&offset=0" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

**Expected Response**:
```json
{
  "earnings": [
    {
      "id": "earning-uuid",
      "amount_usd": 2.40,
      "bandwidth_gb": 100.0,
      "duration_minutes": 180,
      "rate_per_gb": 0.02,
      "status": "confirmed",
      "connection_quality": 95.0,
      "created_at": "2025-10-27T12:00:00Z"
    }
  ],
  "total": 1,
  "limit": 10,
  "offset": 0
}
```

### Step 4: Request Payout

First, add enough earnings to meet minimum ($10):

```sql
UPDATE node_operators
SET pending_payout = 15.00
WHERE email = 'operator1@aureo-vpn.com';
```

Request payout:
```bash
curl -X POST "${API_URL}/operator/payout/request" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

**Expected Response**:
```json
{
  "message": "Payout request submitted successfully. Processing may take 24-48 hours."
}
```

### Step 5: Check Payout History

```bash
curl -X GET "${API_URL}/operator/payouts?limit=10&offset=0" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

**Expected Response**:
```json
{
  "payouts": [
    {
      "id": "payout-uuid",
      "amount_usd": 15.00,
      "crypto_amount": 0.0075,
      "crypto_currency": "ethereum",
      "exchange_rate": 2000.00,
      "wallet_address": "0x742d35...",
      "status": "processing",
      "transaction_hash": "0xabc123...",
      "created_at": "2025-10-27T12:30:00Z"
    }
  ],
  "total": 1,
  "limit": 10,
  "offset": 0
}
```

---

## Test Flow 3: Node Management

### List All Operator Nodes

```bash
curl -X GET "${API_URL}/operator/nodes" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

**Expected Response**:
```json
{
  "nodes": [
    {
      "id": "node-uuid",
      "name": "US-East-Node-1",
      "status": "online",
      "country": "United States",
      "city": "New York",
      "uptime_percentage": 98.5,
      "total_earned_usd": 150.00,
      "created_at": "2025-10-27T10:00:00Z"
    }
  ],
  "count": 1
}
```

### Create Additional Nodes

```bash
curl -X POST "${API_URL}/operator/nodes" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "EU-West-Node-1",
    "hostname": "node2.aureo-vpn.com",
    "public_ip": "5.6.7.8",
    "country": "Germany",
    "country_code": "DE",
    "city": "Frankfurt",
    "wireguard_port": 51820,
    "openvpn_port": 1194,
    "is_operator_owned": true
  }'
```

---

## Test Flow 4: Dashboard

### Get Comprehensive Dashboard Data

```bash
curl -X GET "${API_URL}/operator/dashboard" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

**Expected Response**:
```json
{
  "operator": {
    "id": "operator-uuid",
    "email": "operator1@aureo-vpn.com",
    "wallet_type": "ethereum",
    "wallet_address": "0x742d35...",
    "status": "active",
    "reputation_score": 75.5,
    "total_earned": 150.00,
    "pending_payout": 25.00
  },
  "stats": {
    "total_earned": 150.00,
    "pending_payout": 25.00,
    "total_paid": 125.00,
    "active_nodes": 2,
    "total_sessions": 50,
    "total_bandwidth_gb": 5000.0,
    "current_tier": "gold"
  },
  "active_nodes": [
    {
      "id": "node-uuid-1",
      "name": "US-East-Node-1",
      "status": "online",
      "uptime_percentage": 98.5
    },
    {
      "id": "node-uuid-2",
      "name": "EU-West-Node-1",
      "status": "online",
      "uptime_percentage": 97.2
    }
  ],
  "recent_earnings": [...],
  "recent_payouts": [...]
}
```

---

## Test Flow 5: Public Endpoints

### Get Reward Tiers (No Auth Required)

```bash
curl -X GET "${API_URL}/operator/rewards/tiers"
```

**Expected Response**:
```json
{
  "tiers": [
    {
      "tier_name": "platinum",
      "min_reputation_score": 90.0,
      "min_uptime_percent": 95.0,
      "base_rate_per_gb": 0.03,
      "bonus_multiplier": 2.0,
      "is_active": true
    },
    {
      "tier_name": "gold",
      "min_reputation_score": 75.0,
      "min_uptime_percent": 90.0,
      "base_rate_per_gb": 0.02,
      "bonus_multiplier": 1.5,
      "is_active": true
    },
    {
      "tier_name": "silver",
      "min_reputation_score": 60.0,
      "min_uptime_percent": 80.0,
      "base_rate_per_gb": 0.015,
      "bonus_multiplier": 1.2,
      "is_active": true
    },
    {
      "tier_name": "bronze",
      "min_reputation_score": 0.0,
      "min_uptime_percent": 50.0,
      "base_rate_per_gb": 0.01,
      "bonus_multiplier": 1.0,
      "is_active": true
    }
  ],
  "count": 4
}
```

---

## Test Flow 6: Error Cases

### 1. Register Without Authentication

```bash
curl -X POST "${API_URL}/operator/register" \
  -H "Content-Type: application/json" \
  -d '{
    "wallet_address": "0x123...",
    "wallet_type": "ethereum"
  }'
```

**Expected**: 401 Unauthorized

### 2. Invalid Wallet Address

```bash
curl -X POST "${API_URL}/operator/register" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "wallet_address": "invalid",
    "wallet_type": "ethereum",
    "country": "US",
    "email": "test@test.com"
  }'
```

**Expected**: 400 Bad Request - "invalid wallet address"

### 3. Payout Below Minimum

```sql
-- Set pending payout below $10
UPDATE node_operators
SET pending_payout = 5.00
WHERE email = 'operator1@aureo-vpn.com';
```

```bash
curl -X POST "${API_URL}/operator/payout/request" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

**Expected**: 400 Bad Request - "minimum payout amount is $10.00, current: $5.00"

### 4. Create Node Without Operator Status

```bash
# Register new user who is NOT an operator
curl -X POST "${API_URL}/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "regular@user.com",
    "password": "Password123!",
    "username": "regularuser"
  }'

# Get token
REGULAR_TOKEN="token-from-response"

# Try to create node
curl -X POST "${API_URL}/operator/nodes" \
  -H "Authorization: Bearer ${REGULAR_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Node",
    "hostname": "test.com",
    "public_ip": "1.2.3.4",
    "country": "US",
    "country_code": "US",
    "city": "Test",
    "wireguard_port": 51820,
    "openvpn_port": 1194
  }'
```

**Expected**: 403 Forbidden - "You must be a registered operator to create nodes"

---

## Test Flow 7: Automated Testing Script

Save this as `test-operator-api.sh`:

```bash
#!/bin/bash

API_URL="http://localhost:8080/api/v1"

echo "=== Aureo VPN Operator API Test Suite ==="
echo ""

# 1. User Registration
echo "1. Testing user registration..."
REGISTER_RESPONSE=$(curl -s -X POST "${API_URL}/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test-operator-'$(date +%s)'@aureo-vpn.com",
    "password": "SecurePassword123!",
    "username": "testop'$(date +%s)'"
  }')

ACCESS_TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.access_token')

if [ "$ACCESS_TOKEN" != "null" ]; then
  echo "✅ User registration successful"
else
  echo "❌ User registration failed"
  exit 1
fi

# 2. Operator Registration
echo "2. Testing operator registration..."
OPERATOR_RESPONSE=$(curl -s -X POST "${API_URL}/operator/register" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "wallet_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
    "wallet_type": "ethereum",
    "country": "United States",
    "email": "test@test.com"
  }')

OPERATOR_ID=$(echo $OPERATOR_RESPONSE | jq -r '.operator.id')

if [ "$OPERATOR_ID" != "null" ]; then
  echo "✅ Operator registration successful"
else
  echo "❌ Operator registration failed"
  exit 1
fi

# 3. Get Reward Tiers
echo "3. Testing reward tiers endpoint..."
TIERS_RESPONSE=$(curl -s -X GET "${API_URL}/operator/rewards/tiers")
TIER_COUNT=$(echo $TIERS_RESPONSE | jq '.count')

if [ "$TIER_COUNT" == "4" ]; then
  echo "✅ Reward tiers retrieved (found $TIER_COUNT tiers)"
else
  echo "❌ Reward tiers failed"
fi

# 4. Get Operator Stats
echo "4. Testing operator stats..."
STATS_RESPONSE=$(curl -s -X GET "${API_URL}/operator/stats" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}")

TOTAL_EARNED=$(echo $STATS_RESPONSE | jq -r '.total_earned')

if [ "$TOTAL_EARNED" != "null" ]; then
  echo "✅ Operator stats retrieved"
else
  echo "❌ Operator stats failed"
fi

# 5. Get Dashboard
echo "5. Testing dashboard endpoint..."
DASHBOARD_RESPONSE=$(curl -s -X GET "${API_URL}/operator/dashboard" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}")

DASHBOARD_OPERATOR=$(echo $DASHBOARD_RESPONSE | jq -r '.operator.id')

if [ "$DASHBOARD_OPERATOR" != "null" ]; then
  echo "✅ Dashboard data retrieved"
else
  echo "❌ Dashboard failed"
fi

echo ""
echo "=== Test Suite Complete ==="
```

Run tests:
```bash
chmod +x test-operator-api.sh
./test-operator-api.sh
```

---

## Performance Testing

### Load Testing with Apache Bench

```bash
# Test operator registration endpoint
ab -n 100 -c 10 \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -p operator-payload.json \
  http://localhost:8080/api/v1/operator/stats
```

### Concurrent Payout Requests

```bash
# Test concurrent payout processing
for i in {1..10}; do
  curl -X POST "${API_URL}/operator/payout/request" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" &
done
wait
```

---

## Database Verification

### Check Operator Records

```sql
-- View all operators
SELECT id, email, wallet_type, status, reputation_score, total_earned, pending_payout
FROM node_operators
ORDER BY created_at DESC;

-- View earnings
SELECT oe.amount_usd, oe.bandwidth_gb, oe.status, oe.created_at
FROM operator_earnings oe
JOIN node_operators no ON oe.operator_id = no.id
WHERE no.email = 'operator1@aureo-vpn.com'
ORDER BY oe.created_at DESC;

-- View payouts
SELECT op.amount_usd, op.crypto_currency, op.status, op.transaction_hash, op.created_at
FROM operator_payouts op
JOIN node_operators no ON op.operator_id = no.id
WHERE no.email = 'operator1@aureo-vpn.com'
ORDER BY op.created_at DESC;
```

---

## Postman Collection

Import this JSON into Postman for a complete test suite:

```json
{
  "info": {
    "name": "Aureo VPN Operator API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "variable": [
    {
      "key": "baseUrl",
      "value": "http://localhost:8080/api/v1"
    },
    {
      "key": "accessToken",
      "value": ""
    }
  ],
  "item": [
    {
      "name": "Auth",
      "item": [
        {
          "name": "Register",
          "request": {
            "method": "POST",
            "url": "{{baseUrl}}/auth/register",
            "body": {
              "mode": "raw",
              "raw": "{\n  \"email\": \"operator@test.com\",\n  \"password\": \"Password123!\",\n  \"username\": \"operator1\"\n}"
            }
          }
        }
      ]
    },
    {
      "name": "Operator",
      "item": [
        {
          "name": "Register as Operator",
          "request": {
            "method": "POST",
            "url": "{{baseUrl}}/operator/register",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer {{accessToken}}"
              }
            ]
          }
        }
      ]
    }
  ]
}
```

---

## Troubleshooting

### "Unauthorized" Error

- Check access token is valid
- Token might be expired (refresh it)
- Ensure Bearer prefix in Authorization header

### "Operator not found"

- User might not be registered as operator
- Call `/operator/register` first

### Database Connection Error

- Check PostgreSQL is running
- Verify DB_* environment variables
- Check database exists

### Payout Not Processing

- Check logs for blockchain errors
- Verify blockchain service is configured
- Check wallet has sufficient balance

---

## Next Steps

After testing the API:

1. ✅ **Deploy to Staging** - Test with real blockchain testnet
2. ✅ **Create Frontend Dashboard** - Build React/Vue app using these APIs
3. ✅ **Mobile App** - Develop iOS/Android operator app
4. ✅ **Monitoring** - Set up Grafana dashboards
5. ✅ **Documentation** - Create operator onboarding videos

---

**Version**: 1.0.0
**Last Updated**: 2025-10-27
**Status**: ✅ Ready for Testing
