#!/bin/bash

# Pi Pico WH SHT-30 Sensor Firmware Deployment Script

set -e

echo "Pi Pico WH SHT-30 Sensor Firmware Deployment"
echo "============================================="

# Check if mpremote is installed
if ! command -v mpremote &> /dev/null; then
    echo "mpremote not found. Installing..."
    pip install mpremote
fi

# Check if config.py exists
if [ ! -f "config.py" ]; then
    echo "config.py not found. Creating from template..."
    cp config_template.py config.py
    echo "Please edit config.py with your WiFi and MQTT settings before continuing."
    echo "Opening config.py for editing..."
    ${EDITOR:-nano} config.py
    
    read -p "Press Enter when you've finished configuring config.py..."
fi

echo "Checking if Pico is connected..."
if ! mpremote ls > /dev/null 2>&1; then
    echo "ERROR: Pico not found or not in MicroPython mode"
    echo "Please:"
    echo "1. Connect your Pico WH to USB"
    echo "2. Ensure MicroPython is installed"
    echo "3. Check that the device is detected"
    exit 1
fi

echo "Pico found! Current files on device:"
mpremote ls

echo ""
echo "Uploading firmware files..."

# Upload configuration
echo "Uploading config.py..."
mpremote cp config.py :

# Upload SHT-30 driver
echo "Uploading sht30.py..."
mpremote cp sht30.py :

# Upload main application
echo "Uploading main.py..."
mpremote cp main.py :

echo ""
echo "Files uploaded successfully!"
echo "Current files on Pico:"
mpremote ls

echo ""
echo "Testing configuration..."
mpremote exec "
try:
    from config import *
    print('Configuration loaded successfully')
    print(f'WiFi SSID: {WIFI_SSID}')
    print(f'MQTT Broker: {MQTT_BROKER}:{MQTT_PORT}')
    print(f'Room Number: {ROOM_NUMBER}')
    print(f'Device Name: {DEVICE_NAME}')
except Exception as e:
    print(f'Configuration error: {e}')
"

echo ""
echo "Testing SHT-30 driver..."
mpremote exec "
try:
    from machine import Pin, I2C
    from sht30 import SHT30
    from config import I2C_SDA_PIN, I2C_SCL_PIN
    
    i2c = I2C(0, sda=Pin(I2C_SDA_PIN), scl=Pin(I2C_SCL_PIN), freq=400000)
    devices = i2c.scan()
    print(f'I2C devices found: {[hex(d) for d in devices]}')
    
    if 0x44 in devices:
        sensor = SHT30(i2c)
        temp, hum = sensor.read_temperature_humidity()
        print(f'Sensor test successful - Temp: {temp:.2f}°C, Humidity: {hum:.2f}%')
    else:
        print('SHT-30 sensor not found at address 0x44')
        print('Please check wiring connections')
except Exception as e:
    print(f'Sensor test failed: {e}')
"

echo ""
echo "Deployment complete!"
echo ""
echo "To start the sensor:"
echo "  mpremote run main.py"
echo ""
echo "To monitor output:"
echo "  mpremote"
echo ""
echo "To make it auto-start on boot, rename main.py to boot.py:"
echo "  mpremote exec \"import os; os.rename('main.py', 'boot.py')\""
echo ""
echo "Hardware connections should be:"
echo "  SHT-30 VDD -> Pico 3V3 (Pin 36)"
echo "  SHT-30 GND -> Pico GND (Pin 38)"
echo "  SHT-30 SDA -> Pico GPIO4 (Pin 6)"
echo "  SHT-30 SCL -> Pico GPIO5 (Pin 7)"
