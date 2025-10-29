package rewards

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nikola43/aureo-vpn/pkg/blockchain"
	"github.com/nikola43/aureo-vpn/pkg/database"
	"github.com/nikola43/aureo-vpn/pkg/logger"
	"github.com/nikola43/aureo-vpn/pkg/models"
	"gorm.io/gorm"
)

// RewardService handles crypto rewards for node operators
type RewardService struct {
	db         *gorm.DB
	log        *logger.Logger
	blockchain *blockchain.Service
}

// NewRewardService creates a new reward service
func NewRewardService(log *logger.Logger, blockchainService *blockchain.Service) *RewardService {
	return &RewardService{
		db:         database.GetDB(),
		log:        log,
		blockchain: blockchainService,
	}
}

// RewardTiers defines the standard reward tiers
var DefaultRewardTiers = []models.NodeReward{
	{
		TierName:           "bronze",
		MinReputationScore: 0,
		MinUptimePercent:   50,
		BaseRatePerGB:      0.01,  // $0.01 per GB
		BonusMultiplier:    1.0,
		MinBandwidth:       50,
		MaxLatency:         150,
		IsActive:           true,
	},
	{
		TierName:           "silver",
		MinReputationScore: 60,
		MinUptimePercent:   80,
		BaseRatePerGB:      0.015, // $0.015 per GB
		BonusMultiplier:    1.2,
		MinBandwidth:       100,
		MaxLatency:         100,
		IsActive:           true,
	},
	{
		TierName:           "gold",
		MinReputationScore: 75,
		MinUptimePercent:   90,
		BaseRatePerGB:      0.02,  // $0.02 per GB
		BonusMultiplier:    1.5,
		MinBandwidth:       200,
		MaxLatency:         75,
		IsActive:           true,
	},
	{
		TierName:           "platinum",
		MinReputationScore: 90,
		MinUptimePercent:   95,
		BaseRatePerGB:      0.03,  // $0.03 per GB
		BonusMultiplier:    2.0,
		MinBandwidth:       500,
		MaxLatency:         50,
		IsActive:           true,
	},
}

// InitializeRewardTiers creates default reward tiers
func (rs *RewardService) InitializeRewardTiers() error {
	for _, tier := range DefaultRewardTiers {
		var existing models.NodeReward
		err := rs.db.Where("tier_name = ?", tier.TierName).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := rs.db.Create(&tier).Error; err != nil {
				return fmt.Errorf("failed to create tier %s: %w", tier.TierName, err)
			}
			rs.log.Info("reward tier created", "tier", tier.TierName, "rate_per_gb", tier.BaseRatePerGB)
		}
	}
	return nil
}

// RecordEarning records an earning event for a session
func (rs *RewardService) RecordEarning(ctx context.Context, sessionID uuid.UUID, bandwidthGB float64, durationMinutes int) error {
	// Get session details
	var session models.Session
	if err := rs.db.Preload("Node").First(&session, sessionID).Error; err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Check if node is operator-owned
	if session.Node.OperatorID == nil {
		rs.log.Debug("session on company node, no earnings", "session_id", sessionID)
		return nil // Company-owned nodes don't generate operator earnings
	}

	// Get operator
	var operator models.NodeOperator
	if err := rs.db.First(&operator, session.Node.OperatorID).Error; err != nil {
		return fmt.Errorf("operator not found: %w", err)
	}

	// Get eligible tier
	tier, err := operator.GetEligibleTier(rs.db)
	if err != nil {
		return fmt.Errorf("failed to get tier: %w", err)
	}

	// Calculate quality score based on session
	qualityScore := calculateSessionQuality(&session)

	// Calculate earnings
	amountUSD := models.CalculateEarnings(
		bandwidthGB,
		durationMinutes,
		tier.BaseRatePerGB*tier.BonusMultiplier,
		qualityScore,
	)

	// Create earning record
	earning := models.OperatorEarning{
		OperatorID:        *session.Node.OperatorID,
		NodeID:            session.NodeID,
		SessionID:         sessionID,
		BandwidthGB:       bandwidthGB,
		DurationMinutes:   durationMinutes,
		RatePerGB:         tier.BaseRatePerGB * tier.BonusMultiplier,
		AmountUSD:         amountUSD,
		Status:            "pending",
		ConnectionQuality: qualityScore,
	}

	if err := rs.db.Create(&earning).Error; err != nil {
		return fmt.Errorf("failed to record earning: %w", err)
	}

	rs.log.Info("earning recorded",
		"operator_id", operator.ID,
		"session_id", sessionID,
		"amount_usd", amountUSD,
		"tier", tier.TierName,
	)

	// Update node total earnings
	rs.db.Model(&session.Node).UpdateColumn("total_earned_usd",
		gorm.Expr("total_earned_usd + ?", amountUSD))

	// Update operator stats
	go operator.UpdateStats(rs.db)

	return nil
}

