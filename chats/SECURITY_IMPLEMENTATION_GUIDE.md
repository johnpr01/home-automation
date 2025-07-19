# 🔒 Security Implementation Guide

## 🚨 **CRITICAL FIXES - IMPLEMENT IMMEDIATELY**

### **1. Secure Database Configuration**

Create a secure environment file:

```bash
# Create secure .env file
cat > .env.secure << 'EOF'
# Database Security
POSTGRES_PASSWORD=$(openssl rand -base64 32)
POSTGRES_USER=homeauto_admin
POSTGRES_DB=home_automation_secure

# Grafana Security  
GRAFANA_ADMIN_USER=security_admin
GRAFANA_ADMIN_PASSWORD=$(openssl rand -base64 32)

# JWT Security
JWT_SECRET=$(openssl rand -base64 64)

# MQTT Security
MQTT_USERNAME=homeauto_mqtt
MQTT_PASSWORD=$(openssl rand -base64 32)

# General Security
SECRET_KEY=$(openssl rand -base64 32)
EOF
```

### **2. Secure Docker Compose Configuration**

```yaml
# deployments/docker-compose.secure.yml
services:
  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=${POSTGRES_DB}
      - POSTGRES_USER=${POSTGRES_USER}  
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
    # REMOVE exposed port for security
    # ports:
    #   - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - database_network  # Separate network

  home-automation:
    environment:
      - DATABASE_URL=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=require
    networks:
      - database_network
      - app_network

  grafana:
    environment:
      - GF_SECURITY_ADMIN_USER=${GRAFANA_ADMIN_USER}
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD}
      - GF_USERS_ALLOW_SIGN_UP=false
      - GF_AUTH_ANONYMOUS_ENABLED=false
      - GF_SECURITY_SECRET_KEY=${SECRET_KEY}

networks:
  database_network:
    driver: bridge
    internal: true  # No external access
  app_network:
    driver: bridge
```

### **3. Secure MQTT Configuration**

```bash
# Create MQTT password file
docker exec mosquitto mosquitto_passwd -c /mosquitto/config/passwd ${MQTT_USERNAME}

# Create ACL file
cat > deployments/mosquitto/acl << 'EOF'
# Home automation user permissions
user homeauto_mqtt
topic readwrite home/+/+
topic readwrite sensor/+/+
topic readwrite automation/+/+

# Admin patterns
pattern readwrite admin/%u/+
EOF
```

```properties
# deployments/mosquitto/mosquitto.secure.conf
allow_anonymous false
password_file /mosquitto/config/passwd
acl_file /mosquitto/config/acl

# Enable TLS (after certificate setup)
listener 8883 0.0.0.0
protocol mqtt
cafile /mosquitto/certs/ca.crt
certfile /mosquitto/certs/server.crt
keyfile /mosquitto/certs/server.key
require_certificate false
use_identity_as_username false
```

### **4. Secure Dockerfile**

```dockerfile
# Dockerfile.secure
FROM golang:1.23-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates && \
    addgroup -g 1001 homeauto && \
    adduser -D -s /bin/sh -u 1001 -G homeauto homeauto

WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/web ./web

RUN chown -R homeauto:homeauto /app
USER homeauto

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./main"]
```

---

## 🛡️ **MEDIUM PRIORITY SECURITY ENHANCEMENTS**

### **5. Input Validation Middleware**

```go
// internal/middleware/validation.go
package middleware

import (
    "regexp"
    "strings"
    "net/http"
)

func ValidateInput(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Validate Content-Type
        if r.Method == "POST" || r.Method == "PUT" {
            ct := r.Header.Get("Content-Type")
            if !strings.Contains(ct, "application/json") {
                http.Error(w, "Invalid content type", http.StatusBadRequest)
                return
            }
        }
        
        // Validate path parameters
        if containsSqlInjection(r.URL.Path) {
            http.Error(w, "Invalid request", http.StatusBadRequest)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

func containsSqlInjection(input string) bool {
    sqlPatterns := []string{
        `(?i)(union|select|insert|delete|update|drop|create|alter|exec|execute)`,
        `(?i)(script|javascript|vbscript|onload|onerror)`,
        `[<>\"'%;()&+]`,
    }
    
    for _, pattern := range sqlPatterns {
        if matched, _ := regexp.MatchString(pattern, input); matched {
            return true
        }
    }
    return false
}
```

### **6. Security Headers Middleware**

```go
// internal/middleware/security.go
package middleware

import "net/http"

func SecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Security headers
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY") 
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; object-src 'none';")
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
        
        next.ServeHTTP(w, r)
    })
}
```

### **7. Rate Limiting Middleware**

```go
// internal/middleware/ratelimit.go
package middleware

import (
    "net/http"
    "sync"
    "time"
    "golang.org/x/time/rate"
)

type visitor struct {
    limiter  *rate.Limiter
    lastSeen time.Time
}

type RateLimiter struct {
    visitors map[string]*visitor
    mu       sync.Mutex
    rate     rate.Limit
    burst    int
}

func NewRateLimiter(requestsPerSecond float64, burst int) *RateLimiter {
    rl := &RateLimiter{
        visitors: make(map[string]*visitor),
        rate:     rate.Limit(requestsPerSecond),
        burst:    burst,
    }
    
    // Clean up old visitors every minute
    go rl.cleanupVisitors()
    
    return rl
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip := r.RemoteAddr
        
        rl.mu.Lock()
        v, exists := rl.visitors[ip]
        if !exists {
            v = &visitor{
                limiter:  rate.NewLimiter(rl.rate, rl.burst),
                lastSeen: time.Now(),
            }
            rl.visitors[ip] = v
        }
        v.lastSeen = time.Now()
        rl.mu.Unlock()
        
        if !v.limiter.Allow() {
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

func (rl *RateLimiter) cleanupVisitors() {
    for {
        time.Sleep(time.Minute)
        
        rl.mu.Lock()
        for ip, v := range rl.visitors {
            if time.Since(v.lastSeen) > 3*time.Minute {
                delete(rl.visitors, ip)
            }
        }
        rl.mu.Unlock()
    }
}
```

### **8. Secure Crypto Implementation**

```go
// pkg/crypto/secure.go
package crypto

import (
    "crypto/rand"
    "crypto/subtle"
    "golang.org/x/crypto/argon2"
    "golang.org/x/crypto/bcrypt"
)

// HashPassword using bcrypt (replaces SHA-1)
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

// VerifyPassword checks password against hash
func VerifyPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

// SecureCompare performs constant-time comparison
func SecureCompare(a, b []byte) bool {
    return subtle.ConstantTimeCompare(a, b) == 1
}

// GenerateSecureToken creates cryptographically secure random token
func GenerateSecureToken(length int) ([]byte, error) {
    token := make([]byte, length)
    _, err := rand.Read(token)
    return token, err
}

// Argon2Hash for high-security password hashing
func Argon2Hash(password, salt []byte) []byte {
    return argon2.IDKey(password, salt, 1, 64*1024, 4, 32)
}
```

---

## 🔥 **EMERGENCY DEPLOYMENT SCRIPT**

```bash
#!/bin/bash
# scripts/emergency-security-fix.sh

set -e

echo "🚨 EMERGENCY SECURITY FIXES"
echo "==========================="

# 1. Generate secure passwords
echo "📝 Generating secure credentials..."
POSTGRES_PASS=$(openssl rand -base64 32)
GRAFANA_PASS=$(openssl rand -base64 32)
JWT_SECRET=$(openssl rand -base64 64)
MQTT_PASS=$(openssl rand -base64 32)

# 2. Create secure environment file
echo "📁 Creating secure environment..."
cat > .env.secure << EOF
POSTGRES_PASSWORD=${POSTGRES_PASS}
GRAFANA_ADMIN_PASSWORD=${GRAFANA_PASS}
JWT_SECRET=${JWT_SECRET}
MQTT_PASSWORD=${MQTT_PASS}
POSTGRES_USER=homeauto_admin
GRAFANA_ADMIN_USER=security_admin
MQTT_USERNAME=homeauto_mqtt
EOF

# 3. Backup current configuration
echo "💾 Backing up current config..."
cp deployments/docker-compose.yml deployments/docker-compose.yml.backup.$(date +%Y%m%d_%H%M%S)

# 4. Stop current services
echo "🛑 Stopping current services..."
cd deployments && docker-compose down

# 5. Remove exposed database port
echo "🔒 Securing database access..."
sed -i 's/- "5432:5432"/#- "5432:5432"  # SECURITY: Removed exposed port/' docker-compose.yml

# 6. Update environment references
echo "🔧 Updating configuration..."
sed -i 's/password/${POSTGRES_PASSWORD}/g' docker-compose.yml
sed -i 's/admin/${GRAFANA_ADMIN_PASSWORD}/g' docker-compose.yml

# 7. Setup MQTT security
echo "🔐 Configuring MQTT security..."
mkdir -p mosquitto/config
echo "allow_anonymous false" > mosquitto/mosquitto.secure.conf
echo "password_file /mosquitto/config/passwd" >> mosquitto/mosquitto.secure.conf

# 8. Start with secure configuration
echo "🚀 Starting with secure configuration..."
docker-compose --env-file ../.env.secure up -d

echo "✅ EMERGENCY SECURITY FIXES COMPLETE!"
echo ""
echo "🔑 IMPORTANT: Save these credentials securely:"
echo "   Database Password: ${POSTGRES_PASS}"
echo "   Grafana Password:  ${GRAFANA_PASS}"
echo "   MQTT Password:     ${MQTT_PASS}"
echo ""
echo "📋 Next steps:"
echo "   1. Test all services are working"
echo "   2. Configure TLS certificates"
echo "   3. Setup monitoring and alerting"
echo "   4. Implement remaining security recommendations"
```

---

## 📊 **SECURITY MONITORING**

### **9. Security Event Logging**

```go
// internal/security/logger.go
package security

import (
    "context"
    "log/slog"
    "net/http"
    "time"
)

type SecurityEvent struct {
    Type      string    `json:"type"`
    Source    string    `json:"source"`
    Message   string    `json:"message"`
    Severity  string    `json:"severity"`
    Timestamp time.Time `json:"timestamp"`
    Metadata  map[string]interface{} `json:"metadata"`
}

func LogSecurityEvent(event SecurityEvent) {
    slog.With(
        "security_event", true,
        "type", event.Type,
        "source", event.Source,
        "severity", event.Severity,
    ).Info(event.Message, "metadata", event.Metadata)
}

func LogFailedAuth(r *http.Request, reason string) {
    LogSecurityEvent(SecurityEvent{
        Type:      "authentication_failure",
        Source:    r.RemoteAddr,
        Message:   "Authentication failed",
        Severity:  "HIGH",
        Timestamp: time.Now(),
        Metadata: map[string]interface{}{
            "user_agent": r.UserAgent(),
            "path":       r.URL.Path,
            "reason":     reason,
        },
    })
}
```

### **10. Health Check with Security Status**

```go
// cmd/server/health.go
func securityHealthCheck(w http.ResponseWriter, r *http.Request) {
    checks := map[string]bool{
        "mqtt_auth_enabled":     checkMQTTAuth(),
        "database_ssl_enabled":  checkDatabaseSSL(), 
        "jwt_secret_secure":     checkJWTSecret(),
        "no_default_passwords":  checkDefaultPasswords(),
        "tls_enabled":          checkTLSConfig(),
    }
    
    allSecure := true
    for _, check := range checks {
        if !check {
            allSecure = false
            break
        }
    }
    
    response := map[string]interface{}{
        "status":          "ok",
        "security_status": "secure",
        "checks":         checks,
    }
    
    if !allSecure {
        response["security_status"] = "vulnerable"
        w.WriteHeader(http.StatusServiceUnavailable)
    }
    
    json.NewEncoder(w).Encode(response)
}
```

---

## 🎯 **IMPLEMENTATION PRIORITY**

### **TODAY (Critical)**
1. ✅ Run `scripts/emergency-security-fix.sh`
2. ✅ Change all default passwords
3. ✅ Remove exposed database port
4. ✅ Test system functionality

### **THIS WEEK (High)**
1. 🔄 Implement TLS certificates
2. 🔄 Add input validation middleware
3. 🔄 Configure MQTT authentication
4. 🔄 Add security headers

### **THIS MONTH (Medium)**
1. 📋 Implement rate limiting
2. 📋 Add security monitoring
3. 📋 Network segmentation
4. 📋 Container security hardening

### **ONGOING (Maintenance)**
1. 🔄 Regular dependency updates
2. 🔄 Security scanning automation
3. 🔄 Monitoring and alerting
4. 🔄 Regular security audits

---

## ⚠️ **DEPLOYMENT WARNING**

**🚨 DO NOT DEPLOY TO PRODUCTION** until at least these critical fixes are implemented:

1. ✅ All default passwords changed
2. ✅ Database port not exposed  
3. ✅ MQTT authentication enabled
4. ✅ Strong JWT secrets configured
5. ✅ Non-root containers configured

**🔒 System will be production-ready** after implementing the emergency fixes and high-priority security enhancements listed above.

Use the emergency script to implement critical fixes immediately, then follow the implementation guide for comprehensive security hardening.
