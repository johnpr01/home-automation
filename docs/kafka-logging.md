# Kafka Logging System

The Home Automation system implements a sophisticated dual logging approach using both local file logging and real-time Kafka streaming for centralized monitoring and analytics.

## Overview

The logging system captures all device operations, system events, and errors in structured JSON format, making it easy to monitor, analyze, and alert on system behavior.

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Application   │───▶│  Local Logging  │    │  Kafka Cluster  │
│   Components    │    │   (File-based)  │    │   (Streaming)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       ▲
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────────────┐
                    │     Kafka Client        │
                    │  (Structured Logging)   │
                    └─────────────────────────┘
                                 │
                    ┌─────────────────────────┐
                    │   Log Aggregation       │
                    │   • ELK Stack           │
                    │   • Grafana             │
                    │   • Custom Analytics    │
                    └─────────────────────────┘
```

## Configuration

### Kafka Settings

In `configs/config.yaml`:

```yaml
kafka:
  brokers: ["localhost:9092"]
  log_topic: "home-automation-logs"
  client_id: "home-automation-logger"
  batch_size: 100
  timeout: "5s"
```

### Environment Variables

For Docker deployment:

```bash
KAFKA_BROKERS=kafka:9092
KAFKA_LOG_TOPIC=home-automation-logs
KAFKA_CLIENT_ID=home-automation-logger
```

## Message Format

All log messages follow a consistent JSON structure:

```json
{
  "timestamp": "2025-07-14T10:30:15Z",
  "level": "INFO|WARN|ERROR",
  "service": "DeviceService",
  "message": "Human-readable message",
  "device_id": "device-identifier",
  "action": "command-or-operation",
  "metadata": {
    "key": "value",
    "additional": "context"
  }
}
```

### Field Descriptions

- **timestamp**: ISO 8601 UTC timestamp
- **level**: Log level (INFO, WARN, ERROR)
- **service**: Source service name
- **message**: Human-readable description
- **device_id**: Device identifier (optional)
- **action**: Action or command being performed (optional)
- **metadata**: Additional context data (optional)

## Log Categories

### Device Operations

```json
{
  "timestamp": "2025-07-14T10:30:15Z",
  "level": "INFO",
  "service": "DeviceService",
  "message": "Light living-room-light turned on",
  "device_id": "living-room-light",
  "action": "turn_on",
  "metadata": {
    "device_type": "light",
    "status": "on",
    "power": true
  }
}
```

### Temperature Monitoring

```json
{
  "timestamp": "2025-07-14T10:30:20Z",
  "level": "INFO",
  "service": "DeviceService",
  "message": "Current temperature for device climate-001: 22.50",
  "device_id": "climate-001",
  "action": "get_temperature",
  "metadata": {
    "temperature": 22.5,
    "device_type": "climate"
  }
}
```

### MQTT Operations

```json
{
  "timestamp": "2025-07-14T10:30:21Z",
  "level": "INFO",
  "service": "DeviceService",
  "message": "Successfully published temperature 22.50 to MQTT topic 'temp'",
  "device_id": "climate-001",
  "action": "mqtt_publish",
  "metadata": {
    "temperature": 22.5,
    "mqtt_topic": "temp",
    "qos": 1
  }
}
```

### Error Tracking

```json
{
  "timestamp": "2025-07-14T10:30:25Z",
  "level": "ERROR",
  "service": "DeviceService",
  "message": "Failed to execute command: device unknown-device not found",
  "device_id": "unknown-device",
  "action": "turn_on",
  "metadata": {
    "error_type": "device_not_found",
    "command_value": null
  }
}
```

## Kafka Deployment

### KRaft Mode (Recommended)

The system uses Kafka in KRaft mode, eliminating the need for Zookeeper:

```yaml
kafka:
  image: confluentinc/cp-kafka:latest
  environment:
    KAFKA_NODE_ID: 1
    KAFKA_PROCESS_ROLES: 'broker,controller'
    KAFKA_CONTROLLER_QUORUM_VOTERS: '1@kafka:29093'
    KAFKA_LISTENERS: 'PLAINTEXT://kafka:29092,CONTROLLER://kafka:29093'
    KAFKA_ADVERTISED_LISTENERS: 'PLAINTEXT://kafka:29092'
    KAFKA_AUTO_CREATE_TOPICS_ENABLE: 'true'
```

### Topic Management

Topics are auto-created with these default settings:
- **Replication Factor**: 1 (adjust for production)
- **Partitions**: 1 (can be increased for scaling)
- **Retention**: 7 days (configurable)

## Integration Examples

### Log Aggregation with ELK Stack

```yaml
# logstash.conf
input {
  kafka {
    bootstrap_servers => "kafka:9092"
    topics => ["home-automation-logs"]
    codec => "json"
  }
}

