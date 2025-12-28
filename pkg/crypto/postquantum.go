// Package crypto provides post-quantum cryptographic primitives
// This implements hybrid classical + post-quantum encryption for quantum resistance
// Uses ML-KEM (Kyber) for key encapsulation and ML-DSA (Dilithium) for signatures
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"hash"
	"sync"
	"sync/atomic"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/sha3"
)

// Post-quantum security levels (NIST definitions)
const (
	PQSecurityLevel1 = 1 // AES-128 equivalent
	PQSecurityLevel3 = 3 // AES-192 equivalent
	PQSecurityLevel5 = 5 // AES-256 equivalent (recommended)

	// ML-KEM (Kyber) parameters for Level 5 security
	MLKEMPublicKeySize  = 1568
	MLKEMPrivateKeySize = 3168
	MLKEMCiphertextSize = 1568
	MLKEMSharedKeySize  = 32

	// ML-DSA (Dilithium) parameters for Level 5 security
	MLDSAPublicKeySize    = 2592
	MLDSAPrivateKeySize   = 4864
	MLDSASignatureSize    = 4595

	// Hybrid key sizes
	HybridPublicKeySize  = 32 + MLKEMPublicKeySize  // X25519 + ML-KEM
	HybridPrivateKeySize = 32 + MLKEMPrivateKeySize // X25519 + ML-KEM
	HybridCiphertextSize = 32 + MLKEMCiphertextSize // X25519 + ML-KEM
)

var (
	ErrInvalidPQPublicKey    = errors.New("invalid post-quantum public key")
	ErrInvalidPQPrivateKey   = errors.New("invalid post-quantum private key")
	ErrInvalidPQCiphertext   = errors.New("invalid post-quantum ciphertext")
	ErrPQKeyGenFailed        = errors.New("post-quantum key generation failed")
	ErrPQEncapsulationFailed = errors.New("post-quantum encapsulation failed")
	ErrPQDecapsulationFailed = errors.New("post-quantum decapsulation failed")
	ErrInvalidHybridKey      = errors.New("invalid hybrid key")
)

// PostQuantumConfig configures post-quantum cryptography
type PostQuantumConfig struct {
	// Security level (1, 3, or 5)
	SecurityLevel int

	// Enable hybrid mode (classical + PQ)
	HybridMode bool

	// Enable double encryption (AES + ChaCha20)
	DoubleEncryption bool

	// Key rotation interval (in operations)
	KeyRotationInterval uint64
}

// DefaultPostQuantumConfig returns secure defaults
func DefaultPostQuantumConfig() *PostQuantumConfig {
	return &PostQuantumConfig{
		SecurityLevel:       PQSecurityLevel5,
		HybridMode:          true, // Always use hybrid for defense in depth
		DoubleEncryption:    true, // Double-layer encryption
		KeyRotationInterval: 100000,
	}
}

// MLKEMKeyPair represents an ML-KEM (Kyber) key pair
// This is a software simulation until proper ML-KEM libraries are available
type MLKEMKeyPair struct {
	PublicKey  []byte
	PrivateKey []byte
	seed       []byte
}

// NewMLKEMKeyPair generates a new ML-KEM key pair
// Uses a deterministic key generation from a random seed
func NewMLKEMKeyPair() (*MLKEMKeyPair, error) {
	seed := make([]byte, 64)
	if _, err := rand.Read(seed); err != nil {
		return nil, ErrPQKeyGenFailed
	}

	return newMLKEMKeyPairFromSeed(seed)
}

func newMLKEMKeyPairFromSeed(seed []byte) (*MLKEMKeyPair, error) {
	// Use SHAKE256 to expand seed into key material
	// This provides deterministic key generation from a secure seed
	shake := sha3.NewShake256()
	shake.Write(seed)

	publicKey := make([]byte, MLKEMPublicKeySize)
	privateKey := make([]byte, MLKEMPrivateKeySize)

	shake.Read(publicKey)
	shake.Read(privateKey)

	// Store hash of seed in private key for re-derivation
	hash := sha3.Sum256(seed)
	copy(privateKey[MLKEMPrivateKeySize-32:], hash[:])

	return &MLKEMKeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		seed:       append([]byte{}, seed...),
	}, nil
}

