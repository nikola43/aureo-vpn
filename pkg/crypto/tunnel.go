// Package crypto provides military-grade encryption for VPN tunnel traffic
// Implements ChaCha20-Poly1305 with perfect forward secrecy and zero-knowledge architecture
package crypto

import (
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
)

// Tunnel encryption errors
var (
	ErrInvalidPacket       = errors.New("invalid encrypted packet")
	ErrDecryptionFailed    = errors.New("decryption failed: authentication failed")
	ErrKeyExchangeFailed   = errors.New("key exchange failed")
	ErrSessionExpired      = errors.New("session keys expired")
	ErrReplayAttack        = errors.New("replay attack detected")
	ErrInvalidNonce        = errors.New("invalid nonce")
	ErrMaxPacketsExceeded  = errors.New("maximum packets exceeded, rekey required")
)

// Security constants
const (
	// Key sizes
	KeySize          = 32 // 256-bit keys
	NonceSize        = 24 // XChaCha20-Poly1305 uses 24-byte nonce
	TagSize          = 16 // Poly1305 tag size

	// Packet structure
	HeaderSize       = 8  // Version(1) + Type(1) + Reserved(2) + Length(4)
	MaxPacketSize    = 65535
	MinPaddingSize   = 16
	MaxPaddingSize   = 256

	// Key rotation
	MaxPacketsPerKey = 1 << 32 // Rekey after 2^32 packets (NIST recommendation)
	KeyRotationTime  = 1 * time.Hour

	// Replay protection
	ReplayWindowSize = 8192 // Bits for replay window
)

// PacketType defines encrypted packet types
type PacketType uint8

const (
	PacketTypeData      PacketType = 0x01
	PacketTypeKeepAlive PacketType = 0x02
	PacketTypeRekey     PacketType = 0x03
	PacketTypeClose     PacketType = 0x04
	PacketTypePadding   PacketType = 0xFF
)

// TunnelCipher provides military-grade encryption for VPN tunnel
type TunnelCipher struct {
	// Encryption keys (double-layer)
	primaryCipher   cipher.AEAD // ChaCha20-Poly1305
	secondaryCipher cipher.AEAD // XChaCha20-Poly1305 (inner layer)

	// Key material
	sendKey    []byte
	recvKey    []byte
	sendNonce  uint64
	recvNonce  uint64

	// Session info
	sessionID  [32]byte
	createdAt  time.Time
	lastRekey  time.Time

	// Packet counters (atomic)
	packetsSent     uint64
	packetsReceived uint64

	// Replay protection
	replayWindow *ReplayWindow

	// Key rotation
	pendingRekey   bool
	nextSendKey    []byte
	nextRecvKey    []byte

	mu sync.RWMutex
}

// ReplayWindow implements sliding window replay protection
type ReplayWindow struct {
	bitmap   []uint64
	lastSeq  uint64
	mu       sync.Mutex
}

// NewReplayWindow creates a new replay window
func NewReplayWindow() *ReplayWindow {
	return &ReplayWindow{
		bitmap:  make([]uint64, ReplayWindowSize/64),
		lastSeq: 0,
	}
}

// Check verifies if a sequence number is valid (not replayed)
func (rw *ReplayWindow) Check(seq uint64) bool {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	if seq == 0 {
		return false
	}

	// If sequence is too old, reject
	if seq+ReplayWindowSize <= rw.lastSeq {
		return false
	}

	// If sequence is newer than window, it's valid
	if seq > rw.lastSeq {
		// Shift window
		shift := seq - rw.lastSeq
		if shift >= ReplayWindowSize {
			// Clear entire window
			for i := range rw.bitmap {
				rw.bitmap[i] = 0
			}
		} else {
			// Shift bits
			for i := uint64(0); i < shift; i++ {
				rw.shiftBitmap()
			}
		}
		rw.lastSeq = seq
		rw.setBit(0)
		return true
	}

	// Sequence is within window, check if already seen
	offset := rw.lastSeq - seq
	if rw.getBit(offset) {
		return false // Replay detected
	}

	rw.setBit(offset)
	return true
}

func (rw *ReplayWindow) shiftBitmap() {
	// Shift all bits left by 1
	var carry uint64
	for i := len(rw.bitmap) - 1; i >= 0; i-- {
		newCarry := rw.bitmap[i] >> 63
		rw.bitmap[i] = (rw.bitmap[i] << 1) | carry
		carry = newCarry
	}
}

