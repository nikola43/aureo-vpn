package api

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nikola43/aureo-vpn/pkg/auth"
	"github.com/nikola43/aureo-vpn/pkg/database"
	apperrors "github.com/nikola43/aureo-vpn/pkg/errors"
	"github.com/nikola43/aureo-vpn/pkg/metrics"
	"github.com/nikola43/aureo-vpn/pkg/models"
	"github.com/nikola43/aureo-vpn/pkg/operator"
	"github.com/nikola43/aureo-vpn/pkg/p2p"
	"github.com/nikola43/aureo-vpn/pkg/security"
)

// Privacy filter for secure logging
var privacyFilter = security.NewPrivacyFilter(nil)

// Handlers holds all API handlers
type Handlers struct {
	authService     *auth.Service
	operatorService *operator.Service
	p2pClient       *p2p.Client
}

// NewHandlers creates new API handlers
func NewHandlers(authService *auth.Service, operatorService *operator.Service) *Handlers {
	return &Handlers{
		authService:     authService,
		operatorService: operatorService,
	}
}

// NewHandlersWithP2P creates new API handlers with P2P client
func NewHandlersWithP2P(authService *auth.Service, operatorService *operator.Service, p2pClient *p2p.Client) *Handlers {
	return &Handlers{
		authService:     authService,
		operatorService: operatorService,
		p2pClient:       p2pClient,
	}
}

// SetP2PClient sets the P2P client for decentralized node discovery
func (h *Handlers) SetP2PClient(client *p2p.Client) {
	h.p2pClient = client
}

