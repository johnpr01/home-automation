# Raspberry Pi Pico WH with SHT-30 Sensor
# Temperature and Humidity Monitoring with MQTT

import machine
import network
import time
import ujson
from machine import Pin, I2C
from umqtt.simple import MQTTClient
import gc

# Import configuration
try:
    from config import *
except ImportError:
    print("ERROR: config.py not found! Please copy config_template.py to config.py and configure it.")
    machine.reset()

# Import SHT-30 driver
try:
    from sht30 import SHT30
except ImportError:
    print("ERROR: sht30.py driver not found!")
    machine.reset()

# Generate MQTT topics from configuration
TEMP_TOPIC = TEMP_TOPIC_TEMPLATE.format(room=ROOM_NUMBER)
HUM_TOPIC = HUM_TOPIC_TEMPLATE.format(room=ROOM_NUMBER)

# LED for status indication
led = Pin("LED", Pin.OUT)

class SHT30:
    """SHT-30 Temperature and Humidity Sensor Driver"""
    
    def __init__(self, i2c, addr=0x44):
        self.i2c = i2c
        self.addr = addr
        
    def read_data(self):
        """Read temperature and humidity from SHT-30"""
        try:
            # Send measurement command (high repeatability)
            self.i2c.writeto(self.addr, b'\x2C\x06')
            time.sleep_ms(500)  # Wait for measurement
            
            # Read 6 bytes of data
            data = self.i2c.readfrom(self.addr, 6)
            
            # Convert temperature (first 3 bytes)
            temp_raw = (data[0] << 8) | data[1]
            temperature = -45 + (175 * temp_raw / 65535.0)
            
            # Convert humidity (last 3 bytes)
            hum_raw = (data[3] << 8) | data[4]
            humidity = 100 * hum_raw / 65535.0
            
            return temperature, humidity
            
        except Exception as e:
            print(f"Error reading SHT-30: {e}")
            return None, None

def connect_wifi():
    """Connect to WiFi network"""
    wlan = network.WLAN(network.STA_IF)
    wlan.active(True)
    
    if not wlan.isconnected():
        print(f"Connecting to WiFi: {WIFI_SSID}")
        wlan.connect(WIFI_SSID, WIFI_PASSWORD)
        
        # Wait for connection
        timeout = 0
        while not wlan.isconnected() and timeout < 30:
            print(".", end="")
            time.sleep(1)
            timeout += 1
            
        if wlan.isconnected():
            print(f"\nWiFi connected: {wlan.ifconfig()}")
            return True
        else:
            print("\nWiFi connection failed!")
            return False
    else:
        print(f"Already connected to WiFi: {wlan.ifconfig()}")
        return True

def connect_mqtt():
    """Connect to MQTT broker"""
    try:
        if MQTT_USER and MQTT_PASSWORD:
            client = MQTTClient(DEVICE_NAME, MQTT_BROKER, port=MQTT_PORT, 
                              user=MQTT_USER, password=MQTT_PASSWORD)
        else:
            client = MQTTClient(DEVICE_NAME, MQTT_BROKER, port=MQTT_PORT)
            
        client.connect()
        print(f"MQTT connected to {MQTT_BROKER}:{MQTT_PORT}")
        return client
    except Exception as e:
        print(f"MQTT connection failed: {e}")
        return None

def publish_sensor_data(client, temperature, humidity):
    """Publish temperature and humidity to MQTT topics"""
    try:
        # Create JSON payloads with timestamp
        timestamp = time.time()
        
        temp_payload = ujson.dumps({
            "temperature": round(temperature, 2),
            "unit": "°C",
            "room": ROOM_NUMBER,
            "sensor": "SHT-30",
            "timestamp": timestamp,
            "device_id": DEVICE_NAME
        })
        
        hum_payload = ujson.dumps({
            "humidity": round(humidity, 2),
            "unit": "%",
            "room": ROOM_NUMBER,
            "sensor": "SHT-30", 
            "timestamp": timestamp,
            "device_id": DEVICE_NAME
        })
        
        # Publish temperature
        client.publish(TEMP_TOPIC, temp_payload)
        if ENABLE_DETAILED_LOGGING:
            print(f"Published temp: {temperature:.2f}°C to {TEMP_TOPIC}")
        
        # Publish humidity
        client.publish(HUM_TOPIC, hum_payload)
        if ENABLE_DETAILED_LOGGING:
            print(f"Published humidity: {humidity:.2f}% to {HUM_TOPIC}")
        
        return True
        
    except Exception as e:
        print(f"Failed to publish data: {e}")
        return False

