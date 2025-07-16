# Pi Pico WH + SHT-30 Sensor Firmware

This directory contains MicroPython firmware for the Raspberry Pi Pico WH with an SHT-30 temperature and humidity sensor.

## Hardware Requirements

- Raspberry Pi Pico WH (with WiFi)
- SHT-30 temperature and humidity sensor
- Jumper wires for I2C connection

## Wiring

Connect the SHT-30 sensor to the Pico WH as follows:

| SHT-30 Pin | Pico WH Pin | Description |
|------------|-------------|-------------|
| VDD        | 3V3 (Pin 36)| Power supply |
| GND        | GND (Pin 38)| Ground |
| SDA        | GPIO4 (Pin 6)| I2C Data |
| SCL        | GPIO5 (Pin 7)| I2C Clock |

## Installation

1. **Install MicroPython on Pico WH:**
   - Download the latest MicroPython firmware for Pico W from [micropython.org](https://micropython.org/download/rp2-pico-w/)
   - Hold the BOOTSEL button while connecting the Pico to your computer
   - Copy the `.uf2` file to the RPI-RP2 drive that appears

2. **Upload the firmware:**
   ```bash
   # Install required tools
   pip install mpremote

   # Copy configuration file
   cp config_template.py config.py
   # Edit config.py with your WiFi and MQTT settings

   # Upload files to Pico
   mpremote cp config.py :
   mpremote cp main.py :
   mpremote cp sht30.py :
   ```

3. **Configure settings:**
   - Edit `config.py` with your WiFi credentials
   - Set your Raspberry Pi 5 IP address as the MQTT broker
   - Configure the room number for this sensor
   - Adjust GPIO pins if needed

## Configuration

The `config.py` file contains all configurable settings:

- **WiFi Settings:** SSID and password for your network
- **MQTT Settings:** Broker IP, port, and credentials
- **Device Settings:** Room number and device identifier
- **Sensor Settings:** Reading interval and I2C pins
- **Topics:** MQTT topic templates for temperature and humidity

## MQTT Topics

The sensor publishes data to the following topics:

- Temperature: `room-temp/{room_number}`
- Humidity: `room-hum/{room_number}`

Where `{room_number}` is configured in `config.py`.

## Monitoring

You can monitor the sensor output using:

```bash
# View serial output
mpremote

# Or use mosquitto_sub to monitor MQTT messages
mosquitto_sub -h YOUR_BROKER_IP -t "room-temp/+" -t "room-hum/+"
```

## Features

- Automatic WiFi connection with retry logic
- MQTT connection with automatic reconnection
- Error handling and recovery
- Configurable reading intervals
- Status LED indication (built-in LED on Pico WH)
- Detailed logging for debugging

## Troubleshooting

1. **WiFi Connection Issues:**
   - Check SSID and password in `config.py`
   - Ensure the Pico WH is within WiFi range
   - Check that your network supports 2.4GHz (Pico WH doesn't support 5GHz)

2. **MQTT Connection Issues:**
   - Verify Raspberry Pi 5 IP and MQTT port
   - Check firewall settings on the Pi
   - Ensure MQTT broker is running (docker compose ps)

3. **Sensor Reading Issues:**
   - Check I2C wiring connections
   - Verify SHT-30 sensor is working
   - Check GPIO pin configuration

4. **Serial Output:**
   Use `mpremote` to view detailed logs and error messages from the device.

## File Structure

- `main.py` - Main application code
- `sht30.py` - SHT-30 sensor driver
- `config.py` - Configuration settings (created from template)
- `config_template.py` - Configuration template
- `README.md` - This documentation
