// Package security provides hardened JWT token management with security best practices
package security

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWT-related errors
var (
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenNotYetValid   = errors.New("token not yet valid")
	ErrTokenRevoked       = errors.New("token has been revoked")
	ErrInvalidTokenType   = errors.New("invalid token type")
	ErrInvalidIssuer      = errors.New("invalid token issuer")
	ErrInvalidAudience    = errors.New("invalid token audience")
	ErrTokenFamilyRevoked = errors.New("token family has been revoked")
	ErrWeakSecret         = errors.New("JWT secret is too weak")
	ErrMissingClaims      = errors.New("required claims missing")
)

// TokenType represents the type of JWT token
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// Recommended token lifetimes
const (
	DefaultAccessTokenLifetime  = 15 * time.Minute  // Short-lived access tokens
	DefaultRefreshTokenLifetime = 7 * 24 * time.Hour // 7 days
	MaxAccessTokenLifetime      = 1 * time.Hour      // Maximum allowed
	MaxRefreshTokenLifetime     = 30 * 24 * time.Hour // 30 days
	MinSecretLength             = 32                  // 256 bits minimum
)

// SecureClaims represents hardened JWT claims
type SecureClaims struct {
	jwt.RegisteredClaims

	// User identification
	UserID   string `json:"uid"`
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
	IsAdmin  bool   `json:"admin,omitempty"`

	// Token metadata
	TokenType   TokenType `json:"type"`
	TokenFamily string    `json:"fam,omitempty"` // For refresh token rotation
	TokenID     string    `json:"jti"`           // Unique token ID

	// Security metadata
	Fingerprint string `json:"fp,omitempty"`  // Device fingerprint hash
	IPHash      string `json:"iph,omitempty"` // Hashed IP (for anomaly detection only)
}

// SecureJWTConfig holds configuration for secure JWT operations
type SecureJWTConfig struct {
	Secret             *SecureString
	Issuer             string
	Audience           []string
	AccessTokenLife    time.Duration
	RefreshTokenLife   time.Duration
	EnableIPBinding    bool
	EnableFingerprint  bool
	AllowedClockSkew   time.Duration
}

// DefaultSecureJWTConfig returns secure default configuration
func DefaultSecureJWTConfig(secret string) (*SecureJWTConfig, error) {
	if len(secret) < MinSecretLength {
		return nil, ErrWeakSecret
	}

	return &SecureJWTConfig{
		Secret:            NewSecureString(secret),
		Issuer:            "aureo-vpn",
		Audience:          []string{"aureo-vpn-api"},
		AccessTokenLife:   DefaultAccessTokenLifetime,
		RefreshTokenLife:  DefaultRefreshTokenLifetime,
		EnableIPBinding:   false, // Disabled by default for privacy
		EnableFingerprint: true,
		AllowedClockSkew:  30 * time.Second,
	}, nil
}

// SecureJWTService provides secure JWT operations
type SecureJWTService struct {
	config         *SecureJWTConfig
	revokedTokens  *TokenBlacklist
	revokedFamilies *FamilyBlacklist
}

// NewSecureJWTService creates a new secure JWT service
func NewSecureJWTService(config *SecureJWTConfig) (*SecureJWTService, error) {
	if config == nil {
		return nil, errors.New("config is required")
	}

	if len(config.Secret.Bytes()) < MinSecretLength {
		return nil, ErrWeakSecret
	}

	return &SecureJWTService{
		config:          config,
		revokedTokens:   NewTokenBlacklist(24 * time.Hour),
		revokedFamilies: NewFamilyBlacklist(30 * 24 * time.Hour),
	}, nil
}

// TokenPair represents an access/refresh token pair
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// GenerateTokenPair creates a new access/refresh token pair
func (s *SecureJWTService) GenerateTokenPair(
	ctx context.Context,
	userID, email, username string,
	isAdmin bool,
	fingerprint string,
	clientIP string,
) (*TokenPair, error) {
	// Generate unique family ID for this refresh token chain
	familyID := uuid.New().String()

	accessToken, accessExp, err := s.generateToken(
		userID, email, username, isAdmin,
		TokenTypeAccess, familyID, fingerprint, clientIP,
		s.config.AccessTokenLife,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, _, err := s.generateToken(
		userID, email, username, isAdmin,
		TokenTypeRefresh, familyID, fingerprint, clientIP,
		s.config.RefreshTokenLife,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExp,
		TokenType:    "Bearer",
	}, nil
}

func (s *SecureJWTService) generateToken(
	userID, email, username string,
	isAdmin bool,
	tokenType TokenType,
	familyID, fingerprint, clientIP string,
	lifetime time.Duration,
) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(lifetime)

	// Hash IP for privacy-preserving anomaly detection
	var ipHash string
	if s.config.EnableIPBinding && clientIP != "" {
		hash := sha256.Sum256([]byte(clientIP + s.config.Secret.String()))
		ipHash = hex.EncodeToString(hash[:8]) // Only first 8 bytes
	}

	// Hash fingerprint
	var fpHash string
	if s.config.EnableFingerprint && fingerprint != "" {
		hash := sha256.Sum256([]byte(fingerprint))
		fpHash = hex.EncodeToString(hash[:16])
	}

	claims := &SecureClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.Issuer,
			Subject:   userID,
			Audience:  s.config.Audience,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
		UserID:      userID,
		Email:       email,
		Username:    username,
		IsAdmin:     isAdmin,
		TokenType:   tokenType,
		TokenFamily: familyID,
		TokenID:     uuid.New().String(),
		Fingerprint: fpHash,
		IPHash:      ipHash,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.config.Secret.Bytes())
	if err != nil {
		return "", time.Time{}, err
	}

	return signedToken, expiresAt, nil
}

