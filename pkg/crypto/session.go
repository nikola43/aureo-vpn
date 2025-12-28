// Package crypto provides anonymous session management
// Integrates all cryptographic components for military-grade VPN sessions
package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
)

// Session errors
var (
	ErrSessionInvalid    = errors.New("invalid session")
	ErrSessionNotFound   = errors.New("session not found")
	ErrHandshakeFailed   = errors.New("handshake failed")
	ErrAuthenticationErr = errors.New("authentication failed")
)

// SessionState represents the current state of a session
type SessionState uint8

const (
	SessionStateInitializing SessionState = iota
	SessionStateHandshake
	SessionStateActive
	SessionStateRekeying
	SessionStateClosing
	SessionStateClosed
)

// AnonymousSession provides a fully anonymous, encrypted VPN session
type AnonymousSession struct {
	// Session identification (all blinded/hashed)
	id           [32]byte // Random session ID
	blindedID    []byte   // Blinded version for external use
	peerBlindedID []byte  // Peer's blinded ID

	// Cryptographic components
	tunnel       *TunnelCipher       // Double-layer encryption
	keyExchange  *KeyExchange        // Key exchange
	obfuscator   *TrafficObfuscator  // Traffic obfuscation
	metadata     *MetadataEncryptor  // Metadata encryption
	header       *HeaderEncryptor    // Header encryption

	// Zero-knowledge components
	blinding     *BlindingFactor     // Session blinding
	zkSession    *ZeroKnowledgeSession // Zero-knowledge session

	// Session state
	state        SessionState
	createdAt    time.Time
	lastActivity time.Time
	expiresAt    time.Time

	// Rekeying
	rekeyCounter atomic.Uint64
	lastRekey    time.Time
	rekeyPending bool

	// Statistics (anonymized)
	bytesSent    atomic.Uint64
	bytesRecv    atomic.Uint64
	packetsSent  atomic.Uint64
	packetsRecv  atomic.Uint64

	// Sequence numbers
	sendSeq atomic.Uint64
	recvSeq atomic.Uint64

	// Configuration
	config *SessionConfig

	mu sync.RWMutex
}

// SessionConfig holds session configuration
type SessionConfig struct {
	// Session lifetime
	MaxDuration     time.Duration
	IdleTimeout     time.Duration
	RekeyInterval   time.Duration
	RekeyPacketLimit uint64

	// Security settings
	ObfuscationMode   ObfuscationMode
	EnableDummyTraffic bool
	TimestampGranularity time.Duration

	// Buffer sizes
	SendBufferSize int
	RecvBufferSize int
}

// DefaultSessionConfig returns secure defaults
func DefaultSessionConfig() *SessionConfig {
	return &SessionConfig{
		MaxDuration:          24 * time.Hour,
		IdleTimeout:          30 * time.Minute,
		RekeyInterval:        1 * time.Hour,
		RekeyPacketLimit:     1 << 20, // 1M packets
		ObfuscationMode:      ObfuscationFull,
		EnableDummyTraffic:   true,
		TimestampGranularity: time.Minute,
		SendBufferSize:       64 * 1024,
		RecvBufferSize:       64 * 1024,
	}
}

// HighSecuritySessionConfig returns maximum security settings
func HighSecuritySessionConfig() *SessionConfig {
	return &SessionConfig{
		MaxDuration:          4 * time.Hour,
		IdleTimeout:          10 * time.Minute,
		RekeyInterval:        15 * time.Minute,
		RekeyPacketLimit:     1 << 16, // 64K packets
		ObfuscationMode:      ObfuscationFull,
		EnableDummyTraffic:   true,
		TimestampGranularity: time.Minute,
		SendBufferSize:       32 * 1024,
		RecvBufferSize:       32 * 1024,
	}
}

// SessionManager manages anonymous VPN sessions
type SessionManager struct {
	// Active sessions (keyed by blinded ID)
	sessions map[string]*AnonymousSession

	// Session limits
	maxSessions int
	sessionTTL  time.Duration

	// Master keys (derived from hardware/secure storage)
	masterKey    []byte
	metadataKey  []byte
	blindingKey  []byte

	// IP anonymizer
	ipAnonymizer *IPAnonymizer

	// Data store
	dataStore *AnonymousDataStore

	// Cleanup
	cleanupInterval time.Duration
	stopChan       chan struct{}
	wg             sync.WaitGroup

	mu sync.RWMutex
}

