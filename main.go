package main

import (
	"elev_project/assigner"
	"elev_project/config"

	"elev_project/driver/elevator"
	"elev_project/driver/elevio"
	"elev_project/driver/fsm"
	"elev_project/driver/master"
	"elev_project/driver/runelevator"
	"elev_project/driver/timer"

	"elev_project/network/bcast"
	"elev_project/network/peers"

	"flag"
	"fmt"

	//"os"
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

type requestMsg struct {
	Floor   int
	Button  elevio.ButtonType
	OrderID string
}

type ackMsg struct {
	OrderID string
	AckType string
}

type ObjectMsg struct {
	Message elevator.Elevator
	ID      string
}

var (
	ackchan    = 16572
	orderchan  = 16571
	peerchan   = 64715
	hellochan  = 61569
	masterchan = 21708
	statechan  = 26573
	assignchan = 11901
	backupchan = 56438
)

// 16572,16571,64715,61569,21708,26573,11901

var pendingOrderRequests = make(map[string]requestMsg)
var elevatorStates = make(map[string]elevator.Elevator)
var backupStates = make(map[string]elevator.Elevator)
var pendingMasterOrders = make(map[string][][2]bool)
var Masterid string
var network_connection bool

func main() {
	//From single elevator
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()
	elevio.Init("localhost:1729"+id, config.NumFloors)

	var (
		e                 elevator.Elevator
		prevRequestButton [config.NumFloors][config.NumButtons]bool
		prevFloorSensor   = -1
		obstruction       bool
		ImLost            bool
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

	OrderTx := make(chan requestMsg)
	OrderRx := make(chan requestMsg)
	ackRx := make(chan ackMsg)
	ackTx := make(chan ackMsg)
	stateTx := make(chan ObjectMsg)
	stateRx := make(chan ObjectMsg)
	assignTx := make(chan map[string][][2]bool)
	assignRx := make(chan map[string][][2]bool)
	backupTx := make(chan map[string]elevator.Elevator)
	backupRx := make(chan map[string]elevator.Elevator)

	//From network

	// Our id can be anything. Here we pass it on the command line, using
	//  go run main.go -id=our_id

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)

	go peers.Transmitter(peerchan, id, peerTxEnable)
	go peers.Receiver(peerchan, peerUpdateCh)

	helloTx := make(chan HelloMsg)
	helloRx := make(chan HelloMsg)

	go bcast.Transmitter(hellochan, helloTx)
	go bcast.Receiver(hellochan, helloRx)

	MasterbcastTx := make(chan MasterMsg)
	MasterbcastRx := make(chan MasterMsg)
	go bcast.Receiver(masterchan, MasterbcastRx)

	go func() {
		helloMsg := HelloMsg{"Hello from " + id, 0}
		for {
			helloMsg.Iter++
			helloTx <- helloMsg
			time.Sleep(1 * time.Second)
		}
	}()

	go bcast.Receiver(orderchan, OrderRx)
	go bcast.Transmitter(orderchan, OrderTx)
	go bcast.Transmitter(ackchan, ackTx)
	go bcast.Receiver(ackchan, ackRx)
	go bcast.Transmitter(statechan, stateTx)
	go bcast.Receiver(statechan, stateRx)
	go bcast.Transmitter(assignchan, assignTx)
	go bcast.Receiver(assignchan, assignRx)
	go bcast.Transmitter(backupchan, backupTx)
	go bcast.Receiver(backupchan, backupRx)

	go func() {
		time.Sleep(time.Second)
		for {
			stateTx <- ObjectMsg{Message: e, ID: id}
			time.Sleep(50 * time.Millisecond)
		}
	}()

	fmt.Println("Started")
	//timeout := time.After(5 * time.Second)
	for {
		// If local elevator is dedicated master, starts broadcasting heartbeat. Network connection is not what it seems
		if Masterid == id && Masterid != "" && !network_connection {
			go bcast.Transmitter(masterchan, MasterbcastTx)
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
			var lostElevator string = "99" // To ensure there is no master when initializing the network
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

			if len(p.Lost) > 0 {
				lostElevator = p.Lost[0]
			}
			if lostElevator == Masterid && len(p.Peers) > 0 {
				master.MasterElection(p.Peers, id, &Masterid)
			}
			for _, lostID := range p.Lost {
				if lostID == id {
					ImLost = true
				}
				delete(pendingMasterOrders, lostID)
				delete(elevatorStates, lostID)
			}
			if len(p.New) > 0 {
				ImLost = false
				if Masterid == id {
					backupTx <- backupStates
					fmt.Println("Master sending backup")
					// for i := 0; i < config.NumFloors; i++ {
					// 	fmt.Println(backupStates["1"].Requests[i][elevio.BT_Cab])
					// 	fmt.Println(backupStates["2"].Requests[i][elevio.BT_Cab])
					// }
				}
				master.MasterElection(p.Peers, id, &Masterid)
				assigner.Assigner(backupStates, assignTx, pendingMasterOrders)
			}
		// Activates upon recieved heartbeat from master
		case a := <-MasterbcastRx:
			Masterid = a.MasterID
			//fmt.Printf("Received: %#v\n", a) //DENNE ER TATT UT
			//timeout = time.After(2 * time.Second)

		// Activates if not recieved master heartbeat
		// case <-timeout: // Timeout after 5 seconds
		// 	if Masterid == id {
		// 		//Ute av nettverket
		// 		network_connection = false
		// 	}
		// 	Masterid = id
		// 	fmt.Println("Timeout: No data received making myself master\n ")

		// Activates upon local elevator button press. Adds this to "Elevator" struct "e"
		case button := <-drv_buttons:
			if button.Button == elevio.ButtonType(elevio.BT_Cab) {
				e.Requests[button.Floor][button.Button] = true
				fsm.Fsm_onRequestButtonPress(&e, button.Floor, button.Button)
				elevio.SetButtonLamp(button.Button, button.Floor, true)

			} else if !ImLost {
				if Masterid == id {
					e.Requests[button.Floor][button.Button] = true
					elevatorStates[id] = e
					backupStates[id] = e
					e.Requests[button.Floor][button.Button] = true
					assigner.Assigner(elevatorStates, assignTx, pendingMasterOrders)
					runelevator.RunElev(&e, pendingMasterOrders, id)

				} else {
					e.Requests[button.Floor][button.Button] = true
					request := requestMsg{button.Floor, button.Button, id}
					OrderTx <- request
					pendingOrderRequests[request.OrderID] = request
					fmt.Println("Added order to pendingOrders from:", request.OrderID)
				}
			}

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
				elevio.SetDoorOpenLamp(true)
				timer.StartTimer(config.ObstructionDurationS)
			}
		// Obstruction activated.
		case <-drv_obstr:
			obstruction = !obstruction

		// Stop button pressed.
		case stop := <-drv_stop:
			if stop {
				elevio.SetMotorDirection(elevio.MD_Stop)
				e.Behaviour = elevator.EB_Idle
			}
			time.Sleep(time.Duration(config.InputPollRate))

		case r := <-OrderRx: //r for request
			if Masterid == id {
				ack := ackMsg{OrderID: r.OrderID, AckType: "order"}
				ackTx <- ack
				fmt.Println("Sent ACK from :", ack.OrderID)

				updateDeadline := time.Now().Add(200 * time.Millisecond)
				for time.Now().Before(updateDeadline) {
					select {
					case state := <-stateRx:
						backupStates[state.ID] = state.Message
						elevatorStates[state.ID] = state.Message
						for i := 0; i < config.NumFloors; i++ {
							fmt.Println(backupStates["1"].Requests[i][elevio.BT_Cab])
							fmt.Println(backupStates["2"].Requests[i][elevio.BT_Cab])
						}

					default:
						time.Sleep(50 * time.Millisecond)
					}
				}

				assigner.Assigner(elevatorStates, assignTx, pendingMasterOrders)

			}

		case state := <-stateRx:
			backupStates[state.ID] = state.Message
			elevatorStates[state.ID] = state.Message
			fsm.SetHallLights(elevatorStates)

		case ack := <-ackRx:
			if ack.AckType == "order" {
				fmt.Println("Received ACK from:", ack.OrderID)
				delete(pendingOrderRequests, ack.OrderID) // Remove acknowledged order from pendingOrderRequests
			} else if ack.AckType == "assign" {
				if _, exists := pendingMasterOrders[ack.OrderID]; exists {
					delete(pendingMasterOrders, ack.OrderID)
					fmt.Println("Order acknowledged and removed from pendingMasterOrders:", ack.OrderID)
				}
			}

		case <-time.After(200 * time.Millisecond):
			for _, request := range pendingOrderRequests {
				OrderTx <- request
			}
			for id, orders := range pendingMasterOrders {
				assignTx <- map[string][][2]bool{id: orders}
				fmt.Printf("Id: %s not acknowledged. Resending orders\n", id)
			}

		case AssRec := <-assignRx:

			ack := ackMsg{OrderID: id, AckType: "assign"}
			ackTx <- ack

			runelevator.RunElev(&e, AssRec, id)

		case backup := <-backupRx:
			fmt.Println("Backed up")
			e.Requests = backup[id].Requests

			for i := 0; i < config.NumFloors; i++ {
				if e.Requests[i][elevio.BT_Cab] {
					fsm.Fsm_onRequestButtonPress(&e, i, elevio.BT_Cab)
				}

			}
		}
	}
}
