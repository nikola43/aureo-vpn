package wireguard

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Manager handles WireGuard interface operations
type Manager struct {
	interfaceName string
}

// NewManager creates a new WireGuard manager
func NewManager(interfaceName string) *Manager {
	return &Manager{
		interfaceName: interfaceName,
	}
}

// SetupInterface creates and configures a WireGuard interface
func (m *Manager) SetupInterface(config ServerConfig) error {
	// Create WireGuard interface
	cmd := exec.Command("ip", "link", "add", "dev", m.interfaceName, "type", "wireguard")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create interface: %w", err)
	}

	// Set private key
	setKeyCmd := exec.Command("wg", "set", m.interfaceName, "private-key", "/dev/stdin")
	setKeyCmd.Stdin = strings.NewReader(config.PrivateKey)
	if err := setKeyCmd.Run(); err != nil {
		return fmt.Errorf("failed to set private key: %w", err)
	}

	// Set listen port
	cmd = exec.Command("wg", "set", m.interfaceName, "listen-port", fmt.Sprintf("%d", config.ListenPort))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set listen port: %w", err)
	}

	// Set IP address
	cmd = exec.Command("ip", "address", "add", config.Address, "dev", m.interfaceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set IP address: %w", err)
	}

	// Bring interface up
	cmd = exec.Command("ip", "link", "set", "up", "dev", m.interfaceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to bring interface up: %w", err)
	}

	// Execute PostUp commands
	for _, postUpCmd := range config.PostUp {
		cmd = exec.Command("sh", "-c", postUpCmd)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to execute PostUp command %s: %w", postUpCmd, err)
		}
	}

	return nil
}

// AddPeer adds a peer to the WireGuard interface
func (m *Manager) AddPeer(peer PeerConfig) error {
	args := []string{"set", m.interfaceName, "peer", peer.PublicKey}

	if peer.PresharedKey != "" {
		args = append(args, "preshared-key", "/dev/stdin")
	}

	if len(peer.AllowedIPs) > 0 {
		args = append(args, "allowed-ips", strings.Join(peer.AllowedIPs, ","))
	}

	if peer.PersistentKeepalive > 0 {
		args = append(args, "persistent-keepalive", fmt.Sprintf("%d", peer.PersistentKeepalive))
	}

	cmd := exec.Command("wg", args...)
	if peer.PresharedKey != "" {
		cmd.Stdin = strings.NewReader(peer.PresharedKey)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add peer: %w", err)
	}

	return nil
}

// RemovePeer removes a peer from the WireGuard interface
func (m *Manager) RemovePeer(publicKey string) error {
	cmd := exec.Command("wg", "set", m.interfaceName, "peer", publicKey, "remove")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove peer: %w", err)
	}
	return nil
}

// GetInterfaceStats retrieves statistics for the WireGuard interface
func (m *Manager) GetInterfaceStats() (*InterfaceStats, error) {
	cmd := exec.Command("wg", "show", m.interfaceName, "dump")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get interface stats: %w", err)
	}

	stats := &InterfaceStats{
		InterfaceName: m.interfaceName,
		Peers:         make([]PeerStats, 0),
	}

	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i == 0 || line == "" {
			continue // Skip header and empty lines
		}

		fields := strings.Fields(line)
		if len(fields) < 8 {
			continue
		}

		peer := PeerStats{
			PublicKey:         fields[0],
			Endpoint:          fields[2],
			AllowedIPs:        strings.Split(fields[3], ","),
			LatestHandshake:   parseTimestamp(fields[4]),
			BytesReceived:     parseInt64(fields[5]),
			BytesSent:         parseInt64(fields[6]),
			PersistentKeepalive: parseInt(fields[7]),
		}

		stats.Peers = append(stats.Peers, peer)
	}

	return stats, nil
}

// TeardownInterface removes the WireGuard interface
func (m *Manager) TeardownInterface(config ServerConfig) error {
	// Execute PostDown commands
	for _, postDownCmd := range config.PostDown {
		cmd := exec.Command("sh", "-c", postDownCmd)
		_ = cmd.Run() // Ignore errors for PostDown commands
	}

	// Delete interface
	cmd := exec.Command("ip", "link", "del", "dev", m.interfaceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete interface: %w", err)
	}

	return nil
}

// InterfaceStats represents statistics for a WireGuard interface
type InterfaceStats struct {
	InterfaceName string
	Peers         []PeerStats
}

// PeerStats represents statistics for a WireGuard peer
type PeerStats struct {
	PublicKey           string
	Endpoint            string
	AllowedIPs          []string
	LatestHandshake     time.Time
	BytesReceived       int64
	BytesSent           int64
	PersistentKeepalive int
}

// Helper functions
func parseTimestamp(s string) time.Time {
	timestamp := parseInt64(s)
	if timestamp == 0 {
		return time.Time{}
	}
	return time.Unix(timestamp, 0)
}

func parseInt64(s string) int64 {
	var val int64
	fmt.Sscanf(s, "%d", &val)
	return val
}

func parseInt(s string) int {
	var val int
	fmt.Sscanf(s, "%d", &val)
	return val
}
