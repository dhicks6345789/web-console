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

// The main web server loop - the part that serves files and responds to API calls.
func handleAPI(theResponseWriter http.ResponseWriter, theRequest *http.Request) {
	fmt.Fprintf(theResponseWriter, "API call: %s!", theRequest.URL.Path[1:])
}

// The main body of the program - parse user-provided command-line paramaters, or start the main web server process.
func main() {
	var tasks []string
	items, err := ioutil.ReadDir("tasks")
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range items {
		tasks = append(tasks, item.Name())
	}
	if len(os.Args) == 1 {
		// If no parameters are given, simply start the web server.
		fmt.Println("Starting web server...")
		
		// Handle the "/api/" route.
		http.HandleFunc("/api/", handleAPI)
		
		// Handle the "/" (default, "everything else") route - just try and serve the given path as a static file.
		staticFilesServer := http.FileServer(http.Dir("www"))
		http.Handle("/", staticFilesServer)
		
		http.ListenAndServe(":8090", nil)
	} else if os.Args[1] == "-list" {
		// Print a list of existing IDs.
		fmt.Println("List:")
		for _, task := range tasks {
			fmt.Println(task)
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
