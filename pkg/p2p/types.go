package p2p

import (
	"time"

	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/peer"
)

// NodeInfo represents a VPN node in the P2P network
type NodeInfo struct {
	// Identity
	ID       uuid.UUID `json:"id"`
	PeerID   peer.ID   `json:"peer_id"`
	Name     string    `json:"name"`
	Version  string    `json:"version"`

	// Network
	PublicIP      string   `json:"public_ip"`
	Multiaddrs    []string `json:"multiaddrs"` // libp2p multiaddresses
	WireGuardPort int      `json:"wireguard_port"`
	OpenVPNPort   int      `json:"openvpn_port"`
	PublicKey     string   `json:"public_key"` // WireGuard public key

	// Location
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`

	// Capacity & Load
	MaxConnections     int     `json:"max_connections"`
	CurrentConnections int     `json:"current_connections"`
	LoadScore          float64 `json:"load_score"`
	CPUUsage           float64 `json:"cpu_usage"`
	MemoryUsage        float64 `json:"memory_usage"`
	BandwidthGbps      float64 `json:"bandwidth_gbps"`

	// Status
	Status        string    `json:"status"` // online, offline, maintenance
	LastHeartbeat time.Time `json:"last_heartbeat"`
	Uptime        float64   `json:"uptime_percentage"`

	// Features
	SupportsWireGuard   bool `json:"supports_wireguard"`
	SupportsOpenVPN     bool `json:"supports_openvpn"`
	SupportsMultiHop    bool `json:"supports_multihop"`
	SupportsObfuscation bool `json:"supports_obfuscation"`

	// Operator (for decentralized nodes)
	OperatorID      *uuid.UUID `json:"operator_id,omitempty"`
	IsOperatorOwned bool       `json:"is_operator_owned"`
	Reputation      float64    `json:"reputation"`
	StakeAmount     float64    `json:"stake_amount,omitempty"`

	// Signature for authenticity
	Signature []byte `json:"signature,omitempty"`
	SignedAt  int64  `json:"signed_at,omitempty"`
}

// HeartbeatMessage is broadcast periodically
type HeartbeatMessage struct {
	NodeID             uuid.UUID `json:"node_id"`
	PeerID             string    `json:"peer_id"`
	Status             string    `json:"status"`
	CurrentConnections int       `json:"current_connections"`
	LoadScore          float64   `json:"load_score"`
	CPUUsage           float64   `json:"cpu_usage"`
	MemoryUsage        float64   `json:"memory_usage"`
	BandwidthGbps      float64   `json:"bandwidth_gbps"`
	Timestamp          int64     `json:"timestamp"`
	Signature          []byte    `json:"signature,omitempty"`
}

// AnnounceMessage is sent when a node joins or updates its info
type AnnounceMessage struct {
	Node      NodeInfo `json:"node"`
	Timestamp int64    `json:"timestamp"`
}

// LeaveMessage is sent when a node leaves gracefully
type LeaveMessage struct {
	NodeID    uuid.UUID `json:"node_id"`
	PeerID    string    `json:"peer_id"`
	Timestamp int64     `json:"timestamp"`
	Signature []byte    `json:"signature,omitempty"`
}

// Config for P2P service
type Config struct {
	// Identity
	NodeID     uuid.UUID `json:"node_id"`
	PrivateKey string    `json:"private_key"` // libp2p identity key (base64)

	// Network
	ListenAddrs    []string `json:"listen_addrs"`    // e.g., "/ip4/0.0.0.0/tcp/4001"
	AnnounceAddrs  []string `json:"announce_addrs"`  // Public addresses to announce
	BootstrapPeers []string `json:"bootstrap_peers"` // Bootstrap peer multiaddrs

	// DHT
	EnableDHT       bool `json:"enable_dht"`
	DHTServerMode   bool `json:"dht_server_mode"` // Act as DHT server
	EnableMDNS      bool `json:"enable_mdns"`     // Local discovery

	// PubSub
	EnablePubSub bool `json:"enable_pubsub"`

	// Timing
	HeartbeatInterval time.Duration `json:"heartbeat_interval"`
	AnnounceInterval  time.Duration `json:"announce_interval"`
	NodeTimeout       time.Duration `json:"node_timeout"`

	// Limits
	MaxPeers int `json:"max_peers"`
	MaxNodes int `json:"max_nodes"`

	// Topics
	AnnounceTopic  string `json:"announce_topic"`
	HeartbeatTopic string `json:"heartbeat_topic"`

	// Data persistence
	DataDir string `json:"data_dir"`
}

// DefaultConfig returns default P2P configuration
func DefaultConfig() Config {
	return Config{
		ListenAddrs: []string{
			"/ip4/0.0.0.0/tcp/4001",
			"/ip4/0.0.0.0/udp/4001/quic-v1",
		},
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
		DataDir:           "./data/p2p",
	}
}

// Well-known bootstrap peers (can be extended)
var DefaultBootstrapPeers = []string{
	// Add your bootstrap nodes here
	// "/ip4/1.2.3.4/tcp/4001/p2p/QmBootstrap..."
}

// Protocol IDs
const (
	ProtocolNodeInfo    = "/aureo/nodeinfo/1.0.0"
	ProtocolNodeList    = "/aureo/nodelist/1.0.0"
	ProtocolHealthCheck = "/aureo/health/1.0.0"
)
