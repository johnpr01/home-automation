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
MOTION_TOPIC = MOTION_TOPIC_TEMPLATE.format(room=ROOM_NUMBER)
LIGHT_TOPIC = LIGHT_TOPIC_TEMPLATE.format(room=ROOM_NUMBER)

# LED for status indication
led = Pin("LED", Pin.OUT)

# PIR sensor setup (if enabled)
pir_sensor = None
if PIR_ENABLED:
    pir_sensor = Pin(PIR_SENSOR_PIN, Pin.IN)
    print(f"PIR sensor enabled on GPIO {PIR_SENSOR_PIN}")

# Photo transistor light sensor setup (if enabled)
light_sensor = None
if LIGHT_ENABLED:
    from machine import ADC
    light_sensor = ADC(Pin(LIGHT_SENSOR_PIN))
    print(f"Light sensor enabled on GPIO {LIGHT_SENSOR_PIN} (ADC)")

# Motion detection state
motion_detected = False
last_motion_time = 0
motion_state_sent = False

# Light sensor state
light_level_percent = 0
last_light_reading = 0
last_light_state = "unknown"  # "dark", "normal", "bright"

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
            temperature_celsius = -45 + (175 * temp_raw / 65535.0)
            
            # Convert to Fahrenheit
            temperature = (temperature_celsius * 9.0 / 5.0) + 32.0
            
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
            "unit": "°F",
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

def publish_motion_event(client, motion_detected, motion_start_time=None):
    """Publish motion detection event to MQTT topic"""
    try:
        timestamp = time.time()
        
        motion_payload = ujson.dumps({
            "motion": motion_detected,
            "room": ROOM_NUMBER,
            "sensor": "PIR",
            "timestamp": timestamp,
            "motion_start": motion_start_time if motion_start_time else timestamp,
            "device_id": DEVICE_NAME
        })
        
        # Publish motion event
        client.publish(MOTION_TOPIC, motion_payload)
        if ENABLE_DETAILED_LOGGING:
            status = "DETECTED" if motion_detected else "CLEARED"
            print(f"Published motion {status} to {MOTION_TOPIC}")
        
        return True
        
    except Exception as e:
        print(f"Failed to publish motion data: {e}")
        return False

def check_motion_sensor():
    """Check PIR sensor and return motion state"""
    global motion_detected, last_motion_time, motion_state_sent
    
    if not PIR_ENABLED or not pir_sensor:
        return False
    
    current_time = time.time()
    pir_reading = pir_sensor.value()
    
    # Motion detected
    if pir_reading == 1:
        if not motion_detected:
            # New motion detected
            motion_detected = True
            last_motion_time = current_time
            motion_state_sent = False
            if ENABLE_DETAILED_LOGGING:
                print(f"Motion detected in room {ROOM_NUMBER}")
        else:
            # Motion still ongoing, update last seen time
            last_motion_time = current_time
        return True
    
    # No motion currently detected
    else:
        if motion_detected:
            # Check if motion timeout has elapsed
            if current_time - last_motion_time >= PIR_TIMEOUT:
                motion_detected = False
                motion_state_sent = False
                if ENABLE_DETAILED_LOGGING:
                    print(f"Motion cleared in room {ROOM_NUMBER}")
                return False
        return motion_detected  # Still in timeout period

def read_light_sensor():
    """Read photo transistor and return light level percentage"""
    if not LIGHT_ENABLED or not light_sensor:
        return None
    
    try:
        # Read ADC value (0-65535 on Pico)
        adc_value = light_sensor.read_u16()
        
        # Convert to percentage (0-100%)
        # Higher ADC value = more light (assuming photo transistor pulls voltage high with light)
        light_percentage = (adc_value / 65535.0) * 100.0
        
        return light_percentage
        
    except Exception as e:
        print(f"Error reading light sensor: {e}")
        return None

