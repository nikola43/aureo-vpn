package payment

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nikola43/aureo-vpn/pkg/database"
	"github.com/nikola43/aureo-vpn/pkg/models"
	"gorm.io/gorm"
)

// CryptoPaymentProcessor handles cryptocurrency payments
type CryptoPaymentProcessor struct {
	db              *gorm.DB
	btcAddressPool  []string
	ethAddressPool  []string
	ltcAddressPool  []string
	xmrAddressPool  []string
	confirmations   map[string]int
}

// Payment represents a cryptocurrency payment
type Payment struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key"`
	UserID          uuid.UUID `gorm:"type:uuid"`
	Cryptocurrency  string    // BTC, ETH, LTC, XMR
	Amount          float64
	AmountCrypto    float64
	Address         string // Payment address
	TxHash          string // Transaction hash
	Status          string // pending, confirmed, failed, expired
	Confirmations   int
	RequiredConf    int
	SubscriptionTier string
	Duration        int // months
	ExpiresAt       time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewCryptoPaymentProcessor creates a new crypto payment processor
func NewCryptoPaymentProcessor() *CryptoPaymentProcessor {
	return &CryptoPaymentProcessor{
		db: database.GetDB(),
		confirmations: map[string]int{
			"BTC": 3,
			"ETH": 12,
			"LTC": 6,
			"XMR": 10,
		},
	}
}

// CreatePayment creates a new cryptocurrency payment
func (p *CryptoPaymentProcessor) CreatePayment(userID uuid.UUID, crypto, tier string, duration int) (*Payment, error) {
	// Calculate amount based on tier and duration
	amount := p.calculateAmount(tier, duration)

	// Get current crypto rate
	cryptoAmount, err := p.convertToCrypto(amount, crypto)
	if err != nil {
		return nil, err
	}

	// Generate payment address
	address, err := p.generatePaymentAddress(crypto)
	if err != nil {
		return nil, err
	}

	payment := &Payment{
		ID:               uuid.New(),
		UserID:           userID,
		Cryptocurrency:   crypto,
		Amount:           amount,
		AmountCrypto:     cryptoAmount,
		Address:          address,
		Status:           "pending",
		Confirmations:    0,
		RequiredConf:     p.confirmations[crypto],
		SubscriptionTier: tier,
		Duration:         duration,
		ExpiresAt:        time.Now().Add(24 * time.Hour), // Payment expires in 24h
		CreatedAt:        time.Now(),
	}

	// Save to database
	if err := p.db.Table("payments").Create(payment).Error; err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	log.Printf("Created payment: %s - %s %f %s", payment.ID, tier, cryptoAmount, crypto)
	return payment, nil
}

// CheckPayment checks the status of a payment
func (p *CryptoPaymentProcessor) CheckPayment(paymentID uuid.UUID) (*Payment, error) {
	var payment Payment
	if err := p.db.Table("payments").First(&payment, paymentID).Error; err != nil {
		return nil, err
	}

	// If payment is pending, check blockchain
	if payment.Status == "pending" {
		confirmed, confirmations, txHash, err := p.checkBlockchain(payment.Address, payment.Cryptocurrency, payment.AmountCrypto)
		if err != nil {
			log.Printf("Failed to check blockchain: %v", err)
			return &payment, nil
		}

		// Update confirmations
		payment.Confirmations = confirmations
		payment.TxHash = txHash

		// If confirmed, update status
		if confirmed {
			payment.Status = "confirmed"
			p.activateSubscription(payment.UserID, payment.SubscriptionTier, payment.Duration)
		}

		// Save updated payment
		p.db.Table("payments").Save(&payment)
	}

	return &payment, nil
}

// calculateAmount calculates the USD amount for a subscription
func (p *CryptoPaymentProcessor) calculateAmount(tier string, duration int) float64 {
	basePrices := map[string]float64{
		"basic":   9.99,
		"premium": 12.99,
		"ultimate": 15.99,
	}

	basePrice := basePrices[tier]
	if basePrice == 0 {
		basePrice = 9.99
	}

	// Apply discounts for longer subscriptions
	discount := 1.0
	switch duration {
	case 6:
		discount = 0.85 // 15% off
	case 12:
		discount = 0.70 // 30% off
	case 24:
		discount = 0.60 // 40% off
	}

	return basePrice * float64(duration) * discount
}

// convertToCrypto converts USD to cryptocurrency amount
func (p *CryptoPaymentProcessor) convertToCrypto(usdAmount float64, crypto string) (float64, error) {
	// In production, fetch real-time rates from exchange API
	// For now, use approximate rates
	rates := map[string]float64{
		"BTC": 45000.0,
		"ETH": 3000.0,
		"LTC": 100.0,
		"XMR": 150.0,
	}

	rate, ok := rates[crypto]
	if !ok {
		return 0, fmt.Errorf("unsupported cryptocurrency: %s", crypto)
	}

	return usdAmount / rate, nil
}

