# A basic Python 3 test script that simply prints a sentence a word at a time, with a pause between each word.
# Simulates a script that takes a few seconds to complete. Outputs progress information for the user which is displayed as a progress bar by Web Console.

import sys
import time

sentence = ["The","Quick","Brown","Fox","Jumped","Over","The","Lazy","Dog"]
wordArray = sentence.split(" ")

for pl in range(0, len(wordArray)):
	print (wordArray[pl])
	print ("Progress: Progress " + str(int(round(pl / (len(wordArray)-1), 2) * 100)) + "%")
	sys.stdout.flush()
	time.sleep(2)
