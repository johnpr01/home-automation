package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// AssetType represents the type of asset being discovered
type AssetType string

const (
	AssetTypeSensor         AssetType = "sensor"
	AssetTypeController     AssetType = "controller"
	AssetTypeSmartPlug      AssetType = "smart_plug"
	AssetTypeThermostat     AssetType = "thermostat"
	AssetTypeGateway        AssetType = "gateway"
	AssetTypeBridge         AssetType = "bridge"
	AssetTypeCamera         AssetType = "camera"
	AssetTypeLightBulb      AssetType = "light_bulb"
	AssetTypeMotionSensor   AssetType = "motion_sensor"
	AssetTypeTempSensor     AssetType = "temperature_sensor"
	AssetTypeHumiditySensor AssetType = "humidity_sensor"
	AssetTypeLightSensor    AssetType = "light_sensor"
)

// AssetCapability represents what an asset can do
type AssetCapability string

const (
	CapabilityTemperature    AssetCapability = "temperature"
	CapabilityHumidity       AssetCapability = "humidity"
	CapabilityMotion         AssetCapability = "motion"
	CapabilityLight          AssetCapability = "light"
	CapabilityPower          AssetCapability = "power"
	CapabilityEnergyMonitor  AssetCapability = "energy_monitor"
	CapabilitySwitch         AssetCapability = "switch"
	CapabilityDimmer         AssetCapability = "dimmer"
	CapabilityThermostatCtrl AssetCapability = "thermostat_control"
	CapabilityVideo          AssetCapability = "video"
	CapabilityAudio          AssetCapability = "audio"
	CapabilityMQTT           AssetCapability = "mqtt"
	CapabilityHTTP           AssetCapability = "http"
	CapabilityKLAP           AssetCapability = "klap"
)

// AssetInfo represents information about a discovered asset
type AssetInfo struct {
	// Basic Identity
	ID           string    `json:"id"`           // Unique asset identifier
	Name         string    `json:"name"`         // Human-readable name
	Type         AssetType `json:"type"`         // Type of asset
	Model        string    `json:"model"`        // Device model
	Manufacturer string    `json:"manufacturer"` // Device manufacturer
	Version      string    `json:"version"`      // Firmware/software version

	// Network Information
	IPAddress  string `json:"ip_address"`  // Primary IP address
	MACAddress string `json:"mac_address"` // MAC address
	Hostname   string `json:"hostname"`    // Network hostname
	Ports      []int  `json:"ports"`       // Open/service ports

	// Capabilities and Services
	Capabilities []AssetCapability `json:"capabilities"` // What the asset can do
	Services     []ServiceInfo     `json:"services"`     // Available services

	// Location and Organization
	Room     string            `json:"room"`     // Physical room/location
	Zone     string            `json:"zone"`     // Logical zone
	Tags     []string          `json:"tags"`     // Custom tags
	Metadata map[string]string `json:"metadata"` // Additional metadata

	// Status and Health
	Status       string    `json:"status"`        // online, offline, error
	LastSeen     time.Time `json:"last_seen"`     // Last discovery time
	Health       string    `json:"health"`        // healthy, warning, critical
	BatteryLevel *int      `json:"battery_level"` // Battery percentage (if applicable)

	// Discovery Protocol
	DiscoveryVersion string `json:"discovery_version"` // Protocol version
	Sequence         uint64 `json:"sequence"`          // Message sequence number
	TTL              int    `json:"ttl"`               // Time to live in seconds
}

// ServiceInfo represents a service offered by an asset
type ServiceInfo struct {
	Name        string            `json:"name"`        // Service name
	Protocol    string            `json:"protocol"`    // http, mqtt, tcp, udp
	Port        int               `json:"port"`        // Service port
	Path        string            `json:"path"`        // URL path (for HTTP)
	Topic       string            `json:"topic"`       // MQTT topic (for MQTT)
	Description string            `json:"description"` // Service description
	Properties  map[string]string `json:"properties"`  // Service-specific properties
}

