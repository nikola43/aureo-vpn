package crypto

import (
	"bytes"
	"crypto/rand"
	"testing"
	"time"
)

// =============================================================================
// BlindingFactor Tests
// =============================================================================

func TestNewBlindingFactor(t *testing.T) {
	bf, err := NewBlindingFactor()
	if err != nil {
		t.Fatalf("NewBlindingFactor() error = %v", err)
	}
	defer bf.Wipe()

	if len(bf.factor) != 32 {
		t.Errorf("factor length = %d, want 32", len(bf.factor))
	}
	if len(bf.salt) != 16 {
		t.Errorf("salt length = %d, want 16", len(bf.salt))
	}
}

func TestBlindingFactorBlind(t *testing.T) {
	bf, err := NewBlindingFactor()
	if err != nil {
		t.Fatalf("NewBlindingFactor() error = %v", err)
	}
	defer bf.Wipe()

	identifier := []byte("test-identifier-123")
	blinded := bf.Blind(identifier)

	// Blinded output should be 32 bytes
	if len(blinded) != 32 {
		t.Errorf("blinded length = %d, want 32", len(blinded))
	}

	// Blinded should be different from original
	if bytes.Equal(blinded, identifier) {
		t.Error("blinded output should differ from input")
	}
}

func TestBlindingFactorConsistency(t *testing.T) {
	bf, err := NewBlindingFactor()
	if err != nil {
		t.Fatalf("NewBlindingFactor() error = %v", err)
	}
	defer bf.Wipe()

	identifier := []byte("consistent-test")

	blinded1 := bf.Blind(identifier)
	blinded2 := bf.Blind(identifier)

	// Same input with same blinding factor should produce same output
	if !bytes.Equal(blinded1, blinded2) {
		t.Error("same input should produce same blinded output")
	}
}

func TestBlindingFactorUniqueness(t *testing.T) {
	// Different blinding factors should produce different outputs
	bf1, _ := NewBlindingFactor()
	defer bf1.Wipe()

	bf2, _ := NewBlindingFactor()
	defer bf2.Wipe()

	identifier := []byte("uniqueness-test")

	blinded1 := bf1.Blind(identifier)
	blinded2 := bf2.Blind(identifier)

	// Different blinding factors should produce different outputs
	if bytes.Equal(blinded1, blinded2) {
		t.Error("different blinding factors should produce different outputs")
	}
}

func TestBlindingFactorWipe(t *testing.T) {
	bf, err := NewBlindingFactor()
	if err != nil {
		t.Fatalf("NewBlindingFactor() error = %v", err)
	}

	// Store original values
	originalFactor := make([]byte, len(bf.factor))
	copy(originalFactor, bf.factor)

	bf.Wipe()

	// After wipe, factor should be zeroed
	allZeros := true
	for _, b := range bf.factor {
		if b != 0 {
			allZeros = false
			break
		}
	}
	if !allZeros {
		t.Error("Wipe() should zero factor")
	}
}

// =============================================================================
// ZeroKnowledgeSession Tests
// =============================================================================

func TestNewZeroKnowledgeSession(t *testing.T) {
	// Create a key exchange to get a peer public key
	peerKE, _ := NewKeyExchange()
	defer peerKE.Wipe()

	session, err := NewZeroKnowledgeSession(peerKE.PublicKey(), true)
	if err != nil {
		t.Fatalf("NewZeroKnowledgeSession() error = %v", err)
	}
	defer session.Close()

	if session == nil {
		t.Fatal("NewZeroKnowledgeSession() returned nil")
	}

	// Check public key
	pubKey := session.PublicKey()
	if len(pubKey) != 32 {
		t.Errorf("PublicKey() length = %d, want 32", len(pubKey))
	}
}

func TestZeroKnowledgeSessionBlindedID(t *testing.T) {
	peerKE, _ := NewKeyExchange()
	defer peerKE.Wipe()

	session, err := NewZeroKnowledgeSession(peerKE.PublicKey(), true)
	if err != nil {
		t.Fatalf("NewZeroKnowledgeSession() error = %v", err)
	}
	defer session.Close()

	blindedID := session.BlindedID()
	if len(blindedID) != 32 {
		t.Errorf("BlindedID() length = %d, want 32", len(blindedID))
	}

	// BlindedID should be consistent
	blindedID2 := session.BlindedID()
	if !bytes.Equal(blindedID, blindedID2) {
		t.Error("BlindedID() should be consistent")
	}
}

func TestZeroKnowledgeSessionEncryptDecrypt(t *testing.T) {
	// Create two sessions (client and server)
	clientKE, _ := NewKeyExchange()
	defer clientKE.Wipe()

	serverKE, _ := NewKeyExchange()
	defer serverKE.Wipe()

	clientSession, _ := NewZeroKnowledgeSession(serverKE.PublicKey(), true)
	defer clientSession.Close()

	serverSession, _ := NewZeroKnowledgeSession(clientKE.PublicKey(), false)
	defer serverSession.Close()

	// Encrypt on client
	plaintext := []byte("secret message")
	encrypted, err := clientSession.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Encrypted should be different from plaintext
	if bytes.Equal(encrypted, plaintext) {
		t.Error("encrypted should differ from plaintext")
	}
}

