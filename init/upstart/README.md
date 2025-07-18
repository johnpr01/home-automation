# Home Automation - Upstart Service Configuration

This directory contains Upstart job files for Ubuntu 14.04 and older systems.

## Job Files

- `home-automation.conf` - Main system job (Docker Compose)
- `home-automation-standalone.conf` - Standalone Go binary job
- `tapo-metrics.conf` - Tapo metrics scraper job

## Installation

```bash
# Install main job
sudo cp home-automation.conf /etc/init/
sudo initctl reload-configuration
sudo service home-automation start

# Enable autostart
echo "manual" | sudo tee /etc/init/home-automation.override
# Remove override file to enable autostart
sudo rm /etc/init/home-automation.override
```

## Job Management

```bash
# Check status
sudo status home-automation

# View logs
sudo tail -f /var/log/upstart/home-automation.log

# Control job
sudo start home-automation
sudo stop home-automation
sudo restart home-automation
```

## Configuration

Jobs read environment variables from:
- `/etc/default/home-automation`
- `/opt/home-automation/.env`

## Notes

- Upstart is legacy and primarily used on Ubuntu 14.04 and older
- For modern systems, use systemd instead
- Limited security features compared to systemd
