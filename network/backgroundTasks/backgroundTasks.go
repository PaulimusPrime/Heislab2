package backgroundtasks

import (
	"elev_project/driver/elevator"
	"elev_project/network/networkListeners"
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