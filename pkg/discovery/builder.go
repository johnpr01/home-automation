package discovery

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"time"
)

// AssetBuilder helps create AssetInfo structures
type AssetBuilder struct {
	asset *AssetInfo
}

// NewAssetBuilder creates a new asset builder
func NewAssetBuilder() *AssetBuilder {
	return &AssetBuilder{
		asset: &AssetInfo{
			Capabilities:     make([]AssetCapability, 0),
			Services:         make([]ServiceInfo, 0),
			Tags:             make([]string, 0),
			Metadata:         make(map[string]string),
			Status:           "online",
			Health:           "healthy",
			DiscoveryVersion: DiscoveryVersion,
			TTL:              DefaultTTL,
			LastSeen:         time.Now(),
		},
	}
}

// WithID sets the asset ID
func (ab *AssetBuilder) WithID(id string) *AssetBuilder {
	ab.asset.ID = id
	return ab
}

// WithName sets the asset name
func (ab *AssetBuilder) WithName(name string) *AssetBuilder {
	ab.asset.Name = name
	return ab
}

// WithType sets the asset type
func (ab *AssetBuilder) WithType(assetType AssetType) *AssetBuilder {
	ab.asset.Type = assetType
	return ab
}

// WithModel sets the device model
func (ab *AssetBuilder) WithModel(model string) *AssetBuilder {
	ab.asset.Model = model
	return ab
}

// WithManufacturer sets the device manufacturer
func (ab *AssetBuilder) WithManufacturer(manufacturer string) *AssetBuilder {
	ab.asset.Manufacturer = manufacturer
	return ab
}

// WithVersion sets the firmware/software version
func (ab *AssetBuilder) WithVersion(version string) *AssetBuilder {
	ab.asset.Version = version
	return ab
}

// WithIPAddress sets the IP address
func (ab *AssetBuilder) WithIPAddress(ip string) *AssetBuilder {
	ab.asset.IPAddress = ip
	return ab
}

// WithMACAddress sets the MAC address
func (ab *AssetBuilder) WithMACAddress(mac string) *AssetBuilder {
	ab.asset.MACAddress = mac
	return ab
}

// WithHostname sets the hostname
func (ab *AssetBuilder) WithHostname(hostname string) *AssetBuilder {
	ab.asset.Hostname = hostname
	return ab
}

// WithPort adds a port to the asset
func (ab *AssetBuilder) WithPort(port int) *AssetBuilder {
	ab.asset.Ports = append(ab.asset.Ports, port)
	return ab
}

// WithPorts sets multiple ports
func (ab *AssetBuilder) WithPorts(ports []int) *AssetBuilder {
	ab.asset.Ports = ports
	return ab
}

// WithCapability adds a capability
func (ab *AssetBuilder) WithCapability(capability AssetCapability) *AssetBuilder {
	ab.asset.Capabilities = append(ab.asset.Capabilities, capability)
	return ab
}

// WithCapabilities sets multiple capabilities
func (ab *AssetBuilder) WithCapabilities(capabilities []AssetCapability) *AssetBuilder {
	ab.asset.Capabilities = capabilities
	return ab
}

// WithService adds a service
func (ab *AssetBuilder) WithService(service ServiceInfo) *AssetBuilder {
	ab.asset.Services = append(ab.asset.Services, service)
	return ab
}

// WithHTTPService adds an HTTP service
func (ab *AssetBuilder) WithHTTPService(name string, port int, path string, description string) *AssetBuilder {
	service := ServiceInfo{
		Name:        name,
		Protocol:    "http",
		Port:        port,
		Path:        path,
		Description: description,
		Properties:  make(map[string]string),
	}
	return ab.WithService(service)
}

// WithMQTTService adds an MQTT service
func (ab *AssetBuilder) WithMQTTService(name string, topic string, description string) *AssetBuilder {
	service := ServiceInfo{
		Name:        name,
		Protocol:    "mqtt",
		Topic:       topic,
		Description: description,
		Properties:  make(map[string]string),
	}
	return ab.WithService(service)
}

