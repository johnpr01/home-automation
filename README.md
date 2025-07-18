# Home Automation System

A comprehensive home automation system built in Go, designed to run on Raspberry Pi 5 with integrated sensor networks, MQTT messaging, intelligent thermostat control, and **comprehensive error handling**.

## 🏠 System Overview

This system provides a unified platform for managing home automation devices with real-time environmental monitoring, motion detection, intelligent climate control, and **production-ready error handling and monitoring**.

### Key Features

- **Multi-Sensor Integration**: Temperature, humidity, motion, and light sensors on a single Pi Pico device
- **Unified Sensor Service**: Centralized management of all sensor data with intelligent aggregation
- **Smart Thermostat**: Fahrenheit-based climate control with occupancy awareness
- **Energy Monitoring**: TP-Link Tapo smart plug monitoring with Prometheus time-series storage and KLAP protocol support
- **Modern Protocol Support**: KLAP protocol implementation for Tapo firmware 1.1.0+ with legacy fallback
- **Real-time MQTT**: Low-latency sensor data transmission and device control
- **Microcontroller Sensors**: Pi Pico WH with SHT-30, PIR, and photo transistor sensors
- **Container Orchestration**: Docker Compose with optimized resource allocation
- **Message Streaming**: Kafka integration for data persistence and analytics
- **Motion Detection**: PIR sensor monitoring with room occupancy tracking
- **Ambient Light Sensing**: Photo transistor monitoring with day/night cycle detection
- **Pi Pico Integration**: SHT-30, PIR, and photo transistor sensors via MQTT
- **Multi-Zone Support**: Control multiple rooms independently
- **Orthogonal Architecture**: Services operate independently but can integrate when needed
- **📊 Time-Series Analytics**: Prometheus + Grafana dashboards for energy and sensor monitoring
- **🔌 Smart Plug Control**: TP-Link Tapo integration for power monitoring and device control
- **🛡️ Comprehensive Error Handling**: Structured errors, retry mechanisms, circuit breakers, and health monitoring
- **📊 Production Monitoring**: Structured logging, health checks, and error metrics
- **🔄 Automatic Recovery**: Circuit breakers, retry logic, and graceful degradation
- **🔒 Enterprise Security**: TLS encryption, secure authentication, and comprehensive security auditing

### 🏗️ **Service Architecture:**
- **Thermostat Service**: HVAC temperature control and automation
- **Motion Service**: PIR sensor monitoring and occupancy detection  
- **Light Service**: Photo transistor ambient light tracking
- **Tapo Service**: TP-Link smart plug energy monitoring and control
- **Integrated Service**: Optional combined service with cross-sensor automation
- **Automation Service**: Motion-activated lighting and smart home rules

### 🛡️ **Reliability & Error Handling:**
- **Structured Error Types**: Connection, Device, Service, System, and Validation errors with severity levels
- **Retry Mechanisms**: Exponential backoff with jitter and context-aware timeouts
- **Circuit Breakers**: Prevent cascade failures with automatic recovery
- **Health Monitoring**: Continuous service health checks with detailed reporting
- **Graceful Degradation**: System continues operating with reduced functionality during partial failures
- **Comprehensive Logging**: Structured JSON logging with Kafka integration for high-severity events
- **Error Context**: Rich error information including stack traces, device IDs, and operational context

### 🤖 **Smart Automation Features:**
- **Motion-Activated Lighting**: Automatically turn on lights when motion is detected in dark rooms
- **Intelligent Light Control**: Only activates lights when ambient light is below threshold (20%)

### 🔒 **Enterprise Security Features:**
- **TLS Encryption**: All communications encrypted with TLS 1.2/1.3
- **HTTPS/MQTTS**: Secure web interfaces and MQTT messaging
- **Certificate Management**: Self-signed certificates with proper CA chain
- **Authentication**: Secure password-based authentication for all services
- **Network Security**: Internal Docker networks, no unnecessary port exposure
- **Secret Management**: Secure environment configuration with restricted permissions
- **Security Auditing**: Comprehensive security status monitoring and reporting
- **Emergency Security Fix**: Automated script to quickly secure the system

## 🛡️ **Security & TLS Setup**

This system includes comprehensive security features with TLS encryption for all communications.

