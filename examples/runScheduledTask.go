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
	"io/ioutil"
)

func runCommand (theCommandString string, theCommandArgs ...string) string {
	theCommand := exec.Command(theCommandString, theCommandArgs...)
	commandOutput, commandErr := theCommand.CombinedOutput()
	commandOutputString := strings.TrimSpace(string(commandOutput))
	if commandErr != nil {
		fmt.Println("Error running command: " + theCommandString, theCommandArgs)
		fmt.Println("ERROR: " + commandErr.Error())
	} else if strings.HasSuffix(commandOutputString, "\"Ready\"") {
		return "READY"
	} else if strings.HasSuffix(commandOutputString, "\"Running\"") {
		return "RUNNING"
	}
	return ""
}

func main() {
	if len(os.Args) == 2 {
		var runTimes []int
		runTimesBytes, fileErr := ioutil.ReadFile("runScheduledTask.txt")
		if fileErr == nil {
			for pl, runTimeString := range strings.Split(string(runTimesBytes), "\n") {
				runTimes = append(runTimes, int(runTimeString))
			}
		}
		println(runTimes)
		
		startTime := time.Now().Unix()
		
		fmt.Println("Running \"" + os.Args[1] + "\"...")
		runCommand("C:\\Windows\\System32\\schtasks.exe", "/RUN", "/TN", os.Args[1])
		runState := "RUNNING"
		for runState == "RUNNING" {
			time.Sleep(4 * time.Second)
			fmt.Println("Progress: ")
			runState = runCommand("C:\\Windows\\System32\\schtasks.exe", "/QUERY", "/TN", os.Args[1], "/FO", "CSV", "/NH")
		}
		endTime := time.Now().Unix()
	
		runTime := endTime - startTime
		fmt.Println("Done - runtime " + string(runTime) + " seconds.")
	} else {
		fmt.Println("Usage: runScheduledTask NameOfWindowsScheduledTask")
	}
}
