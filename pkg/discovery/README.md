# Asset Discovery Protocol

A multicast-based asset discovery protocol for home automation systems that enables automatic device discovery and network topology mapping.

## 🌟 **Overview**

The Asset Discovery Protocol allows devices in a home automation network to:

- **Announce their presence** and capabilities automatically
- **Discover other devices** on the network
- **Query for specific device types** or capabilities
- **Maintain an up-to-date inventory** of network assets
- **Handle device arrivals and departures** gracefully

## 🏗️ **Architecture**

### **Protocol Components**

1. **Discovery Protocol** (`pkg/discovery/protocol.go`)
   - Low-level multicast UDP communication
   - Message serialization/deserialization
   - Asset lifecycle management

2. **Discovery Manager** (`pkg/discovery/manager.go`)
   - High-level asset management
   - Event handling and notifications
   - Statistics and logging

3. **Asset Builder** (`pkg/discovery/builder.go`)
   - Fluent API for creating asset definitions
   - Auto-detection of network information
   - Predefined builders for common device types

4. **Discovery CLI** (`cmd/discovery/main.go`)
   - Command-line tool for testing and demonstration
   - Multiple operation modes (discover, announce, query)

### **Key Concepts**

- **Asset**: Any device or service in the network
- **Asset Type**: Category of device (sensor, smart plug, gateway, etc.)
- **Capability**: What an asset can do (temperature sensing, power control, etc.)
- **Service**: Network service offered by an asset (HTTP API, MQTT topic, etc.)

## 📡 **Protocol Specification**

### **Multicast Configuration**
- **Address**: `239.255.42.42`
- **Port**: `42424`
- **Protocol**: UDP
- **Message Format**: JSON
- **Max Message Size**: 8192 bytes

### **Message Types**

1. **Announce** - Device announces its presence
2. **Query** - Request for devices matching criteria
3. **Response** - Reply to a query
4. **Goodbye** - Device leaving the network

### **Message Structure**

```json
{
  "type": "announce|query|response|goodbye",
  "asset": {
    "id": "unique-asset-identifier",
    "name": "Human readable name",
    "type": "sensor|smart_plug|gateway|...",
    "model": "Device model",
    "manufacturer": "Device manufacturer",
    "version": "Firmware version",
    "ip_address": "192.168.1.100",
    "mac_address": "00:11:22:33:44:55",
    "hostname": "device-hostname",
    "ports": [80, 443],
    "capabilities": ["temperature", "humidity", "http"],
    "services": [
      {
        "name": "temperature-api",
        "protocol": "http",
        "port": 8080,
        "path": "/api/temperature",
        "description": "Temperature API"
      }
    ],
    "room": "living-room",
    "zone": "main-floor",
    "tags": ["sensor", "environmental"],
    "metadata": {
      "os": "linux",
      "arch": "arm64"
    },
    "status": "online",
    "health": "healthy",
    "battery_level": 85,
    "last_seen": "2025-07-18T10:30:00Z",
    "discovery_version": "1.0",
    "sequence": 1,
    "ttl": 300
  },
  "query": {
    "asset_types": ["sensor", "smart_plug"],
    "capabilities": ["temperature", "power"],
    "room": "living-room",
    "zone": "main-floor",
    "tags": ["environmental"],
    "max_age": "10m"
  },
  "timestamp": "2025-07-18T10:30:00Z",
  "sender": "gateway-001"
}
```

## 🚀 **Usage**

### **Discovery CLI Tool**

Build and run the discovery tool:

```bash
# Build the discovery tool
cd cmd/discovery
go build -o discovery

# DISCOVER mode: Passively listen for responses to other queries
# Note: This mode only receives responses when other devices send queries
./discovery -mode=discover -duration=60s -verbose

# ANNOUNCE mode: Announce your presence and respond to queries from others
./discovery -mode=announce -type=gateway -name="Home Gateway" -room="office" -duration=300s

# QUERY mode: Actively query for devices (recommended for discovery)
./discovery -mode=query -query-types="sensor,smart_plug" -room="living-room" -duration=30s

# JSON output format for programmatic use
./discovery -mode=query -query-types="gateway" -json -duration=30s > discovered_assets.json
```

### **Discovery Modes Explained**

- **Discover Mode**: Passively listens for discovery traffic. You'll only see assets when other devices send queries.
- **Announce Mode**: Makes your device visible to others. Responds to incoming queries with your device information.
- **Query Mode**: Actively searches for devices by sending discovery queries. This is the most effective way to find devices.

