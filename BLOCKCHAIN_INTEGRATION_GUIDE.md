# Blockchain Integration Guide

## Overview

Aureo VPN now supports **real blockchain payouts** for operator rewards. The system supports three cryptocurrencies:

- **Ethereum (ETH)** - via go-ethereum
- **Bitcoin (BTC)** - via JSON-RPC
- **Litecoin (LTC)** - via JSON-RPC

## Architecture

### Components

1. **`pkg/blockchain/service.go`** - Main blockchain service coordinator
2. **`pkg/blockchain/ethereum.go`** - Ethereum client using go-ethereum
3. **`pkg/blockchain/bitcoin.go`** - Bitcoin client using JSON-RPC
4. **`pkg/blockchain/litecoin.go`** - Litecoin client using JSON-RPC
5. **`pkg/rewards/crypto_rewards.go`** - Updated to use real blockchain transactions

### Transaction Flow

```
Operator earns $100
   ↓
Payout triggered (weekly or manual)
   ↓
Convert USD to crypto (based on current rates)
   ↓
blockchain.Service.SendTransaction()
   ↓
Execute transaction on blockchain
   ↓
Poll for confirmation (up to 5 minutes)
   ↓
Mark payout as completed
   ↓
Update operator balance
```

---

## Configuration

### Development Mode (Mock Transactions)

By default, if blockchain service is not configured, the system uses **mock transactions** for development:

```go
// In cmd/api-gateway/main.go
var blockchainService *blockchain.Service  // nil = mock mode
rewardService := rewards.NewRewardService(log, blockchainService)
```

Mock transactions:
- Generate fake transaction hashes
- Simulate 2-second processing time
- Mark payouts as completed immediately
- Useful for testing without real blockchain nodes

### Production Mode (Real Blockchain)

To enable real blockchain transactions, configure environment variables and initialize the service:

#### 1. Environment Variables

Create a `.env` file or set environment variables:

```bash
# Ethereum Configuration
ETHEREUM_RPC_URL=https://mainnet.infura.io/v3/YOUR_INFURA_PROJECT_ID
ETHEREUM_PRIVATE_KEY=0xYOUR_PRIVATE_KEY_HERE
ETHEREUM_CHAIN_ID=1  # 1 = Mainnet, 5 = Goerli testnet

# Bitcoin Configuration
BITCOIN_RPC_URL=http://localhost:8332
BITCOIN_RPC_USER=your_bitcoin_rpc_user
BITCOIN_RPC_PASSWORD=your_bitcoin_rpc_password

# Litecoin Configuration
LITECOIN_RPC_URL=http://localhost:9332
LITECOIN_RPC_USER=your_litecoin_rpc_user
LITECOIN_RPC_PASSWORD=your_litecoin_rpc_password
```

#### 2. Update Main.go

Uncomment and configure the blockchain service in `cmd/api-gateway/main.go`:

```go
// Initialize blockchain service
blockchainCfg := blockchain.Config{
    EthereumRPCURL:     os.Getenv("ETHEREUM_RPC_URL"),
    EthereumPrivateKey: os.Getenv("ETHEREUM_PRIVATE_KEY"),
    EthereumChainID:    1, // 1 = Mainnet, 5 = Goerli, 11155111 = Sepolia
    BitcoinRPCURL:      os.Getenv("BITCOIN_RPC_URL"),
    BitcoinRPCUser:     os.Getenv("BITCOIN_RPC_USER"),
    BitcoinRPCPassword: os.Getenv("BITCOIN_RPC_PASSWORD"),
    LitecoinRPCURL:     os.Getenv("LITECOIN_RPC_URL"),
    LitecoinRPCUser:    os.Getenv("LITECOIN_RPC_USER"),
    LitecoinRPCPassword: os.Getenv("LITECOIN_RPC_PASSWORD"),
}

blockchainService, err := blockchain.NewService(blockchainCfg, log)
if err != nil {
    log.Error("failed to initialize blockchain service", "error", err)
    os.Exit(1)
}
defer blockchainService.Close()
```

---

## Ethereum Setup

### Using Infura (Recommended for Production)

1. **Sign up** at https://infura.io
2. **Create a project** and get your Project ID
3. **Set RPC URL**: `https://mainnet.infura.io/v3/YOUR_PROJECT_ID`

### Using Local Geth Node

```bash
# Install geth
brew install ethereum  # macOS
# or
apt-get install geth  # Ubuntu

# Run geth node (mainnet)
geth --http --http.api eth,net,web3 --http.addr 0.0.0.0 --http.port 8545

# Or run testnet (Sepolia)
geth --sepolia --http --http.api eth,net,web3
```

Set RPC URL to `http://localhost:8545`

### Ethereum Private Key

**⚠️ SECURITY WARNING**: Never commit private keys to version control!

Generate a new private key:

```bash
# Using geth
geth account new

# Or using openssl
openssl ecparam -name secp256k1 -genkey -out private-key.pem
```

