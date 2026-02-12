# 🖥️ Desktop App Overview

Cross-platform VPN application built with Go and Wails, available for macOS, Windows, and Linux.

---

## ✨ Features

- 🔒 **VPN Connection** — Connect to VPN using WireGuard or OpenVPN
- 🗺️ **Interactive World Map** — Leaflet.js-powered map with node markers and load indicators
- 🌐 **Server Browser** — Browse nodes by country and protocol with real-time load data
- ⚡ **Quick Actions** — Quick Connect, Secure Core, P2P Friendly, Random Server
- 📊 **Real-Time Stats** — Live upload/download speed, data transferred, connection timer
- 🔑 **User Authentication** — Login/register with session persistence
- 🎨 **Premium Dark UI** — Gold-accented design with color-coded load indicators

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Aureo Desktop App                         │
│                                                               │
│  ┌─────────────────────┐     ┌──────────────────────────┐   │
│  │   Frontend (Web)    │     │      Go Backend          │   │
│  │                     │     │                          │   │
│  │  HTML/JS/CSS        │◄───►│  App struct              │   │
│  │  Leaflet.js Map     │     │  ├── api.Client          │   │
│  │  Tabs & Controls    │     │  ├── vpn.WireGuardMgr    │   │
│  │                     │     │  ├── Session state        │   │
│  │  Wails JS Bindings  │     │  └── Config (~/.aureo)   │   │
│  └─────────────────────┘     └──────────┬───────────────┘   │
│                                          │                    │
└──────────────────────────────────────────┼────────────────────┘
                                           │ HTTPS/REST
                                           ▼
                                ┌──────────────────────┐
                                │   Aureo API Gateway   │
                                │   (aureo-vpn backend)  │
                                └──────────────────────┘
```

---

## Project Structure

```
aureo-desktop/
├── main.go                    # Main entry point
├── app.go                     # Application logic exposed to frontend
├── wails.json                 # Wails configuration
├── go.mod                     # Go dependencies
├── internal/
│   ├── api/
│   │   └── client.go         # API client for HTTP requests
│   └── models/
│       └── models.go         # Data models
└── frontend/
    └── dist/
        ├── index.html        # Main HTML
        ├── style.css         # Styling
        └── app.js            # Frontend JavaScript
```
