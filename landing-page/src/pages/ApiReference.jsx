import React, { useState } from 'react';
import { motion } from 'framer-motion';
import PageLayout from '../components/PageLayout';
import {
  Code2,
  Key,
  Server,
  Users,
  Zap,
  Copy,
  Check,
  ChevronDown,
  ChevronRight
} from 'lucide-react';

function CodeBlock({ children, title, language = 'bash' }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(children);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="rounded-xl overflow-hidden border border-white/10">
      {title && (
        <div className="bg-dark-800 px-4 py-2 border-b border-white/10 flex items-center justify-between">
          <span className="text-sm text-gray-400">{title}</span>
          <button
            onClick={handleCopy}
            className="text-gray-400 hover:text-white transition-colors"
          >
            {copied ? <Check className="w-4 h-4 text-green-500" /> : <Copy className="w-4 h-4" />}
          </button>
        </div>
      )}
      <pre className="bg-dark-900 p-4 overflow-x-auto">
        <code className="text-sm text-gray-300 font-mono whitespace-pre">{children}</code>
      </pre>
    </div>
  );
}

function Endpoint({ method, path, description, request, response, auth = true }) {
  const [expanded, setExpanded] = useState(false);

  const methodColors = {
    GET: 'bg-green-500/20 text-green-400',
    POST: 'bg-blue-500/20 text-blue-400',
    PUT: 'bg-yellow-500/20 text-yellow-400',
    DELETE: 'bg-red-500/20 text-red-400',
  };

  return (
    <div className="glass rounded-xl overflow-hidden border border-white/5 hover:border-gold-500/20 transition-all">
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full px-6 py-4 flex items-center justify-between hover:bg-white/5 transition-colors"
      >
        <div className="flex items-center gap-4">
          <span className={`px-3 py-1 rounded-md text-xs font-bold ${methodColors[method]}`}>
            {method}
          </span>
          <code className="text-white font-mono">{path}</code>
          {auth && <Key className="w-4 h-4 text-gold-500" title="Requires authentication" />}
        </div>
        <div className="flex items-center gap-4">
          <span className="text-gray-400 text-sm hidden md:block">{description}</span>
          {expanded ? <ChevronDown className="w-5 h-5 text-gray-400" /> : <ChevronRight className="w-5 h-5 text-gray-400" />}
        </div>
      </button>

      {expanded && (
        <motion.div
          initial={{ opacity: 0, height: 0 }}
          animate={{ opacity: 1, height: 'auto' }}
          exit={{ opacity: 0, height: 0 }}
          className="border-t border-white/5 p-6 space-y-6"
        >
          <p className="text-gray-400 md:hidden">{description}</p>

          {request && (
            <div>
              <h4 className="text-white font-medium mb-3">Request</h4>
              <CodeBlock title="JSON">{request}</CodeBlock>
            </div>
          )}

          {response && (
            <div>
              <h4 className="text-white font-medium mb-3">Response</h4>
              <CodeBlock title="JSON">{response}</CodeBlock>
            </div>
          )}
        </motion.div>
      )}
    </div>
  );
}

