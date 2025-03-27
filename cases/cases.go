package cases

import (
	"elev_project/assigner"
	"elev_project/config"
	"elev_project/driver/elevator"
	"elev_project/driver/elevio"
	"elev_project/driver/fsm"
	"elev_project/driver/runelevator"
	"elev_project/driver/timer"
	"elev_project/network"
	"elev_project/network/peers"
	"fmt"
	"time"
)

// ------ Replacing switch in main with goroutines

func PeersUpdate(
	ch *network.Channels,
	id string,
	Masterid *string,
	ImLost *bool,
	pendingMasterOrders map[string][][2]bool,
	elevatorStates map[string]elevator.Elevator,
	backupStates map[string]elevator.Elevator,
	p peers.PeerUpdate,
	e *elevator.Elevator,
) {
	for {
		select {
		case p := <-ch.PeerUpdateCh:
			var lostElevator string = "99" // To ensure there is no master when initializing the network
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

			if len(p.Lost) > 0 {
				lostElevator = p.Lost[0]
			}
			if lostElevator == *Masterid && len(p.Peers) > 0 {
				network.MasterElection(p.Peers, id, Masterid)
			}
			for _, lostID := range p.Lost {
				if lostID == id {
					*ImLost = true
				}
				delete(pendingMasterOrders, lostID)
				delete(elevatorStates, lostID)

				for i := 0; i < config.NumFloors; i++ {
					temp := backupStates[lostID]

					e.Requests[i][elevio.BT_HallDown] = e.Requests[i][elevio.BT_HallDown] || backupStates[lostID].Requests[i][elevio.BT_HallDown]
					e.Requests[i][elevio.BT_HallUp] = e.Requests[i][elevio.BT_HallUp] || backupStates[lostID].Requests[i][elevio.BT_HallUp]
					temp.Requests[i][elevio.BT_HallDown] = false
					temp.Requests[i][elevio.BT_HallUp] = false
					backupStates[lostID] = temp
					elevatorStates[id] = *e
				}
				if id == *Masterid {
					assigner.Assigner(elevatorStates, ch.AssignTx, pendingMasterOrders)
				}
			}

			if len(p.New) > 0 {
				*ImLost = false
				if *Masterid == id {
					ch.BackupTx <- backupStates
					fmt.Println("Master sending backup")
				}
				network.MasterElection(p.Peers, id, Masterid)
				assigner.Assigner(backupStates, ch.AssignTx, pendingMasterOrders)
			}
		}
	}
}

func HandleButtonPress(
	ch *network.Channels,
	id string,
	Masterid *string,
	ImLost *bool,
	pendingMasterOrders map[string][][2]bool,
	elevatorStates map[string]elevator.Elevator,
	backupStates map[string]elevator.Elevator,
	pendingOrderRequests map[string]network.RequestMsg,
	e *elevator.Elevator,
	Motorstop *bool,
	timeout <-chan time.Time,
) {
	var obstruction = false
	var prevFloorSensor = -1

	for {
		select {
		case button := <-ch.DrvButtons:
			if button.Button == elevio.ButtonType(elevio.BT_Cab) {
				e.Requests[button.Floor][button.Button] = true
				fsm.Fsm_onRequestButtonPress(e, button.Floor, button.Button)
				elevio.SetButtonLamp(button.Button, button.Floor, true)
			} else if !*ImLost {
				if *Masterid == id {
					e.Requests[button.Floor][button.Button] = true
					elevatorStates[id] = *e
					backupStates[id] = *e
					e.Requests[button.Floor][button.Button] = true
					assigner.Assigner(elevatorStates, ch.AssignTx, pendingMasterOrders)
					runelevator.RunElev(e, pendingMasterOrders, id)
				} else {
					e.Requests[button.Floor][button.Button] = true
					request := network.RequestMsg{button.Floor, button.Button, id}
					ch.OrderTx <- request
					pendingOrderRequests[request.OrderID] = request
					fmt.Println("Added order to pendingOrders from:", request.OrderID)
				}
			}

		case floor := <-ch.DrvFloors:

			if floor != -1 && floor != prevFloorSensor {
				fsm.Fsm_onFloorArrival(e, floor)
			} else {
				prevFloorSensor = floor
			}
			// cases.HandleFloorArrival(&e, &floor, &prevFloorSensor)
			timeout = time.After(7 * time.Second)
			*Motorstop = false
			ch.PeerTxEnable <- true

		// Starts door timer if not obstructed
		case <-timer.TimerChannel:
			if !obstruction {
				fsm.Fsm_onDoorTimeout(e)
				obstruction = false
				ch.PeerTxEnable <- true
			} else {
				elevio.SetDoorOpenLamp(true)
				ch.PeerTxEnable <- false
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
		}
	}
}

func HandleAssignments(
	ch *network.Channels,
	id string,
	Masterid *string,
	pendingMasterOrders map[string][][2]bool,
	elevatorStates map[string]elevator.Elevator,
	backupStates map[string]elevator.Elevator,
	pendingOrderRequests map[string]network.RequestMsg,
	e *elevator.Elevator,
	Motorstop *bool,
	timeout <-chan time.Time,
) {
	for {
		select {
		case r := <-ch.OrderRx: //r for request
			if *Masterid == id {
				ack := network.AckMsg{OrderID: r.OrderID, AckType: "order"}
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
			ack := network.AckMsg{OrderID: id, AckType: "assign"}
			ch.AckTx <- ack
			timeout = time.After(7 * time.Second)
			runelevator.RunElev(e, AssRec, id)

		case backup := <-ch.BackupRx:
			e.Requests = backup[id].Requests

			for i := 0; i < config.NumFloors; i++ {
				if e.Requests[i][elevio.BT_Cab] {
					fsm.Fsm_onRequestButtonPress(e, i, elevio.BT_Cab)
				}
			}

		case <-timeout:
			if e.Behaviour == elevator.EB_Moving {
				ch.PeerTxEnable <- false
				*Motorstop = true
				time.Sleep(1000 * time.Millisecond)
				fmt.Println("MOTORSTOP")
			} else {
				timeout = time.After(7 * time.Second)
				fmt.Println("Timer Restarted")
			}
		}
	}
}
