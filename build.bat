@echo off
cls

erase webconsole.exe > nul 2>&1
go get github.com/nfnt/resize
go get github.com/dennwc/gotrace
go get github.com/kodeworks/golang-image-ico
go get golang.org/x/crypto/bcrypt
go get github.com/360EntSecGroup-Skylar/excelize
go build webconsole.go

rem call install.bat --key Usff7rA5eSFtno9kpn9GSP --subdomain somethinggoeshere
rem call install.bat --go
rem net start TunnelTo

rem echo Running...
rem cd "C:\Program Files\WebConsole"
rem webconsole 2>&1
