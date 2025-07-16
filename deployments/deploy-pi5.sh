#!/bin/bash

# Raspberry Pi 5 Home Automation Deployment Script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging function
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
}

# Check if running on Raspberry Pi
check_raspberry_pi() {
    if [[ ! -f /proc/device-tree/model ]] || ! grep -q "Raspberry Pi" /proc/device-tree/model; then
        warn "This script is optimized for Raspberry Pi but can run on other systems"
    else
        local model=$(cat /proc/device-tree/model)
        log "Detected: $model"
    fi
}

# Check system requirements
check_requirements() {
    log "Checking system requirements..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    # Check Docker Compose
    if ! docker compose version &> /dev/null; then
        error "Docker Compose is not installed or not working."
        exit 1
    fi
    
    # Check available memory
    local total_mem=$(awk '/MemTotal/ {print $2}' /proc/meminfo)
    local total_mem_gb=$((total_mem / 1024 / 1024))
    
    if [ $total_mem_gb -lt 2 ]; then
        warn "System has less than 2GB RAM. Performance may be limited."
    else
        log "Available memory: ${total_mem_gb}GB"
    fi
    
    # Check disk space
    local available_space=$(df / | awk 'NR==2 {print $4}')
    local available_gb=$((available_space / 1024 / 1024))
    
    if [ $available_gb -lt 5 ]; then
        error "Insufficient disk space. At least 5GB required."
        exit 1
    else
        log "Available disk space: ${available_gb}GB"
    fi
}

# Setup directories and permissions
setup_directories() {
    log "Setting up directories..."
    
    # Create necessary directories
    mkdir -p ../logs
    mkdir -p mosquitto
    mkdir -p grafana/provisioning/dashboards
    mkdir -p grafana/provisioning/datasources
    mkdir -p init-scripts
    
    # Set proper permissions
    chmod 755 ../logs
    chmod 755 mosquitto
    chmod -R 755 grafana
    
    log "Directories created and permissions set"
}

# Create environment file
create_env_file() {
    if [ ! -f ../.env ]; then
        log "Creating environment configuration..."
        
        # Get Raspberry Pi IP address
        local pi_ip=$(hostname -I | awk '{print $1}')
        
        cat > ../.env << EOF
# Home Automation Environment Configuration for Raspberry Pi 5

# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Database Configuration
DATABASE_URL=postgres://admin:homeauto2024@postgres:5432/home_automation?sslmode=disable
POSTGRES_DB=home_automation
POSTGRES_USER=admin
POSTGRES_PASSWORD=homeauto2024

# MQTT Configuration
MQTT_BROKER=mosquitto
MQTT_PORT=1883
MQTT_CLIENT_ID=home-automation-server

# Kafka Configuration
KAFKA_BROKERS=kafka:9092
KAFKA_LOG_TOPIC=home-automation-logs

# Logging Configuration
LOG_LEVEL=info
LOG_FILE_PATH=/app/logs/home-automation.log

# Grafana Configuration
GF_SECURITY_ADMIN_PASSWORD=homeauto2024

# Raspberry Pi Configuration
PI_IP_ADDRESS=${pi_ip}
ENABLE_PI_MONITORING=true

# Timezone
TZ=UTC
EOF
        
        log "Environment file created at ../.env"
        log "Raspberry Pi IP detected as: $pi_ip"
        warn "Please review and update passwords in ../.env before production deployment"
    else
        log "Environment file already exists"
    fi
}

# Create Mosquitto configuration
create_mosquitto_config() {
    log "Creating Mosquitto configuration..."
    
    cat > mosquitto/mosquitto.conf << EOF
# Mosquitto MQTT Broker Configuration for Raspberry Pi 5

# Network settings
listener 1883 0.0.0.0
protocol mqtt

# WebSocket support
listener 9001 0.0.0.0
protocol websockets

# Security settings
allow_anonymous true
# Note: For production, enable authentication:
# allow_anonymous false
# password_file /mosquitto/config/passwd

# Logging
log_dest file /mosquitto/log/mosquitto.log
log_type error
log_type warning
log_type notice
log_type information
log_timestamp true

# Persistence
persistence true
persistence_location /mosquitto/data/
autosave_interval 1800

# Connection limits (optimized for Pi 5)
max_connections 100
max_inflight_messages 20
max_queued_messages 1000

# Performance tuning for Pi 5
sys_interval 10
store_clean_interval 60

# WebSocket settings
websockets_log_level 0
EOF
    
    log "Mosquitto configuration created"
}

