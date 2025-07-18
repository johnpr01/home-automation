# Home Automation - Runit Service Configuration

This directory contains Runit service configurations for process supervision.

## Service Directories

- `home-automation/` - Main system service (Docker Compose)
- `home-automation-standalone/` - Standalone binary service
- `tapo-metrics/` - Tapo metrics scraper service

## Installation

```bash
# Install runit (Ubuntu/Debian)
sudo apt install runit

# Install runit (Void Linux - already included)

# Copy service directories
sudo cp -r home-automation* /etc/sv/
sudo cp -r tapo-metrics /etc/sv/

# Enable services
sudo ln -s /etc/sv/home-automation /var/service/
sudo ln -s /etc/sv/tapo-metrics /var/service/
```

## Service Management

```bash
# Check status
sudo sv status home-automation
sudo sv status tapo-metrics

# Control services
sudo sv start home-automation
sudo sv stop home-automation
sudo sv restart home-automation

# View logs
sudo svlogtail home-automation
sudo svlogtail tapo-metrics

# Enable/disable services
sudo sv enable home-automation
sudo sv disable home-automation
```

## Configuration

Services read environment variables from:
- `run` scripts (environment setup)
- `/opt/home-automation/.env` (loaded by services)

## Features

- Simple and reliable process supervision
- Automatic restart on failure
- Structured logging with svlogd
- Fast startup and shutdown
- Minimal resource usage
- Clean process trees