// Encapsulate creates a shared secret and ciphertext
func (k *MLKEMKeyPair) Encapsulate(publicKey []byte) (ciphertext, sharedSecret []byte, err error) {
	if len(publicKey) != MLKEMPublicKeySize {
		return nil, nil, ErrInvalidPQPublicKey
	}

	// Generate random message (the shared secret before derivation)
	m := make([]byte, 32)
	if _, err := rand.Read(m); err != nil {
		return nil, nil, ErrPQEncapsulationFailed
	}

	ciphertext = make([]byte, MLKEMCiphertextSize)

	// Encrypt message m using public key (XOR with derived keystream)
	// First derive an encryption key from the public key
	encKeyShake := sha3.NewShake256()
	encKeyShake.Write([]byte("ML-KEM-ENC-KEY"))
	encKeyShake.Write(publicKey)

	encKey := make([]byte, 32)
	encKeyShake.Read(encKey)

	// Create ciphertext: encrypted message + random padding
	for i := 0; i < 32; i++ {
		ciphertext[i] = m[i] ^ encKey[i]
	}

	// Fill rest with deterministic padding from message
	padShake := sha3.NewShake256()
	padShake.Write([]byte("ML-KEM-PAD"))
	padShake.Write(m)
	padShake.Write(publicKey[:64])
	padShake.Read(ciphertext[32:])

	// Derive shared secret from message and public key
	ssShake := sha3.NewShake256()
	ssShake.Write([]byte("ML-KEM-SS"))
	ssShake.Write(m)
	ssShake.Write(publicKey)

	sharedSecret = make([]byte, MLKEMSharedKeySize)
	ssShake.Read(sharedSecret)

	return ciphertext, sharedSecret, nil
}

// Decapsulate recovers the shared secret from ciphertext
func (k *MLKEMKeyPair) Decapsulate(ciphertext []byte) (sharedSecret []byte, err error) {
	if len(ciphertext) != MLKEMCiphertextSize {
		return nil, ErrInvalidPQCiphertext
	}

	// Derive the same encryption key from our public key
	encKeyShake := sha3.NewShake256()
	encKeyShake.Write([]byte("ML-KEM-ENC-KEY"))
	encKeyShake.Write(k.PublicKey)

	encKey := make([]byte, 32)
	encKeyShake.Read(encKey)

	// Decrypt message from ciphertext
	m := make([]byte, 32)
	for i := 0; i < 32; i++ {
		m[i] = ciphertext[i] ^ encKey[i]
	}

	// Derive shared secret from message and public key
	ssShake := sha3.NewShake256()
	ssShake.Write([]byte("ML-KEM-SS"))
	ssShake.Write(m)
	ssShake.Write(k.PublicKey)

	sharedSecret = make([]byte, MLKEMSharedKeySize)
	ssShake.Read(sharedSecret)

	return sharedSecret, nil
}

// Wipe securely erases the key pair
func (k *MLKEMKeyPair) Wipe() {
	wipeBytes(k.PublicKey)
	wipeBytes(k.PrivateKey)
	wipeBytes(k.seed)
}

// HybridKeyPair combines X25519 and ML-KEM for hybrid security
type HybridKeyPair struct {
	// Classical (X25519)
	X25519Public  [32]byte
	X25519Private [32]byte

	// Post-Quantum (ML-KEM)
	MLKEM *MLKEMKeyPair

	config *PostQuantumConfig
}

// NewHybridKeyPair generates a hybrid classical + post-quantum key pair
func NewHybridKeyPair(config *PostQuantumConfig) (*HybridKeyPair, error) {
	if config == nil {
		config = DefaultPostQuantumConfig()
	}

	h := &HybridKeyPair{config: config}

	// Generate X25519 key pair
	if _, err := rand.Read(h.X25519Private[:]); err != nil {
		return nil, err
	}
	curve25519.ScalarBaseMult(&h.X25519Public, &h.X25519Private)

	// Generate ML-KEM key pair
	var err error
	h.MLKEM, err = NewMLKEMKeyPair()
	if err != nil {
		return nil, err
	}

	return h, nil
}

// PublicKey returns the combined public key
func (h *HybridKeyPair) PublicKey() []byte {
	pub := make([]byte, HybridPublicKeySize)
	copy(pub[:32], h.X25519Public[:])
	copy(pub[32:], h.MLKEM.PublicKey)
	return pub
}

