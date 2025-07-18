package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

func main() {
	var subnet = flag.String("subnet", "192.168.68", "Subnet to scan (e.g. 192.168.68)")
	flag.Parse()

	fmt.Printf("Scanning for Tapo devices on subnet %s.x\n", *subnet)
	fmt.Println("This may take a few minutes...")

	var wg sync.WaitGroup
	foundDevices := make(chan string, 255)

	// Scan IPs 1-254
	for i := 1; i <= 254; i++ {
		wg.Add(1)
		go func(ip int) {
			defer wg.Done()

			target := fmt.Sprintf("%s.%d", *subnet, ip)

			// Quick TCP connection test
			conn, err := net.DialTimeout("tcp", target+":80", 1*time.Second)
			if err != nil {
				return // No response
			}
			conn.Close()

			// Test if it's a Tapo device by checking /app endpoint
			client := &http.Client{Timeout: 2 * time.Second}
			resp, err := client.Get(fmt.Sprintf("http://%s/app", target))
			if err == nil {
				foundDevices <- fmt.Sprintf("%s (HTTP Status: %s)", target, resp.Status)
				resp.Body.Close()
			} else {
				foundDevices <- fmt.Sprintf("%s (HTTP device, but not Tapo)", target)
			}
		}(i)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(foundDevices)
	}()

	fmt.Println("\nFound devices:")
	count := 0
	for device := range foundDevices {
		fmt.Printf("📱 %s\n", device)
		count++
	}

	if count == 0 {
		fmt.Println("❌ No devices found on the network")
		fmt.Printf("Make sure:\n")
		fmt.Printf("1. Your Tapo devices are powered on and connected\n")
		fmt.Printf("2. You're on the same network as the devices\n")
		fmt.Printf("3. The subnet '%s' is correct for your network\n", *subnet)
	} else {
		fmt.Printf("\n✓ Found %d device(s)\n", count)
	}
}
