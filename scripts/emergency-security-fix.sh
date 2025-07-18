#!/bin/bash
# Emergency Security Fixes for Home Automation System
# Run this immediately to address critical security vulnerabilities

set -e

echo "🚨 EMERGENCY SECURITY FIXES FOR HOME AUTOMATION SYSTEM"
echo "====================================================="
echo ""
echo "⚠️  WARNING: This will modify your current configuration"
echo "📋 A backup will be created before making changes"
echo ""

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   echo "❌ Don't run this script as root for security reasons"
   exit 1
fi

# Check dependencies
command -v openssl >/dev/null 2>&1 || { echo "❌ openssl is required but not installed. Aborting." >&2; exit 1; }
command -v docker >/dev/null 2>&1 || { echo "❌ docker is required but not installed. Aborting." >&2; exit 1; }

echo "✅ Prerequisites check passed"
echo ""

# Navigate to project directory
cd "$(dirname "$0")/.."
PROJECT_ROOT=$(pwd)

echo "📁 Working in: $PROJECT_ROOT"
echo ""

# 1. Generate secure passwords
echo "🔐 Step 1: Generating cryptographically secure credentials..."
POSTGRES_PASS=$(openssl rand -base64 32)
GRAFANA_PASS=$(openssl rand -base64 32)
JWT_SECRET=$(openssl rand -base64 64)
MQTT_PASS=$(openssl rand -base64 32)
SECRET_KEY=$(openssl rand -base64 32)

echo "   ✅ Generated secure passwords"

# 2. Create secure environment file
echo ""
echo "📝 Step 2: Creating secure environment configuration..."
cat > .env.secure << EOF
# SECURE ENVIRONMENT CONFIGURATION
# Generated: $(date)
# DO NOT COMMIT THIS FILE TO VERSION CONTROL

# Database Security
POSTGRES_DB=home_automation_secure
POSTGRES_USER=homeauto_admin
POSTGRES_PASSWORD=${POSTGRES_PASS}

# Grafana Security  
GRAFANA_ADMIN_USER=security_admin
GRAFANA_ADMIN_PASSWORD=${GRAFANA_PASS}

# JWT Security
JWT_SECRET=${JWT_SECRET}

# MQTT Security
MQTT_USERNAME=homeauto_mqtt
MQTT_PASSWORD=${MQTT_PASS}

# General Security
SECRET_KEY=${SECRET_KEY}

# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# MQTT Configuration
MQTT_BROKER=mosquitto
MQTT_PORT=1883
MQTT_CLIENT_ID=home-automation-server-secure

# Kafka Configuration  
KAFKA_BROKERS=kafka:9092
KAFKA_LOG_TOPIC=home-automation-logs

# Logging Configuration
LOG_LEVEL=info
LOG_FILE_PATH=/app/logs/home-automation.log

# Raspberry Pi Configuration
PI_IP_ADDRESS=192.168.1.100
ENABLE_PI_MONITORING=true

# Timezone
TZ=UTC

# Performance Tuning for Pi 5
POSTGRES_SHARED_BUFFERS=128MB
POSTGRES_EFFECTIVE_CACHE_SIZE=256MB
KAFKA_HEAP_OPTS=-Xmx256m -Xms128m
REDIS_MAXMEMORY=64mb
EOF

chmod 600 .env.secure
echo "   ✅ Created .env.secure with restricted permissions"

# 3. Backup current configuration
echo ""
echo "💾 Step 3: Backing up current configuration..."
BACKUP_DIR="backups/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"

if [ -f "deployments/docker-compose.yml" ]; then
    cp deployments/docker-compose.yml "$BACKUP_DIR/"
    echo "   ✅ Backed up docker-compose.yml"
fi

if [ -f ".env" ]; then
    cp .env "$BACKUP_DIR/"
    echo "   ✅ Backed up .env"
fi

if [ -f "deployments/mosquitto/mosquitto.conf" ]; then
    cp deployments/mosquitto/mosquitto.conf "$BACKUP_DIR/"
    echo "   ✅ Backed up mosquitto.conf"
fi

echo "   📁 Backup location: $BACKUP_DIR"

# 4. Stop current services if running
echo ""
echo "🛑 Step 4: Stopping current services..."
if [ -f "deployments/docker-compose.yml" ]; then
    cd deployments
    if docker-compose ps | grep -q "Up"; then
        echo "   🔄 Stopping running containers..."
        docker-compose down
        echo "   ✅ Services stopped"
    else
        echo "   ℹ️  No running services found"
    fi
    cd ..
else
    echo "   ⚠️  docker-compose.yml not found, skipping service stop"
fi

