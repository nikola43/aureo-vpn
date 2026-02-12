# 🏗️ Mobile Architecture

The mobile app follows a layered architecture with file-based routing, Zustand stores, and an Axios-based API client.

---

## 📁 Directory Structure

```
aureo-app/
├── app/                              # Expo Router (file-based routing)
│   ├── (auth)/                       # Auth flow
│   │   ├── login.tsx                 # Login screen
│   │   ├── register.tsx              # Registration screen
│   │   └── _layout.tsx               # Auth stack layout
│   ├── (tabs)/                       # Main tab interface
│   │   ├── index.tsx                 # Home — connection dashboard
│   │   ├── countries.tsx             # Countries — server browser
│   │   ├── profiles.tsx              # Profiles — saved configs
│   │   ├── settings.tsx              # Settings — preferences
│   │   └── _layout.tsx               # Tab bar layout
│   ├── profile.tsx                   # User profile (modal)
│   ├── change-password.tsx           # Password change (modal)
│   └── _layout.tsx                   # Root layout
├── src/
│   ├── api/                          # API layer
│   │   ├── client.ts                 # Axios + JWT interceptors
│   │   ├── auth.ts                   # Auth endpoints
│   │   ├── nodes.ts                  # Node endpoints
│   │   ├── sessions.ts               # Session endpoints
│   │   ├── config.ts                 # VPN config endpoints
│   │   ├── user.ts                   # User endpoints
│   │   └── types.ts                  # TypeScript interfaces
│   ├── stores/                       # Zustand state management
│   │   ├── auth.store.ts             # Auth, tokens, session restore
│   │   ├── vpn.store.ts              # VPN connection, stats, keys
│   │   ├── settings.store.ts         # User preferences
│   │   └── profiles.store.ts         # VPN profiles
│   ├── hooks/                        # Custom React hooks
│   │   ├── useConnectionTimer.ts     # HH:MM:SS timer
│   │   ├── useGeoLocation.ts         # Real IP geolocation
│   │   ├── useNetworkStatus.ts       # Network connectivity
│   │   ├── useNodes.ts               # Node list (React Query)
│   │   └── useUserStats.ts           # Usage statistics
│   ├── native/
│   │   └── VPNModule.ts              # Native tunnel bridge
│   └── providers/
│       ├── QueryProvider.tsx          # React Query config
│       └── ToastProvider.tsx          # Toast notifications
├── constants/
│   └── theme.ts                      # Design system tokens
├── plugins/
│   ├── withVPNExtension.js           # iOS NetworkExtension
│   └── withVPNService.js             # Android VpnService
└── components/                       # Reusable UI components
```

---

## 🗂️ State Management (Zustand)

| Store | Key State | Persistence |
|-------|-----------|-------------|
| `useAuthStore` | `user`, `accessToken`, `refreshToken`, `isAuthenticated` | Expo SecureStore |
| `useVPNStore` | `connectionStatus`, `activeSession`, `connectedNode`, speeds, bytes | Memory |
| `useSettingsStore` | `protocol`, `killSwitch`, `autoConnect`, `apiBaseUrl`, `theme` | Expo SecureStore (v4) |
| `useProfilesStore` | `profiles[]` (name, icon, color, connectionType, protocol) | Expo SecureStore |

### 🔑 Auth Store

- **Secure Storage:** Tokens in `expo-secure-store` (keys: `aureo_access_token`, `aureo_refresh_token`, `aureo_user_data`)
- **Actions:** `login()`, `register()`, `logout()`, `refreshTokens()`, `restoreSession()`
- **Session Restore:** On app launch, reads SecureStore → optimistic auth with cached user → background profile refresh → fallback to token refresh → logout if both fail
- **Circular Dependency Prevention:** Uses `require()` for lazy-loading related stores

### 🔗 VPN Store

- **State:** `connectionStatus`, `activeSession`, `connectedNode`, `clientIP`, speeds, bytes, recent connections
- **Key Generation:** `tweetnacl.box.keyPair()` for WireGuard X25519 keys, seeded with `expo-crypto` PRNG
- **Connection Flow:**
  1. Generate WireGuard key pair
  2. `POST /config/generate` with public key
  3. Start native tunnel via `VPNModule.startTunnel(config)`
  4. Poll stats every 3 seconds (native + API fallback)
- **Quick Connect:** `GET /nodes/best` → auto-select protocol → connect

---

## 🔗 API Layer

The API client (`src/api/client.ts`) uses **Axios** with:

- **JWT Interceptor** — Attaches `Bearer` token to every request
- **Auto-Refresh** — On 401, mutex-protected token refresh + request retry
- **Dynamic Base URL** — Resolved per-request from settings store
- **Error Normalization** — All errors wrapped in `ApiError` class
- **Circular Dependency Prevention** — Stores accessed via `require()` at call-time

---

## 🧭 Navigation Flow

```
Root Stack
├── (auth)/                    # Shown when !isAuthenticated
│   ├── login
│   └── register
├── (tabs)/                    # Shown when isAuthenticated
│   ├── Home        (🏠)      # Connection dashboard, speed metrics
│   ├── Countries   (🌐)      # Server browser with flags
│   ├── Profiles    (📋)      # Saved VPN configurations
│   └── Settings    (⚙️)      # Preferences, account
└── Modals
    ├── profile                # User profile sheet
    └── change-password        # Password change sheet
```
