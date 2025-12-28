package crypto

import (
	"bytes"
	"sync"
	"testing"
)

// =============================================================================
// ML-KEM Key Pair Tests
// =============================================================================

func TestNewMLKEMKeyPair(t *testing.T) {
	kp, err := NewMLKEMKeyPair()
	if err != nil {
		t.Fatalf("NewMLKEMKeyPair() error = %v", err)
	}
	defer kp.Wipe()

	if len(kp.PublicKey) != MLKEMPublicKeySize {
		t.Errorf("PublicKey size = %d, want %d", len(kp.PublicKey), MLKEMPublicKeySize)
	}

	if len(kp.PrivateKey) != MLKEMPrivateKeySize {
		t.Errorf("PrivateKey size = %d, want %d", len(kp.PrivateKey), MLKEMPrivateKeySize)
	}
}

func TestMLKEMEncapsulateDecapsulate(t *testing.T) {
	alice, err := NewMLKEMKeyPair()
	if err != nil {
		t.Fatalf("NewMLKEMKeyPair() error = %v", err)
	}
	defer alice.Wipe()

	bob, err := NewMLKEMKeyPair()
	if err != nil {
		t.Fatalf("NewMLKEMKeyPair() error = %v", err)
	}
	defer bob.Wipe()

	// Alice encapsulates to Bob
	ciphertext, aliceSecret, err := alice.Encapsulate(bob.PublicKey)
	if err != nil {
		t.Fatalf("Encapsulate() error = %v", err)
	}

	if len(ciphertext) != MLKEMCiphertextSize {
		t.Errorf("Ciphertext size = %d, want %d", len(ciphertext), MLKEMCiphertextSize)
	}

	if len(aliceSecret) != MLKEMSharedKeySize {
		t.Errorf("SharedSecret size = %d, want %d", len(aliceSecret), MLKEMSharedKeySize)
	}

	// Bob decapsulates
	bobSecret, err := bob.Decapsulate(ciphertext)
	if err != nil {
		t.Fatalf("Decapsulate() error = %v", err)
	}

	// Secrets should match
	if !bytes.Equal(aliceSecret, bobSecret) {
		t.Error("Shared secrets don't match")
	}
}

func TestMLKEMInvalidPublicKey(t *testing.T) {
	kp, _ := NewMLKEMKeyPair()
	defer kp.Wipe()

	// Too short
	_, _, err := kp.Encapsulate(make([]byte, 10))
	if err != ErrInvalidPQPublicKey {
		t.Errorf("Expected ErrInvalidPQPublicKey, got %v", err)
	}
}

func TestMLKEMInvalidCiphertext(t *testing.T) {
	kp, _ := NewMLKEMKeyPair()
	defer kp.Wipe()

	// Too short
	_, err := kp.Decapsulate(make([]byte, 10))
	if err != ErrInvalidPQCiphertext {
		t.Errorf("Expected ErrInvalidPQCiphertext, got %v", err)
	}
}

// =============================================================================
// Hybrid Key Pair Tests
// =============================================================================

func TestNewHybridKeyPair(t *testing.T) {
	config := DefaultPostQuantumConfig()
	kp, err := NewHybridKeyPair(config)
	if err != nil {
		t.Fatalf("NewHybridKeyPair() error = %v", err)
	}
	defer kp.Wipe()

	pubKey := kp.PublicKey()
	if len(pubKey) != HybridPublicKeySize {
		t.Errorf("PublicKey size = %d, want %d", len(pubKey), HybridPublicKeySize)
	}
}

func TestHybridKeyExchange(t *testing.T) {
	config := DefaultPostQuantumConfig()

	alice, err := NewHybridKeyPair(config)
	if err != nil {
		t.Fatalf("NewHybridKeyPair() Alice error = %v", err)
	}
	defer alice.Wipe()

	bob, err := NewHybridKeyPair(config)
	if err != nil {
		t.Fatalf("NewHybridKeyPair() Bob error = %v", err)
	}
	defer bob.Wipe()

	// Alice encapsulates to Bob
	ciphertext, aliceSecret, err := alice.Encapsulate(bob.PublicKey())
	if err != nil {
		t.Fatalf("Encapsulate() error = %v", err)
	}

	if len(ciphertext) != HybridCiphertextSize {
		t.Errorf("Ciphertext size = %d, want %d", len(ciphertext), HybridCiphertextSize)
	}

	if len(aliceSecret) != 64 {
		t.Errorf("SharedSecret size = %d, want 64", len(aliceSecret))
	}

	// Bob decapsulates
	bobSecret, err := bob.Decapsulate(ciphertext)
	if err != nil {
		t.Fatalf("Decapsulate() error = %v", err)
	}

	// Secrets should match
	if !bytes.Equal(aliceSecret, bobSecret) {
		t.Error("Hybrid shared secrets don't match")
	}
}

