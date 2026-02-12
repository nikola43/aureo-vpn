# ⛓️ Blockchain & Payments

Multi-chain cryptocurrency payments for node operator rewards.

---

## Payment Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                      PAYMENT FLOW                                        │
└─────────────────────────────────────────────────────────────────────────┘

Earning Calculation:
  ┌──────────────────────────────────────────────────────────────────────┐
  │                                                                      │
  │  Session Ends → Calculate Bandwidth → Apply Tier Rate → Store       │
  │                                                                      │
  │  Example:                                                            │
  │    Bandwidth: 50 GB                                                  │
  │    Tier: Gold ($0.02/GB)                                            │
  │    Quality Score: 85%                                               │
  │    Duration: 4 hours (1.2x bonus)                                   │
  │                                                                      │
  │    Earnings = 50 × $0.02 × (0.5 + 0.85) × 1.2 = $1.62               │
  │                                                                      │
  └──────────────────────────────────────────────────────────────────────┘

Payout Process:
  ┌─────────────┐                          ┌─────────────┐
  │  Operator   │ ──── Request Payout ───▶ │    API      │
  │             │                          │             │
  │  Wallet:    │                          │  • Verify   │
  │  0x1234...  │                          │    balance  │
  └─────────────┘                          │  • Check    │
                                           │    minimum  │
                                           └──────┬──────┘
                                                  │
                                                  ▼
                                           ┌─────────────┐
                                           │ Blockchain  │
                                           │   Service   │
                                           │             │
                                           │ • Create TX │
                                           │ • Sign      │
                                           │ • Broadcast │
                                           └──────┬──────┘
                                                  │
                     ┌────────────────────────────┼────────────────────────────┐
                     ▼                            ▼                            ▼
              ┌─────────────┐              ┌─────────────┐              ┌─────────────┐
              │  Ethereum   │              │   Bitcoin   │              │  Litecoin   │
              │   Network   │              │   Network   │              │   Network   │
              └─────────────┘              └─────────────┘              └─────────────┘
```

## Supported Chains

| Chain | Library | Use Case |
|-------|---------|----------|
| Ethereum | go-ethereum | Primary payments, smart contracts |
| Bitcoin | RPC client | BTC payments |
| Litecoin | RPC client | Fast, low-fee alternative |

## Earning Formula

```
earnings = bandwidth_gb × rate_per_gb × quality_multiplier × duration_bonus

quality_multiplier = 0.5 + (quality_score / 100)
duration_bonus = 1.1 (>60min) or 1.2 (>180min)
```
