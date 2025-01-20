package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	// Server address
	serverAddress := "10.100.23.204:33546"

	// Connect to the TCP server
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		fmt.Println("Error connecting to the server:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to server:", serverAddress)

	listener, err := net.Listen("tcp", "10.100.23.33:20023")
	if err != nil {
		fmt.Println("Error listening to the server:", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Listening to:", "10.100.23.33:20023")

	// Receive the server's response
	buffer := make([]byte, 1024)
	_, err1 := conn.Read(buffer)
	if err1 != nil {
		fmt.Println("Error reading server response:", err)
	}
	// Print the server's response
	//fmt.Printf("Server response: %s\n", string(buffer[:numBytesReceived]))

	message := "Connect to: 10.100.23.33:20023\000"
	_, err2 := conn.Write([]byte(message))
	if err2 != nil {
		fmt.Println("Error sending message:", err)
	}

	buffer2 := make([]byte, 1024)
	numBytesReceived2, err := conn.Read(buffer2)
	if err != nil {
		fmt.Println("Error reading server response:", err)
	}
	// Print the server's response
	fmt.Printf("Server response: %s\n", string(buffer[:numBytesReceived2]))

	// Communication loop: send and receive messages
	for {

		// Read user input
		fmt.Print("Skriv melding eller skriv ut for å stoppe\n")
		reader := bufio.NewReader(os.Stdin)
		message, _ := reader.ReadString('\n')

		// Exit if the user types "exit"
		if message == "ut\n" {
			fmt.Println("Hade bra...")
			break
		}
		message = message + "\000"

		// Send the message to the server
		_, err := conn.Write([]byte(message))
		if err != nil {
			fmt.Println("Error sending message:", err)
			break
		}

		// Receive the server's response
		buffer := make([]byte, 1024)
		numBytesReceived, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading server response:", err)
			break
		}

		// Print the server's response
		fmt.Printf("Server response: %s\n", string(buffer[:numBytesReceived]))
	}
}
