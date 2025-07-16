# Home Automation Project

A comprehensive Go-based home automation system with **Smart Thermostat Framework** optimized for **Raspberry Pi 5** deployment, featuring IoT sensor integration, MQTT communication, and intelligent temperature control using **Fahrenheit**.

## 🌡️ Smart Thermostat System

**NEW: Complete smart thermostat framework with Pi Pico WH sensors!**

- **Intelligent Temperature Control**: Automatic heating/cooling with hysteresis
- **Pi Pico Integration**: SHT-30 temperature/humidity sensors via MQTT
- **Fahrenheit Operation**: All temperatures in Fahrenheit for US users
- **Multi-Zone Support**: Control multiple rooms independently
- **Real-time Processing**: 30-second control loops with instant sensor response

### Quick Thermostat Setup

1. **Deploy Pi Pico sensors:**
   ```bash
   cd firmware/pico-sht30
   # Configure WiFi and MQTT settings
   cp config_template.py config.py
   # Flash to Pi Pico WH with SHT-30 sensor
   ```

2. **Start thermostat service:**
   ```bash
   cd cmd/thermostat
   go build && ./thermostat
   ```

3. **Monitor temperature control:**
   - Listens to: `room-temp/{room_id}` and `room-hum/{room_id}`
   - Controls: Automatic heating/cooling based on target temperature
   - Default: 70°F target with 1°F hysteresis

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
│       └── thermostat_service.go # Thermostat control logic
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
│   └── FAHRENHEIT_CONVERSION.md # Fahrenheit conversion details
│
├── test/                  # Test files
│   └── device_test.go    # Example tests
│
├── firmware/             # IoT device firmware
│   └── pico-sht30/      # Pi Pico WH SHT-30 sensor firmware
│       ├── main.py      # MicroPython main application (Fahrenheit)
│       ├── sht30.py     # SHT-30 sensor driver
│       ├── config_template.py # Configuration template
│       ├── deploy.sh    # Firmware deployment script
│       └── README.md    # Firmware documentation
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
- **Smart Thermostat Service** - Intelligent temperature control (Fahrenheit)
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

### Pi Pico Integration
Deploy SHT-30 temperature/humidity sensors throughout your home:

1. **Configure sensor:**
   ```bash
   cd firmware/pico-sht30
   cp config_template.py config.py
   # Edit with your Pi 5 IP and room assignment
   ```

2. **Flash firmware:**
   ```bash
   # Copy files to Pi Pico WH
   # Sensor automatically sends Fahrenheit temperatures
   ```

3. **Monitor thermostat:**
   ```bash
   # Thermostat service logs show real-time control decisions
   cd cmd/thermostat && go run main.go
   ```

### MQTT Topics (Fahrenheit):
- **Temperature**: `room-temp/{room_number}` (°F)
- **Humidity**: `room-hum/{room_number}` (%)
- **Control**: `room-control/{room_number}` (heating/cooling commands)

### Example Operation:
**Target: 70°F, Hysteresis: 1°F**
- 🔥 Heat ON: Temperature drops below 69.5°F
- 🔥 Heat OFF: Temperature reaches 70°F
- ❄️ Cool ON: Temperature rises above 70.5°F  
- ❄️ Cool OFF: Temperature reaches 70°F

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
- **Device Control**: All device commands and status changes
- **Performance Metrics**: Command execution timing and success rates  
- **Error Tracking**: Centralized error collection and alerting
- **IoT Sensor Data**: Temperature (°F), humidity monitoring from Pi Pico sensors
- **System Health**: Raspberry Pi 5 resource monitoring (CPU, memory, temperature)

## 🏠 Features

### Smart Thermostat System
- **Intelligent Temperature Control**: Automatic heating/cooling with hysteresis (Fahrenheit)
- **Multi-Zone Support**: Independent control for multiple rooms
- **Pi Pico Integration**: SHT-30 sensors with real-time MQTT communication
- **Advanced Control Logic**: Prevents short cycling, maintains comfort
- **Safety Features**: Min/max temperature limits, sensor validation
- **Energy Optimization**: Efficient heating/cooling cycles

### Core Components
- **Device Management**: Control lights, switches, climate systems
- **IoT Sensor Network**: Pi Pico WH sensors with SHT-30 temperature/humidity
- **MQTT Integration**: Standard IoT communication protocol  
- **Kafka Logging**: Real-time log streaming optimized for Pi 5
- **REST API**: Complete HTTP API for integrations
- **Web Dashboard**: Modern, responsive web interface with Grafana
- **CLI Tools**: Command-line utilities for administration

### Device Types Supported
- **Smart Thermostats**: Automatic temperature control (Fahrenheit)
- **Lights**: On/off, dimming, color control
- **Switches**: Simple on/off control  
- **Climate**: Temperature and mode control
- **IoT Sensors**: Pi Pico WH with SHT-30 (temperature/humidity in °F)
- **Environmental**: Various sensor types with real-time readings

### Raspberry Pi 5 Architecture  
- **Optimized Performance**: Resource limits and efficient memory usage
- **Containerized Services**: Docker Compose with Pi-specific optimizations
- **SD Card Friendly**: Log rotation and optimized write patterns
- **Low Power**: Efficient service configuration for 24/7 operation
- **Scalable**: Support for multiple Pi Pico sensors across rooms

## 🔧 Configuration

### Smart Thermostat Configuration
The thermostat service uses Fahrenheit by default. Key settings:

```go
// Default Fahrenheit values
DefaultTargetTemp    = 70.0°F  // Comfortable room temperature
DefaultHysteresis    = 1.0°F   // Prevents short cycling  
DefaultMinTemp       = 50.0°F  // Safety minimum
DefaultMaxTemp       = 95.0°F  // Safety maximum
```

### Main Configuration
Edit `configs/config.yaml` or `.env` to customize:
- Server settings (port, timeouts)
- Database configuration  
- MQTT broker settings
- Kafka logging configuration
- Raspberry Pi specific optimizations

### IoT Sensor Configuration
Edit `firmware/pico-sht30/config.py`:
- WiFi credentials
- Raspberry Pi 5 MQTT broker IP
- Room assignment and device naming
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