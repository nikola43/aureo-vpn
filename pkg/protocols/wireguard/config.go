package wireguard

import (
	"fmt"
	"net"
	"strings"
)

// Config represents WireGuard configuration
type Config struct {
	// Interface settings
	PrivateKey string
	Address    []string // CIDR addresses for the interface
	DNS        []string
	MTU        int
	Table      string // routing table, "auto" or "off"

	// Peer settings
	PeerPublicKey       string
	PeerEndpoint        string // host:port
	AllowedIPs          []string
	PersistentKeepalive int
	PresharedKey        string
}

// ServerConfig represents server-side WireGuard configuration
type ServerConfig struct {
	PrivateKey string
	Address    string // Server's VPN IP address
	ListenPort int
	PostUp     []string // Commands to run after interface is up
	PostDown   []string // Commands to run after interface is down
	SaveConfig bool

	// Peers
	Peers []PeerConfig
}

// PeerConfig represents a peer configuration on the server
type PeerConfig struct {
	PublicKey           string
	PresharedKey        string
	AllowedIPs          []string
	PersistentKeepalive int
}

// GenerateClientConfig generates a WireGuard client configuration file
func GenerateClientConfig(cfg Config) (string, error) {
	var sb strings.Builder

	// Interface section
	sb.WriteString("[Interface]\n")
	sb.WriteString(fmt.Sprintf("PrivateKey = %s\n", cfg.PrivateKey))

	// Addresses
	if len(cfg.Address) > 0 {
		sb.WriteString(fmt.Sprintf("Address = %s\n", strings.Join(cfg.Address, ", ")))
	}

	// DNS
	if len(cfg.DNS) > 0 {
		sb.WriteString(fmt.Sprintf("DNS = %s\n", strings.Join(cfg.DNS, ", ")))
	}

	// MTU
	if cfg.MTU > 0 {
		sb.WriteString(fmt.Sprintf("MTU = %d\n", cfg.MTU))
	}

	// Table
	if cfg.Table != "" {
		sb.WriteString(fmt.Sprintf("Table = %s\n", cfg.Table))
	}

	sb.WriteString("\n")

	// Peer section
	sb.WriteString("[Peer]\n")
	sb.WriteString(fmt.Sprintf("PublicKey = %s\n", cfg.PeerPublicKey))

	if cfg.PresharedKey != "" {
		sb.WriteString(fmt.Sprintf("PresharedKey = %s\n", cfg.PresharedKey))
	}

	if cfg.PeerEndpoint != "" {
		sb.WriteString(fmt.Sprintf("Endpoint = %s\n", cfg.PeerEndpoint))
	}

	if len(cfg.AllowedIPs) > 0 {
		sb.WriteString(fmt.Sprintf("AllowedIPs = %s\n", strings.Join(cfg.AllowedIPs, ", ")))
	}

	if cfg.PersistentKeepalive > 0 {
		sb.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", cfg.PersistentKeepalive))
	}

	return sb.String(), nil
}

// GenerateServerConfig generates a WireGuard server configuration file
func GenerateServerConfig(cfg ServerConfig) (string, error) {
	var sb strings.Builder

	// Interface section
	sb.WriteString("[Interface]\n")
	sb.WriteString(fmt.Sprintf("PrivateKey = %s\n", cfg.PrivateKey))
	sb.WriteString(fmt.Sprintf("Address = %s\n", cfg.Address))
	sb.WriteString(fmt.Sprintf("ListenPort = %d\n", cfg.ListenPort))

	if cfg.SaveConfig {
		sb.WriteString("SaveConfig = true\n")
	}

	// PostUp commands
	for _, cmd := range cfg.PostUp {
		sb.WriteString(fmt.Sprintf("PostUp = %s\n", cmd))
	}

	// PostDown commands
	for _, cmd := range cfg.PostDown {
		sb.WriteString(fmt.Sprintf("PostDown = %s\n", cmd))
	}

	// Peers
	for _, peer := range cfg.Peers {
		sb.WriteString("\n[Peer]\n")
		sb.WriteString(fmt.Sprintf("PublicKey = %s\n", peer.PublicKey))

		if peer.PresharedKey != "" {
			sb.WriteString(fmt.Sprintf("PresharedKey = %s\n", peer.PresharedKey))
		}

		if len(peer.AllowedIPs) > 0 {
			sb.WriteString(fmt.Sprintf("AllowedIPs = %s\n", strings.Join(peer.AllowedIPs, ", ")))
		}

		if peer.PersistentKeepalive > 0 {
			sb.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", peer.PersistentKeepalive))
		}
	}

	return sb.String(), nil
}

// AllocateClientIP allocates a new IP address for a client
func AllocateClientIP(networkCIDR string, usedIPs []string) (string, error) {
	_, ipNet, err := net.ParseCIDR(networkCIDR)
	if err != nil {
		return "", fmt.Errorf("invalid CIDR: %w", err)
	}

	// Create a map of used IPs for quick lookup
	used := make(map[string]bool)
	for _, ip := range usedIPs {
		used[ip] = true
	}

	// Iterate through the network to find an available IP
	for ip := ipNet.IP.Mask(ipNet.Mask); ipNet.Contains(ip); incrementIP(ip) {
		ipStr := ip.String()

		// Skip network address, broadcast address, and gateway (assumed to be .1)
		if ip.Equal(ipNet.IP) || ip[len(ip)-1] == 0 || ip[len(ip)-1] == 255 || ip[len(ip)-1] == 1 {
			continue
		}

		// Check if IP is not used
		if !used[ipStr] {
			ones, _ := ipNet.Mask.Size()
			return fmt.Sprintf("%s/%d", ipStr, ones), nil
		}
	}

	return "", fmt.Errorf("no available IP addresses in network")
}

// incrementIP increments an IP address
func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// ValidateConfig validates a WireGuard configuration
func ValidateConfig(cfg Config) error {
	// Validate private key
	if err := ValidatePrivateKey(cfg.PrivateKey); err != nil {
		return fmt.Errorf("invalid private key: %w", err)
	}

	// Validate peer public key
	if err := ValidatePublicKey(cfg.PeerPublicKey); err != nil {
		return fmt.Errorf("invalid peer public key: %w", err)
	}

	// Validate addresses
	for _, addr := range cfg.Address {
		if _, _, err := net.ParseCIDR(addr); err != nil {
			return fmt.Errorf("invalid address %s: %w", addr, err)
		}
	}

	// Validate allowed IPs
	for _, allowedIP := range cfg.AllowedIPs {
		if _, _, err := net.ParseCIDR(allowedIP); err != nil {
			return fmt.Errorf("invalid allowed IP %s: %w", allowedIP, err)
		}
	}

	return nil
}