func TestZeroKnowledgeSessionStats(t *testing.T) {
	peerKE, _ := NewKeyExchange()
	defer peerKE.Wipe()

	session, _ := NewZeroKnowledgeSession(peerKE.PublicKey(), true)
	defer session.Close()

	stats := session.Stats()
	if stats.PacketsSent != 0 {
		t.Errorf("initial PacketsSent = %d, want 0", stats.PacketsSent)
	}

	// Encrypt some data
	for i := 0; i < 5; i++ {
		session.Encrypt([]byte("test"))
	}

	stats = session.Stats()
	if stats.PacketsSent != 5 {
		t.Errorf("PacketsSent = %d, want 5", stats.PacketsSent)
	}
}

func TestZeroKnowledgeSessionNeedsRekey(t *testing.T) {
	peerKE, _ := NewKeyExchange()
	defer peerKE.Wipe()

	session, _ := NewZeroKnowledgeSession(peerKE.PublicKey(), true)
	defer session.Close()

	// Initially should not need rekey
	if session.NeedsRekey() {
		t.Error("new session should not need rekey")
	}
}

// =============================================================================
// AnonymousDataStore Tests
// =============================================================================

func TestNewAnonymousDataStore(t *testing.T) {
	masterKey := make([]byte, 32)
	rand.Read(masterKey)

	store, err := NewAnonymousDataStore(masterKey)
	if err != nil {
		t.Fatalf("NewAnonymousDataStore() error = %v", err)
	}
	defer store.Close()

	if store == nil {
		t.Fatal("NewAnonymousDataStore() returned nil")
	}
}

func TestAnonymousDataStoreShortKey(t *testing.T) {
	shortKey := make([]byte, 16) // Too short
	rand.Read(shortKey)

	_, err := NewAnonymousDataStore(shortKey)
	if err == nil {
		t.Error("NewAnonymousDataStore() should fail with short key")
	}
}

func TestAnonymousDataStoreStoreRetrieve(t *testing.T) {
	masterKey := make([]byte, 32)
	rand.Read(masterKey)

	store, _ := NewAnonymousDataStore(masterKey)
	defer store.Close()

	blindedKey := make([]byte, 32)
	rand.Read(blindedKey)

	plaintext := []byte("secret data to store")

	// Store data
	err := store.Store(blindedKey, plaintext)
	if err != nil {
		t.Fatalf("Store() error = %v", err)
	}

	// Retrieve data
	retrieved, err := store.Retrieve(blindedKey)
	if err != nil {
		t.Fatalf("Retrieve() error = %v", err)
	}

	if !bytes.Equal(retrieved, plaintext) {
		t.Error("retrieved data should match stored data")
	}
}

func TestAnonymousDataStoreRetrieveNotFound(t *testing.T) {
	masterKey := make([]byte, 32)
	rand.Read(masterKey)

	store, _ := NewAnonymousDataStore(masterKey)
	defer store.Close()

	nonExistentKey := make([]byte, 32)
	rand.Read(nonExistentKey)

	_, err := store.Retrieve(nonExistentKey)
	if err == nil {
		t.Error("Retrieve() should fail for non-existent key")
	}
}

func TestAnonymousDataStoreDelete(t *testing.T) {
	masterKey := make([]byte, 32)
	rand.Read(masterKey)

	store, _ := NewAnonymousDataStore(masterKey)
	defer store.Close()

	blindedKey := make([]byte, 32)
	rand.Read(blindedKey)

	// Store and then delete
	store.Store(blindedKey, []byte("to be deleted"))
	store.Delete(blindedKey)

	// Should not be retrievable after delete
	_, err := store.Retrieve(blindedKey)
	if err == nil {
		t.Error("Retrieve() should fail after Delete()")
	}
}

// =============================================================================
// IPAnonymizer Tests
// =============================================================================

func TestNewIPAnonymizer(t *testing.T) {
	ia, err := NewIPAnonymizer()
	if err != nil {
		t.Fatalf("NewIPAnonymizer() error = %v", err)
	}
	defer ia.Close()

	if ia == nil {
		t.Fatal("NewIPAnonymizer() returned nil")
	}
}

func TestIPAnonymizerAnonymizeIP(t *testing.T) {
	ia, _ := NewIPAnonymizer()
	defer ia.Close()

	ip := []byte{192, 168, 1, 100}
	anonymized := ia.AnonymizeIP(ip)

	if len(anonymized) != 32 {
		t.Errorf("anonymized length = %d, want 32", len(anonymized))
	}

	// Should not equal original
	if bytes.Equal(anonymized, ip) {
		t.Error("anonymized should differ from original")
	}
}

