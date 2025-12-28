import React from 'react';
import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import PageLayout from '../components/PageLayout';
import {
  Shield,
  Users,
  Globe,
  Award,
  Heart,
  Code2,
  Lock,
  Coins,
  Target,
  Lightbulb
} from 'lucide-react';

const values = [
  {
    icon: Lock,
    title: 'Privacy First',
    description: 'We believe privacy is a fundamental human right. Every decision we make puts your privacy first.'
  },
  {
    icon: Code2,
    title: 'Open Source',
    description: 'Our code is open for anyone to inspect. Transparency builds trust, and trust is earned.'
  },
  {
    icon: Users,
    title: 'Community Driven',
    description: 'Powered by thousands of node operators worldwide. A truly decentralized network.'
  },
  {
    icon: Coins,
    title: 'Fair Rewards',
    description: 'Node operators earn cryptocurrency for contributing bandwidth. Everyone benefits.'
  },
];

const stats = [
  { value: '5,000+', label: 'Active Nodes' },
  { value: '60+', label: 'Countries' },
  { value: '100K+', label: 'Users' },
  { value: '$1.2M+', label: 'Paid to Operators' },
];

export default function About() {
  return (
    <PageLayout
      title="About Aureo"
      description="Building the future of decentralized privacy - secure browsing and crypto rewards for the privacy-conscious."
    >
      {/* Mission */}
      <section className="mb-20">
        <div className="glass rounded-2xl p-8 md:p-12 bg-gradient-to-r from-gold-500/5 to-transparent">
          <div className="flex items-start gap-6">
            <div className="w-16 h-16 rounded-xl bg-gold-500/10 flex items-center justify-center flex-shrink-0">
              <Target className="w-8 h-8 text-gold-500" />
            </div>
            <div>
              <h2 className="text-2xl font-bold text-white mb-4">Our Mission</h2>
              <p className="text-gray-300 text-lg leading-relaxed">
                Aureo VPN was founded with a simple mission: to make online privacy accessible to everyone while
                creating a sustainable ecosystem where contributors are fairly rewarded. We believe that the future
                of VPN technology is decentralized, community-driven, and transparent.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Stats */}
      <section className="mb-20">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
          {stats.map((stat, index) => (
            <motion.div
              key={stat.label}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.1 }}
              className="glass rounded-xl p-6 text-center"
            >
              <div className="text-3xl md:text-4xl font-bold text-gold-500 mb-2">{stat.value}</div>
              <div className="text-gray-400">{stat.label}</div>
            </motion.div>
          ))}
        </div>
      </section>

      {/* Values */}
      <section className="mb-20">
        <div className="text-center mb-12">
          <h2 className="text-3xl font-bold text-white">Our Values</h2>
          <p className="text-gray-400 mt-4 max-w-xl mx-auto">
            The principles that guide everything we do.
          </p>
        </div>
        <div className="grid md:grid-cols-2 gap-6">
          {values.map((value, index) => (
            <motion.div
              key={value.title}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.1 }}
              className="glass rounded-xl p-6 hover:border-gold-500/30 border border-transparent transition-all"
            >
              <div className="flex items-start gap-4">
                <div className="w-12 h-12 rounded-xl bg-gold-500/10 flex items-center justify-center flex-shrink-0">
                  <value.icon className="w-6 h-6 text-gold-500" />
                </div>
                <div>
                  <h3 className="text-lg font-bold text-white mb-2">{value.title}</h3>
                  <p className="text-gray-400">{value.description}</p>
                </div>
              </div>
            </motion.div>
          ))}
        </div>
      </section>

      {/* Story */}
      <section className="mb-20">
        <div className="grid md:grid-cols-2 gap-12 items-center">
          <div>
            <h2 className="text-3xl font-bold text-white mb-6">Our Story</h2>
            <div className="space-y-4 text-gray-400">
              <p>
                Aureo VPN started in 2024 as a response to the growing centralization of VPN services.
                We saw a world where a few large companies controlled most of the VPN infrastructure,
                creating single points of failure and potential privacy risks.
              </p>
              <p>
                Our founders, a team of privacy advocates and blockchain developers, envisioned a
                different approach: a VPN network owned and operated by its community, where anyone
                could contribute bandwidth and earn rewards.
              </p>
              <p>
                The name "Aureo" comes from Latin, meaning "gold" - representing the value we place
                on both user privacy and fair compensation for network contributors.
              </p>
            </div>
          </div>
          <div className="glass rounded-2xl p-8">
            <div className="flex items-center gap-4 mb-6">
              <Lightbulb className="w-8 h-8 text-gold-500" />
              <h3 className="text-xl font-bold text-white">Why Decentralized?</h3>
            </div>
            <ul className="space-y-4">
              {[
                'No single point of failure',
                'Censorship resistant',
                'Community-owned infrastructure',
                'Fair reward distribution',
                'True privacy - no central logs'
              ].map((item) => (
                <li key={item} className="flex items-center gap-3">
                  <div className="w-2 h-2 rounded-full bg-gold-500"></div>
                  <span className="text-gray-300">{item}</span>
                </li>
              ))}
            </ul>
          </div>
        </div>
      </section>

      {/* Open Source */}
      <section className="mb-20">
        <div className="glass rounded-2xl p-8 md:p-12">
          <div className="text-center mb-8">
            <Code2 className="w-12 h-12 text-gold-500 mx-auto mb-4" />
            <h2 className="text-3xl font-bold text-white mb-4">Open Source</h2>
            <p className="text-gray-400 max-w-2xl mx-auto">
              We believe in transparency. Our entire codebase is open source, allowing anyone to inspect,
              audit, and contribute to the project.
            </p>
          </div>
          <div className="flex justify-center">
            <a
              href="https://github.com/nikola43/aureo-vpn"
              target="_blank"
              rel="noopener noreferrer"
              className="px-8 py-3 bg-dark-800 hover:bg-dark-700 rounded-lg font-medium transition-all flex items-center gap-3"
            >
              <svg className="w-6 h-6" fill="currentColor" viewBox="0 0 24 24">
                <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
              </svg>
              View on GitHub
            </a>
          </div>
        </div>
      </section>

      {/* CTA */}
      <div className="glass rounded-2xl p-8 text-center bg-gradient-to-r from-gold-500/10 to-transparent">
        <h3 className="text-2xl font-bold text-white mb-4">Join Our Community</h3>
        <p className="text-gray-400 mb-6 max-w-xl mx-auto">
          Whether you want to protect your privacy or earn rewards as a node operator, there's a place for you.
        </p>
        <div className="flex justify-center gap-4 flex-wrap">
          <Link
            to="/downloads"
            className="px-8 py-3 bg-gold-500 text-dark-950 font-bold rounded-lg hover:bg-gold-400 transition-all"
          >
            Get Started
          </Link>
          <Link
            to="/node-operators"
            className="px-8 py-3 glass rounded-lg font-medium hover:bg-white/10 transition-all"
          >
            Become an Operator
          </Link>
          <Link
            to="/contact"
            className="px-8 py-3 glass rounded-lg font-medium hover:bg-white/10 transition-all"
          >
            Contact Us
          </Link>
        </div>
      </div>
    </PageLayout>
  );
}
