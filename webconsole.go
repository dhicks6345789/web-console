package main

/*
Web Console - lets end-users run command-line applications via a web page, complete with authentication and a user interface.
Acts as its own self-contained web server.
*/

import (
	// Standard libraries.
	"fmt"
	"os"
	"log"
	"time"
	"bufio"
	"errors"
	"strings"
	"strconv"
	"os/exec"
	"math/rand"
	"io/ioutil"
	"net/http"
	
	// Bcrypt for password hashing.
	"golang.org/x/crypto/bcrypt"
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
//var taskOutputs = map[string]io.ReadCloser{}
var taskOutputs = map[string][]string{}

// Generate a new, random 16-character ID.
func generateIDString() string {
	rand.Seed(time.Now().UnixNano())
	result := make([]byte, 16)
	for pl := range result {
		result[pl] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}

func hashPassword(thePassword string) (string, error) {
	bytes, cryptErr := bcrypt.GenerateFromPassword([]byte(thePassword), 14)
	return string(bytes), cryptErr
}

func checkPasswordHash(thePassword, theHash string) bool {
	if thePassword == "" && theHash == "" {
		return true
	}
	cryptErr := bcrypt.CompareHashAndPassword([]byte(theHash), []byte(thePassword))
	return cryptErr == nil
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

func startTask(theTaskID string) {
	readBuffer := make([]byte, 10240)
	taskOutputs[theTaskID] = make([]string, 0)
	taskOutput, taskErr := runningTasks[theTaskID].StdoutPipe()
	if taskErr == nil {
		taskErr = runningTasks[theTaskID].Start()
		if taskErr == nil {
			//go func() {
			//	runningTasks[theTaskID].Wait()
			//}()
			taskRunning := true
			for taskRunning {
				println("Reading from task:")
				readSize, readErr := taskOutput.Read(readBuffer)
				if readErr == nil {
					bufferSplit := strings.Split(string(readBuffer[0:readSize]), "\n")
					for pl := 0; pl < len(bufferSplit); pl++ {
						if strings.TrimSpace(bufferSplit[pl]) != "" {
							taskOutputs[theTaskID] = append(taskOutputs[theTaskID], bufferSplit[pl])
						}
					}
					for pl := 0; pl < len(taskOutputs[theTaskID]); pl++ {
						println(taskOutputs[theTaskID][pl])
					}
				} else {
					taskRunning = false
				}
			}
			println("Task done!")
			delete(runningTasks, theTaskID)
			delete(taskOutputs, theTaskID)
		}
	}
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
			taskDetails["public"] = "N"
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
func getUserInput(defaultValue string, messageString string) string {
	inputReader := bufio.NewReader(os.Stdin)
	fmt.Printf(messageString + ": ")
	result, _ := inputReader.ReadString('\n')
	result = strings.TrimSpace(result)
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
			// Handle the getPublicTaskList API call.
			} else if strings.HasPrefix(theRequest.URL.Path, "/api/getPublicTaskList") {
				taskList, taskErr := getTaskList()
				if taskErr == nil {
					taskListString := "{"
					for _, task := range taskList {
						if task["public"]  == "Y" {
							taskListString = taskListString + "\"" + task["taskID"] + "\":\"" + task["title"] + "\","
						}
					}
					if taskListString == "{" {
						fmt.Fprintf(theResponseWriter, "{}")
					} else {
						fmt.Fprintf(theResponseWriter, taskListString[:len(taskListString)-1] + "}")
					}
				} else {
					fmt.Fprintf(theResponseWriter, "ERROR: " + taskErr.Error())
				}
			// Handle a view, run or API request. taskID needs to be provided as a parameter, either via GET or POST.
			} else if strings.HasPrefix(theRequest.URL.Path, "/view") || strings.HasPrefix(theRequest.URL.Path, "/run") || strings.HasPrefix(theRequest.URL.Path, "/api/") {
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
						} else if checkPasswordHash(theRequest.Form.Get("secret"), taskDetails["secret"]) {
							authorised = true
						} else {
							authorisationError = "incorrect secret"
						}
						if authorised {
							if token == "" {
								token = generateIDString()
							}
							tokens[token] = currentTimestamp
							// Handle view and run requests.
							if strings.HasPrefix(theRequest.URL.Path, "/view") || strings.HasPrefix(theRequest.URL.Path, "/run") {
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
								if !taskIsRunning(taskID) {
									commandArray := parseCommandString(taskDetails["command"])
									var commandArgs []string
									if len(commandArray) > 0 {
										commandArgs = commandArray[1:]
									}
									runningTasks[taskID] = exec.Command(commandArray[0], commandArgs...)
									runningTasks[taskID].Dir = "tasks/" + taskID
									go startTask(taskID)
								}
								fmt.Fprintf(theResponseWriter, "OK")
							} else if strings.HasPrefix(theRequest.URL.Path, "/api/getTaskOutput") {
								println("Called getTaskOutput.")
								outputLineNumber := 0
								var atoiErr error
								if theRequest.Form.Get("line") != "" {
									outputLineNumber, atoiErr = strconv.Atoi(theRequest.Form.Get("line"))
									if atoiErr != nil {
										fmt.Fprintf(theResponseWriter, "ERROR: Line number not parsable.")
									}
								}
								for outputLineNumber < len(taskOutputs[taskID]) {
									fmt.Fprintf(theResponseWriter, taskOutputs[taskID][outputLineNumber] + "\n")
									outputLineNumber = outputLineNumber + 1
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
				secret := "Y"
				if task["secret"] == "" {
					secret = "N"
				}
				fmt.Println(task["taskID"] + ": " + task["title"] + ", Secret: " + secret + ", Public: " + task["public"] + ", Command: " + task["command"])
			}
		} else {
			fmt.Println("ERROR: " + taskErr.Error())
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
			
			newTaskTitle := "Task " + newTaskID
			newTaskTitle = getUserInput(newTaskTitle, "Enter a title (hit enter for \"" + newTaskTitle + "\")")
			
			newTaskSecret := ""
			newTaskSecret = getUserInput(newTaskSecret, "Set secret (type secret, or hit enter to skip)")
			
			var newTaskPublic string
			for {
				newTaskPublic = "N"
				newTaskPublic = strings.ToUpper(getUserInput(newTaskPublic, "Make this task public (\"Y\" or \"N\", hit enter for \"N\")"))
				if newTaskPublic == "Y" || newTaskPublic == "N" {
					break
				}
			}
			
			newTaskCommand := ""
			newTaskCommand = getUserInput(newTaskCommand, "Set command (type command, or hit enter to skip)")
			
			outputString := ""
			if newTaskSecret != "" {
				hashedPassword, hashErr := hashPassword(newTaskSecret)
				if hashErr == nil {
					outputString = outputString + "secret: " + hashedPassword + "\n"
				} else {
					fmt.Println("ERROR: Problem hashing password - " + hashErr.Error())
				}
			}
			outputString = outputString + "title: " + newTaskTitle + "\npublic: " + newTaskPublic + "\ncommand: " + newTaskCommand
			writeFileErr := ioutil.WriteFile("tasks/" + newTaskID + "/config.txt", []byte(outputString), 0644)
			if writeFileErr != nil {
				fmt.Println("ERROR: Couldn't write config for Task " + newTaskID + ".")
			}
		} else {
			fmt.Println("ERROR: A task with ID " + newTaskID + " already exists.")
		}		
	}
}
