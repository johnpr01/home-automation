# Pi Pico WH + SHT-30 Sensor Firmware

This directory contains MicroPython firmware for the Raspberry Pi Pico WH with an SHT-30 temperature and humidity sensor, PIR motion sensor, and photo transistor light sensor.

## 📋 Hardware Requirements

- **Raspberry Pi Pico WH** (with WiFi capability)
- **SHT-30** temperature and humidity sensor
- **PIR Motion Sensor** (HC-SR501 or similar)
- **Photo Transistor** or LDR for light sensing
- Jumper wires for connections
- Breadboard (optional but recommended)

## 🔌 Wiring Diagram

### SHT-30 Temperature/Humidity Sensor
| SHT-30 Pin | Pico WH Pin | Description |
|------------|-------------|-------------|
| VDD        | 3V3 (Pin 36)| Power supply |
| GND        | GND (Pin 38)| Ground |
| SDA        | GPIO4 (Pin 6)| I2C Data |
| SCL        | GPIO5 (Pin 7)| I2C Clock |

### PIR Motion Sensor
| PIR Pin    | Pico WH Pin | Description |
|------------|-------------|-------------|
| VCC        | VBUS (Pin 40)| 5V Power supply |
| GND        | GND (Pin 3) | Ground |
| OUT        | GPIO15 (Pin 20)| Digital signal |

### Photo Transistor/LDR Light Sensor
| Sensor Pin | Pico WH Pin | Description |
|------------|-------------|-------------|
| VCC        | 3V3 (Pin 36)| Power supply |
| GND        | GND (Pin 13)| Ground |
| Signal     | GPIO26 (Pin 31)| Analog input (ADC0) |

## 🚀 Flashing Instructions

### Method 1: Complete Setup (Recommended for beginners)

#### Step 1: Install MicroPython Firmware

