# Build stage
FROM golang:1.25.5-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Install build tools
RUN go install github.com/a-h/templ/cmd/templ@latest && \
    go install github.com/magefile/mage@latest

# Copy dependency files first for better Docker layer caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Generate templates and build for production
RUN mage Generate && mage BuildProd

# Verify the binary was created
RUN ls -la bin/ && test -f bin/devrewoh-portfolio

# Runtime stage
FROM alpine:latest

# Install runtime dependencies and create user in one layer
RUN apk --no-cache add ca-certificates tzdata wget && \
    addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup && \
    mkdir -p /app/static

WORKDIR /app

# Copy binary and static files with proper ownership
COPY --from=builder --chown=appuser:appgroup /app/bin/devrewoh-portfolio ./
COPY --from=builder --chown=appuser:appgroup /app/static ./static/

# Switch to non-root user
USER appuser

# Configure container
EXPOSE 8080

# Health check (using wget since it's already installed)
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Use exec form for proper signal handling
CMD ["./devrewoh-portfolio"]