// Register handles user registration
func (h *Handlers) Register(c *fiber.Ctx) error {
	var req auth.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		secErr := security.BadRequest(err)
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	// Validate input
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

	resp, err := h.authService.Register(req)
	if err != nil {
		metrics.LoginAttempts.WithLabelValues("failed").Inc()
		log.Printf("[AUTH] Registration failed: %v", privacyFilter.SanitizeLogMessage(err.Error()))
		// Don't expose internal error details - use generic message
		secErr := security.Conflict(err)
		secErr.Message = "Registration failed. Email or username may already exist."
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	metrics.UserRegistrations.Inc()
	metrics.LoginAttempts.WithLabelValues("success").Inc()

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// Login handles user login
func (h *Handlers) Login(c *fiber.Ctx) error {
	var req auth.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		secErr := security.BadRequest(err)
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	// Basic validation
	if req.Email == "" || req.Password == "" {
		secErr := security.ValidationError("email", "email and password are required")
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	resp, err := h.authService.Login(req)
	if err != nil {
		metrics.LoginAttempts.WithLabelValues("failed").Inc()
		// Don't expose whether email exists or password was wrong
		secErr := security.Unauthorized(err)
		secErr.Message = "Invalid email or password."
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	metrics.LoginAttempts.WithLabelValues("success").Inc()

	return c.JSON(resp)
}

// RefreshToken handles token refresh
func (h *Handlers) RefreshToken(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.BodyParser(&req); err != nil {
		secErr := security.BadRequest(err)
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	if req.RefreshToken == "" {
		secErr := security.ValidationError("refresh_token", "is required")
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	accessToken, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		secErr := security.Unauthorized(err)
		secErr.Message = "Invalid or expired refresh token."
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	metrics.TokenGenerations.WithLabelValues("access").Inc()

	return c.JSON(fiber.Map{
		"access_token": accessToken,
	})
}

// GetProfile returns the authenticated user's profile
func (h *Handlers) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	user, err := h.authService.GetUser(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	return c.JSON(user)
}

// ListNodes returns all available VPN nodes
// Uses P2P discovery when available, falls back to database
func (h *Handlers) ListNodes(c *fiber.Ctx) error {
	country := c.Query("country")
	protocol := c.Query("protocol")
	source := c.Query("source", "auto") // "p2p", "db", or "auto"

	// Try P2P first if available and not explicitly requesting DB
	if h.p2pClient != nil && source != "db" {
		p2pNodes := h.p2pClient.QueryNodes(protocol, country)
		if len(p2pNodes) > 0 || source == "p2p" {
			// Convert P2P nodes to response format
			nodes := make([]fiber.Map, len(p2pNodes))
			for i, n := range p2pNodes {
				nodes[i] = fiber.Map{
					"id":                   n.ID,
					"name":                 n.Name,
					"public_ip":            n.PublicIP,
					"country":              n.Country,
					"country_code":         n.CountryCode,
					"city":                 n.City,
					"wireguard_port":       n.WireGuardPort,
					"openvpn_port":         n.OpenVPNPort,
					"public_key":           n.PublicKey,
					"max_connections":      n.MaxConnections,
					"current_connections":  n.CurrentConnections,
					"load_score":           n.LoadScore,
					"status":               n.Status,
					"supports_wireguard":   n.SupportsWireGuard,
					"supports_openvpn":     n.SupportsOpenVPN,
					"supports_multihop":    n.SupportsMultiHop,
					"supports_obfuscation": n.SupportsObfuscation,
					"is_operator_owned":    n.IsOperatorOwned,
					"reputation":           n.Reputation,
					"last_heartbeat":       n.LastHeartbeat,
				}
			}

			return c.JSON(fiber.Map{
				"nodes":  nodes,
				"count":  len(nodes),
				"source": "p2p",
			})
		}
	}

	// Fall back to database
	db := database.GetDB()

	var nodes []models.VPNNode
	query := db.Where("is_active = ? AND status = ?", true, "online")

	// Optional filters
	if country != "" {
		query = query.Where("country_code = ?", country)
	}

	if protocol != "" {
		if protocol == "wireguard" {
			query = query.Where("supports_wireguard = ?", true)
		} else if protocol == "openvpn" {
			query = query.Where("supports_openvpn = ?", true)
		}
	}

	// Sort by load score (best servers first)
	if err := query.Order("load_score ASC").Find(&nodes).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch nodes",
		})
	}

	return c.JSON(fiber.Map{
		"nodes":  nodes,
		"count":  len(nodes),
		"source": "database",
	})
}

// GetBestNode returns the best available node based on load and latency
// Uses P2P discovery when available, falls back to database
func (h *Handlers) GetBestNode(c *fiber.Ctx) error {
	protocol := c.Query("protocol", "wireguard")
	country := c.Query("country")
	source := c.Query("source", "auto")

	// Try P2P first if available
	if h.p2pClient != nil && source != "db" {
		node := h.p2pClient.GetBestNode(protocol, country)
		if node != nil {
			return c.JSON(fiber.Map{
				"id":                   node.ID,
				"name":                 node.Name,
				"public_ip":            node.PublicIP,
				"country":              node.Country,
				"country_code":         node.CountryCode,
				"city":                 node.City,
				"wireguard_port":       node.WireGuardPort,
				"openvpn_port":         node.OpenVPNPort,
				"public_key":           node.PublicKey,
				"max_connections":      node.MaxConnections,
				"current_connections":  node.CurrentConnections,
				"load_score":           node.LoadScore,
				"status":               node.Status,
				"supports_wireguard":   node.SupportsWireGuard,
				"supports_openvpn":     node.SupportsOpenVPN,
				"is_operator_owned":    node.IsOperatorOwned,
				"reputation":           node.Reputation,
				"last_heartbeat":       node.LastHeartbeat,
				"source":               "p2p",
			})
		}
		if source == "p2p" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "no nodes found via P2P",
			})
		}
	}

	// Fall back to database
	db := database.GetDB()

	query := db.Where("is_active = ? AND status = ?", true, "online")

	if country != "" {
		query = query.Where("country_code = ?", country)
	}

	if protocol == "wireguard" {
		query = query.Where("supports_wireguard = ?", true)
	} else if protocol == "openvpn" {
		query = query.Where("supports_openvpn = ?", true)
	}

	var node models.VPNNode
	if err := query.Order("load_score ASC").First(&node).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "no available nodes found",
		})
	}

	return c.JSON(node)
}

