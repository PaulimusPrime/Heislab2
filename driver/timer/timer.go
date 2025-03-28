package timer

import (
	"time"
)

var TimerChannel = make(chan bool)

func StartTimer(duration time.Duration) {
	go func() {
		time.Sleep(duration)
		TimerChannel <- true
	}()
}
