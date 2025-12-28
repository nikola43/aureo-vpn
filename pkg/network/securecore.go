// Package network provides Secure Core routing
// Like ProtonVPN's Secure Core, this routes traffic through high-security
// server infrastructure before exiting to the destination
package network

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"net"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// SecureCoreConfig configures Secure Core routing
type SecureCoreConfig struct {
	// Enable multi-hop through secure core servers
	Enabled bool

	// Minimum number of hops
	MinHops int

	// Maximum number of hops
	MaxHops int

	// Prefer servers in privacy-friendly jurisdictions
	PreferPrivacyJurisdictions bool

	// Server selection strategy
	SelectionStrategy ServerSelection

	// Circuit rotation interval
	CircuitRotation time.Duration

	// Fallback to direct if secure core unavailable
	AllowFallback bool
}

// ServerSelection defines how servers are selected
type ServerSelection int

const (
	SelectionLowestLatency ServerSelection = iota
	SelectionRandom
	SelectionGeoDiversity
	SelectionLoadBalanced
)

// DefaultSecureCoreConfig returns secure defaults
func DefaultSecureCoreConfig() *SecureCoreConfig {
	return &SecureCoreConfig{
		Enabled:                    true,
		MinHops:                    2,
		MaxHops:                    3,
		PreferPrivacyJurisdictions: true,
		SelectionStrategy:          SelectionGeoDiversity,
		CircuitRotation:            30 * time.Minute,
		AllowFallback:              false,
	}
}

// Jurisdiction represents a server's legal jurisdiction
type Jurisdiction struct {
	Country     string
	Region      string
	PrivacyScore int // 1-10, higher is better
	EyesAlliance bool // Five/Nine/Fourteen Eyes
}

// PrivacyJurisdictions are countries with strong privacy laws
var PrivacyJurisdictions = map[string]Jurisdiction{
	"CH": {Country: "Switzerland", PrivacyScore: 10, EyesAlliance: false},
	"IS": {Country: "Iceland", PrivacyScore: 9, EyesAlliance: false},
	"SE": {Country: "Sweden", PrivacyScore: 8, EyesAlliance: true}, // 14 eyes but good laws
	"PA": {Country: "Panama", PrivacyScore: 9, EyesAlliance: false},
	"RO": {Country: "Romania", PrivacyScore: 8, EyesAlliance: false},
	"MY": {Country: "Malaysia", PrivacyScore: 7, EyesAlliance: false},
	"SG": {Country: "Singapore", PrivacyScore: 7, EyesAlliance: false},
}

// CoreServer represents a secure core server
type CoreServer struct {
	ID           string
	Name         string
	Address      string
	Port         int
	Country      string
	City         string
	Jurisdiction Jurisdiction
	Load         float64
	Latency      time.Duration
	Bandwidth    uint64
	IsCore       bool // Is a secure core server
	IsEntry      bool // Can be an entry point
	IsExit       bool // Can be an exit point
	PublicKey    []byte
	Available    bool
}

// Circuit represents a multi-hop routing path
type Circuit struct {
	ID        [16]byte
	Nodes     []*CoreServer
	CreatedAt time.Time
	ExpiresAt time.Time
	Active    bool
	Stats     CircuitStats
}

// CircuitStats tracks circuit statistics
type CircuitStats struct {
	BytesSent     atomic.Uint64
	BytesReceived atomic.Uint64
	PacketsSent   atomic.Uint64
	PacketsRecv   atomic.Uint64
	TotalLatency  time.Duration
}

// SecureCore manages secure core routing
type SecureCore struct {
	config *SecureCoreConfig

	// Available servers
	servers     []*CoreServer
	coreServers []*CoreServer

	// Active circuits
	circuits   []*Circuit
	circuitsMu sync.RWMutex

	// Current active circuit
	activeCircuit *Circuit

	// Statistics
	stats SecureCoreStats

	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
}

// SecureCoreStats tracks secure core statistics
type SecureCoreStats struct {
	CircuitsCreated   atomic.Uint64
	CircuitsRotated   atomic.Uint64
	CurrentHops       int
	BytesRouted       atomic.Uint64
	AverageLatency    time.Duration
}

