# 🎛️ Control Server

The Control Server is the network orchestrator that monitors VPN node health, recalculates load scores, and cleans up stale data. It does not handle user-facing traffic -- it manages the infrastructure.

Source: `internal/control/server.go`

---

## Background Tasks

The control server runs three background loops:

| Task | Interval | Description |
|---|---|---|
| `healthCheckLoop` | 1 minute | Checks all active nodes' heartbeat timestamps. Marks nodes as `offline` if no heartbeat received in > 2 minutes. Marks nodes back as `online` if heartbeats resume. Updates `last_health_check` timestamp |
| `loadBalancerLoop` | 30 seconds | Recalculates `LoadScore` for all online, active nodes using `CalculateLoadScore()` (40% connections, 30% CPU, 30% memory). Logs a warning if any node's load exceeds 80 |
| `cleanupLoop` | 1 hour | **Old sessions:** Deletes (soft delete) disconnected sessions older than 30 days. **Expired configs:** Deletes configs past their `expires_at` timestamp. **Orphaned sessions:** Finds sessions with `status=active` on nodes with `status=offline`, marks them as `disconnected` with current timestamp |

---

## Health Check Logic

```
For each node WHERE is_active = true:

  IF time.Since(node.LastHeartbeat) > 2 minutes:
    UPDATE status = 'offline'
    LOG "Node {name} marked as offline (no heartbeat)"

  ELSE IF node.Status != 'online':
    UPDATE status = 'online'
    LOG "Node {name} is back online"

  UPDATE last_health_check = NOW()
```

---

## Load Balancer Logic

```
For each node WHERE is_active = true AND status = 'online':

  loadScore = node.CalculateLoadScore()
  UPDATE load_score = loadScore

  IF loadScore > 80:
    LOG WARNING "Node {name} is heavily loaded (score: {loadScore})"
```

The load score formula:

```
LoadScore = (ConnectionLoad * 0.4) + (CPUUsage * 0.3) + (MemoryUsage * 0.3)

Where ConnectionLoad = (CurrentConnections / MaxConnections) * 100
```

---

## Cleanup Logic

```
1. DELETE sessions
   WHERE status = 'disconnected'
   AND disconnected_at < (NOW - 30 days)

2. DELETE configs
   WHERE expires_at IS NOT NULL
   AND expires_at < NOW

3. SELECT sessions
   JOIN vpn_nodes ON sessions.node_id = vpn_nodes.id
   WHERE sessions.status = 'active'
   AND vpn_nodes.status = 'offline'

   For each orphaned session:
     UPDATE status = 'disconnected', disconnected_at = NOW
```

---

## Node Selection

When the API or control server needs to find the best available node for a new connection:

```sql
SELECT * FROM vpn_nodes
WHERE is_active = true
  AND status = 'online'
  AND country_code = ?       -- optional filter
  AND supports_wireguard = ? -- or supports_openvpn, based on protocol
ORDER BY load_score ASC, latency ASC
LIMIT 1
```

The node with the lowest load score (and lowest latency as tiebreaker) is returned to the client.

---

## Additional Functions

| Function | Description |
|---|---|
| `RegisterNode(node)` | Registers a new VPN node. Sets initial status to `offline`, `is_active` to `true`, and `last_heartbeat` to now |
| `UpdateNodeStatus(nodeID, status)` | Updates a node's status and heartbeat timestamp |
| `GetBestNode(protocol, country)` | Returns the best node by load score (see query above) |
| `GetNodeStats()` | Returns aggregate stats: total/online/offline nodes, active connections, nodes by country |
