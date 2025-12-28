package crypto

import (
	"bytes"
	"crypto/rand"
	"testing"
	"time"
)

// =============================================================================
// TunnelCipher Tests
// =============================================================================

func TestNewTunnelCipher(t *testing.T) {
	keys := generateTestKeys(t)
	defer keys.Wipe()

	cipher, err := NewTunnelCipher(keys)
	if err != nil {
		t.Fatalf("NewTunnelCipher() error = %v", err)
	}
	defer cipher.Close()

	if cipher == nil {
		t.Fatal("NewTunnelCipher() returned nil")
	}
}

func TestTunnelCipherEncryptDecrypt(t *testing.T) {
	keys := generateTestKeys(t)
	defer keys.Wipe()

	cipher, err := NewTunnelCipher(keys)
	if err != nil {
		t.Fatalf("NewTunnelCipher() error = %v", err)
	}
	defer cipher.Close()

	testCases := []struct {
		name      string
		plaintext []byte
	}{
		{"empty", []byte{}},
		{"small", []byte("hello world")},
		{"medium", bytes.Repeat([]byte("test"), 100)},
		{"large", bytes.Repeat([]byte("data"), 10000)},
		{"binary", generateRandomBytes(t, 1024)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encrypted, err := cipher.Encrypt(tc.plaintext, PacketTypeData)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			// Encrypted should be different from plaintext
			if len(tc.plaintext) > 0 && bytes.Equal(encrypted, tc.plaintext) {
				t.Error("Encrypt() did not encrypt data")
			}

			decrypted, packetType, err := cipher.Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if packetType != PacketTypeData {
				t.Errorf("Decrypt() packetType = %v, want %v", packetType, PacketTypeData)
			}

			if !bytes.Equal(decrypted, tc.plaintext) {
				t.Errorf("Decrypt() mismatch: got %d bytes, want %d bytes", len(decrypted), len(tc.plaintext))
			}
		})
	}
}

func TestTunnelCipherPacketTypes(t *testing.T) {
	keys := generateTestKeys(t)
	defer keys.Wipe()

	cipher, err := NewTunnelCipher(keys)
	if err != nil {
		t.Fatalf("NewTunnelCipher() error = %v", err)
	}
	defer cipher.Close()

	packetTypes := []PacketType{
		PacketTypeData,
		PacketTypeControl,
		PacketTypeKeepAlive,
		PacketTypeRekey,
	}

	for _, pt := range packetTypes {
		t.Run(pt.String(), func(t *testing.T) {
			plaintext := []byte("test data")
			encrypted, err := cipher.Encrypt(plaintext, pt)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			_, packetType, err := cipher.Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if packetType != pt {
				t.Errorf("packetType = %v, want %v", packetType, pt)
			}
		})
	}
}

func TestTunnelCipherReplayProtection(t *testing.T) {
	keys := generateTestKeys(t)
	defer keys.Wipe()

	cipher, err := NewTunnelCipher(keys)
	if err != nil {
		t.Fatalf("NewTunnelCipher() error = %v", err)
	}
	defer cipher.Close()

	plaintext := []byte("replay test")
	encrypted, err := cipher.Encrypt(plaintext, PacketTypeData)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// First decrypt should succeed
	_, _, err = cipher.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("First Decrypt() error = %v", err)
	}

	// Second decrypt of same packet should fail (replay)
	_, _, err = cipher.Decrypt(encrypted)
	if err == nil {
		t.Error("Second Decrypt() should fail for replay")
	}
}

func TestTunnelCipherStats(t *testing.T) {
	keys := generateTestKeys(t)
	defer keys.Wipe()

	cipher, err := NewTunnelCipher(keys)
	if err != nil {
		t.Fatalf("NewTunnelCipher() error = %v", err)
	}
	defer cipher.Close()

	// Initial stats
	stats := cipher.Stats()
	if stats.PacketsSent != 0 || stats.PacketsReceived != 0 {
		t.Error("Initial stats should be zero")
	}

	// Encrypt some packets
	for i := 0; i < 10; i++ {
		_, err := cipher.Encrypt([]byte("test"), PacketTypeData)
		if err != nil {
			t.Fatalf("Encrypt() error = %v", err)
		}
	}

	stats = cipher.Stats()
	if stats.PacketsSent != 10 {
		t.Errorf("PacketsSent = %d, want 10", stats.PacketsSent)
	}
}

// =============================================================================
// KeyExchange Tests
// =============================================================================

