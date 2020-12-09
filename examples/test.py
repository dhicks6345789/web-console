# A basic Python 3 test script that produces some example out, with a pause between lines to simulate work being done.
# By default, outputs progress information for the user which is displayed as a progress bar by Web Console, but that
# can be turned off if you want Web Console to provide the progress bar instead.

import sys
import time

keywordsArray = ["STATUS","WARNING","ERROR","RESULT","PROGRESS"]

outputLength = 0
outputArray = ["STATUS: Starting...", "RESULT: To the batmobile, let's go!", "Running...", "More running...", "Yet more running...", "ERROR: No Batkeys!", "STATUS: Keys found in Batpocket!", "Fumbling with batkeys...", "RESULT: Atomic batteries to power...", "Efficiency: 73%.", "RESULT: Turbines to speed...", "RESULT: Roger, ready to move out.", "STATUS: Done."]
for pl in range(0, len(outputArray)):
	if outputArray[pl].split(":")[0] in keywordsArray:
		outputLength = outputLength + 1

displayProgress = True
if len(sys.argv) > 1:
	if sys.argv[1] == "--NOPROGRESS":
		displayProgress = False

for pl in range(0, len(outputArray)):
	print (outputArray[pl])
	if displayProgress:
		print("PROGRESS: Progress " + str(int(round(pl / outputLength), 2) * 100) + "%")
	sys.stdout.flush()
	if outputArray[pl].split(":")[0] in keywordsArray:
		time.sleep(2)
