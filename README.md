Elevator-Project - TTK4145 - Real-Time Programming
==================================================

The current iteration requires access to `simelevatorserver` and is run by passing the following code in the terminal:
```
simelevatorserver --port 1729+id
```
Where `id` is its own value, specified when running the program using the following lines in a seperate terminal window:
```
cd Heislab2-mergalization
go run main.go -id=...
```
The actively running loop is located in [`main.go`](main.go). 
All helper functions can be found in the sorted folders under [`network`](network)
________

Project description
---------------
The following project aims to work as a complete algorithm for an elevator network, consisting of $n$ elevators and $m$ floors.

To achieve effective communication between the $n$ elevators, we have implemented a solution using UDP broadcasting on a local network. 
When an elevator is added to the network (i.e. running the terminal lines provided above) 
it first checks if any other elevator is broadcasting a master heartbeat. If so it notifies the network of its arrival, and further serves as a slave.
If however there isn't a master heartbeat, the new elevator starts broadcasting and resumes its task applying this new role.
If at any point the master should be removed from the network, a backup order-que is stored in each of the elevators, and a new master is assigned based on the `id` value.

When an order is recieved in either of the $m$ floors either of the $n$ elevators, said elevator broadcasts its order on a given channel. 
The master listens for such orders, and affirmes the recieval of any order before it assigns the order-handling to the most optimal elevator using a cost_function.

This image illustrates early development planning:
<img width="821" alt="Sequence" src="https://github.com/user-attachments/assets/2d16bf8d-7345-42fa-b360-561b9bcdcc23" />

________
