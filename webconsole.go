package main

/*
Web Console - lets end-users run command-line applications via a web page, complete with authentication and a user interface.
Acts as its own self-contained web server.
*/

import (
	// Standard libraries.
	"fmt"
	"os"
	"io"
	"log"
	"time"
	"bufio"
	"errors"
	"strings"
	"os/exec"
	"math/rand"
	"io/ioutil"
	"net/http"
)

// Characters to use to generate new ID strings. Lowercase only - any user-provided IDs will be lowercased before use.
const letters = "abcdefghijklmnopqrstuvwxyz1234567890"

// The timeout, in seconds, of token validity.
const tokenTimeout = 600
// How often, in seconds, to check token for expired tokens.
const tokenCheckPeriod = 60

// Set up the tokens map.
var tokens = map[string]int64{}

var runningTasks = map[string]*exec.Cmd{}
var taskOutputs = map[string]io.ReadCloser{}

// Generate a new, random 16-character ID.
func generateIDString() string {
	rand.Seed(time.Now().UnixNano())
	result := make([]byte, 16)
	for pl := range result {
		result[pl] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}

// Clear any expired tokens from memory.
func clearExpiredTokens() {
	// This is a periodic task, it runs in a separate thread.
	for true {
		currentTimestamp := time.Now().Unix()
		for token, timestamp := range tokens { 
			if currentTimestamp - tokenTimeout > timestamp {
				delete(tokens, token)
			}
		}
		time.Sleep(tokenCheckPeriod * time.Second)
	}
}

// Split a string representing a command line with paramaters, possibly with quoted sections, into an array of strings.
func parseCommandString(theString string) []string {
	var result []string
	var stringSplit []string
	for theString != "" {
		theString = strings.TrimSpace(theString)
		if strings.HasPrefix(theString, "\"") {
			stringSplit = strings.SplitN(theString[1:], "\"", 2)
		} else {
			stringSplit = strings.SplitN(theString, " ", 2)
		}
		result = append(result, stringSplit[0])
		if len(stringSplit) > 1 {
			theString = stringSplit[1]
		} else {
			theString = ""
		}
	}
	return result
}

// Returns true if the given Task is currently running, false otherwise.
func taskIsRunning(theTaskID string) bool {
	if taskIDValue, taskIDFound := runningTasks[theTaskID]; taskIDFound {
		taskIDValue = taskIDValue
		return true
	}
	return false
}

// Read the Task's details from its config file.
func getTaskDetails(theTaskID string) (map[string]string, error) {
	taskDetails := make(map[string]string)
	configPath := "tasks/" + theTaskID + "/config.txt"
	// Check to see if we have a valid task ID.
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		inFile, inFileErr := os.Open(configPath)
		if inFileErr != nil {
			return taskDetails, errors.New("Can't open Task config file.")
		} else {
			// Read the Task's details from its config file.
			taskDetails["taskID"] = theTaskID
			taskDetails["title"] = ""
			taskDetails["secret"] = ""
			taskDetails["command"] = ""
			scanner := bufio.NewScanner(inFile)
			for scanner.Scan() {
				itemSplit := strings.SplitN(scanner.Text(), ":", 2)
				taskDetails[strings.TrimSpace(itemSplit[0])] = strings.TrimSpace(itemSplit[1])
			}
			inFile.Close()
		}
	} else {
		return taskDetails, errors.New("Invalid taskID")
	}
	return taskDetails, nil
}

// Returns a list of task details.
func getTaskList() ([]map[string]string, error) {
	var taskList []map[string]string
	taskIDs, readDirErr := ioutil.ReadDir("tasks")
	if readDirErr == nil {
		for _, taskID := range taskIDs {
			taskDetails, taskErr := getTaskDetails(taskID.Name())
			if taskErr == nil {
				taskList = append(taskList, taskDetails)
			} else {
				return taskList, taskErr
			}
		}
	} else {
		return taskList, errors.New("Can't read Tasks folder.")
	}
	return taskList, nil
}

// Get an input string from the user.
func getUserInput(defaultValue, messageString) string {
	inputReader := bufio.NewReader(os.Stdin)
	fmt.Print(messageString + ": ")
	result, _ := reader.ReadString('\n')
	if result == "" {
		return defaultValue
	}
	return result
}

