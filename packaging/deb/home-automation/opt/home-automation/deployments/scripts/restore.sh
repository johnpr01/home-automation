#!/bin/bash

# Restore Script for Raspberry Pi 5 Home Automation

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BACKUP_DIR="/home/$USER/home-automation-backups"

echo -e "${BLUE}🏠 Home Automation Restore${NC}"
echo "=========================="

# Check if backup file is provided
if [ $# -eq 0 ]; then
    echo -e "${RED}Usage: $0 <backup-file.tar.gz>${NC}"
    echo ""
    echo "Available backups:"
    ls -lh "$BACKUP_DIR"/*.tar.gz 2>/dev/null || echo "No backups found in $BACKUP_DIR"
    exit 1
fi

BACKUP_FILE="$1"

# Check if backup file exists
if [ ! -f "$BACKUP_DIR/$BACKUP_FILE" ] && [ ! -f "$BACKUP_FILE" ]; then
    echo -e "${RED}Backup file not found: $BACKUP_FILE${NC}"
    echo "Checked locations:"
    echo "  - $BACKUP_DIR/$BACKUP_FILE"
    echo "  - $BACKUP_FILE"
    exit 1
fi

# Use full path if relative path provided
if [ -f "$BACKUP_DIR/$BACKUP_FILE" ]; then
    BACKUP_FILE="$BACKUP_DIR/$BACKUP_FILE"
fi

echo "Restoring from: $BACKUP_FILE"
echo "Restore time: $(date)"
echo ""

# Check if running in deployments directory
if [ ! -f "docker-compose.yml" ]; then
    echo -e "${RED}Please run this script from the deployments directory${NC}"
    exit 1
fi

# Confirm restoration
read -p "This will overwrite current data. Are you sure? (yes/no): " -r
if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
    echo "Restore cancelled"
    exit 1
fi

echo -e "${YELLOW}🛑 Stopping services...${NC}"
docker compose down

# Extract backup
TEMP_RESTORE="/tmp/restore_$(date +%s)"
mkdir -p "$TEMP_RESTORE"

echo -e "${YELLOW}📦 Extracting backup...${NC}"
cd "$TEMP_RESTORE"
tar xzf "$BACKUP_FILE"

# Find the backup directory
BACKUP_NAME=$(ls -1 | head -n 1)
if [ -z "$BACKUP_NAME" ]; then
    echo -e "${RED}Invalid backup file format${NC}"
    rm -rf "$TEMP_RESTORE"
    exit 1
fi

cd "$BACKUP_NAME"

# Show backup information
if [ -f "backup_info.txt" ]; then
    echo -e "${BLUE}📋 Backup Information${NC}"
    cat backup_info.txt
    echo ""
fi

# Restore configuration files
echo -e "${YELLOW}⚙️ Restoring configuration...${NC}"
cd "$TEMP_RESTORE/$BACKUP_NAME"

if [ -d "config" ]; then
    # Backup current config
    if [ -f "../../../.env" ]; then
        cp "../../../.env" "../../../.env.backup.$(date +%s)"
    fi
    
    # Restore configurations
    cp config/.env ../../../ 2>/dev/null || echo "No .env file in backup"
    cp config/docker-compose.yml ../../../deployments/ 2>/dev/null || echo "No docker-compose.yml in backup"
    cp -r config/mosquitto ../../../deployments/ 2>/dev/null || echo "No mosquitto config in backup"
    cp -r config/grafana ../../../deployments/ 2>/dev/null || echo "No grafana config in backup"
    cp -r config/configs ../../../ 2>/dev/null || echo "No configs directory in backup"
fi

# Restore logs
echo -e "${YELLOW}📝 Restoring logs...${NC}"
if [ -d "logs" ]; then
    mkdir -p ../../../logs
    cp -r logs/* ../../../logs/ 2>/dev/null || echo "No logs to restore"
fi

# Return to deployments directory
cd /home/philip/home-automation/deployments

# Start services to create volumes
echo -e "${YELLOW}🚀 Starting services to create volumes...${NC}"
docker compose up -d
sleep 10

# Stop services for volume restoration
echo -e "${YELLOW}🛑 Stopping services for volume restoration...${NC}"
docker compose down

# Restore Docker volumes
echo -e "${YELLOW}💾 Restoring Docker volumes...${NC}"
cd "$TEMP_RESTORE/$BACKUP_NAME"

if [ -d "volumes" ]; then
    # Restore PostgreSQL data
    if [ -f "volumes/postgres_data.tar.gz" ]; then
        echo "Restoring PostgreSQL data..."
        docker run --rm -v "$(pwd)/volumes:/backup" -v "deployments_postgres_data:/data" busybox tar xzf /backup/postgres_data.tar.gz -C /data
    fi
    
    # Restore Grafana data
    if [ -f "volumes/grafana_data.tar.gz" ]; then
        echo "Restoring Grafana data..."
        docker run --rm -v "$(pwd)/volumes:/backup" -v "deployments_grafana_data:/data" busybox tar xzf /backup/grafana_data.tar.gz -C /data
    fi
    
    # Restore Mosquitto data
    if [ -f "volumes/mosquitto_data.tar.gz" ]; then
        echo "Restoring Mosquitto data..."
        docker run --rm -v "$(pwd)/volumes:/backup" -v "deployments_mosquitto_data:/data" busybox tar xzf /backup/mosquitto_data.tar.gz -C /data
    fi
    
    # Restore Redis data
    if [ -f "volumes/redis_data.tar.gz" ]; then
        echo "Restoring Redis data..."
        docker run --rm -v "$(pwd)/volumes:/backup" -v "deployments_redis_data:/data" busybox tar xzf /backup/redis_data.tar.gz -C /data
    fi
    
    # Restore Kafka data
    if [ -f "volumes/kafka_data.tar.gz" ]; then
        echo "Restoring Kafka data..."
        docker run --rm -v "$(pwd)/volumes:/backup" -v "deployments_kafka_data:/data" busybox tar xzf /backup/kafka_data.tar.gz -C /data
    fi
fi

# Restore database from SQL dump
echo -e "${YELLOW}🗄️ Restoring database...${NC}"
if [ -f "database/home_automation.sql" ]; then
    # Start only PostgreSQL for restoration
    docker compose up -d postgres
    sleep 10
    
    # Wait for PostgreSQL to be ready
    until docker compose exec postgres pg_isready -U admin; do
        echo "Waiting for PostgreSQL to be ready..."
        sleep 2
    done
    
    # Restore database
    docker compose exec -T postgres psql -U admin -d home_automation < database/home_automation.sql
    echo "Database restored successfully"
    
    # Stop PostgreSQL
    docker compose down
fi

# Return to deployments directory and start all services
cd /home/philip/home-automation/deployments

echo -e "${YELLOW}🚀 Starting all services...${NC}"
docker compose up -d

# Wait for services to start
echo "Waiting for services to start..."
sleep 30

# Cleanup temporary files
rm -rf "$TEMP_RESTORE"

# Verify restoration
echo -e "${BLUE}🔍 Verifying restoration...${NC}"
echo "Service status:"
docker compose ps

# Check if services are responding
echo ""
echo "Service health:"
services=("postgres" "mosquitto" "redis" "kafka" "grafana" "home-automation")

for service in "${services[@]}"; do
    status=$(docker compose ps "$service" --format "table {{.State}}" | tail -n +2)
    if [[ "$status" == "running" ]]; then
        echo -e "✅ $service: ${GREEN}Running${NC}"
    else
        echo -e "❌ $service: ${RED}$status${NC}"
    fi
done

echo ""
echo -e "${GREEN}✅ Restore completed successfully!${NC}"
echo ""
echo "Access your restored services:"
local_ip=$(hostname -I | awk '{print $1}')
echo "🏠 Home Automation API:  http://${local_ip}:8080"
echo "📊 Grafana Dashboard:    http://${local_ip}:3000"
echo "📡 MQTT Broker:          ${local_ip}:1883"
echo ""
echo "Check logs with: docker compose logs -f"
echo "Monitor status with: ./scripts/health-check.sh"
