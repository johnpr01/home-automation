package main

import (
	"context"
	"crypto/sha1"
	"crypto/sha256"
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
		debug    = flag.Bool("debug", false, "Enable debug output for hash verification")
		help     = flag.Bool("help", false, "Show help message")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Debug KLAP protocol connection issues with Tapo smart plugs.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s -host 192.168.1.100 -username your@email.com -password yourpassword -debug\n", os.Args[0])
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

	// Create a simple logger for testing
	testLogger := logger.NewLogger("tapo-klap-debug", nil)

	fmt.Printf("🔍 Debugging KLAP client connection to %s\n", *host)

	if *debug {
		fmt.Printf("\n📊 Debug Information:\n")
		fmt.Printf("  Username: %s\n", *username)
		fmt.Printf("  Password: %s (length: %d)\n", maskPassword(*password), len(*password))

		// Show hash calculations
		usernameSha1 := sha1Hash([]byte(*username))
		passwordSha1 := sha1Hash([]byte(*password))
		authHash := sha256Hash(concat(usernameSha1, passwordSha1))

		fmt.Printf("\n🔐 Hash Calculations:\n")
		fmt.Printf("  Username SHA1: %x\n", usernameSha1)
		fmt.Printf("  Password SHA1: %x\n", passwordSha1)
		fmt.Printf("  Auth Hash: %x\n", authHash)
		fmt.Printf("\n")
	}

	klapClient := tapo.NewKlapClient(*host, *username, *password, *timeout, *testLogger)

	// Test connection
	ctx := context.Background()
	fmt.Println("🔌 Attempting KLAP connection...")

	if err := klapClient.Connect(ctx); err != nil {
		if *debug {
			fmt.Printf("\n❌ Connection failed with detailed error:\n")
			fmt.Printf("Error: %v\n", err)

			// Check if it's a hash verification error
			if fmt.Sprintf("%v", err) == "server hash verification failed" {
				fmt.Printf("\n💡 Hash Verification Troubleshooting:\n")
				fmt.Printf("1. Verify your TP-Link account credentials are correct\n")
				fmt.Printf("2. Try using your email address (not username) if you're using username\n")
				fmt.Printf("3. Check if your device firmware supports KLAP (1.1.0+)\n")
				fmt.Printf("4. Ensure your account has access to this device\n")
				fmt.Printf("5. Try the legacy protocol if KLAP continues to fail\n")
			}
		}
		log.Fatalf("Failed to connect to device: %v", err)
	}

	fmt.Println("✅ Successfully connected using KLAP protocol!")

	// Get device information
	deviceInfo, err := klapClient.GetDeviceInfo(ctx)
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
	energyUsage, err := klapClient.GetEnergyUsage(ctx)
	if err != nil {
		log.Fatalf("Failed to get energy usage: %v", err)
	}

	fmt.Printf("\n⚡ Energy Usage:\n")
	fmt.Printf("  Current Power: %d mW\n", energyUsage.CurrentPower)
	fmt.Printf("  Today Energy: %d Wh\n", energyUsage.TodayEnergy)
	fmt.Printf("  Month Energy: %d Wh\n", energyUsage.MonthEnergy)
	fmt.Printf("  Today Runtime: %d minutes\n", energyUsage.TodayRuntime)

	fmt.Println("\n🎉 KLAP client debug test completed successfully!")
}

func maskPassword(password string) string {
	if len(password) <= 4 {
		return "****"
	}
	return password[:2] + "****" + password[len(password)-2:]
}

// Helper functions for debug calculations
func sha256Hash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func sha1Hash(data []byte) []byte {
	hash := sha1.Sum(data)
	return hash[:]
}

func concat(arrays ...[]byte) []byte {
	var result []byte
	for _, array := range arrays {
		result = append(result, array...)
	}
	return result
}
