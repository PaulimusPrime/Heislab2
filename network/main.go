package main

import (
	"Network-go/network/bcast"
	"Network-go/network/localip"
	"Network-go/network/peers"
	"flag"
	"fmt"
	"os"
	"time"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//
//	will be received as zero-values.
type HelloMsg struct {
	Message string
	Iter    int
}

type MasterMsg struct {
	Message  string
	MasterID string
}

var Masterid string
var sjekk bool
var stop bool
func main() {
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	// We make channels for sending and receiving our custom data types
	helloTx := make(chan HelloMsg)
	helloRx := make(chan HelloMsg)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(16569, helloTx)
	go bcast.Receiver(16569, helloRx)

	// We make channels for sending and receiving our custom data types
	MasterbcastTx := make(chan MasterMsg)
	MasterbcastRx := make(chan MasterMsg)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Receiver(17082, MasterbcastRx)
	// The example message. We just send one of these every second.
	go func() {
		helloMsg := HelloMsg{"Hello from " + id, 0}
		for {
			helloMsg.Iter++
			helloTx <- helloMsg
			time.Sleep(1 * time.Second)
		}
	}()
	go func() {
			MasterMsg := MasterMsg{"I am something", id}
			for {
				MasterbcastTx <- MasterMsg
				time.Sleep(1 * time.Second)
				if stop {
					break
				}
			}
	}()

	fmt.Println("Started")
	timeout := time.After(5 * time.Second)
	for {
		if Masterid == id  && Masterid != "" && !sjekk{
			go bcast.Transmitter(17082, MasterbcastTx)
			sjekk = true
			stop = true
			go func(){
				MasterMsg := MasterMsg{"I am the master", Masterid}
				for {
					MasterbcastTx <- MasterMsg
					time.Sleep(1 * time.Second)
				}
			}()
		}
		select {
		case p := <-peerUpdateCh:
			var lostLelevator string = "99"
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			if len(p.Lost) > 0 {
				lostLelevator = p.Lost[0]
			}
			if lostLelevator == Masterid {
				MasterElection(p.Peers, id)
			}

		case <-helloRx:
			//fmt.Printf("Received: %#v\n", a)
		case a := <-MasterbcastRx:
			Masterid = a.MasterID
			fmt.Print(Masterid)
			fmt.Printf("Received: %#v\n", a)
			timeout = time.After(5 * time.Second)
		case <-timeout: // Timeout after 5 seconds
			Masterid=id
			fmt.Println("Timeout: No data received making myself master\n ")
		}
	}
}

func MasterElection(peers []string, id string) {
	fmt.Printf("Master election started\n")
	if id == peers[0] {
		fmt.Printf("I am master\n")
		Masterid = id
		fmt.Print(Masterid,"\n")
	} else {
		fmt.Printf("I am slave\n")
		Masterid = peers[0]
		fmt.Print(Masterid,"\n")
	}

}
