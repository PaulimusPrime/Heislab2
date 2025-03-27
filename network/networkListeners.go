package network

import (
	"elev_project/config"
	"elev_project/driver/elevator"
	"elev_project/driver/elevio"
	"elev_project/network/bcast"
	"elev_project/network/peers"
	"time"
)

type HelloMsg struct {
	Message string
	Iter    int
}

type RequestMsg struct {
	Floor   int
	Button  elevio.ButtonType
	OrderID string
}

type AckMsg struct {
	OrderID string
	AckType string
}

type ObjectMsg struct {
	Message elevator.Elevator
	ID      string
}

type Channels struct {
	DrvButtons   chan elevio.ButtonEvent
	DrvFloors    chan int
	DrvObstr     chan bool
	DrvStop      chan bool
	OrderTx      chan RequestMsg
	OrderRx      chan RequestMsg
	AckTx        chan AckMsg
	AckRx        chan AckMsg
	StateTx      chan ObjectMsg
	StateRx      chan ObjectMsg
	AssignTx     chan map[string][][2]bool
	AssignRx     chan map[string][][2]bool
	BackupTx     chan map[string]elevator.Elevator
	BackupRx     chan map[string]elevator.Elevator
	PeerUpdateCh chan peers.PeerUpdate
	PeerTxEnable chan bool
	HelloTx      chan HelloMsg
	HelloRx      chan HelloMsg
}

func InitChannels() *Channels {
	return &Channels{
		DrvButtons:   make(chan elevio.ButtonEvent),
		DrvFloors:    make(chan int),
		DrvObstr:     make(chan bool),
		DrvStop:      make(chan bool),
		OrderTx:      make(chan RequestMsg),
		OrderRx:      make(chan RequestMsg),
		AckTx:        make(chan AckMsg),
		AckRx:        make(chan AckMsg),
		StateTx:      make(chan ObjectMsg),
		StateRx:      make(chan ObjectMsg),
		AssignTx:     make(chan map[string][][2]bool),
		AssignRx:     make(chan map[string][][2]bool),
		BackupTx:     make(chan map[string]elevator.Elevator),
		BackupRx:     make(chan map[string]elevator.Elevator),
		PeerUpdateCh: make(chan peers.PeerUpdate),
		PeerTxEnable: make(chan bool),
		HelloTx:      make(chan HelloMsg),
		HelloRx:      make(chan HelloMsg),
	}
}

func StartListeners(id string, ch *Channels) {
	go peers.Transmitter(config.Peerchan, id, ch.PeerTxEnable)
	go peers.Receiver(config.Peerchan, ch.PeerUpdateCh)
	go elevio.PollButtons(ch.DrvButtons)
	go elevio.PollFloorSensor(ch.DrvFloors)
	go elevio.PollObstructionSwitch(ch.DrvObstr)
	go elevio.PollStopButton(ch.DrvStop)
	go bcast.Transmitter(config.Hellochan, ch.HelloTx)
	go bcast.Receiver(config.Hellochan, ch.HelloRx)
	go bcast.Receiver(config.Orderchan, ch.OrderRx)
	go bcast.Transmitter(config.Orderchan, ch.OrderTx)
	go bcast.Transmitter(config.Ackchan, ch.AckTx)
	go bcast.Receiver(config.Ackchan, ch.AckRx)
	go bcast.Transmitter(config.Statechan, ch.StateTx)
	go bcast.Receiver(config.Statechan, ch.StateRx)
	go bcast.Transmitter(config.Assignchan, ch.AssignTx)
	go bcast.Receiver(config.Assignchan, ch.AssignRx)
	go bcast.Transmitter(config.Backupchan, ch.BackupTx)
	go bcast.Receiver(config.Backupchan, ch.BackupRx)
}

func StartStateSender(id string, ch *Channels, e *elevator.Elevator, Motorstop *bool) {
	go func() {
		time.Sleep(time.Second)
		for {
			if !*Motorstop {
				ch.StateTx <- ObjectMsg{Message: *e, ID: id}
			}
			time.Sleep(50 * time.Millisecond)
		}
	}()

}
