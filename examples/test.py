import sys
import time

wordArray = ["The","Quick","Brown","Fox","Jumped","Over","The","Lazy","Dog"]

for pl in range(0, len(wordArray)):
print (wordArray[pl])
    print ("Progress: Progress " + str(int(round(pl / (len(wordArray)-1), 2) * 100)))
    sys.stdout.flush()
    time.sleep(2)
