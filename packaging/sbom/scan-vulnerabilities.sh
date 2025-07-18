#!/bin/bash
# Vulnerability Scanning Script for Home Automation System

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "🔍 Home Automation System - Vulnerability Scan"
echo "=============================================="
echo

# Function to scan Go dependencies
scan_go_dependencies() {
    echo "🔍 Scanning Go dependencies for vulnerabilities..."
    
    cd "$PROJECT_ROOT"
    
    # Install govulncheck if not available
    if ! command -v govulncheck >/dev/null 2>&1; then
        echo "📦 Installing govulncheck..."
        go install golang.org/x/vuln/cmd/govulncheck@latest
    fi
    
    # Run vulnerability check
    echo "Running govulncheck..."
    if govulncheck ./...; then
        echo "✅ No Go vulnerabilities found"
    else
        echo "⚠️  Go vulnerabilities detected - see output above"
    fi
    echo
}

# Function to scan Docker images
scan_docker_images() {
    echo "🔍 Scanning Docker images for vulnerabilities..."
    
    cd "$PROJECT_ROOT/deployments"
    
    # Extract image names from docker-compose.yml
    images=$(grep -E "image:" docker-compose.yml | sed 's/.*image: *//' | sed 's/[[:space:]]*$//')
    
    for image in $images; do
        echo "Scanning image: $image"
        
        # Try different scanning tools
        if command -v trivy >/dev/null 2>&1; then
            trivy image "$image"
        elif command -v grype >/dev/null 2>&1; then
            grype "$image"
        elif command -v docker >/dev/null 2>&1; then
            # Use Docker Scout if available
            docker scout cves "$image" 2>/dev/null || echo "⚠️  No vulnerability scanner available for Docker images"
        else
            echo "⚠️  No Docker vulnerability scanner found"
            echo "   Install trivy, grype, or Docker Scout for image scanning"
        fi
        echo
    done
}

# Function to check for security updates
check_system_updates() {
    echo "🔍 Checking for system security updates..."
    
    if command -v apt >/dev/null 2>&1; then
        echo "Checking apt packages..."
        apt list --upgradable 2>/dev/null | grep -E "(security|Security)" || echo "No security updates available"
    else
        echo "⚠️  APT not available - cannot check system updates"
    fi
    echo
}

# Function to generate vulnerability report
generate_report() {
    local report_file="$SCRIPT_DIR/vulnerability-report-$(date +%Y%m%d).md"
    
    echo "📄 Generating vulnerability report: $report_file"
    
    cat > "$report_file" << REPORT_EOF
# Vulnerability Scan Report
**Date:** $(date -u +%Y-%m-%dT%H:%M:%SZ)  
**Project:** Home Automation System  
**Scan Type:** Comprehensive Security Assessment

## Scan Summary
- Go Dependencies: Scanned with govulncheck
- Docker Images: Scanned with available tools
- System Packages: Checked for security updates

## Recommendations
1. Review any vulnerabilities found above
2. Update affected dependencies to patched versions
3. Rebuild Docker images with updated base images
4. Apply system security updates
5. Re-run scan after fixes

## Next Steps
- Schedule regular vulnerability scans
- Set up automated dependency updates where appropriate
- Consider integrating scanning into CI/CD pipeline
- Monitor security advisories for used technologies
REPORT_EOF

    echo "✅ Vulnerability report saved to: $report_file"
}

# Main execution
echo "Starting comprehensive vulnerability scan..."
echo

scan_go_dependencies
scan_docker_images  
check_system_updates
generate_report

echo "🎉 Vulnerability scan completed!"
echo "📋 Review the report and address any findings"
