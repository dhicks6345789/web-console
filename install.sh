echo Installing Web Console...
go get golang.org/x/crypto/bcrypt
go get github.com/360EntSecGroup-Skylar/excelize
go build webconsole.go
cp webconsole /usr/local/bin
