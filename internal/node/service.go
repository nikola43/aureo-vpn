package node

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nikola43/aureo-vpn/pkg/database"
	"github.com/nikola43/aureo-vpn/pkg/metrics"
	"github.com/nikola43/aureo-vpn/pkg/models"
	"github.com/nikola43/aureo-vpn/pkg/protocols/wireguard"
	"gorm.io/gorm"
)

// Service manages VPN node operations
type Service struct {
	nodeID         uuid.UUID
	db             *gorm.DB
	wgManager      *wireguard.Manager
	activeSessions map[uuid.UUID]*SessionInfo
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc

	// Traffic monitoring
	lastBytesSent     int64
	lastBytesReceived int64
	lastTrafficCheck  time.Time
	trafficMu         sync.RWMutex
}

// SessionInfo holds session information
type SessionInfo struct {
	Session       *models.Session
	PublicKey     string
	LastKeepalive time.Time
}

// NewService creates a new VPN node service
func NewService(nodeID uuid.UUID) *Service {
	ctx, cancel := context.WithCancel(context.Background())

	return &Service{
		nodeID:         nodeID,
		db:             database.GetDB(),
		wgManager:      wireguard.NewManager("wg0"),
		activeSessions: make(map[uuid.UUID]*SessionInfo),
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Start starts the VPN node service
func (s *Service) Start() error {
	log.Println("Starting VPN Node Service...")

	// Load node configuration
	var node models.VPNNode
	if err := s.db.First(&node, s.nodeID).Error; err != nil {
		return fmt.Errorf("failed to load node: %w", err)
	}

	// Generate server keypair if not exists
	var privateKey string
	if node.PublicKey == "" {
		keyPair, err := wireguard.GenerateKeyPair()
		if err != nil {
			return fmt.Errorf("failed to generate keypair: %w", err)
		}

		node.PublicKey = keyPair.PublicKey
		privateKey = keyPair.PrivateKey
		// Store private key securely (in production, use KMS/Vault)
		// For now, store in PrivateKeyEncrypted field
		if err := s.db.Model(&node).Updates(map[string]interface{}{
			"public_key":            keyPair.PublicKey,
			"private_key_encrypted": keyPair.PrivateKey,
		}).Error; err != nil {
			return fmt.Errorf("failed to save keypair: %w", err)
		}
	} else {
		// Load existing private key from database
		var storedNode models.VPNNode
		if err := s.db.Select("private_key_encrypted").First(&storedNode, s.nodeID).Error; err != nil {
			return fmt.Errorf("failed to load private key: %w", err)
		}
		privateKey = storedNode.PrivateKeyEncrypted

		// If no private key exists, generate a new one
		if privateKey == "" {
			keyPair, err := wireguard.GenerateKeyPair()
			if err != nil {
				return fmt.Errorf("failed to generate keypair: %w", err)
			}
			privateKey = keyPair.PrivateKey

			if err := s.db.Model(&node).UpdateColumn("private_key_encrypted", privateKey).Error; err != nil {
				return fmt.Errorf("failed to save private key: %w", err)
			}
		}
	}

	// Setup WireGuard interface
	if err := s.setupWireGuard(&node, privateKey); err != nil {
		return fmt.Errorf("failed to setup WireGuard: %w", err)
	}

	// Start background tasks
	go s.heartbeatLoop()
	go s.sessionMonitor()
	go s.metricsCollector()
	go s.trafficMonitor()

	log.Println("VPN Node Service started successfully")
	return nil
}

// Stop stops the VPN node service
func (s *Service) Stop() error {
	log.Println("Stopping VPN Node Service...")
	s.cancel()

	// Disconnect all sessions
	s.mu.Lock()
	for sessionID := range s.activeSessions {
		s.disconnectSession(sessionID)
	}
	s.mu.Unlock()

	return nil
}

// setupWireGuard configures the WireGuard interface
func (s *Service) setupWireGuard(node *models.VPNNode, privateKey string) error {
	config := wireguard.ServerConfig{
		PrivateKey: privateKey,
		Address:    node.InternalIP + "/24",
		ListenPort: node.WireGuardPort,
		PostUp: []string{
			"iptables -A FORWARD -i wg0 -j ACCEPT",
			"iptables -A FORWARD -o wg0 -j ACCEPT",
			"iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE",
		},
		PostDown: []string{
			"iptables -D FORWARD -i wg0 -j ACCEPT",
			"iptables -D FORWARD -o wg0 -j ACCEPT",
			"iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE",
		},
	}

	return s.wgManager.SetupInterface(config)
}

// CreateSession creates a new VPN session
func (s *Service) CreateSession(userID uuid.UUID, protocol string) (*models.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Load node
	var node models.VPNNode
	if err := s.db.First(&node, s.nodeID).Error; err != nil {
		return nil, fmt.Errorf("failed to load node: %w", err)
	}

	// Check if node has capacity
	if node.CurrentConnections >= node.MaxConnections {
		return nil, fmt.Errorf("node at maximum capacity")
	}

	// Generate client keypair
	keyPair, err := wireguard.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate keypair: %w", err)
	}

	// Allocate IP address
	var usedIPs []string
	s.db.Model(&models.Session{}).
		Where("node_id = ? AND status = ?", s.nodeID, "active").
		Pluck("tunnel_ip", &usedIPs)

	tunnelIP, err := wireguard.AllocateClientIP(node.InternalIP+"/24", usedIPs)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate IP: %w", err)
	}

	// Create session
	session := &models.Session{
		UserID:             userID,
		NodeID:             s.nodeID,
		Protocol:           protocol,
		TunnelIP:           tunnelIP,
		PublicKey:          keyPair.PublicKey,
		PrivateKey:         keyPair.PrivateKey, // Encrypted in production
		Status:             "active",
		ConnectedAt:        time.Now(),
		LastKeepalive:      time.Now(),
		KillSwitchEnabled:  true,
		DNSLeakProtection:  true,
	}

	if err := s.db.Create(session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Add peer to WireGuard
	peer := wireguard.PeerConfig{
		PublicKey:           keyPair.PublicKey,
		AllowedIPs:          []string{tunnelIP},
		PersistentKeepalive: 25,
	}

	if err := s.wgManager.AddPeer(peer); err != nil {
		s.db.Delete(session)
		return nil, fmt.Errorf("failed to add peer: %w", err)
	}

	// Update node connection count
	s.db.Model(&node).UpdateColumn("current_connections", gorm.Expr("current_connections + ?", 1))

	// Store in active sessions
	s.activeSessions[session.ID] = &SessionInfo{
		Session:       session,
		PublicKey:     keyPair.PublicKey,
		LastKeepalive: time.Now(),
	}

	// Update metrics
	metrics.ActiveConnections.WithLabelValues(protocol, node.Name).Inc()
	metrics.ConnectionsTotal.WithLabelValues(protocol, node.Name, "success").Inc()

	log.Printf("Created session %s for user %s", session.ID, userID)
	return session, nil
}

// DisconnectSession disconnects a VPN session
func (s *Service) DisconnectSession(sessionID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.disconnectSession(sessionID)
}

func (s *Service) disconnectSession(sessionID uuid.UUID) error {
	sessionInfo, ok := s.activeSessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found")
	}

	// Remove peer from WireGuard
	if err := s.wgManager.RemovePeer(sessionInfo.PublicKey); err != nil {
		log.Printf("Failed to remove peer: %v", err)
	}

	// Update session in database
	now := time.Now()
	updates := map[string]interface{}{
		"status":          "disconnected",
		"disconnected_at": &now,
	}

	if err := s.db.Model(&models.Session{}).Where("id = ?", sessionID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	// Update node connection count
	s.db.Model(&models.VPNNode{}).Where("id = ?", s.nodeID).
		UpdateColumn("current_connections", gorm.Expr("current_connections - ?", 1))

	// Remove from active sessions
	delete(s.activeSessions, sessionID)

	// Update metrics
	var node models.VPNNode
	s.db.First(&node, s.nodeID)
	metrics.ActiveConnections.WithLabelValues(sessionInfo.Session.Protocol, node.Name).Dec()

	log.Printf("Disconnected session %s", sessionID)
	return nil
}

// heartbeatLoop sends periodic heartbeats to the control server
func (s *Service) heartbeatLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.sendHeartbeat()
		}
	}
}

