#!/bin/bash

# Home Automation System Uninstaller Script
# This script removes the home automation system from autostart

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

# Function to stop and disable service
stop_service() {
    print_status "Stopping and disabling home automation service"
    
    if systemctl is-active --quiet "$SERVICE_NAME.service"; then
        systemctl stop "$SERVICE_NAME.service"
        print_success "Service stopped"
    else
        print_warning "Service is not running"
    fi
    
    if systemctl is-enabled --quiet "$SERVICE_NAME.service"; then
        systemctl disable "$SERVICE_NAME.service"
        print_success "Service disabled from autostart"
    else
        print_warning "Service is not enabled"
    fi
}

# Function to remove service file
remove_service() {
    print_status "Removing systemd service file"
    
    if [[ -f "/etc/systemd/system/$SERVICE_NAME.service" ]]; then
        rm "/etc/systemd/system/$SERVICE_NAME.service"
        systemctl daemon-reload
        print_success "Service file removed"
    else
        print_warning "Service file not found"
    fi
}

# Function to remove installation directory
remove_install_dir() {
    echo ""
    read -p "Remove installation directory $INSTALL_DIR? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        if [[ -d "$INSTALL_DIR" ]]; then
            rm -rf "$INSTALL_DIR"
            print_success "Installation directory removed"
        else
            print_warning "Installation directory not found"
        fi
    else
        print_warning "Installation directory preserved at $INSTALL_DIR"
    fi
}

# Main uninstallation process
main() {
    echo "=========================================="
    echo "🏠 Home Automation System Uninstaller"
    echo "=========================================="
    echo ""
    
    check_root
    stop_service
    remove_service
    remove_install_dir
    
    echo ""
    print_success "🎉 Home Automation System removed from autostart!"
    print_status "To manually start the system, use docker-compose in your project directory"
    echo ""
}

# Run main function
main "$@"
