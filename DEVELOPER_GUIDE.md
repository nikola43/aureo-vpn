# Aureo VPN Developer Guide

This guide provides in-depth technical documentation for developers working on Aureo VPN. It covers system internals, code architecture, data flows, and implementation details.

## Table of Contents

- [System Architecture Deep Dive](#system-architecture-deep-dive)
- [Database Schema](#database-schema)
- [Authentication System](#authentication-system)
- [VPN Protocol Implementation](#vpn-protocol-implementation)
- [P2P Network Layer](#p2p-network-layer)
- [Session Management](#session-management)
- [Reward System](#reward-system)
- [Blockchain Integration](#blockchain-integration)
- [API Implementation](#api-implementation)
- [Security Implementation](#security-implementation)
- [Extending the Platform](#extending-the-platform)

---

## System Architecture Deep Dive

### Service Communication

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        SERVICE COMMUNICATION DIAGRAM                         │
└─────────────────────────────────────────────────────────────────────────────┘

                              ┌─────────────────┐
                              │     Client      │
                              │  (Desktop/Web)  │
                              └────────┬────────┘
                                       │
                                       │ HTTPS/REST
                                       ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│                              API GATEWAY                                     │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                          Fiber HTTP Server                           │   │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐        │   │
│  │  │  Auth   │ │  Rate   │ │  CORS   │ │ Logger  │ │ Recover │        │   │
│  │  │Middlware│ │ Limiter │ │         │ │         │ │         │        │   │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘        │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│  ┌─────────────┬─────────────┬─────┴─────┬─────────────┬─────────────┐     │
│  │   Auth      │   Node      │  Session  │  Operator   │   Admin     │     │
│  │  Handler    │  Handler    │  Handler  │  Handler    │  Handler    │     │
│  └──────┬──────┴──────┬──────┴─────┬─────┴──────┬──────┴──────┬──────┘     │
│         │             │            │            │             │            │
└─────────┼─────────────┼────────────┼────────────┼─────────────┼────────────┘
          │             │            │            │             │
          ▼             ▼            ▼            ▼             ▼
     ┌─────────────────────────────────────────────────────────────────┐
     │                         SERVICE LAYER                           │
     │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐   │
     │  │  Auth   │ │  Node   │ │ Session │ │Operator │ │ Reward  │   │
     │  │ Service │ │ Service │ │ Service │ │ Service │ │ Service │   │
     │  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘   │
     └───────┼───────────┼───────────┼───────────┼───────────┼────────┘
             │           │           │           │           │
             ▼           ▼           ▼           ▼           ▼
     ┌─────────────────────────────────────────────────────────────────┐
     │                       DATA LAYER                                │
     │  ┌───────────────────┐      ┌───────────────────┐              │
     │  │      SQLite       │      │      Redis        │              │
     │  │    (GORM ORM)     │      │   (Rate Limit)    │              │
     │  │                   │      │   (Sessions)      │              │
     │  │  • Users          │      │   (Cache)         │              │
     │  │  • Nodes          │      │                   │              │
     │  │  • Sessions       │      └───────────────────┘              │
     │  │  • Operators      │                                         │
     │  │  • Earnings       │      ┌───────────────────┐              │
     │  │  • Payouts        │      │    P2P Network    │              │
     │  │  • Configs        │      │     (libp2p)      │              │
     │  └───────────────────┘      └───────────────────┘              │
     └─────────────────────────────────────────────────────────────────┘
```

### Request Lifecycle

```go
// Example: POST /api/v1/sessions (Create VPN Session)

// 1. Request enters Fiber server
app.Post("/api/v1/sessions", middleware.Auth(), handlers.CreateSession)

// 2. Auth middleware validates JWT
func Auth() fiber.Handler {
    return func(c *fiber.Ctx) error {
        token := extractToken(c)
        claims, err := auth.ValidateToken(token)
        if err != nil {
            return c.Status(401).JSON(...)
        }
        c.Locals("user_id", claims.UserID)
        return c.Next()
    }
}

// 3. Handler processes request
func CreateSession(c *fiber.Ctx) error {
    var req CreateSessionRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(...)
    }

    userID := c.Locals("user_id").(uuid.UUID)

    // 4. Service layer handles business logic
    session, config, err := sessionService.Create(userID, req.NodeID, req.Protocol)

    // 5. Database operations via GORM
    // 6. WireGuard peer added
    // 7. Response returned
    return c.Status(201).JSON(...)
}
```

---

## Database Schema

The platform uses **SQLite** as an embedded database, managed through GORM ORM. The database file is created automatically on first run.

### Entity Relationship Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                       ENTITY RELATIONSHIP DIAGRAM                            │
│                              (SQLite Database)                               │
└─────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────┐         ┌──────────────────────┐
│        User          │         │     NodeOperator     │
├──────────────────────┤         ├──────────────────────┤
│ id (PK, UUID)        │         │ id (PK, UUID)        │
│ email (UNIQUE)       │◄───────▶│ user_id (FK)         │
│ username (UNIQUE)    │    1:1  │ wallet_address       │
│ password_hash        │         │ wallet_currency      │
│ subscription_tier    │         │ status               │
│ is_admin             │         │ reputation_score     │
│ is_active            │         │ total_earned_usd     │
│ two_factor_enabled   │         │ pending_payout_usd   │
│ data_used_gb         │         │ stake_amount_usd     │
│ created_at           │         │ created_at           │
│ updated_at           │         │ verified_at          │
└──────────┬───────────┘         └──────────┬───────────┘
           │                                │
           │ 1:N                            │ 1:N
           ▼                                ▼
┌──────────────────────┐         ┌──────────────────────┐
│       Session        │         │       VPNNode        │
├──────────────────────┤         ├──────────────────────┤
│ id (PK, UUID)        │◄────────│ id (PK, UUID)        │
│ user_id (FK)         │    N:1  │ operator_id (FK)     │
│ node_id (FK)         │────────▶│ name                 │
│ protocol             │         │ hostname             │
│ tunnel_ip            │         │ ip_address           │
│ status               │         │ country, city        │
│ bytes_sent           │         │ latitude, longitude  │
│ bytes_received       │         │ wireguard_port       │
│ client_public_key    │         │ wireguard_public_key │
│ server_public_key    │         │ wireguard_private_key│
│ connected_at         │         │ openvpn_port         │
│ disconnected_at      │         │ load_score           │
│ latency_ms           │         │ max_connections      │
│ packet_loss_percent  │         │ status               │
│ enable_kill_switch   │         │ last_heartbeat       │
│ enable_dns_protection│         │ uptime_percent       │
│ next_hop_node_id     │         │ total_earned_usd     │
│ device_type          │         │ is_operator_owned    │
│ os_type              │         │ created_at           │
└──────────────────────┘         └──────────────────────┘
                                            │
                                            │ 1:N
                                            ▼
┌──────────────────────┐         ┌──────────────────────┐
│   OperatorEarning    │         │    OperatorPayout    │
├──────────────────────┤         ├──────────────────────┤
│ id (PK, UUID)        │         │ id (PK, UUID)        │
│ operator_id (FK)     │◄────────│ operator_id (FK)     │
│ node_id (FK)         │    N:1  │ amount_usd           │
│ session_id (FK)      │         │ amount_crypto        │
│ bandwidth_gb         │         │ currency             │
│ amount_usd           │         │ wallet_address       │
│ quality_score        │         │ transaction_hash     │
│ tier_at_time         │         │ status               │
│ status               │         │ fee_usd              │
│ confirmed_at         │         │ exchange_rate        │
│ created_at           │         │ created_at           │
└──────────────────────┘         │ processed_at         │
                                 └──────────────────────┘

┌──────────────────────┐
│       Config         │
├──────────────────────┤
│ id (PK, UUID)        │
│ user_id (FK)         │
│ node_id (FK)         │
│ name                 │
│ protocol             │
│ config_data          │
│ public_key           │
│ private_key_encrypted│
│ dns_servers          │
│ allowed_ips          │
│ mtu                  │
│ keepalive            │
│ expires_at           │
│ created_at           │
└──────────────────────┘
```

### Model Definitions

```go
// pkg/models/user.go
type User struct {
    ID               string         `gorm:"primaryKey"` // UUID stored as string in SQLite
    Email            string         `gorm:"uniqueIndex;not null"`
    Username         string         `gorm:"uniqueIndex;not null"`
    PasswordHash     string         `gorm:"not null" json:"-"`
    SubscriptionTier string         `gorm:"default:'free'"` // free, basic, premium
    IsAdmin          bool           `gorm:"default:false"`
    IsActive         bool           `gorm:"default:true"`
    TwoFactorEnabled bool           `gorm:"default:false"`
    TwoFactorSecret  string         `json:"-"`
    DataUsedGB       float64        `gorm:"default:0"`
    CreatedAt        time.Time
    UpdatedAt        time.Time
    DeletedAt        gorm.DeletedAt `gorm:"index"`
}

// pkg/models/vpn_node.go
type VPNNode struct {
    ID                   string     `gorm:"primaryKey"` // UUID stored as string in SQLite
    OperatorID           *string    `gorm:"index"`
    Name                 string     `gorm:"not null"`
    Hostname             string     `gorm:"not null"`
    IPAddress            string     `gorm:"not null"`
    Country              string     `gorm:"index"`
    City                 string
    Latitude             float64
    Longitude            float64
    WireGuardPort        int        `gorm:"default:51820"`
    WireGuardPublicKey   string
    WireGuardPrivateKey  string     `json:"-"` // Encrypted at rest
    OpenVPNPort          int        `gorm:"default:1194"`
    LoadScore            int        `gorm:"default:0"` // 0-100
    MaxConnections       int        `gorm:"default:1000"`
    CurrentConnections   int        `gorm:"default:0"`
    Status               string     `gorm:"default:'offline'"` // online, offline, maintenance
    LastHeartbeat        time.Time
    UptimePercent        float64    `gorm:"default:0"`
    TotalEarnedUSD       float64    `gorm:"default:0"`
    IsOperatorOwned      bool       `gorm:"default:false"`
    Features             []string   `gorm:"type:text[]"` // multihop, obfuscation, etc.
    CreatedAt            time.Time
    UpdatedAt            time.Time
}

// CalculateLoadScore computes node load based on multiple factors
func (n *VPNNode) CalculateLoadScore(cpuPercent, memPercent float64) int {
    connectionLoad := float64(n.CurrentConnections) / float64(n.MaxConnections) * 100

    // Weighted average: connections 40%, CPU 30%, memory 30%
    score := connectionLoad*0.4 + cpuPercent*0.3 + memPercent*0.3

    if score > 100 {
        return 100
    }
    return int(score)
}
```

### Database Setup

```go
// pkg/database/database.go
import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func NewDatabase(dbPath string) (*gorm.DB, error) {
    // SQLite connection with WAL mode for better concurrency
    db, err := gorm.Open(sqlite.Open(dbPath+"?_journal_mode=WAL&_busy_timeout=5000"), &gorm.Config{})
    if err != nil {
        return nil, err
    }

    // Auto-migrate all models
    return db, AutoMigrate(db)
}

func AutoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &models.User{},
        &models.VPNNode{},
        &models.Session{},
        &models.NodeOperator{},
        &models.OperatorEarning{},
        &models.OperatorPayout{},
        &models.Config{},
    )
}
```

---

## Authentication System

### JWT Token Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          JWT AUTHENTICATION FLOW                             │
└─────────────────────────────────────────────────────────────────────────────┘

    Client                          API Gateway                      Database
       │                                │                               │
       │  POST /auth/login              │                               │
       │  {email, password}             │                               │
       │───────────────────────────────▶│                               │
       │                                │  SELECT * FROM users          │
       │                                │  WHERE email = ?              │
       │                                │──────────────────────────────▶│
       │                                │                               │
       │                                │◀──────────────────────────────│
       │                                │       User record             │
       │                                │                               │
       │                                │  bcrypt.Compare(              │
       │                                │    password,                  │
       │                                │    user.PasswordHash          │
       │                                │  )                            │
       │                                │                               │
       │                                │  Generate JWT tokens:         │
       │                                │  • Access (15min)             │
       │                                │  • Refresh (7 days)           │
       │                                │                               │
       │  {access_token, refresh_token} │                               │
       │◀───────────────────────────────│                               │
       │                                │                               │
       │  GET /api/v1/nodes             │                               │
       │  Authorization: Bearer <token> │                               │
       │───────────────────────────────▶│                               │
       │                                │                               │
       │                                │  Validate JWT:                │
       │                                │  • Check signature            │
       │                                │  • Check expiration           │
       │                                │  • Check token type           │
       │                                │                               │
       │                                │  Extract claims:              │
       │                                │  • UserID                     │
       │                                │  • IsAdmin                    │
       │                                │                               │
       │  {nodes: [...]}                │                               │
       │◀───────────────────────────────│                               │
       │                                │                               │
```

### Implementation

```go
// pkg/auth/jwt.go

type JWTConfig struct {
    Secret          string
    AccessDuration  time.Duration
    RefreshDuration time.Duration
    Issuer          string
}

type Claims struct {
    UserID    uuid.UUID `json:"user_id"`
    Email     string    `json:"email"`
    Username  string    `json:"username"`
    IsAdmin   bool      `json:"is_admin"`
    TokenType string    `json:"token_type"` // "access" or "refresh"
    jwt.RegisteredClaims
}

type JWTService struct {
    config JWTConfig
}

func NewJWTService(config JWTConfig) *JWTService {
    return &JWTService{config: config}
}

// GenerateTokenPair creates both access and refresh tokens
func (s *JWTService) GenerateTokenPair(user *models.User) (*TokenPair, error) {
    accessToken, err := s.generateToken(user, "access", s.config.AccessDuration)
    if err != nil {
        return nil, err
    }

    refreshToken, err := s.generateToken(user, "refresh", s.config.RefreshDuration)
    if err != nil {
        return nil, err
    }

    return &TokenPair{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    int(s.config.AccessDuration.Seconds()),
    }, nil
}

func (s *JWTService) generateToken(user *models.User, tokenType string, duration time.Duration) (string, error) {
    claims := Claims{
        UserID:    user.ID,
        Email:     user.Email,
        Username:  user.Username,
        IsAdmin:   user.IsAdmin,
        TokenType: tokenType,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    s.config.Issuer,
            Subject:   user.ID.String(),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(s.config.Secret))
}

// ValidateToken verifies and parses a JWT token
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        // Validate signing method
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(s.config.Secret), nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, errors.New("invalid token")
}

// pkg/auth/password.go

type PasswordHasher struct {
    cost int
}

func NewPasswordHasher(cost int) *PasswordHasher {
    if cost < bcrypt.MinCost {
        cost = bcrypt.DefaultCost
    }
    return &PasswordHasher{cost: cost}
}

func (h *PasswordHasher) Hash(password string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
    return string(hash), err
}

func (h *PasswordHasher) Verify(password, hash string) bool {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
```

### Auth Middleware

```go
// internal/api/middleware/auth.go

func AuthMiddleware(jwtService *auth.JWTService) fiber.Handler {
    return func(c *fiber.Ctx) error {
        authHeader := c.Get("Authorization")
        if authHeader == "" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "missing authorization header",
            })
        }

        // Extract token from "Bearer <token>"
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "invalid authorization header format",
            })
        }

        claims, err := jwtService.ValidateToken(parts[1])
        if err != nil {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "invalid or expired token",
            })
        }

        // Ensure it's an access token, not refresh
        if claims.TokenType != "access" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "invalid token type",
            })
        }

        // Store claims in context for handlers
        c.Locals("user_id", claims.UserID)
        c.Locals("email", claims.Email)
        c.Locals("is_admin", claims.IsAdmin)

        return c.Next()
    }
}

