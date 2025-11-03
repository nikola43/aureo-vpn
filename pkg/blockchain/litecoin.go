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

// LitecoinClient handles Litecoin blockchain interactions via JSON-RPC
type LitecoinClient struct {
	rpcURL   string
	rpcUser  string
	rpcPass  string
	client   *http.Client
	log      *logger.Logger
}

// litecoinRPCRequest represents a JSON-RPC request
type litecoinRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

// litecoinRPCResponse represents a JSON-RPC response
type litecoinRPCResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	ID string `json:"id"`
}

// NewLitecoinClient creates a new Litecoin client
func NewLitecoinClient(rpcURL, rpcUser, rpcPass string, log *logger.Logger) (*LitecoinClient, error) {
	client := &LitecoinClient{
		rpcURL:  rpcURL,
		rpcUser: rpcUser,
		rpcPass: rpcPass,
		client:  &http.Client{},
		log:     log,
	}

	// Test connection
	_, err := client.call("getblockchaininfo", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to litecoin node: %w", err)
	}

	log.Info("litecoin client initialized", "rpc_url", rpcURL)
	return client, nil
}

// call makes a JSON-RPC call to the Litecoin node
func (lc *LitecoinClient) call(method string, params []interface{}) (json.RawMessage, error) {
	// Build request
	reqBody := litecoinRPCRequest{
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
	req, err := http.NewRequest("POST", lc.rpcURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(lc.rpcUser, lc.rpcPass)

	// Send request
	resp, err := lc.client.Do(req)
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
	var rpcResp litecoinRPCResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for RPC error
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}

// SendTransaction sends LTC to an address
func (lc *LitecoinClient) SendTransaction(ctx context.Context, toAddress string, amountUSD float64) (*Transaction, error) {
	// Validate address
	valid, err := lc.ValidateAddress(toAddress)
	if err != nil || !valid {
		return nil, fmt.Errorf("invalid litecoin address: %s", toAddress)
	}

	// Get current LTC price in USD (simplified - in production, use a price oracle)
	// For now, assume 1 LTC = $80 USD
	ltcPriceUSD := 80.0
	ltcAmount := amountUSD / ltcPriceUSD

	// Send transaction using sendtoaddress RPC method
	params := []interface{}{toAddress, ltcAmount}
	result, err := lc.call("sendtoaddress", params)
	if err != nil {
		return nil, fmt.Errorf("failed to send litecoin transaction: %w", err)
	}

	// Parse transaction hash
	var txHash string
	if err := json.Unmarshal(result, &txHash); err != nil {
		return nil, fmt.Errorf("failed to parse transaction hash: %w", err)
	}

	lc.log.Info("litecoin transaction sent",
		"tx_hash", txHash,
		"to", toAddress,
		"amount_ltc", ltcAmount,
		"amount_usd", amountUSD,
	)

	// Estimate fee (approximate)
	fee := big.NewFloat(0.001) // Typical Litecoin transaction fee

	return &Transaction{
		TxHash:         txHash,
		BlockchainType: "litecoin",
		To:             toAddress,
		Amount:         big.NewFloat(ltcAmount),
		Fee:            fee,
		Status:         "pending",
	}, nil
}

// GetTransactionStatus gets the status of a Litecoin transaction
func (lc *LitecoinClient) GetTransactionStatus(ctx context.Context, txHash string) (*Transaction, error) {
	// Get transaction details
	params := []interface{}{txHash, true}
	result, err := lc.call("gettransaction", params)
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
		BlockchainType: "litecoin",
		To:             toAddress,
		Amount:         big.NewFloat(txInfo.Amount),
		Fee:            big.NewFloat(txInfo.Fee * -1), // Fee is negative in response
		Confirmations:  txInfo.Confirmations,
		Status:         status,
	}, nil
}

// ValidateAddress validates a Litecoin address
func (lc *LitecoinClient) ValidateAddress(address string) (bool, error) {
	// Litecoin addresses start with L, M, or ltc1 and are 26-35 characters
	matched, err := regexp.MatchString("^(L|M|ltc1)[a-zA-HJ-NP-Z0-9]{25,62}$", address)
	if err != nil {
		return false, err
	}

	if !matched {
		return false, nil
	}

	// Verify with Litecoin node
	params := []interface{}{address}
	result, err := lc.call("validateaddress", params)
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

// GetBalance returns the LTC balance of the wallet
func (lc *LitecoinClient) GetBalance(ctx context.Context) (*big.Float, error) {
	result, err := lc.call("getbalance", nil)
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
func (lc *LitecoinClient) EstimateFee(ctx context.Context, amountUSD float64) (*big.Float, error) {
	// Estimate smart fee for 6 block confirmation
	params := []interface{}{6}
	result, err := lc.call("estimatesmartfee", params)
	if err != nil {
		// Fallback to default fee
		return big.NewFloat(0.001), nil
	}

	var feeEstimate struct {
		FeeRate float64 `json:"feerate"`
	}
	if err := json.Unmarshal(result, &feeEstimate); err != nil {
		return big.NewFloat(0.001), nil
	}

	// FeeRate is in LTC/kB, typical transaction is ~250 bytes
	fee := feeEstimate.FeeRate * 0.25

	return big.NewFloat(fee), nil
}

// Close closes the Litecoin client connection
func (lc *LitecoinClient) Close() {
	// HTTP client doesn't need explicit closing
}
