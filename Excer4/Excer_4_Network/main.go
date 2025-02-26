package main

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

var (
	backup_inc    bool = false
	ID            int  = 1
	IsMaster      bool = false
	masterID      int
	port          string = strconv.Itoa(10000 + ID)
	broadCastPort int    = 30000
	electionPort  int    = 30001
	m                    = make(map[int]bool)
)

func main() {

	go BroadcastMasterID(ID)
	go DiscoverPeers()

	print("\n--- Backup phase ---\n")
	for {
		backupChecking()
		if backup_inc {
			fmt.Printf("MasterID: %d \n", masterID)
			//backupCreation(ID)
			fmt.Printf(".. timed out\n")
			backup_inc = false
			fmt.Printf("\n-- Primary phase --\n")
			BroadcastMasterID(ID)
		}
	}
}

func IsMasterAlive(masterID int) bool {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost %d", masterID))
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func BroadcastMasterID(masterID int) {
	addr, _ := net.ResolveUDPAddr("udp", "255.255.255.255:"+strconv.Itoa(broadCastPort))
	conn, _ := net.DialUDP("udp", nil, addr)
	fmt.Print("inside broadcast \n")
	defer conn.Close()
	for {
		message := fmt.Sprintf("Hello from master: %d", masterID)
		_, _ = conn.Write([]byte(message))
		time.Sleep(2 * time.Second)
	}
}

func backupChecking() {
	addr := net.UDPAddr{
		Port: broadCastPort,
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
				DecideNextMaster()
				break
			} else {
				fmt.Print("Error receiving packet", err)
			}
			continue
		}
		fmt.Printf("Received %d bytes from %s: %s\n", n, addr, string(buffer[:n]))
		masterID, err = strconv.Atoi(string(buffer[:n]))
		if err != nil {
			print("error")
		}
	}
}

func DecideNextMaster() int {
	var nextMasterID int

	return nextMasterID
}

func DiscoverPeers() {
	var peerID int
	addr := net.UDPAddr{
		Port: broadCastPort,
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
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))

		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("error")
			return
		}

		fmt.Printf("Received %d bytes from %s: %s\n", n, addr, string(buffer[:n]))
		peerID, err = strconv.Atoi(string(buffer[:n]))
		if err != nil {
			print("error")
			return
		}
		m[peerID] = true
	}

	//strconv.Atoi(string(buffer[:n]))
	//m[ID] := true
}
