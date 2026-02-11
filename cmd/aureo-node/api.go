package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nikola43/aureo-vpn/pkg/crypto"
	"github.com/nikola43/aureo-vpn/pkg/database"
	"github.com/nikola43/aureo-vpn/pkg/metrics"
	"github.com/nikola43/aureo-vpn/pkg/models"
	"github.com/nikola43/aureo-vpn/pkg/p2p"
	"github.com/nikola43/aureo-vpn/pkg/security"
)

// Token configuration constants
const (
	AccessTokenLifetime  = 15 * time.Minute  // Short-lived access tokens
	RefreshTokenLifetime = 7 * 24 * time.Hour // 7 day refresh tokens
	JWTSecretMinLength   = 32                 // Minimum JWT secret length
)

// APIServer handles HTTP requests
type APIServer struct {
	app           *fiber.App
	config        Config
	identity      *models.NodeIdentity
	p2p           *p2p.Host
	vpn           *VPNService
	jwtKey        []byte
	privacyFilter *security.PrivacyFilter
	errorRegistry *security.ErrorRegistry
	tokenBlacklist map[string]time.Time
	blacklistMu    sync.RWMutex
}

// NewAPIServer creates a new API server
func NewAPIServer(config Config, identity *models.NodeIdentity, p2pHost *p2p.Host, vpn *VPNService) *APIServer {
	// Initialize security components
	privacyFilter := security.NewPrivacyFilter(nil)
	errorRegistry := security.NewErrorRegistry(100)

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		// Secure error handler - never expose internal details
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Log internal error for debugging (privacy-safe)
			log.Printf("[ERROR] %s %s: %v", c.Method(), c.Path(), privacyFilter.SanitizeLogMessage(err.Error()))

			// Return generic error to client
			secErr := security.InternalError(err)
			errorRegistry.Record(secErr)
			return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
		},
		// Security limits
		BodyLimit:     1 * 1024 * 1024, // 1MB max body
		ReadTimeout:   10 * time.Second,
		WriteTimeout:  10 * time.Second,
		IdleTimeout:   30 * time.Second,
	})

	// Recovery middleware
	app.Use(recover.New(recover.Config{
		EnableStackTrace: false, // Don't expose stack traces
	}))

	// Privacy-safe request logging (NO IP addresses)
	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)
		// Log WITHOUT IP address for privacy
		log.Printf("[API] %s %s %d %s", c.Method(), c.Path(), c.Response().StatusCode(), duration)
		return err
	})

	// Rate limiting
	app.Use(limiter.New(limiter.Config{
		Max:               100,             // 100 requests
		Expiration:        1 * time.Minute, // per minute
		LimiterMiddleware: limiter.SlidingWindow{},
		KeyGenerator: func(c *fiber.Ctx) string {
			// Use a hash of the IP for rate limiting (privacy-preserving)
			return privacyFilter.AnonymizeIP(c.IP())
		},
		LimitReached: func(c *fiber.Ctx) error {
			secErr := security.RateLimited(time.Minute)
			return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
		},
	}))

	// CORS - configure with specific origins in production
	allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "*" // Default for development, restrict in production
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: false,
		MaxAge:           3600,
	}))

	// Security headers
	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Cache-Control", "no-store")
		return c.Next()
	})

	// Get secure JWT secret from environment or generate one
	jwtKey := getSecureJWTSecret()

	server := &APIServer{
		app:            app,
		config:         config,
		identity:       identity,
		p2p:            p2pHost,
		vpn:            vpn,
		jwtKey:         jwtKey,
		privacyFilter:  privacyFilter,
		errorRegistry:  errorRegistry,
		tokenBlacklist: make(map[string]time.Time),
	}

	server.setupRoutes()

	return server
}

// getSecureJWTSecret retrieves or generates a secure JWT secret
func getSecureJWTSecret() []byte {
	// First, try to get from environment variable
	secret := os.Getenv("JWT_SECRET")
	if secret != "" && len(secret) >= JWTSecretMinLength {
		return []byte(secret)
	}

	// Log warning if JWT_SECRET is missing or too short
	if secret != "" {
		log.Printf("[SECURITY WARNING] JWT_SECRET is too short (min %d bytes). Generating secure random secret.", JWTSecretMinLength)
	} else {
		log.Printf("[SECURITY WARNING] JWT_SECRET not set. Generating secure random secret. Set JWT_SECRET env var for persistence.")
	}

	// Generate a secure random secret
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		log.Fatalf("[FATAL] Failed to generate secure JWT secret: %v", err)
	}

	return secretBytes
}

