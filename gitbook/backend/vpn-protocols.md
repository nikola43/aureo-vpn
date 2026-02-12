# 🔒 VPN Protocols

Aureo VPN supports multiple protocols for secure tunneling. Each protocol has different strengths and is suited for different use cases.

---

## Protocol Comparison

| Feature | WireGuard | OpenVPN |
|---------|-----------|---------|
| Encryption | ChaCha20-Poly1305 | AES-256-GCM |
| Key Exchange | Curve25519 | RSA 2048-bit + TLS |
| Speed | 8-10 Gbps | 1-2 Gbps |
| Latency Overhead | +5-15ms | +10-30ms |
| Codebase Size | ~4,000 lines | ~100,000 lines |
| Default Port | 51820/UDP | 1194/UDP |
| Kernel Integration | Yes (Linux 5.6+) | No (userspace) |
| Best For | Speed, mobile, modern | Compatibility, firewalls |

---

## 🔐 WireGuard Implementation

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        WIREGUARD TUNNEL                                  │
└─────────────────────────────────────────────────────────────────────────┘

Key Generation (Curve25519):
  Private Key: 32 random bytes with bit clamping
  Public Key:  X25519(private_key, basepoint)
  Pre-shared:  32 random bytes (optional extra security)

Connection Flow:
  1. Client generates keypair
  2. Server generates keypair
  3. Exchange public keys via API
  4. Server adds client as peer with allocated IP
  5. Client configures local interface
  6. Noise Protocol handshake (IKpsk2)
  7. ChaCha20-Poly1305 encrypted tunnel

Config Example (Client):
┌────────────────────────────────────────┐
│ [Interface]                            │
│ PrivateKey = <client_private_key>      │
│ Address = 10.8.0.2/32                  │
│ DNS = 1.1.1.1, 8.8.8.8                │
│                                        │
│ [Peer]                                 │
│ PublicKey = <server_public_key>        │
│ Endpoint = vpn.example.com:51820       │
│ AllowedIPs = 0.0.0.0/0                │
│ PersistentKeepalive = 25               │
└────────────────────────────────────────┘
```

### WireGuard Handshake Flow

```
Client                                            Server
  │                                                  │
  │  1. Initiation (Noise IKpsk2)                    │
  │  ┌──────────────────────────────────────────┐    │
  │  │ sender_index                             │    │
  │  │ unencrypted_ephemeral (Curve25519)        │    │
  │  │ encrypted_static                          │    │
  │  │ encrypted_timestamp                       │    │
  │  │ MAC1, MAC2                                │    │
  │  └──────────────────────────────────────────┘    │
  │──────────────────────────────────────────────────▶│
  │                                                  │
  │  2. Response                                     │
  │  ┌──────────────────────────────────────────┐    │
  │  │ sender_index, receiver_index              │    │
  │  │ unencrypted_ephemeral                     │    │
  │  │ encrypted_nothing                         │    │
  │  │ MAC1, MAC2                                │    │
  │  └──────────────────────────────────────────┘    │
  │◀──────────────────────────────────────────────────│
  │                                                  │
  │  3. Transport Data (ChaCha20-Poly1305)           │
  │══════════════════════════════════════════════════▶│
  │◀══════════════════════════════════════════════════│
  │         Encrypted bidirectional tunnel            │
```

### WireGuard Key Management

```go
// Key generation in Go (pkg/protocols/wireguard/)
func GenerateKeyPair() (privateKey, publicKey string, err error) {
    // Generate 32 random bytes for private key
    var key [32]byte
    if _, err := rand.Read(key[:]); err != nil {
        return "", "", err
    }

    // Clamp private key (Curve25519 requirement)
    key[0] &= 248
    key[31] &= 127
    key[31] |= 64

    // Derive public key via X25519
    var pub [32]byte
    curve25519.ScalarBaseMult(&pub, &key)

    return base64.StdEncoding.EncodeToString(key[:]),
           base64.StdEncoding.EncodeToString(pub[:]), nil
}
```

### WireGuard Server Configuration

```ini
[Interface]
PrivateKey = <server_private_key>
Address = 10.8.0.1/24
ListenPort = 51820
PostUp = iptables -A FORWARD -i wg0 -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
PostDown = iptables -D FORWARD -i wg0 -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE

# Peer added dynamically via API when session is created
[Peer]
PublicKey = <client_public_key>
AllowedIPs = 10.8.0.2/32
```

### WireGuard Features

- **Noise Protocol Framework**: IKpsk2 handshake pattern
- **Perfect Forward Secrecy**: New keys for every handshake
- **Roaming**: Seamless reconnection on IP changes
- **PersistentKeepalive**: NAT traversal (25-second interval)
- **Pre-shared Keys**: Optional quantum-resistant layer
- **Cryptokey Routing**: Built-in IP-to-key mapping

---

## 🔒 OpenVPN Implementation

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        OPENVPN TUNNEL                                    │
└─────────────────────────────────────────────────────────────────────────┘

Certificate Generation:
  • RSA 2048-bit keys
  • Self-signed CA certificate
  • Server/Client certificates signed by CA
  • 1-year validity

Connection Flow:
  1. TLS handshake with certificate validation
  2. Key exchange and session key generation
  3. UDP tunnel with AES-256-GCM encryption
  4. HMAC-SHA256 authentication

Config Example (Client):
┌────────────────────────────────────────┐
│ client                                 │
│ dev tun                                │
│ proto udp                              │
│ remote vpn.example.com 1194            │
│ resolv-retry infinite                  │
│ nobind                                 │
│ persist-key                            │
│ persist-tun                            │
│ cipher AES-256-GCM                     │
│ auth SHA256                            │
│ <ca>                                   │
│ -----BEGIN CERTIFICATE-----            │
│ ...                                    │
│ </ca>                                  │
└────────────────────────────────────────┘
```

### OpenVPN Connection Flow

```
Client                                            Server
  │                                                  │
  │  1. TCP/UDP connection                           │
  │──────────────────────────────────────────────────▶│
  │                                                  │
  │  2. TLS Handshake                                │
  │  ┌──────────────────────────────────────────┐    │
  │  │ ClientHello                               │    │
  │  │ ServerHello + Certificate                 │    │
  │  │ Client Certificate (if required)          │    │
  │  │ Key Exchange (RSA/ECDHE)                  │    │
  │  │ Finished                                  │    │
  │  └──────────────────────────────────────────┘    │
  │◀─────────────────────────────────────────────────▶│
  │                                                  │
  │  3. Control Channel (TLS-encrypted)              │
  │  ┌──────────────────────────────────────────┐    │
  │  │ Push configuration                        │    │
  │  │ IP assignment                             │    │
  │  │ Route setup                               │    │
  │  └──────────────────────────────────────────┘    │
  │◀──────────────────────────────────────────────────│
  │                                                  │
  │  4. Data Channel (AES-256-GCM encrypted)         │
  │══════════════════════════════════════════════════▶│
  │◀══════════════════════════════════════════════════│
  │         Encrypted bidirectional tunnel            │
```

### OpenVPN Certificate Generation

```go
// Certificate generation in Go (pkg/protocols/openvpn/)
func GenerateCertificates() (*CertBundle, error) {
    // 1. Generate CA key pair (RSA 2048)
    caKey, _ := rsa.GenerateKey(rand.Reader, 2048)

    // 2. Create self-signed CA certificate
    caTemplate := &x509.Certificate{
        SerialNumber: big.NewInt(1),
        Subject: pkix.Name{
            Organization: []string{"Aureo VPN CA"},
        },
        NotBefore:             time.Now(),
        NotAfter:              time.Now().Add(365 * 24 * time.Hour),
        IsCA:                  true,
        KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
        BasicConstraintsValid: true,
    }

    // 3. Generate server certificate signed by CA
    serverKey, _ := rsa.GenerateKey(rand.Reader, 2048)
    serverCert := signWithCA(caKey, caTemplate, serverKey)

    // 4. Generate client certificate signed by CA
    clientKey, _ := rsa.GenerateKey(rand.Reader, 2048)
    clientCert := signWithCA(caKey, caTemplate, clientKey)

    return &CertBundle{
        CA:         caTemplate,
        ServerKey:  serverKey,
        ServerCert: serverCert,
        ClientKey:  clientKey,
        ClientCert: clientCert,
    }, nil
}
```

### OpenVPN Server Configuration

```ini
port 1194
proto udp
dev tun
ca ca.crt
cert server.crt
key server.key
dh dh2048.pem
server 10.9.0.0 255.255.255.0
push "redirect-gateway def1 bypass-dhcp"
push "dhcp-option DNS 1.1.1.1"
push "dhcp-option DNS 8.8.8.8"
keepalive 10 120
cipher AES-256-GCM
auth SHA256
user nobody
group nogroup
persist-key
persist-tun
status /var/log/openvpn/status.log
verb 3
```

