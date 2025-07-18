#!/bin/bash

# deploy-tls.sh - Deploy TLS-Enabled Home Automation System

set -e

echo "🔒 Deploying TLS-Enabled Home Automation System"
echo "==============================================="

# Get the project root directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DEPLOYMENTS_DIR="$PROJECT_ROOT/deployments"
CERT_DIR="$PROJECT_ROOT/certs"

# Check prerequisites
echo "🔍 Checking prerequisites..."

# Check for required commands
for cmd in openssl docker docker-compose; do
    if ! command -v "$cmd" >/dev/null 2>&1; then
        echo "❌ Required command not found: $cmd"
        echo "   Please install $cmd and try again"
        exit 1
    fi
done

echo "✅ All required commands available"

# Check if certificates exist
if [ ! -f "$CERT_DIR/ca.crt" ] || [ ! -f "$CERT_DIR/server.crt" ] || [ ! -f "$CERT_DIR/server-key.pem" ]; then
    echo "❌ TLS certificates not found in $CERT_DIR"
    echo ""
    echo "🔐 Generating TLS certificates first..."
    "$SCRIPT_DIR/generate-certificates.sh"
else
    echo "✅ TLS certificates found"
fi

echo ""
echo "📁 Step 1: Setting up TLS configuration directories..."

# Create configuration directories
mkdir -p "$DEPLOYMENTS_DIR/nginx"
mkdir -p "$DEPLOYMENTS_DIR/mosquitto"
mkdir -p "$DEPLOYMENTS_DIR/postgres"

echo "   ✅ Configuration directories created"

echo ""
echo "🔧 Step 2: Deploying Nginx TLS proxy configuration..."

# Create Nginx configuration
cat > "$DEPLOYMENTS_DIR/nginx/nginx.conf" << 'EOF'
user nginx;
worker_processes auto;
error_log /var/log/nginx/error.log notice;
pid /var/run/nginx.pid;

