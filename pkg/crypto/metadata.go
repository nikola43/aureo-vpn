// Package crypto provides encrypted metadata handling
// Ensures all packet metadata is encrypted and unlinkable
package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
	"sync"
	"time"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
)

// Metadata errors
var (
	ErrMetadataEncryptFailed = errors.New("metadata encryption failed")
	ErrMetadataDecryptFailed = errors.New("metadata decryption failed")
	ErrInvalidMetadata       = errors.New("invalid metadata format")
)

// MetadataType defines the type of metadata
type MetadataType uint8

const (
	MetadataTypeRouting MetadataType = iota
	MetadataTypeSession
	MetadataTypeTiming
	MetadataTypeNode
	MetadataTypeUser
)

// EncryptedMetadata holds encrypted packet metadata
type EncryptedMetadata struct {
	// Encrypted routing information (onion-style layers)
	RoutingLayers [][]byte

	// Encrypted session identifier
	SessionID []byte

	// Encrypted timestamp (coarse-grained for privacy)
	Timestamp []byte

	// Encrypted hop counter
	HopCount []byte

	// Random padding to normalize size
	Padding []byte

	// MAC for integrity
	MAC []byte
}

// MetadataEncryptor provides military-grade metadata encryption
type MetadataEncryptor struct {
	// Layered encryption keys (one per hop)
	layerKeys [][]byte

	// Session key for session metadata
	sessionKey []byte

	// Timing key for timestamp encryption
	timingKey []byte

	// AEAD cipher for metadata
	aeadCipher aeadCipherInterface

	// Epoch for key rotation
	epoch     uint64
	epochTime time.Time

	// Key derivation parameters
	salt []byte
	info []byte

	mu sync.RWMutex
}

// aeadCipherInterface defines AEAD cipher operations
type aeadCipherInterface interface {
	Seal(dst, nonce, plaintext, additionalData []byte) []byte
	Open(dst, nonce, ciphertext, additionalData []byte) ([]byte, error)
	NonceSize() int
	Overhead() int
}

// MetadataConfig holds configuration for metadata encryption
type MetadataConfig struct {
	// Number of encryption layers (for onion routing)
	LayerCount int

	// Key rotation interval
	RotationInterval time.Duration

	// Timestamp granularity (for privacy)
	TimestampGranularity time.Duration

	// Maximum metadata size
	MaxMetadataSize int
}

// DefaultMetadataConfig returns secure defaults
func DefaultMetadataConfig() *MetadataConfig {
	return &MetadataConfig{
		LayerCount:           3,
		RotationInterval:     1 * time.Hour,
		TimestampGranularity: 1 * time.Minute,
		MaxMetadataSize:      512,
	}
}

// NewMetadataEncryptor creates a new metadata encryptor
func NewMetadataEncryptor(masterKey []byte, config *MetadataConfig) (*MetadataEncryptor, error) {
	if len(masterKey) < 32 {
		return nil, errors.New("master key must be at least 32 bytes")
	}

	if config == nil {
		config = DefaultMetadataConfig()
	}

	// Generate salt
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	me := &MetadataEncryptor{
		salt:      salt,
		info:      []byte("aureo-vpn-metadata-v1"),
		epochTime: time.Now(),
	}

	// Derive layer keys
	me.layerKeys = make([][]byte, config.LayerCount)
	for i := 0; i < config.LayerCount; i++ {
		layerInfo := append(me.info, byte(i))
		hkdfReader := hkdf.New(sha256.New, masterKey, salt, layerInfo)
		key := make([]byte, 32)
		if _, err := io.ReadFull(hkdfReader, key); err != nil {
			return nil, err
		}
		me.layerKeys[i] = key
	}

	// Derive session key
	sessionInfo := append(me.info, []byte("-session")...)
	hkdfReader := hkdf.New(sha256.New, masterKey, salt, sessionInfo)
	me.sessionKey = make([]byte, 32)
	if _, err := io.ReadFull(hkdfReader, me.sessionKey); err != nil {
		return nil, err
	}

	// Derive timing key
	timingInfo := append(me.info, []byte("-timing")...)
	hkdfReader = hkdf.New(sha256.New, masterKey, salt, timingInfo)
	me.timingKey = make([]byte, 32)
	if _, err := io.ReadFull(hkdfReader, me.timingKey); err != nil {
		return nil, err
	}

	// Create AEAD cipher
	aeadCipher, err := chacha20poly1305.NewX(me.sessionKey)
	if err != nil {
		return nil, err
	}
	me.aeadCipher = aeadCipher

	return me, nil
}

