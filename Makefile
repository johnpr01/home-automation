.PHONY: build run test clean install server cli

# Build variables
BINARY_NAME=home-automation
SERVER_BINARY=home-automation-server
CLI_BINARY=home-automation-cli
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build all binaries
build: server cli

# Build server binary
server:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(SERVER_BINARY) ./cmd/server

# Build CLI binary
cli:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(CLI_BINARY) ./cmd/cli

# Run the server
run-server: server
	./$(BUILD_DIR)/$(SERVER_BINARY)

# Run the CLI
run-cli: cli
	./$(BUILD_DIR)/$(CLI_BINARY)

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Install development tools
install-tools:
	$(GOGET) -u golang.org/x/tools/cmd/goimports
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Format code
fmt:
	$(GOCMD) fmt ./...
	goimports -w .

# Lint code
lint:
	golangci-lint run

# Run in development mode with hot reload (requires air)
dev:
	air

# Docker build
docker-build:
	docker build -t $(BINARY_NAME) .

# Docker run
docker-run:
	docker run -p 8080:8080 $(BINARY_NAME)

# Install air for hot reload
install-air:
	$(GOGET) -u github.com/cosmtrek/air@latest

# Help
help:
	@echo "Available commands:"
	@echo "  build         - Build all binaries"
	@echo "  server        - Build server binary"
	@echo "  cli           - Build CLI binary"
	@echo "  run-server    - Build and run server"
	@echo "  run-cli       - Build and run CLI"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Download dependencies"
	@echo "  install-tools - Install development tools"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  dev           - Run in development mode"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  help          - Show this help"