// DiscoveryMessage represents a multicast discovery message
type DiscoveryMessage struct {
	Type      string     `json:"type"`      // announce, query, response, goodbye
	Asset     *AssetInfo `json:"asset"`     // Asset information
	Query     *Query     `json:"query"`     // Query parameters (for query messages)
	Timestamp time.Time  `json:"timestamp"` // Message timestamp
	Sender    string     `json:"sender"`    // Sender identifier
}

// Query represents a discovery query
type Query struct {
	AssetTypes   []AssetType       `json:"asset_types"`  // Filter by asset types
	Capabilities []AssetCapability `json:"capabilities"` // Filter by capabilities
	Room         string            `json:"room"`         // Filter by room
	Zone         string            `json:"zone"`         // Filter by zone
	Tags         []string          `json:"tags"`         // Filter by tags
	MaxAge       time.Duration     `json:"max_age"`      // Maximum age of responses
}

// Protocol constants
const (
	DefaultMulticastAddress = "239.255.42.42"
	DefaultMulticastPort    = 42424
	DefaultTTL              = 300 // 5 minutes
	DiscoveryVersion        = "1.0"
	MaxMessageSize          = 8192
)

// Message types
const (
	MessageTypeAnnounce = "announce"
	MessageTypeQuery    = "query"
	MessageTypeResponse = "response"
	MessageTypeGoodbye  = "goodbye"
)

// DiscoveryProtocol handles asset discovery using multicast
type DiscoveryProtocol struct {
	receiveConn   *net.UDPConn
	sendConn      *net.UDPConn
	multicastAddr *net.UDPAddr
	localAsset    *AssetInfo
	knownAssets   map[string]*AssetInfo
	listeners     []AssetDiscoveryListener
	ctx           context.Context
	cancel        context.CancelFunc
	sequence      uint64
}

// AssetDiscoveryListener handles discovery events
type AssetDiscoveryListener interface {
	OnAssetDiscovered(asset *AssetInfo)
	OnAssetUpdated(asset *AssetInfo)
	OnAssetLost(assetID string)
	OnQueryReceived(query *Query, sender string)
}

