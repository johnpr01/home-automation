# Light Sensor Service Documentation

## ✅ Photo Transistor Light Sensor Service

Your home automation system now includes a **standalone light sensor service** that operates independently using photo transistors for ambient light monitoring. This modular design provides comprehensive lighting intelligence for automation and energy optimization.

### 🏗️ **Architecture Overview**

**Service Independence:**
- **Light Sensor Service**: Dedicated service for photo transistor monitoring and ambient light tracking
- **Motion Service**: PIR sensor monitoring and room occupancy (separate)
- **Thermostat Service**: HVAC temperature control (separate)
- **Optional Integration**: Services can communicate via callbacks for advanced automation

### 🔧 **Hardware Integration**

**Supported Hardware:**
- **Pi Pico WH**: WiFi-enabled microcontroller
- **Photo Transistor**: Ambient light sensor (LTR-3208E, BPW85, TEMT6000, etc.)
- **SHT-30 Sensor**: Temperature/humidity (I2C, existing)
- **PIR Sensor**: Motion detection (existing)

**Wiring Setup:**
```
Pi Pico WH GPIO 28 (ADC2) → Photo Transistor Emitter
Pi Pico WH 3.3V           → Photo Transistor Collector  
Pi Pico WH GND            → 10kΩ Resistor → GPIO 28
```

### 📡 **MQTT Integration**

**New Topic Added:**
- `room-light/{room_number}` - Light level monitoring events

**Light Sensor Payload Format:**
```json
{
  "light_level": 45.2,
  "light_percent": 45.2,
  "light_state": "normal",
  "unit": "%",
  "room": "1",
  "sensor": "PhotoTransistor",
  "timestamp": 1640995200,
  "device_id": "pico-living-room"
}
```

### 🔧 **Configuration Options**

**Pi Pico Configuration (`config.py`):**
```python
LIGHT_SENSOR_PIN = 28         # GPIO pin for photo transistor (ADC2)
LIGHT_ENABLED = True          # Enable/disable light sensor
LIGHT_THRESHOLD_LOW = 10      # Percentage below which it's "dark" (0-100)
LIGHT_THRESHOLD_HIGH = 80     # Percentage above which it's "bright" (0-100)
LIGHT_READING_INTERVAL = 10   # Seconds between light readings
```

### 🏠 **Light Service Features**

**Independent Operation:**
- Processes `room-light/+` MQTT messages
- Tracks ambient light levels for all rooms
- Maintains sensor connectivity (IsOnline flag)
- Provides light level callbacks for integration
- Logs light state changes with detailed context
- Day/night cycle detection

**Room Light Level API:**
```go
// Get specific room light level
lightLevel, exists := lightService.GetRoomLightLevel("1")

// Get all room light data
allRooms := lightService.GetAllLightLevels()

// Add callback for light level changes
lightService.AddLightCallback(func(roomID string, lightState string, lightLevel float64) {
    log.Printf("Room %s light changed: %s (%.1f%%)", roomID, lightState, lightLevel)
})

// Set custom thresholds
lightService.SetThresholds(15.0, 75.0)  // dark < 15%, bright > 75%
```

### 🌞 **Light Detection Logic**

**State Management:**
1. **Dark State**: Light level below configured threshold (< 10%)
2. **Normal State**: Light level between thresholds (10-80%)
3. **Bright State**: Light level above configured threshold (> 80%)
4. **Day/Night Cycle**: Automatic detection based on patterns and time

**LED Status Indicators:**
- **Single blink**: Temperature/humidity reading
- **Double quick blinks**: Motion detected
- **Triple blinks**: MQTT error
- **Five blinks**: Sensor error

### 📊 **Enhanced Monitoring**

**Light Sensor Log Structure:**
```json
{
  "timestamp": "2025-07-16T10:30:15Z",
  "level": "INFO",
  "service": "LightService", 
  "message": "Room 1 light state changed: normal -> dark (8.5%)",
  "room_id": "1",
  "light_level": 8.5,
  "light_state": "dark",
  "device_id": "pico-living-room"
}
```

**MQTT Monitoring:**
```bash
# Monitor all light level events
mosquitto_sub -h YOUR_PI5_IP -t "room-light/+"

# Test light sensor
mosquitto_pub -h YOUR_PI5_IP -t "room-light/1" \
  -m '{"light_level":25.5,"light_state":"normal","sensor":"PhotoTransistor","timestamp":1640995200}'
```

