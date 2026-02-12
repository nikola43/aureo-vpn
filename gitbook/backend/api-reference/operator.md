# 💼 Operator

Endpoints for VPN node operators to register, manage nodes, and track earnings.

---

## Register as Operator

Register the current user as a node operator with a cryptocurrency wallet.

```http
POST /api/v1/operator/register
Authorization: Bearer <token>
Content-Type: application/json

{
  "wallet_address": "0x1234567890abcdef1234567890abcdef12345678",
  "wallet_currency": "ethereum",
  "company_name": "My VPN Nodes LLC",
  "contact_email": "contact@myvpn.com"
}
```

### Request Body

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `wallet_address` | string | Yes | Cryptocurrency wallet address for payouts |
| `wallet_currency` | string | Yes | `ethereum`, `bitcoin`, or `litecoin` |
| `company_name` | string | No | Business name (optional) |
| `contact_email` | string | No | Contact email (defaults to account email) |

### Response `201 Created`

```json
{
  "operator": {
    "id": "uuid",
    "status": "pending",
    "wallet_address": "0x1234...",
    "wallet_currency": "ethereum"
  },
  "message": "Registration submitted. Awaiting verification."
}
```

### Operator Status Values

| Status | Description |
|--------|-------------|
| `pending` | Awaiting admin verification |
| `verified` | Approved and can create nodes |
| `suspended` | Temporarily disabled |
| `rejected` | Application rejected |

{% hint style="info" %}
Operator registration requires admin verification before you can create nodes. You will receive an email when your application is reviewed.
{% endhint %}

---

## Create Operator Node

Register a new VPN node under your operator account (requires verified status).

```http
POST /api/v1/operator/nodes
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "us-east-1",
  "hostname": "vpn.example.com",
  "ip_address": "203.0.113.1",
  "country": "US",
  "city": "New York",
  "latitude": 40.7128,
  "longitude": -74.0060,
  "wireguard_port": 51820,
  "openvpn_port": 1194,
  "max_connections": 500
}
```

### Request Body

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Human-readable node name |
| `hostname` | string | No | DNS hostname |
| `ip_address` | string | Yes | Public IP address |
| `country` | string | Yes | ISO 3166-1 alpha-2 country code |
| `city` | string | No | City name |
| `latitude` | float | No | Geographic latitude |
| `longitude` | float | No | Geographic longitude |
| `wireguard_port` | int | No | WireGuard listen port (default: 51820) |
| `openvpn_port` | int | No | OpenVPN listen port (default: 1194) |
| `max_connections` | int | No | Maximum concurrent connections (default: 500) |

### Response `201 Created`

```json
{
  "node": {
    "id": "uuid",
    "name": "us-east-1",
    "wireguard_public_key": "base64...",
    "subnet": "10.8.0.0/24"
  },
  "private_key": "base64...",
  "setup_instructions": "..."
}
```

{% hint style="danger" %}
The `private_key` is returned only once. Save it securely immediately. It is required to configure the WireGuard interface on your server.
{% endhint %}

### Error Responses

| Status | Error | Description |
|--------|-------|-------------|
| `400` | `VALIDATION_ERROR` | Invalid node parameters |
| `401` | `UNAUTHORIZED` | Missing or invalid token |
| `403` | `FORBIDDEN` | Operator not verified |
| `409` | `CONFLICT` | Node name or IP already registered |
| `429` | `LIMIT_EXCEEDED` | Maximum nodes per operator reached (10) |

---

## Get Operator Dashboard

Get a summary of your operator account including all nodes, earnings, and reputation.

```http
GET /api/v1/operator/dashboard
Authorization: Bearer <token>
```

### Response `200 OK`

```json
{
  "total_earned_usd": 156.78,
  "pending_payout_usd": 23.45,
  "total_paid_usd": 133.33,
  "active_nodes": 2,
  "total_bandwidth_gb": 5678.90,
  "total_sessions": 12345,
  "average_uptime_percent": 99.2,
  "reputation_score": 87,
  "current_tier": "gold",
  "nodes": [
    {
      "id": "uuid",
      "name": "us-east-1",
      "status": "online",
      "load_score": 45,
      "active_connections": 123,
      "bandwidth_today_gb": 456.78,
      "earnings_today_usd": 9.12
    }
  ]
}
```

---

## Get Operator Stats

Get detailed statistics for your operator account.

```http
GET /api/v1/operator/stats
Authorization: Bearer <token>
```

### Response `200 OK`

```json
{
  "total_nodes": 2,
  "online_nodes": 2,
  "total_sessions_served": 12345,
  "total_bandwidth_gb": 5678.90,
  "average_session_duration_minutes": 45,
  "average_quality_score": 87,
  "uptime_percent": 99.2,
  "current_tier": "gold",
  "tier_progress": {
    "current": "gold",
    "next": "platinum",
    "uptime_required": 95,
    "reputation_required": 90,
    "current_uptime": 99.2,
    "current_reputation": 87
  }
}
```

