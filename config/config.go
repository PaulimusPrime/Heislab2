package config

import (
	"time"
)

const (
	NumFloors            int = 4
	NumButtons           int = 3
	DoorOpenDurationS        = 3 * time.Second
	ObstructionDurationS     = 1 * time.Second
	InputPollRate            = 20 * time.Millisecond
	Assignchan           int = 11901
	Orderchan            int = 16571
	Ackchan              int = 16572
	Statechan            int = 26573
	Backupchan           int = 56438
	Hellochan            int = 61569
	Peerchan             int = 64715
)

// 16572,16571,64715,61569,26573,11901,56438