// SessionManagerConfig holds configuration for session manager
type SessionManagerConfig struct {
	MaxSessions     int
	SessionTTL      time.Duration
	CleanupInterval time.Duration
	MasterKey       []byte
}

// DefaultSessionManagerConfig returns secure defaults
func DefaultSessionManagerConfig() *SessionManagerConfig {
	masterKey := make([]byte, 64)
	rand.Read(masterKey)

	return &SessionManagerConfig{
		MaxSessions:     10000,
		SessionTTL:      24 * time.Hour,
		CleanupInterval: 5 * time.Minute,
		MasterKey:       masterKey,
	}
}

// NewSessionManager creates a new anonymous session manager
func NewSessionManager(config *SessionManagerConfig) (*SessionManager, error) {
	if config == nil {
		config = DefaultSessionManagerConfig()
	}

	if len(config.MasterKey) < 32 {
		return nil, errors.New("master key must be at least 32 bytes")
	}

	// Derive sub-keys
	metadataKey := deriveSubKey(config.MasterKey, "metadata")
	blindingKey := deriveSubKey(config.MasterKey, "blinding")

	// Create IP anonymizer
	ipAnon, err := NewIPAnonymizer()
	if err != nil {
		return nil, err
	}

	// Create data store
	dataStore, err := NewAnonymousDataStore(config.MasterKey)
	if err != nil {
		ipAnon.Close()
		return nil, err
	}

	sm := &SessionManager{
		sessions:        make(map[string]*AnonymousSession),
		maxSessions:     config.MaxSessions,
		sessionTTL:      config.SessionTTL,
		masterKey:       config.MasterKey,
		metadataKey:     metadataKey,
		blindingKey:     blindingKey,
		ipAnonymizer:    ipAnon,
		dataStore:       dataStore,
		cleanupInterval: config.CleanupInterval,
		stopChan:        make(chan struct{}),
	}

	// Start cleanup goroutine
	sm.wg.Add(1)
	go sm.cleanupLoop()

	return sm, nil
}

// deriveSubKey derives a sub-key from master key
func deriveSubKey(masterKey []byte, purpose string) []byte {
	hkdfReader := hkdf.New(sha256.New, masterKey, nil, []byte("aureo-vpn-"+purpose))
	key := make([]byte, 32)
	io.ReadFull(hkdfReader, key)
	return key
}

