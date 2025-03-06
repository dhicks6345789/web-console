var webconsole = {
    // A utility function to do a webconsole API call.
    APICall: function(theMethod, theParams, theSuccessFunction, callMethod="POST", APIURLPrefix="") {
        if (!("taskID" in theParams)) {
            if (typeof taskID !== "undefined") {
                theParams["taskID"] = taskID;
            }
        }
        if (!("token" in theParams)) {
            if (typeof token !== "undefined") {
                theParams["token"] = token;
            }
        }
        var apiCall = new XMLHttpRequest();
        apiCall.onreadystatechange = function() {
            if (apiCall.readyState == 4 && apiCall.status == 200) {
                if (apiCall.responseText.toLowerCase().startsWith("error:")) {
                    // console.log(apiCall.responseText);
                    document.getElementById("taskErrors").innerHTML = apiCall.responseText;
                } else {
                    theSuccessFunction(apiCall.responseText);
                }
            }
        }
        URLEncodedParams = "";
        for (const [paramKey, paramValue] of Object.entries(theParams)) {
            URLEncodedParams = URLEncodedParams + encodeURIComponent(paramKey) + "=" + encodeURIComponent(paramValue) + "&";
        }
        URLEncodedParams = URLEncodedParams.slice(0, -1);
        if (callMethod == "GET") {
            apiCall.open("GET", APIURLPrefix + "api/" + theMethod + "?" + URLEncodedParams, true);
            apiCall.send();
        } else {
            apiCall.open("POST", APIURLPrefix + "api/" + theMethod, true);
            apiCall.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
            apiCall.send(URLEncodedParams);
        }
    },
    
    // Trigger a Task running server-side, then poll to check when that Task has finished.
    intervalID: 0,
    polledTasks: {},
    APITask: function(theTaskID, theSuccessFunction, pollPeriod=5, APIURLPrefix="") {
        webconsole.APICall("runTask", {"taskID":theTaskID}, function(result) {
            if (result == "OK") {
                webconsole.polledTasks[theTaskID] = {"taskID":theTaskID, "successFunction":theSuccessFunction, "APIURLPrefix":APIURLPrefix, "period":pollPeriod, "tick":0};
                webconsole.intervalID = setInterval(webconsole.pollTask, 1000);
            }
        }, "GET", APIURLPrefix);
    },
    
    pollTask: function() {
        for (pollTaskID in webconsole.polledTasks) {
            webconsole.polledTasks[pollTaskID]["tick"] = webconsole.polledTasks[pollTaskID]["tick"] + 1;
            if (webconsole.polledTasks[pollTaskID]["tick"] == webconsole.polledTasks[pollTaskID]["period"]) {
                webconsole.polledTasks[pollTaskID]["tick"] = 0;
                webconsole.APICall("getTaskRunning", {"taskID":webconsole.polledTasks[pollTaskID]["taskID"]}, function(result) {
                    if (result == "NO") {
                        getURL = webconsole.polledTasks[pollTaskID]["APIURLPrefix"] + pollTaskID + "/output.json";
                        successFunction = webconsole.polledTasks[pollTaskID]["successFunction"];
                        
                        delete webconsole.polledTasks[pollTaskID];
                        if (Object.keys(webconsole.polledTasks).length == 0) {
                            clearInterval(webconsole.intervalID);
                        }
                        
                        var getCall = new XMLHttpRequest();
                        getCall.onreadystatechange = function() {
                            if (getCall.readyState == 4 && getCall.status == 200) {
                                successFunction(getCall.responseText);
                            }
                        }
                        getCall.open("GET", getURL, true);
                        getCall.send();
                    }
                }, "GET", webconsole.polledTasks[pollTaskID]["APIURLPrefix"]);
            }
        }
    },
    
    // Given a DOM Node, renames any defined Node IDs to include a number on the end.
    // Useful for, after cloning a DOM Node, renaming IDs to be unique.
    numberElementIDs: function(theNode, theNumber) {
        if (theNode.id != undefined && theNode.id != "") {
            theNode.id = theNode.id + "-" + theNumber;
        }
        theNode.childNodes.forEach(function(childNode) {
            webconsole.numberElementIDs(childNode, theNumber);
        });
    },
    
    // Does a GET or POST, optionally in a new tab, with the given variables.
    // Expects to find a form with id "actionForm" on the main page:
    // <form id="actionForm" action="" method=""></form>
    doAction: function(theAction, theURL, theNewTab, theVariables, debug) {
        actionForm = document.getElementById("actionForm");
        if (theAction.toLowerCase() == "post") {
            actionForm.method = "POST";
        } else {
            actionForm.method = "GET";
        }
        actionForm.action = theURL;
        if (theNewTab == true) {
            actionForm.target = "_blank";
        }
        actionFormHTML = "";
        for (var varName in theVariables) {
            actionFormHTML = actionFormHTML + "<input name='" + varName + "' type='hidden' value='" + theVariables[varName] + "'>";
        }
        actionForm.innerHTML = actionFormHTML;
        if (debug == true) {
            console.log(actionForm);
        }
        actionForm.submit();
    }
};
