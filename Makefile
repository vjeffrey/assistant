.PHONY: build clean test check-creds help install run-daemon run-web run-daemon-web fmt lint

# Default target
.DEFAULT_GOAL := help

# Build the assistant binary
build:
	@echo "Building Daily Assistant..."
	@./build.sh

# Install to ~/go/bin
install:
	@echo "Installing to ~/go/bin/assistant..."
	@CGO_ENABLED=1 go install

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f assistant .assistant.tmp
	@echo "✓ Clean complete"

# Run tests
test:
	@echo "Running tests..."
	@CGO_ENABLED=1 go test -v ./...

# Check for credentials in the repository
check-creds:
	@./check-credentials.sh

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Code formatted"

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: brew install golangci-lint"; \
		exit 1; \
	fi

# Run daemon
run-daemon: build
	@./assistant --daemon

# Run web UI only
run-web: build
	@./assistant --web 8080

# Run daemon with web UI
run-daemon-web: build
	@./assistant --daemon --web 8080

# Pre-commit checks (run before committing)
pre-commit: check-creds fmt test
	@echo ""
	@echo "✓ All pre-commit checks passed!"
	@echo "  - No credentials found"
	@echo "  - Code formatted"
	@echo "  - Tests passed"
	@echo ""
	@echo "Safe to commit!"

# Show help
help:
	@echo "Daily Assistant - Makefile Commands"
	@echo ""
	@echo "Build Commands:"
	@echo "  make build           Build the assistant binary"
	@echo "  make install         Install to ~/go/bin/assistant"
	@echo "  make clean           Remove build artifacts"
	@echo ""
	@echo "Development Commands:"
	@echo "  make fmt             Format Go code"
	@echo "  make lint            Run golangci-lint"
	@echo "  make test            Run tests"
	@echo ""
	@echo "Security Commands:"
	@echo "  make check-creds     Check repository for credentials/secrets"
	@echo "  make pre-commit      Run all pre-commit checks (creds, fmt, test)"
	@echo ""
	@echo "Run Commands:"
	@echo "  make run-daemon      Build and run daemon"
	@echo "  make run-web         Build and run web UI (port 8080)"
	@echo "  make run-daemon-web  Build and run daemon with web UI"
	@echo ""
	@echo "Default: make help"
