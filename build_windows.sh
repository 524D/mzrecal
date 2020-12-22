#!/bin/bash
# This script build mzRecal on Windows
# Check README.md for required software

# Obtain the version number from GIT
VERSION=$(git describe --abbrev --dirty --always --tags)
VERSION=${VERSION#"v"}

# The location of the pre-build Windows GSL code
export GSL_INSTALL="${PWD}/windows/gsl"

echo Building mzrecal
CGO_CFLAGS="-I${GSL_INSTALL}/include" CGO_ENABLED=1 CC=/c/TDM-GCC-64/bin/gcc GOOS=windows GOARCH=amd64 /c/go/bin/go build -ldflags "-extldflags \"-static -L${GSL_INSTALL}\" -X main.progVersion=${VERSION}"
