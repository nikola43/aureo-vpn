package operator

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nikola43/aureo-vpn/pkg/database"
	apperrors "github.com/nikola43/aureo-vpn/pkg/errors"
	"github.com/nikola43/aureo-vpn/pkg/logger"
	"github.com/nikola43/aureo-vpn/pkg/models"
	"github.com/nikola43/aureo-vpn/pkg/rewards"
	"gorm.io/gorm"
)

// Service handles node operator operations
type Service struct {
	db            *gorm.DB
	log           *logger.Logger
	rewardService *rewards.RewardService
}

// NewService creates a new operator service
func NewService(log *logger.Logger, rewardService *rewards.RewardService) *Service {
	return &Service{
		db:            database.GetDB(),
		log:           log,
		rewardService: rewardService,
	}
}

// RegisterRequest represents operator registration request
type RegisterRequest struct {
	WalletAddress string `json:"wallet_address" validate:"required"`
	WalletType    string `json:"wallet_type" validate:"required,oneof=ethereum bitcoin litecoin"`
	Country       string `json:"country" validate:"required"`
	Email         string `json:"email" validate:"required,email"`
	PhoneNumber   string `json:"phone_number,omitempty"`
}

// NodeCreateRequest represents node creation request
type NodeCreateRequest struct {
	Name             string  `json:"name" validate:"required,min=3,max=100"`
	Hostname         string  `json:"hostname" validate:"required"`
	PublicIP         string  `json:"public_ip" validate:"required,ip"`
	Country          string  `json:"country" validate:"required"`
	CountryCode      string  `json:"country_code" validate:"required,len=2"`
	City             string  `json:"city" validate:"required"`
	WireGuardPort    int     `json:"wireguard_port" validate:"required,min=1,max=65535"`
	OpenVPNPort      int     `json:"openvpn_port" validate:"required,min=1,max=65535"`
	Latitude         float64 `json:"latitude,omitempty"`
	Longitude        float64 `json:"longitude,omitempty"`
	IsOperatorOwned  bool    `json:"is_operator_owned"`
}

// RegisterOperator registers a new node operator
func (s *Service) RegisterOperator(ctx context.Context, userID uuid.UUID, req RegisterRequest) (*models.NodeOperator, error) {
	// Check if user is already an operator
	var existing models.NodeOperator
	err := s.db.Where("user_id = ?", userID).First(&existing).Error
	if err == nil {
		return nil, apperrors.ErrConflict.WithInternal(fmt.Errorf("user is already an operator"))
	}

	// Validate wallet address format (basic validation)
	if len(req.WalletAddress) < 26 {
		return nil, apperrors.ErrValidation.WithInternal(fmt.Errorf("invalid wallet address"))
	}

	// Check if wallet is already registered
	err = s.db.Where("wallet_address = ?", req.WalletAddress).First(&existing).Error
	if err == nil {
		return nil, apperrors.ErrConflict.WithInternal(fmt.Errorf("wallet address already registered"))
	}

	// Create operator
	operator := &models.NodeOperator{
		UserID:          userID,
		WalletAddress:   req.WalletAddress,
		WalletType:      req.WalletType,
		Status:          "pending",
		IsVerified:      false,
		ReputationScore: 50.0, // Start at base score
		Country:         req.Country,
		Email:           req.Email,
		PhoneNumber:     req.PhoneNumber,
	}

	if err := s.db.Create(operator).Error; err != nil {
		return nil, apperrors.ErrDatabase.WithInternal(err)
	}

	s.log.Info("operator registered",
		"operator_id", operator.ID,
		"user_id", userID,
		"wallet_type", req.WalletType,
	)

	return operator, nil
}

// CreateNode creates a new VPN node for an operator
func (s *Service) CreateNode(ctx context.Context, operatorID uuid.UUID, req NodeCreateRequest) (*models.VPNNode, string, error) {
	// Get operator
	var operator models.NodeOperator
	if err := s.db.First(&operator, operatorID).Error; err != nil {
		return nil, "", apperrors.ErrNotFound.WithInternal(err)
	}

	// Check operator status
	if operator.Status != "active" {
		return nil, "", apperrors.ErrForbidden.WithInternal(fmt.Errorf("operator not active"))
	}

	// Check if operator has reached node limit
	maxNodesPerOperator := 10 // Configurable limit
	if operator.TotalNodesCreated >= maxNodesPerOperator {
		return nil, "", apperrors.ErrQuotaExceeded.WithInternal(fmt.Errorf("maximum nodes limit reached"))
	}

	// Generate WireGuard keypair
	// This would be done using the wireguard package
	publicKey := "OPERATOR_NODE_" + uuid.New().String()[:16] // Placeholder

	// Create node
	node := &models.VPNNode{
		Name:                req.Name,
		Hostname:            req.Hostname,
		PublicIP:            req.PublicIP,
		Country:             req.Country,
		CountryCode:         req.CountryCode,
		City:                req.City,
		Latitude:            req.Latitude,
		Longitude:           req.Longitude,
		WireGuardPort:       req.WireGuardPort,
		OpenVPNPort:         req.OpenVPNPort,
		PublicKey:           publicKey,
		Status:              "offline", // Will be online when node connects
		IsActive:            true,
		SupportsWireGuard:   true,
		SupportsOpenVPN:     true,
		MaxConnections:      1000,
		OperatorID:          &operatorID,
		IsOperatorOwned:     true,
		UptimePercentage:    0,
	}

	if err := s.db.Create(node).Error; err != nil {
		return nil, "", apperrors.ErrDatabase.WithInternal(err)
	}

	// Update operator stats
	s.db.Model(&operator).UpdateColumn("total_nodes_created",
		gorm.Expr("total_nodes_created + ?", 1))

	s.log.Info("operator node created",
		"operator_id", operatorID,
		"node_id", node.ID,
		"location", fmt.Sprintf("%s, %s", node.City, node.Country),
	)

	return node, publicKey, nil
}

