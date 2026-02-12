# 🔑 Authentication

User registration, login, and token management.

---

## Register User

Create a new user account.

```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "username": "johndoe",
  "password": "SecurePassword123"
}
```

### Response `201 Created`

```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "johndoe",
    "subscription_tier": "free",
    "created_at": "2024-01-15T10:00:00Z"
  },
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 900
}
```

### Validation Rules

| Field | Rules |
|-------|-------|
| `email` | Valid email format, unique |
| `username` | 3-30 characters, alphanumeric + underscores, unique |
| `password` | Minimum 8 characters |

### Error Responses

| Status | Error | Description |
|--------|-------|-------------|
| `400` | `VALIDATION_ERROR` | Invalid email, username, or password format |
| `409` | `CONFLICT` | Email or username already registered |

---

## Login

Authenticate an existing user.

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePassword123"
}
```

### Response `200 OK`

```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "johndoe",
    "subscription_tier": "free"
  },
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 900
}
```

### Error Responses

| Status | Error | Description |
|--------|-------|-------------|
| `400` | `VALIDATION_ERROR` | Missing email or password |
| `401` | `UNAUTHORIZED` | Invalid credentials |

---

## Refresh Token

Exchange a refresh token for a new access token.

```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### Response `200 OK`

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 900
}
```

### Error Responses

| Status | Error | Description |
|--------|-------|-------------|
| `400` | `VALIDATION_ERROR` | Missing refresh token |
| `401` | `UNAUTHORIZED` | Invalid or expired refresh token |

---

## Token Lifecycle

```
Registration/Login
       │
       ▼
┌──────────────┐
│ access_token │──── Valid for 15 minutes
│ (JWT)        │
└──────┬───────┘
       │ Expires
       ▼
┌──────────────┐
│refresh_token │──── Valid for 7 days
│ (JWT)        │
└──────┬───────┘
       │ POST /api/v1/auth/refresh
       ▼
┌──────────────┐
│ New tokens   │──── New access + refresh pair
└──────────────┘
```

{% hint style="info" %}
Access tokens expire after **15 minutes**. Refresh tokens expire after **7 days**. Always use the refresh endpoint to obtain new tokens before the access token expires.
{% endhint %}

---

## Usage Example (cURL)

```bash
# 1. Register
RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "johndoe",
    "password": "SecurePassword123"
  }')

# 2. Extract token
TOKEN=$(echo $RESPONSE | jq -r '.access_token')

# 3. Use token for authenticated requests
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/nodes
```
