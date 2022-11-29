# Alternatives

## Web Console Is Not:

- A web-based terminal emulator - you might want to try [Shell In A Box](https://github.com/shellinabox/shellinabox).

- A web-based remote access system - we can recommend [Apache Guacamole](https://guacamole.apache.org/) (and our own [Remote Gateway](https://github.com/dhicks6345789/remote-gateway) project) for web-based access to RDP, SSH and VNC sessions, or [noVNV](https://novnc.com/info.html) for web-based VNC access.

Web Console is not a [low-code](https://en.wikipedia.org/wiki/Low-code_development_platform) development tool - it can remove some complexity around the implementation and hosting of web applications with a basic GUI, but will still need some (hopefully, fairly beginner-level) coding skills.

Web Console is not a full Integrated Development Environment like [Visual Studio](https://visualstudio.microsoft.com/) or [Eclipse](https://www.eclipse.org/). It does have a basic integrated code editing facility, though.

Web Console, importantly, runs code server-side, not in the browser as something like [Replit](https://replit.com/) does.

Web Console is not a library for a specific language, it implements a GUI by interpreting the text (STDOUT, STDERR) output from command-line applications - it should be very much cross-platform. You can even use it to add a web-based graphical user interface to an old DOS-era .BAT file, if you want.

Web Console is not a comprehensive development framework - it isn't suitible for larger applications, and its GUI capabilities aren't on a par with something like [React](https://reactjs.org/). If you want to be implementing proper, modern web applications, Web Console isn't the application for you. It should, however, allow you to write quick-and-dirty scripts that can take some basic user input and get them in front of users in minimal time.

Web Console is not a replacement for [GitHub Actions](https://github.com/features/actions), although Web Console Tasks can easily be triggered from GitHub Actions, so it can be a useful tool to run some code locally on a server when something happens on a repository.

https://glitch.com/
