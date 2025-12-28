// Package blockchain provides price oracle functionality for cryptocurrency conversions
package blockchain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

// ErrPriceUnavailable indicates price data could not be fetched
var ErrPriceUnavailable = errors.New("cryptocurrency price unavailable")

// PriceOracle provides cryptocurrency price data from external sources
type PriceOracle struct {
	client     *http.Client
	cache      map[string]priceCache
	cacheMu    sync.RWMutex
	cacheTTL   time.Duration
	apiKey     string
	apiBaseURL string
}

type priceCache struct {
	price     float64
	timestamp time.Time
}

// CoinGeckoResponse represents the API response from CoinGecko
type CoinGeckoResponse struct {
	Bitcoin  map[string]float64 `json:"bitcoin,omitempty"`
	Ethereum map[string]float64 `json:"ethereum,omitempty"`
	Litecoin map[string]float64 `json:"litecoin,omitempty"`
}

// PriceOracleConfig holds configuration for the price oracle
type PriceOracleConfig struct {
	// CoinGecko API settings
	APIKey     string        // Optional API key for CoinGecko Pro
	CacheTTL   time.Duration // How long to cache prices (default 1 minute)
	Timeout    time.Duration // HTTP request timeout (default 10 seconds)
}

// DefaultPriceOracleConfig returns secure defaults
func DefaultPriceOracleConfig() PriceOracleConfig {
	return PriceOracleConfig{
		APIKey:   os.Getenv("COINGECKO_API_KEY"),
		CacheTTL: 1 * time.Minute,
		Timeout:  10 * time.Second,
	}
}

// NewPriceOracle creates a new price oracle
func NewPriceOracle(cfg PriceOracleConfig) *PriceOracle {
	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = 1 * time.Minute
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}

	baseURL := "https://api.coingecko.com/api/v3"
	if cfg.APIKey != "" {
		baseURL = "https://pro-api.coingecko.com/api/v3"
	}

	return &PriceOracle{
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		cache:      make(map[string]priceCache),
		cacheTTL:   cfg.CacheTTL,
		apiKey:     cfg.APIKey,
		apiBaseURL: baseURL,
	}
}

// GetPrice returns the current USD price for a cryptocurrency
// Uses caching to avoid excessive API calls
func (po *PriceOracle) GetPrice(ctx context.Context, currency string) (float64, error) {
	// Normalize currency name
	coinID := normalizeCurrency(currency)
	if coinID == "" {
		return 0, fmt.Errorf("unsupported currency: %s", currency)
	}

	// Check cache first
	if price, ok := po.getCachedPrice(coinID); ok {
		return price, nil
	}

	// Fetch from API
	price, err := po.fetchPrice(ctx, coinID)
	if err != nil {
		return 0, fmt.Errorf("%w: %s", ErrPriceUnavailable, err.Error())
	}

	// Cache the result
	po.cachePrice(coinID, price)

	return price, nil
}

// GetPriceWithFallback returns price with environment variable fallback
// SECURITY: Fallback should only be used for development/testing
func (po *PriceOracle) GetPriceWithFallback(ctx context.Context, currency string) (float64, error) {
	price, err := po.GetPrice(ctx, currency)
	if err == nil {
		return price, nil
	}

	// Check for environment variable fallback (development only)
	envKey := fmt.Sprintf("%s_PRICE_USD", normalizeCurrencyEnv(currency))
	if fallbackStr := os.Getenv(envKey); fallbackStr != "" {
		var fallback float64
		if _, parseErr := fmt.Sscanf(fallbackStr, "%f", &fallback); parseErr == nil && fallback > 0 {
			return fallback, nil
		}
	}

	return 0, err
}

func (po *PriceOracle) getCachedPrice(coinID string) (float64, bool) {
	po.cacheMu.RLock()
	defer po.cacheMu.RUnlock()

	cached, exists := po.cache[coinID]
	if !exists {
		return 0, false
	}

	if time.Since(cached.timestamp) > po.cacheTTL {
		return 0, false
	}

	return cached.price, true
}

func (po *PriceOracle) cachePrice(coinID string, price float64) {
	po.cacheMu.Lock()
	defer po.cacheMu.Unlock()

	po.cache[coinID] = priceCache{
		price:     price,
		timestamp: time.Now(),
	}
}

func (po *PriceOracle) fetchPrice(ctx context.Context, coinID string) (float64, error) {
	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd", po.apiBaseURL, coinID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Accept", "application/json")
	if po.apiKey != "" {
		req.Header.Set("x-cg-pro-api-key", po.apiKey)
	}

	resp, err := po.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	coinData, ok := result[coinID]
	if !ok {
		return 0, fmt.Errorf("no price data for %s", coinID)
	}

	price, ok := coinData["usd"]
	if !ok {
		return 0, fmt.Errorf("no USD price for %s", coinID)
	}

	if price <= 0 {
		return 0, fmt.Errorf("invalid price for %s: %f", coinID, price)
	}

	return price, nil
}

func normalizeCurrency(currency string) string {
	switch currency {
	case "ethereum", "eth", "ETH":
		return "ethereum"
	case "bitcoin", "btc", "BTC":
		return "bitcoin"
	case "litecoin", "ltc", "LTC":
		return "litecoin"
	default:
		return ""
	}
}

func normalizeCurrencyEnv(currency string) string {
	switch normalizeCurrency(currency) {
	case "ethereum":
		return "ETH"
	case "bitcoin":
		return "BTC"
	case "litecoin":
		return "LTC"
	default:
		return ""
	}
}

// ConvertUSDToCrypto converts USD amount to cryptocurrency amount
// Returns amount and any error encountered
func (po *PriceOracle) ConvertUSDToCrypto(ctx context.Context, currency string, amountUSD float64) (float64, error) {
	if amountUSD <= 0 {
		return 0, errors.New("amount must be positive")
	}

	price, err := po.GetPrice(ctx, currency)
	if err != nil {
		return 0, err
	}

	cryptoAmount := amountUSD / price
	return cryptoAmount, nil
}

// Global price oracle instance
var globalPriceOracle *PriceOracle
var priceOracleOnce sync.Once

// GetPriceOracle returns the global price oracle instance
func GetPriceOracle() *PriceOracle {
	priceOracleOnce.Do(func() {
		globalPriceOracle = NewPriceOracle(DefaultPriceOracleConfig())
	})
	return globalPriceOracle
}
