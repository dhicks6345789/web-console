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
	"strings"
	"math/rand"
	"io/ioutil"
	"net/http"
)

// Characters to use to generate new ID strings. Lowercase only - any user-provided IDs will be lowercased before use.
const letters = "abcdefghijklmnopqrstuvwxyz1234567890"

// Generate a new, random 16-character ID.
func generateTaskID() string {
	rand.Seed(time.Now().UnixNano())
	result := make([]byte, 16)
	for pl := range result {
		result[pl] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}

func getParameter(theRequest *http.Request, theParameterName) string {
	keys, ok := r.URL.Query()[theParameterName]
	
	if !ok || len(keys[0]) < 1 {
		//log.Println("Required parameter " + theParameterName + " is missing.")
		return nil
	}
	return keys[0]
}

// The main body of the program - parse user-provided command-line paramaters, or start the main web server process.
func main() {
	if len(os.Args) == 1 {
		// If no parameters are given, simply start the web server.
		fmt.Println("Starting web server...")
		
		// We write our own function to parse the request URL.
		http.HandleFunc("/", func (theResponseWriter http.ResponseWriter, theRequest *http.Request) {
			// The default root - serve index.html.
			if theRequest.URL.Path == "/" {
				http.ServeFile(theResponseWriter, theRequest, "www/index.html")
			// If the URL matches a task ID, still serve webconsole.html.
			} else if _, err := os.Stat("tasks" + theRequest.URL.Path); !os.IsNotExist(err) {
				fmt.Println("Run task: " + theRequest.URL.Path)
				http.ServeFile(theResponseWriter, theRequest, "www/webconsole.html")
			// Handle API calls.
			} else if strings.HasPrefix(theRequest.URL.Path, "/api/viewTask") {
				taskID := getParameter(theRequest, "taskID")
				if !taskID == nil {
					fmt.Fprintf(theResponseWriter, "View task: %s", taskID)
				}
			} else if strings.HasPrefix(theRequest.URL.Path, "/api/") {
				fmt.Fprintf(theResponseWriter, "API call: %s", theRequest.URL.Path)
			// Otherwise, try and find the static file referred to by the request URL.
			} else {
				http.ServeFile(theResponseWriter, theRequest, "www" + theRequest.URL.Path)
			}
		})
		log.Fatal(http.ListenAndServe(":8090", nil))
	} else if os.Args[1] == "-list" {
		// Print a list of existing IDs.
		items, err := ioutil.ReadDir("tasks")
		if err != nil {
			log.Fatal(err)
		}
		for _, item := range items {
			fmt.Println(item.Name())
		}
	} else if os.Args[1] == "-generate" {
		// Generate a new task ID, and create a matching folder.
		for {
			newTaskID := generateTaskID()
			if _, err := os.Stat("tasks/" + newTaskID); os.IsNotExist(err) {
				os.Mkdir("tasks/" + newTaskID, os.ModePerm)
				fmt.Println("New Task generated: " + newTaskID)
				break
			}
		}
	}
}
