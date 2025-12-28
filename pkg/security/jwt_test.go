package security

import (
	"context"
	"testing"
	"time"
)

// =============================================================================
// JWT Configuration Tests
// =============================================================================

func TestDefaultSecureJWTConfig(t *testing.T) {
	// Weak secret should fail
	_, err := DefaultSecureJWTConfig("short")
	if err == nil {
		t.Error("DefaultSecureJWTConfig() should fail with weak secret")
	}

	// Valid secret should succeed
	config, err := DefaultSecureJWTConfig("this-is-a-very-long-secret-key-that-is-at-least-32-chars")
	if err != nil {
		t.Fatalf("DefaultSecureJWTConfig() error = %v", err)
	}

	if config.AccessTokenLife != DefaultAccessTokenLifetime {
		t.Errorf("AccessTokenLife = %v, want %v", config.AccessTokenLife, DefaultAccessTokenLifetime)
	}

	if config.RefreshTokenLife != DefaultRefreshTokenLifetime {
		t.Errorf("RefreshTokenLife = %v, want %v", config.RefreshTokenLife, DefaultRefreshTokenLifetime)
	}
}

// =============================================================================
// JWT Service Tests
// =============================================================================

func TestNewSecureJWTService(t *testing.T) {
	// Nil config should fail
	_, err := NewSecureJWTService(nil)
	if err == nil {
		t.Error("NewSecureJWTService() should fail with nil config")
	}

	// Valid config should succeed
	config, _ := DefaultSecureJWTConfig("this-is-a-very-long-secret-key-that-is-at-least-32-chars")
	service, err := NewSecureJWTService(config)
	if err != nil {
		t.Fatalf("NewSecureJWTService() error = %v", err)
	}

	if service == nil {
		t.Error("NewSecureJWTService() returned nil service")
	}
}

func TestGenerateTokenPair(t *testing.T) {
	config, _ := DefaultSecureJWTConfig("this-is-a-very-long-secret-key-that-is-at-least-32-chars")
	service, _ := NewSecureJWTService(config)

	ctx := context.Background()
	pair, err := service.GenerateTokenPair(
		ctx,
		"user-123",
		"test@example.com",
		"testuser",
		false,
		"fingerprint-abc",
		"192.168.1.1",
	)
	if err != nil {
		t.Fatalf("GenerateTokenPair() error = %v", err)
	}

	if pair.AccessToken == "" {
		t.Error("GenerateTokenPair() returned empty access token")
	}

	if pair.RefreshToken == "" {
		t.Error("GenerateTokenPair() returned empty refresh token")
	}

	if pair.TokenType != "Bearer" {
		t.Errorf("GenerateTokenPair() TokenType = %v, want Bearer", pair.TokenType)
	}

	if pair.ExpiresAt.IsZero() {
		t.Error("GenerateTokenPair() returned zero ExpiresAt")
	}

	// Expiry should be in the future
	if pair.ExpiresAt.Before(time.Now()) {
		t.Error("GenerateTokenPair() ExpiresAt is in the past")
	}
}

func TestValidateToken(t *testing.T) {
	config, _ := DefaultSecureJWTConfig("this-is-a-very-long-secret-key-that-is-at-least-32-chars")
	service, _ := NewSecureJWTService(config)
	ctx := context.Background()

	pair, _ := service.GenerateTokenPair(ctx, "user-123", "test@example.com", "testuser", false, "fp", "1.2.3.4")

	// Validate access token
	claims, err := service.ValidateToken(ctx, pair.AccessToken, TokenTypeAccess, "fp", "1.2.3.4")
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	if claims.UserID != "user-123" {
		t.Errorf("Claims.UserID = %v, want user-123", claims.UserID)
	}

	if claims.Email != "test@example.com" {
		t.Errorf("Claims.Email = %v, want test@example.com", claims.Email)
	}

	if claims.TokenType != TokenTypeAccess {
		t.Errorf("Claims.TokenType = %v, want %v", claims.TokenType, TokenTypeAccess)
	}
}

func TestValidateTokenWrongType(t *testing.T) {
	config, _ := DefaultSecureJWTConfig("this-is-a-very-long-secret-key-that-is-at-least-32-chars")
	service, _ := NewSecureJWTService(config)
	ctx := context.Background()

	pair, _ := service.GenerateTokenPair(ctx, "user-123", "test@example.com", "testuser", false, "fp", "1.2.3.4")

	// Try to validate access token as refresh token (should fail)
	_, err := service.ValidateToken(ctx, pair.AccessToken, TokenTypeRefresh, "fp", "1.2.3.4")
	if err != ErrInvalidTokenType {
		t.Errorf("ValidateToken() with wrong type error = %v, want %v", err, ErrInvalidTokenType)
	}
}

