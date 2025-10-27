package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Session represents an active VPN connection session
type Session struct {
	ID     uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	NodeID uuid.UUID `gorm:"type:uuid;not null;index" json:"node_id"`

	// Session details
	Protocol      string    `gorm:"not null" json:"protocol"` // wireguard, openvpn
	ClientIP      string    `gorm:"not null" json:"client_ip"`
	TunnelIP      string    `gorm:"not null" json:"tunnel_ip"`
	PublicKey     string    `json:"public_key"`     // For WireGuard
	PrivateKey    string    `json:"-"`              // Encrypted, never exposed
	Status        string    `gorm:"default:'active'" json:"status"` // active, disconnected, terminated
	ConnectedAt   time.Time `gorm:"not null" json:"connected_at"`
	DisconnectedAt *time.Time `json:"disconnected_at,omitempty"`

	// Data transfer tracking
	BytesSent     int64   `gorm:"default:0" json:"bytes_sent"`
	BytesReceived int64   `gorm:"default:0" json:"bytes_received"`
	DataUsedGB    float64 `gorm:"default:0" json:"data_used_gb"`

	// Connection quality
	Latency       int `json:"latency"`        // milliseconds
	PacketLoss    float64 `json:"packet_loss"` // percentage
	LastKeepalive time.Time `json:"last_keepalive"`

	// Features enabled
	SplitTunnelEnabled bool   `gorm:"default:false" json:"split_tunnel_enabled"`
	KillSwitchEnabled  bool   `gorm:"default:true" json:"kill_switch_enabled"`
	DNSLeakProtection  bool   `gorm:"default:true" json:"dns_leak_protection"`
	IsMultiHop         bool   `gorm:"default:false" json:"is_multihop"`
	NextHopNodeID      *uuid.UUID `gorm:"type:uuid" json:"next_hop_node_id,omitempty"`

	// Metadata
	ClientVersion string `json:"client_version"`
	DeviceType    string `json:"device_type"` // desktop, mobile, router
	OSType        string `json:"os_type"`     // linux, macos, windows, ios, android

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User *User    `gorm:"foreignKey:UserID" json:"-"`
	Node *VPNNode `gorm:"foreignKey:NodeID" json:"-"`
}

// BeforeCreate hook
func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.ConnectedAt.IsZero() {
		s.ConnectedAt = time.Now()
	}
	return nil
}

// UpdateDataUsage calculates data usage in GB
func (s *Session) UpdateDataUsage() {
	totalBytes := float64(s.BytesSent + s.BytesReceived)
	s.DataUsedGB = totalBytes / (1024 * 1024 * 1024) // Convert to GB
}

// IsActive checks if session is currently active
func (s *Session) IsActive() bool {
	if s.Status != "active" {
		return false
	}
	// Check if last keepalive was recent (within last 5 minutes)
	return time.Since(s.LastKeepalive) < 5*time.Minute
}

// Duration returns the session duration
func (s *Session) Duration() time.Duration {
	if s.DisconnectedAt != nil {
		return s.DisconnectedAt.Sub(s.ConnectedAt)
	}
	return time.Since(s.ConnectedAt)
}