// Admin-only middleware
func AdminMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        isAdmin, ok := c.Locals("is_admin").(bool)
        if !ok || !isAdmin {
            return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
                "error": "admin access required",
            })
        }
        return c.Next()
    }
}
```

---

## VPN Protocol Implementation

### WireGuard Implementation

```go
// pkg/protocols/wireguard/keys.go

import (
    "crypto/rand"
    "golang.org/x/crypto/curve25519"
    "encoding/base64"
)

// GenerateKeyPair creates a new WireGuard keypair
func GenerateKeyPair() (*KeyPair, error) {
    // Generate 32 random bytes for private key
    privateKey := make([]byte, 32)
    if _, err := rand.Read(privateKey); err != nil {
        return nil, err
    }

    // Apply WireGuard-specific bit clamping
    // This is required by the Curve25519 specification
    privateKey[0] &= 248   // Clear bottom 3 bits
    privateKey[31] &= 127  // Clear top bit
    privateKey[31] |= 64   // Set second-to-top bit

    // Derive public key using X25519
    var publicKey [32]byte
    var privKey [32]byte
    copy(privKey[:], privateKey)
    curve25519.ScalarBaseMult(&publicKey, &privKey)

    return &KeyPair{
        PrivateKey: base64.StdEncoding.EncodeToString(privateKey),
        PublicKey:  base64.StdEncoding.EncodeToString(publicKey[:]),
    }, nil
}

