package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NodeOperator represents a user who operates a VPN node
type NodeOperator struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Identity
	UserID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// Crypto wallet for rewards
	WalletAddress    string  `gorm:"type:varchar(255);uniqueIndex" json:"wallet_address"`
	WalletType       string  `gorm:"type:varchar(50);default:'ethereum'" json:"wallet_type"` // ethereum, bitcoin, litecoin

	// Operator status
	Status           string  `gorm:"type:varchar(50);default:'pending'" json:"status"` // pending, active, suspended, banned
	IsVerified       bool    `gorm:"default:false" json:"is_verified"`
	VerifiedAt       *time.Time `json:"verified_at,omitempty"`

	// Earnings & Statistics
	TotalEarned      float64 `gorm:"type:decimal(20,8);default:0" json:"total_earned"`      // Total earned in USD
	PendingPayout    float64 `gorm:"type:decimal(20,8);default:0" json:"pending_payout"`    // Pending payout in USD
	LastPayoutAt     *time.Time `json:"last_payout_at,omitempty"`

	// Node statistics
	TotalNodesCreated    int     `gorm:"default:0" json:"total_nodes_created"`
	ActiveNodesCount     int     `gorm:"default:0" json:"active_nodes_count"`
	TotalBandwidthGB     float64 `gorm:"type:decimal(20,4);default:0" json:"total_bandwidth_gb"`
	TotalConnectionsServed int64 `gorm:"default:0" json:"total_connections_served"`

	// Performance metrics
	AverageUptime    float64 `gorm:"type:decimal(5,2);default:0" json:"average_uptime"`     // Percentage
	ReputationScore  float64 `gorm:"type:decimal(5,2);default:50" json:"reputation_score"` // 0-100

	// Staking (optional security deposit)
	StakeAmount      float64 `gorm:"type:decimal(20,8);default:0" json:"stake_amount"`
	StakeStatus      string  `gorm:"type:varchar(50);default:'none'" json:"stake_status"` // none, staked, locked, slashed
	StakedAt         *time.Time `json:"staked_at,omitempty"`

	// Contact & Verification
	Email            string  `gorm:"type:varchar(255)" json:"email"`
	PhoneNumber      string  `gorm:"type:varchar(50)" json:"phone_number,omitempty"`
	Country          string  `gorm:"type:varchar(100)" json:"country"`

	// KYC (optional for high-earning operators)
	KYCStatus        string  `gorm:"type:varchar(50);default:'not_required'" json:"kyc_status"` // not_required, pending, approved, rejected
	KYCSubmittedAt   *time.Time `json:"kyc_submitted_at,omitempty"`

	// Tax information (for legal compliance)
	TaxID            string  `gorm:"type:varchar(100)" json:"tax_id,omitempty"`

	// Relationships
	Nodes            []VPNNode         `gorm:"foreignKey:OperatorID" json:"nodes,omitempty"`
	Earnings         []OperatorEarning `gorm:"foreignKey:OperatorID" json:"earnings,omitempty"`
	Payouts          []OperatorPayout  `gorm:"foreignKey:OperatorID" json:"payouts,omitempty"`
}

// OperatorEarning tracks individual earning events
type OperatorEarning struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	OperatorID uuid.UUID      `gorm:"type:uuid;not null;index" json:"operator_id"`
	Operator   *NodeOperator  `gorm:"foreignKey:OperatorID" json:"operator,omitempty"`

	NodeID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"node_id"`
	Node       *VPNNode       `gorm:"foreignKey:NodeID" json:"node,omitempty"`

	SessionID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"session_id"`
	Session    *Session       `gorm:"foreignKey:SessionID" json:"session,omitempty"`

	// Earning details
	BandwidthGB      float64 `gorm:"type:decimal(20,4);not null" json:"bandwidth_gb"`
	DurationMinutes  int     `gorm:"not null" json:"duration_minutes"`
	RatePerGB        float64 `gorm:"type:decimal(10,6);not null" json:"rate_per_gb"` // USD per GB
	AmountUSD        float64 `gorm:"type:decimal(20,8);not null" json:"amount_usd"`

	// Status
	Status           string  `gorm:"type:varchar(50);default:'pending'" json:"status"` // pending, confirmed, paid
	PaidAt           *time.Time `json:"paid_at,omitempty"`

	// Quality metrics (affects future rates)
	ConnectionQuality float64 `gorm:"type:decimal(5,2)" json:"connection_quality"` // 0-100
	UserRating       int      `gorm:"type:int;check:user_rating >= 1 AND user_rating <= 5" json:"user_rating,omitempty"`
}

