# syntax=docker/dockerfile:1

# Build stage
FROM golang:tip-alpine3.21 AS builder
WORKDIR /app

# Install dependencies first (better caching)
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy source code
COPY backend/ .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o elang-app ./cmd/main.go

# Run stage
FROM ubuntu:22.04

# Update and install ca-certificates for HTTPS requests
RUN apt-get update && apt-get upgrade -y && \
    apt-get install -y ca-certificates tzdata wget && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Create non-root user for security
RUN groupadd -r appgroup && useradd -r -g appgroup appuser

# Copy binary from builder stage
COPY --from=builder /app/elang-app .

# Change ownership and switch to non-root user
RUN chown -R appuser:appgroup /app
USER appuser

EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./elang-app"]