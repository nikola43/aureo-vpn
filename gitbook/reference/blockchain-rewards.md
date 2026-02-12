# 💰 Blockchain & Rewards

Aureo VPN supports multi-chain cryptocurrency payouts for node operators. The blockchain service handles transaction sending, status checking, address validation, and fee estimation.

Source: `pkg/blockchain/`, `pkg/rewards/`, `pkg/models/node_operator.go`

---

## Multi-Chain Support

| Chain | Library | Transport | Configuration |
|---|---|---|---|
| **Ethereum** | `go-ethereum` | JSON-RPC | `ETHEREUM_RPC_URL`, `ETHEREUM_PRIVATE_KEY`, `ETHEREUM_CHAIN_ID` |
| **Bitcoin** | RPC client | JSON-RPC | `BITCOIN_RPC_URL`, `BITCOIN_RPC_USER`, `BITCOIN_RPC_PASSWORD` |
| **Litecoin** | RPC client | JSON-RPC | `LITECOIN_RPC_URL`, `LITECOIN_RPC_USER`, `LITECOIN_RPC_PASSWORD` |

---

## Blockchain Service API

```go
// Send a cryptocurrency transaction (converts USD to crypto via price oracle)
func (s *Service) SendTransaction(ctx context.Context, walletType, toAddress string, amountUSD float64) (*Transaction, error)

// Check the status of a transaction by hash
func (s *Service) GetTransactionStatus(ctx context.Context, walletType, txHash string) (*Transaction, error)

// Validate a cryptocurrency address format and checksum
func (s *Service) ValidateAddress(walletType, address string) (bool, error)

// Get the wallet balance for the configured payout wallet
func (s *Service) GetBalance(ctx context.Context, walletType string) (*big.Float, error)

// Estimate the network fee for a transaction
func (s *Service) EstimateFee(ctx context.Context, walletType string, amountUSD float64) (*big.Float, error)
```

### Transaction Result

```go
type Transaction struct {
    TxHash         string
    BlockchainType string
    From           string
    To             string
    Amount         *big.Float
    Fee            *big.Float
    BlockNumber    int64
    Confirmations  int64
    Status         string  // "pending", "confirmed", "failed"
    ErrorMessage   string
}
```

---

## Reward Tiers

Node operators earn based on their performance tier. Higher reputation and uptime unlock better rates:

| Tier | Min Reputation | Min Uptime | Base Rate (USD/GB) | Bonus Multiplier | Min Bandwidth | Max Latency |
|---|---|---|---|---|---|---|
| 🥉 **Bronze** | 0 | 0% | $0.01 | 1.0x | 100 Mbps | 100 ms |
| 🥈 **Silver** | 60 | 90% | $0.02 | 1.2x | 200 Mbps | 75 ms |
| 🥇 **Gold** | 75 | 95% | $0.03 | 1.5x | 500 Mbps | 50 ms |
| 💎 **Platinum** | 90 | 99% | $0.05 | 2.0x | 1000 Mbps | 25 ms |

Tier eligibility is checked with:

```go
func (op *NodeOperator) GetEligibleTier(db *gorm.DB) (*NodeReward, error) {
    var tier NodeReward
    err := db.Where("is_active = ? AND min_reputation_score <= ? AND min_uptime_percent <= ?",
        true, op.ReputationScore, op.AverageUptime).
        Order("base_rate_per_gb DESC").
        First(&tier).Error
    // Falls back to Bronze ($0.01/GB) if no tier matches
}
```

---

## Earnings Calculation

```go
func CalculateEarnings(bandwidthGB float64, durationMinutes int, ratePerGB float64, qualityScore float64) float64 {
    baseEarnings := bandwidthGB * ratePerGB

    // Quality multiplier: 0.5x (score=0) to 1.5x (score=100)
    qualityMultiplier := 0.5 + (qualityScore / 100.0)

    // Duration bonus: encourage stable long sessions
    durationBonus := 1.0
    if durationMinutes > 60 {
        durationBonus = 1.1  // +10% for sessions > 1 hour
    }
    if durationMinutes > 180 {
        durationBonus = 1.2  // +20% for sessions > 3 hours
    }

    return baseEarnings * qualityMultiplier * durationBonus
}
```

**Example:** A Gold-tier operator serving 10 GB over 2 hours with quality score 85:

```
baseEarnings     = 10 GB * $0.03/GB           = $0.30
qualityMultiplier = 0.5 + (85/100)             = 1.35
durationBonus    = 1.1 (> 60 min)              = 1.10
total            = $0.30 * 1.35 * 1.10         = $0.4455
```

---

## Payout Pipeline

```
┌─────────┐     ┌───────────────┐     ┌──────────────┐     ┌───────────────┐
│ Session  │────>│trafficMonitor │────>│ flushEarnings│────>│ RecordEarning │
│ (active) │     │ (every 1s)    │     │ (every 10min)│     │               │
└─────────┘     │               │     │              │     │ Creates       │
                │ Accumulates   │     │ Sends        │     │ OperatorEarning│
                │ PendingBW KB  │     │ bandwidthKB  │     │ status=pending│
                └───────────────┘     └──────────────┘     └───────┬───────┘
                                                                    │
                                                                    ▼
                                                           ┌───────────────┐
                                                           │ConfirmEarnings│
                                                           │               │
                                                           │ Validates     │
                                                           │ quality and   │
                                                           │ sets status=  │
                                                           │ confirmed     │
                                                           └───────┬───────┘
                                                                    │
                                                                    ▼
┌────────────┐     ┌──────────────┐     ┌───────────────────────────────────┐
│ blockchain │<────│ProcessPayouts│<────│ Aggregate confirmed earnings     │
│            │     │              │     │ per operator where                │
│ ETH / BTC  │     │ Creates      │     │ pending_payout >= minimum         │
│ / LTC      │     │ OperatorPayout│    │                                   │
│            │     │ with tx hash │     │ Look up exchange rate via         │
│ SendTx()   │     │              │     │ price oracle                      │
└────────────┘     └──────────────┘     └───────────────────────────────────┘
```

### Pipeline Details

1. **Traffic Monitor** (every 1 second) — Reads WireGuard peer stats, accumulates `PendingBandwidthKB` per session
2. **Flush Earnings** (every 10 minutes per session) — Sends accumulated bandwidth to the reward service via `RecordEarning()`
3. **RecordEarning** — Creates an `OperatorEarning` record with status `pending`. Looks up the operator's tier to determine rate
4. **ConfirmEarnings** — Validates quality metrics and changes status to `confirmed`
5. **ProcessPayouts** — Aggregates all `confirmed` earnings per operator. When total meets minimum threshold, creates an `OperatorPayout` record and sends a blockchain transaction
6. **Blockchain Transaction** — Converts USD to crypto via price oracle, sends transaction, records tx hash. Status progresses: `pending` -> `processing` -> `completed` (or `failed`)

### Operator Stats Update

Every heartbeat (30s), the node updates the operator's aggregate stats:

- `total_earned` — Sum of `paid` earnings
- `pending_payout` — Sum of `confirmed` earnings
- `active_nodes_count` — Count of online, active nodes
- `average_uptime` — Average uptime across all operator nodes
- `total_bandwidth_kb` — Sum of bandwidth from all nodes
