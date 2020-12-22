#!/bin/bash
TARGET_DIR=${HOME}/tools

mkdir -p $TARGET_DIR
VERSION=$(git describe --abbrev --dirty --always --tags)
VERSION=${VERSION#"v"}

# Build for Linux
go build -a -ldflags "-extldflags \"-static\" -X main.progVersion=${VERSION}" -o $TARGET_DIR/mzrecal

## Cross-build for Windows (64 bit)
echo Installing cross compiler
sudo apt install mingw-w64
export GSL_INSTALL="${PWD}/windows/gsl"
echo Building mzrecal
CGO_CFLAGS="-I${GSL_INSTALL}/include" CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build -ldflags "-extldflags \"-static -L${GSL_INSTALL}\" -X main.progVersion=${VERSION}"  -o $TARGET_DIR/mzrecal.exe
