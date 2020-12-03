@echo off

net stop WebConsole
erase "C:\Program Files\WebConsole\webconsole.exe" > nul 2>&1
erase webconsole.exe > nul 2>&1

go get github.com/nfnt/resize
go get github.com/dennwc/gotrace
go get github.com/kodeworks/golang-image-ico
go get golang.org/x/crypto/bcrypt
go get github.com/360EntSecGroup-Skylar/excelize
go build webconsole.go

copy webconsole.exe "C:\Program Files\WebConsole"
net start WebConsole

rem call install.bat --key somekey --subdomain somethinggoeshere
rem call install.bat --go
rem net start TunnelTo
