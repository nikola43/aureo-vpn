package main

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
	"github.com/nikola43/aureo-vpn/pkg/models"
)

// VPNService manages WireGuard VPN connections
type VPNService struct {
	config   Config
	identity *models.NodeIdentity
	sessions map[uuid.UUID]*SessionInfo
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc

	// Stats
	totalBytesIn  int64
	totalBytesOut int64
}

// SessionInfo holds active session information
type SessionInfo struct {
	Session       *models.LocalSession
	PublicKey     string
	LastKeepalive time.Time
	BytesIn       int64
	BytesOut      int64
}

// NewVPNService creates a new VPN service
func NewVPNService(config Config, identity *models.NodeIdentity) (*VPNService, error) {
	ctx, cancel := context.WithCancel(context.Background())

	return &VPNService{
		config:   config,
		identity: identity,
		sessions: make(map[uuid.UUID]*SessionInfo),
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

// Start starts the VPN service
func (v *VPNService) Start() error {
	// Setup WireGuard interface
	if err := v.setupWireGuard(); err != nil {
		return fmt.Errorf("failed to setup WireGuard: %w", err)
	}

	// Load active sessions from database
	v.loadActiveSessions()

	// Start background tasks
	go v.sessionMonitor()

	log.Println("[VPN] WireGuard service started")
	return nil
}

// Stop stops the VPN service
func (v *VPNService) Stop() error {
	v.cancel()

	// Disconnect all sessions
	v.mu.Lock()
	for sessionID := range v.sessions {
		v.disconnectSessionLocked(sessionID)
	}
	v.mu.Unlock()

	// Remove WireGuard interface
	exec.Command("ip", "link", "del", "wg0").Run()

	return nil
}

func (v *VPNService) setupWireGuard() error {
	// Check if wg0 exists
	if _, err := exec.Command("ip", "link", "show", "wg0").Output(); err == nil {
		// Interface exists, remove it first
		exec.Command("ip", "link", "del", "wg0").Run()
	}

	// Create WireGuard interface
	cmds := []string{
		"ip link add wg0 type wireguard",
		fmt.Sprintf("ip addr add 10.10.0.1/24 dev wg0"),
		"ip link set wg0 up",
	}

	for _, cmd := range cmds {
		if err := exec.Command("sh", "-c", cmd).Run(); err != nil {
			return fmt.Errorf("failed to run '%s': %w", cmd, err)
		}
	}

	// Configure WireGuard
	configCmd := fmt.Sprintf("wg set wg0 private-key <(echo '%s') listen-port %d",
		v.identity.WGPrivateKey, v.config.WGPort)
	if err := exec.Command("bash", "-c", configCmd).Run(); err != nil {
		return fmt.Errorf("failed to configure WireGuard: %w", err)
	}

	// Enable IP forwarding and NAT
	natCmds := []string{
		"iptables -A FORWARD -i wg0 -j ACCEPT",
		"iptables -A FORWARD -o wg0 -j ACCEPT",
		"iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE",
	}

	for _, cmd := range natCmds {
		exec.Command("sh", "-c", cmd).Run() // Ignore errors, might already exist
	}

	return nil
}

func (v *VPNService) loadActiveSessions() {
	db := database.GetDB()

	var sessions []models.LocalSession
	db.Where("status = ?", "active").Find(&sessions)

	for _, session := range sessions {
		// Try to restore WireGuard peer
		if session.PublicKey != "" && session.TunnelIP != "" {
			cmd := fmt.Sprintf("wg set wg0 peer %s allowed-ips %s/32",
				session.PublicKey, session.TunnelIP)
			if err := exec.Command("sh", "-c", cmd).Run(); err != nil {
				// Failed to restore, mark as disconnected
				now := time.Now()
				session.Status = "disconnected"
				session.DisconnectedAt = &now
				db.Save(&session)
				continue
			}

			v.mu.Lock()
			v.sessions[session.ID] = &SessionInfo{
				Session:       &session,
				PublicKey:     session.PublicKey,
				LastKeepalive: time.Now(),
			}
			v.mu.Unlock()
		}
	}

	log.Printf("[VPN] Restored %d active sessions", len(v.sessions))
}

// CreateSession creates a new VPN session
func (v *VPNService) CreateSession(userID uuid.UUID, clientPublicKey string) (*models.LocalSession, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Allocate IP address
	tunnelIP := v.allocateIP()
	if tunnelIP == "" {
		return nil, fmt.Errorf("no available IP addresses")
	}

	// Create session
	session := &models.LocalSession{
		ID:            uuid.New(),
		UserID:        userID,
		Protocol:      "wireguard",
		TunnelIP:      tunnelIP,
		PublicKey:     clientPublicKey,
		Status:        "active",
		ConnectedAt:   time.Now(),
		LastKeepalive: time.Now(),
	}

	// Add WireGuard peer
	cmd := fmt.Sprintf("wg set wg0 peer %s allowed-ips %s/32 persistent-keepalive 25",
		clientPublicKey, tunnelIP)
	if err := exec.Command("sh", "-c", cmd).Run(); err != nil {
		return nil, fmt.Errorf("failed to add WireGuard peer: %w", err)
	}

	// Save to database
	db := database.GetDB()
	if err := db.Create(session).Error; err != nil {
		// Rollback WireGuard peer
		exec.Command("sh", "-c", fmt.Sprintf("wg set wg0 peer %s remove", clientPublicKey)).Run()
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	// Track session
	v.sessions[session.ID] = &SessionInfo{
		Session:       session,
		PublicKey:     clientPublicKey,
		LastKeepalive: time.Now(),
	}

	log.Printf("[VPN] Created session %s for user %s (IP: %s)", session.ID, userID, tunnelIP)
	return session, nil
}

// DisconnectSession disconnects a VPN session
func (v *VPNService) DisconnectSession(sessionID uuid.UUID) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	return v.disconnectSessionLocked(sessionID)
}

func (v *VPNService) disconnectSessionLocked(sessionID uuid.UUID) error {
	info, ok := v.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found")
	}

	// Remove WireGuard peer
	cmd := fmt.Sprintf("wg set wg0 peer %s remove", info.PublicKey)
	exec.Command("sh", "-c", cmd).Run()

	// Update database
	db := database.GetDB()
	now := time.Now()
	db.Model(&models.LocalSession{}).Where("id = ?", sessionID).Updates(map[string]interface{}{
		"status":         "disconnected",
		"disconnected_at": &now,
	})

	// Remove from tracking
	delete(v.sessions, sessionID)

	log.Printf("[VPN] Disconnected session %s", sessionID)
	return nil
}

func (v *VPNService) allocateIP() string {
	// Simple IP allocation from 10.10.0.2 to 10.10.0.254
	usedIPs := make(map[string]bool)

	v.mu.RLock()
	for _, info := range v.sessions {
		usedIPs[info.Session.TunnelIP] = true
	}
	v.mu.RUnlock()

	for i := 2; i <= 254; i++ {
		ip := fmt.Sprintf("10.10.0.%d", i)
		if !usedIPs[ip] {
			return ip
		}
	}

	return ""
}

func (v *VPNService) sessionMonitor() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-v.ctx.Done():
			return
		case <-ticker.C:
			v.checkSessions()
		}
	}
}

