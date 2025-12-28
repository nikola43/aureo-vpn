import React from 'react';
import PageLayout from '../components/PageLayout';
import { Shield, Eye, Database, Lock, Users, Globe } from 'lucide-react';

export default function PrivacyPolicy() {
  const lastUpdated = 'December 28, 2024';

  return (
    <PageLayout
      title="Privacy Policy"
      description="How we protect your privacy and handle your data."
    >
      <div className="glass rounded-xl p-8 mb-8">
        <div className="flex items-center gap-3 text-gold-500 mb-4">
          <Shield className="w-6 h-6" />
          <span className="font-bold">Your Privacy Matters</span>
        </div>
        <p className="text-gray-400">
          Last updated: {lastUpdated}
        </p>
      </div>

      <div className="prose prose-invert max-w-none space-y-8">
        <section>
          <h2 className="text-2xl font-bold text-white mb-4">Overview</h2>
          <p className="text-gray-400 leading-relaxed">
            Aureo VPN ("we", "our", or "us") is committed to protecting your privacy. This Privacy Policy
            explains how we collect, use, and safeguard your information when you use our VPN service,
            website, and related applications.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <Eye className="w-6 h-6 text-gold-500" />
            What We DON'T Collect
          </h2>
          <p className="text-gray-400 mb-4">
            As a privacy-first VPN service, we have a strict no-logs policy. We do NOT collect:
          </p>
          <ul className="space-y-2 text-gray-400">
            <li className="flex items-start gap-3">
              <span className="w-2 h-2 rounded-full bg-green-500 mt-2 flex-shrink-0"></span>
              <span>Your IP address when connected to our VPN servers</span>
            </li>
            <li className="flex items-start gap-3">
              <span className="w-2 h-2 rounded-full bg-green-500 mt-2 flex-shrink-0"></span>
              <span>Your browsing history or the websites you visit</span>
            </li>
            <li className="flex items-start gap-3">
              <span className="w-2 h-2 rounded-full bg-green-500 mt-2 flex-shrink-0"></span>
              <span>DNS queries made through our VPN</span>
            </li>
            <li className="flex items-start gap-3">
              <span className="w-2 h-2 rounded-full bg-green-500 mt-2 flex-shrink-0"></span>
              <span>Connection timestamps</span>
            </li>
            <li className="flex items-start gap-3">
              <span className="w-2 h-2 rounded-full bg-green-500 mt-2 flex-shrink-0"></span>
              <span>Bandwidth usage per session</span>
            </li>
            <li className="flex items-start gap-3">
              <span className="w-2 h-2 rounded-full bg-green-500 mt-2 flex-shrink-0"></span>
              <span>Session duration per connection</span>
            </li>
          </ul>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <Database className="w-6 h-6 text-gold-500" />
            What We DO Collect
          </h2>
          <p className="text-gray-400 mb-4">
            To provide our service, we collect minimal information necessary:
          </p>

          <div className="glass rounded-xl p-6 space-y-6">
            <div>
              <h4 className="font-bold text-white mb-2">Account Information</h4>
              <p className="text-gray-400 text-sm">
                Email address (for account recovery and communication), username, and encrypted password hash.
                For anonymous accounts, no email is required.
              </p>
            </div>

            <div>
              <h4 className="font-bold text-white mb-2">Payment Information</h4>
              <p className="text-gray-400 text-sm">
                For cryptocurrency payments, we only store the transaction hash and payment status.
                We do not store your wallet address longer than necessary to process the payment.
              </p>
            </div>

            <div>
              <h4 className="font-bold text-white mb-2">Aggregate Statistics</h4>
              <p className="text-gray-400 text-sm">
                We collect anonymized, aggregate data about server load and network performance to improve
                our service. This data cannot be linked to individual users.
              </p>
            </div>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <Lock className="w-6 h-6 text-gold-500" />
            Data Security
          </h2>
          <p className="text-gray-400 mb-4">
            We implement industry-standard security measures to protect your data:
          </p>
          <ul className="space-y-2 text-gray-400">
            <li className="flex items-start gap-3">
              <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
              <span>All data is encrypted at rest using AES-256</span>
            </li>
            <li className="flex items-start gap-3">
              <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
              <span>Passwords are hashed using bcrypt with salt</span>
            </li>
            <li className="flex items-start gap-3">
              <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
              <span>All API communications use TLS 1.3</span>
            </li>
            <li className="flex items-start gap-3">
              <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
              <span>Regular security audits by third parties</span>
            </li>
            <li className="flex items-start gap-3">
              <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
              <span>RAM-only servers where possible (no persistent storage)</span>
            </li>
          </ul>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <Users className="w-6 h-6 text-gold-500" />
            Node Operators
          </h2>
          <p className="text-gray-400 mb-4">
            For node operators, we collect additional information necessary for the reward system:
          </p>
          <ul className="space-y-2 text-gray-400">
            <li className="flex items-start gap-3">
              <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
              <span>Cryptocurrency wallet address for payouts</span>
            </li>
            <li className="flex items-start gap-3">
              <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
              <span>Node performance metrics (uptime, bandwidth served)</span>
            </li>
            <li className="flex items-start gap-3">
              <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
              <span>Server location (country/city) for node discovery</span>
            </li>
          </ul>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <Globe className="w-6 h-6 text-gold-500" />
            Third-Party Services
          </h2>
          <p className="text-gray-400 mb-4">
            We may use the following third-party services:
          </p>
          <div className="glass rounded-xl p-6 space-y-4">
            <div>
              <h4 className="font-bold text-white mb-1">Analytics</h4>
              <p className="text-gray-400 text-sm">
                We use privacy-focused analytics (Plausible) that don't use cookies or collect personal data.
              </p>
            </div>
            <div>
              <h4 className="font-bold text-white mb-1">Payment Processors</h4>
              <p className="text-gray-400 text-sm">
                Cryptocurrency payments are processed on respective blockchains. We don't use traditional payment processors.
              </p>
            </div>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4">Your Rights</h2>
          <p className="text-gray-400 mb-4">
            You have the right to:
          </p>
          <ul className="space-y-2 text-gray-400">
            <li className="flex items-start gap-3">
              <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
              <span>Access your personal data</span>
            </li>
            <li className="flex items-start gap-3">
              <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
              <span>Request deletion of your account and data</span>
            </li>
            <li className="flex items-start gap-3">
              <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
              <span>Export your data in a portable format</span>
            </li>
            <li className="flex items-start gap-3">
              <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
              <span>Opt out of marketing communications</span>
            </li>
          </ul>
          <p className="text-gray-400 mt-4">
            To exercise these rights, contact us at <a href="mailto:privacy@aureo-vpn.com" className="text-gold-500 hover:underline">privacy@aureo-vpn.com</a>.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4">Changes to This Policy</h2>
          <p className="text-gray-400">
            We may update this Privacy Policy from time to time. We will notify you of any significant changes
            by posting the new policy on this page and updating the "Last updated" date. We encourage you to
            review this policy periodically.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4">Contact Us</h2>
          <p className="text-gray-400">
            If you have questions about this Privacy Policy, please contact us at:<br />
            <a href="mailto:privacy@aureo-vpn.com" className="text-gold-500 hover:underline">privacy@aureo-vpn.com</a>
          </p>
        </section>
      </div>
    </PageLayout>
  );
}
