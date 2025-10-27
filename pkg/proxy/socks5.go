package proxy

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

// SOCKS5Server implements a SOCKS5 proxy server
type SOCKS5Server struct {
	listenAddr string
	authMethod byte // 0x00 = no auth, 0x02 = username/password
	username   string
	password   string
	timeout    time.Duration
}

// SOCKS5 constants
const (
	SOCKS5Version = 0x05

	// Authentication methods
	AuthNone     = 0x00
	AuthGSSAPI   = 0x01
	AuthPassword = 0x02
	AuthNoAcceptable = 0xFF

	// Address types
	AddrIPv4   = 0x01
	AddrDomain = 0x03
	AddrIPv6   = 0x04

	// Commands
	CmdConnect   = 0x01
	CmdBind      = 0x02
	CmdUDPAssoc  = 0x03

	// Reply codes
	ReplySuccess              = 0x00
	ReplyGeneralFailure       = 0x01
	ReplyConnectionNotAllowed = 0x02
	ReplyNetworkUnreachable   = 0x03
	ReplyHostUnreachable      = 0x04
	ReplyConnectionRefused    = 0x05
	ReplyTTLExpired           = 0x06
	ReplyCommandNotSupported  = 0x07
	ReplyAddrTypeNotSupported = 0x08
)

// NewSOCKS5Server creates a new SOCKS5 proxy server
func NewSOCKS5Server(listenAddr string) *SOCKS5Server {
	return &SOCKS5Server{
		listenAddr: listenAddr,
		authMethod: AuthNone,
		timeout:    30 * time.Second,
	}
}

// SetAuth sets authentication credentials
func (s *SOCKS5Server) SetAuth(username, password string) {
	s.username = username
	s.password = password
	s.authMethod = AuthPassword
}

// Start starts the SOCKS5 server
func (s *SOCKS5Server) Start() error {
	listener, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return fmt.Errorf("failed to start SOCKS5 server: %w", err)
	}
	defer listener.Close()

	log.Printf("SOCKS5 server listening on %s", s.listenAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

// handleConnection handles a SOCKS5 client connection
func (s *SOCKS5Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Set deadline
	conn.SetDeadline(time.Now().Add(s.timeout))

	// 1. Negotiation phase
	if err := s.negotiate(conn); err != nil {
		log.Printf("Negotiation failed: %v", err)
		return
	}

	// 2. Authentication (if required)
	if s.authMethod == AuthPassword {
		if err := s.authenticate(conn); err != nil {
			log.Printf("Authentication failed: %v", err)
			return
		}
	}

	// 3. Request phase
	targetConn, err := s.handleRequest(conn)
	if err != nil {
		log.Printf("Request handling failed: %v", err)
		return
	}
	defer targetConn.Close()

	// 4. Relay phase - bidirectional copy
	log.Println("Starting bidirectional relay")
	s.relay(conn, targetConn)
}

// negotiate handles the SOCKS5 negotiation phase
func (s *SOCKS5Server) negotiate(conn net.Conn) error {
	// Read version and methods
	buf := make([]byte, 257)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}

	if n < 2 || buf[0] != SOCKS5Version {
		return fmt.Errorf("invalid SOCKS version: %d", buf[0])
	}

	nMethods := int(buf[1])
	if n < 2+nMethods {
		return fmt.Errorf("invalid number of methods")
	}

	methods := buf[2 : 2+nMethods]

	// Check if client supports our auth method
	var selectedMethod byte = AuthNoAcceptable
	for _, method := range methods {
		if method == s.authMethod {
			selectedMethod = method
			break
		}
	}

	// Send selected method
	_, err = conn.Write([]byte{SOCKS5Version, selectedMethod})
	if err != nil {
		return err
	}

	if selectedMethod == AuthNoAcceptable {
		return fmt.Errorf("no acceptable authentication method")
	}

	return nil
}

// authenticate handles username/password authentication
func (s *SOCKS5Server) authenticate(conn net.Conn) error {
	buf := make([]byte, 513)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}

	if n < 3 || buf[0] != 0x01 {
		return fmt.Errorf("invalid auth version")
	}

	// Parse username
	usernameLen := int(buf[1])
	if n < 2+usernameLen {
		return fmt.Errorf("invalid username length")
	}
	username := string(buf[2 : 2+usernameLen])

	// Parse password
	passwordLen := int(buf[2+usernameLen])
	if n < 3+usernameLen+passwordLen {
		return fmt.Errorf("invalid password length")
	}
	password := string(buf[3+usernameLen : 3+usernameLen+passwordLen])

	// Verify credentials
	var status byte = 0x01 // failure
	if username == s.username && password == s.password {
		status = 0x00 // success
	}

	// Send auth response
	_, err = conn.Write([]byte{0x01, status})
	if err != nil {
		return err
	}

	if status != 0x00 {
		return fmt.Errorf("invalid credentials")
	}

	return nil
}

