package wireguard

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/curve25519"
)

// KeyPair represents a WireGuard key pair
type KeyPair struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

// GenerateKeyPair generates a new WireGuard key pair
func GenerateKeyPair() (*KeyPair, error) {
	// Generate random private key
	privateKey := make([]byte, curve25519.ScalarSize)
	if _, err := rand.Read(privateKey); err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Clamp the private key as per WireGuard specification
	privateKey[0] &= 248
	privateKey[31] &= 127
	privateKey[31] |= 64

	// Derive public key from private key
	publicKey, err := curve25519.X25519(privateKey, curve25519.Basepoint)
	if err != nil {
		return nil, fmt.Errorf("failed to generate public key: %w", err)
	}

	return &KeyPair{
		PrivateKey: base64.StdEncoding.EncodeToString(privateKey),
		PublicKey:  base64.StdEncoding.EncodeToString(publicKey),
	}, nil
}

// GeneratePresharedKey generates a preshared key for additional security
func GeneratePresharedKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("failed to generate preshared key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// ValidatePrivateKey validates a WireGuard private key
func ValidatePrivateKey(key string) error {
	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return fmt.Errorf("invalid base64 encoding: %w", err)
	}

	if len(decoded) != curve25519.ScalarSize {
		return fmt.Errorf("invalid key length: expected %d, got %d", curve25519.ScalarSize, len(decoded))
	}

	return nil
}

// ValidatePublicKey validates a WireGuard public key
func ValidatePublicKey(key string) error {
	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return fmt.Errorf("invalid base64 encoding: %w", err)
	}

	if len(decoded) != curve25519.PointSize {
		return fmt.Errorf("invalid key length: expected %d, got %d", curve25519.PointSize, len(decoded))
	}

	return nil
}

// DerivePublicKey derives a public key from a private key
func DerivePublicKey(privateKey string) (string, error) {
	privKeyBytes, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode private key: %w", err)
	}

	if len(privKeyBytes) != curve25519.ScalarSize {
		return "", fmt.Errorf("invalid private key length")
	}

	publicKey, err := curve25519.X25519(privKeyBytes, curve25519.Basepoint)
	if err != nil {
		return "", fmt.Errorf("failed to derive public key: %w", err)
	}

	return base64.StdEncoding.EncodeToString(publicKey), nil
}
