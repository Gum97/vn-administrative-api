# 🇻🇳 VN Administrative API

> **Production-Ready** RESTful API cung cấp dữ liệu đơn vị hành chính Việt Nam (Tỉnh, Quận/Huyện, Phường/Xã) với khả năng tìm kiếm theo tên cũ/mới sau sáp nhập.

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](https://docker.com/)
[![Redis](https://img.shields.io/badge/Redis-Cache-DC382D?logo=redis)](https://redis.io/)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

## ✨ Features

- 🚀 **High Performance**: 1000+ req/s với Redis cache + Gzip compression
- 🔍 **Smart Search**: Tìm kiếm theo tên hiện tại VÀ tên trước sáp nhập
- 📦 **Docker Ready**: Multi-stage build, distroless image (~2MB)
- 🛡️ **Production Hardened**: Rate limiting, health checks, graceful shutdown
- 📊 **RESTful API**: Chuẩn JSON response format

## 🚀 Quick Start

### Option 1: Docker Compose (Recommended)

```bash
# Clone repository
git clone https://github.com/gum97/vn-administrative-api.git
cd vn-administrative-api

# Copy và chỉnh sửa config
cp .env.example .env

# Start all services (API + PostgreSQL + Redis)
docker-compose up -d

# Check logs
docker-compose logs -f api

# Crawl data (first time or update)
docker-compose run --rm vn-admin-crawler
```

### Option 2: Manual Build

```bash
# Prerequisites: Go 1.23+, PostgreSQL, Redis

# Install dependencies
go mod download

# Copy config
cp .env.example .env
# Edit .env với database credentials của bạn

# Build
go build -o server ./cmd/server

# Run
./server
```

## 📥 Data Population (Crawler)

Data được crawl từ `sapnhap.bando.com.vn`. Bạn cần chạy crawler để populate database:

### Docker Compose
```bash
# Chạy lần đầu hoặc khi cần update data
docker-compose run --rm crawler
```

### Manual
```bash
# Build crawler
go build -o crawler ./cmd/crawler

# Chạy (cần API_COOKIE từ browser)
./crawler
```

> **Lưu ý**: Lấy `API_COOKIE` bằng cách mở DevTools (F12) khi truy cập `sapnhap.bando.com.vn`, copy giá trị `PHPSESSID` từ Request Headers.

## ⚙️ Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | - | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | - | Database username |
| `DB_PASSWORD` | - | Database password |
| `DB_NAME` | - | Database name |
| `DB_SSLMODE` | `disable` | SSL mode |
| `SERVER_PORT` | `8080` | API server port |
| `REDIS_URL` | - | Redis connection URL |
| `CACHE_TTL` | `5m` | Cache time-to-live |

## 📡 API Endpoints

### Health Checks

| Endpoint | Description | Use Case |
|----------|-------------|----------|
| `GET /health` | Liveness probe | Container orchestration |
| `GET /ready` | Readiness probe | Load balancer health check |

### Data Endpoints

#### Lấy danh sách Tỉnh/Thành phố
```
GET /api/v1/provinces
```
```json
{
    "data": [
        {"id": 1, "tentinh": "Thành phố Hà Nội", "mahc": 1}
    ]
}
```

#### Lấy Quận/Huyện/Phường/Xã theo Tỉnh
```
GET /api/v1/provinces/{id}/units
```

#### Tìm kiếm (hỗ trợ tên cũ/mới)
```
GET /api/v1/search?q=Hà%20Tây
```
> Tìm được các đơn vị **từng thuộc tỉnh Hà Tây** trước khi sáp nhập vào Hà Nội

### Response Format

```json
// Success
{"data": [...]}

// Error
{"error": "Error message"}

// Empty result
{"data": []}
```

## 🐳 Deployment

### Production Deployment

```bash
# Build optimized image
docker build -t vn-admin-api:latest .

# Run với external PostgreSQL và Redis
docker run -d \
  --name vn-admin-api \
  -p 8080:8080 \
  -e DB_HOST=your-postgres-host \
  -e DB_USER=postgres \
  -e DB_PASSWORD=secret \
  -e DB_NAME=vn_admin \
  -e REDIS_URL=redis://your-redis-host:6379 \
  vn-admin-api:latest
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: vn-admin-api
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: api
        image: vn-admin-api:latest
        ports:
        - containerPort: 8080
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "128Mi"
            cpu: "500m"
```

## 📊 Performance

| Metric | Value |
|--------|-------|
| **Throughput** | 1000+ req/s |
| **Latency (p99)** | <50ms |
| **Docker Image** | ~2MB (distroless) |
| **Memory Usage** | ~64MB |
| **DB Connections** | Max 100 |

### Load Testing

```bash
# Install hey
go install github.com/rakyll/hey@latest

# Test provinces endpoint (cached)
hey -n 10000 -c 100 http://localhost:8080/api/v1/provinces

# Test search endpoint
hey -n 5000 -c 50 "http://localhost:8080/api/v1/search?q=hanoi"
```

## 🛠️ Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │────▶│  API Server │────▶│   Redis     │
└─────────────┘     │  (Go 1.23)  │     │   Cache     │
                    └──────┬──────┘     └─────────────┘
                           │
                    ┌──────▼──────┐
                    │ PostgreSQL  │
                    │   (Data)    │
                    └─────────────┘
```

## 🔧 Development

```bash
# Run tests
go test ./...

# Run with race detector
go test -race ./...

# Lint
golangci-lint run

# Format
go fmt ./...
```

## 📝 License

MIT License - Xem [LICENSE](LICENSE) để biết thêm chi tiết.
