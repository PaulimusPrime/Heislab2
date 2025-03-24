package fsm

import (
	"elev_project/config"
	"elev_project/driver/elevator"
	"elev_project/driver/elevio"
	"elev_project/driver/requests"
	"elev_project/driver/timer"
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

//Code from sunday 23 by Johannes

//	func SetHallLights(HallCalls map[string][][2]bool) {
//		for i := 1; i < len(HallCalls)+1; i++ {
//			for f := 0; f < config.NumFloors; f++ {
//				elevio.SetButtonLamp(elevio.BT_HallUp, f, HallCalls[strconv.Itoa(i)][f][0])
//				elevio.SetButtonLamp(elevio.BT_HallDown, f, HallCalls[strconv.Itoa(i)][f][1])
//				println("id,", i, HallCalls[strconv.Itoa(i)][f][0])
//				println("id", i, HallCalls[strconv.Itoa(i)][f][1])
//			}
//		}
//	}
func SetHallLights(elevatorStates map[string]elevator.Elevator) {
	// for i := 1; i < len(HallCalls)+1; i++ {
	// 	for f := 0; f < config.NumFloors; f++ {
	// 		elevio.SetButtonLamp(elevio.BT_HallUp, f, HallCalls[strconv.Itoa(i)][f][0])
	// 		elevio.SetButtonLamp(elevio.BT_HallDown, f, HallCalls[strconv.Itoa(i)][f][1])
	// 		println("id,", i, HallCalls[strconv.Itoa(i)][f][0])
	// 		println("id", i, HallCalls[strconv.Itoa(i)][f][1])
	// 	}
	// }
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

//Code end from sunday 23

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

	//SetAllLights(e)
	SetCabLights(e)
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
			//SetAllLights(e)
			SetCabLights(e)
			e.Behaviour = elevator.EB_DoorOpen
		}
	case elevator.EB_DoorOpen:
		elevio.SetDoorOpenLamp(true)
		requests.RequestsClearAtCurrentFloor(e)
		timer.StartTimer(config.DoorOpenDurationS)
		//SetAllLights(e)
		SetCabLights(e)
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
			//SetAllLights(e)
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