export default function ApiReference() {
  return (
    <PageLayout
      title="API Reference"
      description="Complete API documentation for integrating with Aureo VPN services."
    >
      {/* Base URL */}
      <div className="glass rounded-xl p-6 mb-8">
        <h3 className="text-lg font-bold text-white mb-4">Base URL</h3>
        <div className="grid md:grid-cols-2 gap-4">
          <div>
            <span className="text-gray-400 text-sm">Production</span>
            <code className="block text-gold-500 font-mono mt-1">https://api.aureo-vpn.com/api/v1</code>
          </div>
          <div>
            <span className="text-gray-400 text-sm">Staging</span>
            <code className="block text-gray-400 font-mono mt-1">https://staging-api.aureo-vpn.com/api/v1</code>
          </div>
        </div>
      </div>

      {/* Rate Limiting */}
      <div className="glass rounded-xl p-6 mb-8">
        <h3 className="text-lg font-bold text-white mb-4">Rate Limiting</h3>
        <div className="grid md:grid-cols-3 gap-4 mb-4">
          <div className="text-center p-4 bg-dark-800 rounded-lg">
            <span className="text-gray-400 text-sm block">Anonymous</span>
            <span className="text-2xl font-bold text-white">100</span>
            <span className="text-gray-500 text-sm">/minute</span>
          </div>
          <div className="text-center p-4 bg-dark-800 rounded-lg">
            <span className="text-gray-400 text-sm block">Authenticated</span>
            <span className="text-2xl font-bold text-gold-500">1,000</span>
            <span className="text-gray-500 text-sm">/minute</span>
          </div>
          <div className="text-center p-4 bg-dark-800 rounded-lg">
            <span className="text-gray-400 text-sm block">Premium</span>
            <span className="text-2xl font-bold text-cyan-400">5,000</span>
            <span className="text-gray-500 text-sm">/minute</span>
          </div>
        </div>
        <p className="text-gray-400 text-sm">
          Rate limit headers included in all responses: <code className="text-gold-500">X-RateLimit-Limit</code>, <code className="text-gold-500">X-RateLimit-Remaining</code>, <code className="text-gold-500">X-RateLimit-Reset</code>
        </p>
      </div>

      {/* Authentication Section */}
      <section id="auth" className="mb-12">
        <div className="flex items-center gap-3 mb-6">
          <div className="w-10 h-10 rounded-lg bg-gold-500/10 flex items-center justify-center">
            <Key className="w-5 h-5 text-gold-500" />
          </div>
          <h2 className="text-2xl font-bold text-white">Authentication</h2>
        </div>

        <div className="space-y-4">
          <Endpoint
            method="POST"
            path="/auth/register"
            description="Create a new user account"
            auth={false}
            request={`{
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "username": "username"
}`}
            response={`{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "username",
    "subscription_tier": "free"
  }
}`}
          />

          <Endpoint
            method="POST"
            path="/auth/login"
            description="Authenticate a user"
            auth={false}
            request={`{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}`}
            response={`{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "uuid",
    "email": "user@example.com"
  }
}`}
          />

          <Endpoint
            method="POST"
            path="/auth/refresh"
            description="Refresh an access token"
            auth={false}
            request={`{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}`}
            response={`{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}`}
          />
        </div>
      </section>

      {/* Nodes Section */}
      <section id="nodes" className="mb-12">
        <div className="flex items-center gap-3 mb-6">
          <div className="w-10 h-10 rounded-lg bg-gold-500/10 flex items-center justify-center">
            <Server className="w-5 h-5 text-gold-500" />
          </div>
          <h2 className="text-2xl font-bold text-white">VPN Nodes</h2>
        </div>

        <div className="space-y-4">
          <Endpoint
            method="GET"
            path="/nodes"
            description="List available VPN nodes"
            response={`{
  "nodes": [
    {
      "id": "uuid",
      "name": "US-East-1",
      "hostname": "us-east-1.aureo-vpn.com",
      "public_ip": "192.0.2.1",
      "country": "United States",
      "country_code": "US",
      "city": "New York",
      "load_score": 25.5,
      "current_connections": 250,
      "max_connections": 1000,
      "latency": 15,
      "status": "online",
      "supports_wireguard": true,
      "supports_openvpn": true
    }
  ],
  "total": 5234
}`}
          />

          <Endpoint
            method="GET"
            path="/nodes/best"
            description="Get the best node based on criteria"
            response={`{
  "id": "uuid",
  "name": "US-East-1",
  "country": "United States",
  "load_score": 12.3,
  "latency": 8,
  "recommendation_reason": "Lowest latency and load"
}`}
          />

          <Endpoint
            method="GET"
            path="/nodes/:id"
            description="Get details for a specific node"
            response={`{
  "id": "uuid",
  "name": "US-East-1",
  "hostname": "us-east-1.aureo-vpn.com",
  "public_ip": "192.0.2.1",
  "load_score": 25.5,
  "bandwidth_usage_gbps": 2.5,
  "cpu_usage": 35.2,
  "status": "online"
}`}
          />
        </div>
      </section>

      {/* Sessions Section */}
      <section id="sessions" className="mb-12">
        <div className="flex items-center gap-3 mb-6">
          <div className="w-10 h-10 rounded-lg bg-gold-500/10 flex items-center justify-center">
            <Zap className="w-5 h-5 text-gold-500" />
          </div>
          <h2 className="text-2xl font-bold text-white">VPN Sessions</h2>
        </div>

        <div className="space-y-4">
          <Endpoint
            method="POST"
            path="/sessions"
            description="Create a new VPN session"
            request={`{
  "node_id": "uuid",
  "protocol": "wireguard",
  "enable_kill_switch": true,
  "enable_dns_protection": true
}`}
            response={`{
  "session": {
    "id": "uuid",
    "node_id": "uuid",
    "protocol": "wireguard",
    "tunnel_ip": "10.8.0.5",
    "status": "active",
    "connected_at": "2024-01-15T10:00:00Z"
  },
  "config": "[Interface]\\nPrivateKey = ...\\n..."
}`}
          />

          <Endpoint
            method="DELETE"
            path="/sessions/:id"
            description="Disconnect a VPN session"
            response={`{
  "message": "Session disconnected successfully",
  "session_id": "uuid",
  "duration_seconds": 3600,
  "data_used_gb": 1.2
}`}
          />

          <Endpoint
            method="GET"
            path="/user/sessions"
            description="Get active VPN sessions"
            response={`{
  "sessions": [
    {
      "id": "uuid",
      "node": {
        "id": "uuid",
        "name": "US-East-1",
        "country": "United States"
      },
      "protocol": "wireguard",
      "tunnel_ip": "10.8.0.5",
      "status": "active",
      "data_used_gb": 2.5
    }
  ],
  "count": 1
}`}
          />
        </div>
      </section>

      {/* Operator Section */}
      <section id="operator" className="mb-12">
        <div className="flex items-center gap-3 mb-6">
          <div className="w-10 h-10 rounded-lg bg-gold-500/10 flex items-center justify-center">
            <Users className="w-5 h-5 text-gold-500" />
          </div>
          <h2 className="text-2xl font-bold text-white">Operator Endpoints</h2>
        </div>

        <div className="space-y-4">
          <Endpoint
            method="POST"
            path="/operator/register"
            description="Register as a node operator"
            request={`{
  "wallet_address": "0x1234...",
  "wallet_currency": "ethereum",
  "company_name": "My VPN LLC",
  "contact_email": "contact@myvpn.com"
}`}
            response={`{
  "operator": {
    "id": "uuid",
    "status": "pending",
    "wallet_address": "0x1234..."
  },
  "message": "Registration submitted. Awaiting verification."
}`}
          />

          <Endpoint
            method="POST"
            path="/operator/nodes"
            description="Create a new operator node"
            request={`{
  "name": "us-east-1",
  "hostname": "vpn.example.com",
  "ip_address": "203.0.113.1",
  "country": "US",
  "city": "New York",
  "wireguard_port": 51820,
  "max_connections": 500
}`}
            response={`{
  "node": {
    "id": "uuid",
    "name": "us-east-1",
    "wireguard_public_key": "base64...",
    "subnet": "10.8.0.0/24"
  },
  "private_key": "base64..."
}`}
          />

          <Endpoint
            method="GET"
            path="/operator/dashboard"
            description="Get operator dashboard stats"
            response={`{
  "total_earned_usd": 156.78,
  "pending_payout_usd": 23.45,
  "active_nodes": 2,
  "total_bandwidth_gb": 5678.90,
  "average_uptime_percent": 99.2,
  "reputation_score": 87,
  "current_tier": "gold"
}`}
          />

          <Endpoint
            method="POST"
            path="/operator/payout/request"
            description="Request a cryptocurrency payout"
            response={`{
  "payout": {
    "id": "uuid",
    "amount_usd": 23.45,
    "amount_crypto": "0.0117",
    "currency": "ethereum",
    "status": "pending"
  }
}`}
          />
        </div>
      </section>

      {/* Error Codes */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold text-white mb-6">Error Codes</h2>
        <div className="glass rounded-xl overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-gray-400 border-b border-white/10">
                <th className="px-6 py-4">Code</th>
                <th className="px-6 py-4">Status</th>
                <th className="px-6 py-4">Description</th>
              </tr>
            </thead>
            <tbody className="text-gray-300">
              <tr className="border-b border-white/5">
                <td className="px-6 py-4 font-mono text-red-400">UNAUTHORIZED</td>
                <td className="px-6 py-4">401</td>
                <td className="px-6 py-4">Invalid or expired token</td>
              </tr>
              <tr className="border-b border-white/5">
                <td className="px-6 py-4 font-mono text-red-400">FORBIDDEN</td>
                <td className="px-6 py-4">403</td>
                <td className="px-6 py-4">Insufficient permissions</td>
              </tr>
              <tr className="border-b border-white/5">
                <td className="px-6 py-4 font-mono text-yellow-400">NOT_FOUND</td>
                <td className="px-6 py-4">404</td>
                <td className="px-6 py-4">Resource not found</td>
              </tr>
              <tr className="border-b border-white/5">
                <td className="px-6 py-4 font-mono text-yellow-400">VALIDATION_ERROR</td>
                <td className="px-6 py-4">400</td>
                <td className="px-6 py-4">Invalid request data</td>
              </tr>
              <tr className="border-b border-white/5">
                <td className="px-6 py-4 font-mono text-orange-400">RATE_LIMIT_EXCEEDED</td>
                <td className="px-6 py-4">429</td>
                <td className="px-6 py-4">Too many requests</td>
              </tr>
              <tr>
                <td className="px-6 py-4 font-mono text-red-400">SERVER_ERROR</td>
                <td className="px-6 py-4">500</td>
                <td className="px-6 py-4">Internal server error</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      {/* SDK Examples */}
      <section>
        <h2 className="text-2xl font-bold text-white mb-6">SDK Examples</h2>
        <div className="grid md:grid-cols-2 gap-6">
          <div>
            <h4 className="text-lg font-bold text-white mb-3">cURL</h4>
            <CodeBlock title="bash">
{`# Login
curl -X POST https://api.aureo-vpn.com/api/v1/auth/login \\
  -H "Content-Type: application/json" \\
  -d '{"email":"user@example.com","password":"password"}'

# Get nodes
curl -X GET "https://api.aureo-vpn.com/api/v1/nodes?country=US" \\
  -H "Authorization: Bearer <token>"`}
            </CodeBlock>
          </div>
          <div>
            <h4 className="text-lg font-bold text-white mb-3">JavaScript</h4>
            <CodeBlock title="javascript">
{`const client = axios.create({
  baseURL: 'https://api.aureo-vpn.com/api/v1',
  headers: {
    'Authorization': \`Bearer \${token}\`
  }
});

const { data } = await client.get('/nodes', {
  params: { country: 'US', protocol: 'wireguard' }
});`}
            </CodeBlock>
          </div>
        </div>
      </section>
    </PageLayout>
  );
}
