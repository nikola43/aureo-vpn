// Package security provides comprehensive security tests
package security

import (
	"bytes"
	"crypto/rand"
	"sync"
	"testing"
	"time"
)

// =============================================================================
// SecureBytes and Key Generation Tests
// =============================================================================

func TestSecureBytes(t *testing.T) {
	tests := []struct {
		name    string
		size    int
		wantErr bool
	}{
		{"valid 16 bytes", 16, false},
		{"valid 32 bytes", 32, false},
		{"valid 64 bytes", 64, false},
		{"zero bytes", 0, true},
		{"negative bytes", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SecureBytes(tt.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecureBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.size {
				t.Errorf("SecureBytes() returned %d bytes, want %d", len(got), tt.size)
			}
		})
	}
}

func TestSecureBytesUniqueness(t *testing.T) {
	// Generate multiple keys and ensure they're unique
	keys := make(map[string]bool)
	for i := 0; i < 100; i++ {
		key, err := SecureBytes(32)
		if err != nil {
			t.Fatalf("SecureBytes() failed: %v", err)
		}
		keyStr := string(key)
		if keys[keyStr] {
			t.Errorf("SecureBytes() generated duplicate key")
		}
		keys[keyStr] = true
	}
}

func TestSecureKey(t *testing.T) {
	tests := []struct {
		name    string
		size    int
		wantErr bool
	}{
		{"AES-128 key", KeySize128, false},
		{"AES-256 key", KeySize256, false},
		{"HMAC-512 key", KeySize512, false},
		{"invalid size", 24, true},
		{"zero size", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SecureKey(tt.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecureKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.size {
				t.Errorf("SecureKey() returned %d bytes, want %d", len(got), tt.size)
			}
		})
	}
}

func TestSecureNonce(t *testing.T) {
	tests := []struct {
		name    string
		size    int
		wantErr bool
	}{
		{"GCM nonce", NonceSize12, false},
		{"XChaCha nonce", NonceSize24, false},
		{"invalid size", 16, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SecureNonce(tt.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecureNonce() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.size {
				t.Errorf("SecureNonce() returned %d bytes, want %d", len(got), tt.size)
			}
		})
	}
}

// =============================================================================
// SecureString Tests
// =============================================================================

func TestSecureString(t *testing.T) {
	secret := "super-secret-password"
	ss := NewSecureString(secret)

	// Test String() method
	if ss.String() != secret {
		t.Errorf("SecureString.String() = %v, want %v", ss.String(), secret)
	}

	// Test Bytes() method
	if string(ss.Bytes()) != secret {
		t.Errorf("SecureString.Bytes() = %v, want %v", string(ss.Bytes()), secret)
	}

	// Test Wipe() method
	ss.Wipe()
	if ss.String() != "" {
		t.Errorf("SecureString.String() after Wipe() = %v, want empty", ss.String())
	}
}

func TestSecureStringConcurrency(t *testing.T) {
	ss := NewSecureString("test-secret")
	var wg sync.WaitGroup

	// Multiple concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = ss.String()
			_ = ss.Bytes()
		}()
	}

	wg.Wait()
}

// =============================================================================
// WipeBytes Tests
// =============================================================================

func TestWipeBytes(t *testing.T) {
	data := make([]byte, 32)
	rand.Read(data)

	// Copy original for comparison
	original := make([]byte, len(data))
	copy(original, data)

	// Wipe the data
	WipeBytes(data)

	// All bytes should be zero
	for i, b := range data {
		if b != 0 {
			t.Errorf("WipeBytes() byte %d = %v, want 0", i, b)
		}
	}

	// Should be different from original
	if bytes.Equal(data, original) {
		t.Error("WipeBytes() did not change the data")
	}
}

func TestWipeBytesNil(t *testing.T) {
	// Should not panic
	WipeBytes(nil)
}

// =============================================================================
// AEAD Encryption Tests
// =============================================================================

