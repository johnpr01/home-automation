package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/johnpr01/home-automation/pkg/prometheus"
)

func main() {
	// Create a Prometheus client
	client := prometheus.NewClient("http://localhost:9090")

	// Test writing some metrics
	fmt.Println("Writing test energy metrics to Prometheus...")

	// Write some test energy readings
	testDevices := []struct {
		deviceID string
		roomID   string
		powerW   float64
		energyWh float64
		isOn     bool
	}{
		{"test-plug-1", "living-room", 150.5, 2500.0, true},
		{"test-plug-2", "kitchen", 75.2, 1800.0, true},
		{"test-plug-3", "bedroom", 0.0, 950.0, false},
	}

	ctx := context.Background()
	timestamp := time.Now()

	for _, device := range testDevices {
		err := client.WriteEnergyReading(
			ctx,
			device.deviceID,
			device.roomID,
			device.powerW,
			device.energyWh,
			0.0, // voltage
			0.0, // current
			device.isOn,
			timestamp,
		)
		if err != nil {
			log.Printf("Failed to write metrics for device %s: %v", device.deviceID, err)
		} else {
			fmt.Printf("✓ Written metrics for device %s\n", device.deviceID)
		}
	}

	fmt.Println("\nTest metrics written. You can now:")
	fmt.Println("1. Check Prometheus at http://localhost:9090")
	fmt.Println("2. Query for 'tapo_energy_power_watts' or 'tapo_energy_energy_wh'")
	fmt.Println("3. Check Grafana at http://localhost:3000 (admin/admin)")
	fmt.Println("4. View the 'TP-Link Tapo Energy Monitoring (Prometheus)' dashboard")
}
