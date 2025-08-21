# Build stage - using Ubuntu-based golang image
FROM golang:1.23-bookworm AS builder


# Set proxy environment variables for Go modules
ENV HTTP_PROXY=http://appproxy.airtel.com:4145
ENV HTTPS_PROXY=http://appproxy.airtel.com:4145
ENV NO_PROXY=localhost,127.0.0.1

# Disable Go module proxy and checksum database
ENV GOPROXY=direct
ENV GOSUMDB=off
ENV GO111MODULE=on

# Install build dependencies for CGO and Kafka
RUN apt-get update && apt-get install -y \
    build-essential \
    librdkafka-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go mod files and vendor directory
COPY go.mod go.sum ./
COPY vendor ./vendor

# Copy all source code
COPY . .

# Build with vendor and CGO enabled for Kafka support
# Using dynamic linking with system librdkafka
RUN CGO_ENABLED=1 GOOS=linux GOARCH=$(dpkg --print-architecture) \
    go build -mod=vendor \
    -ldflags="-w -s" \
    -tags dynamic \
    -o /product-views ./cmd/api

# Runtime stage - using Ubuntu slim for smaller size
FROM ubuntu:22.04

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    librdkafka1 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binary from builder
COPY --from=builder /product-views /app/product-views

# Copy migration files
COPY --from=builder /app/migrations /app/migrations

# Create non-root user
RUN useradd -m -u 1000 appuser && chown -R appuser:appuser /app
USER appuser

# Expose port
EXPOSE 8080

# Run the application
CMD ["/app/product-views"]