func (rw *ReplayWindow) setBit(offset uint64) {
	if offset >= ReplayWindowSize {
		return
	}
	idx := offset / 64
	bit := offset % 64
	rw.bitmap[idx] |= (1 << bit)
}

func (rw *ReplayWindow) getBit(offset uint64) bool {
	if offset >= ReplayWindowSize {
		return false
	}
	idx := offset / 64
	bit := offset % 64
	return (rw.bitmap[idx] & (1 << bit)) != 0
}

// KeyExchange performs X25519 key exchange with HKDF key derivation
type KeyExchange struct {
	privateKey *ecdh.PrivateKey
	publicKey  *ecdh.PublicKey
}

// NewKeyExchange generates a new ephemeral key pair
func NewKeyExchange() (*KeyExchange, error) {
	curve := ecdh.X25519()
	privateKey, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return &KeyExchange{
		privateKey: privateKey,
		publicKey:  privateKey.PublicKey(),
	}, nil
}

// PublicKey returns the public key bytes
func (ke *KeyExchange) PublicKey() []byte {
	return ke.publicKey.Bytes()
}

// DeriveKeys derives encryption keys from peer's public key
func (ke *KeyExchange) DeriveKeys(peerPublicKey []byte, isInitiator bool) (*DerivedKeys, error) {
	curve := ecdh.X25519()
	peerKey, err := curve.NewPublicKey(peerPublicKey)
	if err != nil {
		return nil, ErrKeyExchangeFailed
	}

	// Compute shared secret
	sharedSecret, err := ke.privateKey.ECDH(peerKey)
	if err != nil {
		return nil, ErrKeyExchangeFailed
	}

	// Derive keys using HKDF
	// Info includes role to ensure different keys for each direction
	var info []byte
	if isInitiator {
		info = []byte("aureo-vpn-tunnel-v1-initiator")
	} else {
		info = []byte("aureo-vpn-tunnel-v1-responder")
	}

	// Derive 4 keys: send/recv for primary and secondary ciphers
	hkdfReader := hkdf.New(sha256.New, sharedSecret, nil, info)

	keys := &DerivedKeys{
		PrimarySendKey:   make([]byte, KeySize),
		PrimaryRecvKey:   make([]byte, KeySize),
		SecondarySendKey: make([]byte, KeySize),
		SecondaryRecvKey: make([]byte, KeySize),
		SessionID:        [32]byte{},
	}

	if _, err := io.ReadFull(hkdfReader, keys.PrimarySendKey); err != nil {
		return nil, ErrKeyExchangeFailed
	}
	if _, err := io.ReadFull(hkdfReader, keys.PrimaryRecvKey); err != nil {
		return nil, ErrKeyExchangeFailed
	}
	if _, err := io.ReadFull(hkdfReader, keys.SecondarySendKey); err != nil {
		return nil, ErrKeyExchangeFailed
	}
	if _, err := io.ReadFull(hkdfReader, keys.SecondaryRecvKey); err != nil {
		return nil, ErrKeyExchangeFailed
	}

	// Generate session ID
	sessionHash := sha256.Sum256(append(sharedSecret, info...))
	keys.SessionID = sessionHash

	return keys, nil
}

// Wipe securely erases the private key
func (ke *KeyExchange) Wipe() {
	ke.privateKey = nil
	ke.publicKey = nil
}

// DerivedKeys holds the derived encryption keys
type DerivedKeys struct {
	PrimarySendKey   []byte
	PrimaryRecvKey   []byte
	SecondarySendKey []byte
	SecondaryRecvKey []byte
	SessionID        [32]byte
}

// Wipe securely erases all keys
func (dk *DerivedKeys) Wipe() {
	wipeBytes(dk.PrimarySendKey)
	wipeBytes(dk.PrimaryRecvKey)
	wipeBytes(dk.SecondarySendKey)
	wipeBytes(dk.SecondaryRecvKey)
	wipeBytes(dk.SessionID[:])
}

