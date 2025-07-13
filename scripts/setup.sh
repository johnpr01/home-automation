#!/bin/bash

# Home Automation Development Setup Script

set -e

echo "🏠 Setting up Home Automation development environment..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

echo "✅ Go is installed: $(go version)"

# Initialize Go module if not already done
if [ ! -f go.mod ]; then
    echo "📦 Initializing Go module..."
    go mod init github.com/johnpr01/home-automation
fi

# Download dependencies
echo "📥 Downloading dependencies..."
go mod tidy

# Create necessary directories
echo "📁 Creating directories..."
mkdir -p bin logs tmp

# Check if make is available
if command -v make &> /dev/null; then
    echo "🔨 Building project..."
    make build
else
    echo "🔨 Building project manually..."
    go build -o bin/home-automation-server ./cmd/server
    go build -o bin/home-automation-cli ./cmd/cli
fi

# Make binaries executable
chmod +x bin/*

echo "✅ Development environment setup complete!"
echo ""
echo "🚀 Quick start commands:"
echo "  make run-server    # Start the server"
echo "  make run-cli       # Run the CLI"
echo "  make test          # Run tests"
echo "  make help          # Show all available commands"
echo ""
echo "🌐 Server will be available at: http://localhost:8080"