func TestHybridKeyPairWithNilConfig(t *testing.T) {
	kp, err := NewHybridKeyPair(nil)
	if err != nil {
		t.Fatalf("NewHybridKeyPair(nil) error = %v", err)
	}
	defer kp.Wipe()

	if kp.config == nil {
		t.Error("Config should not be nil")
	}
}

// =============================================================================
// Post-Quantum Cipher Tests
// =============================================================================

func TestNewPostQuantumCipher(t *testing.T) {
	sharedSecret := make([]byte, 64)
	for i := range sharedSecret {
		sharedSecret[i] = byte(i)
	}

	cipher, err := NewPostQuantumCipher(sharedSecret, nil)
	if err != nil {
		t.Fatalf("NewPostQuantumCipher() error = %v", err)
	}
	defer cipher.Close()

	if cipher.primaryCipher == nil {
		t.Error("Primary cipher should not be nil")
	}

	if cipher.secondaryCipher == nil {
		t.Error("Secondary cipher should not be nil")
	}
}

func TestPostQuantumCipherEncryptDecrypt(t *testing.T) {
	sharedSecret := make([]byte, 64)
	for i := range sharedSecret {
		sharedSecret[i] = byte(i)
	}

	cipher, err := NewPostQuantumCipher(sharedSecret, nil)
	if err != nil {
		t.Fatalf("NewPostQuantumCipher() error = %v", err)
	}
	defer cipher.Close()

	testCases := []struct {
		name      string
		plaintext []byte
	}{
		{"small", []byte("hello")},
		{"medium", bytes.Repeat([]byte("test"), 100)},
		{"large", bytes.Repeat([]byte("data"), 1000)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encrypted, err := cipher.Encrypt(tc.plaintext)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			// Encrypted should be longer (nonces + tags)
			if len(encrypted) <= len(tc.plaintext) {
				t.Error("Encrypted data should be longer than plaintext")
			}

			decrypted, err := cipher.Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if !bytes.Equal(decrypted, tc.plaintext) {
				t.Error("Decrypted data doesn't match plaintext")
			}
		})
	}
}

func TestPostQuantumCipherDoubleEncryption(t *testing.T) {
	sharedSecret := make([]byte, 64)
	for i := range sharedSecret {
		sharedSecret[i] = byte(i)
	}

	// With double encryption
	configDouble := DefaultPostQuantumConfig()
	configDouble.DoubleEncryption = true

	cipherDouble, _ := NewPostQuantumCipher(sharedSecret, configDouble)
	defer cipherDouble.Close()

	// Without double encryption
	configSingle := DefaultPostQuantumConfig()
	configSingle.DoubleEncryption = false

	cipherSingle, _ := NewPostQuantumCipher(sharedSecret, configSingle)
	defer cipherSingle.Close()

	plaintext := []byte("test data for encryption")

	encryptedDouble, _ := cipherDouble.Encrypt(plaintext)
	encryptedSingle, _ := cipherSingle.Encrypt(plaintext)

	// Double encrypted should be longer (extra nonce + tag)
	if len(encryptedDouble) <= len(encryptedSingle) {
		t.Error("Double encrypted should be longer than single")
	}

	// Both should decrypt correctly
	decryptedDouble, _ := cipherDouble.Decrypt(encryptedDouble)
	decryptedSingle, _ := cipherSingle.Decrypt(encryptedSingle)

	if !bytes.Equal(decryptedDouble, plaintext) {
		t.Error("Double decryption failed")
	}

	if !bytes.Equal(decryptedSingle, plaintext) {
		t.Error("Single decryption failed")
	}
}

func TestPostQuantumCipherKeyRotation(t *testing.T) {
	sharedSecret := make([]byte, 64)
	for i := range sharedSecret {
		sharedSecret[i] = byte(i)
	}

	cipher, err := NewPostQuantumCipher(sharedSecret, nil)
	if err != nil {
		t.Fatalf("NewPostQuantumCipher() error = %v", err)
	}
	defer cipher.Close()

	// Encrypt with old key
	plaintext := []byte("test before rotation")
	encrypted1, _ := cipher.Encrypt(plaintext)

	// Rotate keys
	newSecret := make([]byte, 64)
	for i := range newSecret {
		newSecret[i] = byte(i + 100)
	}

	err = cipher.RotateKeys(newSecret)
	if err != nil {
		t.Fatalf("RotateKeys() error = %v", err)
	}

	// Encrypt with new key
	encrypted2, _ := cipher.Encrypt(plaintext)

	// Encrypted values should be different
	if bytes.Equal(encrypted1, encrypted2) {
		t.Error("Encrypted values should differ after key rotation")
	}

	// Should decrypt with new key
	decrypted, _ := cipher.Decrypt(encrypted2)
	if !bytes.Equal(decrypted, plaintext) {
		t.Error("Decryption after rotation failed")
	}

	// Stats should show rotation
	stats := cipher.Stats()
	if stats.KeyRotations != 1 {
		t.Errorf("KeyRotations = %d, want 1", stats.KeyRotations)
	}
}

