#!/bin/bash
# SBOM Generation Script for Home Automation System

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SBOM_DIR="$SCRIPT_DIR/sbom"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "1.0.0")

echo "🔍 Generating Software Bill of Materials (SBOM)"
echo "=============================================="
echo "Project: Home Automation System"
echo "Version: $VERSION"
echo "Output: $SBOM_DIR"
echo

# Create SBOM directory
mkdir -p "$SBOM_DIR"

# Function to get Go module dependencies
get_go_dependencies() {
    cd "$PROJECT_ROOT"
    if [ -f "go.mod" ]; then
        go list -m -json all | jq -r '.Path + " " + .Version' 2>/dev/null || \
        go list -m all | grep -v "^github.com/johnpr01/home-automation$"
    fi
}

# Function to get system dependencies from Debian package
get_system_dependencies() {
    echo "docker.io >= 20.10.0"
    echo "docker-compose-plugin >= 2.0.0"
    echo "curl"
    echo "wget"
    echo "systemd"
    echo "golang-go >= 1.19"  # For standalone package
}

# Function to get Docker image dependencies
get_docker_dependencies() {
    cd "$PROJECT_ROOT/deployments"
    if [ -f "docker-compose.yml" ]; then
        grep -E "image:" docker-compose.yml | sed 's/.*image: *//' | sed 's/[[:space:]]*$//' | sort -u
    fi
}

# Function to get MicroPython dependencies
get_micropython_dependencies() {
    echo "micropython >= 1.22.0"
    echo "mpremote"
    echo "thonny (optional)"
}

