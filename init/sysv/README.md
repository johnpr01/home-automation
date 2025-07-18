# Home Automation - SysV Init Scripts

This directory contains SysV init scripts for older Linux distributions and BSD systems.

## Init Scripts

- `home-automation` - Main system init script (Docker Compose)
- `home-automation-standalone` - Standalone Go binary init script
- `tapo-metrics` - Tapo metrics scraper init script

## Installation

```bash
# Install main script
sudo cp home-automation /etc/init.d/
sudo chmod +x /etc/init.d/home-automation
sudo update-rc.d home-automation defaults

# For Red Hat/CentOS systems:
sudo cp home-automation /etc/rc.d/init.d/
sudo chmod +x /etc/rc.d/init.d/home-automation
sudo chkconfig home-automation on
```

## Script Management

```bash
# Control service
sudo service home-automation start
sudo service home-automation stop
sudo service home-automation restart
sudo service home-automation status

# Or use init script directly
sudo /etc/init.d/home-automation start
sudo /etc/init.d/home-automation stop
sudo /etc/init.d/home-automation restart
sudo /etc/init.d/home-automation status
```

## Configuration

Scripts read configuration from:
- `/etc/default/home-automation`
- `/opt/home-automation/.env`

## Notes

- SysV init is legacy and used on older systems
- Limited process management compared to modern init systems
- Manual PID file management required
- Basic logging to syslog
