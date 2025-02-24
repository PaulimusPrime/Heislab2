package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
	//"transmit/UDP"
)

var backup_inc bool

func main() {
	print("\n--- Backup phase ---\n")
	file, err := os.Open("data.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	// --- Creating backup terminal, if backup doesn't exist ---
	for {
		backupChecking()
		if backup_inc {
			backupCreation()
			fmt.Printf(".. timed out\n")
			backup_inc = false
			fmt.Printf("\n-- Primary phase --\n")
			counter()
		}
	}
}

func counter() {
	num, err := os.ReadFile("data.txt")
	if err != nil {
		log.Fatal(err)
	}
	str := string(num)
	counter, err1 := strconv.Atoi(str)
	if err1 != nil {
		log.Fatal(err1)
	}
	currData := counter
	for i := currData; i < currData+5; i++ {
		fmt.Printf(": %v\n", i)
		err := os.WriteFile("data.txt", []byte(strconv.Itoa(i)), 0666)

		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(2 * time.Second)
	}
}

func backupCreation() {
	killMaster()
	cmd := exec.Command("gnome-terminal", "--", "go", "run", "main.go")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Fatal error\n")
	}
}

func backupChecking() {
	for {
		// ---------------------
		data, err := os.ReadFile("data.txt")
		if err != nil {
			log.Fatal(err)
		}
		prev := string(data)
		time.Sleep(2 * time.Second)
		data, err = os.ReadFile("data.txt")
		if err != nil {
			log.Fatal(err)
		}
		curr := string(data)
		// ---------------------

		if curr == prev {
			print("Master stopped updating\n")
			backup_inc = true
			break
		}
		fmt.Println("data: " + string(data))
	}
}

func killMaster() {
	pid, err := os.ReadFile("pid.txt")
	if err != nil {
		log.Fatal(err)
	}
	str := string(pid)
	pid_int, err1 := strconv.Atoi(str)
	if err1 != nil {
		log.Fatal(err1)
	}
	syscall.Kill(pid_int, syscall.SIGTERM)
	// 	//write new master PID to file
	// 	// Write PID of current process to file so that it can be killed by another motherfucker
	new_pid := os.Getpid()
	err2 := os.WriteFile("pid.txt", []byte(strconv.Itoa(new_pid)), 0666)
	if err2 != nil {
		log.Fatal(err2)
	}
}
