# To Do

## Bugs

* Live messages view not always showing every line, only gets all lines on page refresh.
* Is STDERR being captured okay? Should intersperse with STDOUT, not be stuck at end.
* Output - does output.html get displayed properly?
  * Make sure MIME type is set properly.
  * Default to any file found in www folder to be served as root if index.html not available.
* 404 message not setting 404 code in header.
* Yellow (on white) isn't a good colour for warning messages. Purple, maybe?
* Logging - add date / time signature to start of log output.
* Add / check "error code on exit" message for non-0 results.
* Add Mac support in install.sh.
* Add ChromeOS support in install.sh.
* On Windows, run batch files without having to explicitly run via cmd /c.
* Return error message if task file doesn't run, don't just sit.

## Features

* Add functions to webconsole.js for library functions to do calls to Tasks as API server:
  * Done - Trigger run
  * Done - Poll running task to see when finished - might be long running.
  * Get result (task/www/index.json, proper MIME type set).
  * Done - Run success function with returned value.
    * Make success function a promise instead?
* Additions to the API to provide a mechanism for third-parties to handle authorisation.
  * Mystart.Online
  * Cloudflare
* Chroot (or Windows equivilent) jail per task.
* Inputs from STDIN.
  * Single line text box
  * Radio select
  * Dropdown
  * Typeahead dropdown
  * File upload
  * Photo capture
* Better admin console.
  * UI created by own capabilities!
  * Add root Task that uses own user interface to interact with user.
  * Task run schedualer, with error reporting if tasks fail.
  * Add "New Task" dialog, with (configurable) pre-defined "Type" field for quick starts:
    * Git checkout
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
* Backup tab for Task editor view - download and restore.
  * Task mirroring to secondary server.
* GitHub Actions template to download for individual Tasks
  * Should just be able to use a single template file and insert Task ID with Curl command.
* Custom design system
  * Bootstrap
  * Gov.uk
  * Stack Overflow
  * Tailwind
  * Scott's?
* Python (Flask) implementation to run on (for instance) [PythonAnywhere](https://www.pythonanywhere.com/).
* Optional ability to stop Task(?).
* Auto-generated favicon icon(s)?