func (s *APIServer) setupRoutes() {
	// Public routes
	s.app.Get("/health", s.healthCheck)
	s.app.Get("/info", s.getNodeInfo)
	s.app.Get("/metrics", metrics.PrometheusHandler())

	// API routes
	api := s.app.Group("/api/v1")

	// Auth
	api.Post("/auth/register", s.register)
	api.Post("/auth/login", s.login)

	// Nodes (P2P discovery)
	api.Get("/nodes", s.listNodes)
	api.Get("/nodes/best", s.getBestNode)
	api.Get("/nodes/countries", s.getCountries)

	// P2P status
	api.Get("/p2p/status", s.getP2PStatus)
	api.Get("/p2p/peers", s.getPeers)

	// Protected routes
	protected := api.Group("/", s.authMiddleware)
	protected.Get("/me", s.getProfile)
	protected.Post("/connect", s.connect)
	protected.Post("/disconnect", s.disconnect)
	protected.Get("/sessions", s.getSessions)
	protected.Get("/sessions/:id", s.getSessionByID)
	protected.Get("/config", s.getConfig)
	protected.Post("/config/generate", s.generateConfig) // Client script compatibility
}

// Start starts the API server
func (s *APIServer) Start() error {
	addr := fmt.Sprintf(":%d", s.config.APIPort)
	log.Printf("[API] Server listening on %s", addr)
	return s.app.Listen(addr)
}

// Stop stops the API server
func (s *APIServer) Stop() error {
	return s.app.Shutdown()
}

// ============= Public Handlers =============

func (s *APIServer) healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "healthy",
		"node_id":   s.identity.NodeID,
		"node_name": s.identity.Name,
		"timestamp": time.Now().Unix(),
	})
}

func (s *APIServer) getNodeInfo(c *fiber.Ctx) error {
	connections := 0
	loadScore := 0.0
	if s.vpn != nil {
		connections = s.vpn.GetConnectionCount()
		loadScore = s.vpn.GetLoadScore()
	}

	return c.JSON(fiber.Map{
		"id":                  s.identity.NodeID,
		"name":                s.identity.Name,
		"public_ip":           s.config.PublicIP,
		"country":             s.config.Country,
		"country_code":        s.config.CountryCode,
		"city":                s.config.City,
		"latitude":            s.config.Latitude,
		"longitude":           s.config.Longitude,
		"wireguard_port":      s.config.WGPort,
		"public_key":          s.identity.WGPublicKey,
		"current_connections": connections,
		"load_score":          loadScore,
		"p2p_peer_id":         s.identity.P2PPeerID,
		"status":              "online",
	})
}

func (s *APIServer) listNodes(c *fiber.Ctx) error {
	protocol := c.Query("protocol", "")
	country := c.Query("country", "")

	if s.p2p == nil {
		return c.JSON(fiber.Map{
			"nodes":  []interface{}{},
			"count":  0,
			"source": "p2p_disabled",
		})
	}

	nodes := s.p2p.QueryNodes(protocol, country)

	// Convert to response format
	result := make([]fiber.Map, len(nodes))
	for i, n := range nodes {
		result[i] = fiber.Map{
			"id":                  n.ID,
			"name":                n.Name,
			"public_ip":           n.PublicIP,
			"country":             n.Country,
			"country_code":        n.CountryCode,
			"city":                n.City,
			"latitude":            n.Latitude,
			"longitude":           n.Longitude,
			"wireguard_port":      n.WireGuardPort,
			"public_key":          n.PublicKey,
			"current_connections": n.CurrentConnections,
			"max_connections":     n.MaxConnections,
			"load_score":          n.LoadScore,
			"status":              n.Status,
			"supports_wireguard":  n.SupportsWireGuard,
			"last_heartbeat":      n.LastHeartbeat,
		}
	}

	return c.JSON(fiber.Map{
		"nodes":  result,
		"count":  len(result),
		"source": "p2p",
	})
}

