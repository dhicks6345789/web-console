package main

// A Go application that uses the Windows schtasks tool to run a given schedualed task.
// Doesn't exit until the schedualed task has finished, and prints out a status indicator as the task is running.
// Uses previous run times to guess the total time the shedualed task will take.

import (
	// Standard libraries.
	"os"
	"fmt"
	"time"
	"sort"
	"strings"
	"strconv"
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
		var runTimes []int64
		runTimesBytes, fileErr := ioutil.ReadFile("runScheduledTask.txt")
		if fileErr == nil {
			runTimeSplit := strings.Split(string(runTimesBytes), "\n")
			for pl := 0; pl < len(runTimeSplit); pl = pl + 1 {
				runTimeVal, runTimeErr := strconv.Atoi(runTimeSplit[pl])
				if runTimeErr == nil {
					runTimes = append(runTimes, int64(runTimeVal))
				}
			}
		}
		
		var totalRunTime int64
		totalRunTime = 0
		for pl := 0; pl < len(runTimes); pl = pl + 1 {
			totalRunTime = totalRunTime + runTimes[pl]
		}
		runTimeGuess := totalRunTime / int64(len(runTimes))
		
		startTime := time.Now().Unix()
		
		fmt.Println("Running \"" + os.Args[1] + "\"...")
		runCommand("C:\\Windows\\System32\\schtasks.exe", "/RUN", "/TN", os.Args[1])
		runState := "RUNNING"
		for runState == "RUNNING" {
			time.Sleep(4 * time.Second)
			currentTime := time.Now().Unix()
			fmt.Printf("Progress: " + os.Args[1] + " %d\n", ((currentTime - startTime) / runTimeGuess) * 100)
			runState = runCommand("C:\\Windows\\System32\\schtasks.exe", "/QUERY", "/TN", os.Args[1], "/FO", "CSV", "/NH")
		}
		endTime := time.Now().Unix()
	
		runTime := endTime - startTime
		fmt.Printf("Done - runtime %d seconds.\n", runTime)
		runTimes = append(runTimes, runTime)
		sort.Slice(runTimes, func(i, j int) bool { return runTimes[i] < runTimes[j] })
		for len(runTimes) >= 10 {
			runTimes = runTimes[1:len(runTimes)-2]
		}
		outputString := ""
		for pl := 0; pl < len(runTimes); pl = pl + 1 {
			outputString = outputString + strconv.FormatInt(runTimes[pl], 10)
			if pl < len(runTimes)-1 {
				outputString = outputString + "\n"
			}
		}
		ioutil.WriteFile("runScheduledTask.txt", []byte(outputString), 0644)
	} else {
		fmt.Println("Usage: runScheduledTask NameOfWindowsScheduledTask")
	}
}
