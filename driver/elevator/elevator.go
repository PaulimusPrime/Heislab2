package elevator

import (
	"elev_project/config"
	"elev_project/driver/elevio"
	"time"
)

// Elevatorbehavior gives information about the state that the elevator is in.
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
	//For loop to always start in a defined state
	for {
		e.Floor = elevio.GetFloor()
		if e.Floor == -1 {
			elevio.SetMotorDirection(elevio.MD_Down)
			time.Sleep(time.Duration(config.InputPollRate))
		} else {
			elevio.SetMotorDirection(elevio.MD_Stop)
			//time.Sleep(time.Second)
			break
		}
	}
}