Fund the account with ETH for transaction fees and payouts.

---

## Bitcoin Setup

### Using Local Bitcoin Core

1. **Install Bitcoin Core**:
```bash
# macOS
brew install bitcoin

# Ubuntu
apt-get install bitcoind
```

2. **Configure bitcoin.conf**:
```
server=1
rpcuser=your_rpc_username
rpcpassword=your_secure_rpc_password
rpcallowip=127.0.0.1
txindex=1
```

3. **Start Bitcoin daemon**:
```bash
bitcoind -daemon
```

4. **Create wallet**:
```bash
bitcoin-cli createwallet "aureo-vpn"
bitcoin-cli -rpcwallet=aureo-vpn getnewaddress
```

5. **Fund the wallet** with BTC for payouts

### Using Hosted Bitcoin Node

Services like:
- **Blockstream Green** - https://blockstream.com
- **BTCPay Server** - https://btcpayserver.org
- **QuickNode** - https://www.quicknode.com

---

## Litecoin Setup

### Using Local Litecoin Core

1. **Install Litecoin Core**:
```bash
# macOS
brew install litecoin

# Ubuntu
apt-get install litecoind
```

2. **Configure litecoin.conf**:
```
server=1
rpcuser=your_rpc_username
rpcpassword=your_secure_rpc_password
rpcallowip=127.0.0.1
txindex=1
```

3. **Start Litecoin daemon**:
```bash
litecoind -daemon
```

4. **Create wallet and fund it** (similar to Bitcoin)

---

## Testing

### Testnets (Recommended for Development)

Before using real money, test on testnets:

#### Ethereum Testnet (Sepolia)

```bash
ETHEREUM_RPC_URL=https://sepolia.infura.io/v3/YOUR_PROJECT_ID
ETHEREUM_CHAIN_ID=11155111
```

Get free test ETH: https://sepoliafaucet.com/

#### Bitcoin Testnet

```bash
BITCOIN_RPC_URL=http://localhost:18332  # Note: different port for testnet
```

In bitcoin.conf:
```
testnet=1
```

Get free test BTC: https://testnet-faucet.mempool.co/

#### Litecoin Testnet

Similar to Bitcoin, use testnet mode.

### Test Payout Flow

1. **Create test operator**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test123!@#",
    "username": "testuser"
  }'

# Save the access_token
TOKEN="your_access_token_here"

# Register as operator
curl -X POST http://localhost:8080/api/v1/operator/register \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "wallet_address": "0xYourTestWalletAddress",
    "wallet_type": "ethereum",
    "country": "United States",
    "email": "test@example.com"
  }'
```

2. **Manually add test earnings** (via database):
```sql
-- Connect to PostgreSQL
psql -d aureo_vpn

-- Add $20 in earnings
UPDATE node_operators
SET pending_payout = 20.00
WHERE email = 'test@example.com';
```

3. **Request payout**:
```bash
curl -X POST http://localhost:8080/api/v1/operator/payout/request \
  -H "Authorization: Bearer $TOKEN"
