# Build stage
FROM golang:1.24-alpine AS builder

# Install dependencies
RUN apk add --no-cache git

WORKDIR /app

# Install build tools
RUN go install github.com/a-h/templ/cmd/templ@latest && \
    go install github.com/magefile/mage@latest

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build for production
RUN mage BuildProd

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Copy binary and static files
COPY --from=builder /app/bin/devrewoh-portfolio .
COPY --from=builder /app/static ./static

# Set ownership and switch to non-root user
RUN chown -R appuser:appgroup /app
USER appuser

# Configure container
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./devrewoh-portfolio"]