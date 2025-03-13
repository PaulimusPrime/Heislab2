package UDP

import (
	"fmt"
	"net"
)

func min(){
	//Defining local adress
	addr := net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 20023,
	}
	//Creating socket
	recvSock, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Println("Error creating soccket", err)
		return
	}
	defer recvSock.Close()

	sendSock, err := net.DialUDP("udp", nil, &addr)
	if err != nil {
		fmt.Println("Error creating socket", err)
		return
	}
	defer sendSock.Close()

	//Printing what we are listening to
	fmt.Println("listening on", addr.String())

	//Creating a message buffer
	buffer := make([]byte, 1024)

	//Reading message
	for {
		numBytesRecieved, fromWho, err := recvSock.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error recieving data:", err)
			continue
		}
		message := string(buffer[:numBytesRecieved])

		localIP := "0.0.0.0"
		if fromWho.IP.String() != localIP {
			fmt.Printf("Recieved message from %s: %s\n", fromWho.String(), message)
		}
	}
}