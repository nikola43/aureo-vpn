# Aureo VPN Operator Dashboard

A modern, responsive React dashboard for Aureo VPN node operators to track earnings, manage nodes, and monitor performance.

## Features

### 📊 Comprehensive Dashboard
- **Real-time stats**: Total earned, pending payout, active nodes, reputation score
- **Interactive charts**: Earnings trend and bandwidth usage graphs
- **Node monitoring**: Track all your nodes' status, uptime, and earnings
- **Recent activity**: View latest earnings and payouts

### 💰 Earnings & Payouts
- Track all earnings with detailed breakdowns
- Monitor pending and completed payouts
- Request manual payouts (minimum $10)
- View transaction hashes on blockchain

### 🖥️ Node Management
- View all active nodes at a glance
- Monitor uptime percentage
- Track per-node earnings
- Real-time status indicators

### 🎨 Modern UI/UX
- Clean, professional design with Tailwind CSS
- Fully responsive (mobile, tablet, desktop)
- Dark mode support (coming soon)
- Real-time data updates

## Tech Stack

- **React 18** - UI library
- **TypeScript** - Type safety
- **Vite** - Fast build tool
- **Tailwind CSS** - Utility-first styling
- **Recharts** - Beautiful charts
- **Lucide React** - Modern icons
- **React Router** - Client-side routing
- **Axios** - HTTP client

## Prerequisites

- Node.js 18+ and npm/yarn
- Aureo VPN API running on `localhost:8080` (or configured endpoint)

## Quick Start

### 1. Install Dependencies

```bash
cd web/operator-dashboard
npm install
```

### 2. Configure Environment

```bash
# Copy example env file
cp .env.example .env

# Edit .env if needed
# VITE_API_URL=http://localhost:8080/api/v1
```

### 3. Run Development Server

```bash
npm run dev
```

Dashboard will be available at `http://localhost:3000`

### 4. Login or Register

- **Register**: Create a new account if you don't have one
- **Login**: Use your existing credentials
- You'll need to register as an operator after logging in

## Development

### Project Structure

```
operator-dashboard/
├── src/
│   ├── components/       # Reusable UI components
│   ├── contexts/         # React contexts (Auth, etc.)
│   ├── pages/           # Page components
│   │   ├── Dashboard.tsx
│   │   └── Login.tsx
│   ├── services/        # API service layer
│   │   └── api.ts
│   ├── types/           # TypeScript definitions
│   │   └── index.ts
│   ├── utils/           # Utility functions
│   ├── App.tsx          # Main app component
│   ├── main.tsx         # Entry point
│   └── index.css        # Global styles
├── public/              # Static assets
├── package.json         # Dependencies
├── tsconfig.json        # TypeScript config
├── vite.config.ts       # Vite configuration
└── tailwind.config.js   # Tailwind configuration
```

### Available Scripts

```bash
# Development server with hot reload
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Lint code
npm run lint
```

### Key Files

**API Service** (`src/services/api.ts`):
- Handles all API communication
- Automatic token management
- Error handling and retries

**Auth Context** (`src/contexts/AuthContext.tsx`):
- Manages authentication state
- Provides login/logout methods
- Protects routes

**Dashboard Page** (`src/pages/Dashboard.tsx`):
- Main operator dashboard
- Stats, charts, nodes, earnings, payouts
- Request payout functionality

**Login Page** (`src/pages/Login.tsx`):
- Combined login/register interface
- Form validation
- Error handling

## API Integration

The dashboard communicates with the Aureo VPN API:

### Base URL
```
http://localhost:8080/api/v1
```

### Authentication
JWT tokens stored in localStorage:
```
Authorization: Bearer <access_token>
```

### Key Endpoints Used

```
POST   /auth/register              - User registration
POST   /auth/login                 - User login
GET    /operator/dashboard         - Dashboard data
GET    /operator/stats             - Operator statistics
GET    /operator/nodes             - List operator nodes
POST   /operator/nodes             - Create new node
GET    /operator/earnings          - Earnings history
GET    /operator/payouts           - Payout history
POST   /operator/payout/request    - Request payout
GET    /operator/rewards/tiers     - Reward tiers
```

## Building for Production

### 1. Build

```bash
npm run build
```

This creates optimized files in the `dist/` directory.

### 2. Deploy

#### Static Hosting (Vercel, Netlify)

```bash
# Vercel
vercel --prod

# Netlify
netlify deploy --prod
```

#### Docker

```bash
# Build image
docker build -t aureo-dashboard .

# Run container
docker run -p 3000:80 aureo-dashboard
```

#### Nginx

```nginx
server {
    listen 80;
    server_name dashboard.aureo-vpn.com;
    root /var/www/aureo-dashboard;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://localhost:8080;
    }
}
```

## Environment Variables

Create a `.env` file:

```bash
# API Base URL
VITE_API_URL=http://localhost:8080/api/v1

# Or for production
VITE_API_URL=https://api.aureo-vpn.com/api/v1
```

## Features Walkthrough

### Dashboard Overview

When you first login, you'll see:

1. **Stats Cards**:
   - Total Earned (lifetime)
   - Pending Payout (ready to withdraw)
   - Active Nodes count
   - Reputation Score with tier badge