// GeneratePreSharedKey creates additional symmetric key for extra security
func GeneratePreSharedKey() (string, error) {
    key := make([]byte, 32)
    if _, err := rand.Read(key); err != nil {
        return "", err
    }
    return base64.StdEncoding.EncodeToString(key), nil
}

// pkg/protocols/wireguard/config.go

type ClientConfig struct {
    PrivateKey         string
    Address            string
    DNS                []string
    ServerPublicKey    string
    ServerEndpoint     string
    AllowedIPs         []string
    PersistentKeepalive int
    PreSharedKey       string
}

// GenerateClientConfig creates a WireGuard configuration file
func GenerateClientConfig(cfg ClientConfig) string {
    var builder strings.Builder

    builder.WriteString("[Interface]\n")
    builder.WriteString(fmt.Sprintf("PrivateKey = %s\n", cfg.PrivateKey))
    builder.WriteString(fmt.Sprintf("Address = %s\n", cfg.Address))

    if len(cfg.DNS) > 0 {
        builder.WriteString(fmt.Sprintf("DNS = %s\n", strings.Join(cfg.DNS, ", ")))
    }

    builder.WriteString("\n[Peer]\n")
    builder.WriteString(fmt.Sprintf("PublicKey = %s\n", cfg.ServerPublicKey))
    builder.WriteString(fmt.Sprintf("Endpoint = %s\n", cfg.ServerEndpoint))
    builder.WriteString(fmt.Sprintf("AllowedIPs = %s\n", strings.Join(cfg.AllowedIPs, ", ")))

    if cfg.PersistentKeepalive > 0 {
        builder.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", cfg.PersistentKeepalive))
    }

    if cfg.PreSharedKey != "" {
        builder.WriteString(fmt.Sprintf("PresharedKey = %s\n", cfg.PreSharedKey))
    }

    return builder.String()
}

// pkg/protocols/wireguard/interface.go

type WireGuardManager struct {
    interfaceName string
    listenPort    int
    privateKey    string
    subnet        *net.IPNet
    allocatedIPs  map[string]bool
    mu            sync.RWMutex
}

func NewWireGuardManager(interfaceName string, port int, privateKey string, subnet string) (*WireGuardManager, error) {
    _, ipnet, err := net.ParseCIDR(subnet)
    if err != nil {
        return nil, err
    }

    return &WireGuardManager{
        interfaceName: interfaceName,
        listenPort:    port,
        privateKey:    privateKey,
        subnet:        ipnet,
        allocatedIPs:  make(map[string]bool),
    }, nil
}

// CreateInterface sets up the WireGuard interface
func (m *WireGuardManager) CreateInterface() error {
    // Create interface
    if err := exec.Command("ip", "link", "add", "dev", m.interfaceName, "type", "wireguard").Run(); err != nil {
        // Interface might already exist
        if !strings.Contains(err.Error(), "exists") {
            return err
        }
    }

    // Write private key to temp file
    keyFile, err := os.CreateTemp("", "wg-key-*")
    if err != nil {
        return err
    }
    defer os.Remove(keyFile.Name())

    keyFile.WriteString(m.privateKey)
    keyFile.Close()

    // Configure interface
    if err := exec.Command("wg", "set", m.interfaceName,
        "listen-port", fmt.Sprintf("%d", m.listenPort),
        "private-key", keyFile.Name(),
    ).Run(); err != nil {
        return err
    }

    // Add IP address (gateway IP)
    gatewayIP := m.getGatewayIP()
    if err := exec.Command("ip", "addr", "add",
        fmt.Sprintf("%s/24", gatewayIP),
        "dev", m.interfaceName,
    ).Run(); err != nil {
        // Ignore if already exists
    }

    // Bring interface up
    return exec.Command("ip", "link", "set", m.interfaceName, "up").Run()
}

// AddPeer adds a client peer to the WireGuard interface
func (m *WireGuardManager) AddPeer(publicKey string, allowedIP string) error {
    return exec.Command("wg", "set", m.interfaceName,
        "peer", publicKey,
        "allowed-ips", allowedIP,
    ).Run()
}

// RemovePeer removes a client peer
func (m *WireGuardManager) RemovePeer(publicKey string) error {
    return exec.Command("wg", "set", m.interfaceName,
        "peer", publicKey,
        "remove",
    ).Run()
}

// AllocateIP returns an available IP from the subnet
func (m *WireGuardManager) AllocateIP() (string, error) {
    m.mu.Lock()
    defer m.mu.Unlock()

    // Start from .2 (skip .0 network, .1 gateway)
    ip := m.subnet.IP.Mask(m.subnet.Mask)
    ip = incrementIP(ip)  // .1 (gateway)
    ip = incrementIP(ip)  // .2 (first client)

    for m.subnet.Contains(ip) {
        ipStr := ip.String()
        if !m.allocatedIPs[ipStr] {
            m.allocatedIPs[ipStr] = true
            return ipStr, nil
        }
        ip = incrementIP(ip)
    }

    return "", errors.New("no available IPs in subnet")
}

// ReleaseIP marks an IP as available
func (m *WireGuardManager) ReleaseIP(ip string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    delete(m.allocatedIPs, ip)
}

// GetPeerStats retrieves traffic statistics for all peers
func (m *WireGuardManager) GetPeerStats() (map[string]*PeerStats, error) {
    output, err := exec.Command("wg", "show", m.interfaceName, "dump").Output()
    if err != nil {
        return nil, err
    }

    stats := make(map[string]*PeerStats)
    lines := strings.Split(string(output), "\n")

    // Skip first line (interface info)
    for _, line := range lines[1:] {
        fields := strings.Split(line, "\t")
        if len(fields) < 8 {
            continue
        }

        publicKey := fields[0]
        rxBytes, _ := strconv.ParseInt(fields[5], 10, 64)
        txBytes, _ := strconv.ParseInt(fields[6], 10, 64)
        lastHandshake, _ := strconv.ParseInt(fields[4], 10, 64)

        stats[publicKey] = &PeerStats{
            PublicKey:     publicKey,
            RxBytes:       rxBytes,
            TxBytes:       txBytes,
            LastHandshake: time.Unix(lastHandshake, 0),
        }
    }

    return stats, nil
}
```

### OpenVPN Implementation

```go
// pkg/protocols/openvpn/config.go

type ServerConfig struct {
    Port           int
    Protocol       string // udp or tcp
    Device         string // tun or tap
    Network        string // VPN network CIDR
    Cipher         string
    Auth           string
    TLSVersion     string
    MaxClients     int
    ClientToClient bool
}