filter {
  if [service] == "DeviceService" {
    mutate {
      add_tag => ["home-automation", "device-service"]
    }
  }
}

output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "home-automation-logs-%{+YYYY.MM.dd}"
  }
}
```

### Grafana Dashboard Queries

```sql
-- Device operation counts
SELECT 
  device_id,
  action,
  COUNT(*) as operation_count
FROM logs 
WHERE timestamp > now() - interval '1 hour'
  AND level = 'INFO'
  AND action IS NOT NULL
GROUP BY device_id, action
ORDER BY operation_count DESC
```

### Real-time Alerting

```python
# Example Python consumer for alerting
from kafka import KafkaConsumer
import json

consumer = KafkaConsumer(
    'home-automation-logs',
    bootstrap_servers=['kafka:9092'],
    value_deserializer=lambda x: json.loads(x.decode('utf-8'))
)

for message in consumer:
    log_entry = message.value
    
    # Alert on errors
    if log_entry['level'] == 'ERROR':
        send_alert(f"Error in {log_entry['service']}: {log_entry['message']}")
    
    # Monitor temperature anomalies
    if (log_entry.get('action') == 'get_temperature' and 
        log_entry.get('metadata', {}).get('temperature', 0) > 30):
        send_alert(f"High temperature detected: {log_entry['metadata']['temperature']}°C")
```

## Monitoring and Analytics

### Key Metrics to Track

1. **Device Reliability**
   - Command success/failure rates
   - Response times
   - Device availability

2. **System Performance**
   - Log message throughput
   - Kafka consumer lag
   - Error frequencies

3. **Usage Patterns**
   - Most active devices
   - Peak usage times
   - Command frequency by type

### Sample Queries

```sql
-- Error rate by service
SELECT 
  service,
  SUM(CASE WHEN level = 'ERROR' THEN 1 ELSE 0 END) as errors,
  COUNT(*) as total,
  (SUM(CASE WHEN level = 'ERROR' THEN 1 ELSE 0 END) * 100.0 / COUNT(*)) as error_rate
FROM logs 
WHERE timestamp > now() - interval '1 day'
GROUP BY service;

-- Device activity heatmap
SELECT 
  DATE_TRUNC('hour', timestamp) as hour,
  device_id,
  COUNT(*) as activity_count
FROM logs 
WHERE timestamp > now() - interval '7 days'
  AND device_id IS NOT NULL
GROUP BY hour, device_id
ORDER BY hour, activity_count DESC;
```

## Best Practices

### Performance Optimization

1. **Batch Processing**: Configure appropriate batch sizes for Kafka producers
2. **Async Logging**: Use non-blocking Kafka publishing to avoid impacting device operations
3. **Compression**: Enable Kafka message compression for better throughput
4. **Partitioning**: Use device_id for message partitioning to maintain ordering

### Security Considerations

1. **Authentication**: Enable SASL authentication for Kafka in production
2. **Encryption**: Use SSL/TLS for Kafka communication
3. **Access Control**: Implement topic-level access controls
4. **Data Retention**: Configure appropriate log retention policies

### Troubleshooting

#### Common Issues

1. **Kafka Connection Failures**
   - Check broker connectivity
   - Verify network configuration
   - Review authentication settings

2. **Missing Log Messages**
   - Monitor producer acknowledgments
   - Check for serialization errors
   - Verify topic existence

3. **Performance Issues**
   - Monitor consumer lag
   - Check batch size configuration
   - Review partition count

#### Debug Commands

```bash
# Check Kafka topics
docker exec kafka kafka-topics --bootstrap-server localhost:9092 --list

# Monitor log topic
docker exec kafka kafka-console-consumer --bootstrap-server localhost:9092 --topic home-automation-logs --from-beginning

# Check consumer groups
docker exec kafka kafka-consumer-groups --bootstrap-server localhost:9092 --list
```

## Development

### Local Testing

For development, you can use a simple Kafka consumer to monitor logs:

```bash
# Start the stack
docker-compose up -d

# Monitor logs in real-time
docker exec kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic home-automation-logs \
  --property print.timestamp=true \
  --property print.key=true
```

### Custom Log Handlers

To add custom logging for new services:

```go
// Example custom logger
func (s *YourService) logWithKafka(level, message string, metadata map[string]interface{}) {
    if s.kafkaClient != nil {
        err := s.kafkaClient.PublishLog(level, "YourService", message, "", "", metadata)
        if err != nil {
            s.logger.Printf("Failed to publish log to Kafka: %v", err)
        }
    }
}
```

## Conclusion

The Kafka logging system provides comprehensive observability into the home automation system, enabling real-time monitoring, historical analysis, and proactive alerting. The structured JSON format and dual logging approach ensure both immediate debugging capabilities and long-term analytics potential.
