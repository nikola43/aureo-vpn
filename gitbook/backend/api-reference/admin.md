# 🔐 Admin

Administrative endpoints for managing the platform. All admin endpoints require a valid JWT token with `is_admin: true`.

---

## Authentication

All admin endpoints require:

```
Authorization: Bearer <admin_token>
```

The authenticated user must have the `is_admin` flag set to `true` in the database. Regular users will receive a `403 Forbidden` response.

---

## Endpoints Overview

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/admin/stats` | Get system-wide statistics |
| `GET` | `/api/v1/admin/nodes` | List all nodes |
| `POST` | `/api/v1/admin/nodes` | Create a new node |
| `PUT` | `/api/v1/admin/nodes/:id` | Update a node |
| `DELETE` | `/api/v1/admin/nodes/:id` | Delete a node |
| `GET` | `/api/v1/admin/users` | List all users |
| `GET` | `/api/v1/admin/users/:id` | Get a specific user |
| `PUT` | `/api/v1/admin/users/:id` | Update a user |
| `DELETE` | `/api/v1/admin/users/:id` | Delete a user |
| `GET` | `/api/v1/admin/sessions` | List all active sessions |
| `POST` | `/api/v1/admin/operator/:id/verify` | Verify an operator |

---

## Get System Stats

Get system-wide statistics for the platform.

```http
GET /api/v1/admin/stats
Authorization: Bearer <admin_token>
```

### Response `200 OK`

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

---

## List All Nodes

Get all nodes in the system (including offline and maintenance nodes).

```http
GET /api/v1/admin/nodes
Authorization: Bearer <admin_token>
```

### Response `200 OK`

```json
{
  "nodes": [
    {
      "id": "uuid",
      "name": "us-east-1",
      "ip_address": "203.0.113.1",
      "country": "US",
      "status": "online",
      "operator_id": "uuid",
      "load_score": 25,
      "current_connections": 250,
      "max_connections": 1000,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 5234
}
```

---

## Create Node

Create a new VPN node (admin-provisioned).

```http
POST /api/v1/admin/nodes
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "name": "us-west-1",
  "ip_address": "203.0.113.50",
  "country": "US",
  "city": "Los Angeles",
  "wireguard_port": 51820,
  "max_connections": 1000
}
```

### Response `201 Created`

```json
{
  "node": {
    "id": "uuid",
    "name": "us-west-1",
    "wireguard_public_key": "base64...",
    "status": "online"
  },
  "private_key": "base64..."
}
```

---

## Update Node

Update an existing node's configuration.

```http
PUT /api/v1/admin/nodes/:id
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "max_connections": 2000,
  "status": "maintenance"
}
```

### Response `200 OK`

```json
{
  "node": {
    "id": "uuid",
    "name": "us-west-1",
    "max_connections": 2000,
    "status": "maintenance"
  }
}
```

---

## Delete Node

Remove a node from the network.

```http
DELETE /api/v1/admin/nodes/:id
Authorization: Bearer <admin_token>
```

### Response `200 OK`

```json
{
  "message": "Node deleted successfully"
}
```

---

## List All Users

Get all registered users.

```http
GET /api/v1/admin/users
Authorization: Bearer <admin_token>
```

### Response `200 OK`

```json
{
  "users": [
    {
      "id": "uuid",
      "email": "user@example.com",
      "username": "johndoe",
      "subscription_tier": "premium",
      "is_active": true,
      "is_admin": false,
      "is_operator": false,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 125000
}
```

---

## Get User

Get details for a specific user.

```http
GET /api/v1/admin/users/:id
Authorization: Bearer <admin_token>
```

### Response `200 OK`

```json
{
  "id": "uuid",
  "email": "user@example.com",
  "username": "johndoe",
  "subscription_tier": "premium",
  "is_active": true,
  "is_admin": false,
  "is_operator": true,
  "data_transferred_gb": 125.5,
  "connection_count": 342,
  "created_at": "2024-01-01T00:00:00Z"
}
```

---

## Update User

Update a user's account settings.

```http
PUT /api/v1/admin/users/:id
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "is_active": false,
  "subscription_tier": "free"
}
```

### Response `200 OK`

```json
{
  "user": {
    "id": "uuid",
    "is_active": false,
    "subscription_tier": "free"
  }
}
```

---

## Delete User

Delete a user account and all associated data.

```http
DELETE /api/v1/admin/users/:id
Authorization: Bearer <admin_token>
```

### Response `200 OK`

```json
{
  "message": "User deleted successfully"
}
```

---

## List All Sessions

Get all active VPN sessions across the platform.

```http
GET /api/v1/admin/sessions
Authorization: Bearer <admin_token>
```

### Response `200 OK`

```json
{
  "sessions": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "node_id": "uuid",
      "node_name": "us-east-1",
      "protocol": "wireguard",
      "tunnel_ip": "10.8.0.25",
      "status": "active",
      "bytes_sent": 1234567890,
      "bytes_received": 9876543210,
      "duration_seconds": 1800,
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 38000
}
```

---

## Verify Operator

Approve or reject an operator registration.

```http
POST /api/v1/admin/operator/:id/verify
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "status": "verified"
}
```

### Request Body

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `status` | string | Yes | `verified` or `rejected` |
| `reason` | string | No | Reason for rejection (if rejecting) |

### Response `200 OK`

```json
{
  "operator": {
    "id": "uuid",
    "status": "verified",
    "verified_at": "2024-01-15T10:30:00Z"
  },
  "message": "Operator verified successfully"
}
```

---

## Error Responses

All admin endpoints return these common errors:

| Status | Error | Description |
|--------|-------|-------------|
| `401` | `UNAUTHORIZED` | Missing or invalid token |
| `403` | `FORBIDDEN` | User is not an admin |
| `404` | `NOT_FOUND` | Resource not found |
| `500` | `SERVER_ERROR` | Internal server error |
