import axios from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';
// Base URL without /api/v1 for root endpoints like /info, /health
const ROOT_URL = API_BASE_URL.replace('/api/v1', '');

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

const rootApi = axios.create({
  baseURL: ROOT_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add auth token to requests
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle 401 - redirect to login
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      // Token invalid, clear and redirect
      localStorage.removeItem('access_token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// Auth API
export const authApi = {
  register: (data) => api.post('/auth/register', data),
  login: (data) => api.post('/auth/login', data),
  getProfile: () => api.get('/me'),
};

// Nodes API
export const nodesApi = {
  list: (params) => api.get('/nodes', { params }),
  getBest: (params) => api.get('/nodes/best', { params }),
  getById: (id) => api.get(`/nodes/${id}`),
  getCountries: () => api.get('/nodes/countries'),
};

// VPN Connection API
export const vpnApi = {
  connect: (data) => api.post('/connect', data),
};

// Sessions API
export const sessionsApi = {
  create: (data) => api.post('/sessions', data),
  delete: (id) => api.delete(`/sessions/${id}`),
  list: () => api.get('/sessions'),
};

// Operator API
export const operatorApi = {
  register: (data) => api.post('/operators/register', data),
  getStats: () => api.get('/operators/stats'),
  getEarnings: () => api.get('/operators/earnings'),
  requestPayout: (data) => api.post('/operators/payout', data),
};

// Node Metrics API (for Grafana-style dashboard)
export const metricsApi = {
  // Get node info (current_connections, load_score, status)
  getNodeInfo: () => rootApi.get('/info'),
  // Get health check
  getHealth: () => rootApi.get('/health'),
  // Get P2P network status
  getP2PStatus: () => api.get('/p2p/status'),
  // Get P2P peers
  getPeers: () => api.get('/p2p/peers'),
  // Get raw Prometheus metrics and parse them
  getPrometheusMetrics: async () => {
    const response = await rootApi.get('/metrics', { responseType: 'text' });
    return parsePrometheusMetrics(response.data);
  },
};

// Parse Prometheus text format into JSON
function parsePrometheusMetrics(text) {
  const metrics = {
    bytesReceived: 0,
    bytesSent: 0,
    memoryUsagePercent: 0,
    cpuUsagePercent: 0,
    activeConnections: 0,
    p2pConnectedPeers: 0,
    p2pKnownNodes: 0,
    nodeLoadScore: 0,
  };

  // Regex pattern that handles scientific notation (e.g., 1.679965168e+09)
  const numberPattern = /[\d.]+(?:e[+-]?\d+)?/i;

  const lines = text.split('\n');
  for (const line of lines) {
    // Skip comments and empty lines
    if (line.startsWith('#') || !line.trim()) continue;

    // Parse metric lines - extract the numeric value at the end
    if (line.startsWith('aureo_vpn_bytes_received_total')) {
      const match = line.match(new RegExp(`aureo_vpn_bytes_received_total\\s+(${numberPattern.source})`, 'i'));
      if (match) metrics.bytesReceived = parseFloat(match[1]);
    }
    else if (line.startsWith('aureo_vpn_bytes_sent_total')) {
      const match = line.match(new RegExp(`aureo_vpn_bytes_sent_total\\s+(${numberPattern.source})`, 'i'));
      if (match) metrics.bytesSent = parseFloat(match[1]);
    }
    else if (line.startsWith('aureo_vpn_node_memory_usage_percent')) {
      const match = line.match(new RegExp(`aureo_vpn_node_memory_usage_percent\\{[^}]*\\}\\s+(${numberPattern.source})`, 'i'));
      if (match) metrics.memoryUsagePercent = parseFloat(match[1]);
    }
    else if (line.startsWith('aureo_vpn_node_cpu_usage_percent')) {
      const match = line.match(new RegExp(`aureo_vpn_node_cpu_usage_percent\\{[^}]*\\}\\s+(${numberPattern.source})`, 'i'));
      if (match) metrics.cpuUsagePercent = parseFloat(match[1]);
    }
    else if (line.startsWith('aureo_vpn_active_connections')) {
      const match = line.match(new RegExp(`aureo_vpn_active_connections\\{[^}]*\\}\\s+(${numberPattern.source})`, 'i'));
      if (match) metrics.activeConnections = parseFloat(match[1]);
    }
    else if (line.startsWith('aureo_p2p_connected_peers')) {
      const match = line.match(new RegExp(`aureo_p2p_connected_peers\\s+(${numberPattern.source})`, 'i'));
      if (match) metrics.p2pConnectedPeers = parseFloat(match[1]);
    }
    else if (line.startsWith('aureo_p2p_known_nodes')) {
      const match = line.match(new RegExp(`aureo_p2p_known_nodes\\s+(${numberPattern.source})`, 'i'));
      if (match) metrics.p2pKnownNodes = parseFloat(match[1]);
    }
    else if (line.startsWith('aureo_vpn_node_load_score')) {
      const match = line.match(new RegExp(`aureo_vpn_node_load_score\\{[^}]*\\}\\s+(${numberPattern.source})`, 'i'));
      if (match) metrics.nodeLoadScore = parseFloat(match[1]);
    }
  }

  return metrics;
}

export default api;
