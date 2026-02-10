import React, { useState } from 'react';
import { motion } from 'framer-motion';
import PageLayout from '../components/PageLayout';
import {
  Mail,
  MessageCircle,
  Github,
  Twitter,
  MapPin,
  Send,
  HelpCircle,
  Bug,
  Briefcase
} from 'lucide-react';

const contactMethods = [
  {
    icon: Mail,
    title: 'Email Support',
    description: 'Get help from our support team',
    contact: 'support@aureo-vpn.com',
    link: 'mailto:support@aureo-vpn.com'
  },
  {
    icon: MessageCircle,
    title: 'Discord Community',
    description: 'Join our community chat',
    contact: 'discord.gg/aureo-vpn',
    link: 'https://discord.gg/aureo-vpn'
  },
  {
    icon: Github,
    title: 'GitHub Issues',
    description: 'Report bugs and feature requests',
    contact: 'github.com/nikola43/aureo-vpn',
    link: 'https://github.com/nikola43/aureo-vpn/issues'
  },
  {
    icon: Twitter,
    title: 'Twitter',
    description: 'Follow us for updates',
    contact: '@AureoVPN',
    link: 'https://twitter.com/AureoVPN'
  },
];

const departments = [
  { icon: HelpCircle, label: 'General Support', email: 'support@aureo-vpn.com' },
  { icon: Bug, label: 'Bug Reports', email: 'bugs@aureo-vpn.com' },
  { icon: Briefcase, label: 'Business Inquiries', email: 'business@aureo-vpn.com' },
  { icon: Mail, label: 'Press & Media', email: 'press@aureo-vpn.com' },
];

export default function Contact() {
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    subject: '',
    message: ''
  });

  const handleSubmit = (e) => {
    e.preventDefault();
    // Handle form submission
    console.log('Form submitted:', formData);
    alert('Thank you for your message! We will get back to you soon.');
  };

  return (
    <PageLayout
      title="Contact Us"
      description="Have questions or need support? We're here to help."
    >
      <div className="grid lg:grid-cols-2 gap-12">
        {/* Contact Form */}
        <div className="glass rounded-xl p-8">
          <h2 className="text-2xl font-bold text-white mb-6">Send us a Message</h2>
          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="grid md:grid-cols-2 gap-6">
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-2">Name</label>
                <input
                  type="text"
                  required
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="w-full px-4 py-3 bg-dark-800 border border-white/10 rounded-lg text-white focus:border-gold-500 focus:outline-none transition-colors"
                  placeholder="Your name"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-2">Email</label>
                <input
                  type="email"
                  required
                  value={formData.email}
                  onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                  className="w-full px-4 py-3 bg-dark-800 border border-white/10 rounded-lg text-white focus:border-gold-500 focus:outline-none transition-colors"
                  placeholder="your@email.com"
                />
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-2">Subject</label>
              <select
                value={formData.subject}
                onChange={(e) => setFormData({ ...formData, subject: e.target.value })}
                className="w-full px-4 py-3 bg-dark-800 border border-white/10 rounded-lg text-white focus:border-gold-500 focus:outline-none transition-colors"
              >
                <option value="">Select a topic</option>
                <option value="support">Technical Support</option>
                <option value="billing">Billing Question</option>
                <option value="operator">Node Operator Inquiry</option>
                <option value="business">Business Partnership</option>
                <option value="other">Other</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-2">Message</label>
              <textarea
                required
                rows={5}
                value={formData.message}
                onChange={(e) => setFormData({ ...formData, message: e.target.value })}
                className="w-full px-4 py-3 bg-dark-800 border border-white/10 rounded-lg text-white focus:border-gold-500 focus:outline-none transition-colors resize-none"
                placeholder="How can we help you?"
              />
            </div>
            <button
              type="submit"
              className="w-full px-6 py-3 bg-gold-500 text-dark-950 font-bold rounded-lg hover:bg-gold-400 transition-all flex items-center justify-center gap-2"
            >
              <Send className="w-5 h-5" />
              Send Message
            </button>
          </form>
        </div>

        {/* Contact Methods */}
        <div className="space-y-6">
          <h2 className="text-2xl font-bold text-white mb-6">Other Ways to Reach Us</h2>

          {contactMethods.map((method, index) => (
            <motion.a
              key={method.title}
              href={method.link}
              target="_blank"
              rel="noopener noreferrer"
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: index * 0.1 }}
              className="glass rounded-xl p-6 flex items-start gap-4 hover:border-gold-500/30 border border-transparent transition-all block"
            >
              <div className="w-12 h-12 rounded-xl bg-gold-500/10 flex items-center justify-center flex-shrink-0">
                <method.icon className="w-6 h-6 text-gold-500" />
              </div>
              <div>
                <h3 className="font-bold text-white">{method.title}</h3>
                <p className="text-gray-400 text-sm mb-1">{method.description}</p>
                <span className="text-gold-500 text-sm">{method.contact}</span>
              </div>
            </motion.a>
          ))}

          {/* Departments */}
          <div className="glass rounded-xl p-6 mt-8">
            <h3 className="font-bold text-white mb-4">Department Emails</h3>
            <div className="space-y-3">
              {departments.map((dept) => (
                <div key={dept.label} className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <dept.icon className="w-5 h-5 text-gray-400" />
                    <span className="text-gray-300">{dept.label}</span>
                  </div>
                  <a href={`mailto:${dept.email}`} className="text-gold-500 text-sm hover:underline">
                    {dept.email}
                  </a>
                </div>
              ))}
            </div>
          </div>

          {/* Response Time */}
          <div className="glass rounded-xl p-6">
            <h3 className="font-bold text-white mb-2">Response Times</h3>
            <p className="text-gray-400 text-sm mb-4">
              We typically respond within 24-48 hours. Premium users receive priority support.
            </p>
            <div className="flex gap-4">
              <div className="flex-1 text-center p-3 bg-dark-800 rounded-lg">
                <span className="text-gold-500 font-bold text-lg">24h</span>
                <span className="text-gray-500 text-sm block">Premium</span>
              </div>
              <div className="flex-1 text-center p-3 bg-dark-800 rounded-lg">
                <span className="text-white font-bold text-lg">48h</span>
                <span className="text-gray-500 text-sm block">Standard</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </PageLayout>
  );
}
