package main

import (
	"fmt"
	"os"
	"log"
	"io/ioutil"
	"net/http"
)

func webConsole(theResponseWriter http.ResponseWriter, theRequest *http.Request) {
	fmt.Fprintf(theResponseWriter, "Hello, %s!", theRequest.URL.Path[1:])
}

func main() {
	http.HandleFunc("/", webConsole)
	if len(os.Args) == 1 {
		http.ListenAndServe(":8090", nil)
	} else if os.Args[1] == "-list" {
		fmt.Println("List:")
		items, err := ioutil.ReadDir(".")
		if err != nil {
			log.Fatal(err)
		}
		for _, item := range items {
			fmt.Println(item.Name())
		}
	} else if os.Args[1] == "-generate" {
		String newID := ""
		for pl := 0; pl < 16; pl++ {
			newID := newID + "a"
		}
		fmt.Println(newID)
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
