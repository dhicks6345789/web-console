echo Installing Web Console...
systemctl stop webconsole
curl https://www.sansay.co.uk/binaries/web-console/linux-amd64/webconsole -O /usr/local/bin/webconsole
chmod u+x /usr/local/bin/webconsole
[ ! -d /etc/webconsole ] && mkdir /etc/webconsole
#cp --recursive www /etc/webconsole

## Set up Webconsole as a systemd service - first, stop any existing Webconsole service...
#systemctl stop webconsole
## ...then set up systemd to run Webconsole.
#cp webconsole.service /etc/systemd/system/webconsole.service
#chmod 644 /etc/systemd/system/webconsole.service
systemctl start webconsole
systemctl enable webconsole
