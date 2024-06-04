#!/bin/bash
TARGET_DIR=${HOME}/tools

mkdir -p $TARGET_DIR
# Obtain version number from git
VERSION=$(git describe --abbrev --dirty --always --tags)
VERSION=${VERSION#"v"}

# build flags are set to creat a binairy reproducable, fully statically linked executable that includes the Git version number
# -trimpath: don't inlude full source code path names in the executable. Needed to produce binairy reproducable output.
# -ldflags:
#   -buildid= : clear the buildid, needed to produce binairy reproducable output
#   -X main.progVersion=${VERSION} : include Git version info (from enviroment variable ${VERSION}).
#       This will probably be replaced by using function debug.ReadBuildInfo() from package "runtime/debug" in the future,
#       so that we don't need this build script anymore

echo 'Building mzrecal for Linux/amd64'
GOOS=linux GOARCH=amd64 go build -trimpath -a -ldflags "-buildid= -X main.progVersion=${VERSION}" -o $TARGET_DIR/mzrecal

echo 'Building mzrecal for Windows/amd64'
GOOS=windows GOARCH=amd64 go build -trimpath -a -ldflags "-buildid= -X main.progVersion=${VERSION}" -o $TARGET_DIR/mzrecal.exe

# Generate zip/tar.gz files
# Create a temporary directory
TMPDIR=$(mktemp -d)
# Copy the executable to the temporary directory
cp $TARGET_DIR/mzrecal.exe $TMPDIR
mkdir $TMPDIR/mzrecal
cp $TARGET_DIR/mzrecal $TMPDIR/mzrecal/

cd $TMPDIR
# Check if running on Windows
if [ -n "$WINDIR" ]; then
    # Running on Windows
    "/c/Program Files/7-Zip/7z.exe" a mzrecal-${VERSION}_windows-64bit.zip mzrecal.exe
    "/c/Program Files/7-Zip/7z.exe" a mzrecal-${VERSION}_linux-64bit.tar mzrecal
    "/c/Program Files/7-Zip/7z.exe" a mzrecal-${VERSION}_linux-64bit.tar.gz mzrecal-${VERSION}_linux-64bit.tar
else
    zip mzrecal-${VERSION}_windows-64bit.zip mzrecal.exe
    tar -czf mzrecal-${VERSION}_linux-64bit.tar.gz mzrecal
fi

# Copy the zip/tar.gz files to the target directory
cp mzrecal-${VERSION}_windows-64bit.zip $TARGET_DIR
cp mzrecal-${VERSION}_linux-64bit.tar.gz $TARGET_DIR

# Remove the temporary directory
cd ~
rm -rf $TMPDIR