// CreateSession creates a new anonymous session
func (sm *SessionManager) CreateSession(peerPublicKey []byte, config *SessionConfig) (*AnonymousSession, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if len(sm.sessions) >= sm.maxSessions {
		return nil, errors.New("session limit reached")
	}

	if config == nil {
		config = DefaultSessionConfig()
	}

	// Generate session ID
	var sessionID [32]byte
	if _, err := rand.Read(sessionID[:]); err != nil {
		return nil, err
	}

	// Create key exchange
	keyExchange, err := NewKeyExchange()
	if err != nil {
		return nil, err
	}

	// Derive keys
	keys, err := keyExchange.DeriveKeys(peerPublicKey, true)
	if err != nil {
		keyExchange.Wipe()
		return nil, err
	}

	// Create tunnel cipher
	tunnel, err := NewTunnelCipher(keys)
	if err != nil {
		keys.Wipe()
		keyExchange.Wipe()
		return nil, err
	}

	// Create obfuscator
	obfConfig := &ObfuscationConfig{
		Mode:          config.ObfuscationMode,
		EnableDummy:   config.EnableDummyTraffic,
		MinPacketSize: 64,
		MaxPacketSize: 1400,
		MinDelay:      1 * time.Millisecond,
		MaxDelay:      50 * time.Millisecond,
		Jitter:        0.3,
	}
	obfuscator := NewTrafficObfuscator(obfConfig)

	// Create metadata encryptor
	metadata, err := NewMetadataEncryptor(sm.metadataKey, nil)
	if err != nil {
		obfuscator.Close()
		tunnel.Close()
		keys.Wipe()
		keyExchange.Wipe()
		return nil, err
	}

	// Create header encryptor
	header, err := NewHeaderEncryptor(keys.PrimarySendKey)
	if err != nil {
		metadata.Wipe()
		obfuscator.Close()
		tunnel.Close()
		keys.Wipe()
		keyExchange.Wipe()
		return nil, err
	}

	// Create blinding factor
	blinding, err := NewBlindingFactor()
	if err != nil {
		header.Wipe()
		metadata.Wipe()
		obfuscator.Close()
		tunnel.Close()
		keys.Wipe()
		keyExchange.Wipe()
		return nil, err
	}

	// Create zero-knowledge session
	zkSession, err := NewZeroKnowledgeSession(peerPublicKey, true)
	if err != nil {
		blinding.Wipe()
		header.Wipe()
		metadata.Wipe()
		obfuscator.Close()
		tunnel.Close()
		keys.Wipe()
		keyExchange.Wipe()
		return nil, err
	}

	now := time.Now()
	session := &AnonymousSession{
		id:           sessionID,
		blindedID:    blinding.Blind(sessionID[:]),
		tunnel:       tunnel,
		keyExchange:  keyExchange,
		obfuscator:   obfuscator,
		metadata:     metadata,
		header:       header,
		blinding:     blinding,
		zkSession:    zkSession,
		state:        SessionStateActive,
		createdAt:    now,
		lastActivity: now,
		expiresAt:    now.Add(config.MaxDuration),
		lastRekey:    now,
		config:       config,
	}

	// Store session by blinded ID
	sm.sessions[string(session.blindedID)] = session

	// Wipe derived keys (copies stored in components)
	keys.Wipe()

	return session, nil
}

// GetSession retrieves a session by blinded ID
func (sm *SessionManager) GetSession(blindedID []byte) (*AnonymousSession, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[string(blindedID)]
	if !exists {
		return nil, ErrSessionNotFound
	}

	if session.IsExpired() {
		return nil, ErrSessionExpired
	}

	return session, nil
}

// RemoveSession removes and cleans up a session
func (sm *SessionManager) RemoveSession(blindedID []byte) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[string(blindedID)]
	if !exists {
		return ErrSessionNotFound
	}

	session.Close()
	delete(sm.sessions, string(blindedID))

	return nil
}

// cleanupLoop periodically cleans up expired sessions
func (sm *SessionManager) cleanupLoop() {
	defer sm.wg.Done()

	ticker := time.NewTicker(sm.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-sm.stopChan:
			return
		case <-ticker.C:
			sm.cleanupExpiredSessions()
		}
	}
}

// cleanupExpiredSessions removes expired sessions
func (sm *SessionManager) cleanupExpiredSessions() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for id, session := range sm.sessions {
		if session.IsExpired() || session.state == SessionStateClosed {
			session.Close()
			delete(sm.sessions, id)
		}
	}
}

// AnonymizeIP anonymizes an IP address
func (sm *SessionManager) AnonymizeIP(ip []byte) []byte {
	return sm.ipAnonymizer.AnonymizeIP(ip)
}

// StoreSessionData stores encrypted session data
func (sm *SessionManager) StoreSessionData(blindedID, data []byte) error {
	return sm.dataStore.Store(blindedID, data)
}

// RetrieveSessionData retrieves encrypted session data
func (sm *SessionManager) RetrieveSessionData(blindedID []byte) ([]byte, error) {
	return sm.dataStore.Retrieve(blindedID)
}

// SessionCount returns the number of active sessions
func (sm *SessionManager) SessionCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}

// Close shuts down the session manager
func (sm *SessionManager) Close() {
	close(sm.stopChan)
	sm.wg.Wait()

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Close all sessions
	for id, session := range sm.sessions {
		session.Close()
		delete(sm.sessions, id)
	}

	// Cleanup components
	sm.ipAnonymizer.Close()
	sm.dataStore.Close()

	// Wipe keys
	wipeBytes(sm.masterKey)
	wipeBytes(sm.metadataKey)
	wipeBytes(sm.blindingKey)
}

// Session methods

