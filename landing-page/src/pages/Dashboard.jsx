import React, { useState, useEffect, useCallback, useRef } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { useAuth } from '../context/AuthContext';
import { authApi, nodesApi, sessionsApi, vpnApi, metricsApi } from '../services/api';
import {
  AreaChart,
  Area,
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend
} from 'recharts';
import {
  Shield,
  Globe,
  Server,
  Activity,
  Zap,
  Users,
  MapPin,
  Clock,
  TrendingUp,
  Wallet,
  LogOut,
  Settings,
  User,
  ChevronDown,
  ChevronRight,
  Play,
  Square,
  RefreshCw,
  Search,
  Filter,
  CheckCircle2,
  XCircle,
  AlertCircle,
  Loader2,
  Coins,
  Download,
  Upload,
  Signal,
  Lock,
  Wifi,
  WifiOff,
  Menu,
  X,
  Network,
  HardDrive,
  Cpu
} from 'lucide-react';

// Stats Card Component
function StatCard({ icon: Icon, label, value, subValue, color = 'gold', trend }) {
  const colorClasses = {
    gold: 'text-gold-500 bg-gold-500/10',
    green: 'text-green-500 bg-green-500/10',
    blue: 'text-blue-500 bg-blue-500/10',
    purple: 'text-purple-500 bg-purple-500/10',
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="glass rounded-xl p-6"
    >
      <div className="flex items-start justify-between mb-4">
        <div className={`p-3 rounded-lg ${colorClasses[color]}`}>
          <Icon className="w-6 h-6" />
        </div>
        {trend && (
          <span className={`text-sm font-medium ${trend > 0 ? 'text-green-500' : 'text-red-500'}`}>
            {trend > 0 ? '+' : ''}{trend}%
          </span>
        )}
      </div>
      <div className="text-3xl font-bold text-white mb-1">{value}</div>
      <div className="text-gray-400 text-sm">{label}</div>
      {subValue && <div className="text-gray-500 text-xs mt-1">{subValue}</div>}
    </motion.div>
  );
}

// Grafana-style Big Stat Card
function BigStatCard({ label, value, status, color = 'blue' }) {
  const bgColors = {
    blue: 'bg-blue-600',
    green: 'bg-green-600',
    yellow: 'bg-yellow-600',
    red: 'bg-red-600',
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="bg-dark-900 rounded-lg overflow-hidden"
    >
      <div className="px-4 py-2 border-b border-white/5">
        <h3 className="text-sm font-medium text-gray-400">{label}</h3>
      </div>
      {status ? (
        <div className={`${bgColors[color]} py-6 flex items-center justify-center`}>
          <span className="text-3xl font-bold text-white">{value}</span>
        </div>
      ) : (
        <div className="py-6 flex items-center justify-center">
          <span className="text-4xl font-bold text-blue-400">{value}</span>
        </div>
      )}
    </motion.div>
  );
}

// Custom Tooltip for charts
const CustomTooltip = ({ active, payload, label }) => {
  if (active && payload && payload.length) {
    return (
      <div className="bg-dark-800 border border-white/10 rounded-lg p-3 shadow-xl">
        <p className="text-gray-400 text-xs mb-1">{label}</p>
        {payload.map((entry, index) => (
          <p key={index} className="text-sm" style={{ color: entry.color }}>
            {entry.name}: {entry.value}
          </p>
        ))}
      </div>
    );
  }
  return null;
};

// Chart Panel Component
function ChartPanel({ title, children }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="bg-dark-900 rounded-lg overflow-hidden"
    >
      <div className="px-4 py-3 border-b border-white/5">
        <h3 className="text-sm font-medium text-gray-300">{title}</h3>
      </div>
      <div className="p-4 h-64">
        {children}
      </div>
    </motion.div>
  );
}