func TestAESGCM(t *testing.T) {
	key, _ := SecureKey(KeySize256)
	aead, err := NewAESGCM(key)
	if err != nil {
		t.Fatalf("NewAESGCM() error = %v", err)
	}
	defer aead.Wipe()

	plaintext := []byte("Hello, World!")
	additionalData := []byte("additional authenticated data")

	// Encrypt
	ciphertext, err := aead.Encrypt(plaintext, additionalData)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Ciphertext should be different from plaintext
	if bytes.Equal(ciphertext, plaintext) {
		t.Error("Encrypt() produced ciphertext equal to plaintext")
	}

	// Decrypt
	decrypted, err := aead.Decrypt(ciphertext, additionalData)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	// Decrypted should equal plaintext
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypt() = %v, want %v", decrypted, plaintext)
	}
}

func TestAESGCMInvalidKey(t *testing.T) {
	_, err := NewAESGCM([]byte("short-key"))
	if err == nil {
		t.Error("NewAESGCM() with short key should fail")
	}
}

func TestAESGCMTamperDetection(t *testing.T) {
	key, _ := SecureKey(KeySize256)
	aead, _ := NewAESGCM(key)
	defer aead.Wipe()

	plaintext := []byte("sensitive data")
	ciphertext, _ := aead.Encrypt(plaintext, nil)

	// Tamper with ciphertext
	ciphertext[len(ciphertext)-1] ^= 0xFF

	// Decryption should fail
	_, err := aead.Decrypt(ciphertext, nil)
	if err == nil {
		t.Error("Decrypt() should fail with tampered ciphertext")
	}
}

func TestAESGCMWrongAD(t *testing.T) {
	key, _ := SecureKey(KeySize256)
	aead, _ := NewAESGCM(key)
	defer aead.Wipe()

	plaintext := []byte("sensitive data")
	ciphertext, _ := aead.Encrypt(plaintext, []byte("correct AD"))

	// Decrypt with wrong AD should fail
	_, err := aead.Decrypt(ciphertext, []byte("wrong AD"))
	if err == nil {
		t.Error("Decrypt() should fail with wrong additional data")
	}
}

func TestChaCha20Poly1305(t *testing.T) {
	key, _ := SecureKey(KeySize256)
	aead, err := NewChaCha20Poly1305(key)
	if err != nil {
		t.Fatalf("NewChaCha20Poly1305() error = %v", err)
	}
	defer aead.Wipe()

	plaintext := []byte("Hello, ChaCha!")

	ciphertext, err := aead.Encrypt(plaintext, nil)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	decrypted, err := aead.Decrypt(ciphertext, nil)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypt() = %v, want %v", decrypted, plaintext)
	}
}

func TestXChaCha20Poly1305(t *testing.T) {
	key, _ := SecureKey(KeySize256)
	aead, err := NewXChaCha20Poly1305(key)
	if err != nil {
		t.Fatalf("NewXChaCha20Poly1305() error = %v", err)
	}
	defer aead.Wipe()

	plaintext := []byte("Hello, XChaCha!")

	ciphertext, err := aead.Encrypt(plaintext, nil)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	decrypted, err := aead.Decrypt(ciphertext, nil)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypt() = %v, want %v", decrypted, plaintext)
	}
}

// =============================================================================
// Password Hashing Tests
// =============================================================================

func TestHashPassword(t *testing.T) {
	password := "SecureP@ssw0rd123!"

	hash, err := HashPassword(password, nil)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	// Hash should start with $argon2id$
	if hash[:9] != "$argon2id" {
		t.Errorf("HashPassword() hash format incorrect: %s", hash[:20])
	}

	// Same password should produce different hashes (due to random salt)
	hash2, _ := HashPassword(password, nil)
	if hash == hash2 {
		t.Error("HashPassword() produced identical hashes for same password")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "SecureP@ssw0rd123!"
	hash, _ := HashPassword(password, nil)

	// Correct password should verify
	valid, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}
	if !valid {
		t.Error("VerifyPassword() returned false for correct password")
	}

	// Wrong password should not verify
	valid, err = VerifyPassword("wrong-password", hash)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}
	if valid {
		t.Error("VerifyPassword() returned true for wrong password")
	}
}

