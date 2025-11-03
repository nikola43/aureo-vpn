.PHONY: help build test clean install docker-build docker-up docker-down setup migrate lint

# Variables
BINARY_DIR=bin
API_GATEWAY_BINARY=$(BINARY_DIR)/api-gateway
CONTROL_SERVER_BINARY=$(BINARY_DIR)/control-server
VPN_NODE_BINARY=$(BINARY_DIR)/vpn-node
CLI_BINARY=$(BINARY_DIR)/aureo-vpn

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOLINT=golangci-lint

# Build parameters
LDFLAGS=-ldflags "-s -w"
BUILD_FLAGS=-v

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

setup: ## Run initial setup
	@echo "Running setup script..."
	@chmod +x scripts/setup.sh
	@./scripts/setup.sh

build: ## Build all binaries
	@echo "Building all services..."
	@mkdir -p $(BINARY_DIR)
	@echo "Building API Gateway..."
	@$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(API_GATEWAY_BINARY) ./cmd/api-gateway
	@echo "Building Control Server..."
	@$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(CONTROL_SERVER_BINARY) ./cmd/control-server
	@echo "Building VPN Node..."
	@$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(VPN_NODE_BINARY) ./cmd/vpn-node
	@echo "Building CLI..."
	@$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(CLI_BINARY) ./cmd/cli
	@echo "Build complete!"

build-api: ## Build API Gateway only
	@mkdir -p $(BINARY_DIR)
	@$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(API_GATEWAY_BINARY) ./cmd/api-gateway

build-control: ## Build Control Server only
	@mkdir -p $(BINARY_DIR)
	@$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(CONTROL_SERVER_BINARY) ./cmd/control-server

build-node: ## Build VPN Node only
	@mkdir -p $(BINARY_DIR)
	@$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(VPN_NODE_BINARY) ./cmd/vpn-node

build-cli: ## Build CLI only
	@mkdir -p $(BINARY_DIR)
	@$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(CLI_BINARY) ./cmd/cli

install: ## Install binaries to /usr/local/bin
	@echo "Installing binaries..."
	@sudo cp $(API_GATEWAY_BINARY) /usr/local/bin/
	@sudo cp $(CONTROL_SERVER_BINARY) /usr/local/bin/
	@sudo cp $(VPN_NODE_BINARY) /usr/local/bin/
	@sudo cp $(CLI_BINARY) /usr/local/bin/
	@echo "Installation complete!"

test: ## Run tests
	@echo "Running tests..."
	@$(GOTEST) -v -race -coverprofile=coverage.out ./...

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	@$(GOTEST) -v -race ./tests/unit/...

test-integration: ## Run integration tests only
	@echo "Running integration tests..."
	@$(GOTEST) -v -race ./tests/integration/...

coverage: test ## Generate coverage report
	@echo "Generating coverage report..."
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linter
	@echo "Running linter..."
	@$(GOLINT) run ./...

fmt: ## Format code
	@echo "Formatting code..."
	@$(GOCMD) fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@$(GOCMD) vet ./...

tidy: ## Tidy go modules
	@echo "Tidying modules..."
	@$(GOMOD) tidy

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@$(GOMOD) download

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BINARY_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete!"

docker-build: ## Build Docker images
	@echo "Building Docker images..."
	@cd deployments/docker && \
		docker build -f Dockerfile.api-gateway -t aureo-vpn/api-gateway:latest ../.. && \
		docker build -f Dockerfile.control-server -t aureo-vpn/control-server:latest ../.. && \
		docker build -f Dockerfile.vpn-node -t aureo-vpn/vpn-node:latest ../..
	@echo "Docker images built!"

docker-up: ## Start Docker Compose services
	@echo "Starting services with Docker Compose..."
	@cd deployments/docker && docker-compose up -d
	@echo "Services started!"

docker-down: ## Stop Docker Compose services
	@echo "Stopping services..."
	@cd deployments/docker && docker-compose down
	@echo "Services stopped!"

docker-logs: ## View Docker Compose logs
	@cd deployments/docker && docker-compose logs -f

run-api: build-api ## Run API Gateway
	@echo "Starting API Gateway..."
	@./$(API_GATEWAY_BINARY)

run-control: build-control ## Run Control Server
	@echo "Starting Control Server..."
	@./$(CONTROL_SERVER_BINARY)

run-node: build-node ## Run VPN Node (requires sudo)
	@echo "Starting VPN Node..."
	@sudo ./$(VPN_NODE_BINARY)

dev-api: ## Run API Gateway in dev mode (with live reload)
	@echo "Running API Gateway in dev mode..."
	@$(GOCMD) run ./cmd/api-gateway/main.go

dev-control: ## Run Control Server in dev mode
	@echo "Running Control Server in dev mode..."
	@$(GOCMD) run ./cmd/control-server/main.go

dev-node: ## Run VPN Node in dev mode (requires sudo)
	@echo "Running VPN Node in dev mode..."
	@sudo $(GOCMD) run ./cmd/vpn-node/main.go

migrate-up: ## Run database migrations
	@echo "Running migrations..."
	@$(GOCMD) run cmd/api-gateway/main.go migrate up

migrate-down: ## Rollback database migrations
	@echo "Rolling back migrations..."
	@$(GOCMD) run cmd/api-gateway/main.go migrate down

db-reset: ## Reset database
	@echo "Resetting database..."
	@psql -h localhost -U postgres -c "DROP DATABASE IF EXISTS aureo_vpn;"
	@psql -h localhost -U postgres -c "CREATE DATABASE aureo_vpn;"

k8s-deploy: ## Deploy to Kubernetes
	@echo "Deploying to Kubernetes..."
	@kubectl apply -f deployments/kubernetes/

k8s-delete: ## Delete from Kubernetes
	@echo "Deleting from Kubernetes..."
	@kubectl delete -f deployments/kubernetes/

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	@$(GOTEST) -bench=. -benchmem ./...

security-scan: ## Run security scan
	@echo "Running security scan..."
	@gosec ./...

.DEFAULT_GOAL := help