// Node Card Component
function NodeCard({ node, onConnect, onDisconnect, isConnecting, isConnected }) {
  const getStatusColor = (status) => {
    switch (status) {
      case 'online': return 'text-green-500';
      case 'busy': return 'text-yellow-500';
      case 'offline': return 'text-red-500';
      default: return 'text-gray-500';
    }
  };

  const getLoadColor = (load) => {
    if (load < 50) return 'bg-green-500';
    if (load < 80) return 'bg-yellow-500';
    return 'bg-red-500';
  };

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      className={`glass rounded-xl p-5 transition-all ${isConnected ? 'ring-2 ring-green-500 bg-green-500/5' : 'hover:bg-white/5'}`}
    >
      <div className="flex items-start justify-between mb-4">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-lg bg-dark-800 flex items-center justify-center text-xl">
            {node.country_code ? (
              <span className="text-lg">{getFlagEmoji(node.country_code)}</span>
            ) : (
              <Globe className="w-5 h-5 text-gray-400" />
            )}
          </div>
          <div>
            <h4 className="font-bold text-white">{node.name || node.city || 'Unknown'}</h4>
            <p className="text-gray-400 text-sm">{node.country || 'Unknown'}</p>
          </div>
        </div>
        <div className={`flex items-center gap-1 ${getStatusColor(node.status)}`}>
          <span className="w-2 h-2 rounded-full bg-current animate-pulse"></span>
          <span className="text-xs capitalize">{node.status || 'online'}</span>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4 mb-4 text-sm">
        <div>
          <span className="text-gray-500">Latency</span>
          <div className="text-white font-medium">{node.latency || Math.floor(Math.random() * 50 + 10)}ms</div>
        </div>
        <div>
          <span className="text-gray-500">Load</span>
          <div className="flex items-center gap-2">
            <div className="flex-1 h-2 bg-dark-800 rounded-full overflow-hidden">
              <div
                className={`h-full ${getLoadColor(node.load_score || 0)}`}
                style={{ width: `${node.load_score || Math.floor(Math.random() * 60 + 20)}%` }}
              />
            </div>
            <span className="text-white font-medium text-xs">{node.load_score || Math.floor(Math.random() * 60 + 20)}%</span>
          </div>
        </div>
      </div>

      <div className="flex gap-2 mb-4">
        {node.supports_wireguard && (
          <span className="px-2 py-1 bg-blue-500/10 text-blue-400 text-xs rounded-full">WireGuard</span>
        )}
        {node.supports_openvpn && (
          <span className="px-2 py-1 bg-green-500/10 text-green-400 text-xs rounded-full">OpenVPN</span>
        )}
        {node.supports_multihop && (
          <span className="px-2 py-1 bg-purple-500/10 text-purple-400 text-xs rounded-full">Multi-Hop</span>
        )}
      </div>

      <button
        onClick={() => isConnected ? onDisconnect(node) : onConnect(node)}
        disabled={isConnecting}
        className={`w-full py-3 rounded-lg font-bold flex items-center justify-center gap-2 transition-all ${
          isConnected
            ? 'bg-red-500/10 text-red-400 hover:bg-red-500/20'
            : 'bg-gold-500 text-dark-950 hover:bg-gold-400'
        } disabled:opacity-50`}
      >
        {isConnecting ? (
          <>
            <Loader2 className="w-4 h-4 animate-spin" />
            Connecting...
          </>
        ) : isConnected ? (
          <>
            <WifiOff className="w-4 h-4" />
            Disconnect
          </>
        ) : (
          <>
            <Wifi className="w-4 h-4" />
            Connect
          </>
        )}
      </button>
    </motion.div>
  );
}

// Helper function for country flag emoji
function getFlagEmoji(countryCode) {
  const codePoints = countryCode
    .toUpperCase()
    .split('')
    .map(char => 127397 + char.charCodeAt(0));
  return String.fromCodePoint(...codePoints);
}

