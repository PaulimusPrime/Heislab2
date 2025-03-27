# assigner Package

Generates hall request assignments based on the current state of all elevators using an external executable. Further distributes this assignment, while adding the assignment to `pendingMasterOrders`. These orders must be acknowledged by the reciever in order to be removed from `pendingMasterOrders`.

## Data Structures

- **HRAElevState**: Represents an elevator's state for assignment purposes.  
  - *Fields*: `Behavior` (string), `Floor` (int), `Direction` (string), `CabRequests` ([]bool).

- **HRAInput**: Aggregates hall requests and states from all elevators.  
  - *Fields*: `HallRequests` ([][2]bool), `States` (map[string]HRAElevState).

## Functionality

**Assigner(elevatorStates, assignTx, pendingMasterOrders)**  
- Determines the appropriate external executable based on the OS.  
- Computes aggregated hall requests from all elevators.  
- Constructs a JSON input (`HRAInput`) with hall requests and elevator states.  
- Executes the external hall request assigner, parses the JSON output, and forwards assignments via `assignTx`.  
- Updates `pendingMasterOrders` with the returned assignments.


## Usage

Invoke the `Assigner` function with the current elevator states, a channel for distributing assignments, and a map for pending, un-acknowledged orders.
