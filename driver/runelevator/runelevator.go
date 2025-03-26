package runelevator

import (
	"elev_project/config"
	"elev_project/driver/elevator"
	"elev_project/driver/elevio"
	"elev_project/driver/fsm"
	"fmt"
)

func RunElev(e *elevator.Elevator, assignments map[string][][2]bool, id string) {
	if assignedOrders, exists := assignments[id]; exists {
		for i := 0; i < config.NumFloors; i++ {
			e.Requests[i][elevio.BT_HallUp] = assignedOrders[i][0]
			e.Requests[i][elevio.BT_HallDown] = assignedOrders[i][1]
		}
	} else {
		fmt.Println("Warning: No assignments found for ID:", id)
	}

	for i := 0; i < config.NumFloors; i++ {
		for j := 0; j < 2; j++ {
			if e.Requests[i][elevio.ButtonType(j)] {
				fmt.Printf("Excecute order: %d\n", i)
				fsm.Fsm_onRequestButtonPress(e, i, elevio.ButtonType(j))
			}
		}
	}
}
