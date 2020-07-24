package main

// A Go application that uses the Windows schtasks tool to run a given schedualed task.
// Doesn't exit until the schedualed task has finished, and prints out a status indicator as the task is running.
// Uses previous run times to guess the total time the shedualed task will take.

import (
	// Standard libraries.
	"os"
	"fmt"
	"time"
	"strings"
	"os/exec"
)

func runCommand (theCommandString string, theCommandArgs ...string) string {
	fmt.Println("Running: " + theCommandString, theCommandArgs)
	theCommand := exec.Command(theCommandString, theCommandArgs...)
	commandOutput, commandErr := theCommand.CombinedOutput()
	fmt.Println("Output: " + string(commandOutput))
	if commandErr != nil {
		fmt.Println("Error running command: " + theCommandString, theCommandArgs)
		fmt.Println("ERROR: " + commandErr.Error())
	} else if strings.HasSuffix(string(commandOutput), "\"Ready\"") {
		return "READY"
	} else if strings.HasSuffix(string(commandOutput), "\"Running\"") {
		return "RUNNING"
	}
	return ""
}

func main() {
	if len(os.Args) == 2 {
		startTime := time.Now().Unix()
	
		runCommand("C:\\Windows\\System32\\schtasks.exe", "/RUN", "/TN", os.Args[1])
		runState := "RUNNING"
		for runState == "RUNNING" {
			time.Sleep(4 * time.Second)
			fmt.Println("Progress: ")
			runState = runCommand("C:\\Windows\\System32\\schtasks.exe", "/QUERY", "/TN", os.Args[1], "/FO", "CSV", "/NH")
		}
		endTime := time.Now().Unix()
	
		fmt.Println(endTime - startTime)
	} else {
		fmt.Println("Usage: runScheduledTask NameOfWindowsScheduledTask")
	}
}
