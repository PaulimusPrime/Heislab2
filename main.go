package main

import (
	"elev_project/cases"
	"elev_project/config"
	"elev_project/driver/elevator"
	"elev_project/network"
	"flag"
	"time"
)

func main() {
	var (
		e         elevator.Elevator
		ImLost    bool
		id        string
		Motorstop bool = false

		pendingOrderRequests = make(map[string]network.RequestMsg)
		elevatorStates       = make(map[string]elevator.Elevator)
		backupStates         = make(map[string]elevator.Elevator)
		pendingMasterOrders  = make(map[string][][2]bool)
		Masterid             string
		timeout              = time.After(config.MotorTimeout)
	)

	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()
	elevator.InitializeElevator(&e, id)
	ch := network.InitChannels()
	network.StartListeners(id, ch)
	network.StartStateSender(id, ch, &e, &Motorstop)

	go cases.PeersUpdate(
		ch,
		id,
		&Masterid,
		&ImLost,
		pendingMasterOrders,
		elevatorStates,
		backupStates,
		<-ch.PeerUpdateCh,
		&e)

	go cases.HandleButtonPress(
		ch,
		id,
		&Masterid,
		&ImLost,
		pendingMasterOrders,
		elevatorStates,
		backupStates,
		pendingOrderRequests,
		&e,
		&Motorstop,
		timeout)

	go cases.HandleAssignments(
		ch,
		id,
		&Masterid,
		pendingMasterOrders,
		elevatorStates,
		backupStates,
		pendingOrderRequests,
		&e,
		&Motorstop,
		timeout)
	select {}
}
