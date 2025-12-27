package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/nikola43/aureo-vpn/pkg/database"
	"github.com/nikola43/aureo-vpn/pkg/models"
	"github.com/nikola43/aureo-vpn/pkg/p2p"
)

// Config holds all node configuration
type Config struct {
	// Node identity
	Name        string
	NodeID      string
	DataDir     string

	// Network
	PublicIP    string
	APIPort     int
	P2PPort     int
	WGPort      int

	// Location
	Country     string
	CountryCode string
	City        string

	// P2P
	BootstrapPeers []string
	AnnounceAddrs  []string

	// Features
	EnableAPI bool
	EnableVPN bool
}

func main() {
	config := parseFlags()

	log.Println("=================================================")
	log.Println("       AUREO VPN - Decentralized Node")
	log.Println("=================================================")

	// Create data directory
	if err := os.MkdirAll(config.DataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Initialize SQLite database
	dbPath := filepath.Join(config.DataDir, "node.db")
	if err := database.ConnectSQLite(database.SQLiteConfig{Path: dbPath}); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := database.MigrateSQLite(models.AllLocalModels()...); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Load or create node identity
	identity, err := loadOrCreateIdentity(config)
	if err != nil {
		log.Fatalf("Failed to load identity: %v", err)
	}

	log.Printf("Node ID: %s", identity.NodeID)
	log.Printf("Node Name: %s", identity.Name)

	// Create and start the decentralized node
	node, err := NewDecentralizedNode(config, identity)
	if err != nil {
		log.Fatalf("Failed to create node: %v", err)
	}

	if err := node.Start(); err != nil {
		log.Fatalf("Failed to start node: %v", err)
	}

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Println("=================================================")
	log.Printf("  API Server: http://0.0.0.0:%d", config.APIPort)
	log.Printf("  P2P Port:   %d", config.P2PPort)
	log.Printf("  WireGuard:  %d", config.WGPort)
	log.Println("=================================================")
	log.Println("Node is running. Press Ctrl+C to stop.")

	<-sigChan

	log.Println("\nShutting down...")
	if err := node.Stop(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	log.Println("Node stopped successfully")
}

func parseFlags() Config {
	config := Config{}

	// Node identity
	flag.StringVar(&config.Name, "name", "", "Node name (auto-generated if empty)")
	flag.StringVar(&config.NodeID, "node-id", "", "Node ID (auto-generated if empty)")
	flag.StringVar(&config.DataDir, "data-dir", "./data", "Data directory")

	// Network
	flag.StringVar(&config.PublicIP, "public-ip", "", "Public IP address (auto-detected if empty)")
	flag.IntVar(&config.APIPort, "api-port", 8080, "API server port")
	flag.IntVar(&config.P2PPort, "p2p-port", 4001, "P2P port")
	flag.IntVar(&config.WGPort, "wg-port", 51820, "WireGuard port")

	// Location
	flag.StringVar(&config.Country, "country", "", "Country name")
	flag.StringVar(&config.CountryCode, "country-code", "", "Country code (e.g., US)")
	flag.StringVar(&config.City, "city", "", "City name")

	// P2P
	var bootstrapPeers string
	var announceAddrs string
	flag.StringVar(&bootstrapPeers, "bootstrap", "", "Comma-separated bootstrap peer multiaddrs")
	flag.StringVar(&announceAddrs, "announce", "", "Comma-separated announce addresses")

	// Features
	flag.BoolVar(&config.EnableAPI, "enable-api", true, "Enable API server")
	flag.BoolVar(&config.EnableVPN, "enable-vpn", true, "Enable VPN service")

	flag.Parse()

	// Parse comma-separated values
	if bootstrapPeers != "" {
		config.BootstrapPeers = strings.Split(bootstrapPeers, ",")
	}
	if announceAddrs != "" {
		config.AnnounceAddrs = strings.Split(announceAddrs, ",")
	}

	// Auto-detect public IP if not provided
	if config.PublicIP == "" {
		config.PublicIP = detectPublicIP()
	}

	// Auto-generate name if not provided
	if config.Name == "" {
		hostname, _ := os.Hostname()
		if hostname != "" {
			config.Name = hostname
		} else {
			config.Name = fmt.Sprintf("node-%s", uuid.New().String()[:8])
		}
	}

	// Build announce address if not provided
	if len(config.AnnounceAddrs) == 0 && config.PublicIP != "" {
		config.AnnounceAddrs = []string{
			fmt.Sprintf("/ip4/%s/tcp/%d", config.PublicIP, config.P2PPort),
		}
	}

	return config
}

func loadOrCreateIdentity(config Config) (*models.NodeIdentity, error) {
	db := database.GetDB()

	var identity models.NodeIdentity
	result := db.First(&identity)

	if result.Error == nil {
		// Identity exists
		return &identity, nil
	}

	// Create new identity
	log.Println("Creating new node identity...")

	// Generate node ID
	nodeID := config.NodeID
	if nodeID == "" {
		nodeID = uuid.New().String()
	}

	// Generate WireGuard keys
	wgPrivKey, wgPubKey, err := generateWireGuardKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to generate WireGuard keys: %w", err)
	}

	// P2P keys will be generated when P2P starts
	identity = models.NodeIdentity{
		NodeID:       nodeID,
		Name:         config.Name,
		WGPrivateKey: wgPrivKey,
		WGPublicKey:  wgPubKey,
		CreatedAt:    time.Now(),
	}

	if err := db.Create(&identity).Error; err != nil {
		return nil, fmt.Errorf("failed to save identity: %w", err)
	}

	return &identity, nil
}

func generateWireGuardKeys() (privateKey, publicKey string, err error) {
	// Use wg genkey and wg pubkey commands
	cmd := "wg genkey"
	privKeyBytes, err := runCommand(cmd)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %w", err)
	}
	privateKey = strings.TrimSpace(string(privKeyBytes))

	cmd = fmt.Sprintf("echo '%s' | wg pubkey", privateKey)
	pubKeyBytes, err := runCommand(cmd)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate public key: %w", err)
	}
	publicKey = strings.TrimSpace(string(pubKeyBytes))

	return privateKey, publicKey, nil
}