---

## Get Earnings

Get detailed earnings history.

```http
GET /api/v1/operator/earnings?limit=10&node_id=uuid
Authorization: Bearer <token>
```

### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | int | 50 | Number of records to return |
| `offset` | int | 0 | Pagination offset |
| `node_id` | string | — | Filter by specific node |

### Response `200 OK`

```json
{
  "earnings": [
    {
      "id": "uuid",
      "node_id": "uuid",
      "node_name": "us-east-1",
      "session_id": "uuid",
      "bandwidth_gb": 2.5,
      "rate_per_gb": 0.020,
      "quality_multiplier": 1.35,
      "duration_bonus": 1.2,
      "amount_usd": 0.081,
      "tier": "gold",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total_count": 12345,
  "total_earned_usd": 156.78
}
```

---

## Get Payouts

Get payout history.

```http
GET /api/v1/operator/payouts
Authorization: Bearer <token>
```

### Response `200 OK`

```json
{
  "payouts": [
    {
      "id": "uuid",
      "amount_usd": 45.67,
      "amount_crypto": "0.0228",
      "currency": "ethereum",
      "wallet_address": "0x1234...",
      "transaction_hash": "0xabcdef...",
      "status": "completed",
      "created_at": "2024-01-10T12:00:00Z",
      "processed_at": "2024-01-10T12:15:00Z"
    }
  ]
}
```

### Payout Status Values

| Status | Description |
|--------|-------------|
| `pending` | Payout requested, awaiting processing |
| `processing` | Transaction being created and broadcast |
| `completed` | Transaction confirmed on blockchain |
| `failed` | Transaction failed (will be retried) |

---

## Request Payout

Request a payout of your pending earnings. Minimum payout is $10 USD.

```http
POST /api/v1/operator/payout/request
Authorization: Bearer <token>
Content-Type: application/json
```

### Response `201 Created`

```json
{
  "payout": {
    "id": "uuid",
    "amount_usd": 23.45,
    "amount_crypto": "0.0117",
    "currency": "ethereum",
    "wallet_address": "0x1234...",
    "status": "pending",
    "estimated_completion": "2024-01-15T12:00:00Z",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

### Error Responses

| Status | Error | Description |
|--------|-------|-------------|
| `400` | `INSUFFICIENT_BALANCE` | Pending balance below $10 minimum |
| `401` | `UNAUTHORIZED` | Missing or invalid token |
| `403` | `FORBIDDEN` | Operator not verified |
| `429` | `RATE_LIMITED` | Payout already requested recently |

---

## Get Reward Tiers

Get the current reward tier structure.

```http
GET /api/v1/operator/reward-tiers
Authorization: Bearer <token>
```

### Response `200 OK`

```json
{
  "tiers": [
    {
      "name": "bronze",
      "rate_per_gb_usd": 0.010,
      "min_uptime_percent": 50,
      "min_reputation_score": 0,
      "bonus_multiplier": 1.0
    },
    {
      "name": "silver",
      "rate_per_gb_usd": 0.015,
      "min_uptime_percent": 80,
      "min_reputation_score": 60,
      "bonus_multiplier": 1.2
    },
    {
      "name": "gold",
      "rate_per_gb_usd": 0.020,
      "min_uptime_percent": 90,
      "min_reputation_score": 75,
      "bonus_multiplier": 1.5
    },
    {
      "name": "platinum",
      "rate_per_gb_usd": 0.030,
      "min_uptime_percent": 95,
      "min_reputation_score": 90,
      "bonus_multiplier": 2.0
    }
  ],
  "earning_formula": "bandwidth_gb * rate_per_gb * quality_multiplier * duration_bonus"
}
```

---

## Usage Examples

```bash
# Register as operator
curl -X POST http://localhost:8080/api/v1/operator/register \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "wallet_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f5bC3F",
    "wallet_currency": "ethereum"
  }'

# Create a node
curl -X POST http://localhost:8080/api/v1/operator/nodes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "us-east-1",
    "ip_address": "203.0.113.1",
    "country": "US",
    "city": "New York",
    "wireguard_port": 51820,
    "max_connections": 500
  }'

# View dashboard
curl -X GET http://localhost:8080/api/v1/operator/dashboard \
  -H "Authorization: Bearer $TOKEN"

# View earnings
curl -X GET "http://localhost:8080/api/v1/operator/earnings?limit=10" \
  -H "Authorization: Bearer $TOKEN"

# Request payout
curl -X POST http://localhost:8080/api/v1/operator/payout/request \
  -H "Authorization: Bearer $TOKEN"
```
