package control

import (
	"context"
	"log"
	"time"

	"github.com/nikola43/aureo-vpn/pkg/database"
	"github.com/nikola43/aureo-vpn/pkg/models"
	"gorm.io/gorm"
)

// Server manages the control plane for VPN infrastructure
type Server struct {
	db     *gorm.DB
	ctx    context.Context
	cancel context.CancelFunc
}

// NewServer creates a new control server
func NewServer() *Server {
	ctx, cancel := context.WithCancel(context.Background())

	return &Server{
		db:     database.GetDB(),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start starts the control server
func (s *Server) Start() error {
	log.Println("Starting Control Server...")

	// Start background tasks
	go s.healthCheckLoop()
	go s.loadBalancerLoop()
	go s.cleanupLoop()

	log.Println("Control Server started successfully")
	return nil
}

// Stop stops the control server
func (s *Server) Stop() error {
	log.Println("Stopping Control Server...")
	s.cancel()
	return nil
}

// healthCheckLoop performs periodic health checks on all nodes
func (s *Server) healthCheckLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.performHealthChecks()
		}
	}
}

func (s *Server) performHealthChecks() {
	var nodes []models.VPNNode
	s.db.Where("is_active = ?", true).Find(&nodes)

	for _, node := range nodes {
		// Check if node is healthy based on last heartbeat
		if time.Since(node.LastHeartbeat) > 2*time.Minute {
			// Mark node as offline
			s.db.Model(&node).Updates(map[string]interface{}{
				"status": "offline",
			})
			log.Printf("Node %s marked as offline (no heartbeat)", node.Name)
		} else if node.Status != "online" {
			// Mark node as online if it was offline but is now sending heartbeats
			s.db.Model(&node).Updates(map[string]interface{}{
				"status": "online",
			})
			log.Printf("Node %s is back online", node.Name)
		}

		// Update last health check time
		s.db.Model(&node).UpdateColumn("last_health_check", time.Now())
	}
}

// loadBalancerLoop periodically updates load scores for nodes
func (s *Server) loadBalancerLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.updateLoadScores()
		}
	}
}

func (s *Server) updateLoadScores() {
	var nodes []models.VPNNode
	s.db.Where("is_active = ? AND status = ?", true, "online").Find(&nodes)

	for _, node := range nodes {
		// Calculate load score
		loadScore := node.CalculateLoadScore()

		// Update in database
		s.db.Model(&node).UpdateColumn("load_score", loadScore)

		// Log if node is overloaded
		if loadScore > 80 {
			log.Printf("WARNING: Node %s is heavily loaded (score: %.2f)", node.Name, loadScore)
		}
	}
}

// cleanupLoop performs periodic cleanup of old sessions and data
func (s *Server) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.performCleanup()
		}
	}
}

func (s *Server) performCleanup() {
	// Clean up old disconnected sessions (older than 30 days)
	cutoffTime := time.Now().AddDate(0, 0, -30)
	result := s.db.Where("status = ? AND disconnected_at < ?", "disconnected", cutoffTime).
		Delete(&models.Session{})

	if result.RowsAffected > 0 {
		log.Printf("Cleaned up %d old sessions", result.RowsAffected)
	}

	// Clean up expired configs
	result = s.db.Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).
		Delete(&models.Config{})

	if result.RowsAffected > 0 {
		log.Printf("Cleaned up %d expired configs", result.RowsAffected)
	}

	// Find and fix orphaned sessions (sessions where node is offline)
	var orphanedSessions []models.Session
	s.db.Joins("JOIN vpn_nodes ON vpn_nodes.id = sessions.node_id").
		Where("sessions.status = ? AND vpn_nodes.status = ?", "active", "offline").
		Find(&orphanedSessions)

	for _, session := range orphanedSessions {
		now := time.Now()
		s.db.Model(&session).Updates(map[string]interface{}{
			"status":          "disconnected",
			"disconnected_at": &now,
		})
	}

	if len(orphanedSessions) > 0 {
		log.Printf("Fixed %d orphaned sessions", len(orphanedSessions))
	}
}

// RegisterNode registers a new VPN node
func (s *Server) RegisterNode(node *models.VPNNode) error {
	node.Status = "offline"
	node.IsActive = true
	node.LastHeartbeat = time.Now()

	if err := s.db.Create(node).Error; err != nil {
		return err
	}

	log.Printf("Registered new node: %s (%s)", node.Name, node.PublicIP)
	return nil
}

// UpdateNodeStatus updates the status of a node
func (s *Server) UpdateNodeStatus(nodeID string, status string) error {
	updates := map[string]interface{}{
		"status":         status,
		"last_heartbeat": time.Now(),
	}

	if err := s.db.Model(&models.VPNNode{}).Where("id = ?", nodeID).Updates(updates).Error; err != nil {
		return err
	}

	return nil
}

// GetBestNode returns the best available node for a new connection
func (s *Server) GetBestNode(protocol, country string) (*models.VPNNode, error) {
	query := s.db.Where("is_active = ? AND status = ?", true, "online")

	if country != "" {
		query = query.Where("country_code = ?", country)
	}

	if protocol == "wireguard" {
		query = query.Where("supports_wireguard = ?", true)
	} else if protocol == "openvpn" {
		query = query.Where("supports_openvpn = ?", true)
	}

	var node models.VPNNode
	if err := query.Order("load_score ASC, latency ASC").First(&node).Error; err != nil {
		return nil, err
	}

	return &node, nil
}

// GetNodeStats returns statistics for all nodes
func (s *Server) GetNodeStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalNodes, onlineNodes, offlineNodes int64
	s.db.Model(&models.VPNNode{}).Count(&totalNodes)
	s.db.Model(&models.VPNNode{}).Where("status = ?", "online").Count(&onlineNodes)
	s.db.Model(&models.VPNNode{}).Where("status = ?", "offline").Count(&offlineNodes)

	var totalConnections int64
	s.db.Model(&models.Session{}).Where("status = ?", "active").Count(&totalConnections)

	stats["total_nodes"] = totalNodes
	stats["online_nodes"] = onlineNodes
	stats["offline_nodes"] = offlineNodes
	stats["total_active_connections"] = totalConnections

	// Get nodes by country
	var countryStats []struct {
		Country string
		Count   int64
	}
	s.db.Model(&models.VPNNode{}).
		Select("country, count(*) as count").
		Group("country").
		Scan(&countryStats)

	stats["nodes_by_country"] = countryStats

	return stats, nil
}