func GenerateServerConfig(cfg ServerConfig) string {
    var builder strings.Builder

    builder.WriteString(fmt.Sprintf("port %d\n", cfg.Port))
    builder.WriteString(fmt.Sprintf("proto %s\n", cfg.Protocol))
    builder.WriteString(fmt.Sprintf("dev %s\n", cfg.Device))
    builder.WriteString(fmt.Sprintf("server %s\n", cfg.Network))
    builder.WriteString(fmt.Sprintf("cipher %s\n", cfg.Cipher))
    builder.WriteString(fmt.Sprintf("auth %s\n", cfg.Auth))
    builder.WriteString(fmt.Sprintf("tls-version-min %s\n", cfg.TLSVersion))
    builder.WriteString(fmt.Sprintf("max-clients %d\n", cfg.MaxClients))

    if cfg.ClientToClient {
        builder.WriteString("client-to-client\n")
    }

    builder.WriteString(`
keepalive 10 120
persist-key
persist-tun
status /var/log/openvpn/status.log
log-append /var/log/openvpn/openvpn.log
verb 3
`)

    return builder.String()
}

// pkg/protocols/openvpn/certificate.go

import (
    "crypto/rand"
    "crypto/rsa"
    "crypto/x509"
    "crypto/x509/pkix"
    "encoding/pem"
    "math/big"
    "time"
)

type CertificateBundle struct {
    CACert     string
    CAKey      string
    ServerCert string
    ServerKey  string
}

// GenerateCertificates creates a CA and server certificate
func GenerateCertificates(commonName string) (*CertificateBundle, error) {
    // Generate CA key
    caKey, err := rsa.GenerateKey(rand.Reader, 2048)
    if err != nil {
        return nil, err
    }

    // Create CA certificate
    caTemplate := &x509.Certificate{
        SerialNumber: big.NewInt(1),
        Subject: pkix.Name{
            Organization: []string{"Aureo VPN"},
            CommonName:   "Aureo VPN CA",
        },
        NotBefore:             time.Now(),
        NotAfter:              time.Now().AddDate(10, 0, 0), // 10 years
        IsCA:                  true,
        KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
        BasicConstraintsValid: true,
    }

    caCertDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
    if err != nil {
        return nil, err
    }

    // Generate server key
    serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
    if err != nil {
        return nil, err
    }

    // Create server certificate
    serverTemplate := &x509.Certificate{
        SerialNumber: big.NewInt(2),
        Subject: pkix.Name{
            Organization: []string{"Aureo VPN"},
            CommonName:   commonName,
        },
        NotBefore:   time.Now(),
        NotAfter:    time.Now().AddDate(1, 0, 0), // 1 year
        KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
        ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
    }

    caCert, _ := x509.ParseCertificate(caCertDER)
    serverCertDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caCert, &serverKey.PublicKey, caKey)
    if err != nil {
        return nil, err
    }

    return &CertificateBundle{
        CACert:     encodePEM("CERTIFICATE", caCertDER),
        CAKey:      encodePEM("RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(caKey)),
        ServerCert: encodePEM("CERTIFICATE", serverCertDER),
        ServerKey:  encodePEM("RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(serverKey)),
    }, nil
}

func encodePEM(blockType string, data []byte) string {
    block := &pem.Block{Type: blockType, Bytes: data}
    return string(pem.EncodeToMemory(block))
}
```

---

## P2P Network Layer

### Network Architecture

```go
// pkg/p2p/network.go

import (
    "github.com/libp2p/go-libp2p"
    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/peer"
    dht "github.com/libp2p/go-libp2p-kad-dht"
    pubsub "github.com/libp2p/go-libp2p-pubsub"
)

type P2PNetwork struct {
    host       host.Host
    dht        *dht.IpfsDHT
    pubsub     *pubsub.PubSub
    nodeInfo   *NodeInfo
    peers      map[peer.ID]*PeerInfo
    peersMu    sync.RWMutex
    topics     map[string]*pubsub.Topic
}

type Config struct {
    ListenAddrs     []string
    BootstrapPeers  []string
    MaxPeers        int
    MaxNodes        int
    PrivateKey      crypto.PrivKey
}

func NewP2PNetwork(ctx context.Context, cfg *Config) (*P2PNetwork, error) {
    // Create libp2p host
    h, err := libp2p.New(
        libp2p.ListenAddrStrings(cfg.ListenAddrs...),
        libp2p.Identity(cfg.PrivateKey),
        libp2p.EnableRelay(),
        libp2p.NATPortMap(),
    )
    if err != nil {
        return nil, err
    }

    // Create DHT for peer discovery
    kadDHT, err := dht.New(ctx, h, dht.Mode(dht.ModeServer))
    if err != nil {
        return nil, err
    }

    // Bootstrap DHT
    if err := kadDHT.Bootstrap(ctx); err != nil {
        return nil, err
    }

    // Connect to bootstrap peers
    for _, addrStr := range cfg.BootstrapPeers {
        addr, err := peer.AddrInfoFromString(addrStr)
        if err != nil {
            continue
        }
        h.Connect(ctx, *addr)
    }

    // Create pubsub for messaging
    ps, err := pubsub.NewGossipSub(ctx, h)
    if err != nil {
        return nil, err
    }

    return &P2PNetwork{
        host:   h,
        dht:    kadDHT,
        pubsub: ps,
        peers:  make(map[peer.ID]*PeerInfo),
        topics: make(map[string]*pubsub.Topic),
    }, nil
}

// JoinTopic subscribes to a pubsub topic
func (n *P2PNetwork) JoinTopic(topicName string) (*pubsub.Topic, *pubsub.Subscription, error) {
    topic, err := n.pubsub.Join(topicName)
    if err != nil {
        return nil, nil, err
    }

    sub, err := topic.Subscribe()
    if err != nil {
        return nil, nil, err
    }

    n.topics[topicName] = topic
    return topic, sub, nil
}

// pkg/p2p/messages.go

const (
    TopicNodeAnnounce  = "/aureo/nodes/announce/1.0.0"
    TopicNodeHeartbeat = "/aureo/nodes/heartbeat/1.0.0"
)

type MessageType int

const (
    MessageTypeAnnounce MessageType = iota
    MessageTypeHeartbeat
    MessageTypeLeave
)

type NodeInfo struct {
    NodeID         uuid.UUID   `json:"node_id"`
    PeerID         string      `json:"peer_id"`
    Addresses      []string    `json:"addresses"`
    Country        string      `json:"country"`
    City           string      `json:"city"`
    Latitude       float64     `json:"latitude"`
    Longitude      float64     `json:"longitude"`
    MaxConnections int         `json:"max_connections"`
    LoadScore      int         `json:"load_score"`
    BandwidthMbps  int         `json:"bandwidth_mbps"`
    Status         string      `json:"status"`
    LastHeartbeat  time.Time   `json:"last_heartbeat"`
    Protocols      []string    `json:"protocols"`
    Features       []string    `json:"features"`
    Signature      []byte      `json:"signature"`
}

type P2PMessage struct {
    Type      MessageType `json:"type"`
    NodeInfo  *NodeInfo   `json:"node_info,omitempty"`
    Timestamp time.Time   `json:"timestamp"`
    Signature []byte      `json:"signature"`
}

// Sign creates a cryptographic signature for the message
func (m *P2PMessage) Sign(privateKey crypto.PrivKey) error {
    data, err := json.Marshal(m)
    if err != nil {
        return err
    }

    m.Signature, err = privateKey.Sign(data)
    return err
}

// Verify checks the message signature
func (m *P2PMessage) Verify(publicKey crypto.PubKey) bool {
    sig := m.Signature
    m.Signature = nil

    data, err := json.Marshal(m)
    if err != nil {
        return false
    }

    m.Signature = sig
    ok, _ := publicKey.Verify(data, sig)
    return ok
}

// pkg/p2p/discovery.go

type NodeDiscovery struct {
    network   *P2PNetwork
    nodes     map[uuid.UUID]*NodeInfo
    nodesMu   sync.RWMutex
    callbacks []func(*NodeInfo)
}

func NewNodeDiscovery(network *P2PNetwork) *NodeDiscovery {
    return &NodeDiscovery{
        network: network,
        nodes:   make(map[uuid.UUID]*NodeInfo),
    }
}

// Start begins listening for node announcements
func (d *NodeDiscovery) Start(ctx context.Context) error {
    // Subscribe to announce topic
    _, announceSub, err := d.network.JoinTopic(TopicNodeAnnounce)
    if err != nil {
        return err
    }

    // Subscribe to heartbeat topic
    _, heartbeatSub, err := d.network.JoinTopic(TopicNodeHeartbeat)
    if err != nil {
        return err
    }

    // Handle announcements
    go d.handleMessages(ctx, announceSub)
    go d.handleMessages(ctx, heartbeatSub)

    // Cleanup stale nodes
    go d.cleanupStaleNodes(ctx)

    return nil
}

func (d *NodeDiscovery) handleMessages(ctx context.Context, sub *pubsub.Subscription) {
    for {
        msg, err := sub.Next(ctx)
        if err != nil {
            return
        }

        var p2pMsg P2PMessage
        if err := json.Unmarshal(msg.Data, &p2pMsg); err != nil {
            continue
        }

        switch p2pMsg.Type {
        case MessageTypeAnnounce:
            d.handleAnnounce(p2pMsg.NodeInfo)
        case MessageTypeHeartbeat:
            d.handleHeartbeat(p2pMsg.NodeInfo)
        case MessageTypeLeave:
            d.handleLeave(p2pMsg.NodeInfo)
        }
    }
}

func (d *NodeDiscovery) handleAnnounce(info *NodeInfo) {
    d.nodesMu.Lock()
    defer d.nodesMu.Unlock()

    info.LastHeartbeat = time.Now()
    d.nodes[info.NodeID] = info

    // Notify callbacks
    for _, cb := range d.callbacks {
        go cb(info)
    }
}

func (d *NodeDiscovery) cleanupStaleNodes(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            d.nodesMu.Lock()
            threshold := time.Now().Add(-2 * time.Minute)
            for id, node := range d.nodes {
                if node.LastHeartbeat.Before(threshold) {
                    delete(d.nodes, id)
                }
            }
            d.nodesMu.Unlock()
        }
    }
}