### **Quick Security Setup:**
```bash
# 1. Generate TLS certificates
./scripts/generate-certificates.sh

# 2. Deploy TLS-enabled system
./scripts/deploy-tls.sh

# 3. Verify security status
./scripts/check-security.sh
./scripts/verify-tls.sh
```

### **Secure Access URLs:**
- **🌐 Home Automation API:** `https://localhost:8443`
- **📊 Grafana Dashboard:** `https://localhost:3443`
- **📈 Prometheus:** `https://localhost:9443`
- **⚡ Tapo Metrics:** `https://localhost:2443`
- **📨 MQTTS:** `mqtts://localhost:8883`

### **Emergency Security Fix:**
If you need to quickly secure an existing installation:
```bash
./scripts/emergency-security-fix.sh
```

This will:
- Generate strong passwords
- Create secure configuration
- Disable anonymous access
- Add TLS configuration (if certificates exist)
- Create backup of original settings
- **Cooldown Protection**: Prevents rapid on/off cycles with 5-minute cooldown periods
- **Cross-Sensor Integration**: Coordinates between motion sensors, light sensors, and device control
- **Rule-Based Automation**: Configurable automation rules for different rooms and scenarios
- **MQTT Event Publishing**: Publishes automation events to `automation/{room_id}` topics

### Quick Smart Home Setup

1. **Deploy Pi Pico sensors:**
   ```bash
   cd firmware/pico-sht30
   # Configure WiFi, MQTT, and sensor settings
   cp config_template.py config.py
   # Flash to Pi Pico WH with SHT-30, PIR, and photo transistor
   ```

2. **Start services (choose deployment option):**
   ```bash
   # Option A: Thermostat service only
   cd cmd/thermostat && go run main.go
   
   # Option B: Motion detection service only  
   cd cmd/motion && go run main.go
   
   # Option C: Light sensor service only
   cd cmd/light && go run main.go
   
   # Option D: All services with motion-activated lighting automation
   cd cmd/integrated && go run main.go
   
   # Option E: Demo motion-activated lighting
   cd cmd/automation-demo && go run main.go
   ```

3. **Test Motion-Activated Lighting:**
   ```bash
   # Simulate dark room conditions
   mosquitto_pub -h localhost -t 'room-light/living-room' -m '{"light_level":5.0,"light_percent":5.0,"light_state":"dark","room":"living-room","timestamp":1642118400,"device_id":"pico-living"}'
   
   # Simulate motion detection (should trigger lights ON)
   mosquitto_pub -h localhost -t 'room-motion/living-room' -m '{"motion":true,"room":"living-room","timestamp":1642118410,"device_id":"pico-living"}'
   
   # Check automation events
   mosquitto_sub -h localhost -t 'automation/living-room'
   ```

3. **Monitor smart home control:**
   - **Temperature**: `room-temp/{room_id}` → Automatic HVAC control
   - **Motion**: `room-motion/{room_id}` → Occupancy tracking + light automation
   - **Light**: `room-light/{room_id}` → Ambient light monitoring + automation triggers
   - **Automation**: `automation/{room_id}` → Motion-activated lighting events
   - **Integration**: Cross-sensor automation and energy optimization

## 🏗️ Project Structure

