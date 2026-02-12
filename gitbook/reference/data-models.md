# 🗂️ Data Models

All models use UUID primary keys with GORM. Database: PostgreSQL (production) or SQLite (development/nodes).

---

## User

Represents a VPN service user.

| Field | Type | GORM Tags | JSON | Description |
|---|---|---|---|---|
| `ID` | `uuid.UUID` | `type:uuid;primary_key;default:gen_random_uuid()` | `id` | Primary key, auto-generated UUID |
| `Email` | `string` | `uniqueIndex;not null` | `email` | Unique email address |
| `PasswordHash` | `string` | `not null` | `-` (hidden) | bcrypt/Argon2id hashed password |
| `Username` | `string` | `uniqueIndex;not null` | `username` | Unique username |
| `FullName` | `string` | | `full_name` | Display name |
| `IsActive` | `bool` | `default:true` | `is_active` | Account active status |
| `IsAdmin` | `bool` | `default:false` | `is_admin` | Admin role flag |
| `SubscriptionTier` | `string` | `default:'free'` | `subscription_tier` | `free`, `basic`, `premium` |
| `SubscriptionExpiry` | `time.Time` | | `subscription_expiry` | When subscription expires |
| `DataTransferredGB` | `float64` | `default:0` | `data_transferred_gb` | Cumulative data usage |
| `ConnectionCount` | `int64` | `default:0` | `connection_count` | Total connections ever made |
| `TwoFactorEnabled` | `bool` | `default:false` | `two_factor_enabled` | 2FA enabled flag |
| `TwoFactorSecret` | `string` | | `-` (hidden) | TOTP secret key |
| `CreatedAt` | `time.Time` | | `created_at` | Creation timestamp |
| `UpdatedAt` | `time.Time` | | `updated_at` | Last update timestamp |
| `DeletedAt` | `gorm.DeletedAt` | `index` | `-` (hidden) | Soft delete timestamp |

**Relations:**

- `Sessions []Session` — `foreignKey:UserID` — All VPN sessions for this user
- `Configs []Config` — `foreignKey:UserID` — All VPN configs for this user

**Hooks:** `BeforeCreate` auto-generates UUID if nil.

---

## VPNNode

Represents a VPN server node in the network.