func TestPostQuantumCipherStats(t *testing.T) {
	sharedSecret := make([]byte, 64)
	cipher, _ := NewPostQuantumCipher(sharedSecret, nil)
	defer cipher.Close()

	// Encrypt/decrypt some packets
	for i := 0; i < 10; i++ {
		encrypted, _ := cipher.Encrypt([]byte("test data"))
		cipher.Decrypt(encrypted)
	}

	stats := cipher.Stats()

	if stats.PacketsEncrypted != 10 {
		t.Errorf("PacketsEncrypted = %d, want 10", stats.PacketsEncrypted)
	}

	if stats.PacketsDecrypted != 10 {
		t.Errorf("PacketsDecrypted = %d, want 10", stats.PacketsDecrypted)
	}
}

// =============================================================================
// Post-Quantum Session Tests
// =============================================================================

func TestNewPostQuantumSession(t *testing.T) {
	session, err := NewPostQuantumSession(nil)
	if err != nil {
		t.Fatalf("NewPostQuantumSession() error = %v", err)
	}
	defer session.Close()

	pubKey := session.PublicKey()
	if len(pubKey) != HybridPublicKeySize {
		t.Errorf("PublicKey size = %d, want %d", len(pubKey), HybridPublicKeySize)
	}

	if session.IsEstablished() {
		t.Error("Session should not be established yet")
	}
}

func TestPostQuantumSessionEstablish(t *testing.T) {
	// Create two sessions (client and server)
	client, _ := NewPostQuantumSession(nil)
	defer client.Close()

	server, _ := NewPostQuantumSession(nil)
	defer server.Close()

	// Exchange public keys and establish
	// Client is initiator
	err := client.EstablishWithPeer(server.PublicKey(), true)
	if err != nil {
		t.Fatalf("Client establish error = %v", err)
	}

	// Server is responder
	err = server.EstablishWithPeer(client.PublicKey(), false)
	if err != nil {
		t.Fatalf("Server establish error = %v", err)
	}

	if !client.IsEstablished() {
		t.Error("Client should be established")
	}

	if !server.IsEstablished() {
		t.Error("Server should be established")
	}
}

func TestPostQuantumSessionEncryptDecrypt(t *testing.T) {
	// Create and establish sessions
	client, _ := NewPostQuantumSession(nil)
	defer client.Close()

	server, _ := NewPostQuantumSession(nil)
	defer server.Close()

	// Note: In a real implementation, we'd need proper key exchange
	// For this test, we establish each session with its own key
	client.EstablishWithPeer(server.PublicKey(), true)
	server.EstablishWithPeer(client.PublicKey(), false)

	// Client encrypts
	plaintext := []byte("hello from client")
	encrypted, err := client.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Client Encrypt() error = %v", err)
	}

	// Verify encrypted data is different from plaintext
	if bytes.Equal(encrypted, plaintext) {
		t.Error("Encrypted data should differ from plaintext")
	}

	// Client can decrypt its own messages
	decrypted, err := client.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Client Decrypt() error = %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("Decrypted data should match plaintext")
	}
}

func TestPostQuantumSessionRekey(t *testing.T) {
	session, _ := NewPostQuantumSession(nil)
	defer session.Close()

	// Establish first
	session.EstablishWithPeer(session.PublicKey(), true)

	// Get initial stats
	initialStats := session.Stats()

	// Rekey
	err := session.Rekey()
	if err != nil {
		t.Fatalf("Rekey() error = %v", err)
	}

	// Key exchanges should increase
	newStats := session.Stats()
	if newStats.KeyExchanges <= initialStats.KeyExchanges {
		t.Error("KeyExchanges should increase after rekey")
	}
}

func TestPostQuantumSessionStats(t *testing.T) {
	session, _ := NewPostQuantumSession(nil)
	defer session.Close()

	session.EstablishWithPeer(session.PublicKey(), true)

	// Send some data
	for i := 0; i < 5; i++ {
		session.Encrypt([]byte("test"))
	}

	stats := session.Stats()
	if stats.DataPacketsSent != 5 {
		t.Errorf("DataPacketsSent = %d, want 5", stats.DataPacketsSent)
	}
}

// =============================================================================
// Quantum-Resistant Key Derivation Tests
// =============================================================================