```
home-automation/
├── README.md                 # Project documentation
├── go.mod                   # Go module definition
├── go.sum                   # Go module checksums (generated)
├── Makefile                 # Build and development commands
├── Dockerfile               # Container build configuration
├── .gitignore              # Git ignore patterns
│
├── cmd/                     # Main applications
│   ├── server/             # Web server and API
│   │   └── main.go
│   ├── thermostat/         # Smart thermostat service
│   │   └── main.go         # Thermostat control with MQTT  
│   ├── motion/             # Motion detection service
│   │   └── main.go         # PIR sensor monitoring and occupancy tracking
│   ├── light/              # Light sensor service
│   │   └── main.go         # Photo transistor ambient light monitoring
│   ├── tapo-demo/          # TP-Link Tapo smart plug monitoring
│   │   └── main.go         # Energy monitoring and device control
│   ├── integrated/         # Integrated service with motion-activated lighting
│   │   └── main.go         # Combined services with automation rules
│   ├── automation-demo/    # Motion-activated lighting demo
│   │   └── main.go         # Demo and testing for automation features
│   ├── temp-demo/          # Temperature conversion demo
│   │   └── main.go
│   └── cli/                # Command-line interface
│       └── main.go
│
├── internal/               # Private application code
│   ├── config/            # Configuration management
│   │   └── config.go
│   ├── errors/            # 🛡️ Comprehensive error handling system
│   │   └── errors.go      # Structured errors, severity levels, context
│   ├── logger/            # 📊 Structured logging with Kafka integration
│   │   └── logger.go      # JSON logging, error integration, health monitoring
│   ├── utils/             # 🔄 Retry mechanisms and reliability patterns
│   │   └── retry.go       # Circuit breakers, health checks, exponential backoff
│   ├── handlers/          # HTTP request handlers
│   │   └── handlers.go
│   ├── models/            # Data models
│   │   ├── device.go
│   │   ├── sensor.go
│   │   └── thermostat.go  # Smart thermostat models (Fahrenheit)
│   └── services/          # Business logic
│       ├── device_service.go
│       ├── thermostat_service.go # Thermostat control logic (HVAC focused)
│       ├── motion_service.go     # Motion detection and room occupancy
│       ├── light_service.go      # Light sensor monitoring and ambient light tracking
│       ├── tapo_service.go       # TP-Link Tapo smart plug monitoring
│       └── automation_service.go # Motion-activated lighting and smart home rules
│
├── pkg/                    # Public library code
│   ├── devices/           # Device implementations
│   │   └── light.go
│   ├── mqtt/              # MQTT client
│   │   └── client.go
│   ├── kafka/             # Kafka client for logging
│   │   └── client.go
│   ├── influxdb/          # InfluxDB time-series database client
│   │   └── client.go
│   ├── tapo/              # TP-Link Tapo smart plug client
│   │   └── client.go
│   ├── sensors/           # Sensor implementations
│   │   └── temperature.go
│   └── utils/             # Utility functions
│       ├── temperature.go # Fahrenheit/Celsius conversion
│       └── temperature_test.go
│
├── api/                    # API specifications
│   └── openapi.yaml       # OpenAPI/Swagger documentation
│
├── web/                    # Web interface
│   ├── templates/         # HTML templates
│   │   └── index.html
│   └── static/           # Static assets
│       ├── css/
│       │   └── style.css
│       └── js/
│           └── app.js
│
├── configs/               # Configuration files
│   ├── config.yaml       # Default configuration
│   ├── tapo_template.yml # TP-Link Tapo device template
│   └── tapo.yml          # TP-Link Tapo device configuration
│
├── scripts/               # Development and deployment scripts
│   └── setup.sh          # Development environment setup
│
├── docs/                  # Documentation
│   ├── README.md         # Documentation index
│   ├── THERMOSTAT.md     # Smart thermostat guide
│   ├── MOTION_DETECTION.md # PIR motion sensor guide
│   ├── LIGHT_SENSOR.md   # Photo transistor light sensor guide
│   ├── FAHRENHEIT_CONVERSION.md # Fahrenheit conversion details
│   ├── TAPO_ENERGY_MONITORING.md # TP-Link Tapo energy monitoring guide
│   └── ERROR_HANDLING.md # Comprehensive error handling guide
│
├── test/                  # Test files
│   └── device_test.go    # Example tests
│
├── firmware/             # IoT device firmware
│   └── pico-sht30/      # Pi Pico WH multi-sensor firmware
│       ├── main.py      # MicroPython application (temp/humidity/motion/light)
│       ├── sht30.py     # SHT-30 sensor driver
│       ├── config_template.py # Multi-sensor configuration template
│       ├── deploy.sh    # Firmware deployment script
│       ├── README.md    # Firmware documentation
│       ├── MOTION_SENSOR.md # PIR sensor setup guide
│       └── LIGHT_SENSOR.md  # Photo transistor setup guide
│
├── deployments/          # Raspberry Pi 5 deployment
│   ├── docker-compose.yml # Optimized for Pi 5 with InfluxDB + Grafana
│   ├── deploy-pi5.sh     # Automated Pi 5 deployment
│   ├── mosquitto/        # MQTT broker configuration
│   │   ├── mosquitto.conf # Mosquitto configuration
│   │   ├── acl.example   # Access control template
│   │   └── passwd.example # Password file template
│   ├── influxdb/         # Time-series database configuration
│   │   └── influxdb.conf # InfluxDB configuration
│   ├── grafana/          # Grafana dashboard configuration
│   │   ├── provisioning/ # Data sources and dashboards
│   │   │   ├── datasources/
│   │   │   │   └── datasources.yml
│   │   │   └── dashboards/
│   │   │       ├── dashboard.yml
│   │   │       └── tapo-energy-dashboard.json
│   ├── scripts/          # Management scripts
│   │   ├── health-check.sh # System health monitoring
│   │   ├── backup.sh     # Backup script
│   │   └── restore.sh    # Restore script
│   └── README.md         # Pi 5 deployment guide
```