func TestNewKeyExchange(t *testing.T) {
	ke, err := NewKeyExchange()
	if err != nil {
		t.Fatalf("NewKeyExchange() error = %v", err)
	}
	defer ke.Wipe()

	pubKey := ke.PublicKey()
	if len(pubKey) != 32 {
		t.Errorf("PublicKey() length = %d, want 32", len(pubKey))
	}
}

func TestKeyExchangeDeriveKeys(t *testing.T) {
	// Create two key exchanges (client and server)
	client, err := NewKeyExchange()
	if err != nil {
		t.Fatalf("NewKeyExchange() client error = %v", err)
	}
	defer client.Wipe()

	server, err := NewKeyExchange()
	if err != nil {
		t.Fatalf("NewKeyExchange() server error = %v", err)
	}
	defer server.Wipe()

	// Derive keys on both sides
	clientKeys, err := client.DeriveKeys(server.PublicKey(), true)
	if err != nil {
		t.Fatalf("DeriveKeys() client error = %v", err)
	}
	defer clientKeys.Wipe()

	serverKeys, err := server.DeriveKeys(client.PublicKey(), false)
	if err != nil {
		t.Fatalf("DeriveKeys() server error = %v", err)
	}
	defer serverKeys.Wipe()

	// Client's send should match server's receive
	if !bytes.Equal(clientKeys.PrimarySendKey, serverKeys.PrimaryRecvKey) {
		t.Error("Key mismatch: client send != server recv")
	}

	// Client's receive should match server's send
	if !bytes.Equal(clientKeys.PrimaryRecvKey, serverKeys.PrimarySendKey) {
		t.Error("Key mismatch: client recv != server send")
	}
}

// =============================================================================
// ReplayWindow Tests
// =============================================================================

func TestReplayWindow(t *testing.T) {
	rw := NewReplayWindow(1024)

	// First time seeing a sequence should be valid
	if !rw.Check(100) {
		t.Error("Check(100) should return true for new sequence")
	}

	// Mark it as seen
	rw.Update(100)

	// Second time should be invalid (replay)
	if rw.Check(100) {
		t.Error("Check(100) should return false for seen sequence")
	}

	// Higher sequence should be valid
	if !rw.Check(200) {
		t.Error("Check(200) should return true")
	}
	rw.Update(200)

	// Much lower sequence (beyond window) should be invalid
	if rw.Check(1) {
		t.Error("Check(1) should return false (too old)")
	}
}

