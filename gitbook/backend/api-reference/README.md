# 📖 API Reference

Complete REST API documentation for the Aureo VPN backend.

---

## Base URL

```
/api/v1
```

## Authentication

Most endpoints require a JWT bearer token:

```
Authorization: Bearer <access_token>
```

Access tokens are obtained via the [Authentication](authentication.md) endpoints and expire after 15 minutes. Use the refresh endpoint to obtain new access tokens.

## Endpoints Overview

| Section | Description |
|---------|-------------|
| [Authentication](authentication.md) | Register, login, refresh tokens |
| [Nodes](nodes.md) | List and discover VPN nodes |
| [Sessions](sessions.md) | Create and manage VPN sessions |
| [Operator](operator.md) | Node operator registration and management |
| [Admin](admin.md) | Administrative endpoints |
| [Health & Metrics](health.md) | Health checks and Prometheus metrics |

## Error Responses

All errors follow a consistent format:

```json
{
  "error": "error_code",
  "message": "Human-readable description"
}
```

Common HTTP status codes:
- `400` — Bad request (validation error)
- `401` — Unauthorized (missing or invalid token)
- `403` — Forbidden (insufficient permissions)
- `404` — Resource not found
- `429` — Too many requests (rate limited)
- `500` — Internal server error