### OpenVPN Features

- **TLS Authentication**: Certificate-based mutual authentication
- **Cipher Negotiation**: Automatic selection of strongest supported cipher
- **TCP Fallback**: Can use TCP port 443 to bypass firewalls
- **Compression**: Optional LZO/LZ4 compression
- **HMAC Firewall**: tls-auth for DoS protection
- **IPv6 Support**: Dual-stack tunneling

---

## Protocol Selection Guide

### When to Use WireGuard

- **Mobile devices**: Better battery life, faster reconnection
- **High-bandwidth**: Streaming, large downloads
- **Modern systems**: Linux 5.6+, macOS, Windows 10+
- **Low latency**: Gaming, VoIP
- **Default choice**: Recommended for most users

### When to Use OpenVPN

- **Restrictive networks**: Can use TCP port 443 to bypass firewalls
- **Legacy systems**: Wider OS compatibility
- **Corporate environments**: Certificate-based authentication
- **Maximum compatibility**: Works on virtually any platform

---

## Security Comparison

```
┌──────────────────────────────────────────────────────────────────────┐
│                    SECURITY PROPERTIES                                │
├──────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  WireGuard:                                                           │
│  ├── Encryption:     ChaCha20-Poly1305 (AEAD)                        │
│  ├── Key Exchange:   Curve25519 (ECDH)                                │
│  ├── Hashing:        BLAKE2s                                          │
│  ├── Handshake:      Noise Protocol (IKpsk2)                          │
│  └── PFS:            Every handshake generates new keys                │
│                                                                       │
│  OpenVPN:                                                             │
│  ├── Encryption:     AES-256-GCM (AEAD)                              │
│  ├── Key Exchange:   RSA-2048 / ECDHE                                │
│  ├── Hashing:        SHA-256                                          │
│  ├── Handshake:      TLS 1.3                                         │
│  └── PFS:            With ECDHE cipher suites                         │
│                                                                       │
└──────────────────────────────────────────────────────────────────────┘
```

---

## Kill Switch Implementation

Both protocols support a kill switch that prevents traffic leaks when the VPN disconnects:

```
┌──────────────────────────────────────────────────────────────────────┐
│                      KILL SWITCH FLOW                                 │
├──────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  VPN Connected:                                                       │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐       │
│  │  App     │───▶│ iptables │───▶│ wg0/tun0 │───▶│ Internet │       │
│  │  Traffic │    │ ALLOW    │    │ VPN      │    │          │       │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘       │
│                                                                       │
│  VPN Disconnected (Kill Switch ON):                                   │
│  ┌──────────┐    ┌──────────┐                                        │
│  │  App     │───▶│ iptables │───▶ BLOCKED (DROP all non-VPN)         │
│  │  Traffic │    │ DROP     │                                        │
│  └──────────┘    └──────────┘                                        │
│                                                                       │
│  Exception: VPN server endpoint is always allowed                     │
│                                                                       │
└──────────────────────────────────────────────────────────────────────┘
```

### Kill Switch iptables Rules

```bash
# Block all outgoing traffic
iptables -P OUTPUT DROP

# Allow loopback
iptables -A OUTPUT -o lo -j ACCEPT

# Allow VPN server endpoint
iptables -A OUTPUT -d VPN_SERVER_IP -p udp --dport 51820 -j ACCEPT

# Allow traffic through VPN tunnel
iptables -A OUTPUT -o wg0 -j ACCEPT

# Allow established connections
iptables -A OUTPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
```

---

## DNS Leak Protection

```
┌──────────────────────────────────────────────────────────────────────┐
│                    DNS LEAK PROTECTION                                │
├──────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  Without Protection:                                                  │
│  DNS Query ──▶ ISP DNS Server (LEAKS real IP)                        │
│                                                                       │
│  With Protection:                                                     │
│  DNS Query ──▶ VPN Tunnel ──▶ 1.1.1.1 / 8.8.8.8 (encrypted)         │
│                                                                       │
│  Implementation:                                                      │
│  1. Override /etc/resolv.conf with VPN DNS                           │
│  2. Block DNS queries outside tunnel (port 53)                       │
│  3. Route all DNS through wg0/tun0 interface                         │
│                                                                       │
└──────────────────────────────────────────────────────────────────────┘
```
