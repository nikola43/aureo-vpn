import React from 'react';
import { motion } from 'framer-motion';
import PageLayout from '../components/PageLayout';
import {
  Monitor,
  Smartphone,
  Globe,
  Terminal,
  Download,
  Apple,
  Chrome,
  Check,
  ArrowRight
} from 'lucide-react';

const platforms = [
  {
    name: 'macOS',
    icon: Apple,
    version: 'v2.0.0',
    size: '45 MB',
    requirements: 'macOS 10.15+',
    available: true,
    link: '#'
  },
  {
    name: 'Windows',
    icon: Monitor,
    version: 'v2.0.0',
    size: '38 MB',
    requirements: 'Windows 10/11',
    available: true,
    link: '#'
  },
  {
    name: 'Linux',
    icon: Terminal,
    version: 'v2.0.0',
    size: '32 MB',
    requirements: 'Ubuntu 20.04+, Debian, Fedora, Arch',
    available: true,
    link: '#'
  },
  {
    name: 'iOS',
    icon: Smartphone,
    version: 'Coming Soon',
    size: '-',
    requirements: 'iOS 14+',
    available: false,
    link: '#'
  },
  {
    name: 'Android',
    icon: Smartphone,
    version: 'Coming Soon',
    size: '-',
    requirements: 'Android 8+',
    available: false,
    link: '#'
  },
  {
    name: 'Browser Extension',
    icon: Chrome,
    version: 'Coming Soon',
    size: '-',
    requirements: 'Chrome, Firefox, Edge',
    available: false,
    link: '#'
  },
];

export default function Downloads() {
  return (
    <PageLayout
      title="Downloads"
      description="Download Aureo VPN for your platform and start browsing securely."
    >
      {/* Desktop Apps */}
      <section className="mb-16">
        <h2 className="text-2xl font-bold text-white mb-8">Desktop Applications</h2>
        <div className="grid md:grid-cols-3 gap-6">
          {platforms.filter(p => ['macOS', 'Windows', 'Linux'].includes(p.name)).map((platform, index) => (
            <motion.div
              key={platform.name}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.1 }}
              className="glass rounded-xl p-6 hover:border-gold-500/30 border border-transparent transition-all"
            >
              <platform.icon className="w-12 h-12 text-gold-500 mb-4" />
              <h3 className="text-xl font-bold text-white mb-2">{platform.name}</h3>
              <div className="space-y-2 mb-6">
                <p className="text-gray-400 text-sm">Version: <span className="text-white">{platform.version}</span></p>
                <p className="text-gray-400 text-sm">Size: <span className="text-white">{platform.size}</span></p>
                <p className="text-gray-400 text-sm">{platform.requirements}</p>
              </div>
              {platform.available ? (
                <a
                  href={platform.link}
                  className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-gold-500 text-dark-950 font-bold rounded-lg hover:bg-gold-400 transition-all"
                >
                  <Download className="w-5 h-5" />
                  Download
                </a>
              ) : (
                <button
                  disabled
                  className="w-full px-4 py-3 bg-dark-800 text-gray-500 font-medium rounded-lg cursor-not-allowed"
                >
                  Coming Soon
                </button>
              )}
            </motion.div>
          ))}
        </div>
      </section>

      {/* Mobile Apps */}
      <section className="mb-16">
        <h2 className="text-2xl font-bold text-white mb-8">Mobile Applications</h2>
        <div className="grid md:grid-cols-2 gap-6">
          {platforms.filter(p => ['iOS', 'Android'].includes(p.name)).map((platform, index) => (
            <motion.div
              key={platform.name}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.1 }}
              className="glass rounded-xl p-6 opacity-60"
            >
              <platform.icon className="w-12 h-12 text-gray-500 mb-4" />
              <h3 className="text-xl font-bold text-white mb-2">{platform.name}</h3>
              <p className="text-gray-400 text-sm mb-4">{platform.requirements}</p>
              <div className="flex items-center gap-2 text-gold-500">
                <span className="text-sm font-medium">Coming Q2 2025</span>
              </div>
            </motion.div>
          ))}
        </div>
      </section>

      {/* CLI Tool */}
      <section className="mb-16">
        <h2 className="text-2xl font-bold text-white mb-8">Command Line Tool</h2>
        <div className="glass rounded-xl p-6">
          <div className="flex items-start gap-6">
            <Terminal className="w-12 h-12 text-gold-500 flex-shrink-0" />
            <div className="flex-1">
              <h3 className="text-xl font-bold text-white mb-2">Aureo CLI</h3>
              <p className="text-gray-400 mb-4">
                Full-featured command line interface for power users and automation.
              </p>
              <div className="bg-dark-900 rounded-lg p-4 mb-4 overflow-x-auto">
                <code className="text-sm text-gray-300 font-mono">
                  <span className="text-gray-500"># Install via Go</span><br />
                  go install github.com/nikola43/aureo-vpn/cmd/cli@latest<br /><br />
                  <span className="text-gray-500"># Or download binary</span><br />
                  curl -sSL https://get.aureo-vpn.com | sh
                </code>
              </div>
              <div className="flex flex-wrap gap-2">
                {['connect', 'disconnect', 'status', 'list-servers', 'config'].map((cmd) => (
                  <span key={cmd} className="px-2 py-1 bg-dark-800 rounded text-xs text-gold-500 font-mono">
                    aureo {cmd}
                  </span>
                ))}
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Router Support */}
      <section className="mb-16">
        <h2 className="text-2xl font-bold text-white mb-8">Router Support</h2>
        <div className="glass rounded-xl p-6">
          <div className="flex items-start gap-6">
            <Globe className="w-12 h-12 text-gold-500 flex-shrink-0" />
            <div>
              <h3 className="text-xl font-bold text-white mb-2">Protect Your Entire Network</h3>
              <p className="text-gray-400 mb-4">
                Configure Aureo VPN on your router to protect all connected devices automatically.
              </p>
              <div className="flex flex-wrap gap-3">
                {['OpenWRT', 'DD-WRT', 'Tomato', 'pfSense', 'OPNsense'].map((router) => (
                  <span key={router} className="px-3 py-1 bg-dark-800 rounded text-sm text-gray-300">
                    {router}
                  </span>
                ))}
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* System Requirements */}
      <section>
        <h2 className="text-2xl font-bold text-white mb-8">System Requirements</h2>
        <div className="grid md:grid-cols-3 gap-6">
          {[
            {
              platform: 'Windows',
              requirements: ['Windows 10 or later', '64-bit processor', '100 MB disk space', 'Admin rights for install']
            },
            {
              platform: 'macOS',
              requirements: ['macOS 10.15 (Catalina) or later', '64-bit Intel or Apple Silicon', '100 MB disk space', 'System Extension approval']
            },
            {
              platform: 'Linux',
              requirements: ['Ubuntu 20.04+, Debian 11+, Fedora 34+', 'x86_64 or ARM64', 'WireGuard kernel module', 'sudo access']
            },
          ].map((item) => (
            <div key={item.platform} className="glass rounded-xl p-6">
              <h4 className="font-bold text-white mb-4">{item.platform}</h4>
              <ul className="space-y-2">
                {item.requirements.map((req) => (
                  <li key={req} className="flex items-start gap-2 text-sm">
                    <Check className="w-4 h-4 text-gold-500 flex-shrink-0 mt-0.5" />
                    <span className="text-gray-400">{req}</span>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      </section>
    </PageLayout>
  );
}