// GetNodes returns all known online nodes
func (d *NodeDiscovery) GetNodes() []*NodeInfo {
    d.nodesMu.RLock()
    defer d.nodesMu.RUnlock()

    nodes := make([]*NodeInfo, 0, len(d.nodes))
    for _, node := range d.nodes {
        if node.Status == "online" {
            nodes = append(nodes, node)
        }
    }
    return nodes
}
```

---

## Session Management

### Session Lifecycle

```go
// internal/node/service.go

type SessionService struct {
    db              *gorm.DB
    wgManager       *wireguard.WireGuardManager
    rewardService   *rewards.RewardService
    activeSessions  map[uuid.UUID]*SessionInfo
    sessionsMu      sync.RWMutex
}

type SessionInfo struct {
    Session          *models.Session
    LastTrafficCheck time.Time
    PendingBandwidth float64
    ClientPublicKey  string
}

// CreateSession establishes a new VPN session
func (s *SessionService) CreateSession(ctx context.Context, userID, nodeID uuid.UUID, protocol string) (*models.Session, string, error) {
    // Verify node exists and is online
    var node models.VPNNode
    if err := s.db.First(&node, nodeID).Error; err != nil {
        return nil, "", errors.New("node not found")
    }

    if node.Status != "online" {
        return nil, "", errors.New("node is not available")
    }

    // Check user session limit
    var activeCount int64
    s.db.Model(&models.Session{}).
        Where("user_id = ? AND status = ?", userID, "active").
        Count(&activeCount)

    if activeCount >= MaxSessionsPerUser {
        return nil, "", errors.New("maximum sessions reached")
    }

    // Generate client keypair
    clientKeyPair, err := wireguard.GenerateKeyPair()
    if err != nil {
        return nil, "", err
    }

    // Allocate tunnel IP
    tunnelIP, err := s.wgManager.AllocateIP()
    if err != nil {
        return nil, "", err
    }

    // Add peer to WireGuard
    if err := s.wgManager.AddPeer(clientKeyPair.PublicKey, tunnelIP+"/32"); err != nil {
        s.wgManager.ReleaseIP(tunnelIP)
        return nil, "", err
    }

    // Create session record
    session := &models.Session{
        UserID:           userID,
        NodeID:           nodeID,
        Protocol:         protocol,
        TunnelIP:         tunnelIP,
        ClientPublicKey:  clientKeyPair.PublicKey,
        ServerPublicKey:  node.WireGuardPublicKey,
        Status:           "active",
        ConnectedAt:      time.Now(),
    }

    if err := s.db.Create(session).Error; err != nil {
        s.wgManager.RemovePeer(clientKeyPair.PublicKey)
        s.wgManager.ReleaseIP(tunnelIP)
        return nil, "", err
    }

    // Store in active sessions
    s.sessionsMu.Lock()
    s.activeSessions[session.ID] = &SessionInfo{
        Session:          session,
        LastTrafficCheck: time.Now(),
        ClientPublicKey:  clientKeyPair.PublicKey,
    }
    s.sessionsMu.Unlock()

    // Generate client config
    config := wireguard.GenerateClientConfig(wireguard.ClientConfig{
        PrivateKey:          clientKeyPair.PrivateKey,
        Address:             tunnelIP + "/32",
        DNS:                 []string{"1.1.1.1", "8.8.8.8"},
        ServerPublicKey:     node.WireGuardPublicKey,
        ServerEndpoint:      fmt.Sprintf("%s:%d", node.IPAddress, node.WireGuardPort),
        AllowedIPs:          []string{"0.0.0.0/0"},
        PersistentKeepalive: 25,
    })

    // Increment node connections
    s.db.Model(&node).Update("current_connections", gorm.Expr("current_connections + 1"))

    return session, config, nil
}

// TerminateSession ends a VPN session
func (s *SessionService) TerminateSession(ctx context.Context, sessionID uuid.UUID) error {
    s.sessionsMu.Lock()
    info, exists := s.activeSessions[sessionID]
    if !exists {
        s.sessionsMu.Unlock()
        return errors.New("session not found")
    }
    delete(s.activeSessions, sessionID)
    s.sessionsMu.Unlock()

    // Remove WireGuard peer
    s.wgManager.RemovePeer(info.ClientPublicKey)
    s.wgManager.ReleaseIP(info.Session.TunnelIP)

    // Flush pending earnings
    if info.PendingBandwidth > 0 {
        s.flushEarnings(info)
    }

    // Update session record
    now := time.Now()
    duration := now.Sub(info.Session.ConnectedAt)

    s.db.Model(&models.Session{}).
        Where("id = ?", sessionID).
        Updates(map[string]interface{}{
            "status":          "disconnected",
            "disconnected_at": now,
        })

    // Decrement node connections
    s.db.Model(&models.VPNNode{}).
        Where("id = ?", info.Session.NodeID).
        Update("current_connections", gorm.Expr("current_connections - 1"))

    return nil
}

// MonitorSessions runs in background to track active sessions
func (s *SessionService) MonitorSessions(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    flushTicker := time.NewTicker(15 * time.Minute)
    defer ticker.Stop()
    defer flushTicker.Stop()

    for {
        select {
        case <-ctx.Done():
            return

        case <-ticker.C:
            s.checkSessionHealth()

        case <-flushTicker.C:
            s.flushAllEarnings()
        }
    }
}

func (s *SessionService) checkSessionHealth() {
    // Get peer stats from WireGuard
    stats, err := s.wgManager.GetPeerStats()
    if err != nil {
        return
    }

    s.sessionsMu.Lock()
    defer s.sessionsMu.Unlock()

    for _, info := range s.activeSessions {
        peerStats, ok := stats[info.ClientPublicKey]
        if !ok {
            continue
        }

        // Calculate bandwidth delta
        now := time.Now()
        duration := now.Sub(info.LastTrafficCheck).Seconds()

        bytesTotal := peerStats.RxBytes + peerStats.TxBytes
        prevBytes := info.Session.BytesSent + info.Session.BytesReceived
        bytesDelta := bytesTotal - prevBytes

        if bytesDelta > 0 && duration > 0 {
            bandwidthGB := float64(bytesDelta) / (1024 * 1024 * 1024)
            info.PendingBandwidth += bandwidthGB
        }

        // Update session record
        info.Session.BytesSent = peerStats.TxBytes
        info.Session.BytesReceived = peerStats.RxBytes
        info.LastTrafficCheck = now

        // Check for stale sessions (no handshake in 5 min)
        if time.Since(peerStats.LastHandshake) > 5*time.Minute {
            go s.TerminateSession(context.Background(), info.Session.ID)
        }
    }
}

