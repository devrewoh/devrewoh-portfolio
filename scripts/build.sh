#!/bin/sh
set -eu

echo "▶ generating templates"
templ generate

echo "▶ building Go binary"

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
go build -ldflags="-s -w" -o bin/devrewoh-portfolio .

echo "✅ build complete"