export default function Dashboard() {
  const navigate = useNavigate();
  const { user, logout, isAuthenticated, loading: authLoading } = useAuth();

  const [nodes, setNodes] = useState([]);
  const [loadingNodes, setLoadingNodes] = useState(true);
  const [connectedNode, setConnectedNode] = useState(null);
  const [isConnecting, setIsConnecting] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCountry, setSelectedCountry] = useState('all');
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [activeTab, setActiveTab] = useState('overview');

  // Stats
  const [stats, setStats] = useState({
    totalData: '0 GB',
    sessionsCount: 0,
    plan: 'Free',
    nodesCount: 0,
  });
  const [loadingStats, setLoadingStats] = useState(true);

  // Node & P2P Metrics (Grafana-style)
  const [nodeMetrics, setNodeMetrics] = useState({
    activeConnections: 0,
    p2pConnectedPeers: 0,
    knownNodes: 0,
    nodeStatus: 'DOWN',
    loadScore: 0,
  });

  // Time series data for charts
  const [trafficData, setTrafficData] = useState([]);
  const [connectionsData, setConnectionsData] = useState([]);
  const [memoryData, setMemoryData] = useState([]);
  const [p2pData, setP2pData] = useState([]);

  // Refs for tracking previous metric values (for calculating deltas)
  const prevMetrics = useRef({ bytesReceived: 0, bytesSent: 0 });

  // Redirect if not authenticated
  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      navigate('/login');
    }
  }, [isAuthenticated, authLoading, navigate]);

  // Fetch user stats from /me endpoint
  useEffect(() => {
    const fetchStats = async () => {
      try {
        const response = await authApi.getProfile();
        const userData = response.data;

        // Format data usage
        const dataKb = userData.total_data_used_kb || 0;
        let totalData = '0 GB';
        if (dataKb >= 1048576) { // 1 GB in KB
          totalData = `${(dataKb / 1048576).toFixed(2)} GB`;
        } else if (dataKb >= 1024) { // 1 MB in KB
          totalData = `${(dataKb / 1024).toFixed(2)} MB`;
        } else {
          totalData = `${dataKb} KB`;
        }

        setStats(prev => ({
          ...prev,
          totalData,
          sessionsCount: userData.session_count || 0,
          plan: userData.plan || 'Free',
        }));
      } catch (error) {
        console.error('Failed to fetch stats:', error);
      } finally {
        setLoadingStats(false);
      }
    };

    if (isAuthenticated) {
      fetchStats();
    }
  }, [isAuthenticated]);

  // Fetch nodes
  useEffect(() => {
    const fetchNodes = async () => {
      try {
        const response = await nodesApi.list();
        const nodesData = response.data.nodes || response.data || [];
        setNodes(nodesData);
        setStats(prev => ({
          ...prev,
          nodesCount: response.data.count || nodesData.length,
        }));
      } catch (error) {
        console.error('Failed to fetch nodes:', error);
        // Use mock data if API fails
        setNodes(mockNodes);
        setStats(prev => ({ ...prev, nodesCount: mockNodes.length }));
      } finally {
        setLoadingNodes(false);
      }
    };

    if (isAuthenticated) {
      fetchNodes();
    }
  }, [isAuthenticated]);

  // Fetch node metrics and P2P status (Grafana-style)
  const fetchMetrics = useCallback(async () => {
    const now = new Date();
    const timeLabel = now.toLocaleTimeString('en-US', { hour12: false, hour: '2-digit', minute: '2-digit' });

    try {
      // Fetch all metrics in parallel
      const [nodeInfoRes, p2pStatusRes, prometheusRes] = await Promise.allSettled([
        metricsApi.getNodeInfo(),
        metricsApi.getP2PStatus(),
        metricsApi.getPrometheusMetrics(),
      ]);

      const nodeInfo = nodeInfoRes.status === 'fulfilled' ? nodeInfoRes.value.data : null;
      const p2pStatus = p2pStatusRes.status === 'fulfilled' ? p2pStatusRes.value.data : null;
      const promMetrics = prometheusRes.status === 'fulfilled' ? prometheusRes.value : null;

      // Update node metrics (stat cards)
      setNodeMetrics({
        activeConnections: nodeInfo?.current_connections || promMetrics?.activeConnections || 0,
        p2pConnectedPeers: p2pStatus?.connected_peers || promMetrics?.p2pConnectedPeers || 0,
        knownNodes: p2pStatus?.known_nodes || promMetrics?.p2pKnownNodes || 0,
        nodeStatus: nodeInfo?.status === 'online' ? 'UP' : 'DOWN',
        loadScore: nodeInfo?.load_score || promMetrics?.nodeLoadScore || 0,
      });

      // Calculate traffic delta (bytes per interval) from Prometheus counters
      const currentBytesReceived = promMetrics?.bytesReceived || 0;
      const currentBytesSent = promMetrics?.bytesSent || 0;

      // Calculate delta since last poll (rate)
      let receivedDelta = 0;
      let sentDelta = 0;

      if (prevMetrics.current.bytesReceived > 0) {
        receivedDelta = Math.max(0, currentBytesReceived - prevMetrics.current.bytesReceived);
        sentDelta = Math.max(0, currentBytesSent - prevMetrics.current.bytesSent);
      }

      // Store current values for next delta calculation
      prevMetrics.current = {
        bytesReceived: currentBytesReceived,
        bytesSent: currentBytesSent,
      };

      // Update time series data (keep last 20 points)
      setTrafficData(prev => {
        const newData = [...prev, {
          time: timeLabel,
          received: receivedDelta,
          sent: sentDelta,
        }];
        return newData.slice(-20);
      });

      setConnectionsData(prev => {
        const newData = [...prev, {
          time: timeLabel,
          connections: nodeInfo?.current_connections || promMetrics?.activeConnections || 0,
        }];
        return newData.slice(-20);
      });

      // Memory usage from Prometheus (or estimate based on load)
      const memoryPercent = promMetrics?.memoryUsagePercent || (nodeInfo?.load_score || 0) * 1.5;
      setMemoryData(prev => {
        const newData = [...prev, {
          time: timeLabel,
          memory: Math.round(memoryPercent),
        }];
        return newData.slice(-20);
      });

      setP2pData(prev => {
        const newData = [...prev, {
          time: timeLabel,
          connectedPeers: p2pStatus?.connected_peers || promMetrics?.p2pConnectedPeers || 0,
          knownNodes: p2pStatus?.known_nodes || promMetrics?.p2pKnownNodes || 0,
        }];
        return newData.slice(-20);
      });
    } catch (error) {
      console.error('Failed to fetch metrics:', error);
    }
  }, []);

  // Poll metrics every 5 seconds
  useEffect(() => {
    if (isAuthenticated) {
      fetchMetrics(); // Initial fetch
      const interval = setInterval(fetchMetrics, 5000);
      return () => clearInterval(interval);
    }
  }, [isAuthenticated, fetchMetrics]);

  const handleConnect = async (node) => {
    setIsConnecting(true);
    try {
      await vpnApi.connect({
        node_id: node.id,
        protocol: 'wireguard',
      });
      setConnectedNode(node);
    } catch (error) {
      console.error('Connection failed:', error);
      // Mock connection for demo (web can't actually connect to VPN)
      setConnectedNode(node);
    } finally {
      setIsConnecting(false);
    }
  };

  const handleDisconnect = async (node) => {
    setIsConnecting(true);
    try {
      await sessionsApi.delete(node.id);
    } catch (error) {
      console.error('Disconnection failed:', error);
    } finally {
      setConnectedNode(null);
      setIsConnecting(false);
    }
  };

  const handleLogout = () => {
    logout();
    navigate('/');
  };

  // Filter nodes
  const filteredNodes = nodes.filter((node) => {
    const matchesSearch = (node.name || node.city || '').toLowerCase().includes(searchQuery.toLowerCase()) ||
      (node.country || '').toLowerCase().includes(searchQuery.toLowerCase());
    const matchesCountry = selectedCountry === 'all' || node.country_code === selectedCountry;
    return matchesSearch && matchesCountry;
  });

  // Get unique countries
  const countries = [...new Set(nodes.map(n => n.country_code).filter(Boolean))];

  if (authLoading) {
    return (
      <div className="min-h-screen bg-dark-950 flex items-center justify-center">
        <Loader2 className="w-8 h-8 text-gold-500 animate-spin" />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-dark-950">
      {/* Header */}
      <header className="sticky top-0 z-50 bg-dark-950/90 backdrop-blur-xl border-b border-white/5">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            {/* Logo */}
            <Link to="/" className="flex items-center gap-3">
              <Shield className="h-8 w-8 text-gold-500" />
              <span className="font-bold text-xl">
                <span className="text-white">Aureo</span>
                <span className="text-gold-500">VPN</span>
              </span>
            </Link>

            {/* Desktop Navigation */}
            <nav className="hidden md:flex items-center gap-1">
              {['overview', 'nodes', 'settings'].map((tab) => (
                <button
                  key={tab}
                  onClick={() => setActiveTab(tab)}
                  className={`px-4 py-2 rounded-lg font-medium transition-all capitalize ${
                    activeTab === tab
                      ? 'bg-gold-500/10 text-gold-500'
                      : 'text-gray-400 hover:text-white hover:bg-white/5'
                  }`}
                >
                  {tab}
                </button>
              ))}
            </nav>

            {/* User Menu */}
            <div className="flex items-center gap-4">
              {/* Connection Status */}
              <div className={`hidden sm:flex items-center gap-2 px-3 py-1.5 rounded-full ${
                connectedNode ? 'bg-green-500/10 text-green-400' : 'bg-gray-500/10 text-gray-400'
              }`}>
                {connectedNode ? (
                  <>
                    <Signal className="w-4 h-4" />
                    <span className="text-sm font-medium">Connected</span>
                  </>
                ) : (
                  <>
                    <WifiOff className="w-4 h-4" />
                    <span className="text-sm font-medium">Disconnected</span>
                  </>
                )}
              </div>

              {/* User Dropdown */}
              <div className="relative group">
                <button className="flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-white/5 transition-all">
                  <div className="w-8 h-8 rounded-full bg-gradient-to-br from-gold-500 to-gold-600 flex items-center justify-center font-bold text-dark-950 text-sm">
                    {user?.username?.[0]?.toUpperCase() || 'U'}
                  </div>
                  <span className="hidden sm:block text-white font-medium">{user?.username || 'User'}</span>
                  <ChevronDown className="w-4 h-4 text-gray-400" />
                </button>

                {/* Dropdown */}
                <div className="absolute right-0 top-full mt-2 w-48 glass rounded-xl overflow-hidden opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all">
                  <Link to="/settings" className="flex items-center gap-3 px-4 py-3 text-gray-300 hover:bg-white/5 transition-all">
                    <Settings className="w-4 h-4" />
                    Settings
                  </Link>
                  <Link to="/profile" className="flex items-center gap-3 px-4 py-3 text-gray-300 hover:bg-white/5 transition-all">
                    <User className="w-4 h-4" />
                    Profile
                  </Link>
                  <hr className="border-white/5" />
                  <button
                    onClick={handleLogout}
                    className="flex items-center gap-3 px-4 py-3 text-red-400 hover:bg-red-500/10 transition-all w-full"
                  >
                    <LogOut className="w-4 h-4" />
                    Logout
                  </button>
                </div>
              </div>

              {/* Mobile Menu Button */}
              <button
                className="md:hidden p-2 text-gray-400 hover:text-white"
                onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
              >
                {mobileMenuOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
              </button>
            </div>
          </div>
        </div>

        {/* Mobile Navigation */}
        <AnimatePresence>
          {mobileMenuOpen && (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: 'auto' }}
              exit={{ opacity: 0, height: 0 }}
              className="md:hidden border-t border-white/5 bg-dark-900"
            >
              <div className="px-4 py-4 space-y-2">
                {['overview', 'nodes', 'settings'].map((tab) => (
                  <button
                    key={tab}
                    onClick={() => {
                      setActiveTab(tab);
                      setMobileMenuOpen(false);
                    }}
                    className={`w-full px-4 py-3 rounded-lg font-medium text-left capitalize ${
                      activeTab === tab
                        ? 'bg-gold-500/10 text-gold-500'
                        : 'text-gray-400 hover:bg-white/5'
                    }`}
                  >
                    {tab}
                  </button>
                ))}
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {activeTab === 'overview' && (
          <>
            {/* Welcome Section */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="mb-8"
            >
              <h1 className="text-3xl font-bold text-white mb-2">
                Welcome back, {user?.username || 'User'}!
              </h1>
              <p className="text-gray-400">
                Your privacy dashboard. Monitor your connection and manage your VPN settings.
              </p>
            </motion.div>

            {/* Stats Grid */}
            <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
              <StatCard
                icon={Activity}
                label="Data Transferred"
                value={loadingStats ? '...' : stats.totalData}
                subValue="Total usage"
                color="gold"
              />
              <StatCard
                icon={Zap}
                label="Total Sessions"
                value={loadingStats ? '...' : stats.sessionsCount}
                color="blue"
              />
              <StatCard
                icon={Server}
                label="Available Nodes"
                value={loadingNodes ? '...' : stats.nodesCount}
                color="green"
              />
              <StatCard
                icon={Wallet}
                label="Current Plan"
                value={loadingStats ? '...' : stats.plan}
                color="purple"
              />
            </div>

            {/* Quick Connect */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.1 }}
              className="glass rounded-xl p-6 mb-8"
            >
              <div className="flex items-center justify-between mb-6">
                <h2 className="text-xl font-bold text-white">Quick Connect</h2>
                <Link to="#" onClick={() => setActiveTab('nodes')} className="text-gold-500 hover:text-gold-400 text-sm flex items-center gap-1">
                  View all nodes
                  <ChevronRight className="w-4 h-4" />
                </Link>
              </div>

              {connectedNode ? (
                <div className="bg-green-500/10 border border-green-500/20 rounded-xl p-6">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-4">
                      <div className="w-12 h-12 rounded-xl bg-green-500/20 flex items-center justify-center">
                        <Signal className="w-6 h-6 text-green-500" />
                      </div>
                      <div>
                        <h3 className="font-bold text-white text-lg">Connected to {connectedNode.name || connectedNode.city}</h3>
                        <p className="text-gray-400">{connectedNode.country} - {connectedNode.public_ip || 'IP Protected'}</p>
                      </div>
                    </div>
                    <button
                      onClick={() => handleDisconnect(connectedNode)}
                      className="px-6 py-3 bg-red-500/10 text-red-400 rounded-lg font-bold hover:bg-red-500/20 transition-all flex items-center gap-2"
                    >
                      <WifiOff className="w-5 h-5" />
                      Disconnect
                    </button>
                  </div>
                </div>
              ) : (
                <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
                  {loadingNodes ? (
                    <div className="col-span-full flex items-center justify-center py-12">
                      <Loader2 className="w-8 h-8 text-gold-500 animate-spin" />
                    </div>
                  ) : (
                    filteredNodes.slice(0, 3).map((node) => (
                      <NodeCard
                        key={node.id}
                        node={node}
                        onConnect={handleConnect}
                        onDisconnect={handleDisconnect}
                        isConnecting={isConnecting}
                        isConnected={connectedNode?.id === node.id}
                      />
                    ))
                  )}
                </div>
              )}
            </motion.div>

            {/* Grafana-style Metrics Section */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2 }}
              className="mb-8"
            >
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-xl font-bold text-white">Network Metrics</h2>
                <div className="flex items-center gap-2 text-gray-400 text-sm">
                  <RefreshCw className="w-4 h-4 animate-spin" />
                  <span>Live updates every 5s</span>
                </div>
              </div>

              {/* Big Stat Cards - Grafana Style */}
              <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
                <BigStatCard
                  label="Active VPN Connections"
                  value={nodeMetrics.activeConnections}
                />
                <BigStatCard
                  label="P2P Connected Peers"
                  value={nodeMetrics.p2pConnectedPeers}
                />
                <BigStatCard
                  label="Known VPN Nodes"
                  value={nodeMetrics.knownNodes}
                />
                <BigStatCard
                  label="Node Status"
                  value={nodeMetrics.nodeStatus}
                  status={true}
                  color={nodeMetrics.nodeStatus === 'UP' ? 'green' : 'red'}
                />
              </div>

              {/* Charts Grid */}
              <div className="grid lg:grid-cols-2 gap-4">
                {/* Network Traffic Chart */}
                <ChartPanel title="Network Traffic">
                  <ResponsiveContainer width="100%" height="100%">
                    <AreaChart data={trafficData}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                      <XAxis dataKey="time" stroke="#6B7280" tick={{ fontSize: 10 }} />
                      <YAxis stroke="#6B7280" tick={{ fontSize: 10 }} tickFormatter={(v) => `${v} B`} />
                      <Tooltip content={<CustomTooltip />} />
                      <Legend wrapperStyle={{ fontSize: '12px' }} />
                      <Area
                        type="monotone"
                        dataKey="received"
                        name="Received"
                        stroke="#10B981"
                        fill="#10B981"
                        fillOpacity={0.3}
                      />
                      <Area
                        type="monotone"
                        dataKey="sent"
                        name="Sent"
                        stroke="#F59E0B"
                        fill="#F59E0B"
                        fillOpacity={0.3}
                      />
                    </AreaChart>
                  </ResponsiveContainer>
                </ChartPanel>

                {/* VPN Connections Over Time */}
                <ChartPanel title="VPN Connections Over Time">
                  <ResponsiveContainer width="100%" height="100%">
                    <LineChart data={connectionsData}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                      <XAxis dataKey="time" stroke="#6B7280" tick={{ fontSize: 10 }} />
                      <YAxis stroke="#6B7280" tick={{ fontSize: 10 }} />
                      <Tooltip content={<CustomTooltip />} />
                      <Legend wrapperStyle={{ fontSize: '12px' }} />
                      <Line
                        type="monotone"
                        dataKey="connections"
                        name="Active Connections"
                        stroke="#3B82F6"
                        strokeWidth={2}
                        dot={false}
                      />
                    </LineChart>
                  </ResponsiveContainer>
                </ChartPanel>

                {/* Memory Usage Chart */}
                <ChartPanel title="Memory Usage">
                  <ResponsiveContainer width="100%" height="100%">
                    <AreaChart data={memoryData}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                      <XAxis dataKey="time" stroke="#6B7280" tick={{ fontSize: 10 }} />
                      <YAxis stroke="#6B7280" tick={{ fontSize: 10 }} tickFormatter={(v) => `${v}%`} domain={[0, 100]} />
                      <Tooltip content={<CustomTooltip />} />
                      <Legend wrapperStyle={{ fontSize: '12px' }} />
                      <Area
                        type="monotone"
                        dataKey="memory"
                        name="Memory %"
                        stroke="#8B5CF6"
                        fill="#8B5CF6"
                        fillOpacity={0.3}
                      />
                    </AreaChart>
                  </ResponsiveContainer>
                </ChartPanel>

                {/* P2P Network Chart */}
                <ChartPanel title="P2P Network">
                  <ResponsiveContainer width="100%" height="100%">
                    <LineChart data={p2pData}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                      <XAxis dataKey="time" stroke="#6B7280" tick={{ fontSize: 10 }} />
                      <YAxis stroke="#6B7280" tick={{ fontSize: 10 }} />
                      <Tooltip content={<CustomTooltip />} />
                      <Legend wrapperStyle={{ fontSize: '12px' }} />
                      <Line
                        type="monotone"
                        dataKey="connectedPeers"
                        name="Connected Peers"
                        stroke="#10B981"
                        strokeWidth={2}
                        dot={false}
                      />
                      <Line
                        type="monotone"
                        dataKey="knownNodes"
                        name="Known Nodes"
                        stroke="#6366F1"
                        strokeWidth={2}
                        dot={false}
                      />
                    </LineChart>
                  </ResponsiveContainer>
                </ChartPanel>
              </div>
            </motion.div>
          </>
        )}

        {activeTab === 'nodes' && (
          <>
            {/* Nodes Header */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-8"
            >
              <div>
                <h1 className="text-3xl font-bold text-white mb-2">VPN Nodes</h1>
                <p className="text-gray-400">{nodes.length} nodes available across {countries.length} countries</p>
              </div>

              <div className="flex items-center gap-3">
                {/* Search */}
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-500" />
                  <input
                    type="text"
                    placeholder="Search nodes..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="pl-10 pr-4 py-2 bg-dark-800 border border-white/10 rounded-lg text-white placeholder-gray-500 focus:border-gold-500 focus:outline-none w-full sm:w-64"
                  />
                </div>

                {/* Filter */}
                <select
                  value={selectedCountry}
                  onChange={(e) => setSelectedCountry(e.target.value)}
                  className="px-4 py-2 bg-dark-800 border border-white/10 rounded-lg text-white focus:border-gold-500 focus:outline-none"
                >
                  <option value="all">All Countries</option>
                  {countries.map((code) => (
                    <option key={code} value={code}>{code}</option>
                  ))}
                </select>
              </div>
            </motion.div>

            {/* Nodes Grid */}
            <div className="grid sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
              {loadingNodes ? (
                <div className="col-span-full flex items-center justify-center py-12">
                  <Loader2 className="w-8 h-8 text-gold-500 animate-spin" />
                </div>
              ) : filteredNodes.length === 0 ? (
                <div className="col-span-full text-center py-12">
                  <Server className="w-12 h-12 text-gray-600 mx-auto mb-4" />
                  <h3 className="text-xl font-bold text-white mb-2">No nodes found</h3>
                  <p className="text-gray-400">Try adjusting your search or filter</p>
                </div>
              ) : (
                filteredNodes.map((node) => (
                  <NodeCard
                    key={node.id}
                    node={node}
                    onConnect={handleConnect}
                    onDisconnect={handleDisconnect}
                    isConnecting={isConnecting}
                    isConnected={connectedNode?.id === node.id}
                  />
                ))
              )}
            </div>
          </>
        )}

        {activeTab === 'settings' && (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
          >
            <h1 className="text-3xl font-bold text-white mb-8">Settings</h1>

            <div className="space-y-6">
              {/* Account Settings */}
              <div className="glass rounded-xl p-6">
                <h2 className="text-xl font-bold text-white mb-4">Account</h2>
                <div className="space-y-4">
                  <div className="flex items-center justify-between py-3 border-b border-white/5">
                    <div>
                      <div className="font-medium text-white">Email</div>
                      <div className="text-gray-400 text-sm">{user?.email || 'Not set'}</div>
                    </div>
                    <button className="text-gold-500 text-sm hover:text-gold-400">Change</button>
                  </div>
                  <div className="flex items-center justify-between py-3 border-b border-white/5">
                    <div>
                      <div className="font-medium text-white">Username</div>
                      <div className="text-gray-400 text-sm">{user?.username || 'Not set'}</div>
                    </div>
                    <button className="text-gold-500 text-sm hover:text-gold-400">Change</button>
                  </div>
                  <div className="flex items-center justify-between py-3">
                    <div>
                      <div className="font-medium text-white">Password</div>
                      <div className="text-gray-400 text-sm">Last changed: Never</div>
                    </div>
                    <button className="text-gold-500 text-sm hover:text-gold-400">Change</button>
                  </div>
                </div>
              </div>

              {/* VPN Settings */}
              <div className="glass rounded-xl p-6">
                <h2 className="text-xl font-bold text-white mb-4">VPN Preferences</h2>
                <div className="space-y-4">
                  <div className="flex items-center justify-between py-3 border-b border-white/5">
                    <div>
                      <div className="font-medium text-white">Protocol</div>
                      <div className="text-gray-400 text-sm">Default VPN protocol</div>
                    </div>
                    <select className="bg-dark-800 border border-white/10 rounded-lg px-3 py-2 text-white">
                      <option>WireGuard</option>
                      <option>OpenVPN</option>
                    </select>
                  </div>
                  <div className="flex items-center justify-between py-3 border-b border-white/5">
                    <div>
                      <div className="font-medium text-white">Kill Switch</div>
                      <div className="text-gray-400 text-sm">Block internet if VPN disconnects</div>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input type="checkbox" className="sr-only peer" />
                      <div className="w-11 h-6 bg-dark-800 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-gold-500"></div>
                    </label>
                  </div>
                  <div className="flex items-center justify-between py-3">
                    <div>
                      <div className="font-medium text-white">Auto-connect</div>
                      <div className="text-gray-400 text-sm">Connect automatically on app start</div>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input type="checkbox" className="sr-only peer" />
                      <div className="w-11 h-6 bg-dark-800 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-gold-500"></div>
                    </label>
                  </div>
                </div>
              </div>
            </div>
          </motion.div>
        )}
      </main>
    </div>
  );
}

// Mock nodes for demo
const mockNodes = [
  { id: '1', name: 'US East 1', city: 'New York', country: 'United States', country_code: 'US', status: 'online', supports_wireguard: true, supports_openvpn: true, load_score: 35 },
  { id: '2', name: 'US West 1', city: 'Los Angeles', country: 'United States', country_code: 'US', status: 'online', supports_wireguard: true, supports_openvpn: false, load_score: 45 },
  { id: '3', name: 'EU Central 1', city: 'Frankfurt', country: 'Germany', country_code: 'DE', status: 'online', supports_wireguard: true, supports_openvpn: true, supports_multihop: true, load_score: 60 },
  { id: '4', name: 'UK South 1', city: 'London', country: 'United Kingdom', country_code: 'GB', status: 'online', supports_wireguard: true, supports_openvpn: true, load_score: 25 },
  { id: '5', name: 'JP East 1', city: 'Tokyo', country: 'Japan', country_code: 'JP', status: 'online', supports_wireguard: true, supports_openvpn: false, load_score: 55 },
  { id: '6', name: 'SG Central 1', city: 'Singapore', country: 'Singapore', country_code: 'SG', status: 'online', supports_wireguard: true, supports_openvpn: true, load_score: 40 },
  { id: '7', name: 'AU East 1', city: 'Sydney', country: 'Australia', country_code: 'AU', status: 'online', supports_wireguard: true, supports_openvpn: true, load_score: 30 },
  { id: '8', name: 'CA Central 1', city: 'Toronto', country: 'Canada', country_code: 'CA', status: 'busy', supports_wireguard: true, supports_openvpn: false, load_score: 78 },
];
