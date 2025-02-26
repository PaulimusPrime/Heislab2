package main

import (
	"fmt"
	"net"
	"time"
)

var (
	m = make(map[string]bool) // Map with information of all elements
)

const (
	bufsize       = 1024
	port          = "30001"
	ID            = "PC1" // ID of working computer
	broadCastPort = "30001"
	numNodes      = 3 //The amount of elevators in program
)

func main() {
	//Initiate elevators
	initiateMap()

	// Start UDP listener
	go listenForMessages()

	// Start broadcasting
	broadcastID(ID)
}

// / Function to listen for incoming UDP messages
func listenForMessages() {

	addr := net.UDPAddr{
		Port: 30001,
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
		var peerID string
		n, srcAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading UDP message:", err)
			continue
		}

		if ID == string(buffer[:n]) {
			print("Recieving message from myself\n")
		} else {
			fmt.Printf("Received from %s: %s\n", srcAddr, string(buffer[:n]))
		}
		peerID = string(buffer[:n])
		m[peerID] = true
		fmt.Println(m)
	}

}

// Function to broadcast the computer's ID
func broadcastID(id string) {

	addr, _ := net.ResolveUDPAddr("udp", "255.255.255.255:"+broadCastPort)
	conn, err := net.DialUDP("udp", nil, addr) // Listen to messages on the given address
	fmt.Print("inside broadcast \n")
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
		time.Sleep(3 * time.Second) // Broadcast every 3 seconds
	}
}

func initiateMap() {
	m[ID] = false
	m["PC2"] = false
	m["PC3"] = false
}
