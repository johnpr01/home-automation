# Home Automation Project

A comprehensive Go-based home automation system with web interface, REST API, and MQTT support.

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
└── deployments/          # Deployment configurations
    └── docker-compose.yml # Docker Compose setup
```

## 🚀 Quick Start

1. **Clone and setup:**
   ```bash
   cd /home/philip/home-automation
   ./scripts/setup.sh
   ```

2. **Run the server:**
   ```bash
   make run-server
   ```

3. **Access the web interface:**
   Open http://localhost:8080 in your browser

4. **Use the CLI:**
   ```bash
   make run-cli
   ./bin/home-automation-cli -cmd status
   ```

## 🛠️ Development Commands

- `make build` - Build all binaries
- `make test` - Run tests
- `make fmt` - Format code
- `make lint` - Lint code (requires golangci-lint)
- `make dev` - Run with hot reload (requires air)
- `make help` - Show all available commands

## 🐳 Docker Support

```bash
# Build and run with Docker Compose
cd deployments
docker-compose up --build
```

## 📡 API Endpoints

- `GET /api/status` - System status
- `GET /api/devices` - List devices
- `POST /api/devices/{id}/command` - Control devices
- `GET /api/sensors` - List sensors

## 🏠 Features

### Core Components
- **Device Management**: Control lights, switches, climate systems
- **Sensor Monitoring**: Temperature, humidity, motion, and more
- **MQTT Integration**: Standard IoT communication protocol
- **REST API**: Complete HTTP API for integrations
- **Web Dashboard**: Modern, responsive web interface
- **CLI Tools**: Command-line utilities for administration

### Device Types Supported
- **Lights**: On/off, dimming, color control
- **Switches**: Simple on/off control
- **Climate**: Temperature and mode control
- **Sensors**: Various sensor types with real-time readings

### Architecture
- **Modular Design**: Clean separation of concerns
- **Extensible**: Easy to add new device types
- **Production Ready**: Docker support, logging, configuration management
- **Testing**: Comprehensive test suite

## 🔧 Configuration

Edit `configs/config.yaml` to customize:
- Server settings (port, timeouts)
- Database configuration
- MQTT broker settings
- Device discovery options
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