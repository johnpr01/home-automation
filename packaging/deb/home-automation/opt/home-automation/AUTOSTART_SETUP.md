# Home Automation Autostart Setup

This document explains how to set up the home automation system to start automatically when your Raspberry Pi boots.

## Overview

The autostart system uses:
- **systemd**: Linux service manager to control the application lifecycle
- **Docker Compose**: Container orchestration for all services
- **Installation script**: Automated setup and configuration
- **Service management**: Start, stop, restart, and monitor the system

## Quick Start

### Install Autostart
```bash
# Make sure you're in the home-automation directory
cd /path/to/home-automation

# Run the installation script as root
sudo ./install-autostart.sh
```

### Check Status
```bash
# Check if the service is running
sudo systemctl status home-automation

# View real-time logs
sudo journalctl -u home-automation -f
```

### Control the Service
```bash
# Start the service
sudo systemctl start home-automation

# Stop the service
sudo systemctl stop home-automation

# Restart the service
sudo systemctl restart home-automation

# Disable autostart (but keep service installed)
sudo systemctl disable home-automation

# Enable autostart
sudo systemctl enable home-automation
```

## Installation Details

### What the installer does:
1. **Checks prerequisites**: Verifies Docker and Docker Compose are installed
2. **Creates installation directory**: `/opt/home-automation`
3. **Copies application files**: `docker-compose.yml`, configs, environment files
4. **Installs systemd service**: Registers with system service manager
5. **Enables autostart**: Configures to start on boot
6. **Optionally starts service**: Can start immediately after installation

### File Locations After Installation:
```
/opt/home-automation/           # Main installation directory
├── docker-compose.yml          # Docker services configuration
├── prometheus.yml              # Prometheus configuration
├── .env                        # Environment variables (from .env.example)
├── deployments/                # Deployment configurations
└── configs/                    # Application configurations

/etc/systemd/system/home-automation.service  # systemd service file
```

## Service Configuration

The systemd service (`home-automation.service`) includes:

### Service Properties:
- **User/Group**: Runs as `pi` user for security
- **Working Directory**: `/opt/home-automation`
- **Restart Policy**: Automatically restarts if it crashes
- **Dependencies**: Waits for network and Docker to be ready
- **Security**: Restricted permissions and isolated temporary files

### Commands Used:
- **Start**: `docker compose up`
- **Stop**: `docker compose down`  
- **Restart**: `docker compose restart`
- **Pre-start**: `docker compose pull --quiet` (updates images)

## Environment Configuration

### Required Environment Variables:
The service loads environment variables from `/opt/home-automation/.env`:

```bash
# TP-Link Tapo Credentials
TPLINK_USERNAME=your@email.com
TPLINK_PASSWORD=yourpassword

# Tapo Device IPs
TAPO_DEVICE_1_IP=192.168.68.54
TAPO_DEVICE_2_IP=192.168.68.63
TAPO_DEVICE_3_IP=192.168.68.60
TAPO_DEVICE_4_IP=192.168.68.53

# Protocol Settings (use legacy to avoid error 1003)
TAPO_DEVICE_1_USE_KLAP=false
TAPO_DEVICE_2_USE_KLAP=false
TAPO_DEVICE_3_USE_KLAP=false
TAPO_DEVICE_4_USE_KLAP=false
```

### Editing Configuration:
```bash
# Edit environment variables
sudo nano /opt/home-automation/.env

# Restart service to apply changes
sudo systemctl restart home-automation
```

## Web Interfaces

After the service starts, these interfaces will be available:

| Service | URL | Credentials |
|---------|-----|-------------|
| **Grafana** | http://raspberrypi.local:3000 | admin/admin |
| **Prometheus** | http://raspberrypi.local:9090 | None |
| **Tapo Metrics** | http://raspberrypi.local:2112/metrics | None |
| **Home API** | http://raspberrypi.local:8080/api/status | None |

*Replace `raspberrypi.local` with your Pi's IP address if needed*

## Troubleshooting

### Service Won't Start
```bash
# Check service status and logs
sudo systemctl status home-automation
sudo journalctl -u home-automation -n 50

# Check Docker
sudo docker ps
sudo docker compose logs
```

### Common Issues:

#### 1. Docker Not Running
```bash
# Start Docker service
sudo systemctl start docker
sudo systemctl enable docker
```

#### 2. Permission Issues
```bash
# Fix ownership of installation directory
sudo chown -R pi:pi /opt/home-automation
```

#### 3. Port Conflicts
```bash
# Check what's using ports
sudo netstat -tlnp | grep ':3000\|:9090\|:2112\|:8080'
```

#### 4. Environment Variables
```bash
# Verify .env file exists and has correct values
sudo cat /opt/home-automation/.env
```

### Viewing Logs:
```bash
# Real-time service logs
sudo journalctl -u home-automation -f

# Last 100 log lines
sudo journalctl -u home-automation -n 100

# Docker container logs
cd /opt/home-automation
sudo docker compose logs tapo-metrics
sudo docker compose logs grafana
sudo docker compose logs prometheus
```

## Uninstalling

### Remove Autostart:
```bash
# Run the uninstall script
sudo ./uninstall-autostart.sh
```

### Manual Removal:
```bash
# Stop and disable service
sudo systemctl stop home-automation
sudo systemctl disable home-automation

# Remove service file
sudo rm /etc/systemd/system/home-automation.service
sudo systemctl daemon-reload

# Optionally remove installation directory
sudo rm -rf /opt/home-automation
```

## Security Considerations

### Service Security:
- Runs as non-root `pi` user
- Restricted file system access
- No new privileges allowed
- Private temporary directory

### Network Security:
- Services bind to all interfaces (0.0.0.0)
- Consider firewall rules for production use
- Change default Grafana password

### Recommendations:
1. **Change default passwords** in Grafana immediately
2. **Secure your network** - ensure proper firewall/router configuration
3. **Regular updates** - keep Docker images updated
4. **Monitor logs** - watch for unusual activity

## Performance

### Resource Usage:
- **Memory**: ~500MB-1GB total for all containers
- **CPU**: Low usage during normal operation
- **Storage**: ~2GB for images + logs/data growth
- **Network**: Minimal external traffic (only image pulls)

### Optimization:
- Logs are rotated automatically by systemd
- Docker images are updated on service restart
- Consider SSD storage for better performance

## Advanced Configuration

### Custom Service Settings:
Edit `/etc/systemd/system/home-automation.service` to customize:
- Restart policies
- Resource limits
- Environment variables
- Security settings

### Docker Compose Overrides:
Create `/opt/home-automation/docker-compose.override.yml` for:
- Custom port mappings
- Volume mounts
- Service modifications
- Development settings

## Support

### Log Collection for Support:
```bash
# Collect system information
sudo systemctl status home-automation > support-logs.txt
sudo journalctl -u home-automation -n 100 >> support-logs.txt
cd /opt/home-automation && sudo docker compose logs >> support-logs.txt
```

### Reset to Defaults:
```bash
# Stop service
sudo systemctl stop home-automation

# Reset to original configuration
cd /path/to/home-automation/source
sudo cp docker-compose.yml /opt/home-automation/
sudo cp .env.example /opt/home-automation/.env

# Restart service
sudo systemctl start home-automation
```
