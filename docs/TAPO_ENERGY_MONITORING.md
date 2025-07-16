# TP-Link Tapo Smart Plug Energy Monitoring

This module provides comprehensive energy monitoring for TP-Link Tapo smart plugs, including data storage in InfluxDB and visualization with Grafana dashboards.

## Features

- **Real-time Energy Monitoring**: Track power consumption, voltage, current, and energy usage
- **Multi-device Support**: Monitor multiple Tapo smart plugs simultaneously
- **InfluxDB Integration**: Store time-series energy data for historical analysis
- **MQTT Publishing**: Real-time energy data published to MQTT topics
- **Grafana Dashboards**: Pre-built dashboards for energy visualization
- **Device Control**: Turn devices on/off remotely
- **Automatic Reconnection**: Robust error handling and reconnection logic

## Supported Devices

- TP-Link Tapo P100 (Smart Plug)
- TP-Link Tapo P110 (Smart Plug with Energy Monitoring)
- TP-Link Tapo P115 (Smart Plug with Energy Monitoring)

## Setup Instructions

### 1. Prerequisites

- TP-Link Tapo smart plugs configured on your network
- Tapo app account (username/password)
- Running InfluxDB instance
- Running MQTT broker (Mosquitto)
- Running Grafana instance

### 2. Configure Devices

1. Copy the configuration template:
   ```bash
   cp configs/tapo_template.yml configs/tapo.yml
   ```

2. Edit `configs/tapo.yml` with your device details:
   ```yaml
   devices:
     - device_id: "tapo_living_room_1"
       device_name: "Living Room Lamp"
       room_id: "living_room"
       ip_address: "192.168.1.100"  # Your device IP
       username: "your_tapo_username"
       password: "your_tapo_password"
       poll_interval: 30s
   ```

### 3. Find Device IP Addresses

Use one of these methods to find your Tapo device IP addresses:

**Method 1: Router Admin Panel**
- Log into your router's admin interface
- Look for connected devices named "Tapo" or "TP-Link"

**Method 2: Network Scanner**
```bash
# Install nmap if not available
sudo apt install nmap

# Scan your network (replace with your network range)
nmap -sn 192.168.1.0/24 | grep -B2 "Tapo\|TP-Link"
```

**Method 3: Tapo App**
- Open the Tapo app
- Go to device settings
- Look for "Device Info" or "Network Info"

### 4. Start the Service

Using Docker Compose:
```bash
cd deployments
docker-compose up -d influxdb grafana mosquitto
```

Run the Tapo monitoring service:
```bash
cd cmd/tapo-demo
go run main.go
```

### 5. Access Grafana Dashboard

1. Open Grafana: http://localhost:3000
2. Login: admin/admin
3. Navigate to "Tapo Smart Plug Energy Monitoring" dashboard

## Configuration Options

### Device Configuration
```yaml
devices:
  - device_id: "unique_device_id"        # Unique identifier
    device_name: "Human readable name"   # Display name
    room_id: "room_identifier"           # Room/location
    ip_address: "192.168.1.100"          # Device IP address
    username: "tapo_username"            # Tapo account username
    password: "tapo_password"            # Tapo account password
    poll_interval: 30s                   # How often to poll device
```

### Polling Intervals
- **30s**: Good for frequently used devices
- **60s**: Suitable for always-on devices
- **300s**: For devices that don't change often

## Data Metrics

The following metrics are collected and stored:

| Metric | Description | Unit |
|--------|-------------|------|
| `power_w` | Current power consumption | Watts |
| `energy_wh` | Cumulative energy consumption | Watt-hours |
| `voltage_v` | Supply voltage | Volts |
| `current_a` | Current draw | Amperes |
| `is_on` | Device on/off state | Boolean |
| `signal_strength` | WiFi signal strength | dBm |
| `temperature` | Device temperature (if available) | Celsius |

## MQTT Topics

Energy data is published to MQTT topics in this format:
```
tapo/{device_id}/energy
```

Example payload:
```json
{
  "device_id": "tapo_living_room_1",
  "device_name": "Living Room Lamp",
  "room_id": "living_room",
  "power_w": 12.5,
  "energy_wh": 145.2,
  "is_on": true,
  "signal_strength": -45,
  "timestamp": 1642684800
}
```

## Grafana Dashboard Panels

The included dashboard provides:

1. **Real-time Power Consumption**: Line chart showing current power draw
2. **Current Device Status**: Table with device states and latest readings
3. **Cumulative Energy Consumption**: Stacked area chart of energy usage
4. **Total Power Draw**: Gauge showing total power across all devices
5. **Device Status**: Color-coded on/off indicators
6. **Power Distribution**: Pie chart showing power distribution by device
7. **Today's Total Energy**: Total energy consumed today

## API Endpoints

When integrated with the main home automation service:

### Get Device Status
```http
GET /api/tapo/devices
```

### Control Device
```http
POST /api/tapo/devices/{device_id}/state
Content-Type: application/json

{
  "on": true
}
```

### Get Energy Data
```http
GET /api/tapo/devices/{device_id}/energy?timeRange=24h
```

## Troubleshooting

### Common Issues

**Device Not Connecting**
- Verify IP address is correct
- Check username/password
- Ensure device is on same network
- Try restarting the device

**No Data in InfluxDB**
- Check InfluxDB connection
- Verify bucket and organization names
- Check service logs for errors

**MQTT Not Working**
- Verify MQTT broker is running
- Check connection settings
- Test with MQTT client (mosquitto_sub)

### Debug Mode

Enable debug logging:
```yaml
logging:
  level: debug
```

### Check Device Connectivity
```bash
# Test device connectivity
ping 192.168.1.100

# Test HTTP endpoint
curl -X POST http://192.168.1.100/app \
  -H "Content-Type: application/json" \
  -d '{"method":"get_device_info","params":{}}'
```

## Energy Monitoring Best Practices

1. **Polling Frequency**: Don't poll too frequently (minimum 30s recommended)
2. **Network Stability**: Ensure stable WiFi connection for devices
3. **Data Retention**: Configure InfluxDB retention policies for historical data
4. **Alerting**: Set up alerts for unusual power consumption patterns
5. **Regular Monitoring**: Check device status and connectivity regularly

## Integration with Home Assistant

To integrate with Home Assistant, use MQTT discovery:

```yaml
# configuration.yaml
mqtt:
  sensor:
    - name: "Tapo Living Room Lamp Power"
      state_topic: "tapo/tapo_living_room_1/energy"
      value_template: "{{ value_json.power_w }}"
      unit_of_measurement: "W"
      device_class: "power"
```

## Security Considerations

- Store credentials securely (use environment variables)
- Use strong passwords for Tapo accounts
- Consider network segmentation for IoT devices
- Regularly update device firmware
- Monitor for unusual activity patterns

## Performance Optimization

- Adjust polling intervals based on usage patterns
- Use connection pooling for multiple devices
- Implement proper error handling and retries
- Monitor resource usage (CPU, memory, network)

## Support

For issues and questions:
1. Check the troubleshooting section
2. Review service logs
3. Test device connectivity
4. Verify configuration settings
