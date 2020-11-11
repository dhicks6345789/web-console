								// If a string starts with "ERROR:", format the line in red.
								if (value.toLowerCase().startsWith("error:")) {
									value = "<div style='color:red'>" + value + "</div>"
									$("#taskAlerts").html(value);
								}
