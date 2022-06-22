package main
// Web Console - lets end-users run command-line applications via a web page, complete with authentication and a user interface.
// For more details, see https://www.sansay.co.uk/docs/web-console

import (
	// Standard libraries.
	"io"
	"fmt"
	"os"
	"log"
	"sort"
	"time"
	"bufio"
	"regexp"
	"errors"
	"image"
	"image/png"
	"image/color"
	"strings"
	"strconv"
	"os/exec"
	"net/http"
	"math/rand"
	"io/ioutil"
	"encoding/csv"
	"encoding/json"
	
	// Image resizing library.
	"github.com/nfnt/resize"
	
	// An Image-to-SVG tracing library.
	"github.com/dennwc/gotrace"
	
	// A .ICO format Image encoder.
	"github.com/kodeworks/golang-image-ico"
	
	// Bcrypt for password hashing.
	"golang.org/x/crypto/bcrypt"
	
	// Excelize for loading in Excel files.
	"github.com/360EntSecGroup-Skylar/excelize"
)

// Characters to use to generate new ID strings. Lowercase only - any user-provided IDs will be lowercased before use.
const letters = "abcdefghijklmnopqrstuvwxyz1234567890"

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
// ...and matching permissions.
var permissions = map[string]string{}

// A list of currently running Tasks.
var runningTasks = map[string]*exec.Cmd{}
// The outputs from Tasks.
var taskOutputs = map[string][]string{}
// We record the start time and an array of recent runtimes for each Task so we can guess at this run's liklely time and print a progress report if wanted.
var taskStartTimes = map[string]int64{}
var taskRunTimes = map[string][]int64{}
var taskRuntimeGuesses = map[string]float64{}
// We record the stop time for each Task so we can implement rate limiting.
var taskStopTimes = map[string]int64{}

// Maps of MyStart.Online page names and API keys.
var mystartPageNames = map[string]string{}
var mystartAPIKeys = map[string]string{}

// A struct used to read JSON data from authentication API calls to MyStart.Online.
type mystartStruct struct {
	Login string
	EmailHash string
	EmailDomain string
	LoginType string
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
		return true
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
	readBuffer := make([]byte, 10240)
	taskOutputs[theTaskID] = make([]string, 0)
	taskStdout, taskStdoutErr := runningTasks[theTaskID].StdoutPipe()
	if taskStdoutErr == nil {
		taskStderr, taskStderrErr := runningTasks[theTaskID].StderrPipe()
		if taskStderrErr == nil {
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
						errorString := "ERROR: " + exitErr.Error() + "\n"
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
					// Remove this Task from the runnings Tasks list. We don't remove the output right away - client-side code might
					// still not have received all the output yet.
					delete(runningTasks, theTaskID)
				}
				logfileOutput.Close()
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
	configPath := arguments["taskroot"] + "/" + theTaskID + "/config.txt"
	// Check to see if we have a valid task ID.
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		inFile, inFileErr := os.Open(configPath)
		if inFileErr != nil {
			return taskDetails, errors.New("Can't open Task config file.")
		} else {
			// Read the Task's details from its config file.
			taskDetails["taskID"] = theTaskID
			taskDetails["title"] = ""
			taskDetails["description"] = ""
			taskDetails["secret"] = ""
			taskDetails["public"] = "N"
			taskDetails["ratelimit"] = "0"
			taskDetails["progress"] = "N"
			taskDetails["command"] = ""
			scanner := bufio.NewScanner(inFile)
			for scanner.Scan() {
				itemSplit := strings.SplitN(scanner.Text(), ":", 2)
				taskDetails[strings.TrimSpace(itemSplit[0])] = strings.TrimSpace(itemSplit[1])
			}
			inFile.Close()
			descriptionContents, descriptionContentsErr := ioutil.ReadFile(arguments["taskroot"] + "/" + theTaskID + "/description.txt")
			if descriptionContentsErr == nil {
				taskDetails["description"] = string(descriptionContents)
			}
		}
	} else {
		return taskDetails, errors.New("No config file for taskID: " + theTaskID)
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
					result[csvDataRecord[0]] = csvDataRecord[1]
				}
			}
		} else {
			fmt.Println("ERROR: " + csvErr.Error())
		}
	}
	return result
}