// NewDiscoveryProtocol creates a new discovery protocol instance
func NewDiscoveryProtocol(localAsset *AssetInfo) (*DiscoveryProtocol, error) {
	multicastAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", DefaultMulticastAddress, DefaultMulticastPort))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve multicast address: %w", err)
	}

	// Create receive connection (multicast listener)
	receiveConn, err := net.ListenMulticastUDP("udp", nil, multicastAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on multicast address: %w", err)
	}

	// Create send connection (regular UDP)
	sendConn, err := net.DialUDP("udp", nil, multicastAddr)
	if err != nil {
		receiveConn.Close()
		return nil, fmt.Errorf("failed to create send connection: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Initialize local asset with discovery metadata
	if localAsset != nil {
		localAsset.DiscoveryVersion = DiscoveryVersion
		localAsset.LastSeen = time.Now()
		if localAsset.TTL == 0 {
			localAsset.TTL = DefaultTTL
		}
	}

	return &DiscoveryProtocol{
		receiveConn:   receiveConn,
		sendConn:      sendConn,
		multicastAddr: multicastAddr,
		localAsset:    localAsset,
		knownAssets:   make(map[string]*AssetInfo),
		listeners:     make([]AssetDiscoveryListener, 0),
		ctx:           ctx,
		cancel:        cancel,
		sequence:      0,
	}, nil
}

// AddListener adds a discovery event listener
func (dp *DiscoveryProtocol) AddListener(listener AssetDiscoveryListener) {
	dp.listeners = append(dp.listeners, listener)
}

// Start begins the discovery protocol
func (dp *DiscoveryProtocol) Start() error {
	// Start listening for messages
	go dp.messageListener()

	// Start periodic announcements
	go dp.periodicAnnounce()

	// Start asset cleanup (remove stale assets)
	go dp.assetCleanup()

	// Send initial announcement
	if dp.localAsset != nil {
		return dp.Announce()
	}

	return nil
}

// Stop stops the discovery protocol
func (dp *DiscoveryProtocol) Stop() error {
	// Send goodbye message
	if dp.localAsset != nil {
		dp.Goodbye()
	}

	// Cancel context and close connections
	dp.cancel()
	dp.receiveConn.Close()
	return dp.sendConn.Close()
}

// Announce sends an announcement message for the local asset
func (dp *DiscoveryProtocol) Announce() error {
	if dp.localAsset == nil {
		return fmt.Errorf("no local asset configured")
	}

	dp.sequence++
	dp.localAsset.Sequence = dp.sequence
	dp.localAsset.LastSeen = time.Now()

	message := &DiscoveryMessage{
		Type:      MessageTypeAnnounce,
		Asset:     dp.localAsset,
		Timestamp: time.Now(),
		Sender:    dp.localAsset.ID,
	}

	return dp.sendMessage(message)
}

// Query sends a discovery query
func (dp *DiscoveryProtocol) Query(query *Query) error {
	dp.sequence++

	message := &DiscoveryMessage{
		Type:      MessageTypeQuery,
		Query:     query,
		Timestamp: time.Now(),
		Sender:    dp.getLocalID(),
	}

	return dp.sendMessage(message)
}

// Goodbye sends a goodbye message when leaving the network
func (dp *DiscoveryProtocol) Goodbye() error {
	if dp.localAsset == nil {
		return fmt.Errorf("no local asset configured")
	}

	dp.sequence++
	dp.localAsset.Sequence = dp.sequence

	message := &DiscoveryMessage{
		Type:      MessageTypeGoodbye,
		Asset:     dp.localAsset,
		Timestamp: time.Now(),
		Sender:    dp.localAsset.ID,
	}

	return dp.sendMessage(message)
}

// GetKnownAssets returns all currently known assets
func (dp *DiscoveryProtocol) GetKnownAssets() map[string]*AssetInfo {
	result := make(map[string]*AssetInfo)
	for id, asset := range dp.knownAssets {
		result[id] = asset
	}
	return result
}

// GetAssetsByType returns assets filtered by type
func (dp *DiscoveryProtocol) GetAssetsByType(assetType AssetType) []*AssetInfo {
	var assets []*AssetInfo
	for _, asset := range dp.knownAssets {
		if asset.Type == assetType {
			assets = append(assets, asset)
		}
	}
	return assets
}

// GetAssetsByCapability returns assets with specific capability
func (dp *DiscoveryProtocol) GetAssetsByCapability(capability AssetCapability) []*AssetInfo {
	var assets []*AssetInfo
	for _, asset := range dp.knownAssets {
		for _, cap := range asset.Capabilities {
			if cap == capability {
				assets = append(assets, asset)
				break
			}
		}
	}
	return assets
}

// GetAssetsByRoom returns assets in a specific room
func (dp *DiscoveryProtocol) GetAssetsByRoom(room string) []*AssetInfo {
	var assets []*AssetInfo
	for _, asset := range dp.knownAssets {
		if asset.Room == room {
			assets = append(assets, asset)
		}
	}
	return assets
}

// sendMessage sends a discovery message via multicast
func (dp *DiscoveryProtocol) sendMessage(message *DiscoveryMessage) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if len(data) > MaxMessageSize {
		return fmt.Errorf("message too large: %d bytes (max %d)", len(data), MaxMessageSize)
	}

	_, err = dp.sendConn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send multicast message: %w", err)
	}

	return nil
}

// messageListener listens for incoming discovery messages
func (dp *DiscoveryProtocol) messageListener() {
	buffer := make([]byte, MaxMessageSize)

	for {
		select {
		case <-dp.ctx.Done():
			return
		default:
			dp.receiveConn.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, addr, err := dp.receiveConn.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue // Timeout is expected
				}
				continue // Other errors, keep listening
			}

			dp.handleMessage(buffer[:n], addr)
		}
	}
}

// handleMessage processes an incoming discovery message
func (dp *DiscoveryProtocol) handleMessage(data []byte, sender *net.UDPAddr) {
	var message DiscoveryMessage
	if err := json.Unmarshal(data, &message); err != nil {
		return // Invalid message, ignore
	}

	// Ignore our own messages
	if message.Sender == dp.getLocalID() {
		return
	}

	switch message.Type {
	case MessageTypeAnnounce:
		dp.handleAnnounce(message.Asset)
	case MessageTypeQuery:
		dp.handleQuery(message.Query, message.Sender)
	case MessageTypeResponse:
		dp.handleResponse(message.Asset)
	case MessageTypeGoodbye:
		dp.handleGoodbye(message.Asset)
	}
}