// generatePaymentAddress generates a unique payment address
func (p *CryptoPaymentProcessor) generatePaymentAddress(crypto string) (string, error) {
	// In production, generate real addresses using HD wallets
	// For now, generate mock addresses

	prefix := map[string]string{
		"BTC": "bc1",
		"ETH": "0x",
		"LTC": "ltc1",
		"XMR": "4",
	}

	// Generate unique address
	hash := sha256.Sum256([]byte(uuid.New().String() + time.Now().String()))
	address := prefix[crypto] + hex.EncodeToString(hash[:])[:32]

	return address, nil
}

// checkBlockchain checks if payment is received on blockchain
func (p *CryptoPaymentProcessor) checkBlockchain(address, crypto string, expectedAmount float64) (bool, int, string, error) {
	// In production, integrate with blockchain APIs:
	// - Bitcoin: BlockCypher, Blockchain.info
	// - Ethereum: Etherscan, Infura
	// - Litecoin: BlockCypher
	// - Monero: Monero RPC

	// Mock implementation for demonstration
	// Return: confirmed, confirmations, txHash, error

	// Simulate checking (in production, make actual API calls)
	log.Printf("Checking blockchain for %s on %s address", crypto, address)

	// For demonstration, return pending
	return false, 0, "", nil
}

// activateSubscription activates a user's subscription
func (p *CryptoPaymentProcessor) activateSubscription(userID uuid.UUID, tier string, duration int) error {
	var user models.User
	if err := p.db.First(&user, userID).Error; err != nil {
		return err
	}

	// Update subscription
	user.SubscriptionTier = tier
	user.SubscriptionExpiry = time.Now().AddDate(0, duration, 0)

	if err := p.db.Save(&user).Error; err != nil {
		return err
	}

	log.Printf("Activated %s subscription for user %s (duration: %d months)", tier, userID, duration)
	return nil
}

// GetPaymentHistory returns payment history for a user
func (p *CryptoPaymentProcessor) GetPaymentHistory(userID uuid.UUID) ([]Payment, error) {
	var payments []Payment
	if err := p.db.Table("payments").Where("user_id = ?", userID).Order("created_at DESC").Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}

// GenerateInvoice generates an invoice for a payment
func (p *CryptoPaymentProcessor) GenerateInvoice(payment *Payment) string {
	return fmt.Sprintf(`
INVOICE
=======
Payment ID: %s
Subscription: %s (%d months)
Amount: $%.2f USD
Pay: %.8f %s
Address: %s
Expires: %s
Status: %s

Scan QR code or send payment to the address above.
Payment will be confirmed after %d blockchain confirmations.
`,
		payment.ID,
		payment.SubscriptionTier,
		payment.Duration,
		payment.Amount,
		payment.AmountCrypto,
		payment.Cryptocurrency,
		payment.Address,
		payment.ExpiresAt.Format(time.RFC3339),
		payment.Status,
		payment.RequiredConf,
	)
}

// GetSupportedCryptocurrencies returns list of supported cryptocurrencies
func (p *CryptoPaymentProcessor) GetSupportedCryptocurrencies() []CryptoCurrency {
	return []CryptoCurrency{
		{
			Symbol:      "BTC",
			Name:        "Bitcoin",
			Network:     "Bitcoin",
			Confirmations: 3,
		},
		{
			Symbol:      "ETH",
			Name:        "Ethereum",
			Network:     "Ethereum",
			Confirmations: 12,
		},
		{
			Symbol:      "LTC",
			Name:        "Litecoin",
			Network:     "Litecoin",
			Confirmations: 6,
		},
		{
			Symbol:      "XMR",
			Name:        "Monero",
			Network:     "Monero",
			Confirmations: 10,
		},
	}
}

// CryptoCurrency represents a supported cryptocurrency
type CryptoCurrency struct {
	Symbol        string
	Name          string
	Network       string
	Confirmations int
}

// VerifyPaymentSignature verifies webhook payment signature
func (p *CryptoPaymentProcessor) VerifyPaymentSignature(payload, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// RefundPayment processes a refund (for failed payments)
func (p *CryptoPaymentProcessor) RefundPayment(paymentID uuid.UUID) error {
	var payment Payment
	if err := p.db.Table("payments").First(&payment, paymentID).Error; err != nil {
		return err
	}

	// Mark as refunded
	payment.Status = "refunded"
	payment.UpdatedAt = time.Now()

	return p.db.Table("payments").Save(&payment).Error
}

// GetPaymentQRCode generates a QR code for payment
func (p *CryptoPaymentProcessor) GetPaymentQRCode(payment *Payment) (string, error) {
	// Generate payment URI for QR code
	uri := fmt.Sprintf("%s:%s?amount=%.8f&label=Aureo+VPN+Subscription",
		payment.Cryptocurrency,
		payment.Address,
		payment.AmountCrypto,
	)

	// In production, use a QR code library to generate actual QR code
	// For now, return the URI
	return uri, nil
}
