# Project

This project is part of the course TTK4145 - Real Time Programming, and aims to better our understanding of RTOS (Real-Time-Operating-Systems) through creating a system for operating multiple elevators on a single network. 

The multiple elevator system is designed according to the TTK4145 - Real Time Programming project requirements.

Further descriptions of the different modules can be found in their respective `README.md` files.

## Our approach

The system is finite state machine (FSM) based, with supporting modules which contain the separated systems of the working program. All elevators broadcasts their worldview, hall-orders, and elevator state to the designated master, which acknowledges all recieved messeges, all through the UDP communication protocol.
Every elevator knows which elevators are alive on the network and what states/calls these elevators have. By doing that, every elevator should be able to resume the tasks of the other elevators should they fail to execute their tasks either through;
- Network disconnection
- Power loss
- Loss of motor function
- Door obstructions


A designated master (Port with the current lowest ID on the network) assigns the calls efficiently and distributes them on the network through the `assigner.go` file. The network of elevators work as a primary-backups, where all elevators keeps backup of all current and pending orders, keeping the flow of the system even in cases of critical failure. 

The main consists of a set of global variables (that function as the local elevators world view), elevator- and network initialization functions, and three goroutines that handles all functionality, found in `cases.go`:
```
go cases.HandlePeersUpdate (...)
go cases HandleButtonPress (...)
go cases HandleAssignments (...)
```
## To run

- Running assigner executable:
    -  Access webpage: https://github.com/TTK4145/Project-resources/releases/tag/v1.1.3.
    - Aquire the hall_request_assigner executable from the aforementioned repository in order to run the assigner function.
- Running elevator system: 
    - Access webpage: https://github.com/TTK4145/Simulator-v2 for the elevator simulator.
    - Set up elevator server, either through simulation or a physical model like the one provided by TTK4145. 
    - Run each of the elevators on the system on a set of ports with the terminal line:
        - ```simelevatorserver --port .....``` 
    - Run each local elevator program with the terminal line:
        - ```go run main.go -id="..."```
    - The ports used for the network in this iteration is found in ```elevator.go``` under ```InitializeElevator(e *Elevator, id string)``` as:
        - ```elevio.Init("localhost:1729"+id, config.NumFloors)```
        - The localhost value determines which ports should be used.