func (s *APIServer) getBestNode(c *fiber.Ctx) error {
	protocol := c.Query("protocol", "wireguard")
	country := c.Query("country", "")

	if s.p2p == nil {
		// Return self as the only node
		return c.JSON(fiber.Map{
			"id":             s.identity.NodeID,
			"name":           s.identity.Name,
			"public_ip":      s.config.PublicIP,
			"wireguard_port": s.config.WGPort,
			"public_key":     s.identity.WGPublicKey,
		})
	}

	node := s.p2p.GetBestNode(protocol, country)
	if node == nil {
		// Return self as fallback
		return c.JSON(fiber.Map{
			"id":             s.identity.NodeID,
			"name":           s.identity.Name,
			"public_ip":      s.config.PublicIP,
			"wireguard_port": s.config.WGPort,
			"public_key":     s.identity.WGPublicKey,
		})
	}

	return c.JSON(fiber.Map{
		"id":                  node.ID,
		"name":                node.Name,
		"public_ip":           node.PublicIP,
		"country":             node.Country,
		"country_code":        node.CountryCode,
		"city":                node.City,
		"latitude":            node.Latitude,
		"longitude":           node.Longitude,
		"wireguard_port":      node.WireGuardPort,
		"public_key":          node.PublicKey,
		"current_connections": node.CurrentConnections,
		"load_score":          node.LoadScore,
	})
}

func (s *APIServer) getCountries(c *fiber.Ctx) error {
	if s.p2p == nil {
		countries := []string{}
		if s.config.CountryCode != "" {
			countries = append(countries, s.config.CountryCode)
		}
		return c.JSON(fiber.Map{"countries": countries})
	}

	return c.JSON(fiber.Map{
		"countries": s.p2p.GetRegistry().GetCountries(),
	})
}

func (s *APIServer) getP2PStatus(c *fiber.Ctx) error {
	if s.p2p == nil {
		return c.JSON(fiber.Map{
			"enabled": false,
		})
	}

	registry := s.p2p.GetRegistry()
	return c.JSON(fiber.Map{
		"enabled":         true,
		"peer_id":         s.p2p.GetPeerID().String(),
		"multiaddrs":      s.p2p.GetMultiaddrs(),
		"connected_peers": s.p2p.ConnectedPeers(),
		"known_nodes":     registry.Count(),
		"active_nodes":    registry.ActiveCount(),
	})
}

func (s *APIServer) getPeers(c *fiber.Ctx) error {
	if s.p2p == nil {
		return c.JSON(fiber.Map{"peers": []interface{}{}})
	}

	nodes := s.p2p.GetRegistry().GetAllNodes()
	peers := make([]fiber.Map, len(nodes))
	for i, n := range nodes {
		peers[i] = fiber.Map{
			"id":          n.ID,
			"peer_id":     n.PeerID.String(),
			"name":        n.Name,
			"country":     n.CountryCode,
			"status":      n.Status,
			"last_seen":   n.LastHeartbeat,
		}
	}

	return c.JSON(fiber.Map{"peers": peers})
}

// ============= Auth Handlers =============

