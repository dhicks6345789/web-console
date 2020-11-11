								// If a string starts with "ERROR:", format the line in red.
								if (value.toLowerCase().startsWith("error")) {
									value = "<div style='color:red'>" + value + "</div>"
									$("#taskAlerts").html(value);
								} else if (value.toLowerCase().startsWith("warning") || value.toLowerCase().startsWith("alert")) {
									value = "<div style='color:yellow'>" + value + "</div>"
									$("#taskAlerts").html(value);
								} else if (value.toLowerCase().startsWith("status")) {
									value = "<div style='color:green'>" + value + "</div>"
									$("#taskAlerts").html(value);
								}