// handleAnnounce processes an asset announcement
func (dp *DiscoveryProtocol) handleAnnounce(asset *AssetInfo) {
	if asset == nil || asset.ID == "" {
		return
	}

	existing, exists := dp.knownAssets[asset.ID]
	asset.LastSeen = time.Now()
	dp.knownAssets[asset.ID] = asset

	if exists {
		// Asset updated
		if existing.Sequence < asset.Sequence {
			for _, listener := range dp.listeners {
				listener.OnAssetUpdated(asset)
			}
		}
	} else {
		// New asset discovered
		for _, listener := range dp.listeners {
			listener.OnAssetDiscovered(asset)
		}
	}
}

// handleQuery processes a discovery query
func (dp *DiscoveryProtocol) handleQuery(query *Query, sender string) {
	// Notify listeners about the query
	for _, listener := range dp.listeners {
		listener.OnQueryReceived(query, sender)
	}

	// Respond if our local asset matches the query
	if dp.localAsset != nil && dp.matchesQuery(dp.localAsset, query) {
		response := &DiscoveryMessage{
			Type:      MessageTypeResponse,
			Asset:     dp.localAsset,
			Timestamp: time.Now(),
			Sender:    dp.localAsset.ID,
		}
		dp.sendMessage(response)
	}
}

// handleResponse processes a query response
func (dp *DiscoveryProtocol) handleResponse(asset *AssetInfo) {
	dp.handleAnnounce(asset) // Same logic as announcement
}

// handleGoodbye processes a goodbye message
func (dp *DiscoveryProtocol) handleGoodbye(asset *AssetInfo) {
	if asset == nil || asset.ID == "" {
		return
	}

	if _, exists := dp.knownAssets[asset.ID]; exists {
		delete(dp.knownAssets, asset.ID)
		for _, listener := range dp.listeners {
			listener.OnAssetLost(asset.ID)
		}
	}
}

// matchesQuery checks if an asset matches a discovery query
func (dp *DiscoveryProtocol) matchesQuery(asset *AssetInfo, query *Query) bool {
	if query == nil {
		return true
	}

	// Check asset types
	if len(query.AssetTypes) > 0 {
		found := false
		for _, t := range query.AssetTypes {
			if asset.Type == t {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check capabilities
	if len(query.Capabilities) > 0 {
		for _, queryCap := range query.Capabilities {
			found := false
			for _, assetCap := range asset.Capabilities {
				if assetCap == queryCap {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	// Check room
	if query.Room != "" && asset.Room != query.Room {
		return false
	}

	// Check zone
	if query.Zone != "" && asset.Zone != query.Zone {
		return false
	}

	// Check tags
	if len(query.Tags) > 0 {
		for _, queryTag := range query.Tags {
			found := false
			for _, assetTag := range asset.Tags {
				if assetTag == queryTag {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	// Check age
	if query.MaxAge > 0 {
		age := time.Since(asset.LastSeen)
		if age > query.MaxAge {
			return false
		}
	}

	return true
}

// periodicAnnounce sends periodic announcements
func (dp *DiscoveryProtocol) periodicAnnounce() {
	if dp.localAsset == nil {
		return
	}

	ticker := time.NewTicker(time.Duration(dp.localAsset.TTL/3) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-dp.ctx.Done():
			return
		case <-ticker.C:
			dp.Announce()
		}
	}
}

// assetCleanup removes stale assets
func (dp *DiscoveryProtocol) assetCleanup() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-dp.ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			for id, asset := range dp.knownAssets {
				ttl := time.Duration(asset.TTL) * time.Second
				if now.Sub(asset.LastSeen) > ttl*2 { // Double TTL for cleanup
					delete(dp.knownAssets, id)
					for _, listener := range dp.listeners {
						listener.OnAssetLost(id)
					}
				}
			}
		}
	}
}

// getLocalID returns the local asset ID or a default
func (dp *DiscoveryProtocol) getLocalID() string {
	if dp.localAsset != nil {
		return dp.localAsset.ID
	}
	return "unknown"
}
