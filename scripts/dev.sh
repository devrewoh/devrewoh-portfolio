#!/bin/sh
set -eu

cleanup() {
    kill $TEMPL_PID $SERVER_PID 2>/dev/null || true
}
trap cleanup EXIT INT TERM

# Watch templates
templ generate --watch &
TEMPL_PID=$!

# Watch and rebuild Go
while true; do
    go build -o bin/devrewoh-portfolio .
    ./bin/devrewoh-portfolio &
    SERVER_PID=$!

    # Wait for file changes
    find . -name '*.go' -o -name '*.templ' | entr -d -r echo "rebuilding..."
    kill $SERVER_PID 2>/dev/null || true
done
