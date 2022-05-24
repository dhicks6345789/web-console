# Installation

## Installing Specific Releases

You can pass an argument to the "install" script to tell it to install a specific release:

### Mac / Linux (including Raspberry Pi)
```
curl -s https://www.sansay.co.uk/web-console/install.sh | sudo bash -s -- 0.1-beta.2
```

### Windows
```
powershell -command "& {&'Invoke-WebRequest' -Uri https://www.sansay.co.uk/web-console/install.bat -OutFile install.bat}" && install.bat 0.1-beta.2 && erase install.bat
```

If you use a parameter of "nightly" as the version, the latest version built nightly from the Github source (might have bugs) will be installed:

### Mac / Linux (including Raspberry Pi)
```
curl -s https://www.sansay.co.uk/web-console/install.sh | sudo bash -s -- nightly
```

### Windows
```
powershell -command "& {&'Invoke-WebRequest' -Uri https://www.sansay.co.uk/web-console/install.bat -OutFile install.bat}" && install.bat nightly && erase install.bat
```

You can manually download binary and source packages from the Github [releases page](https://github.com/dhicks6345789/web-console/releases). If you want the very latest version, you can download nightly builds:

| Platform         | Binary
| ---------------- | ----------------------------------------------------------------------- |
| Windows 32-bit   | [Download](https://www.sansay.co.uk/web-console/binaries/win-386.exe)   |
| Windows 64-bit   | [Download](https://www.sansay.co.uk/web-console/binaries/win-amd64.exe) |
| Mac              | [Download](https://www.sansay.co.uk/web-console/binaries/darwin-amd64)  |
| Linux 32-bit     | [Download](https://www.sansay.co.uk/web-console/binaries/linux-386)     |
| Linux 64-bit     | [Download](https://www.sansay.co.uk/web-console/binaries/linux-amd64)   |
| Linux ARM 32-bit | [Download](https://www.sansay.co.uk/web-console/binaries/linux-arm32)   |
| Linux ARM 64-bit | [Download](https://www.sansay.co.uk/web-console/binaries/linux-arm64)   |

As well as the appropriate binary for your platform (place in `/usr/local/bin` on Linux, `C:\Program Files\WebConsole` on Windows), you'll need the contents of the "www" folder (place in `/etc/webconsole/www` on Linux, `C:\Program Files\WebConsole\www` on Windows), available as a [zip file](https://www.sansay.co.uk/web-console/web-console-nightly.zip) for Windows or a [.tar.gz archive](https://www.sansay.co.uk/web-console/web-console-nightly.tar.gz) for MacOS and Linux.

## Building From Source

The source code is available on [Github](https://github.com/dhicks6345789/web-console). Written in Go, the source should be compileable on most platforms with a Go development environment available - the platform's default Go installation is generally fine.

Webconsole depends on the following libraries:
- [Resize](github.com/nfnt/resize): Simple bitmap image resizing library. Used in the implementation of favicons.
- [Gotrace](github.com/dennwc/gotrace): A Go implentation of [Potrace](http://potrace.sourceforge.net/), for tracing bitmaps to SVG files. Used in the implementation of favicons.
- [Golang-Image-ICO](github.com/kodeworks/golang-image-ico): An .ICO format image encoder. Used in the implementation of favicons.
- [Bcrypt](golang.org/x/crypto/bcrypt): For password hashing. Used for basic authentication.
- [Excelize](github.com/360EntSecGroup-Skylar/excelize): For loading Excel files.

A simple bash [build script](https://github.com/dhicks6345789/web-console/blob/master/build.sh) is available in the root of the source tree (or a [batch file](https://github.com/dhicks6345789/web-console/blob/master/build.bat) if you're building on Windows).
