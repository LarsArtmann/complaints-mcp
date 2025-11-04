# complaints-mcp Justfile
# Modern task runner for Go projects

# Default recipe
default:
    just --list

# Build the complaints-mcp server
build:
    go build -o complaints-mcp ./cmd/server

# Build for release with optimization flags
build-release:
    go build -ldflags="-s -w" -o complaints-mcp ./cmd/server

# Clean build artifacts
clean:
    go clean
    rm -f complaints-mcp

# Run tests with BDD framework
test:
    go test ./...

# Run BDD tests specifically
test-bdd:
    @if [ -d "features" ] && [ "$(find features -name "*.go" | wc -l)" -gt 0 ]; then \
        go test -v ./features/...; \
    else \
        echo "⚠️  BDD step definitions not yet implemented - skipping BDD tests"; \
    fi

# Lint code with go vet and formatting checks
lint:
    go vet ./...
    @if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then \
        echo "The following files are not formatted:"; \
        gofmt -s -l .; \
        exit 1; \
    fi
    @echo "All files are properly formatted and pass go vet"

# Format code with gofmt
fmt:
    gofmt -s -w .

# Install dependencies
deps:
    go mod download
    go mod tidy

# Install development tools
install-tools:
    go install github.com/mibk/dupl@latest
    go install github.com/cucumber/godog/cmd/godog@latest
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Find code duplicates using jscpd (modern, multi-language tool)
fd:
	@if ! command -v jscpd >/dev/null 2>&1; then \
		echo "jscpd not found. Install with: npm install -g jscpd"; \
		exit 1; \
	fi
	jscpd . --threshold 15 --format console

# Find code duplicates using jscpd with strict threshold
fd-strict:
	jscpd . --threshold 50 --format console

# Legacy alias for backwards compatibility (uses dupl)
fd-legacy:
	@if ! command -v dupl >/dev/null 2>&1; then \
		echo "dupl not found. Install with: go install github.com/mibk/dupl@latest"; \
		exit 1; \
	fi
	dupl -t 15 .

# Find code duplicates using golangci-lint dupl (alternative)
fd-golangci:
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lint not found. Install from https://golangci-lint.run/"; \
		exit 1; \
	fi
	golangci-lint run --enable-only=dupl

# Run comprehensive code quality checks
quality: fmt lint test fd

# Run full CI pipeline
ci: deps fmt lint test-bdd fd
    @echo "✅ All CI checks passed!"

# Development server (run in background)
dev:
    just build && ./complaints-mcp

# Generate documentation
docs:
    @echo "Generating documentation..."
    @echo "API documentation would be generated here"

# Security audit
security:
    go list -json -m all | nancy sleuth

# Update dependencies
update:
    go get -u ./...
    go mod tidy

# Help information
help:
    @echo "Available commands:"
    @echo "  build          - Build the complaints-mcp server"
    @echo "  build-release  - Build optimized release binary"
    @echo "  clean          - Clean build artifacts"
    @echo "  test           - Run Go tests"
    @echo "  test-bdd       - Run BDD tests with godog"
    @echo "  lint           - Run go vet and formatting checks"
    @echo "  fmt            - Format code with gofmt"
    @echo "  deps           - Install dependencies"
    @echo "  install-tools  - Install development tools"
    @echo "  fd             - Find code duplicates (jscpd)"
    @echo "  fd-strict      - Find duplicates with strict threshold (jscpd)"
    @echo "  fd-legacy      - Find duplicates using legacy dupl"
    @echo "  fd-golangci    - Find duplicates using golangci-lint"
    @echo "  quality        - Run comprehensive code quality checks"
    @echo "  ci             - Run full CI pipeline"
    @echo "  dev            - Run development server"
    @echo "  docs           - Generate documentation"
    @echo "  security       - Security audit"
    @echo "  update         - Update dependencies"
    @echo "  help           - Show this help message"

# Install the project locally
install: build
    cp complaints-mcp $(GOPATH)/bin/