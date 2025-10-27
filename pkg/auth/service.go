package auth

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/nikola43/aureo-vpn/pkg/crypto"
	"github.com/nikola43/aureo-vpn/pkg/database"
	"github.com/nikola43/aureo-vpn/pkg/models"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists        = errors.New("user already exists")
	ErrInactiveUser      = errors.New("user account is inactive")
)

// Service handles authentication operations
type Service struct {
	db             *gorm.DB
	tokenService   *TokenService
	passwordHasher *crypto.PasswordHasher
}

// NewService creates a new authentication service
func NewService(tokenService *TokenService) *Service {
	return &Service{
		db:             database.GetDB(),
		tokenService:   tokenService,
		passwordHasher: crypto.NewPasswordHasher(),
	}
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Username string `json:"username" validate:"required,min=3,max=50"`
	FullName string `json:"full_name"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         *models.User `json:"user"`
}

// Register creates a new user account
func (s *Service) Register(req RegisterRequest) (*AuthResponse, error) {
	// Check if user already exists
	var existingUser models.User
	if err := s.db.Where("email = ? OR username = ?", req.Email, req.Username).First(&existingUser).Error; err == nil {
		return nil, ErrUserExists
	}

	// Hash password
	passwordHash, err := s.passwordHasher.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := models.User{
		Email:        req.Email,
		Username:     req.Username,
		FullName:     req.FullName,
		PasswordHash: passwordHash,
		IsActive:     true,
		SubscriptionTier: "free",
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	accessToken, refreshToken, err := s.tokenService.GenerateTokenPair(
		user.ID,
		user.Email,
		user.Username,
		user.IsAdmin,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         &user,
	}, nil
}

// Login authenticates a user and returns tokens
func (s *Service) Login(req LoginRequest) (*AuthResponse, error) {
	// Find user by email
	var user models.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrInactiveUser
	}

	// Verify password
	valid, err := s.passwordHasher.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to verify password: %w", err)
	}

	if !valid {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, refreshToken, err := s.tokenService.GenerateTokenPair(
		user.ID,
		user.Email,
		user.Username,
		user.IsAdmin,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         &user,
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *Service) RefreshToken(refreshToken string) (string, error) {
	return s.tokenService.RefreshAccessToken(refreshToken)
}

// VerifyToken verifies a JWT token and returns claims
func (s *Service) VerifyToken(token string) (*Claims, error) {
	return s.tokenService.VerifyToken(token)
}

// GetUser retrieves a user by ID
func (s *Service) GetUser(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// UpdatePassword updates a user's password
func (s *Service) UpdatePassword(userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.GetUser(userID)
	if err != nil {
		return err
	}

	// Verify old password
	valid, err := s.passwordHasher.VerifyPassword(oldPassword, user.PasswordHash)
	if err != nil {
		return err
	}

	if !valid {
		return ErrInvalidCredentials
	}

	// Hash new password
	newHash, err := s.passwordHasher.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password
	return s.db.Model(&user).Update("password_hash", newHash).Error
}