// Send encrypts and sends data through the session
func (s *AnonymousSession) Send(data []byte) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != SessionStateActive {
		return nil, ErrSessionInvalid
	}

	if s.IsExpired() {
		s.state = SessionStateClosed
		return nil, ErrSessionExpired
	}

	// Check if rekey is needed
	if s.needsRekey() {
		s.rekeyPending = true
	}

	// Get sequence number
	seq := s.sendSeq.Add(1)

	// Encrypt header
	header, err := s.header.EncryptHeader(
		uint8(PacketTypeData),
		seq,
		time.Now(),
		s.blindedID,
	)
	if err != nil {
		return nil, err
	}

	// Encrypt data with tunnel cipher
	encrypted, err := s.tunnel.Encrypt(data, PacketTypeData)
	if err != nil {
		return nil, err
	}

	// Apply obfuscation
	packet, err := s.obfuscator.Obfuscate(encrypted)
	if err != nil {
		return nil, err
	}

	// Apply delay for timing protection
	if packet.Delay > 0 {
		time.Sleep(packet.Delay)
	}

	// Combine header and data
	headerBytes := serializeHeader(header)
	result := make([]byte, len(headerBytes)+len(packet.Data))
	copy(result, headerBytes)
	copy(result[len(headerBytes):], packet.Data)

	// Update statistics
	s.bytesSent.Add(uint64(len(data)))
	s.packetsSent.Add(1)
	s.lastActivity = time.Now()

	return result, nil
}

// Receive decrypts received data
func (s *AnonymousSession) Receive(data []byte) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != SessionStateActive {
		return nil, ErrSessionInvalid
	}

	if s.IsExpired() {
		s.state = SessionStateClosed
		return nil, ErrSessionExpired
	}

	// Parse header (header size is variable, parse dynamically)
	headerLen, header, err := deserializeHeader(data)
	if err != nil {
		return nil, err
	}

	// Decrypt and verify header
	packetType, seq, _, err := s.header.DecryptHeader(header)
	if err != nil {
		return nil, err
	}

	// Verify sequence number (basic replay protection)
	expectedSeq := s.recvSeq.Add(1)
	if seq < expectedSeq-1000 { // Allow some out-of-order packets
		return nil, errors.New("replay detected")
	}

	// Deobfuscate
	packet := &ObfuscatedPacket{
		Data:    data[headerLen:],
		IsDummy: false,
	}
	deobfuscated, isDummy, err := s.obfuscator.Deobfuscate(packet)
	if err != nil {
		return nil, err
	}

	if isDummy {
		return nil, nil // Dummy packet, discard
	}

	// Decrypt data with tunnel cipher
	plaintext, _, err := s.tunnel.Decrypt(deobfuscated)
	if err != nil {
		return nil, err
	}

	// Handle packet type
	if PacketType(packetType) == PacketTypeRekey {
		s.rekeyPending = true
	}

	// Update statistics
	s.bytesRecv.Add(uint64(len(plaintext)))
	s.packetsRecv.Add(1)
	s.lastActivity = time.Now()

	return plaintext, nil
}

// needsRekey checks if rekeying is needed
func (s *AnonymousSession) needsRekey() bool {
	if s.rekeyPending {
		return true
	}

	// Check packet count
	if s.packetsSent.Load()+s.packetsRecv.Load() > s.config.RekeyPacketLimit {
		return true
	}

	// Check time
	if time.Since(s.lastRekey) > s.config.RekeyInterval {
		return true
	}

	return false
}

