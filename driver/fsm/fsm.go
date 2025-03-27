package fsm

import (
	"elev_project/config"
	"elev_project/driver/elevator"
	"elev_project/driver/elevio"
	"elev_project/driver/requests"
	"elev_project/driver/timer"
)

func SetAllLights(e *elevator.Elevator) {
	for f := 0; f < config.NumFloors; f++ {
		for b := 0; b < config.NumButtons; b++ {
			elevio.SetButtonLamp(elevio.ButtonType(b), f, e.Requests[f][b])
		}
	}
}

func SetHallLights(elevatorStates map[string]elevator.Elevator) {
	elevators := []elevator.Elevator{}
	for _, e := range elevatorStates {
		elevators = append(elevators, e)
	}
	for floor := 0; floor < config.NumFloors; floor++ {
		var upRequest, downRequest bool
		for _, e := range elevators {
			upRequest = upRequest || e.Requests[floor][0]
			downRequest = downRequest || e.Requests[floor][1]
		}
		elevio.SetButtonLamp(elevio.BT_HallUp, floor, upRequest)
		elevio.SetButtonLamp(elevio.BT_HallDown, floor, downRequest)
	}
}

func SetCabLights(e *elevator.Elevator) {
	for f := 0; f < config.NumFloors; f++ {
		elevio.SetButtonLamp(elevio.BT_Cab, f, e.Requests[f][elevio.BT_Cab])
	}

}

func Fsm_onInitBetweenFloors(e *elevator.Elevator) {
	elevio.SetMotorDirection(elevio.MD_Down)
	e.Dirn = elevio.MD_Down
	e.Behaviour = elevator.EB_Moving
}

func Fsm_onRequestButtonPress(e *elevator.Elevator, btn_floor int, btn_type elevio.ButtonType) {
	switch e.Behaviour {
	case elevator.EB_DoorOpen:
		if requests.RequestsShouldClearImmediately(e, btn_floor, btn_type) {
			timer.StartTimer(config.DoorOpenDurationS)

		} else {
			e.Requests[btn_floor][btn_type] = true
		}

	case elevator.EB_Moving:
		e.Requests[btn_floor][btn_type] = true

	case elevator.EB_Idle:
		e.Requests[btn_floor][btn_type] = true
		pair := requests.RequestsChooseDirection(e)
		e.Dirn = pair.Dirn
		e.Behaviour = pair.ElevatorBehaviour
		switch pair.ElevatorBehaviour {
		case elevator.EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			timer.StartTimer(config.DoorOpenDurationS)
			requests.RequestsClearAtCurrentFloor(e)

		case elevator.EB_Moving:
			elevio.SetMotorDirection(e.Dirn)

		case elevator.EB_Idle:
			fallthrough
		default:
			requests.RequestsClearAtCurrentFloor(e)
		}
	}
	SetCabLights(e)
}

func Fsm_onFloorArrival(e *elevator.Elevator, newFloor int) {
	e.Floor = newFloor
	elevio.SetFloorIndicator(e.Floor)
	switch e.Behaviour {
	case elevator.EB_Moving:
		if requests.RequestShouldStop(e) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			requests.RequestsClearAtCurrentFloor(e)
			timer.StartTimer(config.DoorOpenDurationS)
			SetCabLights(e)
			e.Behaviour = elevator.EB_DoorOpen
		}
	case elevator.EB_DoorOpen:
		elevio.SetDoorOpenLamp(true)
		requests.RequestsClearAtCurrentFloor(e)
		timer.StartTimer(config.DoorOpenDurationS)
		SetCabLights(e)
		e.Behaviour = elevator.EB_DoorOpen
	default:
		break
	}
}

func Fsm_onDoorTimeout(e *elevator.Elevator) {
	switch e.Behaviour {
	case elevator.EB_DoorOpen:
		var pair requests.DirnBehaviourPair = requests.RequestsChooseDirection(e)
		e.Dirn = pair.Dirn
		e.Behaviour = pair.ElevatorBehaviour
		switch e.Behaviour {
		case elevator.EB_DoorOpen:
			timer.StartTimer(config.DoorOpenDurationS)
			requests.RequestsClearAtCurrentFloor(e)
			SetCabLights(e)
		case elevator.EB_Moving:
			fallthrough
		case elevator.EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(e.Dirn)
		}
	default:
	}
}
