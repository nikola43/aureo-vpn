package transparency

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"
)

// WarrantCanary represents a warrant canary statement
type WarrantCanary struct {
	ID              string
	Statement       string
	IssuedDate      time.Time
	ExpiryDate      time.Time
	SignerName      string
	SignerTitle     string
	Signature       string
	PublicKey       string
	Status          string // active, expired, compromised
	LastVerified    time.Time
	VerificationURL string
}

// CanaryManager manages warrant canary operations
type CanaryManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// NewCanaryManager creates a new canary manager
func NewCanaryManager() (*CanaryManager, error) {
	// Generate RSA key pair for signing
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	return &CanaryManager{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}, nil
}

// GenerateCanary generates a new warrant canary
func (cm *CanaryManager) GenerateCanary(signerName, signerTitle string) (*WarrantCanary, error) {
	now := time.Now()
	expiry := now.AddDate(0, 3, 0) // Valid for 3 months

	statement := cm.generateStatement(now, expiry)

	// Sign the statement
	signature, err := cm.signStatement(statement)
	if err != nil {
		return nil, fmt.Errorf("failed to sign statement: %w", err)
	}

	// Export public key
	publicKeyPEM, err := cm.exportPublicKey()
	if err != nil {
		return nil, fmt.Errorf("failed to export public key: %w", err)
	}

	canary := &WarrantCanary{
		ID:              fmt.Sprintf("CANARY-%s", now.Format("2006-01-02")),
		Statement:       statement,
		IssuedDate:      now,
		ExpiryDate:      expiry,
		SignerName:      signerName,
		SignerTitle:     signerTitle,
		Signature:       signature,
		PublicKey:       publicKeyPEM,
		Status:          "active",
		LastVerified:    now,
		VerificationURL: "https://aureo-vpn.com/transparency/canary",
	}

	return canary, nil
}

// generateStatement generates the canary statement text
func (cm *CanaryManager) generateStatement(issued, expiry time.Time) string {
	return fmt.Sprintf(`WARRANT CANARY
================

Issued: %s
Expires: %s

This canary will be updated every 90 days.

As of the date above, Aureo VPN:

✓ Has NOT received any National Security Letters
✓ Has NOT received any gag orders
✓ Has NOT been subject to any warrant for user data
✓ Has NOT been forced to modify our systems to facilitate surveillance
✓ Has NOT received any requests to implement backdoors
✓ Has NOT been forced to log user activity beyond what is disclosed in our privacy policy
✓ Has NOT received any court orders requiring us to identify users
✓ Has NOT been contacted by any government agency requesting user information
✓ Has NOT been prohibited from updating this canary
✓ Has NOT been forced to hand over encryption keys
✓ Is NOT under any ongoing investigation that would prevent disclosure
✓ Has NOT received any classified information requests

If this canary is not updated within 14 days of the expiry date,
or if the statements above change, assume that we are under legal
obligation to not disclose certain information.

Our commitment to transparency and user privacy remains unchanged.
We will continue to resist any attempts to compromise user security.

For verification:
- This statement is cryptographically signed
- Public key is published on our website
- Signature can be independently verified
- Canary is archived on the Internet Archive

Contact: transparency@aureo-vpn.com
PGP Key: [KEY ID]

---
This canary should be read in conjunction with our Privacy Policy,
Transparency Report, and Terms of Service.
`, issued.Format("2006-01-02 15:04:05 MST"), expiry.Format("2006-01-02 15:04:05 MST"))
}

// signStatement signs the canary statement
func (cm *CanaryManager) signStatement(statement string) (string, error) {
	hash := sha256.Sum256([]byte(statement))

	signature, err := rsa.SignPKCS1v15(rand.Reader, cm.privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}

	// Encode signature as hex
	return fmt.Sprintf("%x", signature), nil
}

// VerifyCanary verifies the signature of a canary
func (cm *CanaryManager) VerifyCanary(canary *WarrantCanary) (bool, error) {
	hash := sha256.Sum256([]byte(canary.Statement))

	// Decode hex signature
	var signature []byte
	_, err := fmt.Sscanf(canary.Signature, "%x", &signature)
	if err != nil {
		return false, err
	}

	// Verify signature
	err = rsa.VerifyPKCS1v15(cm.publicKey, crypto.SHA256, hash[:], signature)
	return err == nil, err
}

// exportPublicKey exports the public key in PEM format
func (cm *CanaryManager) exportPublicKey() (string, error) {
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(cm.publicKey)
	if err != nil {
		return "", err
	}

	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	return string(pubKeyPEM), nil
}