| Field | Type | GORM Tags | JSON | Description |
|---|---|---|---|---|
| `ID` | `uuid.UUID` | `type:uuid;primary_key;default:gen_random_uuid()` | `id` | Primary key |
| `Name` | `string` | `uniqueIndex;not null` | `name` | Unique node name |
| `Hostname` | `string` | `not null` | `hostname` | System hostname |
| `Country` | `string` | `not null;index` | `country` | Country name |
| `CountryCode` | `string` | `size:2;not null` | `country_code` | ISO 3166-1 alpha-2 |
| `City` | `string` | `not null` | `city` | City name |
| `Latitude` | `float64` | | `latitude` | GPS latitude |
| `Longitude` | `float64` | | `longitude` | GPS longitude |
| `PublicIP` | `string` | `not null` | `public_ip` | Public-facing IP |
| `InternalIP` | `string` | | `internal_ip` | WireGuard tunnel base IP |
| `IPv6Address` | `string` | | `ipv6_address` | IPv6 address |
| `MaxConnections` | `int` | `default:1000` | `max_connections` | Connection capacity |
| `CurrentConnections` | `int` | `default:0` | `current_connections` | Active connections |
| `CPUUsage` | `float64` | `default:0` | `cpu_usage` | CPU usage percentage |
| `MemoryUsage` | `float64` | `default:0` | `memory_usage` | Memory usage percentage |
| `BandwidthUsageGbps` | `float64` | `default:0` | `bandwidth_usage_gbps` | Current bandwidth rate |
| `TotalBandwidthKB` | `int64` | `default:0` | `total_bandwidth_kb` | Cumulative traffic in KB |
| `LoadScore` | `float64` | `default:0;index` | `load_score` | 0-100, lower is better |
| `Status` | `string` | `default:'offline'` | `status` | `online`, `offline`, `maintenance` |
| `IsActive` | `bool` | `default:true` | `is_active` | Node enabled flag |
| `LastHeartbeat` | `time.Time` | | `last_heartbeat` | Last heartbeat received |
| `LastHealthCheck` | `time.Time` | | `last_health_check` | Last health check performed |
| `Latency` | `int` | | `latency` | Latency in milliseconds |
| `SupportsWireGuard` | `bool` | `default:true` | `supports_wireguard` | WireGuard support |
| `SupportsOpenVPN` | `bool` | `default:true` | `supports_openvpn` | OpenVPN support |
| `WireGuardPort` | `int` | `default:51820` | `wireguard_port` | WireGuard listen port |
| `OpenVPNPort` | `int` | `default:1194` | `openvpn_port` | OpenVPN listen port |
| `PublicKey` | `string` | | `public_key` | WireGuard public key |
| `PrivateKeyEncrypted` | `string` | | `-` (hidden) | Encrypted WG private key |
| `SupportsMultiHop` | `bool` | `default:false` | `supports_multihop` | Multi-hop relay support |
| `SupportsObfuscation` | `bool` | `default:false` | `supports_obfuscation` | Traffic obfuscation |
| `SupportsSOCKS5` | `bool` | `default:false` | `supports_socks5` | SOCKS5 proxy support |
| `Version` | `string` | | `version` | Software version |
| `Tags` | `string` | | `tags` | Comma-separated tags |
| `Priority` | `int` | `default:0` | `priority` | Selection priority |
| `OperatorID` | `*uuid.UUID` | `type:uuid;index` | `operator_id` | Owning operator (nullable) |
| `IsOperatorOwned` | `bool` | `default:false` | `is_operator_owned` | Decentralized node flag |
| `UptimePercentage` | `float64` | `type:decimal(5,2);default:0` | `uptime_percentage` | Historical uptime % |
| `TotalEarnedUSD` | `float64` | `type:decimal(20,8);default:0` | `total_earned_usd` | Cumulative earnings |
| `CreatedAt` | `time.Time` | | `created_at` | |
| `UpdatedAt` | `time.Time` | | `updated_at` | |
| `DeletedAt` | `gorm.DeletedAt` | `index` | `-` | Soft delete |

**Relations:** `Sessions []Session` — `foreignKey:NodeID`

### Load Score Formula

```go
func (n *VPNNode) CalculateLoadScore() float64 {
    connectionLoad := float64(n.CurrentConnections) / float64(n.MaxConnections) * 100
    cpuLoad := n.CPUUsage
    memoryLoad := n.MemoryUsage

    // Weighted average: 40% connections, 30% CPU, 30% memory
    return (connectionLoad * 0.4) + (cpuLoad * 0.3) + (memoryLoad * 0.3)
}
```

### Health Check

A node is considered healthy when:

1. `IsActive` is `true` AND `Status` is `"online"`
2. `LastHeartbeat` is within the last 2 minutes
3. `LoadScore` is below 90

---

## Session

Represents an active or historical VPN connection session.

