#!/bin/bash
TARGET_DIR=~/tools

mkdir -p $TARGET_DIR
# Build for Linux
VERSION=$(git describe --abbrev --dirty --always --tags)

echo go build -a -ldflags "-extldflags \"-static\" -X main.progVersion=${VERSION}" -o $TARGET_DIR/mzrecal

# Cross-build for Windows (does not work due to CGO code)
# GOOS=windows GOARCH=amd64 go build -o ~/mzrecal/mzrecal.exe
