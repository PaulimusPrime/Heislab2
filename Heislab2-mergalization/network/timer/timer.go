package timer

import (
	// "fmt"
	// "Driver-go/fsm"
	//"Driver-go/elevio"
	//"fmt"
	//"Driver-go/elevio"
	"time"
)

// getWallTime returns the current wall-clock time in seconds as a float64.
// func GetWallTime() float64 {
// 	return float64(time.Now().UnixNano()) / 1e9
// }

// var (
// 	TimerEndTime float64
// 	TimerActive  bool
// )

// // TimerStart initializes the timer with a given duration in seconds.
// func TimerStart(duration float64) {
// 	TimerEndTime = GetWallTime() + duration
// 	TimerActive = true
// }

// TimerStop deactivates the timer.
// func TimerStop() {
// 	TimerActive = false
// }

// // TimerTimedOut checks if the timer has expired.
// func TimerTimedOut(obstruction bool) bool {
// 	return TimerActive && GetWallTime() > TimerEndTime && !obstruction
// }

// func Timer(receiver chan<- bool) {
// 	fmt.Printf("Start sleeping \n ")
// 	time.Sleep(config.DoorOpenDurationS)
// 	fmt.Printf("Done sleeping \n")
// 	receiver <- true
// }

// func Timer(receiver chan<- bool,obstruction bool) {
// 	prev := false
// 	for {
// 		time.Sleep(config.InputPollRate)
// 		v := TimerTimedOut(obstruction)
// 		if v != prev {
// 			receiver <- v
// 		}
// 		prev = v
// 	}
// }

var TimerChannel = make(chan bool)

func StartTimer(duration time.Duration) {
	go func() {
		time.Sleep(duration)
		TimerChannel <- true
	}()
}