func (s *Service) sendHeartbeat() {
	// Count active WireGuard peers
	peerCount := s.countActivePeers()

	updates := map[string]interface{}{
		"last_heartbeat":      time.Now(),
		"status":              "online",
		"current_connections": peerCount,
	}

	if err := s.db.Model(&models.VPNNode{}).Where("id = ?", s.nodeID).Updates(updates).Error; err != nil {
		log.Printf("Failed to send heartbeat: %v", err)
	}
}

// countActivePeers counts the number of active WireGuard peers
func (s *Service) countActivePeers() int {
	cmd := exec.Command("wg", "show", "wg0", "peers")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	// Count non-empty lines (each line is a peer public key)
	peers := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(peers) == 1 && peers[0] == "" {
		return 0
	}
	return len(peers)
}

// sessionMonitor monitors active sessions
func (s *Service) sessionMonitor() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.checkInactiveSessions()
		}
	}
}

func (s *Service) checkInactiveSessions() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for sessionID, sessionInfo := range s.activeSessions {
		// Disconnect sessions with no keepalive for 10 minutes
		if time.Since(sessionInfo.LastKeepalive) > 10*time.Minute {
			log.Printf("Session %s inactive, disconnecting", sessionID)
			s.disconnectSession(sessionID)
		}
	}
}

