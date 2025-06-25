# To Do

## Bugs

* Live messages view not always showing every line, only gets all lines on page refresh.
* Is STDERR being captured okay? Should intersperse with STDOUT, not be stuck at end.
* There's an upper limit (1,000? - set-able value?) on the number of lines storable by the "output" text box in the user interface, add paging.
  * Pagin added, just need ability to go back / forward to other pages.
* API - Check for ".." paths.
* Output - does output.html get displayed properly?
  * Make sure MIME type is set properly.
  * Default to any file found in www folder to be served as root if index.html not available.
  * Output iframe - seems to be cached version of output sometimes, needs explicit updating somehow.
* Logging - add date / time signature to start of log output.
* Add / check "error code on exit" message for non-0 results.
* Return error message if task file doesn't run, don't just sit.

## Features

* Error reporting for Tasks
  * Via email. Report failures (as given by exit code).
* Optional ability to stop Task.
  * Users with Editor rights can cancel. Runner who runs Task gets permissions to cancel, not other Runners.
* Inputs from STDIN.
  * Needs unique ID per input on page, otherwise browser auto-fill gets very full
  * Single line text box
  * Radio select
  * Dropdown
  * Typeahead dropdown
  * File upload
  * Photo capture
* Handle values passed in standard Webhook headers, pass those to Task's script.
  * Queue size - 0 for none, 1, or "many".
  * Helper script to clone Swagger API to local SQLite database
* Authentication - replace "Error: Unauthorised" page with nicer, more handy "Unauthorised - login here" page.
* Authentication support
  * Tailscale
  * Caddy (SSO features)
  * Implement own Google / MS / Apple Oauth2?
  * Email code
* Edit Mode
  * Set-able editRoot value to limit access to sub-folder.
  * Highlight current/unsaved files, stop exit / run until files saved / discarded.
  * Upload of folders.
  * Rename of folders.
  * Add diff / versioning for file edits.
    * addDiff
    * getDiffList
    * resetToDiff
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
* Better admin console.
  * UI created by own capabilities!
  * Task run schedualer (cross-platform, write back to both cron and Windows Task Schedular), with error reporting if tasks fail.
  * Add "New Task" dialog, with (configurable) pre-defined "Type" field for quick starts:
    * Git checkout
    * Hugo
    * Jekyll
    * 11ty
    * Gatsby
    * Observable
    * Docs-To-Markdown
    * How-To
    * FAQ
    * Dashboard
    * Slideshow
    * Yearbook
    * Certificates
    * Start Menu
    * OCR Layer to PDFs
    * Mailmerge
    * Audio Cues page (plus looper!)
    * Flask App
      * Globals
      * Favicon
      * Oauth2 Auth via redirect?
  * Ability to rename / delete Tasks.
  * Ability to connect cloud storage.
* GitHub Actions template to download for individual Tasks
  * Should just be able to use a single template file and insert Task ID with Curl command.
* Add Mac support in install.sh.
* Add ChromeOS support in install.sh.
* Bootstrap keywords for formatting via STDOUT.
