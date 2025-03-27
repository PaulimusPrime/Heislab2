# fsm Package

Manages the elevator finite state machine and light handling.

## Functions

- **SetAllLights(e *elevator.Elevator)**  
  Updates all button lamps based on the elevator's current request matrix.

- **SetHallLights(elevatorStates map[string]elevator.Elevator)**  
  Aggregates hall requests from all elevators and sets the hall up/down lamps for each floor.

- **SetCabLights(e *elevator.Elevator)**  
  Updates the cab call lights based on the elevator's current requests.

- **Fsm_onInitBetweenFloors(e *elevator.Elevator)**  
  Sets the motor direction to down and marks the elevator as moving when initializing between floors.

- **Fsm_onRequestButtonPress(e *elevator.Elevator, btn_floor int, btn_type elevio.ButtonType)**  
  Handles a request button press. Depending on the current state:
  - If the door is open and the request should be cleared immediately, restarts the door timer.
  - Otherwise, registers the request, determines the next direction and state, and updates the cab lights.

- **Fsm_onFloorArrival(e *elevator.Elevator, newFloor int)**  
  Processes arrival at a new floor:
  - Updates the current floor and floor indicator.
  - If moving and a stop is required, stops the elevator, opens the door, clears requests, starts the door timer, and updates the cab lights.
  - If the door is already open, clears requests and restarts the door timer.

- **Fsm_onDoorTimeout(e *elevator.Elevator)**  
  Handles door timeout events by choosing the next direction, updating the state, clearing current floor requests, and either restarting the door timer (if remaining door open) or setting the motor direction.
