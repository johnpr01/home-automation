package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/johnpr01/home-automation/pkg/discovery"
)

func main() {
	var (
		mode         = flag.String("mode", "discover", "Mode: discover, announce, query")
		assetType    = flag.String("type", "gateway", "Asset type for announce mode")
		assetName    = flag.String("name", "", "Asset name for announce mode")
		room         = flag.String("room", "", "Room for announce mode or query filter")
		ip           = flag.String("ip", "", "IP address for announce mode")
		capabilities = flag.String("capabilities", "", "Comma-separated capabilities for announce mode")
		queryTypes   = flag.String("query-types", "", "Comma-separated asset types to query for")
		queryCaps    = flag.String("query-caps", "", "Comma-separated capabilities to query for")
		duration     = flag.Duration("duration", 60*time.Second, "Duration to run discovery")
		verbose      = flag.Bool("verbose", false, "Verbose output")
		jsonOutput   = flag.Bool("json", false, "JSON output format")
	)
	flag.Parse()

	logger := log.New(os.Stdout, "[DISCOVERY] ", log.LstdFlags)

	switch *mode {
	case "discover":
		runDiscovery(*duration, *verbose, *jsonOutput, logger)
	case "announce":
		runAnnounce(*assetType, *assetName, *room, *ip, *capabilities, *duration, *verbose, logger)
	case "query":
		runQuery(*queryTypes, *queryCaps, *room, *duration, *verbose, *jsonOutput, logger)
	default:
		fmt.Printf("Unknown mode: %s\n", *mode)
		flag.Usage()
		os.Exit(1)
	}
}

// runDiscovery runs asset discovery and displays found assets
func runDiscovery(duration time.Duration, verbose, jsonOutput bool, logger *log.Logger) {
	fmt.Printf("🔍 Starting asset discovery for %v...\n\n", duration)

	// Create discovery manager
	config := discovery.DiscoveryConfig{
		AutoQuery:     true,
		QueryInterval: 30 * time.Second,
		Logger:        logger,
	}

	manager, err := discovery.NewDiscoveryManager(config)
	if err != nil {
		fmt.Printf("Error creating discovery manager: %v\n", err)
		os.Exit(1)
	}

	// Start discovery
	if err := manager.Start(); err != nil {
		fmt.Printf("Error starting discovery: %v\n", err)
		os.Exit(1)
	}
	defer manager.Stop()

	// Create channels for graceful shutdown
	done := make(chan bool)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Start timer
	timer := time.NewTimer(duration)

	// Handle discovery events
	go func() {
		for {
			select {
			case asset := <-manager.GetDiscoveredChannel():
				if jsonOutput {
					data, _ := json.MarshalIndent(asset, "", "  ")
					fmt.Printf("DISCOVERED: %s\n", data)
				} else {
					fmt.Printf("🆕 DISCOVERED: %s (%s) at %s\n",
						asset.Name, asset.Type, asset.IPAddress)
					if verbose {
						printAssetDetails(asset)
					}
				}

			case asset := <-manager.GetUpdatedChannel():
				if jsonOutput {
					data, _ := json.MarshalIndent(asset, "", "  ")
					fmt.Printf("UPDATED: %s\n", data)
				} else if verbose {
					fmt.Printf("🔄 UPDATED: %s (%s)\n", asset.Name, asset.Type)
				}

			case assetID := <-manager.GetLostChannel():
				if jsonOutput {
					fmt.Printf("LOST: {\"asset_id\": \"%s\"}\n", assetID)
				} else {
					fmt.Printf("❌ LOST: %s\n", assetID)
				}

			case queryEvent := <-manager.GetQueryChannel():
				if verbose {
					if jsonOutput {
						data, _ := json.MarshalIndent(queryEvent, "", "  ")
						fmt.Printf("QUERY: %s\n", data)
					} else {
						fmt.Printf("❓ QUERY from %s\n", queryEvent.Sender)
					}
				}

			case <-done:
				return
			}
		}
	}()

	// Wait for completion or interruption
	select {
	case <-timer.C:
		fmt.Printf("\n⏰ Discovery time elapsed\n")
	case <-interrupt:
		fmt.Printf("\n🛑 Discovery interrupted\n")
	}

	close(done)

	// Print summary
	assets := manager.GetAllAssets()
	stats := manager.GetStats()

	if jsonOutput {
		summary := map[string]interface{}{
			"total_assets": len(assets),
			"assets":       assets,
			"stats":        stats,
		}
		data, _ := json.MarshalIndent(summary, "", "  ")
		fmt.Printf("\nSUMMARY: %s\n", data)
	} else {
		fmt.Printf("\n📊 Discovery Summary:\n")
		fmt.Printf("   Total Assets Found: %d\n", len(assets))
		fmt.Printf("   Assets by Type:\n")
		for assetType, count := range stats.AssetsByType {
			fmt.Printf("     %s: %d\n", assetType, count)
		}
		if len(stats.AssetsByRoom) > 0 {
			fmt.Printf("   Assets by Room:\n")
			for room, count := range stats.AssetsByRoom {
				fmt.Printf("     %s: %d\n", room, count)
			}
		}
	}
}

