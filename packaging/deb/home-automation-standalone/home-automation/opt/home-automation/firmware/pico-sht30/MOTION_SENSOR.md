# Raspberry Pi Pico WH Motion Detection README

## PIR Motion Sensor Integration

The firmware now supports PIR (Passive Infrared) motion sensors in addition to temperature and humidity monitoring.

### Hardware Setup

**Required Components:**
- Raspberry Pi Pico WH (with WiFi)
- SHT-30 Temperature/Humidity sensor (I2C)
- PIR Motion Sensor (HC-SR501 or similar)
- Breadboard and jumper wires

**Wiring Connections:**

```
Pi Pico WH          | SHT-30 Sensor    | PIR Sensor
-------------------|------------------|------------------
GPIO 4 (Pin 6)     | SDA              | -
GPIO 5 (Pin 7)     | SCL              | -
GPIO 2 (Pin 4)     | -                | OUT (Data)
3.3V (Pin 36)      | VCC              | VCC
GND (Pin 38)       | GND              | GND
```

### Configuration

Edit `config.py` to enable motion detection:

```python
# PIR Motion Sensor Configuration
PIR_SENSOR_PIN = 2        # GPIO pin for PIR sensor data
PIR_ENABLED = True        # Set to False to disable PIR sensor
PIR_DEBOUNCE_TIME = 2     # Seconds between motion detections
PIR_TIMEOUT = 30          # Seconds before motion is cleared
```

### MQTT Topics

The sensor publishes to these topics:

- **Temperature**: `room-temp/{room_number}` (°F)
- **Humidity**: `room-hum/{room_number}` (%)
- **Motion**: `room-motion/{room_number}` (boolean)

### Motion Detection Payload

```json
{
  "motion": true,
  "room": "1",
  "sensor": "PIR", 
  "timestamp": 1640995200,
  "motion_start": 1640995195,
  "device_id": "pico-sht30-room1"
}
```

### LED Status Indicators

- **Single blink**: Successful temperature/humidity reading
- **Double quick blinks**: Motion detected
- **Triple blinks**: MQTT error
- **Five blinks**: Sensor error

### Motion Detection Logic

1. **Motion Detected**: PIR sensor outputs HIGH
   - Immediately publishes motion=true
   - Updates motion_start timestamp
   - Continues monitoring

2. **Motion Cleared**: No motion for PIR_TIMEOUT seconds
   - Publishes motion=false
   - Resets motion state

3. **Debouncing**: Prevents rapid on/off switching
   - Uses PIR_DEBOUNCE_TIME to filter noise

### Thermostat Integration

The smart thermostat service processes motion events to:

- Update room occupancy status
- Maintain sensor connectivity (IsOnline flag)
- Log motion activity for each room
- Future: Occupancy-based temperature control

### Usage Example

1. **Deploy sensor in living room:**
   ```python
   ROOM_NUMBER = "1"
   PIR_ENABLED = True
   PIR_SENSOR_PIN = 2
   ```

2. **Monitor MQTT messages:**
   ```bash
   mosquitto_sub -h YOUR_PI5_IP -t "room-motion/+"
   ```

3. **View thermostat logs:**
   ```bash
   cd cmd/thermostat && go run main.go
   # Shows motion detection events:
   # Motion DETECTED in room 1 (thermostat: thermostat-001)
   # Motion CLEARED in room 1 (thermostat: thermostat-001)
   ```

### Troubleshooting

**PIR Not Working:**
- Check wiring connections
- Verify PIR_SENSOR_PIN matches hardware
- Ensure PIR_ENABLED = True
- Test PIR sensor with multimeter

**False Triggers:**
- Increase PIR_DEBOUNCE_TIME
- Check for heat sources near sensor
- Adjust PIR sensor sensitivity (potentiometer)

**Motion Not Clearing:**
- Verify PIR_TIMEOUT setting
- Check for continuous motion in room
- Ensure stable power supply

**MQTT Issues:**
- Verify network connectivity
- Check MQTT broker address
- Monitor for error messages in logs

This motion detection system provides real-time occupancy monitoring for your smart home automation system! 🏠👁️
