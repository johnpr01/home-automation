#!/bin/bash

# Health Check Script for Raspberry Pi 5 Home Automation

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if running in deployments directory
if [ ! -f "docker-compose.yml" ]; then
    echo "Please run this script from the deployments directory"
    exit 1
fi

echo -e "${BLUE}🏠 Home Automation Health Check${NC}"
echo "=================================="
echo ""

# System Information
echo -e "${BLUE}📊 System Information${NC}"
echo "CPU Temperature: $(vcgencmd measure_temp 2>/dev/null || echo 'N/A')"
echo "CPU Usage: $(top -bn1 | grep "Cpu(s)" | awk '{print $2+$4"%"}' || echo 'N/A')"
echo "Memory Usage: $(free -h | awk 'NR==2{printf "%.1f%%", $3*100/$2}')"
echo "Disk Usage: $(df -h / | awk 'NR==2{print $5}')"
echo "Load Average: $(uptime | awk -F'load average:' '{ print $2 }')"
echo ""

# Docker System Info
echo -e "${BLUE}🐳 Docker System${NC}"
docker system df
echo ""

# Service Status
echo -e "${BLUE}🔧 Service Status${NC}"
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

# Resource Usage
echo -e "${BLUE}💾 Container Resource Usage${NC}"
docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}"
echo ""

# Port Connectivity
echo -e "${BLUE}🌐 Network Connectivity${NC}"
ports=("8080:Home Automation API" "3000:Grafana" "1883:MQTT" "5432:PostgreSQL" "6379:Redis" "9092:Kafka")

for port_info in "${ports[@]}"; do
    port=$(echo "$port_info" | cut -d':' -f1)
    name=$(echo "$port_info" | cut -d':' -f2)
    
    if netstat -tuln | grep -q ":$port "; then
        echo -e "✅ Port $port ($name): ${GREEN}Open${NC}"
    else
        echo -e "❌ Port $port ($name): ${RED}Closed${NC}"
    fi
done
echo ""

# Service Health Checks
echo -e "${BLUE}🩺 Service Health Checks${NC}"

# Home Automation API
if curl -s http://localhost:8080/health >/dev/null 2>&1; then
    echo -e "✅ Home Automation API: ${GREEN}Healthy${NC}"
else
    echo -e "❌ Home Automation API: ${RED}Unhealthy${NC}"
fi

# Grafana
if curl -s http://localhost:3000/api/health >/dev/null 2>&1; then
    echo -e "✅ Grafana: ${GREEN}Healthy${NC}"
else
    echo -e "❌ Grafana: ${RED}Unhealthy${NC}"
fi

# PostgreSQL
if docker compose exec -T postgres pg_isready -U admin >/dev/null 2>&1; then
    echo -e "✅ PostgreSQL: ${GREEN}Ready${NC}"
else
    echo -e "❌ PostgreSQL: ${RED}Not Ready${NC}"
fi

# Redis
if docker compose exec -T redis redis-cli ping | grep -q PONG; then
    echo -e "✅ Redis: ${GREEN}Responding${NC}"
else
    echo -e "❌ Redis: ${RED}Not Responding${NC}"
fi

# Mosquitto MQTT
if timeout 5 mosquitto_pub -h localhost -p 1883 -t test -m "health-check" >/dev/null 2>&1; then
    echo -e "✅ MQTT Broker: ${GREEN}Accessible${NC}"
else
    echo -e "❌ MQTT Broker: ${RED}Not Accessible${NC}"
fi
echo ""

# Volume Usage
echo -e "${BLUE}💽 Volume Usage${NC}"
docker system df -v | grep -A 20 "Local Volumes:"
echo ""

# Recent Errors
echo -e "${BLUE}🚨 Recent Errors (Last 10 minutes)${NC}"
since_time=$(date -d '10 minutes ago' '+%Y-%m-%dT%H:%M:%S')
error_count=0

for service in "${services[@]}"; do
    errors=$(docker compose logs --since "$since_time" "$service" 2>/dev/null | grep -i -E "(error|exception|fail)" | wc -l)
    if [ "$errors" -gt 0 ]; then
        echo -e "⚠️  $service: ${YELLOW}$errors errors${NC}"
        error_count=$((error_count + errors))
    fi
done

if [ "$error_count" -eq 0 ]; then
    echo -e "✅ ${GREEN}No recent errors found${NC}"
fi
echo ""

# Performance Metrics
echo -e "${BLUE}⚡ Performance Metrics${NC}"

# Database connections
db_connections=$(docker compose exec -T postgres psql -U admin -d home_automation -t -c "SELECT count(*) FROM pg_stat_activity;" 2>/dev/null | tr -d ' ' || echo "N/A")
echo "Database Connections: $db_connections"

# MQTT clients (if mosquitto_sub is available)
if command -v mosquitto_sub >/dev/null 2>&1; then
    # This is a rough estimate - actual implementation would need mosquitto with $SYS topics enabled
    echo "MQTT Clients: Use 'mosquitto_sub -h localhost -t \$SYS/broker/clients/connected' for real-time count"
else
    echo "MQTT Clients: Install mosquitto-clients for monitoring"
fi

# Kafka topics
kafka_topics=$(docker compose exec -T kafka kafka-topics --bootstrap-server localhost:9092 --list 2>/dev/null | wc -l || echo "N/A")
echo "Kafka Topics: $kafka_topics"
echo ""

# Recommendations
echo -e "${BLUE}💡 Recommendations${NC}"

# Check CPU temperature
cpu_temp=$(vcgencmd measure_temp 2>/dev/null | grep -o '[0-9.]*' || echo "0")
if (( $(echo "$cpu_temp > 70" | bc -l 2>/dev/null || echo "0") )); then
    echo -e "🌡️  ${YELLOW}CPU temperature is high ($cpu_temp°C). Consider improving cooling.${NC}"
fi

# Check memory usage
mem_usage=$(free | awk 'NR==2{printf "%.1f", $3*100/$2}')
if (( $(echo "$mem_usage > 80" | bc -l 2>/dev/null || echo "0") )); then
    echo -e "💾 ${YELLOW}Memory usage is high ($mem_usage%). Consider optimizing container limits.${NC}"
fi

# Check disk usage
disk_usage=$(df / | awk 'NR==2{print $5}' | sed 's/%//')
if [ "$disk_usage" -gt 80 ]; then
    echo -e "💽 ${YELLOW}Disk usage is high ($disk_usage%). Consider cleaning up old logs and images.${NC}"
fi

# Check for unhealthy containers
unhealthy=$(docker compose ps | grep -v "Up" | wc -l)
if [ "$unhealthy" -gt 1 ]; then  # Header line counts as 1
    echo -e "🚨 ${YELLOW}Some containers are unhealthy. Check logs with 'docker compose logs'.${NC}"
fi

echo ""
echo -e "${GREEN}Health check complete!${NC}"
echo "Run 'docker compose logs -f' to monitor real-time logs"
echo "Run 'docker stats' to monitor real-time resource usage"
