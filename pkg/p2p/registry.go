package p2p

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/peer"
)

// Registry stores discovered nodes
type Registry struct {
	mu       sync.RWMutex
	nodes    map[uuid.UUID]*NodeInfo
	byPeerID map[peer.ID]uuid.UUID
	config   Config
}

// NewRegistry creates a new node registry
func NewRegistry(config Config) *Registry {
	r := &Registry{
		nodes:    make(map[uuid.UUID]*NodeInfo),
		byPeerID: make(map[peer.ID]uuid.UUID),
		config:   config,
	}

	// Load persisted data
	r.load()

	return r
}

// AddNode adds or updates a node
func (r *Registry) AddNode(node *NodeInfo) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	isNew := false

	// Check if node exists and if update is newer
	if existing, ok := r.nodes[node.ID]; ok {
		if node.LastHeartbeat.Before(existing.LastHeartbeat) {
			return false // Ignore older data
		}
	} else {
		isNew = true
	}

	// Enforce limit
	if isNew && len(r.nodes) >= r.config.MaxNodes {
		r.evictOldestNode()
	}

	r.nodes[node.ID] = node
	r.byPeerID[node.PeerID] = node.ID

	return isNew
}

// RemoveNode removes a node
func (r *Registry) RemoveNode(nodeID uuid.UUID) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if node, ok := r.nodes[nodeID]; ok {
		delete(r.byPeerID, node.PeerID)
		delete(r.nodes, nodeID)
	}
}

// GetNode returns a node by UUID
func (r *Registry) GetNode(nodeID uuid.UUID) (*NodeInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	node, ok := r.nodes[nodeID]
	if ok {
		copy := *node
		return &copy, true
	}
	return nil, false
}

// GetNodeByPeerID returns a node by libp2p peer ID
func (r *Registry) GetNodeByPeerID(peerID peer.ID) (*NodeInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if nodeID, ok := r.byPeerID[peerID]; ok {
		if node, ok := r.nodes[nodeID]; ok {
			copy := *node
			return &copy, true
		}
	}
	return nil, false
}

// GetAllNodes returns all nodes
func (r *Registry) GetAllNodes() []*NodeInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := make([]*NodeInfo, 0, len(r.nodes))
	for _, node := range r.nodes {
		copy := *node
		nodes = append(nodes, &copy)
	}
	return nodes
}

// GetActiveNodes returns online nodes with recent heartbeat
func (r *Registry) GetActiveNodes() []*NodeInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := make([]*NodeInfo, 0)
	timeout := r.config.NodeTimeout

	for _, node := range r.nodes {
		if node.Status == "online" && time.Since(node.LastHeartbeat) < timeout {
			copy := *node
			nodes = append(nodes, &copy)
		}
	}

	// Sort by load score
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].LoadScore < nodes[j].LoadScore
	})

	return nodes
}

// QueryNodes returns nodes matching filters
func (r *Registry) QueryNodes(protocol, countryCode string) []*NodeInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := make([]*NodeInfo, 0)
	timeout := r.config.NodeTimeout

	for _, node := range r.nodes {
		// Check if online
		if node.Status != "online" || time.Since(node.LastHeartbeat) > timeout {
			continue
		}

		// Filter by protocol
		if protocol != "" {
			if protocol == "wireguard" && !node.SupportsWireGuard {
				continue
			}
			if protocol == "openvpn" && !node.SupportsOpenVPN {
				continue
			}
		}

		// Filter by country
		if countryCode != "" && node.CountryCode != countryCode {
			continue
		}

		copy := *node
		nodes = append(nodes, &copy)
	}

	// Sort by load score
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].LoadScore < nodes[j].LoadScore
	})

	return nodes
}

// GetBestNode returns the best available node
func (r *Registry) GetBestNode(protocol, countryCode string) *NodeInfo {
	nodes := r.QueryNodes(protocol, countryCode)
	if len(nodes) == 0 {
		return nil
	}
	return nodes[0] // Already sorted by load
}

// UpdateNodeStatus updates a node's status from heartbeat
func (r *Registry) UpdateNodeStatus(nodeID uuid.UUID, status string, connections int, loadScore, cpu, memory, bandwidth float64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if node, ok := r.nodes[nodeID]; ok {
		node.Status = status
		node.CurrentConnections = connections
		node.LoadScore = loadScore
		node.CPUUsage = cpu
		node.MemoryUsage = memory
		node.BandwidthGbps = bandwidth
		node.LastHeartbeat = time.Now()
	}
}

// CleanupStale removes nodes that haven't sent heartbeat
func (r *Registry) CleanupStale() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	timeout := r.config.NodeTimeout * 3 // Give extra time before removal
	removed := 0

	for id, node := range r.nodes {
		if time.Since(node.LastHeartbeat) > timeout {
			delete(r.byPeerID, node.PeerID)
			delete(r.nodes, id)
			removed++
		}
	}

	return removed
}

// MarkOffline marks nodes as offline if heartbeat is stale
func (r *Registry) MarkOffline() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	timeout := r.config.NodeTimeout
	count := 0

	for _, node := range r.nodes {
		if node.Status == "online" && time.Since(node.LastHeartbeat) > timeout {
			node.Status = "offline"
			count++
		}
	}

	return count
}

// evictOldestNode removes the oldest node
func (r *Registry) evictOldestNode() {
	var oldest *NodeInfo
	for _, node := range r.nodes {
		if oldest == nil || node.LastHeartbeat.Before(oldest.LastHeartbeat) {
			oldest = node
		}
	}
	if oldest != nil {
		delete(r.byPeerID, oldest.PeerID)
		delete(r.nodes, oldest.ID)
	}
}

// Count returns node count
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.nodes)
}

// ActiveCount returns online node count
func (r *Registry) ActiveCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	timeout := r.config.NodeTimeout

	for _, node := range r.nodes {
		if node.Status == "online" && time.Since(node.LastHeartbeat) < timeout {
			count++
		}
	}
	return count
}

// GetCountries returns list of countries with active nodes
func (r *Registry) GetCountries() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	countries := make(map[string]bool)
	timeout := r.config.NodeTimeout

	for _, node := range r.nodes {
		if node.Status == "online" && time.Since(node.LastHeartbeat) < timeout {
			countries[node.CountryCode] = true
		}
	}

	result := make([]string, 0, len(countries))
	for country := range countries {
		result = append(result, country)
	}
	sort.Strings(result)
	return result
}

// Persist saves registry to disk
func (r *Registry) Persist() error {
	if r.config.DataDir == "" {
		return nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Ensure directory exists
	if err := os.MkdirAll(r.config.DataDir, 0755); err != nil {
		return err
	}

	nodes := make([]*NodeInfo, 0, len(r.nodes))
	for _, node := range r.nodes {
		nodes = append(nodes, node)
	}

	data, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(r.config.DataDir, "nodes.json"), data, 0644)
}

// load loads registry from disk
func (r *Registry) load() {
	if r.config.DataDir == "" {
		return
	}

	data, err := os.ReadFile(filepath.Join(r.config.DataDir, "nodes.json"))
	if err != nil {
		return
	}

	var nodes []*NodeInfo
	if err := json.Unmarshal(data, &nodes); err != nil {
		return
	}

	for _, node := range nodes {
		r.nodes[node.ID] = node
		r.byPeerID[node.PeerID] = node.ID
	}
}
