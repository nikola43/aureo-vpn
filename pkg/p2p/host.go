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
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/multiformats/go-multiaddr"
)

// Host manages the libp2p node
type Host struct {
	host       host.Host
	dht        *dht.IpfsDHT
	pubsub     *pubsub.PubSub
	registry   *Registry
	config     Config
	localNode  *NodeInfo
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.RWMutex

	// PubSub topics
	announceTopic  *pubsub.Topic
	heartbeatTopic *pubsub.Topic
	announceSub    *pubsub.Subscription
	heartbeatSub   *pubsub.Subscription

	// Callbacks
	onNodeJoined  func(*NodeInfo)
	onNodeLeft    func(uuid.UUID)
	onNodeUpdated func(*NodeInfo)
}

// NewHost creates a new P2P host
func NewHost(config Config) (*Host, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Generate or load identity key
	var privKey crypto.PrivKey
	if config.PrivateKey != "" {
		keyBytes, err := base64.StdEncoding.DecodeString(config.PrivateKey)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("invalid private key: %w", err)
		}
		privKey, err = crypto.UnmarshalPrivateKey(keyBytes)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to unmarshal private key: %w", err)
		}
	} else {
		var err error
		privKey, _, err = crypto.GenerateEd25519Key(rand.Reader)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to generate key: %w", err)
		}
		keyBytes, _ := crypto.MarshalPrivateKey(privKey)
		config.PrivateKey = base64.StdEncoding.EncodeToString(keyBytes)
	}

	// Build listen addresses
	listenAddrs := make([]multiaddr.Multiaddr, 0, len(config.ListenAddrs))
	for _, addr := range config.ListenAddrs {
		ma, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			log.Printf("[P2P] Invalid listen address %s: %v", addr, err)
			continue
		}
		listenAddrs = append(listenAddrs, ma)
	}

	// Create libp2p host
	opts := []libp2p.Option{
		libp2p.Identity(privKey),
		libp2p.ListenAddrs(listenAddrs...),
		libp2p.EnableNATService(),
		libp2p.EnableRelay(),
		libp2p.EnableHolePunching(),
		libp2p.NATPortMap(),
	}

	// Add announce addresses if specified
	if len(config.AnnounceAddrs) > 0 {
		announceAddrs := make([]multiaddr.Multiaddr, 0, len(config.AnnounceAddrs))
		for _, addr := range config.AnnounceAddrs {
			ma, err := multiaddr.NewMultiaddr(addr)
			if err != nil {
				continue
			}
			announceAddrs = append(announceAddrs, ma)
		}
		if len(announceAddrs) > 0 {
			opts = append(opts, libp2p.AddrsFactory(func([]multiaddr.Multiaddr) []multiaddr.Multiaddr {
				return announceAddrs
			}))
		}
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	log.Printf("[P2P] Host created with ID: %s", h.ID())
	for _, addr := range h.Addrs() {
		log.Printf("[P2P] Listening on: %s/p2p/%s", addr, h.ID())
	}

	ph := &Host{
		host:     h,
		registry: NewRegistry(config),
		config:   config,
		ctx:      ctx,
		cancel:   cancel,
	}

	// Setup DHT
	if config.EnableDHT {
		if err := ph.setupDHT(); err != nil {
			h.Close()
			cancel()
			return nil, fmt.Errorf("failed to setup DHT: %w", err)
		}
	}

	// Setup PubSub
	if config.EnablePubSub {
		if err := ph.setupPubSub(); err != nil {
			h.Close()
			cancel()
			return nil, fmt.Errorf("failed to setup PubSub: %w", err)
		}
	}

	// Setup mDNS for local discovery
	if config.EnableMDNS {
		ph.setupMDNS()
	}

	// Setup protocol handlers
	ph.setupProtocols()

	return ph, nil
}

// setupDHT initializes the Kademlia DHT
func (h *Host) setupDHT() error {
	var mode dht.ModeOpt
	if h.config.DHTServerMode {
		mode = dht.ModeServer
	} else {
		mode = dht.ModeClient
	}

	kademliaDHT, err := dht.New(h.ctx, h.host, dht.Mode(mode))
	if err != nil {
		return err
	}

	if err := kademliaDHT.Bootstrap(h.ctx); err != nil {
		return err
	}

	h.dht = kademliaDHT

	// Connect to bootstrap peers
	go h.connectToBootstrapPeers()

	log.Println("[P2P] DHT initialized")
	return nil
}

