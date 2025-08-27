# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Configure Go proxy for better reliability
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies with retry
RUN go mod download || (sleep 5 && go mod download) || (sleep 10 && go mod download)

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy any necessary files (like .env if needed)
COPY --from=builder /app/.env* ./

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./main"]