// NewSecureCore creates a secure core router
func NewSecureCore(config *SecureCoreConfig) *SecureCore {
	if config == nil {
		config = DefaultSecureCoreConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	sc := &SecureCore{
		config:      config,
		servers:     make([]*CoreServer, 0),
		coreServers: make([]*CoreServer, 0),
		circuits:    make([]*Circuit, 0),
		ctx:         ctx,
		cancel:      cancel,
	}

	if config.CircuitRotation > 0 {
		go sc.circuitRotationLoop()
	}

	return sc
}

// AddServer adds a server to the pool
func (sc *SecureCore) AddServer(server *CoreServer) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.servers = append(sc.servers, server)

	if server.IsCore {
		sc.coreServers = append(sc.coreServers, server)
	}
}

// BuildCircuit creates a new secure core circuit
func (sc *SecureCore) BuildCircuit(exitCountry string) (*Circuit, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if len(sc.coreServers) == 0 {
		if sc.config.AllowFallback {
			return sc.buildDirectCircuit(exitCountry)
		}
		return nil, errors.New("no secure core servers available")
	}

	// Select nodes for circuit
	nodes, err := sc.selectNodes(exitCountry)
	if err != nil {
		return nil, err
	}

	circuit := &Circuit{
		Nodes:     nodes,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(sc.config.CircuitRotation),
		Active:    true,
	}

	rand.Read(circuit.ID[:])

	sc.circuitsMu.Lock()
	sc.circuits = append(sc.circuits, circuit)
	sc.circuitsMu.Unlock()

	sc.stats.CircuitsCreated.Add(1)
	sc.stats.CurrentHops = len(nodes)

	return circuit, nil
}

// selectNodes selects servers for the circuit based on strategy
func (sc *SecureCore) selectNodes(exitCountry string) ([]*CoreServer, error) {
	// Get available servers
	available := make([]*CoreServer, 0)
	for _, s := range sc.servers {
		if s.Available {
			available = append(available, s)
		}
	}

	if len(available) < sc.config.MinHops {
		return nil, errors.New("not enough servers available")
	}

	var selected []*CoreServer

	switch sc.config.SelectionStrategy {
	case SelectionLowestLatency:
		selected = sc.selectByLatency(available, exitCountry)
	case SelectionRandom:
		selected = sc.selectRandom(available, exitCountry)
	case SelectionGeoDiversity:
		selected = sc.selectByGeoDiversity(available, exitCountry)
	case SelectionLoadBalanced:
		selected = sc.selectByLoad(available, exitCountry)
	default:
		selected = sc.selectByGeoDiversity(available, exitCountry)
	}

	return selected, nil
}

// selectByLatency selects servers with lowest latency
func (sc *SecureCore) selectByLatency(servers []*CoreServer, exitCountry string) []*CoreServer {
	// Sort by latency
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Latency < servers[j].Latency
	})

	return sc.pickNodes(servers, exitCountry)
}

// selectRandom randomly selects servers
func (sc *SecureCore) selectRandom(servers []*CoreServer, exitCountry string) []*CoreServer {
	// Fisher-Yates shuffle
	for i := len(servers) - 1; i > 0; i-- {
		randBytes := make([]byte, 4)
		rand.Read(randBytes)
		j := int(binary.BigEndian.Uint32(randBytes)) % (i + 1)
		servers[i], servers[j] = servers[j], servers[i]
	}

	return sc.pickNodes(servers, exitCountry)
}

// selectByGeoDiversity selects servers from different countries
func (sc *SecureCore) selectByGeoDiversity(servers []*CoreServer, exitCountry string) []*CoreServer {
	countries := make(map[string][]*CoreServer)
	for _, s := range servers {
		countries[s.Country] = append(countries[s.Country], s)
	}

	selected := make([]*CoreServer, 0, sc.config.MaxHops)

	// Pick one server from each country
	for country, countryServers := range countries {
		if len(selected) >= sc.config.MaxHops {
			break
		}

		// Skip if we already have enough nodes
		if len(selected) >= sc.config.MinHops && country != exitCountry {
			continue
		}

		// Prefer privacy-friendly jurisdictions for core nodes
		if sc.config.PreferPrivacyJurisdictions {
			if _, ok := PrivacyJurisdictions[country]; !ok && len(selected) < sc.config.MinHops-1 {
				continue
			}
		}

		// Pick the lowest latency server from this country
		sort.Slice(countryServers, func(i, j int) bool {
			return countryServers[i].Latency < countryServers[j].Latency
		})

		selected = append(selected, countryServers[0])
	}

	return selected
}