// setupPubSub initializes GossipSub
func (h *Host) setupPubSub() error {
	ps, err := pubsub.NewGossipSub(h.ctx, h.host)
	if err != nil {
		return err
	}
	h.pubsub = ps

	// Join announce topic
	announceTopic, err := ps.Join(h.config.AnnounceTopic)
	if err != nil {
		return fmt.Errorf("failed to join announce topic: %w", err)
	}
	h.announceTopic = announceTopic

	announceSub, err := announceTopic.Subscribe()
	if err != nil {
		return fmt.Errorf("failed to subscribe to announce topic: %w", err)
	}
	h.announceSub = announceSub

	// Join heartbeat topic
	heartbeatTopic, err := ps.Join(h.config.HeartbeatTopic)
	if err != nil {
		return fmt.Errorf("failed to join heartbeat topic: %w", err)
	}
	h.heartbeatTopic = heartbeatTopic

	heartbeatSub, err := heartbeatTopic.Subscribe()
	if err != nil {
		return fmt.Errorf("failed to subscribe to heartbeat topic: %w", err)
	}
	h.heartbeatSub = heartbeatSub

	log.Println("[P2P] PubSub initialized")
	return nil
}

// setupMDNS initializes mDNS for local network discovery
func (h *Host) setupMDNS() {
	notifee := &mdnsNotifee{host: h}
	service := mdns.NewMdnsService(h.host, "aureo-vpn", notifee)
	if err := service.Start(); err != nil {
		log.Printf("[P2P] mDNS start failed: %v", err)
		return
	}
	log.Println("[P2P] mDNS initialized")
}

// mdnsNotifee handles mDNS peer discovery
type mdnsNotifee struct {
	host *Host
}

func (n *mdnsNotifee) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID == n.host.host.ID() {
		return
	}
	log.Printf("[P2P] mDNS discovered peer: %s", pi.ID)
	n.host.host.Connect(n.host.ctx, pi)
}

// setupProtocols registers protocol handlers
func (h *Host) setupProtocols() {
	// Node info request handler
	h.host.SetStreamHandler(protocol.ID(ProtocolNodeInfo), h.handleNodeInfoRequest)

	// Node list request handler
	h.host.SetStreamHandler(protocol.ID(ProtocolNodeList), h.handleNodeListRequest)

	// Health check handler
	h.host.SetStreamHandler(protocol.ID(ProtocolHealthCheck), h.handleHealthCheck)
}

// handleNodeInfoRequest responds to node info requests
func (h *Host) handleNodeInfoRequest(s network.Stream) {
	defer s.Close()

	h.mu.RLock()
	node := h.localNode
	h.mu.RUnlock()

	if node == nil {
		return
	}

	data, err := json.Marshal(node)
	if err != nil {
		return
	}

	s.Write(data)
}

// handleNodeListRequest responds with known nodes
func (h *Host) handleNodeListRequest(s network.Stream) {
	defer s.Close()

	nodes := h.registry.GetActiveNodes()
	data, err := json.Marshal(nodes)
	if err != nil {
		return
	}

	s.Write(data)
}

// handleHealthCheck responds to health checks
func (h *Host) handleHealthCheck(s network.Stream) {
	defer s.Close()

	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"peer_id":   h.host.ID().String(),
	}

	data, _ := json.Marshal(response)
	s.Write(data)
}

// connectToBootstrapPeers connects to bootstrap nodes
func (h *Host) connectToBootstrapPeers() {
	for _, addrStr := range h.config.BootstrapPeers {
		addr, err := multiaddr.NewMultiaddr(addrStr)
		if err != nil {
			log.Printf("[P2P] Invalid bootstrap address %s: %v", addrStr, err)
			continue
		}

		peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			log.Printf("[P2P] Invalid bootstrap peer info %s: %v", addrStr, err)
			continue
		}

		go func(pi peer.AddrInfo) {
			if err := h.host.Connect(h.ctx, pi); err != nil {
				log.Printf("[P2P] Failed to connect to bootstrap %s: %v", pi.ID, err)
				return
			}
			log.Printf("[P2P] Connected to bootstrap peer: %s", pi.ID)
		}(*peerInfo)
	}
}

// SetLocalNode sets the local node info
func (h *Host) SetLocalNode(node *NodeInfo) {
	h.mu.Lock()
	defer h.mu.Unlock()

	node.PeerID = h.host.ID()

	// Build multiaddrs
	addrs := make([]string, 0, len(h.host.Addrs()))
	for _, addr := range h.host.Addrs() {
		fullAddr := fmt.Sprintf("%s/p2p/%s", addr, h.host.ID())
		addrs = append(addrs, fullAddr)
	}
	node.Multiaddrs = addrs

	h.localNode = node
	h.registry.AddNode(node)
}

