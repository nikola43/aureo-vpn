#!/bin/bash

# Aureo VPN Setup Script
# This script sets up the development environment for Aureo VPN
# Supports: macOS, Linux (Ubuntu/Debian, Fedora/RHEL, Arch)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
MIN_GO_VERSION="1.22"
POSTGRES_VERSION="15"
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DIR="${PROJECT_DIR}/bin"

# Trap errors and cleanup
trap 'error_handler $? $LINENO' ERR

error_handler() {
    echo -e "\n${RED}✗ Error occurred in setup script at line $2 (exit code: $1)${NC}"
    echo -e "${YELLOW}Cleaning up...${NC}"
    cleanup_on_error
    exit 1
}

cleanup_on_error() {
    # Stop any containers we started
    if [ "$POSTGRES_STARTED" = "true" ]; then
        echo "Stopping PostgreSQL container..."
        docker stop aureo-vpn-db 2>/dev/null || true
        docker rm aureo-vpn-db 2>/dev/null || true
    fi
    if [ "$REDIS_STARTED" = "true" ]; then
        echo "Stopping Redis container..."
        docker stop aureo-vpn-redis 2>/dev/null || true
        docker rm aureo-vpn-redis 2>/dev/null || true
    fi
}

# Print header
print_header() {
    echo ""
    echo -e "${BLUE}======================================"
    echo "    Aureo VPN Setup Script"
    echo "======================================"
    echo -e "${NC}"
}

# Check if running as root
check_root() {
    if [ "$EUID" -eq 0 ]; then
        echo -e "${RED}✗ Please do not run this script as root${NC}"
        echo "  Run without sudo. Script will ask for password when needed."
        exit 1
    fi
}

# Detect OS
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="linux"
        if [ -f /etc/os-release ]; then
            . /etc/os-release
            DISTRO=$ID
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="macos"
        DISTRO="macos"
    else
        OS="unknown"
        DISTRO="unknown"
    fi
    echo -e "${BLUE}Detected OS: $OS ($DISTRO)${NC}"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check Go version
check_go_version() {
    local version=$(go version | awk '{print $3}' | sed 's/go//')
    local major=$(echo $version | cut -d. -f1)
    local minor=$(echo $version | cut -d. -f2)
    local min_major=$(echo $MIN_GO_VERSION | cut -d. -f1)
    local min_minor=$(echo $MIN_GO_VERSION | cut -d. -f2)

    if [ "$major" -gt "$min_major" ] || ([ "$major" -eq "$min_major" ] && [ "$minor" -ge "$min_minor" ]); then
        return 0
    else
        return 1
    fi
}

# Install dependencies based on OS
install_dependencies() {
    echo -e "\n${BLUE}Installing system dependencies...${NC}"

    case $DISTRO in
        ubuntu|debian)
            echo "Installing dependencies for Ubuntu/Debian..."
            sudo apt-get update
            sudo apt-get install -y \
                build-essential \
                curl \
                git \
                wget \
                postgresql-client \
                wireguard \
                wireguard-tools \
                iptables \
                iproute2 \
                openssl \
                jq
            ;;
        fedora|rhel|centos)
            echo "Installing dependencies for Fedora/RHEL..."
            sudo dnf install -y \
                gcc \
                make \
                curl \
                git \
                wget \
                postgresql \
                wireguard-tools \
                iptables \
                iproute \
                openssl \
                jq
            ;;
        arch)
            echo "Installing dependencies for Arch Linux..."
            sudo pacman -S --noconfirm \
                base-devel \
                curl \
                git \
                wget \
                postgresql-libs \
                wireguard-tools \
                iptables \
                iproute2 \
                openssl \
                jq
            ;;
        macos)
            echo "Installing dependencies for macOS..."
            if ! command_exists brew; then
                echo -e "${YELLOW}Installing Homebrew...${NC}"
                /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
            fi
            brew install postgresql wireguard-tools openssl jq
            ;;
        *)
            echo -e "${YELLOW}⚠ Unknown distribution. Please install dependencies manually.${NC}"
            ;;
    esac

    echo -e "${GREEN}✓ System dependencies installed${NC}"
}

