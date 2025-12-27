package p2p

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nikola43/aureo-vpn/pkg/models"
)

// NodeInfoFromModel converts a VPNNode model to P2P NodeInfo
func NodeInfoFromModel(node *models.VPNNode, gossipPort int) *NodeInfo {
	info := &NodeInfo{
		ID:                  node.ID,
		Name:                node.Name,
		Version:             node.Version,
		PublicIP:            node.PublicIP,
		WireGuardPort:       node.WireGuardPort,
		OpenVPNPort:         node.OpenVPNPort,
		PublicKey:           node.PublicKey,
		Country:             node.Country,
		CountryCode:         node.CountryCode,
		City:                node.City,
		Latitude:            node.Latitude,
		Longitude:           node.Longitude,
		MaxConnections:      node.MaxConnections,
		CurrentConnections:  node.CurrentConnections,
		LoadScore:           node.LoadScore,
		CPUUsage:            node.CPUUsage,
		MemoryUsage:         node.MemoryUsage,
		BandwidthGbps:       node.BandwidthUsageGbps,
		Status:              node.Status,
		LastHeartbeat:       node.LastHeartbeat,
		Uptime:              node.UptimePercentage,
		SupportsWireGuard:   node.SupportsWireGuard,
		SupportsOpenVPN:     node.SupportsOpenVPN,
		SupportsMultiHop:    node.SupportsMultiHop,
		SupportsObfuscation: node.SupportsObfuscation,
		OperatorID:          node.OperatorID,
		IsOperatorOwned:     node.IsOperatorOwned,
	}

	return info
}

// NodeInfoToModel converts P2P NodeInfo to a VPNNode model (for caching discovered nodes)
func NodeInfoToModel(info *NodeInfo) *models.VPNNode {
	node := &models.VPNNode{
		ID:                  info.ID,
		Name:                info.Name,
		Version:             info.Version,
		PublicIP:            info.PublicIP,
		WireGuardPort:       info.WireGuardPort,
		OpenVPNPort:         info.OpenVPNPort,
		PublicKey:           info.PublicKey,
		Country:             info.Country,
		CountryCode:         info.CountryCode,
		City:                info.City,
		Latitude:            info.Latitude,
		Longitude:           info.Longitude,
		MaxConnections:      info.MaxConnections,
		CurrentConnections:  info.CurrentConnections,
		LoadScore:           info.LoadScore,
		CPUUsage:            info.CPUUsage,
		MemoryUsage:         info.MemoryUsage,
		BandwidthUsageGbps:  info.BandwidthGbps,
		Status:              info.Status,
		LastHeartbeat:       info.LastHeartbeat,
		UptimePercentage:    info.Uptime,
		SupportsWireGuard:   info.SupportsWireGuard,
		SupportsOpenVPN:     info.SupportsOpenVPN,
		SupportsMultiHop:    info.SupportsMultiHop,
		SupportsObfuscation: info.SupportsObfuscation,
		OperatorID:          info.OperatorID,
		IsOperatorOwned:     info.IsOperatorOwned,
	}

	return node
}

// ServiceConfig holds P2P service configuration for VPN node integration
type ServiceConfig struct {
	// Node identity
	NodeID uuid.UUID

	// P2P Network
	ListenPort     int      // Default: 4001
	BootstrapPeers []string // Initial peers to connect to
	AnnounceAddrs  []string // Public addresses to announce

	// Features
	EnableDHT     bool // Use DHT for discovery
	EnableMDNS    bool // Use mDNS for local discovery
	DHTServerMode bool // Act as DHT server

	// Timing
	HeartbeatInterval time.Duration
	AnnounceInterval  time.Duration

	// Storage
	DataDir string
}

// DefaultServiceConfig returns default service configuration
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		ListenPort:        4001,
		EnableDHT:         true,
		EnableMDNS:        true,
		DHTServerMode:     true,
		HeartbeatInterval: 30 * time.Second,
		AnnounceInterval:  5 * time.Minute,
		DataDir:           "./data/p2p",
	}
}

// BuildConfig creates a p2p.Config from ServiceConfig
func (sc ServiceConfig) BuildConfig() Config {
	listenAddrs := []string{
		fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", sc.ListenPort),
		fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", sc.ListenPort),
	}

	return Config{
		NodeID:            sc.NodeID,
		ListenAddrs:       listenAddrs,
		AnnounceAddrs:     sc.AnnounceAddrs,
		BootstrapPeers:    sc.BootstrapPeers,
		EnableDHT:         sc.EnableDHT,
		DHTServerMode:     sc.DHTServerMode,
		EnableMDNS:        sc.EnableMDNS,
		EnablePubSub:      true,
		HeartbeatInterval: sc.HeartbeatInterval,
		AnnounceInterval:  sc.AnnounceInterval,
		NodeTimeout:       2 * time.Minute,
		MaxPeers:          100,
		MaxNodes:          1000,
		AnnounceTopic:     "/aureo/nodes/announce/1.0.0",
		HeartbeatTopic:    "/aureo/nodes/heartbeat/1.0.0",
		DataDir:           sc.DataDir,
	}
}
