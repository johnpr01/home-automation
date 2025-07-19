package discovery

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// DiscoveryManager manages asset discovery for the home automation system
type DiscoveryManager struct {
	protocol    *DiscoveryProtocol
	assets      map[string]*AssetInfo
	assetsMutex sync.RWMutex
	eventLog    []DiscoveryEvent
	logMutex    sync.RWMutex
	maxLogSize  int
	logger      *log.Logger

	// Event channels
	discoveredCh chan *AssetInfo
	updatedCh    chan *AssetInfo
	lostCh       chan string
	queryCh      chan QueryEvent

	// Configuration
	autoQuery     bool
	queryInterval time.Duration
}

// DiscoveryEvent represents a discovery event
type DiscoveryEvent struct {
	Type      string     `json:"type"`      // discovered, updated, lost, query
	AssetID   string     `json:"asset_id"`  // Asset ID
	Asset     *AssetInfo `json:"asset"`     // Asset information (for discovered/updated)
	Query     *Query     `json:"query"`     // Query information (for query events)
	Sender    string     `json:"sender"`    // Query sender (for query events)
	Timestamp time.Time  `json:"timestamp"` // Event timestamp
	Message   string     `json:"message"`   // Human-readable message
}

// QueryEvent represents a received query
type QueryEvent struct {
	Query     *Query
	Sender    string
	Timestamp time.Time
}

// DiscoveryConfig holds configuration for the discovery manager
type DiscoveryConfig struct {
	LocalAsset    *AssetInfo    // Local asset information
	AutoQuery     bool          // Automatically send periodic queries
	QueryInterval time.Duration // Query interval for auto-query
	MaxLogSize    int           // Maximum number of events to keep in log
	Logger        *log.Logger   // Logger for discovery events
}

// NewDiscoveryManager creates a new discovery manager
func NewDiscoveryManager(config DiscoveryConfig) (*DiscoveryManager, error) {
	protocol, err := NewDiscoveryProtocol(config.LocalAsset)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery protocol: %w", err)
	}

	if config.QueryInterval == 0 {
		config.QueryInterval = 5 * time.Minute
	}
	if config.MaxLogSize == 0 {
		config.MaxLogSize = 1000
	}

	dm := &DiscoveryManager{
		protocol:      protocol,
		assets:        make(map[string]*AssetInfo),
		eventLog:      make([]DiscoveryEvent, 0),
		maxLogSize:    config.MaxLogSize,
		logger:        config.Logger,
		discoveredCh:  make(chan *AssetInfo, 100),
		updatedCh:     make(chan *AssetInfo, 100),
		lostCh:        make(chan string, 100),
		queryCh:       make(chan QueryEvent, 100),
		autoQuery:     config.AutoQuery,
		queryInterval: config.QueryInterval,
	}

	// Add ourselves as a listener
	protocol.AddListener(dm)

	return dm, nil
}

// Start starts the discovery manager
func (dm *DiscoveryManager) Start() error {
	if err := dm.protocol.Start(); err != nil {
		return fmt.Errorf("failed to start discovery protocol: %w", err)
	}

	// Start auto-query if enabled
	if dm.autoQuery {
		go dm.autoQueryLoop()
	}

	dm.logEvent("system", "", nil, nil, "", "Discovery manager started")
	return nil
}

// Stop stops the discovery manager
func (dm *DiscoveryManager) Stop() error {
	dm.logEvent("system", "", nil, nil, "", "Discovery manager stopping")
	return dm.protocol.Stop()
}

// GetAllAssets returns all discovered assets
func (dm *DiscoveryManager) GetAllAssets() map[string]*AssetInfo {
	dm.assetsMutex.RLock()
	defer dm.assetsMutex.RUnlock()

	result := make(map[string]*AssetInfo)
	for id, asset := range dm.assets {
		result[id] = asset
	}
	return result
}

// GetAsset returns a specific asset by ID
func (dm *DiscoveryManager) GetAsset(id string) (*AssetInfo, bool) {
	dm.assetsMutex.RLock()
	defer dm.assetsMutex.RUnlock()

	asset, exists := dm.assets[id]
	return asset, exists
}

// GetAssetsByType returns assets of a specific type
func (dm *DiscoveryManager) GetAssetsByType(assetType AssetType) []*AssetInfo {
	dm.assetsMutex.RLock()
	defer dm.assetsMutex.RUnlock()

	var assets []*AssetInfo
	for _, asset := range dm.assets {
		if asset.Type == assetType {
			assets = append(assets, asset)
		}
	}
	return assets
}

// GetAssetsByRoom returns assets in a specific room
func (dm *DiscoveryManager) GetAssetsByRoom(room string) []*AssetInfo {
	dm.assetsMutex.RLock()
	defer dm.assetsMutex.RUnlock()

	var assets []*AssetInfo
	for _, asset := range dm.assets {
		if asset.Room == room {
			assets = append(assets, asset)
		}
	}
	return assets
}

