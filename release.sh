#!/bin/bash
source ./build.sh

# Create release directory
mkdir -p release

# Copy executables to release directory
cp ${HOME}/tools/mzrecal release
cp ${HOME}/tools/mzrecal.exe release

# Create zip file
cd release
zip mzrecal-${VERSION}_Windows-64bit.zip mzrecal.exe

# Create tar.gz file
tar -czf mzrecal-${VERSION}_Linux-64bit.tar.gz mzrecal

# delete executables
rm mzrecal mzrecal.exe