## 🍓 Raspberry Pi 5 Deployment

### Quick Start (Recommended)

1. **Prepare your Raspberry Pi 5:**
   ```bash
   # Update system and install Docker
   sudo apt update && sudo apt upgrade -y
   curl -fsSL https://get.docker.com -o get-docker.sh
   sudo sh get-docker.sh
   sudo usermod -aG docker $USER
   sudo reboot
   ```

2. **Deploy the home automation system:**
   ```bash
   git clone https://github.com/yourname/home-automation.git
   cd home-automation
   ./deploy-with-influxdb.sh
   ```

3. **Access your services:**
   - **Home Automation API**: `http://YOUR_PI_IP:8080`
   - **Smart Thermostat**: Automatic control via MQTT
   - **Grafana Dashboard**: `http://YOUR_PI_IP:3000` (admin/homeauto2024)
   - **InfluxDB**: `http://YOUR_PI_IP:8086` (Time-series data storage)
   - **MQTT Broker**: `YOUR_PI_IP:1883`

4. **Configure Tapo Energy Monitoring:**
   ```bash
   # Configure your TP-Link Tapo devices
   cp configs/tapo_template.yml configs/tapo.yml
   # Edit configs/tapo.yml with your device IPs and credentials
   
   # Start Tapo monitoring service
   cd cmd/tapo-demo && go run main.go
   ```

### Manual Setup

If you prefer manual setup or need customization:

1. **Clone and setup:**
   ```bash
   git clone https://github.com/yourname/home-automation.git
   cd home-automation
   ```

2. **Configure environment:**
   ```bash
   cp .env.example .env
   # Edit .env with your settings
   ```

3. **Deploy services:**
   ```bash
   cd deployments
   docker compose up -d
   ```

### 🚀 Autostart on Boot Setup

For production deployment, you can configure the home automation system to start automatically when your Raspberry Pi boots:

**Quick Install:**
```bash
# Run the automated installation script
sudo ./install-autostart.sh
```

**What this does:**
- ✅ Installs the system as a systemd service
- ✅ Copies files to `/opt/home-automation`
- ✅ Configures automatic startup on boot
- ✅ Sets up proper permissions and security
- ✅ Provides service management commands

**Service Management:**
```bash
# Check service status
sudo systemctl status home-automation

# View real-time logs
sudo journalctl -u home-automation -f

# Stop/start/restart the service
sudo systemctl stop home-automation
sudo systemctl start home-automation
sudo systemctl restart home-automation

# Disable/enable autostart
sudo systemctl disable home-automation
sudo systemctl enable home-automation
```

**Web Interfaces (after autostart):**
- 📊 **Grafana**: `http://raspberrypi.local:3000` (admin/admin)
- 📈 **Prometheus**: `http://raspberrypi.local:9090`
- 🔌 **Tapo Metrics**: `http://raspberrypi.local:2112/metrics`
- 🏠 **Home API**: `http://raspberrypi.local:8080/api/status`

**Remove Autostart:**
```bash
# Run the uninstall script
sudo ./uninstall-autostart.sh
```

For detailed configuration and troubleshooting, see [AUTOSTART_SETUP.md](AUTOSTART_SETUP.md).

## 🛠️ Development Commands

- `make build` - Build all binaries (including thermostat service)
- `make test` - Run tests (including temperature conversion tests)
- `make fmt` - Format code
- `make lint` - Lint code (requires golangci-lint)
- `make dev` - Run with hot reload (requires air)
- `make help` - Show all available commands