// selectByLoad selects servers with lowest load
func (sc *SecureCore) selectByLoad(servers []*CoreServer, exitCountry string) []*CoreServer {
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Load < servers[j].Load
	})

	return sc.pickNodes(servers, exitCountry)
}

// pickNodes picks the required number of nodes
func (sc *SecureCore) pickNodes(servers []*CoreServer, exitCountry string) []*CoreServer {
	numHops := sc.config.MinHops

	// Add variation
	randBytes := make([]byte, 1)
	rand.Read(randBytes)
	if int(randBytes[0])%3 == 0 && numHops < sc.config.MaxHops {
		numHops++
	}

	if numHops > len(servers) {
		numHops = len(servers)
	}

	selected := make([]*CoreServer, 0, numHops)

	// Pick entry (core server in privacy jurisdiction)
	for _, s := range servers {
		if s.IsCore && s.IsEntry {
			if sc.config.PreferPrivacyJurisdictions {
				if jur, ok := PrivacyJurisdictions[s.Country]; ok && jur.PrivacyScore >= 8 {
					selected = append(selected, s)
					break
				}
			} else {
				selected = append(selected, s)
				break
			}
		}
	}

	// Pick middle nodes
	for _, s := range servers {
		if len(selected) >= numHops-1 {
			break
		}
		if s.IsCore && !containsServer(selected, s) {
			selected = append(selected, s)
		}
	}

	// Pick exit (in requested country if specified)
	for _, s := range servers {
		if s.IsExit && (exitCountry == "" || s.Country == exitCountry) {
			if !containsServer(selected, s) {
				selected = append(selected, s)
				break
			}
		}
	}

	return selected
}

func containsServer(servers []*CoreServer, target *CoreServer) bool {
	for _, s := range servers {
		if s.ID == target.ID {
			return true
		}
	}
	return false
}

// buildDirectCircuit builds a direct (non-secure-core) circuit
func (sc *SecureCore) buildDirectCircuit(exitCountry string) (*Circuit, error) {
	for _, s := range sc.servers {
		if s.IsExit && s.Available && (exitCountry == "" || s.Country == exitCountry) {
			circuit := &Circuit{
				Nodes:     []*CoreServer{s},
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(sc.config.CircuitRotation),
				Active:    true,
			}
			rand.Read(circuit.ID[:])
			return circuit, nil
		}
	}
	return nil, errors.New("no exit servers available")
}

// SetActiveCircuit sets the current active circuit
func (sc *SecureCore) SetActiveCircuit(circuit *Circuit) {
	sc.circuitsMu.Lock()
	defer sc.circuitsMu.Unlock()
	sc.activeCircuit = circuit
}

// GetActiveCircuit returns the current active circuit
func (sc *SecureCore) GetActiveCircuit() *Circuit {
	sc.circuitsMu.RLock()
	defer sc.circuitsMu.RUnlock()
	return sc.activeCircuit
}

// RotateCircuit creates a new circuit and retires the old one
func (sc *SecureCore) RotateCircuit() error {
	current := sc.GetActiveCircuit()

	var exitCountry string
	if current != nil && len(current.Nodes) > 0 {
		exitCountry = current.Nodes[len(current.Nodes)-1].Country
	}

	newCircuit, err := sc.BuildCircuit(exitCountry)
	if err != nil {
		return err
	}

	if current != nil {
		current.Active = false
	}

	sc.SetActiveCircuit(newCircuit)
	sc.stats.CircuitsRotated.Add(1)

	return nil
}

// circuitRotationLoop rotates circuits periodically
func (sc *SecureCore) circuitRotationLoop() {
	ticker := time.NewTicker(sc.config.CircuitRotation)
	defer ticker.Stop()

	for {
		select {
		case <-sc.ctx.Done():
			return
		case <-ticker.C:
			sc.RotateCircuit()
		}
	}
}