func runCommand(cmd string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return execCommand(ctx, "sh", "-c", cmd)
}

func detectPublicIP() string {
	// Try to detect public IP
	services := []string{
		"https://api.ipify.org",
		"https://ifconfig.me/ip",
		"https://icanhazip.com",
	}

	for _, url := range services {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		output, err := execCommand(ctx, "curl", "-s", url)
		cancel()
		if err == nil {
			ip := strings.TrimSpace(string(output))
			if ip != "" {
				log.Printf("Detected public IP: %s", ip)
				return ip
			}
		}
	}

	log.Println("Warning: Could not detect public IP")
	return ""
}

func execCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Output()
}

// DecentralizedNode manages all node services
type DecentralizedNode struct {
	config   Config
	identity *models.NodeIdentity
	p2pHost  *p2p.Host
	api      *APIServer
	vpn      *VPNService
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewDecentralizedNode creates a new decentralized node
func NewDecentralizedNode(config Config, identity *models.NodeIdentity) (*DecentralizedNode, error) {
	ctx, cancel := context.WithCancel(context.Background())

	return &DecentralizedNode{
		config:   config,
		identity: identity,
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

// Start starts all node services
func (n *DecentralizedNode) Start() error {
	// Start P2P network
	if err := n.startP2P(); err != nil {
		return fmt.Errorf("failed to start P2P: %w", err)
	}

	// Start VPN service
	if n.config.EnableVPN {
		if err := n.startVPN(); err != nil {
			log.Printf("Warning: VPN service failed to start: %v", err)
		}
	}

	// Start API server
	if n.config.EnableAPI {
		if err := n.startAPI(); err != nil {
			return fmt.Errorf("failed to start API: %w", err)
		}
	}

	// Start background tasks
	go n.heartbeatLoop()
	go n.cleanupLoop()

	return nil
}

// Stop stops all node services
func (n *DecentralizedNode) Stop() error {
	n.cancel()

	if n.api != nil {
		n.api.Stop()
	}

	if n.vpn != nil {
		n.vpn.Stop()
	}

	if n.p2pHost != nil {
		n.p2pHost.Stop()
	}

	return nil
}

func (n *DecentralizedNode) startP2P() error {
	log.Println("Starting P2P network...")

	p2pConfig := p2p.Config{
		NodeID:            uuid.MustParse(n.identity.NodeID),
		ListenAddrs: []string{
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", n.config.P2PPort),
			fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", n.config.P2PPort),
		},
		AnnounceAddrs:     n.config.AnnounceAddrs,
		BootstrapPeers:    n.config.BootstrapPeers,
		EnableDHT:         true,
		DHTServerMode:     true,
		EnableMDNS:        true,
		EnablePubSub:      true,
		HeartbeatInterval: 30 * time.Second,
		AnnounceInterval:  5 * time.Minute,
		NodeTimeout:       2 * time.Minute,
		MaxPeers:          100,
		MaxNodes:          1000,
		AnnounceTopic:     "/aureo/nodes/announce/1.0.0",
		HeartbeatTopic:    "/aureo/nodes/heartbeat/1.0.0",
		DataDir:           filepath.Join(n.config.DataDir, "p2p"),
	}

	// Use existing P2P key if available
	if n.identity.P2PPrivateKey != "" {
		p2pConfig.PrivateKey = n.identity.P2PPrivateKey
	}

	host, err := p2p.NewHost(p2pConfig)
	if err != nil {
		return err
	}
	n.p2pHost = host

	// Save P2P identity if new
	if n.identity.P2PPrivateKey == "" {
		n.identity.P2PPrivateKey = host.GetConfig().PrivateKey
		n.identity.P2PPeerID = host.GetPeerID().String()
		database.GetDB().Save(n.identity)
	}

	// Set local node info
	nodeInfo := &p2p.NodeInfo{
		ID:                uuid.MustParse(n.identity.NodeID),
		PeerID:            host.GetPeerID(),
		Name:              n.identity.Name,
		PublicIP:          n.config.PublicIP,
		WireGuardPort:     n.config.WGPort,
		PublicKey:         n.identity.WGPublicKey,
		Country:           n.config.Country,
		CountryCode:       n.config.CountryCode,
		City:              n.config.City,
		MaxConnections:    100,
		Status:            "online",
		SupportsWireGuard: true,
		SupportsOpenVPN:   false,
		LastHeartbeat:     time.Now(),
		Version:           "1.0.0",
	}
	host.SetLocalNode(nodeInfo)

	// Set callbacks
	host.SetCallbacks(
		func(node *p2p.NodeInfo) {
			log.Printf("[P2P] Node joined: %s (%s)", node.Name, node.CountryCode)
		},
		func(node *p2p.NodeInfo) {
			// Node updated
		},
		func(id uuid.UUID) {
			log.Printf("[P2P] Node left: %s", id)
		},
	)

	if err := host.Start(); err != nil {
		return err
	}

	log.Printf("[P2P] Peer ID: %s", host.GetPeerID())
	log.Printf("[P2P] Listening on: %v", host.GetMultiaddrs())

	return nil
}

func (n *DecentralizedNode) startVPN() error {
	log.Println("Starting VPN service...")

	vpn, err := NewVPNService(n.config, n.identity)
	if err != nil {
		return err
	}
	n.vpn = vpn

	return vpn.Start()
}

func (n *DecentralizedNode) startAPI() error {
	log.Println("Starting API server...")

	api := NewAPIServer(n.config, n.identity, n.p2pHost, n.vpn)
	n.api = api

	go api.Start()

	return nil
}

func (n *DecentralizedNode) heartbeatLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			n.updateStatus()
		}
	}
}

func (n *DecentralizedNode) updateStatus() {
	if n.p2pHost == nil {
		return
	}

	connections := 0
	loadScore := 0.0

	if n.vpn != nil {
		connections = n.vpn.GetConnectionCount()
		loadScore = n.vpn.GetLoadScore()
	}

	n.p2pHost.UpdateLocalNode("online", connections, loadScore, 0, 0, 0)
}

func (n *DecentralizedNode) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			// Cleanup old sessions
			db := database.GetDB()
			cutoff := time.Now().Add(-24 * time.Hour)
			db.Where("status = ? AND disconnected_at < ?", "disconnected", cutoff).
				Delete(&models.LocalSession{})
		}
	}
}