| Field | Type | GORM Tags | JSON | Description |
|---|---|---|---|---|
| `ID` | `uuid.UUID` | `type:uuid;primary_key;default:gen_random_uuid()` | `id` | Primary key |
| `UserID` | `uuid.UUID` | `type:uuid;not null;index` | `user_id` | Owning user |
| `NodeID` | `uuid.UUID` | `type:uuid;not null;index` | `node_id` | Connected node |
| `Protocol` | `string` | `not null` | `protocol` | `wireguard`, `openvpn` |
| `ClientIP` | `string` | `not null` | `client_ip` | Anonymized client IP hash |
| `TunnelIP` | `string` | `not null` | `tunnel_ip` | Allocated tunnel IP |
| `PublicKey` | `string` | | `public_key` | WireGuard public key |
| `PrivateKey` | `string` | | `-` (hidden) | Encrypted, never exposed |
| `Status` | `string` | `default:'active'` | `status` | Session status (see below) |
| `ConnectedAt` | `time.Time` | `not null` | `connected_at` | Connection start time |
| `DisconnectedAt` | `*time.Time` | | `disconnected_at` | Connection end time |
| `BytesSent` | `int64` | `default:0` | `bytes_sent` | Upload bytes |
| `BytesReceived` | `int64` | `default:0` | `bytes_received` | Download bytes |
| `DataUsedGB` | `float64` | `default:0` | `data_used_gb` | Total data in GB |
| `Latency` | `int` | | `latency` | Latency in ms |
| `PacketLoss` | `float64` | | `packet_loss` | Loss percentage |
| `LastKeepalive` | `time.Time` | | `last_keepalive` | Last keepalive received |
| `SplitTunnelEnabled` | `bool` | `default:false` | `split_tunnel_enabled` | Split tunneling |
| `KillSwitchEnabled` | `bool` | `default:true` | `kill_switch_enabled` | Kill switch active |
| `DNSLeakProtection` | `bool` | `default:true` | `dns_leak_protection` | DNS leak protection |
| `IsMultiHop` | `bool` | `default:false` | `is_multihop` | Multi-hop enabled |
| `NextHopNodeID` | `*uuid.UUID` | `type:uuid` | `next_hop_node_id` | Next hop for multi-hop |
| `ClientVersion` | `string` | | `client_version` | Client app version |
| `DeviceType` | `string` | | `device_type` | `desktop`, `mobile`, `router` |
| `OSType` | `string` | | `os_type` | `linux`, `macos`, `windows`, `ios`, `android` |
| `CreatedAt` | `time.Time` | | `created_at` | |
| `UpdatedAt` | `time.Time` | | `updated_at` | |
| `DeletedAt` | `gorm.DeletedAt` | `index` | `-` | Soft delete |

**Status Values:**

| Status | Description |
|---|---|
| `pending` | Session created, waiting for node to provision WireGuard peer |
| `active` | Tunnel is up, traffic flowing |
| `disconnected` | Graceful disconnect completed |
| `terminated` | Force-terminated by server |
| `pending_disconnect` | Client requested disconnect, waiting for node to clean up |

**Relations:**

- `User *User` — `foreignKey:UserID`
- `Node *VPNNode` — `foreignKey:NodeID`

---

## NodeOperator

Represents a user who operates a decentralized VPN node.

| Field | Type | GORM Tags | JSON | Description |
|---|---|---|---|---|
| `ID` | `uuid.UUID` | `type:uuid;primary_key` | `id` | Primary key |
| `UserID` | `uuid.UUID` | `type:uuid;not null;uniqueIndex` | `user_id` | Linked user account |
| `WalletAddress` | `string` | `type:varchar(255);uniqueIndex` | `wallet_address` | Crypto payout address |
| `WalletType` | `string` | `type:varchar(50);default:'ethereum'` | `wallet_type` | `ethereum`, `bitcoin`, `litecoin` |
| `Status` | `string` | `type:varchar(50);default:'pending'` | `status` | `pending`, `active`, `suspended`, `banned` |
| `IsVerified` | `bool` | `default:false` | `is_verified` | Verification status |
| `VerifiedAt` | `*time.Time` | | `verified_at` | Verification timestamp |
| `TotalEarned` | `float64` | `type:decimal(20,8);default:0` | `total_earned` | Cumulative earnings USD |
| `PendingPayout` | `float64` | `type:decimal(20,8);default:0` | `pending_payout` | Pending payout USD |
| `LastPayoutAt` | `*time.Time` | | `last_payout_at` | Last payout timestamp |
| `TotalNodesCreated` | `int` | `default:0` | `total_nodes_created` | Nodes ever created |
| `ActiveNodesCount` | `int` | `default:0` | `active_nodes_count` | Currently active nodes |
| `TotalBandwidthKB` | `int64` | `default:0` | `total_bandwidth_kb` | Total bandwidth served |
| `TotalConnectionsServed` | `int64` | `default:0` | `total_connections_served` | Connections served |
| `AverageUptime` | `float64` | `type:decimal(5,2);default:0` | `average_uptime` | Average uptime % |
| `ReputationScore` | `float64` | `type:decimal(5,2);default:50` | `reputation_score` | 0-100 score |
| `StakeAmount` | `float64` | `type:decimal(20,8);default:0` | `stake_amount` | Staked deposit USD |
| `StakeStatus` | `string` | `type:varchar(50);default:'none'` | `stake_status` | `none`, `staked`, `locked`, `slashed` |
| `StakedAt` | `*time.Time` | | `staked_at` | Stake timestamp |
| `Email` | `string` | `type:varchar(255)` | `email` | Contact email |
| `PhoneNumber` | `string` | `type:varchar(50)` | `phone_number` | Contact phone |
| `Country` | `string` | `type:varchar(100)` | `country` | Operator country |
| `KYCStatus` | `string` | `type:varchar(50);default:'not_required'` | `kyc_status` | `not_required`, `pending`, `approved`, `rejected` |
| `KYCSubmittedAt` | `*time.Time` | | `kyc_submitted_at` | KYC submission time |
| `TaxID` | `string` | `type:varchar(100)` | `tax_id` | Tax ID for compliance |
| `CreatedAt` | `time.Time` | | `created_at` | |
| `UpdatedAt` | `time.Time` | | `updated_at` | |
| `DeletedAt` | `gorm.DeletedAt` | `index` | `-` | Soft delete |

