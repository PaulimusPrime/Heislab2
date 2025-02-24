package masterfunctions

import (
	"fmt"
	"net"
	"time"
)

const (
	broadCastPort = "30000"
)

func MasterDecider() {
	fmt.Printf("I am master")
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
	addr, _ := net.ResolveUDPAddr("udp", "255.255.255.255:"+broadCastPort)
	conn, _ := net.DialUDP("udp", nil, addr)
	fmt.Print("inside broadcast \n")
	defer conn.Close()
	for {
		message := fmt.Sprintf("Master %d", masterID)
		_, _ = conn.Write([]byte(message))
		time.Sleep(2 * time.Second)
	}
}

func ListenForMaster() {
	addr, _ := net.ResolveUDPAddr("udp", ":"+broadCastPort)
	conn, _ := net.ListenUDP("udp", addr)
	defer conn.Close()

	buffer := make([]byte, 1024)
	for {
		n, _, _ := conn.ReadFromUDP(buffer)
		message := string(buffer[:n])
		fmt.Println("Received: ", message)

		var masterID int
		fmt.Sscanf(message, "Master %d", &masterID)

		// currentMasterID = masterID
	}
}