func readUserFile(theConfigPath string) map[string]string {
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
					hashedEmailAddress := csvDataRecord[0]
					emailAddressIsHash := true
					if len(hashedEmailAddress) == 32 {
						for _, addressCharValue := range hashedEmailAddress {
							if !strings.Contains(letters, string(addressCharValue)) {
								emailAddressIsHash = false
							}
						}
					} else {
						emailAddressIsHash = false
					}
					emailAddress := ""
					if emailAddressIsHash {
						if len(csvDataRecord) > 1 {
							emailAddress = csvDataRecord[1]
						}
					} else {
						emailAddress = hashedEmailAddress
						hashedEmailAddress = ""
					}
					result[hashedEmailAddress] = emailAddress
				}
			}
		} else {
			fmt.Println("ERROR: " + csvErr.Error())
		}
	}
	return result
}

// The main body of the program - parse user-provided command-line paramaters, or start the main web server process.
func main() {
	// This application is both a web server for handling API requests and displaying a web-based front end, and a command-line application for handling
	// configuration and setup.
	
	// Set some default argument values.
	arguments["help"] = "false"
	arguments["start"] = "true"
	arguments["list"] = "false"
	arguments["new"] = "false"
	arguments["port"] = "8090"
	arguments["localOnly"] = "true"
	arguments["debug"] = "false"
	setArgumentIfPathExists("config", []string {"config.csv", "/etc/webconsole/config.csv", "C:\\Program Files\\WebConsole\\config.csv"})
	setArgumentIfPathExists("webroot", []string {"www", "/etc/webconsole/www", "C:\\Program Files\\WebConsole\\www", ""})
	setArgumentIfPathExists("taskroot", []string {"tasks", "/etc/webconsole/tasks", "C:\\Program Files\\WebConsole\\tasks", ""})
	arguments["pathPrefix"] = ""
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
		fmt.Println("Webconsole - starting webserver. \"webconsole --help\" for more details.")
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
			if strings.HasSuffix(argName, "APIKey") {
				mystartName = argName[7:len(argName)-6]
			}
			if strings.HasSuffix(argName, "PageName") {
				mystartName = argName[7:len(argName)-8]
			}
			if mystartName == "" {
				mystartName = "default"
			}
			if strings.HasSuffix(argName, "APIKey") {
				mystartAPIKeys[mystartName] = argVal
			}
			if strings.HasSuffix(argName, "PageName") {
				mystartPageNames[mystartName] = argVal
			}
		}
	}
	if arguments["debug"] == "true" {
		fmt.Println("webconsole: MyStart.Online Page Names / API Keys:")
		fmt.Println(mystartPageNames)
		fmt.Println(mystartAPIKeys)
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
			if arguments["debug"] == "true" {
				fmt.Println("webconsole: " + requestPath)
			}
			
			if strings.HasPrefix(requestPath, arguments["pathPrefix"]) {
				requestPath = requestPath[len(arguments["pathPrefix"]):]
			}
			
			serveFile := false
			if requestPath == "/" {
				http.ServeFile(theResponseWriter, theRequest, arguments["webroot"] + "/index.html")
			// Handle the getPublicTaskList API call (the one API call that doesn't require authentication).
			} else if strings.HasPrefix(requestPath, "/api/getPublicTaskList") {
				taskList, taskErr := getTaskList()
				if taskErr == nil {
					// We return the list of public tasks in JSON format. Note that public tasks might still need a secret to run, "public"
					// here just means that they are listed by this API call for display on the landing page.
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
			} else if strings.HasPrefix(requestPath, "/view") || strings.HasPrefix(requestPath, "/run") || strings.HasPrefix(requestPath, "/api/") {
				taskID := theRequest.Form.Get("taskID")
				token := theRequest.Form.Get("token")
				if taskID == "" {
					fmt.Fprintf(theResponseWriter, "ERROR: Missing parameter taskID.")
				} else {
					// If we get to this point, we know we have a valid Task ID.
					taskDetails, taskErr := getTaskDetails(taskID)
					if taskErr == nil {
						authorised := false
						authorisationError := "unknown error"
						permission := "E"
						currentTimestamp := time.Now().Unix()
						rateLimit, rateLimitErr := strconv.Atoi(taskDetails["ratelimit"])
						if rateLimitErr != nil {
							rateLimit = 0
						}
						// Handle a login from MyStart.Online - validate the details passed and check that the user ID given has
						// permission to access this Task.
						if strings.HasPrefix(requestPath, "/api/mystartLogin") {
							if arguments["debug"] == "true" {
								fmt.Println("webconsole: mystartLogin API call.")
							}
							mystartLoginToken := theRequest.Form.Get("loginToken")
							if mystartLoginToken != "" {
								requestURL := fmt.Sprintf("https://dev.mystart.online/api/validateToken?loginToken=%s&pageName=%s", mystartLoginToken, arguments["mystartPageName"])
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
											// Okay - we've authenticated the user, now we need to check authorisation.
											fmt.Println(mystartJSON)
											for taskDetailName, taskDetailValue := range taskDetails {
												if strings.HasPrefix(taskDetailName, "mystart") {
													mystartName := ""
													if strings.HasSuffix(taskDetailName, "Editors") {
														mystartName = taskDetailName[7:len(taskDetailName)-7]
													}
													if mystartName == "" {
														mystartName = "default"
													}
													if strings.HasSuffix(taskDetailName, "Editors") {
														mystartEditorsPath := arguments["taskroot"] + "/" + taskID + "/" + taskDetailValue
														if arguments["debug"] == "true" {
															fmt.Println("webconsole: Looking for MyStart.Online (" + mystartName + ") Editors data in: " + mystartEditorsPath)
														}
														mystartEditors := readUserFile(mystartEditorsPath)
														for editorHash, editorEmail := range mystartEditors {
															if editorHash == mystartJSON.EmailHash {
																authorised = true
																permission = "E"
																if arguments["debug"] == "true" {
																	fmt.Println("webconsole: User authorised via MyStart.Online login, hash: " + editorHash + ", email: " + editorEmail + ", permission: " + permission)
																}
															}
														}
													}
												}
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
								if arguments["debug"] == "true" {
									fmt.Println("webconsole: User authorised - valid token found: " + token + ", permission: " + permission)
								}
							}
						} else if checkPasswordHash(theRequest.Form.Get("secret"), taskDetails["secret"]) {
							authorised = true
							permission = "E"
							if arguments["debug"] == "true" {
								fmt.Println("webconsole: User authorised via Task secret, permission: " + permission)
							}
						} else {
							authorisationError = "incorrect secret"
						}
						if authorised {
							// If we get this far, we know the user is authorised for this Task - they've either provided a valid
							// secret or no secret is set.
							if token == "" {
								token = generateRandomString()
								if arguments["debug"] == "true" {
									fmt.Println("webconsole: New token generated: " + token)
								}
							}
							tokens[token] = currentTimestamp
							permissions[token] = permission
							
							fileToServe := "webconsole.html"
							if strings.HasPrefix(requestPath, "/api/mystartLogin") {
								fileToServe = "redirect.html"
							}
							// Handle view and run requests - no difference server-side, only the client-side treates the URLs differently
							// (the "runTask" method gets called by the client-side code if the URL contains "run" rather than "view").
							if strings.HasPrefix(requestPath, "/view") || strings.HasPrefix(requestPath, "/run") || strings.HasPrefix(requestPath, "/api/mystartLogin") {
								// Serve the "fileToServe" file, first adding in the Task ID and token values to be used client-side, as well
								// as including the appropriate formatting.js file.
								webconsoleBuffer, fileReadErr := ioutil.ReadFile(arguments["webroot"] + fileToServe)
								if fileReadErr == nil {
									formattingJSBuffer, fileReadErr := ioutil.ReadFile(arguments["taskroot"] + "/" + taskID + "/formatting.js")
									if fileReadErr != nil {
										formattingJSBuffer, fileReadErr = ioutil.ReadFile(arguments["taskroot"] + "/formatting.js")
										if fileReadErr != nil {
											formattingJSBuffer, fileReadErr = ioutil.ReadFile(arguments["webroot"] + "/formatting.js")
										}
									}
									if fileReadErr == nil {
										formattingJSString := string(formattingJSBuffer)
										webconsoleString := string(webconsoleBuffer)
										webconsoleString = strings.Replace(webconsoleString, "<<TASKID>>", taskID, -1)
										webconsoleString = strings.Replace(webconsoleString, "<<TOKEN>>", token, -1)
										webconsoleString = strings.Replace(webconsoleString, "<<TITLE>>", taskDetails["title"], -1)
										webconsoleString = strings.Replace(webconsoleString, "<<DESCRIPTION>>", taskDetails["description"], -1)
										webconsoleString = strings.Replace(webconsoleString, "<<FAVICONPATH>>", taskID + "/", -1)
										webconsoleString = strings.Replace(webconsoleString, "// Include formatting.js.", formattingJSString, -1)
										http.ServeContent(theResponseWriter, theRequest, fileToServe, time.Now(), strings.NewReader(webconsoleString))
									} else {
										fmt.Fprintf(theResponseWriter, "ERROR: Couldn't read formatting.js")
									}
								} else {
									fmt.Fprintf(theResponseWriter, "ERROR: Couldn't read " + arguments["webroot"] + fileToServe)
								}
							// API - Exchange the secret for a token.
							} else if strings.HasPrefix(requestPath, "/api/getToken") {
								fmt.Fprintf(theResponseWriter, token)
							// API - Return the Task's title.
							} else if strings.HasPrefix(requestPath, "/api/getTaskDetails") {
								fmt.Fprintf(theResponseWriter, taskDetails["title"] + "\n" + taskDetails["description"])
							// API - Run a given Task.
							} else if strings.HasPrefix(requestPath, "/api/runTask") {
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
										commandArray := parseCommandString(taskDetails["command"])
										for _, batchExtension := range []string{".bat", ".btm", ".cmd"} {
											// If the command is a Windows batch file, we need to run the Windows command shell for it to execute.
											if strings.HasSuffix(strings.ToLower(commandArray[0]), batchExtension) {
												commandArray = parseCommandString("cmd /c " + taskDetails["command"])
											}
										}
										if arguments["debug"] == "true" {
											fmt.Println("Running command: " + strings.Join(commandArray, " "))
										}
										var commandArgs []string
										if len(commandArray) > 0 {
											commandArgs = commandArray[1:]
										}
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
									if _, err := os.Stat(arguments["taskroot"] + "/" + taskID + "/www"); err == nil {
										fmt.Fprintf(theResponseWriter, "ERROR: REDIRECT")
									} else {
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
							// A simple call that doesn't do anything except serve to keep the timestamp for the given Task up-to-date.
							} else if strings.HasPrefix(requestPath, "/api/keepAlive") {
								fmt.Fprintf(theResponseWriter, "OK")
							// To do: return API documentation here.
							} else if strings.HasPrefix(requestPath, "/api/") {
								fmt.Fprintf(theResponseWriter, "ERROR: Unknown API call: %s", requestPath)
							}
						} else {
							fmt.Fprintf(theResponseWriter, "ERROR: Not authorised - %s.", authorisationError)
						}
					} else {
						fmt.Fprintf(theResponseWriter, "ERROR: %s", taskErr.Error())
					}
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
				taskList, taskErr := getTaskList()
				if taskErr == nil {
					for _, task := range taskList {
						if strings.HasPrefix(requestPath, "/" + task["taskID"]) && serveFile == true {
							var filePath = strings.TrimSpace(requestPath[17:])
							if filePath == "" {
								filePath = "/"
							}
							if strings.HasSuffix(filePath, "/") {
								filePath = filePath + "index.html"
							}
							http.ServeFile(theResponseWriter, theRequest, arguments["taskroot"] + "/" + task["taskID"] + "/www" + filePath)
							serveFile = false
						}
					}
				} else {
					fmt.Fprintf(theResponseWriter, "ERROR: " + taskErr.Error())
					serveFile = false
				}
				if serveFile == true {
					http.ServeFile(theResponseWriter, theRequest,  arguments["webroot"] + requestPath)
				}
			}
		})
		// Run the main web server loop.
		hostname := ""
		if (arguments["localOnly"] == "true") {
			fmt.Println("Web server limited to localhost only.")
			hostname = "localhost"
		}
		fmt.Println("Web server using webroot " + arguments["webroot"] + ", taskroot " + arguments["taskroot"] + ".")
		fmt.Println("Web server available at: http://localhost:" + arguments["port"] + "/")
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