# Create Grafana datasource configuration
create_grafana_config() {
    log "Creating Grafana configuration..."
    
    mkdir -p grafana/provisioning/datasources
    
    cat > grafana/provisioning/datasources/home-automation.yml << EOF
apiVersion: 1

datasources:
  - name: PostgreSQL
    type: postgres
    url: postgres:5432
    database: home_automation
    user: admin
    secureJsonData:
      password: homeauto2024
    jsonData:
      sslmode: disable
      maxOpenConns: 5
      maxIdleConns: 2
      connMaxLifetime: 14400
EOF
    
    mkdir -p grafana/provisioning/dashboards
    
    cat > grafana/provisioning/dashboards/home-automation.yml << EOF
apiVersion: 1

providers:
  - name: 'Home Automation'
    folder: 'Home Automation'
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /etc/grafana/provisioning/dashboards
EOF
    
    log "Grafana configuration created"
}

# Optimize Docker for Raspberry Pi
optimize_docker() {
    log "Optimizing Docker for Raspberry Pi 5..."
    
    # Create or update Docker daemon configuration
    local docker_config="/etc/docker/daemon.json"
    
    if [ ! -f "$docker_config" ]; then
        sudo tee "$docker_config" > /dev/null << EOF
{
    "log-driver": "json-file",
    "log-opts": {
        "max-size": "10m",
        "max-file": "3"
    },
    "storage-driver": "overlay2",
    "exec-opts": ["native.cgroupdriver=systemd"],
    "experimental": false,
    "features": {
        "buildkit": true
    }
}
EOF
        
        sudo systemctl restart docker
        log "Docker configuration optimized and restarted"
    else
        log "Docker configuration already exists"
    fi
}

# Build and deploy services
deploy_services() {
    log "Building and deploying services..."
    
    # Pull latest images
    log "Pulling Docker images..."
    docker compose pull
    
    # Build home automation service
    log "Building home automation service..."
    docker compose build home-automation
    
    # Start services
    log "Starting services..."
    docker compose up -d
    
    # Wait for services to be ready
    log "Waiting for services to start..."
    sleep 30
    
    # Check service health
    check_service_health
}

# Check service health
check_service_health() {
    log "Checking service health..."
    
    local services=("postgres" "mosquitto" "redis" "kafka" "grafana" "home-automation")
    local healthy=true
    
    for service in "${services[@]}"; do
        if docker compose ps "$service" | grep -q "Up"; then
            log "✓ $service is running"
        else
            error "✗ $service is not running"
            healthy=false
        fi
    done
    
    if [ "$healthy" = true ]; then
        log "All services are healthy!"
        show_service_urls
    else
        error "Some services are not healthy. Check logs with: docker compose logs"
        exit 1
    fi
}

# Show service URLs
show_service_urls() {
    local pi_ip=$(hostname -I | awk '{print $1}')
    
    log "Deployment complete! Services are available at:"
    echo ""
    echo "🏠 Home Automation API:  http://${pi_ip}:8080"
    echo "📊 Grafana Dashboard:    http://${pi_ip}:3000 (admin/homeauto2024)"
    echo "📡 MQTT Broker:          ${pi_ip}:1883"
    echo "🌐 MQTT WebSocket:       ${pi_ip}:9001"
    echo "🗄️  PostgreSQL:          ${pi_ip}:5432"
    echo "🔄 Redis:                ${pi_ip}:6379"
    echo "📨 Kafka:                ${pi_ip}:9092"
    echo ""
    echo "📝 View logs: docker compose logs -f"
    echo "🔍 Service status: docker compose ps"
    echo "📈 Resource usage: docker stats"
}

# Main deployment function
main() {
    echo "🏠 Raspberry Pi 5 Home Automation Deployment"
    echo "=============================================="
    echo ""
    
    check_raspberry_pi
    check_requirements
    setup_directories
    create_env_file
    create_mosquitto_config
    create_grafana_config
    optimize_docker
    deploy_services
    
    log "Deployment completed successfully!"
}

# Run main function
main "$@"