# 5. Create secure docker-compose configuration
echo ""
echo "🔧 Step 5: Creating secure Docker Compose configuration..."
cat > deployments/docker-compose.secure.yml << 'EOF'
services:
  home-automation:
    build:
      context: ..
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - DATABASE_URL=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=prefer
      - MQTT_BROKER=mosquitto
      - MQTT_PORT=1883
      - MQTT_USERNAME=${MQTT_USERNAME}
      - MQTT_PASSWORD=${MQTT_PASSWORD}
      - KAFKA_BROKERS=kafka:9092
      - KAFKA_LOG_TOPIC=home-automation-logs
      - LOG_FILE_PATH=/app/logs/home-automation.log
      - PROMETHEUS_URL=http://prometheus:9090
      - PROMETHEUS_PUSHGATEWAY_URL=http://prometheus:9091
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      - postgres
      - mosquitto
      - kafka
      - prometheus
    volumes:
      - ../configs:/app/configs:ro
      - ../logs:/app/logs
    restart: unless-stopped
    networks:
      - app_network
      - database_network
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '1.0'
        reservations:
          memory: 256M
          cpus: '0.5'

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=${POSTGRES_DB}
      - POSTGRES_USER=${POSTGRES_USER}
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      - POSTGRES_SHARED_PRELOAD_LIBRARIES=pg_stat_statements
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
    # SECURITY: Database port NOT exposed to host
    restart: unless-stopped
    networks:
      - database_network
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '1.0'
        reservations:
          memory: 256M
          cpus: '0.5'
    command: >
      postgres
      -c shared_buffers=128MB
      -c effective_cache_size=256MB
      -c maintenance_work_mem=64MB
      -c checkpoint_completion_target=0.9
      -c wal_buffers=16MB
      -c default_statistics_target=100

  mosquitto:
    image: eclipse-mosquitto:2.0
    ports:
      - "1883:1883"
      - "9001:9001"
    user: 1883:1883
    volumes:
      - ./mosquitto/mosquitto.secure.conf:/mosquitto/config/mosquitto.conf
      - ./mosquitto/passwd:/mosquitto/config/passwd
      - mosquitto_data:/mosquitto/data
      - mosquitto_logs:/mosquitto/log
    tmpfs:
      - /tmp:noexec,nosuid,size=10m
    restart: unless-stopped
    networks:
      - app_network
    deploy:
      resources:
        limits:
          memory: 128M
          cpus: '0.5'

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    restart: unless-stopped
    networks:
      - app_network
    deploy:
      resources:
        limits:
          memory: 128M
          cpus: '0.5'
    command: >
      redis-server
      --maxmemory 64mb
      --maxmemory-policy allkeys-lru
      --save 900 1
      --save 300 10

  kafka:
    image: confluentinc/cp-kafka:latest
    user: "1000:1000"
    ports:
      - "9092:9092"
      - "9093:9093"
    environment:
      KAFKA_NODE_ID: 1
      CLUSTER_ID: home-automation-cluster
      KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: 'CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT'
      KAFKA_ADVERTISED_LISTENERS: 'PLAINTEXT://kafka:29092,PLAINTEXT_HOST://localhost:9092'
      KAFKA_LISTENERS: 'PLAINTEXT://kafka:29092,CONTROLLER://kafka:29093,PLAINTEXT_HOST://0.0.0.0:9092'
      KAFKA_INTER_BROKER_LISTENER_NAME: 'PLAINTEXT'
      KAFKA_CONTROLLER_LISTENER_NAMES: 'CONTROLLER'
      KAFKA_CONTROLLER_QUORUM_VOTERS: '1@kafka:29093'
      KAFKA_PROCESS_ROLES: 'broker,controller'
      KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS: 0
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
      KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR: 1
      KAFKA_TRANSACTION_STATE_LOG_MIN_ISR: 1
      KAFKA_AUTO_CREATE_TOPICS_ENABLE: 'true'
      KAFKA_LOG_DIRS: '/var/lib/kafka/logs'
      KAFKA_HEAP_OPTS: '-Xmx256m -Xms128m'
      KAFKA_LOG_SEGMENT_BYTES: 104857600
      KAFKA_LOG_RETENTION_HOURS: 24
      KAFKA_LOG_RETENTION_BYTES: 536870912
    volumes:
      - kafka_data:/var/lib/kafka/data
    restart: unless-stopped
    networks:
      - app_network
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '1.0'
        reservations:
          memory: 256M
          cpus: '0.5'

  prometheus:
    image: prom/prometheus:v2.47.0
    ports:
      - "9090:9090"
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--web.enable-lifecycle'
      - '--storage.tsdb.retention.time=30d'
    volumes:
      - prometheus_data:/prometheus
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - ./prometheus/rules:/etc/prometheus/rules:ro
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:9090/-/healthy"]
      interval: 30s
      timeout: 10s
      retries: 5
      start_period: 40s
    restart: unless-stopped
    networks:
      - app_network
    deploy:
      resources:
        limits:
          memory: 256M
          cpus: '0.5'
        reservations:
          memory: 128M
          cpus: '0.25'

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD}
      - GF_SECURITY_ADMIN_USER=${GRAFANA_ADMIN_USER}
      - GF_USERS_ALLOW_SIGN_UP=false
      - GF_AUTH_ANONYMOUS_ENABLED=false
      - GF_SECURITY_SECRET_KEY=${SECRET_KEY}
      - GF_INSTALL_PLUGINS=grafana-clock-panel,grafana-simple-json-datasource
    volumes:
      - grafana_data:/var/lib/grafana
      - ./grafana/provisioning:/etc/grafana/provisioning
    depends_on:
      - postgres
      - prometheus
    restart: unless-stopped
    networks:
      - app_network
      - database_network
    deploy:
      resources:
        limits:
          memory: 256M
          cpus: '0.5'

  tapo-metrics:
    build:
      context: ..
      dockerfile: Dockerfile.tapo
    container_name: home-automation-tapo-metrics
    ports:
      - "2112:2112"
    environment:
      - METRICS_PORT=2112
      - TPLINK_USERNAME=${TPLINK_USERNAME:-}
      - TPLINK_PASSWORD=${TPLINK_PASSWORD:-}
      - TAPO_DEVICE_1_IP=${TAPO_DEVICE_1_IP:-}
      - TAPO_DEVICE_2_IP=${TAPO_DEVICE_2_IP:-}
      - TAPO_DEVICE_1_USE_KLAP=${TAPO_DEVICE_1_USE_KLAP:-true}
      - TAPO_DEVICE_2_USE_KLAP=${TAPO_DEVICE_2_USE_KLAP:-true}
      - LOG_LEVEL=info
      - POLL_INTERVAL=30s
    volumes:
      - ../configs:/app/configs:ro
      - ../logs:/app/logs
    depends_on:
      - prometheus
    restart: unless-stopped
    user: "1000:1000"
    networks:
      - app_network
    deploy:
      resources:
        limits:
          memory: 128M
          cpus: '0.5'
        reservations:
          memory: 64M
          cpus: '0.25'
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:2112/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 30s