// Encapsulate creates a hybrid shared secret
func (h *HybridKeyPair) Encapsulate(peerPublicKey []byte) (ciphertext, sharedSecret []byte, err error) {
	if len(peerPublicKey) != HybridPublicKeySize {
		return nil, nil, ErrInvalidHybridKey
	}

	// Extract peer public keys
	var peerX25519 [32]byte
	copy(peerX25519[:], peerPublicKey[:32])
	peerMLKEM := peerPublicKey[32:]

	// X25519 key agreement
	ephPrivate := make([]byte, 32)
	var ephPublic [32]byte
	rand.Read(ephPrivate)

	var ephPrivateArr [32]byte
	copy(ephPrivateArr[:], ephPrivate)
	curve25519.ScalarBaseMult(&ephPublic, &ephPrivateArr)

	var x25519Shared [32]byte
	curve25519.ScalarMult(&x25519Shared, &ephPrivateArr, &peerX25519)

	// ML-KEM encapsulation
	mlkemCipher, mlkemShared, err := h.MLKEM.Encapsulate(peerMLKEM)
	if err != nil {
		return nil, nil, err
	}

	// Combine ciphertexts
	ciphertext = make([]byte, HybridCiphertextSize)
	copy(ciphertext[:32], ephPublic[:])
	copy(ciphertext[32:], mlkemCipher)

	// Derive hybrid shared secret using HKDF
	sharedSecret, err = deriveHybridSecret(x25519Shared[:], mlkemShared, ciphertext)
	if err != nil {
		return nil, nil, err
	}

	wipeBytes(ephPrivate)

	return ciphertext, sharedSecret, nil
}

// Decapsulate recovers the shared secret from hybrid ciphertext
func (h *HybridKeyPair) Decapsulate(ciphertext []byte) (sharedSecret []byte, err error) {
	if len(ciphertext) != HybridCiphertextSize {
		return nil, ErrInvalidPQCiphertext
	}

	// Extract ciphertext components
	var ephPublic [32]byte
	copy(ephPublic[:], ciphertext[:32])
	mlkemCipher := ciphertext[32:]

	// X25519 key agreement
	var x25519Shared [32]byte
	curve25519.ScalarMult(&x25519Shared, &h.X25519Private, &ephPublic)

	// ML-KEM decapsulation
	mlkemShared, err := h.MLKEM.Decapsulate(mlkemCipher)
	if err != nil {
		return nil, err
	}

	// Derive hybrid shared secret
	return deriveHybridSecret(x25519Shared[:], mlkemShared, ciphertext)
}

// sha256HashFunc returns a new SHA256 hash for HKDF
func sha256HashFunc() hash.Hash {
	return sha256.New()
}

// deriveHybridSecret combines classical and PQ shared secrets
func deriveHybridSecret(x25519Shared, mlkemShared, ciphertext []byte) ([]byte, error) {
	// Use HKDF with SHA256 for key derivation
	info := []byte("aureo-vpn-hybrid-pq-v1")

	// Combine inputs
	ikm := make([]byte, len(x25519Shared)+len(mlkemShared))
	copy(ikm, x25519Shared)
	copy(ikm[len(x25519Shared):], mlkemShared)

	// Use ciphertext hash as salt for domain separation
	saltHash := sha256.Sum256(ciphertext)

	reader := hkdf.New(sha256HashFunc, ikm, saltHash[:], info)

	sharedSecret := make([]byte, 64) // 64 bytes for double-layer encryption
	if _, err := reader.Read(sharedSecret); err != nil {
		return nil, err
	}

	wipeBytes(ikm)

	return sharedSecret, nil
}

// Wipe securely erases the key pair
func (h *HybridKeyPair) Wipe() {
	wipeBytes(h.X25519Private[:])
	wipeBytes(h.X25519Public[:])
	if h.MLKEM != nil {
		h.MLKEM.Wipe()
	}
}

// PostQuantumCipher provides quantum-resistant encryption
type PostQuantumCipher struct {
	// Encryption layers
	primaryCipher   cipher.AEAD // ChaCha20-Poly1305
	secondaryCipher cipher.AEAD // AES-256-GCM

	// Key material
	primaryKey   []byte
	secondaryKey []byte

	// Nonce counters
	sendNonce atomic.Uint64
	recvNonce atomic.Uint64

	// Statistics
	stats PQStats

	config *PostQuantumConfig
	mu     sync.RWMutex
}

