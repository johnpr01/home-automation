# Configuration file for Pico SHT-30 MQTT Sensor
# Copy this file to config.py and update with your settings

# WiFi Configuration
WIFI_SSID = "YOUR_WIFI_NETWORK"
WIFI_PASSWORD = "YOUR_WIFI_PASSWORD"

# MQTT Configuration
MQTT_BROKER = "192.168.1.100"  # Replace with your Raspberry Pi 5 IP address
MQTT_PORT = 1883
MQTT_USER = ""  # Leave empty if no authentication
MQTT_PASSWORD = ""  # Leave empty if no authentication

# Device Configuration
ROOM_NUMBER = "1"  # Change this for different rooms (1, 2, 3, etc.)
DEVICE_NAME = "pico-sht30-room1"  # Unique device identifier

# Sensor Configuration
READING_INTERVAL = 5  # Seconds between readings
I2C_SDA_PIN = 4  # GPIO pin for I2C SDA
I2C_SCL_PIN = 5  # GPIO pin for I2C SCL

# MQTT Topics (will be formatted with room number)
TEMP_TOPIC_TEMPLATE = "room-temp/{room}"
HUM_TOPIC_TEMPLATE = "room-hum/{room}"

# Advanced Settings
MAX_WIFI_RETRIES = 30
MAX_MQTT_RETRIES = 5
MAX_CONSECUTIVE_ERRORS = 5
ENABLE_DETAILED_LOGGING = True
