package security

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/nikola43/aureo-vpn/pkg/database"
	"github.com/nikola43/aureo-vpn/pkg/models"
	"gorm.io/gorm"
)

// MultiHopManager handles multi-hop (double VPN) routing
type MultiHopManager struct {
	db *gorm.DB
}

// HopChain represents a chain of VPN hops
type HopChain struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	EntryNode   *models.VPNNode
	ExitNode    *models.VPNNode
	MiddleNodes []*models.VPNNode // For triple VPN or more
	Protocol    string
	Status      string
	CreatedAt   string
}

// NewMultiHopManager creates a new multi-hop manager
func NewMultiHopManager() *MultiHopManager {
	return &MultiHopManager{
		db: database.GetDB(),
	}
}

// CreateDoubleVPNChain creates a double VPN connection chain
func (m *MultiHopManager) CreateDoubleVPNChain(userID uuid.UUID, entryCountry, exitCountry string) (*HopChain, error) {
	log.Printf("Creating double VPN chain: %s -> %s", entryCountry, exitCountry)

	// Get entry node (first hop)
	var entryNode models.VPNNode
	err := m.db.Where("country_code = ? AND status = ? AND is_active = ? AND supports_multihop = ?",
		entryCountry, "online", true, true).
		Order("load_score ASC").
		First(&entryNode).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find entry node: %w", err)
	}

	// Get exit node (second hop) - must be different from entry
	var exitNode models.VPNNode
	err = m.db.Where("country_code = ? AND status = ? AND is_active = ? AND supports_multihop = ? AND id != ?",
		exitCountry, "online", true, true, entryNode.ID).
		Order("load_score ASC").
		First(&exitNode).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find exit node: %w", err)
	}

	// Validate nodes are in different locations for true multi-hop
	if entryNode.Country == exitNode.Country && entryNode.City == exitNode.City {
		return nil, fmt.Errorf("entry and exit nodes must be in different locations")
	}

	chain := &HopChain{
		ID:        uuid.New(),
		UserID:    userID,
		EntryNode: &entryNode,
		ExitNode:  &exitNode,
		Protocol:  "wireguard", // Use WireGuard for best performance
		Status:    "active",
	}

	log.Printf("Double VPN chain created: %s (%s) -> %s (%s)",
		entryNode.Name, entryNode.Country,
		exitNode.Name, exitNode.Country)

	return chain, nil
}

// CreateTripleVPNChain creates a triple VPN connection chain (maximum security)
func (m *MultiHopManager) CreateTripleVPNChain(userID uuid.UUID, countries []string) (*HopChain, error) {
	if len(countries) != 3 {
		return nil, fmt.Errorf("triple VPN requires exactly 3 countries")
	}

	log.Printf("Creating triple VPN chain: %s -> %s -> %s", countries[0], countries[1], countries[2])

	// Get entry node
	var entryNode models.VPNNode
	err := m.db.Where("country_code = ? AND status = ? AND is_active = ? AND supports_multihop = ?",
		countries[0], "online", true, true).
		Order("load_score ASC").
		First(&entryNode).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find entry node: %w", err)
	}

	// Get middle node
	var middleNode models.VPNNode
	err = m.db.Where("country_code = ? AND status = ? AND is_active = ? AND supports_multihop = ? AND id != ?",
		countries[1], "online", true, true, entryNode.ID).
		Order("load_score ASC").
		First(&middleNode).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find middle node: %w", err)
	}

	// Get exit node
	var exitNode models.VPNNode
	err = m.db.Where("country_code = ? AND status = ? AND is_active = ? AND supports_multihop = ? AND id NOT IN ?",
		countries[2], "online", true, true, []uuid.UUID{entryNode.ID, middleNode.ID}).
		Order("load_score ASC").
		First(&exitNode).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find exit node: %w", err)
	}

	chain := &HopChain{
		ID:          uuid.New(),
		UserID:      userID,
		EntryNode:   &entryNode,
		MiddleNodes: []*models.VPNNode{&middleNode},
		ExitNode:    &exitNode,
		Protocol:    "wireguard",
		Status:      "active",
	}

	log.Printf("Triple VPN chain created: %s -> %s -> %s",
		entryNode.Name, middleNode.Name, exitNode.Name)

	return chain, nil
}

