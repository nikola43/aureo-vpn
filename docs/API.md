# Aureo VPN API Documentation

## Base URL
```
Production: https://api.aureo-vpn.com/v1
Staging: https://staging-api.aureo-vpn.com/v1
```

## Authentication

All authenticated endpoints require a Bearer token in the Authorization header:
```
Authorization: Bearer <access_token>
```

## Rate Limiting

- **Anonymous**: 100 requests/minute
- **Authenticated**: 1000 requests/minute
- **Premium**: 5000 requests/minute

Rate limit headers are included in all responses:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 998
X-RateLimit-Reset: 1640995200
```

## Endpoints

### Authentication

#### POST /auth/register
Create a new user account.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "username": "username",
  "full_name": "John Doe"
}
```

**Response:** `201 Created`
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "username",
    "subscription_tier": "free",
    "created_at": "2024-01-15T10:00:00Z"
  }
}
```

#### POST /auth/login
Authenticate a user.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
```

**Response:** `200 OK`
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "username"
  }
}
```

#### POST /auth/refresh
Refresh an access token.

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Response:** `200 OK`
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### User Management

#### GET /user/profile
Get the authenticated user's profile.

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "username": "username",
  "full_name": "John Doe",
  "subscription_tier": "premium",
  "subscription_expiry": "2025-01-15T10:00:00Z",
  "data_transferred_gb": 125.5,
  "connection_count": 342,
  "is_active": true,
  "created_at": "2024-01-15T10:00:00Z"
}
```

#### GET /user/sessions
Get active VPN sessions.

**Response:** `200 OK`
```json
{
  "sessions": [
    {
      "id": "uuid",
      "node": {
        "id": "uuid",
        "name": "US-East-1",
        "country": "United States",
        "city": "New York"
      },
      "protocol": "wireguard",
      "tunnel_ip": "10.8.0.5",
      "status": "active",
      "connected_at": "2024-01-15T10:00:00Z",
      "data_used_gb": 2.5
    }
  ],
  "count": 1
}
```

#### GET /user/stats
Get user statistics.

**Response:** `200 OK`
```json
{
  "total_sessions": 342,
  "active_sessions": 1,
  "data_transferred_gb": 125.5,
  "avg_session_duration_minutes": 45,
  "favorite_countries": ["US", "GB", "DE"]
}
```

### VPN Nodes

#### GET /nodes
List available VPN nodes.

**Query Parameters:**
- `country` - Filter by country code (e.g., US, GB, DE)
- `protocol` - Filter by protocol (wireguard, openvpn, ipsec)
- `multihop` - Filter multi-hop capable nodes (true/false)
- `limit` - Results per page (default: 50, max: 100)
- `offset` - Pagination offset

**Response:** `200 OK`
```json
{
  "nodes": [
    {
      "id": "uuid",
      "name": "US-East-1",
      "hostname": "us-east-1.aureo-vpn.com",
      "public_ip": "192.0.2.1",
      "country": "United States",
      "country_code": "US",
      "city": "New York",
      "latitude": 40.7128,
      "longitude": -74.0060,
      "load_score": 25.5,
      "current_connections": 250,
      "max_connections": 1000,
      "latency": 15,
      "status": "online",
      "supports_wireguard": true,
      "supports_openvpn": true,
      "supports_ipsec": true,
      "supports_multihop": true
    }
  ],
  "count": 1,
  "total": 5234
}
```

#### GET /nodes/best
Get the best node based on criteria.

**Query Parameters:**
- `protocol` - Preferred protocol (default: wireguard)
- `country` - Target country code
- `feature` - Required feature (streaming, p2p, multihop)

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "name": "US-East-1",
  "country": "United States",
  "load_score": 12.3,
  "latency": 8,
  "recommendation_reason": "Lowest latency and load in US region"
}
```

#### GET /nodes/:id
Get details for a specific node.

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "name": "US-East-1",
  "hostname": "us-east-1.aureo-vpn.com",
  "public_ip": "192.0.2.1",
  "country": "United States",
  "city": "New York",
  "load_score": 25.5,
  "current_connections": 250,
  "max_connections": 1000,
  "bandwidth_usage_gbps": 2.5,
  "cpu_usage": 35.2,
  "memory_usage": 42.1,
  "status": "online",
  "last_heartbeat": "2024-01-15T10:00:00Z"
}
```

### VPN Sessions

#### POST /sessions/create
Create a new VPN session.

**Request:**
```json
{
  "node_id": "uuid",
  "protocol": "wireguard",
  "options": {
    "kill_switch": true,
    "dns_leak_protection": true,
    "split_tunneling": false,
    "obfuscation": "stealth"
  }
}
```

**Response:** `201 Created`
```json
{
  "session": {
    "id": "uuid",
    "node_id": "uuid",
    "protocol": "wireguard",
    "tunnel_ip": "10.8.0.5",
    "status": "active",
    "connected_at": "2024-01-15T10:00:00Z"
  },
  "config": {
    "client_config": "...",
    "dns_servers": ["1.1.1.1", "1.0.0.1"],
    "mtu": 1420
  }
}
```

#### DELETE /sessions/:id
Disconnect a VPN session.

**Response:** `200 OK`
```json
{
  "message": "Session disconnected successfully",
  "session_id": "uuid",
  "duration_seconds": 3600,
  "data_used_gb": 1.2
}
```

### Multi-Hop

#### POST /multihop/create
Create a multi-hop VPN chain.

**Request:**
```json
{
  "entry_country": "CH",
  "exit_country": "IS",
  "type": "double",
  "protocol": "wireguard"
}
```

