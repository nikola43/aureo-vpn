# Aureo VPN - Decentralized Node
# Single binary running API + VPN + P2P with SQLite

FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the decentralized node binary
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /aureo-node ./cmd/aureo-node/

# Runtime stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache \
    wireguard-tools \
    iptables \
    ip6tables \
    iproute2 \
    curl \
    ca-certificates \
    bash

# Create data directory
RUN mkdir -p /data /config

# Copy binary from builder
COPY --from=builder /aureo-node /usr/local/bin/aureo-node
RUN chmod +x /usr/local/bin/aureo-node

# Environment variables with defaults
ENV NODE_NAME="" \
    DATA_DIR="/data" \
    PUBLIC_IP="" \
    API_PORT="8080" \
    P2P_PORT="4001" \
    WG_PORT="51820" \
    COUNTRY="" \
    COUNTRY_CODE="" \
    CITY="" \
    LATITUDE="" \
    LONGITUDE="" \
    BOOTSTRAP_PEERS="" \
    ENABLE_API="true" \
    ENABLE_VPN="true"

# Expose ports
EXPOSE 8080/tcp
EXPOSE 4001/tcp
EXPOSE 4001/udp
EXPOSE 51820/udp

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:${API_PORT}/health || exit 1

# Entry point script
COPY deployments/decentralized/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
