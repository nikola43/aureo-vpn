import React from 'react';
import { motion } from 'framer-motion';
import { Shield, Coins, Globe, Server, Activity, Lock, ArrowRight, CheckCircle2 } from 'lucide-react';

function App() {
  return (
    <div className="min-h-screen bg-gray-900 text-white selection:bg-gold-500 selection:text-white">
      {/* Navigation */}
      <nav className="fixed w-full z-50 bg-gray-900/80 backdrop-blur-sm border-b border-gray-800">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center gap-2">
              <Shield className="h-8 w-8 text-gold-500" />
              <span className="font-bold text-xl tracking-tight">Aureo VPN</span>
            </div>
            <div className="hidden md:block">
              <div className="ml-10 flex items-baseline space-x-8">
                <a href="#features" className="hover:text-gold-400 transition-colors px-3 py-2 rounded-md text-sm font-medium">Features</a>
                <a href="#how-it-works" className="hover:text-gold-400 transition-colors px-3 py-2 rounded-md text-sm font-medium">How it Works</a>
                <a href="#earnings" className="hover:text-gold-400 transition-colors px-3 py-2 rounded-md text-sm font-medium">Earnings</a>
                <a href="/login" className="bg-gold-600 hover:bg-gold-500 text-white px-4 py-2 rounded-full text-sm font-bold transition-all shadow-lg shadow-gold-600/20">
                  Launch App
                </a>
              </div>
            </div>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <section className="relative pt-32 pb-20 lg:pt-48 lg:pb-32 overflow-hidden">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
          <div className="text-center max-w-4xl mx-auto">
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5 }}
            >
              <h1 className="text-5xl md:text-7xl font-extrabold tracking-tight mb-8 bg-clip-text text-transparent bg-gradient-to-r from-white via-gray-200 to-gray-400">
                The Future of <span className="text-gold-500">Decentralized</span> Privacy
              </h1>
              <p className="mt-4 text-xl text-gray-400 mb-12 max-w-2xl mx-auto leading-relaxed">
                Aureo VPN combines military-grade encryption with a decentralized node network. 
                Browse securely or earn crypto by becoming a node operator.
              </p>
              <div className="flex flex-col sm:flex-row justify-center gap-4">
                <a href="/register" className="inline-flex items-center justify-center px-8 py-4 border border-transparent text-lg font-bold rounded-full text-gray-900 bg-gold-500 hover:bg-gold-400 transition-all shadow-lg shadow-gold-500/20">
                  Get Started <ArrowRight className="ml-2 h-5 w-5" />
                </a>
                <a href="#learn-more" className="inline-flex items-center justify-center px-8 py-4 border border-gray-700 text-lg font-medium rounded-full text-gray-300 hover:bg-gray-800 transition-all">
                  Become an Operator
                </a>
              </div>
            </motion.div>
          </div>
        </div>
        
        {/* Background Gradients */}
        <div className="absolute top-0 left-1/2 w-full -translate-x-1/2 h-full z-0 pointer-events-none">
           <div className="absolute top-0 left-1/4 w-96 h-96 bg-gold-600/20 rounded-full blur-3xl" />
           <div className="absolute bottom-0 right-1/4 w-96 h-96 bg-blue-600/10 rounded-full blur-3xl" />
        </div>
      </section>

      {/* Stats Section */}
      <div className="bg-gray-800/50 border-y border-gray-800 backdrop-blur-sm">
        <div className="max-w-7xl mx-auto py-12 px-4 sm:px-6 lg:px-8">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-8 text-center">
            <StatCard number="5,000+" label="Active Nodes" />
            <StatCard number="60+" label="Countries" />
            <StatCard number="10Gbps" label="Max Speed" />
            <StatCard number="$0.10/GB" label="Operator Earnings" />
          </div>
        </div>
      </div>

      {/* Features Section */}
      <section id="features" className="py-24 bg-gray-900">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-base text-gold-500 font-semibold tracking-wide uppercase">Features</h2>
            <p className="mt-2 text-3xl leading-8 font-extrabold tracking-tight text-white sm:text-4xl">
              Why Choose Aureo?
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-12">
            <FeatureCard 
              icon={<Globe className="h-8 w-8 text-gold-500" />}
              title="Decentralized Network"
              description="No central points of failure. Our network is powered by thousands of independent node operators worldwide."
            />
            <FeatureCard 
              icon={<Coins className="h-8 w-8 text-gold-500" />}
              title="Earn Crypto"
              description="Turn your unused bandwidth into passive income. Run a node and get paid in ETH, BTC, or LTC."
            />
            <FeatureCard 
              icon={<Lock className="h-8 w-8 text-gold-500" />}
              title="Zero Logs"
              description="We couldn't track you if we wanted to. Our architecture ensures complete anonymity by design."
            />
          </div>
        </div>
      </section>

      {/* How it Works Section */}
      <section id="how-it-works" className="py-24 bg-gray-800/30 relative overflow-hidden">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
          <div className="lg:grid lg:grid-cols-2 lg:gap-24 items-center">
            <div>
              <h2 className="text-3xl font-extrabold text-white sm:text-4xl mb-6">
                Run a Node. <br/>
                <span className="text-gold-500">Earn Passive Income.</span>
              </h2>
              <p className="text-lg text-gray-400 mb-8">
                Join the Aureo network as a node operator. It takes less than 5 minutes to set up, and you can start earning immediately.
              </p>
              
              <div className="space-y-6">
                <Step 
                  number="01"
                  title="Register as Operator"
                  description="Create an account and connect your crypto wallet."
                />
                <Step 
                  number="02"
                  title="Deploy Node"
                  description="Run our simple Docker container on your server or VPS."
                />
                <Step 
                  number="03"
                  title="Get Paid"
                  description="Receive automatic payouts for every GB of traffic routed."
                />
              </div>
            </div>
            <div className="mt-12 lg:mt-0 relative">
               <div className="bg-gray-900 rounded-2xl p-8 border border-gray-700 shadow-2xl relative">
                  <div className="flex items-center justify-between mb-8 border-b border-gray-800 pb-4">
                    <div>
                      <h3 className="text-lg font-bold text-white">Earnings Dashboard</h3>
                      <p className="text-sm text-gray-400">Live network stats</p>
                    </div>
                    <div className="bg-green-500/20 text-green-400 px-3 py-1 rounded-full text-xs font-bold flex items-center gap-1">
                      <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
                      ONLINE
                    </div>
                  </div>
                  
                  <div className="space-y-6">
                    <div className="bg-gray-800 rounded-lg p-4">
                       <div className="flex justify-between items-end mb-2">
                          <span className="text-gray-400 text-sm">Total Earned</span>
                          <span className="text-gold-500 text-2xl font-mono font-bold">$1,245.50</span>
                       </div>
                       <div className="w-full bg-gray-700 rounded-full h-1.5">
                          <div className="bg-gold-500 h-1.5 rounded-full" style={{width: '75%'}}></div>
                       </div>
                    </div>
                    
                    <div className="grid grid-cols-2 gap-4">
                       <div className="bg-gray-800 rounded-lg p-4">
                          <div className="text-gray-400 text-xs mb-1">Bandwidth</div>
                          <div className="text-white font-mono font-bold">4.2 TB</div>
                       </div>
                       <div className="bg-gray-800 rounded-lg p-4">
                          <div className="text-gray-400 text-xs mb-1">Uptime</div>
                          <div className="text-white font-mono font-bold">99.9%</div>
                       </div>
                    </div>
                  </div>
               </div>
               {/* Decorative elements */}
               <div className="absolute -z-10 top-10 -right-10 w-72 h-72 bg-gold-600/10 rounded-full blur-3xl" />
            </div>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-20 bg-gold-600 relative overflow-hidden">
        <div className="max-w-4xl mx-auto text-center px-4 sm:px-6 lg:px-8 relative z-10">
          <h2 className="text-3xl font-extrabold text-white sm:text-4xl mb-6">
            Ready to reclaim your privacy?
          </h2>
          <p className="text-xl text-gold-100 mb-10">
            Join thousands of users who trust Aureo for secure, decentralized internet access.
          </p>
          <a href="/register" className="inline-flex items-center justify-center px-8 py-4 border border-transparent text-lg font-bold rounded-full text-gold-600 bg-white hover:bg-gray-50 transition-all shadow-xl">
            Get Started for Free
          </a>
        </div>
        <div className="absolute inset-0 bg-pattern opacity-10" />
      </section>

      {/* Footer */}
      <footer className="bg-gray-900 border-t border-gray-800 py-12">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex flex-col md:flex-row justify-between items-center">
            <div className="flex items-center gap-2 mb-4 md:mb-0">
              <Shield className="h-6 w-6 text-gray-600" />
              <span className="font-bold text-gray-500">Aureo VPN</span>
            </div>
            <div className="flex space-x-6 text-gray-400 text-sm">
              <a href="#" className="hover:text-white transition-colors">Privacy Policy</a>
              <a href="#" className="hover:text-white transition-colors">Terms of Service</a>
              <a href="#" className="hover:text-white transition-colors">Documentation</a>
              <a href="#" className="hover:text-white transition-colors">GitHub</a>
            </div>
            <div className="mt-4 md:mt-0 text-gray-600 text-sm">
              &copy; 2024 Aureo Network. All rights reserved.
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}