1. **Download MicroPython:**
   ```bash
   # Download the latest stable MicroPython firmware
   wget https://micropython.org/resources/firmware/rp2-pico-w-20240222-v1.22.2.uf2
   ```
   
   Or visit [micropython.org/download/rp2-pico-w/](https://micropython.org/download/rp2-pico-w/) for the latest version.

2. **Enter Bootloader Mode:**
   - Hold the **BOOTSEL** button on the Pico
   - While holding BOOTSEL, connect the Pico to your computer via USB
   - Release the BOOTSEL button
   - The Pico should appear as a USB drive called **RPI-RP2**

3. **Flash MicroPython:**
   ```bash
   # Copy the firmware file to the Pico
   cp rp2-pico-w-*.uf2 /media/RPI-RP2/
   
   # Or on Windows, drag and drop the .uf2 file to the RPI-RP2 drive
   ```
   
   The Pico will automatically reboot and disconnect/reconnect.

#### Step 2: Install Required Tools

```bash
# Install mpremote for file management
pip install mpremote

# Verify installation
mpremote --help

# Test connection to Pico
mpremote ls
```

#### Step 3: Prepare Application Files

```bash
# Navigate to firmware directory
cd firmware/pico-sht30

# Create configuration from template
cp config_template.py config.py

# Edit configuration with your settings
nano config.py  # or use your preferred editor
```

#### Step 4: Configure Settings

Edit `config.py` with your specific settings:

```python
# WiFi Configuration
WIFI_SSID = "YourWiFiNetwork"
WIFI_PASSWORD = "YourWiFiPassword"

# MQTT Configuration  
MQTT_BROKER = "192.168.1.100"  # Your Raspberry Pi 5 IP
MQTT_PORT = 1883
MQTT_USERNAME = ""  # Optional
MQTT_PASSWORD = ""  # Optional

# Device Configuration
ROOM_NUMBER = 1  # Unique room identifier
DEVICE_ID = "pico-sensor-001"

# Sensor Configuration
TEMP_HUMIDITY_ENABLED = True
MOTION_SENSOR_ENABLED = True
LIGHT_SENSOR_ENABLED = True

# GPIO Pin Configuration (modify if needed)
I2C_SDA_PIN = 4
I2C_SCL_PIN = 5
PIR_PIN = 15
LIGHT_SENSOR_PIN = 26

# Reading Intervals (seconds)
SENSOR_READ_INTERVAL = 30
MOTION_CHECK_INTERVAL = 1
```

#### Step 5: Upload Application Files

```bash
# Upload all required files
mpremote cp config.py :config.py
mpremote cp main.py :main.py
mpremote cp sht30.py :sht30.py

# Verify files were uploaded
mpremote ls

# Expected output:
# config.py
# main.py  
# sht30.py
# boot.py (if present)
```

#### Step 6: Test the Installation

```bash
# Connect to the Pico serial console
mpremote

# You should see output like:
# Starting Pico Multi-Sensor...
# Connecting to WiFi: YourWiFiNetwork
# WiFi connected! IP: 192.168.1.XXX
# Connecting to MQTT broker: 192.168.1.100:1883
# MQTT connected successfully
# Publishing sensor data...
```

Press `Ctrl+C` to exit the serial console.

### Method 2: Advanced Setup (Using Thonny IDE)

#### Step 1: Install Thonny IDE

```bash
# Ubuntu/Debian
sudo apt install thonny

# Or download from https://thonny.org/
```

#### Step 2: Configure Thonny for Pico

1. Open Thonny IDE
2. Go to **Tools** → **Options** → **Interpreter**
3. Select **MicroPython (Raspberry Pi Pico)**
4. Choose the correct port (usually `/dev/ttyACM0` on Linux)
5. Click **Install or update MicroPython**
6. Select your Pico device and install the latest firmware

#### Step 3: Upload Files via Thonny

1. Open each file (`main.py`, `sht30.py`, `config.py`) in Thonny
2. Use **File** → **Save As** → **Raspberry Pi Pico**
3. Save each file with the same name on the Pico

### Method 3: Automated Deployment Script

#### Step 1: Use the Deploy Script

```bash
# Make the deploy script executable
chmod +x deploy.sh

# Run automated deployment
./deploy.sh

# The script will:
# 1. Check for required tools
# 2. Verify Pico connection
# 3. Upload all files
# 4. Test the connection
# 5. Display monitoring commands
```

#### Step 2: Deploy Script Contents

```bash
#!/bin/bash
# deploy.sh - Automated Pico deployment script

set -e

echo "🚀 Pi Pico Multi-Sensor Deployment Script"
echo "========================================="

# Check if mpremote is installed
if ! command -v mpremote &> /dev/null; then
    echo "❌ mpremote not found. Installing..."
    pip install mpremote
fi

# Check if config.py exists
if [ ! -f "config.py" ]; then
    echo "⚠️  config.py not found. Creating from template..."
    cp config_template.py config.py
    echo "📝 Please edit config.py with your settings before continuing."
    echo "   Required: WiFi credentials and MQTT broker IP"
    read -p "Press Enter when ready to continue..."
fi

# Test Pico connection
echo "🔍 Testing Pico connection..."
if ! mpremote ls &> /dev/null; then
    echo "❌ Cannot connect to Pico. Please check:"
    echo "   - Pico is connected via USB"
    echo "   - MicroPython is installed"
    echo "   - No other programs are using the serial port"
    exit 1
fi

echo "✅ Pico connection verified"

# Upload files
echo "📤 Uploading application files..."
mpremote cp config.py :config.py
mpremote cp main.py :main.py  
mpremote cp sht30.py :sht30.py

echo "✅ Files uploaded successfully"

# Verify upload
echo "🔍 Verifying uploaded files..."
mpremote ls

# Test run
echo "🧪 Testing application startup..."
timeout 10s mpremote run main.py || true

echo ""
echo "🎉 Deployment complete!"
echo ""
echo "📊 Monitoring commands:"
echo "   Serial output: mpremote"
echo "   MQTT messages: mosquitto_sub -h YOUR_BROKER_IP -t 'room-+/+'"
echo "   File list: mpremote ls"
echo ""
echo "🔧 Troubleshooting:"
echo "   Reset Pico: mpremote reset"
echo "   Delete files: mpremote rm config.py main.py sht30.py"
echo "   Re-upload: ./deploy.sh"
```

## 🔧 Development Workflow

### Making Code Changes

```bash
# Edit code locally
nano main.py

# Upload changes
mpremote cp main.py :main.py

# Restart the application
mpremote reset

# Monitor output
mpremote
```

### Debugging

```bash
# View detailed logs
mpremote

# Run specific commands
mpremote exec "import os; print(os.listdir())"

# Test WiFi connection
mpremote exec "
import network
wlan = network.WLAN(network.STA_IF)
print('WiFi status:', wlan.isconnected())
print('IP address:', wlan.ifconfig()[0] if wlan.isconnected() else 'Not connected')
"

# Test sensor readings
mpremote exec "
from sht30 import SHT30
from machine import I2C, Pin
i2c = I2C(0, sda=Pin(4), scl=Pin(5), freq=400000)
sensor = SHT30(i2c)
temp, hum = sensor.read_data()
print(f'Temperature: {temp}°C, Humidity: {hum}%')
"
```

### Updating Configuration

```bash
# Download current config
mpremote cp :config.py config_backup.py

# Edit locally
nano config.py

# Upload updated config
mpremote cp config.py :config.py

# Restart to apply changes
mpremote reset
```

## 📊 MQTT Topics

The unified sensor publishes data to the following topics:

- **Temperature**: `room-temp/{room_number}` (°F)
- **Humidity**: `room-hum/{room_number}` (%)
- **Motion**: `room-motion/{room_number}` (0/1 occupancy)
- **Light**: `room-light/{room_number}` (0-100% brightness)

Where `{room_number}` is configured in `config.py`.

### Example MQTT Messages

```json
// Temperature (°F)
{
  "temperature": 72.5,
  "timestamp": "2024-07-16T14:30:00",
  "room": 1,
  "device": "pico-sensor-001"
}

// Motion Detection  
{
  "motion": 1,
  "occupancy": true,
  "timestamp": "2024-07-16T14:30:15",
  "room": 1,
  "device": "pico-sensor-001"
}

// Light Level
{
  "light_level": 45.2,
  "timestamp": "2024-07-16T14:30:30", 
  "room": 1,
  "device": "pico-sensor-001"
}
```

## 📈 Monitoring & Debugging

### Real-time Serial Monitoring

```bash
# Connect to serial console for live debugging
mpremote

# Expected output:
# Starting Pico Multi-Sensor...
# WiFi: Connecting to 'YourNetwork'...
# WiFi: Connected! IP: 192.168.1.123
# MQTT: Connecting to 192.168.1.100:1883...
# MQTT: Connected successfully
# Sensors: Initializing...
# SHT30: OK, PIR: OK, Light: OK
# Publishing sensor data every 30s...
# [14:30:00] Temp: 22.5°C (72.5°F), Humidity: 45%
# [14:30:01] Motion: No movement
# [14:30:02] Light: 45% (daylight)
```

### MQTT Message Monitoring

```bash
# Monitor all sensor messages
mosquitto_sub -h YOUR_BROKER_IP -t "room-+/+"

# Monitor specific room
mosquitto_sub -h YOUR_BROKER_IP -t "room-temp/1" -t "room-hum/1" -t "room-motion/1" -t "room-light/1"

# Monitor with verbose output
mosquitto_sub -h YOUR_BROKER_IP -t "room-+/+" -v

# Save messages to file
mosquitto_sub -h YOUR_BROKER_IP -t "room-+/+" | tee sensor_data.log
```

### Advanced Debugging Commands

```bash
# Check Pico system info
mpremote exec "
import os, gc, machine
print('Files:', os.listdir())
print('Free memory:', gc.mem_free())
print('CPU temp:', machine.temperature())
print('Unique ID:', machine.unique_id().hex())
"

# Test individual sensors
mpremote exec "
from machine import Pin, I2C, ADC
import time

# Test SHT30
i2c = I2C(0, sda=Pin(4), scl=Pin(5))
print('I2C devices:', [hex(addr) for addr in i2c.scan()])

# Test PIR
pir = Pin(15, Pin.IN)
print('PIR state:', pir.value())

# Test light sensor
light = ADC(Pin(26))
print('Light raw:', light.read_u16())
print('Light %:', (light.read_u16() / 65535) * 100)
"

# Network diagnostics
mpremote exec "
import network
wlan = network.WLAN(network.STA_IF)
if wlan.isconnected():
    config = wlan.ifconfig()
    print('IP:', config[0])
    print('Subnet:', config[1]) 
    print('Gateway:', config[2])
    print('DNS:', config[3])
    print('RSSI:', wlan.status('rssi'), 'dBm')
else:
    print('WiFi not connected')
"
```

## 🔧 Configuration Reference

### Complete `config.py` Example

```python
# WiFi Configuration
WIFI_SSID = "HomeNetwork"
WIFI_PASSWORD = "YourSecurePassword"
WIFI_TIMEOUT = 30  # Connection timeout in seconds

# MQTT Configuration
MQTT_BROKER = "192.168.1.100"  # Raspberry Pi 5 IP
MQTT_PORT = 1883
MQTT_USERNAME = ""  # Leave empty if no auth
MQTT_PASSWORD = ""  # Leave empty if no auth
MQTT_TIMEOUT = 10
MQTT_KEEPALIVE = 60

# Device Configuration
ROOM_NUMBER = 1
DEVICE_ID = "pico-sensor-001"
DEVICE_NAME = "Living Room Sensors"

# Sensor Enable/Disable
TEMP_HUMIDITY_ENABLED = True
MOTION_SENSOR_ENABLED = True  
LIGHT_SENSOR_ENABLED = True

# GPIO Pin Configuration
I2C_SDA_PIN = 4          # SHT30 data pin
I2C_SCL_PIN = 5          # SHT30 clock pin
PIR_PIN = 15             # Motion sensor pin
LIGHT_SENSOR_PIN = 26    # Light sensor ADC pin (ADC0)

# Timing Configuration
SENSOR_READ_INTERVAL = 30    # Temperature/humidity reading interval
MOTION_CHECK_INTERVAL = 1    # Motion detection check interval
LIGHT_READ_INTERVAL = 5      # Light level reading interval
WATCHDOG_TIMEOUT = 8000      # System watchdog timeout (ms)

# Sensor Calibration
TEMP_OFFSET = 0.0           # Temperature calibration offset (°C)
HUMIDITY_OFFSET = 0.0       # Humidity calibration offset (%)
LIGHT_CALIBRATION = 1.0     # Light sensor calibration multiplier

# Motion Detection Settings
MOTION_SENSITIVITY = 1       # PIR sensitivity (hardware dependent)
MOTION_TIMEOUT = 30         # Seconds before motion resets to 0

# MQTT Topic Templates
TEMP_TOPIC = "room-temp/{room}"
HUMIDITY_TOPIC = "room-hum/{room}"
MOTION_TOPIC = "room-motion/{room}"  
LIGHT_TOPIC = "room-light/{room}"

# Debug Configuration
DEBUG_MODE = True           # Enable detailed logging
LED_ENABLED = True          # Use onboard LED for status
SERIAL_BAUD = 115200       # Serial communication speed
```

## 🔍 Troubleshooting Guide

### Common Issues and Solutions

#### 1. **Cannot Connect to Pico**

**Symptoms:** `mpremote ls` fails or no device found

**Solutions:**
```bash
# Check USB connection
lsusb | grep -i pico

# Check serial ports
ls /dev/ttyACM*

# Try different USB port or cable
# Ensure no other programs are using the serial port
sudo fuser -v /dev/ttyACM0

# Reset the Pico by disconnecting/reconnecting
# Or use the reset button if available
```

#### 2. **WiFi Connection Problems**

**Symptoms:** "WiFi connection failed" in serial output

**Solutions:**
```bash
# Verify network credentials in config.py
# Check 2.4GHz network (Pico WH doesn't support 5GHz)
# Test WiFi range and signal strength

# Debug WiFi connection
mpremote exec "
import network
wlan = network.WLAN(network.STA_IF)
wlan.active(True)
networks = wlan.scan()
for net in networks:
    print('SSID:', net[0].decode(), 'Signal:', net[3])
"
```

#### 3. **MQTT Connection Issues**

**Symptoms:** "MQTT connection failed" or timeout errors

**Solutions:**
```bash
# Test MQTT broker connectivity from your computer
mosquitto_pub -h YOUR_BROKER_IP -t "test" -m "hello"

# Check Raspberry Pi 5 firewall
sudo ufw status
sudo ufw allow 1883

# Verify MQTT broker is running
docker compose ps | grep mqtt

# Test from Pico
mpremote exec "
import socket
try:
    s = socket.socket()
    s.settimeout(5)
    s.connect(('192.168.1.100', 1883))
    print('MQTT port reachable')
    s.close()
except Exception as e:
    print('MQTT connection error:', e)
"
```

#### 4. **Sensor Reading Problems**

**Symptoms:** Incorrect readings or sensor errors

**Solutions:**
```bash
# Check SHT30 I2C connection
mpremote exec "
from machine import I2C, Pin
i2c = I2C(0, sda=Pin(4), scl=Pin(5), freq=400000)
devices = i2c.scan()
print('I2C devices found:', [hex(d) for d in devices])
# Should show [0x44] for SHT30
"

# Test PIR sensor
mpremote exec "
from machine import Pin
import time
pir = Pin(15, Pin.IN)
for i in range(10):
    print('PIR reading:', pir.value())
    time.sleep(1)
"

# Test light sensor
mpremote exec "
from machine import ADC, Pin
light = ADC(Pin(26))
raw = light.read_u16()
percent = (raw / 65535) * 100
print(f'Light: {raw} raw, {percent:.1f}%')
"
```

#### 5. **Memory Issues**

**Symptoms:** "Out of memory" errors or random reboots

**Solutions:**
```bash
# Check memory usage
mpremote exec "
import gc
print('Free memory:', gc.mem_free())
print('Allocated memory:', gc.mem_alloc())
gc.collect()
print('After cleanup:', gc.mem_free())
"

# Optimize configuration
# - Increase SENSOR_READ_INTERVAL
# - Disable unused sensors
# - Reduce debug output
```

#### 6. **File Upload Issues**

**Symptoms:** "No space left" or upload failures

**Solutions:**
```bash
# Check available space
mpremote exec "
import os
stat = os.statvfs('/')
free = stat[0] * stat[3]
total = stat[0] * stat[2]
print(f'Storage: {free} bytes free of {total} total')
"

# Clean up old files
mpremote rm old_file.py

# Remove all files and start fresh
mpremote exec "
import os
for f in os.listdir():
    if f not in ['boot.py']:
        os.remove(f)
"
```

### Performance Optimization

#### Reduce Power Consumption
```python
# Add to config.py
DEEP_SLEEP_ENABLED = True
SLEEP_DURATION = 300  # 5 minutes between readings

# Disable unused features
LED_ENABLED = False
DEBUG_MODE = False
```

#### Improve Reliability
```python
# Add watchdog timer
WATCHDOG_ENABLED = True
WATCHDOG_TIMEOUT = 8000  # 8 seconds

# Add error recovery
AUTO_RESTART_ON_ERROR = True
MAX_ERROR_COUNT = 5
```

## 📁 File Structure

```
firmware/pico-sht30/
├── README.md              # This comprehensive guide
├── main.py                # Main application (multi-sensor)
├── sht30.py              # SHT-30 sensor driver
├── config_template.py    # Configuration template
├── config.py             # Your configuration (created from template)
├── deploy.sh             # Automated deployment script
├── MOTION_SENSOR.md      # PIR sensor setup guide
├── LIGHT_SENSOR.md       # Light sensor setup guide
└── examples/             # Example configurations
    ├── single_room.py    # Single room setup
    ├── multi_room.py     # Multiple room deployment
    └── advanced.py       # Advanced configuration
```

## 🎯 Next Steps

After successful deployment:

1. **Monitor sensor data** in your home automation dashboard
2. **Set up automation rules** using the motion and light sensors
3. **Scale to multiple rooms** by deploying additional Pico units
4. **Integrate with smart home systems** via MQTT
5. **Create custom automation** using the sensor data streams

For advanced usage, see the main project [README.md](../../README.md) for integration with the complete home automation system.