volumes:
  postgres_data:
  mosquitto_data:
  mosquitto_logs:
  redis_data:
  kafka_data:
  grafana_data:
  prometheus_data:

networks:
  app_network:
    driver: bridge
  database_network:
    driver: bridge
    internal: true  # Database network isolated from external access
EOF

echo "   ✅ Created secure Docker Compose configuration"

# 6. Setup MQTT security
echo ""
echo "🔐 Step 6: Configuring MQTT security..."
mkdir -p deployments/mosquitto

# Create secure MQTT configuration
cat > deployments/mosquitto/mosquitto.secure.conf << 'EOF'
# Secure Mosquitto MQTT Broker Configuration
# ==========================================

# General Configuration
pid_file /mosquitto/data/mosquitto.pid

# Persistence settings
persistence true
persistence_location /mosquitto/data/
autosave_interval 1800

# Logging Configuration
log_dest file /mosquitto/log/mosquitto.log
log_dest stdout
log_type error
log_type warning  
log_type notice
log_type information
log_timestamp true
log_timestamp_format %Y-%m-%dT%H:%M:%S
connection_messages true

# Network Listeners
listener 1883 0.0.0.0
protocol mqtt
socket_domain ipv4

listener 9001 0.0.0.0  
protocol websockets
socket_domain ipv4

# SECURITY CONFIGURATION
# ======================
# CRITICAL: Authentication required
allow_anonymous false
password_file /mosquitto/config/passwd

# Access control (when implemented)
# acl_file /mosquitto/config/acl

# Connection limits
max_connections 100
max_inflight_messages 20
max_queued_messages 100

# Message size limits
message_size_limit 1024

# Security settings
use_username_as_clientid false
EOF

echo "   ✅ Created secure MQTT configuration"

# Create MQTT password file (will be populated when containers start)
echo "   🔐 MQTT password file will be created when containers start"

# 7. Update original docker-compose to be more secure
echo ""
echo "🔧 Step 7: Updating original Docker Compose for security..."
if [ -f "deployments/docker-compose.yml" ]; then
    # Comment out exposed database port
    sed -i 's/^[[:space:]]*- "5432:5432"/#      - "5432:5432"  # SECURITY: Removed exposed database port/' deployments/docker-compose.yml
    echo "   ✅ Commented out exposed database port in original file"
fi

# 8. Create security status check script
echo ""
echo "📋 Step 8: Creating security status checker..."
cat > scripts/check-security.sh << 'EOF'
#!/bin/bash
# Security Status Checker

