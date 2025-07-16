# Home Automation Project

A comprehensive Go-based home automation system optimized for **Raspberry Pi 5** deployment, featuring web interface, REST API, MQTT support, and Kafka logging.

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
│   │   └── sensor.go
│   └── services/          # Business logic
│       └── device_service.go
│
├── pkg/                    # Public library code
│   ├── devices/           # Device implementations
│   │   └── light.go
│   ├── mqtt/              # MQTT client
│   │   └── client.go
│   ├── kafka/             # Kafka client for logging
│   │   └── client.go
│   └── sensors/           # Sensor implementations
│       └── temperature.go
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
│   └── README.md         # Documentation index
│
├── test/                  # Test files
│   └── device_test.go    # Example tests
│
├── firmware/             # IoT device firmware
│   └── pico-sht30/      # Pi Pico WH SHT-30 sensor firmware
│       ├── main.py      # MicroPython main application
│       ├── sht30.py     # SHT-30 sensor driver
│       ├── config_template.py # Configuration template
│       ├── deploy.sh    # Firmware deployment script
│       └── README.md    # Firmware documentation
│
└── deployments/          # Raspberry Pi 5 deployment
    ├── docker-compose.yml # Optimized for Pi 5
    ├── deploy-pi5.sh     # Automated Pi 5 deployment
    ├── scripts/          # Management scripts
    │   ├── health-check.sh # System health monitoring
    │   ├── backup.sh     # Backup script
    │   └── restore.sh    # Restore script
    └── README.md         # Pi 5 deployment guide
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

- `make build` - Build all binaries
- `make test` - Run tests
- `make fmt` - Format code
- `make lint` - Lint code (requires golangci-lint)
- `make dev` - Run with hot reload (requires air)
- `make help` - Show all available commands

## 🐳 Raspberry Pi 5 Services

The system runs the following services optimized for Raspberry Pi 5:

### Core Services:
- **Home Automation API** (Port 8080) - Main application server
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

## 📱 IoT Device Support

### Pi Pico WH + SHT-30 Sensor

Deploy temperature and humidity sensors throughout your home:

1. **Setup the sensor:**
   ```bash
   cd firmware/pico-sht30
   cp config_template.py config.py
   # Edit config.py with your Pi 5 IP and WiFi settings
   ```

2. **Deploy firmware:**
   ```bash
   ./deploy.sh
   ```

3. **Monitor sensor data:**
   ```bash
   python3 mqtt_monitor.py YOUR_PI5_IP
   ```

### MQTT Topics:
- Temperature: `room-temp/{room_number}`
- Humidity: `room-hum/{room_number}`

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

- `GET /api/status` - System status
- `GET /api/devices` - List devices  
- `POST /api/devices/{id}/command` - Control devices
- `GET /api/sensors` - List sensors
- `GET /health` - Health check endpoint

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
  "timestamp": "2025-07-14T10:30:15Z",
  "level": "INFO", 
  "service": "DeviceService",
  "message": "Temperature sensor reading: 22.5°C",
  "device_id": "light-001",
  "action": "turn_on",
  "metadata": {
    "status": "on",
    "power": true,
    "device_type": "light"
  }
  "room": "living-room",
  "device_id": "pico-sht30-room1"
}
```

### Monitoring Capabilities
- **Device Operations**: All device commands and status changes
- **Performance Metrics**: Command execution timing and success rates  
- **Error Tracking**: Centralized error collection and alerting
- **IoT Sensor Data**: Temperature, humidity monitoring from Pi Pico sensors
- **System Health**: Raspberry Pi 5 resource monitoring (CPU, memory, temperature)

## 🏠 Features

### Core Components
- **Device Management**: Control lights, switches, climate systems
- **IoT Sensor Network**: Pi Pico WH sensors with SHT-30 temperature/humidity
- **MQTT Integration**: Standard IoT communication protocol  
- **Kafka Logging**: Real-time log streaming optimized for Pi 5
- **REST API**: Complete HTTP API for integrations
- **Web Dashboard**: Modern, responsive web interface with Grafana
- **CLI Tools**: Command-line utilities for administration

### Device Types Supported
- **Lights**: On/off, dimming, color control
- **Switches**: Simple on/off control  
- **Climate**: Temperature and mode control
- **IoT Sensors**: Pi Pico WH with SHT-30 (temperature/humidity)
- **Environmental**: Various sensor types with real-time readings

### Raspberry Pi 5 Architecture  
- **Optimized Performance**: Resource limits and efficient memory usage
- **Containerized Services**: Docker Compose with Pi-specific optimizations
- **SD Card Friendly**: Log rotation and optimized write patterns
- **Low Power**: Efficient service configuration for 24/7 operation
- **Scalable**: Support for multiple Pi Pico sensors across rooms

## 🔧 Configuration

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
- Reading intervals and GPIO pins
- Logging configuration

## 📚 Documentation

See the `docs/` directory for detailed documentation:
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

**Note**: This is a complete Go project structure following best practices for home automation systems. The code includes working examples for devices, sensors, API handlers, and a modern web interface.