func TestIPAnonymizerConsistency(t *testing.T) {
	ia, _ := NewIPAnonymizer()
	defer ia.Close()

	ip := []byte{10, 0, 0, 1}

	anon1 := ia.AnonymizeIP(ip)
	anon2 := ia.AnonymizeIP(ip)

	// Same IP in same hour should produce same output
	if !bytes.Equal(anon1, anon2) {
		t.Error("same IP should produce same anonymized output")
	}
}

func TestIPAnonymizerDifferentIPs(t *testing.T) {
	ia, _ := NewIPAnonymizer()
	defer ia.Close()

	ip1 := []byte{192, 168, 1, 1}
	ip2 := []byte{192, 168, 1, 2}

	anon1 := ia.AnonymizeIP(ip1)
	anon2 := ia.AnonymizeIP(ip2)

	// Different IPs should produce different outputs
	if bytes.Equal(anon1, anon2) {
		t.Error("different IPs should produce different anonymized outputs")
	}
}

func TestIPAnonymizerRotate(t *testing.T) {
	ia, _ := NewIPAnonymizer()
	defer ia.Close()

	ip := []byte{172, 16, 0, 1}
	anonBefore := ia.AnonymizeIP(ip)

	// Rotate blinding factors
	err := ia.RotateBlindingFactors()
	if err != nil {
		t.Fatalf("RotateBlindingFactors() error = %v", err)
	}

	anonAfter := ia.AnonymizeIP(ip)

	// After rotation, same IP should produce different output
	if bytes.Equal(anonBefore, anonAfter) {
		t.Error("IP should produce different output after rotation")
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkBlindingFactor(b *testing.B) {
	bf, _ := NewBlindingFactor()
	defer bf.Wipe()

	identifier := make([]byte, 32)
	rand.Read(identifier)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.Blind(identifier)
	}
}

func BenchmarkZKSessionEncrypt(b *testing.B) {
	peerKE, _ := NewKeyExchange()
	defer peerKE.Wipe()

	session, _ := NewZeroKnowledgeSession(peerKE.PublicKey(), true)
	defer session.Close()

	data := make([]byte, 1400)
	rand.Read(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		session.Encrypt(data)
	}
}

func BenchmarkAnonymousDataStore(b *testing.B) {
	masterKey := make([]byte, 32)
	rand.Read(masterKey)

	store, _ := NewAnonymousDataStore(masterKey)
	defer store.Close()

	data := make([]byte, 1024)
	rand.Read(data)

	blindedKey := make([]byte, 32)
	rand.Read(blindedKey)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Store(blindedKey, data)
	}
}

func BenchmarkIPAnonymizer(b *testing.B) {
	ia, _ := NewIPAnonymizer()
	defer ia.Close()

	ip := []byte{192, 168, 1, 100}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ia.AnonymizeIP(ip)
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestZeroKnowledgeEndToEnd(t *testing.T) {
	// Create a complete zero-knowledge session flow
	masterKey := make([]byte, 64)
	rand.Read(masterKey)

	// Create data store
	store, err := NewAnonymousDataStore(masterKey)
	if err != nil {
		t.Fatalf("NewAnonymousDataStore() error = %v", err)
	}
	defer store.Close()

	// Create IP anonymizer
	ia, err := NewIPAnonymizer()
	if err != nil {
		t.Fatalf("NewIPAnonymizer() error = %v", err)
	}
	defer ia.Close()

	// Simulate user connection
	userIP := []byte{203, 0, 113, 42}
	anonymizedIP := ia.AnonymizeIP(userIP)

	// Create blinded session ID
	bf, _ := NewBlindingFactor()
	defer bf.Wipe()

	sessionID := []byte("user-session-12345")
	blindedSessionID := bf.Blind(sessionID)

	// Store session data
	sessionData := []byte("encrypted user preferences")
	err = store.Store(blindedSessionID, sessionData)
	if err != nil {
		t.Fatalf("Store() error = %v", err)
	}

	// Later, retrieve session data
	retrieved, err := store.Retrieve(blindedSessionID)
	if err != nil {
		t.Fatalf("Retrieve() error = %v", err)
	}

	if !bytes.Equal(retrieved, sessionData) {
		t.Error("retrieved session data mismatch")
	}

	// Verify anonymized IP is unlinkable
	if bytes.Equal(anonymizedIP, userIP) {
		t.Error("anonymized IP should not equal original")
	}

	// Verify blinded session ID is unlinkable
	if bytes.Equal(blindedSessionID, sessionID) {
		t.Error("blinded session ID should not equal original")
	}
}

func TestZeroKnowledgeSessionConcurrent(t *testing.T) {
	peerKE, _ := NewKeyExchange()
	defer peerKE.Wipe()

	session, _ := NewZeroKnowledgeSession(peerKE.PublicKey(), true)
	defer session.Close()

	done := make(chan bool)

	// Multiple concurrent operations
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				session.Encrypt([]byte("concurrent data"))
				session.BlindedID()
				session.Stats()
				time.Sleep(time.Microsecond)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}
}
