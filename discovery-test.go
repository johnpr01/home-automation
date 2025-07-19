package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

const (
	MulticastAddress = "239.255.42.42"
	MulticastPort    = 42424
)

// Simplified discovery message for testing
type DiscoveryTestMessage struct {
	Type      string    `json:"type"`
	AssetName string    `json:"asset_name"`
	AssetType string    `json:"asset_type"`
	Timestamp time.Time `json:"timestamp"`
	Sender    string    `json:"sender"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: discovery-test [send|receive|both]")
		return
	}

	mode := os.Args[1]

	switch mode {
	case "send":
		sendDiscoveryMessages()
	case "receive":
		receiveDiscoveryMessages()
	case "both":
		runBoth()
	default:
		fmt.Println("Invalid mode. Use 'send', 'receive', or 'both'")
	}
}

func runBoth() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Start receiver
	receiveChan := make(chan string, 10)
	go func() {
		receiveDiscoveryMessagesWithChannel(ctx, receiveChan)
	}()

	// Wait a bit for receiver to start
	time.Sleep(1 * time.Second)

	// Start sender
	go func() {
		sendDiscoveryMessagesWithContext(ctx)
	}()

	// Collect messages
	received := 0
	for {
		select {
		case msg := <-receiveChan:
			fmt.Printf("✅ Discovery Received: %s\n", msg)
			received++
		case <-ctx.Done():
			fmt.Printf("\n📊 Discovery test completed. Received %d messages.\n", received)
			return
		}
	}
}

func sendDiscoveryMessages() {
	sendDiscoveryMessagesWithContext(context.Background())
}

func sendDiscoveryMessagesWithContext(ctx context.Context) {
	// Create separate send connection like our fixed protocol
	multicastAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", MulticastAddress, MulticastPort))
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, multicastAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	for i := 0; i < 5; i++ {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msg := DiscoveryTestMessage{
			Type:      "announce",
			AssetName: fmt.Sprintf("Test Asset %d", i+1),
			AssetType: "gateway",
			Timestamp: time.Now(),
			Sender:    "test-sender",
		}

		data, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Marshal error: %v", err)
			continue
		}

		_, err = conn.Write(data)
		if err != nil {
			log.Printf("Send error: %v", err)
			continue
		}

		fmt.Printf("📤 Discovery Sent: %s (%s)\n", msg.AssetName, msg.AssetType)
		time.Sleep(1 * time.Second)
	}
}

func receiveDiscoveryMessages() {
	receiveDiscoveryMessagesWithChannel(context.Background(), nil)
}

func receiveDiscoveryMessagesWithChannel(ctx context.Context, msgChan chan<- string) {
	// Create multicast listener like our fixed protocol
	conn, err := net.ListenMulticastUDP("udp", nil, &net.UDPAddr{
		IP:   net.ParseIP(MulticastAddress),
		Port: MulticastPort,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	fmt.Printf("📡 Listening for discovery messages on %s:%d\n", MulticastAddress, MulticastPort)

	buffer := make([]byte, 1024)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, sender, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			log.Printf("Read error: %v", err)
			continue
		}

		var msg DiscoveryTestMessage
		if err := json.Unmarshal(buffer[:n], &msg); err != nil {
			fmt.Printf("Failed to unmarshal discovery message: %v\n", err)
			continue
		}

		result := fmt.Sprintf("From %s: %s (%s) at %s", sender, msg.AssetName, msg.AssetType, msg.Timestamp.Format("15:04:05"))
		if msgChan != nil {
			msgChan <- result
		} else {
			fmt.Printf("📥 Discovery Received: %s\n", result)
		}
	}
}
