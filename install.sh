VERSION="0.1-alpha"
echo Installing Web Console $VERSION...
# First, stop any existing Webconsole service...
systemctl stop webconsole
# ...get the binary file to run...
curl https://www.sansay.co.uk/web-console/binaries/linux-amd64 -o /usr/local/bin/webconsole
chmod u+x /usr/local/bin/webconsole
[ ! -d /etc/webconsole ] && mkdir /etc/webconsole
curl -L -s https://github.com/dhicks6345789/web-console/archive/v$VERSION.tar.gz | tar xvz
cp -r web-console-$VERSION/www /usr/share/caddy/web-console

[ ! -d /etc/webconsole/www ] && mkdir /etc/webconsole/www
#curl -s https://www.sansay.co.uk/webconsole/www/browserconfig.xml -o /etc/webconsole/browserconfig.xml
#curl -s https://www.sansay.co.uk/webconsole/www/copyIcon.svg -o /etc/webconsole/copyIcon.svg
#curl -s https://www.sansay.co.uk/webconsole/www/favicon.png -o /etc/webconsole/favicon.png
#curl -s https://www.sansay.co.uk/webconsole/www/formatting.js -o /etc/webconsole/formatting.js
#curl -s https://www.sansay.co.uk/webconsole/www/index.html -o /etc/webconsole/index.html
#curl -s https://www.sansay.co.uk/webconsole/www/site.webmanifest -o /etc/webconsole/site.webmanifest
#curl -s https://www.sansay.co.uk/webconsole/www/webconsole.html -o /etc/webconsole/webconsole.html
#[ ! -d /etc/webconsole/www/favicons ] && mkdir /etc/webconsole/www/favicons
#curl -s https://www.sansay.co.uk/binaries/webconsole/www/favicons/apple.html -o /etc/webconsole/favicons/apple.html
#curl -s https://www.sansay.co.uk/binaries/webconsole/www/favicons/banana.html -o /etc/webconsole/favicons/banana.html

# ...set up systemd to run Webconsole.
#curl https://www.sansay.co.uk/webconsole/webconsole.service -o /etc/systemd/system/webconsole.service
#chmod 644 /etc/systemd/system/webconsole.service
systemctl start webconsole
systemctl enable webconsole
