package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

func main() {
	var ip = flag.String("ip", "", "Device IP address to test")
	flag.Parse()

	if *ip == "" {
		fmt.Println("Usage: network-test -ip=192.168.1.x")
		os.Exit(1)
	}

	fmt.Printf("Testing network connectivity to %s\n", *ip)

	// Test 1: Basic ping (TCP connection test)
	fmt.Println("1. Testing TCP connectivity...")
	conn, err := net.DialTimeout("tcp", *ip+":80", 5*time.Second)
	if err != nil {
		fmt.Printf("❌ TCP connection failed: %v\n", err)
	} else {
		fmt.Printf("✓ TCP connection successful\n")
		conn.Close()
	}

	// Test 2: HTTP GET request to see if device responds
	fmt.Println("2. Testing HTTP response...")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://%s/", *ip))
	if err != nil {
		fmt.Printf("❌ HTTP request failed: %v\n", err)
	} else {
		fmt.Printf("✓ HTTP response received (Status: %s)\n", resp.Status)
		resp.Body.Close()
	}

	// Test 3: Test Tapo app endpoint
	fmt.Println("3. Testing Tapo app endpoint...")
	resp, err = client.Get(fmt.Sprintf("http://%s/app", *ip))
	if err != nil {
		fmt.Printf("❌ Tapo app endpoint failed: %v\n", err)
	} else {
		fmt.Printf("✓ Tapo app endpoint accessible (Status: %s)\n", resp.Status)
		resp.Body.Close()
	}

	fmt.Println("\nNetwork test completed.")
}