func TestVerifyPasswordInvalidHash(t *testing.T) {
	_, err := VerifyPassword("password", "invalid-hash-format")
	if err == nil {
		t.Error("VerifyPassword() should fail with invalid hash format")
	}
}

// =============================================================================
// Key Derivation Tests
// =============================================================================

func TestDeriveKey(t *testing.T) {
	password := []byte("my-secret-password")
	salt, _ := SecureBytes(SaltSize)

	key := DeriveKey(password, salt, nil)

	// Key should be 32 bytes by default
	if len(key) != 32 {
		t.Errorf("DeriveKey() returned %d bytes, want 32", len(key))
	}

	// Same inputs should produce same output
	key2 := DeriveKey(password, salt, nil)
	if !bytes.Equal(key, key2) {
		t.Error("DeriveKey() produced different keys for same inputs")
	}

	// Different salt should produce different key
	salt2, _ := SecureBytes(SaltSize)
	key3 := DeriveKey(password, salt2, nil)
	if bytes.Equal(key, key3) {
		t.Error("DeriveKey() produced same key with different salt")
	}
}

func TestDeriveKeyWithInfo(t *testing.T) {
	secret, _ := SecureBytes(32)
	salt, _ := SecureBytes(16)
	info := []byte("aureo-vpn-encryption")

	key, err := DeriveKeyWithInfo(secret, salt, info, 32)
	if err != nil {
		t.Fatalf("DeriveKeyWithInfo() error = %v", err)
	}

	if len(key) != 32 {
		t.Errorf("DeriveKeyWithInfo() returned %d bytes, want 32", len(key))
	}

	// Different info should produce different key
	key2, _ := DeriveKeyWithInfo(secret, salt, []byte("different-info"), 32)
	if bytes.Equal(key, key2) {
		t.Error("DeriveKeyWithInfo() produced same key with different info")
	}
}

// =============================================================================
// Constant Time Comparison Tests
// =============================================================================

func TestConstantTimeCompare(t *testing.T) {
	a := []byte("secret-data")
	b := []byte("secret-data")
	c := []byte("different-data")

	if !ConstantTimeCompare(a, b) {
		t.Error("ConstantTimeCompare() returned false for equal data")
	}

	if ConstantTimeCompare(a, c) {
		t.Error("ConstantTimeCompare() returned true for different data")
	}
}

func TestConstantTimeStringCompare(t *testing.T) {
	if !ConstantTimeStringCompare("test", "test") {
		t.Error("ConstantTimeStringCompare() returned false for equal strings")
	}

	if ConstantTimeStringCompare("test", "different") {
		t.Error("ConstantTimeStringCompare() returned true for different strings")
	}
}

// =============================================================================
// Key Rotation Tests
// =============================================================================

func TestKeyRotation(t *testing.T) {
	key1, _ := SecureKey(KeySize256)
	kr, err := NewKeyRotation(key1, 1*time.Hour)
	if err != nil {
		t.Fatalf("NewKeyRotation() error = %v", err)
	}
	defer kr.Wipe()

	// Current key should match initial key
	if !bytes.Equal(kr.CurrentKey(), key1) {
		t.Error("CurrentKey() does not match initial key")
	}

	// Rotate to new key
	key2, _ := SecureKey(KeySize256)
	if err := kr.Rotate(key2); err != nil {
		t.Fatalf("Rotate() error = %v", err)
	}

	// Current key should now be key2
	if !bytes.Equal(kr.CurrentKey(), key2) {
		t.Error("CurrentKey() does not match rotated key")
	}
}

func TestKeyRotationInvalidKey(t *testing.T) {
	key, _ := SecureKey(KeySize256)
	kr, _ := NewKeyRotation(key, 1*time.Hour)
	defer kr.Wipe()

	// Rotating with invalid key size should fail
	err := kr.Rotate([]byte("short-key"))
	if err == nil {
		t.Error("Rotate() should fail with invalid key size")
	}
}

