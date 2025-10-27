#!/bin/bash

################################################################################
# Aureo VPN - Complete All-in-One Node Operator Setup Script
#
# This script offers two installation methods:
#   1. System Installation - Install directly on the system
#   2. Docker Installation - Run everything in Docker containers (Recommended)
#
# What it does:
#   - Install all dependencies (PostgreSQL, Go, WireGuard, etc.)
#   - Setup and configure PostgreSQL database
#   - Build and configure the API gateway
#   - Setup and start VPN node (WireGuard + OpenVPN)
#   - Create services for all components
#   - Register you as an operator and create your first node
#
# Usage:
#   sudo ./become-node-operator.sh
#
################################################################################

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
INSTALL_DIR="/opt/aureo-vpn"
CONFIG_DIR="$HOME/.aureo-vpn"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INSTALL_METHOD=""  # Will be set to "system" or "docker"

# Trap errors
trap 'error_handler $? $LINENO' ERR

error_handler() {
    echo -e "\n${RED}✗ Error occurred at line $2 (exit code: $1)${NC}"
    echo -e "${YELLOW}Setup failed. Please check the errors above.${NC}"
    exit 1
}

# Print header
print_header() {
    clear
    echo -e "${PURPLE}"
    cat << "EOF"
╔══════════════════════════════════════════════════════════════════╗
║                                                                  ║
║       🚀 Aureo VPN - Complete Node Operator Setup 🚀            ║
║                                                                  ║
║     All-in-One Setup: Database + API + Node + Gateway          ║
║        Earn Crypto Rewards by Running a VPN Node!              ║
║                                                                  ║
╚══════════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}\n"
}

# Print section header
section() {
    echo -e "\n${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  $1${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
}

# Check if running as root
check_root() {
    # For macOS with Docker installation, root is not required
    if [[ "$OSTYPE" == "darwin"* ]] && [ "$INSTALL_METHOD" = "docker" ]; then
        echo -e "${BLUE}ℹ Running on macOS with Docker - root not required${NC}"
        return 0
    fi

    if [ "$EUID" -ne 0 ]; then
        echo -e "${RED}Please run as root (use sudo)${NC}"
        exit 1
    fi
}

# Command exists helper
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Choose installation method
choose_installation_method() {
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  Choose Installation Method${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"

    echo -e "${BLUE}1)${NC} ${GREEN}Docker Installation${NC} ${YELLOW}(Recommended)${NC}"
    echo -e "   ${BLUE}✓${NC} Isolated containers for each service"
    echo -e "   ${BLUE}✓${NC} Easy to manage and update"
    echo -e "   ${BLUE}✓${NC} Clean uninstall"
    echo -e "   ${BLUE}✓${NC} No system pollution"
    echo ""
    echo -e "${BLUE}2)${NC} ${GREEN}System Installation${NC}"
    echo -e "   ${BLUE}✓${NC} Direct system installation"
    echo -e "   ${BLUE}✓${NC} Better performance (no containerization overhead)"
    echo -e "   ${BLUE}✓${NC} Systemd service management"
    echo -e "   ${BLUE}⚠${NC}  Installs packages system-wide"
    echo ""

    read -p "$(echo -e ${CYAN}Enter your choice [1 or 2]:${NC} )" INSTALL_CHOICE

    case $INSTALL_CHOICE in
        1)
            INSTALL_METHOD="docker"
            echo -e "\n${GREEN}✓ Docker installation selected${NC}"
            ;;
        2)
            INSTALL_METHOD="system"
            echo -e "\n${GREEN}✓ System installation selected${NC}"
            ;;
        *)
            echo -e "${RED}Invalid choice. Defaulting to Docker installation.${NC}"
            INSTALL_METHOD="docker"
            ;;
    esac
}

##############################################################################
# DOCKER INSTALLATION FUNCTIONS
##############################################################################

# Install Docker
install_docker() {
    section "🐳 Installing Docker"

    if command_exists docker; then
        echo -e "${GREEN}✓ Docker already installed${NC}"
        docker --version

        # Check if Docker is running
        if docker ps > /dev/null 2>&1; then
            echo -e "${GREEN}✓ Docker daemon is running${NC}"
        else
            if [ "$OS" = "macos" ]; then
                echo -e "${YELLOW}⚠ Docker daemon is not running${NC}"
                echo -e "${BLUE}Please start Docker Desktop and press Enter to continue...${NC}"
                read
            else
                echo -e "${YELLOW}Starting Docker daemon...${NC}"
                systemctl start docker
                systemctl enable docker
            fi
        fi
    else
        if [ "$OS" = "macos" ]; then
            echo -e "${YELLOW}Docker is not installed${NC}"
            echo ""
            echo -e "${BLUE}Please install Docker Desktop for Mac:${NC}"
            echo -e "${CYAN}1. Visit: https://www.docker.com/products/docker-desktop${NC}"
            echo -e "${CYAN}2. Download and install Docker Desktop${NC}"
            echo -e "${CYAN}3. Start Docker Desktop${NC}"
            echo -e "${CYAN}4. Press Enter to continue...${NC}"
            read

            # Verify Docker is now installed
            if ! command_exists docker; then
                echo -e "${RED}✗ Docker still not found. Please install Docker Desktop first.${NC}"
                exit 1
            fi
            echo -e "${GREEN}✓ Docker detected${NC}"
        else
            echo "Installing Docker..."
            curl -fsSL https://get.docker.com | sh
            usermod -aG docker $SUDO_USER || true
            echo -e "${GREEN}✓ Docker installed${NC}"
        fi
    fi

    # Install Docker Compose (if needed)
    if command_exists docker-compose; then
        echo -e "${GREEN}✓ Docker Compose already installed${NC}"
    else
        # Check if using Docker Desktop (which includes compose)
        if docker compose version > /dev/null 2>&1; then
            echo -e "${GREEN}✓ Docker Compose (v2) is available via Docker Desktop${NC}"
        else
            echo "Installing Docker Compose..."
            curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
            chmod +x /usr/local/bin/docker-compose
            echo -e "${GREEN}✓ Docker Compose installed${NC}"
        fi
    fi

    # Start Docker (Linux only)
    if [ "$OS" != "macos" ]; then
        systemctl start docker
        systemctl enable docker
    fi
}