// GetActiveSessions returns active sessions for the authenticated user
func (h *Handlers) GetActiveSessions(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	db := database.GetDB()

	var sessions []models.Session
	if err := db.Where("user_id = ? AND status = ?", userID, "active").
		Preload("Node").
		Find(&sessions).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch sessions",
		})
	}

	return c.JSON(fiber.Map{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// GetStats returns statistics for the authenticated user
func (h *Handlers) GetStats(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	db := database.GetDB()

	var stats struct {
		TotalSessions     int64   `json:"total_sessions"`
		ActiveSessions    int64   `json:"active_sessions"`
		DataTransferredGB float64 `json:"data_transferred_gb"`
	}

	db.Model(&models.Session{}).Where("user_id = ?", userID).Count(&stats.TotalSessions)
	db.Model(&models.Session{}).Where("user_id = ? AND status = ?", userID, "active").Count(&stats.ActiveSessions)

	var user models.User
	if err := db.First(&user, userID).Error; err == nil {
		stats.DataTransferredGB = user.DataTransferredGB
	}

	return c.JSON(stats)
}

// HealthCheck returns the health status of the API
func (h *Handlers) HealthCheck(c *fiber.Ctx) error {
	// Check database connection
	if err := database.HealthCheck(); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":   "unhealthy",
			"database": "disconnected",
		})
	}

	return c.JSON(fiber.Map{
		"status":   "healthy",
		"database": "connected",
	})
}

// GetP2PStatus returns the P2P network status
func (h *Handlers) GetP2PStatus(c *fiber.Ctx) error {
	if h.p2pClient == nil {
		return c.JSON(fiber.Map{
			"enabled": false,
			"message": "P2P is not enabled on this API gateway",
		})
	}

	stats := h.p2pClient.GetStats()
	return c.JSON(fiber.Map{
		"enabled":         true,
		"peer_id":         stats["peer_id"],
		"connected_peers": stats["connected_peers"],
		"known_nodes":     stats["known_nodes"],
		"active_nodes":    stats["active_nodes"],
		"countries":       stats["countries"],
	})
}

// GetP2PCountries returns countries with active P2P nodes
func (h *Handlers) GetP2PCountries(c *fiber.Ctx) error {
	if h.p2pClient == nil {
		return c.JSON(fiber.Map{
			"countries": []string{},
			"source":    "disabled",
		})
	}

	return c.JSON(fiber.Map{
		"countries": h.p2pClient.GetCountries(),
		"source":    "p2p",
	})
}

// Admin-only handlers

// ListAllNodes returns all nodes (admin only)
func (h *Handlers) ListAllNodes(c *fiber.Ctx) error {
	db := database.GetDB()

	var nodes []models.VPNNode
	if err := db.Find(&nodes).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch nodes",
		})
	}

	return c.JSON(fiber.Map{
		"nodes": nodes,
		"count": len(nodes),
	})
}

// ListAllUsers returns all users (admin only)
func (h *Handlers) ListAllUsers(c *fiber.Ctx) error {
	db := database.GetDB()

	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch users",
		})
	}

	return c.JSON(fiber.Map{
		"users": users,
		"count": len(users),
	})
}

// GetSystemStats returns system-wide statistics (admin only)
func (h *Handlers) GetSystemStats(c *fiber.Ctx) error {
	db := database.GetDB()

	var stats struct {
		TotalUsers     int64 `json:"total_users"`
		ActiveUsers    int64 `json:"active_users"`
		TotalNodes     int64 `json:"total_nodes"`
		OnlineNodes    int64 `json:"online_nodes"`
		TotalSessions  int64 `json:"total_sessions"`
		ActiveSessions int64 `json:"active_sessions"`
	}

	db.Model(&models.User{}).Count(&stats.TotalUsers)
	db.Model(&models.User{}).Where("is_active = ?", true).Count(&stats.ActiveUsers)
	db.Model(&models.VPNNode{}).Count(&stats.TotalNodes)
	db.Model(&models.VPNNode{}).Where("status = ?", "online").Count(&stats.OnlineNodes)
	db.Model(&models.Session{}).Count(&stats.TotalSessions)
	db.Model(&models.Session{}).Where("status = ?", "active").Count(&stats.ActiveSessions)

	return c.JSON(stats)
}