function FeatureCard({ icon, title, description }) {
  return (
    <motion.div 
      whileHover={{ y: -5 }}
      className="bg-gray-800 rounded-xl p-8 border border-gray-700 hover:border-gold-500/50 transition-colors"
    >
      <div className="bg-gray-900 w-16 h-16 rounded-full flex items-center justify-center mb-6">
        {icon}
      </div>
      <h3 className="text-xl font-bold text-white mb-3">{title}</h3>
      <p className="text-gray-400 leading-relaxed">{description}</p>
    </motion.div>
  );
}

function StatCard({ number, label }) {
  return (
    <div>
      <div className="text-4xl font-extrabold text-white mb-2">{number}</div>
      <div className="text-sm font-medium text-gray-400 uppercase tracking-wider">{label}</div>
    </div>
  );
}

function Step({ number, title, description }) {
  return (
    <div className="flex gap-4">
      <div className="flex-shrink-0">
        <div className="w-10 h-10 rounded-full bg-gray-800 border border-gray-700 flex items-center justify-center font-mono font-bold text-gold-500">
          {number}
        </div>
      </div>
      <div>
        <h3 className="text-lg font-bold text-white mb-1">{title}</h3>
        <p className="text-gray-400">{description}</p>
      </div>
    </div>
  );
}

export default App;
