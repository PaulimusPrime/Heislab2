package elevator

import (
	"elev_project/config"
	"elev_project/driver/elevio"
	"time"
)

type ElevatorBehaviour int

const (
	EB_Idle     ElevatorBehaviour = 0
	EB_DoorOpen                   = 1
	EB_Moving                     = 2
)

func (b ElevatorBehaviour) String() string {
	return [...]string{"idle", "doorOpen", "moving"}[b]
}

type Elevator struct {
	Floor     int
	Dirn      elevio.MotorDirection
	Requests  [config.NumFloors][config.NumButtons]bool
	Behaviour ElevatorBehaviour
}

func InitializeElevator(e *Elevator, id string) {
	elevio.Init("localhost:1729"+id, config.NumFloors)
	e.Behaviour = EB_Idle
	e.Floor = elevio.GetFloor()
	e.Dirn = elevio.MD_Stop
	for f := 0; f < config.NumFloors; f++ {
		for b := 0; b < config.NumButtons; b++ {
			e.Requests[f][b] = false
			elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
			elevio.SetDoorOpenLamp(false)
		}

	}

	for {
		e.Floor = elevio.GetFloor()
		if e.Floor == -1 {
			elevio.SetMotorDirection(elevio.MD_Down)
			time.Sleep(time.Duration(config.InputPollRate))
		} else {
			elevio.SetMotorDirection(elevio.MD_Stop)
			time.Sleep(time.Second)
			break
		}
	}
}

// func RunElev(e *Elevator, assignments map[string][][2]bool, id string) {
// 	if assignedOrders, exists := assignments[id]; exists {
// 		for i := 0; i < config.NumFloors; i++ {
// 			e.Requests[i][elevio.BT_HallUp] = assignedOrders[i][0]
// 			e.Requests[i][elevio.BT_HallDown] = assignedOrders[i][1]
// 		}
// 	} else {
// 		fmt.Println("Warning: No assignments found for ID:", id)
// 	}

// 	for i := 0; i < config.NumFloors; i++ {
// 		for j := 0; j < 2; j++ {
// 			if e.Requests[i][elevio.ButtonType(j)] {
// 				fmt.Printf("Excecute order: %d\n", i)
// 				fsm.Fsm_onRequestButtonPress(e, i, elevio.ButtonType(j))
// 			}
// 		}
// 	}
// }
