# cases Package

This package handles elevator event-driven processes, including peer updates, button presses, and order assignments.

## Functions

### PeersUpdate
- Manages peer updates and detects lost elevators.
- Triggers master re-election if needed.
- Redistributes hall orders from disconnected elevators.

### HandleButtonPress
- Processes button presses and assigns orders.
- Handles obstruction detection and timeout events.
- Updates elevator state on floor arrival.

### HandleAssignments
- Manages order distribution and acknowledgments.
- Resends unacknowledged orders.
- Synchronizes elevator states with backups.

## Features
- Master election for fault tolerance.
- Efficient request handling and order assignment.
- Obstruction and motor failure management.
- Backup-based state recovery.