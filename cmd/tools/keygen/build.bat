@echo off

setlocal

set CGO_ENABLED=0

echo | set /p=Building Windows 64-bit...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o ./dist/keygen.exe ./keygen.go
echo  done