# Create docker-compose.yml
create_docker_compose() {
    section "📝 Creating Docker Compose Configuration"

    mkdir -p "$INSTALL_DIR"
    cd "$INSTALL_DIR"

    # Generate JWT secret and DB password (hex to avoid special chars in .env)
    JWT_SECRET=$(openssl rand -hex 48)
    DB_PASSWORD=$(openssl rand -hex 24)

    # Save secrets
    mkdir -p "$CONFIG_DIR"
    cat > "$CONFIG_DIR/docker-secrets" << EOF
JWT_SECRET=$JWT_SECRET
DB_PASSWORD=$DB_PASSWORD
EOF
    chmod 600 "$CONFIG_DIR/docker-secrets"

    # Create .env file for Docker Compose
    # Note: Don't quote values in .env - Docker Compose will use them literally
    cat > "$INSTALL_DIR/.env" << EOF
# Database
DB_PASSWORD=$DB_PASSWORD

# JWT
JWT_SECRET=$JWT_SECRET

# Ports
API_PORT=8080
POSTGRES_PORT=5432
WIREGUARD_PORT=51820
OPENVPN_PORT=1194
DASHBOARD_PORT=3000
EOF

    # Create docker-compose.yml
    cat > "$INSTALL_DIR/docker-compose.yml" << 'EOF'
version: '3.8'

services:
  # PostgreSQL Database
  postgres:
    image: postgres:15-alpine
    container_name: aureo-postgres
    restart: unless-stopped
    environment:
      POSTGRES_DB: aureo_vpn
      POSTGRES_USER: aureo_user
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      PGDATA: /var/lib/postgresql/data/pgdata
    volumes:
      - postgres-data:/var/lib/postgresql/data
    ports:
      - "${POSTGRES_PORT:-5432}:5432"
    networks:
      - aureo-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U aureo_user -d aureo_vpn"]
      interval: 10s
      timeout: 5s
      retries: 5

  # API Gateway
  api:
    build:
      context: .
      dockerfile: Dockerfile.api
    container_name: aureo-api
    restart: unless-stopped
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      # Database
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: aureo_user
      DB_PASSWORD: ${DB_PASSWORD}
      DB_NAME: aureo_vpn
      DB_SSLMODE: disable
      DB_TIMEZONE: UTC
      DB_MAX_IDLE_CONNS: 10
      DB_MAX_OPEN_CONNS: 100
      DB_LOG_LEVEL: warn

      # Server
      SERVER_HOST: 0.0.0.0
      SERVER_PORT: 8080
      SERVER_READ_TIMEOUT: 15s
      SERVER_WRITE_TIMEOUT: 15s
      SERVER_SHUTDOWN_TIMEOUT: 30s

      # JWT
      JWT_SECRET: ${JWT_SECRET}
      JWT_ACCESS_EXPIRY: 24h
      JWT_REFRESH_EXPIRY: 168h

      # Security
      ARGON2_MEMORY: 65536
      ARGON2_ITERATIONS: 3
      ARGON2_PARALLELISM: 4
      ARGON2_SALT_LENGTH: 16
      ARGON2_KEY_LENGTH: 32

      # Environment
      ENVIRONMENT: production

      # Logging
      LOG_LEVEL: info
      LOG_FORMAT: json

      # CORS
      CORS_ALLOWED_ORIGINS: http://localhost:3000,http://localhost:8080,http://dashboard:3000
      CORS_ALLOWED_METHODS: GET,POST,PUT,DELETE,OPTIONS
      CORS_ALLOWED_HEADERS: Content-Type,Authorization
      CORS_ALLOW_CREDENTIALS: "false"

      # Rate Limiting
      RATE_LIMIT_REQUESTS_PER_SECOND: 10
      RATE_LIMIT_BURST: 20

      # VPN
      VPN_NETWORK: 10.8.0.0/24
      WIREGUARD_PORT: 51820
      OPENVPN_PORT: 1194
    ports:
      - "${API_PORT:-8080}:8080"
    networks:
      - aureo-network
    volumes:
      - api-logs:/app/logs
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  # WireGuard VPN
  wireguard:
    image: linuxserver/wireguard:latest
    container_name: aureo-wireguard
    restart: unless-stopped
    cap_add:
      - NET_ADMIN
      - SYS_MODULE
    environment:
      - PUID=1000
      - PGID=1000
      - TZ=UTC
      - SERVERPORT=${WIREGUARD_PORT:-51820}
      - PEERS=10
      - PEERDNS=auto
      - INTERNAL_SUBNET=10.8.0.0/24
    volumes:
      - wireguard-config:/config
    ports:
      - "${WIREGUARD_PORT:-51820}:51820/udp"
    networks:
      - aureo-network
    sysctls:
      - net.ipv4.conf.all.src_valid_mark=1
      - net.ipv4.ip_forward=1

  # Dashboard
  dashboard:
    build:
      context: ./web/operator-dashboard
      dockerfile: Dockerfile
    container_name: aureo-dashboard
    restart: unless-stopped
    depends_on:
      - api
    environment:
      VITE_API_URL: http://api:8080/api/v1
    ports:
      - "${DASHBOARD_PORT:-3000}:3000"
    networks:
      - aureo-network

volumes:
  postgres-data:
    driver: local
  wireguard-config:
    driver: local
  api-logs:
    driver: local

networks:
  aureo-network:
    driver: bridge
EOF

    echo -e "${GREEN}✓ Docker Compose configuration created${NC}"
}

