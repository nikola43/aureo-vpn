# ⚙️ How Aureo Works

A simple experience for users, a rewarding opportunity for operators.

---

## 👤 For Users

Connecting to Aureo is as simple as pressing a button:

```
1. Open the Aureo app
   └── Available on iOS, Android, macOS, Windows, Linux

2. Tap "Quick Connect"
   └── Automatically selects the fastest, lowest-load node

3. You're protected
   └── All traffic encrypted with WireGuard (ChaCha20-Poly1305)
   └── Your real IP is hidden from every website and service
   └── Kill switch prevents any data leaks if connection drops
```

### Behind the Scenes

```
┌────────────┐     1. Generate      ┌──────────────────────┐
│            │     WireGuard keys   │                      │
│   Aureo    │ ──────────────────►  │   Aureo API Gateway  │
│   App      │     2. Register      │                      │
│            │     public key       │  - Select best node  │
│            │ ◄──────────────────  │  - Allocate tunnel IP│
│            │     3. Receive       │  - Return server key │
│            │     server config    │                      │
└─────┬──────┘                      └──────────────────────┘
      │
      │  4. Establish encrypted
      │     WireGuard tunnel
      ▼
┌──────────────┐
│   VPN Node   │  5. All traffic routed
│  (Community  │     through encrypted tunnel
│   Operated)  │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│   Internet   │  Your real IP is hidden
└──────────────┘
```

---

## 📡 For Node Operators

Running an Aureo node turns your server into a revenue-generating privacy infrastructure:

```
1. Register as an operator
   └── Provide your crypto wallet address (ETH, BTC, or LTC)

2. Deploy a node
   └── Run the Aureo node software on your server
   └── Automatic P2P network integration

3. Earn rewards
   └── Get paid per GB of bandwidth served
   └── Higher earnings for better performance
   └── Automated payouts to your crypto wallet
```
