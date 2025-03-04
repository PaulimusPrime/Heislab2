package config

import (
	"time"
)

const(
NumFloors int = 4
NumButtons int =3
DoorOpenDurationS = 3*time.Second
ObstructionDurationS = 1* time.Second
InputPollRate = 20 * time.Millisecond
)

