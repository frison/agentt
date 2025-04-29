# Basic Makefile for Agent Guidance Service

.PHONY: all build server cli test fmt lint clean

BINARY_SERVER=agent-guidance-service
BINARY_CLI=agent-guidance-cli

all: build

build: server cli

server: fmt lint
	@echo "Building server..."
	go build -o $(BINARY_SERVER) ./cmd/server
	@echo "Server binary: $(BINARY_SERVER)"

cli: fmt lint
	@echo "Building cli..."
	go build -o $(BINARY_CLI) ./cmd/cli
	@echo "CLI binary: $(BINARY_CLI)"

test:
	@echo "Running tests..."
	go test ./...

fmt:
	@echo "Formatting code..."
	go fmt ./...

lint:
	@echo "Linting code... (requires golangci-lint)"
	@golangci-lint run || echo "Warning: golangci-lint not found or lint errors detected."

clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_SERVER) $(BINARY_CLI)
	go clean