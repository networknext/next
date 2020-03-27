@echo off

setlocal

set CGO_ENABLED=0

echo | set /p=Building Linux 64-bit...
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -o ./keygen_linux ./src/keygen.go
echo  done

echo | set /p=Building FreeBSD 64-bit...
set GOOS=freebsd
set GOARCH=amd64
go build -ldflags="-s -w" -o ./keygen_freebsd ./src/keygen.go
echo  done

echo | set /p=Building OpenBSD 64-bit...
set GOOS=openbsd
set GOARCH=amd64
go build -ldflags="-s -w" -o ./keygen_openbsd ./src/keygen.go
echo  done

echo | set /p=Building ARM 64-bit...
set GOOS=linux
set GOARCH=arm64
go build -ldflags="-s -w" -o ./keygen_arm ./src/keygen.go
echo  done

echo | set /p=Building MacOS 64-bit...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags="-s -w" -o ./keygen_mac ./src/keygen.go
echo  done

echo | set /p=Building Windows 64-bit...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o ./keygen.exe ./src/keygen.go
echo  done