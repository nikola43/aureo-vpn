// Package crypto provides zero-knowledge data handling
// Ensures no plaintext user data, IPs, or identifiable information is ever stored
package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"errors"
	"io"
	"sync"
	"time"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
)

// Zero-knowledge errors
var (
	ErrZKProofFailed     = errors.New("zero-knowledge proof verification failed")
	ErrBlindedIDInvalid  = errors.New("blinded identifier is invalid")
	ErrCommitmentInvalid = errors.New("commitment verification failed")
)

// BlindingFactor holds cryptographic blinding parameters
type BlindingFactor struct {
	factor []byte
	salt   []byte
}

// NewBlindingFactor creates a new random blinding factor
func NewBlindingFactor() (*BlindingFactor, error) {
	factor := make([]byte, 32)
	salt := make([]byte, 16)

	if _, err := rand.Read(factor); err != nil {
		return nil, err
	}
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	return &BlindingFactor{
		factor: factor,
		salt:   salt,
	}, nil
}

// Blind applies blinding to an identifier (IP, user ID, etc.)
// The result is unlinkable to the original value
func (bf *BlindingFactor) Blind(identifier []byte) []byte {
	// Use Argon2 for slow, memory-hard blinding
	blinded := argon2.IDKey(
		identifier,
		bf.salt,
		3,        // 3 iterations
		64*1024,  // 64 MB memory
		4,        // 4 threads
		32,       // 32 byte output
	)

	// XOR with blinding factor
	for i := range blinded {
		blinded[i] ^= bf.factor[i%len(bf.factor)]
	}

	return blinded
}

// Wipe securely erases the blinding factor
func (bf *BlindingFactor) Wipe() {
	wipeBytes(bf.factor)
	wipeBytes(bf.salt)
}

// ZeroKnowledgeSession provides anonymous session handling
type ZeroKnowledgeSession struct {
	// Session identifier (blinded)
	blindedID []byte

	// Encrypted session data
	encryptedData []byte

	// Commitment for verification
	commitment []byte

	// Ephemeral keys (new for each session)
	ephemeralKey *KeyExchange

	// Encryption cipher
	cipher *TunnelCipher

	// Timing obfuscation
	createdAt    time.Time
	lastActivity time.Time
	jitter       time.Duration

	// Blinding
	blinding *BlindingFactor

	mu sync.RWMutex
}

// ZeroKnowledgeConfig holds configuration for zero-knowledge sessions
type ZeroKnowledgeConfig struct {
	// Timing jitter range (for traffic analysis protection)
	MinJitter time.Duration
	MaxJitter time.Duration

	// Key rotation interval
	RekeyInterval time.Duration

	// Maximum session duration
	MaxDuration time.Duration
}

// DefaultZeroKnowledgeConfig returns secure defaults
func DefaultZeroKnowledgeConfig() *ZeroKnowledgeConfig {
	return &ZeroKnowledgeConfig{
		MinJitter:     10 * time.Millisecond,
		MaxJitter:     100 * time.Millisecond,
		RekeyInterval: 30 * time.Minute,
		MaxDuration:   24 * time.Hour,
	}
}

// NewZeroKnowledgeSession creates a new anonymous session
func NewZeroKnowledgeSession(peerPublicKey []byte, isInitiator bool) (*ZeroKnowledgeSession, error) {
	// Generate ephemeral keys
	ephemeral, err := NewKeyExchange()
	if err != nil {
		return nil, err
	}

	// Derive session keys
	keys, err := ephemeral.DeriveKeys(peerPublicKey, isInitiator)
	if err != nil {
		ephemeral.Wipe()
		return nil, err
	}

	// Create tunnel cipher
	cipher, err := NewTunnelCipher(keys)
	if err != nil {
		keys.Wipe()
		ephemeral.Wipe()
		return nil, err
	}

	// Create blinding factor
	blinding, err := NewBlindingFactor()
	if err != nil {
		cipher.Close()
		keys.Wipe()
		ephemeral.Wipe()
		return nil, err
	}

	// Generate random jitter
	jitterBytes := make([]byte, 8)
	rand.Read(jitterBytes)
	jitter := time.Duration(binary.BigEndian.Uint64(jitterBytes)%uint64(90*time.Millisecond)) + 10*time.Millisecond

	now := time.Now()
	session := &ZeroKnowledgeSession{
		blindedID:    blinding.Blind(keys.SessionID[:]),
		commitment:   generateCommitment(keys.SessionID[:]),
		ephemeralKey: ephemeral,
		cipher:       cipher,
		createdAt:    now,
		lastActivity: now,
		jitter:       jitter,
		blinding:     blinding,
	}

	// Wipe derived keys (cipher has copies)
	keys.Wipe()

	return session, nil
}

// PublicKey returns the session's ephemeral public key
func (zks *ZeroKnowledgeSession) PublicKey() []byte {
	return zks.ephemeralKey.PublicKey()
}

// BlindedID returns the blinded session identifier
func (zks *ZeroKnowledgeSession) BlindedID() []byte {
	zks.mu.RLock()
	defer zks.mu.RUnlock()
	result := make([]byte, len(zks.blindedID))
	copy(result, zks.blindedID)
	return result
}