**Response:** `201 Created`
```json
{
  "chain_id": "uuid",
  "entry_node": {
    "id": "uuid",
    "name": "CH-Zurich-1",
    "country": "Switzerland"
  },
  "exit_node": {
    "id": "uuid",
    "name": "IS-Reykjavik-1",
    "country": "Iceland"
  },
  "estimated_latency_ms": 45,
  "estimated_speed_reduction": 0.70,
  "config": "..."
}
```

### Configuration

#### POST /config/generate
Generate VPN configuration.

**Request:**
```json
{
  "node_id": "uuid",
  "protocol": "wireguard",
  "format": "file"
}
```

**Response:** `200 OK`
```json
{
  "config_id": "uuid",
  "protocol": "wireguard",
  "config_content": "[Interface]\nPrivateKey = ...",
  "qr_code": "data:image/png;base64,...",
  "expires_at": "2025-01-15T10:00:00Z"
}
```

### Payments

#### GET /payment/cryptocurrencies
List supported cryptocurrencies.

**Response:** `200 OK`
```json
{
  "cryptocurrencies": [
    {
      "symbol": "BTC",
      "name": "Bitcoin",
      "network": "Bitcoin",
      "confirmations_required": 3,
      "current_rate_usd": 45000.00
    },
    {
      "symbol": "ETH",
      "name": "Ethereum",
      "network": "Ethereum",
      "confirmations_required": 12,
      "current_rate_usd": 3000.00
    }
  ]
}
```

#### POST /payment/create
Create a cryptocurrency payment.

**Request:**
```json
{
  "cryptocurrency": "BTC",
  "subscription_tier": "premium",
  "duration_months": 12
}
```

**Response:** `201 Created`
```json
{
  "payment_id": "uuid",
  "amount_usd": 107.91,
  "amount_crypto": 0.00239800,
  "cryptocurrency": "BTC",
  "address": "bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh",
  "qr_code": "bitcoin:bc1qxy2...?amount=0.002398",
  "expires_at": "2024-01-16T10:00:00Z",
  "status": "pending"
}
```

#### GET /payment/:id/status
Check payment status.

**Response:** `200 OK`
```json
{
  "payment_id": "uuid",
  "status": "confirmed",
  "confirmations": 3,
  "required_confirmations": 3,
  "tx_hash": "a1b2c3d4e5f6...",
  "subscription_activated": true
}
```

### Admin Endpoints

#### GET /admin/nodes
List all nodes (admin only).

#### GET /admin/users
List all users (admin only).

#### GET /admin/stats
Get system statistics (admin only).

**Response:** `200 OK`
```json
{
  "total_users": 125000,
  "active_users": 45000,
  "total_nodes": 5234,
  "online_nodes": 5180,
  "active_sessions": 38000,
  "total_bandwidth_gbps": 1250.5,
  "avg_latency_ms": 22,
  "uptime_percentage": 99.97
}
```

## Webhooks

### Payment Confirmation
Notifies when a payment is confirmed.

**Payload:**
```json
{
  "event": "payment.confirmed",
  "payment_id": "uuid",
  "user_id": "uuid",
  "amount": 107.91,
  "cryptocurrency": "BTC",
  "tx_hash": "a1b2c3d4e5f6...",
  "subscription_tier": "premium",
  "duration_months": 12,
  "timestamp": "2024-01-15T10:00:00Z"
}
```

### Session Started
Notifies when a VPN session starts.

**Payload:**
```json
{
  "event": "session.started",
  "session_id": "uuid",
  "user_id": "uuid",
  "node_id": "uuid",
  "protocol": "wireguard",
  "timestamp": "2024-01-15T10:00:00Z"
}
```

## Error Responses

All errors follow this format:

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": {
    "field": "Additional context"
  }
}
```

### Common Error Codes

| Code | Status | Description |
|------|--------|-------------|
| `UNAUTHORIZED` | 401 | Invalid or expired token |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `VALIDATION_ERROR` | 400 | Invalid request data |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests |
| `SERVER_ERROR` | 500 | Internal server error |

## SDK Examples

### cURL
```bash
# Login
curl -X POST https://api.aureo-vpn.com/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}'

# Get nodes
curl -X GET "https://api.aureo-vpn.com/v1/nodes?country=US" \
  -H "Authorization: Bearer <token>"
```

### Go
```go
import "github.com/nikola43/aureo-vpn/pkg/client"

client := client.New("https://api.aureo-vpn.com/v1")
client.SetToken("access_token")

nodes, err := client.GetNodes(&client.NodeFilter{
    Country: "US",
    Protocol: "wireguard",
})
```

### Python
```python
import requests

client = requests.Session()
client.headers.update({
    'Authorization': 'Bearer <token>'
})

response = client.get('https://api.aureo-vpn.com/v1/nodes', params={
    'country': 'US',
    'protocol': 'wireguard'
})
nodes = response.json()
```

### JavaScript
```javascript
const client = axios.create({
  baseURL: 'https://api.aureo-vpn.com/v1',
  headers: {
    'Authorization': `Bearer ${token}`
  }
});

const { data } = await client.get('/nodes', {
  params: {
    country: 'US',
    protocol: 'wireguard'
  }
});
```

## Best Practices

1. **Cache tokens** - Access tokens are valid for 15 minutes
2. **Use refresh tokens** - Refresh before expiration
3. **Handle rate limits** - Implement exponential backoff
4. **Verify webhooks** - Always verify webhook signatures
5. **Use HTTPS** - Never send tokens over HTTP
6. **Monitor errors** - Log and monitor API errors

## Support

- Documentation: https://docs.aureo-vpn.com
- API Status: https://status.aureo-vpn.com
- Support: support@aureo-vpn.com