// runAnnounce announces a local asset
func runAnnounce(assetType, assetName, room, ip, capabilities string, duration time.Duration, verbose bool, logger *log.Logger) {
	if assetName == "" {
		fmt.Println("Asset name is required for announce mode")
		os.Exit(1)
	}

	fmt.Printf("📢 Announcing asset '%s' for %v...\n\n", assetName, duration)

	// Parse asset type
	parsedType, err := discovery.ParseAssetType(assetType)
	if err != nil {
		fmt.Printf("Invalid asset type: %v\n", err)
		os.Exit(1)
	}

	// Create asset builder
	builder := discovery.NewAssetBuilder().
		WithType(parsedType).
		WithName(assetName).
		AutoDetectNetwork().
		AutoDetectSystem()

	// Set room if provided
	if room != "" {
		builder.WithRoom(room)
	}

	// Override IP if provided
	if ip != "" {
		builder.WithIPAddress(ip)
	}

	// Parse and add capabilities
	if capabilities != "" {
		caps := strings.Split(capabilities, ",")
		for _, cap := range caps {
			cap = strings.TrimSpace(cap)
			if parsedCap, err := discovery.ParseAssetCapability(cap); err == nil {
				builder.WithCapability(parsedCap)
			} else {
				fmt.Printf("Warning: Invalid capability '%s': %v\n", cap, err)
			}
		}
	}

	// Add default capabilities based on type
	switch parsedType {
	case discovery.AssetTypeGateway:
		builder.WithCapabilities([]discovery.AssetCapability{
			discovery.CapabilityMQTT,
			discovery.CapabilityHTTP,
		}).
			WithHTTPService("api", 8080, "/api", "Home Automation API").
			WithMQTTService("control", "home/control", "Device control")

	case discovery.AssetTypeSmartPlug:
		builder.WithCapabilities([]discovery.AssetCapability{
			discovery.CapabilityPower,
			discovery.CapabilityEnergyMonitor,
			discovery.CapabilitySwitch,
			discovery.CapabilityHTTP,
		}).
			WithHTTPService("tapo-api", 80, "/", "Tapo device API")

	case discovery.AssetTypeSensor:
		builder.WithCapabilities([]discovery.AssetCapability{
			discovery.CapabilityTemperature,
			discovery.CapabilityHumidity,
			discovery.CapabilityMQTT,
		}).
			WithMQTTService("sensors", fmt.Sprintf("sensors/%s", room), "Sensor data")
	}

	asset := builder.Build()

	if verbose {
		fmt.Printf("Asset Configuration:\n")
		printAssetDetails(asset)
		fmt.Printf("\n")
	}

	// Create discovery manager
	config := discovery.DiscoveryConfig{
		LocalAsset: asset,
		Logger:     logger,
	}

	manager, err := discovery.NewDiscoveryManager(config)
	if err != nil {
		fmt.Printf("Error creating discovery manager: %v\n", err)
		os.Exit(1)
	}

	// Start discovery
	if err := manager.Start(); err != nil {
		fmt.Printf("Error starting discovery: %v\n", err)
		os.Exit(1)
	}
	defer manager.Stop()

	// Setup graceful shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Start timer
	timer := time.NewTimer(duration)

	// Send periodic announcements
	announceTicker := time.NewTicker(30 * time.Second)
	defer announceTicker.Stop()

	announceCount := 1
	fmt.Printf("📡 Sent announcement #%d\n", announceCount)

	// Handle events
	for {
		select {
		case <-timer.C:
			fmt.Printf("\n⏰ Announcement time elapsed\n")
			return

		case <-interrupt:
			fmt.Printf("\n🛑 Announcement interrupted\n")
			return

		case <-announceTicker.C:
			manager.Announce()
			announceCount++
			if verbose {
				fmt.Printf("📡 Sent announcement #%d\n", announceCount)
			}

		case queryEvent := <-manager.GetQueryChannel():
			if verbose {
				fmt.Printf("❓ Received query from %s\n", queryEvent.Sender)
			}
		}
	}
}

