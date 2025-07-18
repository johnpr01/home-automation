#!/bin/bash

# verify-tls.sh - Verify TLS Implementation for Home Automation System

set -e

echo "🔒 TLS Verification for Home Automation System"
echo "=============================================="

# Get the project root directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CERT_DIR="$PROJECT_ROOT/certs"

# Load environment variables if available
if [ -f "$PROJECT_ROOT/.env.secure" ]; then
    source "$PROJECT_ROOT/.env.secure"
elif [ -f "$PROJECT_ROOT/.env" ]; then
    source "$PROJECT_ROOT/.env"
fi

# Default values if not set in environment
MQTT_USERNAME=${MQTT_USERNAME:-"homeauto_mqtt"}
MQTT_PASSWORD=${MQTT_PASSWORD:-"secure_mqtt_password"}

echo "🔍 Checking certificate files..."
if [ ! -f "$CERT_DIR/ca.crt" ]; then
    echo "❌ CA certificate not found: $CERT_DIR/ca.crt"
    echo "   Run: ./generate-certificates.sh first"
    exit 1
fi

if [ ! -f "$CERT_DIR/server.crt" ]; then
    echo "❌ Server certificate not found: $CERT_DIR/server.crt"
    echo "   Run: ./generate-certificates.sh first"
    exit 1
fi

echo "✅ Certificate files found"

echo ""
echo "📡 Testing HTTPS endpoints..."

# Test main API HTTPS
echo -n "  🌐 Home Automation API (8443): "
if timeout 10 curl -k --connect-timeout 5 https://localhost:8443/health >/dev/null 2>&1; then
    echo "✅ HTTPS Working"
    API_HTTPS_OK=true
else
    echo "❌ HTTPS Failed (service may not be running)"
    API_HTTPS_OK=false
fi

# Test Grafana HTTPS
echo -n "  📊 Grafana Dashboard (3443): "
if timeout 10 curl -k --connect-timeout 5 https://localhost:3443/api/health >/dev/null 2>&1; then
    echo "✅ HTTPS Working"
    GRAFANA_HTTPS_OK=true
elif timeout 10 curl -k --connect-timeout 5 https://localhost:3443/ >/dev/null 2>&1; then
    echo "✅ HTTPS Working"
    GRAFANA_HTTPS_OK=true
else
    echo "❌ HTTPS Failed (service may not be running)"
    GRAFANA_HTTPS_OK=false
fi

# Test Prometheus HTTPS
echo -n "  📈 Prometheus (9443): "
if timeout 10 curl -k --connect-timeout 5 https://localhost:9443/-/healthy >/dev/null 2>&1; then
    echo "✅ HTTPS Working"
    PROMETHEUS_HTTPS_OK=true
elif timeout 10 curl -k --connect-timeout 5 https://localhost:9443/ >/dev/null 2>&1; then
    echo "✅ HTTPS Working"
    PROMETHEUS_HTTPS_OK=true
else
    echo "❌ HTTPS Failed (service may not be running)"
    PROMETHEUS_HTTPS_OK=false
fi

# Test Tapo Metrics HTTPS
echo -n "  ⚡ Tapo Metrics (2443): "
if timeout 10 curl -k --connect-timeout 5 https://localhost:2443/metrics >/dev/null 2>&1; then
    echo "✅ HTTPS Working"
    TAPO_HTTPS_OK=true
else
    echo "❌ HTTPS Failed (service may not be running)"
    TAPO_HTTPS_OK=false
fi

echo ""
echo "📨 Testing MQTTS endpoint..."

# Test MQTTS with authentication
echo -n "  🏠 MQTT over TLS (8883): "
if command -v mosquitto_pub >/dev/null 2>&1; then
    if timeout 10 mosquitto_pub -h localhost -p 8883 \
        --cafile "$CERT_DIR/ca.crt" \
        -t "test/tls" -m "hello" \
        -u "$MQTT_USERNAME" -P "$MQTT_PASSWORD" >/dev/null 2>&1; then
        echo "✅ MQTTS Working"
        MQTTS_OK=true
    else
        echo "❌ MQTTS Failed (check credentials or service)"
        MQTTS_OK=false
    fi
else
    echo "⚠️  mosquitto_pub not available, skipping MQTT test"
    MQTTS_OK=false
fi

# Test MQTT WebSocket over TLS
echo -n "  🌐 MQTT WebSocket TLS (9443): "
if timeout 5 nc -z localhost 9443 >/dev/null 2>&1; then
    echo "✅ Port Open"
    MQTT_WS_OK=true
