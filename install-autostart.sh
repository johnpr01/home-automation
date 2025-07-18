#!/bin/bash

# Home Automation System Installation Script for Raspberry Pi
# This script sets up the home automation system to start automatically on boot

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
INSTALL_DIR="/opt/home-automation"
SERVICE_NAME="home-automation"
USER="pi"
GROUP="pi"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        print_error "This script must be run as root (use sudo)"
        exit 1
    fi
}

# Function to check if Docker is installed
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first:"
        echo "curl -fsSL https://get.docker.com -o get-docker.sh"
        echo "sudo sh get-docker.sh"
        echo "sudo usermod -aG docker $USER"
        exit 1
    fi
    
    if ! command -v docker compose &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    print_success "Docker and Docker Compose are installed"
}

# Function to create installation directory
create_install_dir() {
    print_status "Creating installation directory: $INSTALL_DIR"
    
    if [[ -d "$INSTALL_DIR" ]]; then
        print_warning "Directory $INSTALL_DIR already exists"
    else
        mkdir -p "$INSTALL_DIR"
        print_success "Created directory $INSTALL_DIR"
    fi
    
    # Set ownership
    chown -R $USER:$GROUP "$INSTALL_DIR"
    chmod 755 "$INSTALL_DIR"
}

# Function to copy application files
copy_files() {
    print_status "Copying application files to $INSTALL_DIR"
    
    # Copy main files
    cp docker-compose.yml "$INSTALL_DIR/"
    cp prometheus.yml "$INSTALL_DIR/"
    
    # Copy .env.example as .env if .env doesn't exist
    if [[ ! -f "$INSTALL_DIR/.env" ]]; then
        cp .env.example "$INSTALL_DIR/.env"
        print_success "Created .env file from .env.example"
        print_warning "Please edit $INSTALL_DIR/.env with your configuration"
    else
        print_warning ".env file already exists, skipping"
    fi
    
    # Copy directories
    if [[ -d "deployments" ]]; then
        cp -r deployments "$INSTALL_DIR/"
    fi
    
    if [[ -d "configs" ]]; then
        cp -r configs "$INSTALL_DIR/"
    fi
    
    # Set ownership
    chown -R $USER:$GROUP "$INSTALL_DIR"
    
    print_success "Application files copied"
}

# Function to install systemd service
install_service() {
    print_status "Installing systemd service"
    
    # Copy service file
    cp home-automation.service "/etc/systemd/system/$SERVICE_NAME.service"
    
    # Reload systemd
    systemctl daemon-reload
    
    # Enable service
    systemctl enable "$SERVICE_NAME.service"
    
    print_success "Systemd service installed and enabled"
}

# Function to start service
start_service() {
    print_status "Starting home automation service"
    
    systemctl start "$SERVICE_NAME.service"
    
    # Check status
    if systemctl is-active --quiet "$SERVICE_NAME.service"; then
        print_success "Home automation service is running"
    else
        print_error "Failed to start home automation service"
        systemctl status "$SERVICE_NAME.service"
        exit 1
    fi
}

# Function to show status
show_status() {
    echo ""
    print_status "Service Status:"
    systemctl status "$SERVICE_NAME.service" --no-pager
    
    echo ""
    print_status "Useful Commands:"
    echo "  View logs:           sudo journalctl -u $SERVICE_NAME.service -f"
    echo "  Stop service:        sudo systemctl stop $SERVICE_NAME.service"
    echo "  Start service:       sudo systemctl start $SERVICE_NAME.service"
    echo "  Restart service:     sudo systemctl restart $SERVICE_NAME.service"
    echo "  Disable autostart:   sudo systemctl disable $SERVICE_NAME.service"
    echo "  Check status:        sudo systemctl status $SERVICE_NAME.service"
}

# Function to show web interfaces
show_interfaces() {
    echo ""
    print_status "Web Interfaces (after startup completes):"
    echo "  📊 Grafana:     http://$(hostname -I | awk '{print $1}'):3000 (admin/admin)"
    echo "  📈 Prometheus:  http://$(hostname -I | awk '{print $1}'):9090"
    echo "  🔌 Tapo Metrics: http://$(hostname -I | awk '{print $1}'):2112/metrics"
    echo "  🏠 Home API:    http://$(hostname -I | awk '{print $1}'):8080/api/status"
}

# Main installation process
main() {
    echo "========================================"
    echo "🏠 Home Automation System Installer"
    echo "========================================"
    echo ""
    
    check_root
    check_docker
    create_install_dir
    copy_files
    install_service
    
    # Ask if user wants to start the service now
    echo ""
    read -p "Start the home automation service now? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        start_service
        show_status
        show_interfaces
    else
        print_success "Installation complete. Start the service with:"
        echo "  sudo systemctl start $SERVICE_NAME.service"
    fi
    
    echo ""
    print_success "🎉 Home Automation System installation complete!"
    print_warning "The system will now start automatically on boot."
    echo ""
}

# Run main function
main "$@"
