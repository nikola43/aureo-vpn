package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VPNNode represents a VPN server node
type VPNNode struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name     string    `gorm:"uniqueIndex;not null" json:"name"`
	Hostname string    `gorm:"not null" json:"hostname"`

	// Location details
	Country     string  `gorm:"not null;index" json:"country"`
	CountryCode string  `gorm:"size:2;not null" json:"country_code"`
	City        string  `gorm:"not null" json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`

	// Server specs
	PublicIP    string `gorm:"not null" json:"public_ip"`
	InternalIP  string `json:"internal_ip"`
	IPv6Address string `json:"ipv6_address"`

	// Capacity and load
	MaxConnections     int     `gorm:"default:1000" json:"max_connections"`
	CurrentConnections int     `gorm:"default:0" json:"current_connections"`
	CPUUsage           float64 `gorm:"default:0" json:"cpu_usage"`
	MemoryUsage        float64 `gorm:"default:0" json:"memory_usage"`
	BandwidthUsageGbps float64 `gorm:"default:0" json:"bandwidth_usage_gbps"`
	LoadScore          float64 `gorm:"default:0;index" json:"load_score"` // 0-100, lower is better

	// Status and health
	Status         string    `gorm:"default:'offline'" json:"status"` // online, offline, maintenance
	IsActive       bool      `gorm:"default:true" json:"is_active"`
	LastHeartbeat  time.Time `json:"last_heartbeat"`
	LastHealthCheck time.Time `json:"last_health_check"`
	Latency        int       `json:"latency"` // in milliseconds

	// Supported protocols
	SupportsWireGuard bool   `gorm:"default:true" json:"supports_wireguard"`
	SupportsOpenVPN   bool   `gorm:"default:true" json:"supports_openvpn"`
	WireGuardPort     int    `gorm:"default:51820" json:"wireguard_port"`
	OpenVPNPort       int    `gorm:"default:1194" json:"openvpn_port"`
	PublicKey         string `json:"public_key"` // WireGuard public key

	// Features
	SupportsMultiHop   bool `gorm:"default:false" json:"supports_multihop"`
	SupportsObfuscation bool `gorm:"default:false" json:"supports_obfuscation"`
	SupportsSOCKS5     bool `gorm:"default:false" json:"supports_socks5"`

	// Metadata
	Version   string `json:"version"`
	Tags      string `json:"tags"` // comma-separated
	Priority  int    `gorm:"default:0" json:"priority"`

	// Operator ownership (for decentralized nodes)
	OperatorID       *uuid.UUID `gorm:"type:uuid;index" json:"operator_id,omitempty"`
	IsOperatorOwned  bool       `gorm:"default:false" json:"is_operator_owned"`
	UptimePercentage float64    `gorm:"type:decimal(5,2);default:0" json:"uptime_percentage"`

	// Earnings (for operator nodes)
	TotalEarnedUSD   float64    `gorm:"type:decimal(20,8);default:0" json:"total_earned_usd,omitempty"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Sessions []Session `gorm:"foreignKey:NodeID" json:"-"`
}

// BeforeCreate hook
func (n *VPNNode) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

// CalculateLoadScore calculates the load score based on connections, CPU, and memory
func (n *VPNNode) CalculateLoadScore() float64 {
	connectionLoad := float64(n.CurrentConnections) / float64(n.MaxConnections) * 100
	cpuLoad := n.CPUUsage
	memoryLoad := n.MemoryUsage

	// Weighted average: 40% connections, 30% CPU, 30% memory
	return (connectionLoad * 0.4) + (cpuLoad * 0.3) + (memoryLoad * 0.3)
}

// IsHealthy checks if the node is healthy
func (n *VPNNode) IsHealthy() bool {
	if !n.IsActive || n.Status != "online" {
		return false
	}

	// Check if heartbeat is recent (within last 2 minutes)
	if time.Since(n.LastHeartbeat) > 2*time.Minute {
		return false
	}

	// Check if load is acceptable
	if n.LoadScore > 90 {
		return false
	}

	return true
}
