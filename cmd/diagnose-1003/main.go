package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/johnpr01/home-automation/internal/logger"
	"github.com/johnpr01/home-automation/pkg/tapo"
)

func main() {
	var (
		ip       = flag.String("ip", "", "Device IP address")
		username = flag.String("username", "", "Tapo username")
		password = flag.String("password", "", "Tapo password")
		useKlap  = flag.Bool("klap", false, "Use KLAP protocol instead of legacy")
	)
	flag.Parse()

	if *ip == "" || *username == "" || *password == "" {
		fmt.Println("Usage: diagnose-1003 -ip=192.168.68.x -username=user -password=pass [-klap]")
		fmt.Println("\nThis tool helps diagnose Tapo error code 1003 (Invalid Request)")
		os.Exit(1)
	}

	serviceLogger := logger.NewLogger("diagnose-1003", nil)

	fmt.Printf("🔍 Diagnosing Tapo Error Code 1003\n")
	fmt.Printf("Device: %s\n", *ip)
	fmt.Printf("Protocol: %s\n", func() string {
		if *useKlap {
			return "KLAP"
		}
		return "Legacy"
	}())
	fmt.Printf("Username: %s\n", *username)
	fmt.Printf("Password: %s\n", func() string {
		if len(*password) > 4 {
			return (*password)[:4] + "****"
		}
		return "****"
	}())
	fmt.Println()

	// Test 1: Try Legacy Protocol
	if !*useKlap {
		fmt.Println("1️⃣ Testing Legacy Protocol...")
		client := tapo.NewTapoClient(*ip, *username, *password, serviceLogger)

		err := client.Connect()
		if err != nil {
			fmt.Printf("❌ Legacy handshake failed: %v\n", err)

			// Check if it's specifically error 1003
			if fmt.Sprintf("%v", err) != "" {
				errorStr := fmt.Sprintf("%v", err)
				if len(errorStr) >= 4 && errorStr[len(errorStr)-4:] == "1003" {
					fmt.Println("\n📋 Error 1003 Troubleshooting Guide:")
					fmt.Println("   • Wrong credentials - verify username/password in Tapo app")
					fmt.Println("   • Account not linked to this device")
					fmt.Println("   • Device firmware might not support legacy protocol")
					fmt.Println("   • Try KLAP protocol with -klap flag")
				}
			}
		} else {
			fmt.Println("✅ Legacy handshake successful!")

			// Test device info
			info, err := client.GetDeviceInfo()
			if err != nil {
				fmt.Printf("❌ Get device info failed: %v\n", err)
			} else {
				fmt.Printf("✅ Device info: %s (Model: %s)\n", info.Nickname, info.Model)
			}
		}
	}

	// Test 2: Try KLAP Protocol
	if *useKlap {
		fmt.Println("2️⃣ Testing KLAP Protocol...")
		testLogger := logger.NewLogger("diagnose", nil)
		klapClient := tapo.NewKlapClient(*ip, *username, *password, 30*time.Second, *testLogger)

		ctx := context.Background()
		err := klapClient.Connect(ctx)
		if err != nil {
			fmt.Printf("❌ KLAP handshake failed: %v\n", err)
		} else {
			fmt.Println("✅ KLAP handshake successful!")

			// Test device info
			info, err := klapClient.GetDeviceInfo(ctx)
			if err != nil {
				fmt.Printf("❌ Get device info failed: %v\n", err)
			} else {
				fmt.Printf("✅ Device info: %s (Model: %s)\n", info.Nickname, info.Model)
			}
		}
	}

	// Test 3: Alternative protocol suggestion
	if !*useKlap {
		fmt.Println("\n3️⃣ Testing KLAP Protocol as alternative...")
		testLogger2 := logger.NewLogger("diagnose", nil)
		klapClient := tapo.NewKlapClient(*ip, *username, *password, 30*time.Second, *testLogger2)
		
		ctx := context.Background()
		err := klapClient.Connect(ctx)
		if err != nil {
			fmt.Printf("❌ KLAP also failed: %v\n", err)
		} else {
			fmt.Println("✅ KLAP works! Use -klap flag for this device")
		}
	}

	fmt.Println("\n📖 Common Solutions for Error 1003:")
	fmt.Println("1. Verify credentials in the Tapo mobile app")
	fmt.Println("2. Ensure your account is linked to this specific device")
	fmt.Println("3. Try the alternative protocol (Legacy vs KLAP)")
	fmt.Println("4. Check if device firmware needs updating")
	fmt.Println("5. Reset device and re-add to your account")

	serviceLogger.Info("Diagnostic completed", map[string]interface{}{
		"device_ip": *ip,
		"protocol": func() string {
			if *useKlap {
				return "KLAP"
			} else {
				return "Legacy"
			}
		}(),
	})
}
