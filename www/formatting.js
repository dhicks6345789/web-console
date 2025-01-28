								// If a string starts with "ERROR:", format the line in red.
								if (value.toLowerCase().startsWith("error: ")) {
									value = "<span style='color:Red'>" + value.substr(7) + "</span>"
									if (displayAlerts == true) {
										document.getElementById("taskAlerts").innerHTML = value;
									}
								} else if (value.toLowerCase().startsWith("warning: ") || value.toLowerCase().startsWith("alert: ")) {
									value = "<span style='color:DarkGoldenRod'>" + value.substr(9) + "</span>"
									if (displayAlerts == true) {
										document.getElementById("taskAlerts").innerHTML = value;
									}
								} else if (value.toLowerCase().startsWith("status: ")) {
									value = "<span style='color:Green'>" + value.substr(8) + "</span>"
									if (displayAlerts == true) {
										document.getElementById("taskAlerts").innerHTML = value;
									}
								} else if (value.toLowerCase().startsWith("result: ")) {
									value = "<span style='color:Black'>" + value.substr(8) + "</span>"
									if (displayAlerts == true) {
										document.getElementById("taskResults").innerHTML = document.getElementById("taskResults").innerHTML + "<div>" + value + "</div>";
									}
								} else if (value.toLowerCase().startsWith("input:")) {
									if (displayAlerts == true) {
										// If a string begins with "INPUT:", we ask the user for some input.
										if (value.toLowerCase().startsWith("input:text:")) {
											console.log("Input - a plain text box.");
											
											// A plain text input box.
											textInputBlock = document.getElementById("textInput").cloneNode(true);
											
											textInputMessage = textInputBlock.childNodes[1];
											textInputMessage.id = "textInputMessage1";
											textInputMessage.innerHTML = value.substr(12);

											textInputBox = textInputBlock.childNodes[3].childNodes[1];
											textInputBox.id = "textInputBox1";

											textInputButton = textInputBlock.childNodes[3].childNodes[3];
											textInputButton.id = "textInputButton1";
											
											document.getElementById("taskInput").appendChild(textInputBlock);
										}
									}
								} else if (!value.toLowerCase().startsWith("progress: ")) {
									value = "<span style='color:LightGray'>" + value + "</span>"
								}
