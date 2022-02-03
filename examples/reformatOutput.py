#!/usr/bin/python3

import sys

def replaceAll(theString, theMatch, theReplace):
	newString = theString.replace(theMatch, theReplace)
	while not newString == theString:
		theString = newString
		newString = theString.replace(theMatch, theReplace)
	return newString

for line in sys.stdin:
	line = line.strip()
	line = replaceAll(line, "  ", " ")
	line = replaceAll(line, " |", ":")
	if line.startswith("WARN "):
		print(line.replace("WARN ","WARNING: "))
	elif line.startswith("* branch"):
		print(line.replace("* branch","Branch:"))
	else:
		if not (line.startswith("| EN") or line.startswith("---") or line.startswith("Start building sites") or line == ""):
			print(line)