// Operator Handlers

// RegisterOperator handles operator registration
func (h *Handlers) RegisterOperator(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	var req operator.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	op, err := h.operatorService.RegisterOperator(c.Context(), userID, req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			return c.Status(appErr.StatusCode).JSON(fiber.Map{
				"error": appErr.Message,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to register operator",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"operator": op,
		"message":  "Operator registered successfully. Please wait for verification.",
	})
}

// CreateOperatorNode handles node creation for operators
func (h *Handlers) CreateOperatorNode(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	// Get operator by user ID
	op, err := h.operatorService.GetOperatorByUserID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You must be a registered operator to create nodes",
		})
	}

	var req operator.NodeCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	node, publicKey, err := h.operatorService.CreateNode(c.Context(), op.ID, req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			return c.Status(appErr.StatusCode).JSON(fiber.Map{
				"error": appErr.Message,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create node",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"node":       node,
		"public_key": publicKey,
		"message":    "Node created successfully. Configure your node software with these credentials.",
	})
}

// GetOperatorNodes returns all nodes for an operator
func (h *Handlers) GetOperatorNodes(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	op, err := h.operatorService.GetOperatorByUserID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You must be a registered operator",
		})
	}

	nodes, err := h.operatorService.GetOperatorNodes(c.Context(), op.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch nodes",
		})
	}

	return c.JSON(fiber.Map{
		"nodes": nodes,
		"count": len(nodes),
	})
}

// GetOperatorStats returns comprehensive statistics for an operator
func (h *Handlers) GetOperatorStats(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	op, err := h.operatorService.GetOperatorByUserID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You must be a registered operator",
		})
	}

	stats, err := h.operatorService.GetOperatorStats(c.Context(), op.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch stats",
		})
	}

	return c.JSON(stats)
}

// GetOperatorEarnings returns earnings history
func (h *Handlers) GetOperatorEarnings(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	op, err := h.operatorService.GetOperatorByUserID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You must be a registered operator",
		})
	}

	// Parse pagination
	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	earnings, total, err := h.operatorService.GetOperatorEarnings(c.Context(), op.ID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch earnings",
		})
	}

	return c.JSON(fiber.Map{
		"earnings": earnings,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetOperatorPayouts returns payout history
func (h *Handlers) GetOperatorPayouts(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	op, err := h.operatorService.GetOperatorByUserID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You must be a registered operator",
		})
	}

	// Parse pagination
	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	payouts, total, err := h.operatorService.GetOperatorPayouts(c.Context(), op.ID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch payouts",
		})
	}

	return c.JSON(fiber.Map{
		"payouts": payouts,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// RequestOperatorPayout handles manual payout requests
func (h *Handlers) RequestOperatorPayout(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	op, err := h.operatorService.GetOperatorByUserID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You must be a registered operator",
		})
	}

	if err := h.operatorService.RequestPayout(c.Context(), op.ID); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			return c.Status(appErr.StatusCode).JSON(fiber.Map{
				"error": appErr.Message,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to process payout request",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Payout request submitted successfully. Processing may take 24-48 hours.",
	})
}

// GetOperatorDashboard returns comprehensive dashboard data
func (h *Handlers) GetOperatorDashboard(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	op, err := h.operatorService.GetOperatorByUserID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You must be a registered operator",
		})
	}

	dashboard, err := h.operatorService.GetOperatorDashboard(c.Context(), op.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch dashboard data",
		})
	}

	return c.JSON(dashboard)
}

// GetRewardTiers returns all reward tiers (public endpoint)
func (h *Handlers) GetRewardTiers(c *fiber.Ctx) error {
	tiers, err := h.operatorService.GetRewardTiers(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch reward tiers",
		})
	}

	return c.JSON(fiber.Map{
		"tiers": tiers,
		"count": len(tiers),
	})
}