events {
    worker_connections 1024;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    # Logging
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for"';
    access_log /var/log/nginx/access.log main;

    # Basic Settings
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;
    client_max_body_size 10M;

    # SSL Settings
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 1d;

    # Security Headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Content-Type-Options nosniff always;
    add_header X-Frame-Options DENY always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Home Automation API (HTTPS)
    server {
        listen 8443 ssl http2;
        server_name homeautomation.local localhost _;

        ssl_certificate /etc/nginx/certs/server.crt;
        ssl_certificate_key /etc/nginx/certs/server-key.pem;

        location / {
            proxy_pass http://home-automation:8080;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_connect_timeout 30;
            proxy_send_timeout 30;
            proxy_read_timeout 30;
        }

        location /health {
            proxy_pass http://home-automation:8080/health;
            access_log off;
        }
    }

    # Grafana Dashboard (HTTPS)
    server {
        listen 3443 ssl http2;
        server_name grafana.homeautomation.local localhost _;

        ssl_certificate /etc/nginx/certs/server.crt;
        ssl_certificate_key /etc/nginx/certs/server-key.pem;

        location / {
            proxy_pass http://grafana:3000;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            
            # WebSocket support for Grafana
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
        }
    }

    # Prometheus (HTTPS)
    server {
        listen 9443 ssl http2;
        server_name prometheus.homeautomation.local localhost _;

        ssl_certificate /etc/nginx/certs/server.crt;
        ssl_certificate_key /etc/nginx/certs/server-key.pem;

        location / {
            proxy_pass http://prometheus:9090;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }

    # Tapo Metrics (HTTPS)
    server {
        listen 2443 ssl http2;
        server_name metrics.homeautomation.local localhost _;

        ssl_certificate /etc/nginx/certs/server.crt;
        ssl_certificate_key /etc/nginx/certs/server-key.pem;

        location / {
            proxy_pass http://tapo-metrics:2112;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }

    # Redirect HTTP to HTTPS
    server {
        listen 80;
        server_name _;
        return 301 https://$host:8443$request_uri;
    }
}
EOF

echo "   ✅ Nginx TLS configuration created"

echo ""
echo "📨 Step 3: Deploying MQTT TLS configuration..."

# Create MQTT TLS configuration
cat > "$DEPLOYMENTS_DIR/mosquitto/mosquitto.tls.conf" << 'EOF'
# Basic Configuration
pid_file /mosquitto/data/mosquitto.pid
persistence true
persistence_location /mosquitto/data/
autosave_interval 1800

# TLS/SSL Configuration
# =====================

# Standard MQTT over TLS (Port 8883)
listener 8883 0.0.0.0
protocol mqtt
socket_domain ipv4

# TLS Settings
cafile /mosquitto/certs/ca.crt
certfile /mosquitto/certs/server.crt
keyfile /mosquitto/certs/server-key.pem

# TLS Security Options
tls_version tlsv1.2
ciphers ECDHE+AESGCM:ECDHE+CHACHA20:DHE+AESGCM:DHE+CHACHA20:!aNULL:!MD5:!DSS
require_certificate false
use_identity_as_username false

# WebSocket over TLS (Port 9443)
listener 9443 0.0.0.0
protocol websockets
socket_domain ipv4

cafile /mosquitto/certs/ca.crt
certfile /mosquitto/certs/server.crt
keyfile /mosquitto/certs/server-key.pem
tls_version tlsv1.2

# Security Configuration
allow_anonymous false
password_file /mosquitto/config/passwd

# Connection Limits
max_connections 100
max_inflight_messages 20
max_queued_messages 100
message_size_limit 1024

# Performance
keepalive_interval 60
retry_interval 30

# Logging
log_dest file /mosquitto/log/mosquitto.log
log_dest stdout
log_type error
log_type warning
log_type notice
log_type information
log_timestamp true
EOF

echo "   ✅ MQTT TLS configuration created"

echo ""
echo "🗄️  Step 4: Deploying PostgreSQL TLS configuration..."

# Create PostgreSQL TLS setup script
cat > "$DEPLOYMENTS_DIR/postgres/setup-tls.sql" << 'EOF'
-- Enable SSL in PostgreSQL
ALTER SYSTEM SET ssl = 'on';
ALTER SYSTEM SET ssl_cert_file = '/var/lib/postgresql/certs/server.crt';
ALTER SYSTEM SET ssl_key_file = '/var/lib/postgresql/certs/server-key.pem';
ALTER SYSTEM SET ssl_ca_file = '/var/lib/postgresql/certs/ca.crt';

-- Configure SSL preferences
ALTER SYSTEM SET ssl_ciphers = 'ECDHE+AESGCM:ECDHE+CHACHA20:DHE+AESGCM:DHE+CHACHA20:!aNULL:!MD5:!DSS';
ALTER SYSTEM SET ssl_prefer_server_ciphers = 'on';

-- Reload configuration
SELECT pg_reload_conf();
EOF

echo "   ✅ PostgreSQL TLS configuration created"

echo ""
echo "🐳 Step 5: Creating TLS-enabled Docker Compose configuration..."

# Create TLS-enabled docker-compose file
cat > "$DEPLOYMENTS_DIR/docker-compose.tls.yml" << 'EOF'
services:
  # Nginx TLS Termination Proxy
  nginx:
    image: nginx:alpine
    container_name: home-automation-nginx
    ports:
      - "80:80"       # HTTP redirect
      - "8443:8443"   # Home Automation API HTTPS
      - "3443:3443"   # Grafana HTTPS
      - "9443:9443"   # Prometheus HTTPS  
      - "2443:2443"   # Tapo Metrics HTTPS
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ../certs:/etc/nginx/certs:ro
      - nginx_logs:/var/log/nginx
    depends_on:
      - home-automation
      - grafana
      - prometheus
      - tapo-metrics
    restart: unless-stopped
    networks:
      - app_network
    deploy:
      resources:
        limits:
          memory: 128M
          cpus: '0.5'
    healthcheck:
      test: ["CMD", "nginx", "-t"]
      interval: 30s
      timeout: 10s
      retries: 3

  home-automation:
    build:
      context: ..
      dockerfile: Dockerfile
    container_name: home-automation-app
    # Remove direct port exposure - accessed via nginx
    environment:
      - PORT=8080
      - DATABASE_URL=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=require
      - MQTT_BROKER=mosquitto
      - MQTT_PORT=8883                    # Use TLS port
      - MQTT_USERNAME=${MQTT_USERNAME}
      - MQTT_PASSWORD=${MQTT_PASSWORD}
      - MQTT_USE_TLS=true                 # Enable TLS
      - MQTT_CA_FILE=/app/certs/ca.crt
      - KAFKA_BROKERS=kafka:9092
      - KAFKA_LOG_TOPIC=home-automation-logs
      - LOG_FILE_PATH=/app/logs/home-automation.log
      - PROMETHEUS_URL=http://prometheus:9090
      - PROMETHEUS_PUSHGATEWAY_URL=http://prometheus:9091
      - JWT_SECRET=${JWT_SECRET}
      - TLS_CERT_FILE=/app/certs/server.crt
      - TLS_KEY_FILE=/app/certs/server-key.pem
    depends_on:
      - postgres
      - mosquitto
      - kafka
      - prometheus
    volumes:
      - ../configs:/app/configs:ro
      - ../logs:/app/logs
      - ../certs:/app/certs:ro
    restart: unless-stopped
    networks:
      - app_network
      - database_network
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '1.0'

  postgres:
    image: postgres:15-alpine
    container_name: home-automation-postgres
    environment:
      - POSTGRES_DB=${POSTGRES_DB}
      - POSTGRES_USER=${POSTGRES_USER}
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      - POSTGRES_SHARED_PRELOAD_LIBRARIES=pg_stat_statements
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ../certs:/var/lib/postgresql/certs:ro
      - ./postgres/setup-tls.sql:/docker-entrypoint-initdb.d/setup-tls.sql
    # No external port exposure - internal network only
    restart: unless-stopped
    networks:
      - database_network
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '1.0'
    command: >
      postgres
      -c ssl=on
      -c ssl_cert_file=/var/lib/postgresql/certs/server.crt
      -c ssl_key_file=/var/lib/postgresql/certs/server-key.pem
      -c ssl_ca_file=/var/lib/postgresql/certs/ca.crt

  mosquitto:
    image: eclipse-mosquitto:2.0
    container_name: home-automation-mosquitto
    ports:
      - "8883:8883"   # MQTTS (TLS)
      - "9001:9443"   # WebSocket over TLS (mapped to external 9001 for compatibility)
    user: 1883:1883
    volumes:
      - ./mosquitto/mosquitto.tls.conf:/mosquitto/config/mosquitto.conf
      - ./mosquitto/passwd:/mosquitto/config/passwd
      - ../certs:/mosquitto/certs:ro
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
    healthcheck:
      test: ["CMD", "timeout", "5", "mosquitto_pub", "-h", "localhost", "-p", "8883", "--cafile", "/mosquitto/certs/ca.crt", "-t", "health", "-m", "test", "-u", "${MQTT_USERNAME}", "-P", "${MQTT_PASSWORD}"]
      interval: 30s
      timeout: 10s
      retries: 3

  redis:
    image: redis:7-alpine
    container_name: home-automation-redis
    # No external port exposure - internal network only
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
    container_name: home-automation-kafka
    user: "1000:1000"
    # No external port exposure - internal network only
    environment:
      KAFKA_NODE_ID: 1
      CLUSTER_ID: home-automation-cluster
      KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: 'CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT'
      KAFKA_ADVERTISED_LISTENERS: 'PLAINTEXT://kafka:29092'
      KAFKA_LISTENERS: 'PLAINTEXT://kafka:29092,CONTROLLER://kafka:29093'
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

  grafana:
    image: grafana/grafana:latest
    container_name: home-automation-grafana
    # Remove direct port exposure - accessed via nginx
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD}
      - GF_SECURITY_ADMIN_USER=${GRAFANA_ADMIN_USER}
      - GF_USERS_ALLOW_SIGN_UP=false
      - GF_AUTH_ANONYMOUS_ENABLED=false
      - GF_SECURITY_SECRET_KEY=${SECRET_KEY}
      - GF_SERVER_PROTOCOL=http            # Internal HTTP, TLS handled by nginx
      - GF_SERVER_HTTP_PORT=3000
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

  prometheus:
    image: prom/prometheus:v2.47.0
    container_name: home-automation-prometheus
    # Remove direct port exposure - accessed via nginx
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--web.enable-lifecycle'
      - '--storage.tsdb.retention.time=30d'
      - '--web.listen-address=0.0.0.0:9090'
    volumes:
      - prometheus_data:/prometheus
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - ./prometheus/rules:/etc/prometheus/rules:ro
    restart: unless-stopped
    networks:
      - app_network

  tapo-metrics:
    build:
      context: ..
      dockerfile: Dockerfile.tapo
    container_name: home-automation-tapo-metrics
    # Remove direct port exposure - accessed via nginx
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

volumes:
  postgres_data:
  mosquitto_data:
  mosquitto_logs:
  kafka_data:
  grafana_data:
  prometheus_data:
  redis_data:
  nginx_logs:

networks:
  app_network:
    driver: bridge
  database_network:
    driver: bridge
    internal: true
EOF

echo "   ✅ TLS Docker Compose configuration created"

echo ""
echo "🔐 Step 6: Creating/updating secure environment configuration..."

# Create/update secure environment file
if [ ! -f "$PROJECT_ROOT/.env.secure" ]; then
    echo "   Creating new .env.secure file..."
    
    # Generate secure passwords
    POSTGRES_PASS=$(openssl rand -base64 32 | tr -d /=+ | cut -c -25)
    MQTT_PASS=$(openssl rand -base64 32 | tr -d /=+ | cut -c -25)
    GRAFANA_PASS=$(openssl rand -base64 32 | tr -d /=+ | cut -c -25)
    JWT_SECRET=$(openssl rand -base64 64 | tr -d /=+ | cut -c -50)
    SECRET_KEY=$(openssl rand -base64 64 | tr -d /=+ | cut -c -50)

    cat > "$PROJECT_ROOT/.env.secure" << EOF
# TLS-Enabled Home Automation System - Secure Configuration
# ==========================================================
# Generated: $(date)
# WARNING: Keep this file secure and do not commit to version control

# TLS Configuration
TLS_ENABLED=true
MQTT_USE_TLS=true
MQTT_PORT=8883
MQTT_CA_FILE=/app/certs/ca.crt
TLS_CERT_FILE=/app/certs/server.crt
TLS_KEY_FILE=/app/certs/server-key.pem

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

# MQTT Configuration (TLS)
MQTT_BROKER=mosquitto
MQTT_CLIENT_ID=home-automation-server-secure

# Kafka Configuration  
KAFKA_BROKERS=kafka:29092
KAFKA_LOG_TOPIC=home-automation-logs

# Logging Configuration
LOG_LEVEL=info
LOG_FILE_PATH=/app/logs/home-automation.log

# Raspberry Pi Configuration
PI_IP_ADDRESS=192.168.68.100
ENABLE_PI_MONITORING=true

# Timezone
TZ=UTC

# Tapo Device Configuration
TPLINK_USERNAME=${TPLINK_USERNAME:-admin}
TPLINK_PASSWORD=${TPLINK_PASSWORD:-admin}
TAPO_DEVICE_1_IP=${TAPO_DEVICE_1_IP:-192.168.68.101}
TAPO_DEVICE_2_IP=${TAPO_DEVICE_2_IP:-192.168.68.102}
TAPO_DEVICE_1_USE_KLAP=${TAPO_DEVICE_1_USE_KLAP:-true}
TAPO_DEVICE_2_USE_KLAP=${TAPO_DEVICE_2_USE_KLAP:-true}
EOF

    chmod 600 "$PROJECT_ROOT/.env.secure"
    echo "   ✅ Secure environment file created with generated passwords"
else
    # Update existing file with TLS settings
    echo "   Updating existing .env.secure with TLS settings..."
    
    # Add TLS settings if not present
    if ! grep -q "TLS_ENABLED" "$PROJECT_ROOT/.env.secure"; then
        cat >> "$PROJECT_ROOT/.env.secure" << 'EOF'

# TLS Configuration (Added by deploy-tls.sh)
TLS_ENABLED=true
MQTT_USE_TLS=true
MQTT_PORT=8883
MQTT_CA_FILE=/app/certs/ca.crt
TLS_CERT_FILE=/app/certs/server.crt
TLS_KEY_FILE=/app/certs/server-key.pem
EOF
        echo "   ✅ TLS settings added to existing .env.secure"
    else
        echo "   ✅ TLS settings already present in .env.secure"
    fi
fi

echo ""
echo "🛑 Step 7: Stopping existing services..."

cd "$DEPLOYMENTS_DIR"

# Stop existing services
if docker-compose ps 2>/dev/null | grep -q "Up"; then
    echo "   🔄 Stopping running containers..."
    docker-compose down
    echo "   ✅ Services stopped"
else
    echo "   ℹ️  No running services found"
fi

echo ""
echo "🚀 Step 8: Starting TLS-enabled services..."

# Start TLS-enabled services
echo "   🐳 Starting containers with TLS configuration..."
docker-compose -f docker-compose.tls.yml --env-file "$PROJECT_ROOT/.env.secure" up -d

echo "   ⏳ Waiting for services to start..."
sleep 10

# Check service status
echo "   📊 Checking service status..."
docker-compose -f docker-compose.tls.yml ps

echo ""
echo "✅ TLS Deployment Complete!"
echo "=========================="
echo ""
echo "🌐 TLS-Enabled Access URLs:"
echo "   📱 Home Automation API:  https://localhost:8443"
echo "   📊 Grafana Dashboard:    https://localhost:3443"
echo "   📈 Prometheus:           https://localhost:9443"
echo "   ⚡ Tapo Metrics:         https://localhost:2443"
echo ""
echo "🔒 MQTT TLS Access:"
echo "   📨 MQTTS:               mqtts://localhost:8883"
echo "   🌐 WebSocket TLS:        wss://localhost:9001"
echo ""
echo "⚠️  Important Notes:"
echo "   • Self-signed certificates are used - browsers will show security warnings"
echo "   • Use -k flag with curl for testing: curl -k https://localhost:8443/health"
echo "   • Add $CERT_DIR/ca.crt to your browser's trusted certificates"
echo "   • MQTT credentials are in .env.secure file"
echo ""
echo "🔍 Next Steps:"
echo "   1. Run verification: ./verify-tls.sh"
echo "   2. Check logs: docker-compose -f docker-compose.tls.yml logs"
echo "   3. Monitor security: ./check-security.sh"
echo ""
echo "🔐 Your Home Automation System is now secured with TLS encryption!"

cd - > /dev/null
