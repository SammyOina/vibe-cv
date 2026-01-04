# Build stage
FROM golang:1.25.3-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s -X main.Version=phase5" \
    -o vibe-cv ./cmd/main.go

# Runtime stage with security hardening
FROM debian:bookworm-slim

# Create non-root user for security
RUN groupadd -g 1000 vibe && useradd -d /app -u 1000 -g vibe -s /sbin/nologin vibe

# Install runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    texlive \
    texlive-latex-extra \
    texlive-fonts-recommended \
    dumb-init \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/vibe-cv .

# Create output directory with correct permissions
RUN mkdir -p /app/outputs && chown -R vibe:vibe /app

# Switch to non-root user
USER vibe

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/api/health || exit 1

# Expose port
EXPOSE 8080

# Use dumb-init to handle signals properly
ENTRYPOINT ["/usr/bin/dumb-init", "--"]

# Run the application
CMD ["./vibe-cv"]

