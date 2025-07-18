# Home Automation - Supervisord Configuration

This directory contains Supervisord configuration files for process management.

## Configuration Files

- `home-automation.ini` - Main system configuration
- `home-automation-standalone.ini` - Standalone binary configuration
- `tapo-metrics.ini` - Tapo metrics scraper configuration
- `supervisord.conf` - Complete supervisord configuration with all programs

## Installation

```bash
# Install supervisord (Ubuntu/Debian)
sudo apt install supervisor

# Install supervisord (CentOS/RHEL)
sudo yum install supervisor

# Copy configuration files
sudo cp *.ini /etc/supervisor/conf.d/

# Reload supervisord configuration
sudo supervisorctl reread
sudo supervisorctl update
```

## Process Management

```bash
# Check status
sudo supervisorctl status

# Control specific programs
sudo supervisorctl start home-automation
sudo supervisorctl stop home-automation
sudo supervisorctl restart home-automation

# Control all programs
sudo supervisorctl start all
sudo supervisorctl stop all
sudo supervisorctl restart all

# View logs
sudo supervisorctl tail home-automation
sudo supervisorctl tail -f home-automation

# Reload configuration
sudo supervisorctl reread
sudo supervisorctl update
```

## Configuration

Programs read environment variables from:
- Configuration files (environment= directive)
- `/opt/home-automation/.env` (loaded by programs)

## Features

- Automatic process restart on failure
- Stdout/stderr logging
- Resource limits
- User/group isolation
- Environment variable support
- Web interface (optional)
