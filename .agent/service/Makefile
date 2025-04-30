# Basic Makefile for Agent Guidance Service

.PHONY: all build server cli test fmt lint clean

BINARY_NAME=agentt

all: build

build: server cli

server: fmt lint
	@echo "Building server (will become part of agentt)..."
	go build -o $(BINARY_NAME)-server-tmp ./cmd/server # Temporary name
	@echo "Server binary (temp): $(BINARY_NAME)-server-tmp"

cli: fmt lint
	@echo "Building cli (will become part of agentt)..."
	go build -o $(BINARY_NAME)-cli-tmp ./cmd/cli # Temporary name
	@echo "CLI binary (temp): $(BINARY_NAME)-cli-tmp"

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
	rm -f $(BINARY_NAME) $(BINARY_NAME)-server-tmp $(BINARY_NAME)-cli-tmp # Update clean target
	go clean