# Check dependencies
check_dependencies() {
    echo -e "\n${BLUE}Checking dependencies...${NC}"
    local missing_deps=0

    # Check Go
    if ! command_exists go; then
        echo -e "${RED}✗ Go is not installed${NC}"
        echo "  Install Go from: https://go.dev/dl/"
        missing_deps=1
    else
        if check_go_version; then
            GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
            echo -e "${GREEN}✓ Go $GO_VERSION installed${NC}"
        else
            echo -e "${RED}✗ Go version must be >= $MIN_GO_VERSION${NC}"
            missing_deps=1
        fi
    fi

    # Check Docker
    if ! command_exists docker; then
        echo -e "${YELLOW}⚠ Docker is not installed (recommended for development)${NC}"
        echo "  Install from: https://docs.docker.com/get-docker/"
    else
        DOCKER_VERSION=$(docker --version | awk '{print $3}' | sed 's/,//')
        echo -e "${GREEN}✓ Docker $DOCKER_VERSION installed${NC}"

        # Check if Docker daemon is running
        if ! docker info >/dev/null 2>&1; then
            echo -e "${YELLOW}⚠ Docker is installed but not running${NC}"
            echo "  Please start Docker Desktop or Docker daemon"
        fi
    fi

    # Check PostgreSQL client
    if ! command_exists psql; then
        echo -e "${YELLOW}⚠ PostgreSQL client not installed (will use Docker)${NC}"
    else
        PSQL_VERSION=$(psql --version | awk '{print $3}')
        echo -e "${GREEN}✓ PostgreSQL client $PSQL_VERSION installed${NC}"
    fi

    # Check WireGuard
    if ! command_exists wg; then
        echo -e "${YELLOW}⚠ WireGuard not installed (required for VPN nodes)${NC}"
        echo "  Install from: https://www.wireguard.com/install/"
    else
        echo -e "${GREEN}✓ WireGuard installed${NC}"
    fi

    # Check git
    if ! command_exists git; then
        echo -e "${RED}✗ Git is not installed${NC}"
        missing_deps=1
    else
        echo -e "${GREEN}✓ Git installed${NC}"
    fi

    # Check make
    if ! command_exists make; then
        echo -e "${YELLOW}⚠ Make is not installed (optional but recommended)${NC}"
    else
        echo -e "${GREEN}✓ Make installed${NC}"
    fi

    # Check openssl
    if ! command_exists openssl; then
        echo -e "${RED}✗ OpenSSL is not installed${NC}"
        missing_deps=1
    else
        echo -e "${GREEN}✓ OpenSSL installed${NC}"
    fi

    if [ $missing_deps -eq 1 ]; then
        echo -e "\n${RED}Some required dependencies are missing.${NC}"
        read -p "Would you like to install them automatically? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            install_dependencies
        else
            echo -e "${RED}Cannot continue without required dependencies.${NC}"
            exit 1
        fi
    fi
}

# Setup Go dependencies
setup_go_dependencies() {
    echo -e "\n${BLUE}Setting up Go dependencies...${NC}"
    cd "$PROJECT_DIR"

    echo "Downloading dependencies..."
    go mod download

    echo "Tidying go.mod..."
    go mod tidy

    echo "Verifying dependencies..."
    go mod verify

    echo -e "${GREEN}✓ Go dependencies installed${NC}"
}

# Setup PostgreSQL
setup_postgresql() {
    echo -e "\n${BLUE}Setting up PostgreSQL...${NC}"

    if command_exists docker && docker info >/dev/null 2>&1; then
        read -p "Do you want to set up PostgreSQL using Docker? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            # Check if container already exists
            if docker ps -a | grep -q aureo-vpn-db; then
                echo "PostgreSQL container already exists."
                read -p "Do you want to remove and recreate it? (y/n) " -n 1 -r
                echo
                if [[ $REPLY =~ ^[Yy]$ ]]; then
                    docker stop aureo-vpn-db 2>/dev/null || true
                    docker rm aureo-vpn-db 2>/dev/null || true
                else
                    echo "Using existing PostgreSQL container."
                    return
                fi
            fi

            echo "Starting PostgreSQL container..."
            docker run -d \
                --name aureo-vpn-db \
                -e POSTGRES_DB=aureo_vpn \
                -e POSTGRES_USER=postgres \
                -e POSTGRES_PASSWORD=postgres \
                -p 5432:5432 \
                --restart unless-stopped \
                postgres:15-alpine

            POSTGRES_STARTED=true

            echo -e "${GREEN}✓ PostgreSQL container started${NC}"
            echo "Waiting for PostgreSQL to be ready..."

            # Wait for PostgreSQL to be ready
            for i in {1..30}; do
                if docker exec aureo-vpn-db pg_isready -U postgres >/dev/null 2>&1; then
                    echo -e "${GREEN}✓ PostgreSQL is ready${NC}"
                    break
                fi
                echo -n "."
                sleep 1
            done
            echo
        fi
    else
        echo -e "${YELLOW}Docker not available. Please set up PostgreSQL manually.${NC}"
        echo "Connection details for manual setup:"
        echo "  Database: aureo_vpn"
        echo "  User: postgres"
        echo "  Port: 5432"
    fi
}

