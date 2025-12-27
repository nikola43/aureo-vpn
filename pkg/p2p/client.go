package p2p

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/multiformats/go-multiaddr"
)

// Client is a lightweight P2P client for querying the network
// Used by API Gateway and other services that don't run VPN nodes
type Client struct {
	host       host.Host
	dht        *dht.IpfsDHT
	pubsub     *pubsub.PubSub
	registry   *Registry
	config     Config
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.RWMutex

	// PubSub subscriptions
	announceSub  *pubsub.Subscription
	heartbeatSub *pubsub.Subscription
}

// ClientConfig holds configuration for the P2P client
type ClientConfig struct {
	ListenPort     int
	BootstrapPeers []string
	EnableDHT      bool
	EnableMDNS     bool
	DataDir        string
}

// DefaultClientConfig returns default client configuration
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		ListenPort: 4002, // Different from node port
		EnableDHT:  true,
		EnableMDNS: true,
		DataDir:    "./data/p2p-client",
	}
}

// NewClient creates a new P2P client
func NewClient(cfg ClientConfig) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Generate ephemeral identity
	privKey, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	keyBytes, _ := crypto.MarshalPrivateKey(privKey)

	// Build listen addresses
	listenAddrs := []multiaddr.Multiaddr{}
	if cfg.ListenPort > 0 {
		tcpAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", cfg.ListenPort))
		quicAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", cfg.ListenPort))
		listenAddrs = append(listenAddrs, tcpAddr, quicAddr)
	}

	// Create libp2p host
	opts := []libp2p.Option{
		libp2p.Identity(privKey),
		libp2p.ListenAddrs(listenAddrs...),
		libp2p.EnableNATService(),
		libp2p.EnableRelay(),
		libp2p.NATPortMap(),
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	config := Config{
		NodeID:            uuid.New(),
		PrivateKey:        base64.StdEncoding.EncodeToString(keyBytes),
		EnableDHT:         cfg.EnableDHT,
		EnableMDNS:        cfg.EnableMDNS,
		EnablePubSub:      true,
		BootstrapPeers:    cfg.BootstrapPeers,
		NodeTimeout:       2 * time.Minute,
		MaxNodes:          1000,
		AnnounceTopic:     "/aureo/nodes/announce/1.0.0",
		HeartbeatTopic:    "/aureo/nodes/heartbeat/1.0.0",
		DataDir:           cfg.DataDir,
	}

	client := &Client{
		host:     h,
		registry: NewRegistry(config),
		config:   config,
		ctx:      ctx,
		cancel:   cancel,
	}

	log.Printf("[P2P Client] Created with ID: %s", h.ID())

	return client, nil
}

// Start starts the P2P client
func (c *Client) Start() error {
	log.Println("[P2P Client] Starting...")

	// Setup DHT
	if c.config.EnableDHT {
		if err := c.setupDHT(); err != nil {
			return fmt.Errorf("failed to setup DHT: %w", err)
		}
	}

	// Setup PubSub
	if err := c.setupPubSub(); err != nil {
		return fmt.Errorf("failed to setup PubSub: %w", err)
	}

	// Setup mDNS
	if c.config.EnableMDNS {
		c.setupMDNS()
	}

	// Start background tasks
	go c.handleAnnounceMessages()
	go c.handleHeartbeatMessages()
	go c.discoveryLoop()
	go c.cleanupLoop()

	log.Println("[P2P Client] Started successfully")
	return nil
}

// Stop stops the P2P client
func (c *Client) Stop() error {
	log.Println("[P2P Client] Stopping...")
	c.cancel()

	c.registry.Persist()

	if c.announceSub != nil {
		c.announceSub.Cancel()
	}
	if c.heartbeatSub != nil {
		c.heartbeatSub.Cancel()
	}

	if c.dht != nil {
		c.dht.Close()
	}

	return c.host.Close()
}

func (c *Client) setupDHT() error {
	kademliaDHT, err := dht.New(c.ctx, c.host, dht.Mode(dht.ModeClient))
	if err != nil {
		return err
	}

	if err := kademliaDHT.Bootstrap(c.ctx); err != nil {
		return err
	}

	c.dht = kademliaDHT

	// Connect to bootstrap peers
	go c.connectToBootstrapPeers()

	log.Println("[P2P Client] DHT initialized")
	return nil
}

func (c *Client) setupPubSub() error {
	ps, err := pubsub.NewGossipSub(c.ctx, c.host)
	if err != nil {
		return err
	}
	c.pubsub = ps

	// Subscribe to announce topic
	announceTopic, err := ps.Join(c.config.AnnounceTopic)
	if err != nil {
		return err
	}
	announceSub, err := announceTopic.Subscribe()
	if err != nil {
		return err
	}
	c.announceSub = announceSub

	// Subscribe to heartbeat topic
	heartbeatTopic, err := ps.Join(c.config.HeartbeatTopic)
	if err != nil {
		return err
	}
	heartbeatSub, err := heartbeatTopic.Subscribe()
	if err != nil {
		return err
	}
	c.heartbeatSub = heartbeatSub

	log.Println("[P2P Client] PubSub initialized")
	return nil
}

