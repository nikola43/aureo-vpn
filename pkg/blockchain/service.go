package blockchain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/nikola43/aureo-vpn/pkg/logger"
)

// Service handles blockchain transactions for crypto payouts
type Service struct {
	ethereum *EthereumClient
	bitcoin  *BitcoinClient
	litecoin *LitecoinClient
	log      *logger.Logger
}

// Config holds blockchain service configuration
type Config struct {
	// Ethereum configuration
	EthereumRPCURL     string
	EthereumPrivateKey string
	EthereumChainID    int64

	// Bitcoin configuration
	BitcoinRPCURL      string
	BitcoinRPCUser     string
	BitcoinRPCPassword string

	// Litecoin configuration
	LitecoinRPCURL      string
	LitecoinRPCUser     string
	LitecoinRPCPassword string
}

// Transaction represents a blockchain transaction result
type Transaction struct {
	TxHash          string
	BlockchainType  string
	From            string
	To              string
	Amount          *big.Float
	Fee             *big.Float
	BlockNumber     int64
	Confirmations   int64
	Status          string // pending, confirmed, failed
	ErrorMessage    string
}

// NewService creates a new blockchain service
func NewService(cfg Config, log *logger.Logger) (*Service, error) {
	service := &Service{
		log: log,
	}

	// Initialize Ethereum client
	if cfg.EthereumRPCURL != "" {
		ethClient, err := NewEthereumClient(
			cfg.EthereumRPCURL,
			cfg.EthereumPrivateKey,
			cfg.EthereumChainID,
			log,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize ethereum client: %w", err)
		}
		service.ethereum = ethClient
	}

	// Initialize Bitcoin client
	if cfg.BitcoinRPCURL != "" {
		btcClient, err := NewBitcoinClient(
			cfg.BitcoinRPCURL,
			cfg.BitcoinRPCUser,
			cfg.BitcoinRPCPassword,
			log,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize bitcoin client: %w", err)
		}
		service.bitcoin = btcClient
	}

	// Initialize Litecoin client
	if cfg.LitecoinRPCURL != "" {
		ltcClient, err := NewLitecoinClient(
			cfg.LitecoinRPCURL,
			cfg.LitecoinRPCUser,
			cfg.LitecoinRPCPassword,
			log,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize litecoin client: %w", err)
		}
		service.litecoin = ltcClient
	}

	return service, nil
}

// SendTransaction sends a cryptocurrency transaction
func (s *Service) SendTransaction(ctx context.Context, walletType, toAddress string, amountUSD float64) (*Transaction, error) {
	s.log.Info("initiating blockchain transaction",
		"wallet_type", walletType,
		"to_address", toAddress,
		"amount_usd", amountUSD,
	)

	switch walletType {
	case "ethereum":
		if s.ethereum == nil {
			return nil, fmt.Errorf("ethereum client not configured")
		}
		return s.ethereum.SendTransaction(ctx, toAddress, amountUSD)

	case "bitcoin":
		if s.bitcoin == nil {
			return nil, fmt.Errorf("bitcoin client not configured")
		}
		return s.bitcoin.SendTransaction(ctx, toAddress, amountUSD)

	case "litecoin":
		if s.litecoin == nil {
			return nil, fmt.Errorf("litecoin client not configured")
		}
		return s.litecoin.SendTransaction(ctx, toAddress, amountUSD)

	default:
		return nil, fmt.Errorf("unsupported wallet type: %s", walletType)
	}
}

// GetTransactionStatus checks the status of a transaction
func (s *Service) GetTransactionStatus(ctx context.Context, walletType, txHash string) (*Transaction, error) {
	switch walletType {
	case "ethereum":
		if s.ethereum == nil {
			return nil, fmt.Errorf("ethereum client not configured")
		}
		return s.ethereum.GetTransactionStatus(ctx, txHash)

	case "bitcoin":
		if s.bitcoin == nil {
			return nil, fmt.Errorf("bitcoin client not configured")
		}
		return s.bitcoin.GetTransactionStatus(ctx, txHash)

	case "litecoin":
		if s.litecoin == nil {
			return nil, fmt.Errorf("litecoin client not configured")
		}
		return s.litecoin.GetTransactionStatus(ctx, txHash)

	default:
		return nil, fmt.Errorf("unsupported wallet type: %s", walletType)
	}
}

// ValidateAddress validates a cryptocurrency address
func (s *Service) ValidateAddress(walletType, address string) (bool, error) {
	switch walletType {
	case "ethereum":
		if s.ethereum == nil {
			return false, fmt.Errorf("ethereum client not configured")
		}
		return s.ethereum.ValidateAddress(address)

	case "bitcoin":
		if s.bitcoin == nil {
			return false, fmt.Errorf("bitcoin client not configured")
		}
		return s.bitcoin.ValidateAddress(address)

	case "litecoin":
		if s.litecoin == nil {
			return false, fmt.Errorf("litecoin client not configured")
		}
		return s.litecoin.ValidateAddress(address)

	default:
		return false, fmt.Errorf("unsupported wallet type: %s", walletType)
	}
}

// GetBalance returns the balance of the configured wallet
func (s *Service) GetBalance(ctx context.Context, walletType string) (*big.Float, error) {
	switch walletType {
	case "ethereum":
		if s.ethereum == nil {
			return nil, fmt.Errorf("ethereum client not configured")
		}
		return s.ethereum.GetBalance(ctx)

	case "bitcoin":
		if s.bitcoin == nil {
			return nil, fmt.Errorf("bitcoin client not configured")
		}
		return s.bitcoin.GetBalance(ctx)

	case "litecoin":
		if s.litecoin == nil {
			return nil, fmt.Errorf("litecoin client not configured")
		}
		return s.litecoin.GetBalance(ctx)

	default:
		return nil, fmt.Errorf("unsupported wallet type: %s", walletType)
	}
}

// EstimateFee estimates the transaction fee
func (s *Service) EstimateFee(ctx context.Context, walletType string, amountUSD float64) (*big.Float, error) {
	switch walletType {
	case "ethereum":
		if s.ethereum == nil {
			return nil, fmt.Errorf("ethereum client not configured")
		}
		return s.ethereum.EstimateFee(ctx, amountUSD)

	case "bitcoin":
		if s.bitcoin == nil {
			return nil, fmt.Errorf("bitcoin client not configured")
		}
		return s.bitcoin.EstimateFee(ctx, amountUSD)

	case "litecoin":
		if s.litecoin == nil {
			return nil, fmt.Errorf("litecoin client not configured")
		}
		return s.litecoin.EstimateFee(ctx, amountUSD)

	default:
		return nil, fmt.Errorf("unsupported wallet type: %s", walletType)
	}
}

// Close closes all blockchain clients
func (s *Service) Close() {
	if s.ethereum != nil {
		s.ethereum.Close()
	}
	if s.bitcoin != nil {
		s.bitcoin.Close()
	}
	if s.litecoin != nil {
		s.litecoin.Close()
	}
}
