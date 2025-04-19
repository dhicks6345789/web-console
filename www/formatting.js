								// If a string starts with "ERROR:", format the line in red.
								if (value.toLowerCase().startsWith("error: ")) {
									value = "<span style='color:Red'>" + value.substr(7) + "</span>"
									if (displayAlerts == true) {
										document.getElementById("taskAlerts").innerHTML = webconsole.renderMarkdown(value);
									}
								} else if (value.toLowerCase().startsWith("warning: ") || value.toLowerCase().startsWith("alert: ")) {
									value = "<span style='color:DarkGoldenRod'>" + value.substr(9) + "</span>"
									if (displayAlerts == true) {
										document.getElementById("taskAlerts").innerHTML = webconsole.renderMarkdown(value);
									}
								} else if (value.toLowerCase().startsWith("status: ")) {
									value = "<span style='color:Green'>" + value.substr(8) + "</span>"
									if (displayAlerts == true) {
										document.getElementById("taskAlerts").innerHTML = webconsole.renderMarkdown(value);
									}
								} else if (value.toLowerCase().startsWith("result: ")) {
									value = "<span style='color:Black'>" + value.substr(8) + "</span>"
									if (displayAlerts == true) {
										document.getElementById("taskResults").innerHTML = document.getElementById("taskResults").innerHTML + "<div>" + value + "</div>";
									}
								} else if (value.toLowerCase().startsWith("input:")) {
									if (displayAlerts == true) {
										// If a string begins with "INPUT:", we ask the user for some input.
										inputCount = inputCount + 1;

										// Clear any existing alert messages shown to the user.
										document.getElementById("taskAlerts").innerHTML = "";

										taskInputBlock = document.getElementById("taskInput");
										taskInputBlock.childNodes.forEach(function(theItem) {
											if (theItem.id.startsWith("textInput")) {
												textInputBox = theItem.childNodes[3].childNodes[1];
												textInputBox.setAttribute("onkeydown", textInputBox.getAttribute("onkeydown").replace("EnterSubmit", "EnterNext"));
												theItem.childNodes[3].childNodes[3].remove();
											}
										});
										
										if (value.toLowerCase().startsWith("input:text:")) {
											// A plain text input box.
											textInputBlock = document.getElementById("textInput").cloneNode(true);
											textInputBlock.id = "textInput-" + inputCount;
											
											textInputMessage = textInputBlock.childNodes[1];
											textInputMessage.id = "textInputMessage-" + inputCount;
											textInputMessage.innerHTML = value.substr(12);

											textInputBox = textInputBlock.childNodes[3].childNodes[1];
											textInputBox.id = "textInputBox-" + inputCount;
											textInputBox.setAttribute("onkeydown", "checkForEnterSubmit(" + inputCount + ")");
											textInputBox.setAttribute("tabindex", inputCount);

											textInputButton = textInputBlock.childNodes[3].childNodes[3];
											textInputButton.id = "textInputButton-" + inputCount;
											textInputButton.setAttribute("onclick", "submitInput()");
											textInputButton.setAttribute("tabindex", inputCount+1);
											
											taskInputBlock.appendChild(textInputBlock);
											if (inputCount == 1) {
												textInputBox.focus();
											}
										} else if (value.toLowerCase().startsWith("input:multichoice:")) {
											multichoiceInputBlock = document.getElementById("multichoiceInput").cloneNode(true);
											multichoiceInputBlock.id = "multichoiceInput-" + inputCount;

											options = value.split(":");
											
											multichoiceInputMessage = multichoiceInputBlock.childNodes[1];
											multichoiceInputMessage.id = "multichoiceInputMessage-" + inputCount;
											multichoiceInputMessage.innerHTML = options[options.length - 1];
											
											multichoiceSelectElement = multichoiceInputBlock.childNodes[3].childNodes[1];
											multichoiceSelectElement.id = "multichoiceSelect-" + inputCount;
											for (pl = 2; pl < options.length - 1; pl = pl + 1) {
												multichoiceSelectElement.innerHTML = multichoiceSelectElement.innerHTML + "<option value=\"" + options[pl] + "\">" + options[pl] + "</option>";
											}
											multichoiceSelectElement.setAttribute("tabindex", inputCount);

											multichoiceInputButton = multichoiceInputBlock.childNodes[3].childNodes[3];
											multichoiceInputButton.id = "multichoiceInputButton-" + inputCount;
											multichoiceInputButton.setAttribute("onclick", "submitInput()");
											multichoiceInputButton.setAttribute("tabindex", inputCount+1);
											
											taskInputBlock.appendChild(multichoiceInputBlock);
											if (inputCount == 1) {
												multichoiceSelectElement.focus();
											}
										}
									}
								} else if (!value.toLowerCase().startsWith("progress: ")) {
									value = "<span style='color:LightGray'>" + value + "</span>"
								}
