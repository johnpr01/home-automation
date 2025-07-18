# Home Automation System - Init Configurations

This directory contains initialization and process management configurations for various init systems and process supervisors.

## 📁 Directory Structure

```
init/
├── README.md                    # This file
├── systemd/                     # systemd service files (modern Linux)
│   ├── README.md
│   ├── home-automation.service  # Main Docker Compose service
│   ├── home-automation-standalone.service
│   ├── home-automation-dev.service
│   └── tapo-metrics.service
├── upstart/                     # Upstart job files (Ubuntu 14.04 and older)
│   ├── README.md
│   ├── home-automation.conf
│   ├── home-automation-standalone.conf
│   └── tapo-metrics.conf
├── sysv/                        # SysV init scripts (legacy systems)
│   ├── README.md
│   ├── home-automation
│   └── home-automation-standalone
├── supervisord/                 # Supervisord process manager
│   ├── README.md
│   ├── home-automation.ini
│   ├── home-automation-standalone.ini
│   ├── tapo-metrics.ini
│   └── supervisord.conf
└── runit/                       # Runit process supervisor
    ├── README.md
    ├── home-automation/
    │   ├── run
    │   ├── finish
    │   └── log/run
    ├── home-automation-standalone/
    │   ├── run
    │   └── log/run
    └── tapo-metrics/
        ├── run
        └── log/run
```

## 🔧 Init System Support

### systemd (Recommended)
**Use for:** Modern Linux distributions (Ubuntu 16.04+, Debian 8+, CentOS 7+, RHEL 7+)

**Features:**
- Advanced security features (user isolation, file system restrictions)
- Dependency management
- Resource limits
- Socket activation
- Detailed logging with journald

**Installation:**
```bash
sudo cp init/systemd/home-automation.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable home-automation
sudo systemctl start home-automation
```

### Upstart (Legacy)
**Use for:** Ubuntu 14.04 and older systems

**Features:**
- Event-based startup
- Automatic respawn
- Simple configuration syntax
- Basic process management

**Installation:**
```bash
sudo cp init/upstart/home-automation.conf /etc/init/
sudo initctl reload-configuration
sudo service home-automation start
```

### SysV Init (Legacy)
**Use for:** Very old Linux distributions, some embedded systems

**Features:**
- Traditional Unix init system
- Shell script based
- Manual PID management
- Basic service control

**Installation:**
```bash
sudo cp init/sysv/home-automation /etc/init.d/
sudo chmod +x /etc/init.d/home-automation
sudo update-rc.d home-automation defaults  # Debian/Ubuntu
sudo chkconfig home-automation on          # RHEL/CentOS
```

### Supervisord (Process Manager)
**Use for:** Process supervision on any Linux system, development environments

**Features:**
- Process monitoring and restart
- Web-based management interface
- Detailed logging
- Group management
- XML-RPC API

**Installation:**
```bash
sudo apt install supervisor
sudo cp init/supervisord/*.ini /etc/supervisor/conf.d/
sudo supervisorctl reread
sudo supervisorctl update
```

### Runit (Process Supervisor)
**Use for:** Void Linux, Alpine Linux, or as systemd alternative

**Features:**
- Simple and reliable
- Fast startup
- Structured logging
- Clean process trees
- No dependencies

**Installation:**
```bash
sudo cp -r init/runit/home-automation /etc/sv/
sudo ln -s /etc/sv/home-automation /var/service/
```

## 🚀 Service Variants

### Main Service (Docker Compose)
Runs the complete home automation system using Docker Compose.

**Components:**
- Prometheus (metrics)
- Grafana (dashboards)
- Tapo metrics scraper
- MQTT broker
- All supporting services

**Use when:** You want the complete system with all features

### Standalone Service
Runs only the Go binary without Docker dependencies.

**Components:**
- Home automation server binary
- Direct device communication
- Simplified deployment

**Use when:** You prefer lightweight deployment or don't want Docker

### Development Service
Runs from source code with hot reload for development.

**Components:**
- Go source compilation
- Development environment
- Debug logging

**Use when:** You're developing or testing the system

### Tapo Metrics Only
Runs only the Tapo smart plug metrics scraper.

**Components:**
- Tapo device monitoring
- Prometheus metrics export
- Lightweight deployment

**Use when:** You only need smart plug monitoring

## ⚙️ Configuration

### Environment Files
All services read configuration from:
- `/opt/home-automation/.env` - Main configuration
- `/etc/default/home-automation` - System defaults
- Service-specific environment variables

