package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/johnpr01/home-automation/internal/logger"
	"github.com/johnpr01/home-automation/pkg/tapo"
)

func main() {
	var (
		host     = flag.String("host", "", "IP address of the Tapo device (required)")
		username = flag.String("username", "", "TP-Link account username (required)")
		password = flag.String("password", "", "TP-Link account password (required)")
		timeout  = flag.Duration("timeout", 30*time.Second, "Connection timeout")
		help     = flag.Bool("help", false, "Show help message")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Test legacy protocol connection to a Tapo smart plug.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s -host 192.168.1.100 -username your@email.com -password yourpassword\n", os.Args[0])
	}

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// Validate required flags
	if *host == "" {
		fmt.Fprintf(os.Stderr, "Error: -host is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if *username == "" {
		fmt.Fprintf(os.Stderr, "Error: -username is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if *password == "" {
		fmt.Fprintf(os.Stderr, "Error: -password is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Create a logger for testing
	testLogger := logger.NewLogger("tapo-legacy-test", nil)

	fmt.Printf("🔌 Testing legacy protocol connection to %s\n", *host)

	// Create legacy client
	client := tapo.NewTapoClient(*host, *username, *password, &testLogger)

	// Test connection
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to device using legacy protocol: %v", err)
	}

	fmt.Println("✅ Successfully connected using legacy protocol!")

	// Get device information
	deviceInfo, err := client.GetDeviceInfo(ctx)
	if err != nil {
		log.Fatalf("Failed to get device info: %v", err)
	}

	fmt.Printf("\n📱 Device Info:\n")
	fmt.Printf("  Device ID: %s\n", deviceInfo.DeviceID)
	fmt.Printf("  Model: %s\n", deviceInfo.Model)
	fmt.Printf("  Firmware: %s\n", deviceInfo.FwVersion)
	fmt.Printf("  Device On: %t\n", deviceInfo.DeviceOn)
	fmt.Printf("  RSSI: %d\n", deviceInfo.RSSI)

	// Get energy usage
	energyUsage, err := client.GetEnergyUsage(ctx)
	if err != nil {
		log.Fatalf("Failed to get energy usage: %v", err)
	}

	fmt.Printf("\n⚡ Energy Usage:\n")
	fmt.Printf("  Current Power: %d mW\n", energyUsage.CurrentPower)
	fmt.Printf("  Today Energy: %d Wh\n", energyUsage.TodayEnergy)
	fmt.Printf("  Month Energy: %d Wh\n", energyUsage.MonthEnergy)
	fmt.Printf("  Today Runtime: %d minutes\n", energyUsage.TodayRuntime)

	fmt.Println("\n🎉 Legacy protocol test completed successfully!")
}
