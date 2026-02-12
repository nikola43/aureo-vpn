# 🛡️ Security Model

Aureo VPN implements defense-in-depth security with input validation, attack detection, privacy filtering, and hardened authentication. All security code lives in `pkg/security/`.

---

## Input Validation

All user-supplied input is validated before processing. The default validator runs in **strict mode** (SQL injection and XSS detection enabled).

| Input Type | Rules |
|---|---|
| **Email** | Max 254 chars. Parsed with `net/mail.ParseAddress()`. Must contain `@` and domain with valid TLD |
| **Username** | 3-64 chars. Must match `^[a-zA-Z][a-zA-Z0-9_-]{2,63}$`. Starts with letter |
| **Password** | 6-128 chars (default). Common password blocklist. Max 5 repeating chars. High-security mode: 16+ chars, upper+lower+digit+special required, max 2 repeating |
| **UUID** | Must match UUID v1-v5 regex: `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5]...` |
| **IP Address** | Validated with `net.ParseIP()` |
| **Port** | Integer 1-65535 |
| **Hostname** | Max 253 chars. RFC 952/1123 compliant: `^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}...` |
| **Country Code** | Exactly 2 chars. Must be in ISO 3166-1 alpha-2 whitelist (35+ countries) |
| **Protocol** | Must be in whitelist: `wireguard`, `openvpn`, `ikev2`, `ipsec` |
| **Crypto Currency** | Must be in whitelist: `ethereum`, `bitcoin`, `litecoin`, `monero` |
| **Wallet Address** | Format validated per chain: ETH (`0x` + 40 hex + EIP-55 checksum), BTC (Legacy/SegWit/Bech32), LTC (L/M/3 prefix or ltc1 Bech32) |
| **Max Input Length** | General text: 10,000 chars. Null bytes stripped. Control chars stripped (except `\n`, `\t`). Invalid UTF-8 replaced |

---

## Attack Detection

### SQL Injection Patterns