# Create Dockerfile for API
create_api_dockerfile() {
    cat > "$INSTALL_DIR/Dockerfile.api" << 'EOF'
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
WORKDIR /build/cmd/api-gateway
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o aureo-api-gateway .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/cmd/api-gateway/aureo-api-gateway .

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run
CMD ["./aureo-api-gateway"]
EOF
}

# Create Dockerfile for Dashboard
create_dashboard_dockerfile() {
    mkdir -p "$INSTALL_DIR/web/operator-dashboard"

    # Create .dockerignore to prevent copying node_modules
    cat > "$INSTALL_DIR/web/operator-dashboard/.dockerignore" << 'EOF'
node_modules
dist
.env
.env.local
*.log
.DS_Store
EOF

    cat > "$INSTALL_DIR/web/operator-dashboard/Dockerfile" << 'EOF'
FROM node:20-alpine AS builder

WORKDIR /build

# Copy package.json
COPY package.json ./

# Clean install dependencies
RUN rm -rf node_modules && \
    npm cache clean --force && \
    npm install

# Copy source code (node_modules excluded via .dockerignore)
COPY . .

# Build for production (skip tsc, just use vite)
RUN npx vite build

# Production stage with nginx
FROM nginx:alpine

# Copy built files
COPY --from=builder /build/dist /usr/share/nginx/html

# Create nginx config
RUN echo 'server {' > /etc/nginx/conf.d/default.conf && \
    echo '    listen 3000;' >> /etc/nginx/conf.d/default.conf && \
    echo '    root /usr/share/nginx/html;' >> /etc/nginx/conf.d/default.conf && \
    echo '    index index.html;' >> /etc/nginx/conf.d/default.conf && \
    echo '    location / {' >> /etc/nginx/conf.d/default.conf && \
    echo '        try_files $uri $uri/ /index.html;' >> /etc/nginx/conf.d/default.conf && \
    echo '    }' >> /etc/nginx/conf.d/default.conf && \
    echo '}' >> /etc/nginx/conf.d/default.conf

EXPOSE 3000

CMD ["nginx", "-g", "daemon off;"]
EOF
}

# Copy project files for Docker
copy_project_for_docker() {
    echo "Copying project files..."

    # Copy API files
    cp -r "$PROJECT_ROOT/cmd" "$INSTALL_DIR/" 2>/dev/null || true
    cp -r "$PROJECT_ROOT/internal" "$INSTALL_DIR/" 2>/dev/null || true
    cp -r "$PROJECT_ROOT/pkg" "$INSTALL_DIR/" 2>/dev/null || true
    cp "$PROJECT_ROOT/go.mod" "$INSTALL_DIR/" 2>/dev/null || true
    cp "$PROJECT_ROOT/go.sum" "$INSTALL_DIR/" 2>/dev/null || true

    # Copy dashboard files
    mkdir -p "$INSTALL_DIR/web"
    cp -r "$PROJECT_ROOT/web/operator-dashboard" "$INSTALL_DIR/web/" 2>/dev/null || true

    echo -e "${GREEN}✓ Project files copied${NC}"
}

# Start Docker services
start_docker_services() {
    section "🚀 Starting Docker Services"

    cd "$INSTALL_DIR"

    echo "Building and starting containers..."
    docker-compose up -d --build

    echo "Waiting for services to be ready..."
    sleep 10

    # Check service health
    if docker-compose ps | grep -q "Up"; then
        echo -e "${GREEN}✓ Docker services started successfully${NC}"
        docker-compose ps
    else
        echo -e "${RED}✗ Some services failed to start${NC}"
        docker-compose ps
        exit 1
    fi
}

##############################################################################
# SYSTEM INSTALLATION FUNCTIONS
##############################################################################

