# 🔗 Sessions

Create and manage VPN sessions.

---

## Create Session

Establish a new VPN session with a specific node.

```http
POST /api/v1/sessions
Authorization: Bearer <token>
Content-Type: application/json

{
  "node_id": "uuid",
  "protocol": "wireguard",
  "enable_kill_switch": true,
  "enable_dns_protection": true
}
```

### Request Body

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `node_id` | string | Yes | Target node UUID |
| `protocol` | string | Yes | `wireguard` or `openvpn` |
| `enable_kill_switch` | bool | No | Block traffic if VPN disconnects (default: `true`) |
| `enable_dns_protection` | bool | No | Route DNS through tunnel (default: `true`) |

### Response `201 Created`

```json
{
  "session": {
    "id": "uuid",
    "node_id": "uuid",
    "tunnel_ip": "10.8.0.25",
    "protocol": "wireguard",
    "status": "active",
    "created_at": "2024-01-15T10:30:00Z"
  },
  "config": "[Interface]\nPrivateKey = ...\nAddress = 10.8.0.25/32\nDNS = 1.1.1.1, 8.8.8.8\n\n[Peer]\nPublicKey = ...\nEndpoint = us1.vpn.aureo.io:51820\nAllowedIPs = 0.0.0.0/0\nPersistentKeepalive = 25"
}
```

### Session Creation Flow

```
Client                    API Gateway                VPN Node
  │                           │                         │
  │  POST /sessions           │                         │
  │──────────────────────────▶│                         │
  │                           │  Allocate tunnel IP     │
  │                           │  Generate keypair       │
  │                           │                         │
  │                           │  Add peer to WireGuard  │
  │                           │────────────────────────▶│
  │                           │                         │
  │                           │  Peer added             │
  │                           │◀────────────────────────│
  │                           │                         │
  │  { session, config }      │                         │
  │◀──────────────────────────│                         │
  │                           │                         │
  │  WireGuard handshake (UDP:51820)                    │
  │════════════════════════════════════════════════════▶│
  │◀════════════════════════════════════════════════════│
  │         Encrypted tunnel established                │
```

### Error Responses

| Status | Error | Description |
|--------|-------|-------------|
| `400` | `VALIDATION_ERROR` | Invalid node_id or protocol |
| `401` | `UNAUTHORIZED` | Missing or invalid token |
| `404` | `NOT_FOUND` | Node not found |
| `409` | `CONFLICT` | Max sessions per user exceeded |
| `503` | `SERVICE_UNAVAILABLE` | Node is offline or at capacity |

{% hint style="info" %}
The maximum number of concurrent sessions per user is configurable via `MAX_SESSIONS_PER_USER` (default: 5).
{% endhint %}

---

## Get Session

Retrieve details about an active session.

```http
GET /api/v1/sessions/:id
Authorization: Bearer <token>
```

### Response `200 OK`

```json
{
  "session": {
    "id": "uuid",
    "node_id": "uuid",
    "node_name": "us-east-1",
    "tunnel_ip": "10.8.0.25",
    "protocol": "wireguard",
    "status": "active",
    "bytes_sent": 1234567890,
    "bytes_received": 9876543210,
    "duration_seconds": 1800,
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

### Session Status Values

| Status | Description |
|--------|-------------|
| `active` | Session is connected and transferring data |
| `disconnected` | Session has been terminated |
| `expired` | Session exceeded the timeout limit |
| `error` | Session terminated due to an error |

### Error Responses

| Status | Error | Description |
|--------|-------|-------------|
| `401` | `UNAUTHORIZED` | Missing or invalid token |
| `404` | `NOT_FOUND` | Session not found or not owned by user |

---

## Disconnect Session

Gracefully disconnect an active VPN session.

```http
DELETE /api/v1/sessions/:id
Authorization: Bearer <token>
```

### Response `200 OK`

```json
{
  "session": {
    "id": "uuid",
    "status": "disconnected",
    "bytes_sent": 1234567890,
    "bytes_received": 9876543210,
    "duration_seconds": 3600
  }
}
```

### Disconnect Flow

```
Client                    API Gateway                VPN Node
  │                           │                         │
  │  DELETE /sessions/:id     │                         │
  │──────────────────────────▶│                         │
  │                           │  Remove peer from WG    │
  │                           │────────────────────────▶│
  │                           │                         │
  │                           │  Finalize stats         │
  │                           │◀────────────────────────│
  │                           │                         │
  │                           │  Calculate earnings     │
  │                           │  Update database        │
  │                           │                         │
  │  { session stats }        │                         │
  │◀──────────────────────────│                         │
```

### Error Responses

| Status | Error | Description |
|--------|-------|-------------|
| `401` | `UNAUTHORIZED` | Missing or invalid token |
| `404` | `NOT_FOUND` | Session not found or already disconnected |

---

## Session Monitoring

Active sessions are monitored in the background:

- **Every 30 seconds**: Track bytes sent/received, calculate bandwidth
- **Every 15 minutes**: Flush accumulated earnings to database
- **On disconnect**: Finalize session stats and calculate total earnings

---

## Usage Examples

```bash
# Create a WireGuard session
curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "node_id": "550e8400-e29b-41d4-a716-446655440000",
    "protocol": "wireguard",
    "enable_kill_switch": true,
    "enable_dns_protection": true
  }'

# Check session status
curl -X GET http://localhost:8080/api/v1/sessions/SESSION_ID \
  -H "Authorization: Bearer $TOKEN"

# Disconnect session
curl -X DELETE http://localhost:8080/api/v1/sessions/SESSION_ID \
  -H "Authorization: Bearer $TOKEN"
```
