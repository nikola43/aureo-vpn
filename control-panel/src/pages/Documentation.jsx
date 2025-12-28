import { useState } from 'react';
import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import PageLayout from '../components/PageLayout';
import {
  Book,
  Code2,
  Server,
  Shield,
  Coins,
  Zap,
  Globe,
  Lock,
  ArrowRight,
  Terminal,
  FileCode,
  Database,
  Copy,
  Check
} from 'lucide-react';

const sections = [
  {
    title: 'Getting Started',
    icon: Book,
    description: 'Quick start guide to get you up and running with Aureo VPN.',
    links: [
      { label: 'Installation', href: '#installation' },
      { label: 'Quick Start', href: '#quick-start' },
      { label: 'Configuration', href: '#configuration' },
    ]
  },
  {
    title: 'Architecture',
    icon: Database,
    description: 'Understand the decentralized architecture of Aureo VPN.',
    links: [
      { label: 'System Overview', href: '#overview' },
      { label: 'Components', href: '#architecture' },
      { label: 'P2P Network', href: '#p2p' },
    ]
  },
  {
    title: 'API Reference',
    icon: Code2,
    description: 'Complete API documentation for developers.',
    links: [
      { label: 'Authentication', href: '/api#auth' },
      { label: 'Nodes API', href: '/api#nodes' },
      { label: 'Sessions API', href: '/api#sessions' },
    ]
  },
  {
    title: 'Node Operators',
    icon: Server,
    description: 'Run your own VPN node and earn cryptocurrency rewards.',
    links: [
      { label: 'Setup Guide', href: '/node-operators#setup' },
      { label: 'Reward Tiers', href: '/node-operators#rewards' },
      { label: 'Maintenance', href: '/node-operators#maintenance' },
    ]
  },
];

function CodeBlock({ children, title }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(children);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="rounded-xl overflow-hidden border border-white/10">
      <div className="bg-dark-800 px-4 py-2 border-b border-white/10 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Terminal className="w-4 h-4 text-gold-500" />
          <span className="text-sm text-gray-400">{title || 'terminal'}</span>
        </div>
        <button
          onClick={handleCopy}
          className="flex items-center gap-1.5 text-xs text-gray-400 hover:text-white transition-colors px-2 py-1 rounded hover:bg-white/5"
          title="Copy to clipboard"
        >
          {copied ? (
            <>
              <Check className="w-3.5 h-3.5 text-green-500" />
              <span className="text-green-500">Copied!</span>
            </>
          ) : (
            <>
              <Copy className="w-3.5 h-3.5" />
              <span>Copy</span>
            </>
          )}
        </button>
      </div>
      <pre className="bg-dark-900 p-4 overflow-x-auto">
        <code className="text-sm text-gray-300 font-mono">{children}</code>
      </pre>
    </div>
  );
}

