package backgroundtasks

import (
	"elev_project/assigner"
	"elev_project/driver/elevator"
	"elev_project/driver/master"
	"elev_project/network/networkListeners"
	"elev_project/network/peers"
	"fmt"
	"time"
)

func StartHelloSender(id string, ch *networkListeners.Channels) {
	helloMsg := networkListeners.HelloMsg{"Hello from " + id, 0}
	go func() {
		for {
			helloMsg.Iter++
			ch.HelloTx <- networkListeners.HelloMsg{}
			time.Sleep(1 * time.Second)
		}
	}()
}

func StartStateSender(id string, ch *networkListeners.Channels, e *elevator.Elevator) {
	go func() {
		time.Sleep(time.Second)
		for {
			ch.StateTx <- networkListeners.ObjectMsg{Message: *e, ID: id}
			time.Sleep(50 * time.Millisecond)
		}
	}()

}

// ------ Replacing switch in main with goroutines

func PeersUpdate(
	ch *networkListeners.Channels,
	id string,
	Masterid *string,
	ImLost *bool,
	pendingMasterOrders map[string][][2]bool,
	elevatorStates map[string]elevator.Elevator,
	backupStates map[string]elevator.Elevator,
	p peers.PeerUpdate,

) {
	var lostElevator string = "99" // To ensure there is no master when initializing the network
	fmt.Printf("Peer update:\n")
	fmt.Printf("  Peers:    %q\n", p.Peers)
	fmt.Printf("  New:      %q\n", p.New)
	fmt.Printf("  Lost:     %q\n", p.Lost)

	if len(p.Lost) > 0 {
		lostElevator = p.Lost[0]
	}
	if lostElevator == *Masterid && len(p.Peers) > 0 {
		master.MasterElection(p.Peers, id, Masterid)
	}
	for _, lostID := range p.Lost {
		if lostID == id {
			*ImLost = true
		}
		delete(pendingMasterOrders, lostID)
		delete(elevatorStates, lostID)
	}
	if len(p.New) > 0 {
		*ImLost = false
		if *Masterid == id {
			ch.BackupTx <- backupStates
			fmt.Println("Master sending backup")
			// for i := 0; i < config.NumFloors; i++ {
			// 	fmt.Println(backupStates["1"].Requests[i][elevio.BT_Cab])
			// 	fmt.Println(backupStates["2"].Requests[i][elevio.BT_Cab])
			// }
		}
		master.MasterElection(p.Peers, id, Masterid)
		assigner.Assigner(backupStates, ch.AssignTx, pendingMasterOrders)
	}
}
