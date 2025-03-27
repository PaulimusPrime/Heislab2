package systemstate

import (
	"elev_project/driver/elevator"
	"elev_project/network/networkListeners"
)

type SystemState struct {
	MasterID             string
	ImLost               bool
	PendingMasterOrders  map[string]*[][2]bool
	ElevatorStates       map[string]*elevator.Elevator
	BackupStates         map[string]*elevator.Elevator
	PendingOrderRequests map[string]*networkListeners.RequestMsg
	Elevator             elevator.Elevator
	Motorstop            bool
}
