import React, { useState, useEffect } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Shield,
  Menu,
  X,
  ArrowRight,
  ArrowLeft,
  Github,
  Twitter,
  MessageCircle,
  Mail
} from 'lucide-react';

function Navigation({ scrolled }) {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const location = useLocation();

  return (
    <nav className={`fixed w-full z-50 transition-all duration-300 ${scrolled ? 'bg-dark-950/90 backdrop-blur-xl border-b border-white/5 py-3' : 'bg-dark-950/80 backdrop-blur-sm py-5'}`}>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between">
          <Link to="/" className="flex items-center gap-3 group">
            <div className="relative">
              <div className="absolute inset-0 bg-gold-500/20 blur-xl rounded-full group-hover:bg-gold-500/30 transition-all" />
              <Shield className="h-9 w-9 text-gold-500 relative" />
            </div>
            <span className="font-bold text-2xl tracking-tight">
              <span className="text-white">Aureo</span>
              <span className="text-gold-500">VPN</span>
            </span>
          </Link>

          <div className="hidden lg:flex items-center gap-1">
            {[
              { label: 'Features', path: '/features' },
              { label: 'Documentation', path: '/docs' },
              { label: 'API', path: '/api' },
              { label: 'Node Operators', path: '/node-operators' },
            ].map((item) => (
              <Link
                key={item.path}
                to={item.path}
                className={`px-4 py-2 rounded-lg transition-all font-medium ${
                  location.pathname === item.path
                    ? 'text-gold-500 bg-gold-500/10'
                    : 'text-gray-300 hover:text-white hover:bg-white/5'
                }`}
              >
                {item.label}
              </Link>
            ))}
          </div>

          <div className="hidden lg:flex items-center gap-3">
            <Link to="/" className="px-5 py-2.5 text-gray-300 hover:text-white font-medium transition-all flex items-center gap-2">
              <ArrowLeft className="w-4 h-4" />
              Back to Home
            </Link>
            <motion.a
              href="/register"
              className="relative group px-6 py-2.5 bg-gradient-to-r from-gold-500 to-gold-600 text-dark-950 font-bold rounded-full overflow-hidden"
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.98 }}
            >
              <span className="relative z-10 flex items-center gap-2">
                Get Started
                <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
              </span>
            </motion.a>
          </div>

          <button
            className="lg:hidden p-2 text-gray-400 hover:text-white"
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
          >
            {mobileMenuOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
          </button>
        </div>
      </div>

      <AnimatePresence>
        {mobileMenuOpen && (
          <motion.div
            initial={{ opacity: 0, y: -20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            className="lg:hidden absolute top-full left-0 right-0 bg-dark-950/95 backdrop-blur-xl border-b border-white/5 p-4"
          >
            <div className="flex flex-col gap-2">
              {[
                { label: 'Features', path: '/features' },
                { label: 'Documentation', path: '/docs' },
                { label: 'API Reference', path: '/api' },
                { label: 'Node Operators', path: '/node-operators' },
              ].map((item) => (
                <Link
                  key={item.path}
                  to={item.path}
                  className="px-4 py-3 text-gray-300 hover:text-white hover:bg-white/5 rounded-lg transition-all font-medium"
                  onClick={() => setMobileMenuOpen(false)}
                >
                  {item.label}
                </Link>
              ))}
              <hr className="border-white/10 my-2" />
              <Link to="/" className="px-4 py-3 text-gray-300 hover:text-white font-medium flex items-center gap-2">
                <ArrowLeft className="w-4 h-4" />
                Back to Home
              </Link>
              <a href="/register" className="px-4 py-3 bg-gold-500 text-dark-950 font-bold rounded-lg text-center">
                Get Started
              </a>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </nav>
  );
}

function Footer() {
  const footerLinks = {
    Product: [
      { label: 'Features', path: '/features' },
      { label: 'Pricing', path: '/#pricing' },
      { label: 'Node Operators', path: '/node-operators' },
      { label: 'Downloads', path: '/downloads' },
    ],
    Resources: [
      { label: 'Documentation', path: '/docs' },
      { label: 'API Reference', path: '/api' },
      { label: 'Status', path: '/status' },
    ],
    Company: [
      { label: 'About', path: '/about' },
      { label: 'Contact', path: '/contact' },
    ],
    Legal: [
      { label: 'Privacy Policy', path: '/privacy' },
      { label: 'Terms of Service', path: '/terms' },
    ]
  };

  return (
    <footer className="relative bg-dark-950 border-t border-white/5 pt-16 pb-8">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="grid grid-cols-2 md:grid-cols-6 gap-8 mb-12">
          <div className="col-span-2">
            <Link to="/" className="flex items-center gap-3 mb-4">
              <Shield className="h-8 w-8 text-gold-500" />
              <span className="font-bold text-xl">
                <span className="text-white">Aureo</span>
                <span className="text-gold-500">VPN</span>
              </span>
            </Link>
            <p className="text-gray-400 mb-6 max-w-sm">
              The future of decentralized privacy. Secure browsing and crypto rewards for the privacy-conscious.
            </p>
            <div className="flex gap-4">
              {[Twitter, Github, MessageCircle, Mail].map((Icon, i) => (
                <a
                  key={i}
                  href="#"
                  className="w-10 h-10 glass rounded-lg flex items-center justify-center hover:bg-white/10 transition-all"
                >
                  <Icon className="w-5 h-5 text-gray-400 hover:text-white" />
                </a>
              ))}
            </div>
          </div>

          {Object.entries(footerLinks).map(([category, links]) => (
            <div key={category}>
              <h4 className="font-bold text-white mb-4">{category}</h4>
              <ul className="space-y-3">
                {links.map((link) => (
                  <li key={link.path}>
                    <Link
                      to={link.path}
                      className="text-gray-400 hover:text-white transition-colors text-sm"
                    >
                      {link.label}
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        <div className="border-t border-white/5 pt-8 flex flex-col md:flex-row justify-between items-center gap-4">
          <div className="text-gray-500 text-sm">
            &copy; {new Date().getFullYear()} Aureo Network. All rights reserved.
          </div>
          <div className="flex items-center gap-6 text-sm text-gray-500">
            <span className="flex items-center gap-2">
              <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
              All systems operational
            </span>
            <span>v2.0.0</span>
          </div>
        </div>
      </div>
    </footer>
  );
}

export default function PageLayout({ children, title, description }) {
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const handleScroll = () => setScrolled(window.scrollY > 50);
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  useEffect(() => {
    window.scrollTo(0, 0);
  }, []);

  return (
    <div className="min-h-screen bg-dark-950 text-white">
      <Navigation scrolled={scrolled} />

      {/* Page Header */}
      <div className="pt-32 pb-16 relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-b from-gold-500/5 via-transparent to-transparent" />
        <div className="absolute inset-0 bg-grid-pattern bg-grid opacity-5" />
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
          >
            <h1 className="text-4xl md:text-5xl font-black mb-4">
              <span className="text-gradient">{title}</span>
            </h1>
            {description && (
              <p className="text-xl text-gray-400 max-w-2xl">{description}</p>
            )}
          </motion.div>
        </div>
      </div>

      {/* Page Content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pb-20">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.1 }}
        >
          {children}
        </motion.div>
      </div>

      <Footer />
    </div>
  );
}
