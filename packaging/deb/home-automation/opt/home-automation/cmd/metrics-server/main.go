package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/johnpr01/home-automation/pkg/prometheus"
)

func main() {
	// Create a Prometheus client
	client := prometheus.NewClient("http://localhost:9090")

	// Start metrics server on port 2112 (standard Prometheus metrics port)
	go func() {
		http.Handle("/metrics", client.GetMetricsHandler())
		log.Println("Starting metrics server on :2112")
		if err := http.ListenAndServe(":2112", nil); err != nil {
			log.Printf("Metrics server failed: %v", err)
		}
	}()

	// Give the server a moment to start
	time.Sleep(time.Second)

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

	fmt.Println("\nTest metrics written and server running. You can now:")
	fmt.Println("1. Check metrics at http://localhost:2112/metrics")
	fmt.Println("2. Check Prometheus at http://localhost:9090")
	fmt.Println("3. Query for 'tapo_power_consumption_watts' or 'tapo_energy_total_wh'")
	fmt.Println("4. Check Grafana at http://localhost:3000 (admin/admin)")
	fmt.Println("5. View the 'TP-Link Tapo Energy Monitoring (Prometheus)' dashboard")
	fmt.Println("\nPress Ctrl+C to stop...")

	// Keep running
	select {}
}
