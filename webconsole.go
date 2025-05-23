package main
// Web Console - lets end-users run command-line applications via a web page, complete with authentication and a user interface.
// For more details, see https://www.sansay.co.uk/docs/web-console

import (
	// Standard libraries.
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"fmt"
	"log"
	"time"
	"sort"
	"bytes"
	"bufio"
	"regexp"
	"errors"
	"image"
	"image/png"
	"image/color"
	"strings"
	"strconv"
	"net/http"
	"net/smtp"
	"math/rand"
	"archive/zip"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"encoding/base64"
	
	// Image resizing library.
	"github.com/nfnt/resize"
	
	// An Image-to-SVG tracing library.
	"github.com/dennwc/gotrace"
	
	// A .ICO format Image encoder.
	"github.com/kodeworks/golang-image-ico"
	
	// Bcrypt for password hashing.
	"golang.org/x/crypto/bcrypt"
	
	// Argon2 for email address hashing - used with MyStart Online.
	"golang.org/x/crypto/argon2"
	
	// Excelize for loading in Excel files.
	// "github.com/360EntSecGroup-Skylar/excelize"
	"github.com/xuri/excelize/v2"
)

// Characters to use to generate new ID strings. Lowercase only - any user-provided IDs will be lowercased before use.
const letters = "abcdefghijklmnopqrstuvwxyz1234567890"

// User-friendly names to use for log levels in emails.
var logLevels = [5]string{"NONE", "ERROR", "WARNING", "MESSAGE", "ALL"}

// The current release version - value provided at compile time.
var buildVersion string

// A map to store any arguments passed on the command line.
var arguments = map[string]string{}

// We use tokens for session management, not cookies.
// The timeout, in seconds, of token validity.
const tokenTimeout = 600
// How often, in seconds, to check for expired tokens.
const tokenCheckPeriod = 60
// A map of current valid tokens...
var tokens = map[string]int64{}
// ...and matching permissions...
var permissions = map[string]string{}
// ...and user IDs (user hashes).
var userIDs = map[string]string{}

// A list of currently running Tasks.
var runningTasks = map[string]*exec.Cmd{}
// The outputs from Tasks.
var taskOutputs = map[string][]string{}
// The input pipes for Tasks.
var taskInputs = map[string]io.WriteCloser{}
// We record the start time and an array of recent runtimes for each Task so we can guess at this run's liklely time and print a progress report if wanted.
var taskStartTimes = map[string]int64{}
var taskRunTimes = map[string][]int64{}
var taskRuntimeGuesses = map[string]float64{}
// We record the user that runs each Task, so we know who has permissions for inputs / stop command.
var taskRunUsers = map[string]string{}
// We record the stop time for each Task so we can implement rate limiting.
var taskStopTimes = map[string]int64{}

// Valid authentication services.
var authServices = []string{"mystart", "cloudflare", "ngrok"}
var authServiceNames = map[string][]string{}

// Items to exclude from being returned from the listFolders API call.
var listFolderExcludes = []string{".git", "__pycache__"}

// Maps of MyStart.Online page names and API keys.
var mystartPageNames = map[string]string{}
var mystartAPIKeys = map[string]string{}

// A map of endpoints to files to serve.
var filesToServeList = map[string]string{"/":"index.html", "/view":"webconsole.html", "/run":"webconsole.html", "/login":"login.html", "/api/mystartLogin":"redirect.html"}

// A struct used to read JSON data from authentication API calls to MyStart.Online.
type mystartStruct struct {
	Login string
	EmailHash string
	EmailDomain string
	LoginType string
}

// Some constant values for use with the Argon2 hashing function.
const argon2Iterations uint32 = 16
const argon2Memory uint32 = 8
const argon2Parallelism uint8 = 1
const argon2KeyLength uint32 = 16

// If the "debug" option has been passed on the command line, print the given information to the (local) console.
func debug(theOutput string) {
	if arguments["debug"] == "true" {
		currentTime := time.Now()
		fmt.Println("webconsole,", currentTime.Format("02/01/2006:15:04:05"), "- " + theOutput)
	}
}

// Print a log file entry in Common Log Format.
func logLine(theOutput string) {
	currentTime := time.Now()
	fmt.Println("log,", currentTime.Format("02/01/2006:15:04:05"), "- " + theOutput)
}

