package requests

import (
	"Network-go/network/config"
	"Network-go/network/elevator"
	"Network-go/network/elevio"
)

type DirnBehaviourPair struct {
	Dirn              elevio.MotorDirection
	ElevatorBehaviour elevator.ElevatorBehaviour
}

// Is the call to a floor above
func RequestsAbove(e *elevator.Elevator) bool {
	for f := e.Floor + 1; f < config.NumFloors; f++ {
		for b := 0; b < config.NumButtons; b++ {
			if e.Requests[f][b] {
				return true
			}
		}
	}
	return false
}

// Is the call to a floor below
func RequestsBelow(e *elevator.Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for b := 0; b < config.NumButtons; b++ {
			if e.Requests[f][b] {
				return true
			}
		}
	}
	return false

}

// Is the call to the current floor
func RequestsHere(e *elevator.Elevator) bool {
	for b := 0; b < config.NumButtons; b++ {
		if e.Requests[e.Floor][b] {
			return true
		}
	}
	return false
}

// Function to decide direction
func RequestsChooseDirection(e *elevator.Elevator) DirnBehaviourPair {
	// fmt.Printf("%v\n",e.Dirn,
	// "\n above: ",RequestsAbove(e),"\n here: ",RequestsHere(e),"\n below: ", RequestsBelow(e))
	switch e.Dirn {
	case elevio.MD_Up:
		if RequestsAbove(e) {
			return DirnBehaviourPair{elevio.MD_Up, elevator.EB_Moving}
		} else if RequestsHere(e) {
			return DirnBehaviourPair{elevio.MD_Down, elevator.EB_DoorOpen}
		} else if RequestsBelow(e) {
			return DirnBehaviourPair{elevio.MD_Down, elevator.EB_Moving}
		} else {
			return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
		}
	case elevio.MD_Down:
		if RequestsBelow(e) {
			return DirnBehaviourPair{elevio.MD_Down, elevator.EB_Moving}
		} else if RequestsHere(e) {
			return DirnBehaviourPair{elevio.MD_Up, elevator.EB_DoorOpen}
		} else if RequestsAbove(e) {
			return DirnBehaviourPair{elevio.MD_Up, elevator.EB_Moving}
		} else {
			return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
		}

	case elevio.MD_Stop:
		if RequestsHere(e) {
			return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_DoorOpen}
		} else if RequestsAbove(e) {
			return DirnBehaviourPair{elevio.MD_Up, elevator.EB_Moving}
		} else if RequestsBelow(e) {
			return DirnBehaviourPair{elevio.MD_Down, elevator.EB_Moving}
		} else {
			return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
		}
	default:
		return DirnBehaviourPair{elevio.MD_Stop, elevator.EB_Idle}
	}
}

// Should the elevator stop at this floor?
func RequestShouldStop(e *elevator.Elevator) bool {
	switch e.Dirn {
	case elevio.MD_Down:
		return e.Requests[e.Floor][elevio.BT_HallDown] || e.Requests[e.Floor][elevio.BT_Cab] || !RequestsBelow(e)
	case elevio.MD_Up:
		return e.Requests[e.Floor][elevio.BT_HallUp] || e.Requests[e.Floor][elevio.BT_Cab] || !RequestsAbove(e)
	case elevio.MD_Stop:
		return true
	default:
		return true
	}
}

func RequestsShouldClearImmediately(e *elevator.Elevator, btn_floor int, btn_type elevio.ButtonType) bool {
	return e.Floor == btn_floor &&
		((e.Dirn == elevio.MD_Up && btn_type == elevio.BT_HallUp) ||
			(e.Dirn == elevio.MD_Down && btn_type == elevio.BT_HallDown) ||
			e.Dirn == elevio.MD_Stop ||
			btn_type == elevio.BT_Cab)
}

func RequestsClearAtCurrentFloor(e *elevator.Elevator) {
	e.Requests[e.Floor][elevio.BT_Cab] = false
	switch e.Dirn {
	case elevio.MD_Up:
		if !RequestsAbove(e) && !e.Requests[e.Floor][elevio.BT_HallUp] {
			e.Requests[e.Floor][elevio.BT_HallDown] = false
		}
		e.Requests[e.Floor][elevio.BT_HallUp] = false

	case elevio.MD_Down:
		if !RequestsBelow(e) && !e.Requests[e.Floor][elevio.BT_HallDown] {
			e.Requests[e.Floor][elevio.BT_HallUp] = false
		}
		e.Requests[e.Floor][elevio.BT_HallDown] = false
	case elevio.MD_Stop:
		fallthrough

	default:
		e.Requests[e.Floor][elevio.BT_HallUp] = false
		e.Requests[e.Floor][elevio.BT_HallDown] = false
	}
}