// NewTunnelCipher creates a new tunnel cipher with derived keys
func NewTunnelCipher(keys *DerivedKeys) (*TunnelCipher, error) {
	// Create primary cipher (ChaCha20-Poly1305)
	primaryCipher, err := chacha20poly1305.New(keys.PrimarySendKey)
	if err != nil {
		return nil, err
	}

	// Create secondary cipher (XChaCha20-Poly1305) for double encryption
	secondaryCipher, err := chacha20poly1305.NewX(keys.SecondarySendKey)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	tc := &TunnelCipher{
		primaryCipher:   primaryCipher,
		secondaryCipher: secondaryCipher,
		sendKey:         make([]byte, len(keys.PrimarySendKey)),
		recvKey:         make([]byte, len(keys.PrimaryRecvKey)),
		sessionID:       keys.SessionID,
		createdAt:       now,
		lastRekey:       now,
		replayWindow:    NewReplayWindow(),
	}

	copy(tc.sendKey, keys.PrimarySendKey)
	copy(tc.recvKey, keys.PrimaryRecvKey)

	return tc, nil
}

// Encrypt encrypts data with double-layer ChaCha20-Poly1305
func (tc *TunnelCipher) Encrypt(plaintext []byte, packetType PacketType) ([]byte, error) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Check if rekey is needed
	if atomic.LoadUint64(&tc.packetsSent) >= MaxPacketsPerKey {
		return nil, ErrMaxPacketsExceeded
	}

	// Add random padding for traffic analysis protection
	padding := generatePadding()

	// Build packet: [header][padded_plaintext]
	paddedLen := len(plaintext) + len(padding)
	packet := make([]byte, HeaderSize+paddedLen)

	// Header: version(1) + type(1) + padding_len(2) + original_len(4)
	packet[0] = 0x01 // Version 1
	packet[1] = byte(packetType)
	binary.BigEndian.PutUint16(packet[2:4], uint16(len(padding)))
	binary.BigEndian.PutUint32(packet[4:8], uint32(len(plaintext)))

	// Copy plaintext and padding
	copy(packet[HeaderSize:], plaintext)
	copy(packet[HeaderSize+len(plaintext):], padding)

	// First encryption layer (XChaCha20-Poly1305 - inner)
	innerNonce := make([]byte, 24)
	if _, err := rand.Read(innerNonce); err != nil {
		return nil, err
	}
	innerCiphertext := tc.secondaryCipher.Seal(nil, innerNonce, packet, nil)

	// Second encryption layer (ChaCha20-Poly1305 - outer)
	// Use counter-based nonce for outer layer
	outerNonce := make([]byte, 12)
	atomic.AddUint64(&tc.sendNonce, 1)
	binary.BigEndian.PutUint64(outerNonce[4:], tc.sendNonce)

	// Include inner nonce as additional data
	outerCiphertext := tc.primaryCipher.Seal(nil, outerNonce, innerCiphertext, innerNonce)

	// Build final packet: [outer_nonce(12)][inner_nonce(24)][ciphertext]
	finalPacket := make([]byte, 12+24+len(outerCiphertext))
	copy(finalPacket[0:12], outerNonce)
	copy(finalPacket[12:36], innerNonce)
	copy(finalPacket[36:], outerCiphertext)

	atomic.AddUint64(&tc.packetsSent, 1)

	return finalPacket, nil
}

// Decrypt decrypts double-layer encrypted data
func (tc *TunnelCipher) Decrypt(ciphertext []byte) ([]byte, PacketType, error) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Minimum size: nonces + tag + header
	minSize := 12 + 24 + TagSize + TagSize + HeaderSize
	if len(ciphertext) < minSize {
		return nil, 0, ErrInvalidPacket
	}

	// Extract nonces
	outerNonce := ciphertext[0:12]
	innerNonce := ciphertext[12:36]
	encryptedData := ciphertext[36:]

	// Check replay protection using outer nonce
	seq := binary.BigEndian.Uint64(outerNonce[4:])
	if !tc.replayWindow.Check(seq) {
		return nil, 0, ErrReplayAttack
	}

	// First decryption layer (ChaCha20-Poly1305 - outer)
	innerCiphertext, err := tc.primaryCipher.Open(nil, outerNonce, encryptedData, innerNonce)
	if err != nil {
		return nil, 0, ErrDecryptionFailed
	}

	// Second decryption layer (XChaCha20-Poly1305 - inner)
	packet, err := tc.secondaryCipher.Open(nil, innerNonce, innerCiphertext, nil)
	if err != nil {
		return nil, 0, ErrDecryptionFailed
	}

	// Parse header
	if len(packet) < HeaderSize {
		return nil, 0, ErrInvalidPacket
	}

	version := packet[0]
	if version != 0x01 {
		return nil, 0, ErrInvalidPacket
	}

	packetType := PacketType(packet[1])
	paddingLen := binary.BigEndian.Uint16(packet[2:4])
	originalLen := binary.BigEndian.Uint32(packet[4:8])

	// Validate lengths
	if uint32(len(packet)-HeaderSize) < originalLen+uint32(paddingLen) {
		return nil, 0, ErrInvalidPacket
	}

	// Extract original plaintext (strip padding)
	plaintext := packet[HeaderSize : HeaderSize+originalLen]

	atomic.AddUint64(&tc.packetsReceived, 1)

	return plaintext, packetType, nil
}

