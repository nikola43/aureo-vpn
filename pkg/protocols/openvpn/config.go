package openvpn

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// ClientConfig represents an OpenVPN client configuration
type ClientConfig struct {
	// Server details
	ServerHost string
	ServerPort int
	Protocol   string // udp or tcp

	// Certificates and keys
	CACert     string
	ClientCert string
	ClientKey  string
	TLSAuth    string

	// Connection settings
	Device       string // tun or tap
	Cipher       string
	Auth         string
	Compression  string
	RemoteCertTLS string
	VerifyX509Name string

	// DNS and routing
	DNS            []string
	RedirectGateway bool
	Routes         []Route

	// Features
	Mute           int
	Verb           int
	PersistKey     bool
	PersistTun     bool
	Keepalive      Keepalive
	FloatingServer bool
}

// ServerConfig represents an OpenVPN server configuration
type ServerConfig struct {
	Port       int
	Protocol   string // udp or tcp
	Device     string // tun or tap
	VPNNetwork string // e.g., "10.8.0.0 255.255.255.0"
	VPNNetmask string

	// Certificates and keys
	CACert     string
	ServerCert string
	ServerKey  string
	DHParams   string
	TLSAuth    string

	// Security
	Cipher              string
	Auth                string
	TLSVersionMin       string
	TLSCipher           string
	RemoteCertTLS       string

	// Client settings
	MaxClients      int
	ClientToClient  bool
	Keepalive       Keepalive
	PersistKey      bool
	PersistTun      bool

	// Routing
	PushRoutes      []Route
	PushDNS         []string
	RedirectGateway bool

	// Logging
	Status    string
	LogAppend string
	Verb      int
}

// Route represents a network route
type Route struct {
	Network string
	Netmask string
	Gateway string
}

// Keepalive represents keepalive settings
type Keepalive struct {
	Interval int
	Timeout  int
}

// GenerateClientConfig generates an OpenVPN client configuration file
func GenerateClientConfig(cfg ClientConfig) (string, error) {
	var sb strings.Builder

	sb.WriteString("client\n")
	sb.WriteString(fmt.Sprintf("dev %s\n", cfg.Device))
	sb.WriteString(fmt.Sprintf("proto %s\n", cfg.Protocol))
	sb.WriteString(fmt.Sprintf("remote %s %d\n", cfg.ServerHost, cfg.ServerPort))

	if cfg.FloatingServer {
		sb.WriteString("float\n")
	}

	sb.WriteString("resolv-retry infinite\n")
	sb.WriteString("nobind\n")

	if cfg.PersistKey {
		sb.WriteString("persist-key\n")
	}

	if cfg.PersistTun {
		sb.WriteString("persist-tun\n")
	}

	// Security settings
	if cfg.Cipher != "" {
		sb.WriteString(fmt.Sprintf("cipher %s\n", cfg.Cipher))
	}

	if cfg.Auth != "" {
		sb.WriteString(fmt.Sprintf("auth %s\n", cfg.Auth))
	}

	if cfg.RemoteCertTLS != "" {
		sb.WriteString(fmt.Sprintf("remote-cert-tls %s\n", cfg.RemoteCertTLS))
	}

	if cfg.VerifyX509Name != "" {
		sb.WriteString(fmt.Sprintf("verify-x509-name %s name\n", cfg.VerifyX509Name))
	}

	// Keepalive
	if cfg.Keepalive.Interval > 0 && cfg.Keepalive.Timeout > 0 {
		sb.WriteString(fmt.Sprintf("keepalive %d %d\n", cfg.Keepalive.Interval, cfg.Keepalive.Timeout))
	}

	// Compression
	if cfg.Compression != "" {
		sb.WriteString(fmt.Sprintf("comp-lzo %s\n", cfg.Compression))
	}

	// Logging
	if cfg.Mute > 0 {
		sb.WriteString(fmt.Sprintf("mute %d\n", cfg.Mute))
	}

	if cfg.Verb > 0 {
		sb.WriteString(fmt.Sprintf("verb %d\n", cfg.Verb))
	}

	// Inline certificates and keys
	sb.WriteString("<ca>\n")
	sb.WriteString(cfg.CACert)
	sb.WriteString("</ca>\n\n")

	sb.WriteString("<cert>\n")
	sb.WriteString(cfg.ClientCert)
	sb.WriteString("</cert>\n\n")

	sb.WriteString("<key>\n")
	sb.WriteString(cfg.ClientKey)
	sb.WriteString("</key>\n\n")

	if cfg.TLSAuth != "" {
		sb.WriteString("key-direction 1\n")
		sb.WriteString("<tls-auth>\n")
		sb.WriteString(cfg.TLSAuth)
		sb.WriteString("</tls-auth>\n")
	}

	return sb.String(), nil
}

