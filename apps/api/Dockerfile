# syntax=docker/dockerfile:1.6
# Multi-stage build for Go backend
# BuildKit required: DOCKER_BUILDKIT=1

#############################################
# STAGE 1: Dependencies
#############################################
FROM golang:1.25-alpine AS deps

# Security: Install ca-certificates for TLS verification
RUN apk add --no-cache ca-certificates git

WORKDIR /app

# Copy dependency files first for layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

#############################################
# STAGE 2: Builder
#############################################
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache ca-certificates git build-base

WORKDIR /app

# Copy dependencies from deps stage
COPY --from=deps /go/pkg/mod /go/pkg/mod

# Copy source code
COPY . .

# Build arguments for versioning
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT
ARG GIT_BRANCH

# Build the binary with security flags
# -ldflags:
#   -s -w: Strip debug info (smaller binary)
#   -extldflags '-static': Static linking
# CGO_ENABLED=0: Disable CGO for static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w \
    -X main.Version=${VERSION} \
    -X main.BuildTime=${BUILD_TIME} \
    -X main.GitCommit=${GIT_COMMIT} \
    -X main.GitBranch=${GIT_BRANCH} \
    -extldflags '-static'" \
    -a -installsuffix cgo \
    -o /app/server \
    ./cmd/server

# Run tests during build (fails if tests fail)
RUN go test -v -race -coverprofile=coverage.out ./...

#############################################
# STAGE 3: Security Scanner
#############################################
FROM golang:1.25-alpine AS security

WORKDIR /app

# Install security scanning tools
RUN go install github.com/securego/gosec/v2/cmd/gosec@latest

COPY --from=builder /app /app

# Run SAST scan
RUN gosec -fmt sarif -out /app/gosec-report.sarif ./... || true

#############################################
# STAGE 4: WireGuard tools builder
#############################################
FROM debian:12-slim AS wg-tools
RUN apt-get update && apt-get install -y --no-install-recommends \
    wireguard-tools iproute2 iptables ca-certificates \
    && rm -rf /var/lib/apt/lists/*

#############################################
# STAGE 5: Production Image
#############################################
FROM debian:12-slim AS production

# Install runtime dependencies (setpriv from util-linux replaces su-exec)
RUN apt-get update && apt-get install -y --no-install-recommends \
    wireguard-tools iproute2 iptables ca-certificates wget \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN groupadd -r vexa && useradd -r -g vexa -d /app -s /sbin/nologin vexa

# Create writable directories
RUN mkdir -p /app/logs /etc/wireguard /var/lib/vexa/secrets && \
    chown -R vexa:vexa /app /etc/wireguard /var/lib/vexa

# Copy binary and entrypoint
COPY --from=builder --chown=vexa:vexa /app/server /app/server
COPY --from=builder --chown=vexa:vexa /app/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

USER root
WORKDIR /app

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/health/live > /dev/null || exit 1

# Expose ports
EXPOSE 8080 51820/udp

# Set environment
ENV PORT=8080
ENV ENV=production
ENV WG_ENABLED=true
ENV WG_MOCK_FALLBACK=true

# Run the server via entrypoint (drops privileges)
ENTRYPOINT ["/app/entrypoint.sh"]
CMD []
