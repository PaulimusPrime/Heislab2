package main

import (
	"fmt"
	"net"
	"os/exec"
	"time"
)

var backup_inc bool
var ID int

func main() {
	ID = 10000
	print("\n--- Backup phase ---\n")
	// file, err := os.Open("data.txt")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer func() {
	// 	if err = file.Close(); err != nil {
	// 		log.Fatal(err)
	// 	}
	// }()

	for {
		backupChecking()
		if backup_inc {
			//backupCreation(ID)
			fmt.Printf(".. timed out\n")
			backup_inc = false
			fmt.Printf("\n-- Primary phase --\n")
			BroadcastMasterID(ID)
		}
	}
}

const (
	broadCastPort = "30000"
)

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
		message := fmt.Sprintf("%d", masterID)
		_, _ = conn.Write([]byte(message))
		time.Sleep(2 * time.Second)
	}
}

func ListenForMaster() int {
	addr, _ := net.ResolveUDPAddr("udp", ":"+broadCastPort)
	conn, _ := net.ListenUDP("udp", addr)
	defer conn.Close()

	buffer := make([]byte, 1024)
	n, _, _ := conn.ReadFromUDP(buffer)
	message := string(buffer[:n])
	fmt.Println("Received: ", message)

	var masterID int
	fmt.Sscanf(message, "Master %d", &masterID)

	return masterID
}

func backupCreation(ID int) {
	//cmd1 := exec.Command("gnome-terminal", "--", "simelevatorserver", "--port", strconv.Itoa(ID))
	cmd2 := exec.Command("gnome-terminal", "--", "go", "run", "main.go")
	/*err1 := cmd1.Run()
	if err1 != nil {
		fmt.Printf("Fatal error\n")
	}*/
	err2 := cmd2.Run()
	if err2 != nil {
		fmt.Printf("Fatal error\n")
	}
}

func backupChecking() {
	addr := net.UDPAddr{
		Port: 30000,
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
		fmt.Printf("Received %d bytes from %s: %s\n", n, addr, string(buffer[:n]))
	}
}
