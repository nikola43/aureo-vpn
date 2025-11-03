package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a VPN service user
type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email        string         `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"not null" json:"-"`
	Username     string         `gorm:"uniqueIndex;not null" json:"username"`
	FullName     string         `json:"full_name"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	IsAdmin      bool           `gorm:"default:false" json:"is_admin"`

	// Subscription details
	SubscriptionTier   string    `gorm:"default:'free'" json:"subscription_tier"` // free, basic, premium
	SubscriptionExpiry time.Time `json:"subscription_expiry"`

	// Usage tracking
	DataTransferredGB float64 `gorm:"default:0" json:"data_transferred_gb"`
	ConnectionCount   int64   `gorm:"default:0" json:"connection_count"`

	// Security
	TwoFactorEnabled bool   `gorm:"default:false" json:"two_factor_enabled"`
	TwoFactorSecret  string `json:"-"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Sessions []Session `gorm:"foreignKey:UserID" json:"-"`
	Configs  []Config  `gorm:"foreignKey:UserID" json:"-"`
}

// BeforeCreate hook to set UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
