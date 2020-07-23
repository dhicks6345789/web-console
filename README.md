# Web Console
Provides a simple web interface for command-line applications - quickly publish your Python / Go / Bash / Powershell / etc script as a basic web app. Turns STDOUT into formatted text, alerts and progress indicators (interface written using Bootstrap / JQuery). Supports any target language, simply runs any command-line based script or executable.

Cross-platform (written in Go), runs as a self-contained executable complete with embedded web server on Windows, Linux and MacOS. The install process includes optional setup as a service / deamon on each platform (uses [NSSM](https://nssm.cc/) on Windows), plus the installer includes setup for [tunnelto.dev](https://tunnelto.dev/) to provide a secure connection through a firewall and a handy subdomain to point a browser at if needed.

Python (Flask) version also available to run on (for instance) [PythonAnywhere](https://www.pythonanywhere.com/).

Simple API, handles authentication (without using cookies), provides a mechanism for third-parties to handle authorisation. Can be used to provide webhook URIs for your scripts for services such as [IFTTT](https://ifttt.com/) and [Zapier](https://zapier.com/).

![Screenshot of Web Console's main user interface](https://raw.githubusercontent.com/dhicks6345789/web-console/master/docs/example1.png)

## Live Demo

Link to live demo goes here.

The above link runs the following simple Python 3 application:

```
import sys
import time

wordArray = ["The","Quick","Brown","Fox","Jumped","Over","The","Lazy","Dog"]

for pl in range(0, len(wordArray)):
    print (wordArray[pl])
    print ("Progress: Progress " + str(int(round(pl / (len(wordArray)-1), 2) * 100)))
    sys.stdout.flush()
    time.sleep(2)
```

The demo application simply prints a sentance a word at a time, one word every two seconds. It also prints a progress percentage, which is displayed by Web Console as a progress bar.

## Installation

Further instructions to go here.

## Dependancies

This project contains binaries from:

[The Non-Sucking Service Manager](https://nssm.cc/) by Iain Patterson, used to set up services on Windows. Public Domain license.

[tunnelto.dev](https://tunnelto.dev), Copyright (c) 2020 Alex Grinman, used to provide secure connections through firewalls. MIT license.
