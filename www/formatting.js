								// If a string starts with "ERROR:", format the line in red.
								if (value.toLowerCase().startsWith("error: ")) {
									value = "<span style='color:red'>" + value.substr(7) + "</span>"
									if (displayAlerts == true) {
										$("#taskAlerts").html(value);
									}
								} else if (value.toLowerCase().startsWith("warning: ") || value.toLowerCase().startsWith("alert: ")) {
									value = "<span style='color:yellow'>" + value.substr(9) + "</span>"
									if (displayAlerts == true) {
										$("#taskAlerts").html(value);
									}
								} else if (value.toLowerCase().startsWith("status: ")) {
									value = "<span style='color:green'>" + value.substr(8) + "</span>"
									if (displayAlerts == true) {
										$("#taskAlerts").html(value);
									}
								} else if (value.toLowerCase().startsWith("result: ")) {
									value = "<span style='color:black'>" + value.substr(8) + "</span>"
									if (displayAlerts == true) {
										$("#taskResults").html($("#taskResults").html() + "<div>" + value + "</div>");
									}
								}
