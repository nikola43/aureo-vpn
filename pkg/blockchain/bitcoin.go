package blockchain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"regexp"

	"github.com/nikola43/aureo-vpn/pkg/logger"
)

// BitcoinClient handles Bitcoin blockchain interactions via JSON-RPC
type BitcoinClient struct {
	rpcURL   string
	rpcUser  string
	rpcPass  string
	client   *http.Client
	log      *logger.Logger
}

// bitcoinRPCRequest represents a JSON-RPC request
type bitcoinRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

// bitcoinRPCResponse represents a JSON-RPC response
type bitcoinRPCResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	ID string `json:"id"`
}

// NewBitcoinClient creates a new Bitcoin client
func NewBitcoinClient(rpcURL, rpcUser, rpcPass string, log *logger.Logger) (*BitcoinClient, error) {
	client := &BitcoinClient{
		rpcURL:  rpcURL,
		rpcUser: rpcUser,
		rpcPass: rpcPass,
		client:  &http.Client{},
		log:     log,
	}

	// Test connection
	_, err := client.call("getblockchaininfo", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to bitcoin node: %w", err)
	}

	log.Info("bitcoin client initialized", "rpc_url", rpcURL)
	return client, nil
}

// call makes a JSON-RPC call to the Bitcoin node
func (bc *BitcoinClient) call(method string, params []interface{}) (json.RawMessage, error) {
	// Build request
	reqBody := bitcoinRPCRequest{
		JSONRPC: "1.0",
		ID:      "aureo-vpn",
		Method:  method,
		Params:  params,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", bc.rpcURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(bc.rpcUser, bc.rpcPass)

	// Send request
	resp, err := bc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var rpcResp bitcoinRPCResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for RPC error
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}

// SendTransaction sends BTC to an address
func (bc *BitcoinClient) SendTransaction(ctx context.Context, toAddress string, amountUSD float64) (*Transaction, error) {
	// Validate address
	valid, err := bc.ValidateAddress(toAddress)
	if err != nil || !valid {
		return nil, fmt.Errorf("invalid bitcoin address: %s", toAddress)
	}

	// Get current BTC price in USD (simplified - in production, use a price oracle)
	// For now, assume 1 BTC = $40,000 USD
	btcPriceUSD := 40000.0
	btcAmount := amountUSD / btcPriceUSD

	// Send transaction using sendtoaddress RPC method
	// This requires the wallet to be unlocked
	params := []interface{}{toAddress, btcAmount}
	result, err := bc.call("sendtoaddress", params)
	if err != nil {
		return nil, fmt.Errorf("failed to send bitcoin transaction: %w", err)
	}

	// Parse transaction hash
	var txHash string
	if err := json.Unmarshal(result, &txHash); err != nil {
		return nil, fmt.Errorf("failed to parse transaction hash: %w", err)
	}

	bc.log.Info("bitcoin transaction sent",
		"tx_hash", txHash,
		"to", toAddress,
		"amount_btc", btcAmount,
		"amount_usd", amountUSD,
	)

	// Estimate fee (approximate)
	fee := big.NewFloat(0.0001) // Typical Bitcoin transaction fee

	return &Transaction{
		TxHash:         txHash,
		BlockchainType: "bitcoin",
		To:             toAddress,
		Amount:         big.NewFloat(btcAmount),
		Fee:            fee,
		Status:         "pending",
	}, nil
}

// GetTransactionStatus gets the status of a Bitcoin transaction
func (bc *BitcoinClient) GetTransactionStatus(ctx context.Context, txHash string) (*Transaction, error) {
	// Get transaction details
	params := []interface{}{txHash, true}
	result, err := bc.call("gettransaction", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	// Parse transaction
	var txInfo struct {
		Amount        float64 `json:"amount"`
		Fee           float64 `json:"fee"`
		Confirmations int64   `json:"confirmations"`
		BlockIndex    int64   `json:"blockindex"`
		BlockHash     string  `json:"blockhash"`
		TxID          string  `json:"txid"`
		Details       []struct {
			Address string  `json:"address"`
			Amount  float64 `json:"amount"`
		} `json:"details"`
	}

	if err := json.Unmarshal(result, &txInfo); err != nil {
		return nil, fmt.Errorf("failed to parse transaction: %w", err)
	}

	// Determine status
	status := "pending"
	if txInfo.Confirmations >= 6 {
		status = "confirmed"
	} else if txInfo.Confirmations > 0 {
		status = "confirming"
	}

	// Get recipient address
	toAddress := ""
	if len(txInfo.Details) > 0 {
		toAddress = txInfo.Details[0].Address
	}

	return &Transaction{
		TxHash:         txHash,
		BlockchainType: "bitcoin",
		To:             toAddress,
		Amount:         big.NewFloat(txInfo.Amount),
		Fee:            big.NewFloat(txInfo.Fee * -1), // Fee is negative in response
		Confirmations:  txInfo.Confirmations,
		Status:         status,
	}, nil
}

// ValidateAddress validates a Bitcoin address
func (bc *BitcoinClient) ValidateAddress(address string) (bool, error) {
	// Bitcoin addresses start with 1, 3, or bc1 and are 26-35 characters
	matched, err := regexp.MatchString("^(1|3|bc1)[a-zA-HJ-NP-Z0-9]{25,62}$", address)
	if err != nil {
		return false, err
	}

	if !matched {
		return false, nil
	}

	// Verify with Bitcoin node
	params := []interface{}{address}
	result, err := bc.call("validateaddress", params)
	if err != nil {
		return false, err
	}

	var validation struct {
		IsValid bool `json:"isvalid"`
	}
	if err := json.Unmarshal(result, &validation); err != nil {
		return false, err
	}

	return validation.IsValid, nil
}

// GetBalance returns the BTC balance of the wallet
func (bc *BitcoinClient) GetBalance(ctx context.Context) (*big.Float, error) {
	result, err := bc.call("getbalance", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	var balance float64
	if err := json.Unmarshal(result, &balance); err != nil {
		return nil, fmt.Errorf("failed to parse balance: %w", err)
	}

	return big.NewFloat(balance), nil
}

// EstimateFee estimates the transaction fee
func (bc *BitcoinClient) EstimateFee(ctx context.Context, amountUSD float64) (*big.Float, error) {
	// Estimate smart fee for 6 block confirmation
	params := []interface{}{6}
	result, err := bc.call("estimatesmartfee", params)
	if err != nil {
		// Fallback to default fee
		return big.NewFloat(0.0001), nil
	}

	var feeEstimate struct {
		FeeRate float64 `json:"feerate"`
	}
	if err := json.Unmarshal(result, &feeEstimate); err != nil {
		return big.NewFloat(0.0001), nil
	}

	// FeeRate is in BTC/kB, typical transaction is ~250 bytes
	fee := feeEstimate.FeeRate * 0.25

	return big.NewFloat(fee), nil
}

// Close closes the Bitcoin client connection
func (bc *BitcoinClient) Close() {
	// HTTP client doesn't need explicit closing
}