// runQuery sends discovery queries
func runQuery(queryTypes, queryCaps, room string, duration time.Duration, verbose, jsonOutput bool, logger *log.Logger) {
	fmt.Printf("❓ Sending discovery queries for %v...\n\n", duration)

	// Create discovery manager
	config := discovery.DiscoveryConfig{
		Logger: logger,
	}

	manager, err := discovery.NewDiscoveryManager(config)
	if err != nil {
		fmt.Printf("Error creating discovery manager: %v\n", err)
		os.Exit(1)
	}

	// Start discovery
	if err := manager.Start(); err != nil {
		fmt.Printf("Error starting discovery: %v\n", err)
		os.Exit(1)
	}
	defer manager.Stop()

	// Create query
	query := &discovery.Query{
		MaxAge: 10 * time.Minute,
	}

	// Parse asset types
	if queryTypes != "" {
		types := strings.Split(queryTypes, ",")
		for _, t := range types {
			t = strings.TrimSpace(t)
			if parsedType, err := discovery.ParseAssetType(t); err == nil {
				query.AssetTypes = append(query.AssetTypes, parsedType)
			} else {
				fmt.Printf("Warning: Invalid asset type '%s': %v\n", t, err)
			}
		}
	}

	// Parse capabilities
	if queryCaps != "" {
		caps := strings.Split(queryCaps, ",")
		for _, cap := range caps {
			cap = strings.TrimSpace(cap)
			if parsedCap, err := discovery.ParseAssetCapability(cap); err == nil {
				query.Capabilities = append(query.Capabilities, parsedCap)
			} else {
				fmt.Printf("Warning: Invalid capability '%s': %v\n", cap, err)
			}
		}
	}

	// Set room filter
	if room != "" {
		query.Room = room
	}

	if verbose {
		fmt.Printf("Query Configuration:\n")
		if len(query.AssetTypes) > 0 {
			fmt.Printf("  Asset Types: %v\n", query.AssetTypes)
		}
		if len(query.Capabilities) > 0 {
			fmt.Printf("  Capabilities: %v\n", query.Capabilities)
		}
		if query.Room != "" {
			fmt.Printf("  Room: %s\n", query.Room)
		}
		fmt.Printf("  Max Age: %v\n", query.MaxAge)
		fmt.Printf("\n")
	}

	// Send initial query
	if err := manager.Query(query); err != nil {
		fmt.Printf("Error sending query: %v\n", err)
		os.Exit(1)
	}

	queryCount := 1
	fmt.Printf("📤 Sent query #%d\n", queryCount)

	// Setup graceful shutdown
	done := make(chan bool)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Start timer
	timer := time.NewTimer(duration)

	// Send periodic queries
	queryTicker := time.NewTicker(15 * time.Second)
	defer queryTicker.Stop()

	// Handle events
	go func() {
		for {
			select {
			case asset := <-manager.GetDiscoveredChannel():
				if jsonOutput {
					data, _ := json.MarshalIndent(asset, "", "  ")
					fmt.Printf("RESPONSE: %s\n", data)
				} else {
					fmt.Printf("📥 RESPONSE: %s (%s) at %s\n",
						asset.Name, asset.Type, asset.IPAddress)
					if verbose {
						printAssetDetails(asset)
					}
				}
			case <-done:
				return
			}
		}
	}()

	// Wait for completion or interruption
	for {
		select {
		case <-timer.C:
			fmt.Printf("\n⏰ Query time elapsed\n")
			close(done)
			return

		case <-interrupt:
			fmt.Printf("\n🛑 Query interrupted\n")
			close(done)
			return

		case <-queryTicker.C:
			manager.Query(query)
			queryCount++
			if verbose {
				fmt.Printf("📤 Sent query #%d\n", queryCount)
			}
		}
	}
}

// printAssetDetails prints detailed asset information
func printAssetDetails(asset *discovery.AssetInfo) {
	fmt.Printf("  ID: %s\n", asset.ID)
	fmt.Printf("  Type: %s\n", asset.Type)
	if asset.Model != "" {
		fmt.Printf("  Model: %s\n", asset.Model)
	}
	if asset.Manufacturer != "" {
		fmt.Printf("  Manufacturer: %s\n", asset.Manufacturer)
	}
	if asset.Version != "" {
		fmt.Printf("  Version: %s\n", asset.Version)
	}
	if asset.IPAddress != "" {
		fmt.Printf("  IP: %s\n", asset.IPAddress)
	}
	if asset.MACAddress != "" {
		fmt.Printf("  MAC: %s\n", asset.MACAddress)
	}
	if asset.Hostname != "" {
		fmt.Printf("  Hostname: %s\n", asset.Hostname)
	}
	if asset.Room != "" {
		fmt.Printf("  Room: %s\n", asset.Room)
	}
	if asset.Zone != "" {
		fmt.Printf("  Zone: %s\n", asset.Zone)
	}
	if len(asset.Capabilities) > 0 {
		fmt.Printf("  Capabilities: %v\n", asset.Capabilities)
	}
	if len(asset.Services) > 0 {
		fmt.Printf("  Services:\n")
		for _, service := range asset.Services {
			fmt.Printf("    - %s (%s", service.Name, service.Protocol)
			if service.Port > 0 {
				fmt.Printf(":%d", service.Port)
			}
			if service.Path != "" {
				fmt.Printf("%s", service.Path)
			}
			if service.Topic != "" {
				fmt.Printf(" topic:%s", service.Topic)
			}
			fmt.Printf(")\n")
		}
	}
	if len(asset.Tags) > 0 {
		fmt.Printf("  Tags: %v\n", asset.Tags)
	}
	fmt.Printf("  Status: %s\n", asset.Status)
	fmt.Printf("  Health: %s\n", asset.Health)
	if asset.BatteryLevel != nil {
		fmt.Printf("  Battery: %d%%\n", *asset.BatteryLevel)
	}
	fmt.Printf("  Last Seen: %s\n", asset.LastSeen.Format(time.RFC3339))
	fmt.Printf("  TTL: %d seconds\n", asset.TTL)
}
