# 🔑 Authentication Flow

Aureo VPN uses JWT-based authentication with short-lived access tokens and long-lived refresh tokens. The implementation lives in `pkg/auth/` (application-level) and `pkg/security/jwt.go` (hardened security layer).

---

## JWT Configuration

| Parameter | Value |
|---|---|
| **Algorithm** | HS256 (HMAC-SHA256) |
| **Issuer** | `aureo-vpn` |
| **Audience** | `aureo-vpn-api` |
| **Access Token Lifetime** | 15 minutes (max 1 hour) |
| **Refresh Token Lifetime** | 7 days (max 30 days) |
| **Minimum Secret Length** | 32 bytes (256 bits) |
| **Clock Skew Tolerance** | 30 seconds |

---

## Token Claims

```go
// Application-level claims (pkg/auth/jwt.go)
type Claims struct {
    UserID    uuid.UUID `json:"user_id"`
    Email     string    `json:"email"`
    Username  string    `json:"username"`
    IsAdmin   bool      `json:"is_admin"`
    TokenType string    `json:"token_type"` // "access" or "refresh"
    jwt.RegisteredClaims
}

// Hardened claims (pkg/security/jwt.go)
type SecureClaims struct {
    jwt.RegisteredClaims
    UserID      string    `json:"uid"`
    Email       string    `json:"email,omitempty"`
    Username    string    `json:"username,omitempty"`
    IsAdmin     bool      `json:"admin,omitempty"`
    TokenType   TokenType `json:"type"`
    TokenFamily string    `json:"fam,omitempty"`  // Refresh token rotation chain
    TokenID     string    `json:"jti"`            // Unique token ID
    Fingerprint string    `json:"fp,omitempty"`   // Device fingerprint hash
    IPHash      string    `json:"iph,omitempty"`  // Hashed IP (anomaly detection)
}
```

---

## Authentication Flow

```
┌──────────┐                          ┌──────────────┐                     ┌────────────┐
│  Client  │                          │  API Gateway │                     │  Database  │
└────┬─────┘                          └──────┬───────┘                     └─────┬──────┘
     │                                       │                                   │
     │  1. POST /api/v1/auth/register        │                                   │
     │  { email, password, username }        │                                   │
     │──────────────────────────────────────>│                                   │
     │                                       │  2. Validate input                │
     │                                       │     ValidateEmail()               │
     │                                       │     ValidateUsername()             │
     │                                       │     ValidatePassword()            │
     │                                       │                                   │
     │                                       │  3. Hash password (bcrypt)        │
     │                                       │                                   │
     │                                       │  4. INSERT user                   │
     │                                       │──────────────────────────────────>│
     │                                       │                                   │
     │                                       │  5. GenerateTokenPair()           │
     │                                       │     - Access token (15min)        │
     │                                       │     - Refresh token (7d)          │
     │                                       │                                   │
     │  6. 201 { access_token,               │                                   │
     │           refresh_token, user }       │                                   │
     │<──────────────────────────────────────│                                   │
     │                                       │                                   │
     │  ════════════════════════════════════ │                                   │
     │                                       │                                   │
     │  7. POST /api/v1/auth/login           │                                   │
     │  { email, password }                  │                                   │
     │──────────────────────────────────────>│                                   │
     │                                       │  8. Find user by email            │
     │                                       │──────────────────────────────────>│
     │                                       │                                   │
     │                                       │  9. Verify password               │
     │                                       │     Check IsActive                │
     │                                       │                                   │
     │                                       │  10. GenerateTokenPair()          │
     │                                       │                                   │
     │  11. 200 { access_token,              │                                   │
     │            refresh_token, user }      │                                   │
     │<──────────────────────────────────────│                                   │
     │                                       │                                   │
     │  ════════════════════════════════════ │                                   │
     │                                       │                                   │
     │  12. GET /api/v1/user/profile         │                                   │
     │  Authorization: Bearer <access_token> │                                   │
     │──────────────────────────────────────>│                                   │
     │                                       │  13. AuthMiddleware               │
     │                                       │      Extract Bearer token         │
     │                                       │      VerifyToken()                │
     │                                       │      Check signing method         │
     │                                       │      Set user_id in Locals        │
     │                                       │                                   │
     │                                       │  14. Handler executes             │
     │                                       │──────────────────────────────────>│
     │                                       │                                   │
     │  15. 200 { user profile }             │                                   │
     │<──────────────────────────────────────│                                   │
     │                                       │                                   │
     │  ════════════════════════════════════ │                                   │
     │                                       │                                   │
     │  16. POST /api/v1/auth/refresh        │                                   │
     │  { refresh_token }                    │                                   │
     │──────────────────────────────────────>│                                   │
     │                                       │  17. VerifyToken(refresh)          │
     │                                       │      Check TokenType == "refresh" │
     │                                       │      GenerateAccessToken()        │
     │                                       │                                   │
     │  18. 200 { access_token }             │                                   │
     │<──────────────────────────────────────│                                   │
     │                                       │                                   │
```

---

## Middleware Stack

Every HTTP request passes through the following middleware chain in order:

```
Request
  │
  ├─> Recover          Panic recovery, stack trace in dev mode
  │
  ├─> RequestID        Generates X-Request-ID header (UUID)
  │
  ├─> Compress         Gzip compression (LevelBestSpeed)
  │
  ├─> Logger           Request logging (time, status, latency, IP, method, path)
  │
  ├─> CORS             Cross-origin resource sharing (configurable origins)
  │
  ├─> Metrics          Prometheus HTTP request metrics
  │
  ├─> RateLimit        Per-IP rate limiting (default: 100 req/min)
  │
  ├─> [AuthMiddleware] JWT verification (protected routes only)
  │   │
  │   ├─ Extract "Bearer" token from Authorization header
  │   ├─ Parse and verify JWT with HMAC-SHA256
  │   ├─ Check token expiry, issuer, signing method
  │   └─ Set user_id, email, username, is_admin in fiber.Locals
  │
  ├─> [AdminMiddleware] Admin role check (admin routes only)
  │   │
  │   └─ Verify is_admin == true from Locals
  │
  └─> Handler          Route handler executes
```

---

## Token Verification Details

```go
func (t *TokenService) VerifyToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{},
        func(token *jwt.Token) (interface{}, error) {
            // Verify signing method is HMAC (prevents algorithm confusion)
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, ErrInvalidToken
            }
            return t.secretKey, nil
        },
    )
    // ...
}
```

**Security checks performed:**

1. Signing method must be `*jwt.SigningMethodHMAC` (prevents `alg: none` and RSA confusion attacks)
2. Token must not be expired (`exp` claim)
3. Token must be valid after `nbf` (not before) claim
4. Claims must parse into the expected struct
5. `token.Valid` must be `true`

---

## Token Refresh (Rotation)

The hardened `SecureJWTService` implements refresh token rotation per OWASP recommendations:

1. Each refresh token belongs to a **token family** (shared UUID across rotations)
2. When a refresh token is used, it is immediately **revoked** (one-time use)
3. A new access + refresh pair is issued with the **same family ID**
4. If a revoked refresh token is reused (token theft detection), the **entire family is revoked**
5. Revoked tokens and families are stored in in-memory blacklists with TTL-based cleanup

---

## Error Responses

Authentication errors never reveal whether the email exists or the password was wrong:

```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid email or password."
  }
}
```

Registration failures use a generic message:

```json
{
  "error": {
    "code": "CONFLICT",
    "message": "Registration failed. Email or username may already exist."
  }
}
```