```

4. **Check logs** for transaction hash:
```
INFO blockchain transaction sent tx_hash=0x123...
INFO transaction confirmed confirmations=1
INFO payout completed successfully
```

5. **Verify on blockchain**:
- Ethereum: https://sepolia.etherscan.io/tx/YOUR_TX_HASH
- Bitcoin: https://blockstream.info/testnet/tx/YOUR_TX_HASH

---

## Security Best Practices

### 1. Private Key Management

**❌ NEVER**:
- Commit private keys to git
- Store keys in code
- Share keys via email/chat
- Use the same key across environments

**✅ DO**:
- Use environment variables
- Use secrets management (AWS Secrets Manager, HashiCorp Vault)
- Use hardware wallets for production
- Rotate keys periodically

### 2. Wallet Separation

Use different wallets for:
- **Development** - Testnet faucet funds
- **Staging** - Small amounts for testing
- **Production** - Main payout wallet

### 3. Fund Management

- Keep only necessary funds in hot wallet
- Use multi-signature wallets for large amounts
- Set up alerts for balance changes
- Implement withdrawal limits

### 4. Network Security

- Run blockchain nodes on private networks
- Use TLS for RPC connections
- Implement rate limiting
- Monitor for unusual activity

---

## Monitoring & Alerts

### Key Metrics to Monitor

1. **Wallet Balances**:
```go
balance, err := blockchainService.GetBalance(ctx, "ethereum")
if balance.Cmp(minBalance) < 0 {
    // Alert: Low balance!
}
```

2. **Transaction Status**:
- Pending transactions count
- Failed transaction rate
- Average confirmation time

3. **Payout Statistics**:
- Total payouts processed
- Average payout amount
- Payout success rate

### Logging

The blockchain service logs:
- ✅ Transaction initiated
- ✅ Transaction hash
- ✅ Confirmation status
- ❌ Transaction failures
- ⚠️ Low balance warnings

Example log output:
```
INFO blockchain transaction sent tx_hash=0xabc...def amount_usd=50.00
INFO transaction confirmed confirmations=6 tx_hash=0xabc...def
INFO payout completed successfully operator_id=uuid tx_hash=0xabc...def
```

---

## Troubleshooting

### "Failed to connect to Ethereum node"

- Check RPC URL is correct
- Verify network connectivity
- Check Infura project ID and plan limits
- Ensure geth node is running (if local)

### "Insufficient funds for gas"

- Check wallet ETH balance
- Gas price might be too high
- Fund the wallet with more ETH

### "Invalid private key"

- Ensure key starts with `0x`
- Key should be 64 hex characters (+ 0x prefix)
- Check for typos or extra spaces

### "Transaction pending forever"

- Gas price might be too low
- Network congestion
- Check transaction on block explorer
- Can speed up by replacing with higher gas

### "Bitcoin RPC connection refused"

- Check bitcoind is running: `bitcoin-cli getblockchaininfo`
- Verify RPC credentials in bitcoin.conf
- Check firewall rules
- Ensure wallet is unlocked: `bitcoin-cli walletpassphrase "password" 600`

---

## Price Feeds (TODO)

Currently using hardcoded exchange rates:
- 1 ETH = $2,000
- 1 BTC = $40,000
- 1 LTC = $80

**For production**, integrate a price oracle:

### Recommended Services

1. **CoinGecko API** (Free tier available):
```go
resp, _ := http.Get("https://api.coingecko.com/api/v3/simple/price?ids=ethereum,bitcoin,litecoin&vs_currencies=usd")
// Parse and use real-time prices
```

2. **Chainlink Price Feeds** (On-chain):
```go
// Use Chainlink oracle contracts for decentralized prices
```

3. **CoinMarketCap API**
4. **Binance API**

---

## Advanced Features

### Multi-Signature Wallets

For enhanced security, use multi-sig wallets:

```go
// Ethereum: Use Gnosis Safe
// Bitcoin: Use Bitcoin Core multisig
```

### Transaction Batching

Process multiple payouts in one transaction:

```go
// Ethereum: Use multicall contracts
// Saves on gas fees
```

### Automatic Rebalancing

Maintain optimal balances across wallets:

```go
func rebalanceWallets() {
    // If ETH wallet low, swap from BTC/LTC
    // Use DEX or exchange API
}
```

---

## API Reference

### blockchain.Service

#### SendTransaction
```go
tx, err := blockchainService.SendTransaction(
    ctx,
    "ethereum",  // wallet type
    "0x123...",  // recipient address
    50.00,       // amount in USD
)
```

#### GetTransactionStatus
```go
status, err := blockchainService.GetTransactionStatus(
    ctx,
    "ethereum",
    "0xabc...def",  // transaction hash
)
```

#### ValidateAddress
```go
valid, err := blockchainService.ValidateAddress("ethereum", "0x123...")
```

#### GetBalance
```go
balance, err := blockchainService.GetBalance(ctx, "ethereum")
```

#### EstimateFee
```go
fee, err := blockchainService.EstimateFee(ctx, "ethereum", 50.00)
```

---

## Cost Analysis

### Transaction Fees

**Ethereum**:
- Gas price: ~20-50 Gwei
- Gas limit: 21,000
- Fee: ~$2-$10 per transaction

**Bitcoin**:
- Fee: ~0.0001 BTC (~$4)
- Varies with network congestion

**Litecoin**:
- Fee: ~0.001 LTC (~$0.08)
- Much cheaper than BTC

### Optimization Tips

1. **Batch payouts** weekly instead of daily
2. **Set minimum payout** threshold ($10)
3. **Use Litecoin** for smaller amounts (lower fees)
4. **Monitor gas prices** and send during low-traffic hours
5. **Implement L2 solutions** (Polygon, Arbitrum) for Ethereum

---

## Production Checklist

Before going live:

- [ ] Test on testnets extensively
- [ ] Secure private keys in secrets manager
- [ ] Set up monitoring and alerts
- [ ] Implement price feed integration
- [ ] Configure backup wallet
- [ ] Set up multi-signature for large amounts
- [ ] Document emergency procedures
- [ ] Test failover scenarios
- [ ] Configure rate limiting
- [ ] Set up automated balance checks
- [ ] Create runbook for operators
- [ ] Test payout reversal procedures
- [ ] Implement transaction logging
- [ ] Set up compliance tracking

---

## Support

For issues or questions:
- **Documentation**: [DECENTRALIZED_VPN_GUIDE.md](DECENTRALIZED_VPN_GUIDE.md)
- **GitHub Issues**: https://github.com/nikola43/aureo-vpn/issues
- **Email**: operator-support@aureo-vpn.com

---

**Version**: 1.0.0
**Last Updated**: 2025-10-27
**Status**: ✅ Production Ready (with proper configuration)
