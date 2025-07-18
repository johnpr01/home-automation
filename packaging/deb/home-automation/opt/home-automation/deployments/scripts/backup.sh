#!/bin/bash

# Backup Script for Raspberry Pi 5 Home Automation

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BACKUP_DIR="/home/$USER/home-automation-backups"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_NAME="home-automation-backup_${TIMESTAMP}"
RETENTION_DAYS=7

# Create backup directory
mkdir -p "$BACKUP_DIR"

echo -e "${BLUE}🏠 Home Automation Backup${NC}"
echo "=========================="
echo "Backup Location: $BACKUP_DIR/$BACKUP_NAME.tar.gz"
echo "Timestamp: $(date)"
echo ""

# Check if running in deployments directory
if [ ! -f "docker-compose.yml" ]; then
    echo -e "${RED}Please run this script from the deployments directory${NC}"
    exit 1
fi

# Create temporary backup directory
TEMP_BACKUP="/tmp/$BACKUP_NAME"
mkdir -p "$TEMP_BACKUP"

echo -e "${YELLOW}📦 Creating backup...${NC}"

# Backup configuration files
echo "Backing up configuration files..."
mkdir -p "$TEMP_BACKUP/config"
cp -r ../configs "$TEMP_BACKUP/config/" 2>/dev/null || echo "No configs directory found"
cp ../.env "$TEMP_BACKUP/config/" 2>/dev/null || echo "No .env file found"
cp docker-compose.yml "$TEMP_BACKUP/config/"
cp -r mosquitto "$TEMP_BACKUP/config/" 2>/dev/null || echo "No mosquitto config found"
cp -r grafana "$TEMP_BACKUP/config/" 2>/dev/null || echo "No grafana config found"

# Backup application logs
echo "Backing up logs..."
mkdir -p "$TEMP_BACKUP/logs"
cp -r ../logs/* "$TEMP_BACKUP/logs/" 2>/dev/null || echo "No logs found"

# Backup database
echo "Backing up PostgreSQL database..."
mkdir -p "$TEMP_BACKUP/database"
docker compose exec -T postgres pg_dump -U admin home_automation > "$TEMP_BACKUP/database/home_automation.sql"

# Backup Docker volumes
echo "Backing up Docker volumes..."
mkdir -p "$TEMP_BACKUP/volumes"

# Export volume data
docker run --rm -v "$(pwd)_postgres_data:/data" -v "$TEMP_BACKUP/volumes:/backup" busybox tar czf /backup/postgres_data.tar.gz -C /data .
docker run --rm -v "$(pwd)_grafana_data:/data" -v "$TEMP_BACKUP/volumes:/backup" busybox tar czf /backup/grafana_data.tar.gz -C /data .
docker run --rm -v "$(pwd)_mosquitto_data:/data" -v "$TEMP_BACKUP/volumes:/backup" busybox tar czf /backup/mosquitto_data.tar.gz -C /data . 2>/dev/null || echo "Mosquitto data backup skipped"
docker run --rm -v "$(pwd)_redis_data:/data" -v "$TEMP_BACKUP/volumes:/backup" busybox tar czf /backup/redis_data.tar.gz -C /data . 2>/dev/null || echo "Redis data backup skipped"
docker run --rm -v "$(pwd)_kafka_data:/data" -v "$TEMP_BACKUP/volumes:/backup" busybox tar czf /backup/kafka_data.tar.gz -C /data . 2>/dev/null || echo "Kafka data backup skipped"

# Create backup metadata
echo "Creating backup metadata..."
cat > "$TEMP_BACKUP/backup_info.txt" << EOF
Home Automation Backup Information
==================================
Backup Date: $(date)
Backup Version: 1.0
Raspberry Pi Model: $(cat /proc/device-tree/model 2>/dev/null || echo "Unknown")
System: $(uname -a)
Docker Version: $(docker --version)
Docker Compose Version: $(docker compose version)

Services Backed Up:
- PostgreSQL Database
- Grafana Configuration and Data
- Mosquitto MQTT Broker Data
- Redis Data (if present)
- Kafka Data (if present)
- Application Configuration
- Application Logs

Docker Images:
$(docker compose images)

Container Status at Backup Time:
$(docker compose ps)
EOF

# Create the compressed backup
echo "Compressing backup..."
cd /tmp
tar czf "$BACKUP_DIR/$BACKUP_NAME.tar.gz" "$BACKUP_NAME"

# Calculate backup size
BACKUP_SIZE=$(du -h "$BACKUP_DIR/$BACKUP_NAME.tar.gz" | cut -f1)

# Cleanup temporary files
rm -rf "$TEMP_BACKUP"

echo -e "${GREEN}✅ Backup completed successfully!${NC}"
echo "Backup file: $BACKUP_DIR/$BACKUP_NAME.tar.gz"
echo "Backup size: $BACKUP_SIZE"
echo ""

# Cleanup old backups
echo "Cleaning up old backups (older than $RETENTION_DAYS days)..."
find "$BACKUP_DIR" -name "home-automation-backup_*.tar.gz" -mtime +$RETENTION_DAYS -delete
REMAINING_BACKUPS=$(ls -1 "$BACKUP_DIR"/home-automation-backup_*.tar.gz 2>/dev/null | wc -l)
echo "Remaining backups: $REMAINING_BACKUPS"

echo ""
echo -e "${BLUE}📋 Backup Summary${NC}"
echo "=================="
ls -lh "$BACKUP_DIR"
echo ""
echo "To restore this backup, run:"
echo "  ./scripts/restore.sh $BACKUP_NAME.tar.gz"
