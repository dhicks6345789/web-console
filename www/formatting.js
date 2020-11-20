								// If a string starts with "ERROR:", format the line in red.
								if (value.toLowerCase().startsWith("error: ")) {
									value = "<span style='color:red'>" + value + "</span>"
									if (displayAlerts == true) {
										$("#taskAlerts").html(value);
									}
								} else if (value.toLowerCase().startsWith("warning: ") || value.toLowerCase().startsWith("alert: ")) {
									value = "<span style='color:yellow'>" + value + "</span>"
									if (displayAlerts == true) {
										$("#taskAlerts").html(value);
									}
								} else if (value.toLowerCase().startsWith("status: ")) {
									value = "<span style='color:green'>" + value.substr(8); + "</span>"
									if (displayAlerts == true) {
										$("#taskAlerts").html(value);
									}
								}
