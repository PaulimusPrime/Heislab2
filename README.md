# Project

This project is part of the course TTK4145 real time programming. 
And aims to better our understanding of RTOS through creating a system for operating multiple elevators on a network. 

Every module contains a seperate README which contains information about the module. The multiple elevator system is designed according to the TTK4145 project description.

## Our approach

The system is finite state machine (FSM) based, with supporting modules which contain the separated systems of the working program. 
Every elevator knows which elevators are alive on the network and what states/calls these elevators have. 


A designated master assigns the calls efficiently and distributes them on the network. The network of elevators work as Primary and backups, keeping the flow of the system, even in cases of critical failure. 

## To run

Access webpage: https://github.com/TTK4145/Project-resources/releases/tag/v1.1.3
Aquire the hall_request_assigner executable.

