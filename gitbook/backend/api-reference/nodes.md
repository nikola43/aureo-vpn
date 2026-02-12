# 📡 Nodes

Discover and query available VPN nodes.

---

## List Nodes

Get a list of available VPN nodes with optional filtering.

```http
GET /api/v1/nodes?country=US&protocol=wireguard&limit=10
Authorization: Bearer <token>
```

### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `country` | string | — | Filter by country code (e.g., `US`, `GB`, `DE`) |
| `protocol` | string | — | Filter by protocol (`wireguard`, `openvpn`, `ipsec`) |
| `multihop` | bool | — | Filter multi-hop capable nodes (`true`/`false`) |
| `limit` | int | 50 | Results per page (max: 100) |
| `offset` | int | 0 | Pagination offset |

### Response `200 OK`

```json
{
  "nodes": [
    {
      "id": "uuid",
      "name": "us-east-1",
      "hostname": "us1.vpn.aureo.io",
      "ip_address": "203.0.113.1",
      "country": "US",
      "city": "New York",
      "latitude": 40.7128,
      "longitude": -74.0060,
      "load_score": 25,
      "current_connections": 250,
      "max_connections": 1000,
      "wireguard_port": 51820,
      "wireguard_public_key": "base64...",
      "status": "online",
      "latency_ms": 15,
      "supports_wireguard": true,
      "supports_openvpn": true,
      "supports_multihop": true,
      "features": ["multihop", "dns_protection"]
    }
  ],
  "count": 1,
  "total": 5234
}
```

### Node Status Values

| Status | Description |
|--------|-------------|
| `online` | Node is active and accepting connections |
| `offline` | Node is not responding |
| `maintenance` | Node is temporarily unavailable |
| `full` | Node has reached max connections |

---

## Get Best Node

Get the optimal node based on load, latency, and location.

```http
GET /api/v1/nodes/best?country=US&protocol=wireguard
Authorization: Bearer <token>
```

### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `country` | string | — | Target country code |
| `protocol` | string | `wireguard` | Preferred protocol |
| `feature` | string | — | Required feature (`streaming`, `p2p`, `multihop`) |

### Response `200 OK`

```json
{
  "node": {
    "id": "uuid",
    "name": "us-east-1",
    "hostname": "us1.vpn.aureo.io",
    "ip_address": "203.0.113.1",
    "country": "US",
    "city": "New York",
    "load_score": 12.3,
    "latency_ms": 8,
    "wireguard_port": 51820,
    "wireguard_public_key": "base64...",
    "status": "online",
    "recommendation_reason": "Lowest latency and load in US region"
  }
}
```

{% hint style="info" %}
The best node is selected using a scoring algorithm that considers load score, latency, geographic proximity, and current connection count.
{% endhint %}

---

## Get Node by ID

Get detailed information about a specific node.

```http
GET /api/v1/nodes/:id
Authorization: Bearer <token>
```

### Response `200 OK`

```json
{
  "id": "uuid",
  "name": "us-east-1",
  "hostname": "us1.vpn.aureo.io",
  "ip_address": "203.0.113.1",
  "country": "US",
  "city": "New York",
  "latitude": 40.7128,
  "longitude": -74.0060,
  "load_score": 25.5,
  "current_connections": 250,
  "max_connections": 1000,
  "bandwidth_usage_gbps": 2.5,
  "cpu_usage": 35.2,
  "memory_usage": 42.1,
  "wireguard_port": 51820,
  "wireguard_public_key": "base64...",
  "openvpn_port": 1194,
  "supports_wireguard": true,
  "supports_openvpn": true,
  "supports_ipsec": true,
  "supports_multihop": true,
  "status": "online",
  "last_heartbeat": "2024-01-15T10:00:00Z"
}
```

### Error Responses

| Status | Error | Description |
|--------|-------|-------------|
| `401` | `UNAUTHORIZED` | Missing or invalid token |
| `404` | `NOT_FOUND` | Node not found |

---

## Load Score Calculation

The load score (0-100) determines how busy a node is:

```
load_score = (active_connections / max_connections) × 50
           + cpu_usage × 0.25
           + memory_usage × 0.25
```

| Score | Classification |
|-------|---------------|
| 0-25 | Low load (ideal) |
| 26-50 | Moderate load |
| 51-75 | High load |
| 76-100 | Near capacity |

---

## Usage Examples

```bash
# List all US WireGuard nodes
curl -X GET "http://localhost:8080/api/v1/nodes?country=US&protocol=wireguard" \
  -H "Authorization: Bearer $TOKEN"

# Get the best node in Germany
curl -X GET "http://localhost:8080/api/v1/nodes/best?country=DE" \
  -H "Authorization: Bearer $TOKEN"

# Get specific node details
curl -X GET "http://localhost:8080/api/v1/nodes/550e8400-e29b-41d4-a716-446655440000" \
  -H "Authorization: Bearer $TOKEN"
```
