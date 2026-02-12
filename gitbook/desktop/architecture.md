# ⚡ Desktop Architecture

The desktop app uses Wails 2.x to bridge a Go backend with an embedded web frontend. All Go methods are automatically exposed to JavaScript via Wails bindings.

---

## Go Backend Methods

The `App` struct methods are exposed to the frontend:

### 🔑 Authentication
- `Login(email, password string)` — Authenticate user
- `Register(email, password, username string)` — Create new account
- `Logout()` — Clear session
- `IsLoggedIn()` — Check if user is authenticated

### 📡 Node Management
- `GetNodes(country, protocol string)` — Get list of nodes
- `GetBestNode()` — Get optimal node
- `GetNode(nodeID string)` — Get specific node details

### 🔗 VPN Connection
- `ConnectToVPN(nodeID, protocol string)` — Connect to VPN
- `DisconnectVPN()` — Disconnect from VPN
- `IsConnected()` — Check connection status
- `GetCurrentSession()` — Get active session details
- `GetAllSessions()` — Get all user sessions

### 👤 User Info
- `GetCurrentUser()` — Get logged-in user
- `GetUserProfile()` — Get user profile from API
- `GetUserStats()` — Get usage statistics

### ⚙️ Configuration
- `SetAPIURL(url string)` — Set API base URL
- `GenerateConfig(nodeID, protocol string)` — Generate VPN config

---

## 🔒 WireGuard Management

The `internal/vpn/wireguard.go` module handles:

- **Key Generation:** `wg genkey` + `wg pubkey` commands
- **Config Path:** `~/.aureo-vpn/wg0.conf`
- **Connection (macOS):** `wg-quick up` via `osascript` for admin elevation
- **Disconnection:** `wg-quick down` with admin privileges
- **Stats:** Parses `sudo wg show` output for bytes sent/received and last handshake

---

## 💾 Session Persistence

Sessions stored in `~/.aureo-vpn/session.json` with 0600 permissions:

```json
{
  "access_token": "...",
  "refresh_token": "...",
  "user": { ... },
  "api_url": "https://api.aureovpn.com"
}
```

---

## 🎨 Frontend UI

- **Design:** Premium dark theme (Gold #F59E0B, Cyber Blue #3B82F6, Dark #030712)
- **Map:** Leaflet.js with node markers, color-coded load indicators
- **Tabs:** Servers (search + list), Stats, Settings
- **Quick Actions:** Quick Connect, Secure Core, P2P Friendly, Random
- **Real-time:** Speed monitoring every 2s, connection timer every 1s
