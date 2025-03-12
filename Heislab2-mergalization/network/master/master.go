package master

import (
	"fmt"
)

/**
jdksjdksjak
**/
func MasterElection(peers []string, id string, Masterid *string) {
	fmt.Printf("Master election started\n")
	if id == peers[0] {
		fmt.Printf("I am master\n")
		*Masterid = id
		fmt.Print(*Masterid,"\n")
	} else {
		fmt.Printf("I am slave\n")
		*Masterid = peers[0]
		fmt.Print(*Masterid,"\n")
	}
}
