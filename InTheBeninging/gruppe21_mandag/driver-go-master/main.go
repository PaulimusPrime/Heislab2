package main

import (
	"Driver-go/config"
	"Driver-go/elevator"
	"Driver-go/elevio"
	"Driver-go/fsm"
	"Driver-go/masterfunctions"
	"Driver-go/timer"
	"fmt"
	"time"
)

func main() {

	elevio.Init("localhost:10000", config.NumFloors)

	var (
		e                 elevator.Elevator
		prevRequestButton [config.NumFloors][config.NumButtons]bool
		prevFloorSensor   = -1
		obstruction       bool
	)

	//initializing elevator
	fmt.Printf("Elevator starting \n")
	elevator.InitializeElevator(&e, &prevRequestButton)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	for {
		select {
		case button := <-drv_buttons:
			e.Requests[button.Floor][button.Button] = true
			elevio.SetButtonLamp(button.Button, button.Floor, true)
			fsm.Fsm_onRequestButtonPress(&e, button.Floor, button.Button)

		case floor := <-drv_floors:
			if floor != -1 && floor != prevFloorSensor {
				fsm.Fsm_onFloorArrival(&e, floor)
			} else {
				prevFloorSensor = floor
			}

		case <-timer.TimerChannel:
			if !obstruction {
				fsm.Fsm_onDoorTimeout(&e)
				obstruction = false
			} else {
				timer.StartTimer(config.ObstructionDurationS)
			}

		case <-drv_obstr:
			if e.Behaviour == elevator.EB_DoorOpen {
				elevio.SetDoorOpenLamp(true)
				obstruction = !obstruction
			}
		case stop := <-drv_stop:
			if stop {
				elevio.SetMotorDirection(elevio.MD_Stop)
				e.Behaviour = elevator.EB_Idle
			}
		}

		//masterfunctions.BroadcastMasterID(10000)
		//masterfunctions.BroadcastMasterID(10001)
		masterfunctions.ListenForMaster()

		// if e.ID==1{

		// }else {
		// 	master_functions.ListenForMaster()
		// }

		time.Sleep(time.Duration(config.InputPollRate))
	}
}
