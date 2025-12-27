//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/nikola43/aureo-vpn/pkg/auth"
	"github.com/nikola43/aureo-vpn/pkg/blockchain"
	"github.com/nikola43/aureo-vpn/pkg/database"
	"github.com/nikola43/aureo-vpn/pkg/logger"
	"github.com/nikola43/aureo-vpn/pkg/models"
	"github.com/nikola43/aureo-vpn/pkg/operator"
	"github.com/nikola43/aureo-vpn/pkg/rewards"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEarningsFlow(t *testing.T) {
	setupEarningsTestDB(t)
	defer teardownEarningsTestDB(t)

	t.Run("Full Earnings Flow", func(t *testing.T) {
		db := database.GetDB()
		log := logger.NewDefault()

		// Services
		blockchainService, _ := blockchain.NewService(blockchain.Config{}, log)
		rewardService := rewards.NewRewardService(log, blockchainService)
		operatorService := operator.NewService(log, rewardService)
		tokenService := auth.NewTokenService("test-secret", 1*time.Hour, 1*time.Hour)
		authService := auth.NewService(tokenService)

		// 1. Register Operator User
		opUserReq := auth.RegisterRequest{
			Email:    "operator@example.com",
			Username: "operator",
			Password: "password123",
			FullName: "Operator User",
		}
		opUserResp, err := authService.Register(opUserReq)
		require.NoError(t, err)

		// 2. Register Operator
		opReq := operator.RegisterRequest{
			WalletAddress: "0x1234567890abcdef1234567890abcdef12345678",
			WalletType:    "ethereum",
			Email:         "operator@example.com",
			Country:       "US",
		}
		op, err := operatorService.RegisterOperator(context.Background(), opUserResp.User.ID, opReq)
		require.NoError(t, err)

		// Verify Operator
		err = operatorService.VerifyOperator(context.Background(), op.ID)
		require.NoError(t, err)

		// Initialize Reward Tiers
		err = rewardService.InitializeRewardTiers()
		require.NoError(t, err)

		// 3. Create Node
		nodeReq := operator.NodeCreateRequest{
			Name:          "OpNode1",
			Hostname:      "opnode1.example.com",
			Country:       "US",
			CountryCode:   "US",
			City:          "New York",
			PublicIP:      "1.2.3.4",
			WireGuardPort: 51820,
			OpenVPNPort:   1194,
		}
		node, _, err := operatorService.CreateNode(context.Background(), op.ID, nodeReq)
		require.NoError(t, err)

		// Mark node online
		db.Model(node).Update("status", "online")

		// 4. Register Client User
		clientUserReq := auth.RegisterRequest{
			Email:    "client@example.com",
			Username: "client",
			Password: "password123",
			FullName: "Client User",
		}
		clientUserResp, err := authService.Register(clientUserReq)
		require.NoError(t, err)

		// 5. Create Session
		session := &models.Session{
			UserID:      clientUserResp.User.ID,
			NodeID:      node.ID,
			Protocol:    "wireguard",
			Status:      "active",
			ConnectedAt: time.Now(),
		}
		err = db.Create(session).Error
		require.NoError(t, err)

		// 6. Record Earnings (Simulate traffic)
		// 100 MB = 100 * 1024 KB = 102400 KB
		trafficKB := int64(102400)
		err = rewardService.RecordEarning(context.Background(), session.ID, trafficKB, 60)
		require.NoError(t, err)

		// Manually update node bandwidth (as simulated by vpn-node service)
		db.Model(node).Update("total_bandwidth_kb", trafficKB)

		// Trigger UpdateStats manually to sync operator stats
		var operator models.NodeOperator
		db.First(&operator, op.ID)
		operator.UpdateStats(db)

		// 7. Verify Earnings
		stats, err := operatorService.GetOperatorStats(context.Background(), op.ID)
		require.NoError(t, err)

		// Verify TotalBandwidthKB
		var kb int64
		if val, ok := stats["total_bandwidth_kb"].(int64); ok {
			kb = val
		} else if val, ok := stats["total_bandwidth_kb"].(float64); ok {
			kb = int64(val)
		}
		assert.Equal(t, trafficKB, kb)

		// Verify Earnings Record
		var earnings []models.OperatorEarning
		db.Where("operator_id = ?", op.ID).Find(&earnings)
		assert.NotEmpty(t, earnings)
		assert.Equal(t, trafficKB, earnings[0].BandwidthKB)

		// Verify Pending Payout
		pendingPayout, ok := stats["pending_payout"].(float64)
		if !ok {
			// Try int or other types
			if val, ok := stats["pending_payout"].(int); ok {
				pendingPayout = float64(val)
			}
		}
		// Rate $0.10/GB. 100MB ~= $0.01
		assert.Greater(t, pendingPayout, 0.009)
		assert.Less(t, pendingPayout, 0.011)
	})
}

// Helper functions

func setupEarningsTestDB(t *testing.T) {
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

func teardownEarningsTestDB(t *testing.T) {
	db := database.GetDB()

	// Clean up test data
	db.Exec("DELETE FROM sessions")
	db.Exec("DELETE FROM configs")
	db.Exec("DELETE FROM vpn_nodes")
	db.Exec("DELETE FROM users")
	db.Exec("DELETE FROM node_operators")
	db.Exec("DELETE FROM operator_earnings")
	db.Exec("DELETE FROM operator_payouts")
	db.Exec("DELETE FROM node_rewards")

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
