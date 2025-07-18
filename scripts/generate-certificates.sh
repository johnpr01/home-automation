#!/bin/bash

# generate-certificates.sh - Generate TLS Certificates for Home Automation System
# WARNING: This script generates self-signed certificates for development/testing only
# For production, use certificates from a trusted Certificate Authority

set -e

echo "🔐 Generating TLS Certificates for Home Automation System"
echo "========================================================="

# Get the project root directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CERT_DIR="$PROJECT_ROOT/certs"

# Configuration
COUNTRY="US"
STATE="State"
CITY="City"
ORG="HomeAutomation"
ORG_UNIT="IT"
CA_CN="HomeAutomation-CA"
SERVER_CN="homeautomation.local"
CLIENT_CN="mqtt-client"
KEY_SIZE=4096
VALIDITY_DAYS=365

# Get local IP for certificate
LOCAL_IP=$(hostname -I | awk '{print $1}' || echo "192.168.1.100")
echo "🌐 Detected local IP: $LOCAL_IP"

# Create certificate directory
echo "📁 Creating certificate directory..."
mkdir -p "$CERT_DIR"
cd "$CERT_DIR"

# Clean up existing certificates
echo "🧹 Cleaning up existing certificates..."
rm -f *.pem *.crt *.csr *.cnf *.srl

echo ""
echo "🔑 Step 1: Generating Certificate Authority (CA)..."

# Generate CA private key
echo "   Generating CA private key..."
openssl genrsa -out ca-key.pem "$KEY_SIZE"

# Generate CA certificate
echo "   Generating CA certificate..."
openssl req -new -x509 -days "$VALIDITY_DAYS" \
    -key ca-key.pem \
    -sha256 \
    -out ca.crt \
    -subj "/C=$COUNTRY/ST=$STATE/L=$CITY/O=$ORG/OU=$ORG_UNIT/CN=$CA_CN"

echo "   ✅ CA certificate generated: ca.crt"

echo ""
echo "🖥️  Step 2: Generating Server Certificate..."

# Generate server private key
echo "   Generating server private key..."
openssl genrsa -out server-key.pem "$KEY_SIZE"

# Generate server certificate signing request
echo "   Generating server CSR..."
openssl req -subj "/C=$COUNTRY/ST=$STATE/L=$CITY/O=$ORG/OU=$ORG_UNIT/CN=$SERVER_CN" \
    -sha256 -new -key server-key.pem -out server.csr

# Create extensions file for server certificate
echo "   Creating server certificate extensions..."
cat > server-extensions.cnf << EOF
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = $SERVER_CN
DNS.2 = localhost
DNS.3 = *.homeautomation.local
DNS.4 = grafana.homeautomation.local
DNS.5 = prometheus.homeautomation.local
DNS.6 = metrics.homeautomation.local
IP.1 = 127.0.0.1
IP.2 = $LOCAL_IP
IP.3 = 192.168.1.100
IP.4 = 192.168.68.100
IP.5 = 10.0.0.100
EOF

# Generate server certificate
echo "   Generating server certificate..."
openssl x509 -req -days "$VALIDITY_DAYS" \
    -in server.csr \
    -CA ca.crt \
    -CAkey ca-key.pem \
    -out server.crt \
    -extensions v3_req \
    -extfile server-extensions.cnf \
    -CAcreateserial

echo "   ✅ Server certificate generated: server.crt"

echo ""
echo "👤 Step 3: Generating Client Certificate (for MQTT)..."

# Generate client private key
echo "   Generating client private key..."
openssl genrsa -out client-key.pem "$KEY_SIZE"

# Generate client certificate signing request
echo "   Generating client CSR..."
openssl req -subj "/C=$COUNTRY/ST=$STATE/L=$CITY/O=$ORG/OU=Client/CN=$CLIENT_CN" \
    -new -key client-key.pem -out client.csr

# Create extensions file for client certificate
echo "   Creating client certificate extensions..."
cat > client-extensions.cnf << 'EOF'
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth
EOF

# Generate client certificate
echo "   Generating client certificate..."
openssl x509 -req -days "$VALIDITY_DAYS" \
    -in client.csr \
    -CA ca.crt \
    -CAkey ca-key.pem \
    -out client.crt \
    -extensions v3_req \
    -extfile client-extensions.cnf \
    -CAcreateserial

echo "   ✅ Client certificate generated: client.crt"

echo ""
echo "🔐 Step 4: Setting proper file permissions..."

# Set proper permissions
chmod 644 *.crt
chmod 600 *-key.pem
chmod 644 *.cnf

# Set ownership (if running as root, change to appropriate user)
if [ "$EUID" -eq 0 ]; then
    chown -R 1000:1000 .
    echo "   ✅ Ownership set to user 1000:1000"
else
    echo "   ✅ Permissions set for current user"
fi

echo ""
echo "🔍 Step 5: Verifying certificates..."

# Verify server certificate
echo "   Verifying server certificate chain..."
if openssl verify -CAfile ca.crt server.crt > /dev/null 2>&1; then
    echo "   ✅ Server certificate chain valid"
else
    echo "   ❌ Server certificate chain invalid"
    exit 1
fi

# Verify client certificate
echo "   Verifying client certificate chain..."
if openssl verify -CAfile ca.crt client.crt > /dev/null 2>&1; then
    echo "   ✅ Client certificate chain valid"
else
    echo "   ❌ Client certificate chain invalid"
    exit 1
fi

# Display certificate information
echo ""
echo "📋 Certificate Information:"
echo "=========================="

echo ""
echo "🏛️  Certificate Authority (CA):"
openssl x509 -in ca.crt -noout -subject -issuer -dates

echo ""
echo "🖥️  Server Certificate:"
openssl x509 -in server.crt -noout -subject -issuer -dates
echo "   Subject Alternative Names:"
openssl x509 -in server.crt -noout -text | grep -A 10 "Subject Alternative Name" | tail -n +2

echo ""
echo "👤 Client Certificate:"
openssl x509 -in client.crt -noout -subject -issuer -dates

# Clean up temporary files
echo ""
echo "🧹 Cleaning up temporary files..."
rm -f *.csr *.cnf *.srl

echo ""
echo "🎉 TLS Certificate Generation Complete!"
echo "======================================"
echo ""
echo "📁 Certificates generated in: $CERT_DIR"
echo ""
echo "📜 Generated files:"
echo "   🔑 ca.crt             - Certificate Authority certificate"
echo "   🔐 ca-key.pem         - Certificate Authority private key"
echo "   🖥️  server.crt        - Server certificate"
echo "   🔐 server-key.pem     - Server private key"
echo "   👤 client.crt         - Client certificate (MQTT)"
echo "   🔐 client-key.pem     - Client private key (MQTT)"
echo ""
echo "⚠️  SECURITY NOTES:"
echo "   • These are self-signed certificates for development/testing"
echo "   • Private keys are protected with 600 permissions"
echo "   • For production, obtain certificates from a trusted CA"
echo "   • Add ca.crt to your browser's trusted certificates"
echo ""
echo "🔗 Next steps:"
echo "   1. Run: ./deploy-tls.sh to deploy TLS configuration"
echo "   2. Run: ./verify-tls.sh to verify TLS is working"
echo ""
