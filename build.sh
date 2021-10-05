#!/bin/bash
TARGET_DIR=${HOME}/tools

mkdir -p $TARGET_DIR
# Obtain version number from git
VERSION=$(git describe --abbrev --dirty --always --tags)
VERSION=${VERSION#"v"}

echo 'Building mzrecal for Linux/amd64'
GOOS=linux GOARCH=amd64 go build -trimpath -a -ldflags "-extldflags \"-static\" -buildid= -X main.progVersion=${VERSION}" -o $TARGET_DIR/mzrecal

echo 'Building mzrecal for Windows/amd64'
GOOS=windows GOARCH=amd64 go build -trimpath -a -ldflags "-extldflags \"-static\" -buildid= -X main.progVersion=${VERSION}" -o $TARGET_DIR/mzrecal.exe
