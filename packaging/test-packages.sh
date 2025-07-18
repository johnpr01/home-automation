#!/bin/bash
# Test installation script for Home Automation Debian packages

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PACKAGES_DIR="$SCRIPT_DIR/output"

echo "🧪 Home Automation Package Installation Test"
echo "============================================"
echo

# Function to test package info
test_package_info() {
    local package_file="$1"
    local package_name="$2"
    
    echo "📦 Testing $package_name package..."
    echo "Package file: $(basename "$package_file")"
    echo "Size: $(du -h "$package_file" | cut -f1)"
    echo
    
    echo "📋 Package information:"
    dpkg-deb --info "$package_file" | grep -E "(Package|Version|Architecture|Depends):"
    echo
    
    echo "🔍 Package integrity check:"
    if dpkg-deb --contents "$package_file" >/dev/null 2>&1; then
        echo "✅ Package structure is valid"
    else
        echo "❌ Package structure is invalid"
        return 1
    fi
    
    local file_count=$(dpkg-deb --contents "$package_file" | wc -l)
    echo "📁 Contains $file_count files/directories"
    echo
}

# Function to show installation preview
show_installation_preview() {
    local package_file="$1"
    local package_name="$2"
    
    echo "📥 Installation preview for $package_name:"
    echo "=========================================="
    echo
    echo "🔧 Installation command:"
    echo "sudo dpkg -i $(basename "$package_file")"
    echo "sudo apt-get install -f  # Fix dependencies if needed"
    echo
    echo "📋 Post-installation steps:"
    echo "1. home-automation-update create    # Create configuration"
    echo "2. home-automation-update edit      # Edit settings"
    echo "3. sudo systemctl start home-automation"
    echo "4. home-automation-status           # Check status"
    echo
    echo "🌐 Web interfaces (after starting):"
    echo "- Grafana: http://$(hostname -I | awk '{print $1}'):3000"
    echo "- Prometheus: http://$(hostname -I | awk '{print $1}'):9090"
    echo
}

# Function to show key files that will be installed
show_key_files() {
    local package_file="$1"
    local package_name="$2"
    
    echo "🗂️  Key files in $package_name package:"
    echo "====================================="
    
    echo "📄 Configuration:"
    dpkg-deb --contents "$package_file" | grep -E "\.(env|yml|yaml|conf)$" | head -5
    
    echo "🔧 Management scripts:"
    dpkg-deb --contents "$package_file" | grep "usr/local/bin" | head -5
    
    echo "⚙️  System integration:"
    dpkg-deb --contents "$package_file" | grep -E "(systemd|init)" | head -3
    
    echo "📖 Documentation:"
    dpkg-deb --contents "$package_file" | grep -E "\.(md|txt|doc)$" | head -3
    echo
}

# Check if packages exist
MAIN_PACKAGE=$(find "$PACKAGES_DIR" -name "home-automation_*.deb" | head -1)
STANDALONE_PACKAGE=$(find "$PACKAGES_DIR" -name "home-automation-standalone_*.deb" | head -1)

if [ ! -f "$MAIN_PACKAGE" ] || [ ! -f "$STANDALONE_PACKAGE" ]; then
    echo "❌ Package files not found in $PACKAGES_DIR"
    echo "Please run ./build-deb.sh first to build the packages"
    exit 1
fi

echo "Found packages:"
echo "- Main: $(basename "$MAIN_PACKAGE")"
echo "- Standalone: $(basename "$STANDALONE_PACKAGE")"
echo

# Test main package
test_package_info "$MAIN_PACKAGE" "Main (Docker-based)"
show_key_files "$MAIN_PACKAGE" "Main"
show_installation_preview "$MAIN_PACKAGE" "Main"

echo "───────────────────────────────────────────────────────────────"
echo

# Test standalone package
test_package_info "$STANDALONE_PACKAGE" "Standalone (Go binary)"
show_key_files "$STANDALONE_PACKAGE" "Standalone"
show_installation_preview "$STANDALONE_PACKAGE" "Standalone"

echo "🎯 Recommendation:"
echo "=================="
echo

echo "For Raspberry Pi 5 with 4GB+ RAM:"
echo "→ Use the main package (full Docker stack)"
echo "  sudo dpkg -i $(basename "$MAIN_PACKAGE")"
echo

echo "For Raspberry Pi with limited resources:"
echo "→ Use the standalone package (lightweight)"
echo "  sudo dpkg -i $(basename "$STANDALONE_PACKAGE")"
echo

echo "🔧 System Requirements:"
echo "======================"
echo
echo "Main package:"
echo "- Docker & Docker Compose"
echo "- 2GB+ RAM recommended"
echo "- 4GB+ disk space"
echo "- systemd-based Linux"
echo
echo "Standalone package:"
echo "- Go compiler (installed automatically)"
echo "- 512MB+ RAM"
echo "- 1GB+ disk space" 
echo "- systemd-based Linux"
echo

echo "📚 Next Steps:"
echo "============="
echo "1. Copy the appropriate .deb file to your Raspberry Pi"
echo "2. Install with: sudo dpkg -i <package-file>"
echo "3. Fix dependencies: sudo apt-get install -f"
echo "4. Configure: home-automation-update create"
echo "5. Start: sudo systemctl start home-automation"
echo "6. Monitor: home-automation-status"
echo

echo "✅ Package testing completed successfully!"