// WithRoom sets the room/location
func (ab *AssetBuilder) WithRoom(room string) *AssetBuilder {
	ab.asset.Room = room
	return ab
}

// WithZone sets the logical zone
func (ab *AssetBuilder) WithZone(zone string) *AssetBuilder {
	ab.asset.Zone = zone
	return ab
}

// WithTag adds a tag
func (ab *AssetBuilder) WithTag(tag string) *AssetBuilder {
	ab.asset.Tags = append(ab.asset.Tags, tag)
	return ab
}

// WithTags sets multiple tags
func (ab *AssetBuilder) WithTags(tags []string) *AssetBuilder {
	ab.asset.Tags = tags
	return ab
}

// WithMetadata adds metadata
func (ab *AssetBuilder) WithMetadata(key, value string) *AssetBuilder {
	ab.asset.Metadata[key] = value
	return ab
}

// WithStatus sets the status
func (ab *AssetBuilder) WithStatus(status string) *AssetBuilder {
	ab.asset.Status = status
	return ab
}

// WithHealth sets the health status
func (ab *AssetBuilder) WithHealth(health string) *AssetBuilder {
	ab.asset.Health = health
	return ab
}

// WithBatteryLevel sets the battery level
func (ab *AssetBuilder) WithBatteryLevel(level int) *AssetBuilder {
	ab.asset.BatteryLevel = &level
	return ab
}

// WithTTL sets the time-to-live
func (ab *AssetBuilder) WithTTL(ttl int) *AssetBuilder {
	ab.asset.TTL = ttl
	return ab
}

// AutoDetectNetwork automatically detects network information
func (ab *AssetBuilder) AutoDetectNetwork() *AssetBuilder {
	// Get hostname
	if hostname, err := os.Hostname(); err == nil {
		ab.asset.Hostname = hostname
	}

	// Get IP address
	if ip := getLocalIP(); ip != "" {
		ab.asset.IPAddress = ip
	}

	// Get MAC address
	if mac := getMACAddress(); mac != "" {
		ab.asset.MACAddress = mac
	}

	return ab
}

// AutoDetectSystem automatically detects system information
func (ab *AssetBuilder) AutoDetectSystem() *AssetBuilder {
	ab.asset.Metadata["os"] = runtime.GOOS
	ab.asset.Metadata["arch"] = runtime.GOARCH
	ab.asset.Metadata["go_version"] = runtime.Version()

	return ab
}

// Build creates the final AssetInfo
func (ab *AssetBuilder) Build() *AssetInfo {
	// Generate ID if not set
	if ab.asset.ID == "" {
		ab.asset.ID = generateAssetID(ab.asset)
	}

	// Set sequence number
	ab.asset.Sequence = 1

	return ab.asset
}

// Predefined asset builders for common device types

// NewHomeAutomationGateway creates a gateway asset
func NewHomeAutomationGateway(name string) *AssetBuilder {
	return NewAssetBuilder().
		WithType(AssetTypeGateway).
		WithName(name).
		WithCapabilities([]AssetCapability{
			CapabilityMQTT,
			CapabilityHTTP,
		}).
		AutoDetectNetwork().
		AutoDetectSystem()
}

// NewTapoSmartPlug creates a Tapo smart plug asset
func NewTapoSmartPlug(name, ip, model string) *AssetBuilder {
	return NewAssetBuilder().
		WithType(AssetTypeSmartPlug).
		WithName(name).
		WithModel(model).
		WithManufacturer("TP-Link").
		WithIPAddress(ip).
		WithCapabilities([]AssetCapability{
			CapabilityPower,
			CapabilityEnergyMonitor,
			CapabilitySwitch,
			CapabilityHTTP,
			CapabilityKLAP,
		}).
		WithHTTPService("tapo-api", 80, "/", "Tapo device API").
		WithPort(80)
}