// The main body of the program - parse user-provided command-line paramaters, or start the main web server process.
func main() {
	if len(os.Args) == 1 {
		// Start the thread that checks for and clears expired tokens.
		go clearExpiredTokens()
		
		// If no parameters are given, simply start the web server.
		fmt.Println("Starting web server...")
		
		// We write our own function to parse the request URL.
		http.HandleFunc("/", func (theResponseWriter http.ResponseWriter, theRequest *http.Request) {
			// Make sure submitted form values are parsed.
			theRequest.ParseForm()
			
			// The default root - serve index.html.
			if theRequest.URL.Path == "/" {
				http.ServeFile(theResponseWriter, theRequest, "www/index.html")
			// Handle a Task or API request. taskID needs to be provided as a parameter, either via GET or POST.
			} else if strings.HasPrefix(theRequest.URL.Path, "/task") || strings.HasPrefix(theRequest.URL.Path, "/api/") {
				taskID := theRequest.Form.Get("taskID")
				token := theRequest.Form.Get("token")
				if taskID == "" {
					fmt.Fprintf(theResponseWriter, "ERROR: Missing parameter taskID.")
				} else {
					taskDetails, taskErr := getTaskDetails(taskID)
					if taskErr == nil {
						authorised := false
						authorisationError := "unknown error"
						currentTimestamp := time.Now().Unix()
						if token != "" {
							if tokens[token] == 0 {
								authorisationError = "invalid or expired token"
							} else {
								authorised = true
							}
						} else if theRequest.Form.Get("secret") == taskDetails["secret"] {								
							authorised = true
						} else {
							authorisationError = "incorrect secret"
						}
						if authorised {
							if token == "" {
								token = generateIDString()
							}
							tokens[token] = currentTimestamp
							// Handle Task requests.
							if strings.HasPrefix(theRequest.URL.Path, "/task") {
								// Serve the webconsole.html file, first adding in the Task ID value so it can be used client-side.
								webconsoleBuffer, fileReadErr := ioutil.ReadFile("www/webconsole.html")
								if fileReadErr == nil {
									webconsoleString := string(webconsoleBuffer)
									webconsoleString = strings.Replace(webconsoleString, "taskID = \"\"", "taskID = \"" + taskID + "\"", -1)
									webconsoleString = strings.Replace(webconsoleString, "token = \"\"", "token = \"" + token + "\"", -1)
									http.ServeContent(theResponseWriter, theRequest, "webconsole.html", time.Now(), strings.NewReader(webconsoleString))
								} else {
									authorisationError = "couldn't read webconsole.html"
								}
							// API - Exchange the secret for a token.
							} else if strings.HasPrefix(theRequest.URL.Path, "/api/getToken") {
								fmt.Fprintf(theResponseWriter, token)
							// API - Return the Task's title.
							} else if strings.HasPrefix(theRequest.URL.Path, "/api/getTaskTitle") {
								fmt.Fprintf(theResponseWriter, taskDetails["title"])
							// API - Run a given Task.
							} else if strings.HasPrefix(theRequest.URL.Path, "/api/runTask") {
								if taskIsRunning(taskID) {
									fmt.Fprintf(theResponseWriter, "OK")
								} else {
									commandArray := parseCommandString(taskDetails["command"])
									var commandArgs []string
									if len(commandArray) > 0 {
										commandArgs = commandArray[1:]
									}
									runningTasks[taskID] = exec.Command(commandArray[0], commandArgs...)
									runningTasks[taskID].Dir = "tasks/" + taskID
									var taskErr error
									taskOutputs[taskID], taskErr = runningTasks[taskID].StdoutPipe()
									if taskErr == nil {
										taskErr = runningTasks[taskID].Start()
									}
									if taskErr == nil {
										fmt.Fprintf(theResponseWriter, "OK")
									} else {
										fmt.Printf("ERROR: " + taskErr.Error())
										fmt.Fprintf(theResponseWriter, "ERROR: " + taskErr.Error())
									}
								}
							} else if strings.HasPrefix(theRequest.URL.Path, "/api/getTaskOutput") {
								readBuffer := make([]byte, 10240)
								readSize, readErr := taskOutputs[taskID].Read(readBuffer)
								if readErr == nil {
									fmt.Fprintf(theResponseWriter, string(readBuffer[0:readSize]))
								} else {
									if readErr.Error() == "EOF" {
										delete(runningTasks, taskID)
										delete(taskOutputs, taskID)
									}
									fmt.Fprintf(theResponseWriter, "ERROR: " + readErr.Error())
								}
							} else if strings.HasPrefix(theRequest.URL.Path, "/api/getTaskRunning") {
								if taskIsRunning(taskID) {
									fmt.Fprintf(theResponseWriter, "YES")
								} else {
									fmt.Fprintf(theResponseWriter, "NO")
								}
							} else if strings.HasPrefix(theRequest.URL.Path, "/api/keepAlive") {
								fmt.Fprintf(theResponseWriter, "OK")
							} else if strings.HasPrefix(theRequest.URL.Path, "/api/") {
								fmt.Fprintf(theResponseWriter, "ERROR: Unknown API call: %s", theRequest.URL.Path)
							}
						} else {
							fmt.Fprintf(theResponseWriter, "ERROR: Not authorised - %s.", authorisationError)
						}
					} else {
						fmt.Fprintf(theResponseWriter, "ERROR: %s", taskErr.Error())
					}
				}
			// Otherwise, try and find the static file referred to by the request URL.
			} else {
				http.ServeFile(theResponseWriter, theRequest, "www" + theRequest.URL.Path)
			}
		})
		log.Fatal(http.ListenAndServe(":8090", nil))
	} else if os.Args[1] == "-list" {
		taskList, taskErr := getTaskList()
		if taskErr == nil {
			for _, task := range taskList {
				fmt.Println(task["taskID"] + ": " + task["title"] + "\n")
			}
		} else {
			fmt.Printf("ERROR: " + taskErr.Error() + "\n")
		}
	} else if os.Args[1] == "-new" {
		// Generate a new, unique Task ID.
		var newTaskID string
		for {
			newTaskID = generateIDString()
			if _, err := os.Stat("tasks/" + newTaskID); os.IsNotExist(err) {
				break
			}
		}
		newTaskID = getUserInput(newTaskID, "Enter a new Task ID (hit enter to generate an ID)")
		if _, err := os.Stat("tasks/" + newTaskID); os.IsNotExist(err) {
			os.Mkdir("tasks/" + newTaskID, os.ModePerm)
			fmt.Println("New Task: " + newTaskID)
		} else {
			fmt.Printf("ERROR: A task with ID " + newTaskID + " already exists.\n")
		}		
	}
}