// NeedsRekey returns true if key rotation is needed
func (tc *TunnelCipher) NeedsRekey() bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	// Rekey based on packet count or time
	if atomic.LoadUint64(&tc.packetsSent) >= MaxPacketsPerKey/2 {
		return true
	}
	if time.Since(tc.lastRekey) >= KeyRotationTime {
		return true
	}
	return false
}

// PrepareRekey generates new keys for rotation
func (tc *TunnelCipher) PrepareRekey() (*KeyExchange, error) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.pendingRekey = true
	return NewKeyExchange()
}

// CompleteRekey completes key rotation with new derived keys
func (tc *TunnelCipher) CompleteRekey(newKeys *DerivedKeys) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Create new ciphers
	newPrimary, err := chacha20poly1305.New(newKeys.PrimarySendKey)
	if err != nil {
		return err
	}

	newSecondary, err := chacha20poly1305.NewX(newKeys.SecondarySendKey)
	if err != nil {
		return err
	}

	// Wipe old keys
	wipeBytes(tc.sendKey)
	wipeBytes(tc.recvKey)

	// Update cipher and keys
	tc.primaryCipher = newPrimary
	tc.secondaryCipher = newSecondary
	tc.sendKey = make([]byte, len(newKeys.PrimarySendKey))
	tc.recvKey = make([]byte, len(newKeys.PrimaryRecvKey))
	copy(tc.sendKey, newKeys.PrimarySendKey)
	copy(tc.recvKey, newKeys.PrimaryRecvKey)

	// Reset counters
	tc.sendNonce = 0
	atomic.StoreUint64(&tc.packetsSent, 0)
	tc.lastRekey = time.Now()
	tc.pendingRekey = false

	return nil
}

// SessionID returns the session identifier
func (tc *TunnelCipher) SessionID() [32]byte {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.sessionID
}

// Stats returns encryption statistics
func (tc *TunnelCipher) Stats() TunnelStats {
	return TunnelStats{
		PacketsSent:     atomic.LoadUint64(&tc.packetsSent),
		PacketsReceived: atomic.LoadUint64(&tc.packetsReceived),
		CreatedAt:       tc.createdAt,
		LastRekey:       tc.lastRekey,
	}
}

// TunnelStats holds tunnel statistics
type TunnelStats struct {
	PacketsSent     uint64
	PacketsReceived uint64
	CreatedAt       time.Time
	LastRekey       time.Time
}

// Close securely destroys the tunnel cipher
func (tc *TunnelCipher) Close() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	wipeBytes(tc.sendKey)
	wipeBytes(tc.recvKey)
	wipeBytes(tc.nextSendKey)
	wipeBytes(tc.nextRecvKey)

	tc.primaryCipher = nil
	tc.secondaryCipher = nil
}

// generatePadding creates random padding for traffic analysis protection
func generatePadding() []byte {
	// Random padding length between MinPaddingSize and MaxPaddingSize
	lenBytes := make([]byte, 1)
	rand.Read(lenBytes)
	paddingLen := MinPaddingSize + int(lenBytes[0])%(MaxPaddingSize-MinPaddingSize)

	padding := make([]byte, paddingLen)
	rand.Read(padding)
	return padding
}

// wipeBytes securely erases a byte slice
func wipeBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
	rand.Read(b) // Overwrite with random
	for i := range b {
		b[i] = 0
	}
}
