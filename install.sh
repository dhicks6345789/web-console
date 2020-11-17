echo Installing Web Console...
go get github.com/nfnt/resize
go get github.com/dennwc/gotrace
go get golang.org/x/crypto/bcrypt
go get github.com/360EntSecGroup-Skylar/excelize
go build webconsole.go
cp webconsole /usr/local/bin
[ ! -d /etc/webconsole ] && mkdir /etc/webconsole
cp --recursive www /etc/webconsole

# Set up Webconsole as a systemd service - first, stop any existing Webconsole service...
systemctl stop webconsole
# ...then set up systemd to run Webconsole.
cp webconsole.service /etc/systemd/system/webconsole.service
chmod 644 /etc/systemd/system/webconsole.service
systemctl start webconsole
systemctl enable webconsole
