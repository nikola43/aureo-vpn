package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/nikola43/aureo-vpn/internal/node"
	"github.com/nikola43/aureo-vpn/pkg/database"
	"github.com/nikola43/aureo-vpn/pkg/p2p"
)

func main() {
	log.Println("Starting Aureo VPN Node...")

	// Load configuration
	config := loadConfig()

	// Connect to database
	dbConfig := database.Config{
		Host:     config.DBHost,
		Port:     config.DBPort,
		User:     config.DBUser,
		Password: config.DBPassword,
		DBName:   config.DBName,
		SSLMode:  config.DBSSLMode,
		TimeZone: "UTC",
	}

	if err := database.Connect(dbConfig); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Parse node ID
	nodeID, err := uuid.Parse(config.NodeID)
	if err != nil {
		log.Fatalf("Invalid node ID: %v", err)
	}

	// Build P2P configuration
	p2pConfig := p2p.ServiceConfig{
		NodeID:            nodeID,
		ListenPort:        config.P2PPort,
		BootstrapPeers:    config.P2PBootstrapPeers,
		AnnounceAddrs:     config.P2PAnnounceAddrs,
		EnableDHT:         config.P2PEnableDHT,
		EnableMDNS:        config.P2PEnableMDNS,
		DHTServerMode:     config.P2PDHTServer,
		HeartbeatInterval: 30 * time.Second,
		AnnounceInterval:  5 * time.Minute,
		DataDir:           config.P2PDataDir,
	}

	// Create and start node service with P2P
	nodeService := node.NewServiceWithP2P(nodeID, p2pConfig)
	if err := nodeService.Start(); err != nil {
		log.Fatalf("Failed to start node service: %v", err)
	}

	// Log P2P stats
	stats := nodeService.GetP2PStats()
	if enabled, ok := stats["enabled"].(bool); ok && enabled {
		log.Printf("P2P enabled - Peer ID: %s", stats["peer_id"])
		log.Printf("P2P listening on: %v", stats["multiaddrs"])
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Println("VPN Node is running. Press Ctrl+C to stop.")
	<-sigChan

	log.Println("Shutting down VPN Node...")
	if err := nodeService.Stop(); err != nil {
		log.Printf("Error stopping node service: %v", err)
	}

	log.Println("VPN Node stopped successfully")
}

type Config struct {
	NodeID     string
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// P2P Configuration
	P2PPort           int
	P2PBootstrapPeers []string
	P2PAnnounceAddrs  []string
	P2PEnableDHT      bool
	P2PEnableMDNS     bool
	P2PDHTServer      bool
	P2PDataDir        string
}

func loadConfig() Config {
	bootstrapPeers := getEnv("P2P_BOOTSTRAP_PEERS", "")
	announceAddrs := getEnv("P2P_ANNOUNCE_ADDRS", "")

	var peers, addrs []string
	if bootstrapPeers != "" {
		peers = strings.Split(bootstrapPeers, ",")
	}
	if announceAddrs != "" {
		addrs = strings.Split(announceAddrs, ",")
	}

	return Config{
		NodeID:     getEnv("NODE_ID", ""),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnvAsInt("DB_PORT", 5432),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "aureo_vpn"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),

		// P2P defaults
		P2PPort:           getEnvAsInt("P2P_PORT", 4001),
		P2PBootstrapPeers: peers,
		P2PAnnounceAddrs:  addrs,
		P2PEnableDHT:      getEnvAsBool("P2P_ENABLE_DHT", true),
		P2PEnableMDNS:     getEnvAsBool("P2P_ENABLE_MDNS", true),
		P2PDHTServer:      getEnvAsBool("P2P_DHT_SERVER", true),
		P2PDataDir:        getEnv("P2P_DATA_DIR", "./data/p2p"),
	}
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
	if _, err := fmt.Sscanf(valueStr, "%d", &value); err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	return valueStr == "true" || valueStr == "1" || valueStr == "yes"
}
