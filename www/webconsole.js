var webconsole = {
    // A utility function to do a webconsole API call.
    APICall: function(theMethod, theParams, theSuccessFunction, callMethod="POST", APIURLPrefix="") {
        console.log(APIURLPrefix);
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
                theSuccessFunction(apiCall.responseText);
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
    APITask: function(theTaskID, pollPeriod=5, APIURLPrefix="") {
        console.log(APIURLPrefix);
        webconsole.APICall("runTask", {"taskID":theTaskID}, function(result) {
            console.log(result);
            // TFLDataFetchTimeout = setTimeout(completeTFLDataFetch, 10000);
        }, APIURLPrefix=APIURLPrefix);
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
    }
};
