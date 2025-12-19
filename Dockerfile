# Build stage
FROM golang:1.25.5-alpine AS builder

WORKDIR /app

# Install build tools
RUN apk add --no-cache ca-certificates git && \
    go install github.com/a-h/templ/cmd/templ@latest

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Generate templates and build
RUN templ generate && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o bin/devrewoh-portfolio .

# Runtime stage
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates

# Copy binary and static files
COPY --from=builder /app/bin/devrewoh-portfolio ./
COPY --from=builder /app/static ./static/

EXPOSE 8080

CMD ["./devrewoh-portfolio"]
