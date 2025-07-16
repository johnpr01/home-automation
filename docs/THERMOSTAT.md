# Smart Thermostat Framework

This directory contains the smart thermostat framework that integrates with Pi Pico sensors to provide intelligent temperature control.

## Overview

The smart thermostat system consists of:

1. **Pi Pico WH + SHT-30 Sensor**: Sends temperature and humidity data via MQTT
2. **Thermostat Service**: Go service that processes sensor data and controls thermostats
3. **MQTT Integration**: Real-time communication between sensors and thermostat logic
4. **Control Logic**: Intelligent heating/cooling decisions with hysteresis

## Components

### Thermostat Models (`internal/models/thermostat.go`)

- `Thermostat`: Core thermostat structure with temperature, humidity, mode, and status
- `ThermostatMode`: Operating modes (off, heat, cool, auto, fan)
- `ThermostatStatus`: Current status (idle, heating, cooling, fan)
- `ThermostatSchedule`: Time-based temperature scheduling
- `ThermostatCommand`: Commands for controlling thermostats

### Thermostat Service (`internal/services/thermostat_service.go`)

- Subscribes to MQTT sensor topics (`room-temp/+`, `room-hum/+`)
- Processes sensor readings and updates thermostat state
- Implements control logic with hysteresis to prevent short cycling
- Publishes control commands to thermostats
- Runs continuous control loop every 30 seconds

### Main Application (`cmd/thermostat/main.go`)

- Demonstrates how to set up and run the thermostat service
- Configures MQTT connection and registers thermostats
- Provides graceful shutdown handling

## MQTT Topics

### Sensor Input Topics
- `room-temp/{room_id}`: Temperature readings from Pi Pico sensors
- `room-hum/{room_id}`: Humidity readings from Pi Pico sensors

**Payload Format (from Pi Pico):**
```json
{
  "temperature": 22.5,
  "unit": "°C",
  "room": "1",
  "sensor": "SHT-30",
  "timestamp": 1640995200,
  "device_id": "pico-living-room"
}
```

### Control Output Topics
- `room-control/{room_id}`: Control commands sent to thermostats
- `thermostat/{thermostat_id}/command`: Direct thermostat commands

## Usage

### 1. Start Infrastructure

```bash
# Start MQTT broker and Kafka
cd deployments
docker-compose up -d
```

### 2. Deploy Pi Pico Firmware

Flash the MicroPython firmware to your Pi Pico WH:
```bash
# Copy firmware/pico-sht30/main.py to your Pi Pico
# Update WIFI credentials and MQTT settings in the file
```

### 3. Run Thermostat Service

```bash
# Build and run the thermostat service
cd cmd/thermostat
go build -o thermostat main.go
./thermostat
```

### 4. Register Thermostats

The service automatically registers thermostats for rooms. You can customize this in `main.go`:

```go
thermostat := &models.Thermostat{
    ID:                "thermostat-001",
    Name:              "Living Room Thermostat",
    RoomID:            "1", // Must match Pi Pico room number
    TargetTemp:        22.0,
    Mode:              models.ModeAuto,
    HeatingEnabled:    true,
    CoolingEnabled:    true,
    Hysteresis:        1.0, // Prevents short cycling
    MinTemp:           10.0,
    MaxTemp:           30.0,
}
```

## Configuration

### Thermostat Parameters

- **Hysteresis**: Temperature dead band to prevent frequent on/off cycling
- **TemperatureOffset**: Calibration offset for sensor readings
- **MinTemp/MaxTemp**: Safety limits for target temperature
- **Mode**: Operating mode (off, heat, cool, auto, fan)

### Control Logic

The thermostat uses hysteresis control:
- **Heating**: Starts when `current_temp < (target_temp - hysteresis/2)`
- **Cooling**: Starts when `current_temp > (target_temp + hysteresis/2)`
- **Stop**: When target temperature is reached

### Example Scenarios

**Heating Mode (Target: 22°C, Hysteresis: 1°C)**
- Heat turns ON when temperature drops below 21.5°C
- Heat turns OFF when temperature reaches 22°C

**Auto Mode**
- Automatically switches between heating and cooling based on temperature

## Integration with Pi Pico

The Pi Pico sensors publish data every 30 seconds by default. The thermostat service:

1. Receives sensor data via MQTT
2. Updates thermostat current temperature and humidity
3. Evaluates control logic (should heat/cool/idle)
4. Publishes control commands if status changes
5. Logs all temperature updates and control decisions

## Monitoring

The service provides detailed logging:
- Sensor data reception and parsing
- Temperature updates for each thermostat
- Control decisions (heating/cooling/idle)
- MQTT connection status and errors

## Next Steps

1. **REST API**: Add HTTP endpoints for thermostat control and monitoring
2. **Database Integration**: Store historical data and schedules
3. **Web Interface**: Create a web dashboard for thermostat management
4. **Advanced Scheduling**: Implement time-based temperature schedules
5. **Multi-Zone Control**: Support multiple heating/cooling zones
6. **Energy Optimization**: Add algorithms for energy-efficient operation

## Troubleshooting

### Common Issues

1. **MQTT Connection Failed**: Check broker address and credentials
2. **No Sensor Data**: Verify Pi Pico WiFi and MQTT configuration
3. **Thermostat Not Responding**: Check control topic subscription
4. **Temperature Oscillation**: Increase hysteresis value

### Debug Commands

```bash
# Check MQTT messages
mosquitto_sub -h localhost -t "room-temp/+"
mosquitto_sub -h localhost -t "room-hum/+"
mosquitto_sub -h localhost -t "room-control/+"

# Test Pi Pico connection
mosquitto_pub -h localhost -t "room-temp/1" -m '{"temperature":20.5,"unit":"°C","room":"1","sensor":"SHT-30","timestamp":1640995200,"device_id":"test"}'
```
