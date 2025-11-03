  # Aureo VPN Deployment Guide

## Table of Contents
- [Prerequisites](#prerequisites)
- [Local Development](#local-development)
- [Docker Deployment](#docker-deployment)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Production Considerations](#production-considerations)
- [Monitoring Setup](#monitoring-setup)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### System Requirements

**API Gateway & Control Server:**
- CPU: 2+ cores
- RAM: 2GB minimum, 4GB recommended
- Storage: 10GB
- OS: Linux (Ubuntu 20.04+ recommended)

**VPN Node:**
- CPU: 4+ cores (8+ for high traffic)
- RAM: 4GB minimum, 8GB+ recommended
- Storage: 20GB
- Network: 1Gbps+ bandwidth
- OS: Linux with kernel 5.6+ (for WireGuard)

### Software Dependencies

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install -y \
    postgresql-client \
    wireguard \
    wireguard-tools \
    iptables \
    iproute2 \
    openssl \
    curl

# Go 1.22+
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

## Local Development

### 1. Clone and Setup

```bash
git clone https://github.com/nikola43/aureo-vpn.git
cd aureo-vpn

# Run setup script
chmod +x scripts/setup.sh
./scripts/setup.sh
```

### 2. Database Setup

```bash
# Option 1: Docker
docker run -d \
  --name aureo-db \
  -e POSTGRES_DB=aureo_vpn \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 \
  postgres:15-alpine

# Option 2: System PostgreSQL
sudo -u postgres createdb aureo_vpn
sudo -u postgres createuser aureo_user
sudo -u postgres psql -c "ALTER USER aureo_user WITH PASSWORD 'secure_password';"
```

### 3. Run Services

```bash
# Terminal 1 - API Gateway
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=aureo_vpn
export JWT_SECRET=your-secret-key
go run cmd/api-gateway/main.go

# Terminal 2 - Control Server
go run cmd/control-server/main.go

# Terminal 3 - Create and start a node
./bin/aureo-vpn node create \
  --name "Dev-Node" \
  --hostname "dev.local" \
  --ip "127.0.0.1" \
  --country "Development" \
  --country-code "DV" \
  --city "Local"

# Copy the node ID and start the node
export NODE_ID=<your-node-id>
sudo -E go run cmd/vpn-node/main.go
```

## Docker Deployment

### 1. Build Images

```bash
cd deployments/docker

# Build all images
docker build -f Dockerfile.api-gateway -t aureo-vpn/api-gateway:latest ../..
docker build -f Dockerfile.control-server -t aureo-vpn/control-server:latest ../..
docker build -f Dockerfile.vpn-node -t aureo-vpn/vpn-node:latest ../..
```

### 2. Configure Environment

Create a `.env` file:

```env
NODE_ID_1=<your-node-uuid>
JWT_SECRET=<generate-secure-secret>
```

### 3. Deploy with Docker Compose

```bash
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f api-gateway
```

### 4. Initialize Database

```bash
# Database migrations run automatically on API Gateway startup
# Check logs to confirm
docker-compose logs api-gateway | grep "migrations"
```

## Kubernetes Deployment

### 1. Prerequisites

```bash
# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Verify cluster access
kubectl cluster-info
```

### 2. Create Secrets

```bash
# Database credentials
kubectl create secret generic postgres-secret \
  --from-literal=username=postgres \
  --from-literal=password=<strong-password> \
  -n aureo-vpn

# JWT secret
kubectl create secret generic jwt-secret \
  --from-literal=secret=$(openssl rand -base64 32) \
  -n aureo-vpn
```

### 3. Deploy Infrastructure

```bash
cd deployments/kubernetes

# Create namespace
kubectl apply -f namespace.yaml

# Deploy PostgreSQL (for development - use managed DB in production)
kubectl apply -f postgres-deployment.yaml

# Deploy Redis
kubectl apply -f redis-deployment.yaml

# Deploy API Gateway
kubectl apply -f api-gateway-deployment.yaml

# Deploy Control Server
kubectl apply -f control-server-deployment.yaml
```

### 4. Deploy VPN Nodes

```bash
# Edit vpn-node-deployment.yaml with node-specific config
kubectl apply -f vpn-node-deployment.yaml

# Verify deployment
kubectl get pods -n aureo-vpn
kubectl get services -n aureo-vpn
```

### 5. Access Services

```bash
# Get API Gateway external IP
kubectl get service api-gateway-service -n aureo-vpn

# Port forward for local access
kubectl port-forward service/api-gateway-service 8080:80 -n aureo-vpn
```

## Production Considerations

### Security

**1. TLS/SSL Configuration**

```bash
# Use cert-manager for automatic certificate management
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Create certificate issuer
kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@yourdomain.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
EOF
```

**2. Secret Management**

Use HashiCorp Vault or AWS Secrets Manager:

```bash
# Example: AWS Secrets Manager
aws secretsmanager create-secret \
  --name aureo-vpn/jwt-secret \
  --secret-string "$(openssl rand -base64 32)"
```

**3. Network Policies**

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: api-gateway-policy
  namespace: aureo-vpn
spec:
  podSelector:
    matchLabels:
      app: api-gateway
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: aureo-vpn
    ports:
    - protocol: TCP
      port: 8080
```

### Database

**1. Use Managed Database Service**

```bash
# AWS RDS
aws rds create-db-instance \
  --db-instance-identifier aureo-vpn-db \
  --db-instance-class db.t3.medium \
  --engine postgres \
  --engine-version 15.3 \
  --master-username postgres \
  --master-user-password <secure-password> \
  --allocated-storage 100
```

**2. Enable Backups**

```bash
# Configure automated backups
aws rds modify-db-instance \
  --db-instance-identifier aureo-vpn-db \
  --backup-retention-period 7 \
  --preferred-backup-window "03:00-04:00"
```

### High Availability

**1. Multi-Region Deployment**

Deploy API Gateway and VPN Nodes across multiple regions:

```bash
# Region 1: us-east-1
# Region 2: eu-west-1
# Region 3: ap-southeast-1
```

**2. Load Balancing**

Use cloud load balancers or NGINX:

```nginx
upstream api_backend {
    least_conn;
    server api-gw-1:8080;
    server api-gw-2:8080;
    server api-gw-3:8080;
}

server {
    listen 443 ssl http2;
    server_name api.aureo-vpn.com;

    ssl_certificate /etc/ssl/certs/aureo-vpn.crt;
    ssl_certificate_key /etc/ssl/private/aureo-vpn.key;

    location / {
        proxy_pass http://api_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

**3. Auto-Scaling**

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: api-gateway-hpa
  namespace: aureo-vpn
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-gateway
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

## Monitoring Setup

### Prometheus

```bash
# Install Prometheus Operator
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/bundle.yaml

# Create ServiceMonitor
kubectl apply -f - <<EOF
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: aureo-vpn-metrics
  namespace: aureo-vpn
spec:
  selector:
    matchLabels:
      app: api-gateway
  endpoints:
  - port: http
    path: /metrics
EOF
```

### Grafana Dashboards

Import pre-built dashboards or create custom ones:

```bash
# Access Grafana
kubectl port-forward service/grafana 3000:3000 -n monitoring

# Default credentials: admin/admin
# Navigate to http://localhost:3000
```

### Alerting

Configure alerts for critical metrics:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: aureo-vpn-alerts
  namespace: aureo-vpn
spec:
  groups:
  - name: aureo-vpn
    interval: 30s
    rules:
    - alert: HighErrorRate
      expr: rate(aureo_vpn_http_requests_total{status=~"5.."}[5m]) > 0.05
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: "High error rate detected"
```

## Troubleshooting

### Common Issues

**1. VPN Node Won't Start**

```bash
# Check kernel modules
sudo modprobe wireguard
lsmod | grep wireguard

# Check capabilities
sudo setcap cap_net_admin+ep /path/to/vpn-node

# Check logs
journalctl -u aureo-vpn-node -f
```

**2. Database Connection Issues**

```bash
# Test connection
psql -h $DB_HOST -U $DB_USER -d $DB_NAME

# Check firewall
sudo ufw allow 5432/tcp

# Verify credentials
env | grep DB_
```

**3. High Latency**

```bash
# Check node load
./bin/aureo-vpn stats

# Test network
ping -c 10 <vpn-node-ip>
mtr <vpn-node-ip>

# Check resource usage
top
htop
```

### Logs

```bash
# Docker
docker-compose logs -f [service-name]

# Kubernetes
kubectl logs -f deployment/api-gateway -n aureo-vpn
kubectl logs -f deployment/vpn-node-1 -n aureo-vpn

# System
journalctl -u aureo-vpn-api -f
```

### Performance Tuning

**1. PostgreSQL**

```sql
-- Increase connections
ALTER SYSTEM SET max_connections = 200;

-- Optimize for SSD
ALTER SYSTEM SET random_page_cost = 1.1;

-- Increase shared buffers
ALTER SYSTEM SET shared_buffers = '2GB';
```

**2. Kernel Parameters**

```bash
# Add to /etc/sysctl.conf
net.core.rmem_max = 134217728
net.core.wmem_max = 134217728
net.ipv4.tcp_rmem = 4096 87380 67108864
net.ipv4.tcp_wmem = 4096 65536 67108864
net.ipv4.ip_forward = 1
net.ipv6.conf.all.forwarding = 1

# Apply
sudo sysctl -p
```

## Support

For additional help:
- Documentation: `/docs`
- Issues: GitHub Issues
- Community: Discord/Slack channel