func TestValidateTokenInvalid(t *testing.T) {
	config, _ := DefaultSecureJWTConfig("this-is-a-very-long-secret-key-that-is-at-least-32-chars")
	service, _ := NewSecureJWTService(config)
	ctx := context.Background()

	_, err := service.ValidateToken(ctx, "invalid-token", TokenTypeAccess, "", "")
	if err == nil {
		t.Error("ValidateToken() should fail with invalid token")
	}
}

func TestValidateTokenDifferentSecret(t *testing.T) {
	config1, _ := DefaultSecureJWTConfig("this-is-the-first-secret-key-for-testing-purposes")
	service1, _ := NewSecureJWTService(config1)

	config2, _ := DefaultSecureJWTConfig("this-is-a-different-secret-key-for-testing-purposes")
	service2, _ := NewSecureJWTService(config2)

	ctx := context.Background()

	// Generate token with service1
	pair, _ := service1.GenerateTokenPair(ctx, "user-123", "test@example.com", "testuser", false, "", "")

	// Try to validate with service2 (should fail)
	_, err := service2.ValidateToken(ctx, pair.AccessToken, TokenTypeAccess, "", "")
	if err == nil {
		t.Error("ValidateToken() should fail with different secret")
	}
}

func TestRefreshTokens(t *testing.T) {
	config, _ := DefaultSecureJWTConfig("this-is-a-very-long-secret-key-that-is-at-least-32-chars")
	service, _ := NewSecureJWTService(config)
	ctx := context.Background()

	// Generate initial pair
	initialPair, _ := service.GenerateTokenPair(ctx, "user-123", "test@example.com", "testuser", false, "fp", "1.2.3.4")

	// Refresh tokens
	newPair, err := service.RefreshTokens(ctx, initialPair.RefreshToken, "fp", "1.2.3.4")
	if err != nil {
		t.Fatalf("RefreshTokens() error = %v", err)
	}

	// New tokens should be different
	if newPair.AccessToken == initialPair.AccessToken {
		t.Error("RefreshTokens() returned same access token")
	}

	if newPair.RefreshToken == initialPair.RefreshToken {
		t.Error("RefreshTokens() returned same refresh token")
	}

	// New access token should be valid
	claims, err := service.ValidateToken(ctx, newPair.AccessToken, TokenTypeAccess, "fp", "1.2.3.4")
	if err != nil {
		t.Fatalf("ValidateToken() on refreshed token error = %v", err)
	}

	if claims.UserID != "user-123" {
		t.Errorf("Claims.UserID = %v, want user-123", claims.UserID)
	}
}

func TestRefreshTokensOneTimeUse(t *testing.T) {
	config, _ := DefaultSecureJWTConfig("this-is-a-very-long-secret-key-that-is-at-least-32-chars")
	service, _ := NewSecureJWTService(config)
	ctx := context.Background()

	pair, _ := service.GenerateTokenPair(ctx, "user-123", "test@example.com", "testuser", false, "fp", "1.2.3.4")

	// First refresh should succeed
	_, err := service.RefreshTokens(ctx, pair.RefreshToken, "fp", "1.2.3.4")
	if err != nil {
		t.Fatalf("First RefreshTokens() error = %v", err)
	}

	// Second refresh with same token should fail (token revoked)
	_, err = service.RefreshTokens(ctx, pair.RefreshToken, "fp", "1.2.3.4")
	if err != ErrTokenRevoked {
		t.Errorf("Second RefreshTokens() error = %v, want %v", err, ErrTokenRevoked)
	}
}

func TestRevokeToken(t *testing.T) {
	config, _ := DefaultSecureJWTConfig("this-is-a-very-long-secret-key-that-is-at-least-32-chars")
	service, _ := NewSecureJWTService(config)
	ctx := context.Background()

	pair, _ := service.GenerateTokenPair(ctx, "user-123", "test@example.com", "testuser", false, "", "")

	// Validate token first (to get claims)
	claims, _ := service.ValidateToken(ctx, pair.AccessToken, TokenTypeAccess, "", "")

	// Revoke the token
	service.RevokeToken(claims.TokenID)

	// Token should now be invalid
	_, err := service.ValidateToken(ctx, pair.AccessToken, TokenTypeAccess, "", "")
	if err != ErrTokenRevoked {
		t.Errorf("ValidateToken() after revoke error = %v, want %v", err, ErrTokenRevoked)
	}
}