// OperatorPayout tracks payout transactions
type OperatorPayout struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	OperatorID uuid.UUID      `gorm:"type:uuid;not null;index" json:"operator_id"`
	Operator   *NodeOperator  `gorm:"foreignKey:OperatorID" json:"operator,omitempty"`

	// Payout details
	AmountUSD        float64 `gorm:"type:decimal(20,8);not null" json:"amount_usd"`
	CryptoAmount     float64 `gorm:"type:decimal(30,18);not null" json:"crypto_amount"`
	CryptoCurrency   string  `gorm:"type:varchar(50);not null" json:"crypto_currency"` // ETH, BTC, LTC
	ExchangeRate     float64 `gorm:"type:decimal(20,8);not null" json:"exchange_rate"` // USD per crypto unit

	// Transaction details
	WalletAddress    string  `gorm:"type:varchar(255);not null" json:"wallet_address"`
	TransactionHash  string  `gorm:"type:varchar(255)" json:"transaction_hash,omitempty"`
	TransactionFee   float64 `gorm:"type:decimal(20,8)" json:"transaction_fee"`

	// Status
	Status           string  `gorm:"type:varchar(50);default:'pending'" json:"status"` // pending, processing, completed, failed
	ProcessedAt      *time.Time `json:"processed_at,omitempty"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	FailureReason    string  `gorm:"type:text" json:"failure_reason,omitempty"`

	// Metadata
	PayoutMethod     string  `gorm:"type:varchar(50)" json:"payout_method"` // blockchain, exchange, manual
	Notes            string  `gorm:"type:text" json:"notes,omitempty"`
}

// NodeReward represents the reward configuration for nodes
type NodeReward struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Reward tiers based on node performance
	TierName         string  `gorm:"type:varchar(50);uniqueIndex" json:"tier_name"` // bronze, silver, gold, platinum
	MinReputationScore float64 `gorm:"type:decimal(5,2)" json:"min_reputation_score"`
	MinUptimePercent  float64 `gorm:"type:decimal(5,2)" json:"min_uptime_percent"`

	// Pricing (USD per GB)
	BaseRatePerGB    float64 `gorm:"type:decimal(10,6);not null" json:"base_rate_per_gb"`
	BonusMultiplier  float64 `gorm:"type:decimal(5,2);default:1.0" json:"bonus_multiplier"`

	// Quality requirements
	MinBandwidth     int     `gorm:"default:100" json:"min_bandwidth"` // Mbps
	MaxLatency       int     `gorm:"default:100" json:"max_latency"`   // ms

	IsActive         bool    `gorm:"default:true" json:"is_active"`
}

// NodePerformanceMetric tracks node performance over time
type NodePerformanceMetric struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	NodeID    uuid.UUID `gorm:"type:uuid;not null;index" json:"node_id"`
	Node      *VPNNode  `gorm:"foreignKey:NodeID" json:"node,omitempty"`

	// Time window
	MetricDate       time.Time `gorm:"type:date;not null;index" json:"metric_date"`
	Hour             int       `gorm:"type:int;check:hour >= 0 AND hour <= 23" json:"hour,omitempty"`

	// Performance data
	UptimeMinutes    int     `gorm:"not null" json:"uptime_minutes"`
	DowntimeMinutes  int     `gorm:"default:0" json:"downtime_minutes"`
	ConnectionsServed int    `gorm:"default:0" json:"connections_served"`
	BandwidthGB      float64 `gorm:"type:decimal(20,4);default:0" json:"bandwidth_gb"`
	AverageLatencyMs int     `gorm:"default:0" json:"average_latency_ms"`

	// Quality scores
	AvailabilityScore float64 `gorm:"type:decimal(5,2)" json:"availability_score"` // 0-100
	PerformanceScore  float64 `gorm:"type:decimal(5,2)" json:"performance_score"`  // 0-100
	UserSatisfaction  float64 `gorm:"type:decimal(5,2)" json:"user_satisfaction"`  // 0-100

	// Earnings for this period
	EarningsUSD      float64 `gorm:"type:decimal(20,8);default:0" json:"earnings_usd"`
}

// CalculateEarnings calculates earnings for a session
func CalculateEarnings(bandwidthGB float64, durationMinutes int, ratePerGB float64, qualityScore float64) float64 {
	baseEarnings := bandwidthGB * ratePerGB

	// Apply quality multiplier (0.5 - 1.5x based on quality)
	qualityMultiplier := 0.5 + (qualityScore / 100.0)

	// Bonus for longer sessions (encourage stability)
	durationBonus := 1.0
	if durationMinutes > 60 {
		durationBonus = 1.1
	}
	if durationMinutes > 180 {
		durationBonus = 1.2
	}

	return baseEarnings * qualityMultiplier * durationBonus
}

// UpdateOperatorStats updates operator statistics
func (op *NodeOperator) UpdateStats(db *gorm.DB) error {
	// Calculate total earnings
	var totalEarned float64
	db.Model(&OperatorEarning{}).
		Where("operator_id = ? AND status = ?", op.ID, "confirmed").
		Select("COALESCE(SUM(amount_usd), 0)").
		Scan(&totalEarned)

	// Calculate total bandwidth from all operator nodes (in KB)
	var totalBandwidthKB int64
	db.Model(&VPNNode{}).
		Where("operator_id = ?", op.ID).
		Select("COALESCE(SUM(total_bandwidth_kb), 0)").
		Scan(&totalBandwidthKB)

	// Calculate pending payout based on bandwidth
	// Rate: $0.10 per GB ($0.0001 per KB)
	// Formula: (totalBandwidthKB / 1024 / 1024) * $0.10
	ratePerKB := 0.0001 / 1024.0 / 1024.0 * 0.10 // $0.10 per GB
	pendingPayout := float64(totalBandwidthKB) * ratePerKB

	// Subtract already paid amounts
	var totalPaid float64
	db.Model(&OperatorEarning{}).
		Where("operator_id = ? AND status = ?", op.ID, "confirmed").
		Select("COALESCE(SUM(amount_usd), 0)").
		Scan(&totalPaid)

	pendingPayout = pendingPayout - totalPaid
	if pendingPayout < 0 {
		pendingPayout = 0
	}

	// Count active nodes
	var activeNodes int64
	db.Model(&VPNNode{}).
		Where("operator_id = ? AND status = ? AND is_active = ?", op.ID, "online", true).
		Count(&activeNodes)

	// Calculate average uptime
	var avgUptime float64
	db.Model(&VPNNode{}).
		Where("operator_id = ?", op.ID).
		Select("COALESCE(AVG(uptime_percentage), 0)").
		Scan(&avgUptime)

	// Update operator record
	return db.Model(op).Updates(map[string]interface{}{
		"total_earned":       totalEarned,
		"pending_payout":     pendingPayout,
		"active_nodes_count": activeNodes,
		"average_uptime":     avgUptime,
		"total_bandwidth_kb": totalBandwidthKB,
	}).Error
}

// CalculateReputationScore calculates operator reputation
func (op *NodeOperator) CalculateReputationScore(db *gorm.DB) float64 {
	score := 50.0 // Base score

	// Uptime contribution (max 30 points)
	score += (op.AverageUptime / 100.0) * 30.0

	// User ratings contribution (max 20 points)
	var avgRating float64
	db.Model(&OperatorEarning{}).
		Where("operator_id = ? AND user_rating > 0", op.ID).
		Select("COALESCE(AVG(user_rating), 0)").
		Scan(&avgRating)
	score += (avgRating / 5.0) * 20.0

	// Bandwidth served contribution (max 10 points)
	if op.TotalBandwidthGB > 1000 {
		score += 10.0
	} else if op.TotalBandwidthGB > 100 {
		score += 5.0
	}

	// Stake contribution (max 10 points)
	if op.StakeAmount >= 1000 {
		score += 10.0
	} else if op.StakeAmount >= 100 {
		score += 5.0
	}

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}

// GetEligibleTier returns the reward tier for this operator
func (op *NodeOperator) GetEligibleTier(db *gorm.DB) (*NodeReward, error) {
	var tier NodeReward
	err := db.Where("is_active = ? AND min_reputation_score <= ? AND min_uptime_percent <= ?",
		true, op.ReputationScore, op.AverageUptime).
		Order("base_rate_per_gb DESC").
		First(&tier).Error

	if err != nil {
		// Return default tier
		return &NodeReward{
			TierName:      "bronze",
			BaseRatePerGB: 0.01, // $0.01 per GB
		}, nil
	}

	return &tier, nil
}
