#!/bin/bash

# Build script for Daily Assistant
# SQLite requires CGO to be enabled

echo "Building Daily Assistant..."
CGO_ENABLED=1 go build -o assistant

if [ $? -eq 0 ]; then
    echo "✓ Build successful!"
    echo "Run './assistant' to get started"
else
    echo "✗ Build failed"
    exit 1
fi
