function doAPICall(theMethod, theParams, theSuccessFunction, callMethod="POST") {
    theParams["taskID"] = taskID;
    theParams["token"] = token;
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
        apiCall.open("GET", "api/" + theMethod + "?" + URLEncodedParams, true);
        apiCall.send();
    } else {
        apiCall.open("POST", "api/" + theMethod, true);
        apiCall.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
        apiCall.send(URLEncodedParams);
    }
}