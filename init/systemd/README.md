# Home Automation - systemd Service Configuration

This directory contains systemd service files for the home automation system.

## Service Files

- `home-automation.service` - Main system service (Docker Compose)
- `home-automation-standalone.service` - Standalone Go binary service
- `home-automation-dev.service` - Development service with file watching
- `tapo-metrics.service` - Tapo metrics scraper service only

## Installation

```bash
# Install main service
sudo cp home-automation.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable home-automation
sudo systemctl start home-automation

# Install standalone service (alternative)
sudo cp home-automation-standalone.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable home-automation-standalone
sudo systemctl start home-automation-standalone
```

## Service Management

```bash
# Check status
sudo systemctl status home-automation

# View logs
sudo journalctl -u home-automation -f

# Control service
sudo systemctl start home-automation
sudo systemctl stop home-automation
sudo systemctl restart home-automation
sudo systemctl reload home-automation

# Enable/disable autostart
sudo systemctl enable home-automation
sudo systemctl disable home-automation
```

## Configuration

Services read environment variables from:
- `/opt/home-automation/.env` (production)
- `/etc/default/home-automation` (system-wide defaults)

## Security Features

- Runs as non-root user (`pi`)
- Restricted file system access
- Private temporary directories
- No new privileges allowed
- Network namespace isolation (optional)
