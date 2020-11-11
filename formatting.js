								// If a string begins with "Progress: ", interpret the following number as a percentage completion value, and
								// update the progress bar accordingly.
								if (value.toLowerCase().startsWith("progress:")) {
									progressBarName = value.substring(value.indexOf(":")+1, value.lastIndexOf(" ")).trim();
									progressBarValue = value.substring(value.lastIndexOf(" ")+1, value.length).replace("%","").trim();
									$("#taskProgress").html(progressBarName + ": " + progressBarValue + "% <div class='progress' style='width:80%'><div class='progress-bar' role='progressbar' style='width:" + progressBarValue + "%' aria-valuenow='" + progressBarValue + "' aria-valuemin='0' aria-valuemax='100'></div></div>");
								// If a string starts with "ERROR:", format the line in red and set the running state to "error".
								} else if (value.toLowerCase().startsWith("error:")) {
									$("#taskAlerts").html("<div style='color:red'>" + value + "</div>");
								// Otherwise, display the line as a message for the user.
								} else {
									$("#taskAlerts").html(value);
									$("#taskOutput").html($("#taskOutput").html() + value + "\n");
								}
								outputLine = outputLine + 1;
