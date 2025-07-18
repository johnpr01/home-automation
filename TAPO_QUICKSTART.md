# Tapo Smart Plug Energy Monitoring - Quick Start

## ✅ Successfully Integrated Tapo Metrics Scraper

The Tapo smart plug energy monitoring system has been successfully integrated into the Docker Compose stack with the following components:

### 🔧 Current Setup

1. **tapo-metrics-scraper**: Dedicated Go service for monitoring Tapo smart plugs
2. **Prometheus**: Time series database collecting metrics from tapo-metrics
3. **Grafana**: Visualization dashboard (ready for configuration)

### 🚀 Quick Start

#### 1. Configure Environment Variables

Copy the example environment file and customize:
```bash
cp .env.example .env
```

Edit `.env` and set your Tapo credentials:
```bash
# Required: Your TP-Link cloud account credentials
TPLINK_USERNAME=your_tplink_email@example.com
TPLINK_PASSWORD=your_tplink_password

# Configure your Tapo device IP addresses
TAPO_DEVICE_1_IP=192.168.1.100
TAPO_DEVICE_2_IP=192.168.1.101

# Protocol settings (KLAP for newer firmware, legacy for older)
TAPO_DEVICE_1_USE_KLAP=true
TAPO_DEVICE_2_USE_KLAP=true
```

#### 2. Start the Services

```bash
# Start Prometheus and Tapo metrics scraper
docker-compose up -d prometheus tapo-metrics

# Verify services are running
docker-compose ps
```

#### 3. Access the Interfaces

- **Tapo Metrics**: http://localhost:2112 (health, metrics, web interface)
- **Prometheus**: http://localhost:9090 (query interface, targets status)
- **Grafana**: http://localhost:3000 (visualization - requires additional setup)

### 📊 Verification Steps

1. **Health Check**: `curl http://localhost:2112/health`
2. **Metrics Endpoint**: `curl http://localhost:2112/metrics`
3. **Prometheus Targets**: Visit http://localhost:9090/targets and verify `tapo-metrics:2112` is UP
4. **Query Metrics**: In Prometheus, query `up{job="tapo-metrics"}` should return `1`

### 🔍 Current Status

✅ **tapo-metrics service**: Running and healthy  
✅ **Prometheus integration**: Successfully scraping metrics  
✅ **KLAP protocol support**: Native Go implementation for modern Tapo devices  
✅ **Health monitoring**: Automated health checks and restart policies  
✅ **Resource optimization**: Configured for Raspberry Pi 5  

### 📁 Key Files

- `cmd/tapo-metrics-scraper/main.go`: Main metrics scraper service
- `pkg/tapo/klap_client.go`: KLAP protocol implementation
- `internal/services/tapo_service.go`: Service logic for device management
- `docker-compose.yml`: Full orchestration configuration
- `prometheus.yml`: Prometheus scrape configuration
- `Dockerfile.tapo`: Container build configuration

### 🐛 Troubleshooting

- **Service won't start**: Check that `TPLINK_PASSWORD` is set in environment
- **Port conflicts**: Ensure port 2112 is not in use by other services
- **Device discovery**: Verify device IP addresses in network settings
- **Protocol issues**: Try toggling `USE_KLAP` setting if connection fails

### 🎯 Next Steps

1. **Configure actual device IPs** in your environment variables
2. **Set up Grafana dashboards** for energy monitoring visualization
3. **Add more devices** by extending the environment configuration
4. **Monitor logs** with `docker-compose logs tapo-metrics`

The system is now ready for production use with real Tapo smart plug devices!