### Thermostat Development
- `go run ./cmd/thermostat/` - Run thermostat service locally
- `go run ./cmd/motion/` - Run motion detection service locally
- `go run ./cmd/light/` - Run light sensor service locally
- `go run ./cmd/tapo-demo/` - Run Tapo energy monitoring service locally
- `go run ./cmd/integrated/` - Run integrated service with motion-activated lighting
- `go run ./cmd/automation-demo/` - Demo motion-activated lighting automation
- `go run ./cmd/temp-demo/` - Demo temperature conversions
- `go test ./pkg/utils/` - Test temperature conversion utilities

### Tapo Testing Utilities
- `go build -o test-klap ./cmd/test-klap && ./test-klap -help` - Build and show KLAP protocol test utility
- `./test-klap -host 192.168.1.100 -username user@email.com -password pass` - Test KLAP protocol connectivity
- `go run ./cmd/integration-test/` - Run comprehensive integration tests
- `go run ./cmd/test-tapo-klap/` - Run Tapo KLAP protocol tests

## 🐳 Raspberry Pi 5 Services

The system runs the following services optimized for Raspberry Pi 5:

### Core Services:
- **Home Automation API** (Port 8080) - Main application server
- **Smart Thermostat Service** - Intelligent HVAC temperature control (Fahrenheit)
- **Motion Detection Service** - PIR sensor monitoring and room occupancy tracking
- **Light Sensor Service** - Photo transistor ambient light monitoring and day/night detection
- **PostgreSQL** (Port 5432) - Database with Pi-optimized settings  
- **Mosquitto MQTT** (Port 1883/9001) - Message broker for IoT devices
- **Redis** (Port 6379) - Caching and session storage
- **Kafka** (Port 9092) - Log streaming and event processing
- **Grafana** (Port 3000) - Monitoring dashboard

### Resource Optimization:
- Memory limits appropriate for Pi 5 (512MB-1GB total usage)
- CPU limits to prevent resource starvation
- Optimized database buffer settings
- Efficient logging with rotation
- SD card-friendly persistence settings

## 🌡️ Smart Thermostat Features

### Temperature Control (Fahrenheit)
- **Intelligent Control**: Automatic heating/cooling with 1°F hysteresis
- **Target Temperature**: Default 70°F (adjustable 50°F - 95°F)
- **Multi-Mode**: Heat, Cool, Auto, Fan, and Off modes
- **Safety Limits**: Configurable min/max temperature protection
- **Calibration**: Temperature offset support for sensor accuracy

### Multi-Sensor Integration
Deploy comprehensive environmental monitoring throughout your home:

1. **Configure sensors:**
   ```bash
   cd firmware/pico-sht30
   cp config_template.py config.py
   # Edit with your Pi 5 IP, room assignment
   # Enable SHT-30 (temp/humidity), PIR (motion), and photo transistor (light)
   ```

2. **Flash firmware:**
   ```bash
   # Copy files to Pi Pico WH
   # Sensors automatically send environmental data via MQTT
   ```

3. **Monitor services:**
   ```bash
   # Individual services
   cd cmd/thermostat && go run main.go  # HVAC control
   cd cmd/motion && go run main.go      # Occupancy tracking
   cd cmd/light && go run main.go       # Ambient light monitoring
   
   # Or integrated service
   cd cmd/integrated && go run main.go  # All sensors with automation
   ```

### MQTT Topics (All Environmental Data):
- **Temperature**: `room-temp/{room_number}` (°F) → Thermostat control
- **Humidity**: `room-hum/{room_number}` (%) → Environmental monitoring
- **Motion**: `room-motion/{room_number}` (occupancy) → Presence detection + light automation
- **Light**: `room-light/{room_number}` (%) → Ambient light levels + automation triggers
- **Control**: `thermostat/{thermostat_id}/control` (HVAC commands)
- **Automation**: `automation/{room_id}` (automation events and light control)

