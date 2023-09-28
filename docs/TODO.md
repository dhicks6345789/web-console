# To Do

## Bugs

* Live messages view not always showing every line, only gets all lines on page refresh.
* Is STDERR being captured okay? Should intersperse with STDOUT, not be stuck at end.
* Seems to be an upper limit on the number of lines storable by the "output" text box on the user interface - might need to set a buffer limit.
* API - Check for ".." paths.
* Output - does output.html get displayed properly?
  * Make sure MIME type is set properly.
  * Default to any file found in www folder to be served as root if index.html not available.
  * Output iframe - seems to be cached version of output sometimes, needs explicit updating somehow.
* 404 message not setting 404 code in header.
* Logging - add date / time signature to start of log output.
* Add / check "error code on exit" message for non-0 results.
* Return error message if task file doesn't run, don't just sit.

## Features

* Edit Mode
  * Highlight current/unsaved files, stop exit / run until files saved / discarded.
  * Upload of files / folders.
  * Rename of files / folder.
  * Add diff / versioning for file edits.
* Add functions to webconsole.js for library functions to do calls to Tasks as API server:
  * Get result (task/www/index.json, proper MIME type set).
  * Done - Run success function with returned value.
    * Make success function a promise instead?
* Make separate API call ID from Task ID so API call can be kept secret without needing external authentication.
  * Viewable to runers / editors only.
* Chroot (or Windows equivilent) jail per task.
  * Actually simple to add via prepend string in arguments?
* Backup tab for Task editor view - download and restore.
  * Task mirroring to secondary server.
* Inputs from STDIN.
  * Single line text box
  * Radio select
  * Dropdown
  * Typeahead dropdown
  * File upload
  * Photo capture
* Better admin console.
  * UI created by own capabilities!
  * Task run schedualer (cross-platform, write back to both cron and Windows Task Schedular), with error reporting if tasks fail.
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
  * Ability to rename / delete Tasks.
  * Ability to connect cloud storage.
* GitHub Actions template to download for individual Tasks
  * Should just be able to use a single template file and insert Task ID with Curl command.
* Optional ability to stop Task.
  * Users with Editor rights can cancel. Runner who runs Task gets permissions to cancel, not other Runners.
* Add Mac support in install.sh.
* Add ChromeOS support in install.sh.
* Web Server / Results folder
  * Add session cookies for authentication.
