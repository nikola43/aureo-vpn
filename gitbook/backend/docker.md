# 🐳 Docker Deployment

Production-ready Docker Compose setup for the Aureo VPN backend.

---

## 🏭 Production Setup

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  redis:
    image: redis:7-alpine
    restart: always
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data

  api-gateway:
    image: ghcr.io/nikola43/aureo-api-gateway:latest
    restart: always
    env_file: .env
    ports:
      - "8080:8080"
    volumes:
      - sqlite_data:/app/data
    depends_on:
      redis:
        condition: service_started
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  control-server:
    image: ghcr.io/nikola43/aureo-control-server:latest
    restart: always
    env_file: .env
    volumes:
      - sqlite_data:/app/data
    depends_on:
      - api-gateway

  prometheus:
    image: prom/prometheus:latest
    restart: always
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus_data:/prometheus
    ports:
      - "9090:9090"

  grafana:
    image: grafana/grafana:latest
    restart: always
    environment:
      GF_SECURITY_ADMIN_PASSWORD: ${GRAFANA_PASSWORD}
    volumes:
      - grafana_data:/var/lib/grafana
    ports:
      - "3000:3000"

volumes:
  sqlite_data:
  redis_data:
  prometheus_data:
  grafana_data:
```

{% hint style="info" %}
SQLite database is stored in a Docker volume for persistence. The database file is created automatically on first run.
{% endhint %}

---

## 🔨 Build Custom Images

```bash
# Build all images
docker-compose -f docker-compose.yml build

# Build specific service
docker build -t aureo-api-gateway -f deployments/docker/Dockerfile.api .
docker build -t aureo-vpn-node -f deployments/docker/Dockerfile.node .
```