### Example Multi-Sensor Operation:
**Smart Home Intelligence with Motion-Activated Lighting**
- 🌡️ **HVAC**: Target 70°F ±1°F hysteresis with automatic heating/cooling
- 👥 **Occupancy**: Motion detection for energy-saving and security
- 🌞 **Lighting**: Ambient light monitoring for automatic lighting control
- 🏠 **Integration**: Cross-sensor automation (e.g., occupied + dark = lights on)
- 🤖 **Automation**: Motion-activated lighting with intelligent dark room detection
- 💡 **Smart Control**: Lights automatically turn on when motion detected in rooms with <20% ambient light
- ⏰ **Cooldown Logic**: 5-minute cooldown prevents rapid on/off cycling
- 📡 **Event Publishing**: Real-time automation events published to MQTT for monitoring

## � Energy Monitoring with TP-Link Tapo

### Smart Plug Integration
Monitor and control TP-Link Tapo smart plugs with comprehensive energy analytics:

- **Real-time Power Monitoring**: Track current power consumption in watts
- **Energy Usage Tracking**: Monitor cumulative energy consumption
- **Device Control**: Turn devices on/off remotely via MQTT and API
- **Prometheus Storage**: Store time-series energy data for historical analysis
- **Grafana Dashboards**: Visualize energy consumption patterns
- **MQTT Integration**: Real-time energy data published to MQTT topics

### Quick Setup
1. **Configure your Tapo devices:**
   ```bash
   cp configs/tapo_template.yml configs/tapo.yml
   # Edit with your device IPs, usernames, and passwords
   ```

2. **Start monitoring:**
   ```bash
   cd cmd/tapo-demo && go run main.go
   ```

3. **View dashboards:**
   - Open Grafana: `http://YOUR_PI_IP:3000`
   - Navigate to "Tapo Smart Plug Energy Monitoring" dashboard

### Energy Data Topics
- **Energy Metrics**: `tapo/{device_id}/energy` (power, voltage, current, energy)
- **Device Control**: `tapo/{device_id}/control` (on/off commands)
- **Status Updates**: `tapo/{device_id}/status` (connectivity, signal strength)

### Supported Metrics
- Power consumption (watts)
- Cumulative energy usage (watt-hours)
- Voltage and current readings
- Device on/off state
- WiFi signal strength
- Device temperature (if available)

For detailed setup instructions, see:
- [docs/TAPO_ENERGY_MONITORING.md](docs/TAPO_ENERGY_MONITORING.md)
- [docs/TAPO_KLAP.md](docs/TAPO_KLAP.md) - KLAP protocol implementation for firmware 1.1.0+

## �🔧 Management & Monitoring

### Health Monitoring
```bash
cd deployments
./scripts/health-check.sh
```

### Backup & Restore
```bash
# Create backup
./scripts/backup.sh

# Restore from backup  
./scripts/restore.sh backup-filename.tar.gz
```

### Service Management
```bash
# View logs
docker compose logs -f

# Restart services
docker compose restart

# Update services
docker compose pull && docker compose up -d
```
- **Home Automation Server**: Main application (port 8080)
- **PostgreSQL**: Database storage (port 5432)
- **MQTT Broker**: IoT device communication (ports 1883, 9001)
- **Kafka**: Log streaming with KRaft mode (ports 9092, 9093)
- **Redis**: Caching layer (port 6379)
- **Grafana**: Monitoring dashboard (port 3000)

## 📡 API Endpoints

### Core System
- `GET /api/status` - System status
- `GET /api/devices` - List devices  
- `POST /api/devices/{id}/command` - Control devices
- `GET /api/sensors` - List sensors
- `GET /health` - Health check endpoint

### Smart Thermostat API (Coming Soon)
- `GET /api/thermostats` - List all thermostats
- `GET /api/thermostats/{id}` - Get thermostat details
- `PUT /api/thermostats/{id}/target` - Set target temperature (°F)
- `PUT /api/thermostats/{id}/mode` - Set operation mode
- `GET /api/thermostats/{id}/history` - Temperature history

## 📊 Logging & Monitoring

### Dual Logging System
The system implements a comprehensive logging approach optimized for Raspberry Pi 5:

- **File Logging**: Local logs stored in `logs/home-automation.log`  
- **Kafka Streaming**: Real-time log streaming to Kafka topics for centralized monitoring
- **Log Rotation**: Automatic rotation to prevent SD card wear

