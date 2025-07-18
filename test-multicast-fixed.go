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

type TestMessage struct {
	Type      string `json:"type"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: multicast-test-fixed [send|receive|both]")
		return
	}

	mode := os.Args[1]

	switch mode {
	case "send":
		sendMessages()
	case "receive":
		receiveMessages()
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
		receiveMessagesWithChannel(ctx, receiveChan)
	}()

	// Wait a bit for receiver to start
	time.Sleep(1 * time.Second)

	// Start sender
	go func() {
		sendMessagesWithContext(ctx)
	}()

	// Collect messages
	received := 0
	for {
		select {
		case msg := <-receiveChan:
			fmt.Printf("✅ Received: %s\n", msg)
			received++
		case <-ctx.Done():
			fmt.Printf("\n📊 Test completed. Received %d messages.\n", received)
			return
		}
	}
}

func sendMessages() {
	sendMessagesWithContext(context.Background())
}

func sendMessagesWithContext(ctx context.Context) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", MulticastAddress, MulticastPort))
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
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

		msg := TestMessage{
			Type:      "test",
			Message:   fmt.Sprintf("Test message %d", i+1),
			Timestamp: time.Now().Format(time.RFC3339),
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

		fmt.Printf("📤 Sent: %s\n", msg.Message)
		time.Sleep(1 * time.Second)
	}
}

func receiveMessages() {
	receiveMessagesWithChannel(context.Background(), nil)
}

func receiveMessagesWithChannel(ctx context.Context, msgChan chan<- string) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", MulticastPort))
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenMulticastUDP("udp", nil, &net.UDPAddr{
		IP:   net.ParseIP(MulticastAddress),
		Port: MulticastPort,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	fmt.Printf("📡 Listening for multicast messages on %s:%d\n", MulticastAddress, MulticastPort)

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

		var msg TestMessage
		if err := json.Unmarshal(buffer[:n], &msg); err != nil {
			fmt.Printf("Failed to unmarshal message: %v\n", err)
			continue
		}

		result := fmt.Sprintf("From %s: %s (at %s)", sender, msg.Message, msg.Timestamp)
		if msgChan != nil {
			msgChan <- result
		} else {
			fmt.Printf("📥 Received %s\n", result)
		}
	}
}