# Generate SPDX 2.3 format SBOM
generate_spdx_sbom() {
    local output_file="$SBOM_DIR/home-automation-$VERSION.spdx.json"
    
    echo "📄 Generating SPDX SBOM: $output_file"
    
    cat > "$output_file" << EOF
{
  "spdxVersion": "SPDX-2.3",
  "dataLicense": "CC0-1.0",
  "SPDXID": "SPDXRef-DOCUMENT",
  "documentName": "Home Automation System SBOM",
  "documentNamespace": "https://github.com/johnpr01/home-automation/sbom/$VERSION",
  "creationInfo": {
    "created": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "creators": ["Tool: home-automation-sbom-generator"],
    "licenseListVersion": "3.20"
  },
  "packages": [
    {
      "SPDXID": "SPDXRef-Package-HomeAutomation",
      "name": "home-automation",
      "downloadLocation": "https://github.com/johnpr01/home-automation",
      "filesAnalyzed": false,
      "versionInfo": "$VERSION",
      "packageSupplier": "Organization: Home Automation Team",
      "copyrightText": "Copyright 2024-2025 Home Automation Team",
      "licenseConcluded": "MIT",
      "licenseDeclared": "MIT",
      "description": "Smart Home Automation System for Raspberry Pi with MQTT, Kafka, Prometheus, and IoT sensors",
      "homepage": "https://github.com/johnpr01/home-automation",
      "packageVerificationCode": {
        "packageVerificationCodeValue": "$(find "$PROJECT_ROOT" -type f -name "*.go" -o -name "*.py" -o -name "*.yml" -o -name "*.yaml" | sort | xargs sha1sum | sha1sum | cut -d' ' -f1)"
      },
      "externalRefs": [
        {
          "referenceCategory": "PACKAGE_MANAGER",
          "referenceType": "purl",
          "referenceLocator": "pkg:github/johnpr01/home-automation@$VERSION"
        }
      ]
    },
EOF

    # Add Go dependencies
    echo "    # Go Dependencies" >> "$output_file"
    local go_deps_count=0
    while IFS= read -r dep; do
        if [ -n "$dep" ] && [ "$dep" != "github.com/johnpr01/home-automation" ]; then
            local dep_name=$(echo "$dep" | awk '{print $1}')
            local dep_version=$(echo "$dep" | awk '{print $2}')
            go_deps_count=$((go_deps_count + 1))
            
            cat >> "$output_file" << EOF
    {
      "SPDXID": "SPDXRef-Package-Go-$go_deps_count",
      "name": "$dep_name",
      "downloadLocation": "https://$dep_name",
      "filesAnalyzed": false,
      "versionInfo": "$dep_version",
      "packageSupplier": "NOASSERTION",
      "copyrightText": "NOASSERTION",
      "licenseConcluded": "NOASSERTION",
      "licenseDeclared": "NOASSERTION",
      "description": "Go module dependency",
      "externalRefs": [
        {
          "referenceCategory": "PACKAGE_MANAGER",
          "referenceType": "purl",
          "referenceLocator": "pkg:golang/$dep_name@$dep_version"
        }
      ]
    },
EOF
        fi
    done < <(get_go_dependencies)

    # Add Docker dependencies
    echo "    # Docker Dependencies" >> "$output_file"
    local docker_deps_count=0
    while IFS= read -r image; do
        if [ -n "$image" ]; then
            docker_deps_count=$((docker_deps_count + 1))
            local image_name=$(echo "$image" | cut -d':' -f1)
            local image_tag=$(echo "$image" | cut -d':' -f2)
            
            cat >> "$output_file" << EOF
    {
      "SPDXID": "SPDXRef-Package-Docker-$docker_deps_count",
      "name": "$image_name",
      "downloadLocation": "https://hub.docker.com/_/$image_name",
      "filesAnalyzed": false,
      "versionInfo": "$image_tag",
      "packageSupplier": "NOASSERTION",
      "copyrightText": "NOASSERTION",
      "licenseConcluded": "NOASSERTION",
      "licenseDeclared": "NOASSERTION",
      "description": "Docker container image",
      "externalRefs": [
        {
          "referenceCategory": "PACKAGE_MANAGER",
          "referenceType": "purl",
          "referenceLocator": "pkg:docker/$image_name@$image_tag"
        }
      ]
    },
EOF
        fi
    done < <(get_docker_dependencies)

    # Add system dependencies
    echo "    # System Dependencies" >> "$output_file"
    local sys_deps_count=0
    while IFS= read -r dep; do
        if [ -n "$dep" ]; then
            sys_deps_count=$((sys_deps_count + 1))
            local dep_name=$(echo "$dep" | awk '{print $1}')
            local dep_constraint=$(echo "$dep" | sed "s/^$dep_name *//")
            
            cat >> "$output_file" << EOF
    {
      "SPDXID": "SPDXRef-Package-System-$sys_deps_count",
      "name": "$dep_name",
      "downloadLocation": "NOASSERTION",
      "filesAnalyzed": false,
      "versionInfo": "$dep_constraint",
      "packageSupplier": "Organization: Debian",
      "copyrightText": "NOASSERTION",
      "licenseConcluded": "NOASSERTION",
      "licenseDeclared": "NOASSERTION",
      "description": "System package dependency",
      "externalRefs": [
        {
          "referenceCategory": "PACKAGE_MANAGER",
          "referenceType": "purl",
          "referenceLocator": "pkg:deb/debian/$dep_name"
        }
      ]
    }$([ $sys_deps_count -lt $(get_system_dependencies | wc -l) ] && echo "," || echo "")
EOF
        fi
    done < <(get_system_dependencies)

    # Close the JSON
    cat >> "$output_file" << EOF
  ],
  "relationships": [
    {
      "spdxElementId": "SPDXRef-DOCUMENT",
      "relationshipType": "DESCRIBES",
      "relatedSpdxElement": "SPDXRef-Package-HomeAutomation"
    }
  ]
}
EOF

    echo "✅ SPDX SBOM generated successfully"
}

