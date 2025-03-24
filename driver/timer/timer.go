package timer

import (
	// "fmt"
	// "Driver-go/fsm"
	//"Driver-go/elevio"
	//"fmt"
	//"Driver-go/elevio"
	"time"
)

var TimerChannel = make(chan bool)

func StartTimer(duration time.Duration) {
	go func() {
		time.Sleep(duration)
		TimerChannel <- true
	}()
}