// NewPicoSensor creates a Pi Pico sensor asset
func NewPicoSensor(name, room string, capabilities []AssetCapability) *AssetBuilder {
	builder := NewAssetBuilder().
		WithType(AssetTypeSensor).
		WithName(name).
		WithModel("Raspberry Pi Pico WH").
		WithManufacturer("Raspberry Pi Foundation").
		WithRoom(room).
		WithCapabilities(capabilities).
		WithCapability(CapabilityMQTT)

	// Add MQTT services based on capabilities
	for _, cap := range capabilities {
		switch cap {
		case CapabilityTemperature:
			builder.WithMQTTService("temperature", fmt.Sprintf("room-temp/%s", room), "Temperature sensor")
		case CapabilityHumidity:
			builder.WithMQTTService("humidity", fmt.Sprintf("room-humidity/%s", room), "Humidity sensor")
		case CapabilityMotion:
			builder.WithMQTTService("motion", fmt.Sprintf("room-motion/%s", room), "Motion sensor")
		case CapabilityLight:
			builder.WithMQTTService("light", fmt.Sprintf("room-light/%s", room), "Light sensor")
		}
	}

	return builder
}

// NewThermostat creates a thermostat asset
func NewThermostat(name, room string) *AssetBuilder {
	return NewAssetBuilder().
		WithType(AssetTypeThermostat).
		WithName(name).
		WithRoom(room).
		WithCapabilities([]AssetCapability{
			CapabilityThermostatCtrl,
			CapabilityTemperature,
			CapabilityMQTT,
			CapabilityHTTP,
		}).
		WithMQTTService("thermostat-control", fmt.Sprintf("thermostat/%s", room), "Thermostat control").
		WithHTTPService("thermostat-api", 8080, "/api/thermostat", "Thermostat HTTP API")
}

// NewCamera creates a camera asset
func NewCamera(name, room, ip string) *AssetBuilder {
	return NewAssetBuilder().
		WithType(AssetTypeCamera).
		WithName(name).
		WithRoom(room).
		WithIPAddress(ip).
		WithCapabilities([]AssetCapability{
			CapabilityVideo,
			CapabilityHTTP,
		}).
		WithHTTPService("video-stream", 80, "/stream", "Video stream").
		WithHTTPService("camera-api", 80, "/api", "Camera control API").
		WithPort(80)
}

// NewLightBulb creates a smart light bulb asset
func NewLightBulb(name, room string) *AssetBuilder {
	return NewAssetBuilder().
		WithType(AssetTypeLightBulb).
		WithName(name).
		WithRoom(room).
		WithCapabilities([]AssetCapability{
			CapabilitySwitch,
			CapabilityDimmer,
			CapabilityMQTT,
		}).
		WithMQTTService("light-control", fmt.Sprintf("light/%s", room), "Light control")
}

// Utility functions

// getLocalIP gets the local IP address
func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// getMACAddress gets the MAC address of the primary network interface
func getMACAddress() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0 {
			return iface.HardwareAddr.String()
		}
	}

	return ""
}

// generateAssetID generates a unique asset ID based on asset properties
func generateAssetID(asset *AssetInfo) string {
	// Use MAC address if available
	if asset.MACAddress != "" {
		return strings.ReplaceAll(asset.MACAddress, ":", "")
	}

	// Use IP address and hostname
	if asset.IPAddress != "" && asset.Hostname != "" {
		return fmt.Sprintf("%s-%s",
			strings.ReplaceAll(asset.IPAddress, ".", "-"),
			asset.Hostname)
	}

	// Use type and name
	if asset.Name != "" {
		return fmt.Sprintf("%s-%s", asset.Type,
			strings.ReplaceAll(strings.ToLower(asset.Name), " ", "-"))
	}

	// Fallback to timestamp
	return fmt.Sprintf("asset-%d", time.Now().Unix())
}