# Detect OS
detect_os() {
    section "🔍 Detecting Operating System"

    # Detect macOS
    if [[ "$OSTYPE" == "darwin"* ]]; then
        OS="macos"
        OS_VERSION=$(sw_vers -productVersion)
        echo -e "${GREEN}✓ Detected: macOS $OS_VERSION${NC}"

        # For Docker installation, macOS is fully supported
        if [ "$INSTALL_METHOD" = "docker" ]; then
            echo -e "${BLUE}ℹ Docker installation on macOS is fully supported${NC}"
        else
            echo -e "${YELLOW}⚠ System installation on macOS has limited support${NC}"
            echo -e "${YELLOW}  Recommendation: Use Docker installation instead${NC}"
            read -p "Continue anyway? (y/n): " CONTINUE_MACOS
            if [[ ! $CONTINUE_MACOS =~ ^[Yy]$ ]]; then
                echo "Installation cancelled. Please re-run and choose Docker installation."
                exit 0
            fi
        fi
    elif [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        OS_VERSION=$VERSION_ID
        echo -e "${GREEN}✓ Detected: $PRETTY_NAME${NC}"
    else
        echo -e "${RED}Cannot detect OS${NC}"
        exit 1
    fi
}

# Install system dependencies
install_system_dependencies() {
    section "📦 Installing System Dependencies"

    case "$OS" in
        ubuntu|debian)
            echo "Updating package lists..."
            apt-get update -qq

            echo "Installing dependencies..."
            DEBIAN_FRONTEND=noninteractive apt-get install -y -qq \
                postgresql postgresql-contrib \
                golang-go \
                wireguard wireguard-tools \
                openvpn easy-rsa \
                iptables curl wget git jq \
                net-tools resolvconf \
                build-essential
            ;;
        centos|rhel|fedora)
            echo "Installing EPEL repository..."
            yum install -y epel-release

            echo "Installing dependencies..."
            yum install -y \
                postgresql postgresql-server postgresql-contrib \
                golang \
                wireguard-tools \
                openvpn easy-rsa \
                iptables curl wget git jq \
                net-tools \
                gcc make
            ;;
        *)
            echo -e "${RED}Unsupported OS: $OS${NC}"
            exit 1
            ;;
    esac

    echo -e "${GREEN}✓ System dependencies installed${NC}"
}

# Setup PostgreSQL database
setup_database() {
    section "🗄️  Setting Up PostgreSQL Database"

    # Start PostgreSQL
    systemctl start postgresql
    systemctl enable postgresql

    echo -e "${BLUE}Creating database and user...${NC}"

    # Generate random password
    DB_PASSWORD=$(openssl rand -base64 32)

    # Create database and user
    sudo -u postgres psql << EOF
CREATE DATABASE aureo_vpn;
CREATE USER aureo_user WITH ENCRYPTED PASSWORD '$DB_PASSWORD';
GRANT ALL PRIVILEGES ON DATABASE aureo_vpn TO aureo_user;
ALTER USER aureo_user CREATEDB;
\q
EOF

    echo -e "${GREEN}✓ Database created${NC}"

    # Save database credentials
    mkdir -p "$CONFIG_DIR"
    cat > "$CONFIG_DIR/db-credentials" << EOF
DB_HOST=localhost
DB_PORT=5432
DB_USER=aureo_user
DB_PASSWORD=$DB_PASSWORD
DB_NAME=aureo_vpn
EOF
    chmod 600 "$CONFIG_DIR/db-credentials"
}

# Build API Gateway
build_api_gateway() {
    section "🏗️  Building API Gateway"

    echo "Building Go application..."
    cd "$PROJECT_ROOT/cmd/api-gateway"

    # Build
    go build -o aureo-api-gateway main.go

    echo -e "${GREEN}✓ API Gateway built${NC}"

    # Create installation directory
    mkdir -p "$INSTALL_DIR/bin"
    cp aureo-api-gateway "$INSTALL_DIR/bin/"
    chmod +x "$INSTALL_DIR/bin/aureo-api-gateway"
}

# Configure environment for system install
configure_system_environment() {
    section "⚙️  Configuring Environment"

    # Generate JWT secret (hex to avoid special chars)
    JWT_SECRET=$(openssl rand -hex 48)

    # Source database credentials
    source "$CONFIG_DIR/db-credentials"

    # Create .env file (no quotes - using hex secrets without special chars)
    cat > "$INSTALL_DIR/.env" << EOF
# Database Configuration
DB_HOST=$DB_HOST
DB_PORT=$DB_PORT
DB_USER=$DB_USER
DB_PASSWORD=$DB_PASSWORD
DB_NAME=$DB_NAME
DB_SSLMODE=disable
DB_TIMEZONE=UTC
DB_MAX_IDLE_CONNS=10
DB_MAX_OPEN_CONNS=100
DB_LOG_LEVEL=warn

# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_READ_TIMEOUT=15s
SERVER_WRITE_TIMEOUT=15s
SERVER_SHUTDOWN_TIMEOUT=30s

# Environment
ENVIRONMENT=production

# JWT Configuration
JWT_SECRET=$JWT_SECRET
JWT_ACCESS_EXPIRY=24h
JWT_REFRESH_EXPIRY=168h

# Security
ARGON2_MEMORY=65536
ARGON2_ITERATIONS=3
ARGON2_PARALLELISM=4
ARGON2_SALT_LENGTH=16
ARGON2_KEY_LENGTH=32

# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization
CORS_ALLOW_CREDENTIALS=true

# Rate Limiting
RATE_LIMIT_REQUESTS_PER_SECOND=10
RATE_LIMIT_BURST=20

# Node Configuration
VPN_NETWORK=10.8.0.0/24
WIREGUARD_PORT=51820
OPENVPN_PORT=1194
EOF

    chmod 600 "$INSTALL_DIR/.env"
    echo -e "${GREEN}✓ Environment configured${NC}"
}

# Create systemd service for API
create_api_service() {
    section "🔧 Creating API Gateway Service"

    cat > /etc/systemd/system/aureo-api.service << EOF
[Unit]
Description=Aureo VPN API Gateway
After=network.target postgresql.service
Requires=postgresql.service

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
EnvironmentFile=$INSTALL_DIR/.env
ExecStart=$INSTALL_DIR/bin/aureo-api-gateway
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable aureo-api.service

    echo -e "${GREEN}✓ API service created${NC}"
}

