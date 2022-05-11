# Build script for WebConsole. Tested on Debian, x86 and Raspberry Pi.

echo Building Web Console...

# Stop any existing running service.
systemctl stop webconsole

go get github.com/nfnt/resize
go get github.com/dennwc/gotrace
go get github.com/kodeworks/golang-image-ico
go get golang.org/x/crypto/bcrypt
go get github.com/360EntSecGroup-Skylar/excelize
go build webconsole.go
cp webconsole /usr/local/bin

# Create the application's data folder and copy the default data files into it.
[ ! -d /etc/webconsole ] && mkdir /etc/webconsole
[ ! -d /etc/webconsole/tasks ] && mkdir /etc/webconsole/tasks
[ ! -d /etc/webconsole/www ] && mkdir /etc/webconsole/www
cp -r www/* /etc/webconsole/www

# Set up systemd to run Webconsole, if it isn't already.
[ ! -f /etc/systemd/system/webconsole.service ] && cp webconsole.service /etc/systemd/system/webconsole.service && chmod 644 /etc/systemd/system/webconsole.service

# Restart the webconsole service.
systemctl start webconsole
systemctl enable webconsole