// GenerateServerConfig generates an OpenVPN server configuration file
func GenerateServerConfig(cfg ServerConfig) (string, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("port %d\n", cfg.Port))
	sb.WriteString(fmt.Sprintf("proto %s\n", cfg.Protocol))
	sb.WriteString(fmt.Sprintf("dev %s\n", cfg.Device))

	sb.WriteString(fmt.Sprintf("server %s\n", cfg.VPNNetwork))

	sb.WriteString("ca ca.crt\n")
	sb.WriteString("cert server.crt\n")
	sb.WriteString("key server.key\n")
	sb.WriteString("dh dh2048.pem\n")

	if cfg.TLSAuth != "" {
		sb.WriteString("tls-auth ta.key 0\n")
	}

	// Security
	sb.WriteString(fmt.Sprintf("cipher %s\n", cfg.Cipher))
	sb.WriteString(fmt.Sprintf("auth %s\n", cfg.Auth))

	if cfg.TLSVersionMin != "" {
		sb.WriteString(fmt.Sprintf("tls-version-min %s\n", cfg.TLSVersionMin))
	}

	if cfg.TLSCipher != "" {
		sb.WriteString(fmt.Sprintf("tls-cipher %s\n", cfg.TLSCipher))
	}

	if cfg.RemoteCertTLS != "" {
		sb.WriteString(fmt.Sprintf("remote-cert-tls %s\n", cfg.RemoteCertTLS))
	}

	// Client settings
	sb.WriteString(fmt.Sprintf("max-clients %d\n", cfg.MaxClients))

	if cfg.ClientToClient {
		sb.WriteString("client-to-client\n")
	}

	if cfg.Keepalive.Interval > 0 && cfg.Keepalive.Timeout > 0 {
		sb.WriteString(fmt.Sprintf("keepalive %d %d\n", cfg.Keepalive.Interval, cfg.Keepalive.Timeout))
	}

	if cfg.PersistKey {
		sb.WriteString("persist-key\n")
	}

	if cfg.PersistTun {
		sb.WriteString("persist-tun\n")
	}

	// Push settings to clients
	for _, dns := range cfg.PushDNS {
		sb.WriteString(fmt.Sprintf("push \"dhcp-option DNS %s\"\n", dns))
	}

	for _, route := range cfg.PushRoutes {
		sb.WriteString(fmt.Sprintf("push \"route %s %s\"\n", route.Network, route.Netmask))
	}

	if cfg.RedirectGateway {
		sb.WriteString("push \"redirect-gateway def1 bypass-dhcp\"\n")
	}

	// Logging
	if cfg.Status != "" {
		sb.WriteString(fmt.Sprintf("status %s\n", cfg.Status))
	}

	if cfg.LogAppend != "" {
		sb.WriteString(fmt.Sprintf("log-append %s\n", cfg.LogAppend))
	}

	sb.WriteString(fmt.Sprintf("verb %d\n", cfg.Verb))

	return sb.String(), nil
}

// GenerateCertificate generates a self-signed certificate
func GenerateCertificate(commonName string, isCA bool) (certPEM, keyPEM string, err error) {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour) // 1 year

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return "", "", fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"Aureo VPN"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	// Create certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to create certificate: %w", err)
	}

	// Encode certificate to PEM
	certPEMBlock := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})

	// Encode private key to PEM
	keyPEMBlock := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	return string(certPEMBlock), string(keyPEMBlock), nil
}

// GenerateTLSAuthKey generates a TLS authentication key
func GenerateTLSAuthKey() (string, error) {
	key := make([]byte, 256)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("failed to generate TLS auth key: %w", err)
	}

	return fmt.Sprintf("%x", key), nil
}
