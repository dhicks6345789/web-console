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

var filesToServe = [...]string {"index.html"}

func arrayContains(theArray []string, testItem string) bool {
	for _, item := range theArray {
		if item == testItem {
			return true
		}
	}
	return false
}

// Generate a new, random 16-character ID.
func generateID() string {
	rand.Seed(time.Now().UnixNano())
	result := make([]byte, 16)
	for pl := range result {
		result[pl] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}

// The main web server loop - the part that serves files and responds to API calls.
func webConsole(theResponseWriter http.ResponseWriter, theRequest *http.Request) {
	if arrayContains(filesToServe, theRequest.URL.Path[1:]) {
		fmt.Fprintf(theResponseWriter, "File served here...")
	} else {
		fmt.Fprintf(theResponseWriter, "Hello, %s!", theRequest.URL.Path[1:])
	}
}

// The main body of the program - parse user-provided command-line paramaters, or start the main web server process.
func main() {
	http.HandleFunc("/", webConsole)
	if len(os.Args) == 1 {
		// If no parameters are given, simply start the web server.
		fmt.Println("Starting web server...")
		http.ListenAndServe(":8090", nil)
	} else if os.Args[1] == "-list" {
		// Print a list of existing IDs.
		fmt.Println("List:")
		items, err := ioutil.ReadDir(".")
		if err != nil {
			log.Fatal(err)
		}
		for _, item := range items {
			fmt.Println(item.Name())
		}
	} else if os.Args[1] == "-generate" {
		// Generate a new ID, and create a matching folder.
		for {
			newID := generateID()
			if _, err := os.Stat(newID); os.IsNotExist(err) {
				os.Mkdir(newID, os.ModePerm)
				fmt.Println("New ID generated: " + newID)
				break
			}
		}
	}
}


/*
# If a public / private key pair doesn't exist, create one.
#@app.route("/getPublicCertificate")
#def getPublicCertificate():
#    return "certifcateGoesHere"

# Authenticate. Needs: encrypted string containing job ID, timestamp and nonce.
#@app.route("/auth")
#def auth:
#    return "Auth!"

#@app.route("/run")
#def run():
#    processRunning = False
#    for psLine in runCommand("ps ax").split("\n"):
#        if not psLine.find("build.sh") == -1:
#            processRunning = True

#    if flask.request.args.get("action") == "run":
#        correctPasswordHash = getFile("/var/local/buildPassword.txt")
#        passedPasswordHash = hashlib.sha256(flask.request.args.get("password").encode("utf-8")).hexdigest()
#        if passedPasswordHash == correctPasswordHash:
#            if not processRunning:
#                os.system("bash /usr/local/bin/build.sh &")
#            return "RUNNING"
#        return "WRONGPASSWORD"
#    elif flask.request.args.get("action") == "getStatus":
#        if processRunning:
#            return "RUNNING"
#        else:
#            return "NOTRUNNING"
#    elif flask.request.args.get("action") == "getLogs":
#        return re.sub(".\[\d*?m", "", getFile("/var/log/build.log"))
#    else:
#        return getFile("/var/www/api/build.html")
*/
