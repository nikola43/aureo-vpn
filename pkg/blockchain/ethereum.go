package blockchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/nikola43/aureo-vpn/pkg/logger"
)

// EthereumClient handles Ethereum blockchain interactions
type EthereumClient struct {
	client     *ethclient.Client
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	address    common.Address
	chainID    *big.Int
	log        *logger.Logger
}

// NewEthereumClient creates a new Ethereum client
func NewEthereumClient(rpcURL, privateKeyHex string, chainID int64, log *logger.Logger) (*EthereumClient, error) {
	// Connect to Ethereum node
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ethereum node: %w", err)
	}

	// Parse private key
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	// Derive public key and address
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	log.Info("ethereum client initialized",
		"address", address.Hex(),
		"chain_id", chainID,
	)

	return &EthereumClient{
		client:     client,
		privateKey: privateKey,
		publicKey:  publicKeyECDSA,
		address:    address,
		chainID:    big.NewInt(chainID),
		log:        log,
	}, nil
}

// SendTransaction sends ETH to an address
func (ec *EthereumClient) SendTransaction(ctx context.Context, toAddress string, amountUSD float64) (*Transaction, error) {
	// Validate address
	if !common.IsHexAddress(toAddress) {
		return nil, fmt.Errorf("invalid ethereum address: %s", toAddress)
	}
	to := common.HexToAddress(toAddress)

	// Get current ETH price in USD (simplified - in production, use a price oracle)
	// For now, assume 1 ETH = $2000 USD
	ethPriceUSD := 2000.0
	ethAmount := amountUSD / ethPriceUSD

	// Convert to Wei (1 ETH = 10^18 Wei)
	amountWei := new(big.Float).Mul(big.NewFloat(ethAmount), big.NewFloat(1e18))
	amount := new(big.Int)
	amountWei.Int(amount)

	// Get nonce
	nonce, err := ec.client.PendingNonceAt(ctx, ec.address)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get gas price
	gasPrice, err := ec.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Estimate gas limit
	gasLimit := uint64(21000) // Standard ETH transfer

	// Create transaction
	tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, nil)

	// Sign transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(ec.chainID), ec.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	err = ec.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	// Calculate fee
	fee := new(big.Float).Mul(
		new(big.Float).SetInt(gasPrice),
		new(big.Float).SetInt(big.NewInt(int64(gasLimit))),
	)
	feeETH := new(big.Float).Quo(fee, big.NewFloat(1e18))

	ec.log.Info("ethereum transaction sent",
		"tx_hash", signedTx.Hash().Hex(),
		"to", toAddress,
		"amount_eth", ethAmount,
		"amount_usd", amountUSD,
		"nonce", nonce,
	)

	return &Transaction{
		TxHash:         signedTx.Hash().Hex(),
		BlockchainType: "ethereum",
		From:           ec.address.Hex(),
		To:             toAddress,
		Amount:         big.NewFloat(ethAmount),
		Fee:            feeETH,
		Status:         "pending",
	}, nil
}

// GetTransactionStatus gets the status of a transaction
func (ec *EthereumClient) GetTransactionStatus(ctx context.Context, txHash string) (*Transaction, error) {
	hash := common.HexToHash(txHash)

	// Get transaction receipt
	receipt, err := ec.client.TransactionReceipt(ctx, hash)
	if err != nil {
		// Transaction might be pending
		return &Transaction{
			TxHash:         txHash,
			BlockchainType: "ethereum",
			Status:         "pending",
		}, nil
	}

	// Get transaction details
	tx, isPending, err := ec.client.TransactionByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	status := "confirmed"
	if isPending {
		status = "pending"
	} else if receipt.Status == 0 {
		status = "failed"
	}

	// Get current block number for confirmations
	currentBlock, err := ec.client.BlockNumber(ctx)
	if err != nil {
		currentBlock = 0
	}
	confirmations := int64(0)
	if receipt.BlockNumber != nil && currentBlock > receipt.BlockNumber.Uint64() {
		confirmations = int64(currentBlock - receipt.BlockNumber.Uint64())
	}

	// Convert amount from Wei to ETH
	amount := new(big.Float).Quo(
		new(big.Float).SetInt(tx.Value()),
		big.NewFloat(1e18),
	)

	// Calculate fee
	gasUsed := new(big.Int).SetUint64(receipt.GasUsed)
	gasPrice := tx.GasPrice()
	fee := new(big.Float).Mul(
		new(big.Float).SetInt(gasPrice),
		new(big.Float).SetInt(gasUsed),
	)
	feeETH := new(big.Float).Quo(fee, big.NewFloat(1e18))

	var toAddress string
	if tx.To() != nil {
		toAddress = tx.To().Hex()
	}

	return &Transaction{
		TxHash:         txHash,
		BlockchainType: "ethereum",
		From:           ec.address.Hex(),
		To:             toAddress,
		Amount:         amount,
		Fee:            feeETH,
		BlockNumber:    receipt.BlockNumber.Int64(),
		Confirmations:  confirmations,
		Status:         status,
	}, nil
}

// ValidateAddress validates an Ethereum address
func (ec *EthereumClient) ValidateAddress(address string) (bool, error) {
	// Check if address matches Ethereum address format (0x followed by 40 hex characters)
	matched, err := regexp.MatchString("^0x[0-9a-fA-F]{40}$", address)
	if err != nil {
		return false, err
	}
	return matched && common.IsHexAddress(address), nil
}

// GetBalance returns the ETH balance of the configured wallet
func (ec *EthereumClient) GetBalance(ctx context.Context) (*big.Float, error) {
	balance, err := ec.client.BalanceAt(ctx, ec.address, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	// Convert Wei to ETH
	ethBalance := new(big.Float).Quo(
		new(big.Float).SetInt(balance),
		big.NewFloat(1e18),
	)

	return ethBalance, nil
}

// EstimateFee estimates the transaction fee for sending a transaction
func (ec *EthereumClient) EstimateFee(ctx context.Context, amountUSD float64) (*big.Float, error) {
	// Get current gas price
	gasPrice, err := ec.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Standard gas limit for ETH transfer
	gasLimit := uint64(21000)

	// Calculate fee in Wei
	fee := new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit)))

	// Convert to ETH
	feeETH := new(big.Float).Quo(
		new(big.Float).SetInt(fee),
		big.NewFloat(1e18),
	)

	return feeETH, nil
}

// Close closes the Ethereum client connection
func (ec *EthereumClient) Close() {
	if ec.client != nil {
		ec.client.Close()
	}
}
