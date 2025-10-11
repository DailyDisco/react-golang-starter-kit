# Development stage
FROM golang:1.25-alpine AS development

# Install development tools and Air for hot reloading
RUN apk add --no-cache \
    git \
    curl \
    ca-certificates \
    tzdata

# Configure Go
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# Set working directory
WORKDIR /app

# Copy go mod files for better caching
COPY go.mod go.sum ./

# Download dependencies with BuildKit cache mount for faster rebuilds
RUN --mount=type=cache,id=gomod,target=/go/pkg/mod \
    --mount=type=cache,id=gobuild,target=/root/.cache/go-build \
    go mod download && \
    go install github.com/air-verse/air@latest

# Copy source code
COPY . .

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Default command for development (will be overridden by docker-compose)
CMD ["air", "-c", ".air.toml"]

# Production build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates

# Configure Go for production build
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies with BuildKit cache mount
RUN --mount=type=cache,id=gomod,target=/go/pkg/mod \
    go mod download && \
    go mod verify

# Copy source code
COPY . .

# Build the application with production optimizations using build cache
RUN --mount=type=cache,id=gomod,target=/go/pkg/mod \
    --mount=type=cache,id=gobuild,target=/root/.cache/go-build \
    go build \
    -ldflags="-w -s -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" \
    -trimpath \
    -o main \
    ./cmd

# Production runtime stage (Alpine with minimal deps)
FROM alpine:3.19 AS production

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    && rm -rf /var/cache/apk/*

# Create a non-root user for runtime
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder --chown=appuser:appgroup /app/main .

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run the application
CMD ["./main"]

# Production runtime stage (Distroless - smallest and most secure)
FROM gcr.io/distroless/static-debian12:nonroot AS production-distroless

# Copy CA certificates and timezone data from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Distroless images run as non-root by default (uid 65532)
# Expose port
EXPOSE 8080

# Note: Distroless doesn't support healthcheck commands with curl
# Health checks should be handled externally (k8s liveness probes, etc.)

# Run the application
CMD ["./main"]
