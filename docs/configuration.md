# Configuration Reference

This document provides a comprehensive reference for configuring the Home Automation system.

## Configuration File Structure

The main configuration is stored in `configs/config.yaml`:

```yaml
server:
  port: "8080"
  host: "0.0.0.0"
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "60s"

database:
  type: "sqlite"
  path: "home_automation.db"

mqtt:
  broker: "localhost"
  port: "1883"
  client_id: "home-automation-server"
  username: ""
  password: ""
  keep_alive: 60
  topics:
    device_commands: "homeautomation/devices/+/commands"
    device_status: "homeautomation/devices/+/status"
    sensor_readings: "homeautomation/sensors/+/readings"

kafka:
  brokers: ["localhost:9092"]
  log_topic: "home-automation-logs"
  client_id: "home-automation-logger"
  batch_size: 100
  timeout: "5s"

devices:
  discovery:
    enabled: true
    interval: "5m"
  auto_configure: true
  default_timeout: "30s"

sensors:
  reading_interval: "1m"
  history_retention: "30d"
  alert_thresholds:
    temperature:
      min: 15.0
      max: 30.0
    humidity:
      min: 30.0
      max: 70.0

logging:
  level: "info"
  format: "json"
  output: "stdout"
  file: "logs/home-automation.log"

security:
  enable_auth: false
  jwt_secret: "your-secret-key-here"
  token_expiry: "24h"
  cors:
    enabled: true
    origins: ["*"]
    methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    headers: ["*"]

features:
  web_ui: true
  api: true
  mqtt: true
  kafka_logging: true
  automation_rules: true
  scheduling: true
```

## Environment Variables

Configuration can be overridden using environment variables:

### Server Configuration
- `PORT`: Server port (default: 8080)
- `HOST`: Server host (default: 0.0.0.0)

### Database Configuration
- `DATABASE_URL`: Database connection string
- `DATABASE_TYPE`: Database type (sqlite, postgres)

### MQTT Configuration
- `MQTT_BROKER`: MQTT broker hostname
- `MQTT_PORT`: MQTT broker port
- `MQTT_USERNAME`: MQTT username
- `MQTT_PASSWORD`: MQTT password

### Kafka Configuration
- `KAFKA_BROKERS`: Comma-separated list of Kafka brokers
- `KAFKA_LOG_TOPIC`: Topic for log messages
- `KAFKA_CLIENT_ID`: Kafka client identifier

### Security Configuration
- `JWT_SECRET`: Secret key for JWT tokens
- `ENABLE_AUTH`: Enable authentication (true/false)

## Section Details

### Server Configuration

Controls the HTTP server behavior:

```yaml
server:
  port: "8080"              # Port to bind to
  host: "0.0.0.0"          # Host to bind to (0.0.0.0 for all interfaces)
  read_timeout: "30s"       # Request read timeout
  write_timeout: "30s"      # Response write timeout
  idle_timeout: "60s"       # Keep-alive timeout
```

### Database Configuration

Database connection settings:

```yaml
database:
  type: "sqlite"                    # Database type: sqlite, postgres
  path: "home_automation.db"        # SQLite file path
  # For PostgreSQL:
  # host: "localhost"
  # port: "5432"
  # name: "home_automation"
  # username: "admin"
  # password: "password"
```

### MQTT Configuration

MQTT broker connection and topic configuration:

```yaml
mqtt:
  broker: "localhost"               # MQTT broker hostname
  port: "1883"                     # MQTT broker port
  client_id: "home-automation-server"  # Unique client identifier
  username: ""                     # MQTT username (optional)
  password: ""                     # MQTT password (optional)
  keep_alive: 60                   # Keep-alive interval in seconds
  topics:
    device_commands: "homeautomation/devices/+/commands"
    device_status: "homeautomation/devices/+/status"
    sensor_readings: "homeautomation/sensors/+/readings"
```

### Kafka Configuration

Kafka logging system configuration:

```yaml
kafka:
  brokers: ["localhost:9092"]      # List of Kafka broker addresses
  log_topic: "home-automation-logs"  # Topic for log messages
  client_id: "home-automation-logger"  # Kafka client identifier
  batch_size: 100                  # Number of messages to batch
  timeout: "5s"                    # Producer timeout
```

### Device Configuration

Device discovery and management settings:

```yaml
devices:
  discovery:
    enabled: true                  # Enable automatic device discovery
    interval: "5m"                # Discovery scan interval
  auto_configure: true            # Automatically configure discovered devices
  default_timeout: "30s"          # Default command timeout
```

### Sensor Configuration

Sensor monitoring and alerting settings:

```yaml
sensors:
  reading_interval: "1m"          # How often to read sensor values
  history_retention: "30d"        # How long to keep sensor history
  alert_thresholds:
    temperature:
      min: 15.0                   # Minimum temperature alert threshold
      max: 30.0                   # Maximum temperature alert threshold
    humidity:
      min: 30.0                   # Minimum humidity alert threshold
      max: 70.0                   # Maximum humidity alert threshold
```