func TestRevokeFamily(t *testing.T) {
	config, _ := DefaultSecureJWTConfig("this-is-a-very-long-secret-key-that-is-at-least-32-chars")
	service, _ := NewSecureJWTService(config)
	ctx := context.Background()

	pair, _ := service.GenerateTokenPair(ctx, "user-123", "test@example.com", "testuser", false, "", "")

	// Get the family ID
	claims, _ := service.ValidateToken(ctx, pair.RefreshToken, TokenTypeRefresh, "", "")
	familyID := claims.TokenFamily

	// Revoke the family
	service.RevokeFamily(familyID)

	// Refresh token should now be invalid
	_, err := service.ValidateToken(ctx, pair.RefreshToken, TokenTypeRefresh, "", "")
	if err != ErrTokenFamilyRevoked {
		t.Errorf("ValidateToken() after family revoke error = %v, want %v", err, ErrTokenFamilyRevoked)
	}
}

// =============================================================================
// Token Blacklist Tests
// =============================================================================

func TestTokenBlacklist(t *testing.T) {
	blacklist := NewTokenBlacklist(1 * time.Hour)

	tokenID := "test-token-123"

	// Should not be revoked initially
	if blacklist.IsRevoked(tokenID) {
		t.Error("IsRevoked() returned true for non-revoked token")
	}

	// Revoke the token
	blacklist.Revoke(tokenID, 1*time.Hour)

	// Should now be revoked
	if !blacklist.IsRevoked(tokenID) {
		t.Error("IsRevoked() returned false for revoked token")
	}
}

func TestTokenBlacklistCleanup(t *testing.T) {
	blacklist := NewTokenBlacklist(1 * time.Millisecond)

	// Add a token with very short expiry
	blacklist.Revoke("short-lived", 1*time.Millisecond)

	// Wait for expiry
	time.Sleep(10 * time.Millisecond)

	// Cleanup
	blacklist.Cleanup()

	// Should no longer be revoked
	if blacklist.IsRevoked("short-lived") {
		t.Error("Token still revoked after cleanup and expiry")
	}
}

// =============================================================================
// Family Blacklist Tests
// =============================================================================

func TestFamilyBlacklist(t *testing.T) {
	blacklist := NewFamilyBlacklist(1 * time.Hour)

	familyID := "family-123"

	if blacklist.IsRevoked(familyID) {
		t.Error("IsRevoked() returned true for non-revoked family")
	}

	blacklist.Revoke(familyID)

	if !blacklist.IsRevoked(familyID) {
		t.Error("IsRevoked() returned false for revoked family")
	}
}

// =============================================================================
// Admin Token Tests
// =============================================================================

func TestAdminTokenClaims(t *testing.T) {
	config, _ := DefaultSecureJWTConfig("this-is-a-very-long-secret-key-that-is-at-least-32-chars")
	service, _ := NewSecureJWTService(config)
	ctx := context.Background()

	// Generate admin token
	pair, _ := service.GenerateTokenPair(ctx, "admin-123", "admin@example.com", "admin", true, "", "")

	claims, err := service.ValidateToken(ctx, pair.AccessToken, TokenTypeAccess, "", "")
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	if !claims.IsAdmin {
		t.Error("IsAdmin should be true for admin token")
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkGenerateTokenPair(b *testing.B) {
	config, _ := DefaultSecureJWTConfig("this-is-a-very-long-secret-key-that-is-at-least-32-chars")
	service, _ := NewSecureJWTService(config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GenerateTokenPair(ctx, "user-123", "test@example.com", "testuser", false, "", "")
	}
}

func BenchmarkValidateToken(b *testing.B) {
	config, _ := DefaultSecureJWTConfig("this-is-a-very-long-secret-key-that-is-at-least-32-chars")
	service, _ := NewSecureJWTService(config)
	ctx := context.Background()

	pair, _ := service.GenerateTokenPair(ctx, "user-123", "test@example.com", "testuser", false, "", "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.ValidateToken(ctx, pair.AccessToken, TokenTypeAccess, "", "")
	}
}
