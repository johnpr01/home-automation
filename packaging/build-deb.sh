#!/bin/bash
# Build script for Home Automation Debian packages

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
PACKAGING_DIR="$SCRIPT_DIR/deb"
OUTPUT_DIR="$SCRIPT_DIR/output"

echo "🏗️  Building Home Automation Debian Package"
echo "==========================================="
echo "Project root: $PROJECT_ROOT"
echo "Packaging dir: $PACKAGING_DIR"
echo "Output dir: $OUTPUT_DIR"
echo

# Clean previous builds
if [ -d "$OUTPUT_DIR" ]; then
    rm -rf "$OUTPUT_DIR"
fi
mkdir -p "$OUTPUT_DIR"

# Get version from git or use default
if command -v git >/dev/null 2>&1 && [ -d "$PROJECT_ROOT/.git" ]; then
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "1.0.0")
    # Clean version for debian (remove 'v' prefix if present)
    VERSION=${VERSION#v}
    # Ensure version starts with a digit for Debian compatibility
    if [[ ! $VERSION =~ ^[0-9] ]]; then
        VERSION="1.0.0-$VERSION"
    fi
else
    VERSION="1.0.0"
fi

echo "📦 Package version: $VERSION"

# Update version in control file
sed -i "s/^Version: .*/Version: $VERSION/" "$PACKAGING_DIR/home-automation/DEBIAN/control"

echo "📋 Copying project files..."

# Copy main project files
cp -r "$PROJECT_ROOT/cmd" "$PACKAGING_DIR/home-automation/opt/home-automation/"
cp -r "$PROJECT_ROOT/internal" "$PACKAGING_DIR/home-automation/opt/home-automation/"
cp -r "$PROJECT_ROOT/pkg" "$PACKAGING_DIR/home-automation/opt/home-automation/"
cp -r "$PROJECT_ROOT/deployments" "$PACKAGING_DIR/home-automation/opt/home-automation/"
cp -r "$PROJECT_ROOT/configs" "$PACKAGING_DIR/home-automation/opt/home-automation/"
cp -r "$PROJECT_ROOT/firmware" "$PACKAGING_DIR/home-automation/opt/home-automation/"

# Copy documentation
cp "$PROJECT_ROOT/README.md" "$PACKAGING_DIR/home-automation/opt/home-automation/"
cp "$PROJECT_ROOT/go.mod" "$PACKAGING_DIR/home-automation/opt/home-automation/"
cp "$PROJECT_ROOT/go.sum" "$PACKAGING_DIR/home-automation/opt/home-automation/"

# Copy environment template
cp "$PROJECT_ROOT/.env.example" "$PACKAGING_DIR/home-automation/opt/home-automation/"

# Copy systemd service file
cp "$PROJECT_ROOT/init/systemd/home-automation.service" "$PACKAGING_DIR/home-automation/etc/systemd/system/"

# Copy autostart documentation
if [ -f "$PROJECT_ROOT/AUTOSTART_SETUP.md" ]; then
    cp "$PROJECT_ROOT/AUTOSTART_SETUP.md" "$PACKAGING_DIR/home-automation/opt/home-automation/"
fi

if [ -f "$PROJECT_ROOT/AUTOSTART_QUICK_REFERENCE.md" ]; then
    cp "$PROJECT_ROOT/AUTOSTART_QUICK_REFERENCE.md" "$PACKAGING_DIR/home-automation/opt/home-automation/"
fi

echo "� Generating SBOM..."

# Generate SBOM if script exists
if [ -f "$SCRIPT_DIR/generate-sbom.sh" ]; then
    cd "$SCRIPT_DIR"
    ./generate-sbom.sh
    
    # Copy SBOM files to package
    if [ -d "$SCRIPT_DIR/sbom" ]; then
        mkdir -p "$PACKAGING_DIR/home-automation/opt/home-automation/sbom"
        cp -r "$SCRIPT_DIR/sbom/"* "$PACKAGING_DIR/home-automation/opt/home-automation/sbom/"
        echo "✅ SBOM files included in package"
    fi
    
    cd "$PACKAGING_DIR"
else
    echo "⚠️  SBOM generation script not found, skipping SBOM inclusion"
fi

echo "�🔧 Setting file permissions..."

# Make scripts executable
chmod +x "$PACKAGING_DIR/home-automation/DEBIAN/preinst"
chmod +x "$PACKAGING_DIR/home-automation/DEBIAN/postinst"
chmod +x "$PACKAGING_DIR/home-automation/DEBIAN/prerm"
chmod +x "$PACKAGING_DIR/home-automation/DEBIAN/postrm"
chmod +x "$PACKAGING_DIR/home-automation/usr/local/bin/"*

# Set proper ownership (will be corrected during installation)
chmod 644 "$PACKAGING_DIR/home-automation/opt/home-automation/.env.example"
chmod 644 "$PACKAGING_DIR/home-automation/etc/systemd/system/home-automation.service"

# Set SBOM file permissions if they exist
if [ -d "$PACKAGING_DIR/home-automation/opt/home-automation/sbom" ]; then
    chmod 644 "$PACKAGING_DIR/home-automation/opt/home-automation/sbom/"*
    chmod +x "$PACKAGING_DIR/home-automation/opt/home-automation/sbom/scan-vulnerabilities.sh"
fi

echo "📦 Building package..."

# Build the main package
PACKAGE_NAME="home-automation_${VERSION}_arm64.deb"
cd "$PACKAGING_DIR"

if dpkg-deb --build home-automation "$OUTPUT_DIR/$PACKAGE_NAME"; then
    echo "✅ Main package built successfully: $PACKAGE_NAME"
else
    echo "❌ Package build failed!"
    exit 1
fi

# Create a standalone package (without Docker dependencies)
echo "📦 Building standalone package..."

# Copy main package structure for standalone version
cp -r home-automation home-automation-standalone

# Update control file for standalone
cat > home-automation-standalone/DEBIAN/control << EOF
Package: home-automation-standalone
Version: $VERSION
Section: utils
Priority: optional
Architecture: arm64
Essential: no
Depends: golang-go (>= 1.19), curl, wget, systemd
Recommends: mosquitto, prometheus, grafana
Suggests: docker.io
Installed-Size: 30000
Maintainer: Home Automation Team <admin@home-automation.local>
Description: Smart Home Automation System (Standalone)
 A lightweight home automation system that runs without Docker:
 .
 * Direct Go binary execution
 * MQTT message processing
 * TP-Link Tapo smart plug monitoring with KLAP protocol support
 * Prometheus metrics export
 * Multi-sensor support (temperature, humidity, motion, light)
 * Raspberry Pi Pico WH firmware for wireless sensors
 * Energy monitoring and data collection
 .
 This standalone package runs the Go binary directly without
 Docker containers, providing a lighter footprint for
 resource-constrained environments.
Homepage: https://github.com/johnpr01/home-automation
EOF

# Update systemd service for standalone
cat > home-automation-standalone/etc/systemd/system/home-automation.service << 'EOF'
[Unit]
Description=Home Automation System (Standalone)
Documentation=https://github.com/johnpr01/home-automation
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=pi
Group=pi
WorkingDirectory=/opt/home-automation
Environment=PATH=/usr/local/bin:/usr/bin:/bin
EnvironmentFile=-/opt/home-automation/.env
ExecStartPre=/bin/sleep 10
ExecStart=/opt/home-automation/bin/home-automation-server
ExecReload=/bin/kill -HUP $MAINPID
Restart=always
RestartSec=10
KillMode=mixed
TimeoutStopSec=30

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/home-automation /var/log/home-automation
CapabilityBoundingSet=CAP_NET_BIND_SERVICE
AmbientCapabilities=CAP_NET_BIND_SERVICE

# Resource limits
LimitNOFILE=65536
MemoryMax=512M
CPUQuota=50%

[Install]
WantedBy=multi-user.target
EOF

# Update postinst for standalone
cat > home-automation-standalone/DEBIAN/postinst << 'EOF'
#!/bin/bash
set -e

echo "🚀 Configuring home-automation standalone system..."

# Set ownership and permissions
echo "🔐 Setting file permissions..."
chown -R pi:pi /opt/home-automation
chmod 755 /opt/home-automation
chmod 644 /opt/home-automation/.env.example

# Build the Go binary
echo "🔨 Building Go application..."
cd /opt/home-automation

if command -v go >/dev/null 2>&1; then
    # Create bin directory
    mkdir -p bin
    chown pi:pi bin
    
    # Build main server
    sudo -u pi go build -o bin/home-automation-server ./cmd/server
    
    # Build utilities
    sudo -u pi go build -o bin/tapo-demo ./cmd/tapo-demo
    sudo -u pi go build -o bin/test-klap ./cmd/test-klap
    
    chmod +x bin/*
    echo "✅ Go binaries built successfully"
else
    echo "❌ Go compiler not found!"
    echo "   Install with: sudo apt install golang-go"
    exit 1
fi

# Create log directory
mkdir -p /var/log/home-automation
chown pi:pi /var/log/home-automation
chmod 755 /var/log/home-automation

# Reload systemd
echo "🔄 Reloading systemd configuration..."
systemctl daemon-reload
systemctl enable home-automation.service

echo ""
echo "🎉 Home Automation Standalone System installed successfully!"
echo ""
echo "📋 Next Steps:"
echo "1. Configure environment: sudo -u pi nano /opt/home-automation/.env"
echo "2. Start service: sudo systemctl start home-automation"
echo "3. Check status: sudo systemctl status home-automation"
echo ""
echo "📊 Management Commands:"
echo "   home-automation-status   - Check system status"
echo "   home-automation-logs     - View service logs"  
echo "   home-automation-restart  - Restart service"
echo "   home-automation-update   - Update configuration"
echo ""
echo "📖 Documentation: /opt/home-automation/README.md"

exit 0
EOF

chmod +x home-automation-standalone/DEBIAN/postinst

# Build standalone package
STANDALONE_PACKAGE_NAME="home-automation-standalone_${VERSION}_arm64.deb"
if dpkg-deb --build home-automation-standalone "$OUTPUT_DIR/$STANDALONE_PACKAGE_NAME"; then
    echo "✅ Standalone package built successfully: $STANDALONE_PACKAGE_NAME"
else
    echo "❌ Standalone package build failed!"
    exit 1
fi

cd "$SCRIPT_DIR"

echo
echo "📦 Package build summary:"
echo "========================"
echo "Main package: $OUTPUT_DIR/$PACKAGE_NAME"
echo "Standalone:   $OUTPUT_DIR/$STANDALONE_PACKAGE_NAME"
echo

# Show package information
echo "📋 Package information:"
echo "======================"
dpkg-deb --info "$OUTPUT_DIR/$PACKAGE_NAME"
echo
echo "📁 Package contents:"
echo "=================="
dpkg-deb --contents "$OUTPUT_DIR/$PACKAGE_NAME" | head -20
echo "... (truncated)"

echo
echo "✅ Build completed successfully!"
echo
echo "📥 Installation commands:"
echo "   sudo dpkg -i $OUTPUT_DIR/$PACKAGE_NAME"
echo "   sudo apt-get install -f  # Fix dependencies if needed"
echo
echo "📥 Standalone installation:"
echo "   sudo dpkg -i $OUTPUT_DIR/$STANDALONE_PACKAGE_NAME" 
echo "   sudo apt-get install -f"
echo
echo "🚀 After installation:"
echo "   home-automation-update create    # Create initial config"
echo "   sudo systemctl start home-automation"
echo "   home-automation-status          # Check status"