// Route routes data through the circuit
func (sc *SecureCore) Route(data []byte) ([]net.Conn, error) {
	circuit := sc.GetActiveCircuit()
	if circuit == nil {
		return nil, errors.New("no active circuit")
	}

	if !circuit.Active || time.Now().After(circuit.ExpiresAt) {
		if err := sc.RotateCircuit(); err != nil {
			return nil, err
		}
		circuit = sc.GetActiveCircuit()
	}

	// Create connections to each hop
	connections := make([]net.Conn, 0, len(circuit.Nodes))

	for i, node := range circuit.Nodes {
		addr := net.JoinHostPort(node.Address, string(rune(node.Port)))

		var conn net.Conn
		var err error

		if i == 0 {
			// First hop: connect directly
			conn, err = net.DialTimeout("tcp", addr, 30*time.Second)
		} else {
			// Subsequent hops: tunnel through previous connection
			// This would use the VPN tunnel protocol
			conn, err = net.DialTimeout("tcp", addr, 30*time.Second)
		}

		if err != nil {
			// Close previous connections on failure
			for _, c := range connections {
				c.Close()
			}
			return nil, err
		}

		connections = append(connections, conn)
	}

	circuit.Stats.PacketsSent.Add(1)
	circuit.Stats.BytesSent.Add(uint64(len(data)))
	sc.stats.BytesRouted.Add(uint64(len(data)))

	return connections, nil
}

// Stats returns secure core statistics
func (sc *SecureCore) Stats() SecureCoreStatsSummary {
	return SecureCoreStatsSummary{
		CircuitsCreated: sc.stats.CircuitsCreated.Load(),
		CircuitsRotated: sc.stats.CircuitsRotated.Load(),
		CurrentHops:     sc.stats.CurrentHops,
		BytesRouted:     sc.stats.BytesRouted.Load(),
		AverageLatency:  sc.stats.AverageLatency,
	}
}

// SecureCoreStatsSummary provides a summary of stats
type SecureCoreStatsSummary struct {
	CircuitsCreated uint64
	CircuitsRotated uint64
	CurrentHops     int
	BytesRouted     uint64
	AverageLatency  time.Duration
}

// Close shuts down secure core
func (sc *SecureCore) Close() {
	sc.cancel()

	sc.circuitsMu.Lock()
	for _, circuit := range sc.circuits {
		circuit.Active = false
	}
	sc.circuitsMu.Unlock()
}

// SmartProtocol handles smart protocol selection
type SmartProtocol struct {
	config *SmartProtocolConfig

	// Available protocols
	protocols map[string]*ProtocolProfile

	// Current selection
	currentProtocol string

	// Network conditions
	conditions NetworkConditions

	mu sync.RWMutex
}

// SmartProtocolConfig configures smart protocol selection
type SmartProtocolConfig struct {
	// Enable automatic protocol selection
	AutoSelect bool

	// Preferred protocol (when auto is disabled)
	PreferredProtocol string

	// Enable protocol fallback
	EnableFallback bool

	// Network detection interval
	DetectionInterval time.Duration
}

// ProtocolProfile describes a VPN protocol
type ProtocolProfile struct {
	Name          string
	Port          int
	OverTCP       bool
	OverUDP       bool
	Speed         int // 1-10
	Reliability   int // 1-10
	Stealth       int // 1-10 (ability to bypass firewalls)
	Security      int // 1-10
	BatteryUsage  int // 1-10 (lower is better)
}

// NetworkConditions represents current network conditions
type NetworkConditions struct {
	Latency        time.Duration
	PacketLoss     float64
	Bandwidth      uint64
	IsRestrictive  bool // Firewall/DPI detected
	IsMobile       bool
	IsMetered      bool
	SignalStrength int // 1-5
}

// DefaultSmartProtocolConfig returns defaults
func DefaultSmartProtocolConfig() *SmartProtocolConfig {
	return &SmartProtocolConfig{
		AutoSelect:        true,
		EnableFallback:    true,
		DetectionInterval: 30 * time.Second,
	}
}

// NewSmartProtocol creates a smart protocol selector
func NewSmartProtocol(config *SmartProtocolConfig) *SmartProtocol {
	if config == nil {
		config = DefaultSmartProtocolConfig()
	}

	sp := &SmartProtocol{
		config:    config,
		protocols: make(map[string]*ProtocolProfile),
	}

	// Register built-in protocols
	sp.registerBuiltinProtocols()

	return sp
}

