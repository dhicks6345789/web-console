# Build script for WebConsole. Tested on Debian, x86 and Raspberry Pi.

source VERSION
CURRENTDATE=`date +"%d/%m/%Y-%H:%M"`
BUILDVERSION="$VERSION-local-$CURRENTDATE"

echo Building Web Console...

# Stop any existing running service.
systemctl stop webconsole

go get github.com/nfnt/resize
go get github.com/dennwc/gotrace
go get github.com/kodeworks/golang-image-ico
go get golang.org/x/crypto/bcrypt
go get golang.org/x/crypto/argon2@v0.14.0
## go get github.com/360EntSecGroup-Skylar/excelize
go get github.com/xuri/excelize/v2

go build -ldflags "-X main.buildVersion=$BUILDVERSION" webconsole.go
cp webconsole /usr/local/bin

# Create the application's data folder and copy the default data files into it.
[ ! -d /etc/webconsole ] && mkdir /etc/webconsole
[ ! -d /etc/webconsole/tasks ] && mkdir /etc/webconsole/tasks
[ ! -d /etc/webconsole/www ] && mkdir /etc/webconsole/www
[ ! -d /etc/webconsole/www/ace ] && mkdir /etc/webconsole/www/ace
cp -r www/* /etc/webconsole/www
cp -r ../ace-builds/src-noconflict/* /etc/webconsole/www/ace

# Set up systemd to run Webconsole, if it isn't already.
[ ! -f /etc/systemd/system/webconsole.service ] && cp webconsole.service /etc/systemd/system/webconsole.service && chmod 644 /etc/systemd/system/webconsole.service

# Restart the webconsole service.
systemctl start webconsole
systemctl enable webconsole