2. **Charts**:
   - Earnings Trend (line chart)
   - Bandwidth Served (bar chart)

3. **Active Nodes**:
   - List of all your nodes
   - Status indicators (online/offline)
   - Per-node earnings
   - Uptime percentage

4. **Recent Activity**:
   - Latest earnings transactions
   - Recent payout history

### Requesting Payouts

1. Check your pending payout amount
2. Must be ≥ $10 to request
3. Click "Request Payout" button
4. Confirm the request
5. Payout processes in 24-48 hours
6. View transaction hash once completed

### Monitoring Earnings

The dashboard automatically:
- Tracks all bandwidth served
- Calculates earnings based on your tier
- Shows quality bonuses
- Displays session durations

### Tier Progression

Watch your tier upgrade as you:
- Maintain higher uptime
- Improve reputation score
- Serve more bandwidth
- Get better user ratings

**Tiers**:
- 🥉 Bronze: $0.01/GB (50%+ uptime)
- 🥈 Silver: $0.015/GB (80%+ uptime)
- 🥇 Gold: $0.02/GB (90%+ uptime)
- 💎 Platinum: $0.03/GB (95%+ uptime)

## Troubleshooting

### Cannot connect to API

**Check**:
1. API server is running: `curl http://localhost:8080/health`
2. VITE_API_URL is correct in `.env`
3. CORS is enabled on API server

**Fix**:
```bash
# Restart API server
cd ../../cmd/api-gateway
go run main.go

# Check proxy in vite.config.ts
```

### Login fails

**Check**:
1. User exists in database
2. Password meets requirements (8+ chars)
3. JWT_SECRET is set on API server

### Dashboard shows no data

**Check**:
1. You're registered as an operator
2. You have created nodes
3. Earnings have been recorded

**Fix**:
```bash
# Register as operator via API
curl -X POST http://localhost:8080/api/v1/operator/register \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "wallet_address": "0xYourAddress",
    "wallet_type": "ethereum",
    "country": "United States",
    "email": "you@example.com"
  }'
```

### Charts not displaying

**Check**:
1. Recent earnings exist
2. Browser console for errors
3. Recharts loaded correctly

## Customization

### Change Theme Colors

Edit `tailwind.config.js`:

```js
theme: {
  extend: {
    colors: {
      primary: {
        500: '#your-color',
        600: '#your-darker-color',
        // ...
      },
    },
  },
}
```

### Add New Pages

1. Create page component in `src/pages/`
2. Add route in `src/App.tsx`
3. Add navigation link

Example:

```tsx
// src/pages/Nodes.tsx
export const Nodes = () => {
  return <div>Nodes Page</div>;
};

// src/App.tsx
<Route path="/nodes" element={
  <PrivateRoute><Nodes /></PrivateRoute>
} />
```

### Modify API Endpoints

Edit `src/services/api.ts`:

```typescript
async getCustomData() {
  const response = await this.api.get('/your/endpoint');
  return response.data;
}
```

## Performance Optimization

### Code Splitting

Already configured with Vite's automatic code splitting.

### Lazy Loading

```tsx
const Dashboard = lazy(() => import('./pages/Dashboard'));

<Suspense fallback={<Loading />}>
  <Dashboard />
</Suspense>
```

### Caching

API responses are cached by Axios. Customize in `api.ts`:

```typescript
this.api.defaults.headers['Cache-Control'] = 'max-age=300';
```

## Security

### Best Practices Implemented

- ✅ JWT tokens in localStorage (with httpOnly cookies recommended for production)
- ✅ Automatic token refresh on 401
- ✅ Input validation
- ✅ HTTPS in production
- ✅ CORS configuration
- ✅ XSS protection via React

### Additional Recommendations

1. **Use httpOnly cookies** for tokens in production
2. **Enable Content Security Policy**
3. **Implement rate limiting**
4. **Use environment-specific configs**

## Testing

### Manual Testing

1. **Registration**: Create account, verify token storage
2. **Login**: Test correct/incorrect credentials
3. **Dashboard**: Verify all widgets load
4. **Payout**: Test request flow
5. **Responsive**: Test mobile/tablet views

### Automated Testing (Coming Soon)

```bash
# Unit tests
npm run test

# E2E tests
npm run test:e2e
```

## Contributing

1. Fork the repository
2. Create feature branch
3. Make changes
4. Test thoroughly
5. Submit pull request

## Support

- **Documentation**: [API_TESTING_GUIDE.md](../../API_TESTING_GUIDE.md)
- **Issues**: https://github.com/nikola43/aureo-vpn/issues
- **Email**: operator-support@aureo-vpn.com

## Roadmap

### Q1 2025
- [ ] Dark mode
- [ ] Mobile app (React Native)
- [ ] Node creation wizard
- [ ] Email notifications
- [ ] Export data (CSV/PDF)

### Q2 2025
- [ ] Real-time WebSocket updates
- [ ] Advanced analytics
- [ ] Multi-language support
- [ ] PWA support
- [ ] Operator chat/forum

## License

MIT License - See [LICENSE](../../LICENSE) for details.

---

**Version**: 1.0.0
**Last Updated**: 2025-10-27
**Status**: ✅ Production Ready

**Built with ❤️ for the Aureo VPN Community**