// Admin: VerifyOperator verifies an operator (admin only)
func (h *Handlers) VerifyOperator(c *fiber.Ctx) error {
	operatorIDStr := c.Params("id")
	operatorID, err := uuid.Parse(operatorIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid operator ID",
		})
	}

	if err := h.operatorService.VerifyOperator(c.Context(), operatorID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to verify operator",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Operator verified successfully",
	})
}

// ReadinessCheck performs a comprehensive readiness check
func (h *Handlers) ReadinessCheck(c *fiber.Ctx) error {
	// Check database connection
	if err := database.HealthCheck(); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":   "not ready",
			"database": "disconnected",
		})
	}

	return c.JSON(fiber.Map{
		"status":   "ready",
		"database": "connected",
	})
}

// UpdateProfile updates the authenticated user's profile
func (h *Handlers) UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	var req struct {
		Username string `json:"username,omitempty"`
		Email    string `json:"email,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	db := database.GetDB()
	updates := make(map[string]interface{})

	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}

	if len(updates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "no fields to update",
		})
	}

	if err := db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update profile",
		})
	}

	// Fetch updated user
	user, err := h.authService.GetUser(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Profile updated successfully",
		"user":    user,
	})
}

// ChangePassword changes the authenticated user's password
func (h *Handlers) ChangePassword(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "old_password and new_password are required",
		})
	}

	if err := h.authService.UpdatePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		if err == auth.ErrInvalidCredentials {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid old password",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update password",
		})
	}

	return c.JSON(fiber.Map{
		"message": "password updated successfully",
	})
}

// GetNode returns a specific node by ID
func (h *Handlers) GetNode(c *fiber.Ctx) error {
	nodeID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid node id",
		})
	}

	var node models.VPNNode
	db := database.GetDB()
	if err := db.First(&node, nodeID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "node not found",
		})
	}

	return c.JSON(node)
}

// CreateSession creates a new VPN session
func (h *Handlers) CreateSession(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	var req struct {
		NodeID   string `json:"node_id"`
		Protocol string `json:"protocol"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	nodeID, err := uuid.Parse(req.NodeID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid node id",
		})
	}

	db := database.GetDB()

	// Check if node exists and is online
	var node models.VPNNode
	if err := db.First(&node, nodeID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "node not found",
		})
	}

	// Create pending session
	// PRIVACY: Never store actual client IP - use anonymized hash for correlation only
	session := models.Session{
		UserID:            userID,
		NodeID:            nodeID,
		Protocol:          req.Protocol,
		Status:            "pending",
		KillSwitchEnabled: true,
		DNSLeakProtection: true,
		ClientIP:          privacyFilter.AnonymizeIP(c.IP()), // Privacy-safe hash, not real IP
		TunnelIP:          "", // Will be assigned by node
	}

	if err := db.Create(&session).Error; err != nil {
		secErr := security.InternalError(err)
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	// Return session (client should poll until active)
	return c.Status(fiber.StatusCreated).JSON(session)
}

// DisconnectSession disconnects an active VPN session
func (h *Handlers) DisconnectSession(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid session id"})
	}

	db := database.GetDB()
	// Update status to pending_disconnect so the node can pick it up
	if err := db.Model(&models.Session{}).
		Where("id = ? AND user_id = ? AND status = ?", sessionID, userID, "active").
		Update("status", "pending_disconnect").Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to disconnect"})
	}

	return c.SendStatus(fiber.StatusOK)
}

// GetSession returns session details
func (h *Handlers) GetSession(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid session id"})
	}

	var session models.Session
	db := database.GetDB()
	if err := db.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "session not found"})
	}

	return c.JSON(session)
}