// GetOperatorStats retrieves comprehensive statistics for an operator
func (s *Service) GetOperatorStats(ctx context.Context, operatorID uuid.UUID) (map[string]interface{}, error) {
	return s.rewardService.GetOperatorStats(ctx, operatorID)
}

// GetOperatorNodes retrieves all nodes for an operator
func (s *Service) GetOperatorNodes(ctx context.Context, operatorID uuid.UUID) ([]models.VPNNode, error) {
	var nodes []models.VPNNode
	err := s.db.Where("operator_id = ?", operatorID).
		Order("created_at DESC").
		Find(&nodes).Error
	return nodes, err
}

// GetOperatorEarnings retrieves earnings history
func (s *Service) GetOperatorEarnings(ctx context.Context, operatorID uuid.UUID, limit, offset int) ([]models.OperatorEarning, int64, error) {
	return s.rewardService.GetOperatorEarnings(ctx, operatorID, limit, offset)
}

// GetOperatorPayouts retrieves payout history
func (s *Service) GetOperatorPayouts(ctx context.Context, operatorID uuid.UUID, limit, offset int) ([]models.OperatorPayout, int64, error) {
	return s.rewardService.GetOperatorPayouts(ctx, operatorID, limit, offset)
}

// RequestPayout requests a manual payout (if threshold not met)
func (s *Service) RequestPayout(ctx context.Context, operatorID uuid.UUID) error {
	var operator models.NodeOperator
	if err := s.db.First(&operator, operatorID).Error; err != nil {
		return apperrors.ErrNotFound.WithInternal(err)
	}

	minPayout := 10.0 // Minimum $10 for payout
	if operator.PendingPayout < minPayout {
		return apperrors.ErrBadRequest.WithInternal(
			fmt.Errorf("minimum payout amount is $%.2f, current: $%.2f", minPayout, operator.PendingPayout))
	}

	// Process payout
	return s.rewardService.ProcessPayouts(ctx, minPayout)
}

// UpdateNodeStatus updates the status of an operator's node
func (s *Service) UpdateNodeStatus(ctx context.Context, nodeID uuid.UUID, status string) error {
	return s.db.Model(&models.VPNNode{}).
		Where("id = ?", nodeID).
		Updates(map[string]interface{}{
			"status":         status,
			"last_heartbeat": time.Now(),
		}).Error
}

// VerifyOperator verifies an operator (admin function)
func (s *Service) VerifyOperator(ctx context.Context, operatorID uuid.UUID) error {
	now := time.Now()
	return s.db.Model(&models.NodeOperator{}).
		Where("id = ?", operatorID).
		Updates(map[string]interface{}{
			"is_verified": true,
			"verified_at": &now,
			"status":      "active",
		}).Error
}

// GetOperatorByUserID retrieves operator by user ID
func (s *Service) GetOperatorByUserID(ctx context.Context, userID uuid.UUID) (*models.NodeOperator, error) {
	var operator models.NodeOperator
	err := s.db.Where("user_id = ?", userID).First(&operator).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound.WithInternal(err)
		}
		return nil, apperrors.ErrDatabase.WithInternal(err)
	}
	return &operator, nil
}

// GetOperatorDashboard retrieves dashboard data
func (s *Service) GetOperatorDashboard(ctx context.Context, operatorID uuid.UUID) (map[string]interface{}, error) {
	var operator models.NodeOperator
	if err := s.db.First(&operator, operatorID).Error; err != nil {
		return nil, apperrors.ErrNotFound.WithInternal(err)
	}

	// Get stats
	stats, err := s.rewardService.GetOperatorStats(ctx, operatorID)
	if err != nil {
		return nil, err
	}

	// Get active nodes
	var activeNodes []models.VPNNode
	s.db.Where("operator_id = ? AND status = ? AND is_active = ?",
		operatorID, "online", true).Find(&activeNodes)

	// Get recent earnings
	earnings, _, _ := s.rewardService.GetOperatorEarnings(ctx, operatorID, 10, 0)

	// Get recent payouts
	payouts, _, _ := s.rewardService.GetOperatorPayouts(ctx, operatorID, 5, 0)

	dashboard := map[string]interface{}{
		"operator":        operator,
		"stats":           stats,
		"active_nodes":    activeNodes,
		"recent_earnings": earnings,
		"recent_payouts":  payouts,
	}

	return dashboard, nil
}

// GetRewardTiers retrieves all reward tiers
func (s *Service) GetRewardTiers(ctx context.Context) ([]models.NodeReward, error) {
	var tiers []models.NodeReward
	err := s.db.Where("is_active = ?", true).
		Order("base_rate_per_gb DESC").
		Find(&tiers).Error
	return tiers, err
}
