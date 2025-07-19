# Home Automation System - Quick Start Guide

This guide will help you get your Home Automation system up and running with Home Assistant integration.

## Prerequisites

- Docker and Docker Compose installed
- Go 1.19+ (if building from source)
- Git

## Quick Setup

### 1. Clone and Setup

```bash
git clone <your-repo-url>
cd home-automation
```

### 2. Configure Environment

Copy the environment file and edit it:

```bash
cp .env.example .env
```

Edit `.env` file with your settings:
```bash
# API Configuration
API_KEY=your-secure-api-key-here
PORT=8080

# MQTT Configuration
MQTT_BROKER=mqtt://localhost:1883
MQTT_USERNAME=homeassistant
MQTT_PASSWORD=mqtt_password

# Database (if using)
DATABASE_URL=sqlite:///app/data/home_automation.db

# Prometheus
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=9090
```

### 3. Create Required Directories

```bash
mkdir -p mosquitto/{config,data,log}
mkdir -p prometheus/rules
mkdir -p grafana/{provisioning,dashboards}
mkdir -p home-assistant-config
mkdir -p config data
```

### 4. Setup MQTT Configuration

Create `mosquitto/config/mosquitto.conf`:

```
listener 1883
allow_anonymous true
persistence true
persistence_location /mosquitto/data/
log_dest file /mosquitto/log/mosquitto.log
```

### 5. Start the Services

```bash
docker-compose up -d
```

This will start:
- Home Automation API (port 8080)
- MQTT Broker (port 1883)
- Prometheus (port 9090)
- Grafana (port 3000)
- Home Assistant (host network mode)
- Redis cache
- Node Exporter and cAdvisor for monitoring

### 6. Verify Services

Check that services are running:

```bash
docker-compose ps
```

Test the API:

```bash
curl http://localhost:8080/api/health
curl http://localhost:8080/api/rooms
```

## Home Assistant Integration

### 1. Install the Custom Integration

The integration is automatically mounted into Home Assistant via Docker volumes.

### 2. Add Integration in Home Assistant

1. Go to **Settings** → **Devices & Services**
2. Click **Add Integration**
3. Search for "Home Automation"
4. Enter your configuration:
   - **Host**: `home-automation` (or your IP)
   - **Port**: `8080`
   - **API Key**: (from your .env file)
   - **MQTT Host**: `mosquitto` (optional)

### 3. Configure Home Assistant

Add to your `configuration.yaml`:

```yaml
# Copy from integrations/home-assistant/configuration_example.yaml
```

## Testing the Setup

### 1. Add a Room via API

```bash
curl -X POST http://localhost:8080/api/rooms \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{"name": "Living Room"}'
```

### 2. Check Home Assistant

The new room should appear as devices in Home Assistant with sensors, switches, lights, and climate controls.

### 3. Test MQTT

Publish a test message:

```bash
docker exec mosquitto mosquitto_pub -t "room-temp/living-room" -m "72.5"
```

### 4. View Metrics

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)
- **Home Assistant**: http://localhost:8123

## Managing Rooms

### Add Room
```bash
curl -X POST http://localhost:8080/api/rooms \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{"name": "Bedroom"}'
```

### List Rooms
```bash
curl http://localhost:8080/api/rooms \
  -H "X-API-Key: your-api-key"
```

### Remove Room
```bash
curl -X DELETE http://localhost:8080/api/rooms/bedroom \
  -H "X-API-Key: your-api-key"
```

## MQTT Topics

The system uses these MQTT topics:

- `room-temp/{room_id}` - Temperature sensor data
- `room-humidity/{room_id}` - Humidity sensor data  
- `room-motion/{room_id}` - Motion detection
- `room-light/{room_id}` - Light level
- `room-occupancy/{room_id}` - Room occupancy
- `room-climate/{room_id}/target_temp/set` - Set target temperature
- `room-light/{room_id}/set` - Control room lights
- `room-switch/{room_id}/set` - Control room switches

## Troubleshooting

### Services Not Starting

Check logs:
```bash
docker-compose logs home-automation
docker-compose logs mosquitto
docker-compose logs homeassistant
```

### API Not Responding

1. Check if the container is running: `docker-compose ps`
2. Check logs: `docker-compose logs home-automation`
3. Verify API key in requests

### Home Assistant Integration Not Found

1. Check custom_components are mounted: `docker-compose exec homeassistant ls /config/custom_components`
2. Restart Home Assistant: `docker-compose restart homeassistant`
3. Check Home Assistant logs for errors

### MQTT Issues

1. Test MQTT connection:
```bash
docker exec mosquitto mosquitto_pub -t "test" -m "hello"
docker exec mosquitto mosquitto_sub -t "test"
```

2. Check MQTT logs: `docker-compose logs mosquitto`

### Prometheus/Grafana Issues

1. Check if metrics endpoint is available: `curl http://localhost:8080/metrics`
2. Verify Prometheus targets: http://localhost:9090/targets
3. Check Grafana data sources: http://localhost:3000

## Advanced Configuration

### Custom Metrics

Add custom metrics to your Go application using the Prometheus client library.

### Additional Sensors

Extend the MQTT topics and Home Assistant entities by modifying the platform files.

### Alerting

Configure Prometheus alerts in `prometheus/rules/` directory.

### Grafana Dashboards

Import or create custom dashboards in Grafana for your specific metrics.

## Development

### Building from Source

```bash
go build -o home-automation ./cmd/server
```

### Running Tests

```bash
go test ./...
```

### Hot Reload (Development)

Use `air` for hot reloading during development:

```bash
go install github.com/cosmtrek/air@latest
air
```

## Support

For issues and questions:

1. Check the logs first
2. Review the configuration files
3. Test individual components
4. Check the GitHub issues page

## Security Notes

- Change default passwords
- Use strong API keys
- Consider enabling MQTT authentication
- Run behind a reverse proxy in production
- Regularly update container images
