package cases

import (
	"elev_project/assigner"
	"elev_project/config"
	"elev_project/driver/elevator"
	"elev_project/driver/elevio"
	"elev_project/driver/fsm"
	"elev_project/driver/master"
	"elev_project/driver/runelevator"
	"elev_project/driver/timer"
	"elev_project/network/networkListeners"
	"elev_project/network/peers"
	"fmt"
	"time"
)

// ------ Replacing switch in main with goroutines

func PeersUpdate(
	ch *networkListeners.Channels,
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
				master.MasterElection(p.Peers, id, Masterid)
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
				master.MasterElection(p.Peers, id, Masterid)
				assigner.Assigner(backupStates, ch.AssignTx, pendingMasterOrders)
			}
		}
	}
}

func HandleButtonPress(
	ch *networkListeners.Channels,
	id string,
	Masterid *string,
	ImLost *bool,
	pendingMasterOrders map[string][][2]bool,
	elevatorStates map[string]elevator.Elevator,
	backupStates map[string]elevator.Elevator,
	pendingOrderRequests map[string]networkListeners.RequestMsg,
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
					request := networkListeners.RequestMsg{button.Floor, button.Button, id}
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
		}
	}
}

func HandleFloorArrival(
	e *elevator.Elevator,
	prevFloorSensor *int,
	floor *int,
) {
	if *floor != -1 && floor != prevFloorSensor {
		fsm.Fsm_onFloorArrival(e, *floor)
	} else {
		prevFloorSensor = floor
	}
}
