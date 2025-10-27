package api

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nikola43/aureo-vpn/pkg/auth"
	"github.com/nikola43/aureo-vpn/pkg/database"
	apperrors "github.com/nikola43/aureo-vpn/pkg/errors"
	"github.com/nikola43/aureo-vpn/pkg/models"
	"github.com/nikola43/aureo-vpn/pkg/metrics"
	"github.com/nikola43/aureo-vpn/pkg/operator"
)

// Handlers holds all API handlers
type Handlers struct {
	authService     *auth.Service
	operatorService *operator.Service
}

// NewHandlers creates new API handlers
func NewHandlers(authService *auth.Service, operatorService *operator.Service) *Handlers {
	return &Handlers{
		authService:     authService,
		operatorService: operatorService,
	}
}

// Register handles user registration
func (h *Handlers) Register(c *fiber.Ctx) error {
	var req auth.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	resp, err := h.authService.Register(req)
	if err != nil {
		metrics.LoginAttempts.WithLabelValues("failed").Inc()
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	metrics.UserRegistrations.Inc()
	metrics.LoginAttempts.WithLabelValues("success").Inc()

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// Login handles user login
func (h *Handlers) Login(c *fiber.Ctx) error {
	var req auth.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	resp, err := h.authService.Login(req)
	if err != nil {
		metrics.LoginAttempts.WithLabelValues("failed").Inc()
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	accessToken, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
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
func (h *Handlers) ListNodes(c *fiber.Ctx) error {
	db := database.GetDB()

	var nodes []models.VPNNode
	query := db.Where("is_active = ? AND status = ?", true, "online")

	// Optional filters
	if country := c.Query("country"); country != "" {
		query = query.Where("country_code = ?", country)
	}

	if protocol := c.Query("protocol"); protocol != "" {
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
		"nodes": nodes,
		"count": len(nodes),
	})
}

// GetBestNode returns the best available node based on load and latency
func (h *Handlers) GetBestNode(c *fiber.Ctx) error {
	db := database.GetDB()

	protocol := c.Query("protocol", "wireguard")
	country := c.Query("country")

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
		TotalUsers        int64 `json:"total_users"`
		ActiveUsers       int64 `json:"active_users"`
		TotalNodes        int64 `json:"total_nodes"`
		OnlineNodes       int64 `json:"online_nodes"`
		TotalSessions     int64 `json:"total_sessions"`
		ActiveSessions    int64 `json:"active_sessions"`
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
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "change password not implemented yet",
	})
}

// GetNode returns a specific node by ID
func (h *Handlers) GetNode(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "get node not implemented yet",
	})
}

// CreateSession creates a new VPN session
func (h *Handlers) CreateSession(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "create session not implemented yet",
	})
}

// DisconnectSession disconnects an active VPN session
func (h *Handlers) DisconnectSession(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "disconnect session not implemented yet",
	})
}

// GetSession returns session details
func (h *Handlers) GetSession(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "get session not implemented yet",
	})
}

// GenerateConfig generates VPN configuration
func (h *Handlers) GenerateConfig(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "generate config not implemented yet",
	})
}

// GetConfig returns a specific configuration
func (h *Handlers) GetConfig(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "get config not implemented yet",
	})
}

// ListConfigs returns all configurations for a user
func (h *Handlers) ListConfigs(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "list configs not implemented yet",
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
