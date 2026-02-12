# 🔨 Building

Build the Aureo VPN desktop application for development or production.

---

## 📦 Prerequisites

### 1. Go (1.21 or later)
```bash
go version
```

### 2. Wails CLI
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 3. Node.js (for frontend dependencies, if needed)

### 4. Platform-specific requirements
- **macOS**: Xcode Command Line Tools
- **Windows**: WebView2
- **Linux**: GTK3 and webkit2gtk

---

## ⚡ Development Mode

Run the application in development mode with hot reload:

```bash
cd aureo-desktop
wails dev
```

---

## 📦 Production Build

Build the application for your platform:

```bash
# Build for current platform
wails build

# Build for specific platform
wails build -platform darwin/universal    # macOS (Universal binary)
wails build -platform windows/amd64       # Windows 64-bit
wails build -platform linux/amd64         # Linux 64-bit
```

The built application will be in the `build/bin` directory.

---

## 🐛 Build Troubleshooting

```bash
# Clean build cache
rm -rf build/

# Update Wails
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Install dependencies
go mod download
```

### Frontend not loading

```bash
# Ensure frontend directory exists
ls -la frontend/dist/

# Verify files are present
ls -la frontend/dist/*.html
```