### **Programmatic Usage**

```go
package main

import (
    "log"
    "time"
    "github.com/johnpr01/home-automation/pkg/discovery"
)

func main() {
    // Create a local asset (this device)
    localAsset := discovery.NewHomeAutomationGateway("Main Gateway").
        WithRoom("office").
        WithZone("main-floor").
        WithTag("primary").
        Build()

    // Configure discovery manager
    config := discovery.DiscoveryConfig{
        LocalAsset:    localAsset,
        AutoQuery:     true,
        QueryInterval: 5 * time.Minute,
        Logger:        log.Default(),
    }

    // Create and start discovery manager
    manager, err := discovery.NewDiscoveryManager(config)
    if err != nil {
        log.Fatal(err)
    }

    if err := manager.Start(); err != nil {
        log.Fatal(err)
    }
    defer manager.Stop()

    // Handle discovery events
    go func() {
        for {
            select {
            case asset := <-manager.GetDiscoveredChannel():
                log.Printf("Discovered: %s (%s) at %s", 
                    asset.Name, asset.Type, asset.IPAddress)

            case asset := <-manager.GetUpdatedChannel():
                log.Printf("Updated: %s (%s)", asset.Name, asset.Type)

            case assetID := <-manager.GetLostChannel():
                log.Printf("Lost: %s", assetID)

            case queryEvent := <-manager.GetQueryChannel():
                log.Printf("Query from: %s", queryEvent.Sender)
            }
        }
    }()

    // Query for specific devices
    manager.QueryByType(discovery.AssetTypeSensor, discovery.AssetTypeSmartPlug)
    manager.QueryByRoom("living-room")
    manager.QueryByCapability(discovery.CapabilityTemperature)

    // Get current assets
    sensors := manager.GetAssetsByType(discovery.AssetTypeSensor)
    smartPlugs := manager.GetAssetsByCapability(discovery.CapabilityPower)
    livingRoomDevices := manager.GetAssetsByRoom("living-room")

    // Get statistics
    stats := manager.GetStats()
    log.Printf("Total assets: %d", stats.TotalAssets)
}
```

## 🏠 **Home Automation Integration**

### **Asset Types for Home Automation**

- **`gateway`** - Main home automation controller
- **`sensor`** - Multi-purpose sensors (Pi Pico with multiple sensors)
- **`smart_plug`** - TP-Link Tapo smart plugs
- **`thermostat`** - HVAC controllers
- **`camera`** - Security cameras
- **`light_bulb`** - Smart lighting
- **`motion_sensor`** - PIR motion detectors
- **`temperature_sensor`** - Temperature monitors
- **`humidity_sensor`** - Humidity monitors
- **`light_sensor`** - Ambient light sensors

### **Common Capabilities**

- **`temperature`** - Temperature sensing
- **`humidity`** - Humidity sensing  
- **`motion`** - Motion detection
- **`light`** - Light level sensing
- **`power`** - Power control/switching
- **`energy_monitor`** - Energy consumption monitoring
- **`thermostat_control`** - HVAC control
- **`mqtt`** - MQTT communication
- **`http`** - HTTP API
- **`klap`** - TP-Link KLAP protocol

### **Predefined Asset Builders**

```go
// Home automation gateway
gateway := discovery.NewHomeAutomationGateway("Main Gateway")

// Tapo smart plug
smartPlug := discovery.NewTapoSmartPlug("Living Room Lamp", "192.168.1.101", "P110")

// Pi Pico sensor with multiple capabilities
picoSensor := discovery.NewPicoSensor("Living Room Sensor", "living-room", 
    []discovery.AssetCapability{
        discovery.CapabilityTemperature,
        discovery.CapabilityHumidity,
        discovery.CapabilityMotion,
        discovery.CapabilityLight,
    })

// Thermostat
thermostat := discovery.NewThermostat("Main Thermostat", "living-room")

// Security camera
camera := discovery.NewCamera("Front Door Camera", "entrance", "192.168.1.102")
```

## 📊 **Discovery Events and Statistics**

### **Event Types**
- **Discovered** - New asset found
- **Updated** - Asset information changed
- **Lost** - Asset no longer responding
- **Query** - Discovery query received

### **Statistics Tracking**
- Total assets discovered
- Assets by type, room, and status
- Event counts by type
- Discovery event log with timestamps

