package main

// A Go application that uses the Windows schtasks tool to run a given schedualed task.
// Doesn't exit until the schedualed task has finished, and prints out a status indicator as the task is running.
// Uses previous run times to guess the total time the shedualed task will take.

import (
	// Standard libraries.
	"time"
	"strings"
	"os/exec"
)

func runCommand (theCommandString string, theCommandArgs []string) string {
	theCommand := exec.Command(theCommandString, theCommandArgs...)
	commandOutput, commandErr := theCommand.CombinedOutput()
	if commandErr != nil {
		println(commandErr.Error())
	} else if strings.HasSuffix(string(commandOutput), "\"Ready\"") {
		return "READY"
	} else if strings.HasSuffix(string(commandOutput), "\"Running\"") {
		return "RUNNING"
	}
	return ""
}

func main() {
	startTime := time.Now().Unix()
	
	runCommand("C:\\Windows\\System32\\schtasks.exe", "/RUN", "/TN", "Salamander - Diary")
	runState := "RUNNING"
	for runState == "RUNNING" {
		time.Sleep(4 * time.Second)
		println("Progress: ")
		runState = runCommand("C:\Windows\System32\schtasks.exe", "/QUERY", "/TN", "Salamander - Diary", "/FO", "CSV", "/NH")
	}
	endTime := time.Now().Unix()
	
	println(endTime - startTime)
}