// EncryptRouting encrypts routing information with onion layers
func (me *MetadataEncryptor) EncryptRouting(destinations [][]byte) ([][]byte, error) {
	me.mu.RLock()
	defer me.mu.RUnlock()

	if len(destinations) > len(me.layerKeys) {
		return nil, errors.New("too many routing hops")
	}

	layers := make([][]byte, len(destinations))

	// Encrypt each destination with its layer key (onion style)
	for i, dest := range destinations {
		cipher, err := chacha20poly1305.NewX(me.layerKeys[i])
		if err != nil {
			return nil, err
		}

		nonce := make([]byte, 24)
		if _, err := rand.Read(nonce); err != nil {
			return nil, err
		}

		// Add layer index to prevent reordering
		layerData := make([]byte, 1+len(dest))
		layerData[0] = byte(i)
		copy(layerData[1:], dest)

		encrypted := cipher.Seal(nonce, nonce, layerData, nil)
		layers[i] = encrypted
	}

	return layers, nil
}

// DecryptRouting decrypts one routing layer
func (me *MetadataEncryptor) DecryptRouting(layer []byte, layerIndex int) ([]byte, error) {
	me.mu.RLock()
	defer me.mu.RUnlock()

	if layerIndex >= len(me.layerKeys) {
		return nil, ErrInvalidMetadata
	}

	cipher, err := chacha20poly1305.NewX(me.layerKeys[layerIndex])
	if err != nil {
		return nil, err
	}

	if len(layer) < 24 {
		return nil, ErrInvalidMetadata
	}

	nonce := layer[:24]
	ciphertext := layer[24:]

	plaintext, err := cipher.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrMetadataDecryptFailed
	}

	// Verify layer index
	if len(plaintext) < 1 || plaintext[0] != byte(layerIndex) {
		return nil, ErrInvalidMetadata
	}

	return plaintext[1:], nil
}

// EncryptSessionID encrypts a session identifier
func (me *MetadataEncryptor) EncryptSessionID(sessionID []byte) ([]byte, error) {
	me.mu.RLock()
	defer me.mu.RUnlock()

	nonce := make([]byte, 24)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	return me.aeadCipher.Seal(nonce, nonce, sessionID, nil), nil
}

// DecryptSessionID decrypts a session identifier
func (me *MetadataEncryptor) DecryptSessionID(encrypted []byte) ([]byte, error) {
	me.mu.RLock()
	defer me.mu.RUnlock()

	if len(encrypted) < 24 {
		return nil, ErrInvalidMetadata
	}

	nonce := encrypted[:24]
	ciphertext := encrypted[24:]

	return me.aeadCipher.Open(nil, nonce, ciphertext, nil)
}

// EncryptTimestamp encrypts a timestamp with coarse granularity
func (me *MetadataEncryptor) EncryptTimestamp(t time.Time, granularity time.Duration) ([]byte, error) {
	me.mu.RLock()
	defer me.mu.RUnlock()

	// Truncate timestamp for privacy
	truncated := t.Truncate(granularity)
	unix := truncated.Unix()

	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, uint64(unix))

	cipher, err := chacha20poly1305.NewX(me.timingKey)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, 24)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	return cipher.Seal(nonce, nonce, data, nil), nil
}

// DecryptTimestamp decrypts an encrypted timestamp
func (me *MetadataEncryptor) DecryptTimestamp(encrypted []byte) (time.Time, error) {
	me.mu.RLock()
	defer me.mu.RUnlock()

	cipher, err := chacha20poly1305.NewX(me.timingKey)
	if err != nil {
		return time.Time{}, err
	}

	if len(encrypted) < 24 {
		return time.Time{}, ErrInvalidMetadata
	}

	nonce := encrypted[:24]
	ciphertext := encrypted[24:]

	data, err := cipher.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return time.Time{}, ErrMetadataDecryptFailed
	}

	if len(data) < 8 {
		return time.Time{}, ErrInvalidMetadata
	}

	unix := binary.BigEndian.Uint64(data)
	return time.Unix(int64(unix), 0), nil
}

