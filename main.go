package main

import (
	// Imports from assigner

	// Imports from cases

	// Imports from config

	// Imports from driver
	"elev_project/assigner"
	"elev_project/cases"
	"elev_project/config"
	"elev_project/driver/elevator"
	"elev_project/driver/elevio"
	"elev_project/driver/fsm"
	"elev_project/driver/runelevator"
	"elev_project/driver/timer"
	"time"

	//"elev_project/driver/master"

	// Imports from network
	backgroundtasks "elev_project/network/backgroundTasks"
	"elev_project/network/networkListeners"

	// Library imports
	"flag"
	"fmt"
)

func main() {
	var (
		//Elevator
		e               elevator.Elevator //elevator struct
		obstruction     bool              //obstruction
		ImLost          bool              //Tells us if we are on the network or not
		id              string            //id of the elevator
		Motorstop       bool              = false
		prevFloorSensor                   = -1

		//Network
		pendingOrderRequests        = make(map[string]networkListeners.RequestMsg)
		elevatorStates              = make(map[string]elevator.Elevator) //map of elevator states for all alive elevators
		backupStates                = make(map[string]elevator.Elevator) //map of elevator states for all elevators that have been or is alive
		pendingMasterOrders         = make(map[string][][2]bool)
		Masterid             string // Local understanding of the current master in the network
		timeout              = time.After(7 * time.Second)
	)

	//initializing elevator
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()
	fmt.Printf("Elevator starting \n")      //simple print to tell us that the elevator is starting
	elevator.InitializeElevator(&e, id)     //initializing elevator
	ch := networkListeners.InitChannels()   // Initializing channels
	networkListeners.StartListeners(id, ch) // Starting listeners/transmitters
	//backgroundtasks.StartHelloSender(id, ch)     //Start sending hellos/alive messages
	backgroundtasks.StartStateSender(id, ch, &e, &Motorstop) //Start sending state messages

	fmt.Println("Started")
	go cases.PeersUpdate(ch, id, &Masterid, &ImLost, pendingMasterOrders, elevatorStates, backupStates, <-ch.PeerUpdateCh, &e)
	go cases.HandleButtonPress(ch, id, &Masterid, &ImLost, pendingMasterOrders, elevatorStates, backupStates, pendingOrderRequests, &e)
	for {
		select {
		// Activates upon change in peers-struct
		//case p := <-ch.PeerUpdateCh:
		//	cases.PeersUpdate(ch, id, &Masterid, &ImLost, pendingMasterOrders, elevatorStates, backupStates, p)
		// var lostElevator string = "99" // To ensure there is no master when initializing the network
		// fmt.Printf("Peer update:\n")
		// fmt.Printf("  Peers:    %q\n", p.Peers)
		// fmt.Printf("  New:      %q\n", p.New)
		// fmt.Printf("  Lost:     %q\n", p.Lost)

		// if len(p.Lost) > 0 {
		// 	lostElevator = p.Lost[0]
		// }
		// if lostElevator == Masterid && len(p.Peers) > 0 {
		// 	master.MasterElection(p.Peers, id, &Masterid)
		// }
		// for _, lostID := range p.Lost {
		// 	if lostID == id {
		// 		ImLost = true
		// 	}
		// 	delete(pendingMasterOrders, lostID)
		// 	delete(elevatorStates, lostID)

		// 	for i := 0; i < config.NumFloors; i++ {
		// 		temp := backupStates[lostID]

		// 		e.Requests[i][elevio.BT_HallDown] = e.Requests[i][elevio.BT_HallDown] || backupStates[lostID].Requests[i][elevio.BT_HallDown]
		// 		e.Requests[i][elevio.BT_HallUp] = e.Requests[i][elevio.BT_HallUp] || backupStates[lostID].Requests[i][elevio.BT_HallUp]
		// 		temp.Requests[i][elevio.BT_HallDown] = false
		// 		temp.Requests[i][elevio.BT_HallUp] = false
		// 		backupStates[lostID] = temp
		// 		elevatorStates[id] = e
		// 	}
		// 	if id == Masterid {
		// 		assigner.Assigner(elevatorStates, ch.AssignTx, pendingMasterOrders)
		// 	}

		// }
		// if len(p.New) > 0 {
		// 	ImLost = false
		// 	if Masterid == id {
		// 		for i := 0; i < 20; i++ {
		// 			ch.BackupTx <- backupStates
		// 			time.Sleep(50 * time.Millisecond)
		// 		}

		// 		fmt.Println("Master sending backup")
		// 	}
		// 	master.MasterElection(p.Peers, id, &Masterid)
		// 	assigner.Assigner(backupStates, ch.AssignTx, pendingMasterOrders)
		// }

		// Activates upon local elevator button press. Adds this to "Elevator" struct "e"
		//case button := <-ch.DrvButtons:
		//	cases.HandleButtonPress(ch, id, &Masterid, &ImLost, pendingMasterOrders, elevatorStates, backupStates, pendingOrderRequests, &e, button)
		/*
			timeout = time.After(7 * time.Second)
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
					assigner.Assigner(elevatorStates, ch.AssignTx, pendingMasterOrders)
					runelevator.RunElev(&e, pendingMasterOrders, id)
				} else {
					e.Requests[button.Floor][button.Button] = true
					request := networkListeners.RequestMsg{button.Floor, button.Button, id}
					ch.OrderTx <- request
					pendingOrderRequests[request.OrderID] = request
					fmt.Println("Added order to pendingOrders from:", request.OrderID)
				}
			}
		*/
		// Activates upon local elevator floor arrival. Updates "Elevator" struct "e".
		case floor := <-ch.DrvFloors:

			if floor != -1 && floor != prevFloorSensor {
				fsm.Fsm_onFloorArrival(&e, floor)
			} else {
				prevFloorSensor = floor
			}
			// cases.HandleFloorArrival(&e, &floor, &prevFloorSensor)
			timeout = time.After(7 * time.Second)
			Motorstop = false
			ch.PeerTxEnable <- true

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
		case <-ch.DrvObstr:
			obstruction = !obstruction

		// Stop button pressed.
		case stop := <-ch.DrvStop:
			if stop {
				elevio.SetMotorDirection(elevio.MD_Stop)
				e.Behaviour = elevator.EB_Idle
			}
			time.Sleep(time.Duration(config.InputPollRate))

		case r := <-ch.OrderRx: //r for request
			if Masterid == id {
				ack := networkListeners.AckMsg{OrderID: r.OrderID, AckType: "order"}
				ch.AckTx <- ack
				fmt.Println("Sent ACK from :", ack.OrderID)
				updateDeadline := time.Now().Add(200 * time.Millisecond)
				for time.Now().Before(updateDeadline) {
					select {
					case state := <-ch.StateRx:
						backupStates[state.ID] = state.Message
						elevatorStates[state.ID] = state.Message
					default:
						time.Sleep(50 * time.Millisecond)
					}
				}
				assigner.Assigner(elevatorStates, ch.AssignTx, pendingMasterOrders)
			}

		case state := <-ch.StateRx:
			backupStates[state.ID] = state.Message
			elevatorStates[state.ID] = state.Message
			fsm.SetHallLights(elevatorStates)

		case ack := <-ch.AckRx:
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
				ch.OrderTx <- request
			}
			for id, orders := range pendingMasterOrders {
				ch.AssignTx <- map[string][][2]bool{id: orders}
				fmt.Printf("Id: %s not acknowledged. Resending orders\n", id)
			}

		case AssRec := <-ch.AssignRx:
			ack := networkListeners.AckMsg{OrderID: id, AckType: "assign"}
			ch.AckTx <- ack
			timeout = time.After(7 * time.Second)
			runelevator.RunElev(&e, AssRec, id)

		case backup := <-ch.BackupRx:
			e.Requests = backup[id].Requests

			for i := 0; i < config.NumFloors; i++ {
				if e.Requests[i][elevio.BT_Cab] {
					fsm.Fsm_onRequestButtonPress(&e, i, elevio.BT_Cab)
				}
			}

		case <-timeout:
			if e.Behaviour == elevator.EB_Moving {
				ch.PeerTxEnable <- false
				Motorstop = true
				time.Sleep(1000 * time.Millisecond)
				fmt.Println("MOTORSTOP")
			} else {
				timeout = time.After(7 * time.Second)
				fmt.Println("Timer Restarted")
			}
		}
	}
}
