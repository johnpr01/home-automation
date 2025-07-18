# Home Automation System - Debian Packages

This directory contains Debian packaging configuration for the Home Automation System, providing easy installation and management on Raspberry Pi 5 and other ARM64 Debian-based systems.

## 📦 Available Packages

### 1. `home-automation` (Main Package)
**Full-featured Docker-based deployment**

- **Dependencies**: Docker, Docker Compose, systemd
- **Components**: Complete stack with Prometheus, Grafana, MQTT, Kafka
- **Resource Usage**: ~1-2GB RAM, requires Docker
- **Best For**: Production deployments with full monitoring

### 2. `home-automation-standalone` (Lightweight Package)
**Standalone Go binary deployment**

- **Dependencies**: Go compiler, systemd (no Docker)
- **Components**: Core automation server, Tapo monitoring, basic metrics
- **Resource Usage**: ~200-500MB RAM, no containers
- **Best For**: Resource-constrained environments, development

## 🏗️ Building Packages

### Prerequisites

```bash
# Install build tools
sudo apt update
sudo apt install dpkg-dev build-essential debhelper

# For cross-compilation (if building on x86_64 for ARM)
sudo apt install gcc-aarch64-linux-gnu
```

### Build Process

```bash
# Navigate to packaging directory
cd packaging

# Build both packages
./build-deb.sh

# Packages will be created in packaging/output/
ls -la output/
# home-automation_1.0.0_arm64.deb
# home-automation-standalone_1.0.0_arm64.deb
```

### Build Script Features

- **Automatic versioning** from git tags
- **File organization** and permission setting
- **Dual package creation** (full and standalone)
- **Dependency validation**
- **Package integrity checks**

## 📥 Installation

### Method 1: Direct Installation

```bash
# Download and install main package
sudo dpkg -i home-automation_1.0.0_arm64.deb

# Install dependencies if needed
sudo apt-get install -f

# Or install standalone version
sudo dpkg -i home-automation-standalone_1.0.0_arm64.deb
sudo apt-get install -f
```

### Method 2: Repository Installation (Future)

```bash
# Add repository (when available)
echo "deb [trusted=yes] https://packages.home-automation.local/debian bullseye main" | sudo tee /etc/apt/sources.list.d/home-automation.list

# Install via apt
sudo apt update
sudo apt install home-automation
```

## ⚙️ Post-Installation Configuration

### 1. Initial Setup

```bash
# Create configuration from template
home-automation-update create

# Edit configuration with your settings
home-automation-update edit

# Validate configuration
home-automation-update validate
```

### 2. Service Management

```bash
# Start the service
sudo systemctl start home-automation

# Enable auto-start
sudo systemctl enable home-automation

# Check status
home-automation-status

# View logs
home-automation-logs -f
```

### 3. Network Configuration

```bash
# Configure your settings in .env file
sudo -u pi nano /opt/home-automation/.env

# Required settings:
TPLINK_USERNAME=your@email.com
TPLINK_PASSWORD=yourpassword
TAPO_DEVICE_1_IP=192.168.1.100
TAPO_DEVICE_2_IP=192.168.1.101
# ... etc
```

## 🔧 Management Commands

The packages install several management utilities:

### `home-automation-status`
Shows comprehensive system status including:
- Service state (running/stopped)
- Docker container health
- Web interface accessibility
- System resource usage
- Configuration validation

### `home-automation-logs`
Advanced log viewer with options:
```bash
home-automation-logs              # Show recent logs
home-automation-logs -f           # Follow logs in real-time
home-automation-logs -n 100       # Show more lines
home-automation-logs --help       # Show all options
```

### `home-automation-restart`
Service restart utility:
```bash
home-automation-restart           # Full service restart
home-automation-restart containers # Docker containers only
home-automation-restart config    # Reload configuration
```

### `home-automation-update`
Configuration management:
```bash
home-automation-update            # Edit configuration
home-automation-update create     # Create initial config
home-automation-update validate   # Check configuration
home-automation-update test       # Test device connectivity
home-automation-update show       # Display current config
```

### `upload-pico-firmware`
Raspberry Pi Pico firmware uploader:
```bash
upload-pico-firmware              # Upload sensor firmware to Pico
```

## 📁 Package Contents

### File Layout

```
/opt/home-automation/              # Main installation directory
├── cmd/                          # Go source code
├── internal/                     # Internal packages
├── pkg/                          # Public packages
├── deployments/                  # Docker Compose configs
│   └── docker-compose.yml        # Main orchestration file
├── configs/                      # Configuration templates
├── firmware/                     # Raspberry Pi Pico firmware
│   └── pico-sht30/              # Multi-sensor firmware
├── .env.example                  # Environment template
├── README.md                     # Documentation
└── data/                         # Runtime data (created on first run)
    ├── prometheus/               # Prometheus data
    ├── grafana/                  # Grafana data
    └── kafka-logs/               # Kafka logs

/etc/systemd/system/
└── home-automation.service       # systemd service file

/usr/local/bin/                   # Management utilities
├── home-automation-status
├── home-automation-logs
├── home-automation-restart
├── home-automation-update
└── upload-pico-firmware

/var/log/home-automation/         # Log directory
```

### Configuration Files

- **Main config**: `/opt/home-automation/.env`
- **Service file**: `/etc/systemd/system/home-automation.service`
- **Docker Compose**: `/opt/home-automation/deployments/docker-compose.yml`
- **Logs**: `/var/log/home-automation/`

## 🔒 Security Features

### File Permissions
- All files owned by `pi:pi` user
- Configuration files have restricted permissions (640)
- Service runs as non-root user
- systemd security restrictions enabled

