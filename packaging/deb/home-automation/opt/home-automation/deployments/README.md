# Raspberry Pi 5 Home Automation Server

This directory contains deployment configurations optimized for Raspberry Pi 5.

## Prerequisites

### Hardware Requirements
- Raspberry Pi 5 (4GB or 8GB RAM recommended)
- MicroSD card (32GB minimum, Class 10 or better)
- Adequate power supply (official Pi 5 power adapter recommended)
- Ethernet connection or WiFi

### Software Requirements
- Raspberry Pi OS (64-bit, Bookworm or later)
- Docker and Docker Compose
- Git

## Initial Setup

### 1. Prepare Raspberry Pi 5

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo apt install docker-compose-plugin -y

# Reboot to apply group changes
sudo reboot
```

### 2. Clone and Setup Project

```bash
# Clone the repository
git clone https://github.com/yourname/home-automation.git
cd home-automation

# Create necessary directories
mkdir -p logs
mkdir -p deployments/mosquitto
mkdir -p deployments/grafana/provisioning

# Set permissions
sudo chown -R $USER:$USER logs
```

### 3. Configure Environment

```bash
# Copy environment template
cp .env.example .env

# Edit configuration for Raspberry Pi 5
nano .env
```

### 4. Deploy Services

```bash
cd deployments

# Start all services
docker compose up -d

# Check status
docker compose ps

# View logs
docker compose logs -f
```

## Raspberry Pi 5 Optimizations

### Memory Management
- Services are configured with memory limits appropriate for Pi 5
- PostgreSQL uses optimized buffer settings
- Kafka heap size limited to 256MB
- Redis configured with memory policies

### CPU Usage
- Each service has CPU limits to prevent resource starvation
- Background processes are throttled appropriately

### Storage Optimization
- Kafka log retention set to 24 hours
- Redis persistence optimized for SD card
- PostgreSQL checkpoint settings tuned for Pi storage

### Network Configuration
- Services use bridge networking for isolation
- Ports exposed only as needed
- MQTT WebSocket support enabled

## Service URLs

Once deployed, services are available at:

- **Home Automation API**: http://your-pi-ip:8080
- **Grafana Dashboard**: http://your-pi-ip:3000 (admin/admin)
- **MQTT Broker**: your-pi-ip:1883
- **MQTT WebSocket**: your-pi-ip:9001
- **PostgreSQL**: your-pi-ip:5432
- **Redis**: your-pi-ip:6379
- **Kafka**: your-pi-ip:9092

## Monitoring and Maintenance

### Health Checks

```bash
# Check service health
./scripts/health-check.sh

# View resource usage
docker stats

# Monitor logs
docker compose logs -f --tail=50
```

### Backup

```bash
# Backup configuration and data
./scripts/backup.sh

# Restore from backup
./scripts/restore.sh backup-filename.tar.gz
```

### Updates

```bash
# Update images
docker compose pull

# Restart services with new images
docker compose up -d

# Clean up old images
docker image prune -f
```

## Performance Tuning

### SD Card Optimization

```bash
# Add to /boot/cmdline.txt for better I/O performance
cgroup_enable=memory cgroup_memory=1

# Mount tmpfs for logs (optional)
echo "tmpfs /tmp tmpfs defaults,noatime,nosuid,size=100m 0 0" | sudo tee -a /etc/fstab
```

### System Configuration

```bash
# Increase file descriptors
echo "* soft nofile 65536" | sudo tee -a /etc/security/limits.conf
echo "* hard nofile 65536" | sudo tee -a /etc/security/limits.conf

# Optimize networking
echo "net.core.rmem_default = 262144" | sudo tee -a /etc/sysctl.conf
echo "net.core.rmem_max = 16777216" | sudo tee -a /etc/sysctl.conf
echo "net.core.wmem_default = 262144" | sudo tee -a /etc/sysctl.conf
echo "net.core.wmem_max = 16777216" | sudo tee -a /etc/sysctl.conf
```

## Troubleshooting

### Common Issues

1. **Out of Memory**
   - Check `docker stats` for memory usage
   - Reduce service memory limits if needed
   - Consider using swap file

2. **SD Card Corruption**
   - Use high-quality SD card
   - Consider mounting data volumes on USB storage
   - Enable regular backups

3. **Network Issues**
   - Check firewall settings: `sudo ufw status`
   - Verify port bindings: `netstat -tulpn`
   - Test connectivity: `ping service-name`

### Log Analysis

```bash
# View service logs
docker compose logs service-name

# Follow logs in real-time
docker compose logs -f service-name

# Search logs for errors
docker compose logs service-name 2>&1 | grep -i error
```

## Security Considerations

### Network Security
- Change default passwords in production
- Consider VPN access for remote management
- Use strong passwords for MQTT if authentication enabled
- Keep Raspberry Pi OS updated

### Container Security
- Regularly update Docker images
- Review exposed ports
- Consider using secrets for sensitive data

## Hardware Monitoring

The system includes monitoring for:
- CPU temperature and usage
- Memory usage
- Disk I/O and space
- Network traffic
- Service health

Access monitoring dashboard at: http://your-pi-ip:3000
