import React from 'react';
import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import PageLayout from '../components/PageLayout';
import {
  Shield,
  Lock,
  Globe,
  Zap,
  Eye,
  Wifi,
  Server,
  Layers,
  Key,
  Fingerprint,
  Split,
  RefreshCw,
  Coins,
  Award,
  Check,
  Gauge
} from 'lucide-react';

const securityFeatures = [
  {
    icon: Shield,
    title: 'WireGuard Protocol',
    description: 'Modern, fast, and secure VPN protocol with ChaCha20-Poly1305 encryption. Built into Linux kernel for maximum performance.',
    badges: ['ChaCha20', 'Poly1305', 'Curve25519']
  },
  {
    icon: Lock,
    title: 'OpenVPN Protocol',
    description: 'Industry-standard VPN protocol with AES-256-GCM encryption. Wide compatibility across all platforms.',
    badges: ['AES-256-GCM', 'RSA-2048', 'SHA256']
  },
  {
    icon: Eye,
    title: 'Kill Switch',
    description: 'System-wide kill switch blocks all internet traffic if VPN disconnects, preventing any IP leaks.',
    badges: ['Auto-activate', 'App-specific']
  },
  {
    icon: Fingerprint,
    title: 'DNS Leak Protection',
    description: 'Routes all DNS queries through encrypted tunnel with custom DNS servers (1.1.1.1, 1.0.0.1).',
    badges: ['DNSCrypt', 'DoH Support']
  },
  {
    icon: Layers,
    title: 'Multi-Hop VPN',
    description: 'Chain through multiple nodes in different countries for enhanced privacy. Triple VPN support available.',
    badges: ['Double VPN', 'Triple VPN']
  },
  {
    icon: Split,
    title: 'Split Tunneling',
    description: 'Route specific apps or domains through VPN while excluding others. Fine-grained control over your traffic.',
    badges: ['App-based', 'IP-based']
  },
];

const platformFeatures = [
  {
    icon: Globe,
    title: 'Global Network',
    description: '60+ countries, 5000+ servers worldwide',
    stats: '99.9% uptime'
  },
  {
    icon: Coins,
    title: 'Crypto Rewards',
    description: 'Earn ETH, BTC, LTC for running nodes',
    stats: '$1.2M+ paid'
  },
  {
    icon: Award,
    title: 'Reputation System',
    description: 'Quality-based rewards with tier progression',
    stats: '4 tiers'
  },
  {
    icon: Gauge,
    title: 'Smart Load Balancing',
    description: 'Auto-select best server based on load & latency',
    stats: '<15ms avg'
  },
];

