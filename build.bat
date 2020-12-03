@echo off
cls

erase webconsole.exe > nul 2>&1
net stop WebConsole
erase "C:\Program Files\WebConsole\webconsole.exe" > nul 2>&1
go build webconsole.go

rem call install.bat --key Usff7rA5eSFtno9kpn9GSP --subdomain somethinggoeshere
rem call install.bat --go
rem net start TunnelTo

rem echo Running...
rem cd "C:\Program Files\WebConsole"
rem webconsole 2>&1
