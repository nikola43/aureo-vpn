import React from 'react';
import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import PageLayout from '../components/PageLayout';
import {
  Server,
  Coins,
  TrendingUp,
  Shield,
  Zap,
  Globe,
  DollarSign,
  Check,
  ArrowRight,
  Terminal,
  HardDrive,
  Wifi
} from 'lucide-react';

function CodeBlock({ children, title }) {
  return (
    <div className="rounded-xl overflow-hidden border border-white/10">
      {title && (
        <div className="bg-dark-800 px-4 py-2 border-b border-white/10">
          <span className="text-sm text-gray-400">{title}</span>
        </div>
      )}
      <pre className="bg-dark-900 p-4 overflow-x-auto">
        <code className="text-sm text-gray-300 font-mono whitespace-pre">{children}</code>
      </pre>
    </div>
  );
}

const tiers = [
  { name: 'Bronze', rate: '$0.010', uptime: '50%', reputation: '0', bonus: '1.0x', color: 'text-amber-600' },
  { name: 'Silver', rate: '$0.015', uptime: '80%', reputation: '60', bonus: '1.2x', color: 'text-gray-400' },
  { name: 'Gold', rate: '$0.020', uptime: '90%', reputation: '75', bonus: '1.5x', color: 'text-gold-500' },
  { name: 'Platinum', rate: '$0.030', uptime: '95%', reputation: '90', bonus: '2.0x', color: 'text-cyan-400' },
];

const requirements = [
  { icon: HardDrive, label: 'CPU', min: '2 cores', rec: '4+ cores' },
  { icon: Server, label: 'RAM', min: '2 GB', rec: '8 GB' },
  { icon: HardDrive, label: 'Storage', min: '50 GB SSD', rec: '100 GB SSD' },
  { icon: Wifi, label: 'Network', min: '100 Mbps', rec: '1 Gbps' },
];

