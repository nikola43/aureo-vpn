# 🚀 Usage Guide

How to use the Aureo VPN desktop application.

---

## Getting Started

### 1. Launch the application
The login screen will appear.

### 2. Configure API URL (optional)
Enter your Aureo VPN API Gateway URL. Default: `http://localhost:8080`

### 3. Login
Enter your email and password, then click "Login".

### 4. Select a VPN node
Browse the list of available nodes. Filter by protocol (WireGuard/OpenVPN) or by country/city. Click on a node to select it.

### 5. Connect to VPN
Click the "Connect" button. The app will create a session with the selected node and update the status to "Connected".

### 6. Monitor your connection
View session information: tunnel IP, protocol, data usage. Real-time statistics update every 5 seconds.

### 7. Disconnect
Click the "Disconnect" button to terminate the session.

---

## ⚙️ Configuration

### API URL

The default API URL is `http://localhost:8080`. You can change this in the login screen before logging in.

To permanently change the default, modify the `startup` function in `app.go`:

```go
func (a *App) startup(ctx context.Context) {
    a.ctx = ctx
    a.apiClient = api.NewClient("http://your-api-gateway:8080")
}
```

---

## 🔐 VPN Protocols

### WireGuard (Recommended)
- Modern, fast, and secure VPN protocol
- Uses UDP for optimal performance
- Recommended for most use cases

### OpenVPN
- Established VPN protocol
- More compatible with restricted networks
- TCP and UDP support

---

## 🐛 Troubleshooting

### Cannot connect to API
- Ensure the API Gateway is running
- Check the API URL in the login screen
- Verify network connectivity

---

## 🛡️ Security Notes

- Access tokens are stored in memory only
- Passwords are never stored locally
- All API communication should use HTTPS in production
- VPN configurations contain sensitive keys — handle with care
