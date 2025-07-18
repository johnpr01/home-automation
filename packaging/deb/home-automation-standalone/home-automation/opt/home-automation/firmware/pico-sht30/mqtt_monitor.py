#!/usr/bin/env python3
"""
MQTT Test Client for Pi Pico SHT-30 Sensor
Subscribe to temperature and humidity topics and display data
"""

import sys
import json
import time
from datetime import datetime

try:
    import paho.mqtt.client as mqtt
except ImportError:
    print("paho-mqtt not installed. Install with: pip install paho-mqtt")
    sys.exit(1)


class MQTTMonitor:
    def __init__(self, broker_host, broker_port=1883, username=None, password=None):
        self.broker_host = broker_host
        self.broker_port = broker_port
        self.username = username
        self.password = password
        self.client = mqtt.Client()
        
        # Set up callbacks
        self.client.on_connect = self.on_connect
        self.client.on_message = self.on_message
        self.client.on_disconnect = self.on_disconnect
        
        # Set credentials if provided
        if username and password:
            self.client.username_pw_set(username, password)
    
    def on_connect(self, client, userdata, flags, rc):
        if rc == 0:
            print(f"Connected to MQTT broker at {self.broker_host}:{self.broker_port}")
            # Subscribe to all room temperature and humidity topics
            client.subscribe("room-temp/+")
            client.subscribe("room-hum/+")
            print("Subscribed to topics: room-temp/+, room-hum/+")
        else:
            print(f"Failed to connect to MQTT broker. Return code: {rc}")
    
    def on_message(self, client, userdata, msg):
        try:
            topic = msg.topic
            payload = msg.payload.decode('utf-8')
            
            # Parse JSON payload
            data = json.loads(payload)
            
            # Extract information
            room = data.get('room', 'Unknown')
            timestamp = data.get('timestamp', time.time())
            sensor = data.get('sensor', 'Unknown')
            device_id = data.get('device_id', 'Unknown')
            
            # Format timestamp
            dt = datetime.fromtimestamp(timestamp)
            time_str = dt.strftime('%Y-%m-%d %H:%M:%S')
            
            if 'temperature' in data:
                temp = data['temperature']
                unit = data.get('unit', '°C')
                print(f"[{time_str}] Room {room} - Temperature: {temp}{unit} (Device: {device_id})")
            
            elif 'humidity' in data:
                humidity = data['humidity']
                unit = data.get('unit', '%')
                print(f"[{time_str}] Room {room} - Humidity: {humidity}{unit} (Device: {device_id})")
            
        except json.JSONDecodeError:
            print(f"Invalid JSON received on topic {topic}: {payload}")
        except Exception as e:
            print(f"Error processing message on topic {topic}: {e}")
    
    def on_disconnect(self, client, userdata, rc):
        if rc != 0:
            print("Unexpected disconnection from MQTT broker")
        else:
            print("Disconnected from MQTT broker")
    
    def start_monitoring(self):
        try:
            print(f"Connecting to MQTT broker at {self.broker_host}:{self.broker_port}...")
            self.client.connect(self.broker_host, self.broker_port, 60)
            
            print("Starting MQTT monitoring... Press Ctrl+C to stop")
            self.client.loop_forever()
            
        except KeyboardInterrupt:
            print("\nStopping MQTT monitoring...")
            self.client.disconnect()
        except Exception as e:
            print(f"Error: {e}")


def main():
    if len(sys.argv) < 2:
        print("Usage: python3 mqtt_monitor.py <broker_host> [broker_port] [username] [password]")
        print("Example: python3 mqtt_monitor.py 192.168.1.100")
        print("Example: python3 mqtt_monitor.py 192.168.1.100 1883 user pass")
        sys.exit(1)
    
    broker_host = sys.argv[1]
    broker_port = int(sys.argv[2]) if len(sys.argv) > 2 else 1883
    username = sys.argv[3] if len(sys.argv) > 3 else None
    password = sys.argv[4] if len(sys.argv) > 4 else None
    
    monitor = MQTTMonitor(broker_host, broker_port, username, password)
    monitor.start_monitoring()


if __name__ == "__main__":
    main()