# Start API server
start_api_server() {
    section "🚀 Starting API Server"

    echo "Starting API Gateway..."
    systemctl start aureo-api.service

    # Wait for API to be ready
    echo "Waiting for API to be ready..."
    sleep 3

    # Check if running
    if systemctl is-active --quiet aureo-api.service; then
        echo -e "${GREEN}✓ API Gateway is running${NC}"

        # Test API
        if curl -s http://localhost:8080/health > /dev/null; then
            echo -e "${GREEN}✓ API health check passed${NC}"
        else
            echo -e "${YELLOW}⚠ API may not be fully ready yet${NC}"
        fi
    else
        echo -e "${RED}✗ Failed to start API Gateway${NC}"
        echo -e "${YELLOW}Check logs: sudo journalctl -u aureo-api.service -f${NC}"
        exit 1
    fi
}

# Setup WireGuard
setup_wireguard() {
    section "🔐 Setting Up WireGuard"

    # Generate WireGuard keys
    WG_PRIVATE_KEY=$(wg genkey)
    WG_PUBLIC_KEY=$(echo "$WG_PRIVATE_KEY" | wg pubkey)

    # Create WireGuard config
    mkdir -p /etc/wireguard
    cat > /etc/wireguard/wg0.conf << EOF
[Interface]
PrivateKey = $WG_PRIVATE_KEY
Address = 10.8.0.1/24
ListenPort = 51820
PostUp = iptables -A FORWARD -i wg0 -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE; iptables -t nat -A POSTROUTING -o ens3 -j MASQUERADE
PostDown = iptables -D FORWARD -i wg0 -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE; iptables -t nat -D POSTROUTING -o ens3 -j MASQUERADE
EOF

    chmod 600 /etc/wireguard/wg0.conf

    # Save keys
    echo "$WG_PUBLIC_KEY" > "$CONFIG_DIR/wireguard-public-key"

    echo -e "${GREEN}✓ WireGuard configured${NC}"
}

# Enable IP forwarding
enable_ip_forwarding() {
    echo "Enabling IP forwarding..."
    echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf
    echo "net.ipv6.conf.all.forwarding=1" >> /etc/sysctl.conf
    sysctl -p > /dev/null 2>&1
    echo -e "${GREEN}✓ IP forwarding enabled${NC}"
}

# Configure firewall
configure_firewall() {
    section "🔥 Configuring Firewall"

    echo "Configuring iptables..."

    # Allow VPN ports
    iptables -A INPUT -p udp --dport 51820 -j ACCEPT
    iptables -A INPUT -p udp --dport 1194 -j ACCEPT
    iptables -A INPUT -p tcp --dport 1194 -j ACCEPT
    iptables -A INPUT -p tcp --dport 8080 -j ACCEPT

    # Save iptables rules
    if command_exists iptables-save; then
        iptables-save > /etc/iptables/rules.v4 2>/dev/null || true
    fi

    echo -e "${GREEN}✓ Firewall configured${NC}"
}

# Start WireGuard
start_wireguard() {
    section "🚀 Starting WireGuard VPN"

    systemctl enable wg-quick@wg0
    systemctl start wg-quick@wg0

    if systemctl is-active --quiet wg-quick@wg0; then
        echo -e "${GREEN}✓ WireGuard is running${NC}"
    else
        echo -e "${RED}✗ Failed to start WireGuard${NC}"
    fi
}

##############################################################################
# SHARED FUNCTIONS (used by both installation methods)
##############################################################################

