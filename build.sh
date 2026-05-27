echo "Build script for WebConsole. Tested on Debian, x86 and Raspberry Pi."
echo "\"build -help\" for more details."

# Set default option values.
INSTALL=false
SERVICE=false

# Read user-defined command-line flags.
while test $# -gt 0; do
    case "$1" in
        -install)
            shift
            INSTALL=true
            ;;
        -service)
            shift
            INSTALL=true
            SERVICE=true
            ;;
        -help)
            echo "Builds the Web Console binary from Go source. Options:"
            echo "-install"
            echo "   Installs the binary in /usr/local/bin and supporting files in /etc/webconsole."
            echo "-service"
            echo "   Sets Web Console up as a service."
            exit 1
            ;;
        *)
            echo "$1 is not a recognized flag."
            exit 1;
            ;;
    esac
done

source VERSION
CURRENTDATE=`date +"%d/%m/%Y-%H:%M"`
BUILDVERSION="$VERSION-local-$CURRENTDATE"

if [ $SERVICE = true ]; then
  # Stop any existing running service.
  systemctl stop webconsole
fi

go get github.com/nfnt/resize
go get github.com/dennwc/gotrace
go get github.com/kodeworks/golang-image-ico
go get golang.org/x/crypto/bcrypt
go get golang.org/x/crypto/argon2@v0.14.0
go get github.com/xuri/excelize/v2

# Clear out any previously-compile binary.
rm webconsole

# Build the executable.
go build -ldflags "-X main.buildVersion=$BUILDVERSION" webconsole.go

# Exit if we didn't manage to build the executable.
[ ! -f webconsole ] && { echo "Error: webconsole not compiled."; exit 1; }

if [ $INSTALL = true ]; then
  # Install the new executable in place.
  cp webconsole /usr/local/bin

  # Create the application's data folder and copy the default data files into it.
  [ ! -d /etc/webconsole ] && mkdir /etc/webconsole
  [ ! -d /etc/webconsole/tasks ] && mkdir /etc/webconsole/tasks
  [ ! -d /etc/webconsole/www ] && mkdir /etc/webconsole/www
  [ ! -d /etc/webconsole/www/ace ] && mkdir /etc/webconsole/www/ace
  cp -r www/* /etc/webconsole/www
  cp -r ../ace-builds/src-noconflict/* /etc/webconsole/www/ace
fi

if [ $SERVICE = true ]; then
  # Set up systemd to run Webconsole, if it isn't already.
  [ ! -f /etc/systemd/system/webconsole.service ] && cp webconsole.service /etc/systemd/system/webconsole.service && chmod 644 /etc/systemd/system/webconsole.service

  # Restart the webconsole service.
  systemctl start webconsole
  systemctl enable webconsole
fi