func TestReplayWindowSliding(t *testing.T) {
	rw := NewReplayWindow(64)

	// Fill up the window
	for i := uint64(0); i < 64; i++ {
		if !rw.Check(i) {
			t.Errorf("Check(%d) should return true", i)
		}
		rw.Update(i)
	}

	// Advance beyond window
	for i := uint64(64); i < 128; i++ {
		if !rw.Check(i) {
			t.Errorf("Check(%d) should return true", i)
		}
		rw.Update(i)
	}

	// Old sequences should now be invalid
	if rw.Check(0) {
		t.Error("Check(0) should return false after window slides")
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func generateTestKeys(t *testing.T) *DerivedKeys {
	t.Helper()

	keys := &DerivedKeys{
		PrimarySendKey:   make([]byte, 32),
		PrimaryRecvKey:   make([]byte, 32),
		SecondarySendKey: make([]byte, 32),
		SecondaryRecvKey: make([]byte, 32),
	}

	rand.Read(keys.PrimarySendKey)
	rand.Read(keys.PrimaryRecvKey)
	rand.Read(keys.SecondarySendKey)
	rand.Read(keys.SecondaryRecvKey)
	rand.Read(keys.SessionID[:])

	return keys
}

func generateRandomBytes(t *testing.T, n int) []byte {
	t.Helper()
	b := make([]byte, n)
	rand.Read(b)
	return b
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkTunnelEncrypt(b *testing.B) {
	keys := &DerivedKeys{
		PrimarySendKey:   make([]byte, 32),
		PrimaryRecvKey:   make([]byte, 32),
		SecondarySendKey: make([]byte, 32),
		SecondaryRecvKey: make([]byte, 32),
	}
	rand.Read(keys.PrimarySendKey)
	rand.Read(keys.PrimaryRecvKey)
	rand.Read(keys.SecondarySendKey)
	rand.Read(keys.SecondaryRecvKey)
	rand.Read(keys.SessionID[:])

	cipher, _ := NewTunnelCipher(keys)
	defer cipher.Close()

	data := make([]byte, 1400)
	rand.Read(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cipher.Encrypt(data, PacketTypeData)
	}
}

func BenchmarkTunnelDecrypt(b *testing.B) {
	keys := &DerivedKeys{
		PrimarySendKey:   make([]byte, 32),
		PrimaryRecvKey:   make([]byte, 32),
		SecondarySendKey: make([]byte, 32),
		SecondaryRecvKey: make([]byte, 32),
	}
	rand.Read(keys.PrimarySendKey)
	rand.Read(keys.PrimaryRecvKey)
	rand.Read(keys.SecondarySendKey)
	rand.Read(keys.SecondaryRecvKey)
	rand.Read(keys.SessionID[:])

	cipher, _ := NewTunnelCipher(keys)
	defer cipher.Close()

	data := make([]byte, 1400)
	rand.Read(data)

	// Pre-encrypt many packets
	packets := make([][]byte, b.N+1)
	for i := 0; i <= b.N; i++ {
		encrypted, _ := cipher.Encrypt(data, PacketTypeData)
		packets[i] = encrypted
	}

	// Create new cipher for decryption (to reset replay window)
	cipher2, _ := NewTunnelCipher(keys)
	defer cipher2.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cipher2.Decrypt(packets[i])
	}
}

func BenchmarkKeyExchange(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ke, _ := NewKeyExchange()
		ke.Wipe()
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestTunnelCipherBidirectional(t *testing.T) {
	// Simulate client-server communication
	clientKE, _ := NewKeyExchange()
	defer clientKE.Wipe()

	serverKE, _ := NewKeyExchange()
	defer serverKE.Wipe()

	// Derive keys
	clientKeys, _ := clientKE.DeriveKeys(serverKE.PublicKey(), true)
	defer clientKeys.Wipe()

	serverKeys, _ := serverKE.DeriveKeys(clientKE.PublicKey(), false)
	defer serverKeys.Wipe()

	// Create ciphers
	clientCipher, _ := NewTunnelCipher(clientKeys)
	defer clientCipher.Close()

	serverCipher, _ := NewTunnelCipher(serverKeys)
	defer serverCipher.Close()

	// Client sends to server
	clientMsg := []byte("Hello from client")
	encrypted, _ := clientCipher.Encrypt(clientMsg, PacketTypeData)

	decrypted, _, err := serverCipher.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Server decrypt error: %v", err)
	}
	if !bytes.Equal(decrypted, clientMsg) {
		t.Error("Client->Server message mismatch")
	}

	// Server sends to client
	serverMsg := []byte("Hello from server")
	encrypted, _ = serverCipher.Encrypt(serverMsg, PacketTypeData)

	decrypted, _, err = clientCipher.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Client decrypt error: %v", err)
	}
	if !bytes.Equal(decrypted, serverMsg) {
		t.Error("Server->Client message mismatch")
	}
}

func TestTunnelCipherRekey(t *testing.T) {
	keys := generateTestKeys(t)
	defer keys.Wipe()

	cipher, err := NewTunnelCipher(keys)
	if err != nil {
		t.Fatalf("NewTunnelCipher() error = %v", err)
	}
	defer cipher.Close()

	// Encrypt some data before rekey
	plaintext := []byte("before rekey")
	_, err = cipher.Encrypt(plaintext, PacketTypeData)
	if err != nil {
		t.Fatalf("Encrypt() before rekey error = %v", err)
	}

	// Prepare rekey
	cipher.PrepareRekey()

	// Complete rekey with new keys
	newKeys := generateTestKeys(t)
	defer newKeys.Wipe()

	err = cipher.CompleteRekey(newKeys)
	if err != nil {
		t.Fatalf("CompleteRekey() error = %v", err)
	}

	// Encrypt data after rekey
	afterRekey := []byte("after rekey")
	encrypted, err := cipher.Encrypt(afterRekey, PacketTypeData)
	if err != nil {
		t.Fatalf("Encrypt() after rekey error = %v", err)
	}

	decrypted, _, err := cipher.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt() after rekey error = %v", err)
	}

	if !bytes.Equal(decrypted, afterRekey) {
		t.Error("Data mismatch after rekey")
	}
}

func TestTunnelCipherConcurrent(t *testing.T) {
	keys := generateTestKeys(t)
	defer keys.Wipe()

	cipher, err := NewTunnelCipher(keys)
	if err != nil {
		t.Fatalf("NewTunnelCipher() error = %v", err)
	}
	defer cipher.Close()

	done := make(chan bool)

	// Multiple concurrent writers
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				data := []byte("concurrent test data")
				_, err := cipher.Encrypt(data, PacketTypeData)
				if err != nil {
					t.Errorf("Goroutine %d: Encrypt() error = %v", id, err)
				}
				time.Sleep(time.Microsecond)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	stats := cipher.Stats()
	if stats.PacketsSent != 1000 {
		t.Errorf("PacketsSent = %d, want 1000", stats.PacketsSent)
	}
}
