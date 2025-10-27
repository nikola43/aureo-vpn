package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// Claims represents JWT claims
type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	IsAdmin   bool      `json:"is_admin"`
	TokenType string    `json:"token_type"` // access, refresh
	jwt.RegisteredClaims
}

// TokenService handles JWT token operations
type TokenService struct {
	secretKey            []byte
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

// NewTokenService creates a new token service
func NewTokenService(secretKey string, accessDuration, refreshDuration time.Duration) *TokenService {
	return &TokenService{
		secretKey:            []byte(secretKey),
		accessTokenDuration:  accessDuration,
		refreshTokenDuration: refreshDuration,
	}
}

// GenerateAccessToken generates an access token for a user
func (t *TokenService) GenerateAccessToken(userID uuid.UUID, email, username string, isAdmin bool) (string, error) {
	claims := Claims{
		UserID:    userID,
		Email:     email,
		Username:  username,
		IsAdmin:   isAdmin,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "aureo-vpn",
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.secretKey)
}

// GenerateRefreshToken generates a refresh token for a user
func (t *TokenService) GenerateRefreshToken(userID uuid.UUID, email, username string, isAdmin bool) (string, error) {
	claims := Claims{
		UserID:    userID,
		Email:     email,
		Username:  username,
		IsAdmin:   isAdmin,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.refreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "aureo-vpn",
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.secretKey)
}

// VerifyToken verifies and parses a JWT token
func (t *TokenService) VerifyToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return t.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token from a valid refresh token
func (t *TokenService) RefreshAccessToken(refreshToken string) (string, error) {
	claims, err := t.VerifyToken(refreshToken)
	if err != nil {
		return "", err
	}

	if claims.TokenType != "refresh" {
		return "", ErrInvalidToken
	}

	return t.GenerateAccessToken(claims.UserID, claims.Email, claims.Username, claims.IsAdmin)
}

// GenerateTokenPair generates both access and refresh tokens
func (t *TokenService) GenerateTokenPair(userID uuid.UUID, email, username string, isAdmin bool) (accessToken, refreshToken string, err error) {
	accessToken, err = t.GenerateAccessToken(userID, email, username, isAdmin)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = t.GenerateRefreshToken(userID, email, username, isAdmin)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
