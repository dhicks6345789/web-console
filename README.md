# Web Console
Provides a simple web interface for command-line applications - quickly publish your Python / Bash / Powershell / Batch / etc script as a basic web app. Turns STDOUT into formatted text, alerts and progress indicators.

Cross-platform, runs as a self-contained web server, binaries are available for Windows, Linux (including Raspberry Pi) and MacOS. The install process includes optional setup as a service / deamon on each platform and for the cross-platform [tunnelto.dev](https://tunnelto.dev/) service to provide an HTTPS-secured connection through a firewall. Web Console can be used behind a proxy server or from a local web browser as a basic user interface for a stand-alone, non-networked system.

As well as providing a user interface, Web Console also provides a simple REST API, providing a webhook URLs for for services such as [IFTTT](https://ifttt.com/) and [Zapier](https://zapier.com/) or letting you trigger tasks from remote systems with command-line tools like [curl](https://curl.se/).

## Live Demo

You can see see a [live demo](https://www.sansay.co.uk/webconsole/view?taskID=4jaknvvu0b4zl3ee) right now.

The above link runs a simple [demo application](https://github.com/dhicks6345789/web-console/blob/master/examples/test.py) that produces some example output, showing the different types of output message supported. It also prints a progress percentage, which is displayed by Web Console as a progress bar.

## Installation

### Mac / Linux (including Raspberry Pi)

On MacOS or Linux, you can download and run an install script (installs the latest release) with one command:
```
curl -s https://www.sansay.co.uk/web-console/install.sh | sudo bash
```

### Windows

On Windows, you can download and run an install batch file (installs the latest release) with one command:
```
powershell -command "& {&'Invoke-WebRequest' -Uri https://www.sansay.co.uk/web-console/install.bat -OutFile install.bat}" && install.bat && erase install.bat
```

### From Source

The source code is available on [Github](https://github.com/dhicks6345789/web-console). Written in Go, the source should be compileable on most platforms with a Go development environment available - the platform's default Go installation is generally fine.

Webconsole depends on the following libraries:
- [Resize](github.com/nfnt/resize): Simple bitmap image resizing library. Used in the implementation of favicons.
- [Gotrace](github.com/dennwc/gotrace): A Go implentation of [Potrace](http://potrace.sourceforge.net/), for tracing bitmaps to SVG files. Used in the implementation of favicons.
- [Golang-Image-ICO](github.com/kodeworks/golang-image-ico): An .ICO format image encoder. Used in the implementation of favicons.
- [Bcrypt](golang.org/x/crypto/bcrypt): For password hashing. Used for basic authentication.
- [Excelize](github.com/360EntSecGroup-Skylar/excelize): For loading Excel files.

A simple bash [build script](https://github.com/dhicks6345789/web-console/blob/master/build.sh) is available in the root of the source tree (or a [batch file](https://github.com/dhicks6345789/web-console/blob/master/build.bat) if you're building on Windows).

### Releases

You can download specific releases from the Github [releases page](https://github.com/dhicks6345789/web-console/releases).

If you don't want to build the source yourself but you want the very latest version (built nightly from the Github source, might have bugs), you can download nightly builds:

| Platform         | Binary
| ---------------- | ----------------------------------------------------------------------- |
| Windows 32-bit   | [Download](https://www.sansay.co.uk/web-console/binaries/win-386.exe)   |
| Windows 64-bit   | [Download](https://www.sansay.co.uk/web-console/binaries/win-amd64.exe) |
| WWW Folder       | [Download](https://www.sansay.co.uk/web-console/www.zip)                |
| ---------------- | ----------------------------------------------------------------------- |
| Mac              | [Download](https://www.sansay.co.uk/web-console/binaries/darwin-amd64)  |
| Linux 32-bit     | [Download](https://www.sansay.co.uk/web-console/binaries/linux-386)     |
| Linux 64-bit     | [Download](https://www.sansay.co.uk/web-console/binaries/linux-amd64)   |
| Linux ARM 32-bit | [Download](https://www.sansay.co.uk/web-console/binaries/linux-arm32)   |
| Linux ARM 64-bit | [Download](https://www.sansay.co.uk/web-console/binaries/linux-arm64)   |
| WWW Folder       | [Download](https://www.sansay.co.uk/web-console/www.tar.gz)             |

The following command on MacOS and Linux should download the appropriate binary for your platform and install it, along with the supporting "www" folder contents:
```
curl -s https://www.sansay.co.uk/web-console/installDev.sh | sudo bash
```

Or, On Windows:
```
powershell -command "& {&'Invoke-WebRequest' -Uri https://www.sansay.co.uk/web-console/installDev.bat -OutFile install.bat}" && install.bat && erase install.bat
```

## Usage

```
webconsole --new
```

Web Console should run pretty much any existing application runable from the command line, returning any console output sent to STDOUT or STDERR to the web user interface. You can use it to run GUI applications that produce no console output, although if they don't exit then the running Task will never end.

Web Console was created with the intention of making it very easy to add a basic web-accesible user interface to command-line applications - the kind of thing a single developer or system administrator might need to quickly write for a specific use case and get in front of end users as quickly as possible. In particular, it's assumed that user inputs and outputs will be provided via some other mechanism, such as files / folders stored on a cloud storage system.

If you are writing a new script or command line utility (or reformatting the output from an existing utility) you can produce output specifically for Web Console to interpret and display in certain ways. Simply including the keywords "ERROR", "WARNING" or "RESULT" at the start of an output line will place those output lines in appropriate places on the output console, highlighted in different colours.

## Dependancies

This project contains binaries from:

[The Non-Sucking Service Manager](https://nssm.cc/) by Iain Patterson, used to set up services on Windows. Public Domain license.

[tunnelto.dev](https://tunnelto.dev), Copyright (c) 2020 Alex Grinman, used to provide secure connections through firewalls. MIT license.

Thw web user interface is constructed using [Bootstrap 5](https://getbootstrap.com/docs/5.0/getting-started/introduction/) and the [JQuery](https://jquery.com/) and [Popper](https://popper.js.org/) JavaScript libraries. All required library files are included in the project and release distributions so Web Console can run as a self-contained application on a non-networked workstation if needed.

## Customisation

### Task Configuration Files

Webconsole will look in the defined "tasks" folder (by default, on Linux, /etc/webconsole) for subfolders. Any subfolders found will be searched for a "config.txt" file and used as a Task ID if found. Task IDs generated by the Webconsole application are random 16-character strings, but any string (no spaces) can be used.

The format of config.txt is as keywords followed by a colon then the given value, i.e.

```
title: Test Site
public: N
command: /root/buildSite.sh
```
Valid keywords are:
title: The title of the Task, displayed in the header on the Task page and as the page title.
description: Descriptive text saying what the task does.
secret: A secret phrase / key / password. If present, must be given during the authentication process - can be passed in via GET (not very secure) or POST.
public: If "Y", this Task will be listed on the index page. Obviously, only use for Tasks you want to be made public.
ratelimit: If more than 0, then this Task will not be allowed to run more often than the given number of seconds.
progress: If "Y", then a progress bar will be presented on the page. The percentages the progress bar shows will be guessed from previous runtimes of this Task.
command: The command line to run. Pretty much any valid command line (or shell / batch script) should work.

Note that changes to config.txt for any Task will be in effect the next time the Task is triggered, without any need to restart / reload anything server side or even refresh the web interface if you already have the Task's page open.

### Custom Output Formatting

Webconsole adds the contents of "formatting.js" to the main HTML user interface to handle text formatting. If you want to customise the way text is formatted you can use your own version. Simpy copy the formatting.js file from the web root folder (/etc/webconsole/www by default on Linux) to the tasks folder (/etc/webconsole/tasks), or to an individual task's folder if you want to customise formatting for one particular task, then make changes to that file as you wish.

The default contents of formatting.js are fairly simple, just formatting text in different colours if a keyword is found at the start of a line.

### Custom Favicon

If you create a new Task via the command-line tool you will be given the option to randomly assign a favicon, selected from the "favicons" folder. You can use your own faviocn if preffered, just copy the appropriate icon to an individual Task's folder, or the root of the "tasks" folder to set the same favicon for all Tasks.

A set of favicons are provided from the free "fruit" [collection](https://www.iconfinder.com/iconsets/fruits-52) from Thiago Silva.

### Custom Description

If you need a longer description than a single line of text, then you can place you custom description in a file called description.txt in the root of an individual Task. You can
embed HTML in this file if you wish, complete with links or whatever other components you like.

## To Do

### Bugs

* Binary download section a bit pointless as change likely to happen in support files - better point people at build.
* Add Pi / Mac support in install.sh.
* Explain new (v0.1.1) "www" hosting feature.
* Authorisation config for "www" folder.
* On Windows, run batch files without having to explicitly run via cmd /c.
* Return error message if batch file doesn't run, don't just sit.
* Live messages view not always showing every line, only gets all lines on page refresh.
* Upgrade Bootstrap.

### Features

* Additions to the API to provide a mechanism for third-parties to handle authorisation.
* Better admin console.
* Python (Flask) implementation to run on (for instance) [PythonAnywhere](https://www.pythonanywhere.com/).
* Inputs from STDIN.
* Optional ability to stop Task(?).
* Auto-generated favicon icon(s)?
