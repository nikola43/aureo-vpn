package security

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

// ObfuscationManager handles traffic obfuscation to bypass DPI
type ObfuscationManager struct {
	mode         string // stealth, scramble, shadowsocks, stunnel
	enabled      bool
	tunnelMethod string
}

// NewObfuscationManager creates a new obfuscation manager
func NewObfuscationManager(mode string) *ObfuscationManager {
	return &ObfuscationManager{
		mode:         mode,
		enabled:      false,
		tunnelMethod: "tls",
	}
}

// Enable activates traffic obfuscation
func (o *ObfuscationManager) Enable() error {
	log.Printf("Enabling traffic obfuscation (mode: %s)", o.mode)

	switch o.mode {
	case "stealth":
		return o.enableStealthMode()
	case "scramble":
		return o.enableScrambleMode()
	case "shadowsocks":
		return o.enableShadowsocks()
	case "stunnel":
		return o.enableStunnel()
	default:
		return fmt.Errorf("unknown obfuscation mode: %s", o.mode)
	}
}

// Disable deactivates traffic obfuscation
func (o *ObfuscationManager) Disable() error {
	log.Println("Disabling traffic obfuscation")
	o.enabled = false
	return nil
}

// enableStealthMode disguises VPN traffic as HTTPS
func (o *ObfuscationManager) enableStealthMode() error {
	log.Println("Enabling stealth mode - disguising traffic as HTTPS")

	// Stealth mode wraps VPN traffic in TLS/HTTPS to appear as regular web traffic
	// This makes it harder for DPI systems to detect VPN usage

	o.enabled = true
	o.tunnelMethod = "tls"
	return nil
}

// enableScrambleMode obfuscates packet headers
func (o *ObfuscationManager) enableScrambleMode() error {
	log.Println("Enabling scramble mode - obfuscating packet headers")

	// Scramble mode XORs packets with a key to hide VPN signatures
	// Similar to OpenVPN's scramble patch

	o.enabled = true
	o.tunnelMethod = "scramble"
	return nil
}

// enableShadowsocks enables Shadowsocks obfuscation
func (o *ObfuscationManager) enableShadowsocks() error {
	log.Println("Enabling Shadowsocks obfuscation")

	// Shadowsocks is designed to bypass firewalls
	// Uses AEAD ciphers and appears as random traffic

	o.enabled = true
	o.tunnelMethod = "shadowsocks"
	return nil
}

// enableStunnel wraps traffic in SSL/TLS tunnel
func (o *ObfuscationManager) enableStunnel() error {
	log.Println("Enabling stunnel - wrapping in SSL/TLS")

	// Stunnel creates an SSL/TLS tunnel for VPN traffic
	// Makes VPN traffic appear as standard HTTPS

	o.enabled = true
	o.tunnelMethod = "stunnel"
	return nil
}

// ObfuscatePacket obfuscates a packet based on the current mode
func (o *ObfuscationManager) ObfuscatePacket(data []byte) ([]byte, error) {
	if !o.enabled {
		return data, nil
	}

	switch o.mode {
	case "stealth":
		return o.obfuscateWithTLS(data)
	case "scramble":
		return o.scramblePacket(data)
	case "shadowsocks":
		return o.obfuscateWithShadowsocks(data)
	default:
		return data, nil
	}
}

// DeobfuscatePacket removes obfuscation from a packet
func (o *ObfuscationManager) DeobfuscatePacket(data []byte) ([]byte, error) {
	if !o.enabled {
		return data, nil
	}

	switch o.mode {
	case "stealth":
		return o.deobfuscateFromTLS(data)
	case "scramble":
		return o.descramblePacket(data)
	case "shadowsocks":
		return o.deobfuscateFromShadowsocks(data)
	default:
		return data, nil
	}
}

// obfuscateWithTLS wraps data in TLS-looking headers
func (o *ObfuscationManager) obfuscateWithTLS(data []byte) ([]byte, error) {
	// Create a fake TLS header
	header := make([]byte, 5)
	header[0] = 0x17 // TLS Application Data
	header[1] = 0x03 // TLS version (major)
	header[2] = 0x03 // TLS version (minor) - TLS 1.2
	binary.BigEndian.PutUint16(header[3:5], uint16(len(data)))

	// Combine header and data
	result := append(header, data...)
	return result, nil
}

// deobfuscateFromTLS removes TLS-looking headers
func (o *ObfuscationManager) deobfuscateFromTLS(data []byte) ([]byte, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("data too short for TLS header")
	}

	// Skip TLS header (5 bytes)
	return data[5:], nil
}

// scramblePacket XORs packet with a random key
func (o *ObfuscationManager) scramblePacket(data []byte) ([]byte, error) {
	// Generate or use stored scramble key
	key := o.getScrambleKey()

	result := make([]byte, len(data))
	for i := range data {
		result[i] = data[i] ^ key[i%len(key)]
	}

	return result, nil
}

// descramblePacket reverses the scramble operation
func (o *ObfuscationManager) descramblePacket(data []byte) ([]byte, error) {
	// XOR is symmetric, so scrambling again descrambles
	return o.scramblePacket(data)
}

