package main

// A Go application that uses the Windows schtasks tool to run a given schedualed task.
// Doesn't exit until the schedualed task has finished. You can use Web Console's "progress" feature
// to add a progress bar if wanted.
import (
	// Standard libraries.
	"os"
	"fmt"
	"time"
	"strings"
	"os/exec"
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
		fmt.Println("Running \"" + os.Args[1] + "\"...")
		startTime := time.Now().Unix()
		runCommand("C:\\Windows\\System32\\schtasks.exe", "/RUN", "/TN", os.Args[1])
		runState := "RUNNING"
		for runState == "RUNNING" {
			time.Sleep(4 * time.Second)
			runState = runCommand("C:\\Windows\\System32\\schtasks.exe", "/QUERY", "/TN", os.Args[1], "/FO", "CSV", "/NH")
		}
		fmt.Printf("Done - runtime %d seconds.\n", time.Now().Unix() - startTime)
	} else {
		fmt.Println("Usage: runScheduledTask NameOfWindowsScheduledTask")
	}
}
