# elevator Package

Manages elevator state and initialization.

## Data Structures

- **ElevatorBehaviour**  
  Represents the elevator's current state:
  - `idle`  
  - `doorOpen`  
  - `moving`

- **Elevator**  
  Contains the current floor, motor direction, request matrix, and behavior state.

## Functionality

- `InitializeElevator(e *Elevator, id string)`
  Initializes the elevator:
  - Connects to the elevator server.
  - Sets initial state (idle, floor, motor direction).
  - Resets request buttons and door lamp.
  - Moves elevator to a defined starting floor.