### Kafka Integration
- **Topic**: `home-automation-logs`
- **Format**: Structured JSON messages with metadata
- **KRaft Mode**: Modern Kafka without Zookeeper (reduced memory usage)
- **Optimized**: Limited retention and segment sizes for Pi 5

### Log Message Structure
```json
{
  "timestamp": "2025-07-16T10:30:15Z",
  "level": "INFO", 
  "service": "ThermostatService",
  "message": "Updated thermostat living-room: 68.5°F -> 69.2°F",
  "thermostat_id": "thermostat-001",
  "room_id": "living-room",
  "action": "temperature_update",
  "metadata": {
    "current_temp": 69.2,
    "target_temp": 70.0,
    "mode": "auto",
    "status": "heating",
    "unit": "fahrenheit"
  }
}
```

### Monitoring Capabilities
- **Thermostat Operations**: Temperature updates, mode changes, heating/cooling cycles
- **Motion Detection**: Room occupancy monitoring from PIR sensors
- **Device Control**: All device commands and status changes
- **Performance Metrics**: Command execution timing and success rates  
- **Error Tracking**: Centralized error collection and alerting
- **IoT Sensor Data**: Temperature (°F), humidity, and motion monitoring from Pi Pico sensors
- **System Health**: Raspberry Pi 5 resource monitoring (CPU, memory, temperature)

## 🛡️ **Production-Ready Error Handling & Reliability**

This system implements comprehensive error handling designed for production home automation environments.

### 🔧 **Error Handling System**
- **Structured Error Types**: Connection, Device, Service, System, and Validation errors with severity levels (Low, Medium, High, Critical)
- **Rich Error Context**: Automatic capture of device IDs, room IDs, stack traces, timestamps, and operational context
- **Error Classification**: Automatic determination of error retryability and criticality
- **Error Wrapping**: Support for error chains with full cause tracking

### 🔄 **Automatic Recovery Mechanisms**
- **Retry Logic**: Exponential backoff with jitter for transient failures
- **Circuit Breakers**: Prevent cascade failures and enable automatic recovery
- **Health Monitoring**: Continuous service health checks with detailed reporting
- **Graceful Degradation**: System continues operating with reduced functionality during partial failures

### 📊 **Monitoring & Observability**
- **Structured Logging**: JSON-formatted logs with automatic error context extraction
- **Kafka Integration**: High-severity errors automatically sent to Kafka for alerting
- **Health Checks**: Real-time monitoring of MQTT connections, service availability, and device status
- **Error Metrics**: Track error rates, retry success rates, and circuit breaker states

### 🚀 **Production Benefits**
- **🔍 Improved Debugging**: Rich error context makes issues easy to diagnose
- **⚡ Faster Recovery**: Automatic reconnection and healing for network issues
- **📈 Higher Reliability**: Circuit breakers prevent system-wide failures
- **🛡️ Better Monitoring**: Comprehensive health checks and structured error reporting

**📖 Complete Documentation**: See [ERROR_HANDLING.md](ERROR_HANDLING.md) for detailed implementation guide, examples, and migration instructions.

## 🏠 Features

## 🏠 Comprehensive Smart Home Features

### Multi-Sensor Environmental Control
- **Intelligent Temperature Control**: Automatic heating/cooling with hysteresis (Fahrenheit)
- **Motion Detection**: PIR sensor monitoring with occupancy tracking
- **Ambient Light Sensing**: Photo transistor monitoring with day/night detection
- **Multi-Zone Support**: Independent control for multiple rooms
- **Cross-Sensor Integration**: Occupancy + light level automation

### Pi Pico Multi-Sensor Platform
- **SHT-30 Integration**: Temperature/humidity sensors with real-time MQTT
- **PIR Motion Sensors**: Room occupancy detection and security monitoring
- **Photo Transistor Light Sensors**: Ambient light levels and circadian rhythm tracking
- **WiFi Connectivity**: Direct MQTT communication with Raspberry Pi 5
- **Multi-Room Deployment**: Scalable sensor network across your home

### Orthogonal Service Architecture
- **Thermostat Service**: HVAC control and temperature automation
- **Motion Service**: Occupancy tracking and presence detection
- **Light Service**: Ambient light monitoring and day/night cycles
- **Integrated Service**: Optional cross-sensor automation and energy optimization