// CreateEncryptedMetadata creates fully encrypted metadata
func (me *MetadataEncryptor) CreateEncryptedMetadata(
	routingDests [][]byte,
	sessionID []byte,
	timestamp time.Time,
	hopCount uint8,
) (*EncryptedMetadata, error) {

	// Encrypt routing layers
	routingLayers, err := me.EncryptRouting(routingDests)
	if err != nil {
		return nil, err
	}

	// Encrypt session ID
	encSessionID, err := me.EncryptSessionID(sessionID)
	if err != nil {
		return nil, err
	}

	// Encrypt timestamp with minute granularity
	encTimestamp, err := me.EncryptTimestamp(timestamp, time.Minute)
	if err != nil {
		return nil, err
	}

	// Encrypt hop count
	hopData := []byte{hopCount}
	encHopCount, err := me.EncryptSessionID(hopData) // Reuse session key
	if err != nil {
		return nil, err
	}

	// Add random padding
	paddingLen := 32 + (time.Now().UnixNano() % 64)
	padding := make([]byte, paddingLen)
	rand.Read(padding)

	// Calculate MAC over all encrypted data
	macData := make([]byte, 0)
	for _, layer := range routingLayers {
		macData = append(macData, layer...)
	}
	macData = append(macData, encSessionID...)
	macData = append(macData, encTimestamp...)
	macData = append(macData, encHopCount...)
	macData = append(macData, padding...)

	mac := sha256.Sum256(macData)

	return &EncryptedMetadata{
		RoutingLayers: routingLayers,
		SessionID:     encSessionID,
		Timestamp:     encTimestamp,
		HopCount:      encHopCount,
		Padding:       padding,
		MAC:           mac[:],
	}, nil
}

// Serialize serializes encrypted metadata to bytes
func (em *EncryptedMetadata) Serialize() ([]byte, error) {
	// Calculate total size
	size := 4 // Layer count
	for _, layer := range em.RoutingLayers {
		size += 4 + len(layer) // Length prefix + data
	}
	size += 4 + len(em.SessionID)
	size += 4 + len(em.Timestamp)
	size += 4 + len(em.HopCount)
	size += 4 + len(em.Padding)
	size += 4 + len(em.MAC)

	data := make([]byte, size)
	offset := 0

	// Write layer count
	binary.BigEndian.PutUint32(data[offset:], uint32(len(em.RoutingLayers)))
	offset += 4

	// Write layers
	for _, layer := range em.RoutingLayers {
		binary.BigEndian.PutUint32(data[offset:], uint32(len(layer)))
		offset += 4
		copy(data[offset:], layer)
		offset += len(layer)
	}

	// Write session ID
	binary.BigEndian.PutUint32(data[offset:], uint32(len(em.SessionID)))
	offset += 4
	copy(data[offset:], em.SessionID)
	offset += len(em.SessionID)

	// Write timestamp
	binary.BigEndian.PutUint32(data[offset:], uint32(len(em.Timestamp)))
	offset += 4
	copy(data[offset:], em.Timestamp)
	offset += len(em.Timestamp)

	// Write hop count
	binary.BigEndian.PutUint32(data[offset:], uint32(len(em.HopCount)))
	offset += 4
	copy(data[offset:], em.HopCount)
	offset += len(em.HopCount)

	// Write padding
	binary.BigEndian.PutUint32(data[offset:], uint32(len(em.Padding)))
	offset += 4
	copy(data[offset:], em.Padding)
	offset += len(em.Padding)

	// Write MAC
	binary.BigEndian.PutUint32(data[offset:], uint32(len(em.MAC)))
	offset += 4
	copy(data[offset:], em.MAC)

	return data, nil
}