// GetAssetsByCapability returns assets with a specific capability
func (dm *DiscoveryManager) GetAssetsByCapability(capability AssetCapability) []*AssetInfo {
	dm.assetsMutex.RLock()
	defer dm.assetsMutex.RUnlock()

	var assets []*AssetInfo
	for _, asset := range dm.assets {
		for _, cap := range asset.Capabilities {
			if cap == capability {
				assets = append(assets, asset)
				break
			}
		}
	}
	return assets
}

// Query sends a discovery query
func (dm *DiscoveryManager) Query(query *Query) error {
	dm.logEvent("query", "", nil, query, "", "Sending discovery query")
	return dm.protocol.Query(query)
}

// QueryByType queries for assets of specific types
func (dm *DiscoveryManager) QueryByType(assetTypes ...AssetType) error {
	query := &Query{
		AssetTypes: assetTypes,
		MaxAge:     10 * time.Minute,
	}
	return dm.Query(query)
}

// QueryByCapability queries for assets with specific capabilities
func (dm *DiscoveryManager) QueryByCapability(capabilities ...AssetCapability) error {
	query := &Query{
		Capabilities: capabilities,
		MaxAge:       10 * time.Minute,
	}
	return dm.Query(query)
}

// QueryByRoom queries for assets in a specific room
func (dm *DiscoveryManager) QueryByRoom(room string) error {
	query := &Query{
		Room:   room,
		MaxAge: 10 * time.Minute,
	}
	return dm.Query(query)
}

// Announce sends an announcement for the local asset
func (dm *DiscoveryManager) Announce() error {
	dm.logEvent("announce", "", nil, nil, "", "Sending asset announcement")
	return dm.protocol.Announce()
}

// GetDiscoveredChannel returns the channel for discovered assets
func (dm *DiscoveryManager) GetDiscoveredChannel() <-chan *AssetInfo {
	return dm.discoveredCh
}

// GetUpdatedChannel returns the channel for updated assets
func (dm *DiscoveryManager) GetUpdatedChannel() <-chan *AssetInfo {
	return dm.updatedCh
}

// GetLostChannel returns the channel for lost assets
func (dm *DiscoveryManager) GetLostChannel() <-chan string {
	return dm.lostCh
}

// GetQueryChannel returns the channel for received queries
func (dm *DiscoveryManager) GetQueryChannel() <-chan QueryEvent {
	return dm.queryCh
}

// GetEventLog returns the discovery event log
func (dm *DiscoveryManager) GetEventLog() []DiscoveryEvent {
	dm.logMutex.RLock()
	defer dm.logMutex.RUnlock()

	result := make([]DiscoveryEvent, len(dm.eventLog))
	copy(result, dm.eventLog)
	return result
}

// GetEventLogSince returns events since a specific time
func (dm *DiscoveryManager) GetEventLogSince(since time.Time) []DiscoveryEvent {
	dm.logMutex.RLock()
	defer dm.logMutex.RUnlock()

	var result []DiscoveryEvent
	for _, event := range dm.eventLog {
		if event.Timestamp.After(since) {
			result = append(result, event)
		}
	}
	return result
}

// GetStats returns discovery statistics
func (dm *DiscoveryManager) GetStats() DiscoveryStats {
	dm.assetsMutex.RLock()
	dm.logMutex.RLock()
	defer dm.assetsMutex.RUnlock()
	defer dm.logMutex.RUnlock()

	stats := DiscoveryStats{
		TotalAssets:    len(dm.assets),
		AssetsByType:   make(map[AssetType]int),
		AssetsByRoom:   make(map[string]int),
		AssetsByStatus: make(map[string]int),
		TotalEvents:    len(dm.eventLog),
	}

	// Count assets by type, room, and status
	for _, asset := range dm.assets {
		stats.AssetsByType[asset.Type]++
		if asset.Room != "" {
			stats.AssetsByRoom[asset.Room]++
		}
		stats.AssetsByStatus[asset.Status]++
	}

	// Count events by type
	stats.EventsByType = make(map[string]int)
	for _, event := range dm.eventLog {
		stats.EventsByType[event.Type]++
	}

	return stats
}

// DiscoveryStats holds discovery statistics
type DiscoveryStats struct {
	TotalAssets    int               `json:"total_assets"`
	AssetsByType   map[AssetType]int `json:"assets_by_type"`
	AssetsByRoom   map[string]int    `json:"assets_by_room"`
	AssetsByStatus map[string]int    `json:"assets_by_status"`
	TotalEvents    int               `json:"total_events"`
	EventsByType   map[string]int    `json:"events_by_type"`
}

// Implement AssetDiscoveryListener interface