def publish_light_data(client, light_level, light_state):
    """Publish light level data to MQTT topic"""
    try:
        timestamp = time.time()
        
        light_payload = ujson.dumps({
            "light_level": round(light_level, 1),
            "light_percent": round(light_level, 1),
            "light_state": light_state,  # "dark", "normal", "bright"
            "unit": "%",
            "room": ROOM_NUMBER,
            "sensor": "PhotoTransistor",
            "timestamp": timestamp,
            "device_id": DEVICE_NAME
        })
        
        # Publish light data
        client.publish(LIGHT_TOPIC, light_payload)
        if ENABLE_DETAILED_LOGGING:
            print(f"Published light: {light_level:.1f}% ({light_state}) to {LIGHT_TOPIC}")
        
        return True
        
    except Exception as e:
        print(f"Failed to publish light data: {e}")
        return False

def determine_light_state(light_percent):
    """Determine light state based on percentage thresholds"""
    if light_percent < LIGHT_THRESHOLD_LOW:
        return "dark"
    elif light_percent > LIGHT_THRESHOLD_HIGH:
        return "bright"
    else:
        return "normal"

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
    if PIR_ENABLED:
        print(f"Motion Topic: {MOTION_TOPIC}")
    if LIGHT_ENABLED:
        print(f"Light Topic: {LIGHT_TOPIC}")
    
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
    if PIR_ENABLED:
        print(f"PIR motion detection enabled for room {ROOM_NUMBER}")
    if LIGHT_ENABLED:
        print(f"Light sensor enabled for room {ROOM_NUMBER} (thresholds: dark<{LIGHT_THRESHOLD_LOW}%, bright>{LIGHT_THRESHOLD_HIGH}%)")
    print("Press Ctrl+C to stop")
    
    # Main sensor loop
    error_count = 0
    last_sensor_reading = 0
    
    while True:
        try:
            current_time = time.time()
            
            # Check motion sensor (check every loop iteration)
            if PIR_ENABLED:
                motion_state = check_motion_sensor()
                
                # Send motion state change if needed
                if motion_state != motion_state_sent:
                    if publish_motion_event(mqtt_client, motion_state, last_motion_time):
                        motion_state_sent = motion_state
                        if motion_state:
                            blink_led(2, 0.05)  # Quick double blink for motion
            
            # Read temperature/humidity sensors at configured interval
            if current_time - last_sensor_reading >= READING_INTERVAL:
                temperature, humidity = sensor.read_temperature_humidity()
                
                # Read light sensor data
                light_level = None
                current_light_state = "unknown"
                if LIGHT_ENABLED:
                    light_level = read_light_sensor()
                    if light_level is not None:
                        current_light_state = determine_light_state(light_level)
                        
                        # Update global state
                        global light_level_percent, last_light_state
                        light_level_percent = light_level
                        
                        # Publish light data if state changed or it's been a while
                        if (current_light_state != last_light_state or 
                            current_time - last_light_reading >= LIGHT_READING_INTERVAL):
                            if publish_light_data(mqtt_client, light_level, current_light_state):
                                last_light_state = current_light_state
                                global last_light_reading
                                last_light_reading = current_time
                
                if ENABLE_DETAILED_LOGGING:
                    motion_status = " [MOTION]" if (PIR_ENABLED and motion_detected) else ""
                    light_status = f" [LIGHT:{current_light_state.upper()}:{light_level:.1f}%]" if LIGHT_ENABLED and light_level is not None else ""
                    print(f"Room {ROOM_NUMBER} - Temp: {temperature:.2f}°F, Humidity: {humidity:.2f}%{motion_status}{light_status}")
                else:
                    motion_indicator = "M" if (PIR_ENABLED and motion_detected) else " "
                    light_indicator = f"L:{light_level:.0f}%" if LIGHT_ENABLED and light_level is not None else ""
                    print(f"T:{temperature:.1f}°F H:{humidity:.1f}% {motion_indicator} {light_indicator}")
                
                # Publish temperature/humidity to MQTT
                if publish_sensor_data(mqtt_client, temperature, humidity):
                    blink_led(1, 0.1)  # Short blink for success
                    error_count = 0  # Reset error count on success
                else:
                    error_count += 1
                    blink_led(3, 0.1)  # Triple blink for MQTT error
                
                last_sensor_reading = current_time
            
            # Check for too many consecutive errors
            if error_count >= MAX_CONSECUTIVE_ERRORS:
                print(f"Too many consecutive errors ({error_count}), restarting...")
                machine.reset()
            
            # Garbage collection
            gc.collect()
            
            # Short sleep to prevent busy waiting
            time.sleep(0.1)
            
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