// ParseAssetType parses a string to AssetType
func ParseAssetType(s string) (AssetType, error) {
	switch strings.ToLower(s) {
	case "sensor":
		return AssetTypeSensor, nil
	case "controller":
		return AssetTypeController, nil
	case "smart_plug", "smartplug":
		return AssetTypeSmartPlug, nil
	case "thermostat":
		return AssetTypeThermostat, nil
	case "gateway":
		return AssetTypeGateway, nil
	case "bridge":
		return AssetTypeBridge, nil
	case "camera":
		return AssetTypeCamera, nil
	case "light_bulb", "lightbulb":
		return AssetTypeLightBulb, nil
	case "motion_sensor", "motionsensor":
		return AssetTypeMotionSensor, nil
	case "temperature_sensor", "temperaturesensor":
		return AssetTypeTempSensor, nil
	case "humidity_sensor", "humiditysensor":
		return AssetTypeHumiditySensor, nil
	case "light_sensor", "lightsensor":
		return AssetTypeLightSensor, nil
	default:
		return "", fmt.Errorf("unknown asset type: %s", s)
	}
}

// ParseAssetCapability parses a string to AssetCapability
func ParseAssetCapability(s string) (AssetCapability, error) {
	switch strings.ToLower(s) {
	case "temperature":
		return CapabilityTemperature, nil
	case "humidity":
		return CapabilityHumidity, nil
	case "motion":
		return CapabilityMotion, nil
	case "light":
		return CapabilityLight, nil
	case "power":
		return CapabilityPower, nil
	case "energy_monitor", "energymonitor":
		return CapabilityEnergyMonitor, nil
	case "switch":
		return CapabilitySwitch, nil
	case "dimmer":
		return CapabilityDimmer, nil
	case "thermostat_control", "thermostatcontrol":
		return CapabilityThermostatCtrl, nil
	case "video":
		return CapabilityVideo, nil
	case "audio":
		return CapabilityAudio, nil
	case "mqtt":
		return CapabilityMQTT, nil
	case "http":
		return CapabilityHTTP, nil
	case "klap":
		return CapabilityKLAP, nil
	default:
		return "", fmt.Errorf("unknown capability: %s", s)
	}
}

// ValidateAssetInfo validates an AssetInfo structure
func ValidateAssetInfo(asset *AssetInfo) error {
	if asset == nil {
		return fmt.Errorf("asset cannot be nil")
	}

	if asset.ID == "" {
		return fmt.Errorf("asset ID is required")
	}

	if asset.Type == "" {
		return fmt.Errorf("asset type is required")
	}

	if asset.TTL <= 0 {
		return fmt.Errorf("TTL must be positive")
	}

	// Validate IP address if provided
	if asset.IPAddress != "" {
		if ip := net.ParseIP(asset.IPAddress); ip == nil {
			return fmt.Errorf("invalid IP address: %s", asset.IPAddress)
		}
	}

	// Validate MAC address format if provided
	if asset.MACAddress != "" {
		if _, err := net.ParseMAC(asset.MACAddress); err != nil {
			return fmt.Errorf("invalid MAC address: %s", asset.MACAddress)
		}
	}

	// Validate ports
	for _, port := range asset.Ports {
		if port < 1 || port > 65535 {
			return fmt.Errorf("invalid port: %d", port)
		}
	}

	// Validate services
	for i, service := range asset.Services {
		if service.Name == "" {
			return fmt.Errorf("service %d: name is required", i)
		}
		if service.Protocol == "" {
			return fmt.Errorf("service %d: protocol is required", i)
		}
		if service.Protocol == "http" || service.Protocol == "tcp" || service.Protocol == "udp" {
			if service.Port < 1 || service.Port > 65535 {
				return fmt.Errorf("service %d: invalid port %d", i, service.Port)
			}
		}
	}

	return nil
}