export default function Features() {
  return (
    <PageLayout
      title="Features"
      description="Discover all the powerful features that make Aureo VPN the most secure and private VPN solution."
    >
      {/* Security Features */}
      <section className="mb-20">
        <div className="text-center mb-12">
          <span className="text-gold-500 font-semibold tracking-wider uppercase text-sm">Security</span>
          <h2 className="text-3xl md:text-4xl font-bold text-white mt-2">Military-Grade Security</h2>
          <p className="text-gray-400 mt-4 max-w-2xl mx-auto">
            Advanced encryption protocols and security features to keep your data safe.
          </p>
        </div>

        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
          {securityFeatures.map((feature, index) => (
            <motion.div
              key={feature.title}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.1 }}
              className="glass rounded-xl p-6 hover:border-gold-500/30 border border-transparent transition-all"
            >
              <div className="w-12 h-12 rounded-xl bg-gold-500/10 flex items-center justify-center mb-4">
                <feature.icon className="w-6 h-6 text-gold-500" />
              </div>
              <h3 className="text-lg font-bold text-white mb-2">{feature.title}</h3>
              <p className="text-gray-400 text-sm mb-4">{feature.description}</p>
              <div className="flex flex-wrap gap-2">
                {feature.badges.map((badge) => (
                  <span key={badge} className="px-2 py-1 bg-dark-800 rounded text-xs text-gray-400">
                    {badge}
                  </span>
                ))}
              </div>
            </motion.div>
          ))}
        </div>
      </section>

      {/* Platform Features */}
      <section className="mb-20">
        <div className="text-center mb-12">
          <span className="text-gold-500 font-semibold tracking-wider uppercase text-sm">Platform</span>
          <h2 className="text-3xl md:text-4xl font-bold text-white mt-2">Decentralized Network</h2>
        </div>

        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
          {platformFeatures.map((feature, index) => (
            <motion.div
              key={feature.title}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.1 }}
              className="glass rounded-xl p-6 text-center"
            >
              <feature.icon className="w-10 h-10 text-gold-500 mx-auto mb-4" />
              <h3 className="font-bold text-white mb-2">{feature.title}</h3>
              <p className="text-gray-400 text-sm mb-3">{feature.description}</p>
              <div className="text-gold-500 font-bold">{feature.stats}</div>
            </motion.div>
          ))}
        </div>
      </section>

      {/* Protocol Comparison */}
      <section className="mb-20">
        <div className="text-center mb-12">
          <span className="text-gold-500 font-semibold tracking-wider uppercase text-sm">Protocols</span>
          <h2 className="text-3xl md:text-4xl font-bold text-white mt-2">Protocol Comparison</h2>
        </div>

        <div className="glass rounded-xl overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="text-left text-gray-400 border-b border-white/10">
                <th className="px-6 py-4">Feature</th>
                <th className="px-6 py-4">WireGuard</th>
                <th className="px-6 py-4">OpenVPN</th>
                <th className="px-6 py-4">IKEv2/IPsec</th>
              </tr>
            </thead>
            <tbody className="text-gray-300">
              <tr className="border-b border-white/5">
                <td className="px-6 py-4 font-medium">Speed</td>
                <td className="px-6 py-4 text-green-400">8-10 Gbps</td>
                <td className="px-6 py-4 text-yellow-400">1-2 Gbps</td>
                <td className="px-6 py-4 text-yellow-400">3-5 Gbps</td>
              </tr>
              <tr className="border-b border-white/5">
                <td className="px-6 py-4 font-medium">Encryption</td>
                <td className="px-6 py-4">ChaCha20-Poly1305</td>
                <td className="px-6 py-4">AES-256-GCM</td>
                <td className="px-6 py-4">AES-256-GCM</td>
              </tr>
              <tr className="border-b border-white/5">
                <td className="px-6 py-4 font-medium">Latency</td>
                <td className="px-6 py-4 text-green-400">+5-15ms</td>
                <td className="px-6 py-4 text-yellow-400">+20-40ms</td>
                <td className="px-6 py-4 text-yellow-400">+10-25ms</td>
              </tr>
              <tr className="border-b border-white/5">
                <td className="px-6 py-4 font-medium">Code Lines</td>
                <td className="px-6 py-4 text-green-400">~4,000</td>
                <td className="px-6 py-4 text-yellow-400">~400,000</td>
                <td className="px-6 py-4 text-yellow-400">~100,000</td>
              </tr>
              <tr>
                <td className="px-6 py-4 font-medium">Best For</td>
                <td className="px-6 py-4">Speed & Security</td>
                <td className="px-6 py-4">Compatibility</td>
                <td className="px-6 py-4">Mobile</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      {/* Privacy Features */}
      <section className="mb-20">
        <div className="text-center mb-12">
          <span className="text-gold-500 font-semibold tracking-wider uppercase text-sm">Privacy</span>
          <h2 className="text-3xl md:text-4xl font-bold text-white mt-2">True Zero-Log Policy</h2>
        </div>

        <div className="grid md:grid-cols-2 gap-8">
          <div className="glass rounded-xl p-8">
            <h3 className="text-xl font-bold text-white mb-6">What We DON'T Log</h3>
            <ul className="space-y-4">
              {[
                'Connection timestamps',
                'IP addresses',
                'Browsing activity',
                'DNS queries',
                'Bandwidth usage per session',
                'Session duration'
              ].map((item) => (
                <li key={item} className="flex items-center gap-3">
                  <div className="w-6 h-6 rounded-full bg-green-500/20 flex items-center justify-center flex-shrink-0">
                    <Check className="w-4 h-4 text-green-500" />
                  </div>
                  <span className="text-gray-300">{item}</span>
                </li>
              ))}
            </ul>
          </div>
          <div className="glass rounded-xl p-8">
            <h3 className="text-xl font-bold text-white mb-6">Our Commitments</h3>
            <ul className="space-y-4">
              {[
                'RAM-only servers (no persistent storage)',
                'Quarterly transparency reports',
                'Cryptographically signed warrant canary',
                'Independent security audits',
                'Open source codebase',
                'GDPR & CCPA compliant'
              ].map((item) => (
                <li key={item} className="flex items-center gap-3">
                  <div className="w-6 h-6 rounded-full bg-gold-500/20 flex items-center justify-center flex-shrink-0">
                    <Shield className="w-4 h-4 text-gold-500" />
                  </div>
                  <span className="text-gray-300">{item}</span>
                </li>
              ))}
            </ul>
          </div>
        </div>
      </section>

      {/* CTA */}
      <div className="glass rounded-2xl p-8 text-center bg-gradient-to-r from-gold-500/10 to-transparent">
        <h3 className="text-2xl font-bold text-white mb-4">Experience True Privacy</h3>
        <p className="text-gray-400 mb-6 max-w-xl mx-auto">
          Join thousands of users enjoying secure, private browsing with Aureo VPN.
        </p>
        <div className="flex justify-center gap-4">
          <Link
            to="/downloads"
            className="px-8 py-3 bg-gold-500 text-dark-950 font-bold rounded-lg hover:bg-gold-400 transition-all"
          >
            Download Now
          </Link>
          <Link
            to="/#pricing"
            className="px-8 py-3 glass rounded-lg font-medium hover:bg-white/10 transition-all"
          >
            View Pricing
          </Link>
        </div>
      </div>
    </PageLayout>
  );
}