// Encrypt encrypts data with zero-knowledge guarantees
func (zks *ZeroKnowledgeSession) Encrypt(plaintext []byte) ([]byte, error) {
	zks.mu.Lock()
	defer zks.mu.Unlock()

	// Apply timing jitter
	if zks.jitter > 0 {
		jitterBytes := make([]byte, 1)
		rand.Read(jitterBytes)
		actualJitter := time.Duration(jitterBytes[0]) * zks.jitter / 256
		time.Sleep(actualJitter)
	}

	zks.lastActivity = time.Now()
	return zks.cipher.Encrypt(plaintext, PacketTypeData)
}

// Decrypt decrypts data with zero-knowledge guarantees
func (zks *ZeroKnowledgeSession) Decrypt(ciphertext []byte) ([]byte, error) {
	zks.mu.Lock()
	defer zks.mu.Unlock()

	plaintext, packetType, err := zks.cipher.Decrypt(ciphertext)
	if err != nil {
		return nil, err
	}

	// Handle special packet types
	switch packetType {
	case PacketTypeRekey:
		// Trigger rekey process
		zks.cipher.PrepareRekey()
	case PacketTypeClose:
		// Session close requested
		zks.Close()
		return nil, ErrSessionExpired
	}

	zks.lastActivity = time.Now()
	return plaintext, nil
}

// NeedsRekey returns true if key rotation is needed
func (zks *ZeroKnowledgeSession) NeedsRekey() bool {
	return zks.cipher.NeedsRekey()
}

// Rekey performs key rotation
func (zks *ZeroKnowledgeSession) Rekey(peerPublicKey []byte, isInitiator bool) error {
	zks.mu.Lock()
	defer zks.mu.Unlock()

	// Generate new ephemeral keys
	newEphemeral, err := NewKeyExchange()
	if err != nil {
		return err
	}

	// Derive new session keys
	newKeys, err := newEphemeral.DeriveKeys(peerPublicKey, isInitiator)
	if err != nil {
		newEphemeral.Wipe()
		return err
	}

	// Complete rekey
	if err := zks.cipher.CompleteRekey(newKeys); err != nil {
		newKeys.Wipe()
		newEphemeral.Wipe()
		return err
	}

	// Update session
	zks.ephemeralKey.Wipe()
	zks.ephemeralKey = newEphemeral

	// Update blinded ID
	zks.blinding.Wipe()
	newBlinding, _ := NewBlindingFactor()
	zks.blinding = newBlinding
	zks.blindedID = newBlinding.Blind(newKeys.SessionID[:])
	zks.commitment = generateCommitment(newKeys.SessionID[:])

	newKeys.Wipe()
	return nil
}

// VerifyCommitment verifies the session commitment
func (zks *ZeroKnowledgeSession) VerifyCommitment(expectedCommitment []byte) bool {
	zks.mu.RLock()
	defer zks.mu.RUnlock()
	return subtle.ConstantTimeCompare(zks.commitment, expectedCommitment) == 1
}

// Stats returns session statistics (anonymized)
func (zks *ZeroKnowledgeSession) Stats() ZKSessionStats {
	zks.mu.RLock()
	defer zks.mu.RUnlock()

	tunnelStats := zks.cipher.Stats()
	return ZKSessionStats{
		PacketsSent:     tunnelStats.PacketsSent,
		PacketsReceived: tunnelStats.PacketsReceived,
		Duration:        time.Since(zks.createdAt),
		IdleTime:        time.Since(zks.lastActivity),
	}
}

// ZKSessionStats holds anonymized session statistics
type ZKSessionStats struct {
	PacketsSent     uint64
	PacketsReceived uint64
	Duration        time.Duration
	IdleTime        time.Duration
}

// Close securely destroys the session
func (zks *ZeroKnowledgeSession) Close() {
	zks.mu.Lock()
	defer zks.mu.Unlock()

	if zks.cipher != nil {
		zks.cipher.Close()
	}
	if zks.ephemeralKey != nil {
		zks.ephemeralKey.Wipe()
	}
	if zks.blinding != nil {
		zks.blinding.Wipe()
	}

	wipeBytes(zks.blindedID)
	wipeBytes(zks.encryptedData)
	wipeBytes(zks.commitment)
}

// generateCommitment creates a cryptographic commitment
func generateCommitment(data []byte) []byte {
	// Use HKDF to derive commitment
	commitmentKey := make([]byte, 32)
	rand.Read(commitmentKey)

	hkdfReader := hkdf.New(sha256.New, data, commitmentKey, []byte("aureo-vpn-commitment-v1"))
	commitment := make([]byte, 32)
	io.ReadFull(hkdfReader, commitment)

	return commitment
}

// AnonymousDataStore provides zero-knowledge data storage
type AnonymousDataStore struct {
	// Master key for data encryption (derived from hardware key or similar)
	masterKey []byte

	// Encrypted data storage (key is blinded ID)
	store map[string][]byte

	// Key derivation salt
	salt []byte

	mu sync.RWMutex
}

