import React, { useEffect, useState } from 'react';
import { api } from '../services/api';
import { DashboardData } from '../types';
import {
  DollarSign, TrendingUp, Server, Activity,
  Wallet, ArrowUpRight, Clock, CheckCircle
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
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-red-600">Error: {error || 'No data available'}</div>
      </div>
    );
  }

  const { stats, operator, active_nodes, recent_earnings, recent_payouts } = data;

  // Prepare chart data
  const earningsChartData = recent_earnings.slice(0, 10).reverse().map(e => ({
    date: format(new Date(e.created_at), 'MM/dd'),
    amount: e.amount_usd,
    bandwidth: e.bandwidth_gb,
  }));

  const getTierColor = (tier: string) => {
    const colors: Record<string, string> = {
      bronze: 'text-amber-700 bg-amber-100',
      silver: 'text-gray-700 bg-gray-200',
      gold: 'text-yellow-700 bg-yellow-100',
      platinum: 'text-purple-700 bg-purple-100',
    };
    return colors[tier] || 'text-gray-700 bg-gray-100';
  };

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Operator Dashboard</h1>
          <p className="text-gray-600 mt-2">
            Welcome back! Track your earnings and manage your nodes.
          </p>
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          {/* Total Earned */}
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Total Earned</p>
                <p className="text-2xl font-bold text-gray-900 mt-2">
                  ${stats.total_earned.toFixed(2)}
                </p>
              </div>
              <div className="bg-green-100 p-3 rounded-full">
                <DollarSign className="h-6 w-6 text-green-600" />
              </div>
            </div>
          </div>

          {/* Pending Payout */}
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Pending Payout</p>
                <p className="text-2xl font-bold text-gray-900 mt-2">
                  ${stats.pending_payout.toFixed(2)}
                </p>
                <button
                  onClick={handleRequestPayout}
                  disabled={stats.pending_payout < 10}
                  className="mt-2 text-xs text-primary-600 hover:text-primary-700 disabled:text-gray-400 disabled:cursor-not-allowed"
                >
                  Request Payout →
                </button>
              </div>
              <div className="bg-blue-100 p-3 rounded-full">
                <Wallet className="h-6 w-6 text-blue-600" />
              </div>
            </div>
          </div>

          {/* Active Nodes */}
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Active Nodes</p>
                <p className="text-2xl font-bold text-gray-900 mt-2">
                  {stats.active_nodes}
                </p>
              </div>
              <div className="bg-purple-100 p-3 rounded-full">
                <Server className="h-6 w-6 text-purple-600" />
              </div>
            </div>
          </div>

          {/* Reputation */}
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Reputation Score</p>
                <p className="text-2xl font-bold text-gray-900 mt-2">
                  {stats.reputation_score.toFixed(1)}
                </p>
                <span className={`mt-2 inline-block px-2 py-1 text-xs font-semibold rounded-full ${getTierColor(stats.current_tier)}`}>
                  {stats.current_tier.toUpperCase()}
                </span>
              </div>
              <div className="bg-yellow-100 p-3 rounded-full">
                <TrendingUp className="h-6 w-6 text-yellow-600" />
              </div>
            </div>
          </div>
        </div>

        {/* Charts Row */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
          {/* Earnings Chart */}
          <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Earnings Trend</h2>
            <ResponsiveContainer width="100%" height={250}>
              <LineChart data={earningsChartData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="date" />
                <YAxis />
                <Tooltip formatter={(value: number) => `$${value.toFixed(2)}`} />
                <Line type="monotone" dataKey="amount" stroke="#3b82f6" strokeWidth={2} />
              </LineChart>
            </ResponsiveContainer>
          </div>

          {/* Bandwidth Chart */}
          <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Bandwidth Served</h2>
            <ResponsiveContainer width="100%" height={250}>
              <BarChart data={earningsChartData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="date" />
                <YAxis />
                <Tooltip formatter={(value: number) => `${value.toFixed(1)} GB`} />
                <Bar dataKey="bandwidth" fill="#8b5cf6" />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Nodes Status */}
        <div className="bg-white rounded-lg shadow mb-8">
          <div className="p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Active Nodes</h2>
            <div className="space-y-4">
              {active_nodes.length === 0 ? (
                <p className="text-gray-500">No active nodes yet</p>
              ) : (
                active_nodes.map(node => (
                  <div key={node.id} className="flex items-center justify-between p-4 border rounded-lg hover:bg-gray-50">
                    <div className="flex items-center space-x-4">
                      <Activity className={`h-5 w-5 ${node.status === 'online' ? 'text-green-500' : 'text-gray-400'}`} />
                      <div>
                        <p className="font-medium text-gray-900">{node.name}</p>
                        <p className="text-sm text-gray-500">{node.city}, {node.country}</p>
                      </div>
                    </div>
                    <div className="text-right">
                      <p className="font-medium text-gray-900">${node.total_earned_usd.toFixed(2)}</p>
                      <p className="text-sm text-gray-500">{node.uptime_percentage.toFixed(1)}% uptime</p>
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
          <div className="bg-white rounded-lg shadow">
            <div className="p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-4">Recent Earnings</h2>
              <div className="space-y-3">
                {recent_earnings.slice(0, 5).map(earning => (
                  <div key={earning.id} className="flex items-center justify-between py-2 border-b last:border-0">
                    <div className="flex items-center space-x-3">
                      <ArrowUpRight className="h-4 w-4 text-green-500" />
                      <div>
                        <p className="text-sm font-medium text-gray-900">
                          {earning.bandwidth_gb.toFixed(1)} GB
                        </p>
                        <p className="text-xs text-gray-500">
                          {format(new Date(earning.created_at), 'MMM d, h:mm a')}
                        </p>
                      </div>
                    </div>
                    <span className="text-sm font-semibold text-green-600">
                      +${earning.amount_usd.toFixed(2)}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Recent Payouts */}
          <div className="bg-white rounded-lg shadow">
            <div className="p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-4">Recent Payouts</h2>
              <div className="space-y-3">
                {recent_payouts.length === 0 ? (
                  <p className="text-gray-500 text-sm">No payouts yet</p>
                ) : (
                  recent_payouts.slice(0, 5).map(payout => (
                    <div key={payout.id} className="flex items-center justify-between py-2 border-b last:border-0">
                      <div className="flex items-center space-x-3">
                        {payout.status === 'completed' ? (
                          <CheckCircle className="h-4 w-4 text-green-500" />
                        ) : (
                          <Clock className="h-4 w-4 text-yellow-500" />
                        )}
                        <div>
                          <p className="text-sm font-medium text-gray-900">
                            {payout.crypto_amount.toFixed(6)} {payout.crypto_currency.toUpperCase()}
                          </p>
                          <p className="text-xs text-gray-500">
                            {format(new Date(payout.created_at), 'MMM d, h:mm a')}
                          </p>
                        </div>
                      </div>
                      <div className="text-right">
                        <span className="text-sm font-semibold text-gray-900">
                          ${payout.amount_usd.toFixed(2)}
                        </span>
                        <p className="text-xs text-gray-500 capitalize">{payout.status}</p>
                      </div>
                    </div>
                  ))
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
