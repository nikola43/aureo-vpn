package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LocalUser represents a user registered on this node
// Each node maintains its own user database
type LocalUser struct {
	ID           uuid.UUID `gorm:"type:text;primary_key" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	Username     string    `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"not null" json:"-"`

	// Subscription
	Plan      string    `gorm:"default:'free'" json:"plan"` // free, pro, premium
	ExpiresAt time.Time `json:"expires_at,omitempty"`

	// Stats
	TotalDataUsedKB int64 `gorm:"default:0" json:"total_data_used_kb"`
	SessionCount    int   `gorm:"default:0" json:"session_count"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (u *LocalUser) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// LocalSession represents an active VPN session on this node
type LocalSession struct {
	ID     uuid.UUID `gorm:"type:text;primary_key" json:"id"`
	UserID uuid.UUID `gorm:"type:text;index;not null" json:"user_id"`

	// Connection info
	Protocol  string `gorm:"not null" json:"protocol"` // wireguard, openvpn
	TunnelIP  string `json:"tunnel_ip"`
	ClientIP  string `json:"client_ip"`
	PublicKey string `json:"public_key"`

	// Status
	Status        string    `gorm:"default:'pending'" json:"status"` // pending, active, disconnected
	ConnectedAt   time.Time `json:"connected_at,omitempty"`
	DisconnectedAt *time.Time `json:"disconnected_at,omitempty"`
	LastKeepalive time.Time `json:"last_keepalive"`

	// Traffic
	BytesSent     int64 `gorm:"default:0" json:"bytes_sent"`
	BytesReceived int64 `gorm:"default:0" json:"bytes_received"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (s *LocalSession) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// LocalNodeConfig stores this node's configuration
type LocalNodeConfig struct {
	ID    uint   `gorm:"primaryKey"`
	Key   string `gorm:"uniqueIndex;not null"`
	Value string `gorm:"not null"`
}

// NodeIdentity stores this node's identity
type NodeIdentity struct {
	ID             uint      `gorm:"primaryKey"`
	NodeID         string    `gorm:"uniqueIndex;not null"` // UUID
	Name           string    `gorm:"not null"`
	P2PPrivateKey  string    `gorm:"not null"` // libp2p identity key (base64)
	P2PPeerID      string    `gorm:"not null"` // libp2p peer ID
	WGPrivateKey   string    `gorm:"not null"` // WireGuard private key
	WGPublicKey    string    `gorm:"not null"` // WireGuard public key
	CreatedAt      time.Time `json:"created_at"`
}

// AllLocalModels returns all models for SQLite migration
func AllLocalModels() []interface{} {
	return []interface{}{
		&LocalUser{},
		&LocalSession{},
		&LocalNodeConfig{},
		&NodeIdentity{},
	}
}
