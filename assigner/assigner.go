package assigner

import (
	"elev_project/config"
	"elev_project/driver/elevator"
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

func Assigner(elevatorStates map[string]elevator.Elevator, assignTx chan<- map[string][][2]bool, pendingMasterOrders map[string][][2]bool) {
	hraExecutable := ""
	switch runtime.GOOS {
	case "linux":
		hraExecutable = "hall_request_assigner"
	case "windows":
		hraExecutable = "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}

	elevators := []elevator.Elevator{}
	for _, e := range elevatorStates {
		elevators = append(elevators, e)
	}

	HallRequests := make([][2]bool, config.NumFloors)

	for floor := 0; floor < config.NumFloors; floor++ {
		var upRequest, downRequest bool
		for _, e := range elevators {
			upRequest = upRequest || e.Requests[floor][0]
			downRequest = downRequest || e.Requests[floor][1]
		}
		HallRequests[floor] = [2]bool{upRequest, downRequest}
	}


	States := make(map[string]HRAElevState)
	for id, e := range elevatorStates { 
		cabRequests := make([]bool, config.NumFloors)
		for floor := 0; floor < config.NumFloors; floor++ {
			cabRequests[floor] = e.Requests[floor][2]
		}

		States[id] = HRAElevState{
			Behavior:    e.Behaviour.String(),
			Floor:       e.Floor,
			Direction:   e.Dirn.String(),
			CabRequests: cabRequests,
		}
	}

	input := HRAInput{
		HallRequests: HallRequests,
		States:       States,
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

	assignTx <- *output
	for id, assignments := range *output {
		pendingMasterOrders[id] = assignments
	}
}