// SetCallbacks sets event callbacks
func (h *Host) SetCallbacks(onJoined, onUpdated func(*NodeInfo), onLeft func(uuid.UUID)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onNodeJoined = onJoined
	h.onNodeUpdated = onUpdated
	h.onNodeLeft = onLeft
}

// Start starts the P2P service
func (h *Host) Start() error {
	log.Println("[P2P] Starting...")

	// Start message handlers
	go h.handleAnnounceMessages()
	go h.handleHeartbeatMessages()

	// Start background tasks
	go h.heartbeatLoop()
	go h.announceLoop()
	go h.discoveryLoop()
	go h.cleanupLoop()

	log.Println("[P2P] Started successfully")
	return nil
}

// Stop stops the P2P service
func (h *Host) Stop() error {
	log.Println("[P2P] Stopping...")
	h.cancel()

	// Persist registry
	h.registry.Persist()

	// Close subscriptions
	if h.announceSub != nil {
		h.announceSub.Cancel()
	}
	if h.heartbeatSub != nil {
		h.heartbeatSub.Cancel()
	}

	// Close topics
	if h.announceTopic != nil {
		h.announceTopic.Close()
	}
	if h.heartbeatTopic != nil {
		h.heartbeatTopic.Close()
	}

	// Close DHT
	if h.dht != nil {
		h.dht.Close()
	}

	// Close host
	return h.host.Close()
}

// handleAnnounceMessages processes announce messages from PubSub
func (h *Host) handleAnnounceMessages() {
	for {
		msg, err := h.announceSub.Next(h.ctx)
		if err != nil {
			if h.ctx.Err() != nil {
				return
			}
			continue
		}

		// Skip own messages
		if msg.ReceivedFrom == h.host.ID() {
			continue
		}

		var announce AnnounceMessage
		if err := json.Unmarshal(msg.Data, &announce); err != nil {
			continue
		}

		// Add/update node
		isNew := h.registry.AddNode(&announce.Node)

		// Trigger callback
		h.mu.RLock()
		if isNew && h.onNodeJoined != nil {
			go h.onNodeJoined(&announce.Node)
		} else if !isNew && h.onNodeUpdated != nil {
			go h.onNodeUpdated(&announce.Node)
		}
		h.mu.RUnlock()

		log.Printf("[P2P] Received announce from %s (%s)", announce.Node.Name, announce.Node.ID)
	}
}

// handleHeartbeatMessages processes heartbeat messages from PubSub
func (h *Host) handleHeartbeatMessages() {
	for {
		msg, err := h.heartbeatSub.Next(h.ctx)
		if err != nil {
			if h.ctx.Err() != nil {
				return
			}
			continue
		}

		// Skip own messages
		if msg.ReceivedFrom == h.host.ID() {
			continue
		}

		var heartbeat HeartbeatMessage
		if err := json.Unmarshal(msg.Data, &heartbeat); err != nil {
			continue
		}

		// Update node status
		h.registry.UpdateNodeStatus(
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

// heartbeatLoop sends periodic heartbeats
func (h *Host) heartbeatLoop() {
	ticker := time.NewTicker(h.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			h.broadcastHeartbeat()
		}
	}
}

// announceLoop sends periodic announcements
func (h *Host) announceLoop() {
	// Send initial announcement
	time.Sleep(2 * time.Second) // Wait for connections
	h.broadcastAnnounce()

	ticker := time.NewTicker(h.config.AnnounceInterval)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			h.broadcastAnnounce()
		}
	}
}

// discoveryLoop uses DHT to find peers
func (h *Host) discoveryLoop() {
	if h.dht == nil {
		return
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Initial discovery
	h.discoverPeers()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			h.discoverPeers()
		}
	}
}

// discoverPeers finds peers using DHT
func (h *Host) discoverPeers() {
	routingDiscovery := drouting.NewRoutingDiscovery(h.dht)

	// Advertise ourselves
	routingDiscovery.Advertise(h.ctx, "aureo-vpn")

	// Find peers
	peerChan, err := routingDiscovery.FindPeers(h.ctx, "aureo-vpn")
	if err != nil {
		return
	}

	for peer := range peerChan {
		if peer.ID == h.host.ID() {
			continue
		}

		if h.host.Network().Connectedness(peer.ID) != network.Connected {
			if err := h.host.Connect(h.ctx, peer); err != nil {
				continue
			}
			log.Printf("[P2P] DHT discovered and connected to: %s", peer.ID)

			// Request node info
			go h.requestNodeInfo(peer.ID)
		}
	}
}