# Create operator account
create_operator_account() {
    section "👤 Creating Your Operator Account"

    echo "Let's create your operator account!"
    echo ""

    read -p "Email: " EMAIL
    read -p "Username: " USERNAME
    read -sp "Password (min 8 characters): " PASSWORD
    echo ""

    # Determine API URL
    if [ "$INSTALL_METHOD" = "docker" ]; then
        API_URL="http://localhost:8080"
    else
        API_URL="http://localhost:8080"
    fi

    # Wait for API to be fully ready
    echo -e "\n${BLUE}Waiting for API to be ready...${NC}"
    for i in {1..30}; do
        if curl -sf "$API_URL/health" > /dev/null 2>&1; then
            echo -e "${GREEN}✓ API is ready${NC}"
            break
        fi
        echo -n "."
        sleep 1
    done
    echo ""

    # Register user
    echo -e "${BLUE}Registering user account...${NC}"

    REGISTER_RESPONSE=$(curl -sf -X POST "$API_URL/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$EMAIL\",
            \"password\": \"$PASSWORD\",
            \"username\": \"$USERNAME\"
        }" || echo '{"error": "Connection failed"}')

    echo "Debug - Registration response: $REGISTER_RESPONSE"

    ACCESS_TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.access_token // empty' 2>/dev/null)

    if [ -n "$ACCESS_TOKEN" ] && [ "$ACCESS_TOKEN" != "null" ]; then
        echo -e "${GREEN}✓ User account created${NC}"
    else
        echo -e "${RED}✗ Failed to create user account${NC}"
        ERROR_MSG=$(echo "$REGISTER_RESPONSE" | jq -r '.error // .message // "Unknown error"' 2>/dev/null)
        echo "Error: $ERROR_MSG"
        echo "Full response: $REGISTER_RESPONSE"
        exit 1
    fi

    # Setup crypto wallet
    echo ""
    echo "Choose cryptocurrency for rewards:"
    echo "  1) Ethereum (ETH)"
    echo "  2) Bitcoin (BTC)"
    echo "  3) Litecoin (LTC)"
    read -p "Choice [1-3]: " CRYPTO_CHOICE

    case $CRYPTO_CHOICE in
        1) CRYPTO_TYPE="ethereum" ;;
        2) CRYPTO_TYPE="bitcoin" ;;
        3) CRYPTO_TYPE="litecoin" ;;
        *) CRYPTO_TYPE="ethereum" ;;
    esac

    read -p "Enter your $CRYPTO_TYPE wallet address: " WALLET_ADDRESS

    # Set country to Unknown (skip API call to avoid hanging)
    COUNTRY="Unknown"

    # Register as operator
    echo -e "\n${BLUE}Registering as node operator...${NC}"

    OPERATOR_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/operator/register" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -d "{
            \"wallet_address\": \"$WALLET_ADDRESS\",
            \"wallet_type\": \"$CRYPTO_TYPE\",
            \"country\": \"$COUNTRY\",
            \"email\": \"$EMAIL\"
        }")

    if echo "$OPERATOR_RESPONSE" | jq -e '.operator' > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Operator account created${NC}"
    else
        echo -e "${RED}✗ Failed to register as operator${NC}"
        echo "Response: $OPERATOR_RESPONSE"
        exit 1
    fi

    # Get fresh token with operator permissions
    echo -e "${CYAN}Getting fresh authorization token...${NC}"
    LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$EMAIL\",
            \"password\": \"$PASSWORD\"
        }")

    ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token // empty' 2>/dev/null)
    if [ -z "$ACCESS_TOKEN" ] || [ "$ACCESS_TOKEN" = "null" ]; then
        echo -e "${RED}✗ Failed to get operator token${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Token refreshed${NC}"

    # Activate operator account (new operators are pending by default)
    echo -e "${CYAN}Activating operator account...${NC}"
    docker exec aureo-postgres psql -U aureo_user -d aureo_vpn -c \
        "UPDATE node_operators SET status='active', is_verified=true, verified_at=NOW() WHERE wallet_address='$WALLET_ADDRESS';" \
        > /dev/null 2>&1

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Operator account activated${NC}"
    else
        echo -e "${YELLOW}⚠ Could not auto-activate operator (may require manual admin approval)${NC}"
    fi

    # Save credentials
    cat > "$CONFIG_DIR/operator-credentials" << EOF
EMAIL=$EMAIL
USERNAME=$USERNAME
ACCESS_TOKEN=$ACCESS_TOKEN
WALLET_ADDRESS=$WALLET_ADDRESS
CRYPTO_TYPE=$CRYPTO_TYPE
API_URL=$API_URL
EOF
    chmod 600 "$CONFIG_DIR/operator-credentials"
}

