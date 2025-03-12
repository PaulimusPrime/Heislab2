package main

import (
	"Network-go/network/bcast"
	"Network-go/network/config"
	"Network-go/network/elevator"
	"Network-go/network/elevio"
	"Network-go/network/fsm"
	"Network-go/network/localip"
	"Network-go/network/master"
	"Network-go/network/peers"
	"Network-go/network/timer"
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

// -------------------------------------------M---------------------------------------------------
// Martin remotework order distribution
type requestMsg struct {
	Floor  int
	Button elevio.ButtonType
}

type prevRequestMsg struct {
	Floor  int
	Button elevio.ButtonType
}

type orderAll struct {
	Orders [config.NumFloors][config.NumButtons]bool
}

//-------------------------------------------M---------------------------------------------------

type ButtonMsg struct {
	Message elevator.Elevator
	ID      string
}

var Masterid string
var network_connection bool

func main() {
	//From single elevator
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()
	elevio.Init("localhost:1729"+id, config.NumFloors)

	var (
		pr                prevRequestMsg //Martin remotework order distribution
		e                 elevator.Elevator
		prevRequestButton [config.NumFloors][config.NumButtons]bool
		prevFloorSensor   = -1
		obstruction       bool
	)

	//initializing elevator
	fmt.Printf("Elevator starting \n")
	elevator.InitializeElevator(&e, &prevRequestButton)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	//-------------------------------------------M---------------------------------------------------
	//Martin remotework order distribution
	OrderTx := make(chan requestMsg)
	OrderRx := make(chan requestMsg)
	allOrdersBcast := make(chan orderAll)
	allOrdersRecv := make(chan orderAll)
	//-------------------------------------------M---------------------------------------------------

	//From network

	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`

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

	// ---------------------------------------------------------------
	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(64715, id, peerTxEnable)
	go peers.Receiver(64715, peerUpdateCh)

	// ---------------------------------------------------------------
	// We make channels for sending and receiving our custom data types
	helloTx := make(chan HelloMsg)
	helloRx := make(chan HelloMsg)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(61569, helloTx)
	go bcast.Receiver(61569, helloRx)

	// ----------------------------------------------------------------
	// We make channels for sending and receiving our custom data types
	MasterbcastTx := make(chan MasterMsg)
	MasterbcastRx := make(chan MasterMsg)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Receiver(21708, MasterbcastRx)

	// --------------------------------------------------------------

	// The example message. We just send one of these every second.
	go func() {
		helloMsg := HelloMsg{"Hello from " + id, 0}
		for {
			helloMsg.Iter++
			helloTx <- helloMsg
			time.Sleep(1 * time.Second)
		}
	}()
	// go func() {
	// 	for {
	// 		if sjekk {
	// 			break
	// 		}
	// 		MasterMsg := MasterMsg{"I am something", id}
	// 		MasterbcastTx <- MasterMsg
	// 		time.Sleep(1 * time.Second)
	// 	}
	// }()

	//-------------------------------------------M---------------------------------------------------
	//Martin remotework order distribution
	if Masterid == id {
		go func() {
			allOrdersBcast <- orderAll{e.Requests}
		}()
		go bcast.Transmitter(16570, allOrdersBcast) //Broadcaster konstant alle ordre
	}
	go bcast.Receiver(16570, allOrdersRecv) //Lytter på alle ordre

	go bcast.Receiver(16571, OrderRx)    //Lytter på enkelt order fra individuelle heiser
	go bcast.Transmitter(16571, OrderTx) //MARTIN FIX DENNE
	//-------------------------------------------M---------------------------------------------------

	fmt.Println("Started")
	timeout := time.After(5 * time.Second)
	for {
		// If local elevator is dedicated master, starts broadcasting heartbeat. Network connection is not what it seems
		if Masterid == id && Masterid != "" && !network_connection {
			go bcast.Transmitter(21708, MasterbcastTx)
			network_connection = true
			go func() {
				MasterMsg := MasterMsg{"I am the master", Masterid}
				for {
					MasterbcastTx <- MasterMsg
					time.Sleep(1 * time.Second)
					if !network_connection {
						break
					}
				}
			}()
		}
		select {
		// Activates upon change in peers-struct
		case p := <-peerUpdateCh:
			var lostLelevator string = "99"
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			if len(p.Lost) > 0 {
				lostLelevator = p.Lost[0]
			}
			if lostLelevator == Masterid && len(p.Peers) > 0 {
				master.MasterElection(p.Peers, id, &Masterid)
			}
		// Activates upon recieved heartbeat from master
		case a := <-MasterbcastRx:
			Masterid = a.MasterID
			//fmt.Printf("Received: %#v\n", a) DENNE ER TATT UT
			timeout = time.After(2 * time.Second)

		// Activates if not recieved master heartbeat
		case <-timeout: // Timeout after 5 seconds
			if Masterid == id {
				//Ute av nettverket
				network_connection = false
				print("I'm out of the network \n")
			}
			Masterid = id
			fmt.Println("Timeout: No data received making myself master\n ")

		// Activates upon local elevator button press. Adds this to "Elevator" struct "e"
		case button := <-drv_buttons:
			//Gamle funksjonalitet:
			//e.Requests[button.Floor][button.Button] = true
			//elevio.SetButtonLamp(button.Button, button.Floor, true)
			//fsm.Fsm_onRequestButtonPress(&e, button.Floor, button.Button)

			//-------------------------------------------M---------------------------------------------------
			//Martin remotework order distribution
			//Om du er master, putt inn ordre, ellers broadcast ordre
			if Masterid == id {
				e.Requests[button.Floor][button.Button] = true
				elevio.SetButtonLamp(button.Button, button.Floor, true)
			} else {
				OrderTx <- requestMsg{button.Floor, button.Button}
				pr.Floor = button.Floor
				pr.Button = button.Button
			}

			//-------------------------------------------M---------------------------------------------------

		// Activates upon local elevator floor arrival. Updates "Elevator" struct "e".
		case floor := <-drv_floors:
			if floor != -1 && floor != prevFloorSensor {
				fsm.Fsm_onFloorArrival(&e, floor)
			} else {
				prevFloorSensor = floor
			}

		// Starts door timer if not obstructed
		case <-timer.TimerChannel:
			if !obstruction {
				fsm.Fsm_onDoorTimeout(&e)
				obstruction = false
			} else {
				timer.StartTimer(config.ObstructionDurationS)
			}
		// Obstruction activated.
		case <-drv_obstr:
			if e.Behaviour == elevator.EB_DoorOpen {
				elevio.SetDoorOpenLamp(true)
				obstruction = !obstruction
			}

		// Stop button pressed.
		case stop := <-drv_stop:
			if stop {
				elevio.SetMotorDirection(elevio.MD_Stop)
				e.Behaviour = elevator.EB_Idle
			}
			time.Sleep(time.Duration(config.InputPollRate))

		//-------------------------------------------M---------------------------------------------------
		//Martin remotework order distribution
		//Mottar order på OrderRx, setter Masters ordre i e.Requests og setter lys på knapp
		case r := <-OrderRx: //r for request
			if Masterid == id {
				e.Requests[r.Floor][r.Button] = true
				elevio.SetButtonLamp(r.Button, r.Floor, true)
			}

		case o := <-allOrdersRecv: //o for ordre
			if Masterid != id {
				if !o.Orders[pr.Button][pr.Floor] {
					OrderTx <- requestMsg(pr)

				}
			}

		case <-drv_stop:
			fmt.Println("Order list:")

			// Print Hall Up buttons
			fmt.Print("Hall Up   : ")
			fmt.Printf("[-] ")
			for floor := 0; floor < config.NumFloors-1; floor++ {
				if e.Requests[floor][elevio.BT_HallUp] {
					fmt.Printf("[1] ")
				} else {
					fmt.Printf("[0] ")
				}
			}
			fmt.Println()
			// Print Hall Down buttons
			fmt.Print("Hall Down : ")
			for floor := 1; floor < config.NumFloors; floor++ {
				if e.Requests[floor][elevio.BT_HallDown] {
					fmt.Printf("[1] ")
				} else {
					fmt.Printf("[0] ")
				}
			}
			fmt.Println("[-] ")
			fmt.Println()
			//-------------------------------------------M---------------------------------------------------
		}
	}
}
