# 📦 Home Automation System - Installation Guide

Quick and easy installation of the Home Automation System on Raspberry Pi 5 using Debian packages.

## 🚀 Quick Start

### Step 1: Download Package
Download the appropriate package for your system:

- **Raspberry Pi 5 (4GB+ RAM)**: `home-automation_*.deb` (Full Docker stack)
- **Raspberry Pi (Limited resources)**: `home-automation-standalone_*.deb` (Lightweight)

### Step 2: Install Package
```bash
# Copy package to your Raspberry Pi
scp home-automation_*.deb pi@your-pi-ip:/home/pi/

# SSH to your Raspberry Pi
ssh pi@your-pi-ip

# Install the package
sudo dpkg -i home-automation_*.deb

# Fix any dependency issues
sudo apt-get install -f
```

### Step 3: Configure System
```bash
# Create initial configuration
home-automation-update create

# Edit configuration with your settings
home-automation-update edit

# Key settings to configure:
# - TPLINK_USERNAME=your@email.com
# - TPLINK_PASSWORD=yourpassword  
# - TAPO_DEVICE_*_IP=192.168.x.x
```

### Step 4: Start Services
```bash
# Start the system
sudo systemctl start home-automation

# Enable auto-start on boot
sudo systemctl enable home-automation

# Check status
home-automation-status
```

### Step 5: Access Web Interfaces
- **Grafana**: http://your-pi-ip:3000 (admin/admin)
- **Prometheus**: http://your-pi-ip:9090

## 📋 Package Comparison

| Feature | Main Package | Standalone Package |
|---------|-------------|-------------------|
| **Dependencies** | Docker, Docker Compose | Go compiler only |
| **Memory Usage** | ~1-2GB | ~200-500MB |
| **Disk Space** | ~4GB | ~1GB |
| **Components** | Full stack (Prometheus, Grafana, MQTT, Kafka) | Core automation + metrics |
| **Web UI** | ✅ Grafana + Prometheus | ✅ Basic metrics endpoint |
| **Best For** | Production, full monitoring | Development, resource-limited |

## 🔧 Management Commands

After installation, you have these management utilities:

### System Status
```bash
home-automation-status          # Comprehensive system overview
```

### View Logs
```bash
home-automation-logs            # Show recent logs
home-automation-logs -f         # Follow logs in real-time
home-automation-logs -n 100     # Show more lines
```

### Restart Services
```bash
home-automation-restart         # Full restart
home-automation-restart docker  # Docker containers only
home-automation-restart config  # Reload configuration
```

### Configuration Management
```bash
home-automation-update          # Edit configuration
home-automation-update validate # Check configuration
home-automation-update test     # Test device connectivity
home-automation-update show     # Display current settings
```

### Firmware Upload (for Raspberry Pi Pico)
```bash
upload-pico-firmware            # Upload sensor firmware to Pico
```

## ⚙️ Configuration

### Required Settings

Edit `/opt/home-automation/.env` with your settings:

```bash
# TP-Link Tapo Credentials
TPLINK_USERNAME=your@email.com
TPLINK_PASSWORD=yourpassword

# Device IP Addresses
TAPO_DEVICE_1_IP=192.168.1.100
TAPO_DEVICE_2_IP=192.168.1.101
TAPO_DEVICE_3_IP=192.168.1.102
TAPO_DEVICE_4_IP=192.168.1.103

# Protocol Settings (use false for older firmware)
TAPO_DEVICE_1_USE_KLAP=false
TAPO_DEVICE_2_USE_KLAP=false
TAPO_DEVICE_3_USE_KLAP=false
TAPO_DEVICE_4_USE_KLAP=false
```

### Optional: Raspberry Pi Pico Sensors

If using Raspberry Pi Pico WH sensors:

1. Flash MicroPython firmware to Pico
2. Configure sensor settings in `/opt/home-automation/firmware/pico-sht30/config.py`
3. Upload firmware: `upload-pico-firmware`

## 🔍 Troubleshooting

### Installation Issues

#### Dependency Problems
```bash
# Update package lists
sudo apt update

# Install missing dependencies manually
sudo apt install docker.io docker-compose-plugin

# Retry package installation
sudo dpkg -i home-automation_*.deb
```

#### Permission Issues
```bash
# Fix file permissions
sudo chown -R pi:pi /opt/home-automation
sudo usermod -aG docker pi

# Logout and login again for group changes
```

### Runtime Issues

#### Service Won't Start
```bash
# Check detailed status
sudo systemctl status home-automation -l

# View recent errors
home-automation-logs -n 50

# Common fixes
home-automation-update validate
home-automation-restart
```

#### Docker Issues (Main Package)
```bash
# Restart Docker
sudo systemctl restart docker

# Check Docker status
docker ps
docker compose -f /opt/home-automation/deployments/docker-compose.yml ps
```

#### Web Interfaces Not Accessible
```bash
# Check if services are running
home-automation-status

# Check firewall
sudo ufw status
sudo ufw allow 3000  # Grafana
sudo ufw allow 9090  # Prometheus
```

### Device Connectivity Issues

#### Tapo Devices Not Responding
```bash
# Test device connectivity
home-automation-update test

# Check network connectivity
ping 192.168.1.100  # Replace with your device IP

# Try legacy protocol
# Edit .env and set TAPO_DEVICE_X_USE_KLAP=false
```

## 🔄 Updates and Maintenance

### Package Updates
```bash
# Download new package version
# Install over existing version
sudo dpkg -i home-automation_new-version_arm64.deb
```

### Configuration Backup
```bash
# Manual backup
cp /opt/home-automation/.env /opt/home-automation/.env.backup

# Automatic backup (done by home-automation-update)
home-automation-update backup
```

### Log Management
```bash
# View log files
ls /var/log/home-automation/

# Clean old logs
sudo journalctl --vacuum-time=7d  # Keep last 7 days
```

## 🗑️ Removal

### Remove Package (Keep Configuration)
```bash
sudo apt remove home-automation
```

### Complete Removal (Delete Everything)
```bash
sudo apt purge home-automation
```

## 📚 Additional Resources

- **Project Documentation**: `/opt/home-automation/README.md`
- **Configuration Template**: `/opt/home-automation/.env.example`
- **Autostart Guide**: `/opt/home-automation/AUTOSTART_SETUP.md`
- **Pico Firmware Guide**: `/opt/home-automation/firmware/pico-sht30/README.md`

## 🆘 Getting Help

### Diagnostic Information
```bash
# System overview
home-automation-status

# Detailed logs
home-automation-logs -f

# Configuration check
home-automation-update validate

# Network test
home-automation-update test
```

### Common Solutions

1. **"Permission denied" errors**: Check user groups, file ownership
2. **"Connection refused" errors**: Verify IP addresses, network connectivity
3. **"Service failed to start" errors**: Check configuration, dependencies
4. **High resource usage**: Consider standalone package, adjust resource limits

## 🎯 Next Steps

After successful installation:

1. **Configure automation rules** using the collected sensor data
2. **Set up Grafana dashboards** for custom monitoring
3. **Add more Tapo devices** by extending the configuration
4. **Deploy Raspberry Pi Pico sensors** for room monitoring
5. **Integrate with other smart home systems** via MQTT

---

**Support**: For issues or questions, check the logs first, then consult the troubleshooting section above.
