# Tasks

Web Console lets you add formatted text to the output of your command-line script for display to the user, and can also display simple progress bars to the user. You add formatted text simply by adding keywords, followed by a colon and a space, to the start of a line output by you script:
* ALERT: Displays text, highlighted in green, to the user in the main output display area.
* WARNING: Displays text, highlighted in yellow, to the user in the main output display area.
* ERROR: Displays text, highlighted in red, to the user in the main output display area.

## Custom Output Formatting

Webconsole adds the contents of "formatting.js" to the main HTML user interface to handle text formatting. If you want to customise the way text is formatted you can use your own version. Simpy copy the formatting.js file from the web root folder (/etc/webconsole/www by default on Linux) to the tasks folder (/etc/webconsole/tasks), or to an individual task's folder if you want to customise formatting for one particular task, then make changes to that file as you wish.

The default contents of formatting.js are fairly simple, just formatting text in different colours if a keyword is found at the start of a line.

## Custom Favicon

If you create a new Task via the command-line tool you will be given the option to randomly assign a favicon, selected from the "favicons" folder. You can use your own favicon if preffered, just copy the appropriate icon (just name it "favicon.png") to an individual Task's folder, or the root of the "tasks" folder to set the same favicon for all Tasks. The web server automatically takes care of providing different versions of the favicon as needed, complete with custom browserconfig.xml, site.webmanifest and Apple mask-icon vector SVG traced from the original file - see the header block in the [HTML of the Task user interface](https://github.com/dhicks6345789/web-console/blob/master/www/webconsole.html) for details.

A set of favicons are provided from the free "fruit" [collection](https://www.iconfinder.com/iconsets/fruits-52) from Thiago Silva.

## Custom Description

If you need a longer description than a single line of text, then you can place you custom description in a file called description.txt in the root of an individual Task. You can embed HTML in this file if you wish, complete with links or whatever other components you like.
