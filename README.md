# Home Automation System

A comprehensive home automation system built in Go, designed to run on Raspberry Pi 5 with integrated sensor networks, MQTT messaging, and intelligent thermostat control.

## 🏠 System Overview

This system provides a unified platform for managing home automation devices with real-time environmental monitoring, motion detection, and intelligent climate control.

### Key Features

- **Multi-Sensor Integration**: Temperature, humidity, motion, and light sensors on a single Pi Pico device
- **Unified Sensor Service**: Centralized management of all sensor data with intelligent aggregation
- **Smart Thermostat**: Fahrenheit-based climate control with occupancy awareness
- **Real-time MQTT**: Low-latency sensor data transmission and device control
- **Microcontroller Sensors**: Pi Pico WH with SHT-30, PIR, and photo transistor sensors
- **Container Orchestration**: Docker Compose with optimized resource allocation
- **Message Streaming**: Kafka integration for data persistence and analytics
- **Motion Detection**: PIR sensor monitoring with room occupancy tracking
- **Ambient Light Sensing**: Photo transistor monitoring with day/night cycle detection
- **Pi Pico Integration**: SHT-30, PIR, and photo transistor sensors via MQTT
- **Multi-Zone Support**: Control multiple rooms independently
- **Orthogonal Architecture**: Services operate independently but can integrate when needed

### 🏗️ **Service Architecture:**
- **Thermostat Service**: HVAC temperature control and automation
- **Motion Service**: PIR sensor monitoring and occupancy detection
- **Light Service**: Photo transistor ambient light tracking
- **Integrated Service**: Optional combined service with cross-sensor automation

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
   
   # Option D: All services with integration
   cd cmd/integrated && go run main.go
   ```

3. **Monitor smart home control:**
   - **Temperature**: `room-temp/{room_id}` → Automatic HVAC control
   - **Motion**: `room-motion/{room_id}` → Occupancy tracking
   - **Light**: `room-light/{room_id}` → Ambient light monitoring
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
│   ├── integrated/         # Optional integrated service
│   │   └── main.go         # Combined motion + thermostat with callbacks
│   ├── temp-demo/          # Temperature conversion demo
│   │   └── main.go
│   └── cli/                # Command-line interface
│       └── main.go
│
├── internal/               # Private application code
│   ├── config/            # Configuration management
│   │   └── config.go
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
│       └── light_service.go      # Light sensor monitoring and ambient light tracking
│
├── pkg/                    # Public library code
│   ├── devices/           # Device implementations
│   │   └── light.go
│   ├── mqtt/              # MQTT client
│   │   └── client.go
│   ├── kafka/             # Kafka client for logging
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
│   └── config.yaml       # Default configuration
│
├── scripts/               # Development and deployment scripts
│   └── setup.sh          # Development environment setup
│
├── docs/                  # Documentation
│   ├── README.md         # Documentation index
│   ├── THERMOSTAT.md     # Smart thermostat guide
│   ├── MOTION_DETECTION.md # PIR motion sensor guide
│   ├── LIGHT_SENSOR.md   # Photo transistor light sensor guide
│   └── FAHRENHEIT_CONVERSION.md # Fahrenheit conversion details
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
│   ├── docker-compose.yml # Optimized for Pi 5
│   ├── deploy-pi5.sh     # Automated Pi 5 deployment
│   ├── mosquitto/        # MQTT broker configuration
│   │   ├── mosquitto.conf # Mosquitto configuration
│   │   ├── acl.example   # Access control template
│   │   └── passwd.example # Password file template
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
   cd home-automation/deployments
   ./deploy-pi5.sh
   ```

3. **Access your services:**
   - **Home Automation API**: `http://YOUR_PI_IP:8080`
   - **Smart Thermostat**: Automatic control via MQTT
   - **Grafana Dashboard**: `http://YOUR_PI_IP:3000` (admin/homeauto2024)
   - **MQTT Broker**: `YOUR_PI_IP:1883`

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

## 🛠️ Development Commands

- `make build` - Build all binaries (including thermostat service)
- `make test` - Run tests (including temperature conversion tests)
- `make fmt` - Format code
- `make lint` - Lint code (requires golangci-lint)
- `make dev` - Run with hot reload (requires air)
- `make help` - Show all available commands

### Thermostat Development
- `go run ./cmd/thermostat/` - Run thermostat service locally
- `go run ./cmd/temp-demo/` - Demo temperature conversions
- `go test ./pkg/utils/` - Test temperature conversion utilities

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
- **Motion**: `room-motion/{room_number}` (occupancy) → Presence detection
- **Light**: `room-light/{room_number}` (%) → Ambient light levels
- **Control**: `thermostat/{thermostat_id}/control` (HVAC commands)

### Example Multi-Sensor Operation:
**Smart Home Intelligence**
- 🌡️ **HVAC**: Target 70°F ±1°F hysteresis with automatic heating/cooling
- 👥 **Occupancy**: Motion detection for energy-saving and security
- 🌞 **Lighting**: Ambient light monitoring for automatic lighting control
- 🏠 **Integration**: Cross-sensor automation (e.g., occupied + dark = lights on)

## 🔧 Management & Monitoring

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