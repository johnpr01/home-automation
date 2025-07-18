# Software Bill of Materials (SBOM) - Quick Reference

## 🎯 **What is an SBOM?**

A Software Bill of Materials (SBOM) is a formal record containing the details and supply chain relationships of various components used in building software. Think of it as an "ingredient list" for software.

## 📁 **Files in this Directory**

| File | Format | Purpose |
|------|--------|---------|
| `*.spdx.json` | SPDX 2.3 | Industry standard, legal compliance |
| `*.cyclonedx.json` | CycloneDX 1.4 | Security analysis, vulnerability scanning |
| `*.sbom.txt` | Text | Human-readable, quick reference |
| `dependency-analysis.md` | Markdown | Risk assessment and recommendations |
| `scan-vulnerabilities.sh` | Script | Automated security scanning |

## 🚀 **Quick Start**

### **View Dependencies**
```bash
# Human-readable format
cat home-automation-*.sbom.txt

# Component count
grep "^-" home-automation-*.sbom.txt | wc -l
```

### **Security Scan**
```bash
# Run vulnerability assessment
./scan-vulnerabilities.sh
```

### **License Review**
```bash
# Check for license information
grep -i license dependency-analysis.md
```

## 🔍 **What's Included**

### **140+ Go Dependencies**
- Prometheus client libraries
- HTTP frameworks (Fiber, Gin) 
- MQTT/Kafka clients
- Cryptographic libraries
- Database drivers

### **6 Docker Images**
- prometheus/prometheus:latest
- grafana/grafana:latest
- confluentinc/cp-kafka:latest
- eclipse-mosquitto:latest
- And more...

### **System Packages**
- Docker & Docker Compose
- System utilities (curl, wget, systemd)
- Go compiler (standalone package)

### **MicroPython Components**
- MicroPython runtime
- Sensor drivers
- Development tools

## 🔒 **Security Use Cases**

### **Vulnerability Management**
- Identify components with known CVEs
- Track security updates needed
- Prioritize patching efforts

### **License Compliance**
- Review third-party licenses
- Generate attribution reports
- Ensure license compatibility

### **Supply Chain Security**
- Complete software inventory
- Track component origins
- Assess supply chain risks

## 📊 **Risk Assessment**

| Component Type | Risk Level | Rationale |
|----------------|------------|-----------|
| Go Dependencies | **Low** | Well-maintained ecosystem |
| Docker Images | **Medium** | Third-party containers |
| System Packages | **Low** | Debian stable packages |
| MicroPython | **Low** | Minimal attack surface |

## 🛠️ **Tools Integration**

### **Vulnerability Scanners**
```bash
# Trivy (recommended)
trivy sbom home-automation-*.spdx.json

# Grype  
grype sbom:home-automation-*.cyclonedx.json

# OWASP Dependency Check
dependency-check.sh --scan .
```

### **License Analysis**
```bash
# SPDX tools
spdx-tools validate home-automation-*.spdx.json

# FOSSology
fossology -f home-automation-*.spdx.json
```

## 📋 **Maintenance**

### **When to Update SBOM**
- ✅ **Dependency changes** (go.mod updates)
- ✅ **New releases** (version updates)
- ✅ **Security patches** (vulnerability fixes)
- ✅ **License changes** (dependency license updates)

### **How to Update**
```bash
# Regenerate SBOM
cd ../
./generate-sbom.sh

# Rebuild package with new SBOM
./build-deb.sh
```

## 🎯 **Standards Compliance**

### **NTIA Minimum Elements**
- ✅ **Component Name**: All components identified
- ✅ **Component Version**: Exact versions specified  
- ✅ **Unique Identifier**: PURLs provided
- ✅ **Dependency Relationships**: Captured in SPDX
- ✅ **SBOM Author**: Tool and timestamp included
- ✅ **SBOM Timestamp**: Generation time recorded

### **Format Standards**
- ✅ **SPDX 2.3**: ISO/IEC 5962:2021 compliant
- ✅ **CycloneDX 1.4**: OWASP specification
- ✅ **Package URLs**: PURL specification

## 📞 **Support**

### **Questions?**
- **Security concerns**: Review `dependency-analysis.md`
- **License issues**: Check SPDX license fields
- **Tool integration**: See `../../../chats/SBOM_DOCUMENTATION.md`
- **Updates needed**: Run `./generate-sbom.sh`

### **Reporting Issues**
If you find missing components or inaccurate information:
1. Check the source code for new dependencies
2. Regenerate SBOM with latest tools
3. Report persistent issues to maintainers

---

**Generated**: $(date -u +%Y-%m-%dT%H:%M:%SZ)  
**Tool**: home-automation-sbom-generator v1.0.0  
**Standards**: SPDX 2.3, CycloneDX 1.4, NTIA Minimum Elements