else
    echo "❌ Port Closed"
    MQTT_WS_OK=false
fi

echo ""
echo "🔐 Testing certificate validity..."

# Check certificate expiry
echo -n "  📜 Certificate expiry: "
EXPIRY=$(openssl x509 -in "$CERT_DIR/server.crt" -noout -enddate | cut -d= -f2)
EXPIRY_EPOCH=$(date -d "$EXPIRY" +%s 2>/dev/null || echo "0")
CURRENT_EPOCH=$(date +%s)
DAYS_UNTIL_EXPIRY=$(( (EXPIRY_EPOCH - CURRENT_EPOCH) / 86400 ))

if [ "$DAYS_UNTIL_EXPIRY" -gt 30 ]; then
    echo "✅ Valid until $EXPIRY ($DAYS_UNTIL_EXPIRY days)"
    CERT_EXPIRY_OK=true
elif [ "$DAYS_UNTIL_EXPIRY" -gt 0 ]; then
    echo "⚠️  Expires soon: $EXPIRY ($DAYS_UNTIL_EXPIRY days)"
    CERT_EXPIRY_OK=true
else
    echo "❌ EXPIRED: $EXPIRY"
    CERT_EXPIRY_OK=false
fi

# Check certificate chain
echo -n "  🔗 Certificate chain: "
if openssl verify -CAfile "$CERT_DIR/ca.crt" "$CERT_DIR/server.crt" >/dev/null 2>&1; then
    echo "✅ Valid"
    CERT_CHAIN_OK=true
else
    echo "❌ Invalid"
    CERT_CHAIN_OK=false
fi

# Check certificate key pair
echo -n "  🔐 Certificate/Key pair: "
CERT_MODULUS=$(openssl x509 -noout -modulus -in "$CERT_DIR/server.crt" | openssl md5)
KEY_MODULUS=$(openssl rsa -noout -modulus -in "$CERT_DIR/server-key.pem" 2>/dev/null | openssl md5)

if [ "$CERT_MODULUS" = "$KEY_MODULUS" ]; then
    echo "✅ Matching"
    CERT_KEY_MATCH=true
else
    echo "❌ Mismatch"
    CERT_KEY_MATCH=false
fi

echo ""
echo "🔒 Testing TLS protocol versions..."

# Test TLS 1.2
echo -n "  🛡️  TLS 1.2: "
if timeout 10 openssl s_client -connect localhost:8443 -tls1_2 -verify_return_error </dev/null >/dev/null 2>&1; then
    echo "✅ Supported"
    TLS12_OK=true
else
    echo "❌ Not supported/service unavailable"
    TLS12_OK=false
fi

# Test TLS 1.3
echo -n "  🛡️  TLS 1.3: "
if timeout 10 openssl s_client -connect localhost:8443 -tls1_3 -verify_return_error </dev/null >/dev/null 2>&1; then
    echo "✅ Supported"
    TLS13_OK=true
else
    echo "⚠️  Not supported/service unavailable"
    TLS13_OK=false
fi

# Test weak protocols (should fail)
echo -n "  🚫 TLS 1.1 (should fail): "
if timeout 10 openssl s_client -connect localhost:8443 -tls1_1 -verify_return_error </dev/null >/dev/null 2>&1; then
    echo "❌ Insecure protocol supported!"
    WEAK_TLS_BLOCKED=false
else
    echo "✅ Properly blocked"
    WEAK_TLS_BLOCKED=true
fi

echo ""
echo "🔍 Testing cipher suites..."

# Get supported cipher suites
echo "  🔐 Supported cipher suites:"
if timeout 10 openssl s_client -connect localhost:8443 -cipher 'ALL' </dev/null 2>/dev/null | grep -E "Cipher|Protocol" | head -2; then
    CIPHER_INFO_OK=true
else
    echo "     ⚠️  Unable to retrieve cipher information (service may not be running)"
    CIPHER_INFO_OK=false
fi

echo ""
echo "🐳 Checking Docker TLS configuration..."

# Check if TLS-enabled docker-compose is running
echo -n "  📋 TLS Docker Compose: "
if [ -f "$PROJECT_ROOT/deployments/docker-compose.tls.yml" ]; then
    echo "✅ TLS configuration file exists"
    TLS_COMPOSE_EXISTS=true
else
    echo "❌ TLS configuration file missing"
    TLS_COMPOSE_EXISTS=false
fi

