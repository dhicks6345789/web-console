@echo off

net stop WebConsole > nul 2>&1
erase "C:\Program Files\WebConsole\webconsole.exe" > nul 2>&1
erase webconsole.exe > nul 2>&1

echo Checking libraries are installed...
go get github.com/nfnt/resize
go get github.com/dennwc/gotrace
go get github.com/kodeworks/golang-image-ico
go get golang.org/x/crypto/bcrypt
go get github.com/xuri/excelize/v2
go get golang.org/x/crypto/argon2

echo Building...
go build webconsole.go

copy webconsole.exe "C:\Program Files\WebConsole" > nul 2>&1
xcopy /E /Y www "C:\Program Files\WebConsole\www" > nul 2>&1
xcopy /E /Y ..\ace-builds\src-noconflict "C:\Program Files\WebConsole\www\ace" > nul 2>&1
net start WebConsole > nul 2>&1
