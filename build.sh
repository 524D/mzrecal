#!/bin/bash
TARGET_DIR=~/tools

mkdir -p $TARGET_DIR
VERSION=$(git describe --abbrev --dirty --always --tags)
VERSION=${VERSION#"v"}

# Build for Linux
go build -a -ldflags "-extldflags \"-static\" -X main.progVersion=${VERSION}" -o $TARGET_DIR/mzrecal

## Cross-build for Windows (64 bit)
echo Install cross compiler
sudo apt install mingw-w64
echo Download gsl source
MZRECALDIR=$PWD
export GSL_WIN="${PWD}/gsl-windows"
export GSL_INSTALL="${GSL_WIN}/install"
mkdir -p "${GSL_WIN}"
cd "${GSL_WIN}"
wget -c https://ftp.gnu.org/gnu/gsl/gsl-2.6.tar.gz
tar -xf gsl-2.6.tar.gz
echo Build GSL for Windows
cd gsl-2.6
./configure --host=x86_64-w64-mingw32 --prefix="${GSL_INSTALL}"
make -j 9
make install
echo Build rzrecal
cd $MZRECALDIR
CGO_CFLAGS="-I ${GSL_INSTALL}/include" CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build -ldflags "-extldflags \"-static -L/home/robm/gsl_windows/lib/\" -X main.progVersion=${VERSION}"  -o $TARGET_DIR/mzrecal.exe
