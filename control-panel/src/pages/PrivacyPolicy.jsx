import { Link } from 'react-router-dom';
import PageLayout from '../components/PageLayout';
import {
  Shield,
  Eye,
  Database,
  Lock,
  Users,
  Globe,
  Server,
  FileText,
  Clock,
  MapPin,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Mail,
  Smartphone,
  Cookie,
  Baby,
  Scale
} from 'lucide-react';

export default function PrivacyPolicy() {
  const lastUpdated = 'December 28, 2025';
  const effectiveDate = 'December 28, 2025';

  return (
    <PageLayout
      title="Privacy Policy"
      description="How we protect your privacy and handle your data."
    >
      {/* Header Card */}
      <div className="glass rounded-xl p-8 mb-8">
        <div className="flex items-center gap-3 text-gold-500 mb-4">
          <Shield className="w-6 h-6" />
          <span className="font-bold">Your Privacy Matters</span>
        </div>
        <div className="grid md:grid-cols-2 gap-4 text-gray-400 text-sm">
          <div className="flex items-center gap-2">
            <Clock className="w-4 h-4" />
            <span>Last updated: {lastUpdated}</span>
          </div>
          <div className="flex items-center gap-2">
            <FileText className="w-4 h-4" />
            <span>Effective date: {effectiveDate}</span>
          </div>
        </div>
      </div>

      {/* Quick Summary */}
      <div className="glass rounded-xl p-6 mb-8 border-l-4 border-gold-500">
        <h3 className="font-bold text-white mb-3">Privacy at a Glance</h3>
        <div className="grid md:grid-cols-3 gap-4 text-sm">
          <div className="flex items-start gap-2">
            <CheckCircle className="w-5 h-5 text-green-500 flex-shrink-0" />
            <span className="text-gray-400">Strict no-logs policy</span>
          </div>
          <div className="flex items-start gap-2">
            <CheckCircle className="w-5 h-5 text-green-500 flex-shrink-0" />
            <span className="text-gray-400">No IP address tracking</span>
          </div>
          <div className="flex items-start gap-2">
            <CheckCircle className="w-5 h-5 text-green-500 flex-shrink-0" />
            <span className="text-gray-400">No browsing history stored</span>
          </div>
          <div className="flex items-start gap-2">
            <CheckCircle className="w-5 h-5 text-green-500 flex-shrink-0" />
            <span className="text-gray-400">Minimal data collection</span>
          </div>
          <div className="flex items-start gap-2">
            <CheckCircle className="w-5 h-5 text-green-500 flex-shrink-0" />
            <span className="text-gray-400">No third-party data sales</span>
          </div>
          <div className="flex items-start gap-2">
            <CheckCircle className="w-5 h-5 text-green-500 flex-shrink-0" />
            <span className="text-gray-400">GDPR & CCPA compliant</span>
          </div>
        </div>
      </div>

      <div className="prose prose-invert max-w-none space-y-10">
        {/* 1. Overview */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4">1. Overview</h2>
          <p className="text-gray-400 leading-relaxed mb-4">
            Aureo VPN ("we", "our", "us", or the "Company") is committed to protecting your privacy and ensuring
            the security of your personal information. This Privacy Policy explains how we collect, use, disclose,
            and safeguard your information when you use our VPN service, website, mobile applications, and related
            services (collectively, the "Services").
          </p>
          <p className="text-gray-400 leading-relaxed">
            By using our Services, you consent to the data practices described in this policy. If you do not agree
            with this policy, please do not use our Services.
          </p>
        </section>

        {/* 2. What We DON'T Collect */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <Eye className="w-6 h-6 text-gold-500" />
            2. What We DO NOT Collect (No-Logs Policy)
          </h2>
          <p className="text-gray-400 mb-4">
            As a privacy-first VPN service, we maintain a strict no-logs policy. We do <strong className="text-white">NOT</strong> collect,
            store, or monitor any of the following:
          </p>

          <div className="grid md:grid-cols-2 gap-4 mb-6">
            <div className="glass rounded-xl p-4">
              <h4 className="font-bold text-white mb-3 text-sm uppercase tracking-wider">Connection Data</h4>
              <ul className="space-y-2 text-gray-400 text-sm">
                <li className="flex items-start gap-2">
                  <XCircle className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                  <span>Your originating IP address</span>
                </li>
                <li className="flex items-start gap-2">
                  <XCircle className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                  <span>VPN IP address assigned to you</span>
                </li>
                <li className="flex items-start gap-2">
                  <XCircle className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                  <span>Connection timestamps</span>
                </li>
                <li className="flex items-start gap-2">
                  <XCircle className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                  <span>Session duration</span>
                </li>
              </ul>
            </div>
            <div className="glass rounded-xl p-4">
              <h4 className="font-bold text-white mb-3 text-sm uppercase tracking-wider">Activity Data</h4>
              <ul className="space-y-2 text-gray-400 text-sm">
                <li className="flex items-start gap-2">
                  <XCircle className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                  <span>Browsing history or websites visited</span>
                </li>
                <li className="flex items-start gap-2">
                  <XCircle className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                  <span>DNS queries</span>
                </li>
                <li className="flex items-start gap-2">
                  <XCircle className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                  <span>Traffic content or metadata</span>
                </li>
                <li className="flex items-start gap-2">
                  <XCircle className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                  <span>Bandwidth usage per session</span>
                </li>
              </ul>
            </div>
          </div>

          <div className="glass rounded-xl p-6 bg-green-500/5 border border-green-500/20">
            <h4 className="font-bold text-white mb-2 flex items-center gap-2">
              <Shield className="w-5 h-5 text-green-500" />
              Warrant Canary
            </h4>
            <p className="text-gray-400 text-sm">
              We have never received a National Security Letter, FISA order, or any other classified request
              for user information. We have never been required to log user activity. We will update this
              canary if any of these statements become untrue. Last verified: {lastUpdated}.
            </p>
          </div>
        </section>

        {/* 3. What We DO Collect */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <Database className="w-6 h-6 text-gold-500" />
            3. Information We Collect
          </h2>
          <p className="text-gray-400 mb-6">
            To provide our Services, we collect minimal information necessary for operation:
          </p>

          <div className="space-y-6">
            {/* Account Information */}
            <div className="glass rounded-xl p-6">
              <h4 className="font-bold text-white mb-4 flex items-center gap-2">
                <Users className="w-5 h-5 text-gold-500" />
                Account Information
              </h4>
              <div className="grid md:grid-cols-2 gap-4 text-sm">
                <div>
                  <span className="text-gray-500 block mb-1">Data Collected:</span>
                  <ul className="text-gray-400 space-y-1">
                    <li>- Email address (optional for anonymous accounts)</li>
                    <li>- Username</li>
                    <li>- Password hash (Argon2id encrypted)</li>
                    <li>- Account creation date</li>
                    <li>- Subscription tier</li>
                  </ul>
                </div>
                <div>
                  <span className="text-gray-500 block mb-1">Purpose:</span>
                  <ul className="text-gray-400 space-y-1">
                    <li>- Account authentication</li>
                    <li>- Password recovery</li>
                    <li>- Service communication</li>
                    <li>- Subscription management</li>
                  </ul>
                </div>
              </div>
            </div>

            {/* Payment Information */}
            <div className="glass rounded-xl p-6">
              <h4 className="font-bold text-white mb-4 flex items-center gap-2">
                <Scale className="w-5 h-5 text-gold-500" />
                Payment Information
              </h4>
              <div className="grid md:grid-cols-2 gap-4 text-sm">
                <div>
                  <span className="text-gray-500 block mb-1">Data Collected:</span>
                  <ul className="text-gray-400 space-y-1">
                    <li>- Blockchain transaction hash</li>
                    <li>- Payment amount and currency</li>
                    <li>- Payment status</li>
                    <li>- Subscription period purchased</li>
                  </ul>
                </div>
                <div>
                  <span className="text-gray-500 block mb-1">What We DON'T Store:</span>
                  <ul className="text-gray-400 space-y-1">
                    <li>- Your wallet address (deleted after verification)</li>
                    <li>- Transaction history beyond purchase</li>
                    <li>- Wallet balances</li>
                  </ul>
                </div>
              </div>
            </div>

            {/* Technical Data */}
            <div className="glass rounded-xl p-6">
              <h4 className="font-bold text-white mb-4 flex items-center gap-2">
                <Server className="w-5 h-5 text-gold-500" />
                Aggregate Technical Data
              </h4>
              <p className="text-gray-400 text-sm mb-3">
                We collect anonymized, aggregate data that cannot be linked to individual users:
              </p>
              <ul className="text-gray-400 text-sm space-y-1">
                <li>- Total server load and capacity utilization</li>
                <li>- Network performance metrics</li>
                <li>- Anonymized error reports for debugging</li>
                <li>- General geographic distribution of server usage (country-level only)</li>
              </ul>
            </div>
          </div>
        </section>

        {/* 4. Node Operators */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <Server className="w-6 h-6 text-gold-500" />
            4. Node Operator Data
          </h2>
          <p className="text-gray-400 mb-4">
            If you operate a VPN node on our network, we collect additional information necessary for
            the reward system and network management:
          </p>

          <div className="glass rounded-xl p-6 space-y-4">
            <div className="grid md:grid-cols-2 gap-6">
              <div>
                <h5 className="font-bold text-white mb-2">Required Information</h5>
                <ul className="text-gray-400 text-sm space-y-2">
                  <li className="flex items-start gap-2">
                    <CheckCircle className="w-4 h-4 text-gold-500 flex-shrink-0 mt-0.5" />
                    <span>Cryptocurrency wallet address (for payouts)</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <CheckCircle className="w-4 h-4 text-gold-500 flex-shrink-0 mt-0.5" />
                    <span>Node public IP address and hostname</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <CheckCircle className="w-4 h-4 text-gold-500 flex-shrink-0 mt-0.5" />
                    <span>Server location (country/city)</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <CheckCircle className="w-4 h-4 text-gold-500 flex-shrink-0 mt-0.5" />
                    <span>WireGuard/OpenVPN public keys</span>
                  </li>
                </ul>
              </div>
              <div>
                <h5 className="font-bold text-white mb-2">Performance Metrics</h5>
                <ul className="text-gray-400 text-sm space-y-2">
                  <li className="flex items-start gap-2">
                    <CheckCircle className="w-4 h-4 text-gold-500 flex-shrink-0 mt-0.5" />
                    <span>Node uptime percentage</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <CheckCircle className="w-4 h-4 text-gold-500 flex-shrink-0 mt-0.5" />
                    <span>Total bandwidth served (aggregate)</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <CheckCircle className="w-4 h-4 text-gold-500 flex-shrink-0 mt-0.5" />
                    <span>Current connection count</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <CheckCircle className="w-4 h-4 text-gold-500 flex-shrink-0 mt-0.5" />
                    <span>Reputation and tier score</span>
                  </li>
                </ul>
              </div>
            </div>
            <p className="text-gray-500 text-xs mt-4">
              Note: Node operators are contractually prohibited from logging user traffic. We perform
              regular audits to ensure compliance.
            </p>
          </div>
        </section>

        {/* 5. Data Security */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <Lock className="w-6 h-6 text-gold-500" />
            5. Data Security
          </h2>
          <p className="text-gray-400 mb-6">
            We implement comprehensive security measures to protect your data:
          </p>

          <div className="grid md:grid-cols-2 gap-6">
            <div className="glass rounded-xl p-6">
              <h4 className="font-bold text-white mb-4">Encryption Standards</h4>
              <ul className="space-y-3 text-gray-400 text-sm">
                <li className="flex items-start gap-3">
                  <CheckCircle className="w-4 h-4 text-green-500 flex-shrink-0 mt-0.5" />
                  <div>
                    <span className="text-white">VPN Traffic:</span> ChaCha20-Poly1305 (WireGuard) or AES-256-GCM (OpenVPN)
                  </div>
                </li>
                <li className="flex items-start gap-3">
                  <CheckCircle className="w-4 h-4 text-green-500 flex-shrink-0 mt-0.5" />
                  <div>
                    <span className="text-white">Password Hashing:</span> Argon2id with unique salt per user
                  </div>
                </li>
                <li className="flex items-start gap-3">
                  <CheckCircle className="w-4 h-4 text-green-500 flex-shrink-0 mt-0.5" />
                  <div>
                    <span className="text-white">Data at Rest:</span> AES-256 encryption
                  </div>
                </li>
                <li className="flex items-start gap-3">
                  <CheckCircle className="w-4 h-4 text-green-500 flex-shrink-0 mt-0.5" />
                  <div>
                    <span className="text-white">API Communications:</span> TLS 1.3 with certificate pinning
                  </div>
                </li>
              </ul>
            </div>
            <div className="glass rounded-xl p-6">
              <h4 className="font-bold text-white mb-4">Infrastructure Security</h4>
              <ul className="space-y-3 text-gray-400 text-sm">
                <li className="flex items-start gap-3">
                  <CheckCircle className="w-4 h-4 text-green-500 flex-shrink-0 mt-0.5" />
                  <span>RAM-only servers (no persistent disk storage)</span>
                </li>
                <li className="flex items-start gap-3">
                  <CheckCircle className="w-4 h-4 text-green-500 flex-shrink-0 mt-0.5" />
                  <span>Regular third-party security audits</span>
                </li>
                <li className="flex items-start gap-3">
                  <CheckCircle className="w-4 h-4 text-green-500 flex-shrink-0 mt-0.5" />
                  <span>Bug bounty program for vulnerability disclosure</span>
                </li>
                <li className="flex items-start gap-3">
                  <CheckCircle className="w-4 h-4 text-green-500 flex-shrink-0 mt-0.5" />
                  <span>24/7 security monitoring and incident response</span>
                </li>
                <li className="flex items-start gap-3">
                  <CheckCircle className="w-4 h-4 text-green-500 flex-shrink-0 mt-0.5" />
                  <span>Open-source codebase for community review</span>
                </li>
              </ul>
            </div>
          </div>
        </section>

        {/* 6. Cookies & Tracking */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <Cookie className="w-6 h-6 text-gold-500" />
            6. Cookies & Tracking Technologies
          </h2>

          <div className="glass rounded-xl p-6 space-y-4">
            <div>
              <h4 className="font-bold text-white mb-2">Essential Cookies Only</h4>
              <p className="text-gray-400 text-sm">
                We only use strictly necessary cookies for authentication and session management.
                We do <strong className="text-white">not</strong> use:
              </p>
              <ul className="text-gray-400 text-sm mt-2 space-y-1">
                <li className="flex items-center gap-2">
                  <XCircle className="w-4 h-4 text-red-500" />
                  <span>Third-party tracking cookies</span>
                </li>
                <li className="flex items-center gap-2">
                  <XCircle className="w-4 h-4 text-red-500" />
                  <span>Advertising or marketing cookies</span>
                </li>
                <li className="flex items-center gap-2">
                  <XCircle className="w-4 h-4 text-red-500" />
                  <span>Social media tracking pixels</span>
                </li>
                <li className="flex items-center gap-2">
                  <XCircle className="w-4 h-4 text-red-500" />
                  <span>Cross-site tracking technologies</span>
                </li>
              </ul>
            </div>
            <div>
              <h4 className="font-bold text-white mb-2">Analytics</h4>
              <p className="text-gray-400 text-sm">
                We use Plausible Analytics, a privacy-focused analytics tool that does not use cookies,
                does not collect personal data, and is fully GDPR compliant. No data is shared with
                third parties.
              </p>
            </div>
          </div>
        </section>

        {/* 7. Third-Party Services */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <Globe className="w-6 h-6 text-gold-500" />
            7. Third-Party Services
          </h2>

          <div className="space-y-4">
            <p className="text-gray-400">
              We minimize the use of third-party services. When we do use them, we select privacy-respecting options:
            </p>

            <div className="grid md:grid-cols-2 gap-4">
              <div className="glass rounded-xl p-4">
                <h4 className="font-bold text-white mb-2">Plausible Analytics</h4>
                <p className="text-gray-400 text-sm">Privacy-focused, cookie-free website analytics. No personal data collected.</p>
              </div>
              <div className="glass rounded-xl p-4">
                <h4 className="font-bold text-white mb-2">Blockchain Networks</h4>
                <p className="text-gray-400 text-sm">ETH, BTC, LTC networks for payment processing. Transaction data is public on respective blockchains.</p>
              </div>
              <div className="glass rounded-xl p-4">
                <h4 className="font-bold text-white mb-2">Cloudflare</h4>
                <p className="text-gray-400 text-sm">DDoS protection and CDN for our website (not VPN traffic). Subject to Cloudflare's privacy policy.</p>
              </div>
              <div className="glass rounded-xl p-4">
                <h4 className="font-bold text-white mb-2">GitHub</h4>
                <p className="text-gray-400 text-sm">Open-source code hosting. Contributions are public and attributed.</p>
              </div>
            </div>
          </div>
        </section>

        {/* 8. Data Retention */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <Clock className="w-6 h-6 text-gold-500" />
            8. Data Retention
          </h2>

          <div className="glass rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-gray-400 border-b border-white/10 bg-dark-800">
                  <th className="px-6 py-3">Data Type</th>
                  <th className="px-6 py-3">Retention Period</th>
                </tr>
              </thead>
              <tbody className="text-gray-300">
                <tr className="border-b border-white/5">
                  <td className="px-6 py-3">Account information</td>
                  <td className="px-6 py-3">Until account deletion + 30 days</td>
                </tr>
                <tr className="border-b border-white/5">
                  <td className="px-6 py-3">Payment records</td>
                  <td className="px-6 py-3">7 years (legal requirement)</td>
                </tr>
                <tr className="border-b border-white/5">
                  <td className="px-6 py-3">Support tickets</td>
                  <td className="px-6 py-3">2 years after resolution</td>
                </tr>
                <tr className="border-b border-white/5">
                  <td className="px-6 py-3">Node operator metrics</td>
                  <td className="px-6 py-3">1 year rolling window</td>
                </tr>
                <tr>
                  <td className="px-6 py-3">VPN connection logs</td>
                  <td className="px-6 py-3 text-green-400 font-medium">Never stored</td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        {/* 9. Your Rights */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <Scale className="w-6 h-6 text-gold-500" />
            9. Your Rights
          </h2>

          <p className="text-gray-400 mb-4">
            Depending on your location, you may have the following rights regarding your personal data:
          </p>

          <div className="grid md:grid-cols-2 gap-4 mb-6">
            <div className="glass rounded-xl p-4">
              <h4 className="font-bold text-white mb-2">GDPR Rights (EU/EEA)</h4>
              <ul className="text-gray-400 text-sm space-y-2">
                <li>- Right to access your data</li>
                <li>- Right to rectification</li>
                <li>- Right to erasure ("right to be forgotten")</li>
                <li>- Right to data portability</li>
                <li>- Right to object to processing</li>
                <li>- Right to withdraw consent</li>
              </ul>
            </div>
            <div className="glass rounded-xl p-4">
              <h4 className="font-bold text-white mb-2">CCPA Rights (California)</h4>
              <ul className="text-gray-400 text-sm space-y-2">
                <li>- Right to know what data is collected</li>
                <li>- Right to delete personal information</li>
                <li>- Right to opt-out of data sales (we don't sell data)</li>
                <li>- Right to non-discrimination</li>
              </ul>
            </div>
          </div>

          <div className="glass rounded-xl p-6">
            <h4 className="font-bold text-white mb-3">How to Exercise Your Rights</h4>
            <p className="text-gray-400 text-sm mb-4">
              To exercise any of these rights, you can:
            </p>
            <ul className="text-gray-400 text-sm space-y-2">
              <li className="flex items-center gap-2">
                <Mail className="w-4 h-4 text-gold-500" />
                <span>Email us at: <a href="mailto:privacy@aureo-vpn.com" className="text-gold-500 hover:underline">privacy@aureo-vpn.com</a></span>
              </li>
              <li className="flex items-center gap-2">
                <Smartphone className="w-4 h-4 text-gold-500" />
                <span>Use the "Delete Account" feature in your dashboard</span>
              </li>
            </ul>
            <p className="text-gray-500 text-xs mt-4">
              We will respond to all data subject requests within 30 days. Identity verification may be required.
            </p>
          </div>
        </section>

        {/* 10. International Data Transfers */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <MapPin className="w-6 h-6 text-gold-500" />
            10. International Data Transfers
          </h2>
          <p className="text-gray-400 mb-4">
            Our decentralized network operates globally. When data crosses borders:
          </p>
          <ul className="text-gray-400 space-y-2 text-sm">
            <li className="flex items-start gap-2">
              <CheckCircle className="w-4 h-4 text-gold-500 flex-shrink-0 mt-0.5" />
              <span>VPN traffic is encrypted end-to-end and not logged, making transfer concerns minimal</span>
            </li>
            <li className="flex items-start gap-2">
              <CheckCircle className="w-4 h-4 text-gold-500 flex-shrink-0 mt-0.5" />
              <span>Account data is processed in Switzerland (adequate protection under EU law)</span>
            </li>
            <li className="flex items-start gap-2">
              <CheckCircle className="w-4 h-4 text-gold-500 flex-shrink-0 mt-0.5" />
              <span>Standard Contractual Clauses (SCCs) are used where required</span>
            </li>
          </ul>
        </section>

        {/* 11. Children's Privacy */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <Baby className="w-6 h-6 text-gold-500" />
            11. Children's Privacy
          </h2>
          <p className="text-gray-400">
            Our Services are not intended for children under 16 years of age. We do not knowingly collect
            personal information from children. If we discover that we have collected information from a
            child under 16, we will delete it immediately. If you believe we have information about a child,
            please contact us at <a href="mailto:privacy@aureo-vpn.com" className="text-gold-500 hover:underline">privacy@aureo-vpn.com</a>.
          </p>
        </section>

        {/* 12. Law Enforcement */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <AlertTriangle className="w-6 h-6 text-gold-500" />
            12. Law Enforcement Requests
          </h2>
          <div className="glass rounded-xl p-6 bg-yellow-500/5 border border-yellow-500/20">
            <p className="text-gray-400 mb-4">
              Because we do not log VPN activity, we cannot provide information we don't have. When we
              receive legal requests:
            </p>
            <ul className="text-gray-400 space-y-2 text-sm">
              <li className="flex items-start gap-2">
                <CheckCircle className="w-4 h-4 text-gold-500 flex-shrink-0 mt-0.5" />
                <span>We verify the request's legitimacy and legal basis</span>
              </li>
              <li className="flex items-start gap-2">
                <CheckCircle className="w-4 h-4 text-gold-500 flex-shrink-0 mt-0.5" />
                <span>We only comply with Swiss law (our jurisdiction)</span>
              </li>
              <li className="flex items-start gap-2">
                <CheckCircle className="w-4 h-4 text-gold-500 flex-shrink-0 mt-0.5" />
                <span>We notify users when legally permitted</span>
              </li>
              <li className="flex items-start gap-2">
                <CheckCircle className="w-4 h-4 text-gold-500 flex-shrink-0 mt-0.5" />
                <span>We publish transparency reports quarterly</span>
              </li>
            </ul>
          </div>
        </section>

        {/* 13. Changes to Policy */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4">13. Changes to This Policy</h2>
          <p className="text-gray-400">
            We may update this Privacy Policy from time to time. When we make significant changes:
          </p>
          <ul className="text-gray-400 mt-2 space-y-1 text-sm">
            <li>- We will update the "Last updated" date at the top</li>
            <li>- We will notify registered users via email</li>
            <li>- We will post a notice on our website for 30 days</li>
            <li>- Significant changes take effect 30 days after notice</li>
          </ul>
        </section>

        {/* 14. Contact Us */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4">14. Contact Us</h2>
          <div className="glass rounded-xl p-6">
            <p className="text-gray-400 mb-4">
              If you have questions about this Privacy Policy or our data practices:
            </p>
            <div className="space-y-3 text-sm">
              <div className="flex items-center gap-3">
                <Mail className="w-5 h-5 text-gold-500" />
                <div>
                  <span className="text-gray-500 block">Privacy inquiries:</span>
                  <a href="mailto:privacy@aureo-vpn.com" className="text-gold-500 hover:underline">privacy@aureo-vpn.com</a>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <Mail className="w-5 h-5 text-gold-500" />
                <div>
                  <span className="text-gray-500 block">Data Protection Officer:</span>
                  <a href="mailto:dpo@aureo-vpn.com" className="text-gold-500 hover:underline">dpo@aureo-vpn.com</a>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <MapPin className="w-5 h-5 text-gold-500" />
                <div>
                  <span className="text-gray-500 block">Registered Address:</span>
                  <span className="text-gray-400">Aureo Network AG, Zurich, Switzerland</span>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* Related Links */}
        <section className="pt-6 border-t border-white/10">
          <h3 className="font-bold text-white mb-4">Related Documents</h3>
          <div className="flex flex-wrap gap-4">
            <Link to="/terms" className="glass px-4 py-2 rounded-lg text-gold-500 hover:bg-white/5 transition-all text-sm">
              Terms of Service
            </Link>
            <Link to="/docs" className="glass px-4 py-2 rounded-lg text-gold-500 hover:bg-white/5 transition-all text-sm">
              Documentation
            </Link>
            <Link to="/contact" className="glass px-4 py-2 rounded-lg text-gold-500 hover:bg-white/5 transition-all text-sm">
              Contact Us
            </Link>
          </div>
        </section>
      </div>
    </PageLayout>
  );
}
