package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Config represents VPN configuration files for users
type Config struct {
	ID     uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	NodeID uuid.UUID `gorm:"type:uuid;not null;index" json:"node_id"`

	// Config details
	Protocol       string `gorm:"not null" json:"protocol"` // wireguard, openvpn
	ConfigName     string `gorm:"not null" json:"config_name"`
	ConfigContent  string `gorm:"type:text;not null" json:"-"` // Encrypted config file content
	ConfigHash     string `gorm:"not null" json:"config_hash"`

	// Keys (encrypted at rest)
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"-"`

	// Settings
	DNSServers         string `json:"dns_servers"` // comma-separated
	AllowedIPs         string `json:"allowed_ips"` // for split tunneling
	MTU                int    `gorm:"default:1420" json:"mtu"`
	PersistentKeepalive int   `gorm:"default:25" json:"persistent_keepalive"`

	// Status
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	LastUsed     *time.Time `json:"last_used,omitempty"`
	TimesUsed    int64      `gorm:"default:0" json:"times_used"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User *User    `gorm:"foreignKey:UserID" json:"-"`
	Node *VPNNode `gorm:"foreignKey:NodeID" json:"-"`
}

// BeforeCreate hook
func (c *Config) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// IsExpired checks if the config has expired
func (c *Config) IsExpired() bool {
	if c.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*c.ExpiresAt)
}

// IsValid checks if config is valid for use
func (c *Config) IsValid() bool {
	return c.IsActive && !c.IsExpired()
}