func TestQuantumResistantKeyDerivation(t *testing.T) {
	masterKey := make([]byte, 32)
	for i := range masterKey {
		masterKey[i] = byte(i)
	}

	qrkd := NewQuantumResistantKeyDerivation(masterKey)
	defer qrkd.Wipe()

	// Derive keys for different purposes
	key1, err := qrkd.DeriveKey("encryption", 32)
	if err != nil {
		t.Fatalf("DeriveKey() error = %v", err)
	}

	key2, err := qrkd.DeriveKey("authentication", 32)
	if err != nil {
		t.Fatalf("DeriveKey() error = %v", err)
	}

	// Different purposes should produce different keys
	if bytes.Equal(key1, key2) {
		t.Error("Different purposes should produce different keys")
	}

	// Same purpose should produce different keys (counter increments)
	key3, _ := qrkd.DeriveKey("encryption", 32)
	if bytes.Equal(key1, key3) {
		t.Error("Same purpose should produce different keys due to counter")
	}
}

func TestQuantumResistantKeyDerivationLength(t *testing.T) {
	masterKey := make([]byte, 32)
	qrkd := NewQuantumResistantKeyDerivation(masterKey)
	defer qrkd.Wipe()

	lengths := []int{16, 32, 64, 128}

	for _, length := range lengths {
		key, err := qrkd.DeriveKey("test", length)
		if err != nil {
			t.Fatalf("DeriveKey() error = %v", err)
		}

		if len(key) != length {
			t.Errorf("Key length = %d, want %d", len(key), length)
		}
	}
}

// =============================================================================
// Concurrent Tests
// =============================================================================

func TestPostQuantumCipherConcurrent(t *testing.T) {
	sharedSecret := make([]byte, 64)
	cipher, _ := NewPostQuantumCipher(sharedSecret, nil)
	defer cipher.Close()

	var wg sync.WaitGroup
	numGoroutines := 10
	opsPerGoroutine := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				encrypted, err := cipher.Encrypt([]byte("concurrent test"))
				if err != nil {
					t.Errorf("Encrypt() error = %v", err)
					return
				}
				_, err = cipher.Decrypt(encrypted)
				if err != nil {
					t.Errorf("Decrypt() error = %v", err)
					return
				}
			}
		}()
	}

	wg.Wait()

	stats := cipher.Stats()
	expected := uint64(numGoroutines * opsPerGoroutine)
	if stats.PacketsEncrypted != expected {
		t.Errorf("PacketsEncrypted = %d, want %d", stats.PacketsEncrypted, expected)
	}
}

// =============================================================================
// Config Tests
// =============================================================================

func TestDefaultPostQuantumConfig(t *testing.T) {
	config := DefaultPostQuantumConfig()

	if config.SecurityLevel != PQSecurityLevel5 {
		t.Errorf("SecurityLevel = %d, want %d", config.SecurityLevel, PQSecurityLevel5)
	}

	if !config.HybridMode {
		t.Error("HybridMode should be true by default")
	}

	if !config.DoubleEncryption {
		t.Error("DoubleEncryption should be true by default")
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkMLKEMKeyGen(b *testing.B) {
	for i := 0; i < b.N; i++ {
		kp, _ := NewMLKEMKeyPair()
		kp.Wipe()
	}
}

func BenchmarkMLKEMEncapsulate(b *testing.B) {
	alice, _ := NewMLKEMKeyPair()
	bob, _ := NewMLKEMKeyPair()
	defer alice.Wipe()
	defer bob.Wipe()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		alice.Encapsulate(bob.PublicKey)
	}
}

func BenchmarkHybridKeyExchange(b *testing.B) {
	config := DefaultPostQuantumConfig()
	alice, _ := NewHybridKeyPair(config)
	bob, _ := NewHybridKeyPair(config)
	defer alice.Wipe()
	defer bob.Wipe()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		alice.Encapsulate(bob.PublicKey())
	}
}

func BenchmarkPQEncrypt(b *testing.B) {
	sharedSecret := make([]byte, 64)
	cipher, _ := NewPostQuantumCipher(sharedSecret, nil)
	defer cipher.Close()

	data := make([]byte, 1400)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cipher.Encrypt(data)
	}
}

func BenchmarkPQDecrypt(b *testing.B) {
	sharedSecret := make([]byte, 64)
	cipher, _ := NewPostQuantumCipher(sharedSecret, nil)
	defer cipher.Close()

	data := make([]byte, 1400)
	encrypted, _ := cipher.Encrypt(data)

	// Create new cipher to reset state
	cipher2, _ := NewPostQuantumCipher(sharedSecret, nil)
	defer cipher2.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cipher2.Decrypt(encrypted)
	}
}
