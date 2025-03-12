package elevator

import (
	"time"
	"Network-go/network/elevio"
	"Network-go/network/config"
)

//Elevatorbehavior gives information about the state that the elevator is in.
type ElevatorBehaviour int
const (
	EB_Idle   ElevatorBehaviour = 0
	EB_DoorOpen                 = 1
	EB_Moving                   = 2
)

//Everyone waiting or only those that want to travel in that direction goes on
// type ClearRequestVariant int
// const(
// 	CV_all ClearRequestVariant = iota
// 	CV_InDirn
// )

type Elevator struct{
	Floor int
	Dirn elevio.MotorDirection
	Requests[config.NumFloors][config.NumButtons] bool
	Behaviour ElevatorBehaviour
}

func InitializeElevator (e *Elevator, prevRequestButton *[config.NumFloors][config.NumButtons]bool){
	e.Behaviour = EB_Idle
	e.Floor = elevio.GetFloor()
	e.Dirn = elevio.MD_Stop
	for f := 0; f < config.NumFloors; f++ {
		for b := 0; b < config.NumButtons; b++ {
			e.Requests[f][b] = false
			prevRequestButton[f][b] = false
			elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
			elevio.SetDoorOpenLamp(false)
		}

	}
	//For loop to always start in first floor
	for {
		e.Floor = elevio.GetFloor()
		if e.Floor != 0 {
			elevio.SetMotorDirection(elevio.MD_Down)
			time.Sleep(time.Duration(config.InputPollRate))
		} else {
			elevio.SetMotorDirection(elevio.MD_Stop)
			break
		}
	}
}