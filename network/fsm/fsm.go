package fsm

import (
	"Network-go/network/config"
	"Network-go/network/elevator"
	"Network-go/network/elevio"
	"Network-go/network/requests"
	"Network-go/network/timer"
)

//static Elevator             elevator
//static ElevOutputDevice     outputDevice

// Her setter jeg bare alle lysene på, men burde blitt sendt til en request light som dealer med lysene
func SetAllLights(e *elevator.Elevator) {
	for f := 0; f < config.NumFloors; f++ {
		for b := 0; b < config.NumButtons; b++ {
			elevio.SetButtonLamp(elevio.ButtonType(b), f, e.Requests[f][b])
		}
	}
}

func Fsm_onInitBetweenFloors(e *elevator.Elevator) {
	elevio.SetMotorDirection(elevio.MD_Down)
	e.Dirn = elevio.MD_Down
	e.Behaviour = elevator.EB_Moving
}

// Alt likt som C bortsett fra printing og at jeg sender elevator e som en parameter til funksjonen
// Istedenfor å bruke en outputdevice sender jeg rett til elevio.setmotordirection
func Fsm_onRequestButtonPress(e *elevator.Elevator, btn_floor int, btn_type elevio.ButtonType) {
	switch e.Behaviour {
	case elevator.EB_DoorOpen:
		// fmt.Printf("%v\n",requests.RequestsShouldClearImmediately(e, btn_floor, btn_type))
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

	SetAllLights(e)
}

// No printing and passing elevator as argument but otherwise the same as in c
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
			SetAllLights(e)
			e.Behaviour = elevator.EB_DoorOpen
		}
	case elevator.EB_DoorOpen:
		elevio.SetDoorOpenLamp(true)
		requests.RequestsClearAtCurrentFloor(e)
		timer.StartTimer(config.DoorOpenDurationS)
		SetAllLights(e)
		e.Behaviour = elevator.EB_DoorOpen
	default:
		break
	}
}

// Same as C without the prints
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
			SetAllLights(e)
		case elevator.EB_Moving:
			fallthrough
		case elevator.EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(e.Dirn)
		}
	default:
	}
}
