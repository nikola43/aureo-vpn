import axios, { AxiosInstance } from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

class ApiService {
  private api: AxiosInstance;

  constructor() {
    this.api = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Add auth token to requests
    this.api.interceptors.request.use((config) => {
      const token = localStorage.getItem('access_token');
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
      return config;
    });

    // Handle auth errors
    this.api.interceptors.response.use(
      (response) => response,
      async (error) => {
        if (error.response?.status === 401) {
          localStorage.removeItem('access_token');
          window.location.href = '/login';
        }
        return Promise.reject(error);
      }
    );
  }

  // Auth
  async register(data: { email: string; password: string; username: string }) {
    const response = await this.api.post('/auth/register', data);
    if (response.data.access_token) {
      localStorage.setItem('access_token', response.data.access_token);
    }
    return response.data;
  }

  async login(data: { email: string; password: string }) {
    const response = await this.api.post('/auth/login', data);
    if (response.data.access_token) {
      localStorage.setItem('access_token', response.data.access_token);
    }
    return response.data;
  }

  logout() {
    localStorage.removeItem('access_token');
  }

  // Operator
  async registerAsOperator(data: {
    wallet_address: string;
    wallet_type: string;
    country: string;
    email: string;
    phone_number?: string;
  }) {
    const response = await this.api.post('/operator/register', data);
    return response.data;
  }

  async getOperatorDashboard() {
    const response = await this.api.get('/operator/dashboard');
    return response.data;
  }

  async getOperatorStats() {
    const response = await this.api.get('/operator/stats');
    return response.data;
  }

  async getOperatorNodes() {
    const response = await this.api.get('/operator/nodes');
    return response.data;
  }

  async createNode(data: {
    name: string;
    hostname: string;
    public_ip: string;
    country: string;
    country_code: string;
    city: string;
    wireguard_port: number;
    openvpn_port: number;
    latitude?: number;
    longitude?: number;
  }) {
    const response = await this.api.post('/operator/nodes', data);
    return response.data;
  }

  async getEarnings(limit = 50, offset = 0) {
    const response = await this.api.get(`/operator/earnings?limit=${limit}&offset=${offset}`);
    return response.data;
  }

  async getPayouts(limit = 50, offset = 0) {
    const response = await this.api.get(`/operator/payouts?limit=${limit}&offset=${offset}`);
    return response.data;
  }

  async requestPayout() {
    const response = await this.api.post('/operator/payout/request');
    return response.data;
  }

  async getRewardTiers() {
    const response = await this.api.get('/operator/rewards/tiers');
    return response.data;
  }
}

export const api = new ApiService();
