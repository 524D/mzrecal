#!/bin/bash
TARGET_DIR=${HOME}/tools

mkdir -p $TARGET_DIR
# Obtain version number from git
VERSION=$(git describe --abbrev --dirty --always --tags)
VERSION=${VERSION#"v"}

echo 'Building mzrecal for Linux'
go build -a -ldflags "-extldflags \"-static\" -X main.progVersion=${VERSION}" -o $TARGET_DIR/mzrecal

## Cross-build for Windows (64 bit)
echo 'Building mzrecal for Windows'
GOOS=windows GOARCH=amd64 go build -a -ldflags "-extldflags \"-static\" -X main.progVersion=${VERSION}" -o $TARGET_DIR/mzrecal.exe