### Logging Configuration

File logging configuration (separate from Kafka):

```yaml
logging:
  level: "info"                   # Log level: debug, info, warn, error
  format: "json"                  # Log format: json, text
  output: "stdout"                # Output: stdout, stderr, file
  file: "logs/home-automation.log"  # Log file path when output is file
```

### Security Configuration

Authentication and authorization settings:

```yaml
security:
  enable_auth: false              # Enable JWT authentication
  jwt_secret: "your-secret-key-here"  # JWT signing secret
  token_expiry: "24h"            # JWT token expiration time
  cors:
    enabled: true                 # Enable CORS
    origins: ["*"]               # Allowed origins
    methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    headers: ["*"]               # Allowed headers
```

### Feature Flags

Enable/disable system features:

```yaml
features:
  web_ui: true                    # Enable web dashboard
  api: true                       # Enable REST API
  mqtt: true                      # Enable MQTT integration
  kafka_logging: true             # Enable Kafka logging
  automation_rules: true          # Enable automation rules
  scheduling: true                # Enable device scheduling
```

## Docker Configuration

When using Docker Compose, configuration is primarily done through environment variables:

```yaml
environment:
  - PORT=8080
  - DATABASE_URL=postgres://admin:password@postgres:5432/home_automation?sslmode=disable
  - MQTT_BROKER=mosquitto
  - MQTT_PORT=1883
  - KAFKA_BROKERS=kafka:9092
  - KAFKA_LOG_TOPIC=home-automation-logs
```

## Configuration Validation

The system validates configuration on startup and will report errors for:

- Invalid port numbers
- Unreachable database connections
- Invalid MQTT broker addresses
- Invalid Kafka broker addresses
- Missing required configuration values

## Production Considerations

### Security Hardening

For production deployments:

1. **Enable Authentication**:
   ```yaml
   security:
     enable_auth: true
     jwt_secret: "strong-random-secret-key"
   ```

2. **Restrict CORS**:
   ```yaml
   security:
     cors:
       origins: ["https://yourdomain.com"]
   ```

3. **Use Environment Variables** for sensitive data instead of config files

### Performance Tuning

1. **Database Connection Pooling**:
   ```yaml
   database:
     max_connections: 25
     max_idle_connections: 5
   ```

2. **Kafka Optimization**:
   ```yaml
   kafka:
     batch_size: 1000
     compression: "gzip"
     max_message_bytes: 1000000
   ```

3. **Sensor Reading Optimization**:
   ```yaml
   sensors:
     reading_interval: "30s"  # Reduce for high-frequency monitoring
     batch_readings: true     # Enable batching for performance
   ```

### High Availability

For HA deployments:

```yaml
database:
  type: "postgres"
  host: "postgres-cluster-endpoint"
  read_replicas: ["replica1:5432", "replica2:5432"]

kafka:
  brokers: ["kafka1:9092", "kafka2:9092", "kafka3:9092"]
  replication_factor: 3

mqtt:
  cluster: true
  brokers: ["mqtt1:1883", "mqtt2:1883"]
```

## Configuration Examples

### Development Environment

```yaml
server:
  port: "8080"
  
database:
  type: "sqlite"
  path: "dev.db"
  
mqtt:
  broker: "localhost"
  port: "1883"
  
kafka:
  brokers: ["localhost:9092"]
  
logging:
  level: "debug"
  
security:
  enable_auth: false
```

### Production Environment

```yaml
server:
  port: "8080"
  read_timeout: "10s"
  write_timeout: "10s"
  
database:
  type: "postgres"
  host: "postgres.internal"
  port: "5432"
  name: "home_automation"
  
mqtt:
  broker: "mqtt.internal"
  port: "1883"
  username: "${MQTT_USERNAME}"
  password: "${MQTT_PASSWORD}"
  
kafka:
  brokers: ["kafka1.internal:9092", "kafka2.internal:9092"]
  
logging:
  level: "info"
  format: "json"
  
security:
  enable_auth: true
  jwt_secret: "${JWT_SECRET}"
```

## Troubleshooting

### Common Configuration Issues

1. **Port Already in Use**:
   - Change the port in configuration
   - Check for conflicting services

2. **Database Connection Failures**:
   - Verify database server is running
   - Check connection string format
   - Verify credentials

3. **MQTT Connection Issues**:
   - Check broker address and port
   - Verify network connectivity
   - Check authentication credentials

4. **Kafka Connection Problems**:
   - Verify broker addresses
   - Check network connectivity
   - Ensure topics exist or auto-creation is enabled

### Configuration Validation Commands

```bash
# Validate configuration file
./bin/home-automation-cli -cmd validate-config

# Test database connection
./bin/home-automation-cli -cmd test-db

# Test MQTT connection
./bin/home-automation-cli -cmd test-mqtt

# Test Kafka connection
./bin/home-automation-cli -cmd test-kafka
```
