package assigner

import (
	"Network-go/network/elevator"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
)

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

// This function will assign a new call to the network
func Assigner(e1 elevator.Elevator, e2 elevator.Elevator, e3 elevator.Elevator) {
	hraExecutable := ""
	switch runtime.GOOS {
	case "linux":
		hraExecutable = "hall_request_assigner"
	case "windows":
		hraExecutable = "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}

	input := HRAInput{
		//HallRequests: HallRequests,
		//HallRequests: [][2]bool{{false, false}, {true, false}, {false, false}, {false, true}},
		HallRequests: [][2]bool{{e1.Requests[0][0]||e2.Requests[0][0]||e3.Requests[0][0], false}, {e1.Requests[1][0]||e2.Requests[1][0]||e3.Requests[1][0], e1.Requests[1][1]||e2.Requests[1][1]||e3.Requests[1][1]}, {e1.Requests[2][0]||e2.Requests[2][0]||e3.Requests[2][0], e1.Requests[2][1]||e2.Requests[2][1]||e3.Requests[2][1]}, {false, e1.Requests[3][1]||e2.Requests[3][1]||e3.Requests[3][1]}},
		States: map[string]HRAElevState{
			"one": HRAElevState{
				Behavior:    e1.Behaviour.String(),
				Floor:       e1.Floor,
				Direction:   e1.Dirn.String(),
				CabRequests: []bool{e1.Requests[0][2], e1.Requests[1][2], e1.Requests[2][2], e1.Requests[3][2]},
			},
			"two": HRAElevState{
				Behavior:    e2.Behaviour.String(),
				Floor:       e2.Floor,
				Direction:   e2.Dirn.String(),
				CabRequests: []bool{e2.Requests[0][2], e2.Requests[1][2], e2.Requests[2][2], e2.Requests[3][2]},
			},
			"three": HRAElevState{
				Behavior:    e3.Behaviour.String(),
				Floor:       e3.Floor,
				Direction:   e3.Dirn.String(),
				CabRequests: []bool{e3.Requests[0][2], e3.Requests[1][2], e3.Requests[2][2], e3.Requests[3][2]},
			},
		},
	}

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
		return
	}

	ret, err := exec.Command("./"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
		return
	}

	output := new(map[string][][2]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		return
	}

	fmt.Printf("output: \n")
	for k, v := range *output {
		fmt.Printf("%6v :  %+v\n", k, v)
	}

}
