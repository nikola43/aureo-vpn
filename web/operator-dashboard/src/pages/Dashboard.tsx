import React, { useEffect, useState } from 'react';
import { api } from '../services/api';
import { DashboardData } from '../types';
import {
  DollarSign, TrendingUp, Server, Activity,
  Wallet, ArrowUpRight, Clock, CheckCircle, Sparkles, Zap, Users
} from 'lucide-react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, BarChart, Bar } from 'recharts';
import { format } from 'date-fns';

export const Dashboard: React.FC = () => {
  const [data, setData] = useState<DashboardData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadDashboard();
  }, []);

  const loadDashboard = async () => {
    try {
      setLoading(true);
      const dashboard = await api.getOperatorDashboard();
      setData(dashboard);
    } catch (err: any) {
      // If 403, user is not registered as operator
      if (err.response?.status === 403) {
        window.location.href = '/register-operator';
        return;
      }
      setError(err.message || 'Failed to load dashboard');
    } finally {
      setLoading(false);
    }
  };

  const handleRequestPayout = async () => {
    if (!data?.stats.pending_payout || data.stats.pending_payout < 10) {
      alert('Minimum payout amount is $10');
      return;
    }

    if (!confirm('Request payout? This will process your pending balance.')) {
      return;
    }

    try {
      await api.requestPayout();
      alert('Payout requested successfully! Processing may take 24-48 hours.');
      loadDashboard();
    } catch (err: any) {
      alert(err.response?.data?.error || 'Failed to request payout');
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen relative overflow-hidden flex items-center justify-center">
        {/* Animated gradient background */}
        <div className="absolute inset-0 bg-gradient-to-br from-indigo-600 via-purple-600 to-pink-600">
          <div className="absolute inset-0 opacity-30">
            <div className="absolute top-0 left-0 w-96 h-96 bg-blue-500 rounded-full mix-blend-multiply filter blur-3xl animate-blob"></div>
            <div className="absolute top-0 right-0 w-96 h-96 bg-purple-500 rounded-full mix-blend-multiply filter blur-3xl animate-blob animation-delay-2000"></div>
            <div className="absolute bottom-0 left-1/2 w-96 h-96 bg-pink-500 rounded-full mix-blend-multiply filter blur-3xl animate-blob animation-delay-4000"></div>
          </div>
        </div>
        <div className="relative backdrop-blur-xl bg-white/10 border border-white/20 rounded-3xl shadow-2xl p-12">
          <div className="flex flex-col items-center space-y-4">
            <div className="animate-spin rounded-full h-16 w-16 border-4 border-white/20 border-t-white"></div>
            <p className="text-white text-lg font-medium">Loading Dashboard...</p>
          </div>
        </div>
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="min-h-screen relative overflow-hidden flex items-center justify-center p-4">
        {/* Animated gradient background */}
        <div className="absolute inset-0 bg-gradient-to-br from-indigo-600 via-purple-600 to-pink-600">
          <div className="absolute inset-0 opacity-30">
            <div className="absolute top-0 left-0 w-96 h-96 bg-blue-500 rounded-full mix-blend-multiply filter blur-3xl animate-blob"></div>
            <div className="absolute top-0 right-0 w-96 h-96 bg-purple-500 rounded-full mix-blend-multiply filter blur-3xl animate-blob animation-delay-2000"></div>
            <div className="absolute bottom-0 left-1/2 w-96 h-96 bg-pink-500 rounded-full mix-blend-multiply filter blur-3xl animate-blob animation-delay-4000"></div>
          </div>
        </div>
        <div className="relative backdrop-blur-xl bg-red-500/20 border border-red-500/30 rounded-3xl shadow-2xl p-8">
          <div className="flex items-center space-x-3">
            <Zap className="h-6 w-6 text-white" />
            <p className="text-white text-lg font-medium">Error: {error || 'No data available'}</p>
          </div>
        </div>
      </div>
    );
  }

  const { stats, active_nodes, recent_earnings, recent_payouts } = data;

  // Calculate current traffic in MB/s from all active nodes
  const currentTrafficMB = active_nodes.reduce((total, node) => {
    // Convert Gbps to MB/s: 1 Gbps = 125 MB/s
    return total + ((node.bandwidth_usage_gbps || 0) * 125);
  }, 0);

  // Calculate total connected users from all active nodes
  const connectedUsers = active_nodes.reduce((total, node) => {
    return total + (node.current_connections || 0);
  }, 0);

  // Prepare chart data
  const earningsChartData = recent_earnings.slice(0, 10).reverse().map(e => ({
    date: format(new Date(e.created_at), 'MM/dd'),
    amount: e.amount_usd,
    bandwidth: e.bandwidth_gb,
  }));

  const getTierColor = (tier: string) => {
    const colors: Record<string, string> = {
      bronze: 'from-amber-600 to-amber-800',
      silver: 'from-gray-400 to-gray-600',
      gold: 'from-yellow-400 to-yellow-600',
      platinum: 'from-purple-400 to-purple-600',
    };
    return colors[tier] || 'from-gray-400 to-gray-600';
  };

  const getTierIcon = (tier: string) => {
    const icons: Record<string, string> = {
      bronze: '🥉',
      silver: '🥈',
      gold: '🥇',
      platinum: '💎',
    };
    return icons[tier] || '⭐';
  };

  return (
    <div className="min-h-screen relative overflow-hidden">
      {/* Animated gradient background */}
      <div className="fixed inset-0 bg-gradient-to-br from-indigo-600 via-purple-600 to-pink-600 -z-10">
        <div className="absolute inset-0 opacity-30">
          <div className="absolute top-0 left-0 w-96 h-96 bg-blue-500 rounded-full mix-blend-multiply filter blur-3xl animate-blob"></div>
          <div className="absolute top-0 right-0 w-96 h-96 bg-purple-500 rounded-full mix-blend-multiply filter blur-3xl animate-blob animation-delay-2000"></div>
          <div className="absolute bottom-0 left-1/2 w-96 h-96 bg-pink-500 rounded-full mix-blend-multiply filter blur-3xl animate-blob animation-delay-4000"></div>
        </div>
      </div>

      {/* Floating particles */}
      <div className="fixed inset-0 overflow-hidden pointer-events-none -z-10">
        <div className="absolute top-1/4 left-1/4 w-2 h-2 bg-white/40 rounded-full animate-float"></div>
        <div className="absolute top-1/3 right-1/4 w-3 h-3 bg-white/30 rounded-full animate-float animation-delay-2000"></div>
        <div className="absolute bottom-1/4 left-1/3 w-2 h-2 bg-white/40 rounded-full animate-float animation-delay-4000"></div>
        <div className="absolute top-2/3 right-1/3 w-2 h-2 bg-white/30 rounded-full animate-float animation-delay-3000"></div>
      </div>

      <div className="relative p-6 min-h-screen">
        <div className="max-w-7xl mx-auto">
          {/* Header */}
          <div className="mb-8">
            <div className="backdrop-blur-xl bg-white/10 border border-white/20 rounded-3xl shadow-2xl p-8">
              <div className="absolute inset-0 bg-gradient-to-br from-white/20 to-transparent rounded-3xl pointer-events-none"></div>
              <div className="relative flex items-center space-x-4">
                <div className="flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br from-blue-500 to-purple-600 shadow-lg">
                  <Sparkles className="h-8 w-8 text-white" />
                </div>
                <div>
                  <h1 className="text-4xl font-bold text-white tracking-tight">Operator Dashboard</h1>
                  <p className="text-white/80 text-lg font-light mt-1">
                    Welcome back! Track your earnings and manage your nodes.
                  </p>
                </div>
              </div>
            </div>
          </div>

          {/* Stats Grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-6 mb-8">
            {/* Total Earned */}
            <div className="group relative backdrop-blur-xl bg-white/10 border border-white/20 rounded-2xl shadow-xl overflow-hidden hover:scale-105 transition-all duration-300">
              <div className="absolute inset-0 bg-gradient-to-br from-white/20 to-transparent pointer-events-none"></div>
              <div className="relative p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-white/80">Total Earned</p>
                    <p className="text-3xl font-bold text-white mt-2">
                      ${(stats.total_earned || 0).toFixed(2)}
                    </p>
                  </div>
                  <div className="bg-gradient-to-br from-green-400 to-green-600 p-4 rounded-2xl shadow-lg">
                    <DollarSign className="h-7 w-7 text-white" />
                  </div>
                </div>
              </div>
              <div className="absolute bottom-0 left-0 right-0 h-1 bg-gradient-to-r from-green-400 to-green-600"></div>
            </div>

            {/* Pending Payout */}
            <div className="group relative backdrop-blur-xl bg-white/10 border border-white/20 rounded-2xl shadow-xl overflow-hidden hover:scale-105 transition-all duration-300">
              <div className="absolute inset-0 bg-gradient-to-br from-white/20 to-transparent pointer-events-none"></div>
              <div className="relative p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-white/80">Pending Payout</p>
                    <p className="text-3xl font-bold text-white mt-2">
                      ${(stats.pending_payout || 0).toFixed(2)}
                    </p>
                    <button
                      onClick={handleRequestPayout}
                      disabled={(stats.pending_payout || 0) < 10}
                      className="mt-3 px-4 py-1.5 bg-white/10 hover:bg-white/20 border border-white/30 rounded-lg text-xs text-white font-medium transition-all duration-300 disabled:opacity-40 disabled:cursor-not-allowed hover:scale-105 active:scale-95"
                    >
                      Request Payout →
                    </button>
                  </div>
                  <div className="bg-gradient-to-br from-blue-400 to-blue-600 p-4 rounded-2xl shadow-lg">
                    <Wallet className="h-7 w-7 text-white" />
                  </div>
                </div>
              </div>
              <div className="absolute bottom-0 left-0 right-0 h-1 bg-gradient-to-r from-blue-400 to-blue-600"></div>
            </div>

            {/* Active Nodes */}
            <div className="group relative backdrop-blur-xl bg-white/10 border border-white/20 rounded-2xl shadow-xl overflow-hidden hover:scale-105 transition-all duration-300">
              <div className="absolute inset-0 bg-gradient-to-br from-white/20 to-transparent pointer-events-none"></div>
              <div className="relative p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-white/80">Active Nodes</p>
                    <p className="text-3xl font-bold text-white mt-2">
                      {stats.active_nodes}
                    </p>
                  </div>
                  <div className="bg-gradient-to-br from-purple-400 to-purple-600 p-4 rounded-2xl shadow-lg">
                    <Server className="h-7 w-7 text-white" />
                  </div>
                </div>
              </div>
              <div className="absolute bottom-0 left-0 right-0 h-1 bg-gradient-to-r from-purple-400 to-purple-600"></div>
            </div>

            {/* Reputation */}
            <div className="group relative backdrop-blur-xl bg-white/10 border border-white/20 rounded-2xl shadow-xl overflow-hidden hover:scale-105 transition-all duration-300">
              <div className="absolute inset-0 bg-gradient-to-br from-white/20 to-transparent pointer-events-none"></div>
              <div className="relative p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-white/80">Reputation Score</p>
                    <p className="text-3xl font-bold text-white mt-2">
                      {(stats.reputation_score || 0).toFixed(1)}
                    </p>
                    <div className={`mt-3 inline-flex items-center px-3 py-1.5 rounded-full bg-gradient-to-r ${getTierColor(stats.current_tier || 'bronze')} text-white text-xs font-bold shadow-lg`}>
                      <span className="mr-1">{getTierIcon(stats.current_tier || 'bronze')}</span>
                      {(stats.current_tier || 'bronze').toUpperCase()}
                    </div>
                  </div>
                  <div className="bg-gradient-to-br from-yellow-400 to-yellow-600 p-4 rounded-2xl shadow-lg">
                    <TrendingUp className="h-7 w-7 text-white" />
                  </div>
                </div>
              </div>
              <div className="absolute bottom-0 left-0 right-0 h-1 bg-gradient-to-r from-yellow-400 to-yellow-600"></div>
            </div>

            {/* Current Traffic */}
            <div className="group relative backdrop-blur-xl bg-white/10 border border-white/20 rounded-2xl shadow-xl overflow-hidden hover:scale-105 transition-all duration-300">
              <div className="absolute inset-0 bg-gradient-to-br from-white/20 to-transparent pointer-events-none"></div>
              <div className="relative p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-white/80">Current Traffic</p>
                    <p className="text-3xl font-bold text-white mt-2">
                      {currentTrafficMB.toFixed(1)}
                    </p>
                    <p className="text-xs text-white/60 mt-1">MB/s</p>
                  </div>
                  <div className="bg-gradient-to-br from-cyan-400 to-cyan-600 p-4 rounded-2xl shadow-lg">
                    <Activity className="h-7 w-7 text-white" />
                  </div>
                </div>
              </div>
              <div className="absolute bottom-0 left-0 right-0 h-1 bg-gradient-to-r from-cyan-400 to-cyan-600"></div>
            </div>

            {/* Connected Users */}
            <div className="group relative backdrop-blur-xl bg-white/10 border border-white/20 rounded-2xl shadow-xl overflow-hidden hover:scale-105 transition-all duration-300">
              <div className="absolute inset-0 bg-gradient-to-br from-white/20 to-transparent pointer-events-none"></div>
              <div className="relative p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-white/80">Connected Users</p>
                    <p className="text-3xl font-bold text-white mt-2">
                      {connectedUsers}
                    </p>
                    <p className="text-xs text-white/60 mt-1">active now</p>
                  </div>
                  <div className="bg-gradient-to-br from-indigo-400 to-indigo-600 p-4 rounded-2xl shadow-lg">
                    <Users className="h-7 w-7 text-white" />
                  </div>
                </div>
              </div>
              <div className="absolute bottom-0 left-0 right-0 h-1 bg-gradient-to-r from-indigo-400 to-indigo-600"></div>
            </div>
          </div>

          {/* Charts Row */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
            {/* Earnings Chart */}
            <div className="relative backdrop-blur-xl bg-white/10 border border-white/20 rounded-2xl shadow-xl overflow-hidden">
              <div className="absolute inset-0 bg-gradient-to-br from-white/10 to-transparent pointer-events-none"></div>
              <div className="relative p-6">
                <h2 className="text-xl font-bold text-white mb-6 flex items-center">
                  <div className="w-1 h-6 bg-gradient-to-b from-blue-400 to-purple-600 rounded-full mr-3"></div>
                  Earnings Trend
                </h2>
                <div className="backdrop-blur-lg bg-white/5 rounded-xl p-4 border border-white/10">
                  <ResponsiveContainer width="100%" height={250}>
                    <LineChart data={earningsChartData}>
                      <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.1)" />
                      <XAxis dataKey="date" stroke="rgba(255,255,255,0.6)" />
                      <YAxis stroke="rgba(255,255,255,0.6)" />
                      <Tooltip
                        contentStyle={{
                          backgroundColor: 'rgba(255,255,255,0.1)',
                          backdropFilter: 'blur(10px)',
                          border: '1px solid rgba(255,255,255,0.2)',
                          borderRadius: '12px',
                          color: 'white'
                        }}
                        formatter={(value: number) => `$${value.toFixed(2)}`}
                      />
                      <Line type="monotone" dataKey="amount" stroke="#60a5fa" strokeWidth={3} dot={{ fill: '#3b82f6', r: 5 }} />
                    </LineChart>
                  </ResponsiveContainer>
                </div>
              </div>
            </div>

            {/* Bandwidth Chart */}
            <div className="relative backdrop-blur-xl bg-white/10 border border-white/20 rounded-2xl shadow-xl overflow-hidden">
              <div className="absolute inset-0 bg-gradient-to-br from-white/10 to-transparent pointer-events-none"></div>
              <div className="relative p-6">
                <h2 className="text-xl font-bold text-white mb-6 flex items-center">
                  <div className="w-1 h-6 bg-gradient-to-b from-purple-400 to-pink-600 rounded-full mr-3"></div>
                  Bandwidth Served
                </h2>
                <div className="backdrop-blur-lg bg-white/5 rounded-xl p-4 border border-white/10">
                  <ResponsiveContainer width="100%" height={250}>
                    <BarChart data={earningsChartData}>
                      <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.1)" />
                      <XAxis dataKey="date" stroke="rgba(255,255,255,0.6)" />
                      <YAxis stroke="rgba(255,255,255,0.6)" />
                      <Tooltip
                        contentStyle={{
                          backgroundColor: 'rgba(255,255,255,0.1)',
                          backdropFilter: 'blur(10px)',
                          border: '1px solid rgba(255,255,255,0.2)',
                          borderRadius: '12px',
                          color: 'white'
                        }}
                        formatter={(value: number) => `${value.toFixed(1)} GB`}
                      />
                      <Bar dataKey="bandwidth" fill="url(#colorBandwidth)" radius={[8, 8, 0, 0]} />
                      <defs>
                        <linearGradient id="colorBandwidth" x1="0" y1="0" x2="0" y2="1">
                          <stop offset="0%" stopColor="#a78bfa" stopOpacity={1} />
                          <stop offset="100%" stopColor="#8b5cf6" stopOpacity={0.8} />
                        </linearGradient>
                      </defs>
                    </BarChart>
                  </ResponsiveContainer>
                </div>
              </div>
            </div>
          </div>

          {/* Nodes Status */}
          <div className="relative backdrop-blur-xl bg-white/10 border border-white/20 rounded-2xl shadow-xl overflow-hidden mb-8">
            <div className="absolute inset-0 bg-gradient-to-br from-white/10 to-transparent pointer-events-none"></div>
            <div className="relative p-6">
              <h2 className="text-xl font-bold text-white mb-6 flex items-center">
                <div className="w-1 h-6 bg-gradient-to-b from-green-400 to-blue-600 rounded-full mr-3"></div>
                Active Nodes
              </h2>
              <div className="space-y-4">
                {active_nodes.length === 0 ? (
                  <div className="backdrop-blur-lg bg-white/5 border border-white/10 rounded-xl p-8 text-center">
                    <Server className="h-12 w-12 text-white/40 mx-auto mb-3" />
                    <p className="text-white/60">No active nodes yet</p>
                  </div>
                ) : (
                  active_nodes.map(node => (
                    <div
                      key={node.id}
                      className="group backdrop-blur-lg bg-white/5 hover:bg-white/10 border border-white/20 rounded-xl p-5 transition-all duration-300 hover:scale-[1.02] hover:shadow-xl"
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-4">
                          <div className={`relative ${node.status === 'online' ? 'bg-gradient-to-br from-green-400 to-green-600' : 'bg-gradient-to-br from-gray-400 to-gray-600'} p-3 rounded-xl shadow-lg`}>
                            <Activity className="h-6 w-6 text-white" />
                            {node.status === 'online' && (
                              <div className="absolute -top-1 -right-1 w-3 h-3 bg-green-400 rounded-full animate-ping"></div>
                            )}
                          </div>
                          <div>
                            <p className="font-semibold text-white text-lg">{node.name}</p>
                            <p className="text-sm text-white/70">{node.city}, {node.country}</p>
                          </div>
                        </div>
                        <div className="text-right">
                          <p className="font-bold text-white text-lg">${(node.total_earned_usd || 0).toFixed(2)}</p>
                          <div className="flex items-center justify-end space-x-2 mt-1">
                            <div className="w-24 h-2 bg-white/20 rounded-full overflow-hidden">
                              <div
                                className="h-full bg-gradient-to-r from-green-400 to-blue-500 rounded-full transition-all duration-500"
                                style={{ width: `${node.uptime_percentage || 0}%` }}
                              ></div>
                            </div>
                            <p className="text-sm text-white/70 font-medium">{(node.uptime_percentage || 0).toFixed(1)}%</p>
                          </div>
                        </div>
                      </div>
                    </div>
                  ))
                )}
              </div>
            </div>
          </div>

          {/* Recent Activity */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Recent Earnings */}
            <div className="relative backdrop-blur-xl bg-white/10 border border-white/20 rounded-2xl shadow-xl overflow-hidden">
              <div className="absolute inset-0 bg-gradient-to-br from-white/10 to-transparent pointer-events-none"></div>
              <div className="relative p-6">
                <h2 className="text-xl font-bold text-white mb-6 flex items-center">
                  <div className="w-1 h-6 bg-gradient-to-b from-green-400 to-emerald-600 rounded-full mr-3"></div>
                  Recent Earnings
                </h2>
                <div className="space-y-3">
                  {recent_earnings.slice(0, 5).map(earning => (
                    <div
                      key={earning.id}
                      className="backdrop-blur-lg bg-white/5 hover:bg-white/10 border border-white/10 rounded-xl p-4 transition-all duration-300 hover:scale-[1.02]"
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-3">
                          <div className="bg-gradient-to-br from-green-400 to-emerald-600 p-2 rounded-lg">
                            <ArrowUpRight className="h-4 w-4 text-white" />
                          </div>
                          <div>
                            <p className="text-sm font-semibold text-white">
                              {(earning.bandwidth_gb || 0).toFixed(1)} GB
                            </p>
                            <p className="text-xs text-white/60">
                              {format(new Date(earning.created_at), 'MMM d, h:mm a')}
                            </p>
                          </div>
                        </div>
                        <span className="text-base font-bold text-green-300">
                          +${(earning.amount_usd || 0).toFixed(2)}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>

            {/* Recent Payouts */}
            <div className="relative backdrop-blur-xl bg-white/10 border border-white/20 rounded-2xl shadow-xl overflow-hidden">
              <div className="absolute inset-0 bg-gradient-to-br from-white/10 to-transparent pointer-events-none"></div>
              <div className="relative p-6">
                <h2 className="text-xl font-bold text-white mb-6 flex items-center">
                  <div className="w-1 h-6 bg-gradient-to-b from-blue-400 to-cyan-600 rounded-full mr-3"></div>
                  Recent Payouts
                </h2>
                <div className="space-y-3">
                  {recent_payouts.length === 0 ? (
                    <div className="backdrop-blur-lg bg-white/5 border border-white/10 rounded-xl p-6 text-center">
                      <Wallet className="h-10 w-10 text-white/40 mx-auto mb-2" />
                      <p className="text-white/60 text-sm">No payouts yet</p>
                    </div>
                  ) : (
                    recent_payouts.slice(0, 5).map(payout => (
                      <div
                        key={payout.id}
                        className="backdrop-blur-lg bg-white/5 hover:bg-white/10 border border-white/10 rounded-xl p-4 transition-all duration-300 hover:scale-[1.02]"
                      >
                        <div className="flex items-center justify-between">
                          <div className="flex items-center space-x-3">
                            <div className={`${payout.status === 'completed' ? 'bg-gradient-to-br from-green-400 to-emerald-600' : 'bg-gradient-to-br from-yellow-400 to-orange-600'} p-2 rounded-lg`}>
                              {payout.status === 'completed' ? (
                                <CheckCircle className="h-4 w-4 text-white" />
                              ) : (
                                <Clock className="h-4 w-4 text-white" />
                              )}
                            </div>
                            <div>
                              <p className="text-sm font-semibold text-white">
                                {(payout.crypto_amount || 0).toFixed(6)} {payout.crypto_currency.toUpperCase()}
                              </p>
                              <p className="text-xs text-white/60">
                                {format(new Date(payout.created_at), 'MMM d, h:mm a')}
                              </p>
                            </div>
                          </div>
                          <div className="text-right">
                            <span className="text-base font-bold text-white">
                              ${(payout.amount_usd || 0).toFixed(2)}
                            </span>
                            <p className="text-xs text-white/70 capitalize">{payout.status}</p>
                          </div>
                        </div>
                      </div>
                    ))
                  )}
                </div>
              </div>
            </div>
          </div>

          {/* Footer spacing */}
          <div className="h-8"></div>
        </div>
      </div>
    </div>
  );
};