# Setup Redis
setup_redis() {
    echo -e "\n${BLUE}Setting up Redis...${NC}"

    if command_exists docker && docker info >/dev/null 2>&1; then
        read -p "Do you want to set up Redis using Docker? (recommended) (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            # Check if container already exists
            if docker ps -a | grep -q aureo-vpn-redis; then
                echo "Redis container already exists."
                read -p "Do you want to remove and recreate it? (y/n) " -n 1 -r
                echo
                if [[ $REPLY =~ ^[Yy]$ ]]; then
                    docker stop aureo-vpn-redis 2>/dev/null || true
                    docker rm aureo-vpn-redis 2>/dev/null || true
                else
                    echo "Using existing Redis container."
                    return
                fi
            fi

            echo "Starting Redis container..."
            docker run -d \
                --name aureo-vpn-redis \
                -p 6379:6379 \
                --restart unless-stopped \
                redis:7-alpine

            REDIS_STARTED=true

            echo -e "${GREEN}✓ Redis container started${NC}"
            sleep 2
        fi
    fi
}

# Create environment file
create_env_file() {
    echo -e "\n${BLUE}Creating environment file...${NC}"

    if [ -f "$PROJECT_DIR/.env" ]; then
        echo "Environment file already exists."
        read -p "Do you want to overwrite it? (y/n) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "Keeping existing .env file."
            return
        fi
        cp "$PROJECT_DIR/.env" "$PROJECT_DIR/.env.backup"
        echo "Backup created at .env.backup"
    fi

    # Generate secure JWT secret
    JWT_SECRET=$(openssl rand -base64 32)

    cat > "$PROJECT_DIR/.env" << EOF
# Aureo VPN Configuration
# Generated on $(date)

# ============================================
# Database Configuration
# ============================================
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=aureo_vpn
DB_SSL_MODE=disable

# ============================================
# JWT Configuration
# ============================================
JWT_SECRET=${JWT_SECRET}

# ============================================
# API Configuration
# ============================================
PORT=8080
ENVIRONMENT=development

# ============================================
# Redis Configuration
# ============================================
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# ============================================
# VPN Configuration
# ============================================
# Set this when running a VPN node
NODE_ID=

# ============================================
# Security Configuration
# ============================================
RATE_LIMIT_ENABLED=true
RATE_LIMIT_MAX_REQUESTS=100
CORS_ENABLED=true

# ============================================
# Monitoring
# ============================================
PROMETHEUS_PORT=9090
GRAFANA_PORT=3000

# ============================================
# Logging
# ============================================
LOG_LEVEL=info
LOG_FORMAT=json
EOF

    echo -e "${GREEN}✓ Environment file created (.env)${NC}"
    echo -e "${YELLOW}⚠ JWT_SECRET has been randomly generated${NC}"
}

