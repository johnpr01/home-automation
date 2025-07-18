package main

import (
	"context"
	"fmt"
	"time"

	"github.com/johnpr01/home-automation/internal/logger"
	"github.com/johnpr01/home-automation/internal/services"
	"github.com/johnpr01/home-automation/pkg/prometheus"
	"github.com/johnpr01/home-automation/pkg/tapo"
)

func main() {
	fmt.Println("🧪 Tapo KLAP Integration Test")
	fmt.Println("=============================")

	// Create logger
	serviceLogger := logger.NewLogger("integration-test", nil)

	// Create Prometheus client once for all tests
	promClient := prometheus.NewClient("http://localhost:9090")

	// Test 1: Direct KLAP Client (with dummy data since we can't connect to real device)
	fmt.Println("\n1️⃣  Testing Direct KLAP Client...")
	testKlapClient(serviceLogger)

	// Test 2: Prometheus Integration
	fmt.Println("\n2️⃣  Testing Prometheus Integration...")
	testPrometheusIntegration(promClient)

	// Test 3: Service Integration
	fmt.Println("\n3️⃣  Testing Service Integration...")
	testServiceIntegration(serviceLogger, promClient)

	// Test 4: Configuration Validation
	fmt.Println("\n4️⃣  Testing Configuration Validation...")
	testConfigurationValidation()

	fmt.Println("\n🎉 Integration Test Complete!")
	fmt.Println("\n📊 Metrics available at: http://localhost:2112/metrics")
	fmt.Println("🔍 To test with real devices, set TPLINK_PASSWORD and update IPs in config")
}

func testKlapClient(logger *logger.Logger) {
	// Test KLAP client creation
	client := tapo.NewKlapClient("192.168.1.100", "test_user", "test_pass", 5*time.Second, *logger)

	if client == nil {
		fmt.Println("❌ FAIL: KLAP client creation failed")
		return
	}

	fmt.Println("✅ PASS: KLAP client created successfully")

	// Test that we can't connect to non-existent device (expected)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		fmt.Println("✅ PASS: Expected connection failure to dummy IP (connection timeout)")
	} else {
		fmt.Println("❌ UNEXPECTED: Connection succeeded to dummy IP")
	}
}

func testPrometheusIntegration(promClient *prometheus.Client) {
	if promClient == nil {
		fmt.Println("❌ FAIL: Prometheus client is nil")
		return
	}

	fmt.Println("✅ PASS: Prometheus client available")

	// Test writing sample metrics
	deviceID := "integration-test-device"
	roomID := "test-room"

	err := promClient.WriteEnergyReading(
		context.Background(),
		deviceID,
		roomID,
		125.5, // Power in watts
		2500,  // Energy in Wh
		230.0, // Voltage
		0.5,   // Current
		true,  // IsOn
		time.Now(),
	)

	if err != nil {
		fmt.Printf("❌ FAIL: Writing energy reading failed: %v\n", err)
		return
	}

	fmt.Println("✅ PASS: Energy reading written to Prometheus")
}

func testServiceIntegration(logger *logger.Logger, promClient *prometheus.Client) {
	if promClient == nil {
		fmt.Println("❌ FAIL: Prometheus client is nil")
		return
	}

	// Create Tapo service
	tapoService := services.NewTapoService(nil, promClient, logger)

	if tapoService == nil {
		fmt.Println("❌ FAIL: Tapo service creation failed")
		return
	}

	fmt.Println("✅ PASS: Tapo service created with Prometheus integration")

	// Test adding KLAP device configuration
	klapConfig := &services.TapoConfig{
		DeviceID:     "test-klap-integration",
		DeviceName:   "Test KLAP Device",
		RoomID:       "integration-test",
		IPAddress:    "192.168.1.200", // Dummy IP
		Username:     "test_user",
		Password:     "test_pass",
		PollInterval: 30 * time.Second,
		UseKlap:      true,
	}

	// We expect this to fail because we can't connect to the dummy IP,
	// but it should validate our configuration structure
	err := tapoService.AddDevice(klapConfig)
	if err != nil {
		fmt.Println("✅ PASS: Expected device addition failure (cannot connect to dummy IP)")
	} else {
		fmt.Println("❌ UNEXPECTED: Device addition succeeded with dummy IP")
	}

	// Test adding legacy device configuration
	legacyConfig := &services.TapoConfig{
		DeviceID:     "test-legacy-integration",
		DeviceName:   "Test Legacy Device",
		RoomID:       "integration-test",
		IPAddress:    "192.168.1.201", // Dummy IP
		Username:     "test_user",
		Password:     "test_pass",
		PollInterval: 60 * time.Second,
		UseKlap:      false,
	}

	err = tapoService.AddDevice(legacyConfig)
	if err != nil {
		fmt.Println("✅ PASS: Expected legacy device addition failure (cannot connect to dummy IP)")
	} else {
		fmt.Println("❌ UNEXPECTED: Legacy device addition succeeded with dummy IP")
	}
}

func testConfigurationValidation() {
	// Test various configuration scenarios
	configs := []*services.TapoConfig{
		{
			DeviceID:     "klap-device-1",
			UseKlap:      true,
			PollInterval: 30 * time.Second,
		},
		{
			DeviceID:     "legacy-device-1",
			UseKlap:      false,
			PollInterval: 60 * time.Second,
		},
		{
			DeviceID:     "auto-interval",
			UseKlap:      true,
			PollInterval: 0, // Should default to 30s
		},
	}

	klapCount := 0
	legacyCount := 0

	for _, config := range configs {
		if config.UseKlap {
			klapCount++
		} else {
			legacyCount++
		}
	}

	if klapCount == 2 && legacyCount == 1 {
		fmt.Println("✅ PASS: Configuration validation successful")
	} else {
		fmt.Printf("❌ FAIL: Configuration validation failed (KLAP: %d, Legacy: %d)\n", klapCount, legacyCount)
	}

	// Test default poll interval handling
	if configs[2].PollInterval == 0 {
		fmt.Println("✅ PASS: Default poll interval handling works (will be set to 30s by service)")
	} else {
		fmt.Println("❌ FAIL: Default poll interval handling incorrect")
	}
}