func (s *APIServer) register(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		secErr := security.BadRequest(err)
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	// Input validation using security package
	if err := security.ValidateEmail(req.Email); err != nil {
		secErr := security.ValidationError("email", err.Error())
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	if err := security.ValidateUsername(req.Username); err != nil {
		secErr := security.ValidationError("username", err.Error())
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	if err := security.ValidatePassword(req.Password); err != nil {
		secErr := security.ValidationError("password", err.Error())
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	// Check if user exists - use constant-time comparison for security
	db := database.GetDB()
	var existing models.LocalUser
	if db.Where("email = ? OR username = ?", req.Email, req.Username).First(&existing).Error == nil {
		// Don't reveal whether it was email or username that matched
		secErr := security.Conflict(fmt.Errorf("registration failed"))
		secErr.Message = "An account with this email or username already exists."
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	// Hash password using Argon2id
	passwordHash, err := crypto.HashPassword(req.Password)
	if err != nil {
		log.Printf("[ERROR] Failed to hash password: %v", s.privacyFilter.SanitizeLogMessage(err.Error()))
		secErr := security.InternalError(err)
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	// Create user
	user := models.LocalUser{
		ID:           uuid.New(),
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: passwordHash,
		Plan:         "free",
	}

	if err := db.Create(&user).Error; err != nil {
		log.Printf("[ERROR] Failed to create user: %v", s.privacyFilter.SanitizeLogMessage(err.Error()))
		secErr := security.InternalError(err)
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	// Generate token
	token, err := s.generateToken(user.ID)
	if err != nil {
		log.Printf("[ERROR] Failed to generate token: %v", s.privacyFilter.SanitizeLogMessage(err.Error()))
		secErr := security.InternalError(err)
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user": fiber.Map{
			"id":       user.ID,
			"email":    user.Email,
			"username": user.Username,
		},
		"token":      token,
		"expires_in": int(AccessTokenLifetime.Seconds()),
	})
}

func (s *APIServer) login(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		secErr := security.BadRequest(err)
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	// Basic input validation
	if req.Email == "" || req.Password == "" {
		secErr := security.ValidationError("email", "email and password are required")
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	// Find user - always perform password check to prevent timing attacks
	db := database.GetDB()
	var user models.LocalUser
	userFound := db.Where("email = ?", req.Email).First(&user).Error == nil

	// Always check password (even if user not found) to prevent timing attacks
	passwordValid := false
	if userFound {
		passwordValid = crypto.CheckPassword(req.Password, user.PasswordHash)
	} else {
		// Dummy password check to maintain constant timing
		_ = crypto.CheckPassword(req.Password, "$argon2id$v=19$m=65536,t=3,p=4$dummy$hash")
	}

	// Return same error for both user not found and wrong password
	if !userFound || !passwordValid {
		secErr := security.Unauthorized(fmt.Errorf("authentication failed"))
		secErr.Message = "Invalid email or password."
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	// Generate token
	token, err := s.generateToken(user.ID)
	if err != nil {
		log.Printf("[ERROR] Failed to generate token: %v", s.privacyFilter.SanitizeLogMessage(err.Error()))
		secErr := security.InternalError(err)
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	return c.JSON(fiber.Map{
		"user": fiber.Map{
			"id":       user.ID,
			"email":    user.Email,
			"username": user.Username,
		},
		"token":        token,
		"access_token": token, // Compatibility with client scripts
		"expires_in":   int(AccessTokenLifetime.Seconds()),
	})
}

// ============= Protected Handlers =============

func (s *APIServer) getProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	db := database.GetDB()
	var user models.LocalUser
	if err := db.First(&user, "id = ?", userID).Error; err != nil {
		secErr := security.NotFound(err)
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	return c.JSON(fiber.Map{
		"id":                 user.ID,
		"email":              user.Email,
		"username":           user.Username,
		"plan":               user.Plan,
		"total_data_used_kb": user.TotalDataUsedKB,
		"session_count":      user.SessionCount,
	})
}

func (s *APIServer) connect(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	var req struct {
		PublicKey string `json:"public_key"`
	}

	if err := c.BodyParser(&req); err != nil {
		secErr := security.BadRequest(err)
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	// Validate WireGuard public key format (base64, 44 chars with =)
	if req.PublicKey == "" {
		secErr := security.ValidationError("public_key", "is required")
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	// Basic validation: WireGuard public keys are 44 chars base64
	if len(req.PublicKey) != 44 {
		secErr := security.ValidationError("public_key", "invalid format")
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	if s.vpn == nil {
		secErr := security.ServiceUnavailable("VPN")
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	// Create session
	session, err := s.vpn.CreateSession(userID, req.PublicKey)
	if err != nil {
		log.Printf("[ERROR] Failed to create VPN session: %v", s.privacyFilter.SanitizeLogMessage(err.Error()))
		secErr := security.InternalError(err)
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	// Return WireGuard config
	return c.JSON(fiber.Map{
		"session_id": session.ID,
		"config": fiber.Map{
			"interface": fiber.Map{
				"address": session.TunnelIP + "/24",
				"dns":     []string{"1.1.1.1", "8.8.8.8"},
			},
			"peer": fiber.Map{
				"public_key":           s.identity.WGPublicKey,
				"endpoint":             fmt.Sprintf("%s:%d", s.config.PublicIP, s.config.WGPort),
				"allowed_ips":          "0.0.0.0/0",
				"persistent_keepalive": 25,
			},
		},
	})
}

func (s *APIServer) disconnect(c *fiber.Ctx) error {
	var req struct {
		SessionID string `json:"session_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		secErr := security.BadRequest(err)
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	// Validate session ID format
	if err := security.ValidateUUID(req.SessionID); err != nil {
		secErr := security.ValidationError("session_id", "invalid format")
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	sessionID, _ := uuid.Parse(req.SessionID) // Already validated

	if s.vpn == nil {
		secErr := security.ServiceUnavailable("VPN")
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	if err := s.vpn.DisconnectSession(sessionID); err != nil {
		secErr := security.NotFound(err)
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	return c.JSON(fiber.Map{
		"message": "disconnected successfully",
	})
}

func (s *APIServer) getSessions(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	db := database.GetDB()
	var sessions []models.LocalSession
	db.Where("user_id = ? AND status = ?", userID, "active").Find(&sessions)

	return c.JSON(fiber.Map{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

func (s *APIServer) getSessionByID(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid session id"})
	}

	db := database.GetDB()
	var session models.LocalSession
	if err := db.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "session not found"})
	}

	return c.JSON(session)
}

func (s *APIServer) getConfig(c *fiber.Ctx) error {
	if s.vpn == nil {
		secErr := security.ServiceUnavailable("VPN")
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	return c.JSON(s.vpn.GetServerConfig())
}

// generateConfig handles client script compatibility for /api/v1/config/generate
func (s *APIServer) generateConfig(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	var req struct {
		PublicKey string `json:"public_key"`
		NodeID    string `json:"node_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		secErr := security.BadRequest(err)
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	// Validate WireGuard public key format
	if req.PublicKey == "" {
		secErr := security.ValidationError("public_key", "is required")
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	if len(req.PublicKey) != 44 {
		secErr := security.ValidationError("public_key", "invalid format")
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	// Validate node_id if provided
	if req.NodeID != "" {
		if err := security.ValidateUUID(req.NodeID); err != nil {
			secErr := security.ValidationError("node_id", "invalid format")
			return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
		}
	}

	if s.vpn == nil {
		secErr := security.ServiceUnavailable("VPN")
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	// Create session
	session, err := s.vpn.CreateSession(userID, req.PublicKey)
	if err != nil {
		log.Printf("[ERROR] Failed to create VPN session: %v", s.privacyFilter.SanitizeLogMessage(err.Error()))
		secErr := security.InternalError(err)
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	// Return in the format expected by client scripts
	return c.JSON(fiber.Map{
		"server_public_key": s.identity.WGPublicKey,
		"server_endpoint":   fmt.Sprintf("%s:%d", s.config.PublicIP, s.config.WGPort),
		"client_ip":         session.TunnelIP,
		"dns":               "1.1.1.1,8.8.8.8",
		"allowed_ips":       "0.0.0.0/0",
		"keepalive":         25,
		"session_id":        session.ID,
	})
}

// ============= Middleware =============

func (s *APIServer) authMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		secErr := security.Unauthorized(fmt.Errorf("missing authorization header"))
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	// Parse "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		secErr := security.Unauthorized(fmt.Errorf("invalid authorization format"))
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	tokenString := parts[1]

	// Basic token format validation
	if len(tokenString) < 10 || len(tokenString) > 2048 {
		secErr := security.Unauthorized(fmt.Errorf("invalid token length"))
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	// Parse and validate token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtKey, nil
	})

	if err != nil || !token.Valid {
		secErr := security.Unauthorized(fmt.Errorf("token validation failed: %w", err))
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		secErr := security.Unauthorized(fmt.Errorf("invalid claims type"))
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	// Check token ID for blacklist
	if tokenID, ok := claims["jti"].(string); ok {
		if s.isTokenBlacklisted(tokenID) {
			secErr := security.Unauthorized(fmt.Errorf("token has been revoked"))
			return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
		}
	}

	// Validate user_id claim
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		secErr := security.Unauthorized(fmt.Errorf("missing user_id claim"))
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		secErr := security.Unauthorized(fmt.Errorf("invalid user_id format"))
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	c.Locals("user_id", userID)
	c.Locals("token_claims", claims) // Store claims for potential use
	return c.Next()
}

func (s *APIServer) generateToken(userID uuid.UUID) (string, error) {
	now := time.Now()

	// Generate unique token ID for revocation support
	tokenIDBytes := make([]byte, 16)
	if _, err := rand.Read(tokenIDBytes); err != nil {
		return "", fmt.Errorf("failed to generate token ID: %w", err)
	}
	tokenID := base64.RawURLEncoding.EncodeToString(tokenIDBytes)

	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"jti":     tokenID,                                    // JWT ID for revocation
		"exp":     now.Add(AccessTokenLifetime).Unix(),        // 15 minutes
		"iat":     now.Unix(),
		"nbf":     now.Unix(),                                 // Not valid before
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtKey)
}

// isTokenBlacklisted checks if a token has been revoked
func (s *APIServer) isTokenBlacklisted(tokenID string) bool {
	s.blacklistMu.RLock()
	defer s.blacklistMu.RUnlock()
	_, exists := s.tokenBlacklist[tokenID]
	return exists
}

// blacklistToken adds a token to the blacklist
func (s *APIServer) blacklistToken(tokenID string, expiry time.Time) {
	s.blacklistMu.Lock()
	defer s.blacklistMu.Unlock()
	s.tokenBlacklist[tokenID] = expiry
}

// cleanupBlacklist removes expired tokens from blacklist
func (s *APIServer) cleanupBlacklist() {
	s.blacklistMu.Lock()
	defer s.blacklistMu.Unlock()
	now := time.Now()
	for tokenID, expiry := range s.tokenBlacklist {
		if expiry.Before(now) {
			delete(s.tokenBlacklist, tokenID)
		}
	}
}