export default function NodeOperators() {
  return (
    <PageLayout
      title="Node Operators"
      description="Run your own VPN node and earn cryptocurrency rewards for contributing to the decentralized network."
    >
      {/* Hero Stats */}
      <div className="grid md:grid-cols-4 gap-6 mb-16">
        {[
          { label: 'Active Nodes', value: '5,234', icon: Server },
          { label: 'Total Paid Out', value: '$1.2M+', icon: DollarSign },
          { label: 'Avg. Monthly Earnings', value: '$450', icon: TrendingUp },
          { label: 'Countries', value: '60+', icon: Globe },
        ].map((stat, index) => (
          <motion.div
            key={stat.label}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: index * 0.1 }}
            className="glass rounded-xl p-6 text-center"
          >
            <stat.icon className="w-8 h-8 text-gold-500 mx-auto mb-3" />
            <div className="text-3xl font-bold text-white mb-1">{stat.value}</div>
            <div className="text-gray-400 text-sm">{stat.label}</div>
          </motion.div>
        ))}
      </div>

      {/* How It Works */}
      <section className="mb-16">
        <h2 className="text-3xl font-bold text-white mb-8">How It Works</h2>
        <div className="grid md:grid-cols-3 gap-8">
          {[
            {
              step: '1',
              title: 'Register as Operator',
              description: 'Create an account and register with your cryptocurrency wallet address to receive payments.',
              icon: Shield
            },
            {
              step: '2',
              title: 'Deploy Your Node',
              description: 'Set up your server using our Docker image or build from source. Takes about 10 minutes.',
              icon: Server
            },
            {
              step: '3',
              title: 'Earn Rewards',
              description: 'Start earning cryptocurrency based on bandwidth served and your tier level.',
              icon: Coins
            },
          ].map((item, index) => (
            <motion.div
              key={item.step}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.15 }}
              className="relative"
            >
              <div className="glass rounded-xl p-8">
                <div className="w-12 h-12 rounded-full bg-gold-500 flex items-center justify-center text-dark-950 font-bold text-xl mb-6">
                  {item.step}
                </div>
                <h3 className="text-xl font-bold text-white mb-3">{item.title}</h3>
                <p className="text-gray-400">{item.description}</p>
              </div>
              {index < 2 && (
                <div className="hidden md:block absolute top-1/2 -right-4 transform -translate-y-1/2">
                  <ArrowRight className="w-8 h-8 text-gold-500/50" />
                </div>
              )}
            </motion.div>
          ))}
        </div>
      </section>

      {/* Reward Tiers */}
      <section id="rewards" className="mb-16">
        <h2 className="text-3xl font-bold text-white mb-8">Reward Tiers</h2>
        <div className="glass rounded-xl overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="text-left text-gray-400 border-b border-white/10">
                <th className="px-6 py-4">Tier</th>
                <th className="px-6 py-4">Rate/GB</th>
                <th className="px-6 py-4">Min Uptime</th>
                <th className="px-6 py-4">Min Reputation</th>
                <th className="px-6 py-4">Bonus</th>
              </tr>
            </thead>
            <tbody>
              {tiers.map((tier) => (
                <tr key={tier.name} className="border-b border-white/5 last:border-0">
                  <td className={`px-6 py-4 font-bold ${tier.color}`}>{tier.name}</td>
                  <td className="px-6 py-4 text-white">{tier.rate}</td>
                  <td className="px-6 py-4 text-gray-300">{tier.uptime}</td>
                  <td className="px-6 py-4 text-gray-300">{tier.reputation}</td>
                  <td className="px-6 py-4 text-gold-500">{tier.bonus}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <div className="mt-8 glass rounded-xl p-6">
          <h4 className="font-bold text-white mb-4">Earnings Example (Gold Tier)</h4>
          <CodeBlock>
{`Bandwidth: 100 GB
Tier: Gold ($0.02/GB)
Quality Score: 90%
Duration: 5 hours (1.2x bonus)

Base: 100 × $0.02 = $2.00
Quality: × (0.5 + 0.9) = $2.80
Duration Bonus: × 1.2 = $3.36
Tier Bonus: × 1.5 = $5.04

Total Earnings: $5.04`}
          </CodeBlock>
        </div>
      </section>

      {/* Requirements */}
      <section className="mb-16">
        <h2 className="text-3xl font-bold text-white mb-8">Server Requirements</h2>
        <div className="grid md:grid-cols-2 gap-8">
          <div className="glass rounded-xl p-6">
            <h4 className="font-bold text-white mb-4 flex items-center gap-2">
              <span className="w-3 h-3 rounded-full bg-yellow-500"></span>
              Minimum
            </h4>
            <div className="space-y-4">
              {requirements.map((req) => (
                <div key={req.label} className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <req.icon className="w-5 h-5 text-gray-400" />
                    <span className="text-gray-300">{req.label}</span>
                  </div>
                  <span className="text-white">{req.min}</span>
                </div>
              ))}
            </div>
          </div>
          <div className="glass rounded-xl p-6">
            <h4 className="font-bold text-white mb-4 flex items-center gap-2">
              <span className="w-3 h-3 rounded-full bg-green-500"></span>
              Recommended
            </h4>
            <div className="space-y-4">
              {requirements.map((req) => (
                <div key={req.label} className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <req.icon className="w-5 h-5 text-gray-400" />
                    <span className="text-gray-300">{req.label}</span>
                  </div>
                  <span className="text-gold-500 font-medium">{req.rec}</span>
                </div>
              ))}
            </div>
          </div>
        </div>

        <div className="mt-6 glass rounded-xl p-6">
          <h4 className="font-bold text-white mb-3">Required Ports</h4>
          <div className="grid md:grid-cols-3 gap-4">
            <div className="flex items-center justify-between p-3 bg-dark-800 rounded-lg">
              <span className="text-gray-400">WireGuard</span>
              <code className="text-gold-500">51820/UDP</code>
            </div>
            <div className="flex items-center justify-between p-3 bg-dark-800 rounded-lg">
              <span className="text-gray-400">OpenVPN</span>
              <code className="text-gold-500">1194/UDP</code>
            </div>
            <div className="flex items-center justify-between p-3 bg-dark-800 rounded-lg">
              <span className="text-gray-400">P2P Network</span>
              <code className="text-gold-500">4001/TCP</code>
            </div>
          </div>
        </div>
      </section>

      {/* Quick Start */}
      <section id="setup" className="mb-16">
        <h2 className="text-3xl font-bold text-white mb-8">Quick Start</h2>

        <div className="space-y-6">
          <div className="glass rounded-xl p-6">
            <h4 className="font-bold text-white mb-4 flex items-center gap-2">
              <span className="w-6 h-6 rounded-full bg-gold-500 flex items-center justify-center text-dark-950 text-sm font-bold">1</span>
              Register as Operator
            </h4>
            <CodeBlock title="terminal">
{`# Register user account
curl -X POST https://api.aureo-vpn.com/api/v1/auth/register \\
  -H "Content-Type: application/json" \\
  -d '{"email":"you@example.com","username":"operator1","password":"secure123"}'

# Save the access token
export TOKEN="<access_token_from_response>"

# Register as operator with your crypto wallet
curl -X POST https://api.aureo-vpn.com/api/v1/operator/register \\
  -H "Authorization: Bearer $TOKEN" \\
  -H "Content-Type: application/json" \\
  -d '{"wallet_address":"0xYourEthWallet","wallet_currency":"ethereum"}'`}
            </CodeBlock>
          </div>

          <div className="glass rounded-xl p-6">
            <h4 className="font-bold text-white mb-4 flex items-center gap-2">
              <span className="w-6 h-6 rounded-full bg-gold-500 flex items-center justify-center text-dark-950 text-sm font-bold">2</span>
              Create Your Node
            </h4>
            <CodeBlock title="terminal">
{`# After admin verification
curl -X POST https://api.aureo-vpn.com/api/v1/operator/nodes \\
  -H "Authorization: Bearer $TOKEN" \\
  -H "Content-Type: application/json" \\
  -d '{
    "name": "my-node",
    "ip_address": "YOUR_SERVER_IP",
    "country": "US",
    "city": "New York",
    "wireguard_port": 51820,
    "max_connections": 500
  }'

# SAVE the node_id and private_key from response!`}
            </CodeBlock>
          </div>

          <div className="glass rounded-xl p-6">
            <h4 className="font-bold text-white mb-4 flex items-center gap-2">
              <span className="w-6 h-6 rounded-full bg-gold-500 flex items-center justify-center text-dark-950 text-sm font-bold">3</span>
              Deploy with Docker
            </h4>
            <CodeBlock title="terminal">
{`docker run -d --name aureo-node \\
  --cap-add=NET_ADMIN \\
  --cap-add=SYS_MODULE \\
  --sysctl net.ipv4.ip_forward=1 \\
  --device /dev/net/tun:/dev/net/tun \\
  -p 51820:51820/udp \\
  -p 1194:1194/udp \\
  -e NODE_ID="YOUR_NODE_ID" \\
  -e NODE_PRIVATE_KEY="YOUR_PRIVATE_KEY" \\
  -e API_URL="https://api.aureo-vpn.com" \\
  -e API_TOKEN="$TOKEN" \\
  ghcr.io/nikola43/aureo-vpn-node:latest`}
            </CodeBlock>
          </div>

          <div className="glass rounded-xl p-6">
            <h4 className="font-bold text-white mb-4 flex items-center gap-2">
              <span className="w-6 h-6 rounded-full bg-gold-500 flex items-center justify-center text-dark-950 text-sm font-bold">4</span>
              Verify Node Status
            </h4>
            <CodeBlock title="terminal">
{`# Check WireGuard interface
docker exec aureo-node wg show wg0

# Check node appears online
curl -s -X GET "https://api.aureo-vpn.com/api/v1/nodes" \\
  -H "Authorization: Bearer $TOKEN" | \\
  jq '.[] | select(.id == "YOUR_NODE_ID")'`}
            </CodeBlock>
          </div>
        </div>
      </section>

      {/* Maintenance */}
      <section id="maintenance" className="mb-16">
        <h2 className="text-3xl font-bold text-white mb-8">Maintenance</h2>
        <div className="grid md:grid-cols-2 gap-6">
          <div className="glass rounded-xl p-6">
            <h4 className="font-bold text-white mb-3">Update Node</h4>
            <CodeBlock>
{`docker pull ghcr.io/nikola43/aureo-vpn-node:latest
docker stop aureo-node
docker rm aureo-node
# Re-run docker run command`}
            </CodeBlock>
          </div>
          <div className="glass rounded-xl p-6">
            <h4 className="font-bold text-white mb-3">View Logs</h4>
            <CodeBlock>
{`# All logs
docker logs -f aureo-node

# Last 100 lines
docker logs --tail=100 aureo-node

# Filter errors
docker logs aureo-node | grep ERROR`}
            </CodeBlock>
          </div>
        </div>
      </section>

      {/* FAQ */}
      <section className="mb-16">
        <h2 className="text-3xl font-bold text-white mb-8">FAQ</h2>
        <div className="space-y-4">
          {[
            { q: 'How long until I start earning?', a: 'Earnings begin as soon as users connect to your node. This depends on your location and network capacity.' },
            { q: "What's the minimum payout?", a: '$10 USD equivalent in your chosen cryptocurrency.' },
            { q: 'How often are payouts processed?', a: 'You can request a payout anytime once you reach the minimum. Transactions typically complete within 10-30 minutes.' },
            { q: 'Can I run multiple nodes?', a: 'Yes, up to 10 nodes per operator account.' },
            { q: 'What happens if my node goes offline?', a: 'Your uptime percentage decreases, which may affect your tier. Active sessions are gracefully disconnected.' },
          ].map((faq) => (
            <div key={faq.q} className="glass rounded-xl p-6">
              <h4 className="font-bold text-white mb-2">{faq.q}</h4>
              <p className="text-gray-400">{faq.a}</p>
            </div>
          ))}
        </div>
      </section>

      {/* CTA */}
      <div className="glass rounded-2xl p-8 text-center bg-gradient-to-r from-gold-500/10 to-transparent">
        <h3 className="text-2xl font-bold text-white mb-4">Ready to start earning?</h3>
        <p className="text-gray-400 mb-6 max-w-xl mx-auto">
          Join thousands of node operators earning passive income by contributing to the decentralized VPN network.
        </p>
        <div className="flex justify-center gap-4">
          <a
            href="/register"
            className="px-8 py-3 bg-gold-500 text-dark-950 font-bold rounded-lg hover:bg-gold-400 transition-all"
          >
            Register as Operator
          </a>
          <Link
            to="/docs"
            className="px-8 py-3 glass rounded-lg font-medium hover:bg-white/10 transition-all"
          >
            Read Full Guide
          </Link>
        </div>
      </div>
    </PageLayout>
  );
}