// ConfirmEarnings confirms pending earnings (called after quality verification)
func (rs *RewardService) ConfirmEarnings(ctx context.Context, earningID uuid.UUID) error {
	return rs.db.Model(&models.OperatorEarning{}).
		Where("id = ? AND status = ?", earningID, "pending").
		Updates(map[string]interface{}{
			"status": "confirmed",
		}).Error
}

// ProcessPayouts processes pending payouts for operators
func (rs *RewardService) ProcessPayouts(ctx context.Context, minPayoutAmount float64) error {
	// Find operators with sufficient pending payouts
	var operators []models.NodeOperator
	err := rs.db.Where("pending_payout >= ? AND status = ?", minPayoutAmount, "active").
		Find(&operators).Error
	if err != nil {
		return err
	}

	for _, operator := range operators {
		if err := rs.createPayout(ctx, &operator); err != nil {
			rs.log.Error("failed to create payout",
				"operator_id", operator.ID,
				"error", err,
			)
			continue
		}
	}

	return nil
}

// createPayout creates a payout for an operator
func (rs *RewardService) createPayout(ctx context.Context, operator *models.NodeOperator) error {
	// Get crypto exchange rate
	exchangeRate, cryptoAmount, err := rs.getCryptoConversion(operator.WalletType, operator.PendingPayout)
	if err != nil {
		return fmt.Errorf("failed to get exchange rate: %w", err)
	}

	// Create payout record
	payout := models.OperatorPayout{
		OperatorID:     operator.ID,
		AmountUSD:      operator.PendingPayout,
		CryptoAmount:   cryptoAmount,
		CryptoCurrency: operator.WalletType,
		ExchangeRate:   exchangeRate,
		WalletAddress:  operator.WalletAddress,
		Status:         "pending",
		PayoutMethod:   "blockchain",
	}

	if err := rs.db.Create(&payout).Error; err != nil {
		return err
	}

	rs.log.Info("payout created",
		"operator_id", operator.ID,
		"amount_usd", payout.AmountUSD,
		"crypto_amount", cryptoAmount,
		"currency", operator.WalletType,
	)

	// In production, this would trigger actual blockchain transaction
	// For now, we'll mark it as processing
	go rs.executeBlockchainTransaction(ctx, &payout)

	return nil
}