// Rekey performs session rekeying
func (s *AnonymousSession) Rekey(peerPublicKey []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.state = SessionStateRekeying

	// Generate new key exchange
	newKeyExchange, err := NewKeyExchange()
	if err != nil {
		s.state = SessionStateActive
		return err
	}

	// Derive new keys
	newKeys, err := newKeyExchange.DeriveKeys(peerPublicKey, true)
	if err != nil {
		newKeyExchange.Wipe()
		s.state = SessionStateActive
		return err
	}

	// Update tunnel
	if err := s.tunnel.CompleteRekey(newKeys); err != nil {
		newKeys.Wipe()
		newKeyExchange.Wipe()
		s.state = SessionStateActive
		return err
	}

	// Update header encryptor
	newHeader, err := NewHeaderEncryptor(newKeys.PrimarySendKey)
	if err != nil {
		newKeys.Wipe()
		newKeyExchange.Wipe()
		s.state = SessionStateActive
		return err
	}

	// Cleanup old components
	s.keyExchange.Wipe()
	s.header.Wipe()

	// Install new components
	s.keyExchange = newKeyExchange
	s.header = newHeader

	// Update blinding
	s.blinding.Wipe()
	newBlinding, _ := NewBlindingFactor()
	s.blinding = newBlinding
	s.blindedID = newBlinding.Blind(s.id[:])

	// Update state
	s.lastRekey = time.Now()
	s.rekeyPending = false
	s.rekeyCounter.Add(1)
	s.state = SessionStateActive

	// Wipe intermediate keys
	newKeys.Wipe()

	return nil
}

// GetPublicKey returns the session's public key
func (s *AnonymousSession) GetPublicKey() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.keyExchange.PublicKey()
}

// GetBlindedID returns the session's blinded ID
func (s *AnonymousSession) GetBlindedID() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]byte, len(s.blindedID))
	copy(result, s.blindedID)
	return result
}

// IsExpired checks if the session has expired
func (s *AnonymousSession) IsExpired() bool {
	now := time.Now()
	return now.After(s.expiresAt) || time.Since(s.lastActivity) > s.config.IdleTimeout
}

// Stats returns anonymized session statistics
func (s *AnonymousSession) Stats() SessionStats {
	return SessionStats{
		BytesSent:     s.bytesSent.Load(),
		BytesReceived: s.bytesRecv.Load(),
		PacketsSent:   s.packetsSent.Load(),
		PacketsReceived: s.packetsRecv.Load(),
		RekeyCount:    s.rekeyCounter.Load(),
		Duration:      time.Since(s.createdAt),
		IdleTime:      time.Since(s.lastActivity),
	}
}

// SessionStats holds anonymized session statistics
type SessionStats struct {
	BytesSent       uint64
	BytesReceived   uint64
	PacketsSent     uint64
	PacketsReceived uint64
	RekeyCount      uint64
	Duration        time.Duration
	IdleTime        time.Duration
}

// Close securely destroys the session
func (s *AnonymousSession) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == SessionStateClosed {
		return
	}

	s.state = SessionStateClosing

	// Close components
	if s.tunnel != nil {
		s.tunnel.Close()
	}
	if s.keyExchange != nil {
		s.keyExchange.Wipe()
	}
	if s.obfuscator != nil {
		s.obfuscator.Close()
	}
	if s.metadata != nil {
		s.metadata.Wipe()
	}
	if s.header != nil {
		s.header.Wipe()
	}
	if s.blinding != nil {
		s.blinding.Wipe()
	}
	if s.zkSession != nil {
		s.zkSession.Close()
	}

	// Wipe sensitive data
	wipeBytes(s.id[:])
	wipeBytes(s.blindedID)
	wipeBytes(s.peerBlindedID)

	s.state = SessionStateClosed
}

// serializeHeader serializes a packet header
func serializeHeader(h *AnonymousPacketHeader) []byte {
	size := 1 + // Version
		4 + len(h.PacketType) +
		4 + len(h.Sequence) +
		4 + len(h.Timestamp) +
		4 + len(h.SessionBinding) +
		4 + len(h.Padding)

	data := make([]byte, size)
	offset := 0

	// Version
	data[offset] = h.Version
	offset++

	// PacketType
	binary.BigEndian.PutUint32(data[offset:], uint32(len(h.PacketType)))
	offset += 4
	copy(data[offset:], h.PacketType)
	offset += len(h.PacketType)

	// Sequence
	binary.BigEndian.PutUint32(data[offset:], uint32(len(h.Sequence)))
	offset += 4
	copy(data[offset:], h.Sequence)
	offset += len(h.Sequence)

	// Timestamp
	binary.BigEndian.PutUint32(data[offset:], uint32(len(h.Timestamp)))
	offset += 4
	copy(data[offset:], h.Timestamp)
	offset += len(h.Timestamp)

	// SessionBinding
	binary.BigEndian.PutUint32(data[offset:], uint32(len(h.SessionBinding)))
	offset += 4
	copy(data[offset:], h.SessionBinding)
	offset += len(h.SessionBinding)

	// Padding
	binary.BigEndian.PutUint32(data[offset:], uint32(len(h.Padding)))
	offset += 4
	copy(data[offset:], h.Padding)

	return data
}

