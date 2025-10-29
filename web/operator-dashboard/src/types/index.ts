export interface User {
  id: string;
  email: string;
  username: string;
  is_active: boolean;
  created_at: string;
}

export interface Operator {
  id: string;
  user_id: string;
  wallet_address: string;
  wallet_type: string;
  status: string;
  is_verified: boolean;
  reputation_score: number;
  total_earned: number;
  pending_payout: number;
  country: string;
  created_at: string;
}

export interface VPNNode {
  id: string;
  name: string;
  hostname: string;
  public_ip: string;
  country: string;
  country_code: string;
  city: string;
  status: string;
  is_active: boolean;
  uptime_percentage: number;
  total_earned_usd: number;
  bandwidth_usage_gbps?: number;
  current_connections?: number;
  created_at: string;
}

export interface OperatorStats {
  total_earned: number;
  pending_payout: number;
  total_paid: number;
  active_nodes: number;
  total_sessions: number;
  total_bandwidth_gb: number;
  reputation_score: number;
  current_tier: string;
  monthly_earnings_estimate: number;
}

export interface Earning {
  id: string;
  operator_id: string;
  node_id: string;
  session_id: string;
  bandwidth_gb: number;
  duration_minutes: number;
  rate_per_gb: number;
  amount_usd: number;
  status: string;
  connection_quality: number;
  user_rating?: number;
  created_at: string;
}

export interface Payout {
  id: string;
  operator_id: string;
  amount_usd: number;
  crypto_amount: number;
  crypto_currency: string;
  exchange_rate: number;
  wallet_address: string;
  status: string;
  transaction_hash?: string;
  payout_method: string;
  created_at: string;
  processed_at?: string;
  completed_at?: string;
}

export interface RewardTier {
  id: string;
  tier_name: string;
  min_reputation_score: number;
  min_uptime_percent: number;
  base_rate_per_gb: number;
  bonus_multiplier: number;
  min_bandwidth: number;
  max_latency: number;
  is_active: boolean;
}

export interface DashboardData {
  operator: Operator;
  stats: OperatorStats;
  active_nodes: VPNNode[];
  recent_earnings: Earning[];
  recent_payouts: Payout[];
}

export interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isOperator: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, username: string) => Promise<void>;
  logout: () => void;
  loading: boolean;
}
