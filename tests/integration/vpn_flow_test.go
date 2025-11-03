// +build integration

package integration

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikola43/aureo-vpn/pkg/auth"
	"github.com/nikola43/aureo-vpn/pkg/crypto"
	"github.com/nikola43/aureo-vpn/pkg/database"
	"github.com/nikola43/aureo-vpn/pkg/models"
)

func TestCompleteVPNFlow(t *testing.T) {
	// Setup test database
	setupTestDB(t)
	defer teardownTestDB(t)

	t.Run("User Registration and Authentication", func(t *testing.T) {
		tokenService := auth.NewTokenService("test-secret", 15*time.Minute, 7*24*time.Hour)
		authService := auth.NewService(tokenService)

		// Register user
		registerReq := auth.RegisterRequest{
			Email:    "testuser@example.com",
			Password: "securepassword123",
			Username: "testuser",
			FullName: "Test User",
		}

		authResp, err := authService.Register(registerReq)
		if err != nil {
			t.Fatalf("Failed to register user: %v", err)
		}
		if authResp.AccessToken == "" {
			t.Error("Access token should not be empty")
		}
		if authResp.RefreshToken == "" {
			t.Error("Refresh token should not be empty")
		}
		if authResp.User.Username != "testuser" {
			t.Errorf("Expected username 'testuser', got '%s'", authResp.User.Username)
		}

		// Verify login
		loginReq := auth.LoginRequest{
			Email:    "testuser@example.com",
			Password: "securepassword123",
		}

		loginResp, err := authService.Login(loginReq)
		require.NoError(t, err)
		assert.NotEmpty(t, loginResp.AccessToken)
	})

	t.Run("VPN Node Management", func(t *testing.T) {
		db := database.GetDB()

		// Create VPN node
		node := &models.VPNNode{
			Name:              "Test-Node-US",
			Hostname:          "test-us.example.com",
			PublicIP:          "192.0.2.1",
			Country:           "United States",
			CountryCode:       "US",
			City:              "New York",
			Status:            "online",
			IsActive:          true,
			SupportsWireGuard: true,
			SupportsOpenVPN:   true,
			MaxConnections:    100,
			WireGuardPort:     51820,
			OpenVPNPort:       1194,
		}

		err := db.Create(node).Error
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, node.ID)

		// Verify node can be retrieved
		var retrievedNode models.VPNNode
		err = db.First(&retrievedNode, node.ID).Error
		require.NoError(t, err)
		assert.Equal(t, "Test-Node-US", retrievedNode.Name)
		assert.True(t, retrievedNode.IsHealthy())
	})

	t.Run("VPN Session Creation", func(t *testing.T) {
		db := database.GetDB()

		// Create test user
		user := &models.User{
			Email:            "sessiontest@example.com",
			Username:         "sessiontest",
			PasswordHash:     "hashed",
			IsActive:         true,
			SubscriptionTier: "premium",
		}
		db.Create(user)

		// Create test node
		node := &models.VPNNode{
			Name:              "Session-Test-Node",
			Hostname:          "session-test.example.com",
			PublicIP:          "192.0.2.2",
			Country:           "Germany",
			CountryCode:       "DE",
			City:              "Berlin",
			Status:            "online",
			IsActive:          true,
			SupportsWireGuard: true,
			MaxConnections:    50,
		}
		db.Create(node)

		// Create session
		session := &models.Session{
			UserID:            user.ID,
			NodeID:            node.ID,
			Protocol:          "wireguard",
			TunnelIP:          "10.8.0.10",
			Status:            "active",
			ConnectedAt:       time.Now(),
			LastKeepalive:     time.Now(),
			KillSwitchEnabled: true,
			DNSLeakProtection: true,
		}

		err := db.Create(session).Error
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, session.ID)
		assert.True(t, session.IsActive())
	})

	t.Run("Multi-Hop VPN", func(t *testing.T) {
		db := database.GetDB()

		// Create entry node
		entryNode := &models.VPNNode{
			Name:              "Entry-Node",
			PublicIP:          "192.0.2.10",
			Country:           "Switzerland",
			CountryCode:       "CH",
			City:              "Zurich",
			Status:            "online",
			IsActive:          true,
			SupportsMultiHop:  true,
			SupportsWireGuard: true,
		}
		db.Create(entryNode)

		// Create exit node
		exitNode := &models.VPNNode{
			Name:              "Exit-Node",
			PublicIP:          "192.0.2.11",
			Country:           "Iceland",
			CountryCode:       "IS",
			City:              "Reykjavik",
			Status:            "online",
			IsActive:          true,
			SupportsMultiHop:  true,
			SupportsWireGuard: true,
		}
		db.Create(exitNode)

		// Verify nodes can be used for multi-hop
		assert.True(t, entryNode.SupportsMultiHop)
		assert.True(t, exitNode.SupportsMultiHop)
		assert.NotEqual(t, entryNode.Country, exitNode.Country)
	})

	t.Run("Load Balancing", func(t *testing.T) {
		db := database.GetDB()

		// Create multiple nodes
		nodes := []models.VPNNode{
			{Name: "Node-1", PublicIP: "192.0.2.20", CountryCode: "US", Status: "online", IsActive: true, LoadScore: 25.0},
			{Name: "Node-2", PublicIP: "192.0.2.21", CountryCode: "US", Status: "online", IsActive: true, LoadScore: 50.0},
			{Name: "Node-3", PublicIP: "192.0.2.22", CountryCode: "US", Status: "online", IsActive: true, LoadScore: 75.0},
		}

		for _, node := range nodes {
			db.Create(&node)
		}

		// Query best node (lowest load score)
		var bestNode models.VPNNode
		err := db.Where("country_code = ? AND status = ? AND is_active = ?", "US", "online", true).
			Order("load_score ASC").
			First(&bestNode).Error

		require.NoError(t, err)
		assert.Equal(t, "Node-1", bestNode.Name)
		assert.Equal(t, 25.0, bestNode.LoadScore)
	})
}