// OnAssetDiscovered handles asset discovery events
func (dm *DiscoveryManager) OnAssetDiscovered(asset *AssetInfo) {
	dm.assetsMutex.Lock()
	dm.assets[asset.ID] = asset
	dm.assetsMutex.Unlock()

	message := fmt.Sprintf("Discovered %s: %s (%s)", asset.Type, asset.Name, asset.IPAddress)
	dm.logEvent("discovered", asset.ID, asset, nil, "", message)

	// Send to channel (non-blocking)
	select {
	case dm.discoveredCh <- asset:
	default:
		// Channel full, skip
	}
}

// OnAssetUpdated handles asset update events
func (dm *DiscoveryManager) OnAssetUpdated(asset *AssetInfo) {
	dm.assetsMutex.Lock()
	dm.assets[asset.ID] = asset
	dm.assetsMutex.Unlock()

	message := fmt.Sprintf("Updated %s: %s", asset.Type, asset.Name)
	dm.logEvent("updated", asset.ID, asset, nil, "", message)

	// Send to channel (non-blocking)
	select {
	case dm.updatedCh <- asset:
	default:
		// Channel full, skip
	}
}

// OnAssetLost handles asset lost events
func (dm *DiscoveryManager) OnAssetLost(assetID string) {
	var assetName string

	dm.assetsMutex.Lock()
	if asset, exists := dm.assets[assetID]; exists {
		assetName = fmt.Sprintf("%s (%s)", asset.Name, asset.Type)
		delete(dm.assets, assetID)
	}
	dm.assetsMutex.Unlock()

	message := fmt.Sprintf("Lost asset: %s", assetName)
	if assetName == "" {
		message = fmt.Sprintf("Lost asset: %s", assetID)
	}

	dm.logEvent("lost", assetID, nil, nil, "", message)

	// Send to channel (non-blocking)
	select {
	case dm.lostCh <- assetID:
	default:
		// Channel full, skip
	}
}

// OnQueryReceived handles query events
func (dm *DiscoveryManager) OnQueryReceived(query *Query, sender string) {
	message := fmt.Sprintf("Received query from %s", sender)
	dm.logEvent("query", "", nil, query, sender, message)

	// Send to channel (non-blocking)
	select {
	case dm.queryCh <- QueryEvent{
		Query:     query,
		Sender:    sender,
		Timestamp: time.Now(),
	}:
	default:
		// Channel full, skip
	}
}

// logEvent adds an event to the discovery log
func (dm *DiscoveryManager) logEvent(eventType, assetID string, asset *AssetInfo, query *Query, sender, message string) {
	event := DiscoveryEvent{
		Type:      eventType,
		AssetID:   assetID,
		Asset:     asset,
		Query:     query,
		Sender:    sender,
		Timestamp: time.Now(),
		Message:   message,
	}

	dm.logMutex.Lock()
	dm.eventLog = append(dm.eventLog, event)

	// Trim log if it's too large
	if len(dm.eventLog) > dm.maxLogSize {
		dm.eventLog = dm.eventLog[len(dm.eventLog)-dm.maxLogSize:]
	}
	dm.logMutex.Unlock()

	// Log to system logger if available
	if dm.logger != nil {
		dm.logger.Printf("[DISCOVERY] %s", message)
	}
}

// autoQueryLoop runs automatic periodic queries
func (dm *DiscoveryManager) autoQueryLoop() {
	ticker := time.NewTicker(dm.queryInterval)
	defer ticker.Stop()

	for range ticker.C {
		// Send a general query for all asset types
		query := &Query{
			MaxAge: dm.queryInterval * 2,
		}
		dm.Query(query)
	}
}

// Helper functions for creating common queries

// CreateSensorQuery creates a query for sensor devices
func CreateSensorQuery(room string) *Query {
	return &Query{
		AssetTypes: []AssetType{
			AssetTypeSensor,
			AssetTypeMotionSensor,
			AssetTypeTempSensor,
			AssetTypeHumiditySensor,
			AssetTypeLightSensor,
		},
		Room:   room,
		MaxAge: 10 * time.Minute,
	}
}

// CreateSmartDeviceQuery creates a query for smart devices
func CreateSmartDeviceQuery(room string) *Query {
	return &Query{
		AssetTypes: []AssetType{
			AssetTypeSmartPlug,
			AssetTypeThermostat,
			AssetTypeLightBulb,
			AssetTypeCamera,
		},
		Room:   room,
		MaxAge: 10 * time.Minute,
	}
}

// CreateControllerQuery creates a query for controllers and gateways
func CreateControllerQuery() *Query {
	return &Query{
		AssetTypes: []AssetType{
			AssetTypeController,
			AssetTypeGateway,
			AssetTypeBridge,
		},
		MaxAge: 15 * time.Minute,
	}
}

// CreateCapabilityQuery creates a query for specific capabilities
func CreateCapabilityQuery(capabilities []AssetCapability, room string) *Query {
	return &Query{
		Capabilities: capabilities,
		Room:         room,
		MaxAge:       10 * time.Minute,
	}
}