### systemd Security
```ini
NoNewPrivileges=true              # Prevent privilege escalation
PrivateTmp=true                   # Isolated temp directories
ProtectSystem=strict              # Read-only system directories
ProtectHome=true                  # No access to user homes
ReadWritePaths=/opt/home-automation # Limited write access
```

### Network Security
- Services bound to specific interfaces
- Docker containers isolated
- No unnecessary port exposure
- Firewall-friendly configuration

## 🔄 Package Lifecycle

### Installation Process
1. **preinst**: System checks, user creation, Docker group setup
2. **Package extraction**: Files copied to target locations
3. **postinst**: Permissions, service setup, initial configuration
4. **Service registration**: systemd integration complete

### Upgrade Process
1. **preinst**: Backup existing config, prepare for upgrade
2. **File replacement**: New files replace old versions
3. **postinst**: Update permissions, restart services
4. **Configuration merge**: Preserve user settings

### Removal Process
1. **prerm**: Stop services, disable auto-start
2. **File removal**: Package files deleted
3. **postrm**: 
   - `remove`: Keep configuration
   - `purge`: Complete removal including config

## 🧪 Testing Packages

### Validation Tests

```bash
# Test package integrity
dpkg-deb --contents home-automation_1.0.0_arm64.deb
dpkg-deb --info home-automation_1.0.0_arm64.deb

# Test installation in clean environment
docker run --rm -it --privileged \
  -v $(pwd)/output:/packages \
  debian:bullseye-slim bash

# Inside container:
apt update && apt install -y systemd
dpkg -i /packages/home-automation_1.0.0_arm64.deb
apt-get install -f
```

### Package Scripts Testing

```bash
# Test individual scripts
sudo bash -x packaging/deb/home-automation/DEBIAN/preinst install
sudo bash -x packaging/deb/home-automation/DEBIAN/postinst configure
```

## 🐛 Troubleshooting

### Common Installation Issues

#### 1. Dependency Problems
```bash
# Error: package depends on docker.io but it is not installable
sudo apt update
sudo apt install -y docker.io docker-compose-plugin

# Then retry installation
sudo dpkg -i home-automation_*.deb
```

#### 2. Permission Issues
```bash
# Error: permission denied
sudo chown -R pi:pi /opt/home-automation
sudo chmod 755 /opt/home-automation
```

#### 3. Service Start Failures
```bash
# Check detailed error messages
sudo systemctl status home-automation -l
sudo journalctl -u home-automation -n 50

# Common fixes
home-automation-update validate
home-automation-restart
```

#### 4. Docker Issues
```bash
# Docker not starting containers
sudo systemctl restart docker
sudo usermod -aG docker pi
# Logout and login again
```

### Package Build Issues

#### Missing Dependencies
```bash
# Install all build dependencies
sudo apt install dpkg-dev build-essential debhelper
sudo apt install golang-go  # For standalone package
```

#### Permission Errors During Build
```bash
# Fix source permissions
chmod +x packaging/build-deb.sh
sudo chown -R $USER:$USER packaging/
```

## 📈 Advanced Configuration

### Custom Package Configuration

Edit `packaging/deb/home-automation/DEBIAN/control` to customize:
- Dependencies
- Package description
- Maintainer information
- Conflicts/Replaces

### Resource Limits

For Raspberry Pi optimization, edit the Docker Compose file:
```yaml
services:
  prometheus:
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '0.5'
```

### Service Customization

Edit systemd service file for custom behavior:
```ini
[Service]
Environment=CUSTOM_VAR=value
ExecStartPre=/path/to/custom/script
MemoryMax=1G
CPUQuota=75%
```

## 🚀 Distribution

### Creating APT Repository

```bash
# Create repository structure
mkdir -p apt-repo/debian/{pool,dists/bullseye/main/binary-arm64}

# Copy packages
cp output/*.deb apt-repo/debian/pool/

# Generate Packages file
cd apt-repo/debian
dpkg-scanpackages pool /dev/null | gzip -9c > dists/bullseye/main/binary-arm64/Packages.gz

# Generate Release file
cat > dists/bullseye/Release << EOF
Origin: Home Automation
Label: Home Automation Packages
Suite: bullseye
Codename: bullseye
Date: $(date -u '+%a, %d %b %Y %H:%M:%S UTC')
Architectures: arm64
Components: main
Description: Home Automation System packages for Raspberry Pi
EOF
```

### GitHub Releases Integration

```bash
# Create release with packages
gh release create v1.0.0 \
  output/home-automation_1.0.0_arm64.deb \
  output/home-automation-standalone_1.0.0_arm64.deb \
  --title "Home Automation v1.0.0" \
  --notes "Initial Debian package release"
```

## 📚 References

- [Debian Policy Manual](https://www.debian.org/doc/debian-policy/)
- [Debian Maintainer's Guide](https://www.debian.org/doc/manuals/maint-guide/)
- [systemd Service Files](https://www.freedesktop.org/software/systemd/man/systemd.service.html)
- [Docker Compose Reference](https://docs.docker.com/compose/compose-file/)

## 🤝 Contributing

To contribute to package development:

1. **Test packages** on clean Raspberry Pi OS installations
2. **Report issues** with specific error messages and logs
3. **Submit improvements** via pull requests
4. **Document** any custom configurations or use cases

### Development Workflow

```bash
# Make changes to package files
edit packaging/deb/home-automation/DEBIAN/control

# Rebuild packages
./packaging/build-deb.sh

# Test in clean environment
# Deploy and validate

# Commit changes
git add packaging/
git commit -m "feat: improve package installation process"
```
