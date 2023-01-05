@echo off

net stop WebConsole > nul 2>&1
erase "C:\Program Files\WebConsole\webconsole.exe" > nul 2>&1
erase webconsole.exe > nul 2>&1

echo Checking libraries are installed...
go install github.com/nfnt/resize@latest
go get github.com/nfnt/resize

go install github.com/dennwc/gotrace@latest
go get github.com/dennwc/gotrace

go install github.com/kodeworks/golang-image-ico@latest
go get github.com/kodeworks/golang-image-ico

go install golang.org/x/crypto/bcrypt@latest
go get golang.org/x/crypto/bcrypt

go install github.com/360EntSecGroup-Skylar/excelize
go get github.com/360EntSecGroup-Skylar/excelize

go install golang.org/x/crypto/argon2@latest
go get golang.org/x/crypto/argon2

echo Building...
go build webconsole.go

copy webconsole.exe "C:\Program Files\WebConsole" > nul 2>&1
xcopy /E /Y www "C:\Program Files\WebConsole\www" > nul 2>&1
net start WebConsole > nul 2>&1

rem call install.bat --key somekey --subdomain somethinggoeshere
rem call install.bat --go
rem net start TunnelTo