func (c *Client) setupMDNS() {
	notifee := &clientMdnsNotifee{client: c}
	service := mdns.NewMdnsService(c.host, "aureo-vpn", notifee)
	if err := service.Start(); err != nil {
		log.Printf("[P2P Client] mDNS start failed: %v", err)
		return
	}
	log.Println("[P2P Client] mDNS initialized")
}

type clientMdnsNotifee struct {
	client *Client
}

func (n *clientMdnsNotifee) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID == n.client.host.ID() {
		return
	}
	n.client.host.Connect(n.client.ctx, pi)
}

func (c *Client) connectToBootstrapPeers() {
	for _, addrStr := range c.config.BootstrapPeers {
		addr, err := multiaddr.NewMultiaddr(addrStr)
		if err != nil {
			continue
		}

		peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			continue
		}

		go func(pi peer.AddrInfo) {
			if err := c.host.Connect(c.ctx, pi); err != nil {
				log.Printf("[P2P Client] Failed to connect to bootstrap %s: %v", pi.ID, err)
				return
			}
			log.Printf("[P2P Client] Connected to bootstrap: %s", pi.ID)

			// Request node list
			c.requestNodeList(pi.ID)
		}(*peerInfo)
	}
}

func (c *Client) requestNodeList(peerID peer.ID) {
	s, err := c.host.NewStream(c.ctx, peerID, protocol.ID(ProtocolNodeList))
	if err != nil {
		return
	}
	defer s.Close()

	data, err := io.ReadAll(io.LimitReader(s, 256*1024))
	if err != nil {
		return
	}

	var nodes []*NodeInfo
	if err := json.Unmarshal(data, &nodes); err != nil {
		return
	}

	for _, node := range nodes {
		c.registry.AddNode(node)
	}

	log.Printf("[P2P Client] Received %d nodes from %s", len(nodes), peerID)
}

func (c *Client) handleAnnounceMessages() {
	for {
		msg, err := c.announceSub.Next(c.ctx)
		if err != nil {
			if c.ctx.Err() != nil {
				return
			}
			continue
		}

		if msg.ReceivedFrom == c.host.ID() {
			continue
		}

		var announce AnnounceMessage
		if err := json.Unmarshal(msg.Data, &announce); err != nil {
			continue
		}

		c.registry.AddNode(&announce.Node)
	}
}

func (c *Client) handleHeartbeatMessages() {
	for {
		msg, err := c.heartbeatSub.Next(c.ctx)
		if err != nil {
			if c.ctx.Err() != nil {
				return
			}
			continue
		}

		if msg.ReceivedFrom == c.host.ID() {
			continue
		}

		var heartbeat HeartbeatMessage
		if err := json.Unmarshal(msg.Data, &heartbeat); err != nil {
			continue
		}

		c.registry.UpdateNodeStatus(
			heartbeat.NodeID,
			heartbeat.Status,
			heartbeat.CurrentConnections,
			heartbeat.LoadScore,
			heartbeat.CPUUsage,
			heartbeat.MemoryUsage,
			heartbeat.BandwidthGbps,
		)
	}
}

func (c *Client) discoveryLoop() {
	if c.dht == nil {
		return
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Initial discovery
	c.discoverPeers()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.discoverPeers()
		}
	}
}

func (c *Client) discoverPeers() {
	routingDiscovery := drouting.NewRoutingDiscovery(c.dht)

	peerChan, err := routingDiscovery.FindPeers(c.ctx, "aureo-vpn")
	if err != nil {
		return
	}

	for peer := range peerChan {
		if peer.ID == c.host.ID() {
			continue
		}

		if err := c.host.Connect(c.ctx, peer); err != nil {
			continue
		}

		// Request node list from new peer
		go c.requestNodeList(peer.ID)
	}
}

func (c *Client) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.registry.MarkOffline()
			c.registry.CleanupStale()
			c.registry.Persist()
		}
	}
}

// GetNodes returns all active nodes
func (c *Client) GetNodes() []*NodeInfo {
	return c.registry.GetActiveNodes()
}

// QueryNodes returns nodes matching filters
func (c *Client) QueryNodes(protocol, countryCode string) []*NodeInfo {
	return c.registry.QueryNodes(protocol, countryCode)
}

// GetBestNode returns the best available node
func (c *Client) GetBestNode(protocol, countryCode string) *NodeInfo {
	return c.registry.GetBestNode(protocol, countryCode)
}

// GetCountries returns list of countries with active nodes
func (c *Client) GetCountries() []string {
	return c.registry.GetCountries()
}

// GetStats returns client statistics
func (c *Client) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"peer_id":         c.host.ID().String(),
		"connected_peers": len(c.host.Network().Peers()),
		"known_nodes":     c.registry.Count(),
		"active_nodes":    c.registry.ActiveCount(),
		"countries":       c.registry.GetCountries(),
	}
}

// AddBootstrapPeer adds a bootstrap peer and connects
func (c *Client) AddBootstrapPeer(addrStr string) error {
	addr, err := multiaddr.NewMultiaddr(addrStr)
	if err != nil {
		return err
	}

	peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return err
	}

	if err := c.host.Connect(c.ctx, *peerInfo); err != nil {
		return err
	}

	c.requestNodeList(peerInfo.ID)
	return nil
}

// GetRegistry returns the node registry
func (c *Client) GetRegistry() *Registry {
	return c.registry
}