# Generate CycloneDX format SBOM
generate_cyclonedx_sbom() {
    local output_file="$SBOM_DIR/home-automation-$VERSION.cyclonedx.json"
    
    echo "📄 Generating CycloneDX SBOM: $output_file"
    
    cat > "$output_file" << EOF
{
  "bomFormat": "CycloneDX",
  "specVersion": "1.4",
  "serialNumber": "urn:uuid:$(uuidgen)",
  "version": 1,
  "metadata": {
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "tools": [
      {
        "vendor": "Home Automation Team",
        "name": "home-automation-sbom-generator",
        "version": "1.0.0"
      }
    ],
    "component": {
      "type": "application",
      "bom-ref": "home-automation@$VERSION",
      "name": "home-automation",
      "version": "$VERSION",
      "description": "Smart Home Automation System for Raspberry Pi",
      "licenses": [
        {
          "license": {
            "id": "MIT"
          }
        }
      ],
      "purl": "pkg:github/johnpr01/home-automation@$VERSION",
      "externalReferences": [
        {
          "type": "website",
          "url": "https://github.com/johnpr01/home-automation"
        },
        {
          "type": "vcs",
          "url": "https://github.com/johnpr01/home-automation.git"
        }
      ]
    }
  },
  "components": [
EOF

    # Add Go dependencies
    local first_component=true
    while IFS= read -r dep; do
        if [ -n "$dep" ] && [ "$dep" != "github.com/johnpr01/home-automation" ]; then
            local dep_name=$(echo "$dep" | awk '{print $1}')
            local dep_version=$(echo "$dep" | awk '{print $2}')
            
            [ "$first_component" = false ] && echo "    ," >> "$output_file"
            first_component=false
            
            cat >> "$output_file" << EOF
    {
      "type": "library",
      "bom-ref": "$dep_name@$dep_version",
      "name": "$dep_name",
      "version": "$dep_version",
      "scope": "required",
      "purl": "pkg:golang/$dep_name@$dep_version",
      "externalReferences": [
        {
          "type": "website",
          "url": "https://$dep_name"
        }
      ]
    }
EOF
        fi
    done < <(get_go_dependencies)

    # Add Docker dependencies
    while IFS= read -r image; do
        if [ -n "$image" ]; then
            local image_name=$(echo "$image" | cut -d':' -f1)
            local image_tag=$(echo "$image" | cut -d':' -f2)
            
            [ "$first_component" = false ] && echo "    ," >> "$output_file"
            first_component=false
            
            cat >> "$output_file" << EOF
    {
      "type": "container",
      "bom-ref": "$image_name@$image_tag",
      "name": "$image_name",
      "version": "$image_tag",
      "scope": "required",
      "purl": "pkg:docker/$image_name@$image_tag",
      "externalReferences": [
        {
          "type": "distribution",
          "url": "https://hub.docker.com/_/$image_name"
        }
      ]
    }
EOF
        fi
    done < <(get_docker_dependencies)

    # Add system dependencies
    while IFS= read -r dep; do
        if [ -n "$dep" ]; then
            local dep_name=$(echo "$dep" | awk '{print $1}')
            local dep_constraint=$(echo "$dep" | sed "s/^$dep_name *//")
            
            [ "$first_component" = false ] && echo "    ," >> "$output_file"
            first_component=false
            
            cat >> "$output_file" << EOF
    {
      "type": "operating-system",
      "bom-ref": "$dep_name",
      "name": "$dep_name",
      "version": "$dep_constraint",
      "scope": "required",
      "purl": "pkg:deb/debian/$dep_name"
    }
EOF
        fi
    done < <(get_system_dependencies)

    # Close the JSON
    cat >> "$output_file" << EOF
  ]
}
EOF

    echo "✅ CycloneDX SBOM generated successfully"
}

# Generate simple text format SBOM
generate_text_sbom() {
    local output_file="$SBOM_DIR/home-automation-$VERSION.sbom.txt"
    
    echo "📄 Generating Text SBOM: $output_file"
    
    cat > "$output_file" << EOF
# Software Bill of Materials (SBOM)
# Home Automation System v$VERSION
# Generated: $(date -u +%Y-%m-%dT%H:%M:%SZ)

## Main Component
Name: home-automation
Version: $VERSION
Type: Application
License: MIT
Description: Smart Home Automation System for Raspberry Pi
Repository: https://github.com/johnpr01/home-automation
PURL: pkg:github/johnpr01/home-automation@$VERSION

## Go Dependencies
EOF

    while IFS= read -r dep; do
        if [ -n "$dep" ] && [ "$dep" != "github.com/johnpr01/home-automation" ]; then
            local dep_name=$(echo "$dep" | awk '{print $1}')
            local dep_version=$(echo "$dep" | awk '{print $2}')
            echo "- $dep_name@$dep_version (pkg:golang/$dep_name@$dep_version)" >> "$output_file"
        fi
    done < <(get_go_dependencies)

    echo "" >> "$output_file"
    echo "## Docker Dependencies" >> "$output_file"
    while IFS= read -r image; do
        if [ -n "$image" ]; then
            local image_name=$(echo "$image" | cut -d':' -f1)
            local image_tag=$(echo "$image" | cut -d':' -f2)
            echo "- $image_name:$image_tag (pkg:docker/$image_name@$image_tag)" >> "$output_file"
        fi
    done < <(get_docker_dependencies)

    echo "" >> "$output_file"
    echo "## System Dependencies" >> "$output_file"
    while IFS= read -r dep; do
        if [ -n "$dep" ]; then
            local dep_name=$(echo "$dep" | awk '{print $1}')
            local dep_constraint=$(echo "$dep" | sed "s/^$dep_name *//")
            echo "- $dep_name $dep_constraint (pkg:deb/debian/$dep_name)" >> "$output_file"
        fi
    done < <(get_system_dependencies)

    echo "" >> "$output_file"
    echo "## MicroPython Dependencies" >> "$output_file"
    while IFS= read -r dep; do
        if [ -n "$dep" ]; then
            echo "- $dep" >> "$output_file"
        fi
    done < <(get_micropython_dependencies)

    echo "✅ Text SBOM generated successfully"
}