// handleRequest handles the SOCKS5 request phase
func (s *SOCKS5Server) handleRequest(conn net.Conn) (net.Conn, error) {
	buf := make([]byte, 262)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	if n < 7 || buf[0] != SOCKS5Version {
		s.sendReply(conn, ReplyGeneralFailure)
		return nil, fmt.Errorf("invalid request")
	}

	cmd := buf[1]
	// buf[2] is reserved
	addrType := buf[3]

	// Parse destination address
	var destAddr string
	var destPort uint16
	offset := 4

	switch addrType {
	case AddrIPv4:
		if n < 10 {
			s.sendReply(conn, ReplyGeneralFailure)
			return nil, fmt.Errorf("invalid IPv4 address")
		}
		destAddr = net.IP(buf[offset : offset+4]).String()
		destPort = binary.BigEndian.Uint16(buf[offset+4 : offset+6])

	case AddrDomain:
		domainLen := int(buf[offset])
		if n < offset+1+domainLen+2 {
			s.sendReply(conn, ReplyGeneralFailure)
			return nil, fmt.Errorf("invalid domain name")
		}
		destAddr = string(buf[offset+1 : offset+1+domainLen])
		destPort = binary.BigEndian.Uint16(buf[offset+1+domainLen : offset+1+domainLen+2])

	case AddrIPv6:
		if n < 22 {
			s.sendReply(conn, ReplyGeneralFailure)
			return nil, fmt.Errorf("invalid IPv6 address")
		}
		destAddr = net.IP(buf[offset : offset+16]).String()
		destPort = binary.BigEndian.Uint16(buf[offset+16 : offset+18])

	default:
		s.sendReply(conn, ReplyAddrTypeNotSupported)
		return nil, fmt.Errorf("unsupported address type: %d", addrType)
	}

	// Handle command
	switch cmd {
	case CmdConnect:
		return s.handleConnect(conn, destAddr, destPort)
	case CmdBind:
		s.sendReply(conn, ReplyCommandNotSupported)
		return nil, fmt.Errorf("BIND command not supported")
	case CmdUDPAssoc:
		s.sendReply(conn, ReplyCommandNotSupported)
		return nil, fmt.Errorf("UDP ASSOCIATE command not supported")
	default:
		s.sendReply(conn, ReplyCommandNotSupported)
		return nil, fmt.Errorf("unknown command: %d", cmd)
	}
}

// handleConnect handles CONNECT command
func (s *SOCKS5Server) handleConnect(conn net.Conn, destAddr string, destPort uint16) (net.Conn, error) {
	target := fmt.Sprintf("%s:%d", destAddr, destPort)
	log.Printf("Connecting to %s", target)

	targetConn, err := net.DialTimeout("tcp", target, 10*time.Second)
	if err != nil {
		log.Printf("Failed to connect to %s: %v", target, err)
		s.sendReply(conn, ReplyConnectionRefused)
		return nil, err
	}

	// Send success reply
	s.sendReply(conn, ReplySuccess)

	return targetConn, nil
}

// sendReply sends a SOCKS5 reply
func (s *SOCKS5Server) sendReply(conn net.Conn, reply byte) error {
	// Reply format: [VER][REP][RSV][ATYP][BND.ADDR][BND.PORT]
	response := []byte{
		SOCKS5Version,
		reply,
		0x00, // reserved
		0x01, // IPv4
		0, 0, 0, 0, // bind address (0.0.0.0)
		0, 0, // bind port (0)
	}

	_, err := conn.Write(response)
	return err
}

// relay performs bidirectional relay between client and target
func (s *SOCKS5Server) relay(client, target net.Conn) {
	// Create channels for completion
	done := make(chan struct{}, 2)

	// Client -> Target
	go func() {
		io.Copy(target, client)
		done <- struct{}{}
	}()

	// Target -> Client
	go func() {
		io.Copy(client, target)
		done <- struct{}{}
	}()

	// Wait for either direction to complete
	<-done
}

// GetStats returns SOCKS5 server statistics
func (s *SOCKS5Server) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"listen_address": s.listenAddr,
		"auth_enabled":   s.authMethod == AuthPassword,
		"timeout":        s.timeout.Seconds(),
	}
}