def blink_led(times=1, delay=0.1):
    """Blink LED for status indication"""
    for _ in range(times):
        led.on()
        time.sleep(delay)
        led.off()
        time.sleep(delay)

def main():
    """Main program loop"""
    print("Starting Pico SHT-30 MQTT Sensor...")
    print(f"Room Number: {ROOM_NUMBER}")
    print(f"Temperature Topic: {TEMP_TOPIC}")
    print(f"Humidity Topic: {HUM_TOPIC}")
    
    # Initialize I2C
    try:
        i2c = I2C(0, sda=Pin(I2C_SDA_PIN), scl=Pin(I2C_SCL_PIN), freq=400000)
        print(f"I2C initialized on SDA={I2C_SDA_PIN}, SCL={I2C_SCL_PIN}")
        
        # Scan for devices
        devices = i2c.scan()
        print(f"I2C devices found: {[hex(device) for device in devices]}")
            
    except Exception as e:
        print(f"I2C initialization failed: {e}")
        return
    
    # Initialize SHT-30 sensor using the new driver
    try:
        sensor = SHT30(i2c)
        print("SHT-30 sensor initialized successfully")
    except Exception as e:
        print(f"SHT-30 sensor initialization failed: {e}")
        return
    
    # Connect to WiFi
    if not connect_wifi():
        print("Cannot continue without WiFi")
        return
    
    # Connect to MQTT
    mqtt_client = connect_mqtt()
    if not mqtt_client:
        print("Cannot continue without MQTT")
        return
    
    print(f"Starting sensor readings every {READING_INTERVAL} seconds...")
    print("Press Ctrl+C to stop")
    
    # Main sensor loop
    error_count = 0
    
    while True:
        try:
            # Read sensor data using the new driver
            temperature, humidity = sensor.read_temperature_humidity()
            
            if ENABLE_DETAILED_LOGGING:
                print(f"Room {ROOM_NUMBER} - Temp: {temperature:.2f}°C, Humidity: {humidity:.2f}%")
            else:
                print(f"T:{temperature:.1f}°C H:{humidity:.1f}%")
            
            # Publish to MQTT
            if publish_sensor_data(mqtt_client, temperature, humidity):
                blink_led(1, 0.1)  # Short blink for success
                error_count = 0  # Reset error count on success
            else:
                error_count += 1
                blink_led(3, 0.1)  # Triple blink for MQTT error
            
            # Check for too many consecutive errors
            if error_count >= MAX_CONSECUTIVE_ERRORS:
                print(f"Too many consecutive errors ({error_count}), restarting...")
                machine.reset()
            
            # Garbage collection
            gc.collect()
            
            # Wait for next reading
            time.sleep(READING_INTERVAL)
            
        except KeyboardInterrupt:
            print("\nStopping sensor readings...")
            break
        except Exception as e:
            print(f"Sensor reading error: {e}")
            error_count += 1
            blink_led(5, 0.1)  # Five blinks for sensor error
            
            # Check for too many consecutive errors
            if error_count >= MAX_CONSECUTIVE_ERRORS:
                print(f"Too many consecutive errors ({error_count}), restarting...")
                machine.reset()
                
            time.sleep(READING_INTERVAL)
    
    # Cleanup
    try:
        mqtt_client.disconnect()
    except:
        pass
    print("Program ended")

if __name__ == "__main__":
    main()
