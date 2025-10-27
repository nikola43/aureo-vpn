package config

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/nikola43/aureo-vpn/pkg/database"
	"github.com/nikola43/aureo-vpn/pkg/models"
	"github.com/nikola43/aureo-vpn/pkg/protocols/openvpn"
	"github.com/nikola43/aureo-vpn/pkg/protocols/wireguard"
	"gorm.io/gorm"
)

// Generator handles VPN configuration generation
type Generator struct {
	db *gorm.DB
}

// NewGenerator creates a new configuration generator
func NewGenerator() *Generator {
	return &Generator{
		db: database.GetDB(),
	}
}

// GenerateWireGuardConfig generates a WireGuard configuration for a user
func (g *Generator) GenerateWireGuardConfig(userID, nodeID uuid.UUID) (*models.Config, string, error) {
	// Get node
	var node models.VPNNode
	if err := g.db.First(&node, nodeID).Error; err != nil {
		return nil, "", fmt.Errorf("node not found: %w", err)
	}

	if !node.SupportsWireGuard {
		return nil, "", fmt.Errorf("node does not support WireGuard")
	}

	// Generate client keypair
	keyPair, err := wireguard.GenerateKeyPair()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate keypair: %w", err)
	}

	// Get used IPs for this node
	var usedIPs []string
	g.db.Model(&models.Config{}).
		Where("node_id = ? AND protocol = ? AND is_active = ?", nodeID, "wireguard", true).
		Pluck("allowed_ips", &usedIPs)

	// Allocate client IP
	clientIP, err := wireguard.AllocateClientIP("10.8.0.0/24", usedIPs)
	if err != nil {
		return nil, "", fmt.Errorf("failed to allocate IP: %w", err)
	}

	// Create WireGuard config
	wgConfig := wireguard.Config{
		PrivateKey: keyPair.PrivateKey,
		Address:    []string{clientIP},
		DNS:        []string{"1.1.1.1", "1.0.0.1"}, // Cloudflare DNS for leak protection
		MTU:        1420,
		Table:      "auto",

		PeerPublicKey:       node.PublicKey,
		PeerEndpoint:        fmt.Sprintf("%s:%d", node.PublicIP, node.WireGuardPort),
		AllowedIPs:          []string{"0.0.0.0/0", "::/0"}, // Route all traffic
		PersistentKeepalive: 25,
	}

	// Generate config content
	configContent, err := wireguard.GenerateClientConfig(wgConfig)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate config: %w", err)
	}

	// Save to database
	config := &models.Config{
		UserID:              userID,
		NodeID:              nodeID,
		Protocol:            "wireguard",
		ConfigName:          fmt.Sprintf("%s-wireguard", node.Name),
		ConfigContent:       configContent,
		PublicKey:           keyPair.PublicKey,
		PrivateKey:          keyPair.PrivateKey, // Should be encrypted in production
		DNSServers:          "1.1.1.1,1.0.0.1",
		AllowedIPs:          clientIP,
		MTU:                 1420,
		PersistentKeepalive: 25,
		IsActive:            true,
	}

	if err := g.db.Create(config).Error; err != nil {
		return nil, "", fmt.Errorf("failed to save config: %w", err)
	}

	return config, configContent, nil
}

// GenerateOpenVPNConfig generates an OpenVPN configuration for a user
func (g *Generator) GenerateOpenVPNConfig(userID, nodeID uuid.UUID) (*models.Config, string, error) {
	// Get node
	var node models.VPNNode
	if err := g.db.First(&node, nodeID).Error; err != nil {
		return nil, "", fmt.Errorf("node not found: %w", err)
	}

	if !node.SupportsOpenVPN {
		return nil, "", fmt.Errorf("node does not support OpenVPN")
	}

	// Generate certificates (simplified - use proper PKI in production)
	caCert, caKey, err := openvpn.GenerateCertificate("Aureo VPN CA", true)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate CA: %w", err)
	}

	clientCert, clientKey, err := openvpn.GenerateCertificate("client", false)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate client cert: %w", err)
	}

	tlsAuth, err := openvpn.GenerateTLSAuthKey()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate TLS auth key: %w", err)
	}

	// Create OpenVPN config
	ovpnConfig := openvpn.ClientConfig{
		ServerHost:        node.PublicIP,
		ServerPort:        node.OpenVPNPort,
		Protocol:          "udp",
		Device:            "tun",
		Cipher:            "AES-256-GCM",
		Auth:              "SHA256",
		RemoteCertTLS:     "server",
		CACert:            caCert,
		ClientCert:        clientCert,
		ClientKey:         clientKey,
		TLSAuth:           tlsAuth,
		DNS:               []string{"1.1.1.1", "1.0.0.1"},
		RedirectGateway:   true,
		PersistKey:        true,
		PersistTun:        true,
		Keepalive:         openvpn.Keepalive{Interval: 10, Timeout: 120},
		Verb:              3,
	}

	// Generate config content
	configContent, err := openvpn.GenerateClientConfig(ovpnConfig)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate config: %w", err)
	}

	// Save to database
	config := &models.Config{
		UserID:        userID,
		NodeID:        nodeID,
		Protocol:      "openvpn",
		ConfigName:    fmt.Sprintf("%s-openvpn", node.Name),
		ConfigContent: configContent,
		PublicKey:     "", // Not used for OpenVPN
		PrivateKey:    clientKey + caKey, // Store both keys (encrypted in production)
		DNSServers:    "1.1.1.1,1.0.0.1",
		AllowedIPs:    "0.0.0.0/0,::/0",
		IsActive:      true,
	}

	if err := g.db.Create(config).Error; err != nil {
		return nil, "", fmt.Errorf("failed to save config: %w", err)
	}

	return config, configContent, nil
}

// GetUserConfigs retrieves all configs for a user
func (g *Generator) GetUserConfigs(userID uuid.UUID) ([]models.Config, error) {
	var configs []models.Config
	if err := g.db.Where("user_id = ?", userID).
		Preload("Node").
		Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

// DeleteConfig deletes a configuration
func (g *Generator) DeleteConfig(configID uuid.UUID) error {
	return g.db.Delete(&models.Config{}, configID).Error
}

// GetConfigContent retrieves the decrypted config content
func (g *Generator) GetConfigContent(configID uuid.UUID) (string, error) {
	var config models.Config
	if err := g.db.First(&config, configID).Error; err != nil {
		return "", err
	}

	// In production, decrypt the config content here
	return config.ConfigContent, nil
}
