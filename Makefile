# Makefile for complaints-mcp

.PHONY: build
.DEFAULT_GOALS := all

# Variables
BINARY_NAME := complaints-mcp
GO_FILES := $(shell find . -name "*.go" | tr '\n' ' ')
GO_FLAGS := -ldflags="-s -w"

# Build commands
build:
	go build $(GO_FLAGS) -o $(BINARY_NAME) $(GO_FILES)

# Development
dev:
	go build $(GO_FLAGS) -o $(BINARY_NAME) $(GO_FILES)

# Testing
test:
	go test ./...

# Linting
lint:
	go vet ./...

# Clean
clean:
	rm -f $(BINARY_NAME)

# Install
install: $(BINARY_NAME)
	cp $(BINARY_NAME) $$GOPATH/bin/

# Run
run: $(BINARY_NAME)
	./$(BINARY_NAME)

# Development server
serve: dev
	./$(BINARY_NAME) --dev

.PHONY: all test lint clean install