// deserializeHeader deserializes a packet header
func deserializeHeader(data []byte) (int, *AnonymousPacketHeader, error) {
	if len(data) < 1 {
		return 0, nil, ErrInvalidMetadata
	}

	h := &AnonymousPacketHeader{}
	offset := 0

	// Version
	h.Version = data[offset]
	offset++

	// PacketType
	if offset+4 > len(data) {
		return 0, nil, ErrInvalidMetadata
	}
	ptLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if offset+ptLen > len(data) {
		return 0, nil, ErrInvalidMetadata
	}
	h.PacketType = make([]byte, ptLen)
	copy(h.PacketType, data[offset:offset+ptLen])
	offset += ptLen

	// Sequence
	if offset+4 > len(data) {
		return 0, nil, ErrInvalidMetadata
	}
	seqLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if offset+seqLen > len(data) {
		return 0, nil, ErrInvalidMetadata
	}
	h.Sequence = make([]byte, seqLen)
	copy(h.Sequence, data[offset:offset+seqLen])
	offset += seqLen

	// Timestamp
	if offset+4 > len(data) {
		return 0, nil, ErrInvalidMetadata
	}
	tsLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if offset+tsLen > len(data) {
		return 0, nil, ErrInvalidMetadata
	}
	h.Timestamp = make([]byte, tsLen)
	copy(h.Timestamp, data[offset:offset+tsLen])
	offset += tsLen

	// SessionBinding
	if offset+4 > len(data) {
		return 0, nil, ErrInvalidMetadata
	}
	sbLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if offset+sbLen > len(data) {
		return 0, nil, ErrInvalidMetadata
	}
	h.SessionBinding = make([]byte, sbLen)
	copy(h.SessionBinding, data[offset:offset+sbLen])
	offset += sbLen

	// Padding
	if offset+4 > len(data) {
		return 0, nil, ErrInvalidMetadata
	}
	padLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if offset+padLen > len(data) {
		return 0, nil, ErrInvalidMetadata
	}
	h.Padding = make([]byte, padLen)
	copy(h.Padding, data[offset:offset+padLen])
	offset += padLen

	return offset, h, nil
}

// VerifyPeer performs mutual authentication
func (s *AnonymousSession) VerifyPeer(expectedPeerID []byte) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Blind the expected ID
	blindedExpected := s.blinding.Blind(expectedPeerID)

	// Constant-time comparison
	return subtle.ConstantTimeCompare(s.peerBlindedID, blindedExpected) == 1
}

// CreateChallenge creates an authentication challenge
func (s *AnonymousSession) CreateChallenge() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	challenge := make([]byte, 32)
	if _, err := rand.Read(challenge); err != nil {
		return nil, err
	}

	// Encrypt challenge
	cipher, err := chacha20poly1305.NewX(s.id[:])
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, 24)
	rand.Read(nonce)

	return cipher.Seal(nonce, nonce, challenge, nil), nil
}

// VerifyChallenge verifies an authentication challenge response
func (s *AnonymousSession) VerifyChallenge(challenge, response []byte) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(challenge) < 24 || len(response) < 24 {
		return false
	}

	// Decrypt challenge
	cipher, err := chacha20poly1305.NewX(s.id[:])
	if err != nil {
		return false
	}

	nonce := challenge[:24]
	ct := challenge[24:]

	plainChallenge, err := cipher.Open(nil, nonce, ct, nil)
	if err != nil {
		return false
	}

	// Decrypt response
	respNonce := response[:24]
	respCt := response[24:]

	plainResponse, err := cipher.Open(nil, respNonce, respCt, nil)
	if err != nil {
		return false
	}

	// Verify response matches challenge
	return subtle.ConstantTimeCompare(plainChallenge, plainResponse) == 1
}
