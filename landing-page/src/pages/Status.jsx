import React from 'react';
import { motion } from 'framer-motion';
import PageLayout from '../components/PageLayout';
import {
  CheckCircle2,
  AlertCircle,
  XCircle,
  Activity,
  Server,
  Globe,
  Zap,
  Clock
} from 'lucide-react';

const services = [
  { name: 'API Gateway', status: 'operational', latency: '12ms' },
  { name: 'Authentication Service', status: 'operational', latency: '8ms' },
  { name: 'VPN Nodes (US)', status: 'operational', latency: '15ms' },
  { name: 'VPN Nodes (EU)', status: 'operational', latency: '22ms' },
  { name: 'VPN Nodes (Asia)', status: 'operational', latency: '45ms' },
  { name: 'P2P Network', status: 'operational', latency: '35ms' },
  { name: 'Blockchain Service', status: 'operational', latency: '120ms' },
  { name: 'Website', status: 'operational', latency: '5ms' },
];

const incidents = [
  {
    date: 'December 20, 2024',
    title: 'Scheduled Maintenance Completed',
    description: 'Routine database maintenance completed successfully. No downtime.',
    status: 'resolved'
  },
  {
    date: 'December 15, 2024',
    title: 'Network Upgrade',
    description: 'Upgraded all EU nodes to WireGuard 2.0. Performance improvements across the board.',
    status: 'resolved'
  },
  {
    date: 'December 1, 2024',
    title: 'API Rate Limiting Adjustment',
    description: 'Increased rate limits for authenticated users from 500 to 1000 requests/minute.',
    status: 'resolved'
  }
];

const stats = [
  { label: 'Uptime (30 days)', value: '99.97%', icon: Activity },
  { label: 'Active Nodes', value: '5,234', icon: Server },
  { label: 'Countries', value: '60+', icon: Globe },
  { label: 'Avg Response Time', value: '18ms', icon: Zap },
];

export default function Status() {
  const getStatusIcon = (status) => {
    switch (status) {
      case 'operational':
        return <CheckCircle2 className="w-5 h-5 text-green-500" />;
      case 'degraded':
        return <AlertCircle className="w-5 h-5 text-yellow-500" />;
      case 'outage':
        return <XCircle className="w-5 h-5 text-red-500" />;
      default:
        return <CheckCircle2 className="w-5 h-5 text-green-500" />;
    }
  };

  const getStatusColor = (status) => {
    switch (status) {
      case 'operational':
        return 'text-green-500';
      case 'degraded':
        return 'text-yellow-500';
      case 'outage':
        return 'text-red-500';
      default:
        return 'text-green-500';
    }
  };

  const allOperational = services.every(s => s.status === 'operational');

  return (
    <PageLayout
      title="System Status"
      description="Real-time status of Aureo VPN services and infrastructure."
    >
      {/* Overall Status */}
      <div className={`glass rounded-xl p-8 mb-8 border-2 ${allOperational ? 'border-green-500/30 bg-green-500/5' : 'border-yellow-500/30 bg-yellow-500/5'}`}>
        <div className="flex items-center gap-4">
          {allOperational ? (
            <CheckCircle2 className="w-12 h-12 text-green-500" />
          ) : (
            <AlertCircle className="w-12 h-12 text-yellow-500" />
          )}
          <div>
            <h2 className={`text-2xl font-bold ${allOperational ? 'text-green-500' : 'text-yellow-500'}`}>
              {allOperational ? 'All Systems Operational' : 'Partial Outage'}
            </h2>
            <p className="text-gray-400">
              {allOperational
                ? 'All Aureo VPN services are running smoothly.'
                : 'Some services are experiencing issues.'}
            </p>
          </div>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
        {stats.map((stat, index) => (
          <motion.div
            key={stat.label}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: index * 0.1 }}
            className="glass rounded-xl p-6 text-center"
          >
            <stat.icon className="w-8 h-8 text-gold-500 mx-auto mb-3" />
            <div className="text-2xl font-bold text-white">{stat.value}</div>
            <div className="text-gray-400 text-sm">{stat.label}</div>
          </motion.div>
        ))}
      </div>

      {/* Service Status */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold text-white mb-6">Service Status</h2>
        <div className="glass rounded-xl overflow-hidden">
          {services.map((service, index) => (
            <div
              key={service.name}
              className={`flex items-center justify-between px-6 py-4 ${index !== services.length - 1 ? 'border-b border-white/5' : ''}`}
            >
              <div className="flex items-center gap-4">
                {getStatusIcon(service.status)}
                <span className="text-white font-medium">{service.name}</span>
              </div>
              <div className="flex items-center gap-6">
                <span className="text-gray-400 text-sm hidden md:block">
                  <Clock className="w-4 h-4 inline mr-1" />
                  {service.latency}
                </span>
                <span className={`capitalize font-medium ${getStatusColor(service.status)}`}>
                  {service.status}
                </span>
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* Uptime Graph Placeholder */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold text-white mb-6">90-Day Uptime</h2>
        <div className="glass rounded-xl p-6">
          <div className="flex items-end gap-1 h-24 mb-4">
            {Array.from({ length: 90 }).map((_, i) => {
              const height = 85 + Math.random() * 15;
              return (
                <div
                  key={i}
                  className="flex-1 bg-green-500 rounded-t"
                  style={{ height: `${height}%` }}
                  title={`Day ${90 - i}: ${height.toFixed(1)}%`}
                />
              );
            })}
          </div>
          <div className="flex justify-between text-sm text-gray-400">
            <span>90 days ago</span>
            <span>Today</span>
          </div>
        </div>
      </section>

      {/* Recent Incidents */}
      <section>
        <h2 className="text-2xl font-bold text-white mb-6">Recent Updates</h2>
        <div className="space-y-4">
          {incidents.map((incident) => (
            <div key={incident.title} className="glass rounded-xl p-6">
              <div className="flex items-start justify-between mb-2">
                <h3 className="font-bold text-white">{incident.title}</h3>
                <span className="text-green-500 text-sm capitalize flex items-center gap-1">
                  <CheckCircle2 className="w-4 h-4" />
                  {incident.status}
                </span>
              </div>
              <p className="text-gray-400 text-sm mb-2">{incident.description}</p>
              <span className="text-gray-500 text-xs">{incident.date}</span>
            </div>
          ))}
        </div>
      </section>

      {/* Subscribe */}
      <div className="mt-12 glass rounded-xl p-6 text-center">
        <h3 className="font-bold text-white mb-2">Get Status Updates</h3>
        <p className="text-gray-400 text-sm mb-4">
          Subscribe to receive notifications about service status changes.
        </p>
        <div className="flex max-w-md mx-auto gap-3">
          <input
            type="email"
            placeholder="your@email.com"
            className="flex-1 px-4 py-2 bg-dark-800 border border-white/10 rounded-lg text-white focus:border-gold-500 focus:outline-none"
          />
          <button className="px-6 py-2 bg-gold-500 text-dark-950 font-bold rounded-lg hover:bg-gold-400 transition-all">
            Subscribe
          </button>
        </div>
      </div>
    </PageLayout>
  );
}
