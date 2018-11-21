#!/bin/bash
TARGET_DIR=~/mzrecal

mkdir -p $TARGET_DIR
# Build for Linux
go build -a -ldflags '-extldflags "-static"' -o $TARGET_DIR/mzrecal

# Cross-build for Windows (does not work due to CGO code)
# GOOS=windows GOARCH=amd64 go build -o ~/mzrecal/mzrecal.exe
