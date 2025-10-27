package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/nikola43/aureo-vpn/pkg/database"
	"github.com/nikola43/aureo-vpn/pkg/models"
	"github.com/nikola43/aureo-vpn/pkg/protocols/wireguard"
	"github.com/spf13/cobra"
)

var (
	dbHost     string
	dbPort     int
	dbUser     string
	dbPassword string
	dbName     string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "aureo-vpn",
		Short: "Aureo VPN Management CLI",
		Long:  "Command-line tool for managing Aureo VPN servers and configurations",
	}

	// Database flags
	rootCmd.PersistentFlags().StringVar(&dbHost, "db-host", "localhost", "Database host")
	rootCmd.PersistentFlags().IntVar(&dbPort, "db-port", 5432, "Database port")
	rootCmd.PersistentFlags().StringVar(&dbUser, "db-user", "postgres", "Database user")
	rootCmd.PersistentFlags().StringVar(&dbPassword, "db-password", "postgres", "Database password")
	rootCmd.PersistentFlags().StringVar(&dbName, "db-name", "aureo_vpn", "Database name")

	// Node commands
	nodeCmd := &cobra.Command{
		Use:   "node",
		Short: "Manage VPN nodes",
	}

	nodeCmd.AddCommand(
		createNodeCmd(),
		listNodesCmd(),
		deleteNodeCmd(),
	)

	// Config commands
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage VPN configurations",
	}

	configCmd.AddCommand(
		generateConfigCmd(),
	)

	// User commands
	userCmd := &cobra.Command{
		Use:   "user",
		Short: "Manage users",
	}

	userCmd.AddCommand(
		listUsersCmd(),
		createUserCmd(),
	)

	// Stats command
	statsCmd := &cobra.Command{
		Use:   "stats",
		Short: "View system statistics",
		Run:   runStats,
	}

	rootCmd.AddCommand(nodeCmd, configCmd, userCmd, statsCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func connectDB() error {
	config := database.Config{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		DBName:   dbName,
		SSLMode:  "disable",
		TimeZone: "UTC",
	}

	return database.Connect(config)
}

func createNodeCmd() *cobra.Command {
	var (
		name        string
		hostname    string
		publicIP    string
		country     string
		countryCode string
		city        string
		wgPort      int
		ovpnPort    int
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new VPN node",
		Run: func(cmd *cobra.Command, args []string) {
			if err := connectDB(); err != nil {
				log.Fatalf("Failed to connect to database: %v", err)
			}
			defer database.Close()

			// Generate keypair
			keyPair, err := wireguard.GenerateKeyPair()
			if err != nil {
				log.Fatalf("Failed to generate keypair: %v", err)
			}

			node := &models.VPNNode{
				Name:              name,
				Hostname:          hostname,
				PublicIP:          publicIP,
				Country:           country,
				CountryCode:       countryCode,
				City:              city,
				WireGuardPort:     wgPort,
				OpenVPNPort:       ovpnPort,
				PublicKey:         keyPair.PublicKey,
				Status:            "offline",
				IsActive:          true,
				SupportsWireGuard: true,
				SupportsOpenVPN:   true,
				MaxConnections:    1000,
			}

			db := database.GetDB()
			if err := db.Create(node).Error; err != nil {
				log.Fatalf("Failed to create node: %v", err)
			}

			fmt.Printf("Node created successfully!\n")
			fmt.Printf("ID: %s\n", node.ID)
			fmt.Printf("Name: %s\n", node.Name)
			fmt.Printf("Public Key: %s\n", node.PublicKey)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Node name (required)")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Node hostname (required)")
	cmd.Flags().StringVar(&publicIP, "ip", "", "Public IP address (required)")
	cmd.Flags().StringVar(&country, "country", "", "Country name (required)")
	cmd.Flags().StringVar(&countryCode, "country-code", "", "Country code (required)")
	cmd.Flags().StringVar(&city, "city", "", "City name (required)")
	cmd.Flags().IntVar(&wgPort, "wg-port", 51820, "WireGuard port")
	cmd.Flags().IntVar(&ovpnPort, "ovpn-port", 1194, "OpenVPN port")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("hostname")
	cmd.MarkFlagRequired("ip")
	cmd.MarkFlagRequired("country")
	cmd.MarkFlagRequired("country-code")
	cmd.MarkFlagRequired("city")

	return cmd
}

func listNodesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all VPN nodes",
		Run: func(cmd *cobra.Command, args []string) {
			if err := connectDB(); err != nil {
				log.Fatalf("Failed to connect to database: %v", err)
			}
			defer database.Close()

			db := database.GetDB()
			var nodes []models.VPNNode
			if err := db.Find(&nodes).Error; err != nil {
				log.Fatalf("Failed to list nodes: %v", err)
			}

			fmt.Printf("Found %d nodes:\n\n", len(nodes))
			for _, node := range nodes {
				fmt.Printf("ID: %s\n", node.ID)
				fmt.Printf("Name: %s\n", node.Name)
				fmt.Printf("Location: %s, %s\n", node.City, node.Country)
				fmt.Printf("IP: %s\n", node.PublicIP)
				fmt.Printf("Status: %s\n", node.Status)
				fmt.Printf("Connections: %d/%d\n", node.CurrentConnections, node.MaxConnections)
				fmt.Printf("Load Score: %.2f\n", node.LoadScore)
				fmt.Println("---")
			}
		},
	}
}

func deleteNodeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [node-id]",
		Short: "Delete a VPN node",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := connectDB(); err != nil {
				log.Fatalf("Failed to connect to database: %v", err)
			}
			defer database.Close()

			nodeID, err := uuid.Parse(args[0])
			if err != nil {
				log.Fatalf("Invalid node ID: %v", err)
			}

			db := database.GetDB()
			if err := db.Delete(&models.VPNNode{}, nodeID).Error; err != nil {
				log.Fatalf("Failed to delete node: %v", err)
			}

			fmt.Println("Node deleted successfully")
		},
	}
}

func generateConfigCmd() *cobra.Command {
	var (
		userID   string
		nodeID   string
		protocol string
		output   string
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate VPN configuration for a user",
		Run: func(cmd *cobra.Command, args []string) {
			if err := connectDB(); err != nil {
				log.Fatalf("Failed to connect to database: %v", err)
			}
			defer database.Close()

			userUUID, err := uuid.Parse(userID)
			if err != nil {
				log.Fatalf("Invalid user ID: %v", err)
			}

			nodeUUID, err := uuid.Parse(nodeID)
			if err != nil {
				log.Fatalf("Invalid node ID: %v", err)
			}

			db := database.GetDB()

			// Get node
			var node models.VPNNode
			if err := db.First(&node, nodeUUID).Error; err != nil {
				log.Fatalf("Failed to find node: %v", err)
			}

			// Generate client keypair
			keyPair, err := wireguard.GenerateKeyPair()
			if err != nil {
				log.Fatalf("Failed to generate keypair: %v", err)
			}

			// Generate config
			config := wireguard.Config{
				PrivateKey:          keyPair.PrivateKey,
				Address:             []string{"10.8.0.2/24"},
				DNS:                 []string{"1.1.1.1", "1.0.0.1"},
				MTU:                 1420,
				PeerPublicKey:       node.PublicKey,
				PeerEndpoint:        fmt.Sprintf("%s:%d", node.PublicIP, node.WireGuardPort),
				AllowedIPs:          []string{"0.0.0.0/0", "::/0"},
				PersistentKeepalive: 25,
			}

			configContent, err := wireguard.GenerateClientConfig(config)
			if err != nil {
				log.Fatalf("Failed to generate config: %v", err)
			}

			// Save config to database
			dbConfig := &models.Config{
				UserID:        userUUID,
				NodeID:        nodeUUID,
				Protocol:      protocol,
				ConfigName:    fmt.Sprintf("%s-%s", node.Name, time.Now().Format("20060102")),
				ConfigContent: configContent,
				PublicKey:     keyPair.PublicKey,
				PrivateKey:    keyPair.PrivateKey,
				DNSServers:    "1.1.1.1,1.0.0.1",
				AllowedIPs:    "0.0.0.0/0,::/0",
				IsActive:      true,
			}

			if err := db.Create(dbConfig).Error; err != nil {
				log.Fatalf("Failed to save config: %v", err)
			}

			// Write to file if output specified
			if output != "" {
				if err := os.WriteFile(output, []byte(configContent), 0600); err != nil {
					log.Fatalf("Failed to write config file: %v", err)
				}
				fmt.Printf("Configuration saved to %s\n", output)
			} else {
				fmt.Println(configContent)
			}

			fmt.Printf("\nConfig ID: %s\n", dbConfig.ID)
		},
	}

	cmd.Flags().StringVar(&userID, "user", "", "User ID (required)")
	cmd.Flags().StringVar(&nodeID, "node", "", "Node ID (required)")
	cmd.Flags().StringVar(&protocol, "protocol", "wireguard", "Protocol (wireguard or openvpn)")
	cmd.Flags().StringVar(&output, "output", "", "Output file path")

	cmd.MarkFlagRequired("user")
	cmd.MarkFlagRequired("node")

	return cmd
}

func listUsersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all users",
		Run: func(cmd *cobra.Command, args []string) {
			if err := connectDB(); err != nil {
				log.Fatalf("Failed to connect to database: %v", err)
			}
			defer database.Close()

			db := database.GetDB()
			var users []models.User
			if err := db.Find(&users).Error; err != nil {
				log.Fatalf("Failed to list users: %v", err)
			}

			fmt.Printf("Found %d users:\n\n", len(users))
			for _, user := range users {
				fmt.Printf("ID: %s\n", user.ID)
				fmt.Printf("Username: %s\n", user.Username)
				fmt.Printf("Email: %s\n", user.Email)
				fmt.Printf("Subscription: %s\n", user.SubscriptionTier)
				fmt.Printf("Active: %v\n", user.IsActive)
				fmt.Printf("Data Used: %.2f GB\n", user.DataTransferredGB)
				fmt.Println("---")
			}
		},
	}
}

func createUserCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create [username] [email]",
		Short: "Create a new user",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Use the API to create users with proper authentication")
		},
	}
}

func runStats(cmd *cobra.Command, args []string) {
	if err := connectDB(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	db := database.GetDB()

	var totalUsers, activeUsers int64
	db.Model(&models.User{}).Count(&totalUsers)
	db.Model(&models.User{}).Where("is_active = ?", true).Count(&activeUsers)

	var totalNodes, onlineNodes int64
	db.Model(&models.VPNNode{}).Count(&totalNodes)
	db.Model(&models.VPNNode{}).Where("status = ?", "online").Count(&onlineNodes)

	var totalSessions, activeSessions int64
	db.Model(&models.Session{}).Count(&totalSessions)
	db.Model(&models.Session{}).Where("status = ?", "active").Count(&activeSessions)

	fmt.Println("=== Aureo VPN Statistics ===\n")
	fmt.Printf("Users:\n")
	fmt.Printf("  Total: %d\n", totalUsers)
	fmt.Printf("  Active: %d\n\n", activeUsers)

	fmt.Printf("Nodes:\n")
	fmt.Printf("  Total: %d\n", totalNodes)
	fmt.Printf("  Online: %d\n\n", onlineNodes)

	fmt.Printf("Sessions:\n")
	fmt.Printf("  Total: %d\n", totalSessions)
	fmt.Printf("  Active: %d\n", activeSessions)
}
