import { Link } from 'react-router-dom';
import PageLayout from '../components/PageLayout';
import {
  FileText,
  AlertTriangle,
  Check,
  X,
  Shield,
  Server,
  Coins,
  Scale,
  Clock,
  Users,
  Globe,
  Lock,
  Zap,
  Ban,
  Gavel,
  Mail,
  MapPin,
  CreditCard,
  RefreshCw,
  AlertCircle,
  Info,
  BookOpen
} from 'lucide-react';

export default function TermsOfService() {
  const lastUpdated = 'December 28, 2025';
  const effectiveDate = 'December 28, 2025';

  return (
    <PageLayout
      title="Terms of Service"
      description="Please read these terms carefully before using Aureo VPN."
    >
      {/* Header Card */}
      <div className="glass rounded-xl p-8 mb-8">
        <div className="flex items-center gap-3 text-gold-500 mb-4">
          <FileText className="w-6 h-6" />
          <span className="font-bold">Terms and Conditions</span>
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
        <h3 className="font-bold text-white mb-3 flex items-center gap-2">
          <Info className="w-5 h-5 text-gold-500" />
          Key Points Summary
        </h3>
        <div className="grid md:grid-cols-2 gap-4 text-sm text-gray-400">
          <div>
            <ul className="space-y-2">
              <li className="flex items-start gap-2">
                <Check className="w-4 h-4 text-green-500 flex-shrink-0 mt-0.5" />
                <span>30-day money-back guarantee</span>
              </li>
              <li className="flex items-start gap-2">
                <Check className="w-4 h-4 text-green-500 flex-shrink-0 mt-0.5" />
                <span>Crypto payments (ETH, BTC, LTC)</span>
              </li>
              <li className="flex items-start gap-2">
                <Check className="w-4 h-4 text-green-500 flex-shrink-0 mt-0.5" />
                <span>Node operators earn rewards</span>
              </li>
            </ul>
          </div>
          <div>
            <ul className="space-y-2">
              <li className="flex items-start gap-2">
                <X className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                <span>No illegal activities</span>
              </li>
              <li className="flex items-start gap-2">
                <X className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                <span>No account sharing</span>
              </li>
              <li className="flex items-start gap-2">
                <X className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                <span>No abuse of service</span>
              </li>
            </ul>
          </div>
        </div>
      </div>

      <div className="prose prose-invert max-w-none space-y-10">
        {/* 1. Acceptance of Terms */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <Gavel className="w-6 h-6 text-gold-500" />
            1. Acceptance of Terms
          </h2>
          <p className="text-gray-400 leading-relaxed mb-4">
            By accessing, downloading, installing, or using Aureo VPN services, applications, or website
            (collectively, the "Services"), you ("User" or "you") agree to be bound by these Terms of Service
            ("Terms"), our <Link to="/privacy" className="text-gold-500 hover:underline">Privacy Policy</Link>,
            and any additional terms that may apply.
          </p>
          <div className="glass rounded-xl p-4 bg-yellow-500/5 border border-yellow-500/20">
            <p className="text-gray-400 text-sm flex items-start gap-2">
              <AlertCircle className="w-5 h-5 text-yellow-500 flex-shrink-0" />
              <span>
                <strong className="text-white">Important:</strong> If you do not agree to these Terms, you must not use our Services.
                By using our Services, you represent that you are at least 16 years old and have the legal capacity
                to enter into binding agreements.
              </span>
            </p>
          </div>
        </section>

        {/* 2. Description of Service */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <Globe className="w-6 h-6 text-gold-500" />
            2. Description of Services
          </h2>
          <p className="text-gray-400 mb-4">
            Aureo VPN provides the following services:
          </p>

          <div className="grid md:grid-cols-2 gap-4">
            <div className="glass rounded-xl p-4">
              <h4 className="font-bold text-white mb-3 flex items-center gap-2">
                <Shield className="w-5 h-5 text-gold-500" />
                VPN Services
              </h4>
              <ul className="space-y-2 text-gray-400 text-sm">
                <li className="flex items-start gap-2">
                  <Check className="w-4 h-4 text-green-500 flex-shrink-0 mt-0.5" />
                  <span>Encrypted VPN tunnels (WireGuard & OpenVPN)</span>
                </li>
                <li className="flex items-start gap-2">
                  <Check className="w-4 h-4 text-green-500 flex-shrink-0 mt-0.5" />
                  <span>Access to global network of VPN servers</span>
                </li>
                <li className="flex items-start gap-2">
                  <Check className="w-4 h-4 text-green-500 flex-shrink-0 mt-0.5" />
                  <span>Kill switch and DNS leak protection</span>
                </li>
                <li className="flex items-start gap-2">
                  <Check className="w-4 h-4 text-green-500 flex-shrink-0 mt-0.5" />
                  <span>Multi-hop VPN connections</span>
                </li>
              </ul>
            </div>
            <div className="glass rounded-xl p-4">
              <h4 className="font-bold text-white mb-3 flex items-center gap-2">
                <Server className="w-5 h-5 text-gold-500" />
                Node Operator Platform
              </h4>
              <ul className="space-y-2 text-gray-400 text-sm">
                <li className="flex items-start gap-2">
                  <Check className="w-4 h-4 text-green-500 flex-shrink-0 mt-0.5" />
                  <span>Decentralized VPN infrastructure</span>
                </li>
                <li className="flex items-start gap-2">
                  <Check className="w-4 h-4 text-green-500 flex-shrink-0 mt-0.5" />
                  <span>Cryptocurrency-based reward system</span>
                </li>
                <li className="flex items-start gap-2">
                  <Check className="w-4 h-4 text-green-500 flex-shrink-0 mt-0.5" />
                  <span>Real-time performance monitoring</span>
                </li>
                <li className="flex items-start gap-2">
                  <Check className="w-4 h-4 text-green-500 flex-shrink-0 mt-0.5" />
                  <span>Automated payouts via blockchain</span>
                </li>
              </ul>
            </div>
          </div>
        </section>

        {/* 3. Account Registration */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <Users className="w-6 h-6 text-gold-500" />
            3. Account Registration & Security
          </h2>

          <div className="space-y-4">
            <div className="glass rounded-xl p-6">
              <h4 className="font-bold text-white mb-3">Account Requirements</h4>
              <ul className="space-y-2 text-gray-400 text-sm">
                <li className="flex items-start gap-2">
                  <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
                  <span>You must provide accurate and complete registration information</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
                  <span>You are responsible for maintaining the confidentiality of your credentials</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
                  <span>You must notify us immediately of any unauthorized account access</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
                  <span>One person may only register one account (multiple accounts may be terminated)</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
                  <span>Account sharing or transfer is prohibited without prior authorization</span>
                </li>
              </ul>
            </div>

            <div className="glass rounded-xl p-6">
              <h4 className="font-bold text-white mb-3">Security Recommendations</h4>
              <ul className="space-y-2 text-gray-400 text-sm">
                <li className="flex items-start gap-2">
                  <Lock className="w-4 h-4 text-gold-500 flex-shrink-0 mt-0.5" />
                  <span>Use a strong, unique password (min. 8 characters with mixed case, numbers, symbols)</span>
                </li>
                <li className="flex items-start gap-2">
                  <Lock className="w-4 h-4 text-gold-500 flex-shrink-0 mt-0.5" />
                  <span>Enable two-factor authentication when available</span>
                </li>
                <li className="flex items-start gap-2">
                  <Lock className="w-4 h-4 text-gold-500 flex-shrink-0 mt-0.5" />
                  <span>Do not reuse passwords from other services</span>
                </li>
              </ul>
            </div>
          </div>
        </section>

        {/* 4. User Responsibilities */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <Scale className="w-6 h-6 text-gold-500" />
            4. User Responsibilities
          </h2>
          <p className="text-gray-400 mb-4">
            As a user of Aureo VPN, you agree to:
          </p>
          <div className="glass rounded-xl p-6 space-y-3">
            <div className="flex items-start gap-3">
              <Check className="w-5 h-5 text-green-500 flex-shrink-0" />
              <span className="text-gray-400">Comply with all applicable local, national, and international laws and regulations</span>
            </div>
            <div className="flex items-start gap-3">
              <Check className="w-5 h-5 text-green-500 flex-shrink-0" />
              <span className="text-gray-400">Use the Services only for lawful purposes</span>
            </div>
            <div className="flex items-start gap-3">
              <Check className="w-5 h-5 text-green-500 flex-shrink-0" />
              <span className="text-gray-400">Not attempt to compromise, reverse-engineer, or circumvent any security measures</span>
            </div>
            <div className="flex items-start gap-3">
              <Check className="w-5 h-5 text-green-500 flex-shrink-0" />
              <span className="text-gray-400">Not use the Services to infringe upon the rights of others</span>
            </div>
            <div className="flex items-start gap-3">
              <Check className="w-5 h-5 text-green-500 flex-shrink-0" />
              <span className="text-gray-400">Not resell, sublicense, or commercially exploit the Services without authorization</span>
            </div>
            <div className="flex items-start gap-3">
              <Check className="w-5 h-5 text-green-500 flex-shrink-0" />
              <span className="text-gray-400">Report any vulnerabilities or security issues responsibly</span>
            </div>
          </div>
        </section>

        {/* 5. Prohibited Activities */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <Ban className="w-6 h-6 text-red-500" />
            5. Prohibited Activities
          </h2>
          <p className="text-gray-400 mb-4">
            The following activities are <strong className="text-red-400">strictly prohibited</strong> and may result in
            immediate account termination without refund:
          </p>

          <div className="grid md:grid-cols-2 gap-4 mb-6">
            <div className="glass rounded-xl p-4 border border-red-500/20">
              <h4 className="font-bold text-white mb-3 text-sm uppercase tracking-wider">Criminal Activities</h4>
              <ul className="space-y-2 text-gray-400 text-sm">
                <li className="flex items-start gap-2">
                  <X className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                  <span>Child sexual abuse material (reported to authorities)</span>
                </li>
                <li className="flex items-start gap-2">
                  <X className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                  <span>Terrorism or violent extremism</span>
                </li>
                <li className="flex items-start gap-2">
                  <X className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                  <span>Drug trafficking</span>
                </li>
                <li className="flex items-start gap-2">
                  <X className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                  <span>Human trafficking</span>
                </li>
                <li className="flex items-start gap-2">
                  <X className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                  <span>Money laundering</span>
                </li>
              </ul>
            </div>
            <div className="glass rounded-xl p-4 border border-red-500/20">
              <h4 className="font-bold text-white mb-3 text-sm uppercase tracking-wider">Cyber Attacks</h4>
              <ul className="space-y-2 text-gray-400 text-sm">
                <li className="flex items-start gap-2">
                  <X className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                  <span>DDoS attacks or flooding</span>
                </li>
                <li className="flex items-start gap-2">
                  <X className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                  <span>Malware distribution</span>
                </li>
                <li className="flex items-start gap-2">
                  <X className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                  <span>Phishing or social engineering</span>
                </li>
                <li className="flex items-start gap-2">
                  <X className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                  <span>Unauthorized system access</span>
                </li>
                <li className="flex items-start gap-2">
                  <X className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                  <span>Port scanning or vulnerability probing</span>
                </li>
              </ul>
            </div>
            <div className="glass rounded-xl p-4 border border-red-500/20">
              <h4 className="font-bold text-white mb-3 text-sm uppercase tracking-wider">Abuse</h4>
              <ul className="space-y-2 text-gray-400 text-sm">
                <li className="flex items-start gap-2">
                  <X className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                  <span>Spam or unsolicited bulk messaging</span>
                </li>
                <li className="flex items-start gap-2">
                  <X className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                  <span>Harassment, threats, or stalking</span>
                </li>
                <li className="flex items-start gap-2">
                  <X className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                  <span>Identity theft or impersonation</span>
                </li>
                <li className="flex items-start gap-2">
                  <X className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                  <span>Fraud or scams</span>
                </li>
              </ul>
            </div>
            <div className="glass rounded-xl p-4 border border-red-500/20">
              <h4 className="font-bold text-white mb-3 text-sm uppercase tracking-wider">Service Abuse</h4>
              <ul className="space-y-2 text-gray-400 text-sm">
                <li className="flex items-start gap-2">
                  <X className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                  <span>Circumventing usage limits</span>
                </li>
                <li className="flex items-start gap-2">
                  <X className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                  <span>Automated abuse or botting</span>
                </li>
                <li className="flex items-start gap-2">
                  <X className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                  <span>Excessive bandwidth consumption</span>
                </li>
                <li className="flex items-start gap-2">
                  <X className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />
                  <span>Interfering with other users' access</span>
                </li>
              </ul>
            </div>
          </div>

          <div className="glass rounded-xl p-4 bg-red-500/5 border border-red-500/20">
            <p className="text-gray-400 text-sm">
              <strong className="text-white">Enforcement:</strong> We reserve the right to investigate suspected violations,
              suspend or terminate accounts, and cooperate with law enforcement when required. Due to our no-logs policy,
              we may be unable to provide specific activity data.
            </p>
          </div>
        </section>

        {/* 6. Node Operator Terms */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <Server className="w-6 h-6 text-gold-500" />
            6. Node Operator Terms
          </h2>
          <p className="text-gray-400 mb-4">
            Additional terms apply if you choose to operate a VPN node on our network:
          </p>

          <div className="space-y-4">
            <div className="glass rounded-xl p-6">
              <h4 className="font-bold text-white mb-3">Operator Requirements</h4>
              <ul className="space-y-2 text-gray-400 text-sm">
                <li className="flex items-start gap-2">
                  <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
                  <span>Maintain minimum uptime requirements for your tier (Bronze: 50%, Silver: 80%, Gold: 90%, Platinum: 95%)</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
                  <span><strong className="text-white">Do NOT log</strong> or monitor any user traffic passing through your node</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
                  <span>Keep node software up to date within 7 days of security releases</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
                  <span>Provide accurate server location and capability information</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
                  <span>Comply with local laws regarding VPN services in your jurisdiction</span>
                </li>
              </ul>
            </div>

            <div className="glass rounded-xl p-6">
              <h4 className="font-bold text-white mb-3 flex items-center gap-2">
                <Coins className="w-5 h-5 text-gold-500" />
                Reward Structure
              </h4>
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="text-left text-gray-400 border-b border-white/10">
                      <th className="py-2 pr-4">Tier</th>
                      <th className="py-2 pr-4">Rate/GB</th>
                      <th className="py-2 pr-4">Min Uptime</th>
                      <th className="py-2">Bonus</th>
                    </tr>
                  </thead>
                  <tbody className="text-gray-300">
                    <tr className="border-b border-white/5">
                      <td className="py-2 pr-4 text-amber-600 font-medium">Bronze</td>
                      <td className="py-2 pr-4">$0.010</td>
                      <td className="py-2 pr-4">50%</td>
                      <td className="py-2">1.0x</td>
                    </tr>
                    <tr className="border-b border-white/5">
                      <td className="py-2 pr-4 text-gray-400 font-medium">Silver</td>
                      <td className="py-2 pr-4">$0.015</td>
                      <td className="py-2 pr-4">80%</td>
                      <td className="py-2">1.2x</td>
                    </tr>
                    <tr className="border-b border-white/5">
                      <td className="py-2 pr-4 text-gold-500 font-medium">Gold</td>
                      <td className="py-2 pr-4">$0.020</td>
                      <td className="py-2 pr-4">90%</td>
                      <td className="py-2">1.5x</td>
                    </tr>
                    <tr>
                      <td className="py-2 pr-4 text-cyan-400 font-medium">Platinum</td>
                      <td className="py-2 pr-4">$0.025</td>
                      <td className="py-2 pr-4">95%</td>
                      <td className="py-2">2.0x</td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <p className="text-gray-500 text-xs mt-4">
                Rewards are calculated based on bandwidth served, quality score, and tier bonuses. We reserve the right
                to modify rates with 30 days notice.
              </p>
            </div>
          </div>
        </section>

        {/* 7. Payment Terms */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <CreditCard className="w-6 h-6 text-gold-500" />
            7. Payment Terms
          </h2>

          <div className="space-y-4">
            <div className="glass rounded-xl p-6">
              <h4 className="font-bold text-white mb-3">Cryptocurrency Payments</h4>
              <ul className="space-y-2 text-gray-400 text-sm">
                <li className="flex items-start gap-2">
                  <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
                  <span>Accepted currencies: Ethereum (ETH), Bitcoin (BTC), Litecoin (LTC)</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
                  <span>Payments are final once confirmed on the blockchain (minimum confirmations apply)</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
                  <span>Exchange rate fluctuations are the responsibility of the user</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
                  <span>Network fees (gas fees, transaction fees) are paid by the user</span>
                </li>
              </ul>
            </div>

            <div className="glass rounded-xl p-6">
              <h4 className="font-bold text-white mb-3 flex items-center gap-2">
                <RefreshCw className="w-5 h-5 text-gold-500" />
                Refund Policy
              </h4>
              <div className="space-y-4 text-gray-400 text-sm">
                <div>
                  <h5 className="text-white font-medium mb-1">30-Day Money-Back Guarantee</h5>
                  <p>New users are eligible for a full refund within 30 days of initial purchase, no questions asked.</p>
                </div>
                <div>
                  <h5 className="text-white font-medium mb-1">Refund Process</h5>
                  <ul className="space-y-1 mt-2">
                    <li>- Submit request via email to <a href="mailto:billing@aureo-vpn.com" className="text-gold-500 hover:underline">billing@aureo-vpn.com</a></li>
                    <li>- Refunds are processed within 5-7 business days</li>
                    <li>- Refunds are issued in the original cryptocurrency at current exchange rates</li>
                    <li>- Network transaction fees are deducted from the refund amount</li>
                  </ul>
                </div>
                <div>
                  <h5 className="text-white font-medium mb-1">Non-Refundable</h5>
                  <p>Accounts terminated for Terms violations are not eligible for refunds.</p>
                </div>
              </div>
            </div>

            <div className="glass rounded-xl p-6">
              <h4 className="font-bold text-white mb-3">Node Operator Payouts</h4>
              <ul className="space-y-2 text-gray-400 text-sm">
                <li className="flex items-start gap-2">
                  <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
                  <span>Minimum payout threshold: $10 USD equivalent</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
                  <span>Payouts are processed within 24 hours of request</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
                  <span>Earnings are calculated in real-time and visible in your dashboard</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
                  <span>Operators are responsible for their own tax obligations</span>
                </li>
              </ul>
            </div>
          </div>
        </section>

        {/* 8. Subscription & Billing */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <Clock className="w-6 h-6 text-gold-500" />
            8. Subscription & Billing
          </h2>

          <div className="glass rounded-xl p-6 space-y-4">
            <div>
              <h4 className="font-bold text-white mb-2">Subscription Plans</h4>
              <p className="text-gray-400 text-sm">
                We offer monthly, annual, and lifetime subscription options. Plan details and pricing are available
                on our <Link to="/#pricing" className="text-gold-500 hover:underline">pricing page</Link>.
              </p>
            </div>
            <div>
              <h4 className="font-bold text-white mb-2">Automatic Renewal</h4>
              <p className="text-gray-400 text-sm">
                Subscriptions do NOT auto-renew. Cryptocurrency payments are one-time. You will receive email reminders
                before your subscription expires.
              </p>
            </div>
            <div>
              <h4 className="font-bold text-white mb-2">Price Changes</h4>
              <p className="text-gray-400 text-sm">
                We may modify pricing with 30 days notice. Existing subscriptions continue at the original price
                until renewal.
              </p>
            </div>
          </div>
        </section>

        {/* 9. Service Availability */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <Zap className="w-6 h-6 text-gold-500" />
            9. Service Availability & SLA
          </h2>

          <div className="glass rounded-xl p-6 space-y-4">
            <div>
              <h4 className="font-bold text-white mb-2">Uptime Target</h4>
              <p className="text-gray-400 text-sm">
                We target 99.9% network availability. However, we do not guarantee uninterrupted service.
                Real-time status is available at <Link to="/status" className="text-gold-500 hover:underline">status.aureo-vpn.com</Link>.
              </p>
            </div>
            <div>
              <h4 className="font-bold text-white mb-2">Scheduled Maintenance</h4>
              <p className="text-gray-400 text-sm">
                Planned maintenance is announced 48 hours in advance via email and status page.
                We minimize maintenance during peak usage hours.
              </p>
            </div>
            <div>
              <h4 className="font-bold text-white mb-2">Force Majeure</h4>
              <p className="text-gray-400 text-sm">
                We are not liable for service interruptions due to circumstances beyond our control, including:
                natural disasters, government actions, cyber attacks, network outages, or third-party failures.
              </p>
            </div>
          </div>
        </section>

        {/* 10. Intellectual Property */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <BookOpen className="w-6 h-6 text-gold-500" />
            10. Intellectual Property
          </h2>

          <div className="glass rounded-xl p-6 space-y-4">
            <div>
              <h4 className="font-bold text-white mb-2">Our Property</h4>
              <p className="text-gray-400 text-sm">
                The Aureo VPN name, logo, and branding are trademarks of Aureo Network AG. You may not use
                our trademarks without prior written permission.
              </p>
            </div>
            <div>
              <h4 className="font-bold text-white mb-2">Open Source</h4>
              <p className="text-gray-400 text-sm">
                Our core software is open source under the MIT License. See our{' '}
                <a href="https://github.com/nikola43/aureo-vpn" target="_blank" rel="noopener noreferrer" className="text-gold-500 hover:underline">
                  GitHub repository
                </a>{' '}
                for license terms.
              </p>
            </div>
            <div>
              <h4 className="font-bold text-white mb-2">User Content</h4>
              <p className="text-gray-400 text-sm">
                You retain ownership of any content you transmit through our VPN. We do not access, store,
                or claim any rights to your content.
              </p>
            </div>
          </div>
        </section>

        {/* 11. Limitation of Liability */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <AlertTriangle className="w-6 h-6 text-gold-500" />
            11. Disclaimers & Limitation of Liability
          </h2>

          <div className="glass rounded-xl p-6 bg-yellow-500/5 border border-yellow-500/20 space-y-4">
            <div>
              <h4 className="font-bold text-white mb-2">AS-IS Provision</h4>
              <p className="text-gray-400 text-sm">
                THE SERVICES ARE PROVIDED "AS IS" AND "AS AVAILABLE" WITHOUT WARRANTIES OF ANY KIND,
                EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO WARRANTIES OF MERCHANTABILITY,
                FITNESS FOR A PARTICULAR PURPOSE, OR NON-INFRINGEMENT.
              </p>
            </div>
            <div>
              <h4 className="font-bold text-white mb-2">No Absolute Privacy Guarantee</h4>
              <p className="text-gray-400 text-sm">
                While we implement strong security measures, no VPN can guarantee complete anonymity or
                protection against all threats. Users are responsible for their own operational security.
              </p>
            </div>
            <div>
              <h4 className="font-bold text-white mb-2">Liability Cap</h4>
              <p className="text-gray-400 text-sm">
                OUR TOTAL LIABILITY FOR ANY CLAIMS ARISING FROM YOUR USE OF THE SERVICES SHALL NOT EXCEED
                THE AMOUNT YOU PAID FOR THE SERVICES IN THE TWELVE (12) MONTHS PRECEDING THE CLAIM.
              </p>
            </div>
            <div>
              <h4 className="font-bold text-white mb-2">Exclusions</h4>
              <p className="text-gray-400 text-sm">
                WE ARE NOT LIABLE FOR ANY INDIRECT, INCIDENTAL, SPECIAL, CONSEQUENTIAL, OR PUNITIVE DAMAGES,
                INCLUDING LOSS OF DATA, REVENUE, OR PROFITS.
              </p>
            </div>
          </div>
        </section>

        {/* 12. Indemnification */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4">12. Indemnification</h2>
          <p className="text-gray-400">
            You agree to indemnify, defend, and hold harmless Aureo Network AG, its officers, directors,
            employees, and agents from any claims, damages, losses, or expenses (including reasonable
            attorneys' fees) arising from:
          </p>
          <ul className="text-gray-400 mt-2 space-y-1 text-sm">
            <li>- Your use of the Services</li>
            <li>- Your violation of these Terms</li>
            <li>- Your violation of any third-party rights</li>
            <li>- Any content you transmit through our Services</li>
          </ul>
        </section>

        {/* 13. Termination */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4">13. Termination</h2>

          <div className="grid md:grid-cols-2 gap-4">
            <div className="glass rounded-xl p-4">
              <h4 className="font-bold text-white mb-2">By You</h4>
              <p className="text-gray-400 text-sm">
                You may cancel your account at any time through your dashboard or by contacting support.
                Refunds are subject to our refund policy.
              </p>
            </div>
            <div className="glass rounded-xl p-4">
              <h4 className="font-bold text-white mb-2">By Us</h4>
              <p className="text-gray-400 text-sm">
                We may suspend or terminate your access for Terms violations, non-payment, or any reason
                with reasonable notice (except for serious violations requiring immediate action).
              </p>
            </div>
          </div>
          <p className="text-gray-400 mt-4 text-sm">
            Upon termination, your right to use the Services ceases immediately. Sections regarding
            intellectual property, limitation of liability, and governing law survive termination.
          </p>
        </section>

        {/* 14. Changes to Terms */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4">14. Changes to Terms</h2>
          <p className="text-gray-400">
            We may modify these Terms at any time. When we make material changes:
          </p>
          <ul className="text-gray-400 mt-2 space-y-1 text-sm">
            <li>- We will update the "Last updated" date</li>
            <li>- We will notify registered users via email</li>
            <li>- We will post a prominent notice on our website</li>
            <li>- Material changes take effect 30 days after notice</li>
            <li>- Continued use after the effective date constitutes acceptance</li>
          </ul>
        </section>

        {/* 15. Governing Law */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <Gavel className="w-6 h-6 text-gold-500" />
            15. Governing Law & Disputes
          </h2>

          <div className="glass rounded-xl p-6 space-y-4">
            <div>
              <h4 className="font-bold text-white mb-2">Jurisdiction</h4>
              <p className="text-gray-400 text-sm">
                These Terms are governed by the laws of Switzerland, without regard to conflict of law principles.
              </p>
            </div>
            <div>
              <h4 className="font-bold text-white mb-2">Dispute Resolution</h4>
              <p className="text-gray-400 text-sm">
                Any disputes shall first be attempted to be resolved through good-faith negotiation.
                If unsuccessful, disputes shall be resolved by binding arbitration in Zurich, Switzerland,
                in accordance with the Swiss Rules of International Arbitration.
              </p>
            </div>
            <div>
              <h4 className="font-bold text-white mb-2">Class Action Waiver</h4>
              <p className="text-gray-400 text-sm">
                You agree to resolve disputes individually and waive any right to participate in a class action lawsuit.
              </p>
            </div>
          </div>
        </section>

        {/* 16. Severability */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4">16. Severability</h2>
          <p className="text-gray-400">
            If any provision of these Terms is found to be unenforceable, the remaining provisions will
            continue in full force and effect. The unenforceable provision will be modified to the minimum
            extent necessary to make it enforceable.
          </p>
        </section>

        {/* 17. Contact Information */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4">17. Contact Information</h2>
          <div className="glass rounded-xl p-6">
            <p className="text-gray-400 mb-4">
              For questions about these Terms of Service:
            </p>
            <div className="space-y-3 text-sm">
              <div className="flex items-center gap-3">
                <Mail className="w-5 h-5 text-gold-500" />
                <div>
                  <span className="text-gray-500 block">Legal inquiries:</span>
                  <a href="mailto:legal@aureo-vpn.com" className="text-gold-500 hover:underline">legal@aureo-vpn.com</a>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <Mail className="w-5 h-5 text-gold-500" />
                <div>
                  <span className="text-gray-500 block">Billing support:</span>
                  <a href="mailto:billing@aureo-vpn.com" className="text-gold-500 hover:underline">billing@aureo-vpn.com</a>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <Mail className="w-5 h-5 text-gold-500" />
                <div>
                  <span className="text-gray-500 block">General support:</span>
                  <a href="mailto:support@aureo-vpn.com" className="text-gold-500 hover:underline">support@aureo-vpn.com</a>
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
            <Link to="/privacy" className="glass px-4 py-2 rounded-lg text-gold-500 hover:bg-white/5 transition-all text-sm">
              Privacy Policy
            </Link>
            <Link to="/node-operators" className="glass px-4 py-2 rounded-lg text-gold-500 hover:bg-white/5 transition-all text-sm">
              Node Operator Guide
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
