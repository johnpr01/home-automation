#!/bin/bash

# check-security.sh - Security Status Checker for Home Automation System

set -e

echo "🔒 Home Automation System - Security Status Check"
echo "================================================="
echo "$(date)"
echo ""

# Get the project root directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Initialize counters
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNING_CHECKS=0

# Function to check a security item
check_security_item() {
    local name="$1"
    local status="$2"
    local message="$3"
    
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    
    case "$status" in
        "pass")
            echo "✅ $name: $message"
            PASSED_CHECKS=$((PASSED_CHECKS + 1))
            ;;
        "fail")
            echo "❌ $name: $message"
            FAILED_CHECKS=$((FAILED_CHECKS + 1))
            ;;
        "warn")
            echo "⚠️  $name: $message"
            WARNING_CHECKS=$((WARNING_CHECKS + 1))
            ;;
    esac
}

echo "🔐 Authentication & Credentials"
echo "==============================="

# Check for secure environment file
if [ -f "$PROJECT_ROOT/.env.secure" ]; then
    check_security_item "Secure Environment" "pass" "Configuration file exists"
    
    # Check file permissions
    PERM=$(stat -c "%a" "$PROJECT_ROOT/.env.secure")
    if [ "$PERM" = "600" ]; then
        check_security_item "Environment Permissions" "pass" "Restrictive permissions (600)"
    else
        check_security_item "Environment Permissions" "fail" "Insecure permissions ($PERM) - should be 600"
    fi
    
    # Check for strong passwords
    if grep -q "POSTGRES_PASSWORD=" "$PROJECT_ROOT/.env.secure"; then
        POSTGRES_PASS_LEN=$(grep "POSTGRES_PASSWORD=" "$PROJECT_ROOT/.env.secure" | cut -d= -f2 | wc -c)
        if [ "$POSTGRES_PASS_LEN" -gt 15 ]; then
            check_security_item "Database Password" "pass" "Strong password length ($POSTGRES_PASS_LEN chars)"
        else
            check_security_item "Database Password" "fail" "Weak password length ($POSTGRES_PASS_LEN chars)"
        fi
    else
        check_security_item "Database Password" "fail" "Not configured"
    fi
    
    # Check for JWT secret
    if grep -q "JWT_SECRET=" "$PROJECT_ROOT/.env.secure"; then
        JWT_LEN=$(grep "JWT_SECRET=" "$PROJECT_ROOT/.env.secure" | cut -d= -f2 | wc -c)
        if [ "$JWT_LEN" -gt 30 ]; then
            check_security_item "JWT Secret" "pass" "Strong secret length ($JWT_LEN chars)"
        else
            check_security_item "JWT Secret" "fail" "Weak secret length ($JWT_LEN chars)"
        fi
    else
        check_security_item "JWT Secret" "fail" "Not configured"
    fi
    
else
    check_security_item "Secure Environment" "fail" "Configuration file missing"
    check_security_item "Environment Permissions" "fail" "File does not exist"
    check_security_item "Database Password" "fail" "No secure config"
    check_security_item "JWT Secret" "fail" "No secure config"
fi

# Check for default credentials
if [ -f "$PROJECT_ROOT/.env" ]; then
    if grep -q "password.*admin\|password.*password\|password.*123" "$PROJECT_ROOT/.env" 2>/dev/null; then
        check_security_item "Default Credentials" "fail" "Default/weak credentials detected in .env"
    else
        check_security_item "Default Credentials" "pass" "No obvious default credentials in .env"
    fi
fi

echo ""
echo "🌐 Network Security"
echo "==================="

# Check Docker Compose configuration
if [ -f "$PROJECT_ROOT/deployments/docker-compose.yml" ]; then
    # Check for exposed database port
    if grep -q "5432:5432" "$PROJECT_ROOT/deployments/docker-compose.yml"; then
        check_security_item "Database Exposure" "fail" "PostgreSQL port exposed to host"
    else
        check_security_item "Database Exposure" "pass" "PostgreSQL not exposed to host"
    fi
    
    # Check for resource limits
    if grep -q "limits:" "$PROJECT_ROOT/deployments/docker-compose.yml"; then
        check_security_item "Resource Limits" "pass" "Container resource limits configured"
    else
        check_security_item "Resource Limits" "warn" "No container resource limits"
    fi
    
    # Check for non-root users
    if grep -q "user:" "$PROJECT_ROOT/deployments/docker-compose.yml"; then
        check_security_item "Non-root Users" "pass" "Non-root user configuration found"
    else
        check_security_item "Non-root Users" "warn" "No explicit non-root user configuration"
    fi
else
    check_security_item "Docker Configuration" "fail" "docker-compose.yml not found"
fi

# Check secure Docker Compose
if [ -f "$PROJECT_ROOT/deployments/docker-compose.secure.yml" ]; then
    check_security_item "Secure Docker Config" "pass" "Secure Docker Compose configuration exists"
else
    check_security_item "Secure Docker Config" "fail" "No secure Docker Compose configuration"
fi

echo ""
echo "🔒 TLS/Encryption Status"
echo "========================"

