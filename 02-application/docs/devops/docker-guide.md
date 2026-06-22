# vexa — Docker Operations

> **Agent:** DevOps Engineer  
> **Status:** Generated 2026-05-28  
> **Version:** 0.1.0

---

## 1. Docker Images

### 1.1 Build All

```bash
# Build all images
docker compose build

# Build specific service
docker compose build api
docker compose build web

# Build with no cache
docker compose build --no-cache
```

### 1.2 Push to Registry

```bash
# Tag
docker tag vexa-api ghcr.io/soumabali/vexa/api:latest
docker tag vexa-web ghcr.io/soumabali/vexa/web:latest

# Push
docker push ghcr.io/soumabali/vexa/api:latest
docker push ghcr.io/soumabali/vexa/web:latest
```

---

## 2. Container Operations

### 2.1 Start/Stop/Restart

```bash
# Start all
docker compose up -d

# Stop all
docker compose down

# Restart service
docker compose restart api
docker compose restart web

# Recreate service
docker compose up -d --force-recreate api
```

### 2.2 Logs

```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f api

# Last 100 lines
docker compose logs --tail=100 api

# Since timestamp
docker compose logs --since=2026-05-28T10:00:00 api
```

### 2.3 Exec

```bash
# Shell into API container
docker compose exec api sh

# Run Go tests
docker compose exec api go test ./...

# Database shell
docker compose exec db psql -U vexa -d vexa

# Redis CLI
docker compose exec redis redis-cli
```

---

## 3. Volume Management

### 3.1 List Volumes

```bash
docker volume ls | grep vexa
```

### 3.2 Backup Volume

```bash
# Backup PostgreSQL data
docker run --rm -v vexa_postgres_data:/data -v $(pwd):/backup alpine \
  tar czf /backup/postgres-backup.tar.gz -C /data .

# Backup Redis data
docker run --rm -v vexa_redis_data:/data -v $(pwd):/backup alpine \
  tar czf /backup/redis-backup.tar.gz -C /data .
```

### 3.3 Restore Volume

```bash
# Restore PostgreSQL data
docker run --rm -v vexa_postgres_data:/data -v $(pwd):/backup alpine \
  tar xzf /backup/postgres-backup.tar.gz -C /data
```

---

## 4. Network

### 4.1 List Networks

```bash
docker network ls | grep vexa
```

### 4.2 Inspect Network

```bash
docker network inspect vexa_default
```

---

## 5. Cleanup

### 5.1 Remove Unused

```bash
# Remove stopped containers
docker container prune

# Remove unused images
docker image prune

# Remove unused volumes
docker volume prune

# Remove unused networks
docker network prune

# Remove everything unused
docker system prune -a
```

### 5.2 Full Cleanup

```bash
# Stop and remove everything
docker compose down -v --rmi all

# Remove all project containers
docker rm -f $(docker ps -a | grep vexa | awk '{print $1}')
```

---

## 6. Multi-Stage Build Optimization

### 6.1 API Multi-Stage

```dockerfile
# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server

# Runtime stage
FROM alpine:3.20
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

### 6.2 Web Multi-Stage

```dockerfile
# Dependencies stage
FROM node:22-alpine AS deps
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN npm install -g pnpm && pnpm install --frozen-lockfile

# Build stage
FROM node:22-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN pnpm build

# Runtime stage
FROM node:22-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public
EXPOSE 3000
CMD ["node", "server.js"]
```
