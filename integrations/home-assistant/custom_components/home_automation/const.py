"""Constants for Home Automation integration."""

DOMAIN = "home_automation"

# Platforms supported by this integration
PLATFORMS = [
    "sensor",
    "binary_sensor", 
    "climate",
    "light",
    "switch",
]

# Default configuration
DEFAULT_HOST = "localhost"
DEFAULT_PORT = 8080
DEFAULT_MQTT_HOST = "localhost"
DEFAULT_MQTT_PORT = 1883

# API Endpoints
API_ENDPOINTS = {
    "status": "/api/status",
    "devices": "/api/devices",
    "sensors": "/api/sensors", 
    "rooms": "/api/rooms",
}

# MQTT Topics
MQTT_TOPICS = {
    "room_temp": "room-temp/+",
    "room_motion": "room-motion/+",
    "room_light": "room-light/+",
    "room_humidity": "room-hum/+",
    "thermostat_control": "thermostat/+/control",
    "automation_events": "automation/+",
}

# Device classes
DEVICE_CLASSES = {
    "temperature": "temperature",
    "humidity": "humidity", 
    "motion": "motion",
    "light": "illuminance",
    "occupancy": "occupancy",
}

# Update intervals
DEFAULT_SCAN_INTERVAL = 30  # seconds
FAST_SCAN_INTERVAL = 10     # seconds for critical sensors
SLOW_SCAN_INTERVAL = 60     # seconds for less critical data

# Error messages
ERROR_CANNOT_CONNECT = "Cannot connect to Home Automation system"
ERROR_INVALID_HOST = "Invalid host specified"
ERROR_TIMEOUT = "Timeout connecting to Home Automation system"