// executeBlockchainTransaction executes the actual blockchain transaction
func (rs *RewardService) executeBlockchainTransaction(ctx context.Context, payout *models.OperatorPayout) {
	// Update status to processing
	rs.db.Model(payout).Updates(map[string]interface{}{
		"status":       "processing",
		"processed_at": time.Now(),
	})

	rs.log.Info("executing blockchain transaction",
		"payout_id", payout.ID,
		"wallet_type", payout.CryptoCurrency,
		"amount_usd", payout.AmountUSD,
		"wallet_address", payout.WalletAddress,
	)

	// Execute blockchain transaction
	var tx *blockchain.Transaction
	var err error

	if rs.blockchain != nil {
		// Use real blockchain service
		tx, err = rs.blockchain.SendTransaction(ctx, payout.CryptoCurrency, payout.WalletAddress, payout.AmountUSD)
		if err != nil {
			rs.log.Error("blockchain transaction failed",
				"payout_id", payout.ID,
				"error", err,
			)
			rs.db.Model(payout).Updates(map[string]interface{}{
				"status":         "failed",
				"failure_reason": err.Error(),
			})
			return
		}
	} else {
		// Fallback to mock transaction if blockchain service is not configured
		rs.log.Warn("blockchain service not configured, using mock transaction", "payout_id", payout.ID)
		time.Sleep(2 * time.Second)
		tx = &blockchain.Transaction{
			TxHash:         fmt.Sprintf("MOCK_%s", uuid.New().String()[:16]),
			BlockchainType: payout.CryptoCurrency,
			To:             payout.WalletAddress,
			Status:         "pending",
		}
	}

	// Update payout with transaction hash
	rs.db.Model(payout).Updates(map[string]interface{}{
		"transaction_hash": tx.TxHash,
		"status":           "processing",
	})

	// Wait for transaction confirmation (simplified - in production, use webhooks or polling)
	if rs.blockchain != nil {
		// Poll for transaction status
		maxAttempts := 30 // 5 minutes with 10-second intervals
		for i := 0; i < maxAttempts; i++ {
			time.Sleep(10 * time.Second)

			status, err := rs.blockchain.GetTransactionStatus(ctx, payout.CryptoCurrency, tx.TxHash)
			if err != nil {
				rs.log.Warn("failed to check transaction status",
					"payout_id", payout.ID,
					"tx_hash", tx.TxHash,
					"attempt", i+1,
					"error", err,
				)
				continue
			}

			if status.Status == "confirmed" {
				rs.log.Info("transaction confirmed",
					"payout_id", payout.ID,
					"tx_hash", tx.TxHash,
					"confirmations", status.Confirmations,
				)
				break
			} else if status.Status == "failed" {
				err := fmt.Errorf("transaction failed: %s", status.ErrorMessage)
				rs.log.Error("transaction failed on blockchain",
					"payout_id", payout.ID,
					"tx_hash", tx.TxHash,
					"error", err,
				)
				rs.db.Model(payout).Updates(map[string]interface{}{
					"status":         "failed",
					"failure_reason": status.ErrorMessage,
				})
				return
			}
		}
	}

	// Mark payout as completed
	now := time.Now()
	err = rs.db.Model(payout).Updates(map[string]interface{}{
		"status":       "completed",
		"completed_at": &now,
	}).Error

	if err != nil {
		rs.log.Error("failed to update payout status", "payout_id", payout.ID, "error", err)
		return
	}

	// Update operator stats
	var operator models.NodeOperator
	if err := rs.db.First(&operator, payout.OperatorID).Error; err == nil {
		rs.db.Model(&operator).Updates(map[string]interface{}{
			"pending_payout": gorm.Expr("pending_payout - ?", payout.AmountUSD),
			"last_payout_at": &now,
		})

		// Mark earnings as paid
		rs.db.Model(&models.OperatorEarning{}).
			Where("operator_id = ? AND status = ?", operator.ID, "confirmed").
			Updates(map[string]interface{}{
				"status":  "paid",
				"paid_at": &now,
			})
	}

	rs.log.Info("payout completed successfully",
		"operator_id", payout.OperatorID,
		"tx_hash", tx.TxHash,
		"amount_usd", payout.AmountUSD,
		"crypto_amount", payout.CryptoAmount,
	)
}

// getCryptoConversion gets the current exchange rate and calculates crypto amount
func (rs *RewardService) getCryptoConversion(cryptoType string, amountUSD float64) (rate float64, cryptoAmount float64, err error) {
	// TODO: Integrate with real price API (CoinGecko, CoinMarketCap, etc.)
	// For now, using mock rates
	rates := map[string]float64{
		"ethereum": 2000.0,  // 1 ETH = $2000
		"bitcoin":  40000.0, // 1 BTC = $40000
		"litecoin": 100.0,   // 1 LTC = $100
	}

	rate, ok := rates[cryptoType]
	if !ok {
		return 0, 0, fmt.Errorf("unsupported crypto type: %s", cryptoType)
	}

	cryptoAmount = amountUSD / rate
	return rate, cryptoAmount, nil
}

