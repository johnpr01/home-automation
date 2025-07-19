#!/bin/bash
# Raspberry Pi 5 Deployment Script with InfluxDB
# This script deploys the home automation system with all components including InfluxDB

set -e

echo "🏠 Home Automation System - Raspberry Pi 5 Deployment"
echo "===================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running on Raspberry Pi
if ! grep -q "Raspberry Pi" /proc/cpuinfo 2>/dev/null; then
    print_warning "This script is optimized for Raspberry Pi 5"
fi

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed. Please install Docker first."
    exit 1
fi

# Check if Docker Compose is available
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    print_error "Docker Compose is not available. Please install Docker Compose."
    exit 1
fi

# Create necessary directories
print_status "Creating necessary directories..."
mkdir -p ./logs
mkdir -p ./configs
mkdir -p ./deployments/mosquitto/data
mkdir -p ./deployments/mosquitto/logs
mkdir -p ./deployments/influxdb/data
mkdir -p ./deployments/grafana/data

# Set proper permissions for Raspberry Pi
print_status "Setting permissions for Raspberry Pi..."
sudo chown -R 1883:1883 ./deployments/mosquitto/data ./deployments/mosquitto/logs || true
sudo chown -R 472:472 ./deployments/grafana/data || true
sudo chown -R 1000:1000 ./deployments/influxdb/data || true

# Copy configuration files if they don't exist
print_status "Setting up configuration files..."

if [ ! -f ./deployments/mosquitto/mosquitto.conf ]; then
    print_warning "Mosquitto config not found. Please configure ./deployments/mosquitto/mosquitto.conf"
fi

if [ ! -f ./deployments/influxdb/influxdb.conf ]; then
    print_warning "InfluxDB config not found. Using default configuration."
fi

if [ ! -f ./configs/config.yaml ]; then
    print_warning "Main config not found. Please configure ./configs/config.yaml"
fi

# Change to deployments directory
cd deployments

# Pull latest images
print_status "Pulling latest Docker images..."
docker-compose pull

# Stop existing services
print_status "Stopping existing services..."
docker-compose down

# Start services with proper dependency order
print_status "Starting core infrastructure services..."

# Start InfluxDB first (needed by other services)
print_status "Starting InfluxDB..."
docker-compose up -d influxdb

# Wait for InfluxDB to be ready
print_status "Waiting for InfluxDB to be ready..."
timeout=60
counter=0
while ! curl -f http://localhost:8086/ping &>/dev/null; do
    if [ $counter -eq $timeout ]; then
        print_error "InfluxDB failed to start within $timeout seconds"
        exit 1
    fi
    sleep 2
    counter=$((counter + 2))
    echo -n "."
done
echo ""
print_status "✅ InfluxDB is ready"

# Start other infrastructure services
print_status "Starting MQTT broker..."
docker-compose up -d mosquitto

print_status "Starting Kafka..."
docker-compose up -d kafka

print_status "Starting PostgreSQL..."
docker-compose up -d postgres

print_status "Starting Redis..."
docker-compose up -d redis

# Wait for services to be ready
print_status "Waiting for services to be ready..."
sleep 10

# Start Grafana
print_status "Starting Grafana..."
docker-compose up -d grafana

# Wait for Grafana to be ready
print_status "Waiting for Grafana to be ready..."
timeout=60
counter=0
while ! curl -f http://localhost:3000/api/health &>/dev/null; do
    if [ $counter -eq $timeout ]; then
        print_warning "Grafana may not be ready yet, continuing anyway..."
        break
    fi
    sleep 2
    counter=$((counter + 2))
    echo -n "."
done
echo ""

# Start main application
print_status "Starting home automation application..."
docker-compose up -d home-automation

# Final status check
print_status "Checking service status..."
sleep 5

# Check InfluxDB
if curl -f http://localhost:8086/ping &>/dev/null; then
    print_status "✅ InfluxDB is running"
else
    print_error "❌ InfluxDB is not responding"
fi

# Check Grafana
if curl -f http://localhost:3000/api/health &>/dev/null; then
    print_status "✅ Grafana is running"
else
    print_warning "⚠️  Grafana may still be starting"
fi

# Check MQTT
if nc -z localhost 1883 &>/dev/null; then
    print_status "✅ MQTT broker is running"
else
    print_error "❌ MQTT broker is not responding"
fi

# Check main application
if curl -f http://localhost:8080/health &>/dev/null; then
    print_status "✅ Home automation API is running"
else
    print_warning "⚠️  Home automation API may still be starting"
fi

echo ""
echo "🎉 Deployment completed!"
echo ""
echo "📊 Service URLs:"
echo "   • Home Automation API: http://$(hostname -I | awk '{print $1}'):8080"
echo "   • Grafana Dashboard:   http://$(hostname -I | awk '{print $1}'):3000"
echo "   • InfluxDB:           http://$(hostname -I | awk '{print $1}'):8086"
echo "   • MQTT Broker:        $(hostname -I | awk '{print $1}'):1883"
echo ""
echo "🔑 Default Credentials:"
echo "   • Grafana: admin/homeauto2024"
echo "   • InfluxDB: admin/password123"
echo "   • Database: admin/password"
echo ""
echo "📚 Next Steps:"
echo "   1. Configure your Tapo devices in ./configs/tapo.yml"
echo "   2. Set up Pi Pico sensors (see firmware/pico-sht30/README.md)"
echo "   3. Access Grafana to view energy monitoring dashboards"
echo "   4. Check logs: docker-compose logs -f"
echo ""
echo "🔧 InfluxDB Setup:"
echo "   • Organization: home-automation"
echo "   • Bucket: sensor-data"
echo "   • Token: home-automation-token"
echo "   • Retention: 30 days"
echo ""
echo "💡 Tapo Energy Monitoring:"
echo "   • Configure devices: cp configs/tapo_template.yml configs/tapo.yml"
echo "   • Set TPLINK_PASSWORD environment variable"
echo "   • Run: cd cmd/tapo-demo && TPLINK_PASSWORD=your_password go run main.go"
echo ""
print_status "For troubleshooting, run: docker-compose logs"
