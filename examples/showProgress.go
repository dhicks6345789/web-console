package main

// A Go application that runs any arbitary command but adds progress (percentage complete) reports every few seconds.
// Guesses run time from previous runtimes.

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

var commandOutput = []string{}

func runCommand (theCommandArgs ...string) string {
	theCommand := exec.Command(theCommandArgs...)
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
		runTimeGuess := float64(totalRunTime / int64(len(runTimes)))
		
		startTime := time.Now().Unix()
		
		fmt.Println("Running \"" + os.Args[1] + "\"...")
		runCommand(os.Args[1:]...)
		runState := "RUNNING"
		for runState == "RUNNING" {
			time.Sleep(4 * time.Second)
			currentTime := time.Now().Unix()
			percentage := int((float64(currentTime - startTime) / runTimeGuess) * 100)
			if percentage > 100 {
				percentage = 100
			}
			//fmt.Printf("Progress: " + os.Args[1] + " %d%%\n", percentage)
			fmt.Printf("Progress: " + os.Args[1] + " %d\n", percentage)
			runState = runCommand("C:\\Windows\\System32\\schtasks.exe", "/QUERY", "/TN", os.Args[1], "/FO", "CSV", "/NH")
		}
		endTime := time.Now().Unix()
	
		runTime := endTime - startTime
		fmt.Printf("Progress: " + os.Args[1] + " 100\n")
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