// Generate a new, random 16-character string, used for tokens and Task IDs.
func generateRandomString() string {
	rand.Seed(time.Now().UnixNano())
	result := make([]byte, 16)
	for pl := range result {
		result[pl] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}

// Use the Bcrypt hashing algorithm to encode a password string.
func hashPassword(thePassword string) (string, error) {
	bytes, cryptErr := bcrypt.GenerateFromPassword([]byte(thePassword), 14)
	return string(bytes), cryptErr
}

// Check a plain text password with a Bcrypt-hashed string, returns true if they match.
func checkPasswordHash(thePassword, theHash string) bool {
	if thePassword == "" && theHash == "" {
		return false
	}
	cryptErr := bcrypt.CompareHashAndPassword([]byte(theHash), []byte(thePassword))
	return cryptErr == nil
}

// Clear any expired tokens from memory.
func clearExpiredTokens() {
	// This is a periodic task, it runs in a separate thread (goroutine) - the time period is set by the tokenCheckPeriod constant set at the top of the script.
	for true {
		currentTimestamp := time.Now().Unix()
		for token, timestamp := range tokens { 
			if currentTimestamp - tokenTimeout > timestamp {
				delete(tokens, token)
				delete(permissions, token)
				delete(userIDs, token)
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

// Runs a task, capturing output from stdout and placing it in a buffer. Designed to be run as a goroutine, so a task can be run in the background
// and output captured while the user does other stuff.
func runTask(theTaskID string) {
	taskDetails, taskErr := getTaskDetails(theTaskID)
	readBuffer := make([]byte, 10240)
	taskOutputs[theTaskID] = make([]string, 0)
	taskStdout, taskStdoutErr := runningTasks[theTaskID].StdoutPipe()
	if taskStdoutErr == nil && taskErr == nil {
		taskStderr, taskStderrErr := runningTasks[theTaskID].StderrPipe()
		if taskStderrErr == nil {
			taskStdin, taskStdinErr := runningTasks[theTaskID].StdinPipe()
			if taskStdinErr == nil {
				taskInputs[theTaskID] = taskStdin
				taskOutput := io.MultiReader(taskStdout, taskStderr)
				logfileOutput, logFileErr := os.Create(arguments["taskroot"] + "/" + theTaskID + "/log.txt")
				if logFileErr == nil {
					taskErr := runningTasks[theTaskID].Start()
					if taskErr == nil {
						taskRunning := true
						// Loop until the Task (an external executable) has finished.
						for taskRunning {
							// Read both STDERR and STDIN.
							readOutputSize, readErr := taskOutput.Read(readBuffer)
							if readErr == nil {
								// Append the output to the log file for the current Task.
								logfileOutput.Write(readBuffer[0:readOutputSize])
								// Append the output as lines of text to the array-of-strings ready for output to the web interface.
								bufferSplit := strings.Split(string(readBuffer[0:readOutputSize]), "\n")
								for pl := 0; pl < len(bufferSplit); pl++ {
									if strings.TrimSpace(bufferSplit[pl]) != "" {
										taskOutputs[theTaskID] = append(taskOutputs[theTaskID], bufferSplit[pl])
									}
								}
							} else {
								taskRunning = false
							}
						}
						// Get the exit status of the running Task. If non-zero, pass the error message back to the user.
						exitErr := runningTasks[theTaskID].Wait()
						if exitErr != nil {
							errorString := "\nERROR: " + exitErr.Error() + "\n"
							logfileOutput.Write([]byte(errorString))
							taskOutputs[theTaskID] = append(taskOutputs[theTaskID], errorString)
						}
						// When we get here, the Task has finished running. We record the finish time and work out the total run time for this run
						// and update (or create) the list of recent run times for this Task.
						taskStopTimes[theTaskID] = time.Now().Unix()
						runTime := taskStopTimes[theTaskID] - taskStartTimes[theTaskID]
						taskRunTimes[theTaskID] = append(taskRunTimes[theTaskID], runTime)
						// We don't just record every runtime, we sort the times and trim them to a set of 10 at most, that way we get a reasonable
						// guess at an average run time, assuming run times are similar each time.
						sort.Slice(taskRunTimes[theTaskID], func(i, j int) bool { return taskRunTimes[theTaskID][i] < taskRunTimes[theTaskID][j] })
						for len(taskRunTimes[theTaskID]) >= 10 {
							taskRunTimes[theTaskID] = taskRunTimes[theTaskID][1:len(taskRunTimes[theTaskID])-2]
						}
						// Write the runTimes.txt file for this Task.
						outputString := ""
						for pl := 0; pl < len(taskRunTimes[theTaskID]); pl = pl + 1 {
							outputString = outputString + strconv.FormatInt(taskRunTimes[theTaskID][pl], 10)
							if pl < len(taskRunTimes[theTaskID])-1 {
								outputString = outputString + "\n"
							}
						}
						ioutil.WriteFile("tasks/" + theTaskID + "/runTimes.txt", []byte(outputString), 0644)
						
						// Depending on the defined logReportingLevel for this Task, email the defined receipiants with the log results.
						logReportingLevel := 0
						highestLogLevelFound := 0
						lowestLogLevelFound := 4
						if taskDetails["logReportingLevel"] == "error" {
							logReportingLevel = 1
						} else if taskDetails["logReportingLevel"] == "warning" {
							logReportingLevel = 2
						} else if taskDetails["logReportingLevel"] == "result" {
							logReportingLevel = 3
						} else if taskDetails["logReportingLevel"] == "message" {
							logReportingLevel = 4
						} else if taskDetails["logReportingLevel"] == "all" {
							logReportingLevel = 5
						}
						if logReportingLevel > 0 {
							debug("Emailing log reports, level: " + taskDetails["logReportingLevel"] + "...")
							logMessageBody := ""
							for pl := 0; pl < len(taskOutputs[theTaskID]); pl = pl + 1 {
								if strings.HasPrefix(strings.ToLower(taskOutputs[theTaskID][pl]), "error") {
									highestLogLevelFound = max(highestLogLevelFound, 1)
									lowestLogLevelFound = min(lowestLogLevelFound, 1)
									logMessageBody = logMessageBody + taskOutputs[theTaskID][pl] + "\n"
								} else if logReportingLevel > 1 {
									if strings.HasPrefix(strings.ToLower(taskOutputs[theTaskID][pl]), "warning") {
										highestLogLevelFound = max(highestLogLevelFound, 2)
										lowestLogLevelFound = min(lowestLogLevelFound, 2)
										logMessageBody = logMessageBody + taskOutputs[theTaskID][pl] + "\n"
									} else if logReportingLevel > 2 {
										if strings.HasPrefix(strings.ToLower(taskOutputs[theTaskID][pl]), "result") {
											highestLogLevelFound = max(highestLogLevelFound, 3)
											lowestLogLevelFound = min(lowestLogLevelFound, 3)
											logMessageBody = logMessageBody + taskOutputs[theTaskID][pl] + "\n"
										} else if logReportingLevel > 3 {
											if strings.HasPrefix(strings.ToLower(taskOutputs[theTaskID][pl]), "status") {
												highestLogLevelFound = max(highestLogLevelFound, 4)
												lowestLogLevelFound = min(lowestLogLevelFound, 4)
												logMessageBody = logMessageBody + taskOutputs[theTaskID][pl] + "\n"
											} else if logReportingLevel > 4 {
												highestLogLevelFound = max(highestLogLevelFound, 5)
												lowestLogLevelFound = min(lowestLogLevelFound, 5)
												logMessageBody = logMessageBody + taskOutputs[theTaskID][pl] + "\n"
											}
										}
									}
								}
							}
							if highestLogLevelFound >= logReportingLevel {
								emailBody := "From: " + taskDetails["smtpFrom"] + "\n"
								emailBody = emailBody + "To: " + taskDetails["smtpTo"] + "\n"
								currentTime := time.Now()
								emailBody = emailBody + "Subject: "
								if lowestLogLevelFound < 4 {
									emailBody = emailBody + "[" + logLevels[lowestLogLevelFound] + "] "
								}
								emailBody = emailBody + "Task \"" + taskDetails["title"] + "\" completed " + currentTime.Format("02/01/2006 15:04:05") + "\n"
								emailBody = emailBody + "\n"
								emailBody = emailBody + logMessageBody
								smtpAuth := smtp.PlainAuth("", taskDetails["smtpFrom"], taskDetails["smtpPassword"], taskDetails["smtpHost"])
								smtpError := smtp.SendMail(taskDetails["smtpHost"] + ":" + taskDetails["smtpPort"], smtpAuth, taskDetails["smtpFrom"], []string{taskDetails["smtpTo"]}, []byte(emailBody))
								if smtpError != nil {
									debug(fmt.Sprint(smtpError))
								} else {
									debug("Email sent ok.")
								}
							}
						}
						
						// Remove this Task from the runnings Tasks list. We don't remove the output right away - client-side code might
						// still not have received all the output yet.
						delete(runningTasks, theTaskID)
					}
					logfileOutput.Close()
					// Copy the just-finished log file to the "output" folder with a unique timestamp.
					logfilePath := arguments["taskroot"] + "/" + theTaskID + "/output"
					if _, err := os.Stat(logfilePath); os.IsNotExist(err) {
						os.Mkdir(logfilePath, os.ModePerm)
					}
					timestampString := time.Unix(taskStartTimes[theTaskID], 0).Format(time.RFC3339) + "-" + strconv.Itoa(int(taskStopTimes[theTaskID] - taskStartTimes[theTaskID]))
					logfileContent, logfileErr := ioutil.ReadFile(arguments["taskroot"] + "/" + theTaskID + "/log.txt")
					if logfileErr == nil {
						logfileName := logfilePath + "/" + strings.ReplaceAll(timestampString, ":", "-") + ".txt"
						logfileContentErr := ioutil.WriteFile(logfileName, logfileContent, 0644)
						if logfileContentErr != nil {
							debug("Some issue writing log file: " + logfileName + ".txt")
						}
					} else {
						debug("Some issue reading log file: " + arguments["taskroot"] + "/" + theTaskID + "/log.txt")
					}
				}
			}
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
	taskDetails["taskID"] = theTaskID
	taskDetails["title"] = ""
	taskDetails["description"] = ""
	taskDetails["secretViewers"] = ""
	taskDetails["secretRunners"] = ""
	taskDetails["secretEditors"] = ""
	taskDetails["public"] = "N"
	taskDetails["ratelimit"] = "0"
	taskDetails["progress"] = "N"
	taskDetails["resultURL"] = ""
	taskDetails["command"] = ""
	taskDetails["authentication"] = ""
	taskDetails["logPeriod"] = "129600"
	taskDetails["logReportingLevel"] = "none"
	ok := true
	if taskDetails["smtpPort"], ok = arguments["smtpport"]; !ok {
		taskDetails["smtpPort"] = "587"
	}
	if taskDetails["smtpHost"], ok = arguments["smtphost"]; !ok {
		taskDetails["smtpHost"] = ""
	}
	if taskDetails["smtpPassword"], ok = arguments["smtppassword"]; !ok {
		taskDetails["smtpPassword"] = ""
	}
	if taskDetails["smtpFrom"], ok = arguments["smtpfrom"]; !ok {
		taskDetails["smtpFrom"] = ""
	}
	if taskDetails["smtpTo"], ok = arguments["smtpto"]; !ok {
		taskDetails["smtpTo"] = ""
	}
	//debug("Finding details for Task: " + theTaskID)
	// Check to see if we have a valid task ID.
	if (theTaskID == "/") {
		// The root Task is always public.
		taskDetails["public"] = "Y"
		
		for _, authService := range authServices {
			// If we have any globally defined authentication service variables then that authentication service is a
			// valid authentication method for the root Task.
			if len(authServiceNames[authService]) > 0 {
				taskDetails["authentication"] = taskDetails["authentication"] + authService + ","
			}
		
			for _, authServiceName := range authServiceNames[authService] {
				editorsName := authService + authServiceName + "Editors"
				editorsPath := arguments["webconsoleroot"] + "/" + editorsName + ".csv"
				if _, err := os.Stat(editorsPath); err == nil {
					taskDetails[editorsName] = editorsPath
					//debug("Editors - found " + editorsName + " file: " + editorsPath)
				}
				runnersName := authService + authServiceName + "Runners"
				runnersPath := arguments["webconsoleroot"] + "/" + runnersName + ".csv"
				if _, err := os.Stat(runnersPath); err == nil {
					taskDetails[runnersName] = runnersPath
					//debug("Runners - found " + runnersName + " file: " + runnersPath)
				}
				viewersName := authService + authServiceName + "Viewers"
				viewersPath := arguments["webconsoleroot"] + "/" + viewersName + ".csv"
				if _, err := os.Stat(viewersPath); err == nil {
					taskDetails[viewersName] = viewersPath
					//debug("Viewers - found " + viewersName + " file: " + viewersPath)
				}
			}
		}
	} else {
		// Get the root Task's details to provide default values if this Task doesn't overwrite them.
		rootTaskDetails, _ := getTaskDetails("/")
		// Read this Task's config file.
		configPath := arguments["taskroot"] + "/" + theTaskID + "/config.txt"
		if _, err := os.Stat(configPath); !os.IsNotExist(err) {
			inFile, inFileErr := os.Open(configPath)
			if inFileErr != nil {
				return taskDetails, errors.New("Can't open Task config file.")
			} else {
				for _, authService := range authServices {
					// If any authorisation service paths are set at the root Task level, use those values as
					// defaults - they can be overwritten by this Task's local settings.
					for rootTaskDetailName, rootTaskDetailValue := range rootTaskDetails {
						if strings.HasPrefix(rootTaskDetailName, authService) {
							taskDetails[rootTaskDetailName] = rootTaskDetailValue
						}
					}
				
					// If we have any globally defined authentication service variables then that authentication service is a
					// valid authentication method for this Task.
					if len(authServiceNames[authService]) > 0 {
						taskDetails["authentication"] = taskDetails["authentication"] + authService + ","
					}
					
					for _, authServiceName := range authServiceNames[authService] {
						editorsName := authService + authServiceName + "Editors"
						editorsPath := arguments["webconsoleroot"] + "/tasks/" + taskDetails["taskID"] + "/" + editorsName + ".csv"
						if _, err := os.Stat(editorsPath); err == nil {
							taskDetails[editorsName] = editorsPath
						}
						runnersName := authService + authServiceName + "Runners"
						runnersPath := arguments["webconsoleroot"] + "/tasks/" + taskDetails["taskID"] + "/" + runnersName + ".csv"
						if _, err := os.Stat(runnersPath); err == nil {
							taskDetails[runnersName] = runnersPath
						}
						viewersName := authService + authServiceName + "Viewers"
						viewersPath := arguments["webconsoleroot"] + "/tasks/" + taskDetails["taskID"] + "/" + viewersName + ".csv"
						if _, err := os.Stat(viewersPath); err == nil {
							taskDetails[viewersName] = viewersPath
						}
					}
				}
				
				// Read the Task's details from its config file.
				scanner := bufio.NewScanner(inFile)
				for scanner.Scan() {
					itemSplit := strings.SplitN(scanner.Text(), ":", 2)
					itemName := strings.TrimSpace(itemSplit[0])
					itemVal := strings.TrimSpace(itemSplit[1])
					taskDetails[itemName] = itemVal
				}
				inFile.Close()
			
				// Figure out what authentication types this Task accepts...
				authTypes := map[string]int{}
				for _, secretType := range []string{"secretViewers","secretRunners","secretEditors"} {
					if taskDetails[secretType] != "" {
						authTypes["secret"] = 1
					}
				}
				// ...and condense them into a comma-separated string value...
				for authType, _ := range authTypes {
					taskDetails["authentication"] = taskDetails["authentication"] + authType + ","
				}
				// ...making sure we trim the last comma off the end.
				if len(taskDetails["authentication"]) > 0 {
					taskDetails["authentication"] = taskDetails["authentication"][0:len(taskDetails["authentication"])-1]
				}
			
				// Get the Task's description - see if there's a "description.txt" file, otherwise use the value set in
				// the "description" value (empty string by default if not otherwise set in the Task's config file).
				descriptionContents, descriptionContentsErr := ioutil.ReadFile(arguments["taskroot"] + "/" + theTaskID + "/description.txt")
				if descriptionContentsErr == nil {
					taskDetails["fullDescription"] = string(descriptionContents)
				}
				if taskDetails["description"] == "" {
					taskDetails["description"] = taskDetails["fullDescription"]
				}
				taskDetails["shortDescription"] = taskDetails["description"]
				delete(taskDetails, "description");
				if taskDetails["fullDescription"] == "" {
					taskDetails["fullDescription"] = taskDetails["shortDescription"]
				}
			}
		} else {
			return taskDetails, errors.New("No config file for taskID: " + theTaskID + " - configPath: " + configPath)
		}
	}
	return taskDetails, nil
}

// Returns a list of task details.
func getTaskList() ([]map[string]string, error) {
	var taskList []map[string]string
	taskIDs, readDirErr := ioutil.ReadDir(arguments["taskroot"])
	if readDirErr == nil {
		for _, taskID := range taskIDs {
			taskDetails, taskErr := getTaskDetails(taskID.Name())
			if taskErr == nil {
				taskList = append(taskList, taskDetails)
			}
		}
	} else {
		return taskList, errors.New("Can't read Tasks folder.")
	}
	return taskList, nil
}

func getTaskPermission(webConsoleRoot string, taskDetails map[string]string, userID string) string {
	//debug("Finding permissions for Task: " + taskDetails["taskID"])
	for taskDetailName, taskDetailValue := range taskDetails {
		if strings.HasPrefix(taskDetailName, "cloudflare") {
			permissionToGrant := ""
			for _, permissionCheck := range [3]string{"Editors", "Runners", "Viewers"} {
				if strings.HasSuffix(taskDetailName, permissionCheck) {
					permissionToGrant = string(permissionCheck[0])
				}
			}
			if permissionToGrant != "" {
				cloudflareUsersPath := taskDetailValue
				//debug("cloudflareUsersPath: " + cloudflareUsersPath)
				if _, err := os.Stat(cloudflareUsersPath); !os.IsNotExist(err) {
					cloudflareUsers := readUserFile(cloudflareUsersPath, "")
					for _, userEmail := range cloudflareUsers {
						if userEmail == userID {
							return permissionToGrant
						}
					}
				}
			}
		} else if strings.HasPrefix(taskDetailName, "ngrok") {
			permissionToGrant := ""
			for _, permissionCheck := range [3]string{"Editors", "Runners", "Viewers"} {
				if strings.HasSuffix(taskDetailName, permissionCheck) {
					permissionToGrant = string(permissionCheck[0])
				}
			}
			if permissionToGrant != "" {
				ngrokUsersPath := taskDetailValue
				//debug("ngrokUsersPath: " + ngrokUsersPath)
				if _, err := os.Stat(ngrokUsersPath); !os.IsNotExist(err) {
					ngrokUsers := readUserFile(ngrokUsersPath, "")
					for _, userEmail := range ngrokUsers {
						if userEmail == userID {
							return permissionToGrant
						}
					}
				}
			}
		}
	}
	return ""
}

// Get an input string from the user via stdin.
func getUserInput(argumentsKey, defaultValue string, messageString string) string {
	if argument, argumentExists := arguments[argumentsKey]; argumentExists {
		return argument
	}
	inputReader := bufio.NewReader(os.Stdin)
	fmt.Printf(messageString + ": ")
	result, _ := inputReader.ReadString('\n')
	result = strings.TrimSpace(result)
	if result == "" {
		return defaultValue
	}
	return result
}

// A helper function that sets the given "arguments" value to the first discovered valid path from a list given as an array of strings.
func setArgumentIfPathExists(theArgument string, thePaths []string) {
	for _, path := range thePaths {
		if _, existsErr := os.Stat(path); !os.IsNotExist(existsErr) {
			arguments[theArgument] = path
			return
		}
	}
}

func readConfigFile(theConfigPath string) map[string]string {
	var result = map[string]string{}
	
	// Is the config file an Excel file?
	if strings.HasSuffix(strings.ToLower(theConfigPath), "xlsx") {
		excelFile, excelErr := excelize.OpenFile(theConfigPath)
		if excelErr == nil {
			excelSheetName := excelFile.GetSheetName(0)
			excelCells, cellErr := excelFile.GetRows(excelSheetName)
			if cellErr == nil {
				fmt.Println(excelCells)
			} else {
				fmt.Println("ERROR: " + cellErr.Error())
			}
		} else {
			fmt.Println("ERROR: " + excelErr.Error())
		}
	} else if strings.HasSuffix(strings.ToLower(theConfigPath), "csv") {
		csvFile, csvErr := os.Open(theConfigPath)
		if csvErr == nil {
			csvData := csv.NewReader(csvFile)
			for {
				csvDataRecord, csvDataErr := csvData.Read()
				if csvDataErr == io.EOF {
					break
				}
				if csvDataErr != nil {
					fmt.Println("ERROR: " + csvDataErr.Error())
				} else {
					csvDataField := strings.ToLower(csvDataRecord[0])
					if csvDataField != "parameter" && !strings.HasPrefix(csvDataField, "#") {
						result[csvDataField] = csvDataRecord[1]
					}
				}
			}
		} else {
			fmt.Println("ERROR: " + csvErr.Error())
		}
	}
	return result
}

// Read a "users" data file - a file telling us which users are valid Editors, Runners or Viewers.
// Files can be in Excel or CSV format, two columns: Email Address, Hash Value
func readUserFile(theConfigPath string, theHashKey string) map[string]string {
	var result = map[string]string{}
	
	// Is the config file an Excel file?
	if strings.HasSuffix(strings.ToLower(theConfigPath), "xlsx") {
		excelFile, excelErr := excelize.OpenFile(theConfigPath)
		if excelErr == nil {
			excelSheetName := excelFile.GetSheetName(0)
			excelCells, cellErr := excelFile.GetRows(excelSheetName)
			if cellErr == nil {
				fmt.Println(excelCells)
			} else {
				fmt.Println("ERROR: " + cellErr.Error())
			}
		} else {
			fmt.Println("ERROR: " + excelErr.Error())
		}
	} else if strings.HasSuffix(strings.ToLower(theConfigPath), "csv") {
		rewriteCSVFile := false
		// If the data file is a CSV file, read it.
		csvFile, csvErr := os.Open(theConfigPath)
		if csvErr == nil {
			csvData := csv.NewReader(csvFile)
			for {
				csvDataRecord, csvDataErr := csvData.Read()
				if csvDataErr == io.EOF {
					break
				}
				if csvDataErr != nil {
					fmt.Println("ERROR: " + csvDataErr.Error())
				} else {
					emailAddress := strings.ToLower(strings.TrimSpace(csvDataRecord[0]))
					if theHashKey == "" {
						result[emailAddress] = emailAddress
					} else {
						// Figure out if the first value is a valid hash.
						emailAddressIsHash := true
						if len(emailAddress) == 32 {
							for _, addressCharValue := range emailAddress {
								if !strings.Contains(letters, string(addressCharValue)) {
									emailAddressIsHash = false
								}
							}
						} else {
							emailAddressIsHash = false
						}
						hashedEmailAddress := ""
						// If the first value is a valid hash, we need to re-write the CSV file in the correct order.
						if emailAddressIsHash {
							hashedEmailAddress = emailAddress
							emailAddress = ""
							rewriteCSVFile = true
						} else {
							if len(csvDataRecord) > 1 {
								// If we already have a valid hash value, read it.
								hashedEmailAddress = csvDataRecord[1]
							} else {
								// If we don't currently have a hash value, we'll need to calculate one, then re-write the CSV file.
								hashedEmailAddress = hex.EncodeToString(argon2.Key([]byte(emailAddress), []byte(theHashKey), argon2Iterations, argon2Memory, argon2Parallelism, argon2KeyLength))
								rewriteCSVFile = true
							}
						}
						result[emailAddress] = hashedEmailAddress
					}
				}
			}
		} else {
			fmt.Println("ERROR: " + csvErr.Error())
		}
		if rewriteCSVFile == true {
			// Re-write the config CSV file.
			csvFile, csvErr := os.Create(theConfigPath)
			if csvErr == nil {
				csvWriter := csv.NewWriter(csvFile)
				var csvData [][]string
				for csvEmailValue, csvHashValue := range result {
					csvData = append(csvData, []string{string(csvEmailValue), string(csvHashValue)})
				}
				csvWriter.WriteAll(csvData)
			} else {
				fmt.Println("ERROR: " + csvErr.Error())
			}
		}
	}
	return result
}

// A utility function that returns true if the string-to-match is in the given array of strings.
func contains(theItems []string, theMatch string) bool {
	for _, item := range theItems {
		if theMatch == item {
			return true
		}
	}
	return false
}

// A function that returns a listing of the given folder as a JSON string. Doesn't recursivly walk down sub-folders, a separate
// API call has to be made on the client side to get the contents of each sub-folder. The original function did simply return all
// sub-folders, but this could return more data than the client side could deal with.
func listOneFolderAsJSON(thePath string) string {
	result := ""

	// Read all items (both sub-folders and files) from the given folder path.
	items, itemErr := os.ReadDir(thePath)
	if itemErr != nil {
		return "Error reading path: " + thePath
	}
	
	// Now, step through each item, producing JSON output as we go.
	for pl := 0; pl < len(items); pl = pl + 1 {
		if contains(listFolderExcludes, items[pl].Name()) == false {
			if items[pl].IsDir() {
				result = result + "[\"" + items[pl].Name() + "\",[]]\n"
			} else {
				result = result + "\"" + items[pl].Name() + "\""
			}
			if pl < len(items)-1 {
				result = result + ","
			}
			result = result + "\n"
		}
	}
	if result == "" {
		result = "\"\"\n"
	}
	return result
}

func normalisePath(thePath string) string {
	otherPathSeparator := "/"
	if string(os.PathSeparator) == "/" {
		otherPathSeparator = "\\"
	}

	previousResult := ""
	result := strings.Replace(thePath, otherPathSeparator, string(os.PathSeparator), -1)
	for result != previousResult {
		previousResult = result
		result = strings.Replace(previousResult, string(os.PathSeparator) + string(os.PathSeparator), string(os.PathSeparator), -1)
	}
	
	return result
}

func getZippedFolderContents(zipWriter *zip.Writer, rootPath string, currentPath string) error {
	// Read all items (both sub-folders and files) from the given folder path...
	readItems, itemErr := os.ReadDir(rootPath + string(os.PathSeparator) + currentPath)
	if itemErr != nil {
		return itemErr
	}
	
	// ...and remove anything we want to exclude.
	var items []fs.DirEntry
	for pl := 0; pl < len(readItems); pl = pl + 1 {
		if contains(listFolderExcludes, readItems[pl].Name()) == false {
			items = append(items, readItems[pl])
		}
	}
	
	// Now, step through each item, adding to the Zip archive as we go.
	for pl := 0; pl < len(items); pl = pl + 1 {
		if contains(listFolderExcludes, items[pl].Name()) == false {
			itemPath := items[pl].Name()
			if currentPath != "" {
				itemPath = currentPath + string(os.PathSeparator) + itemPath
			}
			if items[pl].IsDir() {
				zipErr := getZippedFolderContents(zipWriter, rootPath, itemPath)
				if zipErr != nil {
					return zipErr
				}
			} else {
				filePathToZip := normalisePath(rootPath + string(os.PathSeparator) + itemPath)
				zipPath := strings.Replace(itemPath, string(os.PathSeparator), "/", -1)
				debug("Adding to zip: " + filePathToZip  + " as " + zipPath)
				fileToZip, zipErr := os.Open(filePathToZip)
				if zipErr != nil {
					debug("getZippedFolderContents - error: " + zipErr.Error())
					return zipErr
				}
				zippedFile, zipErr := zipWriter.Create(zipPath)
				if zipErr != nil {
					debug("getZippedFolderContents - error: " + zipErr.Error())
					return zipErr
				}
				_, zipErr = io.Copy(zippedFile, fileToZip)
				if zipErr != nil {
					debug("getZippedFolderContents - error: " + zipErr.Error())
					return zipErr
				}
				fileToZip.Close()
			}
		}
	}
	return nil
}



func doServeFile(theResponseWriter http.ResponseWriter, theRequest *http.Request, theFile string, theTaskID string, theToken string, thePermission string, theTitle string, theShortDescription string, theFullDescription string) {
	// Serve the "fileToServe" file, first adding in the Task ID and token values to be used client-side, as well
	// as including the appropriate formatting.js file.
	debug("Serving file: " + theFile)
	webconsoleBuffer, fileReadErr := ioutil.ReadFile(arguments["webroot"] + "/" + theFile)
	if fileReadErr == nil {
		formattingJSBuffer, fileReadErr := ioutil.ReadFile(arguments["taskroot"] + "/" + theTaskID + "/formatting.js")
		if fileReadErr != nil {
			formattingJSBuffer, fileReadErr = ioutil.ReadFile(arguments["taskroot"] + "/formatting.js")
			if fileReadErr != nil {
				formattingJSBuffer, fileReadErr = ioutil.ReadFile(arguments["webroot"] + "/formatting.js")
			}
		}
		if fileReadErr == nil {
			formattingJSString := string(formattingJSBuffer)
			webconsoleString := string(webconsoleBuffer)
			webconsoleString = strings.Replace(webconsoleString, "<<MYSTARTLOGINPAGE>>", arguments["mystartpagename"], -1)
			webconsoleString = strings.Replace(webconsoleString, "<<TASKID>>", theTaskID, -1)
			webconsoleString = strings.Replace(webconsoleString, "<<TOKEN>>", theToken, -1)
			webconsoleString = strings.Replace(webconsoleString, "<<PERMISSION>>", thePermission, -1)
			webconsoleString = strings.Replace(webconsoleString, "<<TITLE>>", theTitle, -1)
			webconsoleString = strings.Replace(webconsoleString, "<<SHORTDESCRIPTION>>", theShortDescription, -1)
			webconsoleString = strings.Replace(webconsoleString, "<<FULLDESCRIPTION>>", theFullDescription, -1)
			webconsoleString = strings.Replace(webconsoleString, "<<FAVICONPATH>>", theTaskID + "/", -1)
			webconsoleString = strings.Replace(webconsoleString, "// Include formatting.js.", formattingJSString, -1)
			http.ServeContent(theResponseWriter, theRequest, theFile, time.Now(), strings.NewReader(webconsoleString))
		} else {
			fmt.Fprintf(theResponseWriter, "ERROR: Couldn't read formatting.js")
		}
	} else {
		fmt.Fprintf(theResponseWriter, "ERROR: Couldn't read " + arguments["webroot"] + "/" + theFile)
	}
}

// The main body of the program - parse user-provided command-line paramaters, or start the main web server process.
func main() {
	// This application is both a web server for handling API requests and displaying a web-based front end, and a command-line application for handling
	// configuration and setup.
	
	// Set valid authentication services.
	for _, authService := range authServices {
		authServiceNames[authService] = []string{}
	}
	
	// Set some default argument values.
	arguments["help"] = "false"
	arguments["start"] = "true"
	arguments["list"] = "false"
	arguments["new"] = "false"
	arguments["port"] = "8090"
	arguments["localonly"] = "true"
	arguments["debug"] = "false"
	arguments["shellprefix"] = ""
	arguments["cloudflare"] = "false"
	arguments["ngrok"] = "false"
	arguments["logreportinglevel"] = "none"
	setArgumentIfPathExists("webconsoleroot", []string {"/etc/webconsole", "C:\\Program Files\\WebConsole"})
	setArgumentIfPathExists("config", []string {"config.csv", "/etc/webconsole/config.csv", "C:\\Program Files\\WebConsole\\config.csv"})
	setArgumentIfPathExists("webroot", []string {"www", "/etc/webconsole/www", "C:\\Program Files\\WebConsole\\www", ""})
	setArgumentIfPathExists("taskroot", []string {"tasks", "/etc/webconsole/tasks", "C:\\Program Files\\WebConsole\\tasks", ""})
	arguments["pathprefix"] = ""
	if len(os.Args) == 1 {
		arguments["start"] = "true"
	} else {
		arguments["start"] = "false"
	}
	
	// Parse any command line arguments.
	currentArgKey := ""
	for _, argVal := range os.Args {
		if strings.HasPrefix(argVal, "--") {
			if currentArgKey != "" {
				arguments[strings.ToLower(currentArgKey[2:])] = "true"
			}
			currentArgKey = argVal
		} else {
			if currentArgKey != "" {
				arguments[strings.ToLower(currentArgKey[2:])] = argVal
			}
			currentArgKey = ""
		}
	}
	if currentArgKey != "" {
		arguments[strings.ToLower(currentArgKey[2:])] = "true"
	}
	
	if arguments["debug"] == "true" {
		arguments["start"] = "true"
	}
	
	if arguments["start"] == "true" {
		fmt.Println("Webconsole v" + buildVersion + " - starting webserver. \"webconsole --help\" for more details.")
	}
	
	// Print the help / usage documentation if the user wanted.
	if arguments["help"] == "true" {
		//           12345678901234567890123456789012345678901234567890123456789012345678901234567890
		fmt.Println("Webconsole v" + buildVersion + ".")
		fmt.Println("")
		fmt.Println("A simple way to turn a command line application into a web app. Runs as a")
		fmt.Println("web server to host Task pages that allow the end-user to simply click a button")
		fmt.Println("to run a batch / script / etc file.")
		fmt.Println("")
		fmt.Println("Note that by itself, Webconsole doesn't handle HTTPS. If you are")
		fmt.Println("installing on a world-facing server you should use a proxy server that handles")
		fmt.Println("HTTPS - we recommend Caddy as it will automatically handle Let's Encrypt")
		fmt.Println("certificates. If you are behind a firewall then we recommend tunnelto.dev,")
		fmt.Println("giving you an HTTPS-secured URL to access. Both options can be installed via")
		fmt.Println("the install.bat / install.sh scripts.")
		fmt.Println("")
		fmt.Println("Usage: webconsole [--new] [--list] [--start] [--localOnly true/false] [--port int] [--config path] [--webroot path] [--taskroot path]")
		fmt.Println("")
		fmt.Println("--new: creates a new Task. Each Task has a unique 16-character ID which can be")
		fmt.Println("  passed as part of the URL or via a POST request, so for basic security you")
		fmt.Println("  can give a user a URL with an embedded ID. Use an external authentication")
		fmt.Println("  service for better security.")
		fmt.Println("--list: prints a list of existing Tasks.")
		fmt.Println("--start: runs as a web server, waiting for requests. Logs are printed straight to")
		fmt.Println("  stdout - hit Ctrl-C to quit. By itself, the start command can be handy for")
		fmt.Println("  quickly debugging. Run install.bat / install.sh to create a Windows service or")
		fmt.Println("  Linux / MacOS deamon.")
		fmt.Println("--debug: like \"start\", but prints more information.")
		fmt.Println("--localOnly: default is \"true\", in which case the built-in webserver will only")
		fmt.Println("  respond to requests from the local server.")
		fmt.Println("--port: the port number the web server should listen out on. Defaults to 8090.")
		fmt.Println("--config: where to find the config file. By default, on Linux this is")
		fmt.Println("  /etc/webconsole/config.csv.")
		fmt.Println("--webroot: the folder to use for the web root.")
		fmt.Println("--taskroot: the folder to use to store Tasks.")
		os.Exit(0)
	}
	
	// If we have an arument called "config", try and load the given config file (either an Excel or CSV file).
	if configPath, configFound := arguments["config"]; configFound {
		fmt.Println("Using config file: " + configPath)
		for argName, argVal := range readConfigFile(configPath) {
			arguments[argName] = argVal
		}
	}
	
	// See if we have any arguments that start with "mystart" - Page Names and API Keys for MyStart.Online login integration.
	for argName, argVal := range arguments {
		if strings.HasPrefix(argName, "mystart") {
			mystartName := ""
			if strings.HasSuffix(argName, "apikey") {
				mystartName = argName[7:len(argName)-6]
			}
			if strings.HasSuffix(argName, "pagename") {
				mystartName = argName[7:len(argName)-8]
			}
			authServiceNames["mystart"] = append(authServiceNames["mystart"], mystartName)
			if mystartName == "" {
				mystartName = "default"
			}
			if strings.HasSuffix(argName, "apikey") {
				mystartAPIKeys[mystartName] = argVal
			}
			if strings.HasSuffix(argName, "pagename") {
				mystartPageNames[mystartName] = argVal
			}
		} else if strings.HasPrefix(argName, "cloudflare") {
			if argVal != "false" {
				cloudflareName := argName[10:len(argName)]
				authServiceNames["cloudflare"] = append(authServiceNames["cloudflare"], cloudflareName)
			}
		} else if strings.HasPrefix(argName, "ngrok") {
			if argVal != "false" {
				ngrokName := argName[5:len(argName)]
				authServiceNames["ngrok"] = append(authServiceNames["ngrok"], ngrokName)
			}
		}
	}
	
	if arguments["start"] == "true" {
		// Start the thread that checks for and clears expired tokens.
		go clearExpiredTokens()
		
		// Handle the request URL.
		http.HandleFunc("/", func (theResponseWriter http.ResponseWriter, theRequest *http.Request) {
			// Make sure submitted form values are parsed.
			theRequest.ParseForm()
			
			// The default root - serve index.html.
			requestPath := theRequest.URL.Path
			
			// Print the request path.
			debug("Requested URL: " + requestPath)
			
			if strings.HasPrefix(requestPath, arguments["pathPrefix"]) {
				requestPath = requestPath[len(arguments["pathPrefix"]):]
			}
			
			userID := ""
			serveFile := false
			authorised := false
			fileToServe := filesToServeList[requestPath]
			// Handle the getPublicTaskList API call (the one API call that doesn't require authentication).
			if strings.HasPrefix(requestPath, "/api/getPublicTaskList") {
				taskList, taskErr := getTaskList()
				if taskErr == nil {
					// We return the list of public tasks in JSON format. Note that public tasks might still need authentication to run,
					// "public" here just means that they are listed by this API call for display on the landing page.
					taskListString := ""
					for _, task := range taskList {
						if task["public"] == "Y" {
							taskDetailsString, _ := json.Marshal(map[string]string{"title":task["title"], "shortDescription":task["shortDescription"], "fullDescription":task["fullDescription"], "authentication":task["authentication"]})
							taskListString = taskListString + "\"" + task["taskID"] + "\":" + string(taskDetailsString) + ","
						}
					}
					if taskListString == "" {
						fmt.Fprintf(theResponseWriter, "{}")
					} else {
						fmt.Fprintf(theResponseWriter, "{" + taskListString[:len(taskListString)-1] + "}")
					}
				} else {
					fmt.Fprintf(theResponseWriter, "ERROR: " + taskErr.Error())
				}
			// Handle a view, run or API request. If taskID is not provided as a parameter, either via GET or POST, it defaults to "/".
			} else if fileToServe != "" || strings.HasPrefix(requestPath, "/api/") {
				taskID := theRequest.Form.Get("taskID")
				token := theRequest.Form.Get("token")
				if taskID == "" {
					taskID = "/"
				}
				taskDetails, taskErr := getTaskDetails(taskID)
				if taskErr == nil {
					// If we get to this point, we know we have a valid Task ID.
					
					authorisationError := "unknown error"
					permission := "E"
					currentTimestamp := time.Now().Unix()
					rateLimit, rateLimitErr := strconv.Atoi(taskDetails["ratelimit"])
					if rateLimitErr != nil {
						rateLimit = 0
					}
					// Check through request headers - handle a login from Cloudflare's Zero Trust product or ngrok's tunneling service. Validate
					// the details passed and check that the user ID given has permission to access this Task.
					if arguments["cloudflare"] == "true" || arguments["ngrok"] == "true" {
						for headerName, headerValue := range theRequest.Header {
							if (arguments["cloudflare"] == "true" && headerName == "Cf-Access-Authenticated-User-Email") || (arguments["ngrok"] == "true" && headerName == "Ngrok-Auth-User-Email") {
								// To do - actual authentication. Assuming local-only operation, only Cloudflare / ngrok will be passing traffic anyway, but best to check.
								userID = headerValue[0]
								// Okay - we've authenticated the user, now we need to check authorisation.
								permission = getTaskPermission(arguments["webconsoleroot"], taskDetails, userID)
								if permission == "" {
									authorisationError = "authetication attempted via header authorisation (Cloudflare / ngrok), but no valid permissions granted (you're probably missing a users file)"
								} else {
									authorised = true
									//debug("User permissions granted from header " + headerName + ", ID: " + userID + ", permission: " + permission)
								}
							}
						}
					// Handle a login from MyStart.Online - validate the details passed and check that the user ID given has
					// permission to access this Task.
					} else if strings.HasPrefix(requestPath, "/api/mystartLogin") {
						mystartLoginToken := theRequest.Form.Get("loginToken")
						if mystartLoginToken != "" {
							requestURL := fmt.Sprintf("https://dev.mystart.online/api/validateToken?loginToken=%s&pageName=%s", mystartLoginToken, arguments["mystartpagename"])
							mystartResult, mystartErr := http.Get(requestURL)
							if mystartErr != nil {
								fmt.Println("webconsole: mystartLogin - error when doing callback.")
							}
							if mystartResult.StatusCode == 200 {
								defer mystartResult.Body.Close()
								mystartJSON := new(mystartStruct)
								mystartJSONResult := json.NewDecoder(mystartResult.Body).Decode(mystartJSON)
								if mystartJSONResult == nil {
									if mystartJSON.Login == "valid" {
										debug("User authenticated via MyStart.Online login, ID: " + mystartJSON.EmailHash)
										// Okay - we've authenticated the user, now we need to check authorisation.
										permission = getTaskPermission(arguments["webconsoleroot"], taskDetails, mystartJSON.EmailHash)
										if permission != "" {
											authorised = true
											userID = mystartJSON.EmailHash
											debug("User permissions granted via MyStart.Online login, ID: " + userID + ", permission: " + permission)
										}
									}
								}
							}
						} else {
							fmt.Fprintf(theResponseWriter, "ERROR: Missing parameter loginToken.")
						}
					} else if token != "" {
						if tokens[token] == 0 {
							authorisationError = "invalid or expired token"
						} else {
							authorised = true
							permission = permissions[token]
							userID = userIDs[token]
							debug("User authorised - valid token found: " + token + ", permission: " + permission + ", user ID: " + userID)
						}
					} else if checkPasswordHash(theRequest.Form.Get("secret"), taskDetails["secretViewers"]) {
						authorised = true
						permission = "V"
						debug("User authorised via Task secret, permission: " + permission)
					} else if checkPasswordHash(theRequest.Form.Get("secret"), taskDetails["secretRunners"]) {
						authorised = true
						permission = "R"
						debug("User authorised via Task secret, permission: " + permission)
					} else if checkPasswordHash(theRequest.Form.Get("secret"), taskDetails["secretEditors"]) {
						authorised = true
						permission = "E"
						debug("User authorised via Task secret, permission: " + permission)
					} else {
						authorisationError = "no external authorisation used, no valid secret given, no valid token supplied"
					}
					if !authorised && taskDetails["authentication"] == "" {
						if taskDetails["public"] == "Y" {
							debug("User authorised - no other authentication method defined, Task is public, assigning Runner permsisions.")
							permission = "R"
						} else {
							debug("User authorised - no other authentication method defined, Task is not public, assigning Viewer permsisions.")
							permission = "V"
						}
						authorised = true
						authorisationError = ""
					}
					if authorised {
						// If we get this far, we know the user is authorised for this Task - they've either provided a valid
						// secret or no secret is set.
						if token == "" {
							token = generateRandomString()
							debug("New token generated: " + token)
						}
						tokens[token] = currentTimestamp
						permissions[token] = permission
						userIDs[token] = userID
							
						// Handle view and run requests - no difference server-side, only the client-side treates the URLs differently
						// (the "runTask" method gets called by the client-side code if the URL contains "run" rather than "view").
						if fileToServe != "" {
							doServeFile(theResponseWriter, theRequest, fileToServe, taskID, token, permission, taskDetails["title"], taskDetails["shortDescription"], taskDetails["fullDescription"])
						// API - Handle a request for a list of "private" Tasks, i.e. Tasks that the user has explicit
						// authorisation to view, run or edit. We return the list of private tasks in JSON format.
						} else if strings.HasPrefix(requestPath, "/api/getPrivateTaskList") {
							taskList, taskErr := getTaskList()
							taskListString := ""
							if taskErr == nil {
								for _, task := range taskList {
									// Don't list Tasks that would already be listed in the "public" list.
									// Also, don't list special Tasks like the "new-task" Task.
									if task["public"] != "Y" && task["taskID"] != "new-task" {
										listTask := false
										// If we have Edit permissions for the root Task
										// (ID "/"), then we have permissions to view all Tasks.
										if taskID == "/" && permission == "E" {
											listTask = true
										} else {
											// Otherwise, work out permissions for each Task.
											taskPermission := getTaskPermission(arguments["webconsoleroot"], task, userID)
											if taskPermission == "V" || taskPermission == "R" || taskPermission == "E" {
												listTask = true
											}
										}
										if listTask {
											taskDetailsString, _ := json.Marshal(map[string]string{"title":task["title"], "shortDescription":task["shortDescription"], "fullDescription":task["fullDescription"], "authentication":task["authentication"]})
											taskListString = taskListString + "\"" + task["taskID"] + "\":" + string(taskDetailsString) + ","
										}
									}
								}
								if taskListString == "" {
									fmt.Fprintf(theResponseWriter, "{}")
								} else {
									fmt.Fprintf(theResponseWriter, "{" + taskListString[:len(taskListString)-1] + "}")
								}
							} else {
								fmt.Fprintf(theResponseWriter, "ERROR: " + taskErr.Error())
							}
						// API - Exchange the secret for a token.
						} else if strings.HasPrefix(requestPath, "/api/getToken") {
							fmt.Fprintf(theResponseWriter, token)
						// API - Return the Task's title.
						} else if strings.HasPrefix(requestPath, "/api/getTaskDetails") {
							fmt.Fprintf(theResponseWriter, taskDetails["title"] + "\n" + taskDetails["shortDescription"] + "\n" + taskDetails["fullDescription"])
						// API - Return the Task's result URL (or blank if it doesn't have one).
						} else if strings.HasPrefix(requestPath, "/api/getResultURL") {
							_, checkWWWErr := os.Stat(arguments["taskroot"] + "/" + taskID + "/www")
							if taskDetails["resultURL"] == "" && checkWWWErr == nil {
								fmt.Fprintf(theResponseWriter, "www")
							} else {
								fmt.Fprintf(theResponseWriter, taskDetails["resultURL"])
							}
						// API - Run a given Task.
						} else if strings.HasPrefix(requestPath, "/api/runTask") {
							// Check for appropriate permissions (Editor or Runner) for the user to be able to run the given Task.
							if permission == "E" || permission == "R" {
								// If the Task is already running, simply return "OK".
								if taskIsRunning(taskID) {
									fmt.Fprintf(theResponseWriter, "OK")
								} else {
									// Check to see if there's any rate limit set for this task, and don't run the Task if we're still
									// within the rate limited time.
									if currentTimestamp - taskStopTimes[taskID] < int64(rateLimit) {
										fmt.Fprintf(theResponseWriter, "ERROR: Rate limit (%d seconds) exceeded - try again in %d seconds.", rateLimit, int64(rateLimit) - (currentTimestamp - taskStopTimes[taskID]))
									} else {
										// Get ready to run the Task - set up the Task's details...
										if strings.HasPrefix(taskDetails["command"], "webconsole ") {
											taskDetails["command"] = strings.Replace(taskDetails["command"], "webconsole ", "\"" + arguments["webconsoleroot"] + string(os.PathSeparator) + "webconsole\" ", 1)
										} else {
											taskDetails["command"] = strings.TrimSpace(strings.TrimSpace(arguments["shellprefix"]) + " " + taskDetails["command"])
										}
										commandArray := parseCommandString(taskDetails["command"])
										var commandArgs []string
										if len(commandArray) > 0 {
											commandArgs = commandArray[1:]
										}
										debug("Task ID " + taskID + " - running command: " + commandArray[0])
										debug("With arguments: " + strings.Join(commandArgs, ","))
										debug("Ran by user: " + userID)
										
										taskRunUsers[taskID] = userID
										runningTasks[taskID] = exec.Command(commandArray[0], commandArgs...)
										runningTasks[taskID].Dir = arguments["taskroot"] + "/" + taskID
										
										// ...get a list (if available) of recent run times...
										taskRunTimes[taskID] = make([]int64, 0)
										runTimesBytes, fileErr := ioutil.ReadFile(arguments["taskroot"] + "/" + taskID + "/runTimes.txt")
										if fileErr == nil {
											runTimeSplit := strings.Split(string(runTimesBytes), "\n")
											for pl := 0; pl < len(runTimeSplit); pl = pl + 1 {
												runTimeVal, runTimeErr := strconv.Atoi(runTimeSplit[pl])
												if runTimeErr == nil {
													taskRunTimes[taskID] = append(taskRunTimes[taskID], int64(runTimeVal))
												}
											}
										}
										
										// ...use those to guess the run time for this time (just use a simple mean of the
										// existing runtimes)...
										var totalRunTime int64
										totalRunTime = 0
										for pl := 0; pl < len(taskRunTimes[taskID]); pl = pl + 1 {
											totalRunTime = totalRunTime + taskRunTimes[taskID][pl]
										}
										if len(taskRunTimes[taskID]) == 0 {
											taskRuntimeGuesses[taskID] = float64(10)
										} else {
											taskRuntimeGuesses[taskID] = float64(totalRunTime / int64(len(taskRunTimes[taskID])))
										}
										taskStartTimes[taskID] = time.Now().Unix()
										
										// ...then run the Task as a goroutine (thread) in the background.
										go runTask(taskID)
										// Respond to the front-end code that all is okay.
										fmt.Fprintf(theResponseWriter, "OK")
									}
								}
							} else {
								fmt.Fprintf(theResponseWriter, "ERROR: runTask - don't have runner / editor permissions.")
							}
						} else if strings.HasPrefix(requestPath, "/api/cancelTask") {
							// Check for appropriate permissions for the user to be able to cancel the running task...
							if permission == "E" || permission == "R" {
								// ...and that the given Task ID is actually a currently running Task...
								if taskIsRunning(taskID) {
									// and that the current user is the user that ran the Task in the first place.
									if taskRunUsers[taskID] == userID {
										if cancelErr := runningTasks[taskID].Process.Kill(); cancelErr == nil {
											delete(runningTasks, taskID)
											fmt.Fprintf(theResponseWriter, "OK")
										} else {
											fmt.Fprintf(theResponseWriter, "ERROR: cancelTask - Unable to terminate specified Task.")
										}
									} else {
										fmt.Fprintf(theResponseWriter, "ERROR: cancelTask - Current user not original runner of Task.")
									}
								} else {
									fmt.Fprintf(theResponseWriter, "ERROR: cancelTask - given Task not running.")
								}
							} else {
								fmt.Fprintf(theResponseWriter, "ERROR: cancelTask - don't have runner / editor permissions.")
							}
						// Designed to be called periodically, will return the given Tasks' output as a simple string,
						// with lines separated by newlines. Takes one parameter, "line", indicating which output line
						// it should return output from, to save the client-side code having to be sent all of the output each time.
						} else if strings.HasPrefix(requestPath, "/api/getTaskOutput") {
							var atoiErr error
							// Parse the "line" parameter - defaults to 0, so if not set this method will simply return
							// all current output.
							outputLineNumber := 0
							if theRequest.Form.Get("line") != "" {
								outputLineNumber, atoiErr = strconv.Atoi(theRequest.Form.Get("line"))
								if atoiErr != nil {
									fmt.Fprintf(theResponseWriter, "ERROR: Line number not parsable.")
								}
							}
							if _, runningTaskFound := runningTasks[taskID]; !runningTaskFound {
								// If the Task isn't currently running, load the previous run's log file (if it exists)
								// into the Task's output buffer.
								logContents, logContentsErr := ioutil.ReadFile(arguments["taskroot"] + "/" + taskID + "/log.txt")
								if logContentsErr == nil {
									taskOutputs[taskID] = strings.Split(string(logContents), "\n")
								}
							} else if taskDetails["progress"] == "Y" {
								// If the job details have the "progress" option set to "Y", output a (best guess, using previous
								// run times) progresss report line.
								currentTime := time.Now().Unix()
								percentage := int((float64(currentTime - taskStartTimes[taskID]) / taskRuntimeGuesses[taskID]) * 100)
								if percentage > 100 {
									percentage = 100
								}
								taskOutputs[taskID] = append(taskOutputs[taskID], fmt.Sprintf("Progress: Progress %d%%", percentage))
							}
							// Return to the user all the output lines from the given starting point.
							for outputLineNumber < len(taskOutputs[taskID]) {
								fmt.Fprintln(theResponseWriter, taskOutputs[taskID][outputLineNumber])
								outputLineNumber = outputLineNumber + 1
							}
							// If the Task is no longer running, make sure we tell the client-side code that.
							if _, runningTaskFound := runningTasks[taskID]; !runningTaskFound {
								if taskDetails["progress"] == "Y" {
									fmt.Fprintf(theResponseWriter, "Progress: Progress 100%%\n")
								}
								// Remove any files from the task's "output" folder that are past a defined age (in seconds - default is 129600, 36 hours).
								logfilePath := arguments["taskroot"] + "/" + taskID + "/output"
								logItems, itemErr := os.ReadDir(logfilePath)
								if itemErr == nil {
									for pl := 0; pl < len(logItems); pl = pl + 1 {
										parsedTimestamp, timestampErr := time.Parse(time.RFC3339, strings.Split(logItems[pl].Name(), "Z")[0]+"Z")
										if timestampErr == nil {
											taskLogPeriod, _ := strconv.ParseInt(taskDetails["logPeriod"], 10, 64)
											if parsedTimestamp.Unix() < taskStartTimes[taskID] - taskLogPeriod {
												os.Remove(logfilePath + "/" + logItems[pl].Name())
											}
										}
									}
								} else {
									debug("Error reading items in path: " + logfilePath)
								}
								if taskDetails["resultURL"] != "" {
									debug("Sending client resultURL: " + taskDetails["resultURL"])
									fmt.Fprintf(theResponseWriter, "ERROR: REDIRECT " + taskDetails["resultURL"])
								} else if _, err := os.Stat(arguments["taskroot"] + "/" + taskID + "/www"); err == nil {
									debug("www subfolder found, sending client redirect.")
									fmt.Fprintf(theResponseWriter, "ERROR: REDIRECT")
								} else {
									debug("Sending client EOF.")
									fmt.Fprintf(theResponseWriter, "ERROR: EOF")
								}
								//delete(taskOutputs, taskID)
							}
						// Simply returns "YES" if a given Task is running, "NO" otherwise.
						} else if strings.HasPrefix(requestPath, "/api/getTaskRunning") {
							if taskIsRunning(taskID) {
								fmt.Fprintf(theResponseWriter, "YES")
							} else {
								fmt.Fprintf(theResponseWriter, "NO")
							}
						// Submits the given string as input to the given Task. The user needs to be the runner of the Task.
						} else if strings.HasPrefix(requestPath, "/api/submitInputText") {
							// Check the user has permission (Editor or Runner) to submit inputs to the given Task...
							if permission == "E" || permission == "R" {
								// ...and that the given Task is still running...
								if taskIsRunning(taskID) {
									// ...and that the user is the runner of the Task...
									if taskRunUsers[taskID] == userID {
										// ...and that we have a value to submit.
										value := theRequest.Form.Get("value")
										if value != "" {
											debug("submitInputText - taskID: " + taskID + ", value: " + value);
											// Note - we might need to lose the newline addition on Windows.
											io.WriteString(taskInputs[taskID], value + "\n")
										} else {
											fmt.Fprintf(theResponseWriter, "ERROR: submitInputText - missing value.");
										}
									} else {
										fmt.Fprintf(theResponseWriter, "ERROR: submitInputText - current user not original runner of Task.");
									}
								} else {
									fmt.Fprintf(theResponseWriter, "ERROR: submitInputText - given Task not a current running Task.")
								}
							} else {
								fmt.Fprintf(theResponseWriter, "ERROR: submitInput called - don't have runner / editor permissions.")
							}
						// Return a list of editable files for this task, as a JSON structure - needs edit permissions.
						} else if strings.HasPrefix(requestPath, "/api/getEditableFileList") {
							if permission != "E" {
								fmt.Fprintf(theResponseWriter, "ERROR: getEditableFileList called - don't have edit permissions.")
							} else {
								path := theRequest.Form.Get("path")
								debug("getEditableFileList - path: " + path)
								if path != "" {
									outputString := "[\n"
									outputString = outputString + listOneFolderAsJSON(arguments["taskroot"] + "/" + taskID + path)
									outputString = outputString + "]"
									fmt.Fprintf(theResponseWriter, outputString)
								} else {
									fmt.Fprintf(theResponseWriter, "ERROR: getEditableFileList - missing path parameter.")
								}
							}
						// Return the contents of an editable file - needs edit permissions.
						} else if strings.HasPrefix(requestPath, "/api/getEditableFileContents") {
							if permission != "E" {
								fmt.Fprintf(theResponseWriter, "ERROR: getEditableFileContents called - don't have edit permissions.")
							} else {
								filename := theRequest.Form.Get("filename")
								if filename != "" {
									http.ServeFile(theResponseWriter, theRequest, arguments["taskroot"] + "/" + taskID + "/" + filename)
								} else {
									fmt.Fprintf(theResponseWriter, "ERROR: getEditableFileContents - missing filename parameter.")
								}
							}
						// Return the (zipped) contents of a folder - needs edit permissions.
						} else if strings.HasPrefix(requestPath, "/api/getZippedFolderContents") {
							if permission != "E" {
								fmt.Fprintf(theResponseWriter, "ERROR: getZippedFolderContents called - don't have edit permissions.")
							} else {
								filename := theRequest.Form.Get("filename")
								if filename != "" {
									// Create a buffer and writer to write the Zip file to.
									zipBuf := new(bytes.Buffer)
									zipWriter := zip.NewWriter(zipBuf)
									zipFolder := normalisePath(arguments["taskroot"] + string(os.PathSeparator) + taskID + string(os.PathSeparator) + filename)
									debug("Zipping folder: " + zipFolder)
									zipErr := getZippedFolderContents(zipWriter, zipFolder, "")
									if zipErr != nil {
										fmt.Fprintf(theResponseWriter, "ERROR: getZippedFolderContents - %s", zipErr.Error())
									} else {
										// Make sure the Zip buffer / writer is properly finished (the Zip file will be invalid otherwise).
										zipWriter.Close()
										// Return the zipped folder data to the user.
										http.ServeContent(theResponseWriter, theRequest, filename + ".zip", time.Now(), bytes.NewReader(zipBuf.Bytes()))
									}
								} else {
									fmt.Fprintf(theResponseWriter, "ERROR: getZippedFolderContents - missing filename parameter.")
								}
							}
						// Save a file.
						} else if strings.HasPrefix(requestPath, "/api/startSaveFile") {
							if permission != "E" {
								fmt.Fprintf(theResponseWriter, "ERROR: startSaveFile called - don't have edit permissions.")
							} else {
								filename := theRequest.Form.Get("filename")
								if filename != "" {
									debug("Zero: " + arguments["taskroot"] + "/" + taskID + "/" + filename)
									ioutil.WriteFile(arguments["taskroot"] + "/" + taskID + "/" + filename, []byte{}, 0644)
									fmt.Fprintf(theResponseWriter, "OK")
								} else {
									fmt.Fprintf(theResponseWriter, "ERROR: startSaveFile - missing filename parameter.")
								}
							}
						// Save a file.
						} else if strings.HasPrefix(requestPath, "/api/saveFileChunk") {
							if permission != "E" {
								fmt.Fprintf(theResponseWriter, "ERROR: saveFileChunk called - don't have edit permissions.")
							} else {
								filename := theRequest.Form.Get("filename")
								if filename != "" {
									contents := theRequest.Form.Get("contents")
									if contents != "" {
										debug("Write " + arguments["taskroot"] + "/" + taskID + "/" + filename)
										base64Contents, _ := base64.StdEncoding.DecodeString(contents)
										outfile, _ := os.OpenFile(arguments["taskroot"] + "/" + taskID + "/" + filename, os.O_APPEND|os.O_WRONLY, 0644)
										_, _ = outfile.Write(base64Contents)
										outfile.Close()
										// ioutil.WriteFile(arguments["taskroot"] + "/" + taskID + "/" + filename, []byte(base64Contents), 0644)
										fmt.Fprintf(theResponseWriter, "OK")
									} else {
										fmt.Fprintf(theResponseWriter, "ERROR: saveFileChunk - missing contents parameter.")
									}
								} else {
									fmt.Fprintf(theResponseWriter, "ERROR: saveFileChunk - missing filename parameter.")
								}
							}
						// Delete a file.
						} else if strings.HasPrefix(requestPath, "/api/deleteFile") {
							if permission != "E" {
								fmt.Fprintf(theResponseWriter, "ERROR: deleteFile called - don't have edit permissions.")
							} else {
								filename := theRequest.Form.Get("filename")
								if filename != "" {
									debug("Delete file " + arguments["taskroot"] + "/" + taskID + "/" + filename)
									os.Remove(arguments["taskroot"] + "/" + taskID + "/" + filename)
									fmt.Fprintf(theResponseWriter, "OK")
								} else {
									fmt.Fprintf(theResponseWriter, "ERROR: deleteFile - missing filename parameter.")
								}
							}
						// Delete a folder.
						} else if strings.HasPrefix(requestPath, "/api/deleteFolder") {
							if permission != "E" {
								fmt.Fprintf(theResponseWriter, "ERROR: deleteFolder called - don't have edit permissions.")
							} else {
								filename := theRequest.Form.Get("filename")
								if filename != "" {
									debug("Delete folder " + arguments["taskroot"] + "/" + taskID + "/" + filename)
									os.RemoveAll(arguments["taskroot"] + "/" + taskID + "/" + filename)
									fmt.Fprintf(theResponseWriter, "OK")
								} else {
									fmt.Fprintf(theResponseWriter, "ERROR: deleteFolder - missing filename parameter.")
								}
							}
						// Rename a file.
						} else if strings.HasPrefix(requestPath, "/api/renameFile") {
							if permission != "E" {
								fmt.Fprintf(theResponseWriter, "ERROR: renameFile called - don't have edit permissions.")
							} else {
								filename := theRequest.Form.Get("filename")
								if filename != "" {
									newFilename := theRequest.Form.Get("newFilename")
									if newFilename != "" {
										debug("Rename " + arguments["taskroot"] + "/" + taskID + "/" + filename + " to " + arguments["taskroot"] + "/" + taskID + "/" + newFilename)
										os.Rename(arguments["taskroot"] + "/" + taskID + "/" + filename, arguments["taskroot"] + "/" + taskID + "/" + newFilename)
										fmt.Fprintf(theResponseWriter, "OK")
									} else {
										fmt.Fprintf(theResponseWriter, "ERROR: renameFile - missing newFilename parameter.")
									}
								} else {
									fmt.Fprintf(theResponseWriter, "ERROR: renameFile - missing filename parameter.")
								}
							}
						// Create a new, empty, text-format file.
						} else if strings.HasPrefix(requestPath, "/api/newFile") {
							if permission != "E" {
								fmt.Fprintf(theResponseWriter, "ERROR: newFile called - don't have edit permissions.")
							} else {
								filename := theRequest.Form.Get("filename")
								if filename != "" {
									debug("New file " + arguments["taskroot"] + "/" + taskID + "/" + filename)
									newFile, newFileErr := os.OpenFile(arguments["taskroot"] + "/" + taskID + "/" + filename, os.O_RDONLY|os.O_CREATE, 0644)
									if newFileErr != nil {
										fmt.Fprintf(theResponseWriter, "ERROR: newFile - " + newFileErr.Error())
									}
									newFile.Close()
									fmt.Fprintf(theResponseWriter, "OK")
								} else {
									fmt.Fprintf(theResponseWriter, "ERROR: newFile - missing filename parameter.")
								}
							}
						// Create a new folder.
						} else if strings.HasPrefix(requestPath, "/api/newFolder") {
							if permission != "E" {
								fmt.Fprintf(theResponseWriter, "ERROR: newFolder called - don't have edit permissions.")
							} else {
								filename := theRequest.Form.Get("filename")
								if filename != "" {
									debug("New folder " + arguments["taskroot"] + "/" + taskID + "/" + filename)
									os.Mkdir(arguments["taskroot"] + "/" + taskID + "/" + filename, os.ModePerm)
									fmt.Fprintf(theResponseWriter, "OK")
								} else {
									fmt.Fprintf(theResponseWriter, "ERROR: newFolder - missing filename parameter.")
								}
							}
						// Return the hash value of the given secret value.
						} else if strings.HasPrefix(requestPath, "/api/hashSecret") {
							theSecret := theRequest.Form.Get("secret")
							hashedSecret, hashErr := hashPassword(theSecret)
							if hashErr == nil {
								hashPermission := "secret"
								if checkPasswordHash(theRequest.Form.Get("secret"), taskDetails["secretViewers"]) {
									hashPermission = "viewer"
								} else if checkPasswordHash(theRequest.Form.Get("secret"), taskDetails["secretRunners"]) {
									hashPermission = "runner"
								} else if checkPasswordHash(theRequest.Form.Get("secret"), taskDetails["secretEditors"]) {
									hashPermission = "editor"
								}
								fmt.Fprintf(theResponseWriter, hashPermission + "," + hashedSecret)
							} else {
								fmt.Fprintf(theResponseWriter, "ERROR: Problem hashing secret - " + hashErr.Error())
							}
						// A simple call that doesn't do anything except serve to keep the timestamp for the given Task up-to-date.
						} else if strings.HasPrefix(requestPath, "/api/keepAlive") {
							fmt.Fprintf(theResponseWriter, "OK")
						// To do: return API documentation here.
						} else if strings.HasPrefix(requestPath, "/api/") {
							fmt.Fprintf(theResponseWriter, "ERROR: Unknown API call: %s", requestPath)
						}
					} else if strings.HasPrefix(requestPath, "/login") {
						doServeFile(theResponseWriter, theRequest, fileToServe, taskID, "", "", taskDetails["title"], taskDetails["shortDescription"], taskDetails["fullDescription"])
					} else {
						fmt.Fprintf(theResponseWriter, "ERROR: Not authorised - %s.", authorisationError)
					}
				} else {
					fmt.Fprintf(theResponseWriter, "ERROR: %s", taskErr.Error())
				}
			} else if strings.HasSuffix(requestPath, "/site.webmanifest") {
				taskID := ""
				taskList, taskErr := getTaskList()
				if taskErr == nil {
					for _, task := range taskList {
						if strings.HasPrefix(requestPath, "/" + task["taskID"]) {
							taskID = task["taskID"] + "/"
						}
					}
				} else {
					fmt.Fprintf(theResponseWriter, "ERROR: " + taskErr.Error())
				}
				webmanifestBuffer, fileReadErr := ioutil.ReadFile(arguments["webroot"] + "/" + "site.webmanifest")
				if fileReadErr == nil {
					webmanifestString := string(webmanifestBuffer)
					webmanifestString = strings.Replace(webmanifestString, "<<TASKID>>", arguments["pathPrefix"] + "/" + taskID, -1)
					http.ServeContent(theResponseWriter, theRequest, "site.webmanifest", time.Now(), strings.NewReader(webmanifestString))
				} else {
					fmt.Fprintf(theResponseWriter, "ERROR: Couldn't read site.webmanifest.")
				}
			} else {
				// Check to see if the request is for a favicon of some description.
				faviconTitle := ""
				faviconHyphens := 0
				faviconTitles := [7]string{ "favicon.*png", "mstile.*png", "android-chrome.*png", "apple-touch-icon.*png", "safari-pinned-tab.png", "safari-pinned-tab.svg", "favicon.ico" }
				for _, titleMatch := range faviconTitles {
					requestMatch, _ := regexp.MatchString(".*/" + titleMatch + "$", requestPath)
					if requestMatch {
						faviconTitle = titleMatch
						faviconHyphens = strings.Count(titleMatch, "-") + 1
					}
				}
				// If the request was for a favicon, serve something suitible.
				if faviconTitle != "" {
					faviconPath := arguments["webroot"] + "/" + "favicon.png"
					taskList, taskErr := getTaskList()
					if taskErr == nil {
						for _, task := range taskList {
							if strings.HasPrefix(requestPath, "/" + task["taskID"]) {
								// Does this Task have a custom favicon?
								faviconPath = arguments["taskroot"] + "/" + task["taskID"] + "/" + "favicon.png"
								if _, fileExistsErr := os.Stat(faviconPath); os.IsNotExist(fileExistsErr) {
									// Does all Tasks have a custom favicon?
									faviconPath = arguments["taskroot"] + "/" + "favicon.png"
									if _, fileExistsErr := os.Stat(faviconPath); os.IsNotExist(fileExistsErr) {
										// If there is no custom favicon set for this Task, use the default.
										faviconPath = arguments["webroot"] + "/" + "favicon.png"
									}
								}
							}
						}
					} else {
						fmt.Fprintf(theResponseWriter, "ERROR: " + taskErr.Error())
					}
					faviconFile, faviconFileErr := os.Open(faviconPath)
					if faviconFileErr == nil {
						serveFile = true
						faviconImage, _, faviconImageErr := image.Decode(faviconFile)
						faviconFile.Close()
						if faviconImageErr == nil {
							faviconWidth := faviconImage.Bounds().Max.X
							faviconHeight := faviconImage.Bounds().Max.Y
							if faviconTitle == "safari-pinned-tab.png" || faviconTitle == "safari-pinned-tab.svg" {
								silhouetteImage := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{faviconWidth, faviconHeight}})
								for silhouetteY := 0; silhouetteY < faviconHeight; silhouetteY++ {
									for silhouetteX := 0; silhouetteX < faviconWidth; silhouetteX++ {
										r, g, b, a := faviconImage.At(silhouetteX, silhouetteY).RGBA()
										if r > 128 || g > 128 || b > 128 || a < 255 {
											silhouetteImage.Set(silhouetteX, silhouetteY, color.RGBA{255, 255, 255, 0})
										} else {
											silhouetteImage.Set(silhouetteX, silhouetteY, color.RGBA{0, 0, 0, 255})
										}
									}
								}
								if faviconTitle == "safari-pinned-tab.png" {
									pngErr := png.Encode(theResponseWriter, silhouetteImage)
									if pngErr != nil {
										fmt.Fprintf(theResponseWriter, "ERROR: Unable to encode PNG silhouette image.\n")
									}
								} else {
									tracedImage, _ := gotrace.Trace(gotrace.NewBitmapFromImage(silhouetteImage, nil), nil)
									theResponseWriter.Header().Set("Content-Type", "image/svg+xml")
									gotrace.WriteSvg(theResponseWriter, silhouetteImage.Bounds(), tracedImage, "")
								}
								serveFile = false
							} else {
								if faviconTitle == "apple-touch-icon.png" {
									faviconWidth = 180
									faviconHeight = 180
								} else if faviconTitle == "favicon.ico" {
									faviconWidth = 48
									faviconHeight = 48
								}
								// Resize the available (PNG) favicon to match the request.
								faviconSplit := strings.Split(requestPath, "/")
								faviconName := strings.Split(faviconSplit[len(faviconSplit)-1], ".")[0]
								faviconSplit = strings.Split(faviconName, "-")
								if len(faviconSplit) != faviconHyphens {
									faviconSizeSplit := strings.Split(faviconSplit[faviconHyphens], "x")
									if len(faviconSizeSplit) == 2 {
										var atoiError error
										faviconWidth, atoiError = strconv.Atoi(faviconSizeSplit[0])
										if atoiError == nil {
											faviconHeight, atoiError = strconv.Atoi(faviconSizeSplit[1])
										}
										if atoiError != nil {
											fmt.Fprintf(theResponseWriter, "ERROR: Non-integer in image dimensions.\n")
											serveFile = false
										}
									}
								}
								resizedImage := resize.Resize(uint(faviconWidth), uint(faviconHeight), faviconImage, resize.Lanczos3)
								if strings.HasSuffix(faviconTitle, "ico") {
									icoErr := ico.Encode(theResponseWriter, resizedImage)
									if icoErr != nil {
										fmt.Fprintf(theResponseWriter, "ERROR: Unable to encode PNG image.\n")
									}
									serveFile = false
								} else {
									pngErr := png.Encode(theResponseWriter, resizedImage)
									if pngErr != nil {
										fmt.Fprintf(theResponseWriter, "ERROR: Unable to encode PNG image.\n")
									}
									serveFile = false
								}
							}
						} else {
							fmt.Fprintf(theResponseWriter, "ERROR: Couldn't decode favicon file: " + faviconPath + "\n")
						}
					} else {
						fmt.Fprintf(theResponseWriter, "ERROR: Couldn't open favicon file: " + faviconPath + "\n")
					}
				// ...otherwise, just serve the static file referred to by the request URL.
				} else {
					serveFile = true
				}
			}
			if serveFile == true {
				localFilePath := ""
				taskList, taskErr := getTaskList()
				if taskErr == nil {
					for _, task := range taskList {
						if strings.HasPrefix(requestPath, "/" + task["taskID"]) && serveFile == true {
							var filePath = strings.TrimSpace(requestPath[len(task["taskID"])+1:])
							if filePath == "" {
								http.Redirect(theResponseWriter, theRequest, "/" + task["taskID"] + "/", http.StatusPermanentRedirect)
								serveFile = false
							}
							if strings.HasSuffix(filePath, "/") {
								filePath = filePath + "index.html"
							}
							localFilePath = arguments["taskroot"] + "/" + task["taskID"] + "/www" + filePath
						}
					}
				} else {
					fmt.Fprintf(theResponseWriter, "ERROR: " + taskErr.Error())
					serveFile = false
				}
				if serveFile == true {
					// Serve a static file. A file found in the appropriate "webconsole/users/username" folder will override one found in "webconsole/www",
					// so you can serve user-specific content for authenticated users simply by placing files in the appropriate user subfolder.
					if localFilePath == "" {
						usersFilePath := arguments["webconsoleroot"] + "/users/" +  userID + requestPath
						if _, err := os.Stat(usersFilePath); !errors.Is(err, os.ErrNotExist) {
							localFilePath = usersFilePath
						} else {
							localFilePath = arguments["webroot"] + requestPath
						}
					}
					if _, err := os.Stat(localFilePath); !errors.Is(err, os.ErrNotExist) {
						debug("Serving: " + localFilePath)
						http.ServeFile(theResponseWriter, theRequest, localFilePath)
					} else {
						theResponseWriter.WriteHeader(http.StatusNotFound)
						errorFileContent, errorFileErr := ioutil.ReadFile(arguments["webroot"] + "/404.html")
						if errorFileErr == nil {
							fmt.Fprint(theResponseWriter, strings.Replace(string(errorFileContent), "{{FILENAME}}", localFilePath, -1))
						} else {
							fmt.Fprint(theResponseWriter, "Error 404: File " + localFilePath + " not found.")
						}
					}
				}
			}
		})
		// Run the main web server loop.
		hostname := ""
		if (arguments["localonly"] == "true") {
			fmt.Println("Web server limited to localhost only.")
			hostname = "localhost"
		}
		fmt.Println("Web server using webroot " + arguments["webroot"] + ", taskroot " + arguments["taskroot"] + ".")
		fmt.Println("Web server available at: http://localhost:" + arguments["port"] + "/")
		// If we detect ngrok is running, search the syslog for the ngrok URL to display to the user.
		if _, err := os.Stat("/usr/local/bin/ngrok"); err == nil {
			ngrokURL, ngrokErr := exec.Command("bash", "-c", "cat /var/log/syslog | grep ngrok.*localhost | tail -1 | cut -d '=' -f 8").CombinedOutput()
			if ngrokErr != nil {
				fmt.Println("ERROR: " + ngrokErr.Error())
			}
			fmt.Println("ngrok URL: " + strings.TrimSpace(string(ngrokURL)))
		}
		if arguments["debug"] == "true" {
			fmt.Println("Debug mode set - arguments:")
			for argName, argVal := range arguments {
				fmt.Println("   " + argName + ": " + argVal)
			}
		}
		log.Fatal(http.ListenAndServe(hostname + ":" + arguments["port"], nil))
	// Command-line option to print a list of all Tasks.
	} else if arguments["list"] == "true" {
		fmt.Println("Reading Tasks from " + arguments["taskroot"])
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
	// Generate a new Task.
	} else if arguments["newdefaulttask"] == "true" {
		// Generate a new Task ID, checking it doesn't already exist.
		var newTaskID string
		for {
			newTaskID = generateRandomString()
			if _, err := os.Stat(arguments["taskroot"] + "/" + newTaskID); os.IsNotExist(err) {
				break
			}
		}
		
		os.Mkdir(arguments["taskroot"], os.ModePerm)
		os.Mkdir(arguments["taskroot"] + "/" + newTaskID, os.ModePerm)
		fmt.Println("New Task: " + newTaskID)
		
		// Write the config file - a simple text file, one value per line.
		outputString := "title: Task " + newTaskID + "\npublic: N\ncommand: "
		writeFileErr := ioutil.WriteFile(arguments["taskroot"] + "/" + newTaskID + "/config.txt", []byte(outputString), 0644)
		if writeFileErr != nil {
			fmt.Println("ERROR: Couldn't write config for Task " + newTaskID + ".")
		}
	} else if arguments["new"] == "true" {
		// Generate a new, unique Task ID.
		var newTaskID string
		var newTaskIDExists bool
		// Ask the user to provide a Task ID (or they can use the one we just generated).
		if newTaskID, newTaskIDExists = arguments["newtaskid"]; !newTaskIDExists {
			for {
				newTaskID = generateRandomString()
				if _, err := os.Stat(arguments["taskroot"] + "/" + newTaskID); os.IsNotExist(err) {
					break
				}
			}
			newTaskID = getUserInput("newtaskid", newTaskID, "Enter a new Task ID (hit enter to generate an ID)")
		}
		if _, err := os.Stat(arguments["taskroot"] + "/" + newTaskID); os.IsNotExist(err) {
			// We use simple text files in folders for data storage, rather than a database. It seemed the most logical choice - you can stick
			// any resources associated with a Task in that Task's folder, and editing options can be done with a basic text editor.
			os.Mkdir(arguments["taskroot"], os.ModePerm)
			os.Mkdir(arguments["taskroot"] + "/" + newTaskID, os.ModePerm)
			fmt.Println("New Task: " + newTaskID)
			
			// Get a title for the Task.
			newTaskTitle := "Task " + newTaskID
			newTaskTitle = getUserInput("newtasktitle", newTaskTitle, "Enter a title (hit enter for \"" + newTaskTitle + "\")")
			
			// Get a secret for the Task - blank by default, although that's not the same as a public Task.
			newTaskSecret := ""
			newTaskSecret = getUserInput("newtasksecret", newTaskSecret, "Set secret (type secret, or hit enter to skip)")
			
			// Ask the user if this Task should be public, "N" by default.
			var newTaskPublic string
			for {
				newTaskPublic = "N"
				newTaskPublic = strings.ToUpper(getUserInput("newtaskpublic", newTaskPublic, "Make this task public (\"Y\" or \"N\", hit enter for \"N\")"))
				if newTaskPublic == "Y" || newTaskPublic == "N" {
					break
				}
			}
			
			// The command the Task runs. Can be anything the system will run as an executable application, which of course depends on which platform
			// you are running.
			newTaskCommand := ""
			newTaskCommand = getUserInput("newtaskcommand", newTaskCommand, "Set command (type command, or hit enter to skip)")
			
			// Hash the secret (if not just blank).
			outputString := ""
			if newTaskSecret != "" {
				hashedPassword, hashErr := hashPassword(newTaskSecret)
				if hashErr == nil {
					outputString = outputString + "secret: " + hashedPassword + "\n"
				} else {
					fmt.Println("ERROR: Problem hashing password - " + hashErr.Error())
				}
			}
			
			// Write the config file - a simple text file, one value per line.
			outputString = outputString + "title: " + newTaskTitle + "\npublic: " + newTaskPublic + "\ncommand: " + newTaskCommand
			writeFileErr := ioutil.WriteFile(arguments["taskroot"] + "/" + newTaskID + "/config.txt", []byte(outputString), 0644)
			if writeFileErr != nil {
				fmt.Println("ERROR: Couldn't write config for Task " + newTaskID + ".")
			}
		} else {
			fmt.Println("ERROR: A task with ID " + newTaskID + " already exists.")
		}		
	}
}