export default function Documentation() {
  return (
    <PageLayout
      title="Documentation"
      description="Everything you need to know about Aureo VPN - from installation to advanced configuration."
    >
      {/* Quick Links */}
      <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6 mb-16">
        {sections.map((section, index) => (
          <motion.div
            key={section.title}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: index * 0.1 }}
            className="glass rounded-xl p-6 hover:border-gold-500/30 border border-transparent transition-all group"
          >
            <div className="w-12 h-12 rounded-xl bg-gold-500/10 flex items-center justify-center mb-4 group-hover:bg-gold-500/20 transition-all">
              <section.icon className="w-6 h-6 text-gold-500" />
            </div>
            <h3 className="text-lg font-bold text-white mb-2">{section.title}</h3>
            <p className="text-gray-400 text-sm mb-4">{section.description}</p>
            <ul className="space-y-2">
              {section.links.map((link) => (
                <li key={link.href}>
                  {link.href.startsWith('#') ? (
                    <a
                      href={link.href}
                      className="text-sm text-gold-500 hover:text-gold-400 flex items-center gap-1"
                    >
                      <ArrowRight className="w-3 h-3" />
                      {link.label}
                    </a>
                  ) : (
                    <Link
                      to={link.href}
                      className="text-sm text-gold-500 hover:text-gold-400 flex items-center gap-1"
                    >
                      <ArrowRight className="w-3 h-3" />
                      {link.label}
                    </Link>
                  )}
                </li>
              ))}
            </ul>
          </motion.div>
        ))}
      </div>

      {/* Overview Section */}
      <section id="overview" className="mb-16">
        <h2 className="text-3xl font-bold text-white mb-6">Overview</h2>
        <div className="prose prose-invert max-w-none">
          <p className="text-gray-300 text-lg leading-relaxed mb-6">
            Aureo VPN is a decentralized VPN service that allows anyone to run VPN nodes and earn cryptocurrency rewards.
            The platform consists of several key components working together to provide secure, private internet access.
          </p>

          <div className="grid md:grid-cols-2 gap-6 mb-8">
            <div className="glass rounded-xl p-6">
              <h4 className="text-lg font-bold text-white mb-3 flex items-center gap-2">
                <Shield className="w-5 h-5 text-gold-500" />
                Security
              </h4>
              <ul className="space-y-2 text-gray-400">
                <li>WireGuard & OpenVPN protocol support</li>
                <li>ChaCha20-Poly1305 / AES-256-GCM encryption</li>
                <li>Kill switch & DNS leak protection</li>
                <li>Multi-hop routing for enhanced privacy</li>
              </ul>
            </div>
            <div className="glass rounded-xl p-6">
              <h4 className="text-lg font-bold text-white mb-3 flex items-center gap-2">
                <Coins className="w-5 h-5 text-gold-500" />
                Crypto Rewards
              </h4>
              <ul className="space-y-2 text-gray-400">
                <li>Earn ETH, BTC, or LTC for running nodes</li>
                <li>Reputation-based tier system (Bronze to Platinum)</li>
                <li>Quality multipliers based on performance</li>
                <li>Minimum payout: $10 USD equivalent</li>
              </ul>
            </div>
          </div>
        </div>
      </section>

      {/* Installation Section */}
      <section id="installation" className="mb-16">
        <h2 className="text-3xl font-bold text-white mb-6">Installation</h2>

        <h3 className="text-xl font-bold text-white mb-4">Using Docker (Recommended)</h3>
        <CodeBlock title="terminal">
{`# Clone the repository
git clone https://github.com/nikola43/aureo-vpn.git
cd aureo-vpn

# Start all services
cd deployments/docker
docker-compose up -d

# Check status
docker-compose ps`}
        </CodeBlock>

        <div className="mt-6 glass rounded-xl p-6">
          <h4 className="font-bold text-white mb-3">Services Available:</h4>
          <div className="grid md:grid-cols-2 gap-4">
            <div>
              <code className="text-gold-500">API Gateway:</code>
              <span className="text-gray-400 ml-2">http://localhost:8080</span>
            </div>
            <div>
              <code className="text-gold-500">Prometheus:</code>
              <span className="text-gray-400 ml-2">http://localhost:9090</span>
            </div>
            <div>
              <code className="text-gold-500">Grafana:</code>
              <span className="text-gray-400 ml-2">http://localhost:3000</span>
            </div>
            <div>
              <code className="text-gold-500">Dashboard:</code>
              <span className="text-gray-400 ml-2">http://localhost:5000</span>
            </div>
          </div>
        </div>

        <h3 className="text-xl font-bold text-white mt-8 mb-4">Manual Setup</h3>
        <CodeBlock title="terminal">
{`# Prerequisites: Go 1.24+, Redis 7+

# Build all services
make build

# Build individual components
make build-api       # API Gateway
make build-control   # Control Server
make build-node      # VPN Node
make build-cli       # CLI Tool

# Run API Gateway
./bin/api-gateway

# Run VPN Node (requires root)
sudo ./bin/vpn-node`}
        </CodeBlock>
      </section>

      {/* Quick Start Section */}
      <section id="quick-start" className="mb-16">
        <h2 className="text-3xl font-bold text-white mb-6">Quick Start</h2>

        <div className="space-y-6">
          <div className="glass rounded-xl p-6">
            <h4 className="font-bold text-white mb-3 flex items-center gap-2">
              <span className="w-6 h-6 rounded-full bg-gold-500 flex items-center justify-center text-dark-950 text-sm font-bold">1</span>
              Register an Account
            </h4>
            <CodeBlock>
{`curl -X POST https://api.aureo-vpn.com/api/v1/auth/register \\
  -H "Content-Type: application/json" \\
  -d '{
    "email": "user@example.com",
    "username": "johndoe",
    "password": "SecurePassword123"
  }'`}
            </CodeBlock>
          </div>

          <div className="glass rounded-xl p-6">
            <h4 className="font-bold text-white mb-3 flex items-center gap-2">
              <span className="w-6 h-6 rounded-full bg-gold-500 flex items-center justify-center text-dark-950 text-sm font-bold">2</span>
              Get Available Nodes
            </h4>
            <CodeBlock>
{`curl -X GET "https://api.aureo-vpn.com/api/v1/nodes?country=US" \\
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"`}
            </CodeBlock>
          </div>

          <div className="glass rounded-xl p-6">
            <h4 className="font-bold text-white mb-3 flex items-center gap-2">
              <span className="w-6 h-6 rounded-full bg-gold-500 flex items-center justify-center text-dark-950 text-sm font-bold">3</span>
              Create VPN Session
            </h4>
            <CodeBlock>
{`curl -X POST https://api.aureo-vpn.com/api/v1/sessions \\
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \\
  -H "Content-Type: application/json" \\
  -d '{
    "node_id": "NODE_UUID",
    "protocol": "wireguard",
    "enable_kill_switch": true
  }'`}
            </CodeBlock>
          </div>
        </div>
      </section>

      {/* Configuration Section */}
      <section id="configuration" className="mb-16">
        <h2 className="text-3xl font-bold text-white mb-6">Configuration</h2>

        <p className="text-gray-400 mb-6">
          Create a <code className="text-gold-500 bg-gold-500/10 px-2 py-0.5 rounded">.env</code> file with the following configuration:
        </p>

        <CodeBlock title=".env">
{`# Database (SQLite)
DB_PATH=./data/aureo.db

# JWT Authentication
JWT_SECRET=your-32-byte-secret-key-here
JWT_ACCESS_DURATION=15m
JWT_REFRESH_DURATION=168h

# Server Configuration
HOST=0.0.0.0
PORT=8080

# Redis (Optional)
REDIS_HOST=localhost
REDIS_PORT=6379

# VPN Node Configuration
NODE_ID=
NODE_NAME=us-east-1
NODE_COUNTRY=US
DEFAULT_PROTOCOL=wireguard
WIREGUARD_PORT=51820

# Blockchain (for rewards)
ETHEREUM_RPC_URL=https://mainnet.infura.io/v3/YOUR_KEY
ETHEREUM_PRIVATE_KEY=0x...`}
        </CodeBlock>

        <div className="mt-8 glass rounded-xl p-6">
          <h4 className="font-bold text-white mb-4">Reward Tiers Configuration</h4>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-gray-400 border-b border-white/10">
                  <th className="pb-3">Tier</th>
                  <th className="pb-3">Rate/GB</th>
                  <th className="pb-3">Min Uptime</th>
                  <th className="pb-3">Min Reputation</th>
                  <th className="pb-3">Bonus</th>
                </tr>
              </thead>
              <tbody className="text-gray-300">
                <tr className="border-b border-white/5">
                  <td className="py-3 text-amber-600">Bronze</td>
                  <td className="py-3">$0.010</td>
                  <td className="py-3">50%</td>
                  <td className="py-3">0</td>
                  <td className="py-3">1.0x</td>
                </tr>
                <tr className="border-b border-white/5">
                  <td className="py-3 text-gray-400">Silver</td>
                  <td className="py-3">$0.015</td>
                  <td className="py-3">80%</td>
                  <td className="py-3">60</td>
                  <td className="py-3">1.2x</td>
                </tr>
                <tr className="border-b border-white/5">
                  <td className="py-3 text-gold-500">Gold</td>
                  <td className="py-3">$0.020</td>
                  <td className="py-3">90%</td>
                  <td className="py-3">75</td>
                  <td className="py-3">1.5x</td>
                </tr>
                <tr>
                  <td className="py-3 text-cyan-400">Platinum</td>
                  <td className="py-3">$0.025</td>
                  <td className="py-3">95%</td>
                  <td className="py-3">90</td>
                  <td className="py-3">2.0x</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>

      {/* Architecture Section */}
      <section id="architecture" className="mb-16">
        <h2 className="text-3xl font-bold text-white mb-6">Architecture</h2>

        <div className="glass rounded-xl p-6 mb-8">
          <pre className="text-xs text-gray-400 overflow-x-auto font-mono">
{`┌─────────────────────────────────────────────────────────────────────────────┐
│                              AUREO VPN PLATFORM                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐                    │
│  │   Desktop   │     │   Mobile    │     │    Web      │                    │
│  │   Client    │     │   Client    │     │  Dashboard  │                    │
│  └──────┬──────┘     └──────┬──────┘     └──────┬──────┘                    │
│         └───────────────────┼───────────────────┘                           │
│                             ▼                                               │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                         API GATEWAY                                   │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐              │   │
│  │  │   Auth   │  │  Nodes   │  │ Sessions │  │ Operator │              │   │
│  │  │ Handler  │  │ Handler  │  │ Handler  │  │ Handler  │              │   │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘              │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                             │                                               │
│              ┌──────────────┼──────────────┐                                │
│              ▼              ▼              ▼                                │
│  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐                   │
│  │    SQLite     │  │    Redis      │  │ Control Server│                   │
│  │   Database    │  │    Cache      │  │               │                   │
│  └───────────────┘  └───────────────┘  └───────────────┘                   │
│                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                       P2P NETWORK (libp2p)                           │   │
│  │  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐   │   │
│  │  │   Kademlia DHT  │    │    Gossipsub    │    │      mDNS       │   │   │
│  │  │  (Peer Routing) │    │   (Pub/Sub)     │    │  (Local Disc.)  │   │   │
│  │  └─────────────────┘    └─────────────────┘    └─────────────────┘   │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                         VPN NODES                                     │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │   │
│  │  │  Node US-1  │  │  Node EU-1  │  │  Node AP-1  │  │  Node ...   │  │   │
│  │  │  WireGuard  │  │  OpenVPN    │  │  WireGuard  │  │             │  │   │
│  │  │  51820/UDP  │  │  1194/UDP   │  │  51820/UDP  │  │             │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘  │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                      BLOCKCHAIN LAYER                                 │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                   │   │
│  │  │  Ethereum   │  │   Bitcoin   │  │  Litecoin   │                   │   │
│  │  │   Payouts   │  │   Payouts   │  │   Payouts   │                   │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘                   │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘`}
          </pre>
        </div>

        <div className="grid md:grid-cols-3 gap-6">
          <div className="glass rounded-xl p-6">
            <h4 className="font-bold text-white mb-2">API Gateway</h4>
            <p className="text-gray-400 text-sm">REST API server (Fiber) handling authentication, node discovery, and session management.</p>
            <code className="text-gold-500 text-sm">Port: 8080</code>
          </div>
          <div className="glass rounded-xl p-6">
            <h4 className="font-bold text-white mb-2">VPN Node</h4>
            <p className="text-gray-400 text-sm">WireGuard/OpenVPN server providing encrypted tunnels to clients.</p>
            <code className="text-gold-500 text-sm">Port: 51820/UDP</code>
          </div>
          <div className="glass rounded-xl p-6">
            <h4 className="font-bold text-white mb-2">P2P Network</h4>
            <p className="text-gray-400 text-sm">libp2p-based decentralized node discovery using Kademlia DHT.</p>
            <code className="text-gold-500 text-sm">Port: 4001/TCP</code>
          </div>
        </div>
      </section>

      {/* P2P Network Section */}
      <section id="p2p" className="mb-16">
        <h2 className="text-3xl font-bold text-white mb-6">P2P Network</h2>

        <p className="text-gray-300 text-lg leading-relaxed mb-8">
          Aureo VPN uses libp2p for decentralized node discovery. Nodes announce themselves to the
          network and can be discovered without relying on a central server.
        </p>

        <div className="grid md:grid-cols-2 gap-6 mb-8">
          <div className="glass rounded-xl p-6">
            <h4 className="text-lg font-bold text-white mb-3 flex items-center gap-2">
              <Globe className="w-5 h-5 text-gold-500" />
              Kademlia DHT
            </h4>
            <p className="text-gray-400 mb-4">
              Distributed hash table for peer routing. Nodes find each other through a structured
              overlay network with O(log n) lookup complexity.
            </p>
            <ul className="space-y-2 text-gray-400 text-sm">
              <li>• XOR-based distance metric</li>
              <li>• Self-healing network topology</li>
              <li>• No central point of failure</li>
            </ul>
          </div>
          <div className="glass rounded-xl p-6">
            <h4 className="text-lg font-bold text-white mb-3 flex items-center gap-2">
              <Zap className="w-5 h-5 text-gold-500" />
              Gossipsub
            </h4>
            <p className="text-gray-400 mb-4">
              Pub/sub messaging protocol for node announcements and heartbeats.
              Efficient flood-fill with mesh topology.
            </p>
            <ul className="space-y-2 text-gray-400 text-sm">
              <li>• Topic-based subscriptions</li>
              <li>• Heartbeat every 30 seconds</li>
              <li>• Real-time status updates</li>
            </ul>
          </div>
        </div>

        <div className="glass rounded-xl p-6">
          <h4 className="font-bold text-white mb-4">P2P Configuration</h4>
          <CodeBlock title=".env">
{`# P2P Network Settings
P2P_ENABLED=true
P2P_PORT=4001
P2P_ENABLE_DHT=true
P2P_ENABLE_MDNS=true
P2P_DHT_SERVER=true
P2P_BOOTSTRAP_PEERS=/ip4/1.2.3.4/tcp/4001/p2p/QmBootstrap...
P2P_DATA_DIR=/var/lib/aureo/p2p
P2P_MAX_PEERS=100`}
          </CodeBlock>
        </div>
      </section>

      {/* CTA */}
      <div className="mt-16 glass rounded-2xl p-8 text-center">
        <h3 className="text-2xl font-bold text-white mb-4">Ready to get started?</h3>
        <p className="text-gray-400 mb-6">Join thousands of users enjoying secure, private browsing with Aureo VPN.</p>
        <div className="flex justify-center gap-4">
          <Link
            to="/downloads"
            className="px-6 py-3 bg-gold-500 text-dark-950 font-bold rounded-lg hover:bg-gold-400 transition-all"
          >
            Download Now
          </Link>
          <Link
            to="/node-operators"
            className="px-6 py-3 glass rounded-lg font-medium hover:bg-white/10 transition-all"
          >
            Become an Operator
          </Link>
        </div>
      </div>
    </PageLayout>
  );
}
