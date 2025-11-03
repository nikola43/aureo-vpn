package unit

import (
	"testing"
	"time"

	"github.com/nikola43/aureo-vpn/pkg/auth"
	"github.com/google/uuid"
)

func TestTokenGeneration(t *testing.T) {
	tokenService := auth.NewTokenService(
		"test-secret-key",
		15*time.Minute,
		7*24*time.Hour,
	)

	userID := uuid.New()
	email := "test@example.com"
	username := "testuser"
	isAdmin := false

	// Test access token generation
	accessToken, err := tokenService.GenerateAccessToken(userID, email, username, isAdmin)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	if accessToken == "" {
		t.Error("Access token should not be empty")
	}

	// Test refresh token generation
	refreshToken, err := tokenService.GenerateRefreshToken(userID, email, username, isAdmin)
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	if refreshToken == "" {
		t.Error("Refresh token should not be empty")
	}
}

func TestTokenVerification(t *testing.T) {
	tokenService := auth.NewTokenService(
		"test-secret-key",
		15*time.Minute,
		7*24*time.Hour,
	)

	userID := uuid.New()
	email := "test@example.com"
	username := "testuser"
	isAdmin := true

	// Generate token
	token, err := tokenService.GenerateAccessToken(userID, email, username, isAdmin)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Verify token
	claims, err := tokenService.VerifyToken(token)
	if err != nil {
		t.Fatalf("Failed to verify token: %v", err)
	}

	// Check claims
	if claims.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, claims.UserID)
	}

	if claims.Email != email {
		t.Errorf("Expected email %s, got %s", email, claims.Email)
	}

	if claims.Username != username {
		t.Errorf("Expected username %s, got %s", username, claims.Username)
	}

	if claims.IsAdmin != isAdmin {
		t.Errorf("Expected IsAdmin %v, got %v", isAdmin, claims.IsAdmin)
	}

	if claims.TokenType != "access" {
		t.Errorf("Expected token type 'access', got %s", claims.TokenType)
	}
}

func TestInvalidToken(t *testing.T) {
	tokenService := auth.NewTokenService(
		"test-secret-key",
		15*time.Minute,
		7*24*time.Hour,
	)

	// Test with invalid token
	_, err := tokenService.VerifyToken("invalid.token.here")
	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}
}

func TestTokenExpiration(t *testing.T) {
	// Create service with very short expiration
	tokenService := auth.NewTokenService(
		"test-secret-key",
		1*time.Millisecond,
		1*time.Millisecond,
	)

	userID := uuid.New()
	token, err := tokenService.GenerateAccessToken(userID, "test@example.com", "testuser", false)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Try to verify expired token
	_, err = tokenService.VerifyToken(token)
	if err != auth.ErrExpiredToken {
		t.Errorf("Expected ErrExpiredToken, got %v", err)
	}
}

func TestTokenPairGeneration(t *testing.T) {
	tokenService := auth.NewTokenService(
		"test-secret-key",
		15*time.Minute,
		7*24*time.Hour,
	)

	userID := uuid.New()
	accessToken, refreshToken, err := tokenService.GenerateTokenPair(
		userID,
		"test@example.com",
		"testuser",
		false,
	)

	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	if accessToken == "" || refreshToken == "" {
		t.Error("Tokens should not be empty")
	}

	// Verify both tokens
	accessClaims, err := tokenService.VerifyToken(accessToken)
	if err != nil {
		t.Fatalf("Failed to verify access token: %v", err)
	}

	if accessClaims.TokenType != "access" {
		t.Error("Expected access token type")
	}

	refreshClaims, err := tokenService.VerifyToken(refreshToken)
	if err != nil {
		t.Fatalf("Failed to verify refresh token: %v", err)
	}

	if refreshClaims.TokenType != "refresh" {
		t.Error("Expected refresh token type")
	}
}