# Check for TLS certificates
if [ -f "$PROJECT_ROOT/certs/ca.crt" ] && [ -f "$PROJECT_ROOT/certs/server.crt" ]; then
    check_security_item "TLS Certificates" "pass" "TLS certificates present"
    
    # Check certificate validity
    EXPIRY=$(openssl x509 -in "$PROJECT_ROOT/certs/server.crt" -noout -enddate | cut -d= -f2)
    EXPIRY_EPOCH=$(date -d "$EXPIRY" +%s 2>/dev/null || echo "0")
    CURRENT_EPOCH=$(date +%s)
    DAYS_UNTIL_EXPIRY=$(( (EXPIRY_EPOCH - CURRENT_EPOCH) / 86400 ))
    
    if [ "$DAYS_UNTIL_EXPIRY" -gt 30 ]; then
        check_security_item "Certificate Validity" "pass" "Valid ($DAYS_UNTIL_EXPIRY days remaining)"
    elif [ "$DAYS_UNTIL_EXPIRY" -gt 0 ]; then
        check_security_item "Certificate Validity" "warn" "Expires soon ($DAYS_UNTIL_EXPIRY days)"
    else
        check_security_item "Certificate Validity" "fail" "Certificate expired"
    fi
else
    check_security_item "TLS Certificates" "fail" "TLS certificates missing"
    check_security_item "Certificate Validity" "fail" "No certificates to check"
fi

# Check TLS configuration
if [ -f "$PROJECT_ROOT/deployments/docker-compose.tls.yml" ]; then
    check_security_item "TLS Configuration" "pass" "TLS Docker Compose configuration present"
else
    check_security_item "TLS Configuration" "fail" "TLS configuration missing"
fi

# Check for TLS environment variables
if [ -f "$PROJECT_ROOT/.env.secure" ] && grep -q "TLS_ENABLED=true" "$PROJECT_ROOT/.env.secure"; then
    check_security_item "TLS Environment" "pass" "TLS enabled in environment"
else
    check_security_item "TLS Environment" "fail" "TLS not enabled in environment"
fi

echo ""
echo "🌐 Service Security Tests"
echo "========================="

# Test HTTPS endpoints (if running)
if timeout 5 curl -k --connect-timeout 3 https://localhost:8443/health >/dev/null 2>&1; then
    check_security_item "HTTPS API" "pass" "Home Automation API HTTPS accessible"
else
    check_security_item "HTTPS API" "warn" "HTTPS API not accessible (may not be running)"
fi

# Test HTTP redirect
if timeout 5 curl -s -o /dev/null -w "%{http_code}" http://localhost:80 2>/dev/null | grep -q "301\|302"; then
    check_security_item "HTTP Redirect" "pass" "HTTP redirects to HTTPS"
else
    check_security_item "HTTP Redirect" "warn" "HTTP redirect not working (service may not be running)"
fi

# Test MQTTS
if command -v mosquitto_pub >/dev/null 2>&1 && [ -f "$PROJECT_ROOT/certs/ca.crt" ]; then
    if timeout 5 mosquitto_pub -h localhost -p 8883 --cafile "$PROJECT_ROOT/certs/ca.crt" -t test -m test >/dev/null 2>&1; then
        check_security_item "MQTTS" "pass" "MQTTS accessible"
    else
        check_security_item "MQTTS" "warn" "MQTTS not accessible (may need credentials or not running)"
    fi
else
    check_security_item "MQTTS" "warn" "MQTTS test skipped (mosquitto_pub not available or no certificates)"
fi

# Test plain MQTT (should fail if secure)
if timeout 5 nc -z localhost 1883 >/dev/null 2>&1; then
    check_security_item "Plain MQTT" "fail" "Insecure MQTT port still accessible"
else
    check_security_item "Plain MQTT" "pass" "Plain MQTT properly disabled"
fi

echo ""
echo "📁 File Security"
echo "================"

# Check for exposed secrets in version control
if [ -d "$PROJECT_ROOT/.git" ]; then
    if git -C "$PROJECT_ROOT" check-ignore .env.secure >/dev/null 2>&1; then
        check_security_item "Secret Files VCS" "pass" "Secure files properly ignored by git"
    else
        check_security_item "Secret Files VCS" "warn" "Secure files may not be ignored by git"
    fi
else
    check_security_item "Secret Files VCS" "warn" "Not a git repository"
fi

# Check for world-readable secrets
WORLD_READABLE=$(find "$PROJECT_ROOT" -name "*.env*" -o -name "*password*" -o -name "*key*" | xargs ls -la 2>/dev/null | grep "r--r--r--" | wc -l)
if [ "$WORLD_READABLE" -eq 0 ]; then
    check_security_item "File Permissions" "pass" "No world-readable secret files found"
else
    check_security_item "File Permissions" "fail" "$WORLD_READABLE world-readable secret files found"
fi

echo ""
echo "🐳 Container Security"
echo "====================="

