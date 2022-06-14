# Web Console
Provides a simple web interface for command-line applications - quickly publish a Python / Bash / Powershell / Batch / etc script as a basic web app. Turns STDOUT into formatted text, alerts and progress indicators.

Cross-platform, runs as a self-contained web server, binaries are available for Windows, Linux (including Raspberry Pi) and MacOS. The install process includes setup as a service on each platform and, optionally, for the cross-platform [tunnelto.dev](https://tunnelto.dev/) service to provide an HTTPS-secured connection through a firewall.

Web Console is intended to be used behind a proxy server or tunneling service, it doesn't by itself handle HTTPS connections, making it unsuitible as a general-purpose web server open to the outside world. Also, Web Console's implementation prioritises ease of configuration and use over performance, meaning that for larger setups a proxy server is probably going to be a good idea. However, it can be used to provide a basic user interface for a stand-alone, non-networked system, such as a kiosk machine, using HTTP only on a local web browser.

Web Console has been used behind [Apache](https://httpd.apache.org/) and [NGINX](https://www.nginx.com/) acting as reverse-proxy servers, although we recommend [Caddy](https://caddyserver.com/), simply for its ease-of-use in automatically dealing with HTTPS provision for you.

As well as providing a user interface, Web Console also provides a simple REST API, providing a URLs for for services such as [IFTTT](https://ifttt.com/) and [Zapier](https://zapier.com/) or letting you trigger tasks from remote systems with command-line tools like [curl](https://curl.se/).

## Live Demo

You can see see a [live demo](https://www.sansay.co.uk/webconsole/view?taskID=4jaknvvu0b4zl3ee) right now.

The above link runs a simple [demo application](https://github.com/dhicks6345789/web-console/blob/master/examples/test.py) that produces some example output, showing the different types of output message supported. It also prints a progress percentage, which is displayed by Web Console as a progress bar.

## Installation

The intention of Web Console is to give beginner programmers a simple interface for writing small applications and getting those in front of end-users as simply and quickly as possible, with no knowledge of Windows or Linux system administration needed. However, there is some setup and configuration needed to get an instance of Web Console up and running, so the rest of this document is probably best suited for people with some system administration experience. Web Console is, hopefully, quite simple to install and configure, as far as this kind of application goes, and hopefully the following instructions are easy to follow.

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

For more details, including details of building from source, see the [installation](docs/INSTALLATION.md) page.

## Usage

Web Console runs "tasks", individual scripts / applications that each have their own URL where execution can be triggered and output can be seen. Pretty much any existing application runable from the command line should work, returning any console output sent to STDOUT or STDERR to the web-based user interface.

If you are writing a new script or command line utility (or reformatting the output from an existing utility) you can produce output specifically for Web Console to interpret and display in certain ways. For more details, see ...

The Webconsole binary is also used as a command-line tool for configuration. At the command line, type `webconsole --help` for a [command summary](https://www.sansay.co.uk/docs/web-console/COMMANDS.txt).

### Startup

If you install Web Console using the install script then it should run as a service, starting when your machine boots. If you just run the binary, without any other arguments, it will start and listen (by default) on port 8090 for HTTP connections. If you start with the "debug" argument, i.e.:

```
webconsole --debug
```

Then Web Console will start in debug mode, printing information to the console about each received HTTP request.

### Main Configuration

On startup, Web Console attempts to locate a configuration file. This can be in CSV or Excel format, named "config.csv" or "config.xlsx". It will first look in `/etc/webconsole` on MacOS / Linux, or in `C:\Program Files\webconsole` on Windows, followed by the current folder.

The configuration file has a simple `Parameter, Value` format. Valid parameters are:
* pathPrefix: For use when Web Console is used behind a proxy server and isn't at the root of the domain name, tells the web server what path prefix it has been given. E.g.: `/webconsole`.

### Task Configuration

Webconsole will look in the defined "tasks" folder (by default, on Linux / MacOS, `/etc/webconsole/tasks`, on Windows, `C:\Program Files\webconsole\tasks`) for subfolders. Any subfolders found will be searched for a "config.txt" file and used as a Task ID if found. Task IDs generated by the Webconsole application are random 16-character strings, but any string (no spaces) can be used.

The format of config.txt is a basic YAML-style text - keywords followed by a colon then the given value, e.g.:

```
title: Test Site
public: N
command: /root/buildSite.sh
```
Valid keywords are:
* title: The title of the Task, displayed in the header on the Task page and as the page title.
* description: Descriptive text saying what the task does.
* secret: A secret phrase / key / password. If present, must be given during the authentication process - can be passed in via GET (not very secure) or POST.
* public: If "Y", this Task will be listed on the index page. Obviously, only use for Tasks you want to be made public.
* ratelimit: If more than 0, then this Task will not be allowed to run more often than the given number of seconds.
* progress: If "Y", then a progress bar will be presented on the page. The percentages the progress bar shows will be guessed from previous runtimes of this Task.
* command: The command line to run. Pretty much any valid command line (or shell / batch script) should work.

Note that changes to config.txt for any Task will be in effect the next time the Task is triggered, without any need to restart / reload anything server side or even refresh the web interface if you already have the Task's page open.

More details can be found on the [documentation page](docs/TASKS.md).

## Dependancies

This project contains binaries from:

[The Non-Sucking Service Manager](https://nssm.cc/) by Iain Patterson, used to set up services on Windows. Public Domain license.

[tunnelto.dev](https://tunnelto.dev), Copyright (c) 2020 Alex Grinman, used to provide secure connections through firewalls. MIT license.

Thw web user interface is constructed using [Bootstrap 5](https://getbootstrap.com/docs/5.0/getting-started/introduction/) with [Popper](https://popper.js.org/). All required library files are included in the project and release distributions so Web Console can run as a self-contained application on a non-networked workstation / kiosk device if needed.

## To Do

### Bugs

* Output - does output.html get displayed properly?
* Add Mac support in install.sh.
* Authorisation config for "www" folder.
* On Windows, run batch files without having to explicitly run via cmd /c.
* Return error message if task file doesn't run, don't just sit.
* Live messages view not always showing every line, only gets all lines on page refresh.
* Is stderr being captured okay?

### Features

* Additions to the API to provide a mechanism for third-parties to handle authorisation.
  * Mystart.Online
  * Cloudflare
* Chroot (or Windows equivilent) jail per task.
* Inputs from STDIN.
* Better admin console.
  * New account creation
  * Task run schedualer, with error reporting if tasks fail.
  * Add "New Task" dialog, with (configurable) pre-defined "Type" field for quick starts:
    * Hugo
    * Jekyll
    * 11ty
    * Gatsby
    * Docs-To-Markdown
    * How-To
    * FAQ
    * Dashboard
    * Slideshow
    * Yearbook
  * Ability to connect cloud storage.
* Code editor integrated in Task editor view.
* SSH console integrated in Task editor view.
* Custom design system
  * Bootstrap
  * Gov.uk
  * Stack Overflow
  * Tailwind
  * Scott's?
* Python (Flask) implementation to run on (for instance) [PythonAnywhere](https://www.pythonanywhere.com/).
* Optional ability to stop Task(?).
* Auto-generated favicon icon(s)?