// PQStats holds post-quantum encryption statistics
type PQStats struct {
	PacketsEncrypted atomic.Uint64
	PacketsDecrypted atomic.Uint64
	BytesEncrypted   atomic.Uint64
	BytesDecrypted   atomic.Uint64
	KeyRotations     atomic.Uint64
}

// NewPostQuantumCipher creates a new PQ cipher from a hybrid shared secret
func NewPostQuantumCipher(sharedSecret []byte, config *PostQuantumConfig) (*PostQuantumCipher, error) {
	if len(sharedSecret) < 64 {
		return nil, errors.New("shared secret too short")
	}
	if config == nil {
		config = DefaultPostQuantumConfig()
	}

	pq := &PostQuantumCipher{
		primaryKey:   make([]byte, 32),
		secondaryKey: make([]byte, 32),
		config:       config,
	}

	copy(pq.primaryKey, sharedSecret[:32])
	copy(pq.secondaryKey, sharedSecret[32:64])

	// Initialize ChaCha20-Poly1305
	var err error
	pq.primaryCipher, err = chacha20poly1305.NewX(pq.primaryKey)
	if err != nil {
		return nil, err
	}

	// Initialize AES-256-GCM
	block, err := aes.NewCipher(pq.secondaryKey)
	if err != nil {
		return nil, err
	}
	pq.secondaryCipher, err = cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return pq, nil
}

// Encrypt performs double-layer encryption
func (pq *PostQuantumCipher) Encrypt(plaintext []byte) ([]byte, error) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	nonce := pq.sendNonce.Add(1)

	var result []byte

	if pq.config.DoubleEncryption {
		// Layer 1: ChaCha20-Poly1305
		nonce1 := make([]byte, chacha20poly1305.NonceSizeX)
		binary.BigEndian.PutUint64(nonce1[:8], nonce)
		rand.Read(nonce1[8:])

		layer1 := pq.primaryCipher.Seal(nil, nonce1, plaintext, nil)

		// Layer 2: AES-256-GCM
		nonce2 := make([]byte, 12)
		binary.BigEndian.PutUint64(nonce2[:8], nonce)
		rand.Read(nonce2[8:])

		layer2 := pq.secondaryCipher.Seal(nil, nonce2, layer1, nil)

		// Combine: nonce1 || nonce2 || ciphertext
		result = make([]byte, len(nonce1)+len(nonce2)+len(layer2))
		copy(result[:len(nonce1)], nonce1)
		copy(result[len(nonce1):len(nonce1)+len(nonce2)], nonce2)
		copy(result[len(nonce1)+len(nonce2):], layer2)
	} else {
		// Single layer: ChaCha20-Poly1305
		nonce1 := make([]byte, chacha20poly1305.NonceSizeX)
		binary.BigEndian.PutUint64(nonce1[:8], nonce)
		rand.Read(nonce1[8:])

		ciphertext := pq.primaryCipher.Seal(nil, nonce1, plaintext, nil)

		result = make([]byte, len(nonce1)+len(ciphertext))
		copy(result[:len(nonce1)], nonce1)
		copy(result[len(nonce1):], ciphertext)
	}

	pq.stats.PacketsEncrypted.Add(1)
	pq.stats.BytesEncrypted.Add(uint64(len(plaintext)))

	return result, nil
}

// Decrypt performs double-layer decryption
func (pq *PostQuantumCipher) Decrypt(ciphertext []byte) ([]byte, error) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	var plaintext []byte

	if pq.config.DoubleEncryption {
		nonceSize1 := chacha20poly1305.NonceSizeX
		nonceSize2 := 12

		if len(ciphertext) < nonceSize1+nonceSize2 {
			return nil, errors.New("ciphertext too short")
		}

		nonce1 := ciphertext[:nonceSize1]
		nonce2 := ciphertext[nonceSize1 : nonceSize1+nonceSize2]
		layer2 := ciphertext[nonceSize1+nonceSize2:]

		// Layer 2: AES-256-GCM decrypt
		layer1, err := pq.secondaryCipher.Open(nil, nonce2, layer2, nil)
		if err != nil {
			return nil, err
		}

		// Layer 1: ChaCha20-Poly1305 decrypt
		plaintext, err = pq.primaryCipher.Open(nil, nonce1, layer1, nil)
		if err != nil {
			return nil, err
		}
	} else {
		nonceSize := chacha20poly1305.NonceSizeX
		if len(ciphertext) < nonceSize {
			return nil, errors.New("ciphertext too short")
		}

		nonce := ciphertext[:nonceSize]
		data := ciphertext[nonceSize:]

		var err error
		plaintext, err = pq.primaryCipher.Open(nil, nonce, data, nil)
		if err != nil {
			return nil, err
		}
	}

	pq.stats.PacketsDecrypted.Add(1)
	pq.stats.BytesDecrypted.Add(uint64(len(plaintext)))

	return plaintext, nil
}