func (s *SessionService) flushEarnings(info *SessionInfo) {
    if info.PendingBandwidth <= 0 {
        return
    }

    // Calculate session duration
    duration := time.Since(info.Session.ConnectedAt)

    // Record earnings via reward service
    s.rewardService.RecordEarning(
        info.Session.NodeID,
        info.Session.ID,
        info.PendingBandwidth,
        duration,
        100-int(info.Session.PacketLossPercent), // Quality score
    )

    info.PendingBandwidth = 0
}
```

---

## Reward System

### Earning Calculation

```go
// pkg/rewards/rewards.go

type RewardService struct {
    db    *gorm.DB
    tiers []RewardTier
}

type RewardTier struct {
    Name           string
    RatePerGB      float64
    MinUptime      float64
    MinReputation  int
    BonusMultiplier float64
}

var DefaultTiers = []RewardTier{
    {Name: "bronze", RatePerGB: 0.01, MinUptime: 50, MinReputation: 0, BonusMultiplier: 1.0},
    {Name: "silver", RatePerGB: 0.015, MinUptime: 80, MinReputation: 60, BonusMultiplier: 1.2},
    {Name: "gold", RatePerGB: 0.02, MinUptime: 90, MinReputation: 75, BonusMultiplier: 1.5},
    {Name: "platinum", RatePerGB: 0.03, MinUptime: 95, MinReputation: 90, BonusMultiplier: 2.0},
}

func NewRewardService(db *gorm.DB) *RewardService {
    return &RewardService{
        db:    db,
        tiers: DefaultTiers,
    }
}

// RecordEarning calculates and stores an earning record
func (s *RewardService) RecordEarning(nodeID, sessionID uuid.UUID, bandwidthGB float64, duration time.Duration, qualityScore int) error {
    // Get node and operator
    var node models.VPNNode
    if err := s.db.First(&node, nodeID).Error; err != nil {
        return err
    }

    if node.OperatorID == nil {
        return nil // Not an operator node
    }

    var operator models.NodeOperator
    if err := s.db.First(&operator, *node.OperatorID).Error; err != nil {
        return err
    }

    // Determine applicable tier
    tier := s.getEligibleTier(&operator, &node)

    // Calculate quality multiplier (0.5 to 1.5)
    qualityMultiplier := 0.5 + (float64(qualityScore) / 100.0)

    // Calculate duration bonus
    durationBonus := 1.0
    durationMinutes := duration.Minutes()
    if durationMinutes >= 180 {
        durationBonus = 1.2
    } else if durationMinutes >= 60 {
        durationBonus = 1.1
    }

    // Calculate earnings
    baseEarnings := bandwidthGB * tier.RatePerGB
    earnings := baseEarnings * qualityMultiplier * durationBonus * tier.BonusMultiplier

    // Create earning record
    earning := &models.OperatorEarning{
        OperatorID:   operator.ID,
        NodeID:       nodeID,
        SessionID:    sessionID,
        BandwidthGB:  bandwidthGB,
        AmountUSD:    earnings,
        QualityScore: qualityScore,
        TierAtTime:   tier.Name,
        Status:       "pending",
    }

    if err := s.db.Create(earning).Error; err != nil {
        return err
    }

    // Update operator pending payout
    s.db.Model(&operator).
        Update("pending_payout_usd", gorm.Expr("pending_payout_usd + ?", earnings))

    // Update node total earned
    s.db.Model(&node).
        Update("total_earned_usd", gorm.Expr("total_earned_usd + ?", earnings))

    return nil
}

func (s *RewardService) getEligibleTier(operator *models.NodeOperator, node *models.VPNNode) RewardTier {
    // Start from highest tier and work down
    for i := len(s.tiers) - 1; i >= 0; i-- {
        tier := s.tiers[i]

        if node.UptimePercent >= tier.MinUptime &&
            operator.ReputationScore >= tier.MinReputation {
            return tier
        }
    }

    return s.tiers[0] // Default to bronze
}

// ConfirmEarnings moves pending earnings to confirmed status
func (s *RewardService) ConfirmEarnings(ctx context.Context) error {
    // Find pending earnings older than 1 hour (verification period)
    threshold := time.Now().Add(-1 * time.Hour)

    return s.db.Model(&models.OperatorEarning{}).
        Where("status = ? AND created_at < ?", "pending", threshold).
        Update("status", "confirmed").
        Update("confirmed_at", time.Now()).
        Error
}

// ProcessPayouts handles payout requests
func (s *RewardService) ProcessPayouts(ctx context.Context, blockchain *blockchain.Service) error {
    // Find operators with confirmed earnings above threshold
    var operators []models.NodeOperator
    s.db.Where("pending_payout_usd >= ? AND status = ?", MinPayoutThreshold, "active").
        Find(&operators)

    for _, operator := range operators {
        if err := s.processSinglePayout(ctx, &operator, blockchain); err != nil {
            // Log error but continue with other operators
            continue
        }
    }

    return nil
}

func (s *RewardService) processSinglePayout(ctx context.Context, operator *models.NodeOperator, blockchain *blockchain.Service) error {
    // Create payout record
    payout := &models.OperatorPayout{
        OperatorID:    operator.ID,
        AmountUSD:     operator.PendingPayoutUSD,
        Currency:      operator.WalletCurrency,
        WalletAddress: operator.WalletAddress,
        Status:        "processing",
    }

    if err := s.db.Create(payout).Error; err != nil {
        return err
    }

    // Send blockchain transaction
    txHash, amountCrypto, err := blockchain.SendPayment(
        operator.WalletAddress,
        operator.WalletCurrency,
        operator.PendingPayoutUSD,
    )

    if err != nil {
        payout.Status = "failed"
        s.db.Save(payout)
        return err
    }

    // Update payout record
    payout.TransactionHash = txHash
    payout.AmountCrypto = amountCrypto
    payout.Status = "completed"
    payout.ProcessedAt = time.Now()
    s.db.Save(payout)

    // Reset operator pending payout
    s.db.Model(operator).Updates(map[string]interface{}{
        "pending_payout_usd": 0,
        "total_paid_usd":     gorm.Expr("total_paid_usd + ?", operator.PendingPayoutUSD),
    })

    // Mark earnings as paid
    s.db.Model(&models.OperatorEarning{}).
        Where("operator_id = ? AND status = ?", operator.ID, "confirmed").
        Update("status", "paid")

    return nil
}
```

---

## Blockchain Integration

### Ethereum Implementation

```go
// pkg/blockchain/ethereum.go

import (
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
)

type EthereumService struct {
    client     *ethclient.Client
    privateKey *ecdsa.PrivateKey
    address    common.Address
    chainID    *big.Int
}

type Config struct {
    RPCURL     string
    PrivateKey string // Hex encoded with 0x prefix
    ChainID    int64
}

func NewEthereumService(cfg Config) (*EthereumService, error) {
    client, err := ethclient.Dial(cfg.RPCURL)
    if err != nil {
        return nil, err
    }

    privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(cfg.PrivateKey, "0x"))
    if err != nil {
        return nil, err
    }

    publicKey := privateKey.Public().(*ecdsa.PublicKey)
    address := crypto.PubkeyToAddress(*publicKey)

    return &EthereumService{
        client:     client,
        privateKey: privateKey,
        address:    address,
        chainID:    big.NewInt(cfg.ChainID),
    }, nil
}

