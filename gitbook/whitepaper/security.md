# 🛡️ Security & Privacy

Multiple layers of security protect every aspect of the Aureo VPN platform.

---

## 🏰 Defense-in-Depth Architecture

Aureo employs multiple layers of security:

1. 🔒 **Transport Security** — WireGuard with ChaCha20-Poly1305 AEAD encryption
2. 🔑 **Authentication** — JWT tokens with HMAC-SHA256, short-lived access tokens (15 min)
3. 🔐 **Password Security** — bcrypt hashing with common password blocklist
4. ✅ **Input Validation** — Comprehensive validation with SQL injection and XSS detection
5. 🚦 **Rate Limiting** — Per-IP request throttling to prevent abuse
6. 🛡️ **Privacy Filter** — Automatic redaction of PII from all system logs

---

## 🔍 Zero-Knowledge Architecture

Aureo is designed so that no single component has complete information:

- The **API Gateway** knows user accounts but not traffic content
- **VPN Nodes** forward encrypted traffic but do not have user account details
- **Client IPs** are hashed before storage — the real IP is never persisted
- **Payment processing** is on-chain — no centralized payment data

---

## 📜 Compliance

- **GDPR Ready** — Minimal data collection, defined retention periods, right to deletion
- **Privacy by Design** — Privacy protections are architectural, not just policy-based
- **Open Infrastructure** — Community-operated nodes can be independently audited
- **Warrant Canary** — Published and regularly updated statement confirming no government data requests