### Core Infrastructure
- **Device Management**: Control lights, switches, climate systems
- **MQTT Integration**: Standard IoT communication protocol  
- **Kafka Logging**: Real-time log streaming optimized for Pi 5
- **REST API**: Complete HTTP API for integrations
- **Web Dashboard**: Modern, responsive web interface with Grafana
- **CLI Tools**: Command-line utilities for administration

### Device Types Supported
- **Smart Thermostats**: Automatic temperature control (Fahrenheit)
- **Motion Sensors**: PIR motion detection with MQTT alerts and occupancy tracking
- **Light Sensors**: Photo transistor ambient light monitoring
- **Smart Lights**: On/off, dimming, color control with automatic control
- **Switches**: Simple on/off control  
- **Climate Systems**: Temperature and mode control with energy optimization
- **IoT Sensors**: Pi Pico WH with SHT-30 + PIR + photo transistor multi-sensor nodes

### Advanced Automation Features
- **Energy Optimization**: Reduce HVAC when rooms unoccupied or naturally lit
- **Circadian Rhythm Support**: Light-based automation for health and wellness
- **Security Integration**: Motion alerts and unusual activity detection
- **Adaptive Scheduling**: Learn occupancy patterns for proactive climate control
- **Cross-Sensor Logic**: Complex automation rules using multiple sensor inputs

### Raspberry Pi 5 Architecture  
- **Optimized Performance**: Resource limits and efficient memory usage
- **Containerized Services**: Docker Compose with Pi-specific optimizations
- **SD Card Friendly**: Log rotation and optimized write patterns
- **Low Power**: Efficient service configuration for 24/7 operation
- **Scalable**: Support for multiple Pi Pico sensors across rooms

## 🔧 Configuration

### Multi-Sensor Configuration
The services use Fahrenheit and optimized settings by default:

```go
// Thermostat defaults (Fahrenheit)
DefaultTargetTemp    = 70.0°F  // Comfortable room temperature
DefaultHysteresis    = 1.0°F   // Prevents short cycling  
DefaultMinTemp       = 50.0°F  // Safety minimum
DefaultMaxTemp       = 95.0°F  // Safety maximum

// Motion sensor defaults
PIR_DEBOUNCE_TIME    = 2 sec   // Prevent rapid triggering
PIR_TIMEOUT          = 30 sec  // Motion clear delay

// Light sensor defaults  
LIGHT_THRESHOLD_LOW  = 10%     // Below = dark
LIGHT_THRESHOLD_HIGH = 80%     // Above = bright
```

### Main Configuration
Edit `configs/config.yaml` or `.env` to customize:
- Server settings (port, timeouts)
- Database configuration  
- MQTT broker settings
- Kafka logging configuration
- Raspberry Pi specific optimizations

### Pi Pico Multi-Sensor Configuration
Edit `firmware/pico-sht30/config.py`:
- WiFi credentials
- Raspberry Pi 5 MQTT broker IP
- Room assignment and device naming
- Sensor enable/disable flags (PIR, light sensor)
- Sensor thresholds and timing
- Temperature unit (Fahrenheit by default)
- Reading intervals and GPIO pins
- Logging configuration

## 📚 Documentation

See the `docs/` directory for detailed documentation:
- `THERMOSTAT.md` - Complete smart thermostat guide
- `FAHRENHEIT_CONVERSION.md` - Temperature conversion details
- Architecture overview
- API reference
- Device integration guide
- Deployment instructions

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run `make test` and `make lint`
6. Submit a pull request

## 📄 License

This project is licensed under the MIT License.

## 🆘 Support

For questions and support:
- Check the documentation in `docs/`
- Review the API specification in `api/openapi.yaml`
- Open an issue on GitHub
- Review the example configurations in `configs/`

---

🌡️ **Smart Thermostat System**: Your home automation now includes intelligent temperature control using Fahrenheit, with Pi Pico sensors and automatic heating/cooling management optimized for Raspberry Pi 5!

**Note**: This is a complete Go project structure following best practices for home automation systems. The code includes working examples for devices, sensors, API handlers, smart thermostat control, and a modern web interface.