## 🔧 **Configuration**

### **Network Configuration**
```go
const (
    DefaultMulticastAddress = "239.255.42.42"
    DefaultMulticastPort    = 42424
    DefaultTTL              = 300  // 5 minutes
    MaxMessageSize          = 8192
)
```

### **Discovery Manager Configuration**
```go
type DiscoveryConfig struct {
    LocalAsset    *AssetInfo    // This device's information
    AutoQuery     bool          // Send periodic queries
    QueryInterval time.Duration // How often to query
    MaxLogSize    int           // Max events in log
    Logger        *log.Logger   // Event logger
}
```

## 🔍 **Examples**

### **1. Find All Gateway Devices**
```bash
./discovery -mode=query -query-types="gateway" -duration=10s -verbose
```

### **2. Announce a Tapo Smart Plug**
```bash
./discovery -mode=announce \
    -type=smart_plug \
    -name="Living Room Outlet" \
    -room="living-room" \
    -ip="192.168.1.101" \
    -capabilities="power,energy_monitor,switch,http,klap" \
    -duration=300s
```

### **3. Query for Temperature Sensors**
```bash
./discovery -mode=query \
    -query-caps="temperature" \
    -duration=30s \
    -json
```

### **4. Find All Devices in Kitchen**
```bash
./discovery -mode=query \
    -room="kitchen" \
    -duration=30s \
    -verbose
```

### **5. Complete Discovery Workflow**
```bash
# Terminal 1: Start announcing your gateway
./discovery -mode=announce -type=gateway -name="Main Gateway" -room="office" -duration=600s &

# Terminal 2: Query for all devices
./discovery -mode=query -duration=15s -verbose

# Terminal 3: Monitor for new device announcements (passive)
./discovery -mode=discover -duration=300s -verbose
```

### **6. Automated Network Mapping**
```bash
#!/bin/bash
echo "🗺️ Mapping home automation network..."

# Query for all device types
for device_type in gateway sensor smart_plug thermostat camera; do
    echo "🔍 Finding ${device_type} devices..."
    ./discovery -mode=query -query-types="$device_type" -duration=5s -json > "${device_type}_devices.json"
done

echo "✅ Network mapping complete!"
```

## 🛡️ **Security Considerations**

- **Multicast Traffic**: Can be intercepted on the local network
- **Asset Information**: Contains potentially sensitive device details
- **Network Scanning**: Could reveal network topology
- **Firewall Rules**: May need multicast traffic allowance

**Recommendations:**
- Use on trusted networks only
- Implement authentication for sensitive operations
- Consider VPN for remote access
- Monitor for unauthorized discovery traffic

## 🔮 **Future Enhancements**

- **Authentication**: Signed discovery messages
- **Encryption**: TLS for sensitive asset information
- **Discovery Zones**: Hierarchical discovery domains
- **Service Discovery**: DNS-SD integration
- **Cloud Discovery**: Hybrid local/cloud discovery
- **Discovery Mesh**: Multi-hop discovery routing

## 📚 **API Reference**

### **Key Functions**

```go
// Create discovery manager
func NewDiscoveryManager(config DiscoveryConfig) (*DiscoveryManager, error)

// Asset builders
func NewAssetBuilder() *AssetBuilder
func NewHomeAutomationGateway(name string) *AssetBuilder
func NewTapoSmartPlug(name, ip, model string) *AssetBuilder
func NewPicoSensor(name, room string, capabilities []AssetCapability) *AssetBuilder

// Discovery operations
func (dm *DiscoveryManager) Query(query *Query) error
func (dm *DiscoveryManager) QueryByType(assetTypes ...AssetType) error
func (dm *DiscoveryManager) QueryByCapability(capabilities ...AssetCapability) error
func (dm *DiscoveryManager) QueryByRoom(room string) error

// Asset retrieval
func (dm *DiscoveryManager) GetAllAssets() map[string]*AssetInfo
func (dm *DiscoveryManager) GetAssetsByType(assetType AssetType) []*AssetInfo
func (dm *DiscoveryManager) GetAssetsByRoom(room string) []*AssetInfo
func (dm *DiscoveryManager) GetAssetsByCapability(capability AssetCapability) []*AssetInfo
```

The Asset Discovery Protocol provides a robust foundation for automatic device discovery and network asset management in home automation systems! 🏠🔍