// requestNodeInfo requests node info from a peer
func (h *Host) requestNodeInfo(peerID peer.ID) {
	s, err := h.host.NewStream(h.ctx, peerID, protocol.ID(ProtocolNodeInfo))
	if err != nil {
		return
	}
	defer s.Close()

	// Read response
	data, err := io.ReadAll(io.LimitReader(s, 64*1024))
	if err != nil {
		return
	}

	var node NodeInfo
	if err := json.Unmarshal(data, &node); err != nil {
		return
	}

	isNew := h.registry.AddNode(&node)

	h.mu.RLock()
	if isNew && h.onNodeJoined != nil {
		go h.onNodeJoined(&node)
	}
	h.mu.RUnlock()
}

// cleanupLoop cleans up stale entries
func (h *Host) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			h.registry.MarkOffline()
			removed := h.registry.CleanupStale()
			if removed > 0 {
				log.Printf("[P2P] Cleaned up %d stale nodes", removed)
			}
			h.registry.Persist()
		}
	}
}

// broadcastAnnounce publishes node announcement
func (h *Host) broadcastAnnounce() {
	h.mu.RLock()
	node := h.localNode
	h.mu.RUnlock()

	if node == nil || h.announceTopic == nil {
		return
	}

	node.LastHeartbeat = time.Now()

	msg := AnnounceMessage{
		Node:      *node,
		Timestamp: time.Now().UnixNano(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	if err := h.announceTopic.Publish(h.ctx, data); err != nil {
		log.Printf("[P2P] Failed to broadcast announce: %v", err)
	}
}

// broadcastHeartbeat publishes heartbeat
func (h *Host) broadcastHeartbeat() {
	h.mu.RLock()
	node := h.localNode
	h.mu.RUnlock()

	if node == nil || h.heartbeatTopic == nil {
		return
	}

	msg := HeartbeatMessage{
		NodeID:             node.ID,
		PeerID:             h.host.ID().String(),
		Status:             node.Status,
		CurrentConnections: node.CurrentConnections,
		LoadScore:          node.LoadScore,
		CPUUsage:           node.CPUUsage,
		MemoryUsage:        node.MemoryUsage,
		BandwidthGbps:      node.BandwidthGbps,
		Timestamp:          time.Now().UnixNano(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	if err := h.heartbeatTopic.Publish(h.ctx, data); err != nil {
		log.Printf("[P2P] Failed to broadcast heartbeat: %v", err)
	}
}

// UpdateLocalNode updates local node info
func (h *Host) UpdateLocalNode(status string, connections int, load, cpu, memory, bandwidth float64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.localNode == nil {
		return
	}

	h.localNode.Status = status
	h.localNode.CurrentConnections = connections
	h.localNode.LoadScore = load
	h.localNode.CPUUsage = cpu
	h.localNode.MemoryUsage = memory
	h.localNode.BandwidthGbps = bandwidth
	h.localNode.LastHeartbeat = time.Now()
}

// GetRegistry returns the node registry
func (h *Host) GetRegistry() *Registry {
	return h.registry
}

// GetPeerID returns this host's peer ID
func (h *Host) GetPeerID() peer.ID {
	return h.host.ID()
}

// GetMultiaddrs returns this host's addresses
func (h *Host) GetMultiaddrs() []string {
	addrs := make([]string, 0, len(h.host.Addrs()))
	for _, addr := range h.host.Addrs() {
		fullAddr := fmt.Sprintf("%s/p2p/%s", addr, h.host.ID())
		addrs = append(addrs, fullAddr)
	}
	return addrs
}

// ConnectedPeers returns count of connected peers
func (h *Host) ConnectedPeers() int {
	return len(h.host.Network().Peers())
}

// AddPeer manually adds a peer to connect to
func (h *Host) AddPeer(addrStr string) error {
	addr, err := multiaddr.NewMultiaddr(addrStr)
	if err != nil {
		return err
	}

	peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return err
	}

	return h.host.Connect(h.ctx, *peerInfo)
}

// QueryNodes returns nodes matching criteria
func (h *Host) QueryNodes(protocol, countryCode string) []*NodeInfo {
	return h.registry.QueryNodes(protocol, countryCode)
}

// GetBestNode returns the best available node
func (h *Host) GetBestNode(protocol, countryCode string) *NodeInfo {
	return h.registry.GetBestNode(protocol, countryCode)
}

// GetConfig returns the host configuration
func (h *Host) GetConfig() Config {
	return h.config
}
