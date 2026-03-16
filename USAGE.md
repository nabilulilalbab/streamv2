# Usage Guide

## 🚀 How to Run

### Important: Running Go Files

When running Go applications with multiple files in the same package, you must include **ALL files** in the command:

```bash
# ✅ CORRECT - Include all files with wildcard
go run cmd/api/*.go

# ❌ WRONG - Running single file will cause "undefined" errors
go run cmd/api/main.go
go run cmd/api/test_scraper.go
```

---

## 📖 Running Modes

### 1. Test Scraper Mode

Test the scraper functionality without starting the server:

```bash
go run cmd/api/*.go test-scraper
```

**Output:**
```
🧪 Testing IDLIX Scraper...
===================================================
✅ HTTP Client created successfully
✅ IDLIX Repository created

📡 Fetching featured movies...
✅ Found 18 featured movies

Movie 1:
  Title:  Crime 101 (2026)
  URL:    https://tv12.idlixku.com/movie/crime-101-2026/
  ...

✅ All scraper tests passed!
```

---

### 2. Server Mode (Default)

Start the HTTP API server:

```bash
go run cmd/api/*.go
```

**Output:**
```
🚀 Starting IDLIX API Server...
✅ HTTP Client initialized
✅ IDLIX Repository initialized
✅ JeniusPlay Repository initialized
✅ M3U8 Parser initialized
✅ IDLIX Service initialized
✅ Handlers initialized

✅ Server is running on http://0.0.0.0:8080
📡 API Endpoints:
   GET  http://localhost:8080/api/v1/health
   GET  http://localhost:8080/api/v1/featured
   POST http://localhost:8080/api/v1/video/info
```

---

## 🔨 Building Binary

For production, build a single binary:

```bash
# Build
go build -o idlix-api cmd/api/*.go

# Run
./idlix-api
```

**With custom port:**
```bash
PORT=9000 ./idlix-api
```

---

## 🧪 Testing API Endpoints

### 1. Health Check

```bash
curl http://localhost:8080/api/v1/health
```

**Response:**
```json
{
  "status": "ok",
  "version": "1.0.0",
  "message": "IDLIX API is running"
}
```

---

### 2. Featured Movies

```bash
curl http://localhost:8080/api/v1/featured
```

**Response:**
```json
{
  "status": true,
  "message": "Featured movies retrieved successfully",
  "data": {
    "movies": [
      {
        "url": "https://tv12.idlixku.com/movie/crime-101-2026/",
        "title": "Crime 101 (2026)",
        "year": "",
        "type": "",
        "poster": "https://image.tmdb.org/t/p/w185/..."
      }
    ]
  }
}
```

---

### 3. Video Info

```bash
curl -X POST http://localhost:8080/api/v1/video/info \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://tv12.idlixku.com/movie/crime-101-2026/"
  }'
```

**Response:**
```json
{
  "status": true,
  "message": "Video info retrieved successfully",
  "data": {
    "video_id": "163426",
    "video_name": "Crime 101 (2026)",
    "poster": "https://image.tmdb.org/t/p/w185/...",
    "embed_url": "https://jeniusplay.com/video/...",
    "m3u8_url": "https://jeniusplay.com/cdn/hls/.../master.m3u8",
    "is_variant_playlist": true,
    "variants": [
      {
        "id": "0",
        "resolution": "1920x804",
        "bandwidth": 1510000,
        "uri": "https://..."
      }
    ],
    "subtitle": {
      "available": true,
      "url": "https://...",
      "format": "vtt"
    }
  }
}
```

---

## ⚙️ Configuration

### Environment Variables

Create a `.env` file or set environment variables:

```bash
# Server
PORT=8080
HOST=0.0.0.0
GIN_MODE=release

# IDLIX
IDLIX_BASE_URL=https://tv12.idlixku.com/
IDLIX_TIMEOUT=30s
IDLIX_RETRY=3
```

### Loading .env File

```bash
# Install godotenv
go get github.com/joho/godotenv

# Load .env before running
export $(cat .env | xargs) && go run cmd/api/*.go
```

---

## 🐳 Docker

### Build Image

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o idlix-api cmd/api/*.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/idlix-api .
EXPOSE 8080
CMD ["./idlix-api"]
```

```bash
docker build -t idlix-api .
docker run -p 8080:8080 idlix-api
```

---

## 🔧 Development

### Hot Reload (with Air)

Install Air:
```bash
go install github.com/air-verse/air@latest
```

Create `.air.toml`:
```toml
[build]
  cmd = "go build -o ./tmp/main cmd/api/*.go"
  bin = "./tmp/main"
  include_ext = ["go"]
  exclude_dir = ["tmp", "vendor"]
```

Run:
```bash
air
```

---

## 📊 Testing

### Unit Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/utils/
```

### Integration Test

```bash
# Run scraper test
go run cmd/api/*.go test-scraper

# Build and run
go build -o idlix-api cmd/api/*.go
./idlix-api test-scraper
```

---

## 🚨 Common Issues

### Issue 1: "undefined: testScraper"

**Problem:**
```bash
go run cmd/api/main.go
# Error: undefined: testScraper
```

**Solution:**
```bash
# Use wildcard to include all files
go run cmd/api/*.go
```

---

### Issue 2: "connection reset by peer"

**Problem:** Cloudflare rate limiting

**Solution:** Add delay between requests (1-2 seconds)

---

### Issue 3: Module not found

**Problem:**
```bash
go run cmd/api/*.go
# Error: module not found
```

**Solution:**
```bash
# Download dependencies
go mod download
go mod tidy
```

---

## 💡 Tips

### 1. Rate Limiting

Add delays between requests to avoid Cloudflare blocks:

```go
import "time"

// Between requests
time.Sleep(1 * time.Second)
```

### 2. Custom Port

```bash
PORT=9000 go run cmd/api/*.go
```

### 3. Debug Mode

```bash
GIN_MODE=debug go run cmd/api/*.go
```

### 4. Production Build

```bash
# Build with optimizations
go build -ldflags="-s -w" -o idlix-api cmd/api/*.go

# Check size
ls -lh idlix-api
```

---

## 📚 More Documentation

- [API Documentation](docs/API_FLOW_TEST_PLAN.md)
- [Cloudflare Bypass](docs/CLOUDFLARE_FIX_SUMMARY.md)
- [Rate Limiting](docs/RATE_LIMITING_ANALYSIS.md)
- [Implementation Details](docs/IMPLEMENTATION_SUMMARY.md)

---

**Last Updated:** 2026-03-17