# Check running containers
if command -v docker >/dev/null 2>&1; then
    # Check for privileged containers
    PRIVILEGED=$(docker ps --format "table {{.Names}}" | tail -n +2 | xargs -I {} docker inspect {} --format '{{.Name}}: {{.HostConfig.Privileged}}' 2>/dev/null | grep -c "true" || echo "0")
    if [ "$PRIVILEGED" -eq 0 ]; then
        check_security_item "Privileged Containers" "pass" "No privileged containers running"
    else
        check_security_item "Privileged Containers" "fail" "$PRIVILEGED privileged containers found"
    fi
    
    # Check for root processes
    ROOT_CONTAINERS=$(docker ps --format "table {{.Names}}" | tail -n +2 | xargs -I {} docker exec {} ps -o user= 2>/dev/null | grep -c "root" || echo "0")
    if [ "$ROOT_CONTAINERS" -eq 0 ]; then
        check_security_item "Root Processes" "pass" "No root processes in containers"
    else
        check_security_item "Root Processes" "warn" "$ROOT_CONTAINERS containers running as root"
    fi
else
    check_security_item "Container Security" "warn" "Docker not available for security check"
fi

echo ""
echo "📊 Security Score Summary"
echo "========================="

# Calculate security score
TOTAL_POSSIBLE=$((PASSED_CHECKS + FAILED_CHECKS + WARNING_CHECKS))
SECURITY_SCORE=$(( (PASSED_CHECKS * 100) / TOTAL_POSSIBLE ))
WARNING_PENALTY=$(( WARNING_CHECKS * 2 ))
ADJUSTED_SCORE=$(( SECURITY_SCORE - WARNING_PENALTY ))

# Ensure score doesn't go below 0
if [ "$ADJUSTED_SCORE" -lt 0 ]; then
    ADJUSTED_SCORE=0
fi

echo "📈 Security Statistics:"
echo "   ✅ Passed: $PASSED_CHECKS"
echo "   ❌ Failed: $FAILED_CHECKS"
echo "   ⚠️  Warnings: $WARNING_CHECKS"
echo "   📊 Total Checks: $TOTAL_POSSIBLE"
echo ""
echo "🏆 Security Score: $ADJUSTED_SCORE/100"

# Determine security level
if [ "$ADJUSTED_SCORE" -ge 90 ]; then
    echo "🛡️  Security Level: EXCELLENT 🎉"
    echo "   Your system has strong security configurations!"
    SECURITY_LEVEL="excellent"
elif [ "$ADJUSTED_SCORE" -ge 75 ]; then
    echo "🔒 Security Level: GOOD ✅"
    echo "   Your system has decent security with room for improvement."
    SECURITY_LEVEL="good"
elif [ "$ADJUSTED_SCORE" -ge 50 ]; then
    echo "⚠️  Security Level: MODERATE 🟡"
    echo "   Your system needs security improvements."
    SECURITY_LEVEL="moderate"
else
    echo "❌ Security Level: POOR 🔴"
    echo "   Your system requires immediate security attention!"
    SECURITY_LEVEL="poor"
fi

echo ""
echo "🔧 Recommended Actions"
echo "======================"

if [ "$FAILED_CHECKS" -gt 0 ]; then
    echo "🚨 IMMEDIATE ACTIONS NEEDED:"
    
    if [ ! -f "$PROJECT_ROOT/.env.secure" ]; then
        echo "   1. Run emergency security fix: ./scripts/emergency-security-fix.sh"
    fi
    
    if [ ! -f "$PROJECT_ROOT/certs/ca.crt" ]; then
        echo "   2. Generate TLS certificates: ./scripts/generate-certificates.sh"
        echo "   3. Deploy TLS configuration: ./scripts/deploy-tls.sh"
    fi
    
    if grep -q "5432:5432" "$PROJECT_ROOT/deployments/docker-compose.yml" 2>/dev/null; then
        echo "   4. Remove database port exposure from docker-compose.yml"
    fi
fi

if [ "$WARNING_CHECKS" -gt 0 ]; then
    echo ""
    echo "⚠️  RECOMMENDED IMPROVEMENTS:"
    echo "   • Review and implement all security recommendations"
    echo "   • Add container resource limits"
    echo "   • Configure non-root users for all services"
    echo "   • Set up proper log monitoring"
    echo "   • Implement security scanning in CI/CD"
fi

echo ""
echo "📚 Security Resources:"
echo "   📄 Security Implementation Guide: chats/SECURITY_IMPLEMENTATION_GUIDE.md"
echo "   📄 TLS Implementation Guide: chats/TLS_IMPLEMENTATION_GUIDE.md"
echo "   📄 Security Audit Report: chats/SECURITY_AUDIT_REPORT.md"

echo ""
echo "🔄 Re-run this check after implementing fixes:"
echo "   ./scripts/check-security.sh"

# Exit with appropriate code
if [ "$SECURITY_LEVEL" = "poor" ] || [ "$FAILED_CHECKS" -gt 5 ]; then
    exit 1
elif [ "$SECURITY_LEVEL" = "moderate" ] || [ "$FAILED_CHECKS" -gt 2 ]; then
    exit 2
else
    exit 0
fi