// calculateSessionQuality calculates quality score for a session
func calculateSessionQuality(session *models.Session) float64 {
	score := 100.0

	// Deduct for poor connection
	if session.Node.Latency > 100 {
		score -= 10.0
	}
	if session.Node.Latency > 200 {
		score -= 20.0
	}

	// Deduct for short sessions (prefer stability)
	duration := time.Since(session.ConnectedAt)
	if duration < 5*time.Minute {
		score -= 20.0
	}

	// Bonus for long stable sessions
	if duration > time.Hour {
		score += 10.0
	}

	// Cap at 0-100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// UpdateOperatorReputation updates operator reputation score
func (rs *RewardService) UpdateOperatorReputation(ctx context.Context, operatorID uuid.UUID) error {
	var operator models.NodeOperator
	if err := rs.db.First(&operator, operatorID).Error; err != nil {
		return err
	}

	newScore := operator.CalculateReputationScore(rs.db)

	return rs.db.Model(&operator).UpdateColumn("reputation_score", newScore).Error
}

// GetOperatorEarnings retrieves earnings for an operator
func (rs *RewardService) GetOperatorEarnings(ctx context.Context, operatorID uuid.UUID, limit, offset int) ([]models.OperatorEarning, int64, error) {
	var earnings []models.OperatorEarning
	var total int64

	rs.db.Model(&models.OperatorEarning{}).
		Where("operator_id = ?", operatorID).
		Count(&total)

	err := rs.db.Where("operator_id = ?", operatorID).
		Preload("Node").
		Preload("Session").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&earnings).Error

	return earnings, total, err
}

// GetOperatorPayouts retrieves payout history for an operator
func (rs *RewardService) GetOperatorPayouts(ctx context.Context, operatorID uuid.UUID, limit, offset int) ([]models.OperatorPayout, int64, error) {
	var payouts []models.OperatorPayout
	var total int64

	rs.db.Model(&models.OperatorPayout{}).
		Where("operator_id = ?", operatorID).
		Count(&total)

	err := rs.db.Where("operator_id = ?", operatorID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&payouts).Error

	return payouts, total, err
}

// GetOperatorStats retrieves comprehensive stats for an operator
func (rs *RewardService) GetOperatorStats(ctx context.Context, operatorID uuid.UUID) (map[string]interface{}, error) {
	var operator models.NodeOperator
	if err := rs.db.First(&operator, operatorID).Error; err != nil {
		return nil, err
	}

	// Get earnings breakdown
	var earningsToday, earningsWeek, earningsMonth float64
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	weekAgo := today.AddDate(0, 0, -7)
	monthAgo := today.AddDate(0, -1, 0)

	rs.db.Model(&models.OperatorEarning{}).
		Where("operator_id = ? AND created_at >= ?", operatorID, today).
		Select("COALESCE(SUM(amount_usd), 0)").
		Scan(&earningsToday)

	rs.db.Model(&models.OperatorEarning{}).
		Where("operator_id = ? AND created_at >= ?", operatorID, weekAgo).
		Select("COALESCE(SUM(amount_usd), 0)").
		Scan(&earningsWeek)

	rs.db.Model(&models.OperatorEarning{}).
		Where("operator_id = ? AND created_at >= ?", operatorID, monthAgo).
		Select("COALESCE(SUM(amount_usd), 0)").
		Scan(&earningsMonth)

	// Get current tier
	tier, _ := operator.GetEligibleTier(rs.db)

	// Get connected users and current traffic from all operator nodes
	var nodes []models.VPNNode
	rs.db.Where("operator_id = ? AND status = ?", operatorID, "online").Find(&nodes)

	var totalConnectedUsers int
	var totalCurrentTrafficMbps float64
	for _, node := range nodes {
		totalConnectedUsers += node.CurrentConnections
		totalCurrentTrafficMbps += node.BandwidthUsageGbps * 1000.0 // Convert Gbps to Mbps
	}

	stats := map[string]interface{}{
		"operator_id":       operator.ID,
		"total_earned":      operator.TotalEarned,
		"pending_payout":    operator.PendingPayout,
		"earnings_today":    earningsToday,
		"earnings_week":     earningsWeek,
		"earnings_month":    earningsMonth,
		"active_nodes":      operator.ActiveNodesCount,
		"total_bandwidth":   operator.TotalBandwidthGB,
		"reputation_score":  operator.ReputationScore,
		"average_uptime":    operator.AverageUptime,
		"current_tier":      tier.TierName,
		"rate_per_gb":       tier.BaseRatePerGB * tier.BonusMultiplier,
		"connected_users":   totalConnectedUsers,
		"current_traffic":   totalCurrentTrafficMbps,
	}

	return stats, nil
}
