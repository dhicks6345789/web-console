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
[ ! -d /etc/webconsole ] && mkdir /etc/webconsole
cp --recursive www /etc/webconsole

# Restart the webconsole service (if it exists).
systemctl start webconsole
systemctl enable webconsole