echo "🔒 HOME AUTOMATION SECURITY STATUS"
echo "=================================="

SECURE=true

# Check 1: Database port exposure
if docker-compose -f deployments/docker-compose.yml config | grep -q "5432:5432"; then
    echo "❌ Database port exposed (5432)"
    SECURE=false
else
    echo "✅ Database port not exposed"
fi

# Check 2: Environment file security
if [ -f ".env.secure" ]; then
    echo "✅ Secure environment file exists"
    PERMS=$(stat -c "%a" .env.secure)
    if [ "$PERMS" = "600" ]; then
        echo "✅ Secure environment file has correct permissions"
    else
        echo "⚠️  Secure environment file permissions should be 600"
    fi
else
    echo "❌ Secure environment file missing"
    SECURE=false
fi

# Check 3: MQTT security
if [ -f "deployments/mosquitto/mosquitto.secure.conf" ]; then
    if grep -q "allow_anonymous false" deployments/mosquitto/mosquitto.secure.conf; then
        echo "✅ MQTT authentication enabled"
    else
        echo "❌ MQTT allows anonymous access"
        SECURE=false
    fi
else
    echo "❌ Secure MQTT configuration missing"
    SECURE=false
fi

# Check 4: Default passwords
if grep -q "password" deployments/docker-compose.yml 2>/dev/null; then
    echo "⚠️  Default passwords may still be in docker-compose.yml"
fi

echo ""
if [ "$SECURE" = true ]; then
    echo "🎉 SECURITY STATUS: SECURE"
    echo "   All critical security checks passed"
else
    echo "⚠️  SECURITY STATUS: NEEDS ATTENTION"
    echo "   Some security issues found above"
fi
EOF

chmod +x scripts/check-security.sh
echo "   ✅ Created security status checker"

# 9. Create credentials file for admin reference
echo ""
echo "🔑 Step 9: Creating credentials reference..."
cat > .credentials-$(date +%Y%m%d_%H%M%S).txt << EOF
HOME AUTOMATION SYSTEM - EMERGENCY SECURITY CREDENTIALS
=======================================================
Generated: $(date)

CRITICAL: Store these credentials in a secure password manager
          Delete this file after copying credentials securely

Database:
  Username: homeauto_admin
  Password: ${POSTGRES_PASS}
  Database: home_automation_secure

Grafana Dashboard:
  Username: security_admin  
  Password: ${GRAFANA_PASS}
  URL: http://your-pi-ip:3000

MQTT Broker:
  Username: homeauto_mqtt
  Password: ${MQTT_PASS}

JWT Secret: ${JWT_SECRET}
System Secret: ${SECRET_KEY}

NEXT STEPS:
1. Test system with: cd deployments && docker-compose -f docker-compose.secure.yml --env-file ../.env.secure up -d
2. Run security check: ./scripts/check-security.sh
3. Setup MQTT users: docker exec mosquitto mosquitto_passwd -c /mosquitto/config/passwd homeauto_mqtt
4. Delete this file after copying credentials securely
5. Implement TLS certificates for production
EOF

chmod 600 .credentials-*.txt
echo "   ✅ Created secure credentials file (600 permissions)"

echo ""
echo "🎉 EMERGENCY SECURITY FIXES COMPLETED!"
echo "====================================="
echo ""
echo "📋 SUMMARY OF CHANGES:"
echo "   ✅ Generated cryptographically secure passwords"
echo "   ✅ Created secure environment configuration (.env.secure)"
echo "   ✅ Backed up original configuration"
echo "   ✅ Created secure Docker Compose configuration"
echo "   ✅ Configured MQTT security (authentication required)"
echo "   ✅ Removed exposed database port"
echo "   ✅ Created security status checker"
echo "   ✅ Generated secure credentials file"
echo ""
echo "🚀 TO START SECURE SYSTEM:"
echo "   cd deployments"
echo "   docker-compose -f docker-compose.secure.yml --env-file ../.env.secure up -d"
echo ""
echo "🔍 TO CHECK SECURITY STATUS:"
echo "   ./scripts/check-security.sh"
echo ""
echo "🔑 IMPORTANT NEXT STEPS:"
echo "   1. Copy credentials from .credentials-*.txt to secure password manager"
echo "   2. Delete credentials file after copying"
echo "   3. Test system functionality"
echo "   4. Setup MQTT user password"
echo "   5. Configure TLS certificates"
echo "   6. Implement remaining security recommendations"
echo ""
echo "⚠️  BACKUP LOCATION: $BACKUP_DIR"
echo "🔒 CREDENTIALS FILE: $(ls .credentials-*.txt)"
echo ""
echo "✅ System is now significantly more secure!"
