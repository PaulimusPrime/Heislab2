package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

const port = "30000"

func main() {
	// Get computer ID (hostname)
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println("Error getting hostname:", err)
		return
	}

	// Start UDP listener
	go listenForMessages()

	// Start broadcasting
	broadcastID(hostname)
}

// Function to listen for incoming UDP messages
func listenForMessages() {
	addr := net.UDPAddr{
		Port: 30000,
		IP:   net.ParseIP("0.0.0.0"), // Listen on all interfaces
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Println("Error setting up UDP listener:", err)
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1024)
	for {
		n, srcAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading UDP message:", err)
			continue
		}

		fmt.Printf("Received from %s: %s\n", srcAddr, string(buffer[:n]))
	}
}

// Function to broadcast the computer's ID
func broadcastID(id string) {
	broadcastAddr := net.UDPAddr{
		Port: 30000,
		IP:   net.IPv4bcast, // Broadcast address
	}

	conn, err := net.DialUDP("udp", nil, &broadcastAddr)
	if err != nil {
		fmt.Println("Error setting up UDP connection:", err)
		return
	}
	defer conn.Close()

	for {
		_, err := conn.Write([]byte(id))
		if err != nil {
			fmt.Println("Error broadcasting ID:", err)
		}

		time.Sleep(5 * time.Second) // Broadcast every 5 seconds
	}
}