**Relations:**

- `User *User` — `foreignKey:UserID`
- `Nodes []VPNNode` — `foreignKey:OperatorID`
- `Earnings []OperatorEarning` — `foreignKey:OperatorID`
- `Payouts []OperatorPayout` — `foreignKey:OperatorID`

### Reputation Score Calculation

```go
func (op *NodeOperator) CalculateReputationScore(db *gorm.DB) float64 {
    score := 50.0 // Base score

    // Uptime contribution (max 30 points)
    score += (op.AverageUptime / 100.0) * 30.0

    // User ratings contribution (max 20 points)
    var avgRating float64
    db.Model(&OperatorEarning{}).
        Where("operator_id = ? AND user_rating > 0", op.ID).
        Select("COALESCE(AVG(user_rating), 0)").Scan(&avgRating)
    score += (avgRating / 5.0) * 20.0

    // Bandwidth served (max 10 points): 1TB=10, 100GB=5
    // Stake contribution (max 10 points): $1000=10, $100=5

    if score > 100 { score = 100 }
    return score
}
```

---

## OperatorEarning

Tracks individual earning events for node operators.

| Field | Type | GORM Tags | JSON | Description |
|---|---|---|---|---|
| `ID` | `uuid.UUID` | `type:uuid;primary_key` | `id` | Primary key |
| `OperatorID` | `uuid.UUID` | `type:uuid;not null;index` | `operator_id` | Owning operator |
| `NodeID` | `uuid.UUID` | `type:uuid;not null;index` | `node_id` | Node that earned |
| `SessionID` | `uuid.UUID` | `type:uuid;not null;index` | `session_id` | Source session |
| `BandwidthKB` | `int64` | `not null` | `bandwidth_kb` | Bandwidth served |
| `DurationMinutes` | `int` | `not null` | `duration_minutes` | Session duration chunk |
| `RatePerGB` | `float64` | `type:decimal(10,6);not null` | `rate_per_gb` | USD per GB rate |
| `AmountUSD` | `float64` | `type:decimal(20,8);not null` | `amount_usd` | Earned amount |
| `Status` | `string` | `type:varchar(50);default:'pending'` | `status` | `pending`, `confirmed`, `paid` |
| `PaidAt` | `*time.Time` | | `paid_at` | Payment timestamp |
| `ConnectionQuality` | `float64` | `type:decimal(5,2)` | `connection_quality` | 0-100 quality score |
| `UserRating` | `int` | `check:user_rating >= 1 AND user_rating <= 5` | `user_rating` | 1-5 star rating |
| `CreatedAt` | `time.Time` | | `created_at` | |

