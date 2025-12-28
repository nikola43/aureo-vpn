import React from 'react';
import PageLayout from '../components/PageLayout';
import { FileText, AlertTriangle, Check, X } from 'lucide-react';

export default function TermsOfService() {
  const lastUpdated = 'December 28, 2024';

  return (
    <PageLayout
      title="Terms of Service"
      description="Please read these terms carefully before using Aureo VPN."
    >
      <div className="glass rounded-xl p-8 mb-8">
        <div className="flex items-center gap-3 text-gold-500 mb-4">
          <FileText className="w-6 h-6" />
          <span className="font-bold">Terms and Conditions</span>
        </div>
        <p className="text-gray-400">
          Last updated: {lastUpdated}
        </p>
      </div>

      <div className="prose prose-invert max-w-none space-y-8">
        <section>
          <h2 className="text-2xl font-bold text-white mb-4">1. Acceptance of Terms</h2>
          <p className="text-gray-400 leading-relaxed">
            By accessing or using Aureo VPN services, you agree to be bound by these Terms of Service
            and our Privacy Policy. If you do not agree to these terms, please do not use our services.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4">2. Description of Service</h2>
          <p className="text-gray-400 mb-4">
            Aureo VPN provides:
          </p>
          <ul className="space-y-2 text-gray-400">
            <li className="flex items-start gap-3">
              <Check className="w-5 h-5 text-green-500 flex-shrink-0" />
              <span>Virtual Private Network (VPN) services for secure internet browsing</span>
            </li>
            <li className="flex items-start gap-3">
              <Check className="w-5 h-5 text-green-500 flex-shrink-0" />
              <span>Decentralized VPN infrastructure operated by community node operators</span>
            </li>
            <li className="flex items-start gap-3">
              <Check className="w-5 h-5 text-green-500 flex-shrink-0" />
              <span>Cryptocurrency-based reward system for node operators</span>
            </li>
            <li className="flex items-start gap-3">
              <Check className="w-5 h-5 text-green-500 flex-shrink-0" />
              <span>Client applications for various platforms</span>
            </li>
          </ul>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4">3. User Responsibilities</h2>
          <p className="text-gray-400 mb-4">
            As a user of Aureo VPN, you agree to:
          </p>
          <ul className="space-y-2 text-gray-400">
            <li className="flex items-start gap-3">
              <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
              <span>Comply with all applicable laws and regulations in your jurisdiction</span>
            </li>
            <li className="flex items-start gap-3">
              <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
              <span>Not use the service for illegal activities</span>
            </li>
            <li className="flex items-start gap-3">
              <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
              <span>Not attempt to compromise the security or integrity of the service</span>
            </li>
            <li className="flex items-start gap-3">
              <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
              <span>Keep your account credentials secure</span>
            </li>
          </ul>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-3">
            <AlertTriangle className="w-6 h-6 text-red-500" />
            4. Prohibited Activities
          </h2>
          <p className="text-gray-400 mb-4">
            The following activities are strictly prohibited:
          </p>
          <div className="glass rounded-xl p-6 space-y-3">
            {[
              'Distributing malware or engaging in cyberattacks',
              'Sending spam or engaging in phishing',
              'Distributing copyrighted content without authorization',
              'Child exploitation material (immediately reported to authorities)',
              'Harassment or threats against individuals',
              'Identity fraud or impersonation',
              'Attempting to circumvent VPN server security',
              'Reselling or sharing your account without authorization'
            ].map((item) => (
              <div key={item} className="flex items-start gap-3">
                <X className="w-5 h-5 text-red-500 flex-shrink-0" />
                <span className="text-gray-400">{item}</span>
              </div>
            ))}
          </div>
          <p className="text-gray-400 mt-4">
            Violation of these terms may result in immediate termination of your account without refund.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4">5. Node Operator Terms</h2>
          <p className="text-gray-400 mb-4">
            Node operators agree to additional terms:
          </p>
          <ul className="space-y-2 text-gray-400">
            <li className="flex items-start gap-3">
              <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
              <span>Maintain minimum uptime requirements for your tier</span>
            </li>
            <li className="flex items-start gap-3">
              <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
              <span>Not log or monitor user traffic passing through your node</span>
            </li>
            <li className="flex items-start gap-3">
              <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
              <span>Comply with the security guidelines provided</span>
            </li>
            <li className="flex items-start gap-3">
              <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
              <span>Accept cryptocurrency payments as the sole form of compensation</span>
            </li>
            <li className="flex items-start gap-3">
              <span className="w-2 h-2 rounded-full bg-gold-500 mt-2 flex-shrink-0"></span>
              <span>Provide accurate server location and capability information</span>
            </li>
          </ul>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4">6. Payment and Refunds</h2>
          <div className="glass rounded-xl p-6 space-y-4">
            <div>
              <h4 className="font-bold text-white mb-2">Cryptocurrency Payments</h4>
              <p className="text-gray-400 text-sm">
                All payments are made in cryptocurrency (ETH, BTC, LTC). Payments are final once confirmed
                on the blockchain. Exchange rate fluctuations are the responsibility of the user.
              </p>
            </div>
            <div>
              <h4 className="font-bold text-white mb-2">Refund Policy</h4>
              <p className="text-gray-400 text-sm">
                We offer a 30-day money-back guarantee for new users. Refunds are processed in the original
                cryptocurrency at current exchange rates. Refund requests must be submitted within 30 days
                of the initial payment.
              </p>
            </div>
            <div>
              <h4 className="font-bold text-white mb-2">Node Operator Payouts</h4>
              <p className="text-gray-400 text-sm">
                Operator earnings are paid out upon request, with a minimum threshold of $10 USD equivalent.
                Payouts are processed within 24 hours of request and subject to blockchain confirmation times.
              </p>
            </div>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4">7. Service Availability</h2>
          <p className="text-gray-400">
            While we strive for 99.9% uptime, we do not guarantee uninterrupted service. We may occasionally
            need to perform maintenance or updates. We are not responsible for service interruptions due to
            factors beyond our control, including but not limited to network outages, natural disasters, or
            government actions.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4">8. Intellectual Property</h2>
          <p className="text-gray-400">
            The Aureo VPN name, logo, and branding are trademarks of Aureo Network. Our open-source code is
            licensed under the MIT License. You may not use our trademarks without prior written permission.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4">9. Limitation of Liability</h2>
          <p className="text-gray-400">
            Aureo VPN is provided "as is" without warranty of any kind. We are not liable for any indirect,
            incidental, special, consequential, or punitive damages arising from your use of the service.
            Our total liability shall not exceed the amount you paid for the service in the past 12 months.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4">10. Changes to Terms</h2>
          <p className="text-gray-400">
            We reserve the right to modify these terms at any time. Significant changes will be announced
            via email and on our website. Continued use of the service after changes constitutes acceptance
            of the new terms.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4">11. Governing Law</h2>
          <p className="text-gray-400">
            These terms are governed by the laws of Switzerland. Any disputes shall be resolved in the courts
            of Zurich, Switzerland.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4">12. Contact Information</h2>
          <p className="text-gray-400">
            For questions about these Terms of Service, please contact us at:<br />
            <a href="mailto:legal@aureo-vpn.com" className="text-gold-500 hover:underline">legal@aureo-vpn.com</a>
          </p>
        </section>
      </div>
    </PageLayout>
  );
}