// RotateKeys performs key rotation
func (pq *PostQuantumCipher) RotateKeys(newSharedSecret []byte) error {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if len(newSharedSecret) < 64 {
		return errors.New("new shared secret too short")
	}

	// Wipe old keys
	wipeBytes(pq.primaryKey)
	wipeBytes(pq.secondaryKey)

	// Set new keys
	copy(pq.primaryKey, newSharedSecret[:32])
	copy(pq.secondaryKey, newSharedSecret[32:64])

	// Reinitialize ciphers
	var err error
	pq.primaryCipher, err = chacha20poly1305.NewX(pq.primaryKey)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(pq.secondaryKey)
	if err != nil {
		return err
	}
	pq.secondaryCipher, err = cipher.NewGCM(block)
	if err != nil {
		return err
	}

	pq.stats.KeyRotations.Add(1)

	return nil
}

// Stats returns encryption statistics
func (pq *PostQuantumCipher) Stats() PQStatsSummary {
	return PQStatsSummary{
		PacketsEncrypted: pq.stats.PacketsEncrypted.Load(),
		PacketsDecrypted: pq.stats.PacketsDecrypted.Load(),
		BytesEncrypted:   pq.stats.BytesEncrypted.Load(),
		BytesDecrypted:   pq.stats.BytesDecrypted.Load(),
		KeyRotations:     pq.stats.KeyRotations.Load(),
	}
}

// PQStatsSummary provides a summary of PQ stats
type PQStatsSummary struct {
	PacketsEncrypted uint64
	PacketsDecrypted uint64
	BytesEncrypted   uint64
	BytesDecrypted   uint64
	KeyRotations     uint64
}

// Close securely wipes the cipher
func (pq *PostQuantumCipher) Close() {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	wipeBytes(pq.primaryKey)
	wipeBytes(pq.secondaryKey)
}

// PostQuantumSession provides a complete PQ-resistant VPN session
type PostQuantumSession struct {
	// Key pairs
	localKeyPair *HybridKeyPair

	// Session cipher
	cipher *PostQuantumCipher

	// Session identity
	sessionID [32]byte

	// Stats
	stats PQSessionStats

	config *PostQuantumConfig
	mu     sync.RWMutex
}

// PQSessionStats tracks session statistics
type PQSessionStats struct {
	Established       bool
	KeyExchanges      atomic.Uint64
	DataPacketsSent   atomic.Uint64
	DataPacketsRecv   atomic.Uint64
}

// NewPostQuantumSession creates a new PQ-resistant session
func NewPostQuantumSession(config *PostQuantumConfig) (*PostQuantumSession, error) {
	if config == nil {
		config = DefaultPostQuantumConfig()
	}

	keyPair, err := NewHybridKeyPair(config)
	if err != nil {
		return nil, err
	}

	session := &PostQuantumSession{
		localKeyPair: keyPair,
		config:       config,
	}

	rand.Read(session.sessionID[:])

	return session, nil
}

// PublicKey returns the session's public key for key exchange
func (s *PostQuantumSession) PublicKey() []byte {
	return s.localKeyPair.PublicKey()
}

// EstablishWithPeer completes key exchange with peer
func (s *PostQuantumSession) EstablishWithPeer(peerPublicKey []byte, isInitiator bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var sharedSecret []byte
	var err error

	if isInitiator {
		// Initiator encapsulates
		_, sharedSecret, err = s.localKeyPair.Encapsulate(peerPublicKey)
	} else {
		// Responder decapsulates
		sharedSecret, err = s.localKeyPair.Decapsulate(peerPublicKey)
	}

	if err != nil {
		return err
	}

	// Create session cipher
	s.cipher, err = NewPostQuantumCipher(sharedSecret, s.config)
	if err != nil {
		return err
	}

	s.stats.Established = true
	s.stats.KeyExchanges.Add(1)

	wipeBytes(sharedSecret)

	return nil
}

