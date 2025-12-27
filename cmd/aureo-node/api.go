package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nikola43/aureo-vpn/pkg/crypto"
	"github.com/nikola43/aureo-vpn/pkg/database"
	"github.com/nikola43/aureo-vpn/pkg/models"
	"github.com/nikola43/aureo-vpn/pkg/p2p"
)

// APIServer handles HTTP requests
type APIServer struct {
	app      *fiber.App
	config   Config
	identity *models.NodeIdentity
	p2p      *p2p.Host
	vpn      *VPNService
	jwtKey   []byte
}

// NewAPIServer creates a new API server
func NewAPIServer(config Config, identity *models.NodeIdentity, p2pHost *p2p.Host, vpn *VPNService) *APIServer {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format:     "${time} ${method} ${path} ${status} ${latency}\n",
		TimeFormat: "2006/01/02 15:04:05",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Use node ID as JWT key (in production, use a proper secret)
	jwtKey := []byte(identity.NodeID)

	server := &APIServer{
		app:      app,
		config:   config,
		identity: identity,
		p2p:      p2pHost,
		vpn:      vpn,
		jwtKey:   jwtKey,
	}

	server.setupRoutes()

	return server
}

func (s *APIServer) setupRoutes() {
	// Public routes
	s.app.Get("/health", s.healthCheck)
	s.app.Get("/info", s.getNodeInfo)

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
	protected.Get("/config", s.getConfig)
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Email == "" || req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email, username, and password are required",
		})
	}

	// Check if user exists
	db := database.GetDB()
	var existing models.LocalUser
	if db.Where("email = ? OR username = ?", req.Email, req.Username).First(&existing).Error == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "user already exists",
		})
	}

	// Hash password
	passwordHash, err := crypto.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to hash password",
		})
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create user",
		})
	}

	// Generate token
	token, err := s.generateToken(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate token",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user": fiber.Map{
			"id":       user.ID,
			"email":    user.Email,
			"username": user.Username,
		},
		"token": token,
	})
}

func (s *APIServer) login(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Find user
	db := database.GetDB()
	var user models.LocalUser
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid credentials",
		})
	}

	// Check password
	if !crypto.CheckPassword(req.Password, user.PasswordHash) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid credentials",
		})
	}

	// Generate token
	token, err := s.generateToken(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate token",
		})
	}

	return c.JSON(fiber.Map{
		"user": fiber.Map{
			"id":       user.ID,
			"email":    user.Email,
			"username": user.Username,
		},
		"token": token,
	})
}

// ============= Protected Handlers =============

func (s *APIServer) getProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	db := database.GetDB()
	var user models.LocalUser
	if err := db.First(&user, "id = ?", userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	return c.JSON(fiber.Map{
		"id":                user.ID,
		"email":             user.Email,
		"username":          user.Username,
		"plan":              user.Plan,
		"total_data_used_kb": user.TotalDataUsedKB,
		"session_count":     user.SessionCount,
	})
}

func (s *APIServer) connect(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	var req struct {
		PublicKey string `json:"public_key"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.PublicKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "public_key is required",
		})
	}

	if s.vpn == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "VPN service not available",
		})
	}

	// Create session
	session, err := s.vpn.CreateSession(userID, req.PublicKey)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Return WireGuard config
	return c.JSON(fiber.Map{
		"session_id": session.ID,
		"config": fiber.Map{
			"interface": fiber.Map{
				"address":     session.TunnelIP + "/24",
				"dns":         []string{"1.1.1.1", "8.8.8.8"},
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid session_id",
		})
	}

	if s.vpn == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "VPN service not available",
		})
	}

	if err := s.vpn.DisconnectSession(sessionID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
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

func (s *APIServer) getConfig(c *fiber.Ctx) error {
	if s.vpn == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "VPN service not available",
		})
	}

	return c.JSON(s.vpn.GetServerConfig())
}

// ============= Middleware =============

func (s *APIServer) authMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing authorization header",
		})
	}

	// Parse "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid authorization header",
		})
	}

	tokenString := parts[1]

	// Parse and validate token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.jwtKey, nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid token",
		})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid token claims",
		})
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid user_id in token",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid user_id format",
		})
	}

	c.Locals("user_id", userID)
	return c.Next()
}

func (s *APIServer) generateToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtKey)
}
