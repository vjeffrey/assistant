#!/bin/bash

# Build script for Daily Assistant
# SQLite requires CGO to be enabled
# Note: macOS XProtect aggressively blocks CGO binaries
# See XPROTECT_ISSUE.md for details and solutions

echo "Building Daily Assistant..."
CGO_ENABLED=1 go build -ldflags="-w -s" -o .assistant.tmp

if [ $? -eq 0 ]; then
    # Copy (not move) to work around XProtect extended attributes
    cp .assistant.tmp assistant
    rm .assistant.tmp
    echo "✓ Build successful!"
    echo ""
    echo "⚠️  NOTE: macOS XProtect may block this binary on first run."
    echo "   If blocked, go to System Settings > Privacy & Security"
    echo "   and click 'Allow Anyway'. See XPROTECT_ISSUE.md for details."
    echo ""
    echo "Run './assistant' to get started"
else
    echo "✗ Build failed"
    exit 1
fi
