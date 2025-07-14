# Home Automation System Documentation

This directory contains comprehensive documentation for the home automation system.

## Table of Contents

- [Architecture Overview](./architecture.md)
- [API Documentation](./api.md)
- [Device Integration Guide](./devices.md)
- [MQTT Protocol](./mqtt.md)
- [Kafka Logging System](./kafka-logging.md)
- [Configuration Reference](./configuration.md)
- [Deployment Guide](./deployment.md)
- [Development Setup](./development.md)
- [Contributing Guidelines](./contributing.md)

## Quick Start

1. **Installation**: See [Development Setup](./development.md)
2. **Configuration**: Review [Configuration Reference](./configuration.md)
3. **API Usage**: Check [API Documentation](./api.md)
4. **Device Integration**: Follow [Device Integration Guide](./devices.md)
5. **Logging Setup**: Configure [Kafka Logging System](./kafka-logging.md)

## Architecture

The system follows a modular architecture with the following components:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web UI/CLI    │    │   REST API      │    │   MQTT Broker   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
         ┌─────────────────────────────────────────────────┐
         │              Core Services                      │
         │  ┌─────────────┐  ┌─────────────┐              │
         │  │   Device    │  │   Sensor    │              │
         │  │  Service    │  │  Service    │              │
         │  └─────────────┘  └─────────────┘              │
         └─────────────────────────────────────────────────┘
                                 │
         ┌─────────────────────────────────────────────────┐
         │              Data Layer                         │
         │  ┌─────────────┐  ┌─────────────┐              │
         │  │  Database   │  │    Cache    │              │
         │  │ (SQLite/PG) │  │   (Redis)   │              │
         │  └─────────────┘  └─────────────┘              │
         └─────────────────────────────────────────────────┘
                                 │
         ┌─────────────────────────────────────────────────┐
         │            Monitoring Layer                     │
         │  ┌─────────────┐  ┌─────────────┐              │
         │  │    Kafka    │  │   Grafana   │              │
         │  │  (Logging)  │  │(Monitoring) │              │
         │  └─────────────┘  └─────────────┘              │
         └─────────────────────────────────────────────────┘
```

## Key Features

- **Device Management**: Control lights, switches, climate systems, and more
- **Sensor Monitoring**: Real-time data collection from various sensors
- **MQTT Integration**: Standardized communication protocol for IoT devices
- **Kafka Logging**: Real-time log streaming and centralized monitoring
- **REST API**: Complete HTTP API for integration with external systems
- **Web Interface**: User-friendly web dashboard
- **CLI Tools**: Command-line utilities for system administration
- **Docker Support**: Containerized deployment with Docker Compose
- **Extensible Architecture**: Plugin system for custom device types

## Support

For questions, issues, or contributions:

1. Check existing documentation
2. Review the [Contributing Guidelines](./contributing.md)
3. Open an issue on the project repository
4. Join our community discussions

## License

This project is licensed under the MIT License. See the LICENSE file for details.