// SendETH sends ETH to a recipient
func (s *EthereumService) SendETH(ctx context.Context, toAddress string, amountUSD float64) (string, float64, error) {
    // Validate address
    if !common.IsHexAddress(toAddress) {
        return "", 0, errors.New("invalid ethereum address")
    }
    to := common.HexToAddress(toAddress)

    // Convert USD to ETH (in production, use price oracle)
    ethPrice := getETHPrice() // e.g., 2000.0
    amountETH := amountUSD / ethPrice

    // Convert to Wei (1 ETH = 10^18 Wei)
    amountWei := new(big.Float).Mul(
        big.NewFloat(amountETH),
        big.NewFloat(1e18),
    )
    amountWeiBig, _ := amountWei.Int(nil)

    // Get nonce
    nonce, err := s.client.PendingNonceAt(ctx, s.address)
    if err != nil {
        return "", 0, err
    }

    // Get gas price
    gasPrice, err := s.client.SuggestGasPrice(ctx)
    if err != nil {
        return "", 0, err
    }

    // Create transaction
    tx := types.NewTransaction(
        nonce,
        to,
        amountWeiBig,
        21000, // Standard gas limit for ETH transfer
        gasPrice,
        nil,
    )

    // Sign transaction
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(s.chainID), s.privateKey)
    if err != nil {
        return "", 0, err
    }

    // Send transaction
    if err := s.client.SendTransaction(ctx, signedTx); err != nil {
        return "", 0, err
    }

    return signedTx.Hash().Hex(), amountETH, nil
}

// GetBalance returns the wallet balance in ETH
func (s *EthereumService) GetBalance(ctx context.Context) (float64, error) {
    balance, err := s.client.BalanceAt(ctx, s.address, nil)
    if err != nil {
        return 0, err
    }

    // Convert Wei to ETH
    ethBalance := new(big.Float).Quo(
        new(big.Float).SetInt(balance),
        big.NewFloat(1e18),
    )

    result, _ := ethBalance.Float64()
    return result, nil
}

// WaitForConfirmation waits for transaction to be mined
func (s *EthereumService) WaitForConfirmation(ctx context.Context, txHash string) (*types.Receipt, error) {
    hash := common.HexToHash(txHash)

    for {
        receipt, err := s.client.TransactionReceipt(ctx, hash)
        if err == nil {
            return receipt, nil
        }

        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        case <-time.After(5 * time.Second):
            continue
        }
    }
}

// pkg/blockchain/service.go

type BlockchainService struct {
    ethereum *EthereumService
    bitcoin  *BitcoinService
    litecoin *LitecoinService
}

func (s *BlockchainService) SendPayment(address, currency string, amountUSD float64) (string, float64, error) {
    ctx := context.Background()

    switch currency {
    case "ethereum":
        return s.ethereum.SendETH(ctx, address, amountUSD)
    case "bitcoin":
        return s.bitcoin.SendBTC(ctx, address, amountUSD)
    case "litecoin":
        return s.litecoin.SendLTC(ctx, address, amountUSD)
    default:
        return "", 0, errors.New("unsupported currency")
    }
}
```

---

## API Implementation

### Route Registration

```go
// internal/api/routes.go

func SetupRoutes(app *fiber.App, services *Services) {
    api := app.Group("/api/v1")

    // Health checks (public)
    app.Get("/health", handlers.Health)
    app.Get("/ready", handlers.Ready(services.DB))
    app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

    // Auth routes (public)
    auth := api.Group("/auth")
    auth.Post("/register", handlers.Register(services.Auth, services.DB))
    auth.Post("/login", handlers.Login(services.Auth, services.DB))
    auth.Post("/refresh", handlers.RefreshToken(services.Auth))

    // Protected routes
    protected := api.Use(middleware.Auth(services.Auth))

    // User routes
    user := protected.Group("/user")
    user.Get("/profile", handlers.GetProfile(services.DB))
    user.Put("/profile", handlers.UpdateProfile(services.DB))
    user.Get("/sessions", handlers.GetUserSessions(services.DB))
    user.Put("/password", handlers.ChangePassword(services.Auth, services.DB))

    // Node routes
    nodes := protected.Group("/nodes")
    nodes.Get("/", handlers.ListNodes(services.DB, services.P2P))
    nodes.Get("/best", handlers.GetBestNode(services.DB, services.P2P))
    nodes.Get("/:id", handlers.GetNode(services.DB))

    // Session routes
    sessions := protected.Group("/sessions")
    sessions.Post("/", handlers.CreateSession(services.Session))
    sessions.Get("/:id", handlers.GetSession(services.DB))
    sessions.Delete("/:id", handlers.DeleteSession(services.Session))

    // Config routes
    configs := protected.Group("/config")
    configs.Post("/generate", handlers.GenerateConfig(services.DB))
    configs.Get("/", handlers.ListConfigs(services.DB))
    configs.Get("/:id", handlers.GetConfig(services.DB))

    // Operator routes
    operator := protected.Group("/operator")
    operator.Post("/register", handlers.RegisterOperator(services.DB))
    operator.Post("/nodes", handlers.CreateOperatorNode(services.DB))
    operator.Get("/nodes", handlers.ListOperatorNodes(services.DB))
    operator.Get("/dashboard", handlers.OperatorDashboard(services.DB))
    operator.Get("/earnings", handlers.ListEarnings(services.DB))
    operator.Get("/payouts", handlers.ListPayouts(services.DB))
    operator.Post("/payout/request", handlers.RequestPayout(services.Reward))

    // Public reward tiers
    api.Get("/operator/rewards/tiers", handlers.GetRewardTiers)

    // Admin routes
    admin := protected.Group("/admin", middleware.Admin())
    admin.Get("/nodes", handlers.AdminListNodes(services.DB))
    admin.Post("/nodes", handlers.AdminCreateNode(services.DB))
    admin.Put("/nodes/:id", handlers.AdminUpdateNode(services.DB))
    admin.Delete("/nodes/:id", handlers.AdminDeleteNode(services.DB))
    admin.Get("/users", handlers.AdminListUsers(services.DB))
    admin.Put("/operators/:id/verify", handlers.VerifyOperator(services.DB))
    admin.Get("/stats", handlers.SystemStats(services.DB))
}
```

### Handler Example

```go
// internal/api/handlers/session.go

type CreateSessionRequest struct {
    NodeID              uuid.UUID `json:"node_id" validate:"required"`
    Protocol            string    `json:"protocol" validate:"required,oneof=wireguard openvpn"`
    EnableKillSwitch    bool      `json:"enable_kill_switch"`
    EnableDNSProtection bool      `json:"enable_dns_protection"`
}

type CreateSessionResponse struct {
    Session *models.Session `json:"session"`
    Config  string          `json:"config"`
}

func CreateSession(sessionService *node.SessionService) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Parse request
        var req CreateSessionRequest
        if err := c.BodyParser(&req); err != nil {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
                "error": "invalid request body",
            })
        }

        // Validate
        if err := validate.Struct(req); err != nil {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
                "error": formatValidationErrors(err),
            })
        }

        // Get user ID from context (set by auth middleware)
        userID := c.Locals("user_id").(uuid.UUID)

        // Create session
        session, config, err := sessionService.CreateSession(
            c.Context(),
            userID,
            req.NodeID,
            req.Protocol,
        )
        if err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                "error": err.Error(),
            })
        }

        // Update session with options
        session.EnableKillSwitch = req.EnableKillSwitch
        session.EnableDNSProtection = req.EnableDNSProtection

        return c.Status(fiber.StatusCreated).JSON(CreateSessionResponse{
            Session: session,
            Config:  config,
        })
    }
}
```

---

## Security Implementation

### Kill Switch

```go
// internal/security/killswitch.go

type KillSwitch struct {
    enabled    bool
    interface_ string
    rules      []string
}

func NewKillSwitch(vpnInterface string) *KillSwitch {
    return &KillSwitch{
        interface_: vpnInterface,
    }
}

// Enable activates the kill switch (blocks non-VPN traffic)
func (k *KillSwitch) Enable() error {
    k.enabled = true

    rules := [][]string{
        // Allow loopback
        {"-A", "OUTPUT", "-o", "lo", "-j", "ACCEPT"},
        // Allow VPN interface
        {"-A", "OUTPUT", "-o", k.interface_, "-j", "ACCEPT"},
        // Allow established connections
        {"-A", "OUTPUT", "-m", "state", "--state", "ESTABLISHED,RELATED", "-j", "ACCEPT"},
        // Allow VPN server connection (must be set per session)
        // {"-A", "OUTPUT", "-d", vpnServerIP, "-j", "ACCEPT"},
        // Block everything else
        {"-A", "OUTPUT", "-j", "DROP"},
    }

    for _, rule := range rules {
        if err := exec.Command("iptables", rule...).Run(); err != nil {
            k.Disable() // Rollback on failure
            return err
        }
        k.rules = append(k.rules, strings.Join(rule, " "))
    }

    return nil
}

// Disable removes kill switch rules
func (k *KillSwitch) Disable() error {
    k.enabled = false

    // Remove rules in reverse order
    for i := len(k.rules) - 1; i >= 0; i-- {
        rule := strings.Replace(k.rules[i], "-A", "-D", 1)
        args := strings.Split(rule, " ")
        exec.Command("iptables", args...).Run()
    }

    k.rules = nil
    return nil
}

// AllowServer adds exception for VPN server
func (k *KillSwitch) AllowServer(serverIP string, port int) error {
    rule := []string{
        "-I", "OUTPUT", "1",
        "-d", serverIP,
        "-p", "udp",
        "--dport", fmt.Sprintf("%d", port),
        "-j", "ACCEPT",
    }
    return exec.Command("iptables", rule...).Run()
}
```

### DNS Leak Protection

```go
// internal/security/dns.go

type DNSProtection struct {
    originalResolv []byte
    enabled        bool
}

func NewDNSProtection() *DNSProtection {
    return &DNSProtection{}
}

// Enable sets custom DNS servers to prevent leaks
func (d *DNSProtection) Enable(dnsServers []string) error {
    // Backup original resolv.conf
    var err error
    d.originalResolv, err = os.ReadFile("/etc/resolv.conf")
    if err != nil {
        return err
    }

    // Create new resolv.conf with VPN DNS
    var content strings.Builder
    content.WriteString("# Aureo VPN DNS Protection\n")
    for _, server := range dnsServers {
        content.WriteString(fmt.Sprintf("nameserver %s\n", server))
    }

    // Make file immutable to prevent NetworkManager changes
    if err := os.WriteFile("/etc/resolv.conf", []byte(content.String()), 0644); err != nil {
        return err
    }

    // Set immutable attribute
    exec.Command("chattr", "+i", "/etc/resolv.conf").Run()

    d.enabled = true
    return nil
}

// Disable restores original DNS configuration
func (d *DNSProtection) Disable() error {
    if !d.enabled {
        return nil
    }

    // Remove immutable attribute
    exec.Command("chattr", "-i", "/etc/resolv.conf").Run()

    // Restore original
    if err := os.WriteFile("/etc/resolv.conf", d.originalResolv, 0644); err != nil {
        return err
    }

    d.enabled = false
    return nil
}
```

---

## Extending the Platform

### Adding a New Protocol

1. Create protocol package:

```go
// pkg/protocols/newprotocol/newprotocol.go

package newprotocol

type Manager struct {
    // Protocol-specific fields
}

func NewManager(config Config) *Manager {
    // Initialize
}

func (m *Manager) CreateInterface() error {
    // Create network interface
}

func (m *Manager) AddPeer(peerConfig PeerConfig) error {
    // Add client peer
}

func (m *Manager) RemovePeer(peerID string) error {
    // Remove client peer
}

func (m *Manager) GenerateClientConfig(cfg ClientConfig) string {
    // Generate client configuration
}
```

2. Register in node service:

```go
// internal/node/service.go

func (s *NodeService) getProtocolManager(protocol string) ProtocolManager {
    switch protocol {
    case "wireguard":
        return s.wgManager
    case "openvpn":
        return s.ovpnManager
    case "newprotocol":
        return s.newProtocolManager
    default:
        return nil
    }
}
```

### Adding a New Blockchain

1. Create blockchain client:

```go
// pkg/blockchain/newchain/client.go

package newchain

type Client struct {
    rpcURL    string
    // Chain-specific fields
}

func NewClient(cfg Config) (*Client, error) {
    // Initialize connection
}

func (c *Client) SendPayment(address string, amount float64) (string, error) {
    // Send transaction
}

func (c *Client) GetBalance() (float64, error) {
    // Get wallet balance
}
```

2. Register in blockchain service:

```go
// pkg/blockchain/service.go

func (s *BlockchainService) SendPayment(address, currency string, amountUSD float64) (string, float64, error) {
    switch currency {
    case "ethereum":
        return s.ethereum.SendETH(ctx, address, amountUSD)
    case "newchain":
        return s.newchain.Send(ctx, address, amountUSD)
    // ...
    }
}
```

### Adding New API Endpoints

1. Define handler:

```go
// internal/api/handlers/newfeature.go

func NewFeatureHandler(db *gorm.DB) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Handle request
    }
}
```

2. Register route:

```go
// internal/api/routes.go

protected.Get("/new-feature", handlers.NewFeatureHandler(services.DB))
```

3. Add tests:

```go
// tests/api/newfeature_test.go

func TestNewFeature(t *testing.T) {
    // Test implementation
}
```

---

## Performance Optimization

### Database Optimization (SQLite)

```go
// SQLite connection with optimizations
db, _ := gorm.Open(sqlite.Open(dbPath+"?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL"), &gorm.Config{})

// Connection pooling for SQLite
sqlDB, _ := db.DB()
sqlDB.SetMaxIdleConns(1)              // SQLite works best with single connection
sqlDB.SetMaxOpenConns(1)              // Prevents "database is locked" errors
sqlDB.SetConnMaxLifetime(time.Hour)

// Enable WAL mode for better read concurrency
db.Exec("PRAGMA journal_mode=WAL")
db.Exec("PRAGMA synchronous=NORMAL")
db.Exec("PRAGMA cache_size=10000")     // 10MB cache
db.Exec("PRAGMA temp_store=MEMORY")

// Indexes for common queries
type VPNNode struct {
    // ...
    Country string `gorm:"index:idx_country_status"`
    Status  string `gorm:"index:idx_country_status"`
}

// Efficient pagination
func GetNodes(db *gorm.DB, page, pageSize int) ([]models.VPNNode, int64) {
    var nodes []models.VPNNode
    var total int64

    db.Model(&models.VPNNode{}).Count(&total)
    db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&nodes)

    return nodes, total
}
```

### Caching

```go
// Redis caching for hot data
func GetNodeFromCache(redis *redis.Client, nodeID string) (*models.VPNNode, error) {
    data, err := redis.Get(ctx, "node:"+nodeID).Bytes()
    if err == redis.Nil {
        return nil, nil // Cache miss
    }

    var node models.VPNNode
    json.Unmarshal(data, &node)
    return &node, nil
}

func CacheNode(redis *redis.Client, node *models.VPNNode) {
    data, _ := json.Marshal(node)
    redis.Set(ctx, "node:"+node.ID.String(), data, 5*time.Minute)
}
```

---

## Debugging Tips

### Logging

```go
// Structured logging with levels
import "github.com/sirupsen/logrus"

log := logrus.WithFields(logrus.Fields{
    "component": "session-service",
    "user_id":   userID,
    "node_id":   nodeID,
})

log.Info("Creating session")
log.WithError(err).Error("Failed to create session")
```

### WireGuard Debugging

```bash
# View interface status
wg show wg0

# View all peers
wg show wg0 dump

# Monitor in real-time
watch -n1 'wg show wg0'

# Check kernel module
dmesg | grep wireguard

# Packet capture
tcpdump -i wg0 -n
```

### Database Queries

```go
// Enable query logging
db, _ := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info),
})

// Debug queries
db.Debug().Where("status = ?", "active").Find(&nodes)

// SQLite specific: check query plan
db.Exec("EXPLAIN QUERY PLAN SELECT * FROM vpn_nodes WHERE status = 'active'")

// Check database integrity
db.Exec("PRAGMA integrity_check")
```

---

This developer guide provides comprehensive documentation for understanding and extending the Aureo VPN platform. For specific implementation questions, refer to the source code and inline documentation.
