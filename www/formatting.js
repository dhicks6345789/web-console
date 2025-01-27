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
											// A plain text input box.
											document.getElementById("taskResults").innerHTML = document.getElementById("taskResults").innerHTML + "<div>" + value.substr(12) + "</div>";
											document.getElementById("taskInput").innerHTML = "<input type='text' class='form-control' id='input1'>";
										}
									}
								} else if (!value.toLowerCase().startsWith("progress: ")) {
									value = "<span style='color:LightGray'>" + value + "</span>"
								}
