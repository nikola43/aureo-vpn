# Aureo VPN - Decentralized Node

Run your own VPN node in the Aureo decentralized network with a single command.

## Features

- **Fully Decentralized** - No central server, each node is independent
- **P2P Discovery** - Nodes find each other via libp2p (DHT + GossipSub)
- **SQLite Database** - Local storage, no external database required
- **WireGuard VPN** - Fast, modern VPN protocol
- **Auto-detection** - Automatically detects public IP and location
- **Docker-based** - Easy deployment with Docker

## Quick Start

### One-Command Setup

```bash
curl -fsSL https://raw.githubusercontent.com/aureo-vpn/setup.sh | sudo bash
```

### Manual Setup

1. Clone the repository:
```bash
git clone https://github.com/aureo-vpn/aureo-vpn.git
cd aureo-vpn/deployments/decentralized
```

2. Start the node:
```bash
# Using environment variables
NODE_NAME=my-node COUNTRY_CODE=US docker compose up -d

# Or edit docker-compose.yml and run
docker compose up -d
```

## Configuration

Environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `NODE_NAME` | Name for your node | hostname |
| `PUBLIC_IP` | Your public IP | auto-detected |
| `COUNTRY` | Country name | auto-detected |
| `COUNTRY_CODE` | ISO country code (US, DE, etc) | auto-detected |
| `CITY` | City name | auto-detected |
| `API_PORT` | API server port | 8080 |
| `P2P_PORT` | P2P network port | 4001 |
| `WG_PORT` | WireGuard port | 51820 |
| `BOOTSTRAP_PEERS` | Comma-separated peer multiaddrs | (empty) |

## Connecting Nodes

To connect your node to the network, you need bootstrap peers. Get the multiaddr from an existing node:

```bash
# On existing node
curl http://localhost:8080/api/v1/p2p/status | jq -r '.multiaddrs[0]'
# Output: /ip4/1.2.3.4/tcp/4001/p2p/12D3KooW...
```

Then start your node with:
```bash
BOOTSTRAP_PEERS="/ip4/1.2.3.4/tcp/4001/p2p/12D3KooW..." docker compose up -d
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/info` | GET | Node information |
| `/api/v1/nodes` | GET | List all known nodes |
| `/api/v1/nodes/best` | GET | Get best available node |
| `/api/v1/nodes/countries` | GET | List countries with nodes |
| `/api/v1/p2p/status` | GET | P2P network status |
| `/api/v1/p2p/peers` | GET | List connected peers |
| `/api/v1/auth/register` | POST | Register new user |
| `/api/v1/auth/login` | POST | Login user |
| `/api/v1/connect` | POST | Connect to VPN (requires auth) |
| `/api/v1/disconnect` | POST | Disconnect from VPN (requires auth) |

## Management Commands

After installation, use the `aureo` command:

```bash
aureo status   # Show node status
aureo logs     # View logs
aureo peers    # List connected peers
aureo nodes    # List known VPN nodes
aureo restart  # Restart node
aureo stop     # Stop node
aureo start    # Start node
```

## Ports

Make sure these ports are open in your firewall:

| Port | Protocol | Purpose |
|------|----------|---------|
| 8080 | TCP | REST API |
| 4001 | TCP/UDP | P2P (libp2p) |
| 51820 | UDP | WireGuard VPN |

## Data Storage

All data is stored in `/var/lib/aureo-vpn` (or `aureo-data` Docker volume):

- `node.db` - SQLite database (users, sessions, node identity)
- `p2p/` - P2P keys and peer cache

## Troubleshooting

### Check if node is running
```bash
docker ps | grep aureo
curl http://localhost:8080/health
```

### View logs
```bash
docker logs -f aureo-node
```

### Check P2P connections
```bash
curl http://localhost:8080/api/v1/p2p/status
```

### Restart the node
```bash
cd /opt/aureo-vpn
docker compose restart
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      aureo-node                              │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  API Server │  │ VPN Service │  │    P2P Network      │  │
│  │   (Fiber)   │  │ (WireGuard) │  │    (libp2p)         │  │
│  └──────┬──────┘  └──────┬──────┘  └──────────┬──────────┘  │
│         │                │                     │             │
│         └────────────────┼─────────────────────┘             │
│                          │                                   │
│                  ┌───────┴───────┐                          │
│                  │    SQLite     │                          │
│                  │   (local DB)  │                          │
│                  └───────────────┘                          │
└─────────────────────────────────────────────────────────────┘
                          │
                          │ P2P (DHT + GossipSub)
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                    Other Nodes                               │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐        │
│  │ Node A  │  │ Node B  │  │ Node C  │  │ Node D  │  ...   │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘        │
└─────────────────────────────────────────────────────────────┘
```

## License

MIT License
