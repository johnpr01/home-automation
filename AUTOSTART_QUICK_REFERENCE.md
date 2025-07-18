# Home Automation Autostart - Quick Reference

## Installation

```bash
# Install autostart (run from home-automation directory)
sudo ./install-autostart.sh

# Uninstall autostart
sudo ./uninstall-autostart.sh
```

## Service Management

```bash
# Check service status
sudo systemctl status home-automation

# View logs
sudo journalctl -u home-automation -f

# Control service
sudo systemctl start home-automation     # Start
sudo systemctl stop home-automation      # Stop  
sudo systemctl restart home-automation   # Restart
sudo systemctl enable home-automation    # Enable autostart
sudo systemctl disable home-automation   # Disable autostart
```

## Web Access

After installation, access these interfaces:

| Service | URL | Credentials |
|---------|-----|-------------|
| Grafana | http://raspberrypi.local:3000 | admin/admin |
| Prometheus | http://raspberrypi.local:9090 | None |
| Tapo Metrics | http://raspberrypi.local:2112/metrics | None |
| Home API | http://raspberrypi.local:8080/api/status | None |

## Configuration

Edit environment variables:
```bash
sudo nano /opt/home-automation/.env
sudo systemctl restart home-automation  # Apply changes
```

## Troubleshooting

```bash
# View service logs
sudo journalctl -u home-automation -n 100

# Check Docker status
sudo docker ps
cd /opt/home-automation && sudo docker compose logs

# Test network connectivity to Tapo devices
cd /opt/home-automation && sudo docker compose exec tapo-metrics ping 192.168.68.60
```

## Files Created

- `/etc/systemd/system/home-automation.service` - Systemd service
- `/opt/home-automation/` - Installation directory
- `/opt/home-automation/.env` - Environment configuration

## Security Notes

- Service runs as `pi` user (non-root)
- Restricted file system access
- Change default Grafana password immediately
- Consider firewall rules for production use
