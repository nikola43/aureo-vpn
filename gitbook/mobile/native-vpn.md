# 📱 Native VPN Integration

Platform-specific VPN tunnel implementations for iOS and Android.

---

## 🍎 iOS — NetworkExtension

The `withVPNExtension.js` plugin configures:
- `packet-tunnel-provider` entitlement
- `com.apple.developer.networking.vpn.api` capability
- App Group for IPC between app and tunnel extension
- Background modes: `network-authentication`, `fetch`

{% hint style="warning" %}
Requires a separate `AureoTunnel` target in Xcode with WireGuardKit SPM dependency.
{% endhint %}

---

## 🤖 Android — VpnService

The `withVPNService.js` plugin configures:
- `BIND_VPN_SERVICE` permission on `AureoVpnService`
- `FOREGROUND_SERVICE_SPECIAL_USE` (API 34+)
- VPN intent filter and foreground service type

---

## 🔌 Native Bridge API

```typescript
VPNModule = {
  startTunnel(config: TunnelConfig): Promise<void>,
  stopTunnel(): Promise<void>,
  getStatus(): Promise<VPNStatus>,
  getStatistics(): Promise<VPNStatistics>,
  onStatusChanged(callback): Subscription,
  onStatsUpdated(callback): Subscription,
}
```

- **Native Module Name:** `AureoVPN` (registered on both platforms)
- **Events:** `onVPNStatusChanged`, `onVPNStatsUpdated`
- **Dev Mode:** Graceful fallback when native module unavailable

WireGuard keys are generated client-side using `tweetnacl.box.keyPair()` seeded with `expo-crypto` PRNG for X25519 key exchange.