# Check if nginx proxy is running
echo -n "  🌐 Nginx TLS Proxy: "
if docker ps --format "table {{.Names}}" | grep -q "nginx" 2>/dev/null; then
    echo "✅ Running"
    NGINX_RUNNING=true
elif docker ps -a --format "table {{.Names}}" | grep -q "nginx" 2>/dev/null; then
    echo "⚠️  Container exists but not running"
    NGINX_RUNNING=false
else
    echo "❌ Not deployed"
    NGINX_RUNNING=false
fi

echo ""
echo "📊 TLS Security Summary"
echo "======================"

TOTAL_CHECKS=0
PASSED_CHECKS=0

# Count and display results
declare -A checks=(
    ["API HTTPS"]="$API_HTTPS_OK"
    ["Grafana HTTPS"]="$GRAFANA_HTTPS_OK"
    ["Prometheus HTTPS"]="$PROMETHEUS_HTTPS_OK"
    ["Tapo Metrics HTTPS"]="$TAPO_HTTPS_OK"
    ["MQTTS"]="$MQTTS_OK"
    ["MQTT WebSocket TLS"]="$MQTT_WS_OK"
    ["Certificate Expiry"]="$CERT_EXPIRY_OK"
    ["Certificate Chain"]="$CERT_CHAIN_OK"
    ["Certificate/Key Match"]="$CERT_KEY_MATCH"
    ["TLS 1.2 Support"]="$TLS12_OK"
    ["Weak TLS Blocked"]="$WEAK_TLS_BLOCKED"
    ["TLS Config Exists"]="$TLS_COMPOSE_EXISTS"
)

for check in "${!checks[@]}"; do
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    if [ "${checks[$check]}" = "true" ]; then
        echo "✅ $check"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
    else
        echo "❌ $check"
    fi
done

echo ""
echo "📈 Security Score: $PASSED_CHECKS/$TOTAL_CHECKS checks passed"

# Calculate percentage
PERCENTAGE=$(( (PASSED_CHECKS * 100) / TOTAL_CHECKS ))

if [ $PERCENTAGE -ge 90 ]; then
    echo "🎉 Excellent TLS security! ($PERCENTAGE%)"
    EXIT_CODE=0
elif [ $PERCENTAGE -ge 75 ]; then
    echo "✅ Good TLS security ($PERCENTAGE%)"
    EXIT_CODE=0
elif [ $PERCENTAGE -ge 50 ]; then
    echo "⚠️  Moderate TLS security ($PERCENTAGE%) - improvements needed"
    EXIT_CODE=1
else
    echo "❌ Poor TLS security ($PERCENTAGE%) - immediate attention required"
    EXIT_CODE=1
fi

echo ""
echo "🔧 Troubleshooting:"
echo "==================="

if [ "$API_HTTPS_OK" = "false" ] || [ "$GRAFANA_HTTPS_OK" = "false" ] || [ "$PROMETHEUS_HTTPS_OK" = "false" ]; then
    echo "🌐 HTTPS Issues:"
    echo "   • Check if services are running: docker-compose ps"
    echo "   • Verify TLS configuration is deployed: ./deploy-tls.sh"
    echo "   • Check nginx logs: docker logs <nginx-container>"
fi

if [ "$MQTTS_OK" = "false" ]; then
    echo "📨 MQTTS Issues:"
    echo "   • Verify MQTT credentials in .env.secure"
    echo "   • Check mosquitto logs: docker logs <mosquitto-container>"
    echo "   • Ensure mosquitto.tls.conf is deployed"
fi

if [ "$CERT_EXPIRY_OK" = "false" ] || [ "$CERT_CHAIN_OK" = "false" ]; then
    echo "🔐 Certificate Issues:"
    echo "   • Regenerate certificates: ./generate-certificates.sh"
    echo "   • Redeploy TLS configuration: ./deploy-tls.sh"
fi

if [ "$TLS_COMPOSE_EXISTS" = "false" ]; then
    echo "📋 Deployment Issues:"
    echo "   • Deploy TLS configuration: ./deploy-tls.sh"
    echo "   • Check TLS implementation guide: chats/TLS_IMPLEMENTATION_GUIDE.md"
fi

echo ""
echo "📚 For detailed TLS setup instructions, see:"
echo "   📄 chats/TLS_IMPLEMENTATION_GUIDE.md"
echo "   🔒 chats/SECURITY_IMPLEMENTATION_GUIDE.md"

exit $EXIT_CODE