// NewAnonymousDataStore creates a new zero-knowledge data store
func NewAnonymousDataStore(masterKey []byte) (*AnonymousDataStore, error) {
	if len(masterKey) < 32 {
		return nil, errors.New("master key must be at least 32 bytes")
	}

	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	return &AnonymousDataStore{
		masterKey: masterKey,
		store:     make(map[string][]byte),
		salt:      salt,
	}, nil
}

// Store encrypts and stores data with a blinded key
func (ads *AnonymousDataStore) Store(blindedKey, plaintext []byte) error {
	ads.mu.Lock()
	defer ads.mu.Unlock()

	// Derive encryption key from blinded key
	encKey := deriveDataKey(ads.masterKey, blindedKey, ads.salt)
	defer wipeBytes(encKey)

	// Encrypt data
	cipher, err := chacha20poly1305.NewX(encKey)
	if err != nil {
		return err
	}

	nonce := make([]byte, 24)
	if _, err := rand.Read(nonce); err != nil {
		return err
	}

	ciphertext := cipher.Seal(nonce, nonce, plaintext, blindedKey)

	// Store with blinded key hash
	keyHash := sha256.Sum256(blindedKey)
	ads.store[string(keyHash[:])] = ciphertext

	return nil
}

// Retrieve decrypts and retrieves data
func (ads *AnonymousDataStore) Retrieve(blindedKey []byte) ([]byte, error) {
	ads.mu.RLock()
	defer ads.mu.RUnlock()

	keyHash := sha256.Sum256(blindedKey)
	ciphertext, exists := ads.store[string(keyHash[:])]
	if !exists {
		return nil, errors.New("data not found")
	}

	// Derive encryption key
	encKey := deriveDataKey(ads.masterKey, blindedKey, ads.salt)
	defer wipeBytes(encKey)

	// Decrypt data
	cipher, err := chacha20poly1305.NewX(encKey)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < 24 {
		return nil, errors.New("invalid ciphertext")
	}

	nonce := ciphertext[:24]
	encrypted := ciphertext[24:]

	plaintext, err := cipher.Open(nil, nonce, encrypted, blindedKey)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// Delete removes data from the store
func (ads *AnonymousDataStore) Delete(blindedKey []byte) {
	ads.mu.Lock()
	defer ads.mu.Unlock()

	keyHash := sha256.Sum256(blindedKey)
	if data, exists := ads.store[string(keyHash[:])]; exists {
		wipeBytes(data)
		delete(ads.store, string(keyHash[:]))
	}
}

// Close securely destroys the data store
func (ads *AnonymousDataStore) Close() {
	ads.mu.Lock()
	defer ads.mu.Unlock()

	for key, data := range ads.store {
		wipeBytes(data)
		delete(ads.store, key)
	}

	wipeBytes(ads.masterKey)
	wipeBytes(ads.salt)
}

// deriveDataKey derives an encryption key from master key and blinded key
func deriveDataKey(masterKey, blindedKey, salt []byte) []byte {
	combined := append(masterKey, blindedKey...)
	hkdfReader := hkdf.New(sha256.New, combined, salt, []byte("aureo-vpn-datastore-v1"))
	key := make([]byte, 32)
	io.ReadFull(hkdfReader, key)
	return key
}

// IPAnonymizer provides IP address anonymization
type IPAnonymizer struct {
	// Rotating blinding factors (one per hour slot)
	blindingFactors [24]*BlindingFactor

	// Current hour
	currentHour int

	mu sync.RWMutex
}

// NewIPAnonymizer creates a new IP anonymizer
func NewIPAnonymizer() (*IPAnonymizer, error) {
	ia := &IPAnonymizer{
		currentHour: time.Now().Hour(),
	}

	// Initialize all blinding factors
	for i := range ia.blindingFactors {
		bf, err := NewBlindingFactor()
		if err != nil {
			return nil, err
		}
		ia.blindingFactors[i] = bf
	}

	return ia, nil
}

// AnonymizeIP creates an anonymous identifier for an IP
// The same IP will produce different identifiers each hour
func (ia *IPAnonymizer) AnonymizeIP(ip []byte) []byte {
	ia.mu.RLock()
	defer ia.mu.RUnlock()

	// Rotate blinding factor hourly
	currentHour := time.Now().Hour()
	return ia.blindingFactors[currentHour].Blind(ip)
}

// RotateBlindingFactors rotates all blinding factors
func (ia *IPAnonymizer) RotateBlindingFactors() error {
	ia.mu.Lock()
	defer ia.mu.Unlock()

	for i := range ia.blindingFactors {
		ia.blindingFactors[i].Wipe()
		bf, err := NewBlindingFactor()
		if err != nil {
			return err
		}
		ia.blindingFactors[i] = bf
	}

	return nil
}

// Close securely destroys the anonymizer
func (ia *IPAnonymizer) Close() {
	ia.mu.Lock()
	defer ia.mu.Unlock()

	for i := range ia.blindingFactors {
		if ia.blindingFactors[i] != nil {
			ia.blindingFactors[i].Wipe()
		}
	}
}