# Build applications
build_applications() {
    echo -e "\n${BLUE}Building applications...${NC}"
    cd "$PROJECT_DIR"

    # Create bin directory
    mkdir -p "$BIN_DIR"

    echo "Building API Gateway..."
    go build -v -o "$BIN_DIR/api-gateway" ./cmd/api-gateway

    echo "Building Control Server..."
    go build -v -o "$BIN_DIR/control-server" ./cmd/control-server

    echo "Building VPN Node..."
    go build -v -o "$BIN_DIR/vpn-node" ./cmd/vpn-node

    echo "Building CLI..."
    go build -v -o "$BIN_DIR/aureo-vpn" ./cmd/cli

    echo -e "${GREEN}✓ Applications built successfully${NC}"

    # Make binaries executable
    chmod +x "$BIN_DIR"/*
}

# Run tests
run_tests() {
    echo -e "\n${BLUE}Running tests...${NC}"
    read -p "Do you want to run tests? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        cd "$PROJECT_DIR"
        echo "Running unit tests..."
        go test -v -short ./... || echo -e "${YELLOW}⚠ Some tests failed${NC}"
        echo -e "${GREEN}✓ Tests completed${NC}"
    fi
}

# Verify setup
verify_setup() {
    echo -e "\n${BLUE}Verifying setup...${NC}"
    local errors=0

    # Check binaries
    for binary in api-gateway control-server vpn-node aureo-vpn; do
        if [ -f "$BIN_DIR/$binary" ]; then
            echo -e "${GREEN}✓ $binary built${NC}"
        else
            echo -e "${RED}✗ $binary not found${NC}"
            errors=$((errors + 1))
        fi
    done

    # Check environment file
    if [ -f "$PROJECT_DIR/.env" ]; then
        echo -e "${GREEN}✓ .env file exists${NC}"
    else
        echo -e "${RED}✗ .env file not found${NC}"
        errors=$((errors + 1))
    fi

    # Check PostgreSQL
    if command_exists docker && docker ps | grep -q aureo-vpn-db; then
        echo -e "${GREEN}✓ PostgreSQL container running${NC}"
    else
        echo -e "${YELLOW}⚠ PostgreSQL container not running${NC}"
    fi

    # Check Redis
    if command_exists docker && docker ps | grep -q aureo-vpn-redis; then
        echo -e "${GREEN}✓ Redis container running${NC}"
    else
        echo -e "${YELLOW}⚠ Redis container not running${NC}"
    fi

    if [ $errors -eq 0 ]; then
        echo -e "\n${GREEN}✓ All checks passed!${NC}"
    else
        echo -e "\n${YELLOW}⚠ Setup completed with $errors error(s)${NC}"
    fi
}

# Print next steps
print_next_steps() {
    echo -e "\n${GREEN}======================================"
    echo "  Setup Complete!"
    echo "======================================${NC}"
    echo ""
    echo "Next steps:"
    echo ""
    echo -e "${BLUE}1. Start the API Gateway:${NC}"
    echo "   source .env && ./bin/api-gateway"
    echo ""
    echo -e "${BLUE}2. Start the Control Server (in another terminal):${NC}"
    echo "   source .env && ./bin/control-server"
    echo ""
    echo -e "${BLUE}3. Create a VPN node:${NC}"
    echo "   ./bin/aureo-vpn node create \\"
    echo "     --name 'Test-Node' \\"
    echo "     --hostname 'test.local' \\"
    echo "     --ip '127.0.0.1' \\"
    echo "     --country 'Test' \\"
    echo "     --country-code 'TS' \\"
    echo "     --city 'Test City'"
    echo ""
    echo -e "${BLUE}4. Start the VPN node (in another terminal):${NC}"
    echo "   source .env && export NODE_ID=<node-id> && sudo ./bin/vpn-node"
    echo ""
    echo -e "${BLUE}5. Test the API:${NC}"
    echo "   curl http://localhost:8080/health"
    echo ""
    echo -e "${BLUE}6. Or use Docker Compose:${NC}"
    echo "   make docker-up"
    echo ""
    echo -e "${GREEN}Useful commands:${NC}"
    echo "  make help          - Show all make targets"
    echo "  make test          - Run tests"
    echo "  make docker-up     - Start with Docker Compose"
    echo "  make docker-down   - Stop Docker services"
    echo ""
    echo -e "${BLUE}Documentation:${NC}"
    echo "  README.md                    - Quick start guide"
    echo "  QUICK_REFERENCE.md           - Command reference"
    echo "  docs/API.md                  - API documentation"
    echo "  docs/DEPLOYMENT.md           - Deployment guide"
    echo "  IMPLEMENTATION_SUMMARY.md    - Complete feature list"
    echo ""
    echo -e "${YELLOW}Note: For VPN node operations, you need sudo/root privileges${NC}"
    echo ""
}

# Main setup flow
main() {
    print_header
    check_root
    detect_os
    check_dependencies
    setup_go_dependencies
    setup_postgresql
    setup_redis
    create_env_file
    build_applications
    run_tests
    verify_setup
    print_next_steps
}

# Run main function
main

exit 0
