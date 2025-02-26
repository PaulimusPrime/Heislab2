package main

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

var (
	m          = make(map[string]bool) // Map with information of all elements
	masterID   int
	backup_inc bool
)

const (
	bufsize       = 1024
	port          = 30001
	ID            = "PC2" // ID of working computer
	broadCastPort = "30001"
	numNodes      = 3 //The amount of elevators in program
)

func main() {

	// Start UDP listener
	go listenForMessages()

	// Start broadcasting
	broadcastID(ID)

	print("\n--- Backup phase ---\n")
	for {
		heartBeatChecking()
		if backup_inc {
			fmt.Printf("MasterID: %d \n", masterID)
			//backupCreation(ID)
			fmt.Printf(".. timed out\n")
			backup_inc = false
			fmt.Printf("\n-- Primary phase --\n")
			broadcastID(ID)
		}
	}
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

func heartBeatChecking() {
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("0.0.0.0"),
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Println("error")
		return
	}
	defer conn.Close()
	buffer := make([]byte, 1024)

	for {
		conn.SetReadDeadline(time.Now().Add(3 * time.Second))

		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Print("No stream")
				print("Master stopped updating\n")
				backup_inc = true

				break
			} else {
				fmt.Print("Error receiving packet", err)
			}
			continue
		}

		masterID, err = strconv.Atoi(string(buffer[:n]))
		if err != nil {
			print("error")
		}
		fmt.Printf("MasterID: %d at: %s\n", masterID, addr)
	}
}