func (v *VPNService) checkSessions() {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Get WireGuard peer stats
	output, err := exec.Command("wg", "show", "wg0", "dump").Output()
	if err != nil {
		return
	}

	// Parse peer stats
	peerStats := make(map[string]time.Time)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines[1:] { // Skip header
		fields := strings.Fields(line)
		if len(fields) >= 5 {
			pubKey := fields[0]
			handshake := fields[4]
			if handshake != "0" {
				// Parse handshake timestamp
				var ts int64
				fmt.Sscanf(handshake, "%d", &ts)
				if ts > 0 {
					peerStats[pubKey] = time.Unix(ts, 0)
				}
			}
		}
	}

	// Check for inactive sessions
	for sessionID, info := range v.sessions {
		if lastSeen, ok := peerStats[info.PublicKey]; ok {
			info.LastKeepalive = lastSeen
		}

		// Disconnect if no activity for 10 minutes
		if time.Since(info.LastKeepalive) > 10*time.Minute {
			log.Printf("[VPN] Session %s inactive, disconnecting", sessionID)
			v.disconnectSessionLocked(sessionID)
		}
	}
}

// GetConnectionCount returns the number of active connections
func (v *VPNService) GetConnectionCount() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return len(v.sessions)
}

// GetLoadScore returns the current load score (0-100)
func (v *VPNService) GetLoadScore() float64 {
	count := v.GetConnectionCount()
	maxConnections := 100 // Could be configurable

	return float64(count) / float64(maxConnections) * 100
}

// GetSession returns a session by ID
func (v *VPNService) GetSession(sessionID uuid.UUID) *models.LocalSession {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if info, ok := v.sessions[sessionID]; ok {
		return info.Session
	}
	return nil
}

// GetServerConfig returns the WireGuard server configuration for clients
func (v *VPNService) GetServerConfig() map[string]interface{} {
	return map[string]interface{}{
		"public_key": v.identity.WGPublicKey,
		"endpoint":   fmt.Sprintf("%s:%d", v.config.PublicIP, v.config.WGPort),
		"dns":        []string{"1.1.1.1", "8.8.8.8"},
		"allowed_ips": "0.0.0.0/0",
	}
}
