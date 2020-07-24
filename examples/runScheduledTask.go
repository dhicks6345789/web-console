// A Go application that uses the Windows schtasks tool to run a given schedualed task.
// Doesn't exit until the schedualed task has finished, and prints out a status indicator as the task is running.
// Uses previous run times to guess the total time the shedualed task will take.

import (
	// Standard libraries.
	"time"
	"os/exec"
)

startTime := time.Now().Unix()

theCommand = exec.Command("C:\Windows\System32\schtasks.exe", "/RUN", "/TN", "Salamander - Diary")
taskOutput, taskErr := theCommand.StdoutPipe()
if taskErr == nil {
	taskErr = theCommand.Start()
}
if taskErr == nil {
	println("OK")
} else {
	println("ERROR: " + taskErr.Error())
}

time.Sleep(4 * time.Second)

// C:\Windows\System32\schtasks.exe /QUERY /TN "Salamander - Diary" /FO CSV /NH

endTime := time.Now().Unix()

println(endTime - startTime)