// IsExpired checks if the canary has expired
func (c *WarrantCanary) IsExpired() bool {
	return time.Now().After(c.ExpiryDate)
}

// DaysUntilExpiry returns days until canary expires
func (c *WarrantCanary) DaysUntilExpiry() int {
	duration := time.Until(c.ExpiryDate)
	return int(duration.Hours() / 24)
}

// GetHTMLVersion returns HTML-formatted canary for web display
func (c *WarrantCanary) GetHTMLVersion() string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Aureo VPN - Warrant Canary</title>
    <style>
        body { font-family: monospace; max-width: 800px; margin: 50px auto; padding: 20px; }
        .header { background: #2c3e50; color: white; padding: 20px; }
        .statement { background: #ecf0f1; padding: 20px; margin: 20px 0; white-space: pre-wrap; }
        .signature { background: #34495e; color: #ecf0f1; padding: 10px; font-size: 10px; word-break: break-all; }
        .status { padding: 10px; margin: 10px 0; }
        .active { background: #2ecc71; color: white; }
        .expired { background: #e74c3c; color: white; }
        .warning { background: #f39c12; color: white; padding: 15px; margin: 15px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Aureo VPN Warrant Canary</h1>
        <p>ID: %s</p>
        <p>Issued: %s | Expires: %s</p>
    </div>

    <div class="status %s">
        Status: %s | Days until expiry: %d
    </div>

    %s

    <div class="statement">%s</div>

    <h3>Cryptographic Signature</h3>
    <div class="signature">%s</div>

    <h3>Verification</h3>
    <p>To verify this canary:</p>
    <ol>
        <li>Download our public key from <a href="%s">%s</a></li>
        <li>Verify the signature using standard cryptographic tools</li>
        <li>Check that this page hasn't been modified using Internet Archive</li>
    </ol>

    <p><strong>Signed by:</strong> %s, %s</p>
    <p><strong>Last verified:</strong> %s</p>
</body>
</html>`,
		c.ID,
		c.IssuedDate.Format("2006-01-02 15:04:05 MST"),
		c.ExpiryDate.Format("2006-01-02 15:04:05 MST"),
		c.Status,
		c.Status,
		c.DaysUntilExpiry(),
		c.getWarningMessage(),
		c.Statement,
		c.Signature,
		c.VerificationURL,
		c.VerificationURL,
		c.SignerName,
		c.SignerTitle,
		c.LastVerified.Format("2006-01-02 15:04:05 MST"),
	)
}

// getWarningMessage returns a warning if canary is expiring soon or expired
func (c *WarrantCanary) getWarningMessage() string {
	daysLeft := c.DaysUntilExpiry()

	if daysLeft < 0 {
		return `<div class="warning">
        ⚠️ WARNING: This canary has EXPIRED! This may indicate that we are under
        legal obligation not to disclose certain information. Please read our
        latest transparency report and privacy policy for more information.
    </div>`
	}

	if daysLeft <= 14 {
		return fmt.Sprintf(`<div class="warning">
        ⚠️ NOTICE: This canary expires in %d days. A new canary should be
        published soon. If no update appears within 14 days of expiry,
        assume we are under legal constraint.
    </div>`, daysLeft)
	}

	return ""
}

// GetTextVersion returns plain text version for archival
func (c *WarrantCanary) GetTextVersion() string {
	return fmt.Sprintf(`%s

---
Signature: %s

Public Key:
%s

Signer: %s, %s
Issued: %s
Expires: %s
Status: %s
Verification URL: %s
`,
		c.Statement,
		c.Signature,
		c.PublicKey,
		c.SignerName,
		c.SignerTitle,
		c.IssuedDate.Format(time.RFC3339),
		c.ExpiryDate.Format(time.RFC3339),
		c.Status,
		c.VerificationURL,
	)
}

// ArchiveCanary prepares canary for archival on Internet Archive
func (c *WarrantCanary) ArchiveCanary() map[string]interface{} {
	return map[string]interface{}{
		"id":               c.ID,
		"statement":        c.Statement,
		"signature":        c.Signature,
		"public_key":       c.PublicKey,
		"issued_date":      c.IssuedDate.Unix(),
		"expiry_date":      c.ExpiryDate.Unix(),
		"signer_name":      c.SignerName,
		"signer_title":     c.SignerTitle,
		"status":           c.Status,
		"verification_url": c.VerificationURL,
	}
}
