# 💰 Node Operator Economy

The economic model that powers the decentralized network.

---

## 💸 How Operators Earn

Node operators are compensated based on actual bandwidth served, adjusted for quality:

```
Earnings = BandwidthGB x RatePerGB x QualityMultiplier x DurationBonus
```

**Quality Multiplier** (0.5x - 1.5x): Based on connection quality score (latency, stability, uptime).

**Duration Bonus**: Longer stable connections earn more:
- Standard: 1.0x
- 1+ hour sessions: 1.1x
- 3+ hour sessions: 1.2x

---

## ⭐ Reputation System

Every operator starts with a base reputation score of 50 (out of 100). The score is composed of:

| Component | Max Points | Criteria |
|-----------|-----------|----------|
| Base | 50 | Starting score |
| Uptime | 30 | (Average Uptime % / 100) x 30 |
| User Ratings | 20 | (Average Rating / 5) x 20 |
| Bandwidth Served | 10 | 100GB+ = 5pts, 1TB+ = 10pts |
| Stake Amount | 10 | $100+ = 5pts, $1000+ = 10pts |

Higher reputation unlocks better reward tiers, creating a positive feedback loop that incentivizes quality infrastructure.

---

## 🔐 Staking

Operators can optionally stake cryptocurrency as a security deposit, demonstrating commitment to the network:

- Staked funds boost reputation score
- In cases of proven malicious behavior, stakes can be slashed
- Staking creates economic alignment between operators and users

---

## 💳 Payout Pipeline

Earnings flow through a transparent pipeline:

```
Traffic Served → Earnings Recorded (pending)
    → Quality Verified (confirmed)
    → Payout Threshold Reached
    → Crypto Conversion (real-time exchange rate)
    → Blockchain Transaction
    → Operator Wallet
```

Processing time: 24-48 hours from request to wallet receipt.
