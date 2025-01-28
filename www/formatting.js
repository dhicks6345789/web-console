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
											// <div id="textInput0" style="display:none;">
											// <div id="textInputMessage0">Message:</div>
											// <div class="input-group mb-3">
											// <input id="textInputBox0" type="text" class="form-control" onkeydown="checkForEnter(this)">
											// <button id="textInputButton0" type="button" class="btn btn-outline-secondary" onclick="submitInput()">Go</button>
											
											// A plain text input box.
											textInputBlock = document.getElementById("textInput0");
											console.log(textInputBlock);
											newTextInputBlock = textInputBlock.cloneNode(true);
											console.log(newTextInputBlock);
											
											console.log(textInputBlock.childNodes());
											//textInputMessage = textInputBlock.getElementById("textInputMessage0");
											//textInputMessage.id = "textInputMessage1";
											//textInputMessage.innerHTML = value.substr(12);

											//textInputBox = textInputBlock.getElementById("textInputBox0");
											//textInputBox.id = "textInputBox1";

											//textInputButton = textInputBlock.getElementById("textInputButton0");
											//textInputButton.id = "textInputButton1";
											
											//document.getElementById("taskInput").appendChild(textInputBlock);
											//textInputBlock.style.display = "block";
										}
									}
								} else if (!value.toLowerCase().startsWith("progress: ")) {
									value = "<span style='color:LightGray'>" + value + "</span>"
								}
