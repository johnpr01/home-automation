//go:build integration
// +build integration

package test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/johnpr01/home-automation/pkg/influxdb"
)

func TestInfluxDBIntegration(t *testing.T) {
	// Skip if running in unit test mode
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Get InfluxDB configuration from environment
	url := os.Getenv("INFLUXDB_URL")
	if url == "" {
		url = "http://localhost:8086"
	}

	token := os.Getenv("INFLUXDB_TOKEN")
	if token == "" {
		token = "home-automation-token"
	}

	org := os.Getenv("INFLUXDB_ORG")
	if org == "" {
		org = "home-automation"
	}

	bucket := os.Getenv("INFLUXDB_BUCKET")
	if bucket == "" {
		bucket = "sensor-data"
	}

	// Create InfluxDB client
	client := influxdb.NewClient(url, token, org, bucket)
	if client == nil {
		t.Fatal("Failed to create InfluxDB client")
	}

	// Test connection
	err := client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect to InfluxDB: %v", err)
	}
	defer client.Disconnect()

	// Test writing data
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	testDeviceID := "test-device-integration"
	testRoomID := "test-room"

	// Write test energy reading
	err = client.WriteEnergyReading(ctx, testDeviceID, testRoomID, 100.5, 50.2, 220.0, 2.3, true, time.Now())
	if err != nil {
		t.Fatalf("Failed to write energy reading: %v", err)
	}

	// Write test temperature reading
	err = client.WriteTemperatureReading(ctx, testDeviceID, testRoomID, 72.5, 45.2, time.Now())
	if err != nil {
		t.Fatalf("Failed to write temperature reading: %v", err)
	}

	// Give some time for data to be written
	time.Sleep(2 * time.Second)

	// Test querying data
	readings, err := client.GetLatestEnergyReadings(ctx, testDeviceID, 1)
	if err != nil {
		t.Fatalf("Failed to query energy readings: %v", err)
	}

	if len(readings) == 0 {
		t.Fatal("No energy readings returned")
	}

	// Verify the data
	reading := readings[0]
	if reading.DeviceID != testDeviceID {
		t.Errorf("Expected device ID %s, got %s", testDeviceID, reading.DeviceID)
	}

	if reading.PowerW != 100.5 {
		t.Errorf("Expected power 100.5W, got %fW", reading.PowerW)
	}

	t.Logf("✅ InfluxDB integration test passed successfully")
	t.Logf("   - Connected to: %s", url)
	t.Logf("   - Organization: %s", org)
	t.Logf("   - Bucket: %s", bucket)
	t.Logf("   - Written and read energy data successfully")
}

func TestInfluxDBHealthCheck(t *testing.T) {
	// Skip if running in unit test mode
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Get InfluxDB configuration from environment
	url := os.Getenv("INFLUXDB_URL")
	if url == "" {
		url = "http://localhost:8086"
	}

	token := os.Getenv("INFLUXDB_TOKEN")
	if token == "" {
		token = "home-automation-token"
	}

	org := os.Getenv("INFLUXDB_ORG")
	if org == "" {
		org = "home-automation"
	}

	bucket := os.Getenv("INFLUXDB_BUCKET")
	if bucket == "" {
		bucket = "sensor-data"
	}

	// Create InfluxDB client
	client := influxdb.NewClient(url, token, org, bucket)
	if client == nil {
		t.Fatal("Failed to create InfluxDB client")
	}

	// Test connection and health
	err := client.Connect()
	if err != nil {
		t.Fatalf("InfluxDB health check failed: %v", err)
	}
	defer client.Disconnect()

	t.Logf("✅ InfluxDB health check passed")
	t.Logf("   - URL: %s", url)
	t.Logf("   - Service is healthy and accessible")
}