// DeserializeMetadata deserializes encrypted metadata from bytes
func DeserializeMetadata(data []byte) (*EncryptedMetadata, error) {
	if len(data) < 4 {
		return nil, ErrInvalidMetadata
	}

	offset := 0
	em := &EncryptedMetadata{}

	// Read layer count
	layerCount := binary.BigEndian.Uint32(data[offset:])
	offset += 4

	if layerCount > 10 { // Sanity check
		return nil, ErrInvalidMetadata
	}

	// Read layers
	em.RoutingLayers = make([][]byte, layerCount)
	for i := uint32(0); i < layerCount; i++ {
		if offset+4 > len(data) {
			return nil, ErrInvalidMetadata
		}
		length := binary.BigEndian.Uint32(data[offset:])
		offset += 4

		if offset+int(length) > len(data) {
			return nil, ErrInvalidMetadata
		}
		em.RoutingLayers[i] = make([]byte, length)
		copy(em.RoutingLayers[i], data[offset:offset+int(length)])
		offset += int(length)
	}

	// Read session ID
	if offset+4 > len(data) {
		return nil, ErrInvalidMetadata
	}
	length := binary.BigEndian.Uint32(data[offset:])
	offset += 4
	if offset+int(length) > len(data) {
		return nil, ErrInvalidMetadata
	}
	em.SessionID = make([]byte, length)
	copy(em.SessionID, data[offset:offset+int(length)])
	offset += int(length)

	// Read timestamp
	if offset+4 > len(data) {
		return nil, ErrInvalidMetadata
	}
	length = binary.BigEndian.Uint32(data[offset:])
	offset += 4
	if offset+int(length) > len(data) {
		return nil, ErrInvalidMetadata
	}
	em.Timestamp = make([]byte, length)
	copy(em.Timestamp, data[offset:offset+int(length)])
	offset += int(length)

	// Read hop count
	if offset+4 > len(data) {
		return nil, ErrInvalidMetadata
	}
	length = binary.BigEndian.Uint32(data[offset:])
	offset += 4
	if offset+int(length) > len(data) {
		return nil, ErrInvalidMetadata
	}
	em.HopCount = make([]byte, length)
	copy(em.HopCount, data[offset:offset+int(length)])
	offset += int(length)

	// Read padding
	if offset+4 > len(data) {
		return nil, ErrInvalidMetadata
	}
	length = binary.BigEndian.Uint32(data[offset:])
	offset += 4
	if offset+int(length) > len(data) {
		return nil, ErrInvalidMetadata
	}
	em.Padding = make([]byte, length)
	copy(em.Padding, data[offset:offset+int(length)])
	offset += int(length)

	// Read MAC
	if offset+4 > len(data) {
		return nil, ErrInvalidMetadata
	}
	length = binary.BigEndian.Uint32(data[offset:])
	offset += 4
	if offset+int(length) > len(data) {
		return nil, ErrInvalidMetadata
	}
	em.MAC = make([]byte, length)
	copy(em.MAC, data[offset:offset+int(length)])

	return em, nil
}

// VerifyMAC verifies the integrity of encrypted metadata
func (em *EncryptedMetadata) VerifyMAC() bool {
	macData := make([]byte, 0)
	for _, layer := range em.RoutingLayers {
		macData = append(macData, layer...)
	}
	macData = append(macData, em.SessionID...)
	macData = append(macData, em.Timestamp...)
	macData = append(macData, em.HopCount...)
	macData = append(macData, em.Padding...)

	expectedMAC := sha256.Sum256(macData)

	if len(em.MAC) != 32 {
		return false
	}

	// Constant-time comparison
	result := byte(0)
	for i := 0; i < 32; i++ {
		result |= em.MAC[i] ^ expectedMAC[i]
	}
	return result == 0
}

// Wipe securely erases the metadata encryptor
func (me *MetadataEncryptor) Wipe() {
	me.mu.Lock()
	defer me.mu.Unlock()

	for i := range me.layerKeys {
		wipeBytes(me.layerKeys[i])
	}
	wipeBytes(me.sessionKey)
	wipeBytes(me.timingKey)
	wipeBytes(me.salt)
}

// AnonymousPacketHeader provides anonymous packet header handling
type AnonymousPacketHeader struct {
	// Version (1 byte)
	Version uint8

	// Packet type (1 byte) - encrypted
	PacketType []byte

	// Sequence number (8 bytes) - encrypted
	Sequence []byte

	// Timestamp (8 bytes) - encrypted with coarse granularity
	Timestamp []byte

	// Session binding (32 bytes) - blinded
	SessionBinding []byte

	// Padding to fixed size
	Padding []byte
}

// HeaderEncryptor encrypts packet headers
type HeaderEncryptor struct {
	key        []byte
	aeadCipher aeadCipherInterface
	mu         sync.RWMutex
}