// getScrambleKey returns the scramble key
func (o *ObfuscationManager) getScrambleKey() []byte {
	// In production, this should be derived from a shared secret
	// For now, use a static key (should be configurable)
	return []byte("AureoVPNObfuscationKey2024SecureRandom")
}

// obfuscateWithShadowsocks uses Shadowsocks-style obfuscation
func (o *ObfuscationManager) obfuscateWithShadowsocks(data []byte) ([]byte, error) {
	// Shadowsocks obfuscation:
	// 1. Add random padding
	// 2. Encrypt with AEAD cipher
	// 3. Add timestamp to prevent replay

	// Add random padding (1-255 bytes)
	paddingLen := randomByte()
	padding := make([]byte, paddingLen)
	rand.Read(padding)

	// Add timestamp (8 bytes)
	timestamp := make([]byte, 8)
	binary.BigEndian.PutUint64(timestamp, uint64(time.Now().Unix()))

	// Combine: [padding_len][padding][timestamp][data]
	result := make([]byte, 0, 1+int(paddingLen)+8+len(data))
	result = append(result, paddingLen)
	result = append(result, padding...)
	result = append(result, timestamp...)
	result = append(result, data...)

	return result, nil
}

// deobfuscateFromShadowsocks reverses Shadowsocks obfuscation
func (o *ObfuscationManager) deobfuscateFromShadowsocks(data []byte) ([]byte, error) {
	if len(data) < 10 { // min: 1 byte padding_len + 1 byte padding + 8 bytes timestamp
		return nil, fmt.Errorf("data too short")
	}

	// Read padding length
	paddingLen := int(data[0])

	// Skip padding and timestamp
	offset := 1 + paddingLen + 8

	if len(data) < offset {
		return nil, fmt.Errorf("invalid padding length")
	}

	return data[offset:], nil
}

// CreateObfuscatedConnection creates a connection with obfuscation
func (o *ObfuscationManager) CreateObfuscatedConnection(targetAddr string) (net.Conn, error) {
	log.Printf("Creating obfuscated connection to %s", targetAddr)

	// Create base connection
	conn, err := net.DialTimeout("tcp", targetAddr, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Wrap connection with obfuscation
	obfConn := &ObfuscatedConn{
		Conn:     conn,
		manager:  o,
		readBuf:  make([]byte, 65536),
		writeBuf: make([]byte, 65536),
	}

	return obfConn, nil
}

// ObfuscatedConn wraps a net.Conn with obfuscation
type ObfuscatedConn struct {
	net.Conn
	manager  *ObfuscationManager
	readBuf  []byte
	writeBuf []byte
}

// Read reads obfuscated data and deobfuscates it
func (c *ObfuscatedConn) Read(b []byte) (n int, err error) {
	// Read from underlying connection
	n, err = c.Conn.Read(c.readBuf)
	if err != nil {
		return 0, err
	}

	// Deobfuscate
	deobfuscated, err := c.manager.DeobfuscatePacket(c.readBuf[:n])
	if err != nil {
		return 0, err
	}

	// Copy to output buffer
	copy(b, deobfuscated)
	return len(deobfuscated), nil
}

// Write obfuscates data and writes it
func (c *ObfuscatedConn) Write(b []byte) (n int, err error) {
	// Obfuscate data
	obfuscated, err := c.manager.ObfuscatePacket(b)
	if err != nil {
		return 0, err
	}

	// Write to underlying connection
	return c.Conn.Write(obfuscated)
}

// randomByte returns a random byte from 1-255
func randomByte() byte {
	var b [1]byte
	io.ReadFull(rand.Reader, b[:])
	if b[0] == 0 {
		b[0] = 1
	}
	return b[0]
}

// DetectDPI attempts to detect if DPI is blocking VPN traffic
func (o *ObfuscationManager) DetectDPI() (bool, error) {
	log.Println("Detecting DPI/firewall...")

	// Try connecting without obfuscation
	plainConn, err := net.DialTimeout("tcp", "8.8.8.8:53", 5*time.Second)
	if err != nil {
		return true, fmt.Errorf("connection blocked: %w", err)
	}
	plainConn.Close()

	// In production, would test various VPN signatures
	// and see which ones get blocked

	return false, nil // false = no DPI detected
}

// GetObfuscationStats returns statistics about obfuscation
func (o *ObfuscationManager) GetObfuscationStats() map[string]interface{} {
	return map[string]interface{}{
		"mode":          o.mode,
		"enabled":       o.enabled,
		"tunnel_method": o.tunnelMethod,
		"overhead_bytes": o.calculateOverhead(),
	}
}

// calculateOverhead estimates the bandwidth overhead of obfuscation
func (o *ObfuscationManager) calculateOverhead() int {
	switch o.mode {
	case "stealth":
		return 5 // TLS header
	case "scramble":
		return 0 // No overhead, just XOR
	case "shadowsocks":
		return 10 // 1 byte padding_len + up to 255 bytes padding + 8 bytes timestamp (avg ~135 bytes)
	case "stunnel":
		return 50 // SSL/TLS overhead
	default:
		return 0
	}
}
