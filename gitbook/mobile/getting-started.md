# 🚀 Getting Started

Set up the Aureo VPN mobile app for development.

---

## 📦 Prerequisites

- Node.js 18+
- npm or yarn
- iOS: Xcode 15+ (for simulator)
- Android: Android Studio with emulator

---

## ⚙️ Installation

```bash
cd aureo-app

# Install dependencies
npm install

# Start Expo dev server
npm start

# Run on iOS Simulator
npm run ios

# Run on Android Emulator
npm run android
```

---

## 🔧 Configuration

The default API URL is `https://api.aureovpn.com`. To change it:

1. Open the app → Settings → API URL
2. Or modify the default in `src/stores/settings.store.ts`

---

## 📝 Path Aliases

`@/*` maps to the project root, configured in `tsconfig.json`:

```typescript
import { useAuthStore } from '@/src/stores/auth.store';
import { Colors } from '@/constants/theme';
```

---

## ⚡ Experiments

| Feature | Status |
|---------|--------|
| React Compiler | ✅ Enabled |
| Typed Routes | ✅ Enabled |
| New Architecture | ✅ Enabled |

---

## 🔍 Linting

```bash
npm run lint          # ESLint (expo lint)
```