func TestSecurityFeatures(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	t.Run("Password Hashing", func(t *testing.T) {
		hasher := crypto.NewPasswordHasher()

		password := "MySecurePassword123!"
		hash, err := hasher.HashPassword(password)
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.NotEqual(t, password, hash)

		// Verify correct password
		valid, err := hasher.VerifyPassword(password, hash)
		require.NoError(t, err)
		assert.True(t, valid)

		// Verify incorrect password
		valid, err = hasher.VerifyPassword("WrongPassword", hash)
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("JWT Token Security", func(t *testing.T) {
		tokenService := auth.NewTokenService("secure-secret-key", 15*time.Minute, 7*24*time.Hour)

		userID := uuid.New()
		token, err := tokenService.GenerateAccessToken(userID, "test@example.com", "testuser", false)
		require.NoError(t, err)

		// Verify token
		claims, err := tokenService.VerifyToken(token)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, "access", claims.TokenType)

		// Try to verify with wrong secret
		wrongTokenService := auth.NewTokenService("wrong-secret", 15*time.Minute, 7*24*time.Hour)
		_, err = wrongTokenService.VerifyToken(token)
		assert.Error(t, err)
	})
}

func TestPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	t.Run("Concurrent User Registration", func(t *testing.T) {
		tokenService := auth.NewTokenService("test-secret", 15*time.Minute, 7*24*time.Hour)
		authService := auth.NewService(tokenService)

		concurrency := 50
		done := make(chan bool, concurrency)

		start := time.Now()

		for i := 0; i < concurrency; i++ {
			go func(index int) {
				req := auth.RegisterRequest{
					Email:    fmt.Sprintf("perf-test-%d@example.com", index),
					Password: "password123",
					Username: fmt.Sprintf("perftest%d", index),
				}

				_, err := authService.Register(req)
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		// Wait for all registrations
		for i := 0; i < concurrency; i++ {
			<-done
		}

		duration := time.Since(start)
		t.Logf("Registered %d users in %v (%.2f users/sec)", concurrency, duration, float64(concurrency)/duration.Seconds())

		// Should complete in reasonable time
		assert.Less(t, duration, 10*time.Second)
	})
}

// Helper functions

func setupTestDB(t *testing.T) {
	config := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvAsInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "aureo_vpn_test"),
		SSLMode:  "disable",
		TimeZone: "UTC",
	}

	err := database.Connect(config)
	require.NoError(t, err)

	err = database.AutoMigrate()
	require.NoError(t, err)
}

func teardownTestDB(t *testing.T) {
	db := database.GetDB()

	// Clean up test data
	db.Exec("DELETE FROM sessions")
	db.Exec("DELETE FROM configs")
	db.Exec("DELETE FROM vpn_nodes")
	db.Exec("DELETE FROM users")

	database.Close()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	var value int
	fmt.Sscanf(valueStr, "%d", &value)
	return value
}