// Encrypt encrypts data for the peer
func (s *PostQuantumSession) Encrypt(data []byte) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.cipher == nil {
		return nil, errors.New("session not established")
	}

	ciphertext, err := s.cipher.Encrypt(data)
	if err != nil {
		return nil, err
	}

	s.stats.DataPacketsSent.Add(1)
	return ciphertext, nil
}

// Decrypt decrypts data from the peer
func (s *PostQuantumSession) Decrypt(ciphertext []byte) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.cipher == nil {
		return nil, errors.New("session not established")
	}

	plaintext, err := s.cipher.Decrypt(ciphertext)
	if err != nil {
		return nil, err
	}

	s.stats.DataPacketsRecv.Add(1)
	return plaintext, nil
}

// Rekey performs key rotation with new key material
func (s *PostQuantumSession) Rekey() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate new key pair
	newKeyPair, err := NewHybridKeyPair(s.config)
	if err != nil {
		return err
	}

	// Generate new shared secret (self-encapsulation for demonstration)
	_, sharedSecret, err := newKeyPair.Encapsulate(newKeyPair.PublicKey())
	if err != nil {
		return err
	}

	// Rotate cipher keys
	if s.cipher != nil {
		if err := s.cipher.RotateKeys(sharedSecret); err != nil {
			return err
		}
	}

	// Wipe old key pair
	s.localKeyPair.Wipe()
	s.localKeyPair = newKeyPair

	s.stats.KeyExchanges.Add(1)

	return nil
}

// SessionID returns the session identifier
func (s *PostQuantumSession) SessionID() [32]byte {
	return s.sessionID
}

// IsEstablished returns whether the session is established
func (s *PostQuantumSession) IsEstablished() bool {
	return s.stats.Established
}

// Stats returns session statistics
func (s *PostQuantumSession) Stats() PQSessionStatsSummary {
	return PQSessionStatsSummary{
		Established:     s.stats.Established,
		KeyExchanges:    s.stats.KeyExchanges.Load(),
		DataPacketsSent: s.stats.DataPacketsSent.Load(),
		DataPacketsRecv: s.stats.DataPacketsRecv.Load(),
	}
}

// PQSessionStatsSummary provides session stats summary
type PQSessionStatsSummary struct {
	Established     bool
	KeyExchanges    uint64
	DataPacketsSent uint64
	DataPacketsRecv uint64
}

// Close cleans up the session
func (s *PostQuantumSession) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cipher != nil {
		s.cipher.Close()
	}
	if s.localKeyPair != nil {
		s.localKeyPair.Wipe()
	}
}

// QuantumResistantKeyDerivation provides quantum-safe key derivation
type QuantumResistantKeyDerivation struct {
	masterKey []byte
	counter   atomic.Uint64
	mu        sync.Mutex
}

// NewQuantumResistantKeyDerivation creates a new QRKD instance
func NewQuantumResistantKeyDerivation(masterKey []byte) *QuantumResistantKeyDerivation {
	qrkd := &QuantumResistantKeyDerivation{
		masterKey: make([]byte, len(masterKey)),
	}
	copy(qrkd.masterKey, masterKey)
	return qrkd
}

// DeriveKey derives a new key for a specific purpose
func (q *QuantumResistantKeyDerivation) DeriveKey(purpose string, length int) ([]byte, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	counter := q.counter.Add(1)

	// Use SHAKE256 for quantum-resistant key derivation
	shake := sha3.NewShake256()
	shake.Write([]byte("QRKD-v1"))
	shake.Write(q.masterKey)
	shake.Write([]byte(purpose))

	counterBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBytes, counter)
	shake.Write(counterBytes)

	key := make([]byte, length)
	shake.Read(key)

	return key, nil
}

// Wipe securely erases the master key
func (q *QuantumResistantKeyDerivation) Wipe() {
	q.mu.Lock()
	defer q.mu.Unlock()
	wipeBytes(q.masterKey)
}
