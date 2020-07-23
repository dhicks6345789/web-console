# Web Console
Provides a simple web interface for command-line applications - quickly publish your Python / Go / Bash / Powershell / etc script as a basic web app. Turns STDOUT into formatted text, alerts and progress indicators (interface written using Bootstrap / JQuery). Supports any target language, simply runs any command-line based script or executable.

Cross-platform (written in Go), runs as a self-contained executable complete with embedded web server on Windows, Linux and MacOS. The install process includes optional setup as a service / deamon on each platform (uses [NSSM](https://nssm.cc/) on Windows), plus the installer includes setup for [tunnelto.dev](https://tunnelto.dev/) to provide a secure reverse proxy through firewalls if needed.

Python (Flask) version also available to run on (for instance) [PythonAnywhere](https://www.pythonanywhere.com/).

Simple API, handles authentication (without using cookies), provides a mechanism for third-parties to handle authorisation.

![Screenshot of Web Console's main user interface](https://raw.githubusercontent.com/dhicks6345789/web-console/master/docs/example1.png)

## Installation

Download.