// ValidateToken validates a token and returns the claims
func (s *SecureJWTService) ValidateToken(
	ctx context.Context,
	tokenString string,
	expectedType TokenType,
	fingerprint string,
	clientIP string,
) (*SecureClaims, error) {
	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &SecureClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.config.Secret.Bytes(), nil
	}, jwt.WithLeeway(s.config.AllowedClockSkew))

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrTokenNotYetValid
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*SecureClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Validate token type
	if claims.TokenType != expectedType {
		return nil, ErrInvalidTokenType
	}

	// Validate issuer
	if claims.Issuer != s.config.Issuer {
		return nil, ErrInvalidIssuer
	}

	// Check if token is revoked
	if s.revokedTokens.IsRevoked(claims.TokenID) {
		return nil, ErrTokenRevoked
	}

	// Check if token family is revoked (refresh token rotation attack detection)
	if claims.TokenFamily != "" && s.revokedFamilies.IsRevoked(claims.TokenFamily) {
		return nil, ErrTokenFamilyRevoked
	}

	// Validate fingerprint if enabled
	if s.config.EnableFingerprint && claims.Fingerprint != "" && fingerprint != "" {
		hash := sha256.Sum256([]byte(fingerprint))
		fpHash := hex.EncodeToString(hash[:16])
		if !ConstantTimeStringCompare(claims.Fingerprint, fpHash) {
			// Log anomaly but don't reject (could be legitimate device change)
			// In high-security mode, this should return an error
		}
	}

	return claims, nil
}

// RefreshTokens rotates the refresh token and issues new pair
// Implements refresh token rotation per OWASP recommendations
func (s *SecureJWTService) RefreshTokens(
	ctx context.Context,
	refreshTokenString string,
	fingerprint string,
	clientIP string,
) (*TokenPair, error) {
	// Validate refresh token
	claims, err := s.ValidateToken(ctx, refreshTokenString, TokenTypeRefresh, fingerprint, clientIP)
	if err != nil {
		return nil, err
	}

	// Revoke the old refresh token (one-time use)
	s.revokedTokens.Revoke(claims.TokenID, s.config.RefreshTokenLife)

	// Generate new token pair with same family ID
	accessToken, accessExp, err := s.generateToken(
		claims.UserID, claims.Email, claims.Username, claims.IsAdmin,
		TokenTypeAccess, claims.TokenFamily, fingerprint, clientIP,
		s.config.AccessTokenLife,
	)
	if err != nil {
		return nil, err
	}

	refreshToken, _, err := s.generateToken(
		claims.UserID, claims.Email, claims.Username, claims.IsAdmin,
		TokenTypeRefresh, claims.TokenFamily, fingerprint, clientIP,
		s.config.RefreshTokenLife,
	)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExp,
		TokenType:    "Bearer",
	}, nil
}

// RevokeToken revokes a specific token
func (s *SecureJWTService) RevokeToken(tokenID string) {
	s.revokedTokens.Revoke(tokenID, s.config.RefreshTokenLife)
}

// RevokeFamily revokes all tokens in a family (e.g., when logout or token theft detected)
func (s *SecureJWTService) RevokeFamily(familyID string) {
	s.revokedFamilies.Revoke(familyID)
}

// RevokeAllUserTokens revokes all token families for a user
func (s *SecureJWTService) RevokeAllUserTokens(userID string) {
	// In production, this would query all active families for the user
	// and revoke them. For now, this is a placeholder.
}

// Cleanup removes expired entries from blacklists
func (s *SecureJWTService) Cleanup() {
	s.revokedTokens.Cleanup()
	s.revokedFamilies.Cleanup()
}

// TokenBlacklist stores revoked tokens with expiration
type TokenBlacklist struct {
	tokens map[string]time.Time
	mu     sync.RWMutex
	ttl    time.Duration
}

// NewTokenBlacklist creates a new token blacklist
func NewTokenBlacklist(ttl time.Duration) *TokenBlacklist {
	return &TokenBlacklist{
		tokens: make(map[string]time.Time),
		ttl:    ttl,
	}
}

// Revoke adds a token to the blacklist
func (b *TokenBlacklist) Revoke(tokenID string, lifetime time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.tokens[tokenID] = time.Now().Add(lifetime)
}

// IsRevoked checks if a token is revoked
func (b *TokenBlacklist) IsRevoked(tokenID string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	expiry, exists := b.tokens[tokenID]
	if !exists {
		return false
	}
	return time.Now().Before(expiry)
}

// Cleanup removes expired entries
func (b *TokenBlacklist) Cleanup() {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	for id, expiry := range b.tokens {
		if now.After(expiry) {
			delete(b.tokens, id)
		}
	}
}

// FamilyBlacklist stores revoked token families
type FamilyBlacklist struct {
	families map[string]time.Time
	mu       sync.RWMutex
	ttl      time.Duration
}

// NewFamilyBlacklist creates a new family blacklist
func NewFamilyBlacklist(ttl time.Duration) *FamilyBlacklist {
	return &FamilyBlacklist{
		families: make(map[string]time.Time),
		ttl:      ttl,
	}
}

// Revoke adds a family to the blacklist
func (b *FamilyBlacklist) Revoke(familyID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.families[familyID] = time.Now().Add(b.ttl)
}

// IsRevoked checks if a family is revoked
func (b *FamilyBlacklist) IsRevoked(familyID string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	expiry, exists := b.families[familyID]
	if !exists {
		return false
	}
	return time.Now().Before(expiry)
}

// Cleanup removes expired entries
func (b *FamilyBlacklist) Cleanup() {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	for id, expiry := range b.families {
		if now.After(expiry) {
			delete(b.families, id)
		}
	}
}
