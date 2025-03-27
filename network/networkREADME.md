# 1 network, order Packages

Handles inter-elevator communication, peer management, and state synchronization.

## Data Structures

### Messages

- **HelloMsg**: Peer greeting with iteration count.
- **RequestMsg**: Order request with button type and order ID.
- **AckMsg**: Multiple acknowledgment with order ID and type.
- **ObjectMsg**: Elevator state update with an ID.

### Channels

- **Driver Inputs**: `DrvButtons`, `DrvFloors`, `DrvObstr`, `DrvStop`
- **Orders**: `OrderTx`, `OrderRx`
- **Acknowledgments**: `AckTx`, `AckRx`
- **State Updates**: `StateTx`, `StateRx`
- **Assignments**: `AssignTx`, `AssignRx`
- **Backups**: `BackupTx`, `BackupRx`
- **Peer Communication**: `PeerUpdateCh`, `PeerTxEnable`, `HelloTx`, `HelloRx`

## Functions

- **InitChannels()** → Initializes all channels.
- **StartListeners(id, ch)** → Starts peer updates, event polling, and network broadcasting.
- **StartStateSender(id, ch, e, Motorstop)** → After a 1-second delay, continuously broadcasts the current elevator state to `StateTx` every 50ms if the motor is running.
- **MasterElection(peers []string, id string, Masterid *string)** → Performs a simple master election:


---

# 3 bcast, conn, peers Packages

*Provided by TTK4145*

Network module for Go (UDP broadcast only)
==========================================

See [`main.go`](main.go) for usage example. The code is runnable with just `go run main.go`

Add these lines to your `go.mod` file:
```
require Network-go v0.0.0
replace Network-go => ./Network-go
```
Where `./Network-go` is the relative path to this folder, after you have downloaded it.