---

## OperatorPayout

Tracks payout transactions to node operators.

| Field | Type | GORM Tags | JSON | Description |
|---|---|---|---|---|
| `ID` | `uuid.UUID` | `type:uuid;primary_key` | `id` | Primary key |
| `OperatorID` | `uuid.UUID` | `type:uuid;not null;index` | `operator_id` | Receiving operator |
| `AmountUSD` | `float64` | `type:decimal(20,8);not null` | `amount_usd` | USD amount |
| `CryptoAmount` | `float64` | `type:decimal(30,18);not null` | `crypto_amount` | Crypto amount |
| `CryptoCurrency` | `string` | `type:varchar(50);not null` | `crypto_currency` | `ETH`, `BTC`, `LTC` |
| `ExchangeRate` | `float64` | `type:decimal(20,8);not null` | `exchange_rate` | USD per crypto unit |
| `WalletAddress` | `string` | `type:varchar(255);not null` | `wallet_address` | Destination wallet |
| `TransactionHash` | `string` | `type:varchar(255)` | `transaction_hash` | Blockchain tx hash |
| `TransactionFee` | `float64` | `type:decimal(20,8)` | `transaction_fee` | Network fee |
| `Status` | `string` | `type:varchar(50);default:'pending'` | `status` | `pending`, `processing`, `completed`, `failed` |
| `ProcessedAt` | `*time.Time` | | `processed_at` | Processing start |
| `CompletedAt` | `*time.Time` | | `completed_at` | Completion time |
| `FailureReason` | `string` | `type:text` | `failure_reason` | Error details |
| `PayoutMethod` | `string` | `type:varchar(50)` | `payout_method` | `blockchain`, `exchange`, `manual` |
| `Notes` | `string` | `type:text` | `notes` | Admin notes |
| `CreatedAt` | `time.Time` | | `created_at` | |
| `UpdatedAt` | `time.Time` | | `updated_at` | |

---

## NodeReward

Defines reward tier configuration for node operators.

| Field | Type | GORM Tags | JSON | Description |
|---|---|---|---|---|
| `ID` | `uuid.UUID` | `type:uuid;primary_key` | `id` | Primary key |
| `TierName` | `string` | `type:varchar(50);uniqueIndex` | `tier_name` | `bronze`, `silver`, `gold`, `platinum` |
| `MinReputationScore` | `float64` | `type:decimal(5,2)` | `min_reputation_score` | Minimum reputation required |
| `MinUptimePercent` | `float64` | `type:decimal(5,2)` | `min_uptime_percent` | Minimum uptime required |
| `BaseRatePerGB` | `float64` | `type:decimal(10,6);not null` | `base_rate_per_gb` | USD per GB base rate |
| `BonusMultiplier` | `float64` | `type:decimal(5,2);default:1.0` | `bonus_multiplier` | Earnings multiplier |
| `MinBandwidth` | `int` | `default:100` | `min_bandwidth` | Minimum Mbps required |
| `MaxLatency` | `int` | `default:100` | `max_latency` | Maximum latency ms |
| `IsActive` | `bool` | `default:true` | `is_active` | Tier enabled |
| `CreatedAt` | `time.Time` | | `created_at` | |
| `UpdatedAt` | `time.Time` | | `updated_at` | |

---

## NodePerformanceMetric

Tracks node performance over time in hourly/daily windows.