// metricsCollector collects and updates metrics
func (s *Service) metricsCollector() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.collectMetrics()
		}
	}
}

func (s *Service) collectMetrics() {
	var node models.VPNNode
	if err := s.db.First(&node, s.nodeID).Error; err != nil {
		return
	}

	// Update load score
	loadScore := node.CalculateLoadScore()
	s.db.Model(&node).UpdateColumn("load_score", loadScore)

	// Update metrics
	status := 0.0
	if node.Status == "online" {
		status = 1.0
	}

	metrics.NodeStatus.WithLabelValues(node.Name, node.Country, node.City).Set(status)
	metrics.NodeLoad.WithLabelValues(node.Name).Set(loadScore)
	metrics.NodeCPUUsage.WithLabelValues(node.Name).Set(node.CPUUsage)
	metrics.NodeMemoryUsage.WithLabelValues(node.Name).Set(node.MemoryUsage)
	metrics.NodeBandwidth.WithLabelValues(node.Name).Set(node.BandwidthUsageGbps)
}

// trafficMonitor monitors WireGuard traffic and calculates real-time bandwidth usage
func (s *Service) trafficMonitor() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Initialize last check time
	s.trafficMu.Lock()
	s.lastTrafficCheck = time.Now()
	s.trafficMu.Unlock()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.updateTrafficStats()
		}
	}
}

func (s *Service) updateTrafficStats() {
	// Get WireGuard stats
	stats, err := s.wgManager.GetInterfaceStats()
	if err != nil {
		log.Printf("Failed to get WireGuard stats: %v", err)
		return
	}

	// Calculate total bytes from all peers
	var totalBytesSent, totalBytesReceived int64
	for _, peer := range stats.Peers {
		totalBytesSent += peer.BytesSent
		totalBytesReceived += peer.BytesReceived
	}

	s.trafficMu.Lock()
	now := time.Now()
	timeDiff := now.Sub(s.lastTrafficCheck).Seconds()

	// Calculate rate (bytes per second)
	var currentTrafficMbps float64
	if timeDiff > 0 && s.lastTrafficCheck.Unix() > 0 {
		bytesSentDiff := totalBytesSent - s.lastBytesSent
		bytesReceivedDiff := totalBytesReceived - s.lastBytesReceived

		// Total bytes per second (sent + received)
		bytesPerSecond := float64(bytesSentDiff+bytesReceivedDiff) / timeDiff

		// Convert to Mbps (megabits per second)
		currentTrafficMbps = (bytesPerSecond * 8) / 1_000_000
	}

	// Update tracking variables
	s.lastBytesSent = totalBytesSent
	s.lastBytesReceived = totalBytesReceived
	s.lastTrafficCheck = now
	s.trafficMu.Unlock()

	// Update node's bandwidth usage in database
	s.db.Model(&models.VPNNode{}).
		Where("id = ?", s.nodeID).
		UpdateColumn("bandwidth_usage_gbps", currentTrafficMbps/1000.0) // Convert Mbps to Gbps
}

// GetConnectedUsers returns the number of currently connected users
func (s *Service) GetConnectedUsers() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.activeSessions)
}

// GetCurrentTrafficMbps returns the current traffic rate in Mbps
func (s *Service) GetCurrentTrafficMbps() float64 {
	var node models.VPNNode
	if err := s.db.First(&node, s.nodeID).Error; err != nil {
		return 0
	}
	return node.BandwidthUsageGbps * 1000.0 // Convert Gbps to Mbps
}
