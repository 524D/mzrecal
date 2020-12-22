#!/bin/bash
# This script pre-builds the GSL library for Windows >>> ON LINUX <<<
# IT IS NOT NEEDED TO RUN THIS, because we distribute relavant parts of
# the pre-build GSL library for Windows in the mzrecal repository
# The library is cross-compiled on Linux because that is most simple.

echo 'Running this script should not be needed!'
echo 'It will download/cross compile the GSL library for Windows (on Linux)'
read -p "Are you sure? " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]
then
    echo Install cross compiler
    sudo apt install mingw-w64
    echo Download gsl source
    MZRECALDIR=$PWD
    export GSL_WIN="${PWD}/windows"
    export GSL_INSTALL="${GSL_WIN}/gsl"
    mkdir -p "${GSL_WIN}"
    cd "${GSL_WIN}"
    wget -c https://ftp.gnu.org/gnu/gsl/gsl-2.6.tar.gz
    tar -xf gsl-2.6.tar.gz
    echo Build GSL for Windows
    cd gsl-2.6
    ./configure --host=x86_64-w64-mingw32 --prefix="${GSL_INSTALL}"
    make -j 9
    make install
fi

