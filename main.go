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
	backgroundtasks "elev_project/network/backgroundTasks"
	"elev_project/network/networkListeners"
	"flag"
	"fmt"
	"time"
)

func main() {
	var (
		//Elevator
		e               elevator.Elevator //elevator struct
		prevFloorSensor = -1              //previous floor sensor
		obstruction     bool              //obstruction
		ImLost          bool              //Tells us if we are on the network or not
		id              string            //id of the elevator

		//Network
		pendingOrderRequests = make(map[string]networkListeners.RequestMsg)
		elevatorStates       = make(map[string]elevator.Elevator) //map of elevator states for all alive elevators
		backupStates         = make(map[string]elevator.Elevator) //map of elevator states for all elevators that have been or is alive
		pendingMasterOrders  = make(map[string][][2]bool)
		Masterid             string //Id of the current master in the network
	)

	//initializing elevator
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()
	fmt.Printf("Elevator starting \n")           //simple print to tell us that the elevator is starting
	elevator.InitializeElevator(&e, id)          //initializing elevator
	ch := networkListeners.InitChannels()        // Initializing channels
	networkListeners.StartListeners(id, ch)      // Starting listeners/transmitters
	backgroundtasks.StartHelloSender(id, ch)     //Start sending hellos/alive messages
	backgroundtasks.StartStateSender(id, ch, &e) //Start sending state messages

	fmt.Println("Started")
	for {
		select {
		// Activates upon change in peers-struct
		case p := <-ch.PeerUpdateCh:
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
					ch.BackupTx <- backupStates
					fmt.Println("Master sending backup")
					// for i := 0; i < config.NumFloors; i++ {
					// 	fmt.Println(backupStates["1"].Requests[i][elevio.BT_Cab])
					// 	fmt.Println(backupStates["2"].Requests[i][elevio.BT_Cab])
					// }
				}
				master.MasterElection(p.Peers, id, &Masterid)
				assigner.Assigner(backupStates, ch.AssignTx, pendingMasterOrders)
			}

		// Activates upon local elevator button press. Adds this to "Elevator" struct "e"
		case button := <-ch.DrvButtons:
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

		// Activates upon local elevator floor arrival. Updates "Elevator" struct "e".
		case floor := <-ch.DrvFloors:
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
						for i := 0; i < config.NumFloors; i++ {
							fmt.Println(backupStates["1"].Requests[i][elevio.BT_Cab])
							fmt.Println(backupStates["2"].Requests[i][elevio.BT_Cab])
						}
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
			runelevator.RunElev(&e, AssRec, id)

		case backup := <-ch.BackupRx:
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
