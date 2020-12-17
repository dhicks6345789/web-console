VERSION="0.1-beta"
echo Installing Web Console $VERSION...

# First, stop any existing Webconsole service.
systemctl stop webconsole

# Download the appropriate binary file and make sure it's executable.
curl -L -s https://github.com/dhicks6345789/web-console/releases/download/v$VERSION/linux-amd64 -o /usr/local/bin/webconsole
chmod u+x /usr/local/bin/webconsole

# Download the support files bundle and un-bundle it.
curl -L -s https://github.com/dhicks6345789/web-console/archive/v$VERSION.tar.gz | tar xz

# Create the application's data folder and copy the default data files into it.
[ ! -d /etc/webconsole ] && mkdir /etc/webconsole
[ ! -d /etc/webconsole/tasks ] && mkdir /etc/webconsole/tasks
[ ! -d /etc/webconsole/www ] && mkdir /etc/webconsole/www
cp -r web-console-$VERSION/www /etc/webconsole/www

# Set up systemd to run Webconsole.
cp web-console-$VERSION/webconsole.service /etc/systemd/system/webconsole.service
chmod 644 /etc/systemd/system/webconsole.service
systemctl start webconsole
systemctl enable webconsole

# Clear out the temporary bundle folder.
rm -rf web-console-$VERSION