// NewHeaderEncryptor creates a new header encryptor
func NewHeaderEncryptor(key []byte) (*HeaderEncryptor, error) {
	if len(key) < 32 {
		return nil, errors.New("key must be at least 32 bytes")
	}

	aeadCipher, err := chacha20poly1305.NewX(key[:32])
	if err != nil {
		return nil, err
	}

	return &HeaderEncryptor{
		key:        key[:32],
		aeadCipher: aeadCipher,
	}, nil
}

// EncryptHeader encrypts a packet header
func (he *HeaderEncryptor) EncryptHeader(
	packetType uint8,
	sequence uint64,
	timestamp time.Time,
	sessionBinding []byte,
) (*AnonymousPacketHeader, error) {
	he.mu.RLock()
	defer he.mu.RUnlock()

	// Encrypt packet type
	ptData := []byte{packetType}
	nonce := make([]byte, 24)
	rand.Read(nonce)
	encType := he.aeadCipher.Seal(nonce, nonce, ptData, nil)

	// Encrypt sequence
	seqData := make([]byte, 8)
	binary.BigEndian.PutUint64(seqData, sequence)
	rand.Read(nonce)
	encSeq := he.aeadCipher.Seal(nonce, nonce, seqData, nil)

	// Encrypt timestamp (truncated to minute for privacy)
	truncated := timestamp.Truncate(time.Minute)
	tsData := make([]byte, 8)
	binary.BigEndian.PutUint64(tsData, uint64(truncated.Unix()))
	rand.Read(nonce)
	encTs := he.aeadCipher.Seal(nonce, nonce, tsData, nil)

	// Blind session binding
	blindedSession := make([]byte, 32)
	if len(sessionBinding) >= 32 {
		// XOR with random for additional blinding
		randBytes := make([]byte, 32)
		rand.Read(randBytes)
		for i := 0; i < 32; i++ {
			blindedSession[i] = sessionBinding[i] ^ randBytes[i]
		}
	} else {
		rand.Read(blindedSession)
	}

	// Fixed size padding
	padding := make([]byte, 64)
	rand.Read(padding)

	return &AnonymousPacketHeader{
		Version:        1,
		PacketType:     encType,
		Sequence:       encSeq,
		Timestamp:      encTs,
		SessionBinding: blindedSession,
		Padding:        padding,
	}, nil
}

// DecryptHeader decrypts a packet header
func (he *HeaderEncryptor) DecryptHeader(header *AnonymousPacketHeader) (uint8, uint64, time.Time, error) {
	he.mu.RLock()
	defer he.mu.RUnlock()

	if header.Version != 1 {
		return 0, 0, time.Time{}, ErrInvalidMetadata
	}

	// Decrypt packet type
	if len(header.PacketType) < 24 {
		return 0, 0, time.Time{}, ErrInvalidMetadata
	}
	nonce := header.PacketType[:24]
	ct := header.PacketType[24:]
	ptData, err := he.aeadCipher.Open(nil, nonce, ct, nil)
	if err != nil {
		return 0, 0, time.Time{}, ErrMetadataDecryptFailed
	}
	packetType := ptData[0]

	// Decrypt sequence
	if len(header.Sequence) < 24 {
		return 0, 0, time.Time{}, ErrInvalidMetadata
	}
	nonce = header.Sequence[:24]
	ct = header.Sequence[24:]
	seqData, err := he.aeadCipher.Open(nil, nonce, ct, nil)
	if err != nil {
		return 0, 0, time.Time{}, ErrMetadataDecryptFailed
	}
	sequence := binary.BigEndian.Uint64(seqData)

	// Decrypt timestamp
	if len(header.Timestamp) < 24 {
		return 0, 0, time.Time{}, ErrInvalidMetadata
	}
	nonce = header.Timestamp[:24]
	ct = header.Timestamp[24:]
	tsData, err := he.aeadCipher.Open(nil, nonce, ct, nil)
	if err != nil {
		return 0, 0, time.Time{}, ErrMetadataDecryptFailed
	}
	unix := binary.BigEndian.Uint64(tsData)
	timestamp := time.Unix(int64(unix), 0)

	return packetType, sequence, timestamp, nil
}

// Wipe securely erases the header encryptor
func (he *HeaderEncryptor) Wipe() {
	he.mu.Lock()
	defer he.mu.Unlock()
	wipeBytes(he.key)
}
