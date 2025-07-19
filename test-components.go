package main

import (
	"fmt"
	"log"
	"time"

	"github.com/johnpr01/home-automation/pkg/discovery"
)

func main() {
	fmt.Println("🧪 Testing Discovery Protocol Components...")

	// Create a test asset
	testAsset := discovery.NewHomeAutomationGateway("Test Gateway").
		WithRoom("office").
		WithZone("main-floor").
		WithTag("test").
		Build()

	fmt.Printf("✅ Created test asset: %s\n", testAsset.Name)

	// Create discovery manager
	config := discovery.DiscoveryConfig{
		LocalAsset:    testAsset,
		AutoQuery:     false,
		QueryInterval: 5 * time.Minute,
		Logger:        log.Default(),
	}

	manager, err := discovery.NewDiscoveryManager(config)
	if err != nil {
		log.Fatalf("Failed to create discovery manager: %v", err)
	}

	fmt.Println("✅ Created discovery manager")

	// Start the manager
	if err := manager.Start(); err != nil {
		log.Fatalf("Failed to start discovery manager: %v", err)
	}

	fmt.Println("✅ Started discovery manager")

	// Handle discovery events in a goroutine
	eventCount := 0
	go func() {
		for {
			select {
			case asset := <-manager.GetDiscoveredChannel():
				eventCount++
				fmt.Printf("🔍 EVENT: Discovered %s (%s) at %s\n",
					asset.Name, asset.Type, asset.IPAddress)

			case asset := <-manager.GetUpdatedChannel():
				eventCount++
				fmt.Printf("🔄 EVENT: Updated %s (%s)\n", asset.Name, asset.Type)

			case assetID := <-manager.GetLostChannel():
				eventCount++
				fmt.Printf("❌ EVENT: Lost %s\n", assetID)

			case queryEvent := <-manager.GetQueryChannel():
				eventCount++
				fmt.Printf("❓ EVENT: Query from %s\n", queryEvent.Sender)
			}
		}
	}()

	// Wait a bit
	fmt.Println("⏱️  Waiting 5 seconds to see announcements...")
	time.Sleep(5 * time.Second)

	// Check stats
	stats := manager.GetStats()
	fmt.Printf("📊 Stats: %d total assets, %d events received\n",
		stats.TotalAssets, eventCount)

	// Get all assets
	allAssets := manager.GetAllAssets()
	fmt.Printf("📋 All assets: %d found\n", len(allAssets))
	for id, asset := range allAssets {
		fmt.Printf("   - %s: %s (%s)\n", id, asset.Name, asset.Type)
	}

	// Try a manual query
	fmt.Println("🔎 Sending manual query...")
	manager.QueryByType(discovery.AssetTypeGateway)

	// Wait for responses
	fmt.Println("⏱️  Waiting 3 seconds for query responses...")
	time.Sleep(3 * time.Second)

	// Final stats
	finalStats := manager.GetStats()
	fmt.Printf("📊 Final Stats: %d total assets, %d events total\n",
		finalStats.TotalAssets, eventCount)

	// Stop the manager
	if err := manager.Stop(); err != nil {
		log.Printf("Error stopping manager: %v", err)
	}

	fmt.Println("✅ Test completed!")
}