```go
sqlInjectionPatterns = []*regexp.Regexp{
    // SQL keywords
    regexp.MustCompile(`(?i)(\b(select|insert|update|delete|drop|union|alter|create|truncate|exec|execute)\b)`),
    // SQL special chars and functions
    regexp.MustCompile(`(?i)(--|\;|\/\*|\*\/|@@|char\(|nchar\(|varchar\(|nvarchar\()`),
    // SQL system procs and hex literals
    regexp.MustCompile(`(?i)(xp_|sp_|0x[0-9a-f]+)`),
}
```

### XSS Patterns

```go
xssPatterns = []*regexp.Regexp{
    regexp.MustCompile(`(?i)<script[^>]*>.*</script>`),
    regexp.MustCompile(`(?i)<[^>]*(on\w+\s*=)`),
    regexp.MustCompile(`(?i)(javascript:|data:|vbscript:)`),
    regexp.MustCompile(`(?i)<\s*(iframe|object|embed|form|input|button|textarea|select)`),
}
```

When detected in strict mode, the request is rejected with `ErrSQLInjectionDetected` or `ErrXSSDetected`.

---

## Privacy Filter

Source: `pkg/security/privacy.go`

### IP Anonymization

Client IPs are never stored directly. A salted SHA-256 hash is used for correlation:

```go
func (pf *PrivacyFilter) AnonymizeIP(ip string) string {
    data := append(pf.config.AnonymizationSalt, []byte(ip)...)
    hash := sha256.Sum256(data)
    return "ip:" + hex.EncodeToString(hash[:8])
}
// Output: "ip:a1b2c3d4e5f6a7b8"
```

### Sensitive Data Redaction

Log messages are automatically sanitized to remove:

| Pattern | Replacement |
|---|---|
| IPv4 addresses (`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`) | `[ip-redacted]` |
| IPv6 addresses | `[ipv6-redacted]` |
| Email addresses | `[email-redacted]` |
| JWT tokens (`eyJ...`) | `[token-redacted]` |
| API keys / access tokens | `[credential-redacted]` |
| Private keys (PEM blocks) | `[private-key-redacted]` |
| Passwords in URLs | `[password-redacted]` |
| Credit card numbers | `[card-redacted]` |
| Ethereum addresses (`0x...`) | `[eth-address-redacted]` |

### Log Sanitization

```go
var privacyFilter = security.NewPrivacyFilter(nil)

// In handlers:
log.Printf("[AUTH] Registration failed: %v",
    privacyFilter.SanitizeLogMessage(err.Error()))
```

---

## Data Retention Policy (GDPR)

Default retention periods:

| Data Type | Retention | Notes |
|---|---|---|
| Active sessions | While active | No time limit |
| Completed sessions | 24 hours | Minimal metadata only |
| Security logs | 90 days | Sanitized, no PII |
| Access logs | 7 days | Anonymized IPs |
| Error logs | 30 days | Sanitized |
| Inactive users | 1 year | From last activity |
| Deleted users | 30 days | After deletion request |
| Payment records | 7 years | Legal compliance requirement |

### Warrant Canary

The system includes a `WarrantCanary` struct for transparency:

```go
type WarrantCanary struct {
    LastUpdated     time.Time
    Statement       string  // "As of the date above, Aureo VPN has not received..."
    NoGagOrders     bool
    NoNSLs          bool
    NoCourtOrders   bool
    NoGovernmentReq bool
    Signature       string  // SHA-256 HMAC signature
}
```

---

## Password Security

### Default Policy

```go
DefaultPasswordStrength() = &PasswordStrength{
    MinLength:      6,
    RequireUpper:   false,
    RequireLower:   false,
    RequireDigit:   false,
    RequireSpecial: false,
    DisallowCommon: true,   // Blocklist of common passwords
    MaxRepeating:   5,
}
```

### High-Security Policy

```go
HighSecurityPasswordPolicy() = &PasswordStrength{
    MinLength:      16,
    RequireUpper:   true,
    RequireLower:   true,
    RequireDigit:   true,
    RequireSpecial: true,
    DisallowCommon: true,
    MaxRepeating:   2,
}
```

### Common Password Blocklist

The following passwords are always rejected: `password`, `123456`, `123456789`, `qwerty`, `abc123`, `password1`, `password123`, `admin`, `letmein`, `welcome`, `monkey`, `dragon`, `master`, `login`, `trustno1`.

### Password Hashing

Passwords are hashed using **Argon2id** with secure defaults:

```
Algorithm:   Argon2id
Memory:      64 MiB (default) / 256 MiB (high-security)
Iterations:  4 (default) / 6 (high-security)
Parallelism: 4 threads
Salt:        32 bytes (256-bit, random)
Key Length:  32 bytes (256-bit)

Format: $argon2id$v=19$m=65536,t=4,p=4$<salt>$<hash>
```

Verification uses **constant-time comparison** via `subtle.ConstantTimeCompare` to prevent timing attacks.

---

## Authentication Security

### Bearer Token Extraction

The auth middleware extracts the JWT from the `Authorization: Bearer <token>` header. The token is validated with:

1. **Signing method verification** -- Only `*jwt.SigningMethodHMAC` accepted (prevents `alg:none` attacks)
2. **Expiry check** -- `exp` claim must be in the future
3. **Not-before check** -- `nbf` claim must be in the past
4. **Issuer validation** -- Must match `aureo-vpn`

### HMAC-SHA256 Verification

Token signing uses `jwt.SigningMethodHS256` with a shared secret (minimum 32 bytes / 256 bits):

```go
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
signedToken, err := token.SignedString(t.secretKey)
```

### Mutex-Protected Refresh

The hardened JWT service implements refresh token rotation:

- Each refresh token belongs to a **token family** (UUID chain)
- Used refresh tokens are **immediately revoked** (one-time use)
- Token and family blacklists are protected with `sync.RWMutex`
- Blacklist entries auto-expire based on TTL

### Admin Middleware

Admin-only routes require `is_admin == true` in the JWT claims. The `AdminOnlyMiddleware()` checks `c.Locals("is_admin")` and returns 403 Forbidden if not set.

---

## Cryptographic Primitives

Source: `pkg/security/crypto.go`

| Primitive | Implementation | Use Case |
|---|---|---|
| AEAD Encryption | AES-256-GCM | Config file encryption |
| AEAD Encryption | ChaCha20-Poly1305 | Alternative AEAD |
| AEAD Encryption | XChaCha20-Poly1305 | Extended nonce (192-bit) |
| Key Derivation | Argon2id | Password hashing |
| Key Derivation | HKDF-SHA512 | Key expansion with context |
| Key Exchange | X25519 (ECDH) | Hybrid key exchange |
| Random Generation | `crypto/rand` (`/dev/urandom`) | Keys, nonces, salts |
| Secure Memory | Multi-pass wipe + `runtime.KeepAlive` | Key material cleanup |
| Key Rotation | Current + previous key with grace period | Zero-downtime rotation |