| Field | Type | GORM Tags | JSON | Description |
|---|---|---|---|---|
| `ID` | `uuid.UUID` | `type:uuid;primary_key` | `id` | Primary key |
| `NodeID` | `uuid.UUID` | `type:uuid;not null;index` | `node_id` | Measured node |
| `MetricDate` | `time.Time` | `type:date;not null;index` | `metric_date` | Date of measurement |
| `Hour` | `int` | `check:hour >= 0 AND hour <= 23` | `hour` | Hour of day (0-23) |
| `UptimeMinutes` | `int` | `not null` | `uptime_minutes` | Minutes online |
| `DowntimeMinutes` | `int` | `default:0` | `downtime_minutes` | Minutes offline |
| `ConnectionsServed` | `int` | `default:0` | `connections_served` | Connections in window |
| `BandwidthGB` | `float64` | `type:decimal(20,4);default:0` | `bandwidth_gb` | GB transferred |
| `AverageLatencyMs` | `int` | `default:0` | `average_latency_ms` | Average latency |
| `AvailabilityScore` | `float64` | `type:decimal(5,2)` | `availability_score` | 0-100 |
| `PerformanceScore` | `float64` | `type:decimal(5,2)` | `performance_score` | 0-100 |
| `UserSatisfaction` | `float64` | `type:decimal(5,2)` | `user_satisfaction` | 0-100 |
| `EarningsUSD` | `float64` | `type:decimal(20,8);default:0` | `earnings_usd` | Earnings in window |
| `CreatedAt` | `time.Time` | | `created_at` | |

---

## Config

Represents VPN configuration files for users.

| Field | Type | GORM Tags | JSON | Description |
|---|---|---|---|---|
| `ID` | `uuid.UUID` | `type:uuid;primary_key;default:gen_random_uuid()` | `id` | Primary key |
| `UserID` | `uuid.UUID` | `type:uuid;not null;index` | `user_id` | Owning user |
| `NodeID` | `uuid.UUID` | `type:uuid;not null;index` | `node_id` | Target node |
| `Protocol` | `string` | `not null` | `protocol` | `wireguard`, `openvpn` |
| `ConfigName` | `string` | `not null` | `config_name` | Human-readable name |
| `ConfigContent` | `string` | `type:text;not null` | `-` (hidden) | Encrypted config content |
| `ConfigHash` | `string` | `not null` | `config_hash` | Integrity hash |
| `PublicKey` | `string` | | `public_key` | WireGuard public key |
| `PrivateKey` | `string` | | `-` (hidden) | Encrypted private key |
| `DNSServers` | `string` | | `dns_servers` | Comma-separated DNS |
| `AllowedIPs` | `string` | | `allowed_ips` | Split tunnel rules |
| `MTU` | `int` | `default:1420` | `mtu` | MTU size |
| `PersistentKeepalive` | `int` | `default:25` | `persistent_keepalive` | Keepalive interval (seconds) |
| `IsActive` | `bool` | `default:true` | `is_active` | Config enabled |
| `LastUsed` | `*time.Time` | | `last_used` | Last usage time |
| `TimesUsed` | `int64` | `default:0` | `times_used` | Usage counter |
| `ExpiresAt` | `*time.Time` | | `expires_at` | Expiry time (nullable) |
| `CreatedAt` | `time.Time` | | `created_at` | |
| `UpdatedAt` | `time.Time` | | `updated_at` | |
| `DeletedAt` | `gorm.DeletedAt` | `index` | `-` | Soft delete |

**Relations:** `User *User`, `Node *VPNNode`

---

## Local Models (VPN Node Database)

Each VPN node maintains its own local SQLite database with simplified models for standalone operation.

### LocalUser

Lightweight user record synced to the node. Fields: `ID`, `Email`, `Username`, `PasswordHash`, `Plan` (free/pro/premium), `ExpiresAt`, `TotalDataUsedKB`, `SessionCount`, timestamps.

### LocalSession

Active session record on the node. Fields: `ID`, `UserID`, `Protocol`, `TunnelIP`, `ClientIP`, `PublicKey`, `Status` (pending/active/disconnected), `ConnectedAt`, `DisconnectedAt`, `LastKeepalive`, `BytesSent`, `BytesReceived`, timestamps.

### LocalNodeConfig

Key-value configuration store. Fields: `ID` (uint), `Key` (unique), `Value`.

### NodeIdentity

Persistent node identity. Fields: `ID` (uint), `NodeID` (UUID string), `Name`, `P2PPrivateKey` (libp2p base64), `P2PPeerID`, `WGPrivateKey`, `WGPublicKey`, `CreatedAt`.

All local models are migrated together via `AllLocalModels()`.