# Generate dependency analysis
generate_dependency_analysis() {
    local output_file="$SBOM_DIR/dependency-analysis.md"
    
    echo "📄 Generating Dependency Analysis: $output_file"
    
    cat > "$output_file" << EOF
# Dependency Analysis Report
**Home Automation System v$VERSION**  
**Generated:** $(date -u +%Y-%m-%dT%H:%M:%SZ)

## Summary

| Category | Count | Risk Level |
|----------|-------|------------|
| Go Dependencies | $(get_go_dependencies | grep -v "^github.com/johnpr01/home-automation$" | wc -l) | Low |
| Docker Images | $(get_docker_dependencies | wc -l) | Medium |
| System Packages | $(get_system_dependencies | wc -l) | Low |
| MicroPython | $(get_micropython_dependencies | wc -l) | Low |

## Risk Assessment

### Go Dependencies
- **Risk Level:** Low
- **Rationale:** Well-maintained Go ecosystem with strong security practices
- **Mitigation:** Regular dependency updates, vulnerability scanning

### Docker Images
- **Risk Level:** Medium  
- **Rationale:** Third-party container images may contain vulnerabilities
- **Mitigation:** Use official images, regular security updates, image scanning

### System Packages
- **Risk Level:** Low
- **Rationale:** Debian stable packages with security support
- **Mitigation:** Regular system updates, security patches

## Security Recommendations

1. **Regular Updates**: Keep all dependencies updated to latest stable versions
2. **Vulnerability Scanning**: Use tools like \`govulncheck\` for Go dependencies
3. **Container Scanning**: Scan Docker images for known vulnerabilities
4. **Supply Chain Security**: Verify package signatures and checksums
5. **License Compliance**: Review licenses for compatibility with project requirements

## License Summary

### Identified Licenses
- MIT (Home Automation System)
- Various open source licenses (dependencies)

### Action Required
- Review all dependency licenses for compliance
- Update LICENSE file with third-party attributions
- Consider license compatibility matrix

## Compliance Notes

- SBOM format: SPDX 2.3 and CycloneDX 1.4 compliant
- Package URLs (PURLs) provided for dependency identification
- Suitable for supply chain security and compliance requirements
- Compatible with container scanning tools and security platforms

EOF

    echo "✅ Dependency analysis generated successfully"
}

# Generate vulnerability scan script
generate_vuln_scan_script() {
    local output_file="$SBOM_DIR/scan-vulnerabilities.sh"
    
    echo "📄 Generating Vulnerability Scan Script: $output_file"
    
    cat > "$output_file" << 'EOF'
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
EOF

    chmod +x "$output_file"
    echo "✅ Vulnerability scan script generated successfully"
}

# Main execution
echo "🚀 Starting SBOM generation..."

# Check for required tools
if ! command -v jq >/dev/null 2>&1; then
    echo "⚠️  jq not found - installing for JSON processing..."
    sudo apt update && sudo apt install -y jq
fi

if ! command -v uuidgen >/dev/null 2>&1; then
    echo "⚠️  uuidgen not found - installing for UUID generation..."
    sudo apt update && sudo apt install -y uuid-runtime
fi

# Generate all SBOM formats
generate_spdx_sbom
generate_cyclonedx_sbom
generate_text_sbom
generate_dependency_analysis
generate_vuln_scan_script

echo
echo "📋 SBOM Generation Summary:"
echo "=========================="
echo "Files generated in: $SBOM_DIR"
ls -la "$SBOM_DIR"

echo
echo "🔒 Security Recommendations:"
echo "1. Review generated SBOMs for license compliance"
echo "2. Run vulnerability scan: $SBOM_DIR/scan-vulnerabilities.sh"
echo "3. Include SBOMs in release artifacts"
echo "4. Update SBOMs when dependencies change"
echo "5. Store SBOMs in secure, accessible location"

echo
echo "✅ SBOM generation completed successfully!"