// registerBuiltinProtocols registers available protocols
func (sp *SmartProtocol) registerBuiltinProtocols() {
	sp.protocols["wireguard"] = &ProtocolProfile{
		Name:         "WireGuard",
		Port:         51820,
		OverUDP:      true,
		Speed:        10,
		Reliability:  9,
		Stealth:      5,
		Security:     10,
		BatteryUsage: 3,
	}

	sp.protocols["openvpn-udp"] = &ProtocolProfile{
		Name:         "OpenVPN UDP",
		Port:         1194,
		OverUDP:      true,
		Speed:        8,
		Reliability:  8,
		Stealth:      6,
		Security:     9,
		BatteryUsage: 5,
	}

	sp.protocols["openvpn-tcp"] = &ProtocolProfile{
		Name:         "OpenVPN TCP",
		Port:         443,
		OverTCP:      true,
		Speed:        6,
		Reliability:  9,
		Stealth:      8,
		Security:     9,
		BatteryUsage: 6,
	}

	sp.protocols["stealth"] = &ProtocolProfile{
		Name:         "Stealth (obfs4)",
		Port:         443,
		OverTCP:      true,
		Speed:        5,
		Reliability:  7,
		Stealth:      10,
		Security:     8,
		BatteryUsage: 7,
	}

	sp.protocols["ikev2"] = &ProtocolProfile{
		Name:         "IKEv2/IPsec",
		Port:         500,
		OverUDP:      true,
		Speed:        9,
		Reliability:  8,
		Stealth:      4,
		Security:     9,
		BatteryUsage: 4,
	}
}

// SelectProtocol selects the best protocol for current conditions
func (sp *SmartProtocol) SelectProtocol(conditions NetworkConditions) string {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	sp.conditions = conditions

	if !sp.config.AutoSelect && sp.config.PreferredProtocol != "" {
		sp.currentProtocol = sp.config.PreferredProtocol
		return sp.currentProtocol
	}

	// Score each protocol
	type protocolScore struct {
		name  string
		score float64
	}

	scores := make([]protocolScore, 0, len(sp.protocols))

	for name, profile := range sp.protocols {
		score := sp.scoreProtocol(profile, conditions)
		scores = append(scores, protocolScore{name, score})
	}

	// Sort by score descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	if len(scores) > 0 {
		sp.currentProtocol = scores[0].name
	}

	return sp.currentProtocol
}

// scoreProtocol calculates a score for a protocol given conditions
func (sp *SmartProtocol) scoreProtocol(profile *ProtocolProfile, conditions NetworkConditions) float64 {
	var score float64

	// Base scores
	score += float64(profile.Speed) * 0.25
	score += float64(profile.Reliability) * 0.2
	score += float64(profile.Security) * 0.15

	// Adjust for network conditions
	if conditions.IsRestrictive {
		// Prefer stealth protocols in restrictive networks
		score += float64(profile.Stealth) * 0.3
		if profile.OverTCP && profile.Port == 443 {
			score += 2 // HTTPS port bonus
		}
	} else {
		// Normal network - prefer speed
		score += float64(profile.Speed) * 0.1
	}

	if conditions.IsMobile {
		// Mobile: prefer low battery usage and fast handoffs
		score -= float64(profile.BatteryUsage) * 0.1
		if profile.Name == "IKEv2/IPsec" {
			score += 1 // IKEv2 handles network changes well
		}
	}

	if conditions.IsMetered {
		// Metered: prefer protocols with less overhead
		if profile.Name == "WireGuard" {
			score += 1 // Lowest overhead
		}
	}

	if conditions.PacketLoss > 0.05 {
		// High packet loss: prefer TCP-based protocols
		if profile.OverTCP {
			score += 1
		}
	}

	if conditions.Latency > 200*time.Millisecond {
		// High latency: avoid TCP, prefer UDP
		if profile.OverUDP {
			score += 0.5
		}
	}

	return score
}

// GetCurrentProtocol returns the currently selected protocol
func (sp *SmartProtocol) GetCurrentProtocol() string {
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	return sp.currentProtocol
}

// GetProtocolProfile returns the profile for a protocol
func (sp *SmartProtocol) GetProtocolProfile(name string) *ProtocolProfile {
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	return sp.protocols[name]
}

// RegisterProtocol registers a custom protocol
func (sp *SmartProtocol) RegisterProtocol(name string, profile *ProtocolProfile) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.protocols[name] = profile
}
