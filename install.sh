echo Installing Web Console...
# First, stop any existing Webconsole service...
systemctl stop webconsole
# ...get the binary file to run...
curl https://www.sansay.co.uk/binaries/web-console/linux-amd64/webconsole -o /usr/local/bin/webconsole
chmod u+x /usr/local/bin/webconsole
[ ! -d /etc/webconsole ] && mkdir /etc/webconsole
[ ! -d /etc/webconsole/www ] && mkdir /etc/webconsole/www
curl https://www.sansay.co.uk/binaries/web-console/www/browserconfig.xml -o /etc/webconsole/browserconfig.xml
curl https://www.sansay.co.uk/binaries/web-console/www/copyIcon.svg -o /etc/webconsole/copyIcon.svg
curl https://www.sansay.co.uk/binaries/web-console/www/favicon.png -o /etc/webconsole/favicon.png
curl https://www.sansay.co.uk/binaries/web-console/www/formatting.js -o /etc/webconsole/formatting.js
curl https://www.sansay.co.uk/binaries/web-console/www/index.html -o /etc/webconsole/index.html
curl https://www.sansay.co.uk/binaries/web-console/www/site.webmanifest -o /etc/webconsole/site.webmanifest
curl https://www.sansay.co.uk/binaries/web-console/www/webconsole.html -o /etc/webconsole/webconsole.html
[ ! -d /etc/webconsole/www/favicons ] && mkdir /etc/webconsole/www/favicons
curl https://www.sansay.co.uk/binaries/web-console/www/favicons/apple.html -o /etc/webconsole/favicons/apple.html
curl https://www.sansay.co.uk/binaries/web-console/www/favicons/banana.html -o /etc/webconsole/favicons/banana.html
#curl https://www.sansay.co.uk/binaries/web-console/www/favicons/.html /etc/webconsole/favicons/.html

# ...set up systemd to run Webconsole.
curl https://www.sansay.co.uk/binaries/web-console/webconsole.service -o /etc/systemd/system/webconsole.service
chmod 644 /etc/systemd/system/webconsole.service
systemctl start webconsole
systemctl enable webconsole