// =============================================================================
// Hybrid Key Exchange Tests
// =============================================================================

func TestHybridKeyExchange(t *testing.T) {
	// Create two key exchange instances (simulating client and server)
	alice, err := NewHybridKeyExchange()
	if err != nil {
		t.Fatalf("NewHybridKeyExchange() for Alice error = %v", err)
	}
	defer alice.Wipe()

	bob, err := NewHybridKeyExchange()
	if err != nil {
		t.Fatalf("NewHybridKeyExchange() for Bob error = %v", err)
	}
	defer bob.Wipe()

	// Exchange public keys
	alicePublic := alice.PublicKey()
	bobPublic := bob.PublicKey()

	// Derive shared secrets
	aliceSecret, err := alice.DeriveSharedSecret(bobPublic)
	if err != nil {
		t.Fatalf("Alice DeriveSharedSecret() error = %v", err)
	}

	bobSecret, err := bob.DeriveSharedSecret(alicePublic)
	if err != nil {
		t.Fatalf("Bob DeriveSharedSecret() error = %v", err)
	}

	// Shared secrets should be identical
	if !bytes.Equal(aliceSecret, bobSecret) {
		t.Error("Shared secrets do not match")
	}
}

func TestHybridKeyExchangeInvalidPublicKey(t *testing.T) {
	hke, _ := NewHybridKeyExchange()
	defer hke.Wipe()

	// Invalid public key should fail
	_, err := hke.DeriveSharedSecret([]byte("invalid-key"))
	if err == nil {
		t.Error("DeriveSharedSecret() should fail with invalid public key")
	}
}

// =============================================================================
// Hash Tests
// =============================================================================

func TestSecureHash(t *testing.T) {
	data := []byte("test data")
	hash := SecureHash(data)

	// SHA-256 produces 32 bytes
	if len(hash) != 32 {
		t.Errorf("SecureHash() returned %d bytes, want 32", len(hash))
	}

	// Same data should produce same hash
	hash2 := SecureHash(data)
	if !bytes.Equal(hash, hash2) {
		t.Error("SecureHash() produced different hashes for same data")
	}

	// Different data should produce different hash
	hash3 := SecureHash([]byte("different data"))
	if bytes.Equal(hash, hash3) {
		t.Error("SecureHash() produced same hash for different data")
	}
}

func TestSecureHashHex(t *testing.T) {
	data := []byte("test data")
	hexHash := SecureHashHex(data)

	// SHA-256 hex is 64 characters
	if len(hexHash) != 64 {
		t.Errorf("SecureHashHex() returned %d chars, want 64", len(hexHash))
	}
}

func TestSecureHash512(t *testing.T) {
	data := []byte("test data")
	hash := SecureHash512(data)

	// SHA-512 produces 64 bytes
	if len(hash) != 64 {
		t.Errorf("SecureHash512() returned %d bytes, want 64", len(hash))
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkSecureBytes32(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SecureBytes(32)
	}
}

func BenchmarkHashPassword(b *testing.B) {
	password := "BenchmarkPassword123!"
	for i := 0; i < b.N; i++ {
		HashPassword(password, nil)
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	password := "BenchmarkPassword123!"
	hash, _ := HashPassword(password, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifyPassword(password, hash)
	}
}

func BenchmarkAESGCMEncrypt(b *testing.B) {
	key, _ := SecureKey(KeySize256)
	aead, _ := NewAESGCM(key)
	plaintext := make([]byte, 1024)
	rand.Read(plaintext)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		aead.Encrypt(plaintext, nil)
	}
}

func BenchmarkChaCha20Poly1305Encrypt(b *testing.B) {
	key, _ := SecureKey(KeySize256)
	aead, _ := NewChaCha20Poly1305(key)
	plaintext := make([]byte, 1024)
	rand.Read(plaintext)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		aead.Encrypt(plaintext, nil)
	}
}