### Required Variables
```bash
# TP-Link Tapo Credentials
TPLINK_USERNAME=your@email.com
TPLINK_PASSWORD=yourpassword

# Device Configuration
TAPO_DEVICE_1_IP=192.168.68.54
TAPO_DEVICE_2_IP=192.168.68.63
TAPO_DEVICE_3_IP=192.168.68.60
TAPO_DEVICE_4_IP=192.168.68.53

# Protocol Settings
TAPO_DEVICE_1_USE_KLAP=false
TAPO_DEVICE_2_USE_KLAP=false
TAPO_DEVICE_3_USE_KLAP=false
TAPO_DEVICE_4_USE_KLAP=false
```

## 🔒 Security Considerations

### User Isolation
All services run as the `pi` user (non-root) for security:
- Limited file system access
- No privilege escalation
- Restricted resource usage

### systemd Security Features
- `NoNewPrivileges=true` - Prevents privilege escalation
- `PrivateTmp=true` - Isolated temporary directories
- `ProtectSystem=strict` - Read-only system directories
- `ProtectHome=true` - No access to user home directories

### File Permissions
```bash
# Secure configuration files
sudo chown root:pi /opt/home-automation/.env
sudo chmod 640 /opt/home-automation/.env

# Secure service files
sudo chown root:root /etc/systemd/system/home-automation.service
sudo chmod 644 /etc/systemd/system/home-automation.service
```

## 📊 Monitoring and Logging

### systemd Logging
```bash
# View logs
sudo journalctl -u home-automation -f

# View logs with timestamps
sudo journalctl -u home-automation --since "1 hour ago"

# Export logs
sudo journalctl -u home-automation --since today > today-logs.txt
```

### Supervisord Logging
```bash
# View logs
sudo supervisorctl tail home-automation
sudo supervisorctl tail -f home-automation

# Log files location
/var/log/supervisor/home-automation.log
```

### Runit Logging
```bash
# View logs
sudo svlogtail home-automation

# Log files location
/var/log/home-automation/
```

## 🔧 Troubleshooting

### Common Issues

#### Service Won't Start
```bash
# Check service status
sudo systemctl status home-automation  # systemd
sudo service home-automation status    # SysV/Upstart
sudo supervisorctl status               # supervisord
sudo sv status home-automation          # runit

# Check logs for errors
sudo journalctl -u home-automation -n 50  # systemd
sudo tail -f /var/log/upstart/home-automation.log  # upstart
sudo supervisorctl tail home-automation   # supervisord
sudo svlogtail home-automation            # runit
```

#### Permission Errors
```bash
# Fix ownership
sudo chown -R pi:pi /opt/home-automation
sudo chmod +x /opt/home-automation/bin/*
```

#### Environment Issues
```bash
# Verify environment file
sudo cat /opt/home-automation/.env
sudo systemctl show-environment  # systemd
```

#### Docker Issues (for Docker Compose services)
```bash
# Check Docker status
sudo systemctl status docker
sudo docker ps
sudo docker compose logs
```

### Service Management Quick Reference

| Action | systemd | Upstart | SysV | Supervisord | Runit |
|--------|---------|---------|------|-------------|-------|
| Start | `systemctl start` | `service start` | `service start` | `supervisorctl start` | `sv start` |
| Stop | `systemctl stop` | `service stop` | `service stop` | `supervisorctl stop` | `sv stop` |
| Restart | `systemctl restart` | `service restart` | `service restart` | `supervisorctl restart` | `sv restart` |
| Status | `systemctl status` | `status` | `service status` | `supervisorctl status` | `sv status` |
| Logs | `journalctl -u` | `tail /var/log/upstart/` | `tail /var/log/syslog` | `supervisorctl tail` | `svlogtail` |
| Enable | `systemctl enable` | (automatic) | `update-rc.d` | (automatic) | `ln -s /etc/sv/` |

## 🎯 Choosing the Right Init System

### For Production (Raspberry Pi)
**Recommended:** systemd
- Modern Raspberry Pi OS uses systemd
- Best security features
- Comprehensive logging
- Active development

### For Development
**Recommended:** supervisord
- Easy configuration changes
- Web interface for monitoring
- Detailed process control
- Works alongside any init system

### For Minimal Systems
**Recommended:** runit
- Lightweight and fast
- Simple configuration
- Reliable process supervision
- Good for embedded systems

### For Legacy Systems
**Use:** SysV init or Upstart
- Only if systemd is not available
- Limited features but widely compatible
- Manual maintenance required

## 📚 Additional Resources

- [systemd Documentation](https://systemd.io/)
- [Supervisord Documentation](http://supervisord.org/)
- [Runit Documentation](http://smarden.org/runit/)
- [Home Automation Setup Guide](../AUTOSTART_SETUP.md)
- [Installation Scripts](../install-autostart.sh)