# Register VPN node
register_node() {
    section "🖥️  Registering Your VPN Node"

    # Auto-detect information
    PUBLIC_IP=$(curl -s https://api.ipify.org || echo "127.0.0.1")
    HOSTNAME=$(hostname)

    # Set default location values (skip API calls to avoid hanging)
    COUNTRY="Unknown"
    COUNTRY_CODE="US"
    CITY="Unknown"
    LATITUDE="0"
    LONGITUDE="0"

    echo -e "${BLUE}Auto-detected information:${NC}"
    echo "  Public IP: $PUBLIC_IP"
    echo "  Location: $CITY, $COUNTRY"
    echo ""

    read -p "Node name [default: aureo-node-$HOSTNAME]: " NODE_NAME
    NODE_NAME=${NODE_NAME:-"aureo-node-$HOSTNAME"}

    # Source operator credentials
    source "$CONFIG_DIR/operator-credentials"

    # Register node
    echo -e "\n${BLUE}Registering VPN node with network...${NC}"

    NODE_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/operator/nodes" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -d "{
            \"name\": \"$NODE_NAME\",
            \"hostname\": \"$HOSTNAME\",
            \"public_ip\": \"$PUBLIC_IP\",
            \"country\": \"$COUNTRY\",
            \"country_code\": \"$COUNTRY_CODE\",
            \"city\": \"$CITY\",
            \"wireguard_port\": 51820,
            \"openvpn_port\": 1194,
            \"latitude\": $LATITUDE,
            \"longitude\": $LONGITUDE
        }")

    NODE_ID=$(echo "$NODE_RESPONSE" | jq -r '.node.id')

    if [ "$NODE_ID" != "null" ] && [ -n "$NODE_ID" ]; then
        echo -e "${GREEN}✓ VPN node registered successfully${NC}"
        echo -e "${BLUE}  Node ID: $NODE_ID${NC}"

        # Set node status to online (for development/testing - in production, nodes send heartbeats)
        echo -e "${CYAN}Activating node...${NC}"
        docker exec aureo-postgres psql -U aureo_user -d aureo_vpn -c \
            "UPDATE vpn_nodes SET status='online', last_heartbeat=NOW() WHERE id='$NODE_ID';" \
            > /dev/null 2>&1

        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✓ Node activated and set to online${NC}"

            # Recalculate operator stats to reflect the new active node
            docker exec aureo-postgres psql -U aureo_user -d aureo_vpn -c \
                "UPDATE node_operators SET active_nodes_count = (SELECT COUNT(*) FROM vpn_nodes WHERE operator_id = node_operators.id AND status = 'online' AND is_active = true);" \
                > /dev/null 2>&1

            echo -e "${GREEN}✓ Operator stats updated${NC}"

            # Get WireGuard server public key and update node
            echo -e "${CYAN}Configuring WireGuard...${NC}"
            WG_PRIVATE_KEY=$(docker exec aureo-wireguard grep "PrivateKey" /config/wg_confs/wg0.conf 2>/dev/null | cut -d'=' -f2 | tr -d ' ')
            if [ -n "$WG_PRIVATE_KEY" ]; then
                WG_PUBLIC_KEY=$(echo "$WG_PRIVATE_KEY" | wg pubkey 2>/dev/null)
                if [ -n "$WG_PUBLIC_KEY" ]; then
                    docker exec aureo-postgres psql -U aureo_user -d aureo_vpn -c \
                        "UPDATE vpn_nodes SET public_key='$WG_PUBLIC_KEY' WHERE id='$NODE_ID';" \
                        > /dev/null 2>&1
                    echo -e "${GREEN}✓ WireGuard public key configured${NC}"
                fi
            fi
        else
            echo -e "${YELLOW}⚠ Could not auto-activate node (will show as offline until heartbeat)${NC}"
        fi

        # Save node info
        cat >> "$CONFIG_DIR/operator-credentials" << EOF
NODE_ID=$NODE_ID
NODE_NAME=$NODE_NAME
PUBLIC_IP=$PUBLIC_IP
EOF
    else
        echo -e "${RED}✗ Failed to register node${NC}"
        echo "Response: $NODE_RESPONSE"
        exit 1
    fi
}

# Setup web dashboard
setup_dashboard() {
    section "🌐 Setting Up Web Dashboard"

    echo "Installing Node.js and npm..."
    if ! command_exists node; then
        curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
        apt-get install -y nodejs
    fi

    # Build dashboard
    cd "$PROJECT_ROOT/web/operator-dashboard"

    if [ ! -d "node_modules" ]; then
        echo "Installing dashboard dependencies..."
        npm install
    fi

    # Create .env for dashboard
    cat > .env << EOF
VITE_API_URL=http://localhost:8080/api/v1
EOF

    echo -e "${GREEN}✓ Dashboard configured${NC}"
    echo -e "${YELLOW}  Run 'npm run dev' in web/operator-dashboard to start dashboard${NC}"
}

# Print final summary
print_summary() {
    section "🎉 Setup Complete!"

    source "$CONFIG_DIR/operator-credentials"

    # Convert install method to uppercase (bash 3.2 compatible)
    INSTALL_METHOD_UPPER=$(echo "$INSTALL_METHOD" | tr '[:lower:]' '[:upper:]')

    echo -e "${GREEN}✓ Your Aureo VPN node is fully configured and running!${NC}"
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  INSTALLATION METHOD${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "  Installation Type:  ${GREEN}${INSTALL_METHOD_UPPER}${NC}"
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  SERVICES STATUS${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "  ✓ PostgreSQL Database:  Running"
    echo -e "  ✓ API Gateway:          Running on http://localhost:8080"
    echo -e "  ✓ WireGuard VPN:        Running on port 51820"
    echo -e "  ✓ Web Dashboard:        Running on http://localhost:3000"
    echo -e "  ✓ Node Status:          Registered and Active"
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  YOUR ACCOUNT${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "  Email:          $EMAIL"
    echo -e "  Username:       $USERNAME"
    echo -e "  Wallet:         $WALLET_ADDRESS ($CRYPTO_TYPE)"
    echo -e "  Node Name:      $NODE_NAME"
    echo -e "  Node ID:        $NODE_ID"
    echo -e "  Public IP:      $PUBLIC_IP"
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  ACCESS YOUR DASHBOARD${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

    echo ""
    if [ "$INSTALL_METHOD" = "docker" ]; then
        echo -e "  ${GREEN}🌐 Dashboard is already running!${NC}"
        echo ""
        echo -e "  Open your browser and visit:"
        echo -e "     ${GREEN}http://localhost:3000${NC}"
        echo ""
        echo -e "  Login with:"
        echo -e "     ${GREEN}Email: $EMAIL${NC}"
        echo -e "     ${GREEN}Password: <your password>${NC}"
    else
        echo -e "  1. Start the dashboard:"
        echo -e "     ${GREEN}cd $PROJECT_ROOT/web/operator-dashboard${NC}"
        echo -e "     ${GREEN}npm run dev${NC}"
        echo ""
        echo -e "  2. Open your browser:"
        echo -e "     ${GREEN}http://localhost:3000${NC}"
        echo ""
        echo -e "  3. Login with:"
        echo -e "     ${GREEN}Email: $EMAIL${NC}"
        echo -e "     ${GREEN}Password: <your password>${NC}"
    fi

    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  USEFUL COMMANDS${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    if [ "$INSTALL_METHOD" = "docker" ]; then
        echo -e "  ${BLUE}Docker Commands:${NC}"
        echo -e "  View all containers:    ${GREEN}docker-compose -f $INSTALL_DIR/docker-compose.yml ps${NC}"
        echo -e "  View logs (all):        ${GREEN}docker-compose -f $INSTALL_DIR/docker-compose.yml logs -f${NC}"
        echo -e "  View API logs:          ${GREEN}docker-compose -f $INSTALL_DIR/docker-compose.yml logs -f api${NC}"
        echo -e "  View Dashboard logs:    ${GREEN}docker-compose -f $INSTALL_DIR/docker-compose.yml logs -f dashboard${NC}"
        echo -e "  View DB logs:           ${GREEN}docker-compose -f $INSTALL_DIR/docker-compose.yml logs -f postgres${NC}"
        echo ""
        echo -e "  Restart all services:   ${GREEN}docker-compose -f $INSTALL_DIR/docker-compose.yml restart${NC}"
        echo -e "  Stop all services:      ${GREEN}docker-compose -f $INSTALL_DIR/docker-compose.yml down${NC}"
        echo -e "  Start all services:     ${GREEN}docker-compose -f $INSTALL_DIR/docker-compose.yml up -d${NC}"
        echo ""
        echo -e "  Access database:        ${GREEN}docker exec -it aureo-postgres psql -U aureo_user -d aureo_vpn${NC}"
    else
        echo -e "  ${BLUE}System Commands:${NC}"
        echo -e "  Check API status:       ${GREEN}sudo systemctl status aureo-api${NC}"
        echo -e "  View API logs:          ${GREEN}sudo journalctl -u aureo-api -f${NC}"
        echo -e "  Restart API:            ${GREEN}sudo systemctl restart aureo-api${NC}"
        echo ""
        echo -e "  Check WireGuard:        ${GREEN}sudo systemctl status wg-quick@wg0${NC}"
        echo -e "  View WireGuard logs:    ${GREEN}sudo journalctl -u wg-quick@wg0 -f${NC}"
        echo -e "  Restart WireGuard:      ${GREEN}sudo systemctl restart wg-quick@wg0${NC}"
        echo ""
        echo -e "  Check database:         ${GREEN}sudo systemctl status postgresql${NC}"
        echo -e "  Access database:        ${GREEN}sudo -u postgres psql aureo_vpn${NC}"
    fi

    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  EARNINGS INFO${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "  💰 Reward Tiers:"
    echo -e "     🥉 Bronze:    \$0.01/GB  (50%+ uptime)"
    echo -e "     🥈 Silver:    \$0.015/GB (80%+ uptime)"
    echo -e "     🥇 Gold:      \$0.02/GB  (90%+ uptime)"
    echo -e "     💎 Platinum:  \$0.03/GB  (95%+ uptime)"
    echo ""
    echo -e "  📊 Example Earnings:"
    echo -e "     100 GB/day  × 30 days × \$0.01/GB = \$30/month"
    echo -e "     500 GB/day  × 30 days × \$0.02/GB = \$300/month"
    echo -e "     1000 GB/day × 30 days × \$0.03/GB = \$900/month"
    echo ""
    echo -e "  💸 Minimum payout: \$10"
    echo -e "  📅 Payouts: Weekly (Fridays)"
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  CONFIGURATION FILES${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "  Install Directory:     $INSTALL_DIR"
    echo -e "  Config Directory:      $CONFIG_DIR"
    echo -e "  Operator Credentials:  $CONFIG_DIR/operator-credentials"
    echo ""
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}  🎊 Congratulations! You're now earning crypto rewards! 🎊${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

##############################################################################
# MAIN EXECUTION
##############################################################################

main() {
    print_header
    check_root

    docker stop aureo-api aureo-postgres aureo-dashboard aureo-wireguard 2>/dev/null
    docker rm aureo-api aureo-postgres aureo-dashboard aureo-wireguard 2>/dev/null
    docker volume rm aureo-vpn_postgres-data aureo-vpn_api-logs aureo-vpn_wireguard-config 2>/dev/null
    docker network rm aureo-vpn_aureo-network 2>/dev/null
    sudo rm -rf /opt/aureo-vpn

    # Choose installation method
    choose_installation_method

    echo ""
    echo -e "${YELLOW}This script will perform a complete setup:${NC}"
    echo "  ✓ Install all dependencies"
    echo "  ✓ Setup PostgreSQL database"
    echo "  ✓ Build and start API gateway"
    echo "  ✓ Configure and start VPN node"
    echo "  ✓ Create your operator account"
    echo "  ✓ Register your first VPN node"
    echo "  ✓ Setup web dashboard"
    echo ""
    echo -e "${RED}This will take approximately 10-15 minutes.${NC}"
    echo ""

    read -p "Continue with installation? (y/n): " CONTINUE
    if [[ ! $CONTINUE =~ ^[Yy]$ ]]; then
        echo "Installation cancelled."
        exit 0
    fi

    # Run installation based on selected method
    if [ "$INSTALL_METHOD" = "docker" ]; then
        # Docker installation path
        detect_os
        install_docker
        create_docker_compose
        create_api_dockerfile
        create_dashboard_dockerfile
        copy_project_for_docker
        start_docker_services
        sleep 5  # Wait for services to stabilize
        create_operator_account
        register_node
        setup_dashboard
    else
        # System installation path
        detect_os
        install_system_dependencies
        setup_database
        build_api_gateway
        configure_system_environment
        create_api_service
        start_api_server
        setup_wireguard
        enable_ip_forwarding
        configure_firewall
        start_wireguard
        create_operator_account
        register_node
        setup_dashboard
    fi

    print_summary

    echo -e "\n${GREEN}🚀 All-in-One Setup Completed Successfully!${NC}\n"
}

# Run main
main "$@"
