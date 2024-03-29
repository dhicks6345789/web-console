VERSION="nightly"

if [ ! -z "$1" ]
then
  VERSION=$1
fi

echo Installing Web Console $VERSION...

# Work out what architecture we are installing on.
ARCH=$(uname -m)
BINARY=linux-amd64
[[ $ARCH == arm* ]] && BINARY=linux-arm32
[[ $ARCH == aarch64 ]] && BINARY=linux-arm64

# Stop any existing Webconsole service.
systemctl stop webconsole

# Download the appropriate binary file and make sure it's executable.
echo Downloading binary for $BINARY...
if [[ $VERSION == nightly ]]
then
  curl -L -s https://www.sansay.co.uk/web-console/binaries/$BINARY -o /usr/local/bin/webconsole
else
  curl -L -s https://github.com/dhicks6345789/web-console/releases/download/v$VERSION/$BINARY -o /usr/local/bin/webconsole
fi
chmod u+x /usr/local/bin/webconsole

# Download the support files bundle and un-bundle it.
if [[ $VERSION == nightly ]]
then
  mkdir web-console-nightly
  curl -L -s https://www.sansay.co.uk/web-console/web-console-nightly.tar.gz | tar xz -C web-console-nightly
else
  curl -L -s https://github.com/dhicks6345789/web-console/archive/v$VERSION.tar.gz | tar xz
fi

# Create the application's data folder and copy the default data files into it.
[ ! -d /etc/webconsole ] && mkdir /etc/webconsole
[ ! -d /etc/webconsole/tasks ] && mkdir /etc/webconsole/tasks
[ ! -d /etc/webconsole/www ] && mkdir /etc/webconsole/www
cp -r web-console-$VERSION/www/* /etc/webconsole/www

# Set up systemd to run Webconsole.
cp web-console-$VERSION/webconsole.service /etc/systemd/system/webconsole.service
chmod 644 /etc/systemd/system/webconsole.service
systemctl start webconsole
systemctl enable webconsole

# Clear out the temporary bundle folder.
rm -rf web-console-$VERSION