### 🔮 **Service Integration Examples**

**Automatic Lighting Control:**
```go
// Occupancy + Light based automation
lightService.AddLightCallback(func(roomID string, lightState string, lightLevel float64) {
    if occupancy, exists := motionService.GetRoomOccupancy(roomID); exists && occupancy.IsOccupied {
        switch lightState {
        case "dark":
            // Turn on lights when room is occupied and dark
            publishLightCommand(roomID, "turn_on", 75) // 75% brightness
        case "bright":
            // Turn off lights when room is bright (natural light)
            publishLightCommand(roomID, "turn_off", 0)
        }
    }
})
```

**Energy Optimization:**
```go
// Thermostat integration for energy saving
lightService.AddLightCallback(func(roomID string, lightState string, lightLevel float64) {
    if lightState == "dark" && time.Now().Hour() > 22 {
        // Night mode: reduce heating/cooling when it's dark and late
        if thermostat := thermostatService.GetThermostatByRoom(roomID); thermostat != nil {
            // Slightly reduce target temperature for energy savings
            adjustThermostatForNightMode(thermostat.ID)
        }
    }
})
```

**Circadian Rhythm Support:**
```go
// Day/night cycle automation
lightService.AddLightCallback(func(roomID string, lightState string, lightLevel float64) {
    lightLevel, _ := lightService.GetRoomLightLevel(roomID)
    if lightLevel.DayNightCycle == "dawn" {
        // Morning routine automation
        publishMorningScene(roomID)
    } else if lightLevel.DayNightCycle == "dusk" {
        // Evening routine automation  
        publishEveningScene(roomID)
    }
})
```

### 🛠️ **Files Created/Modified**

**New Light Sensor Service:**
- `internal/services/light_service.go` - Standalone light sensor service
- `cmd/light/main.go` - Dedicated light sensor service executable
- `cmd/integrated/main.go` - Updated to include light sensor integration

**Firmware Updates:**
- `firmware/pico-sht30/config_template.py` - Photo transistor configuration
- `firmware/pico-sht30/main.py` - Light sensor logic and MQTT publishing
- `firmware/pico-sht30/LIGHT_SENSOR.md` - Hardware setup guide

**Documentation:**
- `docs/LIGHT_SENSOR.md` - This comprehensive guide
- `README.md` - Updated with light sensor features

### 🚀 **Getting Started**

**Option 1: Run Light Service Independently**
```bash
# Start standalone light sensor service
cd cmd/light && go run main.go
```

**Option 2: Run All Services Independently**
```bash
# Terminal 1: Motion detection
cd cmd/motion && go run main.go

# Terminal 2: Light sensor
cd cmd/light && go run main.go

# Terminal 3: Thermostat control
cd cmd/thermostat && go run main.go
```

**Option 3: Run Integrated Service**
```bash
# Start all services with integration callbacks
cd cmd/integrated && go run main.go
```

**Hardware Setup:**
1. **Update Pi Pico configuration:**
   ```python
   LIGHT_ENABLED = True
   LIGHT_SENSOR_PIN = 28
   LIGHT_THRESHOLD_LOW = 10
   LIGHT_THRESHOLD_HIGH = 80
   ```

2. **Wire photo transistor** to GPIO 28 (ADC2) on Pi Pico WH

3. **Deploy updated firmware** to Pi Pico

4. **Monitor light events:**
   ```bash
   # Light service logs
   mosquitto_sub -h YOUR_PI5_IP -t "room-light/+"
   ```

### 📈 **Use Cases and Applications**

**Home Automation:**
- **Smart Lighting**: Auto on/off based on occupancy + ambient light
- **Energy Efficiency**: Reduce artificial lighting when natural light is sufficient
- **Security**: Detect unusual lighting patterns for intrusion detection
- **Mood Lighting**: Adjust color temperature based on natural light levels

**Health and Wellness:**
- **Circadian Rhythms**: Track natural light exposure for sleep optimization
- **Seasonal Affective Disorder**: Monitor light levels for health insights
- **Work Environment**: Optimize workspace lighting for productivity

**Agricultural/Greenhouse:**
- **Plant Monitoring**: Track light levels for optimal plant growth
- **Grow Light Control**: Supplement natural light when needed
- **Season Tracking**: Monitor seasonal light pattern changes

**Your smart home now has professional ambient light sensing with orthogonal service architecture! 🌞📱🏠**