// GetOptimalMultiHopRoute finds the optimal multi-hop route based on latency
func (m *MultiHopManager) GetOptimalMultiHopRoute(userID uuid.UUID, targetCountry string) (*HopChain, error) {
	// Get all available nodes for multi-hop
	var nodes []models.VPNNode
	m.db.Where("status = ? AND is_active = ? AND supports_multihop = ?",
		"online", true, true).
		Order("latency ASC, load_score ASC").
		Find(&nodes)

	if len(nodes) < 2 {
		return nil, fmt.Errorf("insufficient nodes for multi-hop (need at least 2)")
	}

	// Find entry node with lowest latency
	entryNode := nodes[0]

	// Find exit node in target country with good performance
	var exitNode models.VPNNode
	found := false
	for _, node := range nodes {
		if node.CountryCode == targetCountry && node.ID != entryNode.ID {
			exitNode = node
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("no suitable exit node found in %s", targetCountry)
	}

	chain := &HopChain{
		ID:        uuid.New(),
		UserID:    userID,
		EntryNode: &entryNode,
		ExitNode:  &exitNode,
		Protocol:  "wireguard",
		Status:    "active",
	}

	return chain, nil
}

// CalculateChainLatency estimates total latency for a hop chain
func (m *MultiHopManager) CalculateChainLatency(chain *HopChain) int {
	totalLatency := chain.EntryNode.Latency + chain.ExitNode.Latency

	for _, middleNode := range chain.MiddleNodes {
		totalLatency += middleNode.Latency
	}

	// Add overhead for each hop (encryption/decryption)
	hops := 2 + len(chain.MiddleNodes)
	overhead := hops * 5 // ~5ms per hop

	return totalLatency + overhead
}

// ValidateMultiHopChain ensures the chain is secure and performant
func (m *MultiHopManager) ValidateMultiHopChain(chain *HopChain) error {
	// Check all nodes are online
	if !chain.EntryNode.IsHealthy() {
		return fmt.Errorf("entry node is not healthy")
	}

	if !chain.ExitNode.IsHealthy() {
		return fmt.Errorf("exit node is not healthy")
	}

	for _, node := range chain.MiddleNodes {
		if !node.IsHealthy() {
			return fmt.Errorf("middle node %s is not healthy", node.Name)
		}
	}

	// Check latency is acceptable (< 200ms total)
	if m.CalculateChainLatency(chain) > 200 {
		log.Printf("WARNING: High latency in multi-hop chain: %dms", m.CalculateChainLatency(chain))
	}

	// Check geographic diversity
	if chain.EntryNode.Country == chain.ExitNode.Country {
		return fmt.Errorf("entry and exit nodes should be in different countries")
	}

	return nil
}

// GetMultiHopStatistics returns statistics for multi-hop connections
func (m *MultiHopManager) GetMultiHopStatistics(userID uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count multi-hop sessions
	var multiHopCount int64
	m.db.Model(&models.Session{}).
		Where("user_id = ? AND is_multihop = ?", userID, true).
		Count(&multiHopCount)

	stats["total_multihop_sessions"] = multiHopCount

	// Get most used routes
	type RouteStats struct {
		EntryCountry string
		ExitCountry  string
		Count        int64
	}

	var routeStats []RouteStats
	m.db.Raw(`
		SELECT
			entry.country as entry_country,
			exit.country as exit_country,
			COUNT(*) as count
		FROM sessions s
		INNER JOIN vpn_nodes entry ON s.node_id = entry.id
		INNER JOIN vpn_nodes exit ON s.next_hop_node_id = exit.id
		WHERE s.user_id = ? AND s.is_multihop = true
		GROUP BY entry.country, exit.country
		ORDER BY count DESC
		LIMIT 5
	`, userID).Scan(&routeStats)

	stats["popular_routes"] = routeStats

	return stats, nil
}

// EstimateSpeedReduction calculates expected speed reduction for multi-hop
func (m *MultiHopManager) EstimateSpeedReduction(hops int) float64 {
	// Each additional hop reduces speed by ~20-30%
	switch hops {
	case 2:
		return 0.70 // 70% of original speed
	case 3:
		return 0.50 // 50% of original speed
	case 4:
		return 0.35 // 35% of original speed
	default:
		return 0.70
	}
}

// GetRecommendedMultiHopRoute recommends the best multi-hop route based on user preferences
func (m *MultiHopManager) GetRecommendedMultiHopRoute(userID uuid.UUID, preferences MultiHopPreferences) (*HopChain, error) {
	// Priority: Privacy > Performance > Cost
	switch preferences.Priority {
	case "privacy":
		// Use nodes in privacy-friendly jurisdictions
		return m.CreateDoubleVPNChain(userID, "CH", "IS") // Switzerland -> Iceland
	case "performance":
		// Use nodes with lowest latency
		return m.GetOptimalMultiHopRoute(userID, preferences.ExitCountry)
	case "streaming":
		// Use nodes optimized for streaming
		return m.CreateDoubleVPNChain(userID, preferences.EntryCountry, preferences.ExitCountry)
	default:
		return m.CreateDoubleVPNChain(userID, preferences.EntryCountry, preferences.ExitCountry)
	}
}

// MultiHopPreferences defines user preferences for multi-hop routing
type MultiHopPreferences struct {
	Priority     string // privacy, performance, streaming
	EntryCountry string
	ExitCountry  string
	MaxLatency   int
	MinSpeed     float64 // Mbps
}
