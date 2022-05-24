Webconsole - a simple way to turn a command line application into a web app.
Runs as a simple web server to host Task pages that allow the end-user to
simply click a button to run a batch / script / etc file. Note that by itself,
Webconsole doesn't handle HTTPS. If you are installing on a world-facing server
you should use a proxy server that handles HTTPS - we recommend Caddy as it
will automatically handle Let's Encrypt certificates. If you are behind a
firewall then we recommend tunnelto.dev, giving you an HTTPS-secured URL to
access. Both options can be installed via the install.bat / install.sh
scripts.

Usage: webconsole [--new] [--list] [--start] [--localOnly true/false] [--port int] [--config path] [--webroot path] [--taskroot path]

--new: creates a new Task. Each Task has a unique 16-character ID which can be
  passed as part of the URL or via a POST request, so for basic security you
  can give a user a URL with an embedded ID. Use an external authentication
  service for better security.
  
--list: prints a list of existing Tasks.

--start: runs as a web server, waiting for requests. Logs are printed straight to
  stdout - hit Ctrl-C to quit. By itself, the start command can be handy for
  quickly debugging. Run install.bat / install.sh to create a Windows service or
  Linux / MacOS deamon.
  
--debug: like "start", but prints more information.

--localOnly: default is "true", in which case the built-in webserver will only
  respond to requests from the local server.
  
--port: the port number the web server should listen out on. Defaults to 8090.

--config: where to find the config file. By default, on Linux this is
  /etc/webconsole/config.csv.
  
--webroot: the folder to use for the web root.

--taskroot: the folder to use to store Tasks.
