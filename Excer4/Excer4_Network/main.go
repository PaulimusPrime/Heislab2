package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	broadcastPort = "30000"
	broadcastAddr = "255.255.255.255:" + broadcastPort
	listenAddr    = ":" + broadcastPort
)

var backup_inc bool = false // Flag to check if master exists

func main() {
	fmt.Println("\n--- Backup Phase ---")

	// Start listening for master messages
	go ListenForMaster()

	// Give time for any master messages to be received
	time.Sleep(2 * time.Second)

	// If no master detected, become master
	if !backup_inc {
		fmt.Println("No master detected, becoming master.")
		go BroadcastMasterID()
	} else {
		fmt.Println("Master detected, staying as slave.")
	}

	// Keep main function running
	select {}
}

func BroadcastMasterID() {
	addr, err := net.ResolveUDPAddr("udp", broadcastAddr)
	if err != nil {
		fmt.Println("Error resolving broadcast address:", err)
		return
	}

	// Use a temporary local port for sending (nil means OS picks an available port)
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("Error creating UDP socket:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Broadcasting as master...")

	for {
		message := "Master 1"
		_, err := conn.Write([]byte(message)) // Send UDP packet
		if err != nil {
			fmt.Println("Error sending broadcast:", err)
		}
		time.Sleep(2 * time.Second)
	}
}

func ListenForMaster() {
	addr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		fmt.Println("Error resolving listen address:", err)
		return
	}

	// Bind to the fixed port 30000 for listening
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error creating UDP listener:", err)
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1024)
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading UDP message:", err)
			continue
		}

		message := string(buffer[:n])
		fmt.Println("Received:", message)

		if strings.HasPrefix(message, "Master") {
			backup_inc = true
		}
	}
}
