# Photo Transistor Light Sensor Setup Guide

## 🔌 Hardware Requirements

### Components Needed:
- **Pi Pico WH** (WiFi-enabled)
- **Photo Transistor** (e.g., LTR-3208E, BPW85, or similar)
- **10kΩ Resistor** (pull-down resistor)
- **Jumper Wires**
- **Breadboard** (optional)

### Recommended Photo Transistors:
- **LTR-3208E**: High sensitivity, good for indoor use
- **BPW85**: General purpose, wide spectral range
- **TEMT6000**: Small, easy to use ambient light sensor
- **Any NPN Photo Transistor**: 3-pin package (Collector, Base, Emitter)

## 🔧 Wiring Diagram

### Basic Photo Transistor Circuit:
```
Pi Pico WH                    Photo Transistor
                              (3-pin package)
                                    │
3.3V ────────────────────────── Collector (C)
                                    │
GPIO 28 (ADC2) ──────────────── Emitter (E)
                                    │
                              Base (B) ← Light sensitive
                                    │
GND ─────── 10kΩ Resistor ──────────┘
```

### Alternative: 2-Pin Photo Diode Circuit:
```
Pi Pico WH                    Photo Diode
                              (2-pin package)
                                    │
3.3V ────── 10kΩ Resistor ──── Anode (+)
                    │               │
GPIO 28 (ADC2) ─────┴─────────── Cathode (-)
                                    │
GND ────────────────────────────────┘
```

## 📋 Step-by-Step Wiring

### For 3-Pin Photo Transistor:
1. **Connect Collector (C)** → Pi Pico **3.3V** pin
2. **Connect Emitter (E)** → Pi Pico **GPIO 28** (ADC2)
3. **Connect Base (B)** → Leave open (light sensitive)
4. **Connect 10kΩ Resistor** between GPIO 28 and **GND**

### Pin Locations on Pi Pico WH:
```
                    Pi Pico WH
                   ┌─────────────┐
             GP0 1 │●           ●│ 40 VBUS
             GP1 2 │●           ●│ 39 VSYS  
             GND 3 │●           ●│ 38 GND
             GP2 4 │●           ●│ 37 3V3_EN
             GP3 5 │●           ●│ 36 3V3(OUT) ← Connect to Collector
             GP4 6 │●           ●│ 35 ADC_VREF
             GP5 7 │●           ●│ 34 GP28 ← Connect to Emitter
             GND 8 │●           ●│ 33 GND ← Connect to Resistor
             GP6 9 │●           ●│ 32 GP27
            GP7 10 │●           ●│ 31 GP26
            GP8 11 │●           ●│ 30 RUN
            GP9 12 │●           ●│ 29 GP22
           GP10 13 │●           ●│ 28 GND
           GP11 14 │●           ●│ 27 GP21
           GP12 15 │●           ●│ 26 GP20
           GP13 16 │●           ●│ 25 GP19
           GND 17 │●           ●│ 24 GP18
           GP14 18 │●           ●│ 23 GND
           GP15 19 │●           ●│ 22 GP17
           GP16 20 │●           ●│ 21 GP16
                   └─────────────┘
```

## ⚙️ Configuration

### Update your `config.py`:
```python
# Photo Transistor Light Sensor Configuration
LIGHT_SENSOR_PIN = 28         # GPIO pin for photo transistor (ADC2)
LIGHT_ENABLED = True          # Enable light sensor
LIGHT_THRESHOLD_LOW = 10      # Below 10% = dark
LIGHT_THRESHOLD_HIGH = 80     # Above 80% = bright
LIGHT_READING_INTERVAL = 10   # Read every 10 seconds
```

### MQTT Topic:
- **Topic**: `room-light/{room_number}`
- **Payload**: JSON with light level percentage and state

## 🔬 Testing and Calibration

### 1. Initial Testing:
```python
# Test script to verify sensor readings
from machine import Pin, ADC
import time

light_sensor = ADC(Pin(28))

while True:
    reading = light_sensor.read_u16()
    percentage = (reading / 65535.0) * 100.0
    print(f"ADC: {reading}, Light: {percentage:.1f}%")
    time.sleep(1)
```

### 2. Calibration Steps:
1. **Test in darkness** (cover sensor) → Should read ~0-5%
2. **Test in normal room light** → Should read ~20-60%
3. **Test in bright light** (flashlight/sunlight) → Should read ~80-100%
4. **Adjust thresholds** in config based on your environment

### 3. Expected Readings:
- **Complete Darkness**: 0-5%
- **Dim Room Light**: 10-30%
- **Normal Room Light**: 30-70%
- **Bright Light**: 70-95%
- **Direct Sunlight**: 90-100%

## 🛠️ Troubleshooting

### Common Issues:

#### Readings Always 0%:
- Check wiring connections
- Verify 3.3V power supply
- Test photo transistor with multimeter

#### Readings Always 100%:
- Photo transistor may be wired backwards
- Check collector/emitter connections
- Resistor may be wrong value or missing

#### Erratic Readings:
- Add capacitor (100nF) across ADC input and GND
- Shield sensor from electrical noise
- Check for loose connections

#### Not Responsive to Light:
- Verify photo transistor is light-sensitive type
- Remove any covering from sensor surface
- Test with strong light source (flashlight)

### Advanced Tips:

#### Improve Sensitivity:
- Use larger pull-down resistor (47kΩ instead of 10kΩ)
- Add amplification circuit with op-amp
- Use specialized ambient light sensor IC

#### Reduce Noise:
- Add 100nF ceramic capacitor from ADC pin to GND
- Use shielded cable for longer wire runs
- Add software filtering/averaging

## 📊 Expected MQTT Data

### Sample Light Sensor Message:
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

### Light States:
- **"dark"**: Below configured low threshold (< 10%)
- **"normal"**: Between thresholds (10-80%)
- **"bright"**: Above configured high threshold (> 80%)

## 🎯 Integration Examples

### Home Automation Uses:
- **Automatic Lighting**: Turn on lights when dark
- **Security**: Detect unusual light patterns
- **Energy Saving**: Adjust display brightness
- **Circadian Rhythms**: Track natural light cycles
- **Greenhouse Monitoring**: Optimize plant lighting

**Your Pi Pico now has professional-grade ambient light sensing! 🌞🌙**
