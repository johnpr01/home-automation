# PIR Motion Detection Service

## ✅ Orthogonal Motion Detection Service

Your home automation system now includes a **standalone PIR (Passive Infrared) motion detection service** that operates independently from the smart thermostat system. This modular design ensures better separation of concerns and allows each service to evolve independently.

### 🏗️ **Architecture Overview**

**Service Independence:**
- **Motion Service**: Dedicated service for PIR sensor monitoring and room occupancy tracking
- **Thermostat Service**: Focused on temperature control and HVAC management
- **Optional Integration**: Services can communicate via callbacks when needed

**Key Benefits:**
- **Modular Design**: Each service has a single responsibility
- **Independent Scaling**: Services can be deployed and scaled separately
- **Flexible Integration**: Choose how/when services communicate
- **Better Testing**: Isolated services are easier to test

### 🔧 **Hardware Integration**

**Supported Hardware:**
- **Pi Pico WH**: WiFi-enabled microcontroller
- **SHT-30 Sensor**: Temperature/humidity (I2C)
- **PIR Sensor**: Motion detection (HC-SR501 or similar)

**Wiring Setup:**
```
Pi Pico WH GPIO 2 → PIR Sensor OUT pin
Pi Pico WH 3.3V   → PIR Sensor VCC
Pi Pico WH GND    → PIR Sensor GND
```

### 📡 **MQTT Integration**

**New Topic Added:**
- `room-motion/{room_number}` - Motion detection events

**Motion Payload Format:**
```json
{
  "motion": true,
  "room": "1",
  "sensor": "PIR",
  "timestamp": 1640995200,
  "motion_start": 1640995195,
  "device_id": "pico-living-room"
}
```

### 🔧 **Configuration Options**

**Pi Pico Configuration (`config.py`):**
```python
PIR_SENSOR_PIN = 2        # GPIO pin for PIR data
PIR_ENABLED = True        # Enable/disable motion detection
PIR_DEBOUNCE_TIME = 2     # Seconds between detections
PIR_TIMEOUT = 30          # Seconds before motion clears
```

### 🏠 **Motion Service Features**

**Independent Operation:**
- Processes `room-motion/+` MQTT messages
- Tracks room occupancy status for all rooms
- Maintains sensor connectivity (IsOnline flag)
- Provides occupancy callbacks for integration
- Logs motion events with detailed context

**Room Occupancy API:**
```go
// Get specific room occupancy
occupancy, exists := motionService.GetRoomOccupancy("1")

// Get all room occupancy data
allRooms := motionService.GetAllOccupancy()

// Add callback for occupancy changes
motionService.AddOccupancyCallback(func(roomID string, occupied bool) {
    log.Printf("Room %s occupancy changed: %v", roomID, occupied)
})
```

### 🚨 **Motion Detection Logic**

**State Management:**
1. **Motion Detected**: PIR HIGH → Immediate MQTT publish
2. **Motion Ongoing**: Continuous monitoring, updates timestamp
3. **Motion Cleared**: No motion for 30 seconds → Publishes motion=false
4. **Debouncing**: 2-second filter prevents rapid triggering

**LED Status Indicators:**
- **Single blink**: Temperature/humidity reading
- **Double quick blinks**: Motion detected
- **Triple blinks**: MQTT error
- **Five blinks**: Sensor error

### 📊 **Enhanced Monitoring**

**Enhanced Log Structure:**
```json
{
  "timestamp": "2025-07-16T10:30:15Z",
  "level": "INFO",
  "service": "MotionService", 
  "message": "Motion DETECTED in room 1 (device: pico-living-room)",
  "room_id": "1",
  "motion_detected": true,
  "device_id": "pico-living-room"
}
```

**MQTT Monitoring:**
```bash
# Monitor all motion events
mosquitto_sub -h YOUR_PI5_IP -t "room-motion/+"

# Test motion detection
mosquitto_pub -h YOUR_PI5_IP -t "room-motion/1" \
  -m '{"motion":true,"room":"1","sensor":"PIR","timestamp":1640995200}'
```

### 🔮 **Service Integration Examples**

**Optional Integration:**
- **Occupancy-Based HVAC**: Thermostat service can subscribe to motion callbacks
- **Energy Optimization**: Reduce heating/cooling in unoccupied rooms
- **Security Integration**: Motion alerts for unauthorized access
- **Smart Scheduling**: Occupancy-aware thermostat schedules
- **Analytics**: Room usage patterns and HVAC optimization

**Integration Code Example:**
```go
// Optional: Connect motion service to thermostat service
motionService.AddOccupancyCallback(func(roomID string, occupied bool) {
    // Adjust thermostat behavior based on occupancy
    if thermostat := thermostatService.GetThermostatByRoom(roomID); thermostat != nil {
        // Implement occupancy-based logic here
        log.Printf("Room %s occupancy change affects thermostat %s", roomID, thermostat.ID)
    }
})
```

### 🛠️ **Files Created/Modified**

**New Motion Service:**
- `internal/services/motion_service.go` - Standalone motion detection service
- `cmd/motion/main.go` - Dedicated motion service executable
- `cmd/integrated/main.go` - Example integrated service (optional)

**Thermostat Service (Motion Removed):**
- `internal/services/thermostat_service.go` - Motion handling removed, focused on HVAC

**Firmware (Unchanged):**
- `firmware/pico-sht30/config_template.py` - PIR configuration
- `firmware/pico-sht30/main.py` - Motion detection and MQTT publishing
- `firmware/pico-sht30/MOTION_SENSOR.md` - Hardware setup guide

### 🚀 **Getting Started**

**Option 1: Run Motion Service Independently**
```bash
# Start standalone motion detection service
cd cmd/motion && go run main.go
```

**Option 2: Run Thermostat Service Independently**
```bash
# Start standalone thermostat service
cd cmd/thermostat && go run main.go
```

**Option 3: Run Integrated Service (Optional)**
```bash
# Start both services with optional integration
cd cmd/integrated && go run main.go
```

**Hardware Setup:**
1. **Update Pi Pico configuration:**
   ```python
   PIR_ENABLED = True
   PIR_SENSOR_PIN = 2
   ```

2. **Wire PIR sensor** to GPIO 2 on Pi Pico WH

3. **Deploy updated firmware** to Pi Pico

4. **Monitor motion events:**
   ```bash
   # Motion service logs
   mosquitto_sub -h YOUR_PI5_IP -t "room-motion/+"
   ```

**Your smart home now has modular, orthogonal motion detection and thermostat services! 🏠👁️🌡️**