// GenerateConfig generates VPN configuration for WireGuard peer registration
func (h *Handlers) GenerateConfig(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	var req struct {
		NodeID    string `json:"node_id"`
		PublicKey string `json:"public_key"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.NodeID == "" || req.PublicKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "node_id and public_key are required",
		})
	}

	nodeID, err := uuid.Parse(req.NodeID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid node_id",
		})
	}

	db := database.GetDB()

	// Get node details
	var node models.VPNNode
	if err := db.First(&node, nodeID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "node not found",
		})
	}

	if node.Status != "online" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "node is not online",
		})
	}

	// Check node capacity
	if node.CurrentConnections >= node.MaxConnections {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "node at maximum capacity",
		})
	}

	// Get used IPs for this node
	var usedIPs []string
	db.Model(&models.Session{}).
		Where("node_id = ? AND status IN ?", nodeID, []string{"active", "pending"}).
		Pluck("tunnel_ip", &usedIPs)

	// Allocate client IP (simple allocation from the node's internal subnet)
	clientIP := allocateClientIP(node.InternalIP, usedIPs)
	if clientIP == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "no available IP addresses",
		})
	}

	// Create pending session with client's public key
	// PRIVACY: Never store actual client IP - use anonymized hash for correlation only
	session := models.Session{
		UserID:            userID,
		NodeID:            nodeID,
		Protocol:          "wireguard",
		PublicKey:         req.PublicKey,
		TunnelIP:          clientIP,
		ClientIP:          privacyFilter.AnonymizeIP(c.IP()), // Privacy-safe hash, not real IP
		Status:            "pending",
		KillSwitchEnabled: true,
		DNSLeakProtection: true,
	}

	if err := db.Create(&session).Error; err != nil {
		secErr := security.InternalError(err)
		return c.Status(secErr.HTTPStatus()).JSON(secErr.ClientResponse())
	}

	// Return WireGuard configuration
	// The VPN node will pick up this pending session and add the peer
	serverEndpoint := fmt.Sprintf("%s:%d", node.PublicIP, node.WireGuardPort)
	dns := "1.1.1.1,8.8.8.8"

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"session_id":        session.ID,
		"server_public_key": node.PublicKey,
		"server_endpoint":   serverEndpoint,
		"client_ip":         clientIP,
		"dns":               dns,
		"allowed_ips":       "0.0.0.0/0",
		"keepalive":         25,
	})
}

// allocateClientIP allocates a client IP from the node's subnet
func allocateClientIP(nodeIP string, usedIPs []string) string {
	// Parse node IP to get the base network
	// Assume /24 subnet, node uses .1, clients get .2-.254
	parts := strings.Split(nodeIP, ".")
	if len(parts) != 4 {
		return ""
	}

	base := strings.Join(parts[:3], ".")
	usedSet := make(map[string]bool)
	for _, ip := range usedIPs {
		usedSet[ip] = true
	}

	// Find first available IP (skip .1 which is the node)
	for i := 2; i <= 254; i++ {
		candidateIP := fmt.Sprintf("%s.%d", base, i)
		if !usedSet[candidateIP] {
			return candidateIP
		}
	}

	return ""
}

// GetConfig returns a specific configuration
func (h *Handlers) GetConfig(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	configID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid config id",
		})
	}

	var config models.Config
	if err := database.GetDB().Where("id = ? AND user_id = ?", configID, userID).First(&config).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "config not found",
		})
	}

	return c.JSON(config)
}

// ListConfigs returns all configurations for a user
func (h *Handlers) ListConfigs(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	var configs []models.Config
	if err := database.GetDB().Where("user_id = ?", userID).Find(&configs).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list configs",
		})
	}

	return c.JSON(fiber.Map{
		"configs": configs,
		"count":   len(configs),
	})
}

// Admin handlers

// CreateNode creates a new VPN node (admin only)
func (h *Handlers) CreateNode(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "create node not implemented yet",
	})
}

// UpdateNode updates a VPN node (admin only)
func (h *Handlers) UpdateNode(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "update node not implemented yet",
	})
}

// DeleteNode deletes a VPN node (admin only)
func (h *Handlers) DeleteNode(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "delete node not implemented yet",
	})
}

// GetUser returns a specific user (admin only)
func (h *Handlers) GetUser(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "get user not implemented yet",
	})
}

// UpdateUser updates a user (admin only)
func (h *Handlers) UpdateUser(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "update user not implemented yet",
	})
}

// DeleteUser deletes a user (admin only)
func (h *Handlers) DeleteUser(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "delete user not implemented yet",
	})
}

// GetAllSessions returns all sessions (admin only)
func (h *Handlers) GetAllSessions(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "get all sessions not implemented yet",